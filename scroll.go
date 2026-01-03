package terma

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

// renderScrollbar draws the scrollbar on the right side of the widget.
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

	// Calculate thumb size (proportional to viewport/content ratio)
	thumbHeight := (ctx.Height * ctx.Height) / contentHeight
	if thumbHeight < 1 {
		thumbHeight = 1
	}
	if thumbHeight > trackHeight {
		thumbHeight = trackHeight
	}

	// Calculate thumb position
	maxScroll := s.maxScrollOffset()
	thumbY := 0
	if maxScroll > 0 {
		thumbY = (scrollOffset * (trackHeight - thumbHeight)) / maxScroll
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

	// Draw track and thumb
	for y := 0; y < trackHeight; y++ {
		var char string
		var color Color

		if y >= thumbY && y < thumbY+thumbHeight {
			// Thumb
			char = "█"
			color = thumbColor
		} else {
			// Track
			char = "░"
			color = trackColor
		}

		ctx.DrawStyledText(scrollbarX, y, char, Style{ForegroundColor: color})
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
