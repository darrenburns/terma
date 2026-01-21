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

func wrapWithDimensionConstraints(
	node layout.LayoutNode,
	width, height, minWidth, maxWidth, minHeight, maxHeight Dimension,
	padding, border layout.EdgeInsets,
) layout.LayoutNode {
	if !needsDimensionConstraints(width, height, minWidth, maxWidth, minHeight, maxHeight) {
		return node
	}

	return &dimensionConstraintNode{
		Child:     node,
		Width:     width,
		Height:    height,
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		HInset:    padding.Horizontal() + border.Horizontal(),
		VInset:    padding.Vertical() + border.Vertical(),
	}
}

func needsDimensionConstraints(width, height, minWidth, maxWidth, minHeight, maxHeight Dimension) bool {
	if !minWidth.IsAuto() || !maxWidth.IsAuto() || !minHeight.IsAuto() || !maxHeight.IsAuto() {
		return true
	}
	if width.IsCells() || width.IsPercent() || width.IsFlex() {
		return true
	}
	if height.IsCells() || height.IsPercent() || height.IsFlex() {
		return true
	}
	return false
}

// applyDimensionConstraints resolves Width/Height and Min/Max dimensions into concrete constraints.
// Cells are content-box sizes and are converted to border-box using insets. Percent/Flex are treated
// as border-box sizes relative to the parent's available space.
func applyDimensionConstraints(
	parent layout.Constraints,
	width, height, minWidth, maxWidth, minHeight, maxHeight Dimension,
	hInset, vInset int,
) layout.Constraints {
	maxAvailableW := parent.MaxWidth
	maxAvailableH := parent.MaxHeight

	resolvedMinW := resolveConstraintDimension(minWidth, maxAvailableW, hInset)
	resolvedMaxW := resolveConstraintDimension(maxWidth, maxAvailableW, hInset)
	resolvedMinH := resolveConstraintDimension(minHeight, maxAvailableH, vInset)
	resolvedMaxH := resolveConstraintDimension(maxHeight, maxAvailableH, vInset)

	if resolvedMinW > 0 && resolvedMaxW > 0 && resolvedMinW > resolvedMaxW {
		resolvedMaxW = resolvedMinW
	}
	if resolvedMinH > 0 && resolvedMaxH > 0 && resolvedMinH > resolvedMaxH {
		resolvedMaxH = resolvedMinH
	}

	explicitW, hasExplicitW := resolveExplicitDimension(width, maxAvailableW, hInset)
	explicitH, hasExplicitH := resolveExplicitDimension(height, maxAvailableH, vInset)

	if hasExplicitW && (width.IsPercent() || width.IsFlex()) && parent.MinWidth == parent.MaxWidth {
		hasExplicitW = false
	}
	if hasExplicitH && (height.IsPercent() || height.IsFlex()) && parent.MinHeight == parent.MaxHeight {
		hasExplicitH = false
	}

	if hasExplicitW {
		if resolvedMinW > 0 && explicitW < resolvedMinW {
			explicitW = resolvedMinW
		}
		if resolvedMaxW > 0 && explicitW > resolvedMaxW {
			explicitW = resolvedMaxW
		}
		resolvedMinW = explicitW
		resolvedMaxW = explicitW
	}

	if hasExplicitH {
		if resolvedMinH > 0 && explicitH < resolvedMinH {
			explicitH = resolvedMinH
		}
		if resolvedMaxH > 0 && explicitH > resolvedMaxH {
			explicitH = resolvedMaxH
		}
		resolvedMinH = explicitH
		resolvedMaxH = explicitH
	}

	effective := parent.WithNodeConstraints(resolvedMinW, resolvedMaxW, resolvedMinH, resolvedMaxH)

	if hasExplicitW {
		if explicitW < 0 {
			explicitW = 0
		}
		if explicitW > effective.MaxWidth {
			explicitW = effective.MaxWidth
		}
		effective.MinWidth = explicitW
		effective.MaxWidth = explicitW
	}

	if hasExplicitH {
		if explicitH < 0 {
			explicitH = 0
		}
		if explicitH > effective.MaxHeight {
			explicitH = effective.MaxHeight
		}
		effective.MinHeight = explicitH
		effective.MaxHeight = explicitH
	}

	return effective
}

