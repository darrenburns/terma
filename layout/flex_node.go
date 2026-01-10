package layout

// FlexNode wraps a child with flex behavior for use in LinearNode.
// It's "transparent" - it doesn't produce its own box, just metadata for the parent.
//
// FlexNode is only meaningful as a direct child of LinearNode (Row/Column).
// The parent inspects FlexNode.Flex to determine space distribution,
// then layouts FlexNode.Child with the calculated constraints.
//
// The output ComputedLayout is the child's layout - FlexNode doesn't appear
// in the layout tree. This ensures output indices match input indices,
// which is critical for widget-to-layout mapping during rendering.
//
// Example:
//
//	row := &RowNode{
//	    Children: []LayoutNode{
//	        &BoxNode{Width: 100},              // Fixed 100 cells
//	        &FlexNode{Flex: 1, Child: a},      // Gets 1/3 of remaining
//	        &FlexNode{Flex: 2, Child: b},      // Gets 2/3 of remaining
//	    },
//	}
//
// FlexNode + MainAxisAlignment Composition:
// Flex children absorb remaining space first, then MainAxisAlignment distributes
// any leftover (e.g., if flex children hit MaxWidth). This matches CSS flexbox
// behavior where flex-grow and justify-content interact.
type FlexNode struct {
	// Child is the wrapped layout node.
	Child LayoutNode

	// Flex is the proportion of remaining space this child should receive.
	// Must be > 0. Defaults to 1 if not specified.
	// A child with Flex: 2 gets twice the space of a sibling with Flex: 1.
	Flex float64
}

// ComputeLayout delegates to the child.
// This is only called if FlexNode is used outside a LinearNode context,
// in which case the Flex value is ignored and the child is measured normally.
func (f *FlexNode) ComputeLayout(constraints Constraints) ComputedLayout {
	return f.Child.ComputeLayout(constraints)
}

// FlexValue returns the flex value, defaulting to 1 if not set or invalid.
// This avoids mutating the struct during layout computation.
func (f *FlexNode) FlexValue() float64 {
	if f.Flex <= 0 {
		return 1
	}
	return f.Flex
}

// IsFlexNode returns true if the node is a FlexNode.
// Used by LinearNode to identify flex children during layout.
func IsFlexNode(node LayoutNode) (*FlexNode, bool) {
	flex, ok := node.(*FlexNode)
	return flex, ok
}
