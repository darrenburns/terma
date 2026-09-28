package terma

import (
	"strconv"
)

// pendingFocusID holds the ID of a widget that should receive focus after the next render.
// Set via RequestFocus() and consumed by the app loop.
var pendingFocusID string

// RequestFocus requests that the widget with the given ID receive focus
// after the current render cycle completes. This can be called from anywhere,
// such as keybind actions or callbacks.
func RequestFocus(id string) {
	pendingFocusID = id
}

// BuildContext provides access to framework state during widget building.
// It is passed to Widget.Build() to allow widgets to access focus state,
// hover state, and other framework features in a declarative way.
type BuildContext struct {
	focusManager *FocusManager
	// Signal that holds the currently focused widget (nil if none)
	focusedSignal AnySignal[Focusable]
	// Signal that holds the currently hovered widget (nil if none)
	hoveredSignal AnySignal[Widget]
	// path tracks the current position in the widget tree for auto-ID generation
	path []int
	// floatCollector gathers Floating widgets for deferred rendering
	floatCollector *FloatCollector
	// disabled is true if within a disabled subtree (set by DisabledWhen wrapper)
	disabled bool
}

// NewBuildContext creates a new build context.
func NewBuildContext(fm *FocusManager, focusedSignal AnySignal[Focusable], hoveredSignal AnySignal[Widget], fc *FloatCollector) BuildContext {
	return BuildContext{
		focusManager:   fm,
		focusedSignal:  focusedSignal,
		hoveredSignal:  hoveredSignal,
		path:           []int{0},
		floatCollector: fc,
	}
}

// AutoID returns an automatically generated ID based on tree position.
// This is used for state persistence when widgets don't provide an explicit ID.
func (ctx BuildContext) AutoID() string {
	return "_auto:" + ctx.pathString()
}

// pathString converts the path slice to a dot-separated string (e.g., "0.1.3").
func (ctx BuildContext) pathString() string {
	if len(ctx.path) == 0 {
		return "0"
	}
	b := make([]byte, 0, len(ctx.path)*3)
	for i, idx := range ctx.path {
		if i > 0 {
			b = append(b, '.')
		}
		b = strconv.AppendInt(b, int64(idx), 10)
	}
	return string(b)
}

// PushChild creates a child context with the given index appended to the path.
// Used by container widgets when rendering children.
func (ctx BuildContext) PushChild(index int) BuildContext {
	newPath := make([]int, len(ctx.path)+1)
	copy(newPath, ctx.path)
	newPath[len(ctx.path)] = index
	return BuildContext{
		focusManager:   ctx.focusManager,
		focusedSignal:  ctx.focusedSignal,
		hoveredSignal:  ctx.hoveredSignal,
		path:           newPath,
		floatCollector: ctx.floatCollector,
		disabled:       ctx.disabled,
	}
}

// IsDisabled returns true if within a disabled subtree.
// Widgets can check this to render in a disabled state and skip interactive behavior.
func (ctx BuildContext) IsDisabled() bool {
	return ctx.disabled
}

// WithDisabled returns a copy of the context with disabled=true.
// Used by DisabledWhen wrapper to mark a subtree as disabled.
func (ctx BuildContext) WithDisabled() BuildContext {
	return BuildContext{
		focusManager:   ctx.focusManager,
		focusedSignal:  ctx.focusedSignal,
		hoveredSignal:  ctx.hoveredSignal,
		path:           ctx.path,
		floatCollector: ctx.floatCollector,
		disabled:       true,
	}
}

// IsFocused returns true if the given widget currently has focus.
// Widgets with an explicit ID are matched by that ID; otherwise the
// position-based AutoID is used as a fallback.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when focus changes.
func (ctx BuildContext) IsFocused(widget Widget) bool {
	ctx.subscribeToFocus()
	if ctx.focusManager == nil {
		return false
	}

	focusedID := ctx.focusManager.FocusedID()
	if focusedID == "" {
		return false
	}

	// Check if widget has an explicit ID
	if identifiable, ok := widget.(Identifiable); ok && identifiable.WidgetID() != "" {
		return identifiable.WidgetID() == focusedID
	}

	// Fall back to auto-ID based on tree position
	return ctx.AutoID() == focusedID
}

