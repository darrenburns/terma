package terma

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// RenderContext provides drawing primitives for widgets.
// It tracks the current region where the widget should render.
type RenderContext struct {
	terminal *uv.Terminal
	// Absolute position in terminal
	X, Y int
	// Available size for this widget
	Width, Height int
	// Focus collector for gathering focusable widgets
	focusCollector *FocusCollector
	// Focus manager for checking focus state
	focusManager *FocusManager
	// Build context for passing to widget Build methods
	buildContext BuildContext
	// Widget registry for tracking widget positions
	widgetRegistry *WidgetRegistry
	// Scroll offset for vertical scrolling (content shifted up by this amount)
	scrollYOffset int
	// Virtual content height for scrollable regions
	virtualHeight int
	// Viewport bounds for clipping in scrolled regions (screen coordinates)
	// These stay constant within a scrolled region while X/Y change for virtual positions
	viewportY      int
	viewportHeight int
}

// NewRenderContext creates a root render context for the terminal.
func NewRenderContext(terminal *uv.Terminal, width, height int, fc *FocusCollector, fm *FocusManager, bc BuildContext, wr *WidgetRegistry) *RenderContext {
	return &RenderContext{
		terminal:       terminal,
		X:              0,
		Y:              0,
		Width:          width,
		Height:         height,
		focusCollector: fc,
		focusManager:   fm,
		buildContext:   bc,
		widgetRegistry: wr,
	}
}

// SubContext creates a child context offset from this one.
// The child context is clipped to not exceed the parent's bounds.
func (ctx *RenderContext) SubContext(xOffset, yOffset, width, height int) *RenderContext {
	// Calculate remaining space in parent
	remainingWidth := ctx.Width - xOffset
	// For scrolled contexts, use virtualHeight to allow rendering at virtual positions
	// The actual viewport clipping is handled by draw functions using viewportY/viewportHeight
	parentHeight := ctx.Height
	if ctx.virtualHeight > 0 {
		parentHeight = ctx.virtualHeight
	}
	remainingHeight := parentHeight - yOffset

	// Clamp dimensions to parent bounds
	if width > remainingWidth {
		width = remainingWidth
	}
	if height > remainingHeight {
		height = remainingHeight
	}
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	return &RenderContext{
		terminal:       ctx.terminal,
		X:              ctx.X + xOffset,
		Y:              ctx.Y + yOffset,
		Width:          width,
		Height:         height,
		focusCollector: ctx.focusCollector,
		focusManager:   ctx.focusManager,
		buildContext:   ctx.buildContext,
		widgetRegistry: ctx.widgetRegistry,
		scrollYOffset:  ctx.scrollYOffset,
		virtualHeight:  ctx.virtualHeight,
		viewportY:      ctx.viewportY,
		viewportHeight: ctx.viewportHeight,
	}
}

// ScrolledSubContext creates a child context with vertical scroll offset.
// Content rendered in this context will be shifted up by scrollY pixels,
// with content outside the viewport being clipped.
// virtualHeight is the total content height (may exceed viewport height).
func (ctx *RenderContext) ScrolledSubContext(xOffset, yOffset, width, height, scrollY, virtualHeight int) *RenderContext {
	sub := ctx.SubContext(xOffset, yOffset, width, height)
	sub.scrollYOffset = scrollY
	sub.virtualHeight = virtualHeight
	// Set viewport bounds for clipping - these stay constant within the scrolled region
	// while Y changes as we descend into nested widgets at virtual positions
	sub.viewportY = ctx.Y + yOffset
	sub.viewportHeight = height
	return sub
}

// collectFocusable automatically registers a widget if it implements Focusable.
func (ctx *RenderContext) collectFocusable(widget Widget) {
	if ctx.focusCollector != nil {
		ctx.focusCollector.Collect(widget)
	}
}

// IsFocused returns true if the given widget currently has focus.
// The widget must implement Keyed for reliable focus tracking across rebuilds.
func (ctx *RenderContext) IsFocused(widget Widget) bool {
	if ctx.focusManager == nil {
		return false
	}

	focusedKey := ctx.focusManager.FocusedKey()
	if focusedKey == "" {
		return false
	}

	// Check if widget has an explicit key
	if keyed, ok := widget.(Keyed); ok {
		return keyed.Key() == focusedKey
	}

	// For auto-keyed widgets, check against current path
	autoKey := "_auto:" + ctx.focusCollector.currentPath()
	return autoKey == focusedKey
}

