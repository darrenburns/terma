package terma

import (
	"os"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// screenMode is what differs between running in the alternate screen and
// inline below the shell prompt: where the renderer paints, how a frame gets
// to the terminal, and how the terminal is taken, handed over and given back.
type screenMode interface {
	// enter is called once the terminal has started, before anything is drawn.
	enter(t *uv.Terminal)
	// mouse reports whether to turn on mouse reporting.
	mouse() bool
	// canvas returns what the renderer paints into, and the size to lay the
	// root out in, for a window of width x height cells.
	canvas(t *uv.Terminal, width, height int) (buf CellBuffer, w, h int)
	// present gets the renderer's frame onto the terminal. display draws what
	// the terminal's buffer holds (images included) and flushes it.
	present(t *uv.Terminal, r *Renderer, display func() error) error
	// resize follows the window to width x height cells.
	resize(t *uv.Terminal, r *Renderer, width, height int)
	// pause hands the terminal over to the shell or another program.
	pause(t *uv.Terminal) error
	// resume takes the terminal back after pause.
	resume(t *uv.Terminal) error
	// leave is called on exit, before the terminal is shut down.
	leave(t *uv.Terminal)
	// restored is called after the terminal has been shut down.
	restored()
	// locate maps a mouse event to the renderer's coordinates, or reports
	// false if it can't be placed yet.
	locate(ev uv.Event) (uv.Event, bool)
	// handle consumes the replies the mode asked the terminal for.
	handle(ev uv.Event) bool
}

// fullscreenMode draws over the whole window in the alternate screen.
type fullscreenMode struct{}

func (fullscreenMode) enter(t *uv.Terminal) { t.EnterAltScreen() }
func (fullscreenMode) mouse() bool          { return true }
func (fullscreenMode) canvas(t *uv.Terminal, width, height int) (CellBuffer, int, int) {
	return t, width, height
}
func (fullscreenMode) present(_ *uv.Terminal, _ *Renderer, display func() error) error {
	return display()
}
func (fullscreenMode) resize(t *uv.Terminal, r *Renderer, width, height int) {
	_ = t.Resize(width, height)
	r.Resize(width, height)
	t.Erase()
}
func (fullscreenMode) pause(t *uv.Terminal) error {
	t.ExitAltScreen()
	return t.Pause()
}
func (fullscreenMode) resume(t *uv.Terminal) error {
	err := t.Resume()
	t.EnterAltScreen()
	return err
}
func (fullscreenMode) leave(*uv.Terminal) {}

// restored leaves the alternate screen in case nothing was drawn: Shutdown
// only leaves it after a frame, but Start has already entered it.
func (fullscreenMode) restored() {
	_, _ = os.Stdout.WriteString(ansi.ResetModeAltScreenSaveCursor)
}
func (fullscreenMode) locate(ev uv.Event) (uv.Event, bool) { return ev, true }
func (fullscreenMode) handle(uv.Event) bool                { return false }
