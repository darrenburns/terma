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
// It tracks the widget instance and dirty state for signal subscriptions.
type widgetNode struct {
	widget Widget
	dirty  bool
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
