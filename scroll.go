package terma

// Vertical scrollbar characters for smooth rendering.
// These are "lower eighths" Unicode block elements (U+2581-U+2587).
// Index 0 = 1/8 filled from bottom, index 6 = 7/8 filled, index 7 = space.
var verticalScrollbarChars = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", " "}

const (
	scrollbarFullBlock    = "█"
	scrollbarSubCellCount = 8
)

// ScrollState holds scroll state for a Scrollable widget.
// It is the source of truth for scroll position, and must be provided to Scrollable.
// Share the same state between Scrollable and child widgets that need to
// control scroll position (e.g., List scrolling selection into view).
//
// Example usage:
//
//	scrollState := terma.NewScrollState()
//	scrollable := terma.Scrollable{State: scrollState, ...}
//	list := terma.List[string]{ScrollState: scrollState, ...}
type ScrollState struct {
	Offset         *Signal[int] // Current scroll offset
	viewportHeight int          // Set by Scrollable during layout
	contentHeight  int          // Set by Scrollable during layout

	// OnScrollUp is called when ScrollUp is invoked with the number of lines.
	// If it returns true, the default viewport scrolling is suppressed.
	// Use this for selection-first scrolling (e.g., in List widget).
	OnScrollUp func(lines int) bool

	// OnScrollDown is called when ScrollDown is invoked with the number of lines.
	// If it returns true, the default viewport scrolling is suppressed.
	// Use this for selection-first scrolling (e.g., in List widget).
	OnScrollDown func(lines int) bool
}

// NewScrollState creates a new scroll state with initial offset of 0.
func NewScrollState() *ScrollState {
	return &ScrollState{
		Offset: NewSignal(0),
	}
}

// GetOffset returns the current scroll offset (without subscribing).
func (s *ScrollState) GetOffset() int {
	return s.Offset.Peek()
}

// SetOffset sets the scroll offset directly, clamping to valid bounds.
func (s *ScrollState) SetOffset(offset int) {
	max := s.maxOffset()
	if offset < 0 {
		offset = 0
	} else if offset > max {
		offset = max
	}
	s.Offset.Set(offset)
}

// ScrollToView ensures a region (y to y+height) is visible in the viewport.
// If the region is above the viewport, scrolls up to show it at the top.
// If the region is below the viewport, scrolls down to show it at the bottom.
// If the region is already visible, does nothing.
func (s *ScrollState) ScrollToView(y, height int) {
	if s.viewportHeight <= 0 {
		return
	}

	currentOffset := s.Offset.Peek()
	regionTop := y
	regionBottom := y + height

	// Check if region is above viewport
	if regionTop < currentOffset {
		s.SetOffset(regionTop)
		return
	}

	// Check if region is below viewport
	viewportBottom := currentOffset + s.viewportHeight
	if regionBottom > viewportBottom {
		// Scroll so the region's bottom aligns with viewport bottom
		newOffset := regionBottom - s.viewportHeight
		s.SetOffset(newOffset)
	}
}

// ScrollUp scrolls up by the given number of lines.
// If OnScrollUp is set and returns true, viewport scrolling is suppressed.
func (s *ScrollState) ScrollUp(lines int) {
	if s.OnScrollUp != nil && s.OnScrollUp(lines) {
		return // Callback handled scrolling
	}
	s.SetOffset(s.Offset.Peek() - lines)
}

// ScrollDown scrolls down by the given number of lines.
// If OnScrollDown is set and returns true, viewport scrolling is suppressed.
func (s *ScrollState) ScrollDown(lines int) {
	if s.OnScrollDown != nil && s.OnScrollDown(lines) {
		return // Callback handled scrolling
	}
	s.SetOffset(s.Offset.Peek() + lines)
}

