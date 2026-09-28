package terma

import (
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

const clickChainTimeout = 500 * time.Millisecond

type mouseClickTracker struct {
	lastClickTime time.Time
	lastTargetID  string
	lastButton    uv.MouseButton
	lastX, lastY  int
	clickCount    int

	lastDownTargetID string
	lastDownButton   uv.MouseButton
	lastDownCount    int
}

func (t *mouseClickTracker) nextClick(targetID string, button uv.MouseButton, x, y int, now time.Time) int {
	samePosition := targetID == t.lastTargetID && button == t.lastButton && x == t.lastX && y == t.lastY
	if samePosition && now.Sub(t.lastClickTime) <= clickChainTimeout {
		t.clickCount++
	} else {
		t.clickCount = 1
	}
	t.lastTargetID = targetID
	t.lastButton = button
	t.lastX = x
	t.lastY = y
	t.lastClickTime = now

	t.lastDownTargetID = targetID
	t.lastDownButton = button
	t.lastDownCount = t.clickCount

	return t.clickCount
}

func (t *mouseClickTracker) releaseCount(targetID string, button uv.MouseButton) int {
	if targetID == t.lastDownTargetID && (button == t.lastDownButton || button == uv.MouseNone) {
		return t.lastDownCount
	}
	return 1
}

func buildMouseEvent(m uv.Mouse, subX, subY float64, entry *WidgetEntry, clickCount int) MouseEvent {
	widgetID := ""
	localX, localY := m.X, m.Y
	if entry != nil {
		widgetID = entry.ID
		localX = m.X - entry.Bounds.X
		localY = m.Y - entry.Bounds.Y
	}
	return MouseEvent{
		X:          m.X,
		Y:          m.Y,
		LocalX:     localX,
		LocalY:     localY,
		Button:     m.Button,
		Mod:        m.Mod,
		ClickCount: clickCount,
		WidgetID:   widgetID,
		SubCellX:   subX,
		SubCellY:   subY,
	}
}

// mouseRouter delivers terminal mouse input to widgets. It hit-tests each
// event against the last rendered frame, tracks click chains and hover, and
// keeps the widget that took a press receiving motion and the release until
// the button comes up (implicit pointer capture), wherever the pointer goes.
type mouseRouter struct {
	renderer      *Renderer
	focusManager  *FocusManager
	hoveredSignal AnySignal[Widget]
	clicks        mouseClickTracker
	hover         hoverTracker
	resolveHover  hoverTargetResolver

	// pressed is set from a press until its release. captureID is the widget
	// that took it, if any: a press can also dismiss an overlay or land on a
	// modal's backdrop, and then its release goes nowhere. ownerID is the
	// focusable or pointerOwner around it that was also told of the press
	// (such as the List owning a pressed row), and it follows the drag too.
	pressed       bool
	captureID     string
	ownerID       string
	captureButton uv.MouseButton
}

func newMouseRouter(renderer *Renderer, focusManager *FocusManager, hoveredSignal AnySignal[Widget]) *mouseRouter {
	m := &mouseRouter{
		renderer:      renderer,
		focusManager:  focusManager,
		hoveredSignal: hoveredSignal,
	}
	m.resolveHover = m.hoverTarget
	return m
}

// blocked reports whether a modal overlay covers (x, y), so nothing beneath
// it may receive the event.
func (m *mouseRouter) blocked(x, y int) bool {
	return m.renderer.FloatAt(x, y) == nil && m.renderer.HasModalFloat()
}

// target resolves the widget under (x, y). consumed reports that an overlay
// took the event instead: a modal covers the point, or (when dismiss is set)
// a press outside the top overlay dismissed it.
func (m *mouseRouter) target(x, y int, dismiss bool) (entry *WidgetEntry, consumed bool) {
	if m.renderer.FloatAt(x, y) == nil && m.renderer.HasFloats() {
		if dismiss {
			topFloat := m.renderer.TopFloat()
			if topFloat != nil && topFloat.Config.shouldDismissOnClickOutside() && topFloat.Config.OnDismiss != nil {
				topFloat.Config.OnDismiss()
				return nil, true
			}
		}
		if m.renderer.HasModalFloat() {
			return nil, true
		}
	}
	return m.renderer.WidgetAt(x, y), false
}

func (m *mouseRouter) hoverTarget(x, y int) *WidgetEntry {
	entry, consumed := m.target(x, y, false)
	if consumed {
		return nil
	}
	return entry
}

// captured returns the widget holding the pointer capture, if it still exists.
func (m *mouseRouter) captured() *WidgetEntry {
	return m.renderer.WidgetByID(m.captureID)
}

// capturedOwner returns the owner sharing the pointer capture, if any.
func (m *mouseRouter) capturedOwner() *WidgetEntry {
	if m.ownerID == "" {
		return nil
	}
	return m.renderer.WidgetByID(m.ownerID)
}

// subX and subY place the pointer within its cell (see MouseEvent.SubCellX).
func (m *mouseRouter) press(ev uv.MouseClickEvent, subX, subY float64, now time.Time) {
	m.pressed = true
	m.captureID = ""
	m.ownerID = ""
	m.captureButton = ev.Button

	entry, consumed := m.target(ev.X, ev.Y, true)
	// A disabled widget absorbs the press: nothing is focused or notified.
	if consumed || entry == nil || entry.Disabled {
		return
	}

	// The pressed widget may be a non-focusable child (e.g. Text inside a
	// List), so focus goes to the innermost focusable under the pointer.
	focusEntry := m.renderer.FocusableAt(ev.X, ev.Y)
	if focusEntry != nil {
		m.focusManager.FocusByID(focusEntry.ID)
	}
	clickCount := m.clicks.nextClick(entry.ID, ev.Button, ev.X, ev.Y, now)
	event := buildMouseEvent(uv.Mouse(ev), subX, subY, entry, clickCount)

	m.captureID = entry.ID

	if handler, ok := entry.EventWidget.(MouseDownHandler); ok {
		handler.OnMouseDown(event)
	}
	// Also notify the owner of the pressed widget: the innermost focusable or
	// pointerOwner around it. This lets widgets whose parts are separate
	// widgets (List rows, TextInput text, etc.) place their cursor.
	if owner := m.renderer.PointerOwnerAt(ev.X, ev.Y); owner != nil && owner != entry {
		m.ownerID = owner.ID
		if handler, ok := owner.EventWidget.(MouseDownHandler); ok {
			handler.OnMouseDown(buildMouseEvent(uv.Mouse(ev), subX, subY, owner, clickCount))
		}
	}
	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}
}