func resolveConstraintDimension(d Dimension, maxAvailable, inset int) int {
	if d.IsAuto() {
		return 0
	}
	switch {
	case d.IsCells():
		return addInset(d.CellsValue(), inset)
	case d.IsPercent():
		return int(float64(maxAvailable) * d.PercentValue() / 100.0)
	case d.IsFlex():
		return maxAvailable
	default:
		return 0
	}
}

func resolveExplicitDimension(d Dimension, maxAvailable, inset int) (int, bool) {
	if d.IsAuto() {
		return 0, false
	}
	switch {
	case d.IsCells():
		return addInset(d.CellsValue(), inset), true
	case d.IsPercent():
		return int(float64(maxAvailable) * d.PercentValue() / 100.0), true
	case d.IsFlex():
		return maxAvailable, true
	default:
		return 0, false
	}
}

func addInset(value, inset int) int {
	if value <= 0 {
		return value
	}
	return value + inset
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
	width, height := getWidgetDimensions(widget)
	if horizontal {
		return width
	}
	return height
}

type dimensionConstraintNode struct {
	Child     layout.LayoutNode
	Width     Dimension
	Height    Dimension
	MinWidth  Dimension
	MaxWidth  Dimension
	MinHeight Dimension
	MaxHeight Dimension
	HInset    int
	VInset    int
}

type legacyDimensioned interface {
	GetDimensions() (width, height Dimension)
}

func getWidgetDimensions(widget Widget) (Dimension, Dimension) {
	if dimensioned, ok := widget.(Dimensioned); ok {
		return dimensioned.GetContentDimensions()
	}
	if legacy, ok := widget.(legacyDimensioned); ok {
		return legacy.GetDimensions()
	}
	return Dimension{}, Dimension{}
}

func getWidgetMinMaxDimensions(widget Widget) (Dimension, Dimension, Dimension, Dimension) {
	if constrained, ok := widget.(MinMaxDimensioned); ok {
		return constrained.GetMinMaxDimensions()
	}
	return Dimension{}, Dimension{}, Dimension{}, Dimension{}
}

func getWidgetInsets(widget Widget) (padding, border layout.EdgeInsets) {
	if styled, ok := widget.(Styled); ok {
		style := styled.GetStyle()
		return toLayoutEdgeInsets(style.Padding), borderToEdgeInsets(style.Border)
	}
	return layout.EdgeInsets{}, layout.EdgeInsets{}
}

func (n *dimensionConstraintNode) ComputeLayout(constraints layout.Constraints) layout.ComputedLayout {
	effective := applyDimensionConstraints(
		constraints,
		n.Width,
		n.Height,
		n.MinWidth,
		n.MaxWidth,
		n.MinHeight,
		n.MaxHeight,
		n.HInset,
		n.VInset,
	)
	return n.Child.ComputeLayout(effective)
}

// buildFallbackLayoutNode creates a BoxNode for widgets that don't implement LayoutNodeBuilder.
// It uses the widget's Dimensioned and Styled interfaces to extract dimensions and insets.
//
func buildFallbackLayoutNode(widget Widget, ctx BuildContext) layout.LayoutNode {
	// Extract style for insets first - we need these to compute border-box dimensions
	var padding, border, margin layout.EdgeInsets
	if styled, ok := widget.(Styled); ok {
		style := styled.GetStyle()
		padding = toLayoutEdgeInsets(style.Padding)
		border = borderToEdgeInsets(style.Border)
		margin = toLayoutEdgeInsets(style.Margin)
	}

	node := &layout.BoxNode{
		Padding:   padding,
		Border:    border,
		Margin:    margin,
	}

	if layoutable, ok := widget.(Layoutable); ok {
		node.MeasureFunc = func(constraints layout.Constraints) (int, int) {
			size := layoutable.Layout(ctx, Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			})
			return size.Width, size.Height
		}
	}

	return node
}
