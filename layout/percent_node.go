package layout

// PercentNode wraps a child with percentage-based sizing for use in LinearNode.
// Unlike FlexNode which distributes remaining space, PercentNode calculates
// a fixed size as a percentage of the parent's available space.
//
// PercentNode is only meaningful as a direct child of LinearNode (Row/Column).
// The parent uses PercentNode during the first pass (non-flex measurement),
// where the percentage is resolved to a fixed size based on constraints.
//
// Example:
//
//	row := &RowNode{
//	    Children: []LayoutNode{
//	        &PercentNode{Percent: 30, Axis: Horizontal, Child: a},  // 30% of parent width
//	        &PercentNode{Percent: 70, Axis: Horizontal, Child: b},  // 70% of parent width
//	    },
//	}
type PercentNode struct {
	// Child is the wrapped layout node.
	Child LayoutNode

	// Percent is the percentage of parent space (0-100+).
	// Values over 100 are allowed and will cause overflow.
	Percent float64

	// Axis determines which constraint dimension to use for percentage calculation.
	// Horizontal uses MaxWidth, Vertical uses MaxHeight.
	Axis Axis
}

// ComputeLayout calculates the fixed size from the percentage and applies
// tight constraints on the main axis while preserving cross-axis constraints.
func (p *PercentNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Determine which constraint to use based on axis
	var maxAvailable int
	if p.Axis == Horizontal {
		maxAvailable = constraints.MaxWidth
	} else {
		maxAvailable = constraints.MaxHeight
	}

	// Calculate fixed size from percentage
	fixedSize := int(float64(maxAvailable) * p.Percent / 100.0)

	// Create tight constraint on the percentage axis
	childConstraints := constraints
	if p.Axis == Horizontal {
		childConstraints.MinWidth = fixedSize
		childConstraints.MaxWidth = fixedSize
	} else {
		childConstraints.MinHeight = fixedSize
		childConstraints.MaxHeight = fixedSize
	}

	return p.Child.ComputeLayout(childConstraints)
}

// IsPercentNode returns true if the node is a PercentNode.
// Used by LinearNode to identify percentage children during layout.
func IsPercentNode(node LayoutNode) (*PercentNode, bool) {
	pct, ok := node.(*PercentNode)
	return pct, ok
}