// release ends the press. It goes to the widget that took the press, so a
// drag released over another widget still ends where it started.
func (m *mouseRouter) release(ev uv.MouseReleaseEvent, subX, subY float64) {
	pressed := m.pressed
	pressedID := m.captureID
	entry := m.captured()
	owner := m.capturedOwner()
	m.pressed = false
	m.captureID = ""
	m.ownerID = ""
	m.captureButton = uv.MouseNone

	if !pressed {
		// The press wasn't reported (e.g. it happened before the app started).
		var consumed bool
		// A disabled widget absorbs the release, as it does the press.
		if entry, consumed = m.target(ev.X, ev.Y, false); consumed || entry == nil || entry.Disabled {
			return
		}
		pressedID = entry.ID
	}
	// The pressed widget may be gone (e.g. a row scrolled out of a List)
	// while its owner remains, and the owner still needs the release.
	clickCount := m.clicks.releaseCount(pressedID, ev.Button)
	if entry != nil {
		if handler, ok := entry.EventWidget.(MouseUpHandler); ok {
			handler.OnMouseUp(buildMouseEvent(uv.Mouse(ev), subX, subY, entry, clickCount))
		}
	}
	if owner != nil && owner != entry {
		if handler, ok := owner.EventWidget.(MouseUpHandler); ok {
			handler.OnMouseUp(buildMouseEvent(uv.Mouse(ev), subX, subY, owner, clickCount))
		}
	}
}

