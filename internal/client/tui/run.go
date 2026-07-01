package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// Run starts the Bubble Tea program using the shared session store (AC-055).
// Config.Store is required; prefer Config.Actions from Runtime.Actions().
func Run(ctx context.Context, cfg Config) error {
	if cfg.Store == nil {
		return nil
	}
	m := NewModel(ctx, cfg)
	p := tea.NewProgram(&m, tea.WithContext(ctx), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
