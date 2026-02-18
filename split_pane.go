package terma

import "github.com/darrenburns/terma/layout"

// SplitPaneOrientation determines how the split divides space.
type SplitPaneOrientation int

const (
	// SplitHorizontal divides left/right.
	SplitHorizontal SplitPaneOrientation = iota
	// SplitVertical divides top/bottom.
	SplitVertical
)

const splitPaneKeyStep = 0.05

// SplitPaneState holds the divider position for a SplitPane widget.
type SplitPaneState struct {
	DividerPosition Signal[float64] // 0.0-1.0

	dragging   bool
	dragOffset int

	layoutCache splitPaneLayoutCache
}

type splitPaneLayoutCache struct {
	valid          bool
	contentWidth   int
	contentHeight  int
	contentOffsetX int
	contentOffsetY int
	dividerPos     int
	dividerSize    int
	minPos         float64
	maxPos         float64
	orientation    SplitPaneOrientation
}

// NewSplitPaneState creates a new SplitPaneState with the given initial position.
func NewSplitPaneState(initialPosition float64) *SplitPaneState {
	if initialPosition <= 0 || initialPosition >= 1 {
		initialPosition = 0.5
	}
	return &SplitPaneState{
		DividerPosition: NewSignal(initialPosition),
	}
}

// SetPosition sets the divider position (clamped to valid range).
func (s *SplitPaneState) SetPosition(pos float64) {
	if s == nil || !s.DividerPosition.IsValid() {
		return
	}
	s.DividerPosition.Set(s.clampPosition(pos))
}

// GetPosition returns the current divider position without subscribing.
func (s *SplitPaneState) GetPosition() float64 {
	if s == nil || !s.DividerPosition.IsValid() {
		return 0.5
	}
	return s.DividerPosition.Peek()
}

func (s *SplitPaneState) clampPosition(pos float64) float64 {
	pos = clampFloat(pos, 0, 1)
	if !s.layoutCache.valid {
		return pos
	}
	if pos < s.layoutCache.minPos {
		pos = s.layoutCache.minPos
	}
	if pos > s.layoutCache.maxPos {
		pos = s.layoutCache.maxPos
	}
	return pos
}

// SplitPane divides space between two children with a draggable divider.
type SplitPane struct {
	// Required fields
	ID     string
	State  *SplitPaneState
	First  Widget
	Second Widget

	// Configuration
	Orientation  SplitPaneOrientation
	DividerSize  int
	MinPaneSize  int
	DisableFocus bool
	OnExitFocus  func()

	// Appearance
	DividerForeground      ColorProvider
	DividerBackground      ColorProvider
	DividerFocusForeground ColorProvider
	DividerFocusBackground ColorProvider
	DividerChar            string

	// Standard widget fields
	Width     Dimension
	Height    Dimension
	Style     Style
	Click     func(MouseEvent)
	MouseDown func(MouseEvent)
	MouseUp   func(MouseEvent)
	MouseMove func(MouseEvent)
	Hover     func(bool)
}

// WidgetID returns the split pane's unique identifier.
func (s SplitPane) WidgetID() string {
	return s.ID
}

// GetContentDimensions returns the width and height dimension preferences.
func (s SplitPane) GetContentDimensions() (width, height Dimension) {
	dims := s.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = s.Width
	}
	if height.IsUnset() {
		height = s.Height
	}
	if width.IsUnset() {
		width = Flex(1)
	}
	if height.IsUnset() {
		height = Flex(1)
	}
	return width, height
}

// GetStyle returns the split pane's style.
func (s SplitPane) GetStyle() Style {
	return s.Style
}

// Build returns itself as SplitPane manages its own children.
func (s SplitPane) Build(ctx BuildContext) Widget {
	return s
}

// ChildWidgets returns the split pane's children for render tree construction.
func (s SplitPane) ChildWidgets() []Widget {
	first := s.First
	second := s.Second
	if first == nil {
		first = EmptyWidget{}
	}
	if second == nil {
		second = EmptyWidget{}
	}
	return []Widget{first, second}
}

// IsFocusable returns true if this widget can receive focus.
func (s SplitPane) IsFocusable() bool {
	return !s.DisableFocus
}

// OnKey handles keys not covered by declarative keybindings.
func (s SplitPane) OnKey(event KeyEvent) bool {
	return false
}

