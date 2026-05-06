package agent

import (
	"context"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

func RunStreaming(ctx context.Context, cfg Config, prompt string) <-chan llm.Event {
	out := make(chan llm.Event, 16)
	go func() {
		defer close(out)
	}()
	return out
}
