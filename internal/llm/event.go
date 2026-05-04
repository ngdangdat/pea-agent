package llm

import "encoding/json"

type EventKind int

const (
	EventStart EventKind = iota
	EventTextStart
	EventTextDelta
	EventTextEnd
	EventToolCallStart
	EventToolCallDelta
	EventToolCallEnd
	EventDone
	EventError
)

type Event struct {
	Kind     EventKind
	Delta    string
	Final    *AssistantMessage
	ToolCall *ToolCall
	Err      error
}

type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

type ContentBlock struct {
	Type      string
	Text      string
	ToolCall  *ToolCall
	ToolUseID string
	Content   string
	IsError   bool
}

type AssistantMessage struct {
	Content    []ContentBlock
	StopReason string // stop | error | aborted
	Usage      Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}

type Message struct {
	Role    string
	Content string
	Blocks  []ContentBlock
}

type Tool struct {
	Name        string
	Description string
	InputSchema json.RawMessage
}

type Context struct {
	SystemPrompt string
	Messages     []Message
	Tools        []Tool
}

type Model struct {
	ID       string
	APIKey   string
	BaseURL  string
	Provider string
}
