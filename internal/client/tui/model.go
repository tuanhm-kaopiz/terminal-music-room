package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/actions"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/client/tui/modals"
)

const refreshInterval = 500 * time.Millisecond

// Mode is the TUI interaction layer (ADR-002).
type Mode int

const (
	ModeDashboard Mode = iota
	ModeModalAdd
	ModeModalSearch
	ModeModalSeek
	ModeModalLeave
	ModeHelp
)

// FocusPanel is the keyboard-focused HUD region.
type FocusPanel int

const (
	FocusQueue FocusPanel = iota
	FocusChat
	FocusMembers
)

// Config wires the TUI to an existing CLI session (AC-055).
type Config struct {
	Store   *state.Store
	Actions *actions.Room
	// Send is a legacy fallback when Actions is nil.
	Send  func(ctx context.Context, msgType string, payload any) error
	Leave func(ctx context.Context) error
}

func (c Config) roomActions() *actions.Room {
	if c.Actions != nil {
		return c.Actions
	}
	if c.Store != nil && c.Send != nil {
		return actions.New(c.Send, c.Store)
	}
	return nil
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
	mode    Mode
	focus   FocusPanel
	selectedQueueIdx int
	queueScroll      int
	chatScroll       int
	membersScroll    int
	addModal  modals.AddSource
	seekModal modals.Seek
	leaveModal modals.ConfirmLeave
	quit    bool
	errMsg  string
}

// IsHost reports whether the current session owns the room.
func IsHost(v state.View) bool {
	return v.SessionID != "" && v.Room.HostID == v.SessionID
}

// NewModel builds a model from runtime dependencies.
func NewModel(ctx context.Context, cfg Config) Model {
	in := textinput.New()
	in.Placeholder = "Type message… (? help · q exit TUI)"
	in.CharLimit = 500
	in.Focus()
	in.Prompt = "> "

	return Model{
		ctx:     ctx,
		cfg:     cfg,
		view:    cfg.Store.Snapshot(),
		input:   in,
		storeCh: cfg.Store.SubscribeRoom(),
		mode:    ModeDashboard,
		focus:   FocusChat,
	}
}

func (m *Model) actions() *actions.Room {
	return m.cfg.roomActions()
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
	prevChatLen := len(m.view.Room.Chat)
	m.view = m.cfg.Store.Snapshot()
	if len(m.view.Room.Chat) > prevChatLen {
		m.chatScroll = 0
	}
	if m.view.LastErr != nil && m.view.LastErr.Message != "" {
		m.errMsg = m.view.LastErr.Message
	}
	n := len(m.view.Room.Queue)
	if n == 0 {
		m.selectedQueueIdx = 0
		m.queueScroll = 0
	} else {
		if m.selectedQueueIdx >= n {
			m.selectedQueueIdx = n - 1
		}
		if m.selectedQueueIdx < 0 {
			m.selectedQueueIdx = 0
		}
		m.ensureQueueVisible()
	}
	members := len(m.view.Room.Members)
	if members == 0 {
		m.membersScroll = 0
	} else if m.membersScroll >= members {
		m.membersScroll = members - 1
	}
}
