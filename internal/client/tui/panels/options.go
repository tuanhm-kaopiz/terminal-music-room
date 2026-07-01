package panels

// RenderOpts controls panel focus and scroll offsets.
type RenderOpts struct {
	Focused          bool
	QueueScroll      int
	QueueSelectedIdx int
	ChatScroll       int
	MembersScroll    int
}
