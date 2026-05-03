package agent

import (
	"context"
	"encoding/json"
)

type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
	Execute     func(ctx context.Context, input json.RawMessage) (ToolResult, error)
}

type ToolResult struct {
	Content string
	IsError bool
}
