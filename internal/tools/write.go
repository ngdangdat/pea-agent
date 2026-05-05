package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ngdangdat/pea-agent/internal/agent"
)

func Write() agent.Tool {
	return agent.Tool{
		Name:        "write",
		Description: "Write a UTF-8 file to disk",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {"path": {"type": "string"}, "content": {"type": "string"}},
			"required": ["path", "content"]
		}`),
		Execute: func(ctx context.Context, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct {
				Path    string
				Content string
			}
			if err := json.Unmarshal(raw, &args); err != nil {
				return agent.ToolResult{
					IsError: true,
					Content: fmt.Sprintf("invalid args: %v", err),
				}, nil
			}
			err := os.WriteFile(args.Path, []byte(args.Content), 0644)
			if err != nil {
				return agent.ToolResult{IsError: true, Content: err.Error()}, nil
			}
			return agent.ToolResult{Content: fmt.Sprintf("Wrote %d char to %s", len(args.Content), args.Path)}, nil
		},
	}
}
