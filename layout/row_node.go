package layout

// RowNode lays out children horizontally (left to right).
// This is a thin wrapper around LinearNode with Axis=Horizontal.
type RowNode struct {
	// Spacing is the gap between children.
	Spacing int

	// MainAlign controls horizontal distribution of children.
	MainAlign MainAxisAlignment

	// CrossAlign controls vertical positioning of children.
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

	// Expand flags force the container to fill available space on that axis.
	// Used when the widget's dimension is Flex() to mean "fill available space".
	ExpandWidth  bool
	ExpandHeight bool

	// Preserve flags indicate the container should resist stretching on that axis.
	// Used when the widget's dimension is explicitly Auto to mean "fit content, don't stretch".
	PreserveWidth  bool
	PreserveHeight bool
}

// ComputeLayout computes the row layout by delegating to LinearNode.
func (r *RowNode) ComputeLayout(constraints Constraints) ComputedLayout {
	return (&LinearNode{
		Axis:           Horizontal,
		Spacing:        r.Spacing,
		MainAlign:      r.MainAlign,
		CrossAlign:     r.CrossAlign,
		Children:       r.Children,
		Padding:        r.Padding,
		Border:         r.Border,
		Margin:         r.Margin,
		MinWidth:       r.MinWidth,
		MaxWidth:       r.MaxWidth,
		MinHeight:      r.MinHeight,
		MaxHeight:      r.MaxHeight,
		ExpandWidth:    r.ExpandWidth,
		ExpandHeight:   r.ExpandHeight,
		PreserveWidth:  r.PreserveWidth,
		PreserveHeight: r.PreserveHeight,
	}).ComputeLayout(constraints)
}

// PreservesWidth implements SizePreserver.
func (r *RowNode) PreservesWidth() bool {
	return r.PreserveWidth
}

// PreservesHeight implements SizePreserver.
func (r *RowNode) PreservesHeight() bool {
	return r.PreserveHeight
}