// Keybinds returns the declarative keybindings for resizing the divider.
func (s SplitPane) Keybinds() []Keybind {
	if s.State == nil || !s.State.DividerPosition.IsValid() {
		return nil
	}

	withExit := func(keybinds []Keybind) []Keybind {
		if s.OnExitFocus == nil {
			return keybinds
		}
		return append(keybinds, Keybind{
			Key:    "escape",
			Name:   "Exit divider",
			Action: s.OnExitFocus,
		})
	}

	if s.Orientation == SplitVertical {
		return withExit([]Keybind{
			{Key: "up", Name: "Move divider up", Action: s.moveDividerUp},
			{Key: "down", Name: "Move divider down", Action: s.moveDividerDown},
		})
	}

	return withExit([]Keybind{
		{Key: "left", Name: "Move divider left", Action: s.moveDividerLeft},
		{Key: "h", Name: "Move divider left", Action: s.moveDividerLeft},
		{Key: "right", Name: "Move divider right", Action: s.moveDividerRight},
		{Key: "l", Name: "Move divider right", Action: s.moveDividerRight},
	})
}

func (s SplitPane) moveDividerLeft() {
	s.shiftDivider(-splitPaneKeyStep)
}

func (s SplitPane) moveDividerRight() {
	s.shiftDivider(splitPaneKeyStep)
}

func (s SplitPane) moveDividerUp() {
	s.shiftDivider(-splitPaneKeyStep)
}

func (s SplitPane) moveDividerDown() {
	s.shiftDivider(splitPaneKeyStep)
}

func (s SplitPane) shiftDivider(delta float64) {
	if s.State == nil || !s.State.DividerPosition.IsValid() {
		return
	}
	s.State.DividerPosition.Update(func(pos float64) float64 {
		return s.State.clampPosition(pos + delta)
	})
}

