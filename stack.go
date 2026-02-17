package terma

import "github.com/darrenburns/terma/layout"

// IntPtr returns a pointer to an int value.
// This is a helper for creating Positioned widgets.
func IntPtr(n int) *int {
	return &n
}

// Positioned wraps a child widget to position it within a Stack using edge offsets.
// Use nil for edges that should not constrain the child.
//
// Positioning rules:
//   - If both Top and Bottom are set, the child's height is computed as stack height - top - bottom
//   - If both Left and Right are set, the child's width is computed as stack width - left - right
//   - Otherwise, the child sizes naturally and is positioned from the specified edges
type Positioned struct {
	Top    *int   // Offset from top edge (nil = not constrained)
	Right  *int   // Offset from right edge (nil = not constrained)
	Bottom *int   // Offset from bottom edge (nil = not constrained)
	Left   *int   // Offset from left edge (nil = not constrained)
	Child  Widget // The child widget to position
}

// PositionedFill creates a Positioned that fills the entire Stack.
func PositionedFill(child Widget) Positioned {
	zero := 0
	return Positioned{Top: &zero, Right: &zero, Bottom: &zero, Left: &zero, Child: child}
}

// PositionedAt creates a Positioned at specific top-left offsets.
func PositionedAt(top, left int, child Widget) Positioned {
	return Positioned{Top: &top, Left: &left, Child: child}
}

// Build returns itself as Positioned is handled specially by Stack.
func (p Positioned) Build(ctx BuildContext) Widget {
	return p
}

// Stack overlays children on top of each other in z-order.
// First child is at the bottom, last child is on top.
//
// Children can be:
//   - Regular widgets: positioned using the Stack's Alignment
//   - Positioned wrappers: positioned using edge offsets
//
// Stack sizes itself based on the largest non-positioned child.
// Positioned children do not affect Stack's size.
type Stack struct {
	ID        string    // Optional unique identifier for the widget
	Children  []Widget  // Children to overlay (first at bottom, last on top)
	Alignment Alignment // Default alignment for non-positioned children (default: top-start)
	Width     Dimension // Deprecated: use Style.Width
	Height    Dimension // Deprecated: use Style.Height
	Style     Style     // Optional styling
	Click     func(MouseEvent) // Optional callback invoked when clicked
	MouseDown func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent) // Optional callback invoked when mouse is released
	Hover     func(bool)       // Optional callback invoked when hover state changes
}

// GetContentDimensions returns the width and height dimension preferences.
func (s Stack) GetContentDimensions() (width, height Dimension) {
	dims := s.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	return width, height
}

// GetStyle returns the style of the stack.
func (s Stack) GetStyle() Style {
	return s.Style
}

// WidgetID returns the stack's unique identifier.
func (s Stack) WidgetID() string {
	return s.ID
}

// OnClick is called when the widget is clicked.
func (s Stack) OnClick(event MouseEvent) {
	if s.Click != nil {
		s.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (s Stack) OnMouseDown(event MouseEvent) {
	if s.MouseDown != nil {
		s.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (s Stack) OnMouseUp(event MouseEvent) {
	if s.MouseUp != nil {
		s.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
func (s Stack) OnHover(hovered bool) {
	if s.Hover != nil {
		s.Hover(hovered)
	}
}

// Build returns itself as Stack manages its own children.
func (s Stack) Build(ctx BuildContext) Widget {
	return s
}

// BuildLayoutNode creates a StackNode for the layout system.
func (s Stack) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	children := make([]layout.StackChild, len(s.Children))

	for i, child := range s.Children {
		childCtx := ctx.PushChild(i)
		built := child.Build(childCtx)

		// Check if this is a Positioned wrapper
		var stackChild layout.StackChild
		if positioned, ok := built.(Positioned); ok {
			// Build the inner child's layout node
			innerBuilt := positioned.Child.Build(childCtx)
			var childNode layout.LayoutNode
			if builder, ok := innerBuilt.(LayoutNodeBuilder); ok {
				childNode = builder.BuildLayoutNode(childCtx)
			} else {
				childNode = buildFallbackLayoutNode(innerBuilt, childCtx)
			}

			// Wrap in PercentNode for width/height if the inner child has percent dimensions
			childNode = wrapInPercentNodesForStack(childNode, innerBuilt)

			stackChild = layout.StackChild{
				Node:         childNode,
				IsPositioned: true,
				Top:          positioned.Top,
				Right:        positioned.Right,
				Bottom:       positioned.Bottom,
				Left:         positioned.Left,
			}
		} else {
			// Regular child - will use Stack's alignment
			var childNode layout.LayoutNode
			if builder, ok := built.(LayoutNodeBuilder); ok {
				childNode = builder.BuildLayoutNode(childCtx)
			} else {
				childNode = buildFallbackLayoutNode(built, childCtx)
			}

			// Wrap in PercentNode for width/height if child has percent dimensions
			childNode = wrapInPercentNodesForStack(childNode, built)

			stackChild = layout.StackChild{
				Node:         childNode,
				IsPositioned: false,
			}
		}

		children[i] = stackChild
	}

	padding := toLayoutEdgeInsets(s.Style.Padding)
	border := borderToEdgeInsets(s.Style.Border)
	dims := GetWidgetDimensionSet(s)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.StackNode{
		Children:      children,
		DefaultHAlign: toLayoutHAlign(s.Alignment.Horizontal),
		DefaultVAlign: toLayoutVAlign(s.Alignment.Vertical),
		Padding:       padding,
		Border:        border,
		Margin:        toLayoutEdgeInsets(s.Style.Margin),
		MinWidth:      minWidth,
		MaxWidth:      maxWidth,
		MinHeight:     minHeight,
		MaxHeight:     maxHeight,
		ExpandWidth:   dims.Width.IsFlex(),
		ExpandHeight:  dims.Height.IsFlex(),
	})

	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{
			child:     node,
			minWidth:  dims.MinWidth,
			maxWidth:  dims.MaxWidth,
			minHeight: dims.MinHeight,
			maxHeight: dims.MaxHeight,
			padding:   padding,
			border:    border,
		}
	}

	return node
}

// Render is a no-op for Stack when using the RenderTree-based rendering path.
func (s Stack) Render(ctx *RenderContext) {
	// No-op: children are positioned by renderTree() using ComputedLayout.Children
}

// AllChildren returns all child widgets, unwrapping Positioned to get inner children.
// This is used by the render tree to match widgets with their computed layouts.
func (s Stack) AllChildren() []Widget {
	children := make([]Widget, len(s.Children))
	for i, child := range s.Children {
		if positioned, ok := child.(Positioned); ok {
			children[i] = positioned.Child
		} else {
			children[i] = child
		}
	}
	return children
}

// toLayoutHAlign converts terma.HorizontalAlignment to layout.HorizontalAlignment.
func toLayoutHAlign(a HorizontalAlignment) layout.HorizontalAlignment {
	switch a {
	case HAlignCenter:
		return layout.HAlignCenter
	case HAlignEnd:
		return layout.HAlignEnd
	default: // HAlignStart
		return layout.HAlignStart
	}
}

// toLayoutVAlign converts terma.VerticalAlignment to layout.VerticalAlignment.
func toLayoutVAlign(a VerticalAlignment) layout.VerticalAlignment {
	switch a {
	case VAlignCenter:
		return layout.VAlignCenter
	case VAlignBottom:
		return layout.VAlignBottom
	default: // VAlignTop
		return layout.VAlignTop
	}
}