// Focused returns the currently focused widget, or nil if none.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when focus changes.
func (ctx BuildContext) Focused() Focusable {
	if !ctx.focusedSignal.IsValid() {
		return nil
	}
	return ctx.focusedSignal.Get()
}

// subscribeToFocus records a dependency on the focused widget. Focus state is
// owned by the FocusManager, so reads through it are otherwise invisible to
// the reactive system and a skipped Build would keep stale focus styling.
func (ctx BuildContext) subscribeToFocus() {
	if ctx.focusedSignal.IsValid() {
		_ = ctx.focusedSignal.Get()
	}
}

// FocusedSignal returns the signal holding the focused widget.
// Useful for more advanced reactive patterns.
func (ctx BuildContext) FocusedSignal() AnySignal[Focusable] {
	return ctx.focusedSignal
}

// ActiveKeybinds returns all declarative keybindings currently active
// based on the focused widget and its ancestors.
// Useful for displaying available keybindings in a footer or help screen.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when focus changes.
func (ctx BuildContext) ActiveKeybinds() []Keybind {
	ctx.subscribeToFocus()
	if ctx.focusManager == nil {
		return nil
	}
	return ctx.focusManager.ActiveKeybinds()
}

// IsHovered returns true if the given widget is currently being hovered.
// The widget must implement Identifiable for hover comparison.
// This is a reactive value, but only the widgets whose answer changes are
// rebuilt: moving between two widgets notifies just those two.
func (ctx BuildContext) IsHovered(widget Widget) bool {
	// Compare by ID to avoid issues with incomparable types (e.g., slices in Column)
	identifiable, ok := widget.(Identifiable)
	if !ok || identifiable.WidgetID() == "" || !ctx.hoveredSignal.IsValid() {
		return false
	}
	id := identifiable.WidgetID()
	return SelectAny(ctx.hoveredSignal, func(hovered Widget) bool {
		return hoveredWidgetID(hovered) == id
	})
}

// Hovered returns the currently hovered widget, or nil if none.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when hover changes.
func (ctx BuildContext) Hovered() Widget {
	if !ctx.hoveredSignal.IsValid() {
		return nil
	}
	return ctx.hoveredSignal.Get()
}

// HoveredID returns the ID of the currently hovered widget ("" if none).
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when hover changes.
func (ctx BuildContext) HoveredID() string {
	if !ctx.hoveredSignal.IsValid() {
		return ""
	}
	return hoveredWidgetID(ctx.hoveredSignal.Get())
}

func hoveredWidgetID(hovered Widget) string {
	if identifiable, ok := hovered.(Identifiable); ok {
		return identifiable.WidgetID()
	}
	return ""
}

// Theme returns the current theme data.
// This is a reactive value - reading it during Build() will cause
// the widget to rebuild when the theme changes.
//
// Example:
//
//	func (w *MyWidget) Build(ctx BuildContext) Widget {
//	    theme := ctx.Theme()
//	    return Text{
//	        Content: "Hello",
//	        Style: Style{
//	            ForegroundColor: theme.Text,
//	            BackgroundColor: theme.Surface,
//	        },
//	    }
//	}
func (ctx BuildContext) Theme() ThemeData {
	return getTheme()
}

// RequestFocus requests that the widget with the given ID receive focus
// after the current render cycle completes. This is useful for programmatically
// moving focus, such as when showing inline edit fields.
//
// Example:
//
//	func (a *App) startEdit() {
//	    a.editingIndex.Set(currentIndex)
//	    // Focus will be applied after the next render
//	}
//
//	func (a *App) Build(ctx BuildContext) Widget {
//	    if a.editingIndex.Get() >= 0 {
//	        ctx.RequestFocus("edit-input")
//	    }
//	    // ...
//	}
func (ctx BuildContext) RequestFocus(id string) {
	pendingFocusID = id
}
