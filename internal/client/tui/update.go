package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/terminal-music-room/music-room/internal/client/tui/keys"
	"github.com/terminal-music-room/music-room/internal/client/tui/layout"
	"github.com/terminal-music-room/music-room/internal/client/tui/modals"
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
		if m.mode == ModeModalAdd {
			m.addModal = m.addModal.WithWidth(msg.Width)
		}
		if m.mode == ModeModalSeek {
			m.seekModal = m.seekModal.WithWidth(msg.Width)
		}
		return m, nil

	case tickMsg:
		m.refresh()
		return m, tickCmd()

	case storeUpdateMsg:
		m.refresh()
		return m, waitStoreCmd(m.storeCh)

	case tea.KeyMsg:
		if key := msg.String(); key == "ctrl+c" || key == keys.KeyQuit {
			m.quit = true
			return m, tea.Quit
		}
		if m.mode == ModeModalAdd {
			return m.updateAddModal(msg)
		}
		if m.mode == ModeModalSeek {
			return m.updateSeekModal(msg)
		}
		if m.mode == ModeModalLeave {
			return m.updateLeaveModal(msg)
		}
		if m.mode == ModeHelp {
			if msg.String() == "esc" || msg.String() == keys.KeyHelp {
				m.mode = ModeDashboard
				return m, nil
			}
			return m, nil
		}
		if cmd, handled := m.handleKey(msg.String()); handled {
			return m, cmd
		}
	}

	if m.mode == ModeModalAdd {
		var cmd tea.Cmd
		m.addModal, cmd = m.addModal.Update(msg)
		return m, cmd
	}
	if m.mode == ModeModalSeek {
		var cmd tea.Cmd
		m.seekModal, cmd = m.seekModal.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) handleKey(key string) (tea.Cmd, bool) {
	switch key {
	case "ctrl+c", keys.KeyQuit:
		m.quit = true
		return tea.Quit, true
	case "esc":
		if m.mode == ModeHelp {
			m.mode = ModeDashboard
			return nil, true
		}
		if m.mode != ModeDashboard {
			m.closeModal()
			return nil, true
		}
		return nil, false
	case keys.KeyHelp:
		if m.mode == ModeHelp {
			m.mode = ModeDashboard
		} else if m.mode == ModeDashboard {
			m.mode = ModeHelp
		}
		return nil, true
	case "enter":
		return m.handleChatEnter(), true
	case keys.KeyPauseToggle:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handlePlaybackToggle(), true
	case keys.KeySkip:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handlePlaybackSkip(), true
	case keys.KeyAddSource:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.openAddModal(), true
	case keys.KeySeek, "shift+s":
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.openSeekModal(), true
	case keys.KeyVoteSkip:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleVoteSkip(), true
	case keys.KeyVotePriority, "shift+v":
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleVotePriority(), true
	case keys.KeyQueueRemove:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleQueueRemove(), true
	case keys.KeyQueueUp:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleQueueReorder(-1), true
	case keys.KeyQueueDown:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleQueueReorder(1), true
	case keys.KeyLeave:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.openLeaveModal(), true
	case keys.KeyTab:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.cycleFocus(1), true
	case keys.KeyShiftTab:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.cycleFocus(-1), true
	case keys.KeyUp:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleFocusScroll(-1), true
	case keys.KeyDown:
		if m.mode != ModeDashboard {
			return nil, true
		}
		return m.handleFocusScroll(1), true
	default:
		if emoji, ok := keys.ReactionForKey(key); ok {
			if m.mode != ModeDashboard {
				return nil, true
			}
			return m.handleReaction(emoji), true
		}
		return nil, false
	}
}

func (m *Model) handleChatEnter() tea.Cmd {
	if m.mode == ModeHelp {
		return nil
	}
	body := strings.TrimSpace(m.input.Value())
	if body == "" {
		return nil
	}
	act := m.actions()
	if act == nil {
		return textinput.Blink
	}
	if err := act.Chat(m.ctx, body); err != nil {
		m.errMsg = err.Error()
		return textinput.Blink
	}
	m.errMsg = ""
	m.input.SetValue("")
	return textinput.Blink
}

