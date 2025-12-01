package tools

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

//
// ============================================================
// 工具内部辅助：格式化 content
// ============================================================
//

// formatBashContent 生成统一格式的 Content：
// stdout
// [stderr]:
// ...
// [bash_id]:
// ...
// [exit_code]:
// ...
func formatBashContent(stdout, stderr string, exitCode int, bashID string) string {
	var b strings.Builder

	if stdout != "" {
		b.WriteString(stdout)
	}

	if stderr != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[stderr]:\n")
		b.WriteString(stderr)
	}

	if bashID != "" {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[bash_id]:\n")
		b.WriteString(bashID)
	}

	if b.Len() > 0 {
		b.WriteString("\n")
	}
	b.WriteString("[exit_code]:\n")
	b.WriteString(fmt.Sprintf("%d", exitCode))

	if b.Len() == 0 {
		return "(no output)"
	}
	return b.String()
}

//
// ============================================================
// BackgroundShell —— 后台进程状态容器
// ============================================================
//

type BackgroundShell struct {
	BashID       string
	Command      string
	Cmd          *exec.Cmd
	StdoutReader *bufio.Reader // stdout 和 stderr 已合并

	OutputLines   []string
	LastReadIndex int

	Status   string // running / completed / failed / terminated / error
	ExitCode *int
	Start    time.Time

	mu sync.Mutex
}

func (s *BackgroundShell) AddOutput(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.OutputLines = append(s.OutputLines, line)
}

func (s *BackgroundShell) GetNewOutput() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	newLines := s.OutputLines[s.LastReadIndex:]
	s.LastReadIndex = len(s.OutputLines)
	return newLines
}

func (s *BackgroundShell) UpdateStatus(alive bool, code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if alive {
		s.Status = "running"
		return
	}
	if code == 0 {
		s.Status = "completed"
	} else {
		s.Status = "failed"
	}
	s.ExitCode = &code
}

func (s *BackgroundShell) SetErrorStatus(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = "error"
	s.OutputLines = append(s.OutputLines, "Monitor error: "+msg)
}

func (s *BackgroundShell) Terminate() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Cmd != nil && s.Cmd.Process != nil {
		_ = s.Cmd.Process.Kill()
	}
	s.Status = "terminated"
	code := -1
	s.ExitCode = &code
}

//
// ============================================================
// BackgroundShellManager
// ============================================================
//

type BackgroundShellManager struct {
	mu     sync.Mutex
	shells map[string]*BackgroundShell
}

var globalShellManager = &BackgroundShellManager{
	shells: make(map[string]*BackgroundShell),
}

func (m *BackgroundShellManager) Add(shell *BackgroundShell) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shells[shell.BashID] = shell
}

func (m *BackgroundShellManager) Get(id string) *BackgroundShell {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.shells[id]
}

func (m *BackgroundShellManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.shells, id)
}

func (m *BackgroundShellManager) ListIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	ids := make([]string, 0, len(m.shells))
	for id := range m.shells {
		ids = append(ids, id)
	}
	return ids
}

//
// ============================================================
// 监控 goroutine —— 读取后台输出 + 更新状态
// ============================================================
//

func monitorShellOutput(shell *BackgroundShell) {
	reader := shell.StdoutReader
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			shell.AddOutput(strings.TrimRight(line, "\n"))
		}
		if err != nil {
			// 流关闭 / 读取错误 -> 进程可能已退出
			if err == io.EOF {
				// 正常结束，等待进程真正退出，拿到退出码
				if shell.Cmd != nil {
					_ = shell.Cmd.Wait()
					if shell.Cmd.ProcessState != nil {
						shell.UpdateStatus(false, shell.Cmd.ProcessState.ExitCode())
					} else {
						shell.UpdateStatus(false, -1)
					}
				} else {
					shell.UpdateStatus(false, -1)
				}
			} else {
				// 其他读取错误
				shell.SetErrorStatus(err.Error())
			}
			return
		}
	}
}

//
// ============================================================
// 参数解析小工具
// ============================================================
//

func getIntArg(args map[string]any, key string, def int) int {
	v, ok := args[key]
	if !ok {
		return def
	}
	switch vv := v.(type) {
	case int:
		return vv
	case int32:
		return int(vv)
	case int64:
		return int(vv)
	case float64:
		return int(vv)
	default:
		return def
	}
}

func getBoolArg(args map[string]any, key string, def bool) bool {
	v, ok := args[key]
	if !ok {
		return def
	}
	if b, ok := v.(bool); ok {
		return b
	}
	return def
}

