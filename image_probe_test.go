package terma

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/darrenburns/terma/layout"
	"github.com/stretchr/testify/require"
)

func TestImageProbeIndependentSubcontexts(t *testing.T) {
	for _, kind := range []string{"normal", "overflow", "scrolled"} {
		t.Run(kind, func(t *testing.T) {
			b := newImageBuffer(10, 2)
			root := NewRenderContext(b, 10, 2, nil, nil, BuildContext{}, nil)
			sourceA, sourceB := testImage(t, 2, 2), testImage(t, 4, 4)
			sub := func(x int) *RenderContext {
				switch kind {
				case "overflow":
					return root.OverflowSubContext(x, 0, 4, 2)
				case "scrolled":
					return root.ScrolledSubContext(x, 0, 4, 2, 0, 0)
				default:
					return root.SubContext(x, 0, 4, 2)
				}
			}
			left, right := sub(0), sub(5)
			left.DrawImage(0, 0, 4, 2, sourceA, ImageStretch)
			right.DrawImage(0, 0, 4, 2, sourceB, ImageStretch)
			require.Len(t, b.records, 2)
			a, z := b.cells[b.index(0, 0)].owner, b.cells[b.index(5, 0)].owner
			require.NotSame(t, a, z)
			require.Same(t, sourceA, a.source)
			require.Same(t, sourceB, z.source)
			require.True(t, b.owns(0, 0, a))
		})
	}
}

func TestImageProbeFailedKittyUploadSettles(t *testing.T) {
	now := time.Now()
	k := newKittyImages()
	src := testImage(t, 2, 2)
	u := k.upload(imageVariant{src, image.Rect(0, 0, 2, 2), 2, 2})
	u.placement(kittyPlacementKey{cols: 1, rows: 1, cw: 2, ch: 2})
	for i := 0; i < 2; i++ {
		require.NotEmpty(t, k.batch(now))
		k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id}, Payload: []byte("ERROR")})
	}
	require.True(t, u.failed)
	require.False(t, k.pending(), "permanent fallback must stop scheduling frames")
}

func TestImageProbePartialSixelFailureErased(t *testing.T) {
	terminal := &imageTestTerminal{ScreenBuffer: uv.NewScreenBuffer(8, 5), failAt: "\x1bP"}
	r := NewRenderer(terminal, 8, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Render(Image{Source: testImage(t, 4, 4), Fit: ImageStretch, Style: Style{Width: Cells(8), Height: Cells(5)}})
	s := &imageSession{detector: &imageDetector{requested: "sixel", started: true, done: true}, pointer: &pixelPointer{cellWidth: 8, cellHeight: 16}, kitty: newKittyImages(), sixel: newSixelImages()}
	require.Error(t, s.present(terminal, r, func() {}, 0))
	terminal.failAt = ""
	terminal.calls = nil
	r.Render(Text{Content: "removed"})
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.Contains(t, terminal.calls, "ERASE", "a failed write may have emitted partial image bytes")
}

func TestImageProbeLateReplies(t *testing.T) {
	now := time.Now()
	d := &imageDetector{requested: "auto", profile: colorprofile.ANSI256}
	d.query(now)
	require.Equal(t, "blocks", d.protocol(now.Add(time.Second), true))
	d.handle(uv.PrimaryDeviceAttributesEvent{62, 4})
	require.Equal(t, "sixel", d.protocol(now.Add(time.Second), true))
	d.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: imageProbeID}, Payload: []byte("OK")})
	d.handle(uv.TerminalVersionEvent{Name: "kitty(0.30.0)"})
	require.Equal(t, "kitty", d.protocol(now.Add(time.Second), true))
}

func TestImageProbeKittyAlphaPreserved(t *testing.T) {
	k := newKittyImages()
	src := testImage(t, 3, 3)
	u := k.upload(imageVariant{src, src.pixels.Bounds(), 6, 6})
	require.NotNil(t, u)
	require.NotEmpty(t, u.data)
	decodedBytes, err := base64.StdEncoding.DecodeString(u.data)
	require.NoError(t, err)
	decoded, err := png.Decode(bytes.NewReader(decodedBytes))
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 6, 6), decoded.Bounds())
	for y := 0; y < 6; y++ {
		for x := 0; x < 6; x++ {
			require.Equal(t, src.pixels.NRGBAAt(x/2, y/2), color.NRGBAModel.Convert(decoded.At(x, y)))
		}
	}
	require.Contains(t, k.batch(time.Now()), "f=100")
	require.False(t, strings.Contains(k.cleanup(), "d=A"), "cleanup must never delete other applications' images")
}

