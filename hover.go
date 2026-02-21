package terma

import uv "github.com/charmbracelet/ultraviolet"

// hoverTargetResolver resolves the event target at the given screen coordinates.
type hoverTargetResolver func(x, y int) *WidgetEntry

// hoverTracker tracks pointer hover state and emits transition events on changes.
type hoverTracker struct {
	currentID     string
	currentWidget Widget
	currentBounds Rect

	pointerX      int
	pointerY      int
	pointerMod    uv.KeyMod
	pointerButton uv.MouseButton
	pointerKnown  bool
}

// UpdatePointer records the latest pointer state, resolves the hover target,
// and emits hover transition events if needed.
func (h *hoverTracker) UpdatePointer(x, y int, mod uv.KeyMod, button uv.MouseButton, resolve hoverTargetResolver, hoveredSignal AnySignal[Widget]) bool {
	h.pointerX = x
	h.pointerY = y
	h.pointerMod = mod
	h.pointerButton = button
	h.pointerKnown = true

	if resolve == nil {
		return h.applyTarget(nil, x, y, mod, button, hoveredSignal)
	}
	return h.applyTarget(resolve(x, y), x, y, mod, button, hoveredSignal)
}

// Reconcile re-checks hover against the most recently known pointer position.
// This allows hover leave/enter to fire when layout changes while the pointer
// is stationary.
func (h *hoverTracker) Reconcile(resolve hoverTargetResolver, hoveredSignal AnySignal[Widget]) bool {
	if !h.pointerKnown || resolve == nil {
		return false
	}
	entry := resolve(h.pointerX, h.pointerY)
	return h.applyTarget(entry, h.pointerX, h.pointerY, h.pointerMod, h.pointerButton, hoveredSignal)
}

func (h *hoverTracker) applyTarget(entry *WidgetEntry, x, y int, mod uv.KeyMod, button uv.MouseButton, hoveredSignal AnySignal[Widget]) bool {
	var (
		newID     string
		newWidget Widget
		newBounds Rect
	)
	if entry != nil {
		newID = entry.ID
		newWidget = entry.EventWidget
		newBounds = entry.Bounds
	}

	// No transition if identity didn't change.
	if newID == h.currentID {
		if entry != nil {
			// Keep bounds fresh for accurate local coordinates on future leave events.
			h.currentWidget = newWidget
			h.currentBounds = newBounds
		}
		return false
	}

	oldID := h.currentID
	oldWidget := h.currentWidget
	oldBounds := h.currentBounds

	// Leave fires before enter when switching targets.
	if oldWidget != nil {
		dispatchHoverEvent(oldWidget, HoverEvent{
			Type:             HoverLeave,
			X:                x,
			Y:                y,
			LocalX:           x - oldBounds.X,
			LocalY:           y - oldBounds.Y,
			Button:           button,
			Mod:              mod,
			WidgetID:         oldID,
			PreviousWidgetID: oldID,
			NextWidgetID:     newID,
		})
	}

	// Only update the hovered signal when identity changes.
	hoveredSignal.Set(newWidget)

	h.currentID = newID
	h.currentWidget = newWidget
	h.currentBounds = newBounds

	if newWidget != nil {
		dispatchHoverEvent(newWidget, HoverEvent{
			Type:             HoverEnter,
			X:                x,
			Y:                y,
			LocalX:           x - newBounds.X,
			LocalY:           y - newBounds.Y,
			Button:           button,
			Mod:              mod,
			WidgetID:         newID,
			PreviousWidgetID: oldID,
			NextWidgetID:     newID,
		})
	}

	return true
}

// dispatchHoverEvent dispatches hover transitions to the target widget.
func dispatchHoverEvent(widget Widget, event HoverEvent) {
	if widget == nil {
		return
	}
	if hoverable, ok := widget.(Hoverable); ok {
		hoverable.OnHover(event)
	}
}