// motion handles pointer movement, reporting whether a render is needed.
func (m *mouseRouter) motion(ev uv.MouseMotionEvent, subX, subY float64) bool {
	changed := false
	if m.pressed {
		if ev.Button == uv.MouseNone {
			// Motion with no button held means the release was never reported,
			// e.g. it happened outside the terminal window.
			m.release(uv.MouseReleaseEvent{X: ev.X, Y: ev.Y, Button: m.captureButton, Mod: ev.Mod}, subX, subY)
			changed = true
		} else {
			pointer := uv.Mouse{X: ev.X, Y: ev.Y, Button: m.captureButton, Mod: ev.Mod}
			entry := m.captured()
			if entry != nil {
				if handler, ok := entry.EventWidget.(MouseMoveHandler); ok {
					handler.OnMouseMove(buildMouseEvent(pointer, subX, subY, entry, 1))
					changed = true
				}
			}
			if owner := m.capturedOwner(); owner != nil && owner != entry {
				if handler, ok := owner.EventWidget.(MouseMoveHandler); ok {
					handler.OnMouseMove(buildMouseEvent(pointer, subX, subY, owner, 1))
					changed = true
				}
			}
		}
	}
	// Pixel reporting also sends motion within a cell. The hover target can't
	// change until the pointer changes cell or a frame is drawn, and every
	// frame reconciles hover itself.
	if m.hover.pointerChanged(ev.X, ev.Y, ev.Mod, ev.Button) &&
		m.hover.UpdatePointer(ev.X, ev.Y, ev.Mod, ev.Button, m.resolveHover, m.hoveredSignal) {
		changed = true
	}
	return changed
}

// wheel scrolls the innermost scrollable under the pointer that can move,
// reporting whether one did.
func (m *mouseRouter) wheel(ev uv.MouseWheelEvent) bool {
	if m.blocked(ev.X, ev.Y) {
		return false
	}
	return dispatchMouseWheel(m.renderer, ev.X, ev.Y, ev.Button)
}

// reconcileHover re-resolves hover at the last pointer position, so enter and
// leave fire when layout moves widgets under a stationary pointer.
func (m *mouseRouter) reconcileHover() bool {
	return m.hover.Reconcile(m.resolveHover, m.hoveredSignal)
}

// dispatchMouseWheel routes wheel events to scrollable widgets under the cursor.
// Scrollables are tried from innermost to outermost until one handles the event.
func dispatchMouseWheel(renderer *Renderer, x int, y int, button uv.MouseButton) bool {
	if renderer == nil {
		return false
	}
	for _, scrollable := range renderer.ScrollablesAt(x, y) {
		var handled bool
		switch button {
		case uv.MouseWheelUp:
			handled = scrollable.ScrollUp(1)
		case uv.MouseWheelDown:
			handled = scrollable.ScrollDown(1)
		case uv.MouseWheelLeft:
			handled = scrollable.ScrollLeft(1)
		case uv.MouseWheelRight:
			handled = scrollable.ScrollRight(1)
		}
		if handled {
			return true
		}
	}
	return false
}

// coalesceMouseMotion folds motion events already queued on events into ev,
// so a burst the event loop fell behind on costs one hit test and one handler
// call rather than one each. Only motion with the same buttons and modifiers
// is folded. It returns the latest motion and the first queued event it could
// not fold, or nil.
func coalesceMouseMotion(ev uv.MouseMotionEvent, events <-chan uv.Event) (uv.MouseMotionEvent, uv.Event) {
	for {
		select {
		case next, ok := <-events:
			if !ok {
				return ev, nil
			}
			motion, isMotion := next.(uv.MouseMotionEvent)
			if !isMotion || motion.Button != ev.Button || motion.Mod != ev.Mod {
				return ev, next
			}
			ev = motion
		default:
			return ev, nil
		}
	}
}
