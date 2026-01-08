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
}

// ComputeLayout computes the layout for this linear container.
func (l *LinearNode) ComputeLayout(constraints Constraints) ComputedLayout {
	if len(l.Children) == 0 {
		return l.emptyLayout(constraints)
	}

	// Step 1: Convert to content-box constraints (space available for children)
	contentConstraints := l.toContentConstraints(constraints)

	// Step 2: First pass - measure all children
	childLayouts, totalMain, maxCross := l.measureChildren(contentConstraints)

	// Step 3: Determine final container size
	containerMain, containerCross := l.determineContainerSize(contentConstraints, totalMain, maxCross)

	// Step 4: Calculate main-axis positions based on alignment
	mainPositions := l.calculateMainPositions(childLayouts, containerMain)

	// Step 5: Position children (apply cross-axis alignment, possibly re-layout for stretch)
	positionedChildren := l.positionChildren(childLayouts, mainPositions, containerCross, contentConstraints)

	// Step 6: Build the final BoxModel
	return l.buildResult(constraints, containerMain, containerCross, positionedChildren)
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

// measureChildren performs the first pass: measure each child and track totals.
func (l *LinearNode) measureChildren(contentConstraints Constraints) ([]ComputedLayout, int, int) {
	childLayouts := make([]ComputedLayout, len(l.Children))
	totalMain := 0
	maxCross := 0

	for i, child := range l.Children {
		// Give each child loose constraints on main axis, full cross constraint
		childConstraints := l.makeChildConstraints(contentConstraints)
		childLayouts[i] = child.ComputeLayout(childConstraints)

		// Accumulate sizes (use margin-box for spacing calculations)
		childMain := l.mainSize(childLayouts[i].Box.MarginBoxWidth(), childLayouts[i].Box.MarginBoxHeight())
		childCross := l.crossSize(childLayouts[i].Box.MarginBoxWidth(), childLayouts[i].Box.MarginBoxHeight())

		totalMain += childMain
		if childCross > maxCross {
			maxCross = childCross
		}
	}

	// Add spacing between children
	if len(l.Children) > 1 {
		totalMain += l.Spacing * (len(l.Children) - 1)
	}

	return childLayouts, totalMain, maxCross
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

// determineContainerSize calculates the container's content size.
func (l *LinearNode) determineContainerSize(contentConstraints Constraints, totalMain, maxCross int) (int, int) {
	mainMin, mainMax := l.mainConstraint(contentConstraints)
	crossMin, crossMax := l.crossConstraint(contentConstraints)

	// Clamp to constraints
	containerMain := max(mainMin, min(mainMax, totalMain))
	containerCross := max(crossMin, min(crossMax, maxCross))

	return containerMain, containerCross
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
			// Re-layout child with tight cross constraint
			if childCross < containerCross {
				stretchedConstraints := l.makeStretchConstraints(contentConstraints, containerCross)
				layout = l.Children[i].ComputeLayout(stretchedConstraints)
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

// makeStretchConstraints creates constraints for stretching a child on the cross axis.
func (l *LinearNode) makeStretchConstraints(contentConstraints Constraints, crossSize int) Constraints {
	_, mainMax := l.mainConstraint(contentConstraints)
	// Main axis: loose (0 to max) - child keeps its natural main-axis size
	// Cross axis: tight (crossSize) - child stretches to fill cross-axis
	return l.makeConstraints(0, mainMax, crossSize, crossSize, 0, 0)
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
