package llm

import (
	"context"
	"fmt"
)

type Provider interface {
	Stream(ctx context.Context, model Model, c Context) <-chan Event
}

var providers = map[string]Provider{}

func Register(name string, p Provider) {
	if _, exists := providers[name]; exists {
		panic("provider already registered: " + name)
	}
	providers[name] = p
}

func Stream(ctx context.Context, model Model, c Context) <-chan Event {
	p, ok := providers[model.Provider]
	if !ok {
		ch := make(chan Event, 1)
		ch <- Event{Kind: EventError, Err: fmt.Errorf("no provider: %s", model.Provider)}
		close(ch)
		return ch
	}
	return p.Stream(ctx, model, c)
}
