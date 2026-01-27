package terma

// EmptyWidget is a placeholder widget that renders nothing and takes no space.
// Use this directly or via ShowWhen/HideWhen for conditional rendering.
type EmptyWidget struct{}

// Build returns itself since EmptyWidget handles its own layout and rendering.
func (e EmptyWidget) Build(ctx BuildContext) Widget {
	return e
}

// Layout returns zero size - EmptyWidget takes no space in the layout.
func (e EmptyWidget) Layout(_ BuildContext, _ Constraints) Size {
	return Size{}
}

// Render does nothing - EmptyWidget is invisible.
func (e EmptyWidget) Render(_ *RenderContext) {}

// ShowWhen conditionally renders a child widget.
// When condition is true, returns the child widget normally.
// When condition is false, returns EmptyWidget (no space, no render).
//
// Example:
//
//	ShowWhen(user.IsAdmin(), AdminPanel{})
//	ShowWhen(items.Len() > 0, ItemList{})
func ShowWhen(condition bool, child Widget) Widget {
	if condition {
		return child
	}
	return EmptyWidget{}
}

// HideWhen is the inverse of ShowWhen.
// When condition is true, returns EmptyWidget (hidden).
// When condition is false, returns the child widget.
//
// Example:
//
//	HideWhen(isLoading.Get(), Content{})
func HideWhen(condition bool, child Widget) Widget {
	return ShowWhen(!condition, child)
}

// invisibleWrapper reserves layout space for a child but optionally skips rendering.
// Used by VisibleWhen/InvisibleWhen for CSS visibility-like behavior.
type invisibleWrapper struct {
	child   Widget
	visible bool
}

// Build returns itself to handle Layout and Render.
func (w invisibleWrapper) Build(_ BuildContext) Widget {
	return w
}

// Layout builds the child and returns its size, reserving space regardless of visibility.
func (w invisibleWrapper) Layout(ctx BuildContext, constraints Constraints) Size {
	built := w.child.Build(ctx)
	if layoutable, ok := built.(Layoutable); ok {
		return layoutable.Layout(ctx, constraints)
	}
	return Size{}
}

// Render only renders the child when visible.
// When invisible, the space is reserved but nothing is drawn and no focus/hover events occur.
func (w invisibleWrapper) Render(ctx *RenderContext) {
	// No-op - renderTree takes care of rendering now.
}

// VisibleWhen reserves space for a child regardless of visibility.
// When condition is true, the child is rendered normally.
// When condition is false, space is reserved but nothing is rendered
// and the widget does not receive focus or hover events.
//
// This is similar to CSS `visibility: hidden` - the element is invisible
// but still affects layout.
//
// Example:
//
//	VisibleWhen(hasData.Get(), Chart{})  // reserves chart space even when no data
func VisibleWhen(condition bool, child Widget) Widget {
	return invisibleWrapper{child: child, visible: condition}
}

// InvisibleWhen is the inverse of VisibleWhen.
// When condition is true, space is reserved but child is not rendered.
// When condition is false, child is rendered normally.
//
// Example:
//
//	InvisibleWhen(isSecret, SensitiveData{})  // space held but content hidden
func InvisibleWhen(condition bool, child Widget) Widget {
	return VisibleWhen(!condition, child)
}

// inertWrapper prevents focus collection for a subtree without applying disabled styling.
// Unlike disabledWrapper, inert widgets render normally but cannot receive keyboard focus.
// The wrapper is detected by BuildRenderTree, which passes nil for the focus collector
// so no focusable widgets in the subtree are registered.
type inertWrapper struct {
	child Widget
}

// Build returns the child widget directly.
func (w inertWrapper) Build(ctx BuildContext) Widget {
	return w.child.Build(ctx)
}

// Inert prevents all focusable widgets in the subtree from receiving keyboard focus.
// Unlike DisabledWhen, the child renders normally with no visual changes.
// Use this for display-only instances of interactive widgets (e.g., a Table used
// for layout inside another Table's cell).
//
// Example:
//
//	Inert(Table[T]{State: previewState, ...})
func Inert(child Widget) Widget {
	return inertWrapper{child: child}
}

// disabledWrapper marks a subtree as disabled, preventing focus and showing disabled styling.
// The child is still built and laid out normally, but focusable widgets within the subtree
// cannot receive focus and should render in a disabled state.
//
// The wrapper is detected by BuildRenderTree BEFORE Build() is called, which sets
// ctx.disabled = true. Then Build() returns the child directly, so the child
// is built with the disabled context and renders/lays out normally.
type disabledWrapper struct {
	child    Widget
	disabled bool
}

// Build returns the child widget directly.
// The disabled context is set by BuildRenderTree before this is called.
func (w disabledWrapper) Build(ctx BuildContext) Widget {
	return w.child.Build(ctx)
}

// DisabledWhen disables all focusable widgets in the subtree when condition is true.
// Disabled widgets cannot receive keyboard focus and should render with disabled styling.
// The child is still rendered and takes up space in layout.
//
// Example:
//
//	DisabledWhen(!form.IsValid(), Button{Label: "Submit", OnPress: submit})
func DisabledWhen(condition bool, child Widget) Widget {
	return disabledWrapper{child: child, disabled: condition}
}

// EnabledWhen is the inverse of DisabledWhen.
// When condition is true, the child is rendered normally.
// When condition is false, the child is disabled.
//
// Example:
//
//	EnabledWhen(form.IsValid(), Button{Label: "Submit", OnPress: submit})
func EnabledWhen(condition bool, child Widget) Widget {
	return DisabledWhen(!condition, child)
}
