// cmd/cli/main.go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	prompt "github.com/c-bata/go-prompt"

	"gopilot-cli/internal/agent"
	"gopilot-cli/internal/config"
	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/retry"
	"gopilot-cli/internal/tools"
	tw "gopilot-cli/internal/utils/terminal"
)

//
// ANSI Colors（和之前版本保持一致）
//

const (
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorCyan   = "\033[36m"

	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"
)

//
// CLI 参数解析
//

type CLIArgs struct {
	Workspace string
}

func parseArgs() *CLIArgs {
	var workspace string

	flag.StringVar(&workspace, "workspace", "", "Workspace directory (default: current directory)")
	flag.StringVar(&workspace, "w", workspace, "Workspace directory (shorthand)")

	flag.Parse()

	return &CLIArgs{
		Workspace: workspace,
	}
}

//
// Banner & 帮助 & Session Info & Stats
//

func printBanner() {
	const boxWidth = 58
	text := fmt.Sprintf("%s🤖 Gopilot - Multi-turn Interactive Session%s", ColorBold, ColorReset)
	width := tw.CalculateDisplayWidth(text)

	totalPadding := boxWidth - width
	if totalPadding < 0 {
		totalPadding = 0
	}
	left := totalPadding / 2
	right := totalPadding - left

	fmt.Println()
	fmt.Printf("%s%s╔%s╗%s\n", ColorBold, ColorBrightCyan, strings.Repeat("═", boxWidth), ColorReset)
	fmt.Printf("%s%s║%s%s%s%s║%s\n",
		ColorBold, ColorBrightCyan,
		strings.Repeat(" ", left),
		text,
		strings.Repeat(" ", right),
		ColorBrightCyan,
		ColorReset,
	)
	fmt.Printf("%s%s╚%s╝%s\n", ColorBold, ColorBrightCyan, strings.Repeat("═", boxWidth), ColorReset)
	fmt.Println()
}

func printHelp() {
	fmt.Printf(`
%s%sAvailable Commands:%s
  %s/help%s      - Show this help message
  %s/clear%s     - Clear session history (keep system prompt)
  %s/history%s   - Show current session message count
  %s/stats%s     - Show session statistics
  %s/exit%s      - Exit program (also: exit, quit, q)

%s%sNotes (Go version):%s
  - 直接输入任务回车即可
  - 使用 Tab 可以补全 /help /exit 等命令
`,
		ColorBold, ColorBrightYellow, ColorReset,
		ColorBrightGreen, ColorReset,
		ColorBrightGreen, ColorReset,
		ColorBrightGreen, ColorReset,
		ColorBrightGreen, ColorReset,
		ColorBrightGreen, ColorReset,

		ColorBold, ColorBrightYellow, ColorReset,
	)
}

func printSessionInfo(ag *agent.Agent, workspaceDir string, model string, toolCount int) {
	const boxWidth = 58

	printInfoLine := func(text string) {
		textWidth := tw.CalculateDisplayWidth(text)
		padding := boxWidth - 1 - textWidth // -1 for leading space
		if padding < 0 {
			padding = 0
		}
		fmt.Printf("%s│%s %s%s%s│%s\n",
			ColorDim, ColorReset,
			text,
			strings.Repeat(" ", padding),
			ColorDim, ColorReset)
	}

	fmt.Printf("%s┌%s┐%s\n", ColorDim, strings.Repeat("─", boxWidth), ColorReset)

	header := fmt.Sprintf("%sSession Info%s", ColorBrightCyan, ColorReset)
	headerWidth := tw.CalculateDisplayWidth(header)
	totalPad := boxWidth - 1 - headerWidth
	if totalPad < 0 {
		totalPad = 0
	}
	left := totalPad / 2
	right := totalPad - left

	fmt.Printf("%s│%s %s%s%s%s│%s\n",
		ColorDim, ColorReset,
		strings.Repeat(" ", left),
		header,
		strings.Repeat(" ", right),
		ColorDim, ColorReset)

	fmt.Printf("%s├%s┤%s\n", ColorDim, strings.Repeat("─", boxWidth), ColorReset)

	history := ag.History()
	printInfoLine(fmt.Sprintf("Model: %s", model))
	printInfoLine(fmt.Sprintf("Workspace: %s", workspaceDir))
	printInfoLine(fmt.Sprintf("Message History: %d messages", len(history)))
	printInfoLine(fmt.Sprintf("Available Tools: %d tools", toolCount))

	fmt.Printf("%s└%s┘%s\n", ColorDim, strings.Repeat("─", boxWidth), ColorReset)
	fmt.Println()
	fmt.Printf("%sType %s/help%s for help, %s/exit%s to quit%s\n",
		ColorDim, ColorBrightGreen, ColorDim, ColorBrightGreen, ColorDim, ColorReset)
	fmt.Println()
}

