package terma

import (
	"strings"

	"terma/layout"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

// CellBuffer is the interface for cell-based rendering.
// Both *uv.Terminal and *uv.Buffer satisfy this interface.
type CellBuffer interface {
	SetCell(x, y int, c *uv.Cell)
	CellAt(x, y int) *uv.Cell
}

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

// toUVUnderline converts a terma UnderlineStyle to an ultraviolet Underline.
func toUVUnderline(u UnderlineStyle) uv.Underline {
	switch u {
	case UnderlineSingle:
		return uv.UnderlineSingle
	case UnderlineDouble:
		return uv.UnderlineDouble
	case UnderlineCurly:
		return uv.UnderlineCurly
	case UnderlineDotted:
		return uv.UnderlineDotted
	case UnderlineDashed:
		return uv.UnderlineDashed
	default:
		return uv.UnderlineNone
	}
}

// RenderContext provides drawing primitives for widgets.
// It tracks the current region where the widget should render.
type RenderContext struct {
	terminal CellBuffer
	// Absolute position in terminal (may be outside clip for virtual/scrolled positioning)
	X, Y int
	// Available size for this widget's content
	Width, Height int
	// Clip rect in absolute screen coordinates - all drawing is clipped to this rect
	clip Rect
	// Focus collector for gathering focusable widgets
	focusCollector *FocusCollector
	// Focus manager for checking focus state
	focusManager *FocusManager
	// Build context for passing to widget Build methods
	buildContext BuildContext
	// Widget registry for tracking widget positions
	widgetRegistry *WidgetRegistry
	// Returns inherited background color at absolute screen position.
	// Used for transparent backgrounds to sample from underlying widgets with gradient/solid backgrounds.
	// Takes both X and Y to support arbitrary-angle gradients.
	// Nil means no inherited background (use terminal default).
	inheritedBgAt func(absX, absY int) Color
}

// NewRenderContext creates a root render context for the terminal.
func NewRenderContext(terminal CellBuffer, width, height int, fc *FocusCollector, fm *FocusManager, bc BuildContext, wr *WidgetRegistry) *RenderContext {
	return &RenderContext{
		terminal:       terminal,
		X:              0,
		Y:              0,
		Width:          width,
		Height:         height,
		clip:           Rect{X: 0, Y: 0, Width: width, Height: height},
		focusCollector: fc,
		focusManager:   fm,
		buildContext:   bc,
		widgetRegistry: wr,
	}
}

// IsVisible returns whether a point is within the clip rect.
func (ctx *RenderContext) IsVisible(absX, absY int) bool {
	return ctx.clip.Contains(absX, absY)
}

// ClipBounds returns the current clip rect.
func (ctx *RenderContext) ClipBounds() Rect {
	return ctx.clip
}

// SubContext creates a child context offset from this one.
// The child's clip rect is the intersection of the parent's clip rect and the child's bounds.
func (ctx *RenderContext) SubContext(xOffset, yOffset, width, height int) *RenderContext {
	// Ensure non-negative dimensions
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	// Child's absolute position
	childX := ctx.X + xOffset
	childY := ctx.Y + yOffset

	// Child's bounds
	childBounds := Rect{X: childX, Y: childY, Width: width, Height: height}

	// New clip = intersection of parent clip and child bounds
	newClip := ctx.clip.Intersect(childBounds)

	return &RenderContext{
		terminal:       ctx.terminal,
		X:              childX,
		Y:              childY,
		Width:          width,
		Height:         height,
		clip:           newClip,
		focusCollector: ctx.focusCollector,
		focusManager:   ctx.focusManager,
		buildContext:   ctx.buildContext,
		widgetRegistry: ctx.widgetRegistry,
		inheritedBgAt:  ctx.inheritedBgAt,
	}
}

// OverflowSubContext creates a child context offset from this one that allows overflow.
// Unlike SubContext, this keeps the parent's clip rect instead of intersecting with child bounds.
// This allows children to render outside their container's bounds (e.g., Stack children with
// negative positioning like badges that overflow their parent).
func (ctx *RenderContext) OverflowSubContext(xOffset, yOffset, width, height int) *RenderContext {
	// Ensure non-negative dimensions
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	// Child's absolute position
	childX := ctx.X + xOffset
	childY := ctx.Y + yOffset

	// Keep parent's clip rect to allow overflow
	return &RenderContext{
		terminal:       ctx.terminal,
		X:              childX,
		Y:              childY,
		Width:          width,
		Height:         height,
		clip:           ctx.clip, // Don't intersect - allow overflow
		focusCollector: ctx.focusCollector,
		focusManager:   ctx.focusManager,
		buildContext:   ctx.buildContext,
		widgetRegistry: ctx.widgetRegistry,
		inheritedBgAt:  ctx.inheritedBgAt,
	}
}

// ScrolledSubContext creates a child context with scroll offset applied.
// The clip rect remains the viewport bounds, but content positions are offset.
// scrollY is how much the content has been scrolled (content moves up).
func (ctx *RenderContext) ScrolledSubContext(xOffset, yOffset, width, height, scrollY int) *RenderContext {
	// Ensure non-negative dimensions
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	// Viewport's absolute screen position
	viewportX := ctx.X + xOffset
	viewportY := ctx.Y + yOffset

	// Viewport bounds (what's visible on screen)
	viewportBounds := Rect{X: viewportX, Y: viewportY, Width: width, Height: height}

	// Clip rect = intersection of parent clip and viewport bounds
	newClip := ctx.clip.Intersect(viewportBounds)

	// Content position is shifted up by scroll offset
	// When scrollY=0, content starts at viewportY
	// When scrollY=10, content is shifted up by 10 (virtual position 10 appears at viewport top)
	contentY := viewportY - scrollY

	return &RenderContext{
		terminal:       ctx.terminal,
		X:              viewportX,
		Y:              contentY,
		Width:          width,
		Height:         height,
		clip:           newClip,
		focusCollector: ctx.focusCollector,
		focusManager:   ctx.focusManager,
		buildContext:   ctx.buildContext,
		widgetRegistry: ctx.widgetRegistry,
		inheritedBgAt:  ctx.inheritedBgAt,
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

// FillRect fills a rectangular region with a background color.
// If the color is semi-transparent, it blends with the inherited background.
func (ctx *RenderContext) FillRect(x, y, width, height int, bgColor Color) {
	if !bgColor.IsSet() {
		return
	}

	for row := 0; row < height; row++ {
		absY := ctx.Y + y + row
		// Skip rows outside vertical clip bounds
		if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
			continue
		}

		for col := 0; col < width; col++ {
			absX := ctx.X + x + col
			// Skip columns outside horizontal clip bounds
			if absX < ctx.clip.X || absX >= ctx.clip.X+ctx.clip.Width {
				continue
			}

			// Determine effective background color for this cell
			effectiveBg := bgColor
			if !bgColor.IsOpaque() {
				// Semi-transparent: blend over inherited background
				inherited := Color{}
				if ctx.inheritedBgAt != nil {
					inherited = ctx.inheritedBgAt(absX, absY)
				}
				if !inherited.IsSet() {
					inherited = Black
				}
				effectiveBg = bgColor.BlendOver(inherited)
			}

			cellStyle := uv.Style{
				Bg: effectiveBg.toANSI(),
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
		// Skip rows outside vertical clip bounds
		if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
			continue
		}

		for col := 0; col < width; col++ {
			absX := ctx.X + x + col
			// Skip columns outside horizontal clip bounds
			if absX < ctx.clip.X || absX >= ctx.clip.X+ctx.clip.Width {
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
	chars := GetBorderCharSet(border.Style)
	if chars.TopLeft == "" {
		return // BorderNone or unknown style
	}
	tl, tr, bl, br := chars.TopLeft, chars.TopRight, chars.BottomLeft, chars.BottomRight
	h, v := chars.Top, chars.Left

	// Check if border color is a ColorProvider (for gradient borders)
	borderColorProvider := border.Color

	// Helper to set a cell with a specific foreground color
	// Reads existing cell to preserve background from underlying content (for transparency)
	setCellStyled := func(cx, cy int, content string, fgColor Color) {
		absX := ctx.X + cx
		absY := ctx.Y + cy
		// Skip if outside clip bounds
		if !ctx.clip.Contains(absX, absY) {
			return
		}
		// Read existing cell to get actual background (for transparency over siblings)
		var bg Color
		if existing := ctx.terminal.CellAt(absX, absY); existing != nil {
			bg = FromANSI(existing.Style.Bg)
		}
		if !bg.IsSet() && ctx.inheritedBgAt != nil {
			bg = ctx.inheritedBgAt(absX, absY)
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

	// Helper to set a cell with border color (samples from ColorProvider at cell position)
	setCell := func(cx, cy int, content string) {
		var fgColor Color
		if borderColorProvider != nil && borderColorProvider.IsSet() {
			// Sample border color at this cell's position within the border box
			fgColor = borderColorProvider.ColorAt(width, height, cx-x, cy-y)
		}
		setCellStyled(cx, cy, content, fgColor)
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
			color ColorProvider
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
			runes := []rune(p.text)
			textLen := len(runes)
			for i, r := range runes {
				if p.start+i < edgeWidth {
					cellX := x + 1 + p.start + i
					// Determine foreground color for this decoration character
					var fgColor Color
					if p.color != nil && p.color.IsSet() {
						// Sample from decoration's color provider
						// For horizontal gradients use angle=90; vertical (angle=0) on
						// single-line text falls back to midpoint color
						fgColor = p.color.ColorAt(textLen, 1, i, 0)
					} else if borderColorProvider != nil && borderColorProvider.IsSet() {
						// Fall back to border color gradient at this position
						fgColor = borderColorProvider.ColorAt(width, height, cellX-x, edgeY-y)
					}
					setCellStyled(cellX, edgeY, string(r), fgColor)
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
	absY := ctx.Y + y

	// Skip if outside vertical clip bounds
	if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
		return
	}

	// Check if we have explicit foreground/background colors
	hasExplicitFg := style.ForegroundColor != nil && style.ForegroundColor.IsSet()
	hasExplicitBg := style.BackgroundColor != nil && style.BackgroundColor.IsSet()

	// Calculate text width for foreground gradient sampling
	// Use trimmed text (no trailing spaces) so gradients span actual content, not padding
	textWidth := ansi.StringWidth(strings.TrimRight(text, " "))

	// Build text attributes bitmask
	var attrs uint8
	if style.Bold {
		attrs |= uv.AttrBold
	}
	if style.Faint {
		attrs |= uv.AttrFaint
	}
	if style.Italic {
		attrs |= uv.AttrItalic
	}
	if style.Blink {
		attrs |= uv.AttrBlink
	}
	if style.Reverse {
		attrs |= uv.AttrReverse
	}
	if style.Conceal {
		attrs |= uv.AttrConceal
	}
	if style.Strikethrough {
		attrs |= uv.AttrStrikethrough
	}

	// Draw each grapheme cluster as a cell, advancing by its display width
	col := 0
	remaining := text
	for len(remaining) > 0 {
		grapheme, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		cellX := absX + col
		// Stop if we've passed the right edge of clip rect
		if cellX >= ctx.clip.X+ctx.clip.Width {
			break
		}
		// Only draw if within horizontal clip bounds
		if cellX >= ctx.clip.X {
			// Get actual background from terminal buffer (for transparency over siblings)
			var screenBg Color
			if existing := ctx.terminal.CellAt(cellX, absY); existing != nil {
				screenBg = FromANSI(existing.Style.Bg)
			}
			if !screenBg.IsSet() && ctx.inheritedBgAt != nil {
				screenBg = ctx.inheritedBgAt(cellX, absY)
			}

			// Sample background at this cell's position
			var bg Color
			if hasExplicitBg {
				bg = style.BackgroundColor.ColorAt(1, 1, 0, 0)
				if !bg.IsOpaque() {
					if !screenBg.IsSet() {
						screenBg = Black
					}
					bg = bg.BlendOver(screenBg)
				}
			} else {
				bg = screenBg
			}

			// Sample foreground at this cell's position (supports gradients)
			var fg Color
			if hasExplicitFg {
				// Sample from ColorProvider using text width and context height
				// This allows horizontal gradients per-line and vertical gradients across lines
				fg = style.ForegroundColor.ColorAt(textWidth, ctx.Height, col, y)
				fg = blendForeground(fg, bg)
			}

			cellStyle := uv.Style{
				Fg:        fg.toANSI(),
				Bg:        bg.toANSI(),
				Attrs:     attrs,
				Underline: toUVUnderline(style.Underline),
			}
			if style.UnderlineColor.IsSet() {
				cellStyle.UnderlineColor = style.UnderlineColor.toANSI()
			}

			cell := &uv.Cell{Content: grapheme, Width: width, Style: cellStyle}
			ctx.terminal.SetCell(cellX, absY, cell)
		}
		col += width
		remaining = remaining[len(grapheme):]
	}
}

// DrawSpan draws a styled span at the given position relative to this context.
// The baseStyle provides default colors when span style doesn't specify them.
// Returns the number of characters drawn (for positioning subsequent spans).
func (ctx *RenderContext) DrawSpan(x, y int, span Span, baseStyle Style) int {
	absX := ctx.X + x
	absY := ctx.Y + y

	// Calculate span width for foreground gradient sampling
	spanWidth := ansi.StringWidth(span.Text)

	// Skip if outside vertical clip bounds
	if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
		return spanWidth
	}

	// Track whether span has explicit foreground (Color) or falls back to base (ColorProvider)
	spanHasFg := span.Style.Foreground.IsSet()
	baseFgProvider := baseStyle.ForegroundColor
	hasBaseFg := baseFgProvider != nil && baseFgProvider.IsSet()

	// Check if span has explicit background
	explicitBg := span.Style.Background
	hasExplicitBg := explicitBg.IsSet()
	hasBaseBg := baseStyle.BackgroundColor != nil && baseStyle.BackgroundColor.IsSet()

	// Build text attributes bitmask
	var attrs uint8
	if span.Style.Bold {
		attrs |= uv.AttrBold
	}
	if span.Style.Faint {
		attrs |= uv.AttrFaint
	}
	if span.Style.Italic {
		attrs |= uv.AttrItalic
	}
	if span.Style.Blink {
		attrs |= uv.AttrBlink
	}
	if span.Style.Reverse {
		attrs |= uv.AttrReverse
	}
	if span.Style.Conceal {
		attrs |= uv.AttrConceal
	}
	if span.Style.Strikethrough {
		attrs |= uv.AttrStrikethrough
	}

	// Draw each grapheme cluster as a cell, advancing by its display width
	col := 0
	remaining := span.Text
	for len(remaining) > 0 {
		grapheme, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		cellX := absX + col
		// Stop if we've passed the right edge of clip rect
		if cellX >= ctx.clip.X+ctx.clip.Width {
			break
		}
		// Only draw if within horizontal clip bounds
		if cellX >= ctx.clip.X {
			// Get actual background from terminal buffer (for transparency over siblings)
			var screenBg Color
			if existing := ctx.terminal.CellAt(cellX, absY); existing != nil {
				screenBg = FromANSI(existing.Style.Bg)
			}
			if !screenBg.IsSet() && ctx.inheritedBgAt != nil {
				screenBg = ctx.inheritedBgAt(cellX, absY)
			}

			// Sample background at this cell's position
			var bg Color
			if hasExplicitBg {
				bg = explicitBg
				if !bg.IsOpaque() {
					if !screenBg.IsSet() {
						screenBg = Black
					}
					bg = bg.BlendOver(screenBg)
				}
			} else if hasBaseBg {
				bg = baseStyle.BackgroundColor.ColorAt(1, 1, 0, 0)
				if !bg.IsOpaque() {
					if !screenBg.IsSet() {
						screenBg = Black
					}
					bg = bg.BlendOver(screenBg)
				}
			} else {
				bg = screenBg
			}

			// Determine foreground color: span style (Color) overrides base style (ColorProvider)
			var cellFg Color
			if spanHasFg {
				// Span has explicit foreground color
				cellFg = blendForeground(span.Style.Foreground, bg)
			} else if hasBaseFg {
				// Fall back to base style foreground (sample from ColorProvider)
				// Use span width for horizontal and context height for vertical gradients
				cellFg = baseFgProvider.ColorAt(spanWidth, ctx.Height, col, y)
				cellFg = blendForeground(cellFg, bg)
			}

			cellStyle := uv.Style{
				Fg:        cellFg.toANSI(),
				Bg:        bg.toANSI(),
				Attrs:     attrs,
				Underline: toUVUnderline(span.Style.Underline),
			}
			if span.Style.UnderlineColor.IsSet() {
				cellStyle.UnderlineColor = span.Style.UnderlineColor.toANSI()
			}

			cell := &uv.Cell{Content: grapheme, Width: width, Style: cellStyle}
			ctx.terminal.SetCell(cellX, absY, cell)
		}
		col += width
		remaining = remaining[len(grapheme):]
	}

	return spanWidth
}

// Renderer handles the widget tree rendering pipeline.
type Renderer struct {
	terminal       CellBuffer
	width          int
	height         int
	focusCollector *FocusCollector
	focusManager   *FocusManager
	focusedSignal  AnySignal[Focusable]
	hoveredSignal  AnySignal[Widget]
	widgetRegistry *WidgetRegistry
	floatCollector *FloatCollector
	// modalFocusTarget is the ID of the first focusable in a modal float.
	// Used to auto-focus into modals when they open.
	modalFocusTarget string
}

// NewRenderer creates a new renderer for the given terminal.
func NewRenderer(terminal CellBuffer, width, height int, fm *FocusManager, focusedSignal AnySignal[Focusable], hoveredSignal AnySignal[Widget]) *Renderer {
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
// This uses the tree-based rendering path which builds the complete layout tree first,
// then renders using BoxModel utilities for clean separation of layout and painting.
func (r *Renderer) Render(root Widget) []FocusableEntry {
	// Reset collectors and widget registry for this render pass
	r.focusCollector.Reset()
	r.widgetRegistry.Reset()
	r.floatCollector.Reset()
	r.modalFocusTarget = ""

	// Create build context
	buildCtx := NewBuildContext(r.focusManager, r.focusedSignal, r.hoveredSignal, r.floatCollector)

	// Phase 1+2: Build complete render tree (layout + focus collection)
	constraints := layout.Loose(r.width, r.height)
	renderTree := BuildRenderTree(root, buildCtx, constraints, r.focusCollector)

	// Phase 3: Render from the tree (pure painting - no layout or focus logic)
	ctx := NewRenderContext(r.terminal, r.width, r.height, nil, r.focusManager, buildCtx, r.widgetRegistry)
	r.renderTree(ctx, renderTree, 0, 0)

	// Handle floats
	r.renderFloats(ctx, buildCtx)

	return r.focusCollector.Focusables()
}

// renderTree paints a render tree to the terminal.
// All positions come from BoxModel utilities - no manual offset calculations.
// This is the new rendering path that uses computed layout geometry.
func (r *Renderer) renderTree(ctx *RenderContext, tree RenderTree, screenX, screenY int) {
	box := tree.Layout.Box

	// Get positions using BoxModel utilities
	borderX, borderY := box.BorderOrigin()    // Offset from margin-box to border-box
	contentX, contentY := box.ContentOrigin() // Offset from margin-box to content

	// Positions relative to context origin (used for SubContext calls)
	absBorderX := screenX + borderX
	absBorderY := screenY + borderY
	absContentX := screenX + contentX
	absContentY := screenY + contentY

	// True absolute screen positions (for gradient sampling and cell rendering)
	trueAbsBorderX := ctx.X + absBorderX
	trueAbsBorderY := ctx.Y + absBorderY

	// Extract style for painting
	var style Style
	if styled, ok := tree.Widget.(Styled); ok {
		style = styled.GetStyle()
	}

	// 1. Fill background (border-box area) using ColorProvider
	if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
		// Check if background is semi-transparent (needs backdrop blending)
		// Sample at (0,0) to check - for gradients this may not be fully accurate
		// but is a reasonable heuristic
		sampleColor := style.BackgroundColor.ColorAt(box.Width, box.Height, 0, 0)
		useBackdrop := !sampleColor.IsOpaque()

		if useBackdrop {
			// Use DrawBackdrop to preserve underlying content and blend colors
			backdropCtx := ctx.SubContext(absBorderX, absBorderY, box.Width, box.Height)
			backdropCtx.DrawBackdrop(0, 0, box.Width, box.Height, sampleColor)
		} else {
			// Opaque background - fill with solid color
			for row := 0; row < box.Height; row++ {
				absY := trueAbsBorderY + row
				if absY < ctx.clip.Y || absY >= ctx.clip.Y+ctx.clip.Height {
					continue
				}

				for col := 0; col < box.Width; col++ {
					absX := trueAbsBorderX + col
					if absX < ctx.clip.X || absX >= ctx.clip.X+ctx.clip.Width {
						continue
					}

					// Sample color at this position (works for both solid colors and gradients)
					cellColor := style.BackgroundColor.ColorAt(box.Width, box.Height, col, row)

					cellStyle := uv.Style{Bg: cellColor.toANSI()}
					cell := &uv.Cell{Content: " ", Width: 1, Style: cellStyle}
					ctx.terminal.SetCell(absX, absY, cell)
				}
			}
		}
	}

	// 2. Draw border
	if !style.Border.IsZero() {
		borderCtx := ctx.SubContext(absBorderX, absBorderY, box.Width, box.Height)
		// Set inherited background for border cells using ColorProvider
		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
			parentCallback := ctx.inheritedBgAt

			borderCtx.inheritedBgAt = func(absX, absY int) Color {
				relX := absX - originX
				relY := absY - originY
				cellColor := bg.ColorAt(w, h, relX, relY)

				// Blend with parent if not opaque
				if !cellColor.IsOpaque() && parentCallback != nil {
					inherited := parentCallback(absX, absY)
					if !inherited.IsSet() {
						inherited = Black
					}
					cellColor = cellColor.BlendOver(inherited)
				}
				return cellColor
			}
		}
		borderCtx.DrawBorder(0, 0, box.Width, box.Height, style.Border)
	}

	// 3. Render widget content at content origin
	if renderable, ok := tree.Widget.(Renderable); ok {
		contentCtx := ctx.SubContext(absContentX, absContentY, box.ContentWidth(), box.ContentHeight())
		// Set inherited background for content using ColorProvider
		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
			parentCallback := ctx.inheritedBgAt

			contentCtx.inheritedBgAt = func(absX, absY int) Color {
				relX := absX - originX
				relY := absY - originY
				cellColor := bg.ColorAt(w, h, relX, relY)

				// Blend with parent if not opaque
				if !cellColor.IsOpaque() && parentCallback != nil {
					inherited := parentCallback(absX, absY)
					if !inherited.IsSet() {
						inherited = Black
					}
					cellColor = cellColor.BlendOver(inherited)
				}
				return cellColor
			}
		}
		renderable.Render(contentCtx)
	}

	// 4. Render children at their computed positions
	// If tree.Children is empty but widget has children, the widget handles them in Render() (fallback)
	// If tree.Children is populated, we handle positioning here (new path)
	if len(tree.Children) > 0 {
		// Determine the context for rendering children
		var childClipCtx *RenderContext

		// Stack positions children relative to border-box, not content-box
		// Stack allows children to overflow (e.g., badges with negative positioning)
		_, isStack := tree.Widget.(Stack)
		if isStack {
			// Stack: children positioned relative to border-box, can overflow
			childClipCtx = ctx.OverflowSubContext(absBorderX, absBorderY, box.Width, box.Height)
		} else if box.IsScrollableY() {
			// Scrollable: apply scroll offset so content is shifted up
			usableBox := box.UsableContentBox()
			childClipCtx = ctx.ScrolledSubContext(absContentX, absContentY, usableBox.Width, usableBox.Height, box.ScrollOffsetY)
		} else {
			// Standard containers: children positioned relative to content area
			// Use SubContext which clips children to the container bounds
			// Stack uses OverflowSubContext to allow positioned children to overflow
			usableBox := box.UsableContentBox()
			childClipCtx = ctx.SubContext(absContentX, absContentY, usableBox.Width, usableBox.Height)
		}

		// Set inherited background for children using ColorProvider
		if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
			bg := style.BackgroundColor
			w, h := box.Width, box.Height
			originX, originY := trueAbsBorderX, trueAbsBorderY
			parentCallback := ctx.inheritedBgAt

			childClipCtx.inheritedBgAt = func(absX, absY int) Color {
				relX := absX - originX
				relY := absY - originY
				cellColor := bg.ColorAt(w, h, relX, relY)

				// Blend with parent if not opaque
				if !cellColor.IsOpaque() && parentCallback != nil {
					inherited := parentCallback(absX, absY)
					if !inherited.IsSet() {
						inherited = Black
					}
					cellColor = cellColor.BlendOver(inherited)
				}
				return cellColor
			}
		}

		for i, childTree := range tree.Children {
			if i >= len(tree.Layout.Children) {
				break
			}
			pos := tree.Layout.Children[i]
			// Pass relative positions - childClipCtx.X/Y already contains the origin offset
			r.renderTree(childClipCtx, childTree, pos.X, pos.Y)
		}
	}

	// 4b. Render scrollbar and update ScrollState if widget is scrollable
	if scrollable, ok := tree.Widget.(Scrollable); ok {
		// Update ScrollState with computed dimensions from layout
		// This enables keyboard scrolling to work correctly
		if scrollable.State != nil {
			usableBox := box.UsableContentBox()
			scrollable.State.updateLayout(usableBox.Height, box.VirtualHeight)
		}

		// Render scrollbar if scrolling is enabled and content overflows
		if box.IsScrollableY() && !scrollable.DisableScroll {
			// Get focus state
			focused := ctx.focusManager != nil && ctx.IsFocused(tree.Widget)

			// Create context for scrollbar (at content area, not affected by scroll offset)
			scrollbarCtx := ctx.SubContext(absContentX, absContentY, box.ContentWidth(), box.ContentHeight())
			if style.BackgroundColor != nil && style.BackgroundColor.IsSet() {
				bg := style.BackgroundColor
				w, h := box.Width, box.Height
				originX, originY := trueAbsBorderX, trueAbsBorderY
				scrollbarCtx.inheritedBgAt = func(absX, absY int) Color {
					relX := absX - originX
					relY := absY - originY
					return bg.ColorAt(w, h, relX, relY)
				}
			}
			scrollable.renderScrollbar(scrollbarCtx, box.ScrollOffsetY, focused)
		}
	}

	// 5. Register for hit testing
	// Get widget ID if available
	var widgetID string
	if identifiable, ok := tree.Widget.(Identifiable); ok {
		widgetID = identifiable.WidgetID()
	}
	r.widgetRegistry.Record(tree.Widget, widgetID, Rect{
		X:      absBorderX,
		Y:      absBorderY,
		Width:  box.Width,
		Height: box.Height,
	})
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

// ScrollablesAt returns all Scrollable widgets at the given coordinates,
// ordered from innermost to outermost.
func (r *Renderer) ScrollablesAt(x, y int) []*Scrollable {
	return r.widgetRegistry.ScrollablesAt(x, y)
}

// renderFloats renders all floating widgets collected during the build phase.
// Floats are rendered in order (first registered = bottom, last = top).
// Modal floats render a backdrop before their content.
func (r *Renderer) renderFloats(ctx *RenderContext, buildCtx BuildContext) {
	if r.floatCollector.Len() == 0 {
		return
	}

	for i := range r.floatCollector.entries {
		entry := &r.floatCollector.entries[i]

		// Record focusable count before building so we can find the first focusable
		// in this float's subtree without calling Build() again
		focusableCountBefore := r.focusCollector.Len()

		// Build the float's widget tree to determine its size
		// Use loose constraints - floats size to their content
		constraints := layout.Loose(r.width, r.height)
		floatTree := BuildRenderTree(entry.Child, buildCtx, constraints, r.focusCollector)

		floatWidth := floatTree.Layout.Box.MarginBoxWidth()
		floatHeight := floatTree.Layout.Box.MarginBoxHeight()

		// Calculate position based on config
		var x, y int
		if entry.Config.AnchorID != "" {
			// Anchor-based positioning
			anchor := r.widgetRegistry.WidgetByID(entry.Config.AnchorID)
			x, y = calculateAnchorPosition(anchor, entry.Config.Anchor, floatWidth, floatHeight, entry.Config.Offset)
		} else {
			// Absolute positioning
			x, y = calculateAbsolutePosition(entry.Config.Position, r.width, r.height, floatWidth, floatHeight, entry.Config.Offset)
		}

		// Clamp to screen bounds
		x, y = clampToScreen(x, y, floatWidth, floatHeight, r.width, r.height)

		// Store computed position and size for hit testing
		entry.X = x
		entry.Y = y
		entry.Width = floatWidth
		entry.Height = floatHeight

		// Render modal backdrop if needed
		if entry.Config.Modal {
			r.renderModalBackdrop(ctx, entry.Config.BackdropColor)

			// Track the first focusable in this modal for auto-focus
			// Uses focusables already collected during BuildRenderTree above
			if r.modalFocusTarget == "" {
				r.modalFocusTarget = r.focusCollector.FirstIDAfter(focusableCountBefore)
			}
		}

		// Render the float at its computed position
		r.renderTree(ctx, floatTree, x, y)
	}
}

// renderModalBackdrop renders a semi-transparent backdrop over the entire screen.
func (r *Renderer) renderModalBackdrop(ctx *RenderContext, backdropColor Color) {
	if !backdropColor.IsSet() {
		backdropColor = DefaultModalBackdropColor
	}

	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; {
			// Get existing cell to blend with
			existing := ctx.terminal.CellAt(x, y)
			var bgColor Color
			if existing != nil && existing.Style.Bg != nil {
				bgColor = FromANSI(existing.Style.Bg)
			}
			if !bgColor.IsSet() {
				bgColor = Black
			}

			// Blend backdrop color over existing background
			blended := backdropColor.BlendOver(bgColor)

			// Set cell with blended background, preserving content
			content := " "
			width := 1
			if existing != nil && existing.Content != "" {
				content = existing.Content
				width = existing.Width
				if width < 1 {
					width = 1
				}
			}

			cellStyle := uv.Style{Bg: blended.toANSI()}
			// Dim the foreground text by blending it with the backdrop
			if existing != nil && existing.Style.Fg != nil {
				fgColor := FromANSI(existing.Style.Fg)
				if fgColor.IsSet() {
					// Reduce foreground opacity to make it appear dimmed
					dimmedFg := fgColor.WithAlpha(0.3).BlendOver(blended)
					cellStyle.Fg = dimmedFg.toANSI()
				}
			}

			cell := &uv.Cell{Content: content, Width: width, Style: cellStyle}
			ctx.terminal.SetCell(x, y, cell)

			// Advance by cell width to skip continuation cells for wide characters
			x += width
		}
	}
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