// generateBashID 生成一个 8 字符的随机 ID（对应 Python 的 str(uuid.uuid4())[:8]）
func generateBashID() string {
	return uuid.New().String()[:8]
}

//
// ============================================================
// BashTool
// ============================================================
//

type BashTool struct {
	isWindows bool
}

func NewBashTool() *BashTool {
	return &BashTool{
		isWindows: runtime.GOOS == "windows",
	}
}

func (t *BashTool) Name() string {
	return "bash"
}

func (t *BashTool) Description() string {
	if t.isWindows {
		return `Execute PowerShell commands in foreground or background.

For terminal operations like git, npm, docker, etc. DO NOT use for file operations - use specialized tools.

Parameters:
  - command (required): PowerShell command to execute
  - timeout (optional): Timeout in seconds (default: 120, max: 600) for foreground commands
  - run_in_background (optional): Set true for long-running commands (servers, etc.)

Tips:
  - Quote file paths with spaces: cd "My Documents"
  - Chain dependent commands with semicolon: git add . ; git commit -m "msg"
  - Use absolute paths instead of cd when possible
  - For background commands, monitor with bash_output and terminate with bash_kill`
	}
	return `Execute bash commands in foreground or background.

For terminal operations like git, npm, docker, etc. DO NOT use for file operations - use specialized tools.

Parameters:
  - command (required): Bash command to execute
  - timeout (optional): Timeout in seconds (default: 120, max: 600) for foreground commands
  - run_in_background (optional): Set true for long-running commands (servers, etc.)

Tips:
  - Quote file paths with spaces: cd "My Documents"
  - Chain dependent commands with &&: git add . && git commit -m "msg"
  - Use absolute paths instead of cd when possible
  - For background commands, monitor with bash_output and terminate with bash_kill`
}

func (t *BashTool) Parameters() map[string]any {
	shellName := "bash"
	if t.isWindows {
		shellName = "PowerShell"
	}
	cmdDesc := fmt.Sprintf("The %s command to execute. Quote file paths with spaces using double quotes.", shellName)

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"command": map[string]any{
				"type":        "string",
				"description": cmdDesc,
			},
			"timeout": map[string]any{
				"type":        "integer",
				"description": "Optional: Timeout in seconds (default: 120, max: 600). Only applies to foreground commands.",
			},
			"run_in_background": map[string]any{
				"type":        "boolean",
				"description": "Optional: Set to true to run the command in the background. Use this for long-running commands like servers. You can monitor output using bash_output tool.",
			},
		},
		"required": []string{"command"},
	}
}

// Execute 对应 Python BashTool.execute
func (t *BashTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	command, _ := args["command"].(string)
	if strings.TrimSpace(command) == "" {
		return &ToolResult{
			Success: false,
			Error:   "command is required",
		}, nil
	}

	timeout := getIntArg(args, "timeout", 120)
	if timeout > 600 {
		timeout = 600
	} else if timeout < 1 {
		timeout = 120
	}
	runBG := getBoolArg(args, "run_in_background", false)

	var cmd *exec.Cmd
	if t.isWindows {
		cmd = exec.Command("powershell.exe", "-NoProfile", "-Command", command)
	} else {
		cmd = exec.Command("bash", "-c", command)
	}

	// -----------------------------
	// 后台执行
	// -----------------------------
	if runBG {
		id := generateBashID()

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("failed to get stdout pipe: %v", err),
			}, nil
		}
		cmd.Stderr = cmd.Stdout // stderr 合并到 stdout

		if err := cmd.Start(); err != nil {
			return &ToolResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}

		shell := &BackgroundShell{
			BashID:       id,
			Command:      command,
			Cmd:          cmd,
			StdoutReader: bufio.NewReader(stdoutPipe),
			Start:        time.Now(),
			Status:       "running",
		}
		globalShellManager.Add(shell)

		go monitorShellOutput(shell)

		message := fmt.Sprintf("Command started in background. Use bash_output to monitor (bash_id='%s').", id)
		formattedContent := fmt.Sprintf("%s\n\nCommand: %s\nBash ID: %s", message, command, id)

		return &ToolResult{
			Success:  true,
			Content:  formattedContent,
			Stdout:   fmt.Sprintf("Background command started with ID: %s", id),
			Stderr:   "",
			ExitCode: 0,
			BashID:   id,
		}, nil
	}

	// -----------------------------
	// 前台执行
	// -----------------------------
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	var err error
	select {
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		err = fmt.Errorf("command cancelled: %w", ctx.Err())
	case e := <-done:
		err = e
	case <-time.After(time.Duration(timeout) * time.Second):
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		err = fmt.Errorf("command timed out after %d seconds", timeout)
	}

	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	} else if err != nil {
		exitCode = -1
	}

	content := formatBashContent(stdout, stderr, exitCode, "")

	if err != nil {
		return &ToolResult{
			Success:  false,
			Content:  content,
			Error:    err.Error(),
			Stdout:   stdout,
			Stderr:   stderr,
			ExitCode: exitCode,
		}, nil
	}

	return &ToolResult{
		Success:  true,
		Content:  content,
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
	}, nil
}

