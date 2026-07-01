package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-music-room/music-room/internal/client/tui/layout"
	"github.com/terminal-music-room/music-room/internal/client/tui/modals"
	"github.com/terminal-music-room/music-room/internal/client/tui/panels"
	"github.com/terminal-music-room/music-room/internal/client/tui/theme"
)

// View renders the sci-fi HUD dashboard (AC-009, AC-040).
func (m Model) View() string {
	if m.quit {
		return theme.Default().Muted().Render("Exiting TUI…\n")
	}
	if m.width == 0 || m.height == 0 {
		return theme.Default().Muted().Render("Loading…\n")
	}

	tm := theme.Default()
	reg := layout.Compute(m.width, m.height, false)
	v := m.view
	isHost := IsHost(v)
	queueOpts := m.renderOpts(FocusQueue)
	chatOpts := m.renderOpts(FocusChat)
	membersOpts := m.renderOpts(FocusMembers)

	var parts []string
	parts = append(parts, panels.Header(tm, v, reg.Header.Width, reg.Header.Height))

	if reg.Degraded {
		parts = append(parts, tm.Warning().Render(
			"⚠ terminal < 80×24 — degraded HUD (room + now playing prioritized)",
		))
		parts = append(parts, panels.NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, panels.RenderOpts{}))
		if !reg.Members.IsZero() {
			parts = append(parts, panels.Members(tm, v, reg.Members.Width, reg.Members.Height, membersOpts))
		}
	} else {
		topRow := lipgloss.JoinHorizontal(lipgloss.Top,
			panels.NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, panels.RenderOpts{}),
			panels.Members(tm, v, reg.Members.Width, reg.Members.Height, membersOpts),
			panels.Signals(tm, v, reg.Signals.Width, reg.Signals.Height, panels.RenderOpts{}),
		)
		parts = append(parts, topRow)
		parts = append(parts, panels.Queue(tm, v, reg.Queue.Width, reg.Queue.Height, queueOpts))
		parts = append(parts, panels.Chat(tm, v, reg.Chat.Width, reg.Chat.Height, chatOpts))
	}

	if m.mode == ModeHelp {
		parts = append(parts, modals.Help(tm, m.width, isHost))
		parts = append(parts, panels.StatusBar(tm, v, reg.Width, isHost))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if m.mode == ModeModalSeek {
		if m.errMsg != "" {
			parts = append(parts, tm.Error().Render(m.errMsg))
		}
		parts = append(parts, m.seekModal.View(tm, m.width))
		parts = append(parts, panels.StatusBar(tm, v, reg.Width, isHost))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if m.mode == ModeModalAdd {
		if m.errMsg != "" {
			parts = append(parts, tm.Error().Render(m.errMsg))
		}
		parts = append(parts, m.addModal.View(tm, m.width))
		parts = append(parts, panels.StatusBar(tm, v, reg.Width, isHost))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	if m.mode == ModeModalLeave {
		if m.errMsg != "" {
			parts = append(parts, tm.Error().Render(m.errMsg))
		}
		parts = append(parts, m.leaveModal.View(tm, m.width))
		parts = append(parts, panels.StatusBar(tm, v, reg.Width, isHost))
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	input := m.input.View()
	if m.errMsg != "" {
		input = tm.Error().Render(m.errMsg) + "\n" + input
	}
	parts = append(parts, input)
	parts = append(parts, panels.StatusBar(tm, v, reg.Width, isHost))

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m Model) renderOpts(panel FocusPanel) panels.RenderOpts {
	opts := panels.RenderOpts{
		Focused:     m.focus == panel && m.mode == ModeDashboard,
		QueueScroll: m.queueScroll,
	}
	if panel == FocusChat {
		opts.ChatScroll = m.chatScroll
	}
	if panel == FocusQueue {
		opts.QueueSelectedIdx = m.selectedQueueIdx
	}
	if panel == FocusMembers {
		opts.MembersScroll = m.membersScroll
	}
	return opts
}
