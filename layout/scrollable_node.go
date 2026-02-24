package layout

import "math"

// ScrollableNode wraps a child and applies viewport/scrolling semantics.
// Unlike DockNode (which partitions space), ScrollableNode clips content
// that exceeds the viewport and tracks virtual dimensions.
//
// The layout algorithm:
// 1. Reserve scrollbar space if needed (based on AlwaysReserveScrollbarSpace or content size)
// 2. Measure child with both bounded and unbounded heights, then choose virtual content size
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

const scrollableUnboundedMeasureHeight = math.MaxInt32

func clampScrollableMeasureHeight(maxHeight int) int {
	if maxHeight < 0 {
		return 0
	}
	if maxHeight > scrollableUnboundedMeasureHeight {
		return scrollableUnboundedMeasureHeight
	}
	return maxHeight
}

func shouldPreferBoundedHeightScale(boundedHeight, boundedMaxHeight, unboundedHeight, unboundedMaxHeight int) bool {
	if boundedMaxHeight <= 0 || unboundedMaxHeight <= 0 {
		return false
	}
	if !isUnbounded(unboundedHeight) {
		return false
	}
	if unboundedHeight <= boundedHeight {
		return false
	}

	boundedRatio := float64(boundedHeight) / float64(boundedMaxHeight)
	unboundedRatio := float64(unboundedHeight) / float64(unboundedMaxHeight)
	diff := boundedRatio - unboundedRatio
	if diff < 0 {
		diff = -diff
	}

	// Small bounded viewports produce coarse percentage rounding.
	// Allow a small tolerance so Percent/Flex still map to viewport-relative sizing.
	tolerance := (1.0 / float64(max(1, boundedMaxHeight))) + 0.02
	return diff <= tolerance
}

func (s *ScrollableNode) chooseViewportMeasurement(boundedLayout, unboundedLayout ComputedLayout, boundedMaxHeight int) ComputedLayout {
	boundedVirtualHeight := boundedLayout.Box.MarginBoxHeight()
	unboundedVirtualHeight := unboundedLayout.Box.MarginBoxHeight()

	// In unbounded probes, some Flex layouts collapse. Use bounded measurement
	// so Flex/Percent heights resolve against the visible viewport.
	if unboundedVirtualHeight < boundedVirtualHeight {
		return boundedLayout
	}

	if shouldPreferBoundedHeightScale(
		boundedLayout.Box.BorderBoxHeight(),
		boundedMaxHeight,
		unboundedLayout.Box.BorderBoxHeight(),
		scrollableUnboundedMeasureHeight,
	) {
		return boundedLayout
	}

	return unboundedLayout
}

func (s *ScrollableNode) measureChildForViewport(maxWidth, viewportMaxHeight int) (ComputedLayout, int, int) {
	boundedMaxHeight := clampScrollableMeasureHeight(viewportMaxHeight)

	boundedConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: boundedMaxHeight,
	}
	unboundedConstraints := Constraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: scrollableUnboundedMeasureHeight,
	}

	boundedLayout := s.Child.ComputeLayout(boundedConstraints)
	unboundedLayout := s.Child.ComputeLayout(unboundedConstraints)
	selected := s.chooseViewportMeasurement(boundedLayout, unboundedLayout, boundedMaxHeight)

	return selected, selected.Box.MarginBoxWidth(), selected.Box.MarginBoxHeight()
}

// ComputeLayout computes the scrollable layout.
// The child is measured with both viewport-bounded and unbounded height probes
// to derive a stable virtual size. The viewport is then constrained to the
// parent constraints, and scroll offsets are applied.
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

	// Phase 1: Measure child to get virtual content size
	// Reserve scrollbar space pessimistically if AlwaysReserveScrollbarSpace is set
	measureWidth := contentMaxWidth
	if s.AlwaysReserveScrollbarSpace && scrollbarWidth > 0 {
		measureWidth = max(0, measureWidth-scrollbarWidth)
	}

	childLayout, virtualWidth, virtualHeight := s.measureChildForViewport(measureWidth, contentMaxHeight)

	// Phase 2: Determine if scrolling is needed
	needsVerticalScroll := virtualHeight > contentMaxHeight
	needsHorizontalScroll := virtualWidth > contentMaxWidth

	// Phase 3: Re-measure if vertical scrollbar needed but space wasn't reserved initially.
	// Reducing width for scrollbar may cause content to reflow (text wrapping) or
	// reveal horizontal overflow (fixed-width elements).
	if needsVerticalScroll && !s.AlwaysReserveScrollbarSpace && scrollbarWidth > 0 {
		measureWidth = max(0, contentMaxWidth-scrollbarWidth)

		childLayout, virtualWidth, virtualHeight = s.measureChildForViewport(measureWidth, contentMaxHeight)

		// Measure with unbounded width to detect horizontal overflow.
		// This reveals if the child has minimum width requirements that exceed
		// the reduced viewport (e.g., fixed-width elements, long words).
		// Flex-based layouts will return 0 width in unbounded contexts (no natural size),
		// so we fall back to the bounded measurement for them.
		unboundedWidthConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  scrollableUnboundedMeasureHeight,
			MinHeight: 0,
			MaxHeight: scrollableUnboundedMeasureHeight,
		}
		naturalWidthLayout := s.Child.ComputeLayout(unboundedWidthConstraints)
		naturalWidth := naturalWidthLayout.Box.MarginBoxWidth()
		if naturalWidth > 0 && !isUnbounded(naturalWidth) && naturalWidth < unboundedWidthConstraints.MaxWidth {
			virtualWidth = naturalWidth
		} else {
			// Some layouts (e.g. percent/fill) report unbounded sentinels when probed with
			// unbounded width. Those aren't real intrinsic widths, so use bounded measurement.
			virtualWidth = childLayout.Box.MarginBoxWidth()
		}

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
	//
	// If a scrollbar is present, reserve its space in the border-box so the
	// content width/height isn't clipped during shrink-wrap.
	viewportWidth := min(virtualWidth, contentMaxWidth)
	viewportHeight := min(virtualHeight, contentMaxHeight)
	if actualScrollbarWidth > 0 {
		viewportWidth = min(viewportWidth+actualScrollbarWidth, contentMaxWidth)
	}
	if actualScrollbarHeight > 0 {
		viewportHeight = min(viewportHeight+actualScrollbarHeight, contentMaxHeight)
	}

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