// pushChild enters a child context for auto-key generation.
func (ctx *RenderContext) pushChild(index int) {
	if ctx.focusCollector != nil {
		ctx.focusCollector.PushChild(index)
	}
}

// popChild exits the current child context.
func (ctx *RenderContext) popChild() {
	if ctx.focusCollector != nil {
		ctx.focusCollector.PopChild()
	}
}

// RenderChild handles the complete rendering of a child widget.
// It manages key generation, focus collection, layout, and rendering.
// Automatically applies padding, margin, and border from the widget's style.
// Returns the size of the rendered child including padding, margin, and border.
func (ctx *RenderContext) RenderChild(index int, child Widget, xOffset, yOffset, maxWidth, maxHeight int) Size {
	ctx.pushChild(index)
	defer ctx.popChild()

	// Track if this widget handles keys (KeyHandler or KeybindProvider) for bubbling
	if ctx.focusCollector != nil && ctx.focusCollector.ShouldTrackAncestor(child) {
		ctx.focusCollector.PushAncestor(child)
		defer ctx.focusCollector.PopAncestor()
	}

	// Auto-collect focusable widgets (collect original, not built, for composite widgets)
	ctx.collectFocusable(child)

	// Build the widget with build context
	built := child.Build(ctx.buildContext)

	// Extract style for padding, margin, and border
	var style Style
	if styled, ok := built.(Styled); ok {
		style = styled.GetStyle()
	}
	margin := style.Margin
	padding := style.Padding
	border := style.Border
	borderWidth := border.Width()

	// Calculate total insets (margin + border + padding)
	hInset := margin.Horizontal() + borderWidth*2 + padding.Horizontal()
	vInset := margin.Vertical() + borderWidth*2 + padding.Vertical()

	// Reduce available space for layout by insets
	contentMaxWidth := maxWidth - hInset
	contentMaxHeight := maxHeight - vInset
	if contentMaxWidth < 0 {
		contentMaxWidth = 0
	}
	if contentMaxHeight < 0 {
		contentMaxHeight = 0
	}

	// Compute child size via layout with reduced constraints
	contentWidth := contentMaxWidth
	contentHeight := contentMaxHeight
	if layoutable, ok := built.(Layoutable); ok {
		size := layoutable.Layout(ctx.buildContext, Constraints{
			MinWidth:  0,
			MaxWidth:  contentMaxWidth,
			MinHeight: 0,
			MaxHeight: contentMaxHeight,
		})
		contentWidth = size.Width
		contentHeight = size.Height
	}

	// Calculate the bordered area (content + padding + border, excludes margin)
	// Border surrounds padding, so bordered area includes border + padding + content
	borderedWidth := contentWidth + padding.Horizontal() + borderWidth*2
	borderedHeight := contentHeight + padding.Vertical() + borderWidth*2
	borderedXOffset := xOffset + margin.Left
	borderedYOffset := yOffset + margin.Top

	// Fill the inner area (padding + content) with background color if set
	if style.BackgroundColor != DefaultColor {
		innerXOffset := borderedXOffset + borderWidth
		innerYOffset := borderedYOffset + borderWidth
		innerWidth := contentWidth + padding.Horizontal()
		innerHeight := contentHeight + padding.Vertical()
		innerCtx := ctx.SubContext(innerXOffset, innerYOffset, innerWidth, innerHeight)
		innerCtx.FillRect(0, 0, innerWidth, innerHeight, style.BackgroundColor)
	}

	// Draw border if set
	if !border.IsZero() {
		borderCtx := ctx.SubContext(borderedXOffset, borderedYOffset, borderedWidth, borderedHeight)
		borderCtx.DrawBorder(0, 0, borderedWidth, borderedHeight, border)
	}

	// Return total size including padding, margin, and border
	totalWidth := contentWidth + hInset
	totalHeight := contentHeight + vInset

	// Record widget in registry BEFORE rendering children.
	// This ensures parents are recorded before children, so when searching
	// back-to-front we find the deepest (most nested) widget first.
	if ctx.widgetRegistry != nil {
		// Get the widget's key if it has one
		var key string
		if keyed, ok := child.(Keyed); ok {
			key = keyed.Key()
		}

		// Calculate absolute position of the bordered area (excludes margin).
		// Margin is space outside the widget, so clicks on margin should not
		// register as clicks on the widget.
		absX := ctx.X + borderedXOffset
		absY := ctx.Y + borderedYOffset

		ctx.widgetRegistry.Record(child, key, Rect{
			X:      absX,
			Y:      absY,
			Width:  borderedWidth,
			Height: borderedHeight,
		})
	}

	// Render the child in its content region, offset by margin, border, and padding
	// This happens AFTER recording so children are recorded after their parent
	if renderable, ok := built.(Renderable); ok {
		contentXOffset := xOffset + margin.Left + borderWidth + padding.Left
		contentYOffset := yOffset + margin.Top + borderWidth + padding.Top
		childCtx := ctx.SubContext(contentXOffset, contentYOffset, contentWidth, contentHeight)
		renderable.Render(childCtx)
	}

	return Size{Width: totalWidth, Height: totalHeight}
}

