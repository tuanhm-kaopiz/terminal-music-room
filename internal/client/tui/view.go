package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/terminal-music-room/music-room/internal/client/state"
	"github.com/terminal-music-room/music-room/internal/protocol"
)

var (
	panelStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	mutedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

// View renders the full TUI layout (AC-053).
func (m Model) View() string {
	if m.quit {
		return "Leaving room…\n"
	}
	if m.width == 0 || m.height == 0 {
		return "Loading…\n"
	}

	headerH := 3
	footerH := 3
	bodyH := max(6, m.height-headerH-footerH)
	topH := max(4, bodyH/3)
	midH := max(3, bodyH/3)
	chatH := max(3, bodyH-topH-midH)

	leftW := max(20, m.width*3/5)
	rightW := max(10, m.width-leftW)

	header := renderHeader(m.view, m.width)
	topRow := lipgloss.JoinHorizontal(lipgloss.Top,
		renderNowPlaying(m.view, leftW, topH),
		renderMembers(m.view, rightW, topH),
	)
	queue := renderQueue(m.view, m.width, midH)
	chat := renderChat(m.view, m.width, chatH)
	footer := m.input.View()
	if m.errMsg != "" {
		footer = errStyle.Render(truncate(m.errMsg, m.width-2)) + "\n" + footer
	}
	help := mutedStyle.Render(truncate("q quit · enter send chat · "+string(m.view.Status), m.width))

	return lipgloss.JoinVertical(lipgloss.Left, header, topRow, queue, chat, footer, help)
}

func renderHeader(v state.View, width int) string {
	room := v.Room.Slug
	if room == "" {
		room = "(no room)"
	}
	title := fmt.Sprintf("Room: %s", room)
	if v.Room.HostID != "" {
		title += fmt.Sprintf(" · %d online", len(v.Room.Members))
	}
	content := titleStyle.Render(truncate(title, width-4))
	return panelStyle.Width(width).Height(3).Render(content)
}

func renderNowPlaying(v state.View, width, height int) string {
	innerW := max(1, width-4)
	innerH := max(1, height-2)
	lines := make([]string, 0, innerH)
	lines = append(lines, titleStyle.Render("Now Playing"))
	pb := v.Room.Playback
	title := "(nothing queued)"
	if pb.Track != nil && pb.Track.Title != "" {
		title = pb.Track.Title
	}
	lines = append(lines, truncate(title, innerW))
	status := string(pb.Status)
	pos := formatMs(pb.PositionMs)
	dur := formatMs(pb.DurationMs)
	if dur == "0:00" && pb.Track != nil && pb.Track.DurationMs > 0 {
		dur = formatMs(pb.Track.DurationMs)
	}
	lines = append(lines, truncate(fmt.Sprintf("%s  %s / %s", status, pos, dur), innerW))
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines[:innerH], "\n"))
}

func renderMembers(v state.View, width, height int) string {
	innerW := max(1, width-4)
	innerH := max(1, height-2)
	lines := []string{titleStyle.Render("Members")}
	for _, m := range v.Room.Members {
		prefix := " "
		name := m.DisplayName
		if name == "" {
			name = m.Nickname
		}
		if m.IsHost {
			prefix = "*"
		}
		lines = append(lines, truncate(prefix+" "+name, innerW))
	}
	if len(v.Room.Members) == 0 {
		lines = append(lines, mutedStyle.Render("(empty)"))
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines[:innerH], "\n"))
}

func renderQueue(v state.View, width, height int) string {
	innerW := max(1, width-4)
	innerH := max(1, height-2)
	lines := []string{titleStyle.Render("Queue")}
	if len(v.Room.Queue) == 0 {
		lines = append(lines, mutedStyle.Render("(empty)"))
	} else {
		for i, item := range v.Room.Queue {
			if len(lines) >= innerH {
				break
			}
			line := fmt.Sprintf("%d. %s", i+1, item.Title)
			if item.AddedBy != "" {
				line += " · " + item.AddedBy
			}
			lines = append(lines, truncate(line, innerW))
		}
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines[:innerH], "\n"))
}

func renderChat(v state.View, width, height int) string {
	innerW := max(1, width-4)
	innerH := max(1, height-2)
	lines := []string{titleStyle.Render("Chat")}
	msgs := v.Room.Chat
	start := 0
	if len(msgs) > innerH-1 {
		start = len(msgs) - (innerH - 1)
	}
	for _, msg := range msgs[start:] {
		lines = append(lines, truncate(formatChat(msg), innerW))
	}
	if len(v.Room.Chat) == 0 {
		lines = append(lines, mutedStyle.Render("(no messages yet)"))
	}
	for len(lines) < innerH {
		lines = append(lines, "")
	}
	return panelStyle.Width(width).Height(height).Render(strings.Join(lines[:innerH], "\n"))
}

func formatChat(msg protocol.ChatMessage) string {
	if msg.Kind == protocol.ChatKindSystem {
		return "[sys] " + msg.Body
	}
	author := msg.Author
	if author == "" {
		author = "?"
	}
	return author + ": " + msg.Body
}

func formatMs(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	sec := ms / 1000
	min := sec / 60
	sec %= 60
	return fmt.Sprintf("%d:%02d", min, sec)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= max {
		return s
	}
	if max <= 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := runewidth.RuneWidth(r)
		if w+rw > max-1 {
			break
		}
		b.WriteRune(r)
		w += rw
	}
	b.WriteRune('…')
	return b.String()
}
