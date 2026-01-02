package terma

// BuildContext provides access to framework state during widget building.
// It is passed to Widget.Build() to allow widgets to access focus state,
// hover state, and other framework features in a declarative way.
type BuildContext struct {
	focusManager *FocusManager
	// Signal that holds the currently focused widget (nil if none)
	focusedSignal *AnySignal[Focusable]
	// Signal that holds the currently hovered widget (nil if none)
	hoveredSignal *AnySignal[Widget]
}

// NewBuildContext creates a new build context.
func NewBuildContext(fm *FocusManager, focusedSignal *AnySignal[Focusable], hoveredSignal *AnySignal[Widget]) BuildContext {
	return BuildContext{
		focusManager:  fm,
		focusedSignal: focusedSignal,
		hoveredSignal: hoveredSignal,
	}
}

// IsFocused returns true if the given widget currently has focus.
// The widget should implement Identifiable for reliable focus tracking across rebuilds.
func (ctx BuildContext) IsFocused(widget Widget) bool {
	if ctx.focusManager == nil {
		return false
	}

	focusedID := ctx.focusManager.FocusedID()
	if focusedID == "" {
		return false
	}

	// Check if widget has an explicit ID
	if identifiable, ok := widget.(Identifiable); ok {
		return identifiable.WidgetID() == focusedID
	}

	return false
}

// Focused returns the currently focused widget, or nil if none.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when focus changes.
func (ctx BuildContext) Focused() Focusable {
	if ctx.focusedSignal == nil {
		return nil
	}
	return ctx.focusedSignal.Get()
}

// FocusedSignal returns the signal holding the focused widget.
// Useful for more advanced reactive patterns.
func (ctx BuildContext) FocusedSignal() *AnySignal[Focusable] {
	return ctx.focusedSignal
}

// ActiveKeybinds returns all declarative keybindings currently active
// based on the focused widget and its ancestors.
// Useful for displaying available keybindings in a footer or help screen.
func (ctx BuildContext) ActiveKeybinds() []Keybind {
	if ctx.focusManager == nil {
		return nil
	}
	return ctx.focusManager.ActiveKeybinds()
}

// IsHovered returns true if the given widget is currently being hovered.
// The widget must implement Identifiable for hover comparison.
func (ctx BuildContext) IsHovered(widget Widget) bool {
	hoveredID := ctx.HoveredID()
	if hoveredID == "" {
		return false
	}

	// Compare by ID to avoid issues with incomparable types (e.g., slices in Column)
	if identifiable, ok := widget.(Identifiable); ok {
		return identifiable.WidgetID() == hoveredID
	}

	return false
}

// Hovered returns the currently hovered widget, or nil if none.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when hover changes.
func (ctx BuildContext) Hovered() Widget {
	if ctx.hoveredSignal == nil {
		return nil
	}
	return ctx.hoveredSignal.Get()
}

// HoveredID returns the ID of the currently hovered widget ("" if none).
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when hover changes.
func (ctx BuildContext) HoveredID() string {
	if ctx.hoveredSignal == nil {
		return ""
	}
	hovered := ctx.hoveredSignal.Get()
	if hovered == nil {
		return ""
	}
	if identifiable, ok := hovered.(Identifiable); ok {
		return identifiable.WidgetID()
	}
	return ""
}
