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
	headerRows  = 3
	inputRows   = 1
	statusRows  = 1
	footerRows  = inputRows + statusRows
	minTopRowH  = 5
	minQueueH   = 4
	minChatH    = 3
	minNowPlayW = 28
	minCrewW    = 12
	minSignalsW = 14
)

// Compute returns HUD regions for the given terminal size.
// When width < MinWidth or height < MinHeight, Degraded is true and
// non-essential panels (signals, queue, chat) are collapsed (AC-011).
func Compute(width, height int, asciiBorders bool) Regions {
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
	if r.Degraded {
		r.computeDegraded()
		return r
	}
	r.computeFull()
	return r
}

func (r *Regions) computeDegraded() {
	w, h := r.Width, r.Height
	body := h - headerRows - footerRows
	if body < 1 {
		body = 1
	}
	r.Header = Region{Width: w, Height: minInt(headerRows, h)}
	// Priority: room + now playing + online (AC-011).
	r.NowPlaying = Region{Width: w, Height: maxInt(3, body-2)}
	r.Members = Region{Width: w, Height: minInt(2, body)}
	r.Signals = Region{}
	r.Queue = Region{}
	r.Chat = Region{}
}

func (r *Regions) computeFull() {
	w, h := r.Width, r.Height
	r.Header = Region{Width: w, Height: headerRows}

	body := h - headerRows - footerRows
	if body < minTopRowH+minQueueH+minChatH {
		body = minTopRowH + minQueueH + minChatH
	}

	topH := maxInt(minTopRowH, body*2/7)
	queueH := maxInt(minQueueH, body*2/7)
	chatH := body - topH - queueH
	if chatH < minChatH {
		chatH = minChatH
		queueH = body - topH - chatH
		if queueH < minQueueH {
			queueH = minQueueH
			topH = body - queueH - chatH
		}
	}

	crewW := maxInt(minCrewW, w/5)
	signalsW := maxInt(minSignalsW, w/4)
	nowW := w - crewW - signalsW
	if nowW < minNowPlayW {
		nowW = minNowPlayW
		remaining := w - nowW
		crewW = remaining / 2
		signalsW = remaining - crewW
	}

	r.NowPlaying = Region{Width: nowW, Height: topH}
	r.Members = Region{Width: crewW, Height: topH}
	r.Signals = Region{Width: signalsW, Height: topH}
	r.Queue = Region{Width: w, Height: queueH}
	r.Chat = Region{Width: w, Height: chatH}
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
