package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type shadowTestApp struct {
	position Signal[Offset]
	visible  Signal[bool]
	effect   Signal[FloatShadow]
	text     Signal[string]
	underlay string
}

func (a *shadowTestApp) Build(BuildContext) Widget {
	underlay := a.underlay
	if underlay == "" {
		underlay = strings.Repeat("underlying text 0123456789\n", 10)
	}
	return Column{Children: []Widget{
		Text{Content: underlay, Width: Cells(30), Height: Cells(10), Style: Style{ForegroundColor: White, BackgroundColor: RGB(40, 50, 60)}},
		shadowTestFloat{app: a},
	}}
}

type shadowTestFloat struct{ app *shadowTestApp }

func (f shadowTestFloat) Build(BuildContext) Widget {
	a := f.app
	shadow := a.effect.Get()
	return Floating{Visible: a.visible.Get(), Config: FloatConfig{Offset: a.position.Get(), Shadow: &shadow}, Child: SignalText(a.text, func(s string) string { return s })}
}

func TestFloatShadowRetainedDamage(t *testing.T) {
	a := &shadowTestApp{position: NewSignal(Offset{X: 3, Y: 2}), visible: NewSignal(true), effect: NewSignal(FloatShadow{Color: Black.WithAlpha(.5), Offset: Offset{X: 1, Y: 1}, BlurRadius: 2, Spread: 1}), text: NewSignal("FLOAT")}
	m, r := renderForMouse(a, 30, 10)
	_ = m
	check := func() {
		expected := RenderToBuffer(a, 30, 10)
		for y := 0; y < 10; y++ {
			for x := 0; x < 30; x++ {
				require.Equal(t, expected.CellAt(x, y), r.terminal.CellAt(x, y), "cell %d,%d", x, y)
			}
		}
	}
	check()
	a.position.Set(Offset{X: 18, Y: 6})
	r.Update(a)
	check()
	a.effect.Set(FloatShadow{Color: RGB(20, 180, 255).WithAlpha(.6), BlurRadius: 3})
	r.Update(a)
	check()
	a.text.Set("SMALL")
	r.Update(a)
	check()
	a.visible.Set(false)
	r.Update(a)
	check()
	assert.NotContains(t, r.ScreenText(), "SMALL")
}

func TestFloatShadowDamagePreservesWideGlyphs(t *testing.T) {
	a := &shadowTestApp{position: NewSignal(Offset{X: 4, Y: 2}), visible: NewSignal(true), effect: NewSignal(FloatShadow{Color: Black.WithAlpha(.5), BlurRadius: 1}), text: NewSignal("FLOAT"), underlay: strings.Repeat("界界界界界界界界界界界界界界界\n", 10)}
	_, r := renderForMouse(a, 30, 10)
	for _, offset := range []Offset{{X: 20, Y: 6}, {X: 9, Y: 3}} {
		a.position.Set(offset)
		r.Update(a)
		expected := RenderToBuffer(a, 30, 10)
		for y := 0; y < 10; y++ {
			for x := 0; x < 30; x++ {
				require.Equal(t, expected.CellAt(x, y), r.terminal.CellAt(x, y), "cell %d,%d after move %+v", x, y, offset)
			}
		}
	}
}

func TestFloatShadowPreservesWideGlyphAndLink(t *testing.T) {
	buffer := uv.NewBuffer(10, 3)
	cell := uv.Cell{Content: "界", Width: 2, Style: uv.Style{Fg: White.toANSI(), Bg: Blue.toANSI(), Attrs: uv.AttrBold}, Link: uv.Link{URL: "https://example.test"}}
	buffer.SetCell(2, 1, &cell)
	ctx := NewRenderContext(buffer, 10, 3, nil, nil, BuildContext{}, nil)
	paintFloatShadow(ctx, Rect{X: 2, Y: 1, Width: 2, Height: 1}, FloatShadow{Color: Black.WithAlpha(.5), BlurRadius: 1})
	after := buffer.CellAt(2, 1)
	require.NotNil(t, after)
	assert.Equal(t, cell.Content, after.Content)
	assert.Equal(t, cell.Width, after.Width)
	assert.Equal(t, cell.Link, after.Link)
	assert.Equal(t, cell.Style.Attrs, after.Style.Attrs)
	assert.NotEqual(t, cell.Style.Fg, after.Style.Fg)
	assert.Equal(t, 0, buffer.CellAt(3, 1).Width)
}
