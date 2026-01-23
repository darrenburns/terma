package terma

import (
	"sync/atomic"

	"terma/layout"
)

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
	dirty  atomic.Bool
}

// newWidgetNode creates a new widget node.
func newWidgetNode(widget Widget) *widgetNode {
	node := &widgetNode{
		widget: widget,
	}
	node.dirty.Store(true)
	return node
}

// markDirty marks this node for rebuild.
// Thread-safe: can be called from any goroutine.
func (n *widgetNode) markDirty() {
	n.dirty.Store(true)
}

// isDirty returns whether this node needs rebuild.
// Thread-safe.
func (n *widgetNode) isDirty() bool {
	return n.dirty.Load()
}

// clearDirty marks this node as clean.
// Thread-safe.
func (n *widgetNode) clearDirty() {
	n.dirty.Store(false)
}