func printStats(ag *agent.Agent, start time.Time, totalTools int) {
	dur := time.Since(start)
	totalSec := int(dur.Seconds())
	hours := totalSec / 3600
	minutes := (totalSec % 3600) / 60
	seconds := totalSec % 60

	history := ag.History()
	var userCount, assistantCount, toolMsgCount int
	for _, m := range history {
		switch m.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
		case "tool":
			toolMsgCount++
		}
	}

	fmt.Printf("\n%s%sSession Statistics:%s\n", ColorBold, ColorBrightCyan, ColorReset)
	fmt.Printf("%s%s%s\n", ColorDim, strings.Repeat("─", 40), ColorReset)
	fmt.Printf("  Session Duration: %02d:%02d:%02d\n", hours, minutes, seconds)
	fmt.Printf("  Total Messages: %d\n", len(history))
	fmt.Printf("    - User Messages: %s%d%s\n", ColorBrightGreen, userCount, ColorReset)
	fmt.Printf("    - Assistant Replies: %s%d%s\n", ColorBrightBlue, assistantCount, ColorReset)
	fmt.Printf("    - Tool Calls: %s%d%s\n", ColorBrightYellow, toolMsgCount, ColorReset)
	fmt.Printf("  Available Tools: %d\n", totalTools)
	fmt.Printf("%s%s%s\n\n", ColorDim, strings.Repeat("─", 40), ColorReset)
}

//
// System Prompt
//

func loadSystemPrompt(path string) string {
	if path == "" {
		return defaultSystemPrompt()
	}
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return defaultSystemPrompt()
	}
	return string(data)
}

func defaultSystemPrompt() string {
	return `You are a coding agent running in a CLI environment.

- You can edit files in the current workspace using the available tools.
- You can run shell commands for building, testing and inspecting the project.
- Always be explicit about what files you read or modify.
- Prefer small, incremental changes and keep outputs concise.`
}

//
// runAgent：对应 Python 的 run_agent + 交互 loop，用 go-prompt 实现
//

