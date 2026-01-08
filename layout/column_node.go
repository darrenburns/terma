package layout

// ColumnNode lays out children vertically (top to bottom).
// This is a thin wrapper around LinearNode with Axis=Vertical.
type ColumnNode struct {
	// Spacing is the gap between children.
	Spacing int

	// MainAlign controls vertical distribution of children.
	MainAlign MainAxisAlignment

	// CrossAlign controls horizontal positioning of children.
	CrossAlign CrossAxisAlignment

	// Children to lay out.
	Children []LayoutNode

	// Container's own insets (optional).
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Container's own size constraints (optional, 0 means unconstrained).
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// ComputeLayout computes the column layout by delegating to LinearNode.
func (c *ColumnNode) ComputeLayout(constraints Constraints) ComputedLayout {
	return (&LinearNode{
		Axis:       Vertical,
		Spacing:    c.Spacing,
		MainAlign:  c.MainAlign,
		CrossAlign: c.CrossAlign,
		Children:   c.Children,
		Padding:    c.Padding,
		Border:     c.Border,
		Margin:     c.Margin,
		MinWidth:   c.MinWidth,
		MaxWidth:   c.MaxWidth,
		MinHeight:  c.MinHeight,
		MaxHeight:  c.MaxHeight,
	}).ComputeLayout(constraints)
}
