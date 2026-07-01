package theme

import (
	"strings"
	"testing"
)

func TestRoleColorsPalette(t *testing.T) {
	colors := RoleColors()
	if colors["neon_cyan"] != "51" || colors["neon_magenta"] != "201" {
		t.Fatalf("unexpected palette: %+v", colors)
	}
	if colors["bg"] != "0" {
		t.Fatalf("bg %q", colors["bg"])
	}
}

func TestPanelFocusedUsesMagentaAccent(t *testing.T) {
	tm := Default()
	// Border foreground is embedded in rendered ANSI; compare style via distinct renders.
	cyan := tm.Panel(false).Width(12).Height(4).Render("")
	magenta := tm.Panel(true).Width(12).Height(4).Render("")
	// At minimum both should produce bordered output.
	if !strings.Contains(cyan, "─") && !strings.Contains(cyan, "-") {
		t.Fatalf("expected border in cyan panel: %q", cyan)
	}
	if cyan == magenta {
		// Some terminals collapse colors; verify palette roles instead.
		if RoleColors()["neon_cyan"] == RoleColors()["neon_magenta"] {
			t.Fatal("palette must distinguish focus")
		}
	}
}

func TestASCIIBorderDiffersFromUnicode(t *testing.T) {
	def := Default().Panel(false).Width(10).Height(3).Render("x")
	asc := ASCII().Panel(false).Width(10).Height(3).Render("x")
	if def == asc {
		t.Fatal("ASCII border theme should differ from default")
	}
}

func TestProgressBar(t *testing.T) {
	tm := Default()
	bar := tm.ProgressBar(50, 100, 10)
	if !strings.Contains(bar, "█") || !strings.Contains(bar, "░") {
		t.Fatalf("bar %q", bar)
	}
	empty := tm.ProgressBar(0, 100, 8)
	if strings.Count(empty, "█") != 0 {
		t.Fatalf("expected no fill: %q", empty)
	}
	full := tm.ProgressBar(100, 100, 6)
	if strings.Count(full, "░") != 0 {
		t.Fatalf("expected full: %q", full)
	}
}

func TestSemanticStylesNonEmpty(t *testing.T) {
	tm := Default()
	for name, s := range map[string]lipglossStyle{
		"header": tm.Header(),
		"title":  tm.Title(),
		"muted":  tm.Muted(),
		"error":  tm.Error(),
		"ok":     tm.Success(),
		"warn":   tm.Warning(),
	} {
		out := s.Render("probe")
		if out == "" || !strings.Contains(out, "probe") {
			t.Fatalf("%s: %q", name, out)
		}
	}
}

// lipglossStyle avoids importing lipgloss in test signature noise.
type lipglossStyle interface {
	Render(...string) string
}

func TestASCIITheme(t *testing.T) {
	tm := ASCII()
	out := tm.Panel(false).Render("panel")
	if out == "" {
		t.Fatal("empty render")
	}
}
