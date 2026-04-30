package anthropic

import (
	"encoding/json"
	"strings"
)

type streamState struct {
	text         strings.Builder // accumulate text_data chunks
	inputTokens  int             // message start
	outputTokens int             // message delta
	stopReason   string          // message delta
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

type textDelta struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicMessageEvent string

const (
	messageStart      anthropicMessageEvent = "message_start"
	messageDelta      anthropicMessageEvent = "message_delta"
	messageStop       anthropicMessageEvent = "message_stop"
	contentBlockStart anthropicMessageEvent = "content_block_start"
	contentBlockDelta anthropicMessageEvent = "content_block_delta"
	contentBlockStop  anthropicMessageEvent = "content_block_stop"
)