// FillRect fills a rectangular region with a background color.
func (ctx *RenderContext) FillRect(x, y, width, height int, bgColor Color) {
	if bgColor == DefaultColor {
		return
	}

	cellStyle := uv.Style{
		Bg: bgColor.toANSI(),
	}

	// Use viewport bounds for clipping if in a scrolled region
	clipY, clipHeight := ctx.Y, ctx.Height
	if ctx.viewportHeight > 0 {
		clipY = ctx.viewportY
		clipHeight = ctx.viewportHeight
	}

	for row := 0; row < height; row++ {
		absY := ctx.Y + y + row - ctx.scrollYOffset
		if absY < clipY || absY >= clipY+clipHeight {
			continue
		}
		for col := 0; col < width; col++ {
			absX := ctx.X + x + col
			if absX < ctx.X || absX >= ctx.X+ctx.Width {
				continue
			}
			cell := &uv.Cell{Content: " ", Width: 1, Style: cellStyle}
			ctx.terminal.SetCell(absX, absY, cell)
		}
	}
}

// DrawBorder draws a border around a rectangular region.
// The border is drawn at the edges of the specified rectangle.
func (ctx *RenderContext) DrawBorder(x, y, width, height int, border Border) {
	if border.Style == BorderNone || width < 2 || height < 2 {
		return
	}

	// Border characters based on style
	var tl, tr, bl, br, h, v string
	switch border.Style {
	case BorderSquare:
		tl, tr, bl, br, h, v = "┌", "┐", "└", "┘", "─", "│"
	case BorderRounded:
		tl, tr, bl, br, h, v = "╭", "╮", "╰", "╯", "─", "│"
	default:
		return
	}

	borderStyle := uv.Style{
		Fg: border.Color.toANSI(),
	}

	// Use viewport bounds for clipping if in a scrolled region
	clipY, clipHeight := ctx.Y, ctx.Height
	if ctx.viewportHeight > 0 {
		clipY = ctx.viewportY
		clipHeight = ctx.viewportHeight
	}

	// Helper to set a cell with a specific style
	setCellStyled := func(cx, cy int, content string, style uv.Style) {
		absX := ctx.X + cx
		absY := ctx.Y + cy - ctx.scrollYOffset
		if absX < ctx.X || absX >= ctx.X+ctx.Width || absY < clipY || absY >= clipY+clipHeight {
			return
		}
		cell := &uv.Cell{Content: content, Width: 1, Style: style}
		ctx.terminal.SetCell(absX, absY, cell)
	}

	// Helper to set a cell with border style
	setCell := func(cx, cy int, content string) {
		setCellStyled(cx, cy, content, borderStyle)
	}

	// Draw corners
	setCell(x, y, tl)
	setCell(x+width-1, y, tr)
	setCell(x, y+height-1, bl)
	setCell(x+width-1, y+height-1, br)

	// Available width for horizontal edges (excluding corners)
	edgeWidth := width - 2

	// Group decorations by edge (top or bottom)
	var topDecorations, bottomDecorations []BorderDecoration
	for _, dec := range border.Decorations {
		switch dec.Position {
		case DecorationTopLeft, DecorationTopCenter, DecorationTopRight:
			topDecorations = append(topDecorations, dec)
		case DecorationBottomLeft, DecorationBottomCenter, DecorationBottomRight:
			bottomDecorations = append(bottomDecorations, dec)
		}
	}

	// Draw horizontal edge with decorations
	drawHorizontalEdge := func(edgeY int, decorations []BorderDecoration) {
		// Create a slice to track which positions are occupied by decoration text
		// true = occupied by decoration, false = draw border character
		occupied := make([]bool, edgeWidth)

		// Calculate decoration positions and mark occupied cells
		type placedDecoration struct {
			text  string
			start int
			color Color
		}
		var placed []placedDecoration

		for _, dec := range decorations {
			// Add spacing around text: " text "
			text := " " + dec.Text + " "
			textLen := ansi.StringWidth(text)

			if textLen > edgeWidth {
				// Truncate if too long (using display width)
				text = ansi.Truncate(text, edgeWidth, "")
				textLen = edgeWidth
			}

			var startPos int
			switch dec.Position {
			case DecorationTopLeft, DecorationBottomLeft:
				startPos = 0
			case DecorationTopCenter, DecorationBottomCenter:
				startPos = (edgeWidth - textLen) / 2
			case DecorationTopRight, DecorationBottomRight:
				startPos = edgeWidth - textLen
			}

			// Clamp to valid range
			if startPos < 0 {
				startPos = 0
			}
			if startPos+textLen > edgeWidth {
				startPos = edgeWidth - textLen
			}

			// Mark cells as occupied
			for i := 0; i < textLen && startPos+i < edgeWidth; i++ {
				occupied[startPos+i] = true
			}

			placed = append(placed, placedDecoration{
				text:  text,
				start: startPos,
				color: dec.Color,
			})
		}

		// Draw border characters where not occupied
		for col := 0; col < edgeWidth; col++ {
			if !occupied[col] {
				setCell(x+1+col, edgeY, h)
			}
		}

		// Draw decoration text
		for _, p := range placed {
			decorationStyle := borderStyle
			if p.color != DefaultColor {
				decorationStyle = uv.Style{Fg: p.color.toANSI()}
			}
			for i, r := range p.text {
				if p.start+i < edgeWidth {
					setCellStyled(x+1+p.start+i, edgeY, string(r), decorationStyle)
				}
			}
		}
	}

	// Draw top edge with decorations
	drawHorizontalEdge(y, topDecorations)

	// Draw bottom edge with decorations
	drawHorizontalEdge(y+height-1, bottomDecorations)

	// Draw left and right edges
	for row := 1; row < height-1; row++ {
		setCell(x, y+row, v)
		setCell(x+width-1, y+row, v)
	}
}

