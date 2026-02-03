package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"terma/layout"
)

// WrapMode defines how text should wrap within available width.
type WrapMode int

const (
	// WrapNone disables wrapping - text is truncated if too long.
	WrapNone WrapMode = iota
	// WrapSoft breaks at word boundaries (spaces), only breaking words if necessary (default).
	WrapSoft
	// WrapHard breaks at exact character boundary when line exceeds width.
	WrapHard
)

// TextAlign defines horizontal alignment for text content within available width.
type TextAlign int

const (
	// TextAlignLeft aligns text to the left edge (default).
	TextAlignLeft TextAlign = iota
	// TextAlignCenter centers text horizontally.
	TextAlignCenter
	// TextAlignRight aligns text to the right edge.
	TextAlignRight
)

// Text is a leaf widget that displays text content.
type Text struct {
	ID        string             // Optional unique identifier for the widget
	Content   string             // Plain text (used if Spans is empty)
	Spans     []Span             // Rich text segments (takes precedence if non-empty)
	Wrap      WrapMode           // Wrapping mode (default = WrapNone)
	TextAlign TextAlign          // Horizontal alignment (default = TextAlignLeft)
	Width     Dimension          // Deprecated: use Style.Width
	Height    Dimension          // Deprecated: use Style.Height
	Style     Style              // Optional styling (colors, inherited by spans)
	Click     func(MouseEvent)   // Optional callback invoked when clicked
	MouseDown func(MouseEvent)   // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent)   // Optional callback invoked when mouse is released
	Hover     func(bool)         // Optional callback invoked when hover state changes
}

// Build returns itself as Text is a leaf widget.
func (t Text) Build(ctx BuildContext) Widget {
	return t
}

