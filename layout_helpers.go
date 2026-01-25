package terma

import (
	"terma/layout"
)

// toLayoutEdgeInsets converts terma.EdgeInsets to layout.EdgeInsets.
func toLayoutEdgeInsets(e EdgeInsets) layout.EdgeInsets {
	return layout.EdgeInsets{
		Top:    e.Top,
		Right:  e.Right,
		Bottom: e.Bottom,
		Left:   e.Left,
	}
}

// borderToEdgeInsets converts a Border to layout.EdgeInsets based on border width.
func borderToEdgeInsets(b Border) layout.EdgeInsets {
	w := b.Width()
	return layout.EdgeInsetsAll(w)
}

// toLayoutWrapMode converts terma.WrapMode to layout.WrapMode.
func toLayoutWrapMode(w WrapMode) layout.WrapMode {
	switch w {
	case WrapNone:
		return layout.WrapNone
	case WrapHard:
		return layout.WrapChar
	default: // WrapSoft
		return layout.WrapWord
	}
}

// toLayoutMainAlign converts terma.MainAxisAlign to layout.MainAxisAlignment.
func toLayoutMainAlign(a MainAxisAlign) layout.MainAxisAlignment {
	switch a {
	case MainAxisCenter:
		return layout.MainAxisCenter
	case MainAxisEnd:
		return layout.MainAxisEnd
	default: // MainAxisStart
		return layout.MainAxisStart
	}
}

// toLayoutCrossAlign converts terma.CrossAxisAlign to layout.CrossAxisAlignment.
func toLayoutCrossAlign(a CrossAxisAlign) layout.CrossAxisAlignment {
	switch a {
	case CrossAxisStart:
		return layout.CrossAxisStart
	case CrossAxisCenter:
		return layout.CrossAxisCenter
	case CrossAxisEnd:
		return layout.CrossAxisEnd
	default: // CrossAxisStretch
		return layout.CrossAxisStretch
	}
}

// dimensionToMinMax converts a terma Dimension to min/max constraints.
// For Cells (fixed), both min and max are set to the value.
// For Auto/Flex/Percent, returns 0,0 (no constraints from dimension).
func dimensionToMinMax(d Dimension) (min, max int) {
	if d.IsCells() {
		v := d.CellsValue()
		return v, v
	}
	return 0, 0
}

func dimensionToCells(d Dimension) int {
	if d.IsCells() {
		return d.CellsValue()
	}
	return 0
}

func clampFixedDimension(value, minValue, maxValue int) int {
	if minValue > 0 && maxValue > 0 && maxValue < minValue {
		return minValue
	}
	if minValue > 0 && value < minValue {
		value = minValue
	}
	if maxValue > 0 && value > maxValue {
		value = maxValue
	}
	return value
}

func dimensionSetToMinMax(ds DimensionSet, padding, border layout.EdgeInsets) (minW, maxW, minH, maxH int) {
	explicitMinW := dimensionToCells(ds.MinWidth)
	explicitMaxW := dimensionToCells(ds.MaxWidth)
	explicitMinH := dimensionToCells(ds.MinHeight)
	explicitMaxH := dimensionToCells(ds.MaxHeight)

	if ds.Width.IsCells() {
		width := ds.Width.CellsValue()
		if explicitMinW > 0 || explicitMaxW > 0 {
			width = clampFixedDimension(width, explicitMinW, explicitMaxW)
		}
		minW, maxW = width, width
	} else {
		if explicitMinW > 0 {
			minW = explicitMinW
		}
		if explicitMaxW > 0 {
			maxW = explicitMaxW
		}
	}

	if ds.Height.IsCells() {
		height := ds.Height.CellsValue()
		if explicitMinH > 0 || explicitMaxH > 0 {
			height = clampFixedDimension(height, explicitMinH, explicitMaxH)
		}
		minH, maxH = height, height
	} else {
		if explicitMinH > 0 {
			minH = explicitMinH
		}
		if explicitMaxH > 0 {
			maxH = explicitMaxH
		}
	}

	hInset := padding.Horizontal() + border.Horizontal()
	vInset := padding.Vertical() + border.Vertical()
	if minW > 0 {
		minW += hInset
	}
	if maxW > 0 {
		maxW += hInset
	}
	if minH > 0 {
		minH += vInset
	}
	if maxH > 0 {
		maxH += vInset
	}

	return minW, maxW, minH, maxH
}

func dimensionToPercentConstraint(d Dimension, maxAvailable int) int {
	if d.IsPercent() {
		return int(float64(maxAvailable) * d.PercentValue() / 100.0)
	}
	return 0
}

func hasPercentMinMax(ds DimensionSet) bool {
	return ds.MinWidth.IsPercent() || ds.MaxWidth.IsPercent() ||
		ds.MinHeight.IsPercent() || ds.MaxHeight.IsPercent()
}

type percentConstraintWrapper struct {
	child                 layout.LayoutNode
	minWidth, maxWidth     Dimension
	minHeight, maxHeight   Dimension
	padding, border        layout.EdgeInsets
}