// DrawText draws text at the given position relative to this context.
func (ctx *RenderContext) DrawText(x, y int, text string) {
	ctx.DrawStyledText(x, y, text, Style{})
}

// DrawStyledText draws styled text at the given position relative to this context.
func (ctx *RenderContext) DrawStyledText(x, y int, text string, style Style) {
	absX := ctx.X + x
	absY := ctx.Y + y - ctx.scrollYOffset

	// Use viewport bounds for clipping if in a scrolled region
	clipY, clipHeight := ctx.Y, ctx.Height
	if ctx.viewportHeight > 0 {
		clipY = ctx.viewportY
		clipHeight = ctx.viewportHeight
	}

	// Skip if outside render region (accounting for scroll)
	if absY < clipY || absY >= clipY+clipHeight {
		return
	}

	// Build the cell style
	cellStyle := uv.Style{
		Fg: style.ForegroundColor.toANSI(),
		Bg: style.BackgroundColor.toANSI(),
	}

	// Draw each grapheme cluster as a cell, advancing by its display width
	col := 0
	remaining := text
	for len(remaining) > 0 {
		grapheme, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		cellX := absX + col
		if cellX >= ctx.X+ctx.Width {
			break
		}
		cell := &uv.Cell{Content: grapheme, Width: width, Style: cellStyle}
		ctx.terminal.SetCell(cellX, absY, cell)
		col += width
		remaining = remaining[len(grapheme):]
	}
}