func (m *Model) handlePlaybackToggle() tea.Cmd {
	act := m.actions()
	if act == nil {
		return nil
	}
	action, err := keys.PlaybackToggle(m.view)
	if err != nil {
		m.errMsg = err.Error()
		return nil
	}
	var sendErr error
	switch action {
	case keys.PlaybackPause:
		sendErr = act.Pause(m.ctx)
	case keys.PlaybackResume:
		sendErr = act.Resume(m.ctx)
	}
	if sendErr != nil {
		m.errMsg = sendErr.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) handlePlaybackSkip() tea.Cmd {
	act := m.actions()
	if act == nil {
		return nil
	}
	if err := keys.RequireTrack(m.view); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	if err := act.Skip(m.ctx); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) selectedQueueItemID() string {
	items := m.view.Room.Queue
	if len(items) == 0 {
		return ""
	}
	idx := m.selectedQueueIdx
	if idx < 0 || idx >= len(items) {
		idx = 0
	}
	return items[idx].ID
}

func (m *Model) handleVoteSkip() tea.Cmd {
	act := m.actions()
	if act == nil {
		return nil
	}
	if err := keys.RequireTrack(m.view); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	if err := act.VoteSkip(m.ctx); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) handleVotePriority() tea.Cmd {
	act := m.actions()
	if act == nil {
		return nil
	}
	itemID := m.selectedQueueItemID()
	if itemID == "" {
		m.errMsg = "no queue item selected"
		return nil
	}
	if err := act.VotePriority(m.ctx, itemID); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) handleReaction(emoji string) tea.Cmd {
	act := m.actions()
	if act == nil {
		return nil
	}
	if err := keys.RequireTrack(m.view); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	if err := act.React(m.ctx, emoji); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) handleQueueRemove() tea.Cmd {
	if err := keys.RequireHost(IsHost(m.view)); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	itemID := m.selectedQueueItemID()
	if itemID == "" {
		m.errMsg = "no queue item selected"
		return nil
	}
	act := m.actions()
	if act == nil {
		return nil
	}
	if err := act.QueueRemove(m.ctx, itemID); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

func (m *Model) handleQueueReorder(direction int) tea.Cmd {
	if err := keys.RequireHost(IsHost(m.view)); err != nil {
		m.errMsg = err.Error()
		return nil
	}
	items := m.view.Room.Queue
	idx := m.selectedQueueIdx
	if idx < 0 || idx >= len(items) {
		idx = 0
	}
	itemID, afterID, ok := keys.QueueReorderTargets(items, idx, direction)
	if !ok {
		m.errMsg = "cannot reorder queue item"
		return nil
	}
	act := m.actions()
	if act == nil {
		return nil
	}
	if err := act.QueueReorder(m.ctx, itemID, afterID); err != nil {
		m.errMsg = err.Error()
	} else {
		m.errMsg = ""
	}
	return nil
}

var focusOrder = []FocusPanel{FocusQueue, FocusChat, FocusMembers}

func (m *Model) cycleFocus(delta int) tea.Cmd {
	idx := 0
	for i, p := range focusOrder {
		if p == m.focus {
			idx = i
			break
		}
	}
	n := len(focusOrder)
	idx = keys.CycleIndex(idx, delta, n)
	m.focus = focusOrder[idx]
	if m.focus == FocusChat {
		return m.input.Focus()
	}
	m.input.Blur()
	return textinput.Blink
}

func (m *Model) handleFocusScroll(delta int) tea.Cmd {
	switch m.focus {
	case FocusQueue:
		m.moveQueueSelection(delta)
	case FocusChat:
		if delta < 0 {
			m.chatScroll++
		} else if m.chatScroll > 0 {
			m.chatScroll--
		}
	case FocusMembers:
		if delta < 0 {
			m.membersScroll++
		} else if m.membersScroll > 0 {
			m.membersScroll--
		}
	}
	return nil
}

func (m *Model) moveQueueSelection(delta int) {
	n := len(m.view.Room.Queue)
	if n == 0 {
		return
	}
	m.selectedQueueIdx += delta
	if m.selectedQueueIdx < 0 {
		m.selectedQueueIdx = 0
	}
	if m.selectedQueueIdx >= n {
		m.selectedQueueIdx = n - 1
	}
	m.ensureQueueVisible()
}

func (m *Model) ensureQueueVisible() {
	visible := m.queueVisibleRows()
	if visible < 1 {
		visible = 1
	}
	if m.selectedQueueIdx < m.queueScroll {
		m.queueScroll = m.selectedQueueIdx
	}
	if m.selectedQueueIdx >= m.queueScroll+visible {
		m.queueScroll = m.selectedQueueIdx - visible + 1
	}
}

func (m *Model) queueVisibleRows() int {
	if m.width == 0 || m.height == 0 {
		return 1
	}
	reg := layout.Compute(m.width, m.height, false)
	if reg.Degraded {
		return 1
	}
	h := reg.Queue.Height
	if h < 3 {
		return 1
	}
	return h - 2
}

func (m *Model) openAddModal() tea.Cmd {
	m.mode = ModeModalAdd
	m.addModal = modals.NewAddSource(m.width)
	return textinput.Blink
}

func (m *Model) closeModal() {
	m.addModal = modals.AddSource{}
	m.seekModal = modals.Seek{}
	m.leaveModal = modals.ConfirmLeave{}
	m.mode = ModeDashboard
}

func (m *Model) openLeaveModal() tea.Cmd {
	m.mode = ModeModalLeave
	m.leaveModal = modals.NewConfirmLeave(m.view.Room.Slug)
	return nil
}

func (m *Model) updateLeaveModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeModal()
		return m, nil
	case "y", "enter":
		return m.confirmLeave()
	default:
		return m, nil
	}
}

