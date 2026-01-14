package layout

// HorizontalAlignment specifies horizontal positioning within available space.
type HorizontalAlignment int

const (
	// HAlignStart aligns content at the start (left).
	HAlignStart HorizontalAlignment = iota
	// HAlignCenter centers content horizontally.
	HAlignCenter
	// HAlignEnd aligns content at the end (right).
	HAlignEnd
)

// VerticalAlignment specifies vertical positioning within available space.
type VerticalAlignment int

const (
	// VAlignTop aligns content at the top.
	VAlignTop VerticalAlignment = iota
	// VAlignCenter centers content vertically.
	VAlignCenter
	// VAlignBottom aligns content at the bottom.
	VAlignBottom
)

// StackChild represents a child within a Stack, with optional positioning info.
type StackChild struct {
	Node         LayoutNode // The child's layout node
	IsPositioned bool       // True if this child uses edge-based positioning

	// Edge offsets for positioned children (nil = not constrained).
	// If both Top and Bottom are set, child height = stack height - top - bottom.
	// If both Left and Right are set, child width = stack width - left - right.
	Top    *int
	Right  *int
	Bottom *int
	Left   *int
}

// StackNode overlays children on top of each other.
// First child is at the bottom, last child is on top.
//
// Stack sizes itself based on the largest non-positioned child.
// Positioned children do not affect Stack's size (they can overflow).
type StackNode struct {
	Children []StackChild // Children to overlay

	// Default alignment for non-positioned children.
	DefaultHAlign HorizontalAlignment
	DefaultVAlign VerticalAlignment

	// Insets
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Node's own min/max constraints.
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	// Expand flags for flex sizing.
	ExpandWidth  bool
	ExpandHeight bool
}

// ComputeLayout computes the StackNode's layout given parent constraints.
// Unlike other containers, Stack positions children relative to its border-box,
// not content-box. This allows Positioned children to overlap borders.
func (s *StackNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Step 1: Compute effective constraints (border-box).
	effective := s.effectiveConstraints(constraints)

	// Step 2: Determine stack border-box size.
	hInset := s.Padding.Horizontal() + s.Border.Horizontal()
	vInset := s.Padding.Vertical() + s.Border.Vertical()

	// For measuring children, give them content-box constraints
	contentMaxW := max(0, effective.MaxWidth-hInset)
	contentMaxH := max(0, effective.MaxHeight-vInset)
	looseConstraints := Loose(contentMaxW, contentMaxH)

	var stackBorderBoxWidth, stackBorderBoxHeight int

	// Check if dimensions are explicitly constrained (Cells dimension)
	if effective.IsTightWidth() {
		stackBorderBoxWidth = effective.MaxWidth
	}
	if effective.IsTightHeight() {
		stackBorderBoxHeight = effective.MaxHeight
	}

	// First pass: measure non-positioned children to determine stack size.
	childLayouts := make([]ComputedLayout, len(s.Children))
	for i, child := range s.Children {
		if !child.IsPositioned {
			layout := child.Node.ComputeLayout(looseConstraints)
			childLayouts[i] = layout
			// Stack size is the maximum of all non-positioned children (if not explicit).
			// Add insets to convert child size to border-box contribution.
			if !effective.IsTightWidth() {
				childContribution := layout.Box.MarginBoxWidth() + hInset
				if childContribution > stackBorderBoxWidth {
					stackBorderBoxWidth = childContribution
				}
			}
			if !effective.IsTightHeight() {
				childContribution := layout.Box.MarginBoxHeight() + vInset
				if childContribution > stackBorderBoxHeight {
					stackBorderBoxHeight = childContribution
				}
			}
		}
	}

	// Apply effective constraints.
	stackBorderBoxWidth, stackBorderBoxHeight = effective.Constrain(stackBorderBoxWidth, stackBorderBoxHeight)

	// If expand flags are set, fill available space.
	if s.ExpandWidth {
		stackBorderBoxWidth = effective.MaxWidth
	}
	if s.ExpandHeight {
		stackBorderBoxHeight = effective.MaxHeight
	}

	// Step 3: Layout positioned children now that we know stack size.
	// Positioned children are constrained/positioned relative to border-box.
	for i, child := range s.Children {
		if child.IsPositioned {
			childConstraints := s.computePositionedConstraints(child, stackBorderBoxWidth, stackBorderBoxHeight)
			childLayouts[i] = child.Node.ComputeLayout(childConstraints)
		}
	}

	// Step 4: Position all children relative to border-box origin.
	// The renderer will need to handle that children may overlap border/padding.
	positionedChildren := make([]PositionedChild, len(s.Children))
	for i, child := range s.Children {
		layout := childLayouts[i]
		var x, y int

		if child.IsPositioned {
			x, y = s.computePositionedPosition(child, layout.Box, stackBorderBoxWidth, stackBorderBoxHeight)
		} else {
			// Non-positioned children are aligned within content area
			contentW := stackBorderBoxWidth - hInset
			contentH := stackBorderBoxHeight - vInset
			x, y = s.computeAlignedPosition(layout.Box, contentW, contentH)
			// Offset to content area origin
			x += s.Border.Left + s.Padding.Left
			y += s.Border.Top + s.Padding.Top
		}

		positionedChildren[i] = PositionedChild{
			X:      x,
			Y:      y,
			Layout: layout,
		}
	}

	// Step 5: Build the result.
	return ComputedLayout{
		Box: BoxModel{
			Width:   stackBorderBoxWidth,
			Height:  stackBorderBoxHeight,
			Padding: s.Padding,
			Border:  s.Border,
			Margin:  s.Margin,
		},
		Children: positionedChildren,
	}
}

