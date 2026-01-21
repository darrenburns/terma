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
//
// GetContentDimensions returns the content-box dimensions - the space needed for
// the widget's content, NOT including padding or border. The framework automatically
// adds padding and border from the widget's Style to compute the final outer size.
//
// For example, a TextInput might return Cells(1) for height (one line of text content),
// and if it has Style{Padding: EdgeInsetsXY(1,1)}, the final height will be 3 cells.
type Dimensioned interface {
	GetContentDimensions() (width, height Dimension)
}

// MinMaxDimensioned is implemented by widgets that expose min/max size preferences.
// These dimensions are resolved against parent constraints at layout time.
type MinMaxDimensioned interface {
	GetMinMaxDimensions() (minWidth, maxWidth, minHeight, maxHeight Dimension)
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

// LayoutObserver is implemented by widgets that want access to computed layout data.
// OnLayout is called after layout is computed for the widget, before child render trees are built.
// Use this to read resolved child positions/sizes without re-measuring.
type LayoutObserver interface {
	OnLayout(ctx BuildContext, metrics LayoutMetrics)
}

// ChildProvider exposes a widget's children for render tree construction.
// Implement this for custom containers so computed child layouts are rendered.
type ChildProvider interface {
	ChildWidgets() []Widget
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
