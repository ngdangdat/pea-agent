package anthropic

import (
	"encoding/json"
	"strings"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

type streamState struct {
	text             strings.Builder // accumulate text_data chunks
	toolInputBuf     strings.Builder // accumulate tool_use chunks
	inputTokens      int             // message start
	outputTokens     int             // message delta
	stopReason       string          // message delta
	currentTool      *llm.ToolCall
	assistantContent []llm.ContentBlock
}
type sseEnvelope struct {
	Type    anthropicMessageEvent `json:"type"`
	Index   int                   `json:"index"`
	Delta   json.RawMessage       `json:"delta"`
	Message json.RawMessage       `json:"message"`
	Usage   *anthropicUsage       `json:"usage"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicMessageEvent string
type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}
type anthropicContentBlock struct {
	Type string `json:"type"`
	// type=text
	Text string `json:"text,omitempty"`
	// type=tool_use
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// type=tool_result
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

type anthropicMessageContent struct {
	Text   string                  `json:"text,omitempty"`
	Blocks []anthropicContentBlock `json:"blocks,omitempty"`
}

func (c anthropicMessageContent) MarshalJSON() ([]byte, error) {
	if c.Blocks != nil {
		return json.Marshal(c.Blocks)
	}
	return json.Marshal(c.Text)
}

type anthropicMessage struct {
	Role    string                  `json:"role"`
	Content anthropicMessageContent `json:"content"`
}

const (
	messageStart      anthropicMessageEvent = "message_start"
	messageDelta      anthropicMessageEvent = "message_delta"
	messageStop       anthropicMessageEvent = "message_stop"
	contentBlockStart anthropicMessageEvent = "content_block_start"
	contentBlockDelta anthropicMessageEvent = "content_block_delta"
	contentBlockStop  anthropicMessageEvent = "content_block_stop"
)
