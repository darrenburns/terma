package terma

import (
	"math"
	"strconv"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// pixelPointer locates the mouse to a fraction of a cell, as Textual does for
// smooth scrolling. Terminals that support SGR-Pixels mouse mode (1016), such
// as kitty and Ghostty, can report the pointer in pixels; this
// divides those by the size of a cell to get the cell plus the position
// within it. Dragging a scrollbar thumb then follows the pointer exactly.
//
// The mode is switched on only once the terminal has reported it recognises
// it (in reply to a DECRQM query) and the window's size in pixels is known.
// Set TERMA_DISABLE_PIXEL_MOUSE to keep cell-based reporting.
type pixelPointer struct {
	disabled  bool // Switched off by TERMA_DISABLE_PIXEL_MOUSE.
	supported bool // The terminal recognises mode 1016.
	enabled   bool // Mode 1016 has been switched on.

	cols, rows              int
	pixelWidth, pixelHeight int
}

func newPixelPointer() *pixelPointer {
	return &pixelPointer{disabled: boolEnv("TERMA_DISABLE_PIXEL_MOUSE")}
}

// query returns the sequence asking whether the terminal supports mode 1016.
func (p *pixelPointer) query() string {
	if p.disabled {
		return ""
	}
	return ansi.RequestModeMouseExtSgrPixel
}

// handle records a size or mode report, returning the sequence that switches
// pixel reporting on once everything it needs is known.
func (p *pixelPointer) handle(event uv.Event) string {
	switch ev := event.(type) {
	case uv.WindowSizeEvent:
		p.cols, p.rows = ev.Width, ev.Height
	case uv.WindowPixelSizeEvent:
		p.pixelWidth, p.pixelHeight = ev.Width, ev.Height
	case uv.ModeReportEvent:
		if ev.Mode != ansi.ModeMouseExtSgrPixel {
			return ""
		}
		p.supported = !ev.Value.IsNotRecognized() && !ev.Value.IsPermanentlyReset()
	default:
		return ""
	}
	if p.enabled || p.disabled || !p.supported {
		return ""
	}
	if _, _, ok := p.cellSize(); !ok {
		return ""
	}
	p.enabled = true
	return ansi.SetModeMouseExtSgrPixel
}

// cellSize returns the size of a cell in pixels.
func (p *pixelPointer) cellSize() (width, height float64, ok bool) {
	if p.cols <= 0 || p.rows <= 0 || p.pixelWidth < p.cols || p.pixelHeight < p.rows {
		return 0, 0, false
	}
	return float64(p.pixelWidth) / float64(p.cols), float64(p.pixelHeight) / float64(p.rows), true
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
