package terma

import "terma/layout"

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
	Offset         Signal[int] // Current scroll offset
	viewportHeight int         // Set by Scrollable during layout
	contentHeight  int         // Set by Scrollable during layout

	// PinToBottom enables auto-scroll when content grows while at bottom.
	// Scrolling up breaks the pin; scrolling to bottom re-engages it.
	PinToBottom bool
	isPinned    bool // internal: tracks current pinned state

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
// isPinned starts true so that initial content (offset 0) is considered "at bottom".
func NewScrollState() *ScrollState {
	return &ScrollState{
		Offset:   NewSignal(0),
		isPinned: true,
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
// Returns true if scrolling was handled (callback handled it or offset changed).
// Returns false if already at the top and no callback handled it.
// If OnScrollUp is set and returns true, viewport scrolling is suppressed.
// If PinToBottom is enabled, scrolling up breaks the pin.
func (s *ScrollState) ScrollUp(lines int) bool {
	if s.OnScrollUp != nil && s.OnScrollUp(lines) {
		return true // Callback handled scrolling
	}
	// Break pin when user scrolls up
	if s.PinToBottom && s.isPinned {
		s.isPinned = false
	}
	oldOffset := s.Offset.Peek()
	s.SetOffset(oldOffset - lines)
	return s.Offset.Peek() != oldOffset
}

// ScrollDown scrolls down by the given number of lines.
// Returns true if scrolling was handled (callback handled it or offset changed).
// Returns false if already at the bottom and no callback handled it.
// If OnScrollDown is set and returns true, viewport scrolling is suppressed.
// If PinToBottom is enabled, reaching the bottom re-engages the pin.
func (s *ScrollState) ScrollDown(lines int) bool {
	if s.OnScrollDown != nil && s.OnScrollDown(lines) {
		return true // Callback handled scrolling
	}
	oldOffset := s.Offset.Peek()
	s.SetOffset(oldOffset + lines)
	// Re-engage pin when reaching bottom
	if s.PinToBottom && s.IsAtBottom() {
		s.isPinned = true
	}
	return s.Offset.Peek() != oldOffset
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

// IsAtBottom returns true if currently scrolled to the bottom.
func (s *ScrollState) IsAtBottom() bool {
	return s.Offset.Peek() >= s.maxOffset()
}

// IsPinned returns true if PinToBottom is enabled and currently pinned.
func (s *ScrollState) IsPinned() bool {
	return s.PinToBottom && s.isPinned
}

// ScrollToBottom scrolls to the bottom and re-engages the pin if PinToBottom is enabled.
func (s *ScrollState) ScrollToBottom() {
	s.SetOffset(s.maxOffset())
	if s.PinToBottom {
		s.isPinned = true
	}
}

// updateLayout is called by Scrollable to update viewport/content dimensions.
// Note: Does not clamp offset here because Layout may be called multiple times
// with different constraints (e.g., by floating widgets). Clamping is deferred
// to Render where we have the final dimensions.
// If PinToBottom is enabled and pinned, auto-scrolls when content grows.
func (s *ScrollState) updateLayout(viewportHeight, contentHeight int) {
	oldContentHeight := s.contentHeight
	s.viewportHeight = viewportHeight
	s.contentHeight = contentHeight

	// Auto-scroll to bottom when pinned and content grows
	if s.PinToBottom && s.isPinned && contentHeight > oldContentHeight && oldContentHeight > 0 {
		s.Offset.Set(s.maxOffset())
	}
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
	DisableScroll bool         // If true, scrolling is disabled and scrollbar hidden (default: false)
	DisableFocus  bool         // If true, widget cannot receive focus (default: false = focusable)
	Width         Dimension    // Optional width (zero value = auto)
	Height        Dimension    // Optional height (zero value = auto)
	Style         Style        // Optional styling
	Click         func(MouseEvent) // Optional callback invoked when clicked
	MouseDown     func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp       func(MouseEvent) // Optional callback invoked when mouse is released
	Hover         func(bool)   // Optional callback invoked when hover state changes

	// Scrollbar appearance customization
	ScrollbarThumbColor Color // Custom thumb color (default: White unfocused, BrightCyan focused)
	ScrollbarTrackColor Color // Custom track color (default: BrightBlack)
}

// WidgetID returns the widget's unique identifier.
// Implements the Identifiable interface.
func (s Scrollable) WidgetID() string {
	return s.ID
}

// GetContentDimensions returns the width and height dimension preferences.
func (s Scrollable) GetContentDimensions() (width, height Dimension) {
	return s.Width, s.Height
}

// GetStyle returns the style of the scrollable widget.
// Implements the Styled interface.
func (s Scrollable) GetStyle() Style {
	return s.Style
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (s Scrollable) OnClick(event MouseEvent) {
	if s.Click != nil {
		s.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (s Scrollable) OnMouseDown(event MouseEvent) {
	if s.MouseDown != nil {
		s.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (s Scrollable) OnMouseUp(event MouseEvent) {
	if s.MouseUp != nil {
		s.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (s Scrollable) OnHover(hovered bool) {
	if s.Hover != nil {
		s.Hover(hovered)
	}
}

// Build returns itself as Scrollable manages its own child.
func (s Scrollable) Build(ctx BuildContext) Widget {
	return s
}

// BuildLayoutNode builds a layout node for this Scrollable widget.
// Implements the LayoutNodeBuilder interface.
func (s Scrollable) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	if s.Child == nil {
		// No child - return empty box
		return &layout.BoxNode{}
	}

	// Create child context once and reuse
	childCtx := ctx.PushChild(0)

	// Build the child widget
	built := s.Child.Build(childCtx)

	// Get child's layout node
	var childNode layout.LayoutNode
	if builder, ok := built.(LayoutNodeBuilder); ok {
		childNode = builder.BuildLayoutNode(childCtx)
	} else {
		childNode = buildFallbackLayoutNode(built, childCtx)
	}

	// Get scroll offset from state
	// Use Get() to subscribe to offset changes so PinToBottom auto-scroll triggers re-render
	scrollOffsetY := 0
	if s.State != nil {
		scrollOffsetY = s.State.Offset.Get()
	}

	// Scrollbar width: 1 if scrolling enabled, 0 if disabled
	scrollbarWidth := 0
	if !s.DisableScroll {
		scrollbarWidth = 1
	}

	// Get content-box constraints from dimensions
	minWidth, maxWidth := dimensionToMinMax(s.Width)
	minHeight, maxHeight := dimensionToMinMax(s.Height)

	padding := toLayoutEdgeInsets(s.Style.Padding)
	border := borderToEdgeInsets(s.Style.Border)

	// Add padding and border to convert content-box to border-box constraints
	hInset := padding.Horizontal() + border.Horizontal()
	vInset := padding.Vertical() + border.Vertical()
	if minWidth > 0 {
		minWidth += hInset
	}
	if maxWidth > 0 {
		maxWidth += hInset
	}
	if minHeight > 0 {
		minHeight += vInset
	}
	if maxHeight > 0 {
		maxHeight += vInset
	}

	return &layout.ScrollableNode{
		Child:           childNode,
		ScrollOffsetY:   scrollOffsetY,
		ScrollbarWidth:  scrollbarWidth,
		ScrollbarHeight: 0, // Horizontal scrolling not supported yet
		Padding:         padding,
		Border:          border,
		Margin:          toLayoutEdgeInsets(s.Style.Margin),
		MinWidth:        minWidth,
		MaxWidth:        maxWidth,
		MinHeight:       minHeight,
		MaxHeight:       maxHeight,
	}
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

// ScrollUp scrolls the content up by the given number of lines.
// Returns true if scrolling was handled, false if scroll is disabled or at limit.
func (s Scrollable) ScrollUp(lines int) bool {
	if !s.canScroll() {
		return false
	}
	return s.State.ScrollUp(lines)
}

// ScrollDown scrolls the content down by the given number of lines.
// Returns true if scrolling was handled, false if scroll is disabled or at limit.
func (s Scrollable) ScrollDown(lines int) bool {
	if !s.canScroll() {
		return false
	}
	return s.State.ScrollDown(lines)
}

// Render draws the scrollable widget and its child.
func (s Scrollable) Render(ctx *RenderContext) {
	// No-op - rendering is done via renderTree
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
	theme := getTheme()
	var trackColor, thumbColor Color
	if s.ScrollbarTrackColor.IsSet() {
		trackColor = s.ScrollbarTrackColor
	} else {
		trackColor = theme.ScrollbarTrack
	}
	if s.ScrollbarThumbColor.IsSet() {
		thumbColor = s.ScrollbarThumbColor
	} else if focused {
		thumbColor = theme.Primary
	} else {
		thumbColor = theme.ScrollbarThumb
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
		// Break pin when going to top
		if s.State.PinToBottom && s.State.isPinned {
			s.State.isPinned = false
		}
		s.setScrollOffset(0)
		Log("Scrollable[%s].OnKey: home, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("end", "G"):
		maxOff := s.maxScrollOffset()
		Log("Scrollable[%s].OnKey: end BEFORE - stateViewport=%d, stateContent=%d, maxOffset=%d",
			s.ID, s.State.viewportHeight, s.State.contentHeight, maxOff)
		// Use ScrollToBottom to re-engage pin
		s.State.ScrollToBottom()
		Log("Scrollable[%s].OnKey: end AFTER - offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	}

	return false
}
