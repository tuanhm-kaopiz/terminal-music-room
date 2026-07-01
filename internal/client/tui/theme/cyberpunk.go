package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ANSI palette indices (16-color baseline, ADR-004).
const (
	colorBG          = "0"
	colorSurface     = "236"
	colorText        = "15"
	colorMuted       = "245"
	colorNeonCyan    = "51"
	colorNeonMagenta = "201"
	colorNeonYellow  = "226"
	colorError       = "196"
	colorSuccess     = "46"
	colorBorder      = "39"
)

// Theme is the fixed cyberpunk HUD palette for TUI v2.
type Theme struct {
	bg          lipgloss.Color
	surface     lipgloss.Color
	text        lipgloss.Color
	muted       lipgloss.Color
	neonCyan    lipgloss.Color
	neonMagenta lipgloss.Color
	neonYellow  lipgloss.Color
	err         lipgloss.Color
	success     lipgloss.Color
	border      lipgloss.Color
	asciiBorder bool
}

// Default returns the standard cyberpunk theme with Unicode box drawing.
func Default() Theme {
	return newTheme(false)
}

// ASCII returns the theme using ASCII +/- borders.
func ASCII() Theme {
	return newTheme(true)
}

func newTheme(ascii bool) Theme {
	return Theme{
		bg:          lipgloss.Color(colorBG),
		surface:     lipgloss.Color(colorSurface),
		text:        lipgloss.Color(colorText),
		muted:       lipgloss.Color(colorMuted),
		neonCyan:    lipgloss.Color(colorNeonCyan),
		neonMagenta: lipgloss.Color(colorNeonMagenta),
		neonYellow:  lipgloss.Color(colorNeonYellow),
		err:         lipgloss.Color(colorError),
		success:     lipgloss.Color(colorSuccess),
		border:      lipgloss.Color(colorBorder),
		asciiBorder: ascii,
	}
}

// Panel returns a bordered panel style; focused panels use magenta accent.
func (t Theme) Panel(focused bool) lipgloss.Style {
	borderFG := t.neonCyan
	if focused {
		borderFG = t.neonMagenta
	}
	return lipgloss.NewStyle().
		Border(t.borderStyle()).
		BorderForeground(borderFG).
		Background(t.surface).
		Foreground(t.text).
		Padding(0, 1)
}

// Header styles the top HUD strip (room slug, online count, conn).
func (t Theme) Header() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.neonCyan).
		Background(t.bg).
		Bold(true)
}

// Title styles panel headings.
func (t Theme) Title() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(t.neonCyan).
		Bold(true)
}

// Muted styles secondary metadata and help text.
func (t Theme) Muted() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.muted)
}

// Error styles error toasts and validation messages.
func (t Theme) Error() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.err).Bold(true)
}

// Success styles connected / OK badges.
func (t Theme) Success() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.success)
}

// Warning styles vote progress and cautions.
func (t Theme) Warning() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.neonYellow)
}

// HostMarker styles the host indicator in member lists.
func (t Theme) HostMarker() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(t.neonMagenta).Bold(true)
}

// ProgressBar renders an ASCII fill bar with neon accents.
func (t Theme) ProgressBar(filled, total, width int) string {
	if width < 1 {
		width = 1
	}
	if total <= 0 {
		total = 1
	}
	if filled < 0 {
		filled = 0
	}
	if filled > total {
		filled = total
	}
	barW := width
	if barW > 40 {
		barW = 40
	}
	nFilled := filled * barW / total
	if nFilled > barW {
		nFilled = barW
	}
	filledStr := strings.Repeat("█", nFilled)
	emptyStr := strings.Repeat("░", barW-nFilled)
	return lipgloss.NewStyle().Foreground(t.neonCyan).Render(filledStr) +
		lipgloss.NewStyle().Foreground(t.muted).Render(emptyStr)
}

func (t Theme) borderStyle() lipgloss.Border {
	if t.asciiBorder {
		return lipgloss.ASCIIBorder()
	}
	return lipgloss.NormalBorder()
}

// RoleColors exposes palette indices for tests and documentation.
func RoleColors() map[string]string {
	return map[string]string{
		"bg":           colorBG,
		"surface":      colorSurface,
		"text":         colorText,
		"muted":        colorMuted,
		"neon_cyan":    colorNeonCyan,
		"neon_magenta": colorNeonMagenta,
		"neon_yellow":  colorNeonYellow,
		"error":        colorError,
		"success":      colorSuccess,
		"border":       colorBorder,
	}
}