// maxOffset returns the maximum valid scroll offset.
func (s *ScrollState) maxOffset() int {
	max := s.contentHeight - s.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// canScroll returns true if scrolling is possible (content exceeds viewport).
func (s *ScrollState) canScroll() bool {
	return s.contentHeight > s.viewportHeight
}

// updateLayout is called by Scrollable to update viewport/content dimensions.
func (s *ScrollState) updateLayout(viewportHeight, contentHeight int) {
	s.viewportHeight = viewportHeight
	s.contentHeight = contentHeight
	// Clamp offset in case content shrunk
	s.SetOffset(s.Offset.Peek())
}

// Scrollable is a container widget that enables vertical scrolling of its child
// when the child's content exceeds the available viewport height.
// A scrollbar is displayed on the right side when scrolling is active.
//
// Example usage:
//
//	scrollState := terma.NewScrollState()
//	scrollable := terma.Scrollable{
//	    State:  scrollState,
//	    Height: terma.Cells(10),
//	    Child:  myContent,
//	}
type Scrollable struct {
	ID            string       // Optional unique identifier for the widget
	Child         Widget       // The child widget to scroll
	State         *ScrollState // Required - holds scroll position
	Width         Dimension    // Optional width (zero value = auto)
	Height        Dimension    // Optional height (zero value = auto)
	Style         Style        // Optional styling
	DisableScroll bool         // If true, scrolling is disabled and scrollbar hidden (default: false)
	DisableFocus  bool         // If true, widget cannot receive focus (default: false = focusable)

	// Scrollbar appearance customization
	ScrollbarThumbColor Color // Custom thumb color (default: White unfocused, BrightCyan focused)
	ScrollbarTrackColor Color // Custom track color (default: BrightBlack)
}

// WidgetID returns the widget's unique identifier.
// Implements the Identifiable interface.
func (s Scrollable) WidgetID() string {
	return s.ID
}

// GetDimensions returns the width and height dimension preferences.
func (s Scrollable) GetDimensions() (width, height Dimension) {
	return s.Width, s.Height
}

// GetStyle returns the style of the scrollable widget.
func (s Scrollable) GetStyle() Style {
	return s.Style
}

// Build returns itself as Scrollable manages its own child.
func (s Scrollable) Build(ctx BuildContext) Widget {
	return s
}

// getScrollOffset returns the current scroll offset.
func (s Scrollable) getScrollOffset() int {
	if s.State == nil {
		return 0
	}
	return s.State.Offset.Peek()
}

// setScrollOffset sets the scroll offset.
func (s Scrollable) setScrollOffset(offset int) {
	if s.State != nil {
		s.State.SetOffset(offset)
	}
}

// canScroll returns true if scrolling is possible (content exceeds viewport).
func (s Scrollable) canScroll() bool {
	if s.DisableScroll || s.State == nil {
		return false
	}
	return s.State.canScroll()
}

// maxScrollOffset returns the maximum valid scroll offset.
func (s Scrollable) maxScrollOffset() int {
	if s.State == nil {
		return 0
	}
	return s.State.maxOffset()
}

// clampScrollOffset ensures the scroll offset is within valid bounds.
func (s Scrollable) clampScrollOffset() {
	if s.State != nil {
		s.State.SetOffset(s.State.Offset.Peek())
	}
}

// ScrollUp scrolls the content up by the given number of lines.
func (s Scrollable) ScrollUp(lines int) {
	if !s.canScroll() {
		return
	}
	s.State.ScrollUp(lines)
}

// ScrollDown scrolls the content down by the given number of lines.
func (s Scrollable) ScrollDown(lines int) {
	if !s.canScroll() {
		return
	}
	s.State.ScrollDown(lines)
}

// Layout computes the size of the scrollable widget.
func (s Scrollable) Layout(ctx BuildContext, constraints Constraints) Size {
	if s.Child == nil || s.State == nil {
		return Size{Width: constraints.MinWidth, Height: constraints.MinHeight}
	}

	// Build the child first
	built := s.Child.Build(ctx)

	// Determine if we need space for scrollbar
	scrollbarWidth := 0
	if !s.DisableScroll {
		scrollbarWidth = 1 // Reserve space for scrollbar
	}

	// Get child's style insets (RenderChild will apply these during render)
	var childHInset, childVInset int
	if styled, ok := built.(Styled); ok {
		style := styled.GetStyle()
		borderWidth := style.Border.Width()
		childHInset = style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
		childVInset = style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
	}

	// Calculate available content width (subtract scrollbar width only)
	// Note: RenderChild will handle the child's own style insets (border, padding, margin)
	contentMaxWidth := constraints.MaxWidth - scrollbarWidth
	if contentMaxWidth < 0 {
		contentMaxWidth = 0
	}

	// Layout child with UNBOUNDED height to get natural content height
	childWidth := contentMaxWidth
	var contentHeight int
	if layoutable, ok := built.(Layoutable); ok {
		childConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  contentMaxWidth,
			MinHeight: 0,
			MaxHeight: 100000, // Large value to allow natural height
		}
		size := layoutable.Layout(ctx, childConstraints)
		childWidth = size.Width
		// Add child's vertical insets since RenderChild will apply them
		contentHeight = size.Height + childVInset
	} else {
		contentHeight = constraints.MaxHeight
	}

	// Determine viewport dimensions
	var width int
	switch {
	case s.Width.IsCells():
		width = s.Width.CellsValue()
	case s.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		// Use child's natural width plus scrollbar and child's horizontal insets
		width = childWidth + scrollbarWidth + childHInset
	}

	var height int
	switch {
	case s.Height.IsCells():
		height = s.Height.CellsValue()
	case s.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		// Use natural content height, but clamp to constraints
		height = contentHeight
	}

	// Clamp to constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}

	// Update state with layout info
	s.State.updateLayout(height, contentHeight)

	// Clamp scroll offset after layout in case content shrunk
	s.clampScrollOffset()

	return Size{Width: width, Height: height}
}

