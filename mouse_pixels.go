package terma

import (
	"math"
	"strconv"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// pixelPointer locates the mouse to a fraction of a cell, as Textual does for
// smooth scrolling. Terminals that support SGR-Pixels mouse mode (1016), such
// as kitty and Ghostty, can report the pointer in pixels; this
// divides those by the size of a cell to get the cell plus the position
// within it. Dragging a scrollbar thumb then follows the pointer exactly.
//
// The cell size comes from the terminal's own report (CSI 16 t) where it
// gives one. The window's size in pixels divided by its columns and rows is
// only a stand-in: some terminals (Ghostty, for one) count their window
// padding in that size, which makes the cells too big and puts the pointer
// further off the further it is from the top-left. So it stands in only when
// it divides evenly (padding rarely does), or, unevenly, for a report the
// window has outgrown.
//
// Changing the font size changes the cells' size, and is seen only as a
// resize. Resizing the window leaves the pixels not covered by cells (the
// padding, and less than a cell besides) within a cell of what they were, so
// a resize that moves them further than that asks the terminal again. One
// question is out at a time, and none are asked of a terminal that hasn't
// answered, or when the window's size in pixels explains the resize.
//
// The mode is switched on only once the terminal has reported it recognises
// it (in reply to a DECRQM query) and a cell size is known.
// Set TERMA_DISABLE_PIXEL_MOUSE to keep cell-based reporting.
type pixelPointer struct {
	disabled  bool // Switched off by TERMA_DISABLE_PIXEL_MOUSE.
	supported bool // The terminal recognises mode 1016.
	enabled   bool // Mode 1016 has been switched on.

	// The window: in cells, and in pixels where the terminal says (0 if not).
	window windowGeometry
	// readWindow reads the window's current size, if it can. The size in
	// pixels arrives in an event of its own after the size in cells (and only
	// if it changed), so reading both at once keeps them in step. Nil in tests.
	readWindow func() (windowGeometry, bool)

	// The cell size in pixels, as the terminal last reported it.
	cellWidth, cellHeight int
	// The window's pixels not covered by cells when the cell size was last
	// confirmed, if the window's size in pixels was known.
	slackX, slackY int
	haveSlack      bool

	answered bool           // The terminal has answered a cell size query.
	awaiting bool           // A cell size query is unanswered.
	askedFor windowGeometry // The window when the unanswered query was sent.
}

// windowGeometry is the window's size in cells and in pixels.
type windowGeometry struct {
	cols, rows              int
	pixelWidth, pixelHeight int
}

// requestCellSize asks the terminal for the size of a cell in pixels.
var requestCellSize = ansi.WindowOp(ansi.RequestCellSizeWinOp)

// cellSizeReplyTimeout bounds the wait for a cell size reply before the
// terminal is handed over (see replyDue).
const cellSizeReplyTimeout = 200 * time.Millisecond

func newPixelPointer() *pixelPointer {
	return &pixelPointer{disabled: boolEnv("TERMA_DISABLE_PIXEL_MOUSE")}
}

// query returns the sequences asking whether the terminal supports mode 1016
// and how big its cells are.
func (p *pixelPointer) query() string {
	if p.disabled {
		return ""
	}
	p.awaiting = true
	return ansi.RequestModeMouseExtSgrPixel + requestCellSize
}

// handle records a mode, cell size or window size report, returning the
// sequences to write: a fresh cell size query when the window has changed in
// a way the last reported cell size can't explain, and the one that switches
// pixel reporting on once everything it needs is known.
func (p *pixelPointer) handle(event uv.Event) string {
	var seq string
	switch ev := event.(type) {
	case uv.WindowSizeEvent:
		window := p.window
		window.cols, window.rows = ev.Width, ev.Height
		seq = p.resized(window)
	case uv.WindowPixelSizeEvent:
		window := p.window
		window.pixelWidth, window.pixelHeight = ev.Width, ev.Height
		seq = p.resized(window)
	case uv.CellSizeEvent:
		if p.recordCellSize(ev) && p.askedFor.cols > 0 && p.askedFor != p.window {
			// The window changed while the question was out; the answer may
			// be for the font size it had then.
			seq = p.ask()
		}
	case uv.ModeReportEvent:
		if ev.Mode != ansi.ModeMouseExtSgrPixel {
			return ""
		}
		p.supported = !ev.Value.IsNotRecognized() && !ev.Value.IsPermanentlyReset()
	default:
		return ""
	}
	return seq + p.enable()
}

// resized records the window's new size, returning a cell size query if the
// cells may have changed size.
func (p *pixelPointer) resized(window windowGeometry) string {
	p.setWindow(window)
	if p.disabled || !p.answered || p.awaiting {
		// Nothing to ask, or the answer on its way is checked against the
		// window as it is by then.
		return ""
	}
	if p.window.pixelWidth > 0 {
		if !p.haveSlack {
			// The first size in pixels since the cell size was reported.
			p.noteSlack()
			return ""
		}
		if p.reportFits() {
			return ""
		}
	}
	return p.ask()
}

// setWindow records the window's size, read afresh where that is possible.
func (p *pixelPointer) setWindow(window windowGeometry) {
	if p.readWindow != nil {
		if current, ok := p.readWindow(); ok {
			window = current
		}
	}
	if window.pixelWidth <= 0 || window.pixelHeight <= 0 {
		// Keep the last size in pixels known, rather than none.
		window.pixelWidth, window.pixelHeight = p.window.pixelWidth, p.window.pixelHeight
	}
	p.window = window
}

// ask returns a cell size query, noting that it is out.
func (p *pixelPointer) ask() string {
	p.awaiting = true
	p.askedFor = p.window
	return requestCellSize
}

// recordCellSize records the reply to a cell size query, reporting whether
// it gave a size. Terminals may reply 0 (while minimised, say); the last size
// given is kept then.
func (p *pixelPointer) recordCellSize(ev uv.CellSizeEvent) bool {
	p.awaiting = false
	if ev.Width <= 0 || ev.Height <= 0 {
		return false
	}
	p.answered = true
	p.cellWidth, p.cellHeight = ev.Width, ev.Height
	p.noteSlack()
	return true
}

// noteSlack records the window's pixels not covered by cells, the measure
// later resizes are checked against.
func (p *pixelPointer) noteSlack() {
	w := p.window
	p.haveSlack = p.cellWidth > 0 && w.cols > 0 && w.pixelWidth > 0
	if p.haveSlack {
		p.slackX = w.pixelWidth - w.cols*p.cellWidth
		p.slackY = w.pixelHeight - w.rows*p.cellHeight
	}
}

// reportFits reports whether the reported cell size still explains the
// window: a resize leaves the pixels not covered by cells within a cell of
// where they were, while a new font size moves them by a little per cell.
// Without the window's size in pixels there is nothing to check against.
func (p *pixelPointer) reportFits() bool {
	if !p.haveSlack || p.window.pixelWidth <= 0 {
		return true
	}
	slackX := p.window.pixelWidth - p.window.cols*p.cellWidth
	slackY := p.window.pixelHeight - p.window.rows*p.cellHeight
	return absInt(slackX-p.slackX) < p.cellWidth && absInt(slackY-p.slackY) < p.cellHeight
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// resume is called when the terminal is taken back after being handed to
// another program, returning the sequences to write. Pixel reporting was
// switched off meanwhile, and the font size may have changed, so the cell
// size is asked for again before it is switched back on.
func (p *pixelPointer) resume() string {
	p.enabled = false
	p.setWindow(p.window)
	if p.disabled || !p.answered {
		return p.enable()
	}
	p.cellWidth, p.cellHeight = 0, 0
	p.haveSlack = false
	return p.ask() + p.enable()
}

// replyDue reports whether a cell size reply is on its way from a terminal
// known to answer. Left unread when the terminal is handed to another program
// (or back to the shell), it would reach that as typing.
func (p *pixelPointer) replyDue() bool {
	return p.answered && p.awaiting
}

// enable returns the sequence that switches pixel reporting on, once the
// terminal supports it and a cell size is known.
func (p *pixelPointer) enable() string {
	if p.enabled || p.disabled || !p.supported {
		return ""
	}
	if _, _, ok := p.cellSize(); !ok {
		return ""
	}
	p.enabled = true
	return ansi.SetModeMouseExtSgrPixel
}

// cellSize returns the size of a cell in pixels: the terminal's report while
// the window fits it, or else the window's size in pixels over its cells.
func (p *pixelPointer) cellSize() (width, height float64, ok bool) {
	reported := p.cellWidth > 0 && p.cellHeight > 0
	if reported && p.reportFits() {
		return float64(p.cellWidth), float64(p.cellHeight), true
	}
	// A report the window no longer fits (the font size changed, and the
	// terminal is being asked again) is further off than the window's size
	// divided unevenly, which is off by at most the padding.
	if width, height, ok := p.windowCellSize(reported); ok {
		return width, height, true
	}
	if reported {
		return float64(p.cellWidth), float64(p.cellHeight), true
	}
	return 0, 0, false
}

// windowCellSize returns the window's size in pixels over its columns and
// rows, only if they divide evenly unless uneven is set.
func (p *pixelPointer) windowCellSize(uneven bool) (width, height float64, ok bool) {
	w := p.window
	if w.cols <= 0 || w.rows <= 0 || w.pixelWidth < w.cols || w.pixelHeight < w.rows {
		return 0, 0, false
	}
	if !uneven && (w.pixelWidth%w.cols != 0 || w.pixelHeight%w.rows != 0) {
		return 0, 0, false
	}
	return float64(w.pixelWidth) / float64(w.cols), float64(w.pixelHeight) / float64(w.rows), true
}

// locateEvent applies locate to mouse events and returns other events as
// they are.
func (p *pixelPointer) locateEvent(event uv.Event) (uv.Event, float64, float64) {
	switch ev := event.(type) {
	case uv.MouseClickEvent:
		m, subX, subY := p.locate(uv.Mouse(ev))
		return uv.MouseClickEvent(m), subX, subY
	case uv.MouseReleaseEvent:
		m, subX, subY := p.locate(uv.Mouse(ev))
		return uv.MouseReleaseEvent(m), subX, subY
	case uv.MouseMotionEvent:
		return p.locateMotion(ev)
	case uv.MouseWheelEvent:
		m, subX, subY := p.locate(uv.Mouse(ev))
		return uv.MouseWheelEvent(m), subX, subY
	}
	return event, 0.5, 0.5
}

// locateMotion applies locate to a motion event.
func (p *pixelPointer) locateMotion(ev uv.MouseMotionEvent) (uv.MouseMotionEvent, float64, float64) {
	m, subX, subY := p.locate(uv.Mouse(ev))
	return uv.MouseMotionEvent(m), subX, subY
}

// locate converts a mouse position to cells, and returns where in its cell
// the pointer is (from 0 up to 1). Without pixel reporting that is unknown,
// so the cell's centre is assumed.
func (p *pixelPointer) locate(m uv.Mouse) (cell uv.Mouse, subX, subY float64) {
	cellWidth, cellHeight, ok := p.cellSize()
	if !p.enabled || !ok {
		return m, 0.5, 0.5
	}
	// Some terminals report positions in the window's padding as negative.
	x := math.Max(float64(m.X), 0) / cellWidth
	y := math.Max(float64(m.Y), 0) / cellHeight
	cellX, cellY := math.Floor(x), math.Floor(y)
	m.X, m.Y = int(cellX), int(cellY)
	return m, x - cellX, y - cellY
}

// sgrMouseRepair reassembles SGR mouse reports that the event decoder splits
// apart. With pixel reporting, kitty and Ghostty report a pointer outside the
// window with negative coordinates, which the decoder doesn't accept: it stops
// at the "-", returning the report's start as an UnknownEvent and the rest as
// key presses. Left alone, those keys reach the app (the pointer is reported
// leaving the window as it gains or loses focus, so "m" or "M" arrives then),
// and a release outside the window is lost, leaving a drag stuck.
type sgrMouseRepair struct {
	held []byte // The report so far, or nil when none is being held.
}

// sgrMouseMaxLen bounds a held report, so a stray "\x1b[<" can't swallow
// typing for long.
const sgrMouseMaxLen = 32

// feed returns the event to handle in place of event: nil while a split
// report is being held, the mouse event it reports once it is complete, or
// event itself.
func (r *sgrMouseRepair) feed(event uv.Event) uv.Event {
	switch ev := event.(type) {
	case uv.UnknownEvent:
		if strings.HasPrefix(string(ev), "\x1b[<") {
			r.held = []byte(ev)
			return nil
		}
	case uv.KeyPressEvent:
		if r.held == nil || len(ev.Text) != 1 {
			break
		}
		switch c := ev.Text[0]; {
		case c >= '0' && c <= '9', c == ';', c == '-':
			r.held = append(r.held, c)
			if len(r.held) > sgrMouseMaxLen {
				r.reset()
			}
			return nil
		case c == 'M', c == 'm':
			repaired := repairSgrMouse(append(r.held, c))
			r.reset()
			return repaired
		}
	}
	// Anything else ends the report: drop what was held.
	r.reset()
	return event
}

// reset drops any held report.
func (r *sgrMouseRepair) reset() {
	r.held = nil
}

// repairSgrMouse decodes an SGR mouse report with its negative coordinates
// replaced by 0, returning nil if it isn't a mouse report.
func repairSgrMouse(seq []byte) uv.Event {
	body := string(seq[len("\x1b[<") : len(seq)-1])
	params := strings.Split(body, ";")
	if len(params) != 3 {
		return nil
	}
	for i, param := range params {
		n, err := strconv.Atoi(param)
		if err != nil {
			return nil
		}
		params[i] = strconv.Itoa(max(n, 0))
	}
	repaired := "\x1b[<" + strings.Join(params, ";") + string(seq[len(seq)-1])
	var decoder uv.EventDecoder
	n, event := decoder.Decode([]byte(repaired))
	if n != len(repaired) {
		return nil
	}
	switch event.(type) {
	case uv.MouseClickEvent, uv.MouseReleaseEvent, uv.MouseMotionEvent, uv.MouseWheelEvent:
		return event
	}
	return nil
}
