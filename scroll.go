package terma

// scrollStateRegistry stores scroll state by widget ID to persist across renders.
// This follows the pattern used by React (useState), Flutter (State), and SwiftUI (@State)
// where state is stored externally and associated with widget identity.
var scrollStateRegistry = make(map[string]*Signal[int])

// ScrollController provides programmatic control over a Scrollable.
// Share the same controller between Scrollable and widgets that need
// to control scroll position (e.g., ListView scrolling selection into view).
//
// Example usage:
//
//	controller := terma.NewScrollController()
//	// Pass to both Scrollable and a child that needs to control scrolling
//	scrollable := &terma.Scrollable{Controller: controller, ...}
//	listView := &MyListView{Controller: controller, ...}
//	// In MyListView, call controller.ScrollToView(y, height) when selection changes
type ScrollController struct {
	offset         *Signal[int]
	viewportHeight int // Set by Scrollable during layout
	contentHeight  int // Set by Scrollable during layout

	// OnScrollUp is called when ScrollUp is invoked with the number of lines.
	// If it returns true, the default viewport scrolling is suppressed.
	// Use this for selection-first scrolling (e.g., in List widget).
	OnScrollUp func(lines int) bool

	// OnScrollDown is called when ScrollDown is invoked with the number of lines.
	// If it returns true, the default viewport scrolling is suppressed.
	// Use this for selection-first scrolling (e.g., in List widget).
	OnScrollDown func(lines int) bool
}

// NewScrollController creates a new scroll controller with initial offset of 0.
func NewScrollController() *ScrollController {
	return &ScrollController{
		offset: NewSignal(0),
	}
}

// Offset returns the current scroll offset.
func (c *ScrollController) Offset() int {
	return c.offset.Peek()
}

// SetOffset sets the scroll offset directly, clamping to valid bounds.
func (c *ScrollController) SetOffset(offset int) {
	max := c.maxOffset()
	if offset < 0 {
		offset = 0
	} else if offset > max {
		offset = max
	}
	c.offset.Set(offset)
}

// ScrollToView ensures a region (y to y+height) is visible in the viewport.
// If the region is above the viewport, scrolls up to show it at the top.
// If the region is below the viewport, scrolls down to show it at the bottom.
// If the region is already visible, does nothing.
func (c *ScrollController) ScrollToView(y, height int) {
	if c.viewportHeight <= 0 {
		return
	}

	currentOffset := c.offset.Peek()
	regionTop := y
	regionBottom := y + height

	// Check if region is above viewport
	if regionTop < currentOffset {
		c.SetOffset(regionTop)
		return
	}

	// Check if region is below viewport
	viewportBottom := currentOffset + c.viewportHeight
	if regionBottom > viewportBottom {
		// Scroll so the region's bottom aligns with viewport bottom
		newOffset := regionBottom - c.viewportHeight
		c.SetOffset(newOffset)
	}
}

// ScrollUp scrolls up by the given number of lines.
// If OnScrollUp is set and returns true, viewport scrolling is suppressed.
func (c *ScrollController) ScrollUp(lines int) {
	if c.OnScrollUp != nil && c.OnScrollUp(lines) {
		return // Callback handled scrolling
	}
	c.SetOffset(c.offset.Peek() - lines)
}

// ScrollDown scrolls down by the given number of lines.
// If OnScrollDown is set and returns true, viewport scrolling is suppressed.
func (c *ScrollController) ScrollDown(lines int) {
	if c.OnScrollDown != nil && c.OnScrollDown(lines) {
		return // Callback handled scrolling
	}
	c.SetOffset(c.offset.Peek() + lines)
}

