package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

const (
	apiUrl    = "https://api.anthropic.com/v1/messages"
	apiVer    = "2023-06-01"
	maxTokens = 1024
)

func buildRequest(ctx context.Context, model llm.Model, llmContext llm.Context) (*http.Request, error) {
	messages := []map[string]string{}
	for _, msg := range llmContext.Messages {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	body, err := json.Marshal(map[string]any{
		"model":      model.ID,
		"max_tokens": maxTokens,
		"stream":     true,
		"messages":   messages,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", apiUrl, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("error while building request: %v", err)
	}
	var apiKey string = model.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key is not passed")
	}
	req.Header.Set("x-api-key", model.APIKey)
	req.Header.Set("anthropic-version", apiVer)
	req.Header.Set("content-type", "application/json")
	return req, nil
}

func emitError(out chan<- llm.Event, err error, stopReason string) {
	out <- llm.Event{
		Kind:  llm.EventError,
		Err:   err,
		Final: &llm.AssistantMessage{StopReason: stopReason},
	}
}

func readAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	var parsed struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && parsed.Error.Message != "" {
		return fmt.Errorf("anthropic %d (%s): %s", resp.StatusCode, parsed.Error.Type, parsed.Error.Message)
	}
	return fmt.Errorf("anthropic %d: %s", resp.StatusCode, body)
}

func classifyAbort(ctx context.Context) string {
	if ctx.Err() != nil {
		return "aborted"
	}
	return "error"
}

func parseSSE(body io.Reader, out chan<- llm.Event, model llm.Model) error {
	state := streamState{}
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var dataLine string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "":
			if dataLine != "" {
				handleSSEEvent(dataLine, out, &state)
			}
			dataLine = ""
		case strings.HasPrefix(line, "data:"):
			dataLine = line[5:]
		}
	}
	return scanner.Err()
}

func Stream(ctx context.Context, model llm.Model, llmContext llm.Context) chan llm.Event {
	out := make(chan llm.Event, 16)
	go func() {
		defer close(out)
		req, err := buildRequest(ctx, model, llmContext)
		if err != nil {
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// return fmt.Errorf("send request: %w", err)
			emitError(out, err, classifyAbort(ctx))
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			// return fmt.Errorf("anthropic %d: %s", resp.StatusCode, msg)
			emitError(out, readAPIError(resp), "error")
			return
		}
		if err := parseSSE(resp.Body, out, model); err != nil {
			emitError(out, err, "error")
		}

	}()
	return out
}
