package terma

// scrollStateRegistry stores scroll state by widget ID to persist across renders.
// This follows the pattern used by React (useState), Flutter (State), and SwiftUI (@State)
// where state is stored externally and associated with widget identity.
var scrollStateRegistry = make(map[string]*Signal[int])

// Scrollable is a container widget that enables vertical scrolling of its child
// when the child's content exceeds the available viewport height.
// A scrollbar is displayed on the right side when scrolling is active.
type Scrollable struct {
	ID            string    // Optional unique identifier for the widget
	Child         Widget    // The child widget to scroll
	Width         Dimension // Optional width (zero value = auto)
	Height        Dimension // Optional height (zero value = auto)
	Style         Style     // Optional styling
	DisableScroll bool      // If true, scrolling is disabled (default: false = scrolling enabled)

	// Internal state - lazily initialized, persisted via scrollStateRegistry
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
// If the widget has an ID, state is persisted in the registry across renders.
func (s *Scrollable) ensureScrollOffset() {
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

// canScroll returns true if scrolling is possible (content exceeds viewport).
func (s *Scrollable) canScroll() bool {
	return !s.DisableScroll && s.contentHeight > s.viewportHeight
}

// maxScrollOffset returns the maximum valid scroll offset.
func (s *Scrollable) maxScrollOffset() int {
	max := s.contentHeight - s.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// clampScrollOffset ensures the scroll offset is within valid bounds.
func (s *Scrollable) clampScrollOffset() {
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

	// Clamp scroll offset after layout in case content shrunk
	s.clampScrollOffset()

	return Size{Width: width, Height: height}
}

// Render draws the scrollable widget and its child.
func (s *Scrollable) Render(ctx *RenderContext) {
	if s.Child == nil {
		return
	}

	s.ensureScrollOffset()
	scrollOffset := s.scrollOffset.Peek()

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
// Returns true if scrolling is enabled (not disabled) and the widget has an ID.
// Note: We can't check canScroll() here because Layout hasn't run yet during focus collection.
func (s *Scrollable) IsFocusable() bool {
	return !s.DisableScroll && s.ID != ""
}

// OnKey handles key events when the widget is focused.
func (s *Scrollable) OnKey(event KeyEvent) bool {
	if !s.canScroll() {
		Log("Scrollable[%s].OnKey: cannot scroll (content=%d, viewport=%d), ignoring key",
			s.ID, s.contentHeight, s.viewportHeight)
		return false
	}

	s.ensureScrollOffset()
	oldOffset := s.scrollOffset.Peek()

	switch {
	case event.MatchString("up", "k"):
		s.ScrollUp(1)
		Log("Scrollable[%s].OnKey: scroll up, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	case event.MatchString("down", "j"):
		s.ScrollDown(1)
		Log("Scrollable[%s].OnKey: scroll down, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	case event.MatchString("pageup", "ctrl+u"):
		s.ScrollUp(s.viewportHeight / 2)
		Log("Scrollable[%s].OnKey: page up, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	case event.MatchString("pagedown", "ctrl+d"):
		s.ScrollDown(s.viewportHeight / 2)
		Log("Scrollable[%s].OnKey: page down, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	case event.MatchString("home", "g"):
		s.scrollOffset.Set(0)
		Log("Scrollable[%s].OnKey: home, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	case event.MatchString("end", "G"):
		s.scrollOffset.Set(s.maxScrollOffset())
		Log("Scrollable[%s].OnKey: end, offset %d -> %d", s.ID, oldOffset, s.scrollOffset.Peek())
		return true
	}

	return false
}
