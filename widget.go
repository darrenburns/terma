package terma

import "terma/layout"

// Widget is the base interface for all UI elements.
// Leaf widgets (like Text) return themselves from Build().
// Container widgets return a composed widget tree.
type Widget interface {
	Build(ctx BuildContext) Widget
}

// Constraints define the min/max dimensions a widget can occupy.
type Constraints struct {
	MinWidth, MaxWidth   int
	MinHeight, MaxHeight int
}

// Size represents the computed dimensions of a widget.
type Size struct {
	Width, Height int
}

// Layoutable is implemented by widgets that can compute their size
// given constraints and perform layout on their children.
type Layoutable interface {
	Layout(ctx BuildContext, constraints Constraints) Size
}

// Renderable is implemented by widgets that can render themselves.
type Renderable interface {
	Render(ctx *RenderContext)
}

// Dimensioned is implemented by widgets that have explicit dimension preferences.
// This allows parent containers to query child dimensions for fractional layout.
type Dimensioned interface {
	GetDimensions() (width, height Dimension)
}

// Styled is implemented by widgets that have a Style.
// The framework uses this to extract padding and margin for automatic layout.
type Styled interface {
	GetStyle() Style
}

// LayoutNodeBuilder is implemented by widgets that can build a layout node for themselves.
// This enables integration with the new layout system in the layout package.
type LayoutNodeBuilder interface {
	BuildLayoutNode(ctx BuildContext) layout.LayoutNode
}

// widgetNode is an internal node in the widget tree.
// It tracks the widget instance, layout info, and dirty state.
type widgetNode struct {
	widget   Widget
	dirty    bool
	built    Widget // Result of calling Build()
	children []*widgetNode

	// Layout info
	x, y          int
	width, height int
}

// newWidgetNode creates a new widget node.
func newWidgetNode(widget Widget) *widgetNode {
	return &widgetNode{
		widget: widget,
		dirty:  true,
	}
}

// markDirty marks this node for rebuild.
func (n *widgetNode) markDirty() {
	n.dirty = true
}

// build rebuilds this node if dirty, tracking signal subscriptions.
func (n *widgetNode) build(ctx BuildContext) Widget {
	if !n.dirty && n.built != nil {
		return n.built
	}

	// Set this node as the current building node so signals can subscribe
	previousNode := currentBuildingNode
	currentBuildingNode = n

	// Call Build() on the widget
	n.built = n.widget.Build(ctx)

	// Restore previous node
	currentBuildingNode = previousNode

	n.dirty = false
	return n.built
}

// layout computes the size and position of this node and its children.
func (n *widgetNode) layout(ctx BuildContext, constraints Constraints, x, y int) Size {
	n.x = x
	n.y = y

	// Build the widget first
	built := n.build(ctx)

	// If the built widget implements Layoutable, use it
	if layoutable, ok := built.(Layoutable); ok {
		size := layoutable.Layout(ctx, constraints)
		n.width = size.Width
		n.height = size.Height
		return size
	}

	// Default size
	n.width = constraints.MaxWidth
	n.height = constraints.MaxHeight
	return Size{Width: n.width, Height: n.height}
}