// Render draws the scrollable widget and its child.
func (s Scrollable) Render(ctx *RenderContext) {
	if s.Child == nil || s.State == nil {
		return
	}

	scrollOffset := s.getScrollOffset()
	contentHeight := s.State.contentHeight

	// Determine if we need to show scrollbar
	needsScrollbar := s.canScroll()
	scrollbarWidth := 0
	if needsScrollbar {
		scrollbarWidth = 1
	}

	// Content area width (excluding scrollbar)
	contentWidth := ctx.Width - scrollbarWidth

	// Create a scrolled sub-context for the child
	childCtx := ctx.ScrolledSubContext(0, 0, contentWidth, ctx.Height, scrollOffset, contentHeight)

	// Render the child through the scrolled context
	childCtx.RenderChild(0, s.Child, 0, 0, contentWidth, contentHeight)

	// Render scrollbar if needed (on the main context, not scrolled)
	if needsScrollbar {
		focused := ctx.IsFocused(s)
		s.renderScrollbar(ctx, scrollOffset, focused)
	}
}

// scrollbarThumbMetrics calculates smooth scrollbar thumb position and size.
// Returns position and size as floats for sub-cell precision.
func scrollbarThumbMetrics(scrollOffset, maxScroll, viewportHeight, contentHeight int) (position, size float64) {
	if contentHeight <= 0 || viewportHeight <= 0 {
		return 0, float64(viewportHeight)
	}

	// Calculate thumb size proportional to viewport/content ratio
	sizeRatio := float64(viewportHeight) / float64(contentHeight)
	size = float64(viewportHeight) * sizeRatio
	if size < 1.0 {
		size = 1.0
	}
	if size > float64(viewportHeight) {
		size = float64(viewportHeight)
	}

	// Calculate thumb position within available track space
	if maxScroll > 0 {
		availableTrack := float64(viewportHeight) - size
		positionRatio := float64(scrollOffset) / float64(maxScroll)
		position = availableTrack * positionRatio
	}

	return position, size
}

// getTopEdgeChar returns the character for the top edge of the scrollbar thumb.
// startSubOffset indicates how far into the cell the thumb starts (0-7).
// Since lower-eighth blocks fill from bottom, we return the complementary block.
func getTopEdgeChar(startSubOffset int) string {
	if startSubOffset <= 0 {
		return scrollbarFullBlock
	}
	if startSubOffset >= scrollbarSubCellCount {
		return " "
	}
	// Thumb fills from bottom, so invert: offset 2 means 6/8 filled = index 5
	return verticalScrollbarChars[scrollbarSubCellCount-1-startSubOffset]
}

// getBottomEdgeChar returns the character for the bottom edge of the scrollbar thumb.
// endSubOffset indicates how far into the cell the thumb extends from the top (0-8).
// We draw track color filling from the bottom, so we need the complement.
func getBottomEdgeChar(endSubOffset int) string {
	if endSubOffset >= scrollbarSubCellCount {
		// Thumb fills entire cell - return space (shows background = thumb)
		return " "
	}
	if endSubOffset <= 0 {
		// Thumb doesn't extend into this cell - return full block (shows foreground = track)
		return scrollbarFullBlock
	}
	// Thumb extends endSubOffset/8 from top, track fills (8-endSubOffset)/8 from bottom
	trackFill := scrollbarSubCellCount - endSubOffset
	return verticalScrollbarChars[trackFill-1]
}

