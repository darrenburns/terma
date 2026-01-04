package terma

// Offset represents X, Y coordinates for positioning.
type Offset struct {
	X, Y int
}

// FloatPosition specifies absolute positioning for floating widgets.
type FloatPosition int

const (
	// FloatPositionAbsolute uses the Offset as absolute screen coordinates.
	FloatPositionAbsolute FloatPosition = iota
	// FloatPositionCenter centers the float on the screen.
	FloatPositionCenter
	// FloatPositionTopCenter positions the float at the top center of the screen.
	FloatPositionTopCenter
	// FloatPositionBottomCenter positions the float at the bottom center of the screen.
	FloatPositionBottomCenter
)

// AnchorPoint specifies where on an anchor widget to attach a floating widget.
type AnchorPoint int

const (
	// AnchorTopLeft positions the float at the top-left of the anchor.
	AnchorTopLeft AnchorPoint = iota
	// AnchorTopCenter positions the float at the top-center of the anchor.
	AnchorTopCenter
	// AnchorTopRight positions the float at the top-right of the anchor.
	AnchorTopRight
	// AnchorBottomLeft positions the float below the anchor, aligned to its left edge.
	AnchorBottomLeft
	// AnchorBottomCenter positions the float below the anchor, centered.
	AnchorBottomCenter
	// AnchorBottomRight positions the float below the anchor, aligned to its right edge.
	AnchorBottomRight
	// AnchorLeftTop positions the float to the left of the anchor, aligned to its top.
	AnchorLeftTop
	// AnchorLeftCenter positions the float to the left of the anchor, centered vertically.
	AnchorLeftCenter
	// AnchorLeftBottom positions the float to the left of the anchor, aligned to its bottom.
	AnchorLeftBottom
	// AnchorRightTop positions the float to the right of the anchor, aligned to its top.
	AnchorRightTop
	// AnchorRightCenter positions the float to the right of the anchor, centered vertically.
	AnchorRightCenter
	// AnchorRightBottom positions the float to the right of the anchor, aligned to its bottom.
	AnchorRightBottom
)

// FloatConfig configures positioning and behavior for a floating widget.
type FloatConfig struct {
	// Anchor-based positioning (use AnchorID + Anchor).
	// If AnchorID is set, the float is positioned relative to that widget.
	AnchorID string      // ID of the widget to anchor to
	Anchor   AnchorPoint // Where on the anchor widget to attach

	// Absolute positioning (used when AnchorID is empty).
	Position FloatPosition

	// Offset from the calculated position.
	Offset Offset

	// Modal behavior - when true, traps focus and shows a backdrop.
	Modal bool

	// DismissOnEsc dismisses the float when Escape is pressed.
	// Defaults to true if OnDismiss is set.
	DismissOnEsc *bool

	// DismissOnClickOutside dismisses the float when clicking outside it.
	// Defaults to true for non-modal floats when OnDismiss is set.
	DismissOnClickOutside *bool

	// OnDismiss is called when the float should be dismissed.
	OnDismiss func()

	// BackdropColor is the color of the modal backdrop.
	// Only used when Modal is true. Defaults to semi-transparent black.
	BackdropColor Color
}

// shouldDismissOnEsc returns whether the float should dismiss on Escape key.
func (c FloatConfig) shouldDismissOnEsc() bool {
	if c.DismissOnEsc != nil {
		return *c.DismissOnEsc
	}
	// Default: dismiss on Esc if OnDismiss is set
	return c.OnDismiss != nil
}

// shouldDismissOnClickOutside returns whether the float should dismiss on click outside.
func (c FloatConfig) shouldDismissOnClickOutside() bool {
	if c.DismissOnClickOutside != nil {
		return *c.DismissOnClickOutside
	}
	// Default: dismiss on click outside for non-modal floats if OnDismiss is set
	return !c.Modal && c.OnDismiss != nil
}

// Floating is a widget that renders its child as an overlay on top of other widgets.
// The child is rendered after the main widget tree, ensuring it appears on top.
//
// For modal floats, the backdrop blocks clicks to underlying widgets and Escape
// dismisses the modal. Note: Full focus trapping (Tab cycling only within modal)
// is not yet implemented - Tab will cycle through all focusables including those
// behind the modal.
//
// Example - dropdown menu anchored to a button:
//
//	Floating{
//	    Visible: m.showMenu.Get(),
//	    Config: FloatConfig{
//	        AnchorID:  "file-btn",
//	        Anchor:    AnchorBottomLeft,
//	        OnDismiss: func() { m.showMenu.Set(false) },
//	    },
//	    Child: Menu{Items: menuItems},
//	}
//
// Example - centered modal dialog:
//
//	Floating{
//	    Visible: a.showDialog.Get(),
//	    Config: FloatConfig{
//	        Position:  FloatPositionCenter,
//	        Modal:     true,
//	        OnDismiss: func() { a.showDialog.Set(false) },
//	    },
//	    Child: Dialog{Title: "Confirm", ...},
//	}
type Floating struct {
	// Visible controls whether the floating widget is shown.
	Visible bool

	// Config specifies positioning and behavior.
	Config FloatConfig

	// Child is the widget to render as an overlay.
	Child Widget
}