// DrawSpan draws a styled span at the given position relative to this context.
// The baseStyle provides default colors when span style doesn't specify them.
// Returns the number of characters drawn (for positioning subsequent spans).
func (ctx *RenderContext) DrawSpan(x, y int, span Span, baseStyle Style) int {
	absX := ctx.X + x
	absY := ctx.Y + y - ctx.scrollYOffset

	// Use viewport bounds for clipping if in a scrolled region
	clipY, clipHeight := ctx.Y, ctx.Height
	if ctx.viewportHeight > 0 {
		clipY = ctx.viewportY
		clipHeight = ctx.viewportHeight
	}

	// Skip if outside render region (accounting for scroll)
	if absY < clipY || absY >= clipY+clipHeight {
		return ansi.StringWidth(span.Text)
	}

	// Determine colors: span style overrides base style
	fg := span.Style.Foreground
	if fg == DefaultColor {
		fg = baseStyle.ForegroundColor
	}
	bg := span.Style.Background
	if bg == DefaultColor {
		bg = baseStyle.BackgroundColor
	}

	// Build text attributes bitmask
	var attrs uint8
	if span.Style.Bold {
		attrs |= uv.AttrBold
	}
	if span.Style.Italic {
		attrs |= uv.AttrItalic
	}

	// Build underline style
	var underline uv.Underline
	if span.Style.Underline {
		underline = uv.UnderlineSingle
	}

	// Build the cell style with text attributes
	cellStyle := uv.Style{
		Fg:        fg.toANSI(),
		Bg:        bg.toANSI(),
		Attrs:     attrs,
		Underline: underline,
	}

	// Draw each grapheme cluster as a cell, advancing by its display width
	col := 0
	remaining := span.Text
	for len(remaining) > 0 {
		grapheme, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		cellX := absX + col
		if cellX >= ctx.X+ctx.Width {
			break
		}
		cell := &uv.Cell{Content: grapheme, Width: width, Style: cellStyle}
		ctx.terminal.SetCell(cellX, absY, cell)
		col += width
		remaining = remaining[len(grapheme):]
	}

	return ansi.StringWidth(span.Text)
}

// Renderer handles the widget tree rendering pipeline.
type Renderer struct {
	terminal       *uv.Terminal
	width          int
	height         int
	focusCollector *FocusCollector
	focusManager   *FocusManager
	focusedSignal  *AnySignal[Focusable]
	hoveredSignal  *AnySignal[Widget]
	widgetRegistry *WidgetRegistry
}

// NewRenderer creates a new renderer for the given terminal.
func NewRenderer(terminal *uv.Terminal, width, height int, fm *FocusManager, focusedSignal *AnySignal[Focusable], hoveredSignal *AnySignal[Widget]) *Renderer {
	return &Renderer{
		terminal:       terminal,
		width:          width,
		height:         height,
		focusCollector: NewFocusCollector(),
		focusManager:   fm,
		focusedSignal:  focusedSignal,
		hoveredSignal:  hoveredSignal,
		widgetRegistry: NewWidgetRegistry(),
	}
}

// Resize updates the renderer dimensions.
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
}

