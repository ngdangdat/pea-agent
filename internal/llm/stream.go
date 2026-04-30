package llm

import "context"

type Message struct {
	Role    string
	Content string
}

type Context struct {
	SystemPrompt string
	Messages     []Message
}

type Model struct {
	ID     string
	APIKey string
}

func Stream(ctx context.Context, model Model, c Context) <-chan Event {
	out := make(chan Event, 16)
	go func() {
		close(out)
		// 1. build HTTP request with ctx
		// 2. read SSE lines
		// 3. for each line, decode JSON and send the right Event on `out`
		// 4. on error or ctx cancellation: send EventError, return
		// 5. on natural end: send EventDone, return
	}()

	return out
}