func (m *Model) confirmLeave() (tea.Model, tea.Cmd) {
	if m.cfg.Leave != nil {
		if err := m.cfg.Leave(m.ctx); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
	} else if act := m.actions(); act != nil {
		if err := act.Leave(m.ctx); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
	}
	m.errMsg = ""
	m.closeModal()
	m.quit = true
	return m, tea.Quit
}

func (m *Model) openSeekModal() tea.Cmd {
	m.mode = ModeModalSeek
	m.seekModal = modals.NewSeek(m.width)
	return textinput.Blink
}

func (m *Model) updateSeekModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeModal()
		return m, nil
	case "enter":
		return m.submitSeekModal()
	default:
		var cmd tea.Cmd
		m.seekModal, cmd = m.seekModal.Update(msg)
		return m, cmd
	}
}

func (m *Model) submitSeekModal() (tea.Model, tea.Cmd) {
	act := m.actions()
	if act == nil {
		return m, nil
	}
	if err := m.seekModal.Submit(m.ctx, act, m.view); err != nil {
		m.errMsg = err.Error()
		return m, textinput.Blink
	}
	m.errMsg = ""
	m.closeModal()
	return m, nil
}

func (m *Model) updateAddModal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeModal()
		return m, nil
	case "tab":
		m.addModal.Intent = modals.ToggleIntent(m.addModal.Intent)
		return m, nil
	case "enter":
		return m.submitAddModal()
	default:
		var cmd tea.Cmd
		m.addModal, cmd = m.addModal.Update(msg)
		return m, cmd
	}
}

func (m *Model) submitAddModal() (tea.Model, tea.Cmd) {
	act := m.actions()
	if act == nil {
		return m, nil
	}
	if err := m.addModal.Submit(m.ctx, act); err != nil {
		m.errMsg = err.Error()
		return m, textinput.Blink
	}
	m.errMsg = ""
	m.closeModal()
	return m, nil
}
