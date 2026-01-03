package terma

import (
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// blendForeground blends a semi-transparent foreground color over a background.
// Returns the foreground unchanged if it's opaque or not set.
func blendForeground(fg, bg Color) Color {
	if !fg.IsSet() || fg.IsOpaque() {
		return fg
	}
	blendTarget := bg
	if !blendTarget.IsSet() {
		blendTarget = Black
	}
	return fg.BlendOver(blendTarget)
}

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
	// Returns inherited background color at absolute screen Y.
	// Used for transparent backgrounds to sample from underlying widgets (e.g., GradientBox).
	// Nil means no inherited background (use terminal default).
	inheritedBgAt func(absY int) Color
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
		inheritedBgAt:  ctx.inheritedBgAt,
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
		ctx.focusCollector.Collect(widget, ctx.buildContext.AutoID())
	}
}

// IsFocused returns true if the given widget currently has focus.
// The widget must implement Identifiable for reliable focus tracking across rebuilds.
func (ctx *RenderContext) IsFocused(widget Widget) bool {
	if ctx.focusManager == nil {
		return false
	}

	focusedID := ctx.focusManager.FocusedID()
	if focusedID == "" {
		return false
	}

	// Check if widget has an explicit ID
	if identifiable, ok := widget.(Identifiable); ok {
		return identifiable.WidgetID() == focusedID
	}

	// For auto-ID widgets, check against current path from BuildContext
	return ctx.buildContext.AutoID() == focusedID
}

