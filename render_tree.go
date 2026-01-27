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

	// EventWidget is the original widget (before Build()).
	// Used for hit testing and event dispatch.
	EventWidget Widget

	// EventID is the ID used for hit testing and focus (explicit or auto).
	EventID string

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
	// Handle disabledWrapper specially - recurse into child with disabled context
	// This makes the wrapper completely transparent to layout and focus collection
	if dw, ok := widget.(disabledWrapper); ok {
		if dw.disabled {
			ctx = ctx.WithDisabled()
		}
		return BuildRenderTree(dw.child, ctx, constraints, fc)
	}

	// Handle FocusTrap specially - recurse into child with trap scope pushed.
	// This makes FocusTrap transparent to layout while correctly scoping focus.
	if ft, ok := widget.(FocusTrap); ok {
		if fc != nil && ft.TrapsFocus() {
			trapID := ft.WidgetID()
			if trapID == "" {
				trapID = ctx.AutoID()
			}
			fc.PushTrap(trapID)
			defer fc.PopTrap()
		}
		if ft.Child == nil {
			return BuildRenderTree(EmptyWidget{}, ctx, constraints, fc)
		}
		return BuildRenderTree(ft.Child, ctx, constraints, fc)
	}

	autoID := ctx.AutoID()

	// Determine event ID (explicit ID or auto)
	eventID := autoID
	if identifiable, ok := widget.(Identifiable); ok && identifiable.WidgetID() != "" {
		eventID = identifiable.WidgetID()
	}

	built := widget.Build(ctx)

	// Collect focusables during tree build (not during render)
	if fc != nil {
		if trapper, ok := widget.(FocusTrapper); ok && trapper.TrapsFocus() {
			fc.PushTrap(eventID)
			defer fc.PopTrap()
		}
		fc.Collect(widget, ctx.AutoID(), ctx)
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

	// Provide computed layout to observers before building child render trees.
	if observer, ok := built.(LayoutObserver); ok {
		observer.OnLayout(ctx, LayoutMetrics{layout: computed})
	}

	// Recursively build children
	children := buildChildTrees(built, ctx, computed, fc)

	return RenderTree{
		Widget:      built,
		EventWidget: widget,
		EventID:     eventID,
		Layout:      computed,
		Children:    children,
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
	if provider, ok := widget.(ChildProvider); ok {
		return provider.ChildWidgets()
	}
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
	case Dock:
		return w.AllChildren()
	case Stack:
		return w.AllChildren()
	case Switcher:
		if child, ok := w.Children[w.Active]; ok {
			return []Widget{child}
		}
		return nil
	default:
		return nil
	}
}

// layoutFromLayoutable creates a ComputedLayout for widgets that don't implement LayoutNodeBuilder.
// This provides fallback support so widgets can be migrated incrementally.
//
// Important: Layout() returns content-box dimensions (content size only, not including padding/border).
// This function converts those to border-box for the BoxModel by adding padding and border.
func layoutFromLayoutable(widget Widget, ctx BuildContext, constraints layout.Constraints) layout.ComputedLayout {
	// Extract style for insets
	var style Style
	if styled, ok := widget.(Styled); ok {
		style = styled.GetStyle()
	}

	padding := toLayoutEdgeInsets(style.Padding)
	border := borderToEdgeInsets(style.Border)
	margin := toLayoutEdgeInsets(style.Margin)

	// Compute insets to convert between content-box and border-box
	hInset := padding.Horizontal() + border.Horizontal()
	vInset := padding.Vertical() + border.Vertical()

	// Use existing Layoutable interface to get content size
	// Pass content-box constraints (subtract insets from parent constraints)
	contentWidth, contentHeight := max(0, constraints.MaxWidth-hInset), max(0, constraints.MaxHeight-vInset)
	if layoutable, ok := widget.(Layoutable); ok {
		size := layoutable.Layout(ctx, Constraints{
			MinWidth:  max(0, constraints.MinWidth-hInset),
			MaxWidth:  max(0, constraints.MaxWidth-hInset),
			MinHeight: max(0, constraints.MinHeight-vInset),
			MaxHeight: max(0, constraints.MaxHeight-vInset),
		})
		contentWidth, contentHeight = size.Width, size.Height
	}

	// Create BoxModel with border-box dimensions (content + padding + border)
	return layout.ComputedLayout{
		Box: layout.BoxModel{
			Width:   contentWidth + hInset,
			Height:  contentHeight + vInset,
			Padding: padding,
			Border:  border,
			Margin:  margin,
		},
		Children: nil, // Fallback widgets handle their own children in Render()
	}
}