func (p *percentConstraintWrapper) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	minW := dimensionToPercentConstraint(p.minWidth, constraints.MaxWidth)
	maxW := dimensionToPercentConstraint(p.maxWidth, constraints.MaxWidth)
	minH := dimensionToPercentConstraint(p.minHeight, constraints.MaxHeight)
	maxH := dimensionToPercentConstraint(p.maxHeight, constraints.MaxHeight)

	if minW == 0 && maxW == 0 && minH == 0 && maxH == 0 {
		return p.child.ComputeLayout(constraints)
	}

	hInset := p.padding.Horizontal() + p.border.Horizontal()
	vInset := p.padding.Vertical() + p.border.Vertical()
	if minW > 0 {
		minW += hInset
	}
	if maxW > 0 {
		maxW += hInset
	}
	if minH > 0 {
		minH += vInset
	}
	if maxH > 0 {
		maxH += vInset
	}

	effective := constraints.WithNodeConstraints(minW, maxW, minH, maxH)
	return p.child.ComputeLayout(effective)
}

// wrapInFlexIfNeeded wraps a layout node in FlexNode if the dimension is Flex().
// This is used when building layout trees from widgets - children with Flex dimensions
// on the main axis should be wrapped in FlexNode so LinearNode can distribute space.
//
// Parameters:
//   - node: The layout node to potentially wrap
//   - mainAxisDim: The dimension on the main axis (Width for Row, Height for Column)
//
// Returns:
//   - The original node if mainAxisDim is not Flex()
//   - A FlexNode wrapping the original if mainAxisDim is Flex()
func wrapInFlexIfNeeded(node layout.LayoutNode, mainAxisDim Dimension) layout.LayoutNode {
	if mainAxisDim.IsFlex() {
		return &layout.FlexNode{
			Flex:  mainAxisDim.FlexValue(),
			Child: node,
		}
	}
	return node
}

// wrapInPercentIfNeeded wraps a layout node in PercentNode if the dimension is Percent().
// This is used when building layout trees from widgets - children with Percent dimensions
// on the main axis should be wrapped in PercentNode so the percentage can be resolved
// from the parent's constraints.
//
// Parameters:
//   - node: The layout node to potentially wrap
//   - mainAxisDim: The dimension on the main axis (Width for Row, Height for Column)
//   - axis: The layout axis (Horizontal for Row, Vertical for Column)
//
// Returns:
//   - The original node if mainAxisDim is not Percent()
//   - A PercentNode wrapping the original if mainAxisDim is Percent()
func wrapInPercentIfNeeded(node layout.LayoutNode, mainAxisDim Dimension, axis layout.Axis) layout.LayoutNode {
	if mainAxisDim.IsPercent() {
		return &layout.PercentNode{
			Percent: mainAxisDim.PercentValue(),
			Child:   node,
			Axis:    axis,
		}
	}
	return node
}

// getChildMainAxisDimension returns the main-axis dimension for a widget.
// For horizontal (Row): returns Width
// For vertical (Column): returns Height
func getChildMainAxisDimension(widget Widget, horizontal bool) Dimension {
	dims := GetWidgetDimensionSet(widget)
	if horizontal {
		return dims.Width
	}
	return dims.Height
}

// wrapInPercentNodesForStack wraps a layout node in PercentNode(s) for Stack children.
// Unlike Row/Column which have a single main axis, Stack children can have percent
// dimensions on both width and height independently.
//
// Parameters:
//   - node: The layout node to potentially wrap
//   - widget: The widget to check for percent dimensions
//
// Returns:
//   - The original node if no percent dimensions
//   - A PercentNode (or nested PercentNodes) wrapping the original if percent dimensions exist
func wrapInPercentNodesForStack(node layout.LayoutNode, widget Widget) layout.LayoutNode {
	dims := GetWidgetDimensionSet(widget)

	// Wrap for width percent first
	if dims.Width.IsPercent() {
		node = &layout.PercentNode{
			Percent: dims.Width.PercentValue(),
			Child:   node,
			Axis:    layout.Horizontal,
		}
	}

	// Wrap for height percent (wraps around any width percent wrapper)
	if dims.Height.IsPercent() {
		node = &layout.PercentNode{
			Percent: dims.Height.PercentValue(),
			Child:   node,
			Axis:    layout.Vertical,
		}
	}

	return node
}

// buildFallbackLayoutNode creates a BoxNode for widgets that don't implement LayoutNodeBuilder.
// It uses the widget's Dimensioned and Styled interfaces to extract dimensions and insets.
//
// Content dimensions from GetContentDimensions() are converted to border-box constraints
// by adding padding and border. This allows widgets to specify their content size without
// worrying about the box model - the framework handles adding space for decoration.
func buildFallbackLayoutNode(widget Widget, ctx BuildContext) layout.LayoutNode {
	dims := GetWidgetDimensionSet(widget)

	// Extract style for insets first - we need these to compute border-box dimensions
	var padding, border, margin layout.EdgeInsets
	if styled, ok := widget.(Styled); ok {
		style := styled.GetStyle()
		padding = toLayoutEdgeInsets(style.Padding)
		border = borderToEdgeInsets(style.Border)
		margin = toLayoutEdgeInsets(style.Margin)
	}

	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.BoxNode{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		Padding:   padding,
		Border:    border,
		Margin:    margin,
		ExpandWidth:  dims.Width.IsFlex() || dims.Width.IsPercent(),
		ExpandHeight: dims.Height.IsFlex() || dims.Height.IsPercent(),
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