//
// ============================================================
// BashOutputTool
// ============================================================
//

type BashOutputTool struct{}

func NewBashOutputTool() *BashOutputTool {
	return &BashOutputTool{}
}

func (t *BashOutputTool) Name() string {
	return "bash_output"
}

func (t *BashOutputTool) Description() string {
	return `Retrieves output from a running or completed background bash shell.

- Takes a bash_id parameter identifying the shell
- Always returns only new output since the last check
- Returns stdout and stderr output (combined) along with exit_code
- Supports optional regex filtering to show only lines matching a pattern
- Use this tool to monitor long-running commands started with bash(run_in_background=true)`
}

func (t *BashOutputTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bash_id": map[string]any{
				"type":        "string",
				"description": "The ID of the background shell to retrieve output from.",
			},
			"filter_str": map[string]any{
				"type":        "string",
				"description": "Optional regular expression to filter the output lines. Non-matching new lines will be discarded.",
			},
		},
		"required": []string{"bash_id"},
	}
}

func (t *BashOutputTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	id, _ := args["bash_id"].(string)
	filterStr, _ := args["filter_str"].(string)

	shell := globalShellManager.Get(id)
	if shell == nil {
		available := globalShellManager.ListIDs()
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Shell not found: %s. Available: %v", id, available),
		}, nil
	}

	lines := shell.GetNewOutput()

	if filterStr != "" {
		if re, err := regexp.Compile(filterStr); err == nil {
			filtered := make([]string, 0, len(lines))
			for _, line := range lines {
				if re.MatchString(line) {
					filtered = append(filtered, line)
				}
			}
			lines = filtered
		}
		// 如果正则错误：按 Python 行为，忽略错误，返回所有新行
	}

	stdout := strings.Join(lines, "\n")

	exitCode := 0
	if shell.ExitCode != nil {
		exitCode = *shell.ExitCode
	}

	content := formatBashContent(stdout, "", exitCode, id)

	return &ToolResult{
		Success:  true,
		Content:  content,
		Stdout:   stdout,
		Stderr:   "",
		ExitCode: exitCode,
		BashID:   id,
	}, nil
}

//
// ============================================================
// BashKillTool
// ============================================================
//

type BashKillTool struct{}

func NewBashKillTool() *BashKillTool {
	return &BashKillTool{}
}

func (t *BashKillTool) Name() string {
	return "bash_kill"
}

func (t *BashKillTool) Description() string {
	return `Kills a running background bash shell by its ID.

- Takes a bash_id parameter identifying the shell to kill
- Attempts termination and returns remaining output
- Cleans up all resources associated with the shell
- Use this tool when you need to terminate long-running commands started with bash(run_in_background=true)`
}

func (t *BashKillTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"bash_id": map[string]any{
				"type":        "string",
				"description": "The ID of the background shell to terminate.",
			},
		},
		"required": []string{"bash_id"},
	}
}

func (t *BashKillTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	id, _ := args["bash_id"].(string)

	shell := globalShellManager.Get(id)
	if shell == nil {
		available := globalShellManager.ListIDs()
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("Shell not found: %s. Available: %v", id, available),
		}, nil
	}

	// 先取剩余输出
	lines := shell.GetNewOutput()
	stdout := strings.Join(lines, "\n")

	// 终止并移除
	shell.Terminate()
	globalShellManager.Remove(id)

	exitCode := 0
	if shell.ExitCode != nil {
		exitCode = *shell.ExitCode
	} else {
		exitCode = -1
	}

	content := formatBashContent(stdout, "", exitCode, id)

	return &ToolResult{
		Success:  true,
		Content:  content,
		Stdout:   stdout,
		Stderr:   "",
		ExitCode: exitCode,
		BashID:   id,
	}, nil
}