// renderScrollbar draws the scrollbar on the right side of the widget.
// Uses Unicode lower-eighth block characters for sub-cell precision.
func (s Scrollable) renderScrollbar(ctx *RenderContext, scrollOffset int, focused bool) {
	if s.State == nil {
		return
	}

	scrollbarX := ctx.Width - 1
	trackHeight := ctx.Height
	contentHeight := s.State.contentHeight

	if trackHeight <= 0 || contentHeight <= 0 {
		return
	}

	// Determine scrollbar colors based on focus state and custom settings
	var trackColor, thumbColor Color
	if s.ScrollbarTrackColor.IsSet() {
		trackColor = s.ScrollbarTrackColor
	} else {
		trackColor = BrightBlack
	}
	if s.ScrollbarThumbColor.IsSet() {
		thumbColor = s.ScrollbarThumbColor
	} else if focused {
		thumbColor = BrightCyan
	} else {
		thumbColor = White
	}

	// Calculate thumb position and size with floating-point precision
	maxScroll := s.maxScrollOffset()
	thumbPos, thumbSize := scrollbarThumbMetrics(scrollOffset, maxScroll, trackHeight, contentHeight)

	// Convert to sub-cell units (multiply by 8)
	startSubCell := thumbPos * float64(scrollbarSubCellCount)
	endSubCell := (thumbPos + thumbSize) * float64(scrollbarSubCellCount)

	// Get cell indices and sub-cell offsets
	startCellIndex := int(startSubCell) / scrollbarSubCellCount
	startSubOffset := int(startSubCell) % scrollbarSubCellCount
	endCellIndex := int(endSubCell) / scrollbarSubCellCount
	endSubOffset := int(endSubCell) % scrollbarSubCellCount

	// Draw each cell of the track
	for y := 0; y < trackHeight; y++ {
		var char string
		var style Style

		// Check if this cell is outside the thumb
		// Note: when endSubOffset=0, thumb ends exactly at cell boundary, so endCellIndex is outside
		outsideThumb := y < startCellIndex || y > endCellIndex || (y == endCellIndex && endSubOffset == 0)

		if outsideThumb {
			// Track (outside thumb) - use space with background color for consistent appearance
			char = " "
			style = Style{BackgroundColor: trackColor}
		} else if y == startCellIndex && y == endCellIndex {
			// Thumb fits within single cell
			fillAmount := endSubOffset - startSubOffset
			if fillAmount <= 0 {
				char = " "
			} else if fillAmount >= scrollbarSubCellCount {
				char = scrollbarFullBlock
			} else {
				char = verticalScrollbarChars[fillAmount-1]
			}
			style = Style{ForegroundColor: thumbColor, BackgroundColor: trackColor}
		} else if y == startCellIndex {
			// Top edge of thumb - partial block, thumb fills from bottom
			char = getTopEdgeChar(startSubOffset)
			style = Style{ForegroundColor: thumbColor, BackgroundColor: trackColor}
		} else if y == endCellIndex {
			// Bottom edge of thumb - partial block, thumb is on top
			// Use Reverse so terminal swaps fg/bg, matching how track renders
			char = getBottomEdgeChar(endSubOffset)
			style = Style{ForegroundColor: thumbColor, BackgroundColor: trackColor, Reverse: true}
		} else {
			// Middle of thumb - full block
			char = scrollbarFullBlock
			style = Style{ForegroundColor: thumbColor}
		}

		ctx.DrawStyledText(scrollbarX, y, char, style)
	}
}

// IsFocusable returns true if this widget can receive focus.
// Returns true if scrolling is enabled, focus is not disabled, and the widget has an ID.
// Set DisableFocus=true to prevent focus while still showing the scrollbar.
// Note: We can't check canScroll() here because Layout hasn't run yet during focus collection.
func (s Scrollable) IsFocusable() bool {
	return !s.DisableScroll && !s.DisableFocus && s.ID != ""
}

// OnKey handles key events when the widget is focused.
func (s Scrollable) OnKey(event KeyEvent) bool {
	if !s.canScroll() || s.State == nil {
		Log("Scrollable[%s].OnKey: cannot scroll, ignoring key", s.ID)
		return false
	}

	oldOffset := s.getScrollOffset()
	viewportHeight := s.State.viewportHeight

	switch {
	case event.MatchString("up", "k"):
		s.ScrollUp(1)
		Log("Scrollable[%s].OnKey: scroll up, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("down", "j"):
		s.ScrollDown(1)
		Log("Scrollable[%s].OnKey: scroll down, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("pageup", "ctrl+u"):
		s.ScrollUp(viewportHeight / 2)
		Log("Scrollable[%s].OnKey: page up, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("pagedown", "ctrl+d"):
		s.ScrollDown(viewportHeight / 2)
		Log("Scrollable[%s].OnKey: page down, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("home", "g"):
		s.setScrollOffset(0)
		Log("Scrollable[%s].OnKey: home, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("end", "G"):
		s.setScrollOffset(s.maxScrollOffset())
		Log("Scrollable[%s].OnKey: end, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	}

	return false
}
