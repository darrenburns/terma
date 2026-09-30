package terma

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/colorprofile"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/charmbracelet/x/ansi/sixel"
	"github.com/stretchr/testify/require"
)

func TestImageDetection(t *testing.T) {
	now := time.Now()
	d := &imageDetector{requested: "auto", profile: colorprofile.ANSI256}
	q := d.query(now)
	require.True(t, strings.HasSuffix(q, "\x1b[c"))
	require.Less(t, strings.Index(q, "a=q"), strings.Index(q, "\x1b[c"))
	require.Contains(t, q, requestCellSize)
	require.Contains(t, q, ansi.RequestModeSynchronizedOutput)
	require.Equal(t, "blocks", d.protocol(now, true))
	d.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: imageProbeID}, Payload: []byte("OK")})
	d.handle(uv.TerminalVersionEvent{Name: "kitty(0.28.0)"})
	d.handle(uv.PrimaryDeviceAttributesEvent{62, 4})
	require.Equal(t, "kitty", d.protocol(now, true))
	require.Equal(t, "blocks", d.protocol(now, false))
	d.profile = colorprofile.ANSI
	require.Equal(t, "sixel", d.protocol(now, true))
	d.profile = colorprofile.TrueColor
	d.knownPlaceholders = false
	require.Equal(t, "sixel", d.protocol(now, true))
	d.sixel = false
	require.Equal(t, "blocks", d.protocol(now, true))
	d.multiplexer = true
	d.requested = "kitty"
	require.Equal(t, "kitty", d.protocol(now, true))
	d.requested = "auto"
	require.Equal(t, "blocks", d.protocol(now, true))
	d = &imageDetector{requested: "auto", profile: colorprofile.TrueColor}
	d.query(now)
	require.Equal(t, "blocks", d.protocol(now.Add(time.Second), true))
	require.True(t, d.done)
	for _, name := range []string{"kitty(0.27.1)", "not-ghostty 1.2.0", "ghosttyish 1.2.0", "unknown"} {
		d.knownPlaceholders = false
		d.handle(uv.TerminalVersionEvent{Name: name})
		require.False(t, d.knownPlaceholders, name)
	}
}
func TestImageGeometryPrecedence(t *testing.T) {
	p := &pixelPointer{}
	p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
	p.handle(uv.WindowPixelSizeEvent{Width: 800, Height: 480})
	w, h, ok := p.cellSize()
	require.True(t, ok)
	require.Equal(t, float64(10), w)
	require.Equal(t, float64(20), h)
	p.handle(uv.CellSizeEvent{Width: 12, Height: 25})
	p.handle(uv.WindowPixelSizeEvent{Width: 1600, Height: 960})
	// Images keep the report until the terminal answers again.
	iw, ih, _ := p.imageCellSize()
	require.Equal(t, 12, iw)
	require.Equal(t, 25, ih)
	t.Setenv("TERMA_IMAGE_CELL_SIZE", "9x18")
	p = newPixelPointer()
	p.handle(uv.CellSizeEvent{Width: 12, Height: 25})
	w, h, _ = p.cellSize()
	require.Equal(t, float64(9), w)
	require.Equal(t, float64(18), h)
	iw, ih, _ = p.imageCellSize()
	require.Equal(t, 9, iw)
	require.Equal(t, 18, ih)
	require.Empty(t, p.handle(uv.WindowSizeEvent{Width: 40, Height: 12}), "nothing to ask with the size given")
	for _, s := range []string{"0x10", "2x0", "-1x5", "bad", "8x16x4"} {
		t.Setenv("TERMA_IMAGE_CELL_SIZE", s)
		w, h := explicitImageCellSize()
		require.Zero(t, w)
		require.Zero(t, h)
	}
}
func TestImageKittyTransfersAndRetry(t *testing.T) {
	now := time.Now()
	k := newKittyImages()
	// Held for the test: caches refer to images weakly.
	src := testImage(t, 128, 128)
	defer runtime.KeepAlive(src)
	u := k.upload(newImageVariant(src, image.Rect(0, 0, 128, 128), 128, 128))
	require.NotNil(t, u)
	// A deliberately long encoded payload exposes multipart batching without
	// depending on PNG compressibility.
	// Longer than one frame's budget, so it takes two.
	u.data = strings.Repeat("A", kittyBatchBytes+3*kittyChunkBytes)
	p := u.placement(kittyPlacementKey{cols: 16, rows: 8, cw: 8, ch: 16})
	first := k.batch(now)
	require.Equal(t, kittyBatchBytes/kittyChunkBytes, strings.Count(first, "\x1b_G"))
	require.NotContains(t, first, "a=p")
	require.NotNil(t, k.active)
	u2 := k.upload(newImageVariant(testImage(t, 2, 2), image.Rect(0, 0, 2, 2), 2, 2))
	u2.placement(kittyPlacementKey{cols: 1, rows: 1, cw: 2, ch: 2})
	second := k.batch(now)
	require.Contains(t, second, "m=0")
	require.Less(t, strings.Index(second, "m=0"), strings.Index(second, "a=t"))
	require.True(t, u.pending)
	require.False(t, p.ready)
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id}, Payload: []byte("OK")})
	require.True(t, u.ready)
	require.Contains(t, k.batch(now), "a=p,U=1")
	require.True(t, p.pending)
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id, PlacementID: p.id}, Payload: []byte("ENOENT: missing")})
	require.False(t, u.ready)
	require.False(t, u.failed)
	for k.active != nil || !u.pending {
		k.batch(now)
	}
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id}, Payload: []byte("ERROR")})
	require.True(t, u.failed)
}
func TestImageKittyLimitsAndMovement(t *testing.T) {
	src := testImage(t, 2, 2)
	k := newKittyImages()
	for i := 1; i <= 255; i++ {
		u := k.upload(newImageVariant(src, image.Rect(0, 0, 2, 2), i, 1))
		require.NotNil(t, u)
		require.Equal(t, i, u.id)
	}
	require.Nil(t, k.upload(newImageVariant(src, image.Rect(0, 0, 2, 2), 256, 1)))
	k.finish()
	k.begin(map[imageVariant]bool{})
	require.Nil(t, k.upload(newImageVariant(src, image.Rect(0, 0, 2, 2), 256, 1)), "previous presentation still references IDs")
	k.finish()
	u := k.upload(newImageVariant(src, image.Rect(0, 0, 2, 2), 256, 1))
	require.NotNil(t, u)
	require.Equal(t, 1, u.id)
	require.Contains(t, k.batch(time.Now()), "a=d,d=I,i=1")
	k = newKittyImages()
	b := newImageBuffer(8, 6)
	ctx := NewRenderContext(b, 8, 6, nil, nil, BuildContext{}, nil)
	ctx.DrawImage(0, 0, 4, 4, src, ImageStretch)
	out := uv.NewBuffer(8, 6)
	k.paint(out, b, 8, 16)
	require.Len(t, k.uploads, 1)
	b.Clear()
	ctx.DrawImage(2, 1, 4, 4, src, ImageStretch)
	k.paint(out, b, 8, 16)
	require.Len(t, k.uploads, 1, "movement must reuse upload")
	for _, u := range k.uploads {
		require.Len(t, u.placements, 1)
	}
	b = newImageBuffer(len(imageDiacritics)+1, 1)
	ctx = NewRenderContext(b, b.buffer.Width(), 1, nil, nil, BuildContext{}, nil)
	ctx.DrawImage(0, 0, b.buffer.Width(), 1, src, ImageStretch)
	k = newKittyImages()
	out = uv.NewBuffer(b.buffer.Width(), 1)
	k.paint(out, b, 1, 1)
	for _, u := range k.uploads {
		require.Len(t, u.placements, 2)
		u.ready = true
		for _, p := range u.placements {
			p.ready = true
		}
	}
	k.paint(out, b, 1, 1)
	require.Equal(t, string([]rune{0x10eeee, imageDiacritics[0], imageDiacritics[0]}), out.CellAt(len(imageDiacritics), 0).Content)
}
func TestImageSixelPaletteCropAndBands(t *testing.T) {
	s := newSixelImages()
	src := testImage(t, 64, 64)
	key := sixelKey{newImageVariant(src, image.Rect(0, 0, 64, 64), 31, 19), color.NRGBA{20, 30, 40, 255}}
	p := s.prepare(key)
	require.LessOrEqual(t, len(p.pixels.Palette), 256)
	_, _, _, alpha := p.pixels.Palette[0].RGBA()
	require.Zero(t, alpha)
	crop := sixelCrop(p.pixels, image.Rect(3, 5, 14, 18), func(x, y int) bool { return x == 7 })
	require.Equal(t, image.Rect(0, 0, 11, 13), crop.Bounds())
	require.Zero(t, crop.ColorIndexAt(4, 0))
	require.Equal(t, p.pixels.ColorIndexAt(3, 5), crop.ColorIndexAt(0, 0))
	payload, err := encodeSixel(crop)
	require.NoError(t, err)
	require.Contains(t, payload, "\x1bP0;1q\"1;1;11;13")
	require.NotContains(t, payload, "-\x1b\\")
	decoded, err := (&sixel.Decoder{}).Decode(strings.NewReader(strings.TrimPrefix(payload, "\x1bP0;1q")))
	require.NoError(t, err)
	require.Equal(t, crop.Bounds(), decoded.Bounds())
	_, _, _, alpha = decoded.At(4, 0).RGBA()
	require.Zero(t, alpha)
	s.prepare(key)
	s.payload(p, image.Rect(3, 5, 14, 18))
	s.payload(p, image.Rect(3, 5, 14, 18))
	require.Equal(t, 1, s.preparations)
	require.Len(t, p.crops, 1)
	for i := 0; i < 60; i++ {
		s.prepare(sixelKey{newImageVariant(src, image.Rect(0, 0, 64, 64), i+1, 3), color.NRGBA{A: 255}})
	}
	require.LessOrEqual(t, len(s.prepared), 32)
	require.LessOrEqual(t, s.bytes, maxImageCacheBytes)
}
func TestImageSixelOcclusion(t *testing.T) {
	b := newImageBuffer(8, 5)
	ctx := NewRenderContext(b, 8, 5, nil, nil, BuildContext{}, nil)
	ctx.DrawImage(0, 0, 8, 5, testImage(t, 8, 10), ImageStretch)
	b.SetCell(3, 2, &uv.Cell{Content: "x", Width: 1})
	out := uv.NewBuffer(8, 5)
	copyImageCells(out, b, 8, 5)
	regions := sixelRegions(b, out, 1)
	for _, r := range regions {
		require.False(t, r.rect.Contains(3, 2))
		require.GreaterOrEqual(t, r.rect.Y, 1)
	}
	s := newSixelImages()
	payload, _ := s.output(s.draws(b, out, 2, 3, 1))
	require.NotEmpty(t, payload)
	require.NotContains(t, payload, "-\x1b\\")
	require.Equal(t, 1, s.preparations)
}

