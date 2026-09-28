package terma

import (
	"math"

	"github.com/darrenburns/terma/layout"
)

// Vertical scrollbar characters for smooth rendering.
// These are "lower eighths" Unicode block elements (U+2581-U+2587):
// index i fills the bottom (i+1)/8 of a cell.
var verticalScrollbarChars = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇"}

// scrollbarSubCellCount is how many steps a scrollbar cell is divided into.
const scrollbarSubCellCount = 8

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
	OffsetX Signal[int] // Current horizontal scroll offset
	Offset  Signal[int] // Current vertical scroll offset

	// position is the exact vertical scroll position in lines, which Offset
	// rounds. Content can only be drawn at whole lines, but the scrollbar
	// thumb is drawn at position, so dragging it with a pixel-precise pointer
	// moves it smoothly. It is read while painting the scrollbar, so changing
	// it alone repaints only the scrollbar.
	position Signal[float64]

	viewportWidth  int // Set by Scrollable during layout
	viewportHeight int // Set by Scrollable during layout
	contentWidth   int // Set by Scrollable during layout
	contentHeight  int // Set by Scrollable during layout

	scrollbarDragging   bool
	scrollbarDragOffset float64
	layoutCache         scrollableLayoutCache

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

	// OnScrollLeft is called when ScrollLeft is invoked with the number of columns.
	// If it returns true, the default viewport scrolling is suppressed.
	OnScrollLeft func(cols int) bool

	// OnScrollRight is called when ScrollRight is invoked with the number of columns.
	// If it returns true, the default viewport scrolling is suppressed.
	OnScrollRight func(cols int) bool
}

type scrollableLayoutCache struct {
	valid          bool
	contentWidth   int
	contentHeight  int
	contentOffsetX int
	contentOffsetY int
	scrollableY    bool
}

// NewScrollState creates a new scroll state with initial offset of 0.
// isPinned starts true so that initial content (offset 0) is considered "at bottom".
func NewScrollState() *ScrollState {
	return &ScrollState{
		OffsetX:  NewSignal(0),
		Offset:   NewSignal(0),
		position: NewSignal(0.0),
		isPinned: true,
	}
}

// GetOffsetX returns the current horizontal scroll offset (without subscribing).
func (s *ScrollState) GetOffsetX() int {
	return s.OffsetX.Peek()
}

// GetOffset returns the current scroll offset (without subscribing).
func (s *ScrollState) GetOffset() int {
	return s.Offset.Peek()
}

// SetOffsetX sets the horizontal scroll offset directly, clamping to valid bounds.
func (s *ScrollState) SetOffsetX(offset int) {
	max := s.maxOffsetX()
	if offset < 0 {
		offset = 0
	} else if offset > max {
		offset = max
	}
	s.OffsetX.Set(offset)
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
	s.position.Set(float64(offset))
}

// setPosition scrolls to an exact position in lines, clamped to valid bounds.
// Content is drawn at the nearest whole line; the scrollbar thumb follows the
// exact position.
func (s *ScrollState) setPosition(position float64) {
	position = clampFloat(position, 0, float64(s.maxOffset()))
	s.Offset.Set(int(math.Round(position)))
	s.position.Set(position)
}

