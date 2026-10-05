package terma

import (
	"os"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// InlineOptions configures RunInline.
type InlineOptions struct {
	// MaxHeight caps the live region's rows. 0 means the terminal height.
	// Root is laid out with Loose(width, maxHeight): Auto heights follow
	// content, Flex fills MaxHeight.
	MaxHeight int
	// Mouse enables mouse reporting. Off by default inline, because it takes
	// the wheel away from the terminal's own scrollback.
	Mouse bool
	// OnExit decides what happens to the live region when the app exits.
	OnExit InlineExit
}

// InlineExit is what happens to an inline app's live region when it exits.
type InlineExit int

const (
	// InlineExitKeep leaves the final frame in the scrollback, with the
	// cursor on the line below it.
	InlineExitKeep InlineExit = iota
	// InlineExitClear erases the live region, leaving the cursor where the
	// region started.
	InlineExitClear
)

// RunInline runs root in the normal screen below the cursor, like a CLI
// prompt, and blocks until it exits. The app lives in a region that starts on
// the cursor's row and is as tall as its content (root's height and any
// overlay reaching below it), up to opts.MaxHeight. The terminal's scrollback
// above it stays as it was; PrintAbove adds to it.
//
// Sixel images position themselves absolutely, so they don't work inline.
func RunInline(root Widget, opts InlineOptions) error {
	return run(root, &inlineMode{opts: opts})
}

// inlineMode draws the app in a live region of the normal screen. The
// renderer paints into frame as if it were a window maxHeight rows tall;
// the region shows its top rows, as many as the content needs.
type inlineMode struct {
	opts       InlineOptions
	frame      uv.ScreenBuffer
	width      int
	screenRows int
	maxHeight  int
	// height is the region's rows on screen, as last presented.
	height int
	// drawn is set once a frame has been displayed.
	drawn  bool
	origin inlineOrigin
}

// inlineTerminal is the part of uv.Terminal the live region is drawn with.
type inlineTerminal interface {
	CellBuffer
	Bounds() uv.Rectangle
	Resize(width, height int) error
	ClearArea(area uv.Rectangle)
	Position() (x, y int)
	SetPosition(x, y int)
	WriteString(s string) (int, error)
	Erase()
}

// inlineOrigin tracks the screen row of the region's top, which mouse
// reports are relative to. The terminal says where the cursor is (DSR);
// after that, each frame's move is worked out from how many rows were
// printed above the region and how tall it is.
type inlineOrigin struct {
	row   int
	known bool
	// ask is set when the next frame should ask where the cursor is.
	ask bool
	// outstanding counts the questions not yet answered; only the answer to
	// the latest one is used.
	outstanding int
	// queryRow is the cursor's row in the region when the latest question
	// was asked, and since the frames presented after it.
	queryRow int
	since    []inlineMove
}

// inlineMove is a presented frame: the rows printed above the region, and
// the region's height.
type inlineMove struct{ printed, height int }

func inlineMaxHeight(maxHeight, screenRows int) int {
	if maxHeight <= 0 || maxHeight > screenRows {
		maxHeight = screenRows
	}
	return max(maxHeight, 1)
}

// regionHeight is the region's height for content that reaches content rows
// down the canvas.
func (m *inlineMode) regionHeight(content int) int {
	return min(max(content, 1), m.maxHeight)
}

// moved records a frame that printed rows above the region and drew it
// height rows tall. Printed rows push the region down; a region reaching
// past the bottom of the screen scrolls it up.
func (m *inlineMode) moved(printed, height int) {
	switch {
	case m.origin.known:
		m.origin.row = min(m.origin.row+printed, m.screenRows-height)
	case m.origin.outstanding > 0:
		m.origin.since = append(m.origin.since, inlineMove{printed, height})
	}
}

// answered takes the cursor's screen row from a cursor position report.
func (m *inlineMode) answered(row int) {
	m.origin.outstanding--
	if m.origin.outstanding > 0 {
		return
	}
	m.origin.row = row - m.origin.queryRow
	m.origin.known = true
	for _, move := range m.origin.since {
		m.moved(move.printed, move.height)
	}
	m.origin.since = nil
}

// forget marks the region's place on screen unknown until the terminal says
// again.
func (m *inlineMode) forget() {
	m.origin.known = false
	m.origin.ask = m.opts.Mouse
	m.origin.since = nil
}

func (m *inlineMode) enter(t *uv.Terminal) {
	// Start has queued entering the alternate screen, saving the cursor;
	// leaving it restores the cursor to the row the region starts on. Draw
	// an empty region now so output printed before the first frame lands in
	// the normal screen.
	t.ExitAltScreen()
	_ = t.Resize(t.Bounds().Dx(), 1)
	m.drawn = t.Display() == nil
	m.forget()
	openPrintAbove()
}

func (m *inlineMode) mouse() bool { return m.opts.Mouse }

func (m *inlineMode) canvas(_ *uv.Terminal, width, height int) (CellBuffer, int, int) {
	m.width, m.screenRows = width, height
	m.maxHeight = inlineMaxHeight(m.opts.MaxHeight, height)
	m.frame = uv.NewScreenBuffer(width, m.maxHeight)
	return m.frame, width, m.maxHeight
}

func (m *inlineMode) present(t *uv.Terminal, r *Renderer, display func() error) error {
	printed := m.printAbove(t, takePrintAbove(false))
	m.show(t, m.regionHeight(r.contentHeight()), 0)
	if err := display(); err != nil {
		return err
	}
	m.drawn = true
	m.moved(printed, m.height)
	if m.origin.ask {
		m.origin.ask = false
		m.origin.outstanding++
		_, m.origin.queryRow = t.Position()
		m.origin.since = nil
		_, _ = t.WriteString(ansi.RequestCursorPosition)
		_ = t.Flush()
	}
	return nil
}

// show sizes the terminal's buffer to height rows of the canvas followed by
// blank rows. Only cells that changed are marked for drawing, so resizing
// the region redraws nothing else, and no widget is built or painted again.
func (m *inlineMode) show(t inlineTerminal, height, blank int) {
	b := t.Bounds()
	resized := b.Dx() != m.width || b.Dy() != height+blank
	if resized {
		_ = t.Resize(m.width, height+blank)
	}
	copyCells(t, m.frame, m.width, height)
	t.ClearArea(uv.Rect(0, height, m.width, blank))
	if resized {
		// When the region grows and its new last row is blank, uv erases the
		// row above it, but redraws it only if marked changed. Resizing again
		// (to the same size) marks every row, so it is redrawn.
		_ = t.Resize(m.width, height+blank)
	}
	m.height = height
}

// printAbove writes items where the region is, so they scroll up into the
// scrollback like any other output, and has the region drawn again below
// them. It returns how many rows were written.
func (m *inlineMode) printAbove(t inlineTerminal, items []printAboveItem) int {
	var out strings.Builder
	rows := 0
	for _, item := range items {
		for _, line := range item.lines(m.width, m.screenRows) {
			out.WriteString(ansi.Truncate(line, m.width, ""))
			out.WriteString(ansi.ResetStyle + "\r\n")
			rows++
		}
	}
	if rows == 0 {
		return 0
	}
	// Writing the lines ourselves rather than with PrependString: it counts a
	// line exactly as wide as the terminal as taking no rows, and goes wrong
	// when the lines and the region together are taller than the screen.
	_, y := t.Position()
	start := "\r"
	if y > 0 {
		start += ansi.CursorUp(y)
	}
	_, _ = t.WriteString(start + ansi.EraseScreenBelow + out.String())
	// The cursor is now on the region's new top row, which is blank.
	t.SetPosition(-1, -1)
	t.Erase()
	return rows
}

func (m *inlineMode) resize(t *uv.Terminal, r *Renderer, width, height int) {
	m.width, m.screenRows = width, height
	m.maxHeight = inlineMaxHeight(m.opts.MaxHeight, height)
	r.Resize(width, m.maxHeight)
	// Erasing redraws the region from where the terminal thinks its top is.
	// Terminals that rewrap lines when the width changes can move the old
	// region, leaving some of it behind.
	t.Erase()
	m.forget()
}

func (m *inlineMode) pause(t *uv.Terminal) error {
	// An empty region erases the app, leaving the cursor where it started
	// for the program being handed the terminal.
	m.show(t, 0, 1)
	_ = t.Display()
	return t.Pause()
}

func (m *inlineMode) resume(t *uv.Terminal) error {
	// Resume leaves the alternate screen again, which moves the cursor back
	// to where it was saved. Some terminals (xterm) restore what DECSC saved,
	// others (tmux) where the cursor was when the app started. Saving it
	// before and restoring it after puts it back where the program left it,
	// for the region to start there.
	_, _ = os.Stdout.WriteString(ansi.SaveCursor)
	err := t.Resume()
	_, _ = os.Stdout.WriteString(ansi.RestoreCursor)
	m.forget()
	return err
}

func (m *inlineMode) leave(t *uv.Terminal) {
	pending := takePrintAbove(true)
	if !m.drawn {
		return
	}
	m.printAbove(t, pending)
	t.HideCursor()
	// Shutdown erases the last row of the region and leaves the cursor
	// there, so a blank row is added for it.
	switch m.opts.OnExit {
	case InlineExitClear:
		m.show(t, 0, 1)
	default:
		m.show(t, m.height, 1)
	}
	_ = t.Display()
}

// restored leaves the alternate screen Start entered if no frame was drawn
// to leave it. After a frame that would restore the cursor saved at start.
func (m *inlineMode) restored() {
	if !m.drawn {
		_, _ = os.Stdout.WriteString(ansi.ResetModeAltScreenSaveCursor)
	}
}

// locate makes mouse rows relative to the region's top. Mouse events are
// dropped until the terminal has said where the region is.
func (m *inlineMode) locate(ev uv.Event) (uv.Event, bool) {
	var mouse uv.Mouse
	switch e := ev.(type) {
	case uv.MouseClickEvent:
		mouse = uv.Mouse(e)
	case uv.MouseReleaseEvent:
		mouse = uv.Mouse(e)
	case uv.MouseMotionEvent:
		mouse = uv.Mouse(e)
	case uv.MouseWheelEvent:
		mouse = uv.Mouse(e)
	default:
		return ev, true
	}
	if !m.origin.known {
		return nil, false
	}
	mouse.Y -= m.origin.row
	switch ev.(type) {
	case uv.MouseClickEvent:
		return uv.MouseClickEvent(mouse), true
	case uv.MouseReleaseEvent:
		return uv.MouseReleaseEvent(mouse), true
	case uv.MouseMotionEvent:
		return uv.MouseMotionEvent(mouse), true
	default:
		return uv.MouseWheelEvent(mouse), true
	}
}

func (m *inlineMode) handle(ev uv.Event) bool {
	report, ok := ev.(uv.CursorPositionEvent)
	if !ok || m.origin.outstanding == 0 {
		return false
	}
	m.answered(report.Y)
	return true
}

// contentHeight is how far down the last frame painted: the root's border
// box, or an overlay reaching below it.
func (r *Renderer) contentHeight() int {
	height := r.lastLayoutHeight
	for _, float := range r.retainedFloats {
		height = max(height, float.paintBounds.Y+float.paintBounds.Height)
	}
	return height
}
