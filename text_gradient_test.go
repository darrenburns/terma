package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextGradientBackground(t *testing.T) {
	for _, tc := range []struct {
		name string
		text Text
	}{
		{"plain", Text{Content: "ABCDE"}},
		{"spans", Text{Spans: []Span{{Text: "AB", Style: SpanStyle{Bold: true}}, {Text: "CDE"}}}},
		{"wrapped", Text{Content: "ABCDEFGHIJKLMNO", Wrap: WrapHard}},
		{"wrapped spans", Text{Spans: []Span{{Text: "ABCDEFGHIJ"}, {Text: "KLMNO", Style: SpanStyle{Italic: true}}}, Wrap: WrapHard}},
		{"aligned", Text{Content: "AB\nC", TextAlign: TextAlignCenter}},
		{"border and padding", Text{Content: "ABCDE", Wrap: WrapHard, Style: Style{Padding: EdgeInsetsAll(1), Border: SquareBorder(White)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gradient := NewGradient(RGB(255, 0, 0), RGB(0, 0, 255)).WithAngle(45)
			tc.text.Style.Width = Cells(9)
			tc.text.Style.Height = Cells(6)
			tc.text.Style.BackgroundColor = gradient
			buf, width, height := RenderToBufferWithSize(tc.text, 12, 9)
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					require.Equal(t, gradient.ColorAt(width, height, x, y), FromANSI(buf.CellAt(x, y).Style.Bg), "background at (%d,%d)", x, y)
				}
			}
			assert.NotEqual(t, buf.CellAt(0, 0).Style.Bg, buf.CellAt(width-1, height-1).Style.Bg)
		})
	}
}

func TestTextGradientBackgroundSpanOverrides(t *testing.T) {
	gradient := NewGradient(RGB(255, 0, 0), RGB(0, 0, 255)).WithAngle(90)
	translucent := RGBA(0, 255, 0, 0.5)
	foreground := RGBA(255, 255, 255, 0.5)
	buf := RenderToBuffer(Text{
		Spans: []Span{
			{Text: "AB", Style: SpanStyle{Bold: true}},
			{Text: "C", Style: SpanStyle{Background: White, Foreground: Black}},
			{Text: "DE", Style: SpanStyle{Background: translucent, Foreground: foreground}},
		},
		Style: Style{BackgroundColor: gradient, ForegroundColor: White},
	}, 5, 1)
	for x := 0; x < 5; x++ {
		wantBg := gradient.ColorAt(5, 1, x, 0)
		wantFg := White
		if x == 2 {
			wantBg, wantFg = White, Black
		} else if x > 2 {
			wantBg = translucent.BlendOver(wantBg)
			wantFg = foreground.BlendOver(wantBg)
		}
		assert.Equal(t, wantBg, FromANSI(buf.CellAt(x, 0).Style.Bg), "background at %d", x)
		assert.Equal(t, wantFg, FromANSI(buf.CellAt(x, 0).Style.Fg), "foreground at %d", x)
	}
	assert.NotZero(t, buf.CellAt(0, 0).Style.Attrs&uv.AttrBold)
}

func TestTextGradientBackgroundAlphaAndForeground(t *testing.T) {
	gradient := NewGradient(RGBA(255, 0, 0, 0.5), RGBA(0, 0, 255, 0.5)).WithAngle(90)
	foreground := NewGradient(RGBA(255, 255, 255, 0.5), RGBA(0, 255, 0, 0.5)).WithAngle(90)
	base := RGB(20, 40, 60)
	for _, rich := range []bool{false, true} {
		text := Text{Content: "ABCDE", Style: Style{BackgroundColor: gradient, ForegroundColor: foreground}}
		if rich {
			text.Spans = []Span{{Text: "ABCDE"}}
		}
		buf := RenderToBuffer(Column{Style: Style{BackgroundColor: base}, Children: []Widget{text}}, 5, 1)
		for x := 0; x < 5; x++ {
			bg := gradient.ColorAt(5, 1, x, 0).BlendOver(base)
			assert.Equal(t, bg, FromANSI(buf.CellAt(x, 0).Style.Bg), "rich=%v background at %d", rich, x)
			assert.Equal(t, foreground.ColorAt(5, 1, x, 0).BlendOver(bg), FromANSI(buf.CellAt(x, 0).Style.Fg), "rich=%v foreground at %d", rich, x)
		}
	}
}

func TestTextGradientBackgroundWideGlyph(t *testing.T) {
	gradient := NewGradient(RGB(255, 0, 0), RGB(0, 0, 255)).WithAngle(90)
	for _, rich := range []bool{false, true} {
		text := Text{Content: "你BC界", Style: Style{BackgroundColor: gradient}}
		if rich {
			text.Spans = []Span{{Text: "你BC界"}}
		}
		buf := RenderToBuffer(text, 6, 1)
		// A terminal wide glyph has one style across its occupied columns. Sample
		// its leading column; UV reserves empty continuation cells for the rest.
		for _, x := range []int{0, 2, 3, 4} {
			assert.Equal(t, gradient.ColorAt(6, 1, x, 0), FromANSI(buf.CellAt(x, 0).Style.Bg), "rich=%v glyph at %d", rich, x)
		}
		assert.Equal(t, 2, buf.CellAt(0, 0).Width)
		assert.Equal(t, "你", buf.CellAt(0, 0).Content)
		assert.Equal(t, 0, buf.CellAt(1, 0).Width)
	}
}

func TestSnapshot_Text_GradientBackground(t *testing.T) {
	widget := Column{Spacing: 1, Children: []Widget{
		Text{Content: "Direct Text gradient", Style: Style{Width: Cells(28), BackgroundColor: NewGradient(RGB(150, 40, 20), RGB(20, 40, 150)).WithAngle(90)}},
		Text{Content: "Wrapped and centered text", Wrap: WrapSoft, TextAlign: TextAlignCenter, Style: Style{Width: Cells(28), Height: Cells(7), Padding: EdgeInsetsAll(1), Border: RoundedBorder(White), BackgroundColor: NewGradient(RGB(10, 90, 70), RGB(90, 10, 70)).WithAngle(45)}},
		Text{Spans: []Span{{Text: "Inherited "}, {Text: "override", Style: SpanStyle{Background: White, Foreground: Black}}, {Text: " gradient"}}, Style: Style{Width: Cells(28), BackgroundColor: NewGradient(RGB(150, 40, 20), RGB(20, 40, 150)).WithAngle(90)}},
	}}
	AssertSnapshot(t, widget, 30, 12, "Direct Text backgrounds sample gradients across glyphs, wrapped lines, alignment, padding and borders; an explicit span background overrides its part of the gradient.")
}
