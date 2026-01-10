package layout

import "math"

// ScrollableNode wraps a child and applies viewport/scrolling semantics.
// Unlike DockNode (which partitions space), ScrollableNode clips content
// that exceeds the viewport and tracks virtual dimensions.
//
// The layout algorithm:
// 1. Reserve scrollbar space if needed (based on AlwaysReserveScrollbarSpace or content size)
// 2. Measure child with unbounded height to determine virtual content size
// 3. If virtual size exceeds viewport, scrollbar space is reserved
// 4. Build BoxModel with VirtualHeight/ScrollOffsetY populated
//
// Example:
//
//	scrollable := &ScrollableNode{
//	    Child:         longContent,
//	    ScrollOffsetY: 10,
//	    ScrollbarWidth: 1,
//	}
//	result := scrollable.ComputeLayout(Tight(80, 24))
//	// result.Box.VirtualHeight may exceed 24
//	// result.Box.IsScrollableY() returns true if content overflows
type ScrollableNode struct {
	// Child is the content to scroll.
	Child LayoutNode

	// ScrollOffsetX is the horizontal scroll offset in cells.
	ScrollOffsetX int

	// ScrollOffsetY is the vertical scroll offset in cells.
	ScrollOffsetY int

	// ScrollbarWidth is the space reserved for a vertical scrollbar (default 1).
	// Set to 0 to disable vertical scrollbar space reservation.
	ScrollbarWidth int

	// ScrollbarHeight is the space reserved for a horizontal scrollbar (default 1).
	// Set to 0 to disable horizontal scrollbar space reservation.
	ScrollbarHeight int

	// AlwaysReserveScrollbarSpace forces scrollbar space reservation even when
	// content doesn't require scrolling. This prevents layout "jumping" when
	// content crosses the scrollable threshold.
	AlwaysReserveScrollbarSpace bool

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

// ComputeLayout computes the scrollable layout.
// The child is measured with unbounded constraints on the scroll axis to
// determine its natural (virtual) size. The viewport is then constrained
// to the parent constraints, and scroll offsets are applied.
func (s *ScrollableNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Apply node's own constraints
	effective := constraints.WithNodeConstraints(s.MinWidth, s.MaxWidth, s.MinHeight, s.MaxHeight)

	// Use scrollbar dimensions as specified (0 = no space reserved, e.g. for overlay/hidden scrollbars)
	scrollbarWidth := s.ScrollbarWidth
	scrollbarHeight := s.ScrollbarHeight

	// Convert to content-box constraints (subtract our own padding/border)
	hInset := s.Padding.Horizontal() + s.Border.Horizontal()
	vInset := s.Padding.Vertical() + s.Border.Vertical()

	contentMaxWidth := max(0, effective.MaxWidth-hInset)
	contentMaxHeight := max(0, effective.MaxHeight-vInset)
	contentMinWidth := max(0, effective.MinWidth-hInset)
	contentMinHeight := max(0, effective.MinHeight-vInset)

	// Phase 1: Measure child with unbounded height to get virtual content size
	// Reserve scrollbar space pessimistically if AlwaysReserveScrollbarSpace is set
	measureWidth := contentMaxWidth
	if s.AlwaysReserveScrollbarSpace && scrollbarWidth > 0 {
		measureWidth = max(0, measureWidth-scrollbarWidth)
	}

	unboundedConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  measureWidth,
		MinHeight: 0,
		MaxHeight: math.MaxInt32, // Unbounded height for measuring
	}

	childLayout := s.Child.ComputeLayout(unboundedConstraints)
	virtualHeight := childLayout.Box.MarginBoxHeight()
	virtualWidth := childLayout.Box.MarginBoxWidth()

	// Phase 2: Determine if scrolling is needed
	needsVerticalScroll := virtualHeight > contentMaxHeight
	needsHorizontalScroll := virtualWidth > contentMaxWidth

	// Phase 3: Re-measure if vertical scrollbar needed but space wasn't reserved initially.
	// Reducing width for scrollbar may cause content to reflow (text wrapping) or
	// reveal horizontal overflow (fixed-width elements).
	if needsVerticalScroll && !s.AlwaysReserveScrollbarSpace && scrollbarWidth > 0 {
		measureWidth = max(0, contentMaxWidth-scrollbarWidth)

		// Measure with bounded width for proper text wrapping and height calculation
		boundedConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  measureWidth,
			MinHeight: 0,
			MaxHeight: math.MaxInt32,
		}
		childLayout = s.Child.ComputeLayout(boundedConstraints)
		virtualHeight = childLayout.Box.MarginBoxHeight()

		// Measure with unbounded width to detect horizontal overflow.
		// This reveals if the child has minimum width requirements that exceed
		// the reduced viewport (e.g., fixed-width elements, long words).
		unboundedWidthConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  math.MaxInt32, // Unbounded width
			MinHeight: 0,
			MaxHeight: math.MaxInt32,
		}
		naturalLayout := s.Child.ComputeLayout(unboundedWidthConstraints)
		virtualWidth = naturalLayout.Box.MarginBoxWidth()

		// Re-evaluate horizontal scroll: does natural width exceed reduced viewport?
		needsHorizontalScroll = virtualWidth > measureWidth
	}

	// Determine actual scrollbar space to reserve
	actualScrollbarWidth := 0
	if needsVerticalScroll && scrollbarWidth > 0 {
		actualScrollbarWidth = scrollbarWidth
	} else if s.AlwaysReserveScrollbarSpace && scrollbarWidth > 0 {
		actualScrollbarWidth = scrollbarWidth
	}

	actualScrollbarHeight := 0
	if needsHorizontalScroll && scrollbarHeight > 0 {
		actualScrollbarHeight = scrollbarHeight
	} else if s.AlwaysReserveScrollbarSpace && scrollbarHeight > 0 {
		actualScrollbarHeight = scrollbarHeight
	}

	// Phase 4: Determine viewport size
	// Viewport is the smaller of virtual content size and available space.
	// This shrink-wraps when content fits, and caps at constraints when content overflows.
	// ScrollableNode acts as a "boundary" - it contains content tightly rather than
	// expanding to fill available space like layout containers do.
	viewportWidth := min(virtualWidth, contentMaxWidth)
	viewportHeight := min(virtualHeight, contentMaxHeight)

	// Clamp viewport to min constraints
	viewportWidth = max(contentMinWidth, viewportWidth)
	viewportHeight = max(contentMinHeight, viewportHeight)

	// Phase 5: Build the result
	// The BoxModel represents the viewport, with virtual dimensions for scrolling
	borderBoxWidth := viewportWidth + hInset
	borderBoxHeight := viewportHeight + vInset

	// Clamp scroll offsets to valid bounds
	maxScrollY := max(0, virtualHeight-viewportHeight+actualScrollbarHeight)
	maxScrollX := max(0, virtualWidth-viewportWidth+actualScrollbarWidth)

	clampedScrollY := max(0, min(s.ScrollOffsetY, maxScrollY))
	clampedScrollX := max(0, min(s.ScrollOffsetX, maxScrollX))

	box := BoxModel{
		Width:           borderBoxWidth,
		Height:          borderBoxHeight,
		Padding:         s.Padding,
		Border:          s.Border,
		Margin:          s.Margin,
		VirtualWidth:    virtualWidth,
		VirtualHeight:   virtualHeight,
		ScrollOffsetX:   clampedScrollX,
		ScrollOffsetY:   clampedScrollY,
		ScrollbarWidth:  actualScrollbarWidth,
		ScrollbarHeight: actualScrollbarHeight,
	}

	// Child is positioned at (0, 0) relative to our content area
	// The scroll offset affects rendering, not layout position
	return ComputedLayout{
		Box: box,
		Children: []PositionedChild{{
			X:      0,
			Y:      0,
			Layout: childLayout,
		}},
	}
}
