package layout

// Minimum terminal size for full HUD (AC-009, NFR-001).
const (
	MinWidth  = 80
	MinHeight = 24
)

// Region is a rectangular panel slot in the terminal grid.
type Region struct {
	Width  int
	Height int
}

// IsZero reports whether the region is hidden (degraded layout).
func (r Region) IsZero() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Regions describes computed HUD panel geometry.
type Regions struct {
	Width        int
	Height       int
	Degraded     bool
	AsciiBorders bool

	Header     Region
	NowPlaying Region
	Members    Region
	Signals    Region
	Queue      Region
	Chat       Region
	Input      Region
	StatusBar  Region
}

const (
	headerRows  = 1 // plain text strip (no bordered panel)
	inputRows   = 1
	statusRows  = 1
	footerRows  = inputRows + statusRows
	panelBorder = 2 // lipgloss border adds 2 cells per panel edge
	bodyRows    = 3 // row1 playing+crew, row2 queue+signals, row3 chat
	minRow1H    = 5
	minRow2H    = 4
	minRow3H    = 3
	minNowPlayW = 28
	minCrewW    = 14
	minSignalsW = 12
	minQueueW   = 24
)

// Compute returns HUD regions for the given terminal size.
// extraFooterRows subtracts from body height (error toast, degraded banner).
func Compute(width, height int, asciiBorders bool, extraFooterRows ...int) Regions {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	r := Regions{
		Width:        width,
		Height:       height,
		AsciiBorders: asciiBorders,
		Degraded:     width < MinWidth || height < MinHeight,
		Input:        Region{Width: width, Height: minInt(inputRows, height)},
		StatusBar:    Region{Width: width, Height: minInt(statusRows, height)},
	}
	footerExtra := 0
	if len(extraFooterRows) > 0 && extraFooterRows[0] > 0 {
		footerExtra = extraFooterRows[0]
	}
	layoutH := height
	if footerExtra > 0 {
		layoutH = maxInt(1, height-footerExtra)
	}
	r.Height = layoutH
	if r.Degraded {
		r.computeDegraded()
		return r
	}
	r.computeFull()
	return r
}

// BodyHeight is the vertical space for the three content rows (excludes header/footer).
func (r Regions) BodyHeight() int {
	return r.NowPlaying.Height + r.Queue.Height + r.Chat.Height
}

func (r *Regions) computeDegraded() {
	w, h := r.Width, r.Height
	body := h - headerRows - footerRows
	if body < 1 {
		body = 1
	}
	r.Header = Region{Width: w, Height: minInt(headerRows, h)}
	r1 := maxInt(3, body*2/3)
	r2 := body - r1
	r.NowPlaying = Region{Width: w, Height: r1}
	r.Members = Region{Width: w, Height: minInt(2, r2)}
	r.Signals = Region{}
	r.Queue = Region{}
	r.Chat = Region{}
}

func (r *Regions) computeFull() {
	w, h := r.Width, r.Height
	r.Header = Region{Width: w, Height: headerRows}

	body := h - headerRows - footerRows
	if body < minRow1H+minRow2H+minRow3H {
		body = minRow1H + minRow2H + minRow3H
	}

	row1H := maxInt(minRow1H, body/bodyRows)
	row2H := maxInt(minRow2H, body/bodyRows)
	row3H := body - row1H - row2H
	if row3H < minRow3H {
		row3H = minRow3H
		row2H = maxInt(minRow2H, (body-row3H)/2)
		row1H = body - row2H - row3H
	}

	// Row 1: now playing + crew (two abreast; borders add panelBorder each).
	pairW := w - 2*panelBorder
	crewW := maxInt(minCrewW, pairW/3)
	nowW := pairW - crewW
	if nowW < minNowPlayW {
		nowW = minNowPlayW
		crewW = pairW - nowW
	}

	// Row 2: queue + signals (two abreast).
	signalsW := maxInt(minSignalsW, pairW*2/5)
	queueW := pairW - signalsW
	if queueW < minQueueW {
		queueW = minQueueW
		signalsW = pairW - queueW
	}

	r.NowPlaying = Region{Width: nowW, Height: row1H}
	r.Members = Region{Width: crewW, Height: row1H}
	r.Signals = Region{Width: signalsW, Height: row2H}
	r.Queue = Region{Width: queueW, Height: row2H}
	r.Chat = Region{Width: w, Height: row3H}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
