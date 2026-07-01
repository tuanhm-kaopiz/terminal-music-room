package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

type tickMsg time.Time

type storeUpdateMsg struct{}

// Update handles Bubble Tea messages.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(0, msg.Width-4)
		return m, nil

	case tickMsg:
		m.refresh()
		return m, tickCmd()

	case storeUpdateMsg:
		m.refresh()
		return m, waitStoreCmd(m.storeCh)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quit = true
			if m.cfg.Leave != nil {
				_ = m.cfg.Leave(m.ctx)
			}
			return m, tea.Quit
		case "enter":
			body := strings.TrimSpace(m.input.Value())
			if body == "" {
				return m, nil
			}
			if m.cfg.Send != nil {
				if err := m.cfg.Send(m.ctx, protocol.MsgChatSend, protocol.ChatSendPayload{Body: body}); err != nil {
					m.errMsg = err.Error()
				} else {
					m.errMsg = ""
				}
			}
			m.input.SetValue("")
			return m, textinput.Blink
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}
