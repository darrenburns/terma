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
	itemHover     itemHoverTracker
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
	if renderer != nil {
		m.hover.targetSignal = renderer.hoverTarget
	}
	return m
}

// blocked reports whether a modal overlay covers (x, y), so nothing beneath
// it may receive the event.
func (m *mouseRouter) blocked(x, y int) bool {
	return m.renderer.pointerFloatAt(x, y) == nil && m.renderer.HasModalFloat()
}

// target resolves the widget under (x, y). consumed reports that an overlay
// took the event instead: a modal covers the point, or (when dismiss is set)
// a press outside the top overlay dismissed it.
func (m *mouseRouter) target(x, y int, dismiss bool) (entry *WidgetEntry, consumed bool) {
	hitFloat := m.renderer.pointerFloatAt(x, y)
	if dismiss {
		topFloat := m.renderer.topPointerFloat()
		// A parent overlay is still outside its nested popup. Consume the
		// dismissing press so it cannot also activate the content underneath.
		if topFloat != nil && hitFloat != topFloat && topFloat.Config.shouldDismissOnClickOutside() && topFloat.Config.OnDismiss != nil {
			topFloat.Config.OnDismiss()
			return nil, true
		}
	}
	if hitFloat == nil && m.renderer.HasModalFloat() {
		return nil, true
	}
	return m.renderer.WidgetAt(x, y), false
}

func (m *mouseRouter) hoverTarget(x, y int) *WidgetEntry {
	entry, consumed := m.target(x, y, false)
	if consumed {
		return nil
	}
	if entry != nil && !entry.Disabled {
		if regions, ok := entry.EventWidget.(hoverRegionResolver); ok {
			if region := regions.hoverRegionAt(entry, x, y); region != nil {
				return region
			}
		}
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
	if m.hover.pointerChanged(ev.X, ev.Y, ev.Mod, ev.Button) {
		if m.hover.UpdatePointer(ev.X, ev.Y, ev.Mod, ev.Button, m.resolveHover, m.hoveredSignal) {
			changed = true
		}
		if m.updateItemHover() {
			changed = true
		}
	}
	return changed
}

// updateItemHover moves the item hover highlight to the collection item under
// the last known pointer position, reporting whether it moved.
func (m *mouseRouter) updateItemHover() bool {
	if !m.hover.pointerKnown {
		return false
	}
	x, y := m.hover.pointerX, m.hover.pointerY
	var item hoverItem
	if !m.blocked(x, y) {
		item = m.renderer.hoverItemAt(x, y)
	}
	return m.itemHover.update(item)
}

// wheel delivers cell-based wheel input; wheelAt also preserves pixel offsets.
func (m *mouseRouter) wheel(ev uv.MouseWheelEvent) bool {
	return m.wheelAt(ev, 0.5, 0.5)
}

func (m *mouseRouter) wheelAt(ev uv.MouseWheelEvent, subX, subY float64) bool {
	if m.blocked(ev.X, ev.Y) {
		return false
	}
	return dispatchMouseWheelEvent(m.renderer, ev, subX, subY)
}

// reconcileHover re-resolves hover at the last pointer position, so enter and
// leave fire when layout moves widgets under a stationary pointer.
func (m *mouseRouter) reconcileHover() bool {
	changed := m.hover.Reconcile(m.resolveHover, m.hoveredSignal)
	if m.updateItemHover() {
		changed = true
	}
	return changed
}

// dispatchMouseWheel is the cell-based convenience path used by scroll tests.
func dispatchMouseWheel(renderer *Renderer, x, y int, button uv.MouseButton) bool {
	return dispatchMouseWheelEvent(renderer, uv.MouseWheelEvent{X: x, Y: y, Button: button}, 0.5, 0.5)
}

// dispatchMouseWheelEvent follows the topmost target's ancestor chain within
// its pointer layer. Each handler gets first refusal before normal scrolling
// at that widget, so a declining child still allows its viewport to scroll.
func dispatchMouseWheelEvent(renderer *Renderer, ev uv.MouseWheelEvent, subX, subY float64) bool {
	if renderer == nil {
		return false
	}
	lo, hi := renderer.pointerLayer(ev.X, ev.Y)
	entry := renderer.widgetRegistry.widgetAtIn(ev.X, ev.Y, lo, hi)
	for entry != nil {
		if !entry.Disabled && entry.Visible.Contains(ev.X, ev.Y) {
			event := buildMouseEvent(uv.Mouse(ev), subX, subY, entry, 0)
			handler, ok := entry.EventWidget.(MouseWheelHandler)
			if !ok {
				handler, ok = entry.Widget.(MouseWheelHandler)
			}
			if ok && handler.OnMouseWheel(event) {
				return true
			}
			var scrollable *Scrollable
			switch widget := entry.Widget.(type) {
			case Scrollable:
				scrollable = &widget
			case *Scrollable:
				scrollable = widget
			}
			if scrollable != nil {
				var handled bool
				switch ev.Button {
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
		}
		// Search only earlier entries in this layer: ancestors are recorded before
		// descendants, and no sibling beneath the target may receive the event.
		parentID := entry.parentID
		entry = nil
		if parentID != "" {
			for i := hi - 1; i >= lo; i-- {
				if renderer.widgetRegistry.entries[i].ID == parentID {
					entry = &renderer.widgetRegistry.entries[i]
					hi = i
					break
				}
			}
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
