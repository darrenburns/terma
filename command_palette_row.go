package terma

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
)

// commandPaletteHintGap is the least space between an item's label and hint.
const commandPaletteHintGap = 2

// commandPaletteItemLine is the first line of a command palette item: the
// label on the left, the hint right-aligned, and markers (the nested-level
// indicator, the current-item check) at the right edge.
//
// The label has priority: the hint gets the space the label leaves, and is
// elided from the left, so a long path keeps the end that tells paths apart.
// Only a label wider than the whole line is cut short, with an ellipsis.
type commandPaletteItemLine struct {
	label      []Span
	labelStyle Style
	hint       string
	hintStyle  Style
	markers    []Span // Drawn at the right edge, styled over hintStyle.
}

func (l commandPaletteItemLine) Build(ctx BuildContext) Widget {
	return l
}

func (l commandPaletteItemLine) GetContentDimensions() (width, height Dimension) {
	return Flex(1), Cells(1)
}

func (l commandPaletteItemLine) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(
		DimensionSet{Width: Flex(1), Height: Cells(1)}, layout.EdgeInsets{}, layout.EdgeInsets{})
	return &layout.BoxNode{
		MinWidth:    minWidth,
		MaxWidth:    maxWidth,
		MinHeight:   minHeight,
		MaxHeight:   maxHeight,
		ExpandWidth: true,
	}
}

func (l commandPaletteItemLine) Render(ctx *RenderContext) {
	markersWidth := spansWidth(l.markers)
	room := max(0, ctx.Width-markersWidth)

	labelEnd := drawSpansWithin(ctx, 0, l.label, l.labelStyle, room)

	if hintRoom := room - labelEnd - commandPaletteHintGap; l.hint != "" && hintRoom > 0 {
		hint := elideStart(l.hint, hintRoom)
		ctx.DrawStyledText(room-ansi.StringWidth(hint), 0, hint, l.hintStyle)
	}

	x := ctx.Width - markersWidth
	for _, span := range l.markers {
		ctx.DrawStyledText(x, 0, span.Text, applySpanStyle(l.hintStyle, span.Style))
		x += ansi.StringWidth(span.Text)
	}
}

func spansWidth(spans []Span) int {
	width := 0
	for _, span := range spans {
		width += ansi.StringWidth(span.Text)
	}
	return width
}

// drawSpansWithin draws spans from x, ending them with an ellipsis if they
// don't fit in width cells. It returns the width drawn.
func drawSpansWithin(ctx *RenderContext, x int, spans []Span, base Style, width int) int {
	if width <= 0 {
		return 0
	}
	if spansWidth(spans) > width {
		spans = truncateSpans(spans, width-1)
		spans = append(spans, Span{Text: "…"})
	}
	drawn := 0
	for _, span := range spans {
		ctx.DrawStyledText(x+drawn, 0, span.Text, applySpanStyle(base, span.Style))
		drawn += ansi.StringWidth(span.Text)
	}
	return drawn
}

// truncateSpans cuts spans to at most width cells.
func truncateSpans(spans []Span, width int) []Span {
	result := make([]Span, 0, len(spans))
	for _, span := range spans {
		if width <= 0 {
			break
		}
		spanWidth := ansi.StringWidth(span.Text)
		if spanWidth > width {
			span.Text = ansi.Truncate(span.Text, width, "")
			spanWidth = ansi.StringWidth(span.Text)
		}
		result = append(result, span)
		width -= spanWidth
	}
	return result
}

// elideStart shortens text to at most width cells by replacing its start
// with an ellipsis, keeping its end.
func elideStart(text string, width int) string {
	if ansi.StringWidth(text) <= width {
		return text
	}
	if width <= 0 {
		return ""
	}
	graphemes := splitGraphemes(text)
	kept := 0
	start := len(graphemes)
	for start > 0 {
		w := graphemeWidth(graphemes[start-1])
		if kept+w > width-1 {
			break
		}
		kept += w
		start--
	}
	return "…" + joinGraphemes(graphemes[start:])
}
