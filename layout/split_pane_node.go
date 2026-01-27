package layout

// SplitPaneNode lays out two children separated by a divider.
// The container always fills the available space in its constraints.
type SplitPaneNode struct {
	First  LayoutNode
	Second LayoutNode
	Axis   Axis

	Position    float64 // 0.0-1.0 along the main axis
	DividerSize int
	MinPaneSize int

	// Container insets and constraints.
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	// Preserve flags resist cross-axis stretching when Auto is explicitly set.
	PreserveWidth  bool
	PreserveHeight bool
}

// ComputeLayout computes the layout for a split pane container.
func (n *SplitPaneNode) ComputeLayout(constraints Constraints) ComputedLayout {
	effective := n.effectiveConstraints(constraints)
	contentConstraints := n.toContentConstraints(effective)

	axisSize := contentConstraints.MaxWidth
	crossSize := contentConstraints.MaxHeight
	if n.Axis == Vertical {
		axisSize = contentConstraints.MaxHeight
		crossSize = contentConstraints.MaxWidth
	}

	metrics := computeSplitPaneMetrics(axisSize, n.dividerSize(), n.minPaneSize(), n.Position)
	dividerOffset := metrics.offset
	dividerSize := metrics.dividerSize
	availableMain := metrics.available

	firstNode := n.First
	if firstNode == nil {
		firstNode = &BoxNode{}
	}
	secondNode := n.Second
	if secondNode == nil {
		secondNode = &BoxNode{}
	}

	var firstConstraints, secondConstraints Constraints
	if n.Axis == Horizontal {
		firstConstraints = Constraints{
			MinWidth:  dividerOffset,
			MaxWidth:  dividerOffset,
			MinHeight: crossSize,
			MaxHeight: crossSize,
		}
		secondConstraints = Constraints{
			MinWidth:  availableMain - dividerOffset,
			MaxWidth:  availableMain - dividerOffset,
			MinHeight: crossSize,
			MaxHeight: crossSize,
		}
	} else {
		firstConstraints = Constraints{
			MinWidth:  crossSize,
			MaxWidth:  crossSize,
			MinHeight: dividerOffset,
			MaxHeight: dividerOffset,
		}
		secondConstraints = Constraints{
			MinWidth:  crossSize,
			MaxWidth:  crossSize,
			MinHeight: availableMain - dividerOffset,
			MaxHeight: availableMain - dividerOffset,
		}
	}

	firstLayout := firstNode.ComputeLayout(firstConstraints)
	secondLayout := secondNode.ComputeLayout(secondConstraints)

	firstX, firstY := 0, 0
	secondX, secondY := 0, 0
	if n.Axis == Horizontal {
		secondX = dividerOffset + dividerSize
	} else {
		secondY = dividerOffset + dividerSize
	}

	// Adjust for child margins (positions are border-box, margin is external).
	firstX += firstLayout.Box.Margin.Left
	firstY += firstLayout.Box.Margin.Top
	secondX += secondLayout.Box.Margin.Left
	secondY += secondLayout.Box.Margin.Top

	children := []PositionedChild{
		{
			X:      firstX,
			Y:      firstY,
			Layout: firstLayout,
		},
		{
			X:      secondX,
			Y:      secondY,
			Layout: secondLayout,
		},
	}

	// Fill available space based on max constraints.
	contentWidth := contentConstraints.MaxWidth
	contentHeight := contentConstraints.MaxHeight
	borderBoxWidth := contentWidth + n.Padding.Horizontal() + n.Border.Horizontal()
	borderBoxHeight := contentHeight + n.Padding.Vertical() + n.Border.Vertical()
	borderBoxWidth, borderBoxHeight = effective.Constrain(borderBoxWidth, borderBoxHeight)

	return ComputedLayout{
		Box: BoxModel{
			Width:   borderBoxWidth,
			Height:  borderBoxHeight,
			Padding: n.Padding,
			Border:  n.Border,
			Margin:  n.Margin,
		},
		Children: children,
	}
}

// PreservesWidth indicates whether this node resists horizontal stretching.
func (n *SplitPaneNode) PreservesWidth() bool {
	return n.PreserveWidth
}

// PreservesHeight indicates whether this node resists vertical stretching.
func (n *SplitPaneNode) PreservesHeight() bool {
	return n.PreserveHeight
}

func (n *SplitPaneNode) effectiveConstraints(parent Constraints) Constraints {
	return parent.WithNodeConstraints(n.MinWidth, n.MaxWidth, n.MinHeight, n.MaxHeight)
}

func (n *SplitPaneNode) toContentConstraints(constraints Constraints) Constraints {
	hInset := n.Padding.Horizontal() + n.Border.Horizontal()
	vInset := n.Padding.Vertical() + n.Border.Vertical()

	return Constraints{
		MinWidth:  max(0, constraints.MinWidth-hInset),
		MaxWidth:  max(0, constraints.MaxWidth-hInset),
		MinHeight: max(0, constraints.MinHeight-vInset),
		MaxHeight: max(0, constraints.MaxHeight-vInset),
	}
}

func (n *SplitPaneNode) dividerSize() int {
	if n.DividerSize <= 0 {
		return 1
	}
	return n.DividerSize
}

func (n *SplitPaneNode) minPaneSize() int {
	if n.MinPaneSize <= 0 {
		return 1
	}
	return n.MinPaneSize
}

type splitPaneMetrics struct {
	offset      int
	available   int
	dividerSize int
}

func computeSplitPaneMetrics(axisSize, dividerSize, minPaneSize int, position float64) splitPaneMetrics {
	if axisSize < 0 {
		axisSize = 0
	}
	if dividerSize <= 0 {
		dividerSize = 1
	}
	if dividerSize > axisSize {
		dividerSize = axisSize
	}

	available := max(0, axisSize-dividerSize)

	pos := clampFloat64(position, 0, 1)
	offset := 0
	if available > 0 {
		offset = int(float64(available) * pos)
	}

	minPane := max(0, minPaneSize)
	if available <= 0 {
		return splitPaneMetrics{offset: 0, available: 0, dividerSize: dividerSize}
	}

	if available < 2*minPane {
		offset = available / 2
		return splitPaneMetrics{offset: offset, available: available, dividerSize: dividerSize}
	}

	minOffset := minPane
	maxOffset := available - minPane
	if offset < minOffset {
		offset = minOffset
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	return splitPaneMetrics{offset: offset, available: available, dividerSize: dividerSize}
}

func clampFloat64(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
