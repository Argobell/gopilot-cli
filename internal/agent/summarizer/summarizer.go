package summarizer

import (
	"context"
	"fmt"
	"strings"

	"log/slog"

	"gopilot-cli/internal/agent/colors"
	"gopilot-cli/internal/agent/tokenizer"
	"gopilot-cli/internal/llm"
	"gopilot-cli/internal/schema"
)

// Summarizer 用于对较长的 agent 消息历史进行摘要，
// 以保证消息内容不会超过设定的 token 限制。
type Summarizer struct {
	client     *llm.Client
	tokenLimit int
}

// 新建 Summarizer 实例
func NewSummarizer(client *llm.Client, tokenLimit int) *Summarizer {
	return &Summarizer{
		client:     client,
		tokenLimit: tokenLimit,
	}
}

// SummarizeMessages 当消息历史的 token 估算值超过限制时，
// 对消息历史进行摘要，返回可能已更新的消息切片。
func (s *Summarizer) SummarizeMessages(ctx context.Context, messages []schema.Message) ([]schema.Message, error) {
	tokens := tokenizer.EstimateTokens(messages)
	if tokens <= s.tokenLimit {
		return messages, nil
	}

	fmt.Printf("\n%s📊 Token estimate: %d/%d%s\n",
		colors.BRIGHT_YELLOW, tokens, s.tokenLimit, colors.RESET)
	fmt.Printf("%s🔄 Summarizing message history...%s\n", colors.BRIGHT_YELLOW, colors.RESET)

	// Collect all user message indices (skip system)
	userIdx := []int{}
	for i, m := range messages {
		if i > 0 && m.Role == "user" {
			userIdx = append(userIdx, i)
		}
	}
	if len(userIdx) == 0 {
		fmt.Printf("%s⚠️ No user messages to summarize%s\n", colors.BRIGHT_YELLOW, colors.RESET)
		return messages, nil
	}

	var newMsgs []schema.Message
	newMsgs = append(newMsgs, messages[0]) // system

	for i, ui := range userIdx {
		newMsgs = append(newMsgs, messages[ui])

		var end int
		if i < len(userIdx)-1 {
			end = userIdx[i+1]
		} else {
			end = len(messages)
		}

		execMsgs := messages[ui+1 : end]
		if len(execMsgs) == 0 {
			continue
		}

		// Create summary text
		summary, err := s.createSummary(ctx, execMsgs, i+1)
		if err != nil {
			slog.Warn("Summary failed", slog.String("err", err.Error()))
			continue
		}

		newMsgs = append(newMsgs, schema.Message{
			Role:    "user",
			Content: "[Execution Summary]\n\n" + summary,
		})
	}

	newTokens := tokenizer.EstimateTokens(newMsgs)
	fmt.Printf("%s✓ Summary complete (tokens %d → %d)%s\n",
		colors.BRIGHT_GREEN, tokens, newTokens, colors.RESET)

	return newMsgs, nil
}

func (s *Summarizer) createSummary(ctx context.Context, msgs []schema.Message, round int) (string, error) {
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

	resp, err := s.client.Generate(ctx, req, nil)
	if err != nil {
		return sb.String(), err
	}

	return resp.Content, nil
}

