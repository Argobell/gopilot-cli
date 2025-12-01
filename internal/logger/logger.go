package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopilot-cli/internal/schema"
	"gopilot-cli/internal/tools"
)

//
// ---------------------------------------------------------
// Agent Logger
// ---------------------------------------------------------
//

// AgentLogger 用于记录一次 Agent 运行过程中的所有信息。
// 包括：LLM 请求内容、LLM 响应内容、工具调用结果等。
// 内部使用互斥锁（mutex）确保多协程访问时的并发安全。
type AgentLogger struct {
	logDir   string     // 日志目录 (~/.gopilot/log)
	logFile  *os.File   // 当前运行的日志文件句柄
	logIndex int        // 日志条目计数器
	mu       sync.Mutex // 互斥锁，保证所有操作并发安全
}

// NewAgentLogger 创建日志管理器实例，并初始化日志目录。
// 若目录或用户 Home 路径不存在，会自动尝试创建。
func NewAgentLogger() (*AgentLogger, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine user home directory: %w", err)
	}

	logDir := filepath.Join(home, ".gopilot", "log")

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create log directory: %w", err)
	}

	return &AgentLogger{
		logDir:   logDir,
		logIndex: 0,
	}, nil
}

//
// ---------------------------------------------------------
// Log File Control
// ---------------------------------------------------------
//

// StartNewRun 开启一次新的日志会话。
// 会创建一个带时间戳的日志文件，并写入基础头部信息。
func (l *AgentLogger) StartNewRun() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 关闭前一次运行的日志文件
	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}

	timestamp := time.Now().Format("20060102_150405")
	logFilename := fmt.Sprintf("agent_run_%s.log", timestamp)
	logPath := filepath.Join(l.logDir, logFilename)

	file, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	l.logFile = file
	l.logIndex = 0

	// 写入文件头
	header := fmt.Sprintf("%s\nAgent Run Log - %s\n%s\n",
		strings.Repeat("=", 80),
		time.Now().Format("2006-01-02 15:04:05"),
		strings.Repeat("=", 80),
	)

	if _, err := file.WriteString(header); err != nil {
		return fmt.Errorf("failed writing header: %w", err)
	}

	return nil
}

//
// ---------------------------------------------------------
// JSON Helper
// ---------------------------------------------------------
//

// safeJSON 对数据进行格式化 JSON 序列化。
// 若序列化失败，则返回带错误提示的 JSON 字符串。
func safeJSON(v any) []byte {
	j, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Appendf(nil, `{"error": "json marshal failed: %v"}`, err)
	}
	return j
}

//
// ---------------------------------------------------------
// Write to Log File
// ---------------------------------------------------------
//

// writeLog 向日志文件写入一条日志记录。
// 每条记录都会包含：日志类型、条目编号、时间戳、内容。
func (l *AgentLogger) writeLog(logType, content string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile == nil {
		return fmt.Errorf("log file not initialized (StartNewRun not called? )")
	}

	l.logIndex++

	entry := fmt.Sprintf(
		"\n%s\n[%d] %s\nTimestamp: %s\n%s\n%s\n",
		strings.Repeat("-", 80),
		l.logIndex,
		logType,
		time.Now().Format("2006-01-02 15:04:05. 000"),
		strings.Repeat("-", 80),
		content,
	)

	if _, err := l.logFile.WriteString(entry); err != nil {
		return fmt.Errorf("write log failed: %w", err)
	}

	return l.logFile.Sync() // 确保写入磁盘
}

//
// ---------------------------------------------------------
// Log LLM Request
// ---------------------------------------------------------
//

// LogRequest 记录一次完整的 LLM 请求。
// 包含 messages 内容、tool 列表（仅输出名称）、tool_calls 等。
func (l *AgentLogger) LogRequest(messages []schema.Message, toolList []tools.Tool) error {
	msgList := make([]map[string]any, 0, len(messages))

	for _, msg := range messages {
		m := map[string]any{
			"role":    msg.Role,
			"content": msg.Content,
		}
		if msg.Thinking != "" {
			m["thinking"] = msg.Thinking
		}
		if msg.ToolCallID != "" {
			m["tool_call_id"] = msg.ToolCallID
		}
		if msg.Name != "" {
			m["name"] = msg.Name
		}
		if len(msg.ToolCalls) > 0 {
			dumps := make([]map[string]any, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				dumps[i] = map[string]any{
					"id":   tc.ID,
					"type": tc.Type,
					"function": map[string]any{
						"name":      tc.Function.Name,
						"arguments": tc.Function.Arguments,
					},
				}
			}
			m["tool_calls"] = dumps
		}
		msgList = append(msgList, m)
	}

	req := map[string]any{
		"messages": msgList,
		"tools":    []string{},
	}

	if len(toolList) > 0 {
		names := make([]string, len(toolList))
		for i, t := range toolList {
			names[i] = t.Name()
		}
		req["tools"] = names
	}

	return l.writeLog("REQUEST", "LLM Request:\n\n"+string(safeJSON(req)))
}

//
// ---------------------------------------------------------
// Log LLM Response
// ---------------------------------------------------------
//

// LogResponse 记录 LLM 模型返回的数据。
// 包括 content、thinking（推理内容）、tool_calls、finish_reason 等。
func (l *AgentLogger) LogResponse(
	content string,
	thinking string,
	toolCalls []schema.ToolCall,
	finishReason string,
) error {
	resp := map[string]any{
		"content": content,
	}

	if thinking != "" {
		resp["thinking"] = thinking
	}
	if finishReason != "" {
		resp["finish_reason"] = finishReason
	}
	if len(toolCalls) > 0 {
		dumps := make([]map[string]any, len(toolCalls))
		for i, tc := range toolCalls {
			dumps[i] = map[string]any{
				"id":   tc.ID,
				"type": tc.Type,
				"function": map[string]any{
					"name":      tc.Function.Name,
					"arguments": tc.Function.Arguments,
				},
			}
		}
		resp["tool_calls"] = dumps
	}

	return l.writeLog("RESPONSE", "LLM Response:\n\n"+string(safeJSON(resp)))
}

//
// ---------------------------------------------------------
// Log Tool Execution Result
// ---------------------------------------------------------
//

// LogToolResult 记录工具执行的结果。
// 包括工具名称、参数、成功/失败、输出内容或错误信息。
func (l *AgentLogger) LogToolResult(
	toolName string,
	arguments map[string]any,
	success bool,
	resultContent string,
	resultError string,
) error {
	data := map[string]any{
		"tool_name": toolName,
		"arguments": arguments,
		"success":   success,
	}

	if success {
		data["result"] = resultContent
	} else {
		data["error"] = resultError
	}

	return l.writeLog("TOOL_RESULT", "Tool Execution:\n\n"+string(safeJSON(data)))
}

//
// ---------------------------------------------------------
// File Control
// ---------------------------------------------------------
//

// GetLogFilePath 返回当前日志文件的路径。
func (l *AgentLogger) GetLogFilePath() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile == nil {
		return ""
	}
	return l.logFile.Name()
}

// Close 关闭日志文件。
func (l *AgentLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.logFile != nil {
		err := l.logFile.Close()
		l.logFile = nil
		return err
	}
	return nil
}
