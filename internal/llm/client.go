package llm

import (
	"context"
	"encoding/json"
	"fmt"

	"log/slog"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"

	"gopilot-cli/internal/retry"
	"gopilot-cli/internal/schema"
	"gopilot-cli/internal/tools"
)

// Client LLM 客户端
type Client struct {
	client      openai.Client
	model       string
	retryConfig *retry.Config
	onRetry     retry.OnRetryFunc
}

// ClientOption 客户端选项
type ClientOption func(*Client)

// WithRetryConfig 设置重试配置
func WithRetryConfig(cfg *retry.Config) ClientOption {
	return func(c *Client) {
		c.retryConfig = cfg
	}
}

// WithRetryCallback 设置重试回调
func WithRetryCallback(fn retry.OnRetryFunc) ClientOption {
	return func(c *Client) {
		c.onRetry = fn
	}
}

// NewClient 创建 LLM 客户端
func NewClient(apiKey, baseURL, model string, opts ...ClientOption) *Client {
	clientOpts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if baseURL != "" {
		clientOpts = append(clientOpts, option.WithBaseURL(baseURL))
	}

	c := &Client{
		client:      openai.NewClient(clientOpts...),
		model:       model,
		retryConfig: retry.DefaultConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	slog.Info("Initialized LLM client",
		slog.String("model", model),
		slog.String("baseURL", baseURL),
	)

	return c
}

// Generate 生成 LLM 响应
func (c *Client) Generate(ctx context.Context, messages []schema.Message, toolRegistry *tools.ToolRegistry) (*schema.LLMResponse, error) {
	return retry.Do(ctx, c.retryConfig, func() (*schema.LLMResponse, error) {
		return c.doGenerate(ctx, messages, toolRegistry)
	}, c.onRetry)
}

func (c *Client) doGenerate(ctx context.Context, messages []schema.Message, toolRegistry *tools.ToolRegistry) (*schema.LLMResponse, error) {
	chatMessages := c.convertMessages(messages)

	params := openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: chatMessages,
	}

	if toolRegistry != nil && len(toolRegistry.List()) > 0 {
		params.Tools = c.convertTools(toolRegistry)
	}

	completion, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("chat completion failed: %w", err)
	}

	return c.parseResponse(completion), nil
}

// convertMessages 转换消息格式
func (c *Client) convertMessages(messages []schema.Message) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, 0, len(messages))

	for _, msg := range messages {
		switch msg.Role {
		case "system":
			// 使用辅助函数 SystemMessage
			result = append(result, openai.SystemMessage(msg.Content))

		case "user":
			// 使用辅助函数 UserMessage
			result = append(result, openai.UserMessage(msg.Content))

		case "assistant":
			if len(msg.ToolCalls) > 0 {
				// 构建带工具调用的 assistant 消息
				// 需要使用 ChatCompletionMessageToolCallUnionParam
				toolCalls := make([]openai.ChatCompletionMessageToolCallUnionParam, 0, len(msg.ToolCalls))
				for _, tc := range msg.ToolCalls {
					argsJSON, _ := json.Marshal(tc.Function.Arguments)
					// 使用 OfFunction 字段包装
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCallUnionParam{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: tc.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      tc.Function.Name,
								Arguments: string(argsJSON),
							},
						},
					})
				}

				// 构建 assistant 消息参数
				assistantParam := openai.ChatCompletionAssistantMessageParam{
					ToolCalls: toolCalls,
				}
				// 设置 content（如果有）
				if msg.Content != "" {
					assistantParam.Content.OfString = param.NewOpt(msg.Content)
				}

				// 使用 Union 包装
				result = append(result, openai.ChatCompletionMessageParamUnion{
					OfAssistant: &assistantParam,
				})
			} else {
				// 没有工具调用，使用辅助函数
				result = append(result, openai.AssistantMessage(msg.Content))
			}

		case "tool":
			// 使用辅助函数 ToolMessage
			result = append(result, openai.ToolMessage(msg.Content, msg.ToolCallID))
		}
	}

	return result
}

// convertTools 转换工具格式
func (c *Client) convertTools(registry *tools.ToolRegistry) []openai.ChatCompletionToolUnionParam {
	toolList := registry.List()
	result := make([]openai.ChatCompletionToolUnionParam, 0, len(toolList))

	for _, tool := range toolList {
		result = append(result, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
			Parameters:  openai.FunctionParameters(tool.Parameters()),
		}))
	}

	return result
}

// parseResponse 解析 API 响应
func (c *Client) parseResponse(completion *openai.ChatCompletion) *schema.LLMResponse {
	if len(completion.Choices) == 0 {
		return &schema.LLMResponse{FinishReason: "unknown"}
	}

	message := completion.Choices[0].Message
	response := &schema.LLMResponse{
		Content:      message.Content,
		FinishReason: string(completion.Choices[0].FinishReason),
	}

	// 提取 thinking 内容
	for k, v := range message.JSON.ExtraFields {
		switch k {
		case "reasoning_content",
			"thoughts",
			"internal_thoughts",
			"reasoning":
			response.Thinking = v.Raw()
		}
	}

	// 解析工具调用
	for _, tc := range message.ToolCalls {
		var args map[string]any
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		response.ToolCalls = append(response.ToolCalls, schema.ToolCall{
			ID:   tc.ID,
			Type: "function",
			Function: schema.FunctionCall{
				Name:      tc.Function.Name,
				Arguments: args,
			},
		})
	}

	return response
}
