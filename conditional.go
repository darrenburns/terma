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
	if w.visible {
		ctx.RenderChild(0, w.child, 0, 0, ctx.Width, ctx.Height)
	}
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
