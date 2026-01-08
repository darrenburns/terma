package layout

// BoxNode is a leaf node representing a fixed or measured box.
// It implements LayoutNode and produces a BoxModel with no children.
type BoxNode struct {
	// Fixed size (if MeasureFunc is nil).
	// These are border-box dimensions.
	Width  int
	Height int

	// Node's own min/max constraints (0 = no constraint).
	// These are merged with parent constraints to form effective constraints.
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int

	// Insets
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

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
// The result is the tightest constraints that satisfy both.
func (b *BoxNode) effectiveConstraints(parent Constraints) Constraints {
	effective := parent

	// Tighten min constraints (take the larger minimum)
	if b.MinWidth > 0 && b.MinWidth > effective.MinWidth {
		effective.MinWidth = b.MinWidth
	}
	if b.MinHeight > 0 && b.MinHeight > effective.MinHeight {
		effective.MinHeight = b.MinHeight
	}

	// Tighten max constraints (take the smaller maximum)
	if b.MaxWidth > 0 && b.MaxWidth < effective.MaxWidth {
		effective.MaxWidth = b.MaxWidth
	}
	if b.MaxHeight > 0 && b.MaxHeight < effective.MaxHeight {
		effective.MaxHeight = b.MaxHeight
	}

	// Ensure min doesn't exceed max (min wins if conflict)
	if effective.MinWidth > effective.MaxWidth {
		effective.MaxWidth = effective.MinWidth
	}
	if effective.MinHeight > effective.MaxHeight {
		effective.MaxHeight = effective.MinHeight
	}

	return effective
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
