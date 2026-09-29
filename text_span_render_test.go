package terma

import (
	uv "github.com/charmbracelet/ultraviolet"
	"reflect"
	"strings"
	"testing"
)

func TestSpanStreamingPreservesGradientRendering(t *testing.T) {
	for _, spans := range [][]Span{
		{{Text: "abc 你好 é abcdefgh", Style: SpanStyle{Bold: true}}, {Text: "\nemoji 👨‍👩‍👧‍👦", Style: SpanStyle{Italic: true, Underline: UnderlineSingle}}},
		{{Text: "🇬"}, {Text: "🇧 abc"}},
		{{Text: "👨‍"}, {Text: "👩 x"}},
	} {
		for _, wrap := range []WrapMode{WrapNone, WrapHard, WrapSoft} {
			for _, align := range []TextAlign{TextAlignLeft, TextAlignCenter, TextAlignRight} {
				for _, width := range []int{1, 3, 8, 20} {
					const height = 5
					style := Style{ForegroundColor: NewGradient(Red, Blue).WithAngle(90)}
					widget := Text{Spans: spans, Wrap: wrap, TextAlign: align, Style: style}
					got, want := uv.NewBuffer(width, height), uv.NewBuffer(width, height)
					widget.Render(NewRenderContext(got, width, height, nil, nil, BuildContext{}, nil))
					ctx := NewRenderContext(want, width, height, nil, nil, BuildContext{}, nil)
					lines := widget.referenceCollectSpanLines(width, height)
					for y, line := range lines {
						offset := alignLine(line.width, width, align)
						if offset > 0 {
							ctx.DrawStyledText(0, y, strings.Repeat(" ", offset), style)
						}
						for _, seg := range line.segments {
							ctx.DrawSpan(offset+seg.relX, y, seg.span, style)
						}
						if padding := width - offset - line.width; padding > 0 {
							ctx.DrawStyledText(offset+line.width, y, strings.Repeat(" ", padding), style)
						}
					}
					for y := len(lines); y < height; y++ {
						ctx.DrawStyledText(0, y, strings.Repeat(" ", width), style)
					}
					for y := 0; y < height; y++ {
						for x := 0; x < width; x++ {
							if !reflect.DeepEqual(got.CellAt(x, y), want.CellAt(x, y)) {
								t.Fatalf("wrap=%d align=%d width=%d cell=(%d,%d) got=%#v want=%#v", wrap, align, width, x, y, got.CellAt(x, y), want.CellAt(x, y))
							}
						}
					}
				}
			}
		}
	}
}