// Render renders the widget tree to the terminal and returns collected focusables.
func (r *Renderer) Render(root Widget) []FocusableEntry {
	// Reset focus collector and widget registry for this render pass
	r.focusCollector.Reset()
	r.widgetRegistry.Reset()

	// Create build context
	buildCtx := NewBuildContext(r.focusManager, r.focusedSignal, r.hoveredSignal)

	// Create render context
	ctx := NewRenderContext(r.terminal, r.width, r.height, r.focusCollector, r.focusManager, buildCtx, r.widgetRegistry)

	// Track if root widget handles keys (KeyHandler or KeybindProvider) for bubbling
	if r.focusCollector.ShouldTrackAncestor(root) {
		r.focusCollector.PushAncestor(root)
	}

	// Auto-collect focusable for root widget (collect original, not built)
	ctx.collectFocusable(root)

	// Build the root widget
	built := root.Build(buildCtx)

	// Extract style for padding, margin, and border
	var style Style
	if styled, ok := built.(Styled); ok {
		style = styled.GetStyle()
	}
	margin := style.Margin
	padding := style.Padding
	border := style.Border
	borderWidth := border.Width()

	// Calculate total insets (margin + border + padding)
	hInset := margin.Horizontal() + borderWidth*2 + padding.Horizontal()
	vInset := margin.Vertical() + borderWidth*2 + padding.Vertical()

	// Reduce available space for layout by insets
	contentMaxWidth := r.width - hInset
	contentMaxHeight := r.height - vInset
	if contentMaxWidth < 0 {
		contentMaxWidth = 0
	}
	if contentMaxHeight < 0 {
		contentMaxHeight = 0
	}

	// Create constraints with reduced size
	constraints := Constraints{
		MinWidth:  0,
		MaxWidth:  contentMaxWidth,
		MinHeight: 0,
		MaxHeight: contentMaxHeight,
	}

	// Layout the widget
	var contentWidth, contentHeight int
	if layoutable, ok := built.(Layoutable); ok {
		size := layoutable.Layout(buildCtx, constraints)
		contentWidth = size.Width
		contentHeight = size.Height
	} else {
		contentWidth = contentMaxWidth
		contentHeight = contentMaxHeight
	}

	// Calculate the bordered area (content + padding + border, excludes margin)
	borderedWidth := contentWidth + padding.Horizontal() + borderWidth*2
	borderedHeight := contentHeight + padding.Vertical() + borderWidth*2
	borderedXOffset := margin.Left
	borderedYOffset := margin.Top

	// Fill the inner area (padding + content) with background color if set
	if style.BackgroundColor != DefaultColor {
		innerXOffset := borderedXOffset + borderWidth
		innerYOffset := borderedYOffset + borderWidth
		innerWidth := contentWidth + padding.Horizontal()
		innerHeight := contentHeight + padding.Vertical()
		innerCtx := ctx.SubContext(innerXOffset, innerYOffset, innerWidth, innerHeight)
		innerCtx.FillRect(0, 0, innerWidth, innerHeight, style.BackgroundColor)
	}

	// Draw border if set
	if !border.IsZero() {
		borderCtx := ctx.SubContext(borderedXOffset, borderedYOffset, borderedWidth, borderedHeight)
		borderCtx.DrawBorder(0, 0, borderedWidth, borderedHeight, border)
	}

	// Record root widget in registry BEFORE rendering children.
	// This ensures the root is recorded first, so when searching back-to-front
	// we find the deepest (most nested) widget first.
	// Use bordered area (excludes margin) - margin is outside the widget.
	var rootKey string
	if keyed, ok := root.(Keyed); ok {
		rootKey = keyed.Key()
	}
	r.widgetRegistry.Record(root, rootKey, Rect{
		X:      borderedXOffset,
		Y:      borderedYOffset,
		Width:  borderedWidth,
		Height: borderedHeight,
	})

	// Render the widget offset by margin, border, and padding
	// This happens AFTER recording so children are recorded after root
	if renderable, ok := built.(Renderable); ok {
		contentXOffset := margin.Left + borderWidth + padding.Left
		contentYOffset := margin.Top + borderWidth + padding.Top
		childCtx := ctx.SubContext(contentXOffset, contentYOffset, contentWidth, contentHeight)
		renderable.Render(childCtx)
	}

	return r.focusCollector.Focusables()
}

// WidgetAt returns the topmost widget at the given terminal coordinates.
// Returns nil if no widget is at that position.
func (r *Renderer) WidgetAt(x, y int) *WidgetEntry {
	return r.widgetRegistry.WidgetAt(x, y)
}

// WidgetByKey returns the widget entry with the given key.
// Returns nil if no widget has that key.
func (r *Renderer) WidgetByKey(key string) *WidgetEntry {
	return r.widgetRegistry.WidgetByKey(key)
}

// ScrollableAt returns the innermost Scrollable widget at the given coordinates.
// Returns nil if no Scrollable is at that position.
func (r *Renderer) ScrollableAt(x, y int) *Scrollable {
	return r.widgetRegistry.ScrollableAt(x, y)
}