func TestImageProbeTransparentResourceOverTranslucentBackground(t *testing.T) {
	resource, err := NewImageResource(image.NewNRGBA(image.Rect(0, 0, 2, 2)))
	require.NoError(t, err)
	r := imageTestRenderer(2, 1)
	r.Render(Image{Source: resource, Fit: ImageStretch, Style: Style{Width: Cells(2), Height: Cells(1), BackgroundColor: Red.WithAlpha(.5)}})
	expected := imageBackground(Red.WithAlpha(.5).BlendOver(Black))
	for x := 0; x < 2; x++ {
		require.Equal(t, expected, r.images.cells[r.images.index(x, 0)].background)
		require.Equal(t, expected, r.images.CellAt(x, 0).Style.Fg)
		require.Equal(t, expected, r.images.CellAt(x, 0).Style.Bg)
	}
}

func TestImageProbeConstrainedAutoLayout(t *testing.T) {
	src := testImage(t, 800, 800)
	for _, tc := range []struct {
		name string
		c    layout.Constraints
		w, h int
	}{
		{"width", layout.Loose(10, 100), 10, 5},
		{"height", layout.Loose(100, 5), 10, 5},
		{"unbounded", layout.Unbounded(), 100, 50},
		{"tight width", layout.TightWidth(20, 100), 20, 10},
		{"tight height", layout.TightHeight(100, 8), 16, 8},
		{"both tight", layout.Tight(20, 20), 20, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := (Image{Source: src}).BuildLayoutNode(BuildContext{}).ComputeLayout(tc.c).Box
			require.Equal(t, tc.w, b.ContentWidth())
			require.Equal(t, tc.h, b.ContentHeight())
		})
	}
}

func TestImageProbeRemovingRootClearsImages(t *testing.T) {
	r := imageTestRenderer(8, 3)
	r.Render(Image{Source: testImage(t, 2, 2), Fit: ImageStretch, Style: Style{Width: Cells(8), Height: Cells(3)}})
	require.NotEmpty(t, r.images.records)
	r.Render(nil)
	require.Empty(t, r.images.records)
	require.NotContains(t, r.ScreenText(), "▀")
	for y := 0; y < 3; y++ {
		for x := 0; x < 8; x++ {
			if c := r.presentation.CellAt(x, y); c != nil {
				require.Empty(t, strings.TrimSpace(c.Content))
			}
		}
	}
}

func TestImageProbeFirstSwitchSettlesDispatch(t *testing.T) {
	r := imageTestRenderer(20, 5)
	label := NewSignal("x")
	scheduled := false
	root := headlessDispatchPlainRoot{build: func() Widget {
		return Column{Children: []Widget{
			Image{Source: testImage(t, 2, 2), Style: Style{Width: Cells(4), Height: Cells(1)}},
			headlessDispatchLeaf{Text: Text{Content: label.Get()}, onLayout: func() {
				if r.images == nil || scheduled {
					return
				}
				scheduled = true
				Dispatch(func() { label.Set("settled") })
			}},
		}}
	}}
	w, h := r.RenderWithSize(root)
	require.True(t, scheduled)
	require.Contains(t, r.ScreenText(), "settled")
	require.Equal(t, 7, w)
	require.Equal(t, 2, h)
}

type imageProbeComposite struct {
	sources [2]*ImageResource
	step    Signal[int]
}

