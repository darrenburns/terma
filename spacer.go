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
	MinMaxDimensions
}

// Build returns itself as Spacer is a leaf widget.
func (s Spacer) Build(ctx BuildContext) Widget {
	return s
}

// GetContentDimensions returns the width and height dimension preferences.
// Unset dimensions default to Flex(1).
func (s Spacer) GetContentDimensions() (width, height Dimension) {
	w, h := s.Width, s.Height
	if w.IsUnset() {
		w = Flex(1)
	}
	if h.IsUnset() {
		h = Flex(1)
	}
	return w, h
}

// BuildLayoutNode builds a layout node for this Spacer widget.
func (s Spacer) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	return &layout.BoxNode{}
}

// Render is a no-op for Spacer as it has no visual appearance.
func (s Spacer) Render(ctx *RenderContext) {
	// Intentionally empty - spacers don't render anything
}
