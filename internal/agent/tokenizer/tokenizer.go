package tokenizer

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"

	"gopilot-cli/internal/schema"
)

// EstimateTokens 估算消息历史的 token 数量。
// 优先使用 tiktoken-go 进行编码统计，若不可用则回退到字符长度估算。
// 对每条消息，统计 Content、Thinking、ToolCalls 的 token 数，并加上元数据开销。
func EstimateTokens(messages []schema.Message) int {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return EstimateTokensFallback(messages)
	}

	total := 0

	for _, m := range messages {
		// 统计消息正文的 token 数
		total += countTokens(enc, m.Content)
		// 统计思考内容的 token 数
		total += countTokens(enc, m.Thinking)
		// 若有工具调用，统计其 token 数
		if len(m.ToolCalls) > 0 {
			raw := fmt.Sprintf("%v", m.ToolCalls)
			total += len(enc.Encode(raw, nil, nil))
		}

		// 每条消息加约 4 个 token 的元数据开销
		total += 4
	}

	return total
}

// countTokens 用编码器统计文本的 token 数。
// 若文本为空则返回 0。
func countTokens(enc *tiktoken.Tiktoken, text string) int {
	if text == "" {
		return 0
	}
	return len(enc.Encode(text, nil, nil))
}

// EstimateTokensFallback 在无法使用编码器时，采用字符长度除以 2.5 的方式估算 token 数量。
func EstimateTokensFallback(messages []schema.Message) int {
    total := 0
    for _, m := range messages {
        total += len(m.Content)
        total += len(m.Thinking)
        if len(m.ToolCalls) > 0 {
            total += len(fmt.Sprintf("%v", m.ToolCalls))
        }
    }
	// 按 2.5 字符约等于 1 token 进行估算
    return int(float64(total) / 2.5)
}

