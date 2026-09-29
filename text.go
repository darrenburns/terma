package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/internal/textutil"
	"github.com/darrenburns/terma/layout"
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
	ID        string           // Optional unique identifier for the widget
	Content   string           // Plain text (used if Spans is empty)
	Spans     []Span           // Rich text segments (takes precedence if non-empty)
	Wrap      WrapMode         // Wrapping mode (default = WrapNone)
	TextAlign TextAlign        // Horizontal alignment (default = TextAlignLeft)
	Width     Dimension        // Deprecated: use Style.Width
	Height    Dimension        // Deprecated: use Style.Height
	Style     Style            // Optional styling (colors, inherited by spans)
	Click     func(MouseEvent) // Optional callback invoked when clicked
	MouseDown func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent) // Optional callback invoked when mouse is released
	Hover     func(HoverEvent) // Optional callback invoked when hover state changes
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

// OnHover is called on hover enter/leave transitions.
// Implements the Hoverable interface.
func (t Text) OnHover(event HoverEvent) {
	if t.Hover != nil {
		t.Hover(event)
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
			result = append(result, textutil.HardWrap(line, maxWidth)...)
		case WrapSoft:
			// Soft wrap: break at word boundaries
			wrapped := ansi.Wordwrap(line, maxWidth, "")
			wrappedLines := strings.Split(wrapped, "\n")
			for _, wl := range wrappedLines {
				if ansi.StringWidth(wl) > maxWidth {
					// Word longer than maxWidth, hard-break it.
					result = append(result, textutil.HardWrap(wl, maxWidth)...)
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
	span  Span
	relX  int // x position relative to line start
	width int
}

// lineData holds all span segments for a single line.
type lineData struct {
	segments []spanSegment
	width    int // total width of the line
}

// spanGrapheme carries a reference to the source style, rather than copying the
// entire style for every character. The iterator keeps no input-sized buffer.
type spanGrapheme struct {
	text  string
	style *SpanStyle
	width int
}

type spanGraphemeIterator struct {
	spans     []Span
	spanIndex int
	remaining string
}

func (it *spanGraphemeIterator) next() (spanGrapheme, bool) {
	for len(it.remaining) == 0 {
		if it.spanIndex >= len(it.spans) {
			return spanGrapheme{}, false
		}
		it.remaining = it.spans[it.spanIndex].Text
		it.spanIndex++
	}
	g, width := ansi.FirstGraphemeCluster(it.remaining, ansi.GraphemeWidth)
	it.remaining = it.remaining[len(g):]
	return spanGrapheme{text: g, style: &it.spans[it.spanIndex-1].Style, width: width}, true
}

// spanLineBuilder joins adjacent graphemes without copying the accumulated
// string for every character. It is never copied while its builder is in use.
// finish returns only the completed line data, then resets the builder.
type spanLineBuilder struct {
	lineData
	text strings.Builder
}

func (line *spanLineBuilder) finish(width int) lineData {
	if len(line.segments) > 0 {
		line.segments[len(line.segments)-1].span.Text = line.text.String()
	}
	result := line.lineData
	result.width = width
	*line = spanLineBuilder{}
	return result
}

func appendStyledGrapheme(line *spanLineBuilder, g spanGrapheme, x *int) {
	if g.width == 0 {
		return
	}
	if len(line.segments) > 0 {
		last := &line.segments[len(line.segments)-1]
		if last.span.Style == *g.style && last.relX+last.width == *x {
			line.text.WriteString(g.text)
			last.width += g.width
			*x += g.width
			return
		}
		last.span.Text = line.text.String()
		line.text.Reset()
	}
	line.text.WriteString(g.text)
	line.segments = append(line.segments, spanSegment{
		span:  Span{Style: *g.style},
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

	// First pass: collect visible spans per line
	lines := t.collectSpanLines(ctx.Width, ctx.Height)

	reuseWidth := !spanBoundariesMayJoin(t.Spans)

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
			if !reuseWidth {
				ctx.DrawSpan(xOffset+seg.relX, y, seg.span, drawBaseStyle)
			} else {
				ctx.drawSpan(xOffset+seg.relX, y, seg.span, drawBaseStyle, seg.width)
			}
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

// collectSpanLines collects only the visible prefix, keeping word lookahead
// bounded by the wrap width rather than materializing all input graphemes.
func (t Text) collectSpanLines(width, height int) []lineData {
	if height == 0 {
		return nil
	}
	it := spanGraphemeIterator{spans: t.Spans}
	collector := spanLineCollector{width: width, height: height}
	if width <= 0 || t.Wrap == WrapNone {
		for g, ok := it.next(); ok; g, ok = it.next() {
			if g.text == "\n" {
				if collector.finish() {
					return collector.lines
				}
			} else if width <= 0 || collector.x+g.width <= width {
				appendStyledGrapheme(&collector.current, g, &collector.x)
			} else if height > 0 && len(collector.lines)+1 == height && collector.x == width {
				// The last visible row is full; later input cannot affect it.
				collector.finish()
				return collector.lines
			}
		}
		collector.finish()
		return collector.lines
	}
	if t.Wrap == WrapHard {
		for g, ok := it.next(); ok; g, ok = it.next() {
			if g.text == "\n" {
				if collector.finish() {
					return collector.lines
				}
			} else if collector.append(g) {
				return collector.lines
			}
		}
		collector.finish()
		return collector.lines
	}
	if spanBoundariesMayJoin(t.Spans) {
		return hardWrapSpanLinesBoundary(collectSpanLinesSoftBoundary(&it, width, height), width, height)
	}
	collector.skipZero = true
	return collectSpanLinesSoft(&it, &collector)
}

// spanLineCollector also hard-wraps overlong soft words as they are emitted,
// reusing the iterator's measured widths instead of segmenting their text again.
type spanLineCollector struct {
	lines    []lineData
	current  spanLineBuilder
	x        int
	width    int
	height   int
	skipZero bool
}

func (c *spanLineCollector) finish() bool {
	c.lines = append(c.lines, c.current.finish(c.x))
	c.x = 0
	return c.height > 0 && len(c.lines) >= c.height
}

func (c *spanLineCollector) append(g spanGrapheme) bool {
	if c.skipZero && g.width == 0 {
		return false
	}
	if c.x > 0 && c.x+g.width > c.width {
		if c.finish() {
			return true
		}
	}
	appendStyledGrapheme(&c.current, g, &c.x)
	return false
}

func collectSpanLinesSoft(it *spanGraphemeIterator, c *spanLineCollector) []lineData {
	// x is the soft line's width, independently of the physical lines emitted
	// for an overlong word. This preserves word/whitespace wrapping decisions.
	x := 0
	var word []spanGrapheme
	var spaceStart spanGraphemeIterator
	spaceCount := 0
	wordWidth, spaceWidth := 0, 0
	streamingWord := false
	wordPresent := false
	flushWord := func() bool {
		if !wordPresent && !streamingWord {
			return false
		}
		for i := 0; i < spaceCount; i++ {
			g, _ := spaceStart.next()
			if c.append(g) {
				return true
			}
			x += g.width
		}
		spaceCount = 0
		spaceWidth = 0
		for _, g := range word {
			if c.append(g) {
				return true
			}
			x += g.width
		}
		word = word[:0]
		wordWidth = 0
		wordPresent = false
		streamingWord = false
		return false
	}
	for {
		before := *it
		g, ok := it.next()
		if !ok {
			break
		}
		if g.text == "\n" {
			if flushWord() {
				return c.lines
			}
			spaceCount = 0
			spaceWidth = 0
			if c.finish() {
				return c.lines
			}
			x = 0
			continue
		}
		if g.text == " " {
			if flushWord() {
				return c.lines
			}
			if spaceCount == 0 {
				spaceStart = before
			}
			spaceCount++
			spaceWidth += g.width
			continue
		}
		if streamingWord {
			if c.append(g) {
				return c.lines
			}
			x += g.width
			continue
		}
		wordPresent = true
		if g.width > 0 {
			word = append(word, g)
		}
		wordWidth += g.width
		if x+spaceWidth+wordWidth > c.width && wordWidth < c.width {
			if x > 0 {
				if c.finish() {
					return c.lines
				}
				x = 0
			}
			spaceCount = 0
			spaceWidth = 0
		}
		if wordWidth >= c.width {
			// Once a word is this wide, no later character can trigger a soft
			// break before it. Emit now so even a huge word stops at height.
			if flushWord() {
				return c.lines
			}
			streamingWord = true
		}
	}
	if flushWord() {
		return c.lines
	}
	c.finish()
	return c.lines
}

// Graphemes split between equally styled spans can join when the segment text
// is drawn (for example, two regional indicators become one flag). Retain the
// previous soft-line segmentation and gradient width for these rare boundaries.
// ASCII boundaries cannot change display width, so they use the streaming path.
func spanBoundariesMayJoin(spans []Span) bool {
	var previous *Span
	for i := range spans {
		span := &spans[i]
		if span.Text == "" {
			continue
		}
		if previous != nil && previous.Style == span.Style &&
			(previous.Text[len(previous.Text)-1] >= 0x80 || span.Text[0] >= 0x80) {
			return true
		}
		previous = span
	}
	return false
}

func collectSpanLinesSoftBoundary(it *spanGraphemeIterator, width, height int) []lineData {
	if width <= 0 {
		panic("soft boundary collector requires positive width")
	}

	var lines []lineData
	var currentLine spanLineBuilder
	x := 0

	var word []spanGrapheme
	wordWidth := 0
	var space []spanGrapheme
	spaceWidth := 0

	flushLine := func() bool {
		lines = append(lines, currentLine.finish(x))
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

	for g, ok := it.next(); ok; g, ok = it.next() {
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

func hardWrapSpanLinesBoundary(lines []lineData, width, height int) []lineData {
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

		var currentLine spanLineBuilder
		x := 0
		for _, seg := range line.segments {
			for remaining := seg.span.Text; len(remaining) > 0; {
				g, gWidth := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
				remaining = remaining[len(g):]
				if x > 0 && x+gWidth > width {
					if appendLine(currentLine.finish(x)) {
						return wrapped
					}
					x = 0
				}
				appendStyledGrapheme(&currentLine, spanGrapheme{
					text:  g,
					style: &seg.span.Style,
					width: gWidth,
				}, &x)
			}
		}
		if appendLine(currentLine.finish(x)) {
			return wrapped
		}
	}

	return wrapped
}
