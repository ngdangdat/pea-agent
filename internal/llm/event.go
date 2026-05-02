package llm

type EventKind int

const (
	EventStart EventKind = iota
	EventTextStart
	EventTextDelta
	EventTextEnd
	EventDone
	EventError
)

type Event struct {
	Kind  EventKind
	Delta string
	Final *AssistantMessage
	Err   error
}

type AssistantMessage struct {
	Text       string
	StopReason string // stop | error | aborted
	Usage      Usage
}

type Usage struct {
	InputTokens  int
	OutputTokens int
}
