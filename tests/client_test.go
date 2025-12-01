package tests

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gopilot-cli/internal/config"
	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/schema"
	"gopilot-cli/internal/tools"
)

//
// ---------------------------------------------------------
// Helper: Load Config
// ---------------------------------------------------------
//

// 获取项目根目录（因为 Go test 工作目录在 tests/ 下）
func projectRoot() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..")
}

// 加载 configs/config.yaml，若失败直接报错
func loadTestConfig(t *testing.T) *config.Config {
	path := filepath.Join(projectRoot(), "configs/config.yaml")
	cfg, err := config.LoadFromFile(path)
	require.NoError(t, err, "failed to load config.yaml")
	return cfg
}

//
// ---------------------------------------------------------
// Logger Initialization
// ---------------------------------------------------------
//

// 初始化 slog 日志系统
func init() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)
}

//
// ---------------------------------------------------------
// Test 1: Simple Completion
// ---------------------------------------------------------
//

// 测试基础对话能力（简单补全）
func TestOpenAI_SimpleCompletion(t *testing.T) {
	slog.Info("=== TestOpenAI_SimpleCompletion ===")

	// 加载配置
	cfg := loadTestConfig(t)

	// 初始化 LLM 客户端
	client := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.APIBase, cfg.LLM.Model)

	// 构造简单消息
	msgs := []schema.Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Say 'Hello from OpenAI Go Client!' and nothing else."},
	}

	// 设置请求超时，避免测试阻塞
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 执行调用
	resp, err := client.Generate(ctx, msgs, nil)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// 调试输出
	slog.Info("模型返回",
		slog.String("content", resp.Content),
		slog.String("thinking", resp.Thinking),
		slog.String("finish_reason", resp.FinishReason),
	)

	// 内容检查
	require.Contains(t, resp.Content, "Hello")
}

//
// ---------------------------------------------------------
// Weather Tool (示例工具)
// ---------------------------------------------------------
//

// 示例工具：查询天气
type WeatherTool struct{}

func (WeatherTool) Name() string        { return "get_weather" }
func (WeatherTool) Description() string { return "Get weather of a location" }
func (WeatherTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]any{
				"type":        "string",
				"description": "City name",
			},
		},
		"required": []string{"location"},
	}
}
func (WeatherTool) Execute(ctx context.Context, args map[string]any) (*tools.ToolResult, error) {
	location := args["location"].(string)
	return &tools.ToolResult{
		Success: true,
		Content: fmt.Sprintf("Sunny in %s", location),
	}, nil
}

//
// ---------------------------------------------------------
// Test 2: Tool Calling
// ---------------------------------------------------------
//

// 测试模型是否能正确触发工具调用
func TestOpenAI_ToolCalling(t *testing.T) {
	slog.Info("=== TestOpenAI_ToolCalling ===")

	// 加载配置
	cfg := loadTestConfig(t)

	// 初始化 LLM 客户端
	client := llm.NewClient(cfg.LLM.APIKey, cfg.LLM.APIBase, cfg.LLM.Model)

	// 注册测试工具
	reg := tools.NewToolRegistry()
	reg.Register(WeatherTool{})

	// 构造对话（用户请求天气）
	msgs := []schema.Message{
		{Role: "user", Content: "What's the weather in New York?"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// 调用模型
	resp, err := client.Generate(ctx, msgs, reg)
	require.NoError(t, err)
	require.NotNil(t, resp)

	// 打印模型响应
	slog.Info("LLM Response",
		slog.String("content", resp.Content),
		slog.String("thinking", resp.Thinking),
		slog.Any("tool_calls", resp.ToolCalls),
	)

	// 必须触发工具调用
	require.NotEmpty(t, resp.ToolCalls, "should call get_weather tool")
	require.Equal(t, "get_weather", resp.ToolCalls[0].Function.Name)
}