// Build registers the floating widget with the collector if visible.
// Returns an empty widget since the actual rendering happens in the overlay phase.
func (f Floating) Build(ctx BuildContext) Widget {
	if !f.Visible || f.Child == nil {
		return EmptyWidget{}
	}

	// Register with the float collector for deferred rendering
	if ctx.floatCollector != nil {
		ctx.floatCollector.Add(FloatEntry{
			Config: f.Config,
			Child:  f.Child,
		})
	}

	return EmptyWidget{}
}

// FloatEntry stores a registered floating widget for deferred rendering.
type FloatEntry struct {
	Config FloatConfig
	Child  Widget
	// Computed position after layout (set during render phase)
	X, Y          int
	Width, Height int
}

// FloatCollector gathers Floating widgets during the build phase
// for rendering after the main widget tree.
type FloatCollector struct {
	entries []FloatEntry
}

// NewFloatCollector creates a new float collector.
func NewFloatCollector() *FloatCollector {
	return &FloatCollector{}
}

// Add registers a floating widget entry.
func (c *FloatCollector) Add(entry FloatEntry) {
	c.entries = append(c.entries, entry)
}

// Entries returns all registered floating widgets.
func (c *FloatCollector) Entries() []FloatEntry {
	return c.entries
}

// HasModal returns true if any registered float is modal.
func (c *FloatCollector) HasModal() bool {
	for _, entry := range c.entries {
		if entry.Config.Modal {
			return true
		}
	}
	return false
}

// TopModal returns the topmost modal float entry, or nil if none.
func (c *FloatCollector) TopModal() *FloatEntry {
	for i := len(c.entries) - 1; i >= 0; i-- {
		if c.entries[i].Config.Modal {
			return &c.entries[i]
		}
	}
	return nil
}

// Reset clears all entries for a new render pass.
func (c *FloatCollector) Reset() {
	c.entries = c.entries[:0]
}

// Len returns the number of registered floats.
func (c *FloatCollector) Len() int {
	return len(c.entries)
}

// calculateAnchorPosition computes the position for an anchor-based float.
func calculateAnchorPosition(anchor *WidgetEntry, anchorPoint AnchorPoint, floatWidth, floatHeight int, offset Offset) (x, y int) {
	if anchor == nil {
		return offset.X, offset.Y
	}

	bounds := anchor.Bounds

	switch anchorPoint {
	case AnchorTopLeft:
		x = bounds.X
		y = bounds.Y - floatHeight
	case AnchorTopCenter:
		x = bounds.X + (bounds.Width-floatWidth)/2
		y = bounds.Y - floatHeight
	case AnchorTopRight:
		x = bounds.X + bounds.Width - floatWidth
		y = bounds.Y - floatHeight
	case AnchorBottomLeft:
		x = bounds.X
		y = bounds.Y + bounds.Height
	case AnchorBottomCenter:
		x = bounds.X + (bounds.Width-floatWidth)/2
		y = bounds.Y + bounds.Height
	case AnchorBottomRight:
		x = bounds.X + bounds.Width - floatWidth
		y = bounds.Y + bounds.Height
	case AnchorLeftTop:
		x = bounds.X - floatWidth
		y = bounds.Y
	case AnchorLeftCenter:
		x = bounds.X - floatWidth
		y = bounds.Y + (bounds.Height-floatHeight)/2
	case AnchorLeftBottom:
		x = bounds.X - floatWidth
		y = bounds.Y + bounds.Height - floatHeight
	case AnchorRightTop:
		x = bounds.X + bounds.Width
		y = bounds.Y
	case AnchorRightCenter:
		x = bounds.X + bounds.Width
		y = bounds.Y + (bounds.Height-floatHeight)/2
	case AnchorRightBottom:
		x = bounds.X + bounds.Width
		y = bounds.Y + bounds.Height - floatHeight
	}

	return x + offset.X, y + offset.Y
}

// calculateAbsolutePosition computes the position for an absolutely positioned float.
func calculateAbsolutePosition(position FloatPosition, screenWidth, screenHeight, floatWidth, floatHeight int, offset Offset) (x, y int) {
	switch position {
	case FloatPositionCenter:
		x = (screenWidth - floatWidth) / 2
		y = (screenHeight - floatHeight) / 2
	case FloatPositionTopCenter:
		x = (screenWidth - floatWidth) / 2
		y = 0
	case FloatPositionBottomCenter:
		x = (screenWidth - floatWidth) / 2
		y = screenHeight - floatHeight
	case FloatPositionAbsolute:
		x = 0
		y = 0
	}

	return x + offset.X, y + offset.Y
}

// clampToScreen ensures the float position keeps it visible on screen.
func clampToScreen(x, y, floatWidth, floatHeight, screenWidth, screenHeight int) (int, int) {
	// Clamp X to keep float on screen
	if x < 0 {
		x = 0
	} else if x+floatWidth > screenWidth {
		x = screenWidth - floatWidth
	}

	// Clamp Y to keep float on screen
	if y < 0 {
		y = 0
	} else if y+floatHeight > screenHeight {
		y = screenHeight - floatHeight
	}

	return x, y
}
