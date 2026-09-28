package terma

import (
	"math"

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
