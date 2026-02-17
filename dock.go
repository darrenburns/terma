package terma

import "github.com/darrenburns/terma/layout"

// Edge specifies which edge a child is docked to.
// This is a re-export of layout.DockEdge for widget API.
type Edge = layout.DockEdge

// Edge constants for dock order.
const (
	Top    = layout.DockTop
	Bottom = layout.DockBottom
	Left   = layout.DockLeft
	Right  = layout.DockRight
)

// Dock arranges children by docking them to edges.
// Works like WPF DockPanel: edges consume space in order, body fills remainder.
//
// Example:
//
//	Dock{
//	    Top:    []Widget{Header{}},
//	    Bottom: []Widget{KeybindBar{}},
//	    Left:   []Widget{Sidebar{}},
//	    Body:   MainContent{},
//	}
type Dock struct {
	ID        string    // Optional unique identifier
	Top       []Widget  // Widgets docked to top edge
	Bottom    []Widget  // Widgets docked to bottom edge
	Left      []Widget  // Widgets docked to left edge
	Right     []Widget  // Widgets docked to right edge
	Body      Widget    // Widget that fills remaining space
	DockOrder []Edge    // Order in which edges are processed (default: Top, Bottom, Left, Right)
	Width     Dimension // Deprecated: use Style.Width
	Height    Dimension // Deprecated: use Style.Height
	Style     Style     // Optional styling
}

// GetContentDimensions returns dimensions (defaults to Flex(1) for both).
func (d Dock) GetContentDimensions() (Dimension, Dimension) {
	dims := d.Style.GetDimensions()
	width, height := dims.Width, dims.Height
	if width.IsUnset() {
		width = d.Width
	}
	if height.IsUnset() {
		height = d.Height
	}
	if width.IsUnset() {
		width = Flex(1)
	}
	if height.IsUnset() {
		height = Flex(1)
	}
	return width, height
}

// GetStyle returns the dock's style.
func (d Dock) GetStyle() Style {
	return d.Style
}

// WidgetID returns the dock's identifier.
func (d Dock) WidgetID() string {
	return d.ID
}

// Build returns self (Dock manages its own children).
func (d Dock) Build(ctx BuildContext) Widget {
	return d
}

// Render is a no-op (children positioned by renderTree).
func (d Dock) Render(ctx *RenderContext) {}

// BuildLayoutNode creates a DockNode for the layout system.
func (d Dock) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	childIndex := 0

	// Convert edge children in dock order
	// We must process edges in the same order as AllChildren() returns them
	order := d.dockOrder()

	var top, bottom, left, right []layout.LayoutNode

	for _, edge := range order {
		switch edge {
		case Top:
			top = d.buildEdgeChildren(ctx, d.Top, &childIndex, edge)
		case Bottom:
			bottom = d.buildEdgeChildren(ctx, d.Bottom, &childIndex, edge)
		case Left:
			left = d.buildEdgeChildren(ctx, d.Left, &childIndex, edge)
		case Right:
			right = d.buildEdgeChildren(ctx, d.Right, &childIndex, edge)
		}
	}

	// Convert body
	var body layout.LayoutNode
	if d.Body != nil {
		bodyCtx := ctx.PushChild(childIndex)
		built := d.Body.Build(bodyCtx)
		if builder, ok := built.(LayoutNodeBuilder); ok {
			body = builder.BuildLayoutNode(bodyCtx)
		} else {
			body = buildFallbackLayoutNode(built, bodyCtx)
		}
	}

	padding := toLayoutEdgeInsets(d.Style.Padding)
	border := borderToEdgeInsets(d.Style.Border)
	dims := GetWidgetDimensionSet(d)
	minW, maxW, minH, maxH := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.DockNode{
		Top:       top,
		Bottom:    bottom,
		Left:      left,
		Right:     right,
		Body:      body,
		DockOrder: d.DockOrder,
		Padding:   padding,
		Border:    border,
		Margin:    toLayoutEdgeInsets(d.Style.Margin),
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
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

// buildEdgeChildren converts widgets for a dock edge.
func (d Dock) buildEdgeChildren(ctx BuildContext, widgets []Widget, index *int, edge Edge) []layout.LayoutNode {
	nodes := make([]layout.LayoutNode, len(widgets))
	for i, child := range widgets {
		childCtx := ctx.PushChild(*index)
		built := child.Build(childCtx)
		var node layout.LayoutNode
		if builder, ok := built.(LayoutNodeBuilder); ok {
			node = builder.BuildLayoutNode(childCtx)
		} else {
			node = buildFallbackLayoutNode(built, childCtx)
		}

		// Wrap in PercentNode if the child has a percentage dimension on the relevant axis
		// Top/Bottom: check Height (Vertical), Left/Right: check Width (Horizontal)
		dims := GetWidgetDimensionSet(built)
		switch edge {
		case Top, Bottom:
			if dims.Height.IsPercent() {
				node = &layout.PercentNode{
					Percent: dims.Height.PercentValue(),
					Child:   node,
					Axis:    layout.Vertical,
				}
			}
		case Left, Right:
			if dims.Width.IsPercent() {
				node = &layout.PercentNode{
					Percent: dims.Width.PercentValue(),
					Child:   node,
					Axis:    layout.Horizontal,
				}
			}
		}

		nodes[i] = node
		*index++
	}
	return nodes
}

// dockOrder returns the edge processing order.
func (d Dock) dockOrder() []Edge {
	if len(d.DockOrder) > 0 {
		return d.DockOrder
	}
	return []Edge{Top, Bottom, Left, Right}
}

// AllChildren returns all children in layout order (for extractChildren).
// The order matches how BuildLayoutNode processes edges plus body.
func (d Dock) AllChildren() []Widget {
	var all []Widget
	order := d.dockOrder()
	for _, edge := range order {
		switch edge {
		case Top:
			all = append(all, d.Top...)
		case Bottom:
			all = append(all, d.Bottom...)
		case Left:
			all = append(all, d.Left...)
		case Right:
			all = append(all, d.Right...)
		}
	}
	if d.Body != nil {
		all = append(all, d.Body)
	}
	return all
}
