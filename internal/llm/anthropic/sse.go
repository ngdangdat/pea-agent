package anthropic

import (
	"encoding/json"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

func handleSSEEvent(line string, out chan<- llm.Event, state *streamState) {
	var env sseEnvelope
	if err := json.Unmarshal([]byte(line), &env); err != nil {
		out <- llm.Event{Kind: llm.EventError, Err: err}
	}
	// log.Printf("[SSE envelope] Type=[%s] Delta=[%s] Message=[%s]\n", env.Type, string(env.Delta), string(env.Message))
	// log.Printf("[SSE debug] line=[%s]", line)
	switch env.Type {
	case messageStart:
		var s struct {
			Usage anthropicUsage `json:"usage"`
		}
		if err := json.Unmarshal(env.Message, &s); err == nil {
			state.inputTokens = s.Usage.InputTokens
		}
		out <- llm.Event{Kind: llm.EventStart}
	case contentBlockStart:
		var s struct {
			ContentBlock struct {
				Type, ID, Name string
			} `json:"content_block"`
		}
		json.Unmarshal([]byte(line), &s)
		switch s.ContentBlock.Type {
		case "text":
			out <- llm.Event{Kind: llm.EventTextStart}
		case "tool_use":
			state.currentTool = &llm.ToolCall{ID: s.ContentBlock.ID, Name: s.ContentBlock.Name}
			state.toolInputBuf.Reset()
			out <- llm.Event{Kind: llm.EventToolCallStart}
		}

	case contentBlockDelta:
		var d struct {
			Delta struct {
				Type        string `json:"type"`
				Text        string `json:"text"`
				PartialJSON string `json:"partial_json"`
			} `json:"delta"`
		}
		json.Unmarshal([]byte(line), &d)
		switch d.Delta.Type {
		case "text_delta":
			state.text.WriteString(d.Delta.Text)
			out <- llm.Event{Kind: llm.EventTextDelta, Delta: d.Delta.Text}
		case "input_json_delta":
			state.toolInputBuf.WriteString(d.Delta.PartialJSON)
			out <- llm.Event{Kind: llm.EventToolCallDelta, Delta: d.Delta.PartialJSON}
		}
	case contentBlockStop:
		if state.currentTool != nil {
			state.currentTool.Input = json.RawMessage(state.toolInputBuf.String())
			state.assistantContent = append(state.assistantContent, llm.ContentBlock{
				Type: "tool_use", ToolCall: state.currentTool,
			})
			out <- llm.Event{Kind: llm.EventToolCallEnd, ToolCall: state.currentTool}
			state.currentTool = nil
		} else {
			state.assistantContent = append(state.assistantContent, llm.ContentBlock{Type: "text", Text: state.text.String()})
			state.text.Reset()
			out <- llm.Event{Kind: llm.EventTextEnd}
		}
	case messageDelta:
		out <- llm.Event{}
		if env.Usage != nil {
			state.outputTokens = env.Usage.OutputTokens
		}
		var d struct {
			StopReason string `json:"stop_reason"`
		}
		if err := json.Unmarshal(env.Delta, &d); err == nil && d.StopReason != "" {
			state.stopReason = d.StopReason
		}

	case messageStop:
		out <- llm.Event{Kind: llm.EventDone, Final: &llm.AssistantMessage{
			Content:    state.assistantContent,
			StopReason: state.stopReason,
			Usage: llm.Usage{
				InputTokens:  state.inputTokens,
				OutputTokens: state.outputTokens,
			},
		}}
	}
}