// maxOffset returns the maximum valid scroll offset.
func (c *ScrollController) maxOffset() int {
	max := c.contentHeight - c.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// canScroll returns true if scrolling is possible (content exceeds viewport).
func (c *ScrollController) canScroll() bool {
	return c.contentHeight > c.viewportHeight
}

// updateLayout is called by Scrollable to update viewport/content dimensions.
func (c *ScrollController) updateLayout(viewportHeight, contentHeight int) {
	c.viewportHeight = viewportHeight
	c.contentHeight = contentHeight
	// Clamp offset in case content shrunk
	c.SetOffset(c.offset.Peek())
}

// Scrollable is a container widget that enables vertical scrolling of its child
// when the child's content exceeds the available viewport height.
// A scrollbar is displayed on the right side when scrolling is active.
type Scrollable struct {
	ID            string    // Optional unique identifier for the widget
	Child         Widget    // The child widget to scroll
	Width         Dimension // Optional width (zero value = auto)
	Height        Dimension // Optional height (zero value = auto)
	Style         Style     // Optional styling
	DisableScroll bool      // If true, scrolling is disabled and scrollbar hidden (default: false)
	DisableFocus  bool      // If true, widget cannot receive focus (default: false = focusable)

	// Controller provides programmatic scroll control. If provided, it takes
	// precedence over internal state management. Share this with child widgets
	// (like ListView) that need to scroll items into view.
	Controller *ScrollController

	// Internal state - lazily initialized, persisted via scrollStateRegistry
	// Only used when Controller is nil.
	scrollOffset   *Signal[int]
	contentHeight  int // Cached content height from last layout
	viewportHeight int // Cached viewport height from last layout
}

// Key returns the widget's unique identifier.
func (s *Scrollable) Key() string {
	return s.ID
}

// GetDimensions returns the width and height dimension preferences.
func (s *Scrollable) GetDimensions() (width, height Dimension) {
	return s.Width, s.Height
}

// GetStyle returns the style of the scrollable widget.
func (s *Scrollable) GetStyle() Style {
	return s.Style
}

// Build returns itself as Scrollable manages its own child.
func (s *Scrollable) Build(ctx BuildContext) Widget {
	return s
}

// ensureScrollOffset initializes the scroll offset signal if needed.
// If a Controller is provided, this is a no-op (controller manages its own state).
// If the widget has an ID, state is persisted in the registry across renders.
func (s *Scrollable) ensureScrollOffset() {
	// If we have a controller, it manages the offset
	if s.Controller != nil {
		return
	}

	if s.scrollOffset != nil {
		return
	}

	// Look up existing state by ID to persist across renders
	if s.ID != "" {
		if existing, ok := scrollStateRegistry[s.ID]; ok {
			s.scrollOffset = existing
			return
		}
	}

	// Create new state
	s.scrollOffset = NewSignal(0)

	// Register it for persistence if we have an ID
	if s.ID != "" {
		scrollStateRegistry[s.ID] = s.scrollOffset
	}
}

// getScrollOffset returns the current scroll offset from either the controller or internal state.
func (s *Scrollable) getScrollOffset() int {
	if s.Controller != nil {
		return s.Controller.Offset()
	}
	s.ensureScrollOffset()
	return s.scrollOffset.Peek()
}

// setScrollOffset sets the scroll offset on either the controller or internal state.
func (s *Scrollable) setScrollOffset(offset int) {
	if s.Controller != nil {
		s.Controller.SetOffset(offset)
		return
	}
	s.ensureScrollOffset()
	s.scrollOffset.Set(offset)
}

// canScroll returns true if scrolling is possible (content exceeds viewport).
func (s *Scrollable) canScroll() bool {
	if s.DisableScroll {
		return false
	}
	if s.Controller != nil {
		return s.Controller.canScroll()
	}
	return s.contentHeight > s.viewportHeight
}

// maxScrollOffset returns the maximum valid scroll offset.
func (s *Scrollable) maxScrollOffset() int {
	if s.Controller != nil {
		return s.Controller.maxOffset()
	}
	max := s.contentHeight - s.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// clampScrollOffset ensures the scroll offset is within valid bounds.
func (s *Scrollable) clampScrollOffset() {
	if s.Controller != nil {
		// Controller clamps automatically when offset is set
		s.Controller.SetOffset(s.Controller.Offset())
		return
	}
	s.ensureScrollOffset()
	offset := s.scrollOffset.Peek()
	max := s.maxScrollOffset()
	if offset < 0 {
		s.scrollOffset.Set(0)
	} else if offset > max {
		s.scrollOffset.Set(max)
	}
}

// ScrollUp scrolls the content up by the given number of lines.
func (s *Scrollable) ScrollUp(lines int) {
	if !s.canScroll() {
		return
	}
	if s.Controller != nil {
		s.Controller.ScrollUp(lines)
		return
	}
	s.ensureScrollOffset()
	newOffset := s.scrollOffset.Peek() - lines
	if newOffset < 0 {
		newOffset = 0
	}
	s.scrollOffset.Set(newOffset)
}

// ScrollDown scrolls the content down by the given number of lines.
func (s *Scrollable) ScrollDown(lines int) {
	if !s.canScroll() {
		return
	}
	if s.Controller != nil {
		s.Controller.ScrollDown(lines)
		return
	}
	s.ensureScrollOffset()
	newOffset := s.scrollOffset.Peek() + lines
	max := s.maxScrollOffset()
	if newOffset > max {
		newOffset = max
	}
	s.scrollOffset.Set(newOffset)
}

// Layout computes the size of the scrollable widget.
func (s *Scrollable) Layout(constraints Constraints) Size {
	if s.Child == nil {
		return Size{Width: constraints.MinWidth, Height: constraints.MinHeight}
	}

	// Build the child first
	built := s.Child.Build(BuildContext{})

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
	if layoutable, ok := built.(Layoutable); ok {
		childConstraints := Constraints{
			MinWidth:  0,
			MaxWidth:  contentMaxWidth,
			MinHeight: 0,
			MaxHeight: 100000, // Large value to allow natural height
		}
		size := layoutable.Layout(childConstraints)
		childWidth = size.Width
		// Add child's vertical insets since RenderChild will apply them
		s.contentHeight = size.Height + childVInset
	} else {
		s.contentHeight = constraints.MaxHeight
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
		height = s.contentHeight
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

	s.viewportHeight = height

	// Update controller with layout info if present
	if s.Controller != nil {
		s.Controller.updateLayout(s.viewportHeight, s.contentHeight)
	}

	// Clamp scroll offset after layout in case content shrunk
	s.clampScrollOffset()

	return Size{Width: width, Height: height}
}

// Render draws the scrollable widget and its child.
func (s *Scrollable) Render(ctx *RenderContext) {
	if s.Child == nil {
		return
	}

	scrollOffset := s.getScrollOffset()

	// Determine if we need to show scrollbar
	needsScrollbar := s.canScroll()
	scrollbarWidth := 0
	if needsScrollbar {
		scrollbarWidth = 1
	}

	// Content area width (excluding scrollbar)
	contentWidth := ctx.Width - scrollbarWidth

	// Create a scrolled sub-context for the child
	childCtx := ctx.ScrolledSubContext(0, 0, contentWidth, ctx.Height, scrollOffset, s.contentHeight)

	// Render the child through the scrolled context
	childCtx.RenderChild(0, s.Child, 0, 0, contentWidth, s.contentHeight)

	// Render scrollbar if needed (on the main context, not scrolled)
	if needsScrollbar {
		focused := ctx.IsFocused(s)
		s.renderScrollbar(ctx, scrollOffset, focused)
	}
}

// renderScrollbar draws the scrollbar on the right side of the widget.
func (s *Scrollable) renderScrollbar(ctx *RenderContext, scrollOffset int, focused bool) {
	scrollbarX := ctx.Width - 1
	trackHeight := ctx.Height

	if trackHeight <= 0 || s.contentHeight <= 0 {
		return
	}

	// Calculate thumb size (proportional to viewport/content ratio)
	thumbHeight := (ctx.Height * ctx.Height) / s.contentHeight
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

	// Determine scrollbar colors based on focus state
	var trackColor, thumbColor Color
	if focused {
		trackColor = BrightBlack
		thumbColor = BrightCyan
	} else {
		trackColor = BrightBlack
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
func (s *Scrollable) IsFocusable() bool {
	return !s.DisableScroll && !s.DisableFocus && s.ID != ""
}

// OnKey handles key events when the widget is focused.
func (s *Scrollable) OnKey(event KeyEvent) bool {
	if !s.canScroll() {
		Log("Scrollable[%s].OnKey: cannot scroll (content=%d, viewport=%d), ignoring key",
			s.ID, s.contentHeight, s.viewportHeight)
		return false
	}

	oldOffset := s.getScrollOffset()

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
		s.ScrollUp(s.viewportHeight / 2)
		Log("Scrollable[%s].OnKey: page up, offset %d -> %d", s.ID, oldOffset, s.getScrollOffset())
		return true
	case event.MatchString("pagedown", "ctrl+d"):
		s.ScrollDown(s.viewportHeight / 2)
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