type imageTestTerminal struct {
	uv.ScreenBuffer
	output bytes.Buffer
	failAt string
	calls  []string
}

func (t *imageTestTerminal) WriteString(s string) (int, error) {
	t.calls = append(t.calls, s)
	if t.failAt != "" && strings.Contains(s, t.failAt) {
		return 0, errors.New("write failed")
	}
	return t.output.WriteString(s)
}
func (t *imageTestTerminal) Flush() error   { t.calls = append(t.calls, "FLUSH"); return nil }
func (t *imageTestTerminal) Display() error { t.calls = append(t.calls, "DISPLAY"); return t.Flush() }
func (t *imageTestTerminal) Erase()         { t.calls = append(t.calls, "ERASE"); t.Clear() }
func TestImagePresentationCleanup(t *testing.T) {
	for _, fail := range []string{"", "\x1bP", "\x1b[?2026h"} {
		t.Run(fail, func(t *testing.T) {
			terminal := &imageTestTerminal{ScreenBuffer: uv.NewScreenBuffer(8, 5), failAt: fail}
			r := NewRenderer(terminal, 8, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
			r.Render(Image{Source: testImage(t, 4, 4), Fit: ImageStretch, Style: Style{Width: Cells(8), Height: Cells(5)}})
			s := &imageSession{detector: &imageDetector{requested: "sixel", started: true, done: true, syncOutput: true}, pointer: &pixelPointer{cellWidth: 8, cellHeight: 16}, kitty: newKittyImages(), sixel: newSixelImages()}
			err := s.present(terminal, r, func() {}, 0)
			if fail != "" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			joined := strings.Join(terminal.calls, "|")
			require.Contains(t, joined, ansi.ResetModeSynchronizedOutput)
			require.True(t, strings.HasSuffix(joined, ansi.ResetModeSynchronizedOutput+"|FLUSH"))
			if fail == "" {
				require.Less(t, strings.Index(joined, "DISPLAY"), strings.Index(joined, "\x1bP"))
				require.Contains(t, joined, "\x1b8")
				terminal.calls = nil
				r.Render(Text{Content: "removed"})
				require.NoError(t, s.present(terminal, r, func() {}, 0))
				require.Contains(t, terminal.calls, "ERASE")
				require.Empty(t, s.sixelShown)
			}
		})
	}
}
func TestImagePresentationDoesNotHardwareScroll(t *testing.T) {
	var output bytes.Buffer
	renderer := uv.NewTerminalRenderer(&output, nil)
	renderer.SetFullscreen(true)
	// Terminal's default keeps SetScrollOptim false. Exercise a whole-screen
	// shift, which would otherwise be a strong candidate for hardware scrolling.
	b := uv.NewBuffer(30, 12)
	draw := func(offset int) {
		for y := 0; y < 12; y++ {
			for x := 0; x < 30; x++ {
				b.SetCell(x, y, &uv.Cell{Content: string(rune('A' + y + offset)), Width: 1})
			}
		}
		renderer.Render(b)
		require.NoError(t, renderer.Flush())
	}
	draw(0)
	output.Reset()
	draw(1)
	for _, bad := range []string{"\x1b[S", "\x1b[T", "\x1b[1S", "\x1b[1T", "\x1b[M", "\x1b[L"} {
		require.NotContains(t, output.String(), bad)
	}
}

func TestImageDetectionKeepsOrdinaryInput(t *testing.T) {
	events := decodeAll(t, "a\x1b_Gi=2147483647;OK\x1b\\b\x1b[6;16;8t\x1b[?62;4c")
	d := &imageDetector{requested: "auto", profile: colorprofile.TrueColor}
	d.query(time.Now())
	p := &pixelPointer{}
	var keys strings.Builder
	for _, event := range events {
		d.handle(event)
		p.handle(event)
		if key, ok := event.(uv.KeyPressEvent); ok {
			keys.WriteString(key.Text)
		}
	}
	require.Equal(t, "ab", keys.String())
	require.True(t, d.graphics)
	require.True(t, d.sixel)
	require.True(t, d.done)
	w, h, ok := p.cellSize()
	require.True(t, ok)
	require.Equal(t, 8.0, w)
	require.Equal(t, 16.0, h)
}

func TestImageResizeResetAndDebugOverlay(t *testing.T) {
	terminal := &imageTestTerminal{ScreenBuffer: uv.NewScreenBuffer(8, 5)}
	root := Image{Source: testImage(t, 4, 4), Fit: ImageStretch, Style: Style{Width: Flex(1), Height: Flex(1)}}
	r := NewRenderer(terminal, 8, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Render(root)
	session := &imageSession{detector: &imageDetector{requested: "sixel", started: true, done: true}, pointer: &pixelPointer{cellWidth: 8, cellHeight: 16}, kitty: newKittyImages(), sixel: newSixelImages()}
	debug := func() { terminal.SetCell(0, 0, &uv.Cell{Content: "D", Width: 1}) }
	require.NoError(t, session.present(terminal, r, debug, 1))
	require.Equal(t, "D", terminal.CellAt(0, 0).Content)
	require.NotContains(t, terminal.output.String(), "\x1b[1;1H\x1bP", "Sixel must not paint over debug rows")
	session.kitty.upload(newImageVariant(root.Source, image.Rect(0, 0, 4, 4), 4, 4))
	session.reset(terminal)
	require.Empty(t, session.kitty.uploads)
	require.Contains(t, terminal.output.String(), "a=d,d=I,i=1")
	r.Resize(12, 7)
	r.Update(root)
	require.Equal(t, uv.Rect(0, 0, 12, 7), r.images.Bounds())
	require.Equal(t, uv.Rect(0, 0, 12, 7), terminal.Bounds())
	terminal.Erase()
	require.NoError(t, session.present(terminal, r, debug, 1))
	require.Equal(t, "▀", terminal.CellAt(11, 6).Content)
	require.Equal(t, "D", terminal.CellAt(0, 0).Content)
}

func TestImageGeometryRefreshWithoutPixelMouse(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		p := &pixelPointer{disabled: disabled, imageGeometry: true}
		p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
		p.handle(uv.WindowPixelSizeEvent{Width: 804, Height: 484})
		p.handle(uv.CellSizeEvent{Width: 10, Height: 20})
		p.handle(uv.CellSizeEvent{Width: 8192, Height: 25}) // Not believable.
		w, h, _ := p.imageCellSize()
		require.Equal(t, [2]int{10, 20}, [2]int{w, h})
		// The font grows to 12x25 cells in the same window: asked again, even
		// with pixel mouse reporting switched off.
		require.Equal(t, requestCellSize, p.handle(uv.WindowSizeEvent{Width: 66, Height: 19}))
		w, h, _ = p.imageCellSize()
		require.Equal(t, [2]int{10, 20}, [2]int{w, h}, "kept until answered")
		p.handle(uv.CellSizeEvent{Width: 12, Height: 25})
		w, h, _ = p.imageCellSize()
		require.Equal(t, [2]int{12, 25}, [2]int{w, h})
		require.False(t, p.enabled)
	}
}

func TestImageGhosttyAutoDetection(t *testing.T) {
	for _, name := range []string{"ghostty 1.0.0", "Ghostty 1.2.0", "ghostty 1.3.1", "ghostty 1.4.0-dev+abc", "Ghostty(1.3.1)"} {
		t.Run(name, func(t *testing.T) {
			for _, profile := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.TrueColor} {
				now := time.Now()
				d := &imageDetector{requested: "auto", profile: profile}
				d.query(now)
				p := &pixelPointer{}
				// Decode the same in-band identity, geometry and DA1 replies
				// handled by Run, then deliver the graphics acknowledgement late.
				for _, ev := range decodeAll(t, "\x1bP>|"+name+"\x1b\\\x1b[6;16;8t\x1b[?62c") {
					d.handle(ev)
					p.handle(ev)
				}
				_, _, geometry := p.cellSize()
				require.True(t, geometry)
				require.Equal(t, "blocks", d.protocol(now, geometry), "identity alone is insufficient")
				for _, ev := range decodeAll(t, "\x1b_Gi=2147483647;OK\x1b\\") {
					d.handle(ev)
				}
				require.Equal(t, "kitty", d.protocol(now, geometry))
				require.Equal(t, "blocks", d.protocol(now, false))
				d.profile = colorprofile.ANSI
				require.Equal(t, "blocks", d.protocol(now, geometry))
				d.profile = profile
				d.multiplexer = true
				require.Equal(t, "blocks", d.protocol(now, geometry))
				d.multiplexer = false
				d.requested = "blocks"
				require.Equal(t, "blocks", d.protocol(now, geometry))
			}
		})
	}
}
