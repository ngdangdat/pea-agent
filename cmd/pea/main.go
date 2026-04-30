package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/ngdangdat/pea-agent/internal/llm"
	"github.com/ngdangdat/pea-agent/internal/llm/anthropic"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pea <prompt>")
		os.Exit(1)
	}

	prompt := os.Args[1]
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	stream := anthropic.Stream(ctx, llm.Model{
		ID:     "claude-haiku-4-5-20251001",
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}, llm.Context{Messages: []llm.Message{{Role: "user", Content: prompt}}})
	for ev := range stream {
		switch ev.Kind {
		case llm.EventTextDelta:
			fmt.Print(ev.Delta)
		case llm.EventDone:
			fmt.Println()
		case llm.EventError:
			fmt.Fprintln(os.Stderr, "\nerror:", ev.Err)
			os.Exit(1)
		}
	}
}
