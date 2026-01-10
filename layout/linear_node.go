package layout

// LinearNode lays out children along a single axis (horizontal or vertical).
// This is the shared implementation for Row (Horizontal) and Column (Vertical).
// The algorithm operates on "main axis" and "cross axis" concepts, making
// the code axis-agnostic.
type LinearNode struct {
	// Axis determines the layout direction.
	// Horizontal: children flow left-to-right (like Row)
	// Vertical: children flow top-to-bottom (like Column)
	Axis Axis

	// Spacing is the gap between children along the main axis.
	Spacing int

	// MainAlign controls distribution of children along the main axis.
	MainAlign MainAxisAlignment

	// CrossAlign controls positioning of children along the cross axis.
	CrossAlign CrossAxisAlignment

	// Children to lay out.
	Children []LayoutNode

	// Container's own insets (optional).
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Container's own size constraints (optional, 0 means unconstrained).
	// These are border-box constraints, applied after content-based sizing
	// but before parent constraints. Allows containers to enforce minimum
	// sizes independent of their children.
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// ComputeLayout computes the layout for this linear container.
func (l *LinearNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Combine parent constraints with node's own constraints
	effective := l.effectiveConstraints(constraints)

	if len(l.Children) == 0 {
		return l.emptyLayout(effective)
	}

	// Step 1: Convert to content-box constraints (space available for children)
	contentConstraints := l.toContentConstraints(effective)

	// Step 2: First pass - measure non-flex children, identify flex children
	childLayouts, fixedInfo := l.measureNonFlexChildren(contentConstraints)

	// Step 3: Determine container size (expand to max if flex children present)
	containerMain, containerCross := l.determineContainerSizeWithFlex(contentConstraints, fixedInfo)

	// Step 4: Second pass - measure flex children with allocated space
	l.measureFlexChildren(childLayouts, containerMain, containerCross, fixedInfo, contentConstraints)

	// Step 5: Recalculate maxCross including flex children
	maxCross := fixedInfo.maxCross
	for i, layout := range childLayouts {
		if fixedInfo.isFlexChild[i] {
			childCross := l.crossSize(layout.Box.MarginBoxWidth(), layout.Box.MarginBoxHeight())
			if childCross > maxCross {
				maxCross = childCross
			}
		}
	}

	// Step 6: Re-determine container cross size with flex children included
	_, crossMax := l.crossConstraint(contentConstraints)
	crossMin, _ := l.crossConstraint(contentConstraints)
	containerCross = max(crossMin, min(crossMax, maxCross))

	// Step 7: Calculate main-axis positions based on alignment
	mainPositions := l.calculateMainPositionsWithFlex(childLayouts, containerMain)

	// Step 8: Position children (apply cross-axis alignment, possibly re-layout for stretch)
	positionedChildren := l.positionChildren(childLayouts, mainPositions, containerCross, contentConstraints)

	// Step 9: Build the final BoxModel
	return l.buildResult(effective, containerMain, containerCross, positionedChildren)
}

// fixedLayoutInfo holds information from the first pass (non-flex measurement).
type fixedLayoutInfo struct {
	totalFixedContent int       // Sum of non-flex children's main-axis sizes (excludes spacing)
	totalFlex         float64   // Sum of all Flex values
	maxCross          int       // Maximum cross-axis size from non-flex children
	hasFlex           bool      // True if any child is a FlexNode
	isFlexChild       []bool    // Per-child: true if it's a FlexNode
	flexValues        []float64 // Per-child: Flex value (0 for non-flex)
}

// effectiveConstraints combines parent constraints with node's own min/max constraints.
func (l *LinearNode) effectiveConstraints(parent Constraints) Constraints {
	return parent.WithNodeConstraints(l.MinWidth, l.MaxWidth, l.MinHeight, l.MaxHeight)
}

// emptyLayout handles the case of no children.
func (l *LinearNode) emptyLayout(constraints Constraints) ComputedLayout {
	// Natural empty size is just the insets (padding + border)
	hInset := l.Padding.Horizontal() + l.Border.Horizontal()
	vInset := l.Padding.Vertical() + l.Border.Vertical()

	// Clamp to constraints once
	width, height := constraints.Constrain(hInset, vInset)

	return ComputedLayout{
		Box: BoxModel{
			Width:   width,
			Height:  height,
			Padding: l.Padding,
			Border:  l.Border,
			Margin:  l.Margin,
		},
		Children: nil,
	}
}

// toContentConstraints converts border-box constraints to content-box constraints.
func (l *LinearNode) toContentConstraints(constraints Constraints) Constraints {
	hInset := l.Padding.Horizontal() + l.Border.Horizontal()
	vInset := l.Padding.Vertical() + l.Border.Vertical()

	return Constraints{
		MinWidth:  max(0, constraints.MinWidth-hInset),
		MaxWidth:  max(0, constraints.MaxWidth-hInset),
		MinHeight: max(0, constraints.MinHeight-vInset),
		MaxHeight: max(0, constraints.MaxHeight-vInset),
	}
}

// measureNonFlexChildren performs the first pass: measure non-flex children and identify flex children.
func (l *LinearNode) measureNonFlexChildren(contentConstraints Constraints) ([]ComputedLayout, fixedLayoutInfo) {
	n := len(l.Children)
	childLayouts := make([]ComputedLayout, n)
	info := fixedLayoutInfo{
		isFlexChild: make([]bool, n),
		flexValues:  make([]float64, n),
	}

	for i, child := range l.Children {
		// Check if this is a FlexNode
		if flex, ok := IsFlexNode(child); ok {
			info.isFlexChild[i] = true
			info.hasFlex = true

			// Get normalized flex value (defaults to 1 if invalid)
			flexVal := flex.FlexValue()
			info.flexValues[i] = flexVal
			info.totalFlex += flexVal

			// Skip measuring - will be done in second pass
			continue
		}

		// Non-flex child: measure now
		childConstraints := l.makeChildConstraints(contentConstraints)
		childLayouts[i] = child.ComputeLayout(childConstraints)

		// Accumulate sizes (use margin-box for spacing calculations)
		childMain := l.mainSize(childLayouts[i].Box.MarginBoxWidth(), childLayouts[i].Box.MarginBoxHeight())
		childCross := l.crossSize(childLayouts[i].Box.MarginBoxWidth(), childLayouts[i].Box.MarginBoxHeight())

		info.totalFixedContent += childMain
		if childCross > info.maxCross {
			info.maxCross = childCross
		}
	}

	return childLayouts, info
}

// determineContainerSizeWithFlex calculates the container's content size,
// expanding to maxWidth/maxHeight if flex children are present.
func (l *LinearNode) determineContainerSizeWithFlex(contentConstraints Constraints, info fixedLayoutInfo) (int, int) {
	mainMin, mainMax := l.mainConstraint(contentConstraints)
	crossMin, crossMax := l.crossConstraint(contentConstraints)

	var containerMain int
	if info.hasFlex {
		// With flex children, container expands to fill available space
		// This gives flex children space to fill
		containerMain = mainMax
	} else {
		// Without flex children, shrink-wrap to content (including spacing)
		totalSpacing := l.totalSpacing()
		totalFixedMain := info.totalFixedContent + totalSpacing
		containerMain = max(mainMin, min(mainMax, totalFixedMain))
	}

	// Cross-axis size is determined by tallest child (so far only non-flex)
	containerCross := max(crossMin, min(crossMax, info.maxCross))

	return containerMain, containerCross
}

// totalSpacing returns the total spacing between all children.
func (l *LinearNode) totalSpacing() int {
	n := len(l.Children)
	if n > 1 {
		return l.Spacing * (n - 1)
	}
	return 0
}

// measureFlexChildren performs the second pass: allocate remaining space to flex children.
func (l *LinearNode) measureFlexChildren(
	childLayouts []ComputedLayout,
	containerMain, containerCross int,
	info fixedLayoutInfo,
	contentConstraints Constraints,
) {
	if !info.hasFlex || info.totalFlex == 0 {
		return
	}

	// Remaining space for flex children = container - fixed content - spacing
	remaining := containerMain - info.totalFixedContent - l.totalSpacing()
	if remaining < 0 {
		remaining = 0
	}

	// Distribute remaining space to flex children using cumulative distribution
	// This ensures no pixel loss due to rounding
	allocatedSoFar := 0.0
	actualAllocatedSoFar := 0

	_, crossMax := l.crossConstraint(contentConstraints)

	for i := range l.Children {
		if !info.isFlexChild[i] {
			continue
		}

		// Calculate this child's share using cumulative distribution
		allocatedSoFar += info.flexValues[i]
		targetTotal := float64(remaining) * allocatedSoFar / info.totalFlex
		thisAllocation := int(targetTotal) - actualAllocatedSoFar
		actualAllocatedSoFar += thisAllocation

		// Create tight constraint on main axis for flex child
		flexConstraints := l.makeFlexChildConstraints(contentConstraints, thisAllocation, crossMax)

		// Get the actual child (unwrap FlexNode)
		actualChild := l.Children[i]
		if flex, ok := IsFlexNode(actualChild); ok {
			actualChild = flex.Child
		}

		childLayouts[i] = actualChild.ComputeLayout(flexConstraints)
	}
}

// makeFlexChildConstraints creates constraints for a flex child.
func (l *LinearNode) makeFlexChildConstraints(contentConstraints Constraints, mainSize, crossMax int) Constraints {
	// Main axis: tight (exactly mainSize)
	// Cross axis: loose (0 to max)
	return l.makeConstraints(mainSize, mainSize, 0, crossMax, 0, 0)
}

// calculateMainPositionsWithFlex calculates positions accounting for actual child sizes.
func (l *LinearNode) calculateMainPositionsWithFlex(childLayouts []ComputedLayout, containerMain int) []int {
	// This is the same as calculateMainPositions but using actual child sizes
	return l.calculateMainPositions(childLayouts, containerMain)
}

// makeChildConstraints creates constraints for measuring a child.
func (l *LinearNode) makeChildConstraints(contentConstraints Constraints) Constraints {
	// Main axis: loose (0 to max) - children can be any size up to available space
	// Cross axis: loose (0 to max), will be tightened later if CrossAxisStretch
	_, mainMax := l.mainConstraint(contentConstraints)
	_, crossMax := l.crossConstraint(contentConstraints)

	// Children get loose constraints - they report their natural size
	return l.makeConstraints(0, mainMax, 0, crossMax, 0, 0)
}

// calculateMainPositions calculates the position of each child along the main axis.
func (l *LinearNode) calculateMainPositions(childLayouts []ComputedLayout, containerMain int) []int {
	n := len(childLayouts)
	positions := make([]int, n)

	// Calculate total child size and child sizes array
	childSizes := make([]int, n)
	totalChildMain := 0
	for i, layout := range childLayouts {
		size := l.mainSize(layout.Box.MarginBoxWidth(), layout.Box.MarginBoxHeight())
		childSizes[i] = size
		totalChildMain += size
	}

	// Calculate extra space (space not occupied by children)
	extraSpace := containerMain - totalChildMain

	// For Start/Center/End, we also account for explicit spacing
	totalSpacing := 0
	if n > 1 {
		totalSpacing = l.Spacing * (n - 1)
	}

	switch l.MainAlign {
	case MainAxisStart:
		pos := 0
		for i := 0; i < n; i++ {
			positions[i] = pos
			pos += childSizes[i] + l.Spacing
		}

	case MainAxisCenter:
		// Extra space minus spacing, divided by 2 for centering
		startOffset := (extraSpace - totalSpacing) / 2
		pos := startOffset
		for i := 0; i < n; i++ {
			positions[i] = pos
			pos += childSizes[i] + l.Spacing
		}

	case MainAxisEnd:
		// All extra space (minus spacing) goes at the start
		startOffset := extraSpace - totalSpacing
		pos := startOffset
		for i := 0; i < n; i++ {
			positions[i] = pos
			pos += childSizes[i] + l.Spacing
		}

	case MainAxisSpaceBetween:
		// First child at 0, last child flush with end
		// Distribute extra space into (n-1) gaps
		if n == 1 {
			positions[0] = 0
		} else {
			numGaps := n - 1
			pos := 0
			for i := 0; i < n; i++ {
				positions[i] = pos
				pos += childSizes[i]
				if i < numGaps {
					// Use cumulative division to distribute remainder
					// Gap i gets: (extraSpace * (i+1) / numGaps) - (extraSpace * i / numGaps)
					pos += (extraSpace * (i + 1) / numGaps) - (extraSpace * i / numGaps)
				}
			}
		}

	case MainAxisSpaceAround:
		// Each child gets equal space on both sides
		// Effectively: half-gap | child | gap | child | gap | child | half-gap
		// This means n gaps worth of space, with edges getting half
		if n == 0 {
			break
		}
		numGaps := n // n children means n units of space to distribute
		for i := 0; i < n; i++ {
			// Space before child i: half unit for first, full unit for others
			// Using cumulative: spaceBeforeChild[i] = extraSpace * (2*i + 1) / (2*n)
			spaceBefore := (extraSpace * (2*i + 1)) / (2 * numGaps)
			positions[i] = spaceBefore + sumInts(childSizes[:i])
		}

	case MainAxisSpaceEvenly:
		// Equal gaps everywhere: gap | child | gap | child | gap | child | gap
		// (n+1) gaps total
		if n == 0 {
			break
		}
		numGaps := n + 1
		for i := 0; i < n; i++ {
			// Space before child i = (i+1) gaps worth
			spaceBefore := (extraSpace * (i + 1)) / numGaps
			positions[i] = spaceBefore + sumInts(childSizes[:i])
		}
	}

	return positions
}

// sumInts returns the sum of a slice of ints.
func sumInts(vals []int) int {
	total := 0
	for _, v := range vals {
		total += v
	}
	return total
}

// positionChildren creates the final positioned children with cross-axis alignment.
func (l *LinearNode) positionChildren(
	childLayouts []ComputedLayout,
	mainPositions []int,
	containerCross int,
	contentConstraints Constraints,
) []PositionedChild {
	positioned := make([]PositionedChild, len(childLayouts))

	for i, layout := range childLayouts {
		childCross := l.crossSize(layout.Box.MarginBoxWidth(), layout.Box.MarginBoxHeight())

		// Calculate cross-axis position
		var crossPos int
		switch l.CrossAlign {
		case CrossAxisStart:
			crossPos = 0

		case CrossAxisCenter:
			crossPos = (containerCross - childCross) / 2

		case CrossAxisEnd:
			crossPos = containerCross - childCross

		case CrossAxisStretch:
			// Re-layout child with tight cross constraint, accounting for child's margins.
			// The available space for the border-box is container minus margins.
			// Stretch forces the child's layout box to match exactly, even if that means
			// shrinking. Content may visually overflow, but layout integrity is preserved.
			//
			// IMPORTANT: For flex children, we must preserve their allocated main-axis size.
			// The first layout pass gave them a specific width/height based on flex distribution.
			// We use tight main-axis constraints to prevent reverting to natural size.
			childMarginCross := l.crossSize(layout.Box.Margin.Horizontal(), layout.Box.Margin.Vertical())
			availableCross := containerCross - childMarginCross
			if availableCross > 0 {
				// Get the main-axis size from the first layout pass (preserves flex allocation)
				childMainSize := l.mainSize(layout.Box.Width, layout.Box.Height)

				// Create constraints: tight on main-axis (preserve flex), tight on cross-axis (stretch)
				stretchedConstraints := l.makeStretchConstraintsPreserveMain(childMainSize, availableCross)

				// Get the actual child (unwrap FlexNode if present)
				actualChild := l.Children[i]
				if flex, ok := IsFlexNode(actualChild); ok {
					actualChild = flex.Child
				}

				layout = actualChild.ComputeLayout(stretchedConstraints)
			}
			crossPos = 0
		}

		// Convert main/cross positions to x/y
		x, y := l.makePosition(mainPositions[i], crossPos)

		// Adjust for child's margin (position is for margin-box, but we report border-box position)
		x += layout.Box.Margin.Left
		y += layout.Box.Margin.Top

		positioned[i] = PositionedChild{
			X:      x,
			Y:      y,
			Layout: layout,
		}
	}

	return positioned
}

// makeStretchConstraintsPreserveMain creates constraints for stretching a child on the cross axis
// while preserving its main-axis size. This is critical for flex children whose main-axis
// size was determined by flex distribution and must not revert to natural size.
func (l *LinearNode) makeStretchConstraintsPreserveMain(mainSize, crossSize int) Constraints {
	// Main axis: tight (preserve the allocated/computed size from first pass)
	// Cross axis: tight (stretch to fill container)
	return l.makeConstraints(mainSize, mainSize, crossSize, crossSize, 0, 0)
}

// buildResult constructs the final ComputedLayout.
func (l *LinearNode) buildResult(
	constraints Constraints,
	containerMain, containerCross int,
	children []PositionedChild,
) ComputedLayout {
	// Convert content size to border-box
	contentWidth, contentHeight := l.makeSize(containerMain, containerCross)
	borderBoxWidth := contentWidth + l.Padding.Horizontal() + l.Border.Horizontal()
	borderBoxHeight := contentHeight + l.Padding.Vertical() + l.Border.Vertical()

	// Clamp to original constraints
	borderBoxWidth, borderBoxHeight = constraints.Constrain(borderBoxWidth, borderBoxHeight)

	return ComputedLayout{
		Box: BoxModel{
			Width:   borderBoxWidth,
			Height:  borderBoxHeight,
			Padding: l.Padding,
			Border:  l.Border,
			Margin:  l.Margin,
		},
		Children: children,
	}
}

// --- Axis abstraction helpers ---

// mainSize extracts the main-axis dimension from width and height.
func (l *LinearNode) mainSize(w, h int) int {
	if l.Axis == Horizontal {
		return w
	}
	return h
}

// crossSize extracts the cross-axis dimension from width and height.
func (l *LinearNode) crossSize(w, h int) int {
	if l.Axis == Horizontal {
		return h
	}
	return w
}

// makeSize constructs width and height from main and cross dimensions.
func (l *LinearNode) makeSize(main, cross int) (w, h int) {
	if l.Axis == Horizontal {
		return main, cross
	}
	return cross, main
}

// mainConstraint extracts the main-axis constraint (min, max).
func (l *LinearNode) mainConstraint(c Constraints) (min, max int) {
	if l.Axis == Horizontal {
		return c.MinWidth, c.MaxWidth
	}
	return c.MinHeight, c.MaxHeight
}

// crossConstraint extracts the cross-axis constraint (min, max).
func (l *LinearNode) crossConstraint(c Constraints) (min, max int) {
	if l.Axis == Horizontal {
		return c.MinHeight, c.MaxHeight
	}
	return c.MinWidth, c.MaxWidth
}

// makeConstraints constructs a Constraints from main/cross values.
func (l *LinearNode) makeConstraints(mainMin, mainMax, crossMin, crossMax, origMainMin, origCrossMin int) Constraints {
	if l.Axis == Horizontal {
		return Constraints{
			MinWidth:  max(mainMin, origMainMin),
			MaxWidth:  mainMax,
			MinHeight: max(crossMin, origCrossMin),
			MaxHeight: crossMax,
		}
	}
	return Constraints{
		MinWidth:  max(crossMin, origCrossMin),
		MaxWidth:  crossMax,
		MinHeight: max(mainMin, origMainMin),
		MaxHeight: mainMax,
	}
}

// makePosition converts main/cross position to x/y coordinates.
func (l *LinearNode) makePosition(main, cross int) (x, y int) {
	if l.Axis == Horizontal {
		return main, cross
	}
	return cross, main
}