// WidgetID returns the text widget's unique identifier.
// Implements the Identifiable interface.
func (t Text) WidgetID() string {
	return t.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (t Text) OnClick(event MouseEvent) {
	if t.Click != nil {
		t.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (t Text) OnMouseDown(event MouseEvent) {
	if t.MouseDown != nil {
		t.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (t Text) OnMouseUp(event MouseEvent) {
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (t Text) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// GetContentDimensions returns the width and height dimension preferences.
func (t Text) GetContentDimensions() (width, height Dimension) {
	dims := t.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = t.Width
	}
	if height.IsUnset() {
		height = t.Height
	}
	return width, height
}

// GetStyle returns the style of the text widget.
func (t Text) GetStyle() Style {
	return t.Style
}

// BuildLayoutNode builds a layout node for this Text widget.
// Implements the LayoutNodeBuilder interface.
func (t Text) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	// Get the text content (spans concatenated or plain content)
	content := t.textContent()

	padding := toLayoutEdgeInsets(t.Style.Padding)
	border := borderToEdgeInsets(t.Style.Border)
	dims := GetWidgetDimensionSet(t)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.TextNode{
		Content:   content,
		Wrap:      toLayoutWrapMode(t.Wrap),
		Padding:   padding,
		Border:    border,
		Margin:    toLayoutEdgeInsets(t.Style.Margin),
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
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

// textContent returns the effective text content.
// If Spans is non-empty, concatenates all span text; otherwise returns Content.
func (t Text) textContent() string {
	if len(t.Spans) > 0 {
		var sb strings.Builder
		for _, span := range t.Spans {
			sb.WriteString(span.Text)
		}
		return sb.String()
	}
	return t.Content
}

// wrapText wraps the given text content to fit within maxWidth based on the wrap mode.
func wrapText(content string, maxWidth int, mode WrapMode) []string {
	if maxWidth <= 0 || mode == WrapNone {
		return strings.Split(content, "\n")
	}

	inputLines := strings.Split(content, "\n")
	var result []string

	for _, line := range inputLines {
		lineWidth := ansi.StringWidth(line)

		// If line fits, add as-is
		if lineWidth <= maxWidth {
			result = append(result, line)
			continue
		}

		switch mode {
		case WrapHard:
			// Hard wrap: break at exact character boundary using Truncate
			remaining := line
			for len(remaining) > 0 {
				if ansi.StringWidth(remaining) <= maxWidth {
					result = append(result, remaining)
					break
				}
				chunk := ansi.Truncate(remaining, maxWidth, "")
				result = append(result, chunk)
				remaining = remaining[len(chunk):]
			}
		case WrapSoft:
			// Soft wrap: break at word boundaries
			wrapped := ansi.Wordwrap(line, maxWidth, "")
			wrappedLines := strings.Split(wrapped, "\n")
			for _, wl := range wrappedLines {
				if ansi.StringWidth(wl) > maxWidth {
					// Word longer than maxWidth, hard-break it
					remaining := wl
					for len(remaining) > 0 {
						if ansi.StringWidth(remaining) <= maxWidth {
							result = append(result, remaining)
							break
						}
						chunk := ansi.Truncate(remaining, maxWidth, "")
						result = append(result, chunk)
						remaining = remaining[len(chunk):]
					}
				} else {
					result = append(result, wl)
				}
			}
		}
	}

	return result
}

// alignLine calculates the x-offset for a line based on text alignment.
func alignLine(lineWidth, availableWidth int, align TextAlign) int {
	switch align {
	case TextAlignCenter:
		return (availableWidth - lineWidth) / 2
	case TextAlignRight:
		return availableWidth - lineWidth
	default:
		return 0
	}
}

// Render draws the text to the render context.
func (t Text) Render(ctx *RenderContext) {
	if len(t.Spans) > 0 {
		t.renderSpans(ctx)
	} else {
		t.renderPlain(ctx)
	}
}

// renderPlain renders plain text content.
func (t Text) renderPlain(ctx *RenderContext) {
	// Start with the full style, then ensure foreground color has a default
	style := t.Style
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = ctx.buildContext.Theme().Text
	}
	drawStyle := style
	if drawStyle.BackgroundColor != nil && drawStyle.BackgroundColor.IsSet() {
		drawStyle.BackgroundColor = nil
	}

	// Get lines with wrapping applied
	lines := wrapText(t.Content, ctx.Width, t.Wrap)

	// Check if we need to draw text and padding separately
	// (when strikethrough/underline is set but FillLine is false)
	hasLineDecoration := style.Strikethrough || style.Underline != UnderlineNone
	separatePadding := hasLineDecoration && !style.FillLine

	for i := 0; i < ctx.Height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		// Truncate line if it exceeds width (fallback for WrapNone or edge cases)
		lineWidth := ansi.StringWidth(line)
		if lineWidth > ctx.Width {
			line = ansi.Truncate(line, ctx.Width, "")
			lineWidth = ctx.Width
		}

		// Calculate alignment offset
		xOffset := alignLine(lineWidth, ctx.Width, t.TextAlign)
		leftPadding := xOffset
		rightPadding := ctx.Width - lineWidth - xOffset

		if separatePadding && lineWidth < ctx.Width {
			// Style for padding (without strikethrough/underline)
			paddingStyle := drawStyle
			paddingStyle.Strikethrough = false
			paddingStyle.Underline = UnderlineNone

			// Draw left padding
			if leftPadding > 0 {
				ctx.DrawStyledText(0, i, strings.Repeat(" ", leftPadding), paddingStyle)
			}
			// Draw text with full style (including strikethrough/underline)
			ctx.DrawStyledText(xOffset, i, line, drawStyle)
			// Draw right padding
			if rightPadding > 0 {
				ctx.DrawStyledText(xOffset+lineWidth, i, strings.Repeat(" ", rightPadding), paddingStyle)
			}
		} else {
			// Build aligned line with padding
			alignedLine := strings.Repeat(" ", leftPadding) + line + strings.Repeat(" ", rightPadding)
			ctx.DrawStyledText(0, i, alignedLine, drawStyle)
		}
	}
}

// spanSegment holds a span segment at a relative x position within a line.
type spanSegment struct {
	span   Span
	relX   int // x position relative to line start
	width  int
}

// lineData holds all span segments for a single line.
type lineData struct {
	segments []spanSegment
	width    int // total width of the line
}

type styledGrapheme struct {
	text  string
	style SpanStyle
	width int
}

func collectSpanGraphemes(spans []Span) []styledGrapheme {
	if len(spans) == 0 {
		return nil
	}
	result := make([]styledGrapheme, 0, len(spans))
	for _, span := range spans {
		if span.Text == "" {
			continue
		}
		for _, g := range splitGraphemes(span.Text) {
			result = append(result, styledGrapheme{
				text:  g,
				style: span.Style,
				width: graphemeWidth(g),
			})
		}
	}
	return result
}

func appendStyledGrapheme(line *lineData, g styledGrapheme, x *int) {
	if g.width == 0 {
		return
	}
	if len(line.segments) > 0 {
		last := &line.segments[len(line.segments)-1]
		if last.span.Style == g.style && last.relX+last.width == *x {
			last.span.Text += g.text
			last.width += g.width
			*x += g.width
			return
		}
	}
	line.segments = append(line.segments, spanSegment{
		span:  Span{Text: g.text, Style: g.style},
		relX:  *x,
		width: g.width,
	})
	*x += g.width
}

// renderSpans renders rich text with multiple styled spans.
func (t Text) renderSpans(ctx *RenderContext) {
	// Start with the full style, then ensure foreground color has a default
	baseStyle := t.Style
	if baseStyle.ForegroundColor == nil || !baseStyle.ForegroundColor.IsSet() {
		baseStyle.ForegroundColor = ctx.buildContext.Theme().Text
	}
	drawBaseStyle := baseStyle
	if drawBaseStyle.BackgroundColor != nil && drawBaseStyle.BackgroundColor.IsSet() {
		drawBaseStyle.BackgroundColor = nil
	}

	// First pass: collect all spans per line
	lines := t.collectSpanLines(ctx.Width, ctx.Height)

	// Second pass: render each line with alignment
	for y, line := range lines {
		if y >= ctx.Height {
			break
		}

		// Calculate alignment offset for this line
		xOffset := alignLine(line.width, ctx.Width, t.TextAlign)

		// Draw left padding if needed
		if xOffset > 0 {
			ctx.DrawStyledText(0, y, strings.Repeat(" ", xOffset), drawBaseStyle)
		}

		// Draw all spans in the line
		for _, seg := range line.segments {
			ctx.DrawSpan(xOffset+seg.relX, y, seg.span, drawBaseStyle)
		}

		// Draw right padding to fill remaining width
		rightPadding := ctx.Width - xOffset - line.width
		if rightPadding > 0 {
			ctx.DrawStyledText(xOffset+line.width, y, strings.Repeat(" ", rightPadding), drawBaseStyle)
		}
	}

	// Fill any remaining lines with empty space
	for y := len(lines); y < ctx.Height; y++ {
		ctx.DrawStyledText(0, y, strings.Repeat(" ", ctx.Width), drawBaseStyle)
	}
}

// collectSpanLines collects all span segments organized by line.
func (t Text) collectSpanLines(width, height int) []lineData {
	if height == 0 {
		return nil
	}
	graphemes := collectSpanGraphemes(t.Spans)
	if len(graphemes) == 0 {
		return []lineData{{}}
	}

	if width <= 0 || t.Wrap == WrapNone {
		return collectSpanLinesNoWrap(graphemes, width, height)
	}

	switch t.Wrap {
	case WrapHard:
		return collectSpanLinesHard(graphemes, width, height)
	default:
		lines := collectSpanLinesSoft(graphemes, width, height)
		return hardWrapSpanLines(lines, width, height)
	}
}

func collectSpanLinesNoWrap(graphemes []styledGrapheme, width, height int) []lineData {
	var lines []lineData
	var currentLine lineData
	x := 0

	flushLine := func() bool {
		currentLine.width = x
		lines = append(lines, currentLine)
		currentLine = lineData{}
		x = 0
		return height > 0 && len(lines) >= height
	}

	for _, g := range graphemes {
		if g.text == "\n" {
			if flushLine() {
				return lines
			}
			continue
		}
		if width > 0 && x+g.width > width {
			continue
		}
		appendStyledGrapheme(&currentLine, g, &x)
	}

	flushLine()
	if height > 0 && len(lines) > height {
		return lines[:height]
	}
	return lines
}

func collectSpanLinesHard(graphemes []styledGrapheme, width, height int) []lineData {
	if width <= 0 {
		return collectSpanLinesNoWrap(graphemes, width, height)
	}

	var lines []lineData
	var currentLine lineData
	x := 0

	flushLine := func() bool {
		currentLine.width = x
		lines = append(lines, currentLine)
		currentLine = lineData{}
		x = 0
		return height > 0 && len(lines) >= height
	}

	for _, g := range graphemes {
		if g.text == "\n" {
			if flushLine() {
				return lines
			}
			continue
		}

		if x > 0 && x+g.width > width {
			if flushLine() {
				return lines
			}
		}

		appendStyledGrapheme(&currentLine, g, &x)
	}

	flushLine()
	if height > 0 && len(lines) > height {
		return lines[:height]
	}
	return lines
}

func collectSpanLinesSoft(graphemes []styledGrapheme, width, height int) []lineData {
	if width <= 0 {
		return collectSpanLinesNoWrap(graphemes, width, height)
	}

	var lines []lineData
	var currentLine lineData
	x := 0

	var word []styledGrapheme
	wordWidth := 0
	var space []styledGrapheme
	spaceWidth := 0

	flushLine := func() bool {
		currentLine.width = x
		lines = append(lines, currentLine)
		currentLine = lineData{}
		x = 0
		return height > 0 && len(lines) >= height
	}

	flushWord := func() {
		if len(word) == 0 {
			return
		}
		if len(space) > 0 {
			for _, g := range space {
				appendStyledGrapheme(&currentLine, g, &x)
			}
			space = nil
			spaceWidth = 0
		}
		for _, g := range word {
			appendStyledGrapheme(&currentLine, g, &x)
		}
		word = nil
		wordWidth = 0
	}

	for _, g := range graphemes {
		if g.text == "\n" {
			flushWord()
			space = nil
			spaceWidth = 0
			if flushLine() {
				return lines
			}
			continue
		}

		if g.text == " " {
			flushWord()
			space = append(space, g)
			spaceWidth += g.width
			continue
		}

		word = append(word, g)
		wordWidth += g.width

		if x+spaceWidth+wordWidth > width && wordWidth < width {
			if x > 0 {
				if flushLine() {
					return lines
				}
			}
			space = nil
			spaceWidth = 0
		}
	}

	flushWord()
	flushLine()
	if height > 0 && len(lines) > height {
		return lines[:height]
	}
	return lines
}

func hardWrapSpanLines(lines []lineData, width, height int) []lineData {
	if width <= 0 {
		return lines
	}

	var wrapped []lineData

	appendLine := func(line lineData) bool {
		wrapped = append(wrapped, line)
		return height > 0 && len(wrapped) >= height
	}

	for _, line := range lines {
		if line.width <= width {
			if appendLine(line) {
				return wrapped
			}
			continue
		}

		var currentLine lineData
		x := 0
		for _, seg := range line.segments {
			for _, g := range splitGraphemes(seg.span.Text) {
				gWidth := graphemeWidth(g)
				if x > 0 && x+gWidth > width {
					currentLine.width = x
					if appendLine(currentLine) {
						return wrapped
					}
					currentLine = lineData{}
					x = 0
				}
				appendStyledGrapheme(&currentLine, styledGrapheme{
					text:  g,
					style: seg.span.Style,
					width: gWidth,
				}, &x)
			}
		}
		currentLine.width = x
		if appendLine(currentLine) {
			return wrapped
		}
	}

	return wrapped
}
