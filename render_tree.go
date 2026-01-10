package terma

import "terma/layout"

// RenderTree pairs a widget with its computed layout for rendering.
// This enables complete separation of layout and rendering phases:
// 1. Build phase: widget.Build() creates widget tree
// 2. Layout phase: BuildRenderTree() computes all positions
// 3. Render phase: renderTree() paints using computed geometry
type RenderTree struct {
	// Widget is the built widget (result of Build()).
	// Used for calling Render() and extracting style.
	Widget Widget

	// Layout is the computed geometry for this widget.
	// Contains BoxModel with all dimensions and insets.
	Layout layout.ComputedLayout

	// Children are the child render trees.
	// Positions come from Layout.Children[i].X and Layout.Children[i].Y.
	Children []RenderTree
}

// BuildRenderTree constructs the complete render tree with all layout computed.
// Focus collection also happens here, keeping the render phase pure painting.
func BuildRenderTree(widget Widget, ctx BuildContext, constraints layout.Constraints, fc *FocusCollector) RenderTree {
	built := widget.Build(ctx)

	// Collect focusables during tree build (not during render)
	if fc != nil {
		fc.Collect(widget, ctx.AutoID())
		if fc.ShouldTrackAncestor(widget) {
			fc.PushAncestor(widget)
			defer fc.PopAncestor()
		}
	}

	// Build layout node and compute layout
	var computed layout.ComputedLayout
	if builder, ok := built.(LayoutNodeBuilder); ok {
		node := builder.BuildLayoutNode(ctx)
		computed = node.ComputeLayout(constraints)
	} else {
		// Fallback for widgets without LayoutNodeBuilder
		// Use existing Layoutable interface to get size, create minimal BoxModel
		computed = layoutFromLayoutable(built, ctx, constraints)
	}

	// Recursively build children
	children := buildChildTrees(built, ctx, computed, fc)

	return RenderTree{
		Widget:   built,
		Layout:   computed,
		Children: children,
	}
}

// buildChildTrees recursively builds RenderTrees for all children.
func buildChildTrees(widget Widget, ctx BuildContext, computed layout.ComputedLayout, fc *FocusCollector) []RenderTree {
	// Extract children from the widget
	widgetChildren := extractChildren(widget)

	// If no layout children or no widget children, return empty
	if len(widgetChildren) == 0 || len(computed.Children) == 0 {
		return nil
	}

	trees := make([]RenderTree, len(widgetChildren))

	for i, child := range widgetChildren {
		if i >= len(computed.Children) {
			break
		}
		pos := computed.Children[i]
		childCtx := ctx.PushChild(i)

		// Child constraints are tight to computed size
		childConstraints := layout.Tight(
			pos.Layout.Box.BorderBoxWidth(),
			pos.Layout.Box.BorderBoxHeight(),
		)

		trees[i] = BuildRenderTree(child, childCtx, childConstraints, fc)
	}
	return trees
}

// extractChildren extracts the child widgets from a container widget.
func extractChildren(widget Widget) []Widget {
	switch w := widget.(type) {
	case Row:
		return w.Children
	case Column:
		return w.Children
	case Scrollable:
		if w.Child != nil {
			return []Widget{w.Child}
		}
		return nil
	// Add other container types as they implement LayoutNodeBuilder
	default:
		return nil
	}
}

// layoutFromLayoutable creates a ComputedLayout for widgets that don't implement LayoutNodeBuilder.
// This provides fallback support so widgets can be migrated incrementally.
func layoutFromLayoutable(widget Widget, ctx BuildContext, constraints layout.Constraints) layout.ComputedLayout {
	// Extract style for insets
	var style Style
	if styled, ok := widget.(Styled); ok {
		style = styled.GetStyle()
	}

	// Use existing Layoutable interface to get size
	width, height := constraints.MaxWidth, constraints.MaxHeight
	if layoutable, ok := widget.(Layoutable); ok {
		size := layoutable.Layout(ctx, Constraints{
			MinWidth:  constraints.MinWidth,
			MaxWidth:  constraints.MaxWidth,
			MinHeight: constraints.MinHeight,
			MaxHeight: constraints.MaxHeight,
		})
		width, height = size.Width, size.Height
	}

	// Create BoxModel from style + computed size
	return layout.ComputedLayout{
		Box: layout.BoxModel{
			Width:   width,
			Height:  height,
			Padding: toLayoutEdgeInsets(style.Padding),
			Border:  borderToEdgeInsets(style.Border),
			Margin:  toLayoutEdgeInsets(style.Margin),
		},
		Children: nil, // Fallback widgets handle their own children in Render()
	}
}