// effectiveConstraints computes the intersection of parent constraints and node's own min/max.
func (s *StackNode) effectiveConstraints(parent Constraints) Constraints {
	return parent.WithNodeConstraints(s.MinWidth, s.MaxWidth, s.MinHeight, s.MaxHeight)
}

// computePositionedConstraints determines constraints for a positioned child.
func (s *StackNode) computePositionedConstraints(child StackChild, stackWidth, stackHeight int) Constraints {
	// Start with loose constraints.
	minW, maxW := 0, stackWidth
	minH, maxH := 0, stackHeight

	// If both horizontal edges are set, child width is determined.
	if child.Left != nil && child.Right != nil {
		w := stackWidth - *child.Left - *child.Right
		if w < 0 {
			w = 0
		}
		minW, maxW = w, w
	}

	// If both vertical edges are set, child height is determined.
	if child.Top != nil && child.Bottom != nil {
		h := stackHeight - *child.Top - *child.Bottom
		if h < 0 {
			h = 0
		}
		minH, maxH = h, h
	}

	return Constraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// computePositionedPosition calculates position for a positioned child.
func (s *StackNode) computePositionedPosition(child StackChild, box BoxModel, stackWidth, stackHeight int) (x, y int) {
	childWidth := box.MarginBoxWidth()
	childHeight := box.MarginBoxHeight()

	// Horizontal positioning.
	if child.Left != nil {
		x = *child.Left
	} else if child.Right != nil {
		x = stackWidth - childWidth - *child.Right
	}

	// Vertical positioning.
	if child.Top != nil {
		y = *child.Top
	} else if child.Bottom != nil {
		y = stackHeight - childHeight - *child.Bottom
	}

	// Add margin offset (X,Y should point to border-box, not margin-box).
	x += box.Margin.Left
	y += box.Margin.Top

	return x, y
}

// computeAlignedPosition calculates position for a non-positioned child using alignment.
func (s *StackNode) computeAlignedPosition(box BoxModel, stackWidth, stackHeight int) (x, y int) {
	childWidth := box.MarginBoxWidth()
	childHeight := box.MarginBoxHeight()

	// Horizontal alignment.
	switch s.DefaultHAlign {
	case HAlignCenter:
		x = (stackWidth - childWidth) / 2
	case HAlignEnd:
		x = stackWidth - childWidth
	default: // HAlignStart
		x = 0
	}

	// Vertical alignment.
	switch s.DefaultVAlign {
	case VAlignCenter:
		y = (stackHeight - childHeight) / 2
	case VAlignBottom:
		y = stackHeight - childHeight
	default: // VAlignTop
		y = 0
	}

	// Add margin offset (X,Y should point to border-box, not margin-box).
	x += box.Margin.Left
	y += box.Margin.Top

	return x, y
}
