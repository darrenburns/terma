package terma

import "terma/layout"

// Switcher displays one child at a time based on a string key.
// Only the child matching the Active key is rendered; others are not built.
//
// State preservation: Widget state lives in Signals and State objects
// (e.g., ListState, ScrollState) held by the App, not in widgets themselves.
// When switching between children, their state objects persist and the
// widget rebuilds with that state when reactivated.
//
// Example:
//
//	Switcher{
//	    Active: a.activeTab.Get(),
//	    Children: map[string]Widget{
//	        "home":     HomeTab{listState: a.homeListState},
//	        "settings": SettingsTab{},
//	        "profile":  ProfileTab{},
//	    },
//	}
type Switcher struct {
	// Active is the key of the currently visible child.
	Active string

	// Children maps string keys to widgets. Only the widget matching
	// Active is rendered.
	Children map[string]Widget

	// Width specifies the width of the switcher container.
	// Deprecated: use Style.Width.
	Width Dimension

	// Height specifies the height of the switcher container.
	// Deprecated: use Style.Height.
	Height Dimension

	// Style applies styling to the switcher container.
	Style Style
}

// Build returns itself since Switcher manages its own child rendering.
func (s Switcher) Build(ctx BuildContext) Widget {
	return s
}

// GetContentDimensions returns the configured width and height.
func (s Switcher) GetContentDimensions() (Dimension, Dimension) {
	dims := s.Style.GetDimensions()
	width, height := dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	return width, height
}

// GetStyle returns the configured style.
func (s Switcher) GetStyle() Style {
	return s.Style
}

// BuildLayoutNode builds a layout node for this Switcher widget.
// Implements the LayoutNodeBuilder interface.
func (s Switcher) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	// Get the active child
	var child Widget
	if c, ok := s.Children[s.Active]; ok {
		child = c
	} else {
		child = EmptyWidget{}
	}

	// Build the child
	childCtx := ctx.PushChild(0)
	built := child.Build(childCtx)

	// Get child's layout node
	var childNode layout.LayoutNode
	if builder, ok := built.(LayoutNodeBuilder); ok {
		childNode = builder.BuildLayoutNode(childCtx)
	} else {
		childNode = buildFallbackLayoutNode(built, childCtx)
	}

	// Wrap in FlexNode if child has Flex height (Switcher acts as a vertical container)
	mainAxisDim := getChildMainAxisDimension(built, false) // false = vertical axis
	childNode = wrapInPercentIfNeeded(childNode, mainAxisDim, layout.Vertical)
	childNode = wrapInFlexIfNeeded(childNode, mainAxisDim)

	// Get Switcher's own dimensions and style
	style := s.Style
	padding := toLayoutEdgeInsets(style.Padding)
	border := borderToEdgeInsets(style.Border)
	dims := style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = s.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = s.Height
	}
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	// Explicit Auto means "fit content, don't stretch" - set preserve flags
	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	// Create a column node that wraps the single child
	node := layout.LayoutNode(&layout.ColumnNode{
		Children:       []layout.LayoutNode{childNode},
		Padding:        padding,
		Border:         border,
		Margin:         toLayoutEdgeInsets(style.Margin),
		MinWidth:       minWidth,
		MaxWidth:       maxWidth,
		MinHeight:      minHeight,
		MaxHeight:      maxHeight,
		ExpandWidth:    dims.Width.IsFlex(),
		ExpandHeight:   dims.Height.IsFlex(),
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
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
