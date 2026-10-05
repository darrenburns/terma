package terma

import (
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// appSession turns input events into frames: it owns the retained renderer,
// focus, modal focus transitions, hover and event routing. Run wraps it with a
// terminal and an event loop; Pilot drives it directly, so tests go through
// the same code path as a running app.
type appSession struct {
	root     Widget
	renderer *Renderer
	focus    *FocusManager
	focused  AnySignal[Focusable]
	mouse    *mouseRouter
	now      func() time.Time
	// onInput runs before a key press, paste or click is routed (Run uses it
	// to restart the cursor blink).
	onInput func()

	lastFocusedID  string
	lastModalCount int
}

// inputResult tells the event loop what an event needs next.
type inputResult int

const (
	inputIgnored inputResult = iota
	inputRender
	inputQuit
)

func newAppSession(root Widget, screen CellBuffer, width, height int, now func() time.Time) *appSession {
	focus := NewFocusManager()
	focus.SetRootWidget(root)
	focused := NewAnySignal[Focusable](nil)
	hovered := NewAnySignal[Widget](nil)
	renderer := NewRenderer(screen, width, height, focus, focused, hovered)
	return &appSession{
		root:     root,
		renderer: renderer,
		focus:    focus,
		focused:  focused,
		mouse:    newMouseRouter(renderer, focus, hovered),
		now:      now,
		onInput:  func() {},
	}
}

// syncFocus publishes the focused widget, reporting whether focus or the
// active keybinds changed.
func (s *appSession) syncFocus() bool {
	keybindsChanged := s.focus.syncKeybinds()
	focusedID := s.focus.FocusedID()
	if focusedID == s.lastFocusedID {
		return keybindsChanged
	}
	s.lastFocusedID = focusedID
	s.focused.Set(s.focus.Focused())
	return true
}

// frame runs dispatched work and draws the root, then settles focus and hover
// against what was drawn.
func (s *appSession) frame() {
	drainPendingDispatches()
	// Update the focused signal BEFORE render so widgets can read it
	s.syncFocus()

	s.focus.SetFocusables(s.renderer.Update(s.root))

	// If focus changed after render (auto-focus or focus removal), re-render
	if s.syncFocus() {
		s.renderer.Update(s.root)
	}

	// Manage modal focus transitions (open/close) and keep focus inside topmost modal.
	modalCount := s.renderer.ModalCount()
	for i := s.lastModalCount; i < modalCount; i++ {
		s.focus.SaveFocus()
	}
	for i := modalCount; i < s.lastModalCount; i++ {
		s.focus.RestoreFocus()
	}
	s.lastModalCount = modalCount

	// Apply pending focus request from ctx.RequestFocus() after modal
	// restore logic so explicit focus requests win.
	if pendingFocusID != "" {
		s.focus.FocusByID(pendingFocusID)
		pendingFocusID = ""
	}

	// Update the signal and re-render so the focused widget shows focus style
	if s.syncFocus() {
		s.renderer.Update(s.root)
	}

	// Reconcile hover after render so enter/leave transitions still fire when
	// layout changes under a stationary pointer.
	if s.mouse.reconcileHover() {
		s.renderer.Update(s.root)
	}
}

// handle routes one input event to the widgets on screen. subX and subY place
// a mouse pointer within its cell.
func (s *appSession) handle(ev uv.Event, subX, subY float64) inputResult {
	switch ev := ev.(type) {
	case uv.KeyPressEvent:
		key := KeyEvent{event: ev}
		if !s.focus.capturesKey(key) && ev.MatchString("ctrl+c") {
			return inputQuit
		}
		s.onInput()
		dispatchKey(s.renderer, s.focus, s.root, key)
		return inputRender

	case uv.PasteEvent:
		// Like a key, a paste goes to what's on screen.
		s.onInput()
		if !dispatchPaste(s.focus, s.root, ev.Content) {
			Log("Paste not handled (%d bytes)", len(ev.Content))
		}
		return inputRender

	case uv.ClipboardEvent:
		deliverClipboard(ev.Content)
		return inputRender

	case uv.MouseClickEvent:
		s.onInput()
		s.mouse.press(ev, subX, subY, s.now())
		return inputRender

	case uv.MouseReleaseEvent:
		s.mouse.release(ev, subX, subY)
		return inputRender

	case uv.MouseMotionEvent:
		if s.mouse.motion(ev, subX, subY) {
			return inputRender
		}

	case uv.MouseWheelEvent:
		if s.mouse.wheelAt(ev, subX, subY) {
			return inputRender
		}
	}
	return inputIgnored
}