// OnClick is called when the widget is clicked.
func (s SplitPane) OnClick(event MouseEvent) {
	if s.Click != nil {
		s.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
func (s SplitPane) OnMouseDown(event MouseEvent) {
	if s.MouseDown != nil {
		s.MouseDown(event)
	}
	if s.State == nil {
		return
	}
	cache := s.State.layoutCache
	if !cache.valid {
		return
	}
	if !s.isOnDivider(event, cache) {
		s.State.dragging = false
		return
	}

	s.State.dragging = true
	coord := s.contentCoord(event, cache)
	s.State.dragOffset = coord - cache.dividerPos
}

// OnMouseMove is called when the mouse is moved while dragging.
func (s SplitPane) OnMouseMove(event MouseEvent) {
	if s.MouseMove != nil {
		s.MouseMove(event)
	}
	if s.State == nil || !s.State.dragging {
		return
	}
	cache := s.State.layoutCache
	if !cache.valid {
		return
	}

	available := cache.contentWidth - cache.dividerSize
	if cache.orientation == SplitVertical {
		available = cache.contentHeight - cache.dividerSize
	}
	if available <= 0 {
		return
	}

	coord := s.contentCoord(event, cache)
	newOffset := coord - s.State.dragOffset
	newOffset = clampInt(newOffset, 0, available)
	newPos := float64(newOffset) / float64(available)
	newPos = s.State.clampPosition(newPos)
	s.State.DividerPosition.Set(newPos)
}

// OnMouseUp is called when the mouse is released on the widget.
func (s SplitPane) OnMouseUp(event MouseEvent) {
	if s.MouseUp != nil {
		s.MouseUp(event)
	}
	if s.State != nil {
		s.State.dragging = false
	}
}

// OnHover is called when the hover state changes.
func (s SplitPane) OnHover(hovered bool) {
	if s.Hover != nil {
		s.Hover(hovered)
	}
}

// OnLayout caches layout metrics for divider hit-testing and dragging.
func (s SplitPane) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if s.State == nil {
		return
	}

	box := metrics.Box()
	contentWidth := box.ContentWidth()
	contentHeight := box.ContentHeight()
	contentOffsetX := box.Border.Left + box.Padding.Left
	contentOffsetY := box.Border.Top + box.Padding.Top

	dividerSize := s.dividerSize()
	minPane := s.minPaneSize()
	position := 0.5
	if s.State.DividerPosition.IsValid() {
		position = s.State.DividerPosition.Peek()
	}

	axisSize := contentWidth
	if s.Orientation == SplitVertical {
		axisSize = contentHeight
	}
	metricsResult := computeSplitPaneMetrics(axisSize, dividerSize, minPane, position)

	cache := splitPaneLayoutCache{
		valid:          true,
		contentWidth:   contentWidth,
		contentHeight:  contentHeight,
		contentOffsetX: contentOffsetX,
		contentOffsetY: contentOffsetY,
		dividerPos:     metricsResult.offset,
		dividerSize:    metricsResult.dividerSize,
		minPos:         metricsResult.minPos,
		maxPos:         metricsResult.maxPos,
		orientation:    s.Orientation,
	}
	s.State.layoutCache = cache
}

// BuildLayoutNode builds a layout node for this SplitPane widget.
func (s SplitPane) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	first := s.First
	second := s.Second
	if first == nil {
		first = EmptyWidget{}
	}
	if second == nil {
		second = EmptyWidget{}
	}

	firstCtx := ctx.PushChild(0)
	secondCtx := ctx.PushChild(1)

	builtFirst := first.Build(firstCtx)
	builtSecond := second.Build(secondCtx)

	var firstNode layout.LayoutNode
	if builder, ok := builtFirst.(LayoutNodeBuilder); ok {
		firstNode = builder.BuildLayoutNode(firstCtx)
	} else {
		firstNode = buildFallbackLayoutNode(builtFirst, firstCtx)
	}

	var secondNode layout.LayoutNode
	if builder, ok := builtSecond.(LayoutNodeBuilder); ok {
		secondNode = builder.BuildLayoutNode(secondCtx)
	} else {
		secondNode = buildFallbackLayoutNode(builtSecond, secondCtx)
	}

	position := 0.5
	if s.State != nil {
		if s.State.DividerPosition.IsValid() {
			position = s.State.DividerPosition.Get()
		} else {
			Log("SplitPane[%s]: DividerPosition is invalid, defaulting to 0.5", s.ID)
		}
	} else {
		Log("SplitPane[%s]: State is nil, defaulting to 0.5", s.ID)
	}

	axis := layout.Horizontal
	if s.Orientation == SplitVertical {
		axis = layout.Vertical
	}

	padding := toLayoutEdgeInsets(s.Style.Padding)
	border := borderToEdgeInsets(s.Style.Border)
	dims := GetWidgetDimensionSet(s)
	minW, maxW, minH, maxH := dimensionSetToMinMax(dims, padding, border)
	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	node := layout.LayoutNode(&layout.SplitPaneNode{
		First:          firstNode,
		Second:         secondNode,
		Axis:           axis,
		Position:       clampFloat(position, 0, 1),
		DividerSize:    s.dividerSize(),
		MinPaneSize:    s.minPaneSize(),
		Padding:        padding,
		Border:         border,
		Margin:         toLayoutEdgeInsets(s.Style.Margin),
		MinWidth:       minW,
		MaxWidth:       maxW,
		MinHeight:      minH,
		MaxHeight:      maxH,
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
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

// Render draws the divider for the split pane.
func (s SplitPane) Render(ctx *RenderContext) {
	contentWidth := ctx.Width
	contentHeight := ctx.Height
	dividerSize := s.dividerSize()
	position := 0.5
	cache := splitPaneLayoutCache{}
	if s.State != nil {
		cache = s.State.layoutCache
		if s.State.DividerPosition.IsValid() {
			position = s.State.DividerPosition.Peek()
		}
	}

	dividerPos := 0
	if cache.valid && cache.contentWidth == contentWidth && cache.contentHeight == contentHeight && cache.orientation == s.Orientation {
		dividerPos = cache.dividerPos
		dividerSize = cache.dividerSize
	} else {
		axisSize := contentWidth
		if s.Orientation == SplitVertical {
			axisSize = contentHeight
		}
		metrics := computeSplitPaneMetrics(axisSize, dividerSize, s.minPaneSize(), position)
		dividerPos = metrics.offset
		dividerSize = metrics.dividerSize
	}

	if dividerSize <= 0 {
		return
	}

	dividerHighlighted := ctx.IsFocused(s)
	if s.State != nil && s.State.dragging {
		// Keep focus colors while the divider is actively being dragged.
		dividerHighlighted = true
	}
	fgProvider, bgProvider := s.dividerProviders(dividerHighlighted)

	dividerChar := s.dividerChar()
	if s.Orientation == SplitVertical {
		for i := 0; i < dividerSize; i++ {
			y := dividerPos + i
			if y < 0 || y >= contentHeight {
				continue
			}
			for x := 0; x < contentWidth; x++ {
				style := dividerCellStyle(fgProvider, bgProvider, contentWidth, contentHeight, x, y)
				ctx.DrawStyledText(x, y, dividerChar, style)
			}
		}
		return
	}

	for i := 0; i < dividerSize; i++ {
		x := dividerPos + i
		if x < 0 || x >= contentWidth {
			continue
		}
		for y := 0; y < contentHeight; y++ {
			style := dividerCellStyle(fgProvider, bgProvider, contentWidth, contentHeight, x, y)
			ctx.DrawStyledText(x, y, dividerChar, style)
		}
	}
}

func (s SplitPane) dividerSize() int {
	if s.DividerSize <= 0 {
		return 1
	}
	return s.DividerSize
}

func (s SplitPane) minPaneSize() int {
	if s.MinPaneSize <= 0 {
		return 1
	}
	return s.MinPaneSize
}

func (s SplitPane) dividerChar() string {
	if s.DividerChar != "" {
		return s.DividerChar
	}
	if s.Orientation == SplitHorizontal {
		return "│"
	}
	return "─"
}

func (s SplitPane) isOnDivider(event MouseEvent, cache splitPaneLayoutCache) bool {
	coord := s.contentCoord(event, cache)
	if cache.orientation == SplitVertical {
		return coord >= cache.dividerPos && coord < cache.dividerPos+cache.dividerSize
	}
	return coord >= cache.dividerPos && coord < cache.dividerPos+cache.dividerSize
}

func (s SplitPane) contentCoord(event MouseEvent, cache splitPaneLayoutCache) int {
	if cache.orientation == SplitVertical {
		return event.LocalY - cache.contentOffsetY
	}
	return event.LocalX - cache.contentOffsetX
}

func (s SplitPane) dividerProviders(dividerHighlighted bool) (ColorProvider, ColorProvider) {
	var fg ColorProvider
	var bg ColorProvider

	if dividerHighlighted {
		if colorProviderIsSet(s.DividerFocusForeground) {
			fg = s.DividerFocusForeground
		} else if colorProviderIsSet(s.DividerForeground) {
			fg = s.DividerForeground
		}
		if colorProviderIsSet(s.DividerFocusBackground) {
			bg = s.DividerFocusBackground
		} else if colorProviderIsSet(s.DividerBackground) {
			bg = s.DividerBackground
		}
	} else {
		if colorProviderIsSet(s.DividerForeground) {
			fg = s.DividerForeground
		}
		if colorProviderIsSet(s.DividerBackground) {
			bg = s.DividerBackground
		}
	}

	theme := getTheme()
	if !colorProviderIsSet(fg) {
		if dividerHighlighted {
			fg = theme.Primary
		} else {
			fg = theme.Border
		}
	}

	return fg, bg
}

func dividerCellStyle(fgProvider, bgProvider ColorProvider, width, height, x, y int) Style {
	style := Style{}
	if colorProviderIsSet(fgProvider) {
		style.ForegroundColor = fgProvider.ColorAt(width, height, x, y)
	}
	if colorProviderIsSet(bgProvider) {
		style.BackgroundColor = bgProvider.ColorAt(width, height, x, y)
	}
	return style
}

func colorProviderIsSet(provider ColorProvider) bool {
	return provider != nil && provider.IsSet()
}

type splitPaneMetrics struct {
	offset      int
	available   int
	dividerSize int
	minPos      float64
	maxPos      float64
}

func computeSplitPaneMetrics(axisSize, dividerSize, minPaneSize int, position float64) splitPaneMetrics {
	if axisSize < 0 {
		axisSize = 0
	}
	if dividerSize <= 0 {
		dividerSize = 1
	}
	if dividerSize > axisSize {
		dividerSize = axisSize
	}

	available := max(0, axisSize-dividerSize)

	pos := clampFloat(position, 0, 1)
	offset := 0
	if available > 0 {
		offset = int(float64(available) * pos)
	}

	minPane := max(0, minPaneSize)
	if available <= 0 {
		return splitPaneMetrics{offset: 0, available: 0, dividerSize: dividerSize, minPos: 0, maxPos: 0}
	}

	if available < 2*minPane {
		offset = available / 2
		pos := float64(offset) / float64(available)
		return splitPaneMetrics{offset: offset, available: available, dividerSize: dividerSize, minPos: pos, maxPos: pos}
	}

	minOffset := minPane
	maxOffset := available - minPane
	if offset < minOffset {
		offset = minOffset
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	minPos := float64(minOffset) / float64(available)
	maxPos := float64(maxOffset) / float64(available)

	return splitPaneMetrics{
		offset:      offset,
		available:   available,
		dividerSize: dividerSize,
		minPos:      minPos,
		maxPos:      maxPos,
	}
}

func clampFloat(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
