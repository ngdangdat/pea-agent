package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

const defaultBaseURL = "https://api.openai.com/v1"

type Provider struct{}

func (Provider) Stream(ctx context.Context, model llm.Model, c llm.Context) <-chan llm.Event {
	out := make(chan llm.Event, 16)
	go func() {
		defer close(out)
		req, err := buildRequest(ctx, model, c)
		if err != nil {
			emitErr(out, err, "error")
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			emitErr(out, err, classifyAbort(ctx))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			emitErr(out, readAPIError(resp), classifyAbort(ctx))
			return
		}
		parseSSE(resp.Body, out)
	}()
	return out
}

func parseSSE(body io.Reader, out chan<- llm.Event) error {
	state := &openaiStreamState{toolCalls: map[int]*openaiToolBuf{}}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	out <- llm.Event{Kind: llm.EventStart}
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(line[5:])
		if payload == "[DONE]" {
			finalize(state, out)
			return nil
		}

		var chunk openaiChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		handleSSEEvent(&chunk, out, state)
	}

	return nil
}

func handleSSEEvent(chunk *openaiChunk, out chan<- llm.Event, state *openaiStreamState) {
	if chunk.Usage != nil {
		state.inputTokens = chunk.Usage.PromptTokens
		state.outputTokens = chunk.Usage.CompletionTokens
	}
	if len(chunk.Choices) == 0 {
		return
	}
	choice := chunk.Choices[0]
	if choice.Delta.Content != "" {
		if !state.textOpen {
			state.textOpen = true
			out <- llm.Event{Kind: llm.EventTextStart}
		}
		state.textBuf.WriteString(choice.Delta.Content)
		out <- llm.Event{Kind: llm.EventTextDelta, Delta: choice.Delta.Content}
	}
	for _, tc := range choice.Delta.ToolCalls {
		buf, ok := state.toolCalls[tc.Index]
		if !ok {
			buf = &openaiToolBuf{}
			state.toolCalls[tc.Index] = buf
			state.toolOrder = append(state.toolOrder, tc.Index)
		}
		if tc.ID != "" {
			buf.id = tc.ID
		}
		if tc.Function.Name != "" {
			buf.name = tc.Function.Name
		}
		if !buf.started && buf.id != "" && buf.name != "" {
			buf.started = true
			out <- llm.Event{Kind: llm.EventToolCallStart, ToolCall: &llm.ToolCall{
				ID:   buf.id,
				Name: buf.name,
			},
			}
		}
		if tc.Function.Arguments != "" {
			buf.args.WriteString(tc.Function.Arguments)
			out <- llm.Event{Kind: llm.EventToolCallDelta, Delta: tc.Function.Arguments}
		}
	}
	if choice.FinishReason != "" {
		state.stopReason = normalizeStopReason(choice.FinishReason)
		if state.textOpen {
			state.assistantContent = append(state.assistantContent, llm.ContentBlock{
				Type: "text",
				Text: state.textBuf.String(),
			})
			state.textBuf.Reset()
			state.textOpen = false
			out <- llm.Event{Kind: llm.EventTextEnd}
		}
		for _, tOrder := range state.toolOrder {
			buf := state.toolCalls[tOrder]
			tc := &llm.ToolCall{
				ID:    buf.id,
				Name:  buf.name,
				Input: json.RawMessage(buf.args.String()),
			}
			state.assistantContent = append(state.assistantContent, llm.ContentBlock{
				Type:     "tool_use",
				ToolCall: tc,
			})
			out <- llm.Event{Kind: llm.EventToolCallEnd, ToolCall: tc}
		}
	}
}

func finalize(state *openaiStreamState, out chan<- llm.Event) {
	out <- llm.Event{
		Kind: llm.EventDone,
		Final: &llm.AssistantMessage{
			Content:    state.assistantContent,
			StopReason: state.stopReason,
			Usage: llm.Usage{
				InputTokens:  state.inputTokens,
				OutputTokens: state.outputTokens,
			},
		},
	}
}

func emitErr(out chan<- llm.Event, err error, stopReason string) {
	out <- llm.Event{
		Kind:  llm.EventError,
		Err:   err,
		Final: &llm.AssistantMessage{StopReason: stopReason},
	}
}

func normalizeStopReason(r string) string {
	if r == "tool_calls" {
		return "tool_use"
	}
	return r
}

func classifyAbort(ctx context.Context) string {
	if ctx.Err() != nil {
		return "aborted"
	}
	return "error"
}

func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Error struct {
			Message string `json:"message"`
			Type    string `json:"type"`
			Code    string `json:"code"`
		}
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return fmt.Errorf("openai http_status=%d code=%s (%s): %s",
			resp.StatusCode, parsed.Error.Code, parsed.Error.Type, parsed.Error.Message,
		)
	}
	return fmt.Errorf("openai http_status=%d: %s", resp.StatusCode, body)
}

func init() {
	llm.Register("openai", Provider{})
}
