package tui

import (
	"context"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ngdangdat/pea-agent/internal/agent"
	"github.com/ngdangdat/pea-agent/internal/llm"
)

type Command string

const (
	helloText = `Hello, I'm pea-agent. Ask me anything . . .`
)

type Model struct {
	input         textinput.Model
	history       []string
	width         int
	streamingLine string
	cfg           agent.Config
	appCtx        context.Context
	//       turnCancel    context.CancelFunc    // cancels the current turn; nil when idle
	//       totalTokens   int                   // running token count, shown in footer
	//       spinner       spinner.Model         // animated indicator during tool calls
	//       busyReason    string                // "" when idle, e.g. "running bash" otherwise
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = helloText
	ti.Focus()
	return Model{input: ti}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			prompt := m.input.Value()
			if prompt == "/exit" {
				return m, tea.Quit
			}
			m.input.SetValue("")
			m.history = append(m.history, "> "+prompt)
			m.streamingLine = ""
			ch := agent.RunStreaming(m.appCtx, m.cfg, prompt)
			return m, waitForAgent(ch)
		}
	case agentEventMsg:
		switch msg.ev.Kind {
		}
	case agentDoneMsg:
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	lines := make([]string, 0, len(m.history))
	for _, line := range m.history {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n") + m.input.View()
}

func Run() (tea.Model, error) {
	return tea.NewProgram(New()).Run()
}

type agentEventMsg struct {
	ev llm.Event
	ch <-chan llm.Event
}

type agentDoneMsg struct{}

func waitForAgent(ch <-chan llm.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return agentDoneMsg{}
		}
		return agentEventMsg{ev: ev, ch: ch}
	}
}
