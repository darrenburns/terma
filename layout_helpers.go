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
// For Auto or Fr, returns 0,0 (no constraints from dimension).
func dimensionToMinMax(d Dimension) (min, max int) {
	if d.IsCells() {
		v := d.CellsValue()
		return v, v
	}
	return 0, 0
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

// getChildMainAxisDimension returns the main-axis dimension for a widget.
// For horizontal (Row): returns Width
// For vertical (Column): returns Height
func getChildMainAxisDimension(widget Widget, horizontal bool) Dimension {
	if dimensioned, ok := widget.(Dimensioned); ok {
		width, height := dimensioned.GetDimensions()
		if horizontal {
			return width
		}
		return height
	}
	return Dimension{} // unset/auto
}

// buildFallbackLayoutNode creates a BoxNode for widgets that don't implement LayoutNodeBuilder.
// It uses the widget's Dimensioned and Styled interfaces to extract dimensions and insets.
func buildFallbackLayoutNode(widget Widget, ctx BuildContext) layout.LayoutNode {
	var widthDim, heightDim Dimension
	if dimensioned, ok := widget.(Dimensioned); ok {
		widthDim, heightDim = dimensioned.GetDimensions()
	}

	// Convert dimensions to min/max constraints
	minWidth, maxWidth := dimensionToMinMax(widthDim)
	minHeight, maxHeight := dimensionToMinMax(heightDim)

	// Extract style for insets
	var padding, border, margin layout.EdgeInsets
	if styled, ok := widget.(Styled); ok {
		style := styled.GetStyle()
		padding = toLayoutEdgeInsets(style.Padding)
		border = borderToEdgeInsets(style.Border)
		margin = toLayoutEdgeInsets(style.Margin)
	}

	return &layout.BoxNode{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		Padding:   padding,
		Border:    border,
		Margin:    margin,
	}
}
