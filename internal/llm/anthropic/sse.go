package anthropic

import (
	"encoding/json"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

func handleSSEEvent(line string, out chan<- llm.Event, state *streamState) {
	var env sseEnvelope
	if err := json.Unmarshal([]byte(line), &env); err != nil {
	}
	switch env.Type {
	case messageStart:
		out <- llm.Event{Kind: llm.EventStart}
	case contentBlockStart:
		out <- llm.Event{Kind: llm.EventTextStart}

	case contentBlockDelta:
		var d textDelta
		json.Unmarshal(env.Delta, &d)
		if d.Type == "text_delta" {
			state.text.WriteString(d.Text)
			out <- llm.Event{Kind: llm.EventTextDelta, Delta: d.Text}
		}
	case contentBlockStop:
		out <- llm.Event{Kind: llm.EventTextEnd}
	case messageDelta:
		out <- llm.Event{}
		if env.Usage != nil {
			state.outputTokens = env.Usage.OutputTokens
		}
	case messageStop:
		out <- llm.Event{Kind: llm.EventDone, Final: &llm.AssistantMessage{
			Text:       state.text.String(),
			StopReason: state.stopReason,
			Usage: llm.Usage{
				InputTokens:  state.inputTokens,
				OutputTokens: state.outputTokens,
			},
		}}
	}
}
