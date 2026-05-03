package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ngdangdat/pea-agent/internal/agent"
)

func Read() agent.Tool {
	return agent.Tool{
		Name:        "read",
		Description: "Read a UTF-8 file from disk",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {"path": {"type": "string"}},
			"required": ["path"]
		}`),
		Execute: func(ctx context.Context, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct{ Path string }
			if err := json.Unmarshal(raw, &args); err != nil {
				return agent.ToolResult{
					IsError: true,
					Content: fmt.Sprintf("invalid args: %v", err),
				}, nil
			}
			data, err := os.ReadFile(args.Path)
			if err != nil {
				return agent.ToolResult{IsError: true, Content: err.Error()}, nil
			}
			return agent.ToolResult{Content: string(data)}, nil
		},
	}
}
