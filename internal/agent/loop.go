package agent

import (
	"context"
	"fmt"

	"github.com/ngdangdat/pea-agent/internal/llm"
	"github.com/ngdangdat/pea-agent/internal/llm/anthropic"
)

type Config struct {
	Model llm.Model
	Tools []Tool
}

type Event = llm.Event

func Run(ctx context.Context, cfg Config, prompt string) error {
	history := []llm.Message{{Role: "user", Content: prompt}}
	toolByName := map[string]Tool{}
	for _, t := range cfg.Tools {
		toolByName[t.Name] = t
	}

	for {
		c := llm.Context{Messages: history, Tools: toolsForLLM(cfg.Tools)}
		stream := anthropic.Stream(ctx, cfg.Model, c)

		var final *llm.AssistantMessage
		for ev := range stream {
			switch ev.Kind {
			case llm.EventTextDelta:
				fmt.Println(ev.Delta)
			case llm.EventDone:
				final = ev.Final
			case llm.EventError:
				return ev.Err
			}
		}
		fmt.Println()
		history = append(history, llm.Message{Role: "assistant", Blocks: final.Content})

		if final.StopReason != "tool_use" {
			return nil
		}

		var resultBlocks []llm.ContentBlock
		for _, b := range final.Content {
			if b.Type != "tool_use" {
				continue
			}
			tool, ok := toolByName[b.ToolCall.Name]
			if !ok {
				resultBlocks = append(resultBlocks, llm.ContentBlock{
					Type: "tool_result", ToolUseID: b.ToolCall.ID, Content: "no such tool", IsError: true,
				})
				continue
			}
			res, _ := tool.Execute(ctx, b.ToolCall.Input)
			resultBlocks = append(resultBlocks, llm.ContentBlock{
				Type:      "tool_result",
				ToolUseID: b.ToolCall.ID,
				Content:   res.Content,
				IsError:   res.IsError,
			})
		}
		history = append(history, llm.Message{
			Role:   "user",
			Blocks: resultBlocks,
		})

	}
}

func toolsForLLM(tools []Tool) []llm.Tool {
	out := make([]llm.Tool, 0, len(tools))
	for _, t := range tools {
		out = append(out, llm.Tool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}

	return out
}