// RenderChild handles the complete rendering of a child widget.
// It manages key generation, focus collection, layout, and rendering.
// Automatically applies padding, margin, and border from the widget's style.
// Returns the size of the rendered child including padding, margin, and border.
func (ctx *RenderContext) RenderChild(index int, child Widget, xOffset, yOffset, maxWidth, maxHeight int) Size {
	// Update build context with child path
	parentBuildContext := ctx.buildContext
	ctx.buildContext = ctx.buildContext.PushChild(index)
	defer func() { ctx.buildContext = parentBuildContext }()

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

	// Fill the bordered area (border + padding + content) with background color if set
	// This ensures border cells also have the correct background color
	if style.BackgroundColor.IsSet() {
		borderedCtx := ctx.SubContext(borderedXOffset, borderedYOffset, borderedWidth, borderedHeight)
		borderedCtx.FillRect(0, 0, borderedWidth, borderedHeight, style.BackgroundColor)
	}

	// Draw border if set (on top of the background)
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
		// Get the widget's ID if it has one
		var id string
		if identifiable, ok := child.(Identifiable); ok {
			id = identifiable.WidgetID()
		}

		// Calculate absolute position of the bordered area (excludes margin).
		// Margin is space outside the widget, so clicks on margin should not
		// register as clicks on the widget.
		absX := ctx.X + borderedXOffset
		absY := ctx.Y + borderedYOffset

		ctx.widgetRegistry.Record(child, id, Rect{
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
		// Set inherited background callback for this widget's children.
		// IMPORTANT: We set this to return the blended result AFTER the widget renders.
		// The widget itself uses ctx.inheritedBgAt (parent's callback) for its own
		// background blending. The widget's children need to see the blended result.
		if style.BackgroundColor.IsSet() {
			widgetBg := style.BackgroundColor
			parentCallback := ctx.inheritedBgAt

			if widgetBg.IsOpaque() {
				childCtx.inheritedBgAt = func(absY int) Color { return widgetBg }
			} else {
				// For semi-transparent backgrounds, children see the blended result.
				// The widget's own background was drawn by FillRect using parent's callback.
				childCtx.inheritedBgAt = func(absY int) Color {
					inherited := Black
					if parentCallback != nil {
						inherited = parentCallback(absY)
					}
					if !inherited.IsSet() {
						inherited = Black
					}
					return widgetBg.BlendOver(inherited)
				}
			}
		}
		renderable.Render(childCtx)
	}

	return Size{Width: totalWidth, Height: totalHeight}
}

// FillRect fills a rectangular region with a background color.
// If the color is semi-transparent, it blends with the inherited background.
func (ctx *RenderContext) FillRect(x, y, width, height int, bgColor Color) {
	if !bgColor.IsSet() {
		return
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

		// Determine effective background color for this row
		effectiveBg := bgColor
		if !bgColor.IsOpaque() {
			// Semi-transparent: blend over inherited background
			inherited := Color{}
			if ctx.inheritedBgAt != nil {
				inherited = ctx.inheritedBgAt(absY)
			}
			if !inherited.IsSet() {
				inherited = Black
			}
			effectiveBg = bgColor.BlendOver(inherited)
		}

		cellStyle := uv.Style{
			Bg: effectiveBg.toANSI(),
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

// DrawBackdrop applies a semi-transparent overlay over existing content.
// Unlike FillRect, this preserves the underlying characters and blends
// the backdrop color over both foreground and background colors.
// This creates a true transparency effect where text behind is still visible.
func (ctx *RenderContext) DrawBackdrop(x, y, width, height int, backdropColor Color) {
	if !backdropColor.IsSet() {
		return
	}

	for row := 0; row < height; row++ {
		absY := ctx.Y + y + row
		if absY < 0 || absY >= ctx.Height {
			continue
		}

		for col := 0; col < width; col++ {
			absX := ctx.X + x + col
			if absX < 0 || absX >= ctx.Width {
				continue
			}

			// Read existing cell from terminal buffer
			existingCell := ctx.terminal.CellAt(absX, absY)
			if existingCell == nil {
				// No existing cell, just fill with backdrop color
				cell := &uv.Cell{
					Content: " ",
					Width:   1,
					Style:   uv.Style{Bg: backdropColor.toANSI()},
				}
				ctx.terminal.SetCell(absX, absY, cell)
				continue
			}

			// Convert existing colors and blend with backdrop
			existingFg := FromANSI(existingCell.Style.Fg)
			existingBg := FromANSI(existingCell.Style.Bg)

			// Blend backdrop over both foreground and background
			blendedFg := backdropColor.BlendOver(existingFg)
			blendedBg := backdropColor.BlendOver(existingBg)

			// Re-write cell with same content but blended colors
			newCell := &uv.Cell{
				Content: existingCell.Content,
				Width:   existingCell.Width,
				Style: uv.Style{
					Fg:        blendedFg.toANSI(),
					Bg:        blendedBg.toANSI(),
					Attrs:     existingCell.Style.Attrs,
					Underline: existingCell.Style.Underline,
				},
			}
			ctx.terminal.SetCell(absX, absY, newCell)
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

	// Use viewport bounds for clipping if in a scrolled region
	clipY, clipHeight := ctx.Y, ctx.Height
	if ctx.viewportHeight > 0 {
		clipY = ctx.viewportY
		clipHeight = ctx.viewportHeight
	}

	// Helper to set a cell with a specific foreground color
	setCellStyled := func(cx, cy int, content string, fgColor Color) {
		absX := ctx.X + cx
		absY := ctx.Y + cy - ctx.scrollYOffset
		if absX < ctx.X || absX >= ctx.X+ctx.Width || absY < clipY || absY >= clipY+clipHeight {
			return
		}
		// Use inherited background if available
		var bg Color
		if ctx.inheritedBgAt != nil {
			bg = ctx.inheritedBgAt(absY)
		}
		// Blend foreground if semi-transparent
		effectiveFg := blendForeground(fgColor, bg)
		style := uv.Style{
			Fg: effectiveFg.toANSI(),
			Bg: bg.toANSI(),
		}
		cell := &uv.Cell{Content: content, Width: 1, Style: style}
		ctx.terminal.SetCell(absX, absY, cell)
	}

	// Helper to set a cell with border color
	setCell := func(cx, cy int, content string) {
		setCellStyled(cx, cy, content, border.Color)
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
			fgColor := border.Color
			if p.color.IsSet() {
				fgColor = p.color
			}
			for i, r := range p.text {
				if p.start+i < edgeWidth {
					setCellStyled(x+1+p.start+i, edgeY, string(r), fgColor)
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

	// Determine background color: use inherited if no explicit background,
	// or blend if semi-transparent
	bg := style.BackgroundColor
	if bg.IsSet() && !bg.IsOpaque() {
		// Semi-transparent: blend over inherited background
		inherited := Color{}
		if ctx.inheritedBgAt != nil {
			inherited = ctx.inheritedBgAt(absY)
		}
		if !inherited.IsSet() {
			inherited = Black
		}
		bg = bg.BlendOver(inherited)
	} else if !bg.IsSet() && ctx.inheritedBgAt != nil {
		bg = ctx.inheritedBgAt(absY)
	}

	// Blend foreground if semi-transparent
	fg := blendForeground(style.ForegroundColor, bg)

	// Build the cell style
	cellStyle := uv.Style{
		Fg: fg.toANSI(),
		Bg: bg.toANSI(),
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

	// Determine colors: span style overrides base style, then inherited
	fg := span.Style.Foreground
	if !fg.IsSet() {
		fg = baseStyle.ForegroundColor
	}
	bg := span.Style.Background
	if !bg.IsSet() {
		bg = baseStyle.BackgroundColor
	}
	// Handle semi-transparent backgrounds or fall back to inherited
	if bg.IsSet() && !bg.IsOpaque() {
		// Semi-transparent: blend over inherited background
		inherited := Color{}
		if ctx.inheritedBgAt != nil {
			inherited = ctx.inheritedBgAt(absY)
		}
		if !inherited.IsSet() {
			inherited = Black
		}
		bg = bg.BlendOver(inherited)
	} else if !bg.IsSet() && ctx.inheritedBgAt != nil {
		bg = ctx.inheritedBgAt(absY)
	}

	// Blend foreground if semi-transparent
	fg = blendForeground(fg, bg)

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
	floatCollector *FloatCollector
	// modalFocusTarget is the ID of the first focusable in a modal float.
	// Used to auto-focus into modals when they open.
	modalFocusTarget string
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
		floatCollector: NewFloatCollector(),
	}
}

// Resize updates the renderer dimensions.
func (r *Renderer) Resize(width, height int) {
	r.width = width
	r.height = height
}

// ScreenText returns the current screen content as plain text.
// Each row is separated by a newline. Wide characters are handled correctly.
func (r *Renderer) ScreenText() string {
	var builder strings.Builder

	for y := 0; y < r.height; y++ {
		x := 0
		for x < r.width {
			cell := r.terminal.CellAt(x, y)

			if cell == nil || cell.Content == "" {
				builder.WriteByte(' ')
				x++
				continue
			}

			builder.WriteString(cell.Content)

			// Skip continuation cells for wide characters
			if cell.Width > 1 {
				x += cell.Width
			} else {
				x++
			}
		}

		if y < r.height-1 {
			builder.WriteByte('\n')
		}
	}

	return builder.String()
}

// Render renders the widget tree to the terminal and returns collected focusables.
func (r *Renderer) Render(root Widget) []FocusableEntry {
	// Reset collectors and widget registry for this render pass
	r.focusCollector.Reset()
	r.widgetRegistry.Reset()
	r.floatCollector.Reset()
	r.modalFocusTarget = ""

	// Create build context
	buildCtx := NewBuildContext(r.focusManager, r.focusedSignal, r.hoveredSignal, r.floatCollector)

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

	// Fill the bordered area (border + padding + content) with background color if set
	// This ensures border cells also have the correct background color
	if style.BackgroundColor.IsSet() {
		borderedCtx := ctx.SubContext(borderedXOffset, borderedYOffset, borderedWidth, borderedHeight)
		borderedCtx.FillRect(0, 0, borderedWidth, borderedHeight, style.BackgroundColor)
	}

	// Draw border if set (on top of the background)
	if !border.IsZero() {
		borderCtx := ctx.SubContext(borderedXOffset, borderedYOffset, borderedWidth, borderedHeight)
		borderCtx.DrawBorder(0, 0, borderedWidth, borderedHeight, border)
	}

	// Record root widget in registry BEFORE rendering children.
	// This ensures the root is recorded first, so when searching back-to-front
	// we find the deepest (most nested) widget first.
	// Use bordered area (excludes margin) - margin is outside the widget.
	var rootID string
	if identifiable, ok := root.(Identifiable); ok {
		rootID = identifiable.WidgetID()
	}
	r.widgetRegistry.Record(root, rootID, Rect{
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
		// Set inherited background callback for this widget's children.
		if style.BackgroundColor.IsSet() {
			widgetBg := style.BackgroundColor
			parentCallback := ctx.inheritedBgAt

			if widgetBg.IsOpaque() {
				childCtx.inheritedBgAt = func(absY int) Color { return widgetBg }
			} else {
				// For semi-transparent backgrounds, children see the blended result.
				childCtx.inheritedBgAt = func(absY int) Color {
					inherited := Black
					if parentCallback != nil {
						inherited = parentCallback(absY)
					}
					if !inherited.IsSet() {
						inherited = Black
					}
					return widgetBg.BlendOver(inherited)
				}
			}
		}
		renderable.Render(childCtx)
	}

	// Phase 2: Render floating widgets on top
	// Set up inherited background on root context so float backdrops can blend properly
	if style.BackgroundColor.IsSet() {
		rootBg := style.BackgroundColor
		ctx.inheritedBgAt = func(absY int) Color { return rootBg }
	}
	r.renderFloats(ctx, buildCtx)

	return r.focusCollector.Focusables()
}

// renderFloats renders all collected floating widgets on top of the main widget tree.
func (r *Renderer) renderFloats(ctx *RenderContext, buildCtx BuildContext) {
	for i := range r.floatCollector.entries {
		entry := &r.floatCollector.entries[i]

		// For modal floats, track focusables so we can auto-focus into the modal
		var focusableCountBefore int
		if entry.Config.Modal {
			focusableCountBefore = len(r.focusCollector.Focusables())
		}

		r.renderFloat(ctx, buildCtx, entry)

		// If this was a modal and new focusables were added, check if we need to auto-focus
		if entry.Config.Modal && r.modalFocusTarget == "" {
			focusables := r.focusCollector.Focusables()
			if len(focusables) > focusableCountBefore {
				// Get the modal's focusables (those added during this float's render)
				modalFocusables := focusables[focusableCountBefore:]

				// Check if focus is already inside the modal
				currentFocusID := r.focusManager.FocusedID()
				focusInModal := false
				for _, f := range modalFocusables {
					if f.ID == currentFocusID {
						focusInModal = true
						break
					}
				}

				// Only set focus target if focus is NOT already in the modal
				if !focusInModal {
					r.modalFocusTarget = modalFocusables[0].ID
				}
			}
		}
	}
}

// renderFloat renders a single floating widget.
func (r *Renderer) renderFloat(ctx *RenderContext, buildCtx BuildContext, entry *FloatEntry) {
	// First, we need to calculate the float's size to determine position.
	// Build the child and compute its layout size.
	built := entry.Child.Build(buildCtx)

	// Extract style to account for insets in size calculation
	var style Style
	if styled, ok := built.(Styled); ok {
		style = styled.GetStyle()
	}
	margin := style.Margin
	padding := style.Padding
	border := style.Border
	borderWidth := border.Width()
	hInset := margin.Horizontal() + borderWidth*2 + padding.Horizontal()
	vInset := margin.Vertical() + borderWidth*2 + padding.Vertical()

	// Get the content size via layout (constraints reduced by insets)
	var contentWidth, contentHeight int
	if layoutable, ok := built.(Layoutable); ok {
		constraints := Constraints{
			MinWidth:  0,
			MaxWidth:  r.width - hInset,
			MinHeight: 0,
			MaxHeight: r.height - vInset,
		}
		size := layoutable.Layout(buildCtx, constraints)
		contentWidth = size.Width
		contentHeight = size.Height
	} else {
		contentWidth = r.width/2 - hInset
		contentHeight = r.height/2 - vInset
	}

	// Total float size includes insets
	floatWidth := contentWidth + hInset
	floatHeight := contentHeight + vInset

	// Calculate position
	var x, y int
	if entry.Config.AnchorID != "" {
		anchor := r.widgetRegistry.WidgetByID(entry.Config.AnchorID)
		x, y = calculateAnchorPosition(anchor, entry.Config.Anchor, floatWidth, floatHeight, entry.Config.Offset)
	} else {
		x, y = calculateAbsolutePosition(entry.Config.Position, r.width, r.height, floatWidth, floatHeight, entry.Config.Offset)
	}

	// Clamp to screen bounds
	x, y = clampToScreen(x, y, floatWidth, floatHeight, r.width, r.height)

	// Store computed position and size for event handling
	entry.X = x
	entry.Y = y
	entry.Width = floatWidth
	entry.Height = floatHeight

	// For modal floats, draw a transparent backdrop that preserves underlying content
	if entry.Config.Modal {
		backdropColor := entry.Config.BackdropColor
		if !backdropColor.IsSet() {
			backdropColor = RGBA(0, 0, 0, 128)
		}
		ctx.DrawBackdrop(0, 0, r.width, r.height, backdropColor)
	}

	// Use RenderChild to properly handle all rendering (background, border, content)
	// Create a sub-context positioned at the float location
	floatCtx := ctx.SubContext(x, y, floatWidth, floatHeight)
	floatCtx.RenderChild(1000, entry.Child, 0, 0, floatWidth, floatHeight)
}

// WidgetAt returns the topmost widget at the given terminal coordinates.
// Returns nil if no widget is at that position.
func (r *Renderer) WidgetAt(x, y int) *WidgetEntry {
	return r.widgetRegistry.WidgetAt(x, y)
}

// WidgetByID returns the widget entry with the given ID.
// Returns nil if no widget has that ID.
func (r *Renderer) WidgetByID(id string) *WidgetEntry {
	return r.widgetRegistry.WidgetByID(id)
}

// ScrollableAt returns the innermost Scrollable widget at the given coordinates.
// Returns nil if no Scrollable is at that position.
func (r *Renderer) ScrollableAt(x, y int) *Scrollable {
	return r.widgetRegistry.ScrollableAt(x, y)
}

// HasFloats returns true if there are any floating widgets.
func (r *Renderer) HasFloats() bool {
	return r.floatCollector.Len() > 0
}

// FloatAt returns the topmost float entry containing the point (x, y).
// Returns nil if no float contains the point.
func (r *Renderer) FloatAt(x, y int) *FloatEntry {
	// Search back-to-front (topmost floats are last)
	for i := len(r.floatCollector.entries) - 1; i >= 0; i-- {
		entry := &r.floatCollector.entries[i]
		if x >= entry.X && x < entry.X+entry.Width &&
			y >= entry.Y && y < entry.Y+entry.Height {
			return entry
		}
	}
	return nil
}

// TopFloat returns the topmost (last registered) float entry, or nil if none.
func (r *Renderer) TopFloat() *FloatEntry {
	if r.floatCollector.Len() == 0 {
		return nil
	}
	return &r.floatCollector.entries[len(r.floatCollector.entries)-1]
}

// HasModalFloat returns true if any float is modal.
func (r *Renderer) HasModalFloat() bool {
	return r.floatCollector.HasModal()
}

// ModalFocusTarget returns the ID of the first focusable widget in a modal float.
// Returns empty string if there's no modal or no focusables in the modal.
// Used to auto-focus into modals when they open.
func (r *Renderer) ModalFocusTarget() string {
	return r.modalFocusTarget
}
