package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log/slog"

	"gopilot-cli/internal/agent/colors"
	"gopilot-cli/internal/agent/summarizer"
	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/logger"
	"gopilot-cli/internal/schema"
	"gopilot-cli/internal/tools"
	terminal "gopilot-cli/internal/utils/terminal"
)

//
// ============================================================
// Agent Structure
// ============================================================
//

type Agent struct {
	llm          *llm.Client
	systemPrompt string
	tools        map[string]tools.Tool
	maxSteps     int
	tokenLimit   int
	workspace    string

	messages []schema.Message
	log      *logger.AgentLogger
}

func NewAgent(
	client *llm.Client,
	systemPrompt string,
	toolList []tools.Tool,
	maxSteps int,
	workspace string,
	tokenLimit int,
) (*Agent, error) {

	wp := workspace
	if wp == "" {
		wp = "./workspace"
	}

	abs, _ := filepath.Abs(wp)
	_ = os.MkdirAll(abs, 0755)

	// 向系统提示注入 workspace 信息
	if !strings.Contains(systemPrompt, "Current Workspace") {
		systemPrompt += fmt.Sprintf(
			"\n\n## Current Workspace\nCurrent workspace: `%s`\nAll relative paths will resolve here.",
			abs,
		)
	}

	toolMap := map[string]tools.Tool{}
	for _, t := range toolList {
		toolMap[t.Name()] = t
	}

	ag := &Agent{
		llm:          client,
		systemPrompt: systemPrompt,
		tools:        toolMap,
		maxSteps:     maxSteps,
		tokenLimit:   tokenLimit,
		workspace:    abs,
		messages: []schema.Message{
			{Role: "system", Content: systemPrompt},
		},
	}

	log, err := logger.NewAgentLogger()
	if err != nil {
		return nil, err
	}
	ag.log = log

	return ag, nil
}

func (a *Agent) AddUserMessage(content string) {
	a.messages = append(a.messages, schema.Message{
		Role:    "user",
		Content: content,
	})
}

//
// ============================================================
// Main Run Loop
// ============================================================
//

func (a *Agent) Run(ctx context.Context) (string, error) {
	// 新建日志会话
	if err := a.log.StartNewRun(); err != nil {
		return "", err
	}

	fmt.Printf("%s📝 Log file: %s%s\n",
		colors.DIM, a.log.GetLogFilePath(), colors.RESET)

	step := 0
	msgSummarizer := summarizer.NewSummarizer(a.llm, a.tokenLimit)

	for step < a.maxSteps {

		// 触发摘要
		newMsgs, err := msgSummarizer.SummarizeMessages(ctx, a.messages)
		if err != nil {
			slog.Warn("Summarization failed", slog.String("err", err.Error()))
		} else {
			a.messages = newMsgs
		}

		// 打印 Step 框
		stepText := fmt.Sprintf("%s%s💭 Step %d/%d%s",
			colors.BOLD, colors.BRIGHT_CYAN, step+1, a.maxSteps, colors.RESET)
		width := terminal.CalculateDisplayWidth(stepText)
		box := 58
		padding := box - 1 - width

		fmt.Printf("\n%s╭%s╮%s\n", colors.DIM, strings.Repeat("─", box), colors.RESET)
		fmt.Printf("%s│%s %s%s%s│%s\n",
			colors.DIM, colors.RESET,
			stepText,
			strings.Repeat(" ", padding),
			colors.DIM, colors.RESET)
		fmt.Printf("%s╰%s╯%s\n",
			colors.DIM, strings.Repeat("─", box), colors.RESET)

		toolList := []tools.Tool{}
		for _, t := range a.tools {
			toolList = append(toolList, t)
		}
		reg := tools.NewToolRegistry()
		for _, t := range toolList {
			reg.Register(t)
		}

		// 日志：请求
		a.log.LogRequest(a.messages, toolList)

		// 调用模型
		resp, err := a.llm.Generate(ctx, a.messages, reg)
		if err != nil {
			fmt.Printf("\n%s❌ LLM Error: %s%s\n", colors.BRIGHT_RED, err.Error(), colors.RESET)
			return err.Error(), err
		}

		// 日志：响应
		a.log.LogResponse(
			resp.Content,
			resp.Thinking,
			resp.ToolCalls,
			resp.FinishReason,
		)

		// 加入 assistant 消息
		a.messages = append(a.messages, schema.Message{
			Role:      "assistant",
			Content:   resp.Content,
			Thinking:  resp.Thinking,
			ToolCalls: resp.ToolCalls,
		})

		// 打印思考
		if resp.Thinking != "" {
			fmt.Printf("\n%s🧠 Thinking:%s\n", colors.BOLD+colors.MAGENTA, colors.RESET)
			fmt.Println("%s%s%s\n", colors.DIM, resp.Thinking, colors.RESET)
		}

		// 打印模型输出
		if resp.Content != "" {
			fmt.Printf("\n%s🤖 Assistant:%s\n", colors.BOLD+colors.BRIGHT_BLUE, colors.RESET)
			fmt.Println(resp.Content)
		}

		// 若无工具调用，任务结束
		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		// =========================
		// 工具调用处理
		// =========================

		for _, tc := range resp.ToolCalls {
			fname := tc.Function.Name
			args := tc.Function.Arguments

			fmt.Printf("\n%s🔧 Tool Call:%s %s%s%s\n",
				colors.BRIGHT_YELLOW, colors.RESET, colors.BOLD, colors.CYAN, fname)

			// 打印参数
			fmt.Printf("%s   Arguments:%s\n", colors.DIM, colors.RESET)
			b, _ := json.MarshalIndent(args, "", "  ")
			for _, line := range strings.Split(string(b), "\n") {
				fmt.Printf("   %s%s%s\n", colors.DIM, line, colors.RESET)
			}

			tool, ok := a.tools[fname]
			var result *tools.ToolResult

			if !ok {
				result = &tools.ToolResult{
					Success: false,
					Error:   fmt.Sprintf("Unknown tool: %s", fname),
				}
			} else {
				result, err = tool.Execute(ctx, args)
				if err != nil {
					result = &tools.ToolResult{
						Success: false,
						Error:   err.Error(),
					}
				}
			}

			// 日志：工具调用
			a.log.LogToolResult(
				fname,
				args,
				result.Success,
				result.Content,
				result.Error,
			)

			// 打印执行结果
			if result.Success {
				text := result.Content
				if len(text) > 300 {
					text = text[:300] + colors.DIM + "..." + colors.RESET
				}
				fmt.Printf("%s✓ Result:%s %s\n", colors.BRIGHT_GREEN, colors.RESET, text)
			} else {
				fmt.Printf("%s✗ Error:%s %s%s%s\n",
					colors.BRIGHT_RED, colors.RESET, colors.RED, result.Error, colors.RESET)
			}

			// 添加到消息历史
			retval := result.Content
			if !result.Success {
				retval = "Error: " + result.Error
			}

			a.messages = append(a.messages, schema.Message{
				Role:       "tool",
				Content:    retval,
				ToolCallID: tc.ID,
				Name:       fname,
			})
		}

		step++
	}

	msg := fmt.Sprintf("Task could not complete in %d steps.", a.maxSteps)
	fmt.Printf("\n%s⚠️ %s%s\n", colors.BRIGHT_YELLOW, msg, colors.RESET)
	return msg, nil
}

func (a *Agent) History() []schema.Message {
	out := make([]schema.Message, len(a.messages))
	copy(out, a.messages)
	return out
}
