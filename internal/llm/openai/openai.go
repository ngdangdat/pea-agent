package openai

import (
	"context"

	"github.com/ngdangdat/pea-agent/internal/llm"
)

const defaultBaseURL = "https://api.openai.com/v1"

type Provider struct{}

func (Provider) Stream(ctx context.Context, model llm.Model, c llm.Context) <-chan llm.Event {
	out := make(chan llm.Event, 16)
	go func() {
		defer close(out)


	}
}
