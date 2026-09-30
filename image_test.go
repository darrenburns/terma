package terma

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func testImage(t testing.TB, w, h int) *ImageResource {
	t.Helper()
	p := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p.SetNRGBA(x, y, color.NRGBA{uint8(x * 255 / max(1, w-1)), uint8(y * 255 / max(1, h-1)), 120, 200})
		}
	}
	r, err := NewImageResource(p)
	require.NoError(t, err)
	return r
}
func imageTestRenderer(w, h int) *Renderer {
	return NewRenderer(uv.NewScreenBuffer(w, h), w, h, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
}
func TestImageResourceImmutable(t *testing.T) {
	_, err := NewImageResource(nil)
	require.Error(t, err)
	var nilImage *image.NRGBA
	_, err = NewImageResource(nilImage)
	require.Error(t, err)
	_, err = NewImageResource(image.NewNRGBA(image.Rectangle{}))
	require.Error(t, err)
	p := image.NewNRGBA(image.Rect(5, 7, 7, 9))
	p.SetNRGBA(5, 7, color.NRGBA{200, 40, 90, 100})
	r, err := NewImageResource(p)
	require.NoError(t, err)
	p.SetNRGBA(5, 7, color.NRGBA{})
	require.Equal(t, image.Rect(0, 0, 2, 2), r.pixels.Bounds())
	require.Equal(t, color.NRGBA{200, 40, 90, 100}, r.pixels.NRGBAAt(0, 0))
}
func TestImageFitAndLayout(t *testing.T) {
	src := testImage(t, 160, 80)
	for _, tc := range []struct {
		fit  ImageFit
		dest Rect
		crop image.Rectangle
	}{{ImageContain, Rect{0, 2, 20, 5}, image.Rect(0, 0, 160, 80)}, {ImageCover, Rect{0, 0, 20, 10}, image.Rect(40, 0, 120, 80)}, {ImageStretch, Rect{0, 0, 20, 10}, image.Rect(0, 0, 160, 80)}} {
		m := mapImage(src, Rect{Width: 20, Height: 10}, tc.fit, 8, 16)
		require.Equal(t, tc.dest, m.dest)
		require.Equal(t, tc.crop, m.crop)
	}
	for _, protocol := range []string{"auto", "kitty", "sixel", "blocks"} {
		t.Run(protocol, func(t *testing.T) {
			t.Setenv("TERMA_IMAGE_PROTOCOL", protocol)
			r := imageTestRenderer(50, 30)
			r.Render(Image{Source: src, Style: Style{Padding: EdgeInsetsAll(1), Margin: EdgeInsetsAll(1), Border: RoundedBorder(White)}})
			require.Equal(t, 24, r.lastLayoutWidth)
			require.Equal(t, 9, r.lastLayoutHeight)
			require.Contains(t, r.ScreenText(), "▀")
			require.NotContains(t, r.ScreenText(), string(rune(0x10eeee)))
		})
	}
}
func TestImageBufferMutations(t *testing.T) {
	for _, op := range []struct {
		name string
		run  func(*imageBuffer)
	}{
		{"identical", func(b *imageBuffer) { c := *b.CellAt(2, 1); b.SetCell(2, 1, &c) }},
		{"clear", func(b *imageBuffer) { b.Clear() }},
		{"clear area", func(b *imageBuffer) { b.ClearArea(uv.Rect(2, 1, 1, 1)) }},
		{"fill", func(b *imageBuffer) { b.Fill(&uv.Cell{Content: "x", Width: 1}) }},
		{"fill area", func(b *imageBuffer) { b.FillArea(&uv.Cell{Content: "x", Width: 1}, uv.Rect(2, 1, 1, 1)) }},
		{"resize", func(b *imageBuffer) { b.Resize(6, 4) }},
		{"wide", func(b *imageBuffer) { b.SetCell(1, 1, &uv.Cell{Content: "界", Width: 2}) }},
		{"pointer mutation", func(b *imageBuffer) { b.CellAt(2, 1).Content = "x" }},
	} {
		t.Run(op.name, func(t *testing.T) {
			b := newImageBuffer(6, 4)
			ctx := NewRenderContext(b, 6, 4, nil, nil, BuildContext{}, nil)
			ctx.DrawImage(0, 0, 6, 4, testImage(t, 6, 8), ImageStretch)
			rec := b.cells[b.index(2, 1)].owner
			require.True(t, b.owns(2, 1, rec))
			op.run(b)
			require.False(t, b.owns(2, 1, rec))
		})
	}
	b := newImageBuffer(6, 4)
	b.SetCell(1, 1, &uv.Cell{Content: "界", Width: 2})
	rec := &imageRecord{visible: Rect{Width: 6, Height: 4}}
	b.claim(1, 1, rec, *b.CellAt(1, 1), color.NRGBA{})
	b.SetCell(2, 1, nil)
	require.Nil(t, b.cells[b.index(1, 1)].owner)
}
func TestImageBackdropAndClip(t *testing.T) {
	b := newImageBuffer(8, 4)
	ctx := NewRenderContext(b, 8, 4, nil, nil, BuildContext{}, nil)
	ctx.clip = Rect{X: 2, Y: 1, Width: 3, Height: 2}
	ctx.DrawImage(0, 0, 8, 4, testImage(t, 16, 16), ImageStretch)
	rec := b.cells[b.index(2, 1)].owner
	require.Equal(t, Rect{Width: 8, Height: 4}, rec.visible)
	before := *b.CellAt(2, 1)
	ctx.DrawBackdrop(0, 0, 8, 4, Black.WithAlpha(.5))
	require.Equal(t, "▀", b.CellAt(2, 1).Content)
	require.False(t, b.owns(2, 1, rec))
	require.False(t, before.Equal(b.CellAt(2, 1)))
}

type imageSequenceApp struct {
	source      *ImageResource
	show, cover Signal[bool]
	size        Signal[int]
	scroll      *ScrollState
	paint       Signal[string]
}

func (a *imageSequenceApp) Build(BuildContext) Widget {
	var children []Widget
	if a.show.Get() {
		for i := 0; i < 8; i++ {
			children = append(children, Image{ID: string(rune('a' + i)), Source: a.source, Fit: ImageStretch, Style: Style{Width: Cells(a.size.Get()), Height: Cells(4)}})
		}
	}
	label := SignalText(a.paint, func(s string) string { return s })
	label.LayoutStyle = Style{Width: Cells(20), Height: Cells(1)}
	return Column{Children: []Widget{label, Scrollable{State: a.scroll, Height: Cells(9), Width: Cells(20), Child: Column{Children: children}}, ShowWhen(a.cover.Get(), Floating{Visible: true, Config: FloatConfig{Modal: true, Position: FloatPositionCenter, BackdropColor: Black.WithAlpha(.5)}, Child: Text{Content: "overlay"}})}}
}
func imageRecordSignature(b *imageBuffer) []any {
	if b == nil {
		return nil
	}
	out := make([]any, len(b.cells))
	for y := 0; y < b.buffer.Height(); y++ {
		for x := 0; x < b.buffer.Width(); x++ {
			c := b.cells[b.index(x, y)]
			if c.owner != nil && b.owns(x, y, c.owner) {
				out[b.index(x, y)] = struct {
					Source     *ImageResource
					Mapping    imageMapping
					Visible    Rect
					Background color.NRGBA
				}{c.owner.source, c.owner.mapping, c.owner.visible, c.background}
			}
		}
	}
	return out
}
func TestImageRetainedSequence(t *testing.T) {
	source := testImage(t, 40, 40)
	seq := newReactivitySequence(t, 24, 13, func() *imageSequenceApp {
		return &imageSequenceApp{source, NewSignal(false), NewSignal(false), NewSignal(12), NewScrollState(), NewSignal("clean sibling")}
	})
	frame := func(name string, change func(*imageSequenceApp)) {
		seq.frame(name, change)
		a, e := seq.actual.renderer, seq.expected.renderer
		require.Equal(t, imageRecordSignature(e.images), imageRecordSignature(a.images), name)
		if a.images != nil {
			ap, ep := uv.NewBuffer(24, 13), uv.NewBuffer(24, 13)
			paint := func(b *imageBuffer, p *uv.Buffer) {
				copyImageCells(p, b, 24, 13)
				k := newKittyImages()
				k.paint(p, b, 8, 16)
				for _, u := range k.uploads {
					u.ready = true
					for _, v := range u.placements {
						v.ready = true
					}
				}
				k.paint(p, b, 8, 16)
			}
			paint(a.images, ap)
			paint(e.images, ep)
			require.Zero(t, CompareBuffers(ep, ap, 24, 13).MismatchedCells, name)
		}
	}
	frame("text only", nil)
	require.Nil(t, seq.actual.renderer.images)
	frame("first image", func(a *imageSequenceApp) { a.show.Set(true) })
	frame("clean reuse", nil)
	frame("scroll", func(a *imageSequenceApp) { a.scroll.ScrollDown(3) })
	frame("partial sibling", func(a *imageSequenceApp) { a.paint.Set("changed") })
	frame("resize image", func(a *imageSequenceApp) { a.size.Set(16) })
	frame("backdrop", func(a *imageSequenceApp) { a.cover.Set(true) })
	frame("remove backdrop", func(a *imageSequenceApp) { a.cover.Set(false) })
	frame("last image removed", func(a *imageSequenceApp) { a.show.Set(false) })
	require.Empty(t, seq.actual.renderer.images.records)
	seq.actual.renderer.fullRenderRequired = true
	frame("full after removal", nil)
	require.NotNil(t, seq.actual.renderer.images)
}
func TestImageIndexedEmittedBytes(t *testing.T) {
	for _, profile := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.TrueColor} {
		var out strings.Builder
		renderer := uv.NewTerminalRenderer(&out, nil)
		renderer.SetColorProfile(profile)
		renderer.SetFullscreen(true)
		b := uv.NewBuffer(3, 1)
		for i, id := range []int{16, 0, 17} {
			b.SetCell(i, 0, &uv.Cell{Content: string([]rune{0x10eeee, 0x305, 0x305}), Width: 1, Style: uv.Style{Fg: ansi.IndexedColor(id), UnderlineColor: ansi.IndexedColor(31 + i)}})
		}
		renderer.Render(b)
		require.NoError(t, renderer.Flush())
		require.Contains(t, out.String(), "38;5;16")
		require.Contains(t, out.String(), "38;5;0")
		require.Contains(t, out.String(), "58;5;32")
		out.Reset()
		b.SetCell(1, 0, &uv.Cell{Content: "▀", Width: 1, Style: uv.Style{Fg: color.NRGBA{A: 255}}})
		renderer.Render(b)
		require.NoError(t, renderer.Flush())
		require.Contains(t, out.String(), "▀")
	}
}