// thumbPosition returns the position the scrollbar thumb shows while content
// is drawn at offset. That is the exact position, unless the offset has since
// been moved away from it (Offset is a public signal, and layout clamps it).
func (s *ScrollState) thumbPosition(offset int) float64 {
	position := s.position.Get()
	if math.Abs(position-float64(offset)) > 0.5 {
		return float64(offset)
	}
	return position
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

// ScrollLeft scrolls left by the given number of columns.
// Returns true if scrolling was handled (callback handled it or offset changed).
// Returns false if already at the left edge and no callback handled it.
func (s *ScrollState) ScrollLeft(cols int) bool {
	if s.OnScrollLeft != nil && s.OnScrollLeft(cols) {
		return true
	}
	oldOffset := s.OffsetX.Peek()
	s.SetOffsetX(oldOffset - cols)
	return s.OffsetX.Peek() != oldOffset
}

// ScrollRight scrolls right by the given number of columns.
// Returns true if scrolling was handled (callback handled it or offset changed).
// Returns false if already at the right edge and no callback handled it.
func (s *ScrollState) ScrollRight(cols int) bool {
	if s.OnScrollRight != nil && s.OnScrollRight(cols) {
		return true
	}
	oldOffset := s.OffsetX.Peek()
	s.SetOffsetX(oldOffset + cols)
	return s.OffsetX.Peek() != oldOffset
}

// maxOffset returns the maximum valid scroll offset.
func (s *ScrollState) maxOffset() int {
	max := s.contentHeight - s.viewportHeight
	if max < 0 {
		return 0
	}
	return max
}

// maxOffsetX returns the maximum valid horizontal scroll offset.
func (s *ScrollState) maxOffsetX() int {
	max := s.contentWidth - s.viewportWidth
	if max < 0 {
		return 0
	}
	return max
}

// canScrollY returns true if vertical scrolling is possible.
func (s *ScrollState) canScrollY() bool {
	return s.contentHeight > s.viewportHeight
}

// canScrollX returns true if horizontal scrolling is possible.
func (s *ScrollState) canScrollX() bool {
	return s.contentWidth > s.viewportWidth
}

func (s *ScrollState) hasHorizontalCallbacks() bool {
	return s.OnScrollLeft != nil || s.OnScrollRight != nil
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

// pinnedToEnd reports whether the next layout should show the end of the
// content: PinToBottom is on, the pin is engaged, and the last frame was
// scrolled to the bottom. The first layout is never pinned, so content that
// starts out taller than the viewport is shown from the top.
func (s *ScrollState) pinnedToEnd() bool {
	return s.PinToBottom && s.isPinned && s.contentHeight > 0 && s.IsAtBottom()
}

// updateLayout is called by Scrollable to update viewport/content dimensions.
// Note: Does not clamp offset here because Layout may be called multiple times
// with different constraints (e.g., by floating widgets). Clamping is deferred
// to Render where we have the final dimensions.
//
// If the content was pinned to the bottom, the layout has already scrolled to
// the new end (see pinnedToEnd), so the offset is brought into line with it.
func (s *ScrollState) updateLayout(viewportHeight, contentHeight int) {
	pinned := s.pinnedToEnd()
	s.viewportHeight = viewportHeight
	s.contentHeight = contentHeight
	if pinned {
		s.SetOffset(s.maxOffset())
	}
}

// updateHorizontalLayout is called by Scrollable to update horizontal dimensions.
func (s *ScrollState) updateHorizontalLayout(viewportWidth, contentWidth int) {
	s.viewportWidth = viewportWidth
	s.contentWidth = contentWidth
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
	ID            string           // Optional unique identifier for the widget
	Child         Widget           // The child widget to scroll
	State         *ScrollState     // Required - holds scroll position
	DisableScroll bool             // If true, scrolling is disabled and scrollbar hidden (default: false)
	Focusable     bool             // If true, widget can receive keyboard focus for scroll navigation
	DisableFocus  bool             // If true, prevent keyboard focus
	Width         Dimension        // Deprecated: use Style.Width
	Height        Dimension        // Deprecated: use Style.Height
	Style         Style            // Optional styling
	Click         func(MouseEvent) // Optional callback invoked when clicked
	MouseDown     func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp       func(MouseEvent) // Optional callback invoked when mouse is released
	MouseMove     func(MouseEvent) // Optional callback invoked when mouse is moved while dragging
	Hover         func(HoverEvent) // Optional callback invoked when hover state changes

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
	dims := s.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	return width, height
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
	if s.State == nil {
		return
	}

	s.State.scrollbarDragging = false
	cache := s.State.layoutCache
	if !cache.valid || !cache.scrollableY {
		return
	}

	localX, localY := s.contentCoords(event, cache)
	if !s.isOnScrollbar(localX, localY, cache) {
		return
	}

	thumb, ok := s.scrollbarThumb(cache)
	if !ok {
		return
	}

	pointerY := float64(localY) + event.SubCellY
	s.State.scrollbarDragging = true
	if thumb.contains(pointerY) {
		// Preserve the grab position within the thumb so drag feels natural.
		s.State.scrollbarDragOffset = pointerY - thumb.startCells()
		return
	}

	// Clicking the track centres the thumb on the pointer and starts dragging.
	s.State.scrollbarDragOffset = thumb.lengthCells() / 2
	s.dragScrollbar(pointerY)
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (s Scrollable) OnMouseUp(event MouseEvent) {
	if s.MouseUp != nil {
		s.MouseUp(event)
	}
	if s.State != nil {
		s.State.scrollbarDragging = false
	}
}

// OnMouseMove is called when the mouse is moved while dragging.
// Implements the MouseMoveHandler interface.
func (s Scrollable) OnMouseMove(event MouseEvent) {
	if s.MouseMove != nil {
		s.MouseMove(event)
	}
	if s.State == nil || !s.State.scrollbarDragging {
		return
	}
	cache := s.State.layoutCache
	if !cache.valid || !cache.scrollableY {
		return
	}

	_, localY := s.contentCoords(event, cache)
	s.dragScrollbar(float64(localY) + event.SubCellY)
}

// OnLayout caches layout metrics for scrollbar hit-testing and dragging.
func (s Scrollable) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if s.State == nil {
		return
	}

	box := metrics.Box()
	cache := scrollableLayoutCache{
		valid:          true,
		contentWidth:   box.ContentWidth(),
		contentHeight:  box.ContentHeight(),
		contentOffsetX: box.Border.Left + box.Padding.Left,
		contentOffsetY: box.Border.Top + box.Padding.Top,
		scrollableY:    box.IsScrollableY() && !s.DisableScroll,
	}

	s.State.layoutCache = cache
	if !cache.scrollableY {
		s.State.scrollbarDragging = false
	}
}

// OnHover is called on hover enter/leave transitions.
// Implements the Hoverable interface.
func (s Scrollable) OnHover(event HoverEvent) {
	if s.Hover != nil {
		s.Hover(event)
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

	return s.BuildContainerLayoutNode(ctx, []layout.LayoutNode{childNode})
}

func (s Scrollable) BuildContainerLayoutNode(ctx BuildContext, children []layout.LayoutNode) layout.LayoutNode {
	if len(children) == 0 {
		return &layout.BoxNode{}
	}

	childNode := children[0]

	scrollOffsetX := 0
	scrollOffsetY := 0
	pinToEnd := false
	if s.State != nil {
		scrollOffsetX = s.State.OffsetX.Get()
		scrollOffsetY = s.State.Offset.Get()
		pinToEnd = s.State.pinnedToEnd()
	}

	scrollbarWidth := 0
	if !s.DisableScroll {
		scrollbarWidth = 1
	}

	padding := toLayoutEdgeInsets(s.Style.Padding)
	border := borderToEdgeInsets(s.Style.Border)
	dims := GetWidgetDimensionSet(s)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.ScrollableNode{
		Child:           childNode,
		ScrollOffsetX:   scrollOffsetX,
		ScrollOffsetY:   scrollOffsetY,
		PinToEndY:       pinToEnd,
		ScrollbarWidth:  scrollbarWidth,
		ScrollbarHeight: 0,
		Padding:         padding,
		Border:          border,
		Margin:          toLayoutEdgeInsets(s.Style.Margin),
		MinWidth:        minWidth,
		MaxWidth:        maxWidth,
		MinHeight:       minHeight,
		MaxHeight:       maxHeight,
	})

	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{
			child:     node,
			minWidth:  dims.MinWidth,
			maxWidth:  dims.MaxWidth,
			minHeight: dims.MinHeight,
			maxHeight: dims.MaxHeight,
			padding:   padding,
			border:    border,
		}
	}

	return node
}

// getScrollOffset returns the current scroll offset.
func (s Scrollable) getScrollOffset() int {
	if s.State == nil {
		return 0
	}
	return s.State.Offset.Peek()
}

// getScrollOffsetX returns the current horizontal scroll offset.
func (s Scrollable) getScrollOffsetX() int {
	if s.State == nil {
		return 0
	}
	return s.State.OffsetX.Peek()
}

// setScrollOffset sets the scroll offset.
func (s Scrollable) setScrollOffset(offset int) {
	if s.State != nil {
		s.State.SetOffset(offset)
	}
}

// canScrollY returns true if vertical scrolling is possible.
func (s Scrollable) canScrollY() bool {
	if s.DisableScroll || s.State == nil {
		return false
	}
	return s.State.canScrollY()
}

// canScrollX returns true if horizontal scrolling is possible.
func (s Scrollable) canScrollX() bool {
	if s.DisableScroll || s.State == nil {
		return false
	}
	return s.State.canScrollX() || s.State.hasHorizontalCallbacks()
}

// canScrollAny returns true if vertical or horizontal scrolling is possible.
func (s Scrollable) canScrollAny() bool {
	return s.canScrollY() || s.canScrollX()
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
	if !s.canScrollY() {
		return false
	}
	return s.State.ScrollUp(lines)
}

// ScrollDown scrolls the content down by the given number of lines.
// Returns true if scrolling was handled, false if scroll is disabled or at limit.
func (s Scrollable) ScrollDown(lines int) bool {
	if !s.canScrollY() {
		return false
	}
	return s.State.ScrollDown(lines)
}

// ScrollLeft scrolls the content left by the given number of columns.
// Returns true if scrolling was handled, false if scroll is disabled or at limit.
func (s Scrollable) ScrollLeft(cols int) bool {
	if !s.canScrollX() {
		return false
	}
	return s.State.ScrollLeft(cols)
}

// ScrollRight scrolls the content right by the given number of columns.
// Returns true if scrolling was handled, false if scroll is disabled or at limit.
func (s Scrollable) ScrollRight(cols int) bool {
	if !s.canScrollX() {
		return false
	}
	return s.State.ScrollRight(cols)
}

func (s Scrollable) contentCoords(event MouseEvent, cache scrollableLayoutCache) (x, y int) {
	return event.LocalX - cache.contentOffsetX, event.LocalY - cache.contentOffsetY
}

func (s Scrollable) isOnScrollbar(localX, localY int, cache scrollableLayoutCache) bool {
	if cache.contentWidth <= 0 || cache.contentHeight <= 0 {
		return false
	}
	if localY < 0 || localY >= cache.contentHeight {
		return false
	}
	return localX == cache.contentWidth-1
}

// dragScrollbar places the thumb under the pointer, keeping the point where it
// was grabbed, and scrolls to the exact position that draws it there.
func (s Scrollable) dragScrollbar(pointerY float64) {
	if s.State == nil {
		return
	}

	cache := s.State.layoutCache
	if !cache.valid || !cache.scrollableY {
		return
	}

	thumb, ok := s.scrollbarThumb(cache)
	travel := thumb.travelCells()
	if !ok || travel <= 0 {
		s.State.setPosition(0)
		return
	}

	thumbStart := clampFloat(pointerY-s.State.scrollbarDragOffset, 0, travel)
	s.State.setPosition(thumbStart / travel * float64(s.maxScrollOffset()))

	if s.State.PinToBottom {
		s.State.isPinned = s.State.IsAtBottom()
	}
}

// scrollbarThumb returns the thumb as it is currently drawn, or false if the
// content can't scroll.
func (s Scrollable) scrollbarThumb(cache scrollableLayoutCache) (scrollbarThumb, bool) {
	maxScroll := s.maxScrollOffset()
	if maxScroll <= 0 || cache.contentHeight <= 0 {
		return scrollbarThumb{}, false
	}
	position := s.State.thumbPosition(s.getScrollOffset())
	return newScrollbarThumb(position, maxScroll, cache.contentHeight, s.State.contentHeight), true
}

// Render draws the scrollable widget and its child.
func (s Scrollable) Render(ctx *RenderContext) {
	// No-op - rendering is done via renderTree
}

// scrollbarThumb is the thumb's extent along the scrollbar track, in eighths
// of a cell.
type scrollbarThumb struct {
	start, length, track int
}

// newScrollbarThumb sizes and places the thumb on a track of trackHeight cells
// for content of contentHeight lines scrolled to position (of maxScroll).
// The length depends only on the track and content heights, so the thumb
// keeps exactly the same size wherever it is drawn.
func newScrollbarThumb(position float64, maxScroll, trackHeight, contentHeight int) scrollbarThumb {
	track := trackHeight * scrollbarSubCellCount
	thumb := scrollbarThumb{length: track, track: track}
	if track <= 0 || contentHeight <= trackHeight {
		return thumb
	}
	length := int(math.Round(float64(track) * float64(trackHeight) / float64(contentHeight)))
	thumb.length = min(max(length, scrollbarSubCellCount), track)
	if maxScroll > 0 {
		ratio := clampFloat(position/float64(maxScroll), 0, 1)
		thumb.start = int(math.Round(ratio * float64(thumb.track-thumb.length)))
	}
	return thumb
}

func (t scrollbarThumb) startCells() float64 {
	return float64(t.start) / scrollbarSubCellCount
}

func (t scrollbarThumb) lengthCells() float64 {
	return float64(t.length) / scrollbarSubCellCount
}

// travelCells is how far the thumb can move along the track.
func (t scrollbarThumb) travelCells() float64 {
	return float64(t.track-t.length) / scrollbarSubCellCount
}

func (t scrollbarThumb) contains(y float64) bool {
	start := t.startCells()
	return y >= start && y < start+t.lengthCells()
}

// cover returns the part of track cell y that the thumb covers, as eighths
// measured down from the top of the cell: [top, bottom).
func (t scrollbarThumb) cover(y int) (top, bottom int) {
	cellTop := y * scrollbarSubCellCount
	top = min(max(t.start-cellTop, 0), scrollbarSubCellCount)
	bottom = min(max(t.start+t.length-cellTop, 0), scrollbarSubCellCount)
	return top, bottom
}

// renderScrollbar draws the scrollbar down the right edge of ctx. The thumb is
// placed to an eighth of a cell: cells it partly covers use lower-block
// characters, and cells it fills are drawn as background colour, so the thumb
// looks the same whatever the font.
func (s Scrollable) renderScrollbar(ctx *RenderContext, scrollOffset int, focused bool) {
	if s.State == nil {
		return
	}

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

	position := s.State.thumbPosition(scrollOffset)
	thumb := newScrollbarThumb(position, s.maxScrollOffset(), trackHeight, contentHeight)
	x := ctx.Width - 1
	for y := 0; y < trackHeight; y++ {
		top, bottom := thumb.cover(y)
		switch {
		case bottom <= top:
			ctx.DrawStyledText(x, y, " ", Style{BackgroundColor: trackColor})
		case top == 0 && bottom == scrollbarSubCellCount:
			ctx.DrawStyledText(x, y, " ", Style{BackgroundColor: thumbColor})
		case top > 0:
			// The thumb starts partway down the cell and fills the rest.
			char := verticalScrollbarChars[scrollbarSubCellCount-top-1]
			ctx.DrawStyledText(x, y, char, Style{ForegroundColor: thumbColor, BackgroundColor: trackColor})
		default:
			// The thumb ends partway down the cell; the track fills the rest.
			char := verticalScrollbarChars[scrollbarSubCellCount-bottom-1]
			ctx.DrawStyledText(x, y, char, Style{ForegroundColor: trackColor, BackgroundColor: thumbColor})
		}
	}
}

// IsFocusable returns true if this widget can receive focus.
// Returns true if Focusable is set and scrolling is enabled.
// Note: We can't check canScroll() here because Layout hasn't run yet during focus collection.
func (s Scrollable) IsFocusable() bool {
	return s.Focusable && !s.DisableScroll && !s.DisableFocus
}

// OnKey handles key events when the widget is focused.
func (s Scrollable) OnKey(event KeyEvent) bool {
	if !s.canScrollAny() || s.State == nil {
		Log("Scrollable[%s].OnKey: cannot scroll, ignoring key", s.ID)
		return false
	}

	oldOffsetY := s.getScrollOffset()
	oldOffsetX := s.getScrollOffsetX()
	viewportHeight := s.State.viewportHeight

	switch {
	case event.MatchString("up", "k"):
		if !s.canScrollY() {
			return false
		}
		s.ScrollUp(1)
		Log("Scrollable[%s].OnKey: scroll up, offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("down", "j"):
		if !s.canScrollY() {
			return false
		}
		s.ScrollDown(1)
		Log("Scrollable[%s].OnKey: scroll down, offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("pgup", "pageup", "ctrl+u"):
		if !s.canScrollY() {
			return false
		}
		s.ScrollUp(viewportHeight)
		Log("Scrollable[%s].OnKey: page up, offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("pgdown", "pagedown", "ctrl+d"):
		if !s.canScrollY() {
			return false
		}
		s.ScrollDown(viewportHeight)
		Log("Scrollable[%s].OnKey: page down, offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("home", "g"):
		if !s.canScrollY() {
			return false
		}
		// Break pin when going to top
		if s.State.PinToBottom && s.State.isPinned {
			s.State.isPinned = false
		}
		s.setScrollOffset(0)
		Log("Scrollable[%s].OnKey: home, offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("end", "G"):
		if !s.canScrollY() {
			return false
		}
		maxOff := s.maxScrollOffset()
		Log("Scrollable[%s].OnKey: end BEFORE - stateViewport=%d, stateContent=%d, maxOffset=%d",
			s.ID, s.State.viewportHeight, s.State.contentHeight, maxOff)
		// Use ScrollToBottom to re-engage pin
		s.State.ScrollToBottom()
		Log("Scrollable[%s].OnKey: end AFTER - offset %d -> %d", s.ID, oldOffsetY, s.getScrollOffset())
		return true
	case event.MatchString("left", "h"):
		if !s.canScrollX() {
			return false
		}
		s.ScrollLeft(1)
		Log("Scrollable[%s].OnKey: scroll left, offsetX %d -> %d", s.ID, oldOffsetX, s.getScrollOffsetX())
		return true
	case event.MatchString("right", "l"):
		if !s.canScrollX() {
			return false
		}
		s.ScrollRight(1)
		Log("Scrollable[%s].OnKey: scroll right, offsetX %d -> %d", s.ID, oldOffsetX, s.getScrollOffsetX())
		return true
	}

	return false
}
