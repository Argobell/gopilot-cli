package tools

import (
	"context"
)

// ToolResult 工具执行结果
type ToolResult struct {
	Success bool   `json:"success"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`

	// 下面这几个是给 Bash 这种需要结构化输出的工具用的
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	ExitCode int    `json:"exit_code,omitempty"`
	BashID   string `json:"bash_id,omitempty"`
}

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(ctx context.Context, args map[string]any) (*ToolResult, error)
}

// ToOpenAISchema 将 Tool 转换为 OpenAI 工具格式
func ToOpenAISchema(tool Tool) map[string]any {
	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        tool.Name(),
			"description": tool.Description(),
			"parameters":  tool.Parameters(),
		},
	}
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]Tool
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// ToOpenAISchemas 转换所有工具为 OpenAI 格式
func (r *ToolRegistry) ToOpenAISchemas() []map[string]any {
	schemas := make([]map[string]any, 0, len(r.tools))
	for _, tool := range r.tools {
		schemas = append(schemas, ToOpenAISchema(tool))
	}
	return schemas
}
