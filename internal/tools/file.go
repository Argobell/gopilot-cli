package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

//
// ---------------------------------------------------------
// Token-Based Text Truncation （基于 Token 的文本截断）
// ---------------------------------------------------------

// TruncateTextByTokens 按 token 限制截断文本（等价 Python truncate_text_by_tokens）
func TruncateTextByTokens(text string, maxTokens int) string {
	// 空字符串直接返回
	if len(text) == 0 {
		return text
	}

	// 获取 CL100K 编码器（与 Python 一致）
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return text // 编码器加载失败则不截断
	}

	// 计算 token 数
	tokenCount := len(enc.Encode(text, nil, nil))
	if tokenCount <= maxTokens {
		return text
	}

	// Token/字符比例，用于估算保留区间
	runes := []rune(text)
	charCount := len(runes)
	ratio := float64(tokenCount) / float64(charCount)

	// 前后各保留一半（含 5% 安全边界）
	charsPerHalf := int((float64(maxTokens) / 2) / ratio * 0.95)
	if charsPerHalf < 1 {
		charsPerHalf = 1
	}

	// -------------------------
	// 截断头部
	// -------------------------
	head := runes[:min(charsPerHalf, len(runes))]
	headStr := string(head)

	// 头部对齐换行符，尽量不截断句子结构
	if idx := strings.LastIndex(headStr, "\n"); idx > 0 {
		headStr = headStr[:idx]
	}

	// -------------------------
	// 截断尾部
	// -------------------------
	tail := runes[max(0, len(runes)-charsPerHalf):]
	tailStr := string(tail)

	// 尾部对齐换行符
	if idx := strings.Index(tailStr, "\n"); idx > 0 {
		tailStr = tailStr[idx+1:]
	}

	// -------------------------
	// 合并结果
	// -------------------------
	note := fmt.Sprintf(
		"\n\n... [Content truncated: %d tokens -> ~%d tokens limit] ...\n\n",
		tokenCount, maxTokens,
	)

	return headStr + note + tailStr
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//
// ---------------------------------------------------------
// ReadTool（读取文件，带行号 + offset/limit + token 截断）
// ---------------------------------------------------------

type ReadTool struct {
	workspace string
}

// NewReadTool 创建文件读取工具
func NewReadTool(workspace string) *ReadTool {
	return &ReadTool{workspace: workspace}
}

func (t *ReadTool) Name() string {
	return "read_file"
}

func (t *ReadTool) Description() string {
	return "Read file content with line numbers. Supports offset/limit and token truncation."
}

func (t *ReadTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type":        "string",
				"description": "File path (absolute or relative to workspace)",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Starting line number (1-indexed)",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Number of lines to read",
			},
		},
		"required": []string{"path"},
	}
}

func (t *ReadTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	// 解析参数
	path := args["path"].(string)

	var offset, limit *int
	if v, ok := args["offset"].(int); ok {
		offset = &v
	}
	if v, ok := args["limit"].(int); ok {
		limit = &v
	}

	// 解析文件路径（相对路径基于 workspace）
	file := filepath.Join(t.workspace, path)

	data, err := os.ReadFile(file)
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("File not found: %s", path)}, nil
	}

	lines := strings.Split(string(data), "\n")

	// -------------------------
	// 处理 offset / limit
	// -------------------------
	start := 0
	if offset != nil {
		start = *offset - 1
		if start < 0 {
			start = 0
		}
	}
	end := len(lines)
	if limit != nil {
		end = min(start+*limit, len(lines))
	}
	if start > len(lines) {
		start = len(lines)
	}

	selected := lines[start:end]

	// -------------------------
	// 添加行号（右对齐 6 格）
	// -------------------------
	formatted := make([]string, len(selected))
	for i, line := range selected {
		formatted[i] = fmt.Sprintf("%6d|%s", start+i+1, line)
	}

	content := strings.Join(formatted, "\n")

	// Token 截断（保持与 Python 32000 限制一致）
	content = TruncateTextByTokens(content, 32000)

	return &ToolResult{Success: true, Content: content}, nil
}

//
// ---------------------------------------------------------
// WriteTool（写入文件，覆盖模式）
// ---------------------------------------------------------

type WriteTool struct {
	workspace string
}

func NewWriteTool(workspace string) *WriteTool {
	return &WriteTool{workspace: workspace}
}

func (t *WriteTool) Name() string {
	return "write_file"
}

func (t *WriteTool) Description() string {
	return "Write full content to a file. Overwrites existing content."
}

func (t *WriteTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type": "string",
			},
			"content": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"path", "content"},
	}
}

func (t *WriteTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	path := args["path"].(string)
	content := args["content"].(string)

	file := filepath.Join(t.workspace, path)

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	// 写入内容
	err := os.WriteFile(file, []byte(content), 0644)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Content: fmt.Sprintf("Successfully wrote to %s", file)}, nil
}

//
// ---------------------------------------------------------
// EditTool（精确替换，仅替换一个 old_str）
// ---------------------------------------------------------

type EditTool struct {
	workspace string
}

func NewEditTool(workspace string) *EditTool {
	return &EditTool{workspace: workspace}
}

func (t *EditTool) Name() string {
	return "edit_file"
}

func (t *EditTool) Description() string {
	return "Perform exact string replacement in a file. old_str must appear exactly once."
}

func (t *EditTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"path": map[string]any{
				"type": "string",
			},
			"old_str": map[string]any{
				"type": "string",
			},
			"new_str": map[string]any{
				"type": "string",
			},
		},
		"required": []string{"path", "old_str", "new_str"},
	}
}

func (t *EditTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	path := args["path"].(string)
	oldStr := args["old_str"].(string)
	newStr := args["new_str"].(string)

	file := filepath.Join(t.workspace, path)

	data, err := os.ReadFile(file)
	if err != nil {
		return &ToolResult{Success: false, Error: fmt.Sprintf("File not found: %s", path)}, nil
	}

	content := string(data)

	if !strings.Contains(content, oldStr) {
		return &ToolResult{Success: false, Error: fmt.Sprintf("Text not found: %s", oldStr)}, nil
	}

	// 精确替换一个
	updated := strings.Replace(content, oldStr, newStr, 1)

	err = os.WriteFile(file, []byte(updated), 0644)
	if err != nil {
		return &ToolResult{Success: false, Error: err.Error()}, nil
	}

	return &ToolResult{Success: true, Content: fmt.Sprintf("Successfully edited %s", file)}, nil
}
