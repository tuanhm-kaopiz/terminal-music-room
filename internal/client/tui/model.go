package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/state"
)

const refreshInterval = 500 * time.Millisecond

// Config wires the TUI to an existing CLI session (AC-055).
type Config struct {
	Store *state.Store
	Send  func(ctx context.Context, msgType string, payload any) error
	Leave func(ctx context.Context) error
}

// Model is the Bubble Tea root model.
type Model struct {
	ctx     context.Context
	cfg     Config
	view    state.View
	input   textinput.Model
	storeCh <-chan struct{}
	width   int
	height  int
	quit    bool
	errMsg  string
}

// NewModel builds a model from runtime dependencies.
func NewModel(ctx context.Context, cfg Config) Model {
	in := textinput.New()
	in.Placeholder = "Type message… (q quit)"
	in.CharLimit = 500
	in.Focus()
	in.Prompt = "> "

	return Model{
		ctx:     ctx,
		cfg:     cfg,
		view:    cfg.Store.Snapshot(),
		input:   in,
		storeCh: cfg.Store.SubscribeRoom(),
	}
}

// Init starts periodic refresh and store subscription (AC-054).
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.input.Focus(),
		textinput.Blink,
		tickCmd(),
		waitStoreCmd(m.storeCh),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func waitStoreCmd(ch <-chan struct{}) tea.Cmd {
	return func() tea.Msg {
		_, ok := <-ch
		if !ok {
			return nil
		}
		return storeUpdateMsg{}
	}
}

func (m *Model) refresh() {
	m.view = m.cfg.Store.Snapshot()
}
