package layout

// BoxNode is a leaf node representing a fixed or measured box.
// It implements LayoutNode and produces a BoxModel with no children.
type BoxNode struct {
	// Fixed size (if MeasureFunc is nil).
	// These are border-box dimensions.
	Width  int
	Height int

	// Node's own min/max constraints.
	// These are merged with parent constraints to form effective constraints.
	// NOTE: 0 means "no constraint" (unconstrained), not "zero size".
	// This is a pragmatic trade-off for API simplicity in TUI contexts.
	// If you need a box that shrinks to zero, use MeasureFunc or set Width/Height directly.
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	// Insets
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Expand flags force the box to fill available space on that axis.
	// When true and Width/Height is 0, the box expands to MaxWidth/MaxHeight.
	// This is used when a widget's dimension is Flex() or Percent().
	ExpandWidth  bool
	ExpandHeight bool

	// MeasureFunc for dynamic sizing (overrides Width/Height if set).
	// Receives CONTENT-BOX constraints (available space for content, after subtracting padding/border).
	// Returns CONTENT-BOX dimensions (just the content size, not including padding/border).
	// ComputeLayout adds padding/border back automatically.
	// This keeps MeasureFunc simple - it only measures content, not decoration.
	// Can use constraints.IsTightWidth() etc. to detect forced vs flexible sizing.
	MeasureFunc func(constraints Constraints) (width, height int)
}

// ComputeLayout computes the BoxNode's layout given parent constraints.
// It returns a ComputedLayout with the resulting BoxModel and no children.
func (b *BoxNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Step 1: Compute effective constraints (intersection of parent and node's own min/max).
	// These are border-box constraints.
	effective := b.effectiveConstraints(constraints)

	// Step 2: Determine the desired border-box size
	var width, height int
	if b.MeasureFunc != nil {
		// Dynamic sizing - convert to content-box constraints for MeasureFunc
		contentConstraints := b.toContentBoxConstraints(effective)

		// MeasureFunc returns content-box dimensions
		contentWidth, contentHeight := b.MeasureFunc(contentConstraints)

		// Convert back to border-box by adding padding and border
		width = contentWidth + b.Padding.Horizontal() + b.Border.Horizontal()
		height = contentHeight + b.Padding.Vertical() + b.Border.Vertical()
	} else {
		// Fixed sizing - Width/Height are already border-box
		width, height = b.Width, b.Height

		// If expand flags are set and size is 0, fill available space.
		// In unbounded contexts (e.g. Scrollable measuring with infinite height),
		// Flex has no meaningful natural size, so we return 0 (or MinWidth/MinHeight).
		// This follows Flutter/CSS behavior where Flex content in unbounded space
		// collapses to its minimum size.
		if b.ExpandWidth && width == 0 {
			if !isUnbounded(effective.MaxWidth) {
				width = effective.MaxWidth
			}
			// else: width stays 0 - Flex has no natural size in unbounded space
		}
		if b.ExpandHeight && height == 0 {
			if !isUnbounded(effective.MaxHeight) {
				height = effective.MaxHeight
			}
			// else: height stays 0 - Flex has no natural size in unbounded space
		}
	}

	// Step 3: Clamp to effective constraints (border-box)
	width, height = effective.Constrain(width, height)

	// Step 4: Build the BoxModel
	box := BoxModel{
		Width:   width,
		Height:  height,
		Padding: b.Padding,
		Border:  b.Border,
		Margin:  b.Margin,
	}

	return ComputedLayout{
		Box:      box,
		Children: nil, // Leaf node - no children
	}
}

// effectiveConstraints computes the intersection of parent constraints and node's own min/max.
func (b *BoxNode) effectiveConstraints(parent Constraints) Constraints {
	return parent.WithNodeConstraints(b.MinWidth, b.MaxWidth, b.MinHeight, b.MaxHeight)
}

// toContentBoxConstraints converts border-box constraints to content-box constraints
// by subtracting padding and border. Used before calling MeasureFunc.
func (b *BoxNode) toContentBoxConstraints(borderBox Constraints) Constraints {
	hInset := b.Padding.Horizontal() + b.Border.Horizontal()
	vInset := b.Padding.Vertical() + b.Border.Vertical()

	return Constraints{
		MinWidth:  max(0, borderBox.MinWidth-hInset),
		MaxWidth:  max(0, borderBox.MaxWidth-hInset),
		MinHeight: max(0, borderBox.MinHeight-vInset),
		MaxHeight: max(0, borderBox.MaxHeight-vInset),
	}
}
