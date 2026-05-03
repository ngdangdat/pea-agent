package tools

import (
	"context"
	"encoding/json"
	"os/exec"

	"github.com/ngdangdat/pea-agent/internal/agent"
)

func Bash() agent.Tool {
	return agent.Tool{
		Name:        "bash",
		Description: "Run a shell command. Return combied stdout+stderr.",
		InputSchema: json.RawMessage(`
			{
				"type": "object",
				"properties": {"command": {"type": "string"}},
				"required": ["command"]
			}
		`),
		Execute: func(ctx context.Context, raw json.RawMessage) (agent.ToolResult, error) {
			var args struct{ Command string }
			json.Unmarshal(raw, &args)
			cmd := exec.CommandContext(ctx, "bash", "-c", args.Command)
			out, err := cmd.CombinedOutput()
			res := agent.ToolResult{Content: string(out)}
			if err != nil {
				res.IsError = true
				res.Content += "\n" + err.Error()
			}
			return res, nil
		},
	}
}
