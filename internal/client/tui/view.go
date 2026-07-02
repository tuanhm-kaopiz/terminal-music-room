package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/terminal-music-room/music-room/internal/client/state"
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
	reg := m.layoutRegions()
	v := m.view
	isHost := IsHost(v)

	header := panels.Header(tm, v, reg.Header.Width, reg.Header.Height)
	footer := m.renderFooter(tm, v, reg, isHost)

	switch m.mode {
	case ModeHelp:
		body := m.placeModal(reg, modals.Help(tm, m.width, isHost))
		return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
	case ModeModalAdd:
		modal := m.addModal.View(tm, m.width)
		if m.errMsg != "" {
			modal = tm.Error().Render(m.errMsg) + "\n" + modal
		}
		body := m.placeModal(reg, modal)
		return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
	case ModeModalSeek:
		modal := m.seekModal.View(tm, m.width)
		if m.errMsg != "" {
			modal = tm.Error().Render(m.errMsg) + "\n" + modal
		}
		body := m.placeModal(reg, modal)
		return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
	case ModeModalLeave:
		modal := m.leaveModal.View(tm, m.width)
		if m.errMsg != "" {
			modal = tm.Error().Render(m.errMsg) + "\n" + modal
		}
		body := m.placeModal(reg, modal)
		return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
	case ModeModalPassword:
		modal := m.passwordModal.View(tm, m.width)
		if m.errMsg != "" {
			modal = tm.Error().Render(m.errMsg) + "\n" + modal
		}
		body := m.placeModal(reg, modal)
		return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
	}

	body := m.renderBody(tm, v, reg)
	return clipView(joinHeaderBodyFooter(header, body, footer), m.width, m.height)
}

func (m Model) renderBody(tm theme.Theme, v state.View, reg layout.Regions) string {
	queueOpts := m.renderOpts(FocusQueue)
	chatOpts := m.renderOpts(FocusChat)
	membersOpts := m.renderOpts(FocusMembers)

	if reg.Degraded {
		var parts []string
		parts = append(parts, tm.Warning().Render(
			"⚠ terminal < 80×24 — degraded HUD (room + now playing prioritized)",
		))
		parts = append(parts, panels.NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, panels.RenderOpts{}))
		if !reg.Members.IsZero() {
			parts = append(parts, panels.Members(tm, v, reg.Members.Width, reg.Members.Height, membersOpts))
		}
		return lipgloss.JoinVertical(lipgloss.Left, parts...)
	}

	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		panels.NowPlaying(tm, v, reg.NowPlaying.Width, reg.NowPlaying.Height, panels.RenderOpts{}),
		panels.Members(tm, v, reg.Members.Width, reg.Members.Height, membersOpts),
	)
	row1 = lipgloss.NewStyle().Width(reg.Width).Render(row1)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		panels.Queue(tm, v, reg.Queue.Width, reg.Queue.Height, queueOpts),
		panels.Signals(tm, v, reg.Signals.Width, reg.Signals.Height, panels.RenderOpts{}),
	)
	row2 = lipgloss.NewStyle().Width(reg.Width).Render(row2)

	row3 := panels.Chat(tm, v, reg.Chat.Width, reg.Chat.Height, chatOpts)

	return lipgloss.JoinVertical(lipgloss.Left, row1, row2, row3)
}

func (m Model) renderFooter(tm theme.Theme, v state.View, reg layout.Regions, isHost bool) string {
	input := m.input.View()
	if m.errMsg != "" && m.mode == ModeDashboard {
		input = tm.Error().Render(m.errMsg) + "\n" + input
	}
	return lipgloss.JoinVertical(lipgloss.Left, input, panels.StatusBar(tm, v, reg.Width, isHost))
}

func (m Model) placeModal(reg layout.Regions, modal string) string {
	bodyH := reg.BodyHeight()
	if bodyH < 1 {
		bodyH = 1
	}
	return lipgloss.Place(reg.Width, bodyH, lipgloss.Center, lipgloss.Center, modal)
}

func joinHeaderBodyFooter(header, body, footer string) string {
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) layoutRegions() layout.Regions {
	extra := 0
	if m.errMsg != "" && m.mode == ModeDashboard {
		extra++
	}
	if m.width < layout.MinWidth || m.height < layout.MinHeight {
		extra++
	}
	if m.view.LastErr != nil && m.view.LastErr.Message != "" && m.mode == ModeDashboard {
		extra++
	}
	return layout.Compute(m.width, m.height, false, extra)
}

func clipView(content string, width, height int) string {
	if width < 1 || height < 1 {
		return content
	}
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	return lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, strings.Join(lines, "\n"))
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
		opts.MembersSelectedIdx = m.selectedMemberIdx
	}
	return opts
}
