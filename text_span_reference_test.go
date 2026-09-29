package terma

import (
	"github.com/charmbracelet/x/ansi"
	"strings"
)

// Reference collector retained to compare streaming output with the previous implementation.
type referenceStyledGrapheme struct {
	text  string
	style SpanStyle
	width int
}

func referenceCollectSpanGraphemes(spans []Span) []referenceStyledGrapheme {
	if len(spans) == 0 {
		return nil
	}
	result := make([]referenceStyledGrapheme, 0, len(spans))
	for _, span := range spans {
		for remaining := span.Text; len(remaining) > 0; {
			g, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
			result = append(result, referenceStyledGrapheme{
				text:  g,
				style: span.Style,
				width: width,
			})
			remaining = remaining[len(g):]
		}
	}
	return result
}

// referenceSpanLineBuilder joins adjacent graphemes without copying the accumulated
// string for every character. It is never copied while its builder is in use.
// finish returns only the completed line data, then resets the builder.
type referenceSpanLineBuilder struct {
	lineData
	text strings.Builder
}

func (line *referenceSpanLineBuilder) finish(width int) lineData {
	if len(line.segments) > 0 {
		line.segments[len(line.segments)-1].span.Text = line.text.String()
	}
	result := line.lineData
	result.width = width
	*line = referenceSpanLineBuilder{}
	return result
}

func referenceAppendStyledGrapheme(line *referenceSpanLineBuilder, g referenceStyledGrapheme, x *int) {
	if g.width == 0 {
		return
	}
	if len(line.segments) > 0 {
		last := &line.segments[len(line.segments)-1]
		if last.span.Style == g.style && last.relX+last.width == *x {
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
		span:  Span{Style: g.style},
		relX:  *x,
		width: g.width,
	})
	*x += g.width
}

// referenceCollectSpanLines collects all span segments organized by line.
func (t Text) referenceCollectSpanLines(width, height int) []lineData {
	if height == 0 {
		return nil
	}
	graphemes := referenceCollectSpanGraphemes(t.Spans)
	if len(graphemes) == 0 {
		return []lineData{{}}
	}

	if width <= 0 || t.Wrap == WrapNone {
		return referenceCollectSpanLinesNoWrap(graphemes, width, height)
	}

	switch t.Wrap {
	case WrapHard:
		return referenceCollectSpanLinesHard(graphemes, width, height)
	default:
		lines := referenceCollectSpanLinesSoft(graphemes, width, height)
		return referenceHardWrapSpanLines(lines, width, height)
	}
}

func referenceCollectSpanLinesNoWrap(graphemes []referenceStyledGrapheme, width, height int) []lineData {
	var lines []lineData
	var currentLine referenceSpanLineBuilder
	x := 0

	flushLine := func() bool {
		lines = append(lines, currentLine.finish(x))
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
		referenceAppendStyledGrapheme(&currentLine, g, &x)
	}

	flushLine()
	if height > 0 && len(lines) > height {
		return lines[:height]
	}
	return lines
}

func referenceCollectSpanLinesHard(graphemes []referenceStyledGrapheme, width, height int) []lineData {
	if width <= 0 {
		return referenceCollectSpanLinesNoWrap(graphemes, width, height)
	}

	var lines []lineData
	var currentLine referenceSpanLineBuilder
	x := 0

	flushLine := func() bool {
		lines = append(lines, currentLine.finish(x))
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

		referenceAppendStyledGrapheme(&currentLine, g, &x)
	}

	flushLine()
	if height > 0 && len(lines) > height {
		return lines[:height]
	}
	return lines
}

func referenceCollectSpanLinesSoft(graphemes []referenceStyledGrapheme, width, height int) []lineData {
	if width <= 0 {
		return referenceCollectSpanLinesNoWrap(graphemes, width, height)
	}

	var lines []lineData
	var currentLine referenceSpanLineBuilder
	x := 0

	var word []referenceStyledGrapheme
	wordWidth := 0
	var space []referenceStyledGrapheme
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
				referenceAppendStyledGrapheme(&currentLine, g, &x)
			}
			space = nil
			spaceWidth = 0
		}
		for _, g := range word {
			referenceAppendStyledGrapheme(&currentLine, g, &x)
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

func referenceHardWrapSpanLines(lines []lineData, width, height int) []lineData {
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

		var currentLine referenceSpanLineBuilder
		x := 0
		for _, seg := range line.segments {
			for _, g := range splitGraphemes(seg.span.Text) {
				gWidth := graphemeWidth(g)
				if x > 0 && x+gWidth > width {
					if appendLine(currentLine.finish(x)) {
						return wrapped
					}
					x = 0
				}
				referenceAppendStyledGrapheme(&currentLine, referenceStyledGrapheme{
					text:  g,
					style: seg.span.Style,
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
