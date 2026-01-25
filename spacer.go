package terma

import "terma/layout"

// Spacer is a non-visual widget that occupies space in layouts.
// It renders nothing but participates in layout calculations.
//
// Common uses:
//   - Push widgets to opposite ends of a Row or Column
//   - Create flexible gaps that expand to fill available space
//   - Add fixed-size empty regions with Cells()
//
// Default behavior: An unset dimension defaults to Flex(1), so a bare
// Spacer{} expands to fill available space in both directions.
//
// Note: Explicitly setting Auto results in 0 size since Spacer has no
// content to fit. Use Flex(1) instead if you want the spacer to expand.
type Spacer struct {
	Width  Dimension // Defaults to Flex(1) if unset
	Height Dimension // Defaults to Flex(1) if unset

	MinWidth  Dimension
	MaxWidth  Dimension
	MinHeight Dimension
	MaxHeight Dimension
}

// Build returns itself as Spacer is a leaf widget.
func (s Spacer) Build(ctx BuildContext) Widget {
	return s
}

// GetContentDimensions returns the width and height dimension preferences.
// If both dimensions are unset, both default to Flex(1) (bare Spacer{} expands in both directions).
// If one dimension is explicitly set to Flex/Percent, the other defaults to Auto (size 0),
// making Spacer{Width: Flex(1)} behave as a horizontal-only spacer.
func (s Spacer) GetContentDimensions() (width, height Dimension) {
	w, h := s.Width, s.Height

	// Both unset: default both to Flex(1) for bare Spacer{}
	if w.IsUnset() && h.IsUnset() {
		return Flex(1), Flex(1)
	}

	// One dimension explicitly set: the unset one defaults to Auto (0 size)
	// This makes Spacer{Width: Flex(1)} only expand horizontally
	if w.IsUnset() {
		w = Auto
	}
	if h.IsUnset() {
		h = Auto
	}
	return w, h
}

// BuildLayoutNode builds a layout node for this Spacer widget.
func (s Spacer) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	w, h := s.GetContentDimensions()
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(
		DimensionSet{
			Width:     w,
			Height:    h,
			MinWidth:  s.MinWidth,
			MaxWidth:  s.MaxWidth,
			MinHeight: s.MinHeight,
			MaxHeight: s.MaxHeight,
		},
		layout.EdgeInsets{},
		layout.EdgeInsets{},
	)

	return &layout.BoxNode{
		MinWidth:     minWidth,
		MaxWidth:     maxWidth,
		MinHeight:    minHeight,
		MaxHeight:    maxHeight,
		ExpandWidth:  w.IsFlex() || w.IsPercent(),
		ExpandHeight: h.IsFlex() || h.IsPercent(),
	}
}

// Render is a no-op for Spacer as it has no visual appearance.
func (s Spacer) Render(ctx *RenderContext) {
	// Intentionally empty - spacers don't render anything
}