func (w *imageProbeComposite) Build(BuildContext) Widget { return w }
func (w *imageProbeComposite) GetStyle() Style           { return Style{Width: Cells(20), Height: Cells(6)} }
func (w *imageProbeComposite) Render(ctx *RenderContext) {
	step := w.step.Get()
	x := step%19 - 6
	left := ctx.SubContext(x, 0, 9, 5)
	right := ctx.ScrolledSubContext(7, 1, 11, 5, step%4, step%3)
	if step%7 != 0 {
		left.DrawImage(0, 0, 9, 5, w.sources[step%2], ImageStretch)
	}
	right.DrawImage(0, 0, 11, 5, w.sources[1-step%2], ImageFit(step%3))
	ctx.DrawText(step%18, 2, "界")
	if step%5 == 0 {
		ctx.DrawBackdrop(3, 1, 7, 4, Black.WithAlpha(.5))
	}
}
func TestImageProbeMovingOverlappingRetained(t *testing.T) {
	sources := [2]*ImageResource{testImage(t, 16, 8), testImage(t, 24, 48)}
	seq := newReactivitySequence(t, 22, 8, func() *imageProbeComposite { return &imageProbeComposite{sources: sources, step: NewSignal(0)} })
	for step := 0; step < 42; step++ {
		name := fmt.Sprintf("step-%d", step)
		seq.frame(name, func(w *imageProbeComposite) { w.step.Set(step) })
		a, e := seq.actual.renderer.images, seq.expected.renderer.images
		require.Equal(t, imageRecordSignature(e), imageRecordSignature(a), name)
		type nativeRegion struct {
			cursor string
			pixels *image.NRGBA
		}
		native := func(b *imageBuffer) []nativeRegion {
			out := uv.NewBuffer(22, 8)
			copyImageCells(out, b, 22, 8)
			payload := newSixelImages().output(b, out, 2, 3, 0)
			var regions []nativeRegion
			for _, part := range strings.Split(payload, "\x1b\\") {
				if part == "" {
					continue
				}
				cursor, data, ok := strings.Cut(part, "\x1bP0;1q")
				require.True(t, ok)
				decoded, err := (&sixel.Decoder{}).Decode(strings.NewReader(data))
				require.NoError(t, err)
				pixels := image.NewNRGBA(decoded.Bounds())
				draw.Draw(pixels, pixels.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
				regions = append(regions, nativeRegion{cursor, pixels})
			}
			return regions
		}
		require.Equal(t, native(e), native(a), name)
	}
}

func TestImageProbeFlexAndInsets(t *testing.T) {
	source := testImage(t, 80, 80)
	for _, tc := range []struct {
		name        string
		style       Style
		constraints layout.Constraints
		w, h        int
	}{
		{"flex width auto height", Style{Width: Flex(1), Height: Auto}, layout.Loose(40, 30), 40, 20},
		{"auto width flex height", Style{Width: Auto, Height: Flex(1)}, layout.Loose(80, 12), 24, 12},
		{"both flex", Style{Width: Flex(1), Height: Flex(1)}, layout.Loose(40, 30), 40, 30},
		{"flex padding", Style{Width: Flex(1), Height: Auto, Padding: EdgeInsetsAll(2)}, layout.Loose(44, 50), 44, 24},
		{"flex border", Style{Width: Flex(1), Height: Flex(1), Padding: EdgeInsetsAll(2), Border: RoundedBorder(White)}, layout.Loose(44, 30), 44, 30},
		{"unbounded padded", Style{Width: Flex(1), Height: Flex(1), Padding: EdgeInsetsAll(2)}, layout.Loose(math.MaxInt32, math.MaxInt32), 14, 9},
		{"scroll height padded", Style{Width: Flex(1), Height: Flex(1), Padding: EdgeInsetsAll(2)}, layout.Loose(44, math.MaxInt32), 44, 24},
		{"fully unbounded padded", Style{Width: Flex(1), Height: Flex(1), Padding: EdgeInsetsAll(2)}, layout.Unbounded(), 14, 9},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := (Image{Source: source, Style: tc.style}).BuildLayoutNode(BuildContext{}).ComputeLayout(tc.constraints).Box
			require.Equal(t, tc.w, b.Width)
			require.Equal(t, tc.h, b.Height)
		})
	}
}
func TestImageProbeParentResolvedPercent(t *testing.T) {
	source := testImage(t, 80, 80)
	for _, tc := range []struct {
		name string
		root Widget
		dest Rect
	}{
		{"row width", Row{Style: Style{Width: Cells(40), Height: Cells(20)}, Children: []Widget{Image{Source: source, Fit: ImageStretch, Style: Style{Width: Percent(50), Height: Cells(4)}}}}, Rect{Width: 20, Height: 4}},
		{"column height", Column{Style: Style{Width: Cells(40), Height: Cells(20)}, Children: []Widget{Image{Source: source, Fit: ImageStretch, Style: Style{Width: Cells(8), Height: Percent(50)}}}}, Rect{Width: 8, Height: 10}},
		{"stack both", Stack{Style: Style{Width: Cells(40), Height: Cells(20)}, Children: []Widget{Image{Source: source, Fit: ImageStretch, Style: Style{Width: Percent(50), Height: Percent(50)}}}}, Rect{Width: 20, Height: 10}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := imageTestRenderer(40, 20)
			r.Render(tc.root)
			require.Len(t, r.images.records, 1)
			for _, record := range r.images.records {
				require.Equal(t, tc.dest, record.mapping.dest)
			}
		})
	}
}

func TestImageProbePaddedFlexScrollable(t *testing.T) {
	state := NewScrollState()
	root := Scrollable{State: state, Width: Cells(30), Height: Cells(10), Child: Image{Source: testImage(t, 80, 80), Fit: ImageStretch, Style: Style{Width: Flex(1), Height: Flex(1), Padding: EdgeInsetsAll(2)}}}
	r := imageTestRenderer(30, 10)
	r.Render(root)
	require.Equal(t, 10, state.viewportHeight)
	require.Equal(t, 17, state.contentHeight)
	require.Len(t, r.images.records, 1)
	for _, record := range r.images.records {
		require.Equal(t, 13, record.mapping.dest.Height)
	}
}
