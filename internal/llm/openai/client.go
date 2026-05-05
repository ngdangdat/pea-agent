package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

// based on Anthropic provider, write these functions
// `buildRequest`, `parseSSE`, `emitErr`, `classifyAbort`, `readAPIError`
func buildRequest(ctx context.Context, model llm.Model, c llm.Context) (*http.Request, error) {
	body, err := json.Marshal(map[string]any{
		"stream":   true,
		"model":    model.ID,
		"messages": toOpenAIMessages(c),
		"tools":    toOpenAITools(c.Tools),
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("%s%s", defaultBaseURL, "/chat/completions")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	var apiKey string = model.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI API key is not set")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Set("content-type", "application/json")
	return req, nil
}

func toOpenAIMessages(c llm.Context) []openaiMessage {
	// in case it has system prompt, too
	out := make([]openaiMessage, 0, len(c.Messages)+1)

	if c.SystemPrompt != "" {
		out = append(out, openaiMessage{
			Role:    "system",
			Content: c.SystemPrompt,
		})
	}
	for _, msg := range c.Messages {
		switch msg.Role {
		case "user":
			for _, b := range msg.Blocks {
				if b.Type == "tool_result" {
					out = append(out, openaiMessage{
						Role:       "tool",
						ToolCallID: b.ToolUseID,
						Content:    b.Content,
					})
				}
			}
			if msg.Content != "" {
				out = append(out, openaiMessage{
					Role:    "user",
					Content: msg.Content,
				})
			}
		case "assistant":
			m := openaiMessage{
				Role: msg.Role,
			}
			for _, b := range msg.Blocks {
				switch b.Type {
				case "text":
					m.Content = b.Text
				case "tool_use":
					m.ToolCalls = append(m.ToolCalls, openaiToolCall{
						ID:   b.ToolCall.ID,
						Type: "function",
						Function: openaiToolCallFunction{
							Name:      b.ToolCall.Name,
							Arguments: string(b.ToolCall.Input),
						},
					})
				}
			}
			out = append(out, m)
		}
	}
	return out
}

func toOpenAITools(tools []llm.Tool) []openaiTool {
	out := make([]openaiTool, 0, len(tools))
	for _, tool := range tools {
		out = append(out, openaiTool{
			Type: "function",
			Function: openaiToolFunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.InputSchema,
			},
		})
	}
	return out
}
