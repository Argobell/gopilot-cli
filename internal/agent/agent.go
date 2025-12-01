package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"log/slog"

	"github.com/pkoukk/tiktoken-go"

	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/logger"
	"gopilot-cli/internal/schema"
	"gopilot-cli/internal/tools"
	terminal "gopilot-cli/internal/utils/terminal"
)

//
// ============================================================
// Terminal Color Codes
// ============================================================
//

var (
	RESET = "\033[0m"
	BOLD  = "\033[1m"
	DIM   = "\033[2m"

	RED     = "\033[31m"
	GREEN   = "\033[32m"
	YELLOW  = "\033[33m"
	BLUE    = "\033[34m"
	MAGENTA = "\033[35m"
	CYAN    = "\033[36m"

	BRIGHT_RED     = "\033[91m"
	BRIGHT_GREEN   = "\033[92m"
	BRIGHT_YELLOW  = "\033[93m"
	BRIGHT_BLUE    = "\033[94m"
	BRIGHT_MAGENTA = "\033[95m"
	BRIGHT_CYAN    = "\033[96m"
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
// Token Estimation (tiktoken-go)
// ============================================================
//

func (a *Agent) estimateTokens() int {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return a.estimateTokensFallback()
	}

	total := 0

	for _, m := range a.messages {
		total += countTokens(enc, m.Content)
		total += countTokens(enc, m.Thinking)

		if len(m.ToolCalls) > 0 {
			raw := fmt.Sprintf("%v", m.ToolCalls)
			total += len(enc.Encode(raw, nil, nil))
		}

		total += 4 // metadata overhead
	}

	return total
}

func countTokens(enc *tiktoken.Tiktoken, text string) int {
	if text == "" {
		return 0
	}
	return len(enc.Encode(text, nil, nil))
}

func (a *Agent) estimateTokensFallback() int {
	total := 0
	for _, m := range a.messages {
		total += len(m.Content) + len(m.Thinking)
	}
	return total / 3
}

//
// ============================================================
// Message Summarization
// ============================================================
//

func (a *Agent) summarizeMessages(ctx context.Context) error {
	tokens := a.estimateTokens()
	if tokens <= a.tokenLimit {
		return nil
	}

	fmt.Printf("\n%s📊 Token estimate: %d/%d%s\n",
		BRIGHT_YELLOW, tokens, a.tokenLimit, RESET)
	fmt.Printf("%s🔄 Summarizing message history...%s\n", BRIGHT_YELLOW, RESET)

	// 收集所有 user 消息索引（跳过 system）
	userIdx := []int{}
	for i, m := range a.messages {
		if i > 0 && m.Role == "user" {
			userIdx = append(userIdx, i)
		}
	}
	if len(userIdx) == 0 {
		fmt.Printf("%s⚠️ No user messages to summarize%s\n", BRIGHT_YELLOW, RESET)
		return nil
	}

	var newMsgs []schema.Message
	newMsgs = append(newMsgs, a.messages[0]) // system

	summaryCount := 0

	for i, ui := range userIdx {
		newMsgs = append(newMsgs, a.messages[ui])

		var end int
		if i < len(userIdx)-1 {
			end = userIdx[i+1]
		} else {
			end = len(a.messages)
		}

		execMsgs := a.messages[ui+1 : end]
		if len(execMsgs) == 0 {
			continue
		}

		// 创建总结文本
		summary, err := a.createSummary(ctx, execMsgs, i+1)
		if err != nil {
			slog.Warn("Summary failed", slog.String("err", err.Error()))
			continue
		}

		newMsgs = append(newMsgs, schema.Message{
			Role:    "user",
			Content: "[Execution Summary]\n\n" + summary,
		})
		summaryCount++
	}

	a.messages = newMsgs
	newTokens := a.estimateTokens()
	fmt.Printf("%s✓ Summary complete (tokens %d → %d)%s\n",
		BRIGHT_GREEN, tokens, newTokens, RESET)

	return nil
}

func (a *Agent) createSummary(ctx context.Context, msgs []schema.Message, round int) (string, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Round %d execution process:\n\n", round))

	for _, m := range msgs {
		switch m.Role {
		case "assistant":
			sb.WriteString("Assistant: " + m.Content + "\n")
			if len(m.ToolCalls) > 0 {
				names := []string{}
				for _, tc := range m.ToolCalls {
					names = append(names, tc.Function.Name)
				}
				sb.WriteString("  → Called tools: " + strings.Join(names, ", ") + "\n")
			}
		case "tool":
			sb.WriteString("  ← Tool returned: " + m.Content + "\n")
		}
	}

	prompt := fmt.Sprintf(`
Please summarize the following agent execution process:

%s

Rules:
- Focus on what the agent did and which tools were used
- Concise, English, < 800 words
- Summarize execution only (no user content)
`, sb.String())

	req := []schema.Message{
		{Role: "system", Content: "You summarize agent execution processes."},
		{Role: "user", Content: prompt},
	}

	resp, err := a.llm.Generate(ctx, req, nil)
	if err != nil {
		return sb.String(), err
	}

	return resp.Content, nil
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
		DIM, a.log.GetLogFilePath(), RESET)

	step := 0

	for step < a.maxSteps {

		// 触发摘要
		if err := a.summarizeMessages(ctx); err != nil {
			slog.Warn("Summarization failed", slog.String("err", err.Error()))
		}

		// 打印 Step 框
		stepText := fmt.Sprintf("%s%s💭 Step %d/%d%s",
			BOLD, BRIGHT_CYAN, step+1, a.maxSteps, RESET)
		width := terminal.CalculateDisplayWidth(stepText)
		box := 58
		padding := box - 1 - width

		fmt.Printf("\n%s╭%s╮%s\n", DIM, strings.Repeat("─", box), RESET)
		fmt.Printf("%s│%s %s%s%s│%s\n",
			DIM, RESET,
			stepText,
			strings.Repeat(" ", padding),
			DIM, RESET)
		fmt.Printf("%s╰%s╯%s\n",
			DIM, strings.Repeat("─", box), RESET)

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
			fmt.Printf("\n%s❌ LLM Error: %s%s\n", BRIGHT_RED, err.Error(), RESET)
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
			fmt.Printf("\n%s🧠 Thinking:%s\n", BOLD+MAGENTA, RESET)
			fmt.Printf("%s%s%s\n", DIM, resp.Thinking, RESET)
		}

		// 打印模型输出
		if resp.Content != "" {
			fmt.Printf("\n%s🤖 Assistant:%s\n", BOLD+BRIGHT_BLUE, RESET)
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
				BRIGHT_YELLOW, RESET, BOLD, CYAN, fname)

			// 打印参数
			fmt.Printf("%s   Arguments:%s\n", DIM, RESET)
			b, _ := json.MarshalIndent(args, "", "  ")
			for _, line := range strings.Split(string(b), "\n") {
				fmt.Printf("   %s%s%s\n", DIM, line, RESET)
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
					text = text[:300] + DIM + "..." + RESET
				}
				fmt.Printf("%s✓ Result:%s %s\n", BRIGHT_GREEN, RESET, text)
			} else {
				fmt.Printf("%s✗ Error:%s %s%s%s\n",
					BRIGHT_RED, RESET, RED, result.Error, RESET)
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
	fmt.Printf("\n%s⚠️ %s%s\n", BRIGHT_YELLOW, msg, RESET)
	return msg, nil
}

func (a *Agent) History() []schema.Message {
	out := make([]schema.Message, len(a.messages))
	copy(out, a.messages)
	return out
}
