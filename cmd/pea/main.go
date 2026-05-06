package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ngdangdat/pea-agent/internal/agent"
	"github.com/ngdangdat/pea-agent/internal/llm"
	_ "github.com/ngdangdat/pea-agent/internal/llm/anthropic"
	_ "github.com/ngdangdat/pea-agent/internal/llm/openai"
	"github.com/ngdangdat/pea-agent/internal/tools"
	"github.com/ngdangdat/pea-agent/internal/tui"
)

func main() {
	if len(os.Args) < 2 {
		tui.New()
		return
	}

	prompt := os.Args[1]
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	modelCfg := llm.Model{
		Provider: "openai",
		// Provider: "anthropic",
		ID: "gpt-4o-mini",
		// ID:       "claude-haiku-4-5-20251001",
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	agentCfg := agent.Config{
		Model: modelCfg,
		Tools: []agent.Tool{tools.Read(), tools.Bash(), tools.Write()},
	}
	err := agent.Run(ctx, agentCfg, prompt)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
