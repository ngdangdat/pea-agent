# pea-agent

A small Go CLI that streams responses from Claude.

## Requirements

- Go 1.25+
- `ANTHROPIC_API_KEY` in your environment

## Build

```sh
go build -o bin/pea ./cmd/pea
```

## Usage

```sh
export ANTHROPIC_API_KEY=sk-ant-...
./bin/pea "why is the sky blue?"
```

Or run directly:

```sh
go run ./cmd/pea "hello"
```

Text deltas stream to stdout. Press `Ctrl+C` to abort.

## Layout

- `cmd/pea` — CLI entrypoint
- `internal/llm` — provider-agnostic stream/event types
- `internal/llm/anthropic` — Anthropic Messages API streaming client (SSE)
