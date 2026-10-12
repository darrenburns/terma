package terma

// PointerShape is a mouse pointer shape from the Kitty OSC 22 protocol.
// Terminals without support leave the pointer unchanged. This does not change
// the text cursor. Set a widget's Style.Pointer to select a shape on hover.
type PointerShape string

const (
	PointerInherit      PointerShape = "" // Use the nearest ancestor's shape (the default).
	PointerDefault      PointerShape = "default"
	PointerPointer      PointerShape = "pointer" // A pointing hand.
	PointerText         PointerShape = "text"
	PointerVerticalText PointerShape = "vertical-text"
	PointerHelp         PointerShape = "help"
	PointerWait         PointerShape = "wait"
	PointerProgress     PointerShape = "progress"
	PointerCrosshair    PointerShape = "crosshair"
	PointerCell         PointerShape = "cell"
	PointerMove         PointerShape = "move"
	PointerGrab         PointerShape = "grab"
	PointerGrabbing     PointerShape = "grabbing"
	PointerNotAllowed   PointerShape = "not-allowed"
	PointerNoDrop       PointerShape = "no-drop"
	PointerAlias        PointerShape = "alias"
	PointerCopy         PointerShape = "copy"
	PointerZoomIn       PointerShape = "zoom-in"
	PointerZoomOut      PointerShape = "zoom-out"
	PointerEWResize     PointerShape = "ew-resize"
	PointerNSResize     PointerShape = "ns-resize"
	PointerNESWResize   PointerShape = "nesw-resize"
	PointerNWSEResize   PointerShape = "nwse-resize"
	PointerNResize      PointerShape = "n-resize"
	PointerNEResize     PointerShape = "ne-resize"
	PointerEResize      PointerShape = "e-resize"
	PointerSEResize     PointerShape = "se-resize"
	PointerSResize      PointerShape = "s-resize"
	PointerSWResize     PointerShape = "sw-resize"
	PointerWResize      PointerShape = "w-resize"
	PointerNWResize     PointerShape = "nw-resize"
)

const resetPointerShape = "\x1b]22;default\x07"

func (p PointerShape) sequence() string {
	// Only protocol names may enter the OSC payload, never arbitrary escapes.
	switch p {
	case PointerDefault, PointerPointer, PointerText, PointerVerticalText,
		PointerHelp, PointerWait, PointerProgress, PointerCrosshair, PointerCell,
		PointerMove, PointerGrab, PointerGrabbing, PointerNotAllowed, PointerNoDrop,
		PointerAlias, PointerCopy, PointerZoomIn, PointerZoomOut,
		PointerEWResize, PointerNSResize, PointerNESWResize, PointerNWSEResize,
		PointerNResize, PointerNEResize, PointerEResize, PointerSEResize,
		PointerSResize, PointerSWResize, PointerWResize, PointerNWResize:
		return "\x1b]22;" + string(p) + "\x07"
	default:
		return resetPointerShape
	}
}

// pointerShapeResolver lets painted regions, such as a divider, choose a
// pointer independently of the containing widget's style.
type pointerShapeResolver interface {
	pointerShapeAt(entry *WidgetEntry, x, y int) PointerShape
}

func (m *mouseRouter) updatePointerShape() {
	if !m.pointerKnown {
		return
	}
	x, y := m.pointerX, m.pointerY
	entry := m.captured()
	if entry == nil {
		entry, _ = m.target(x, y, false)
	}
	shape := PointerDefault
	for entry != nil && !entry.Disabled {
		resolver, ok := entry.EventWidget.(pointerShapeResolver)
		if !ok {
			resolver, ok = entry.Widget.(pointerShapeResolver)
		}
		if ok {
			if resolved := resolver.pointerShapeAt(entry, x, y); resolved != PointerInherit {
				shape = resolved
				break
			}
		}
		if styled, ok := entry.Widget.(Styled); ok {
			if pointer := styled.GetStyle().Pointer; pointer != PointerInherit {
				shape = pointer
				break
			}
		}
		entry = m.renderer.WidgetByID(entry.parentID)
	}
	sequence := shape.sequence()
	if sequence != m.pointerSequence {
		m.pointerSequence = sequence
		WriteTerminal(sequence)
	}
}
