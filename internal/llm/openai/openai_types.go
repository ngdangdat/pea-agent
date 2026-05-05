package openai

import (
	"encoding/json"
	"strings"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

type openaiStreamState struct {
	textOpen     bool
	textBuf      strings.Builder
	toolCalls    map[int]*openaiToolBuf
	toolOrder    []int
	stopReason   string
	inputTokens  int
	outputTokens int

	assistantContent []llm.ContentBlock
}

type openaiToolBuf struct {
	id      string
	name    string
	args    strings.Builder
	started bool
}

type openaiChunk struct {
	Choices []openaiChunkChoice `json:"choices"`
	Usage   *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

type openaiChunkChoice struct {
	Delta struct {
		Content   string `json:"content"`
		ToolCalls []struct {
			Index    int    `json:"index"`
			ID       string `json:"id"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	} `json:"delta"`
	FinishReason string `json:"finish_reason"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type openaiToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function openaiToolCallFunction `json:"function"`
}

type openaiToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string                       `json:"type"`
	Function openaiToolFunctionDefinition `json:"function"`
}

type openaiToolFunctionDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}