func runAgent(workspaceDir string) error {
	sessionStart := time.Now()

	// 1. 加载配置
	cfg, err := config.LoadFromFile("configs/config.yaml")
	if err != nil {
		fmt.Printf("%s❌ Failed to load config: %v%s\n", ColorRed, err, ColorReset)
		return err
	}

	// 2. 初始化重试配置 + LLM client
	rc := &retry.Config{
		Enabled:         cfg.LLM.Retry.Enabled,
		MaxRetries:      cfg.LLM.Retry.MaxRetries,
		InitialDelay:    time.Duration(cfg.LLM.Retry.InitialDelay * float64(time.Second)),
		MaxDelay:        time.Duration(cfg.LLM.Retry.MaxDelay * float64(time.Second)),
		ExponentialBase: cfg.LLM.Retry.ExponentialBase,
	}

	onRetry := func(err error, attempt int) {
		fmt.Printf("\n%s⚠️  LLM call failed (attempt %d): %s%s\n",
			ColorBrightYellow, attempt, err.Error(), ColorReset)
		delay := rc.CalculateDelay(attempt - 1)
		fmt.Printf("%s   Retrying in %s (attempt %d)...%s\n",
			ColorDim, delay.String(), attempt+1, ColorReset)
	}

	apiKey := cfg.LLM.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		fmt.Printf("%s❌ No API key provided (config.llm.api_key or OPENAI_API_KEY)%s\n", ColorRed, ColorReset)
		return fmt.Errorf("no api key")
	}

	llmClient := llm.NewClient(
		apiKey,
		cfg.LLM.APIBase,
		cfg.LLM.Model,
		llm.WithRetryConfig(rc),
		llm.WithRetryCallback(onRetry),
	)

	if cfg.LLM.Retry.Enabled {
		fmt.Printf("%s✅ LLM retry enabled (max %d retries)%s\n",
			ColorGreen, cfg.LLM.Retry.MaxRetries, ColorReset)
	}

	// 3. 初始化工具
	absWs, err := filepath.Abs(workspaceDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absWs, 0o755); err != nil {
		return err
	}

	var toolList []tools.Tool
	toolList = append(toolList,
		tools.NewBashTool(),
		tools.NewBashOutputTool(),
		tools.NewBashKillTool(),
	)
	fmt.Printf("%s✅ Loaded Bash tools%s\n", ColorGreen, ColorReset)

	toolList = append(toolList,
		tools.NewReadTool(absWs),
		tools.NewWriteTool(absWs),
		tools.NewEditTool(absWs),
	)
	fmt.Printf("%s✅ Loaded file tools (workspace: %s)%s\n", ColorGreen, absWs, ColorReset)

	// 4. System Prompt
	systemPrompt := loadSystemPrompt(cfg.Agent.SystemPromptPath)
	fmt.Printf("%s✅ System prompt loaded%s\n", ColorGreen, ColorReset)

	// 5. 创建 Agent
	ag, err := agent.NewAgent(
		llmClient,
		systemPrompt,
		toolList,
		cfg.Agent.MaxSteps,
		absWs,
		cfg.Agent.TokenLimit,
	)
	if err != nil {
		return err
	}

	// 6. 打印欢迎信息
	printBanner()
	printSessionInfo(ag, absWs, cfg.LLM.Model, len(toolList))

	// 7. go-prompt：补全器
	completer := func(d prompt.Document) []prompt.Suggest {
		text := strings.TrimSpace(d.TextBeforeCursor())
		// 仅在开头位置补全命令
		if len(text) == 0 || strings.HasPrefix(text, "/") {
			suggestions := []prompt.Suggest{
				{Text: "/help", Description: "Show help message"},
				{Text: "/clear", Description: "Clear session history"},
				{Text: "/history", Description: "Show message count"},
				{Text: "/stats", Description: "Show session statistics"},
				{Text: "/exit", Description: "Exit program"},
			}
			return prompt.FilterHasPrefix(suggestions, text, true)
		}
		return []prompt.Suggest{}
	}

	// 8. go-prompt：执行器
	executor := func(in string) {
		input := strings.TrimSpace(in)
		if input == "" {
			return
		}

		// 命令（以 / 开头）
		if strings.HasPrefix(input, "/") {
			cmd := strings.ToLower(input)

			switch cmd {
			case "/exit", "/quit", "/q":
				fmt.Printf("\n%s👋 Goodbye! Thanks for using Mini Agent%s\n\n", ColorBrightYellow, ColorReset)
				printStats(ag, sessionStart, len(toolList))
				os.Exit(0)
			case "/help":
				printHelp()
				return
			case "/clear":
				oldCount := len(ag.History())
				fmt.Printf("%s✅ Cleared %d messages, starting new session%s\n\n",
					ColorGreen, oldCount-1, ColorReset)

				var err error
				ag, err = agent.NewAgent(
					llmClient,
					systemPrompt,
					toolList,
					cfg.Agent.MaxSteps,
					absWs,
					cfg.Agent.TokenLimit,
				)
				if err != nil {
					fmt.Printf("%s❌ Failed to reset agent: %v%s\n", ColorRed, err, ColorReset)
				}
				return
			case "/history":
				fmt.Printf("\n%sCurrent session message count: %d%s\n\n",
					ColorBrightCyan, len(ag.History()), ColorReset)
				return
			case "/stats":
				printStats(ag, sessionStart, len(toolList))
				return
			default:
				fmt.Printf("%s❌ Unknown command: %s%s\n", ColorRed, input, ColorReset)
				fmt.Printf("%sType /help to see available commands%s\n\n", ColorDim, ColorReset)
				return
			}
		}

		// 非 / 命令：允许 exit/quit/q
		lower := strings.ToLower(input)
		if lower == "exit" || lower == "quit" || lower == "q" {
			fmt.Printf("\n%s👋 Goodbye! Thanks for using Mini Agent%s\n\n", ColorBrightYellow, ColorReset)
			printStats(ag, sessionStart, len(toolList))
			os.Exit(0)
		}

		// 普通对话：丢给 Agent
		fmt.Printf("\n%sAgent%s %s›%s %sThinking...%s\n\n",
			ColorBrightBlue, ColorReset, ColorDim, ColorReset, ColorDim, ColorReset)

		ag.AddUserMessage(input)

		ctx := context.Background()
		_, err := ag.Run(ctx)
		if err != nil {
			fmt.Printf("\n%s❌ Error: %v%s\n", ColorRed, err, ColorReset)
		}

		fmt.Printf("\n%s%s%s\n\n", ColorDim, strings.Repeat("─", 60), ColorReset)
	}

	// 9. 启动 go-prompt
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix("You › "),
		prompt.OptionTitle("gopilot-cli"),
		prompt.OptionInputTextColor(prompt.Yellow),
	)
	p.Run()

	return nil
}

//
// main：CLI 入口
//

func main() {
	args := parseArgs()

	var workspaceDir string
	if args.Workspace != "" {
		workspaceDir = args.Workspace
	} else {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("%s❌ Failed to get current directory: %v%s\n", ColorRed, err, ColorReset)
			os.Exit(1)
		}
		workspaceDir = wd
	}

	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		fmt.Printf("%s❌ Failed to create workspace dir: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}

	if err := runAgent(workspaceDir); err != nil {
		os.Exit(1)
	}
}
