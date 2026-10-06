package terma

import (
	"image"
	"image/color"
	"runtime"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi/kitty"
	"github.com/stretchr/testify/require"
)

func TestImageKittyResetEndsChunkedUpload(t *testing.T) {
	k := newKittyImages()
	u := k.upload(newImageVariant(testImage(t, 64, 64), image.Rect(0, 0, 64, 64), 64, 64))
	u.data = strings.Repeat("A", kittyBatchBytes+kittyChunkBytes)
	k.batch(time.Now())
	require.NotNil(t, k.active, "upload still has chunks to send")

	out := k.reset()
	// The open transfer is ended before any other graphics command: the
	// terminal would take the deletes as more of its data.
	require.True(t, strings.HasPrefix(out, "\x1b_Gm=0,q=2\x1b\\"), "got %q", out)
	require.Contains(t, out, "a=d,d=I,i=1,q=2")
	require.Nil(t, k.active)
	require.Empty(t, k.uploads)

	// The next image gets a new id, so a late reply about the old one isn't
	// taken for it.
	next := k.upload(newImageVariant(testImage(t, 2, 2), image.Rect(0, 0, 2, 2), 2, 2))
	require.Equal(t, 2, next.id)
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: 1}, Payload: []byte("EBADPNG")})
	require.False(t, next.failed)
}

func TestImageKittyCleanupWithoutTransfer(t *testing.T) {
	k := newKittyImages()
	k.upload(newImageVariant(testImage(t, 2, 2), image.Rect(0, 0, 2, 2), 2, 2))
	out := k.cleanup()
	require.NotContains(t, out, "m=0")
	require.Equal(t, "\x1b_Ga=d,d=I,i=1,q=2\x1b\\", out)
}

func TestImageKittyReleasesSentData(t *testing.T) {
	k := newKittyImages()
	src := testImage(t, 8, 8)
	defer runtime.KeepAlive(src) // Caches refer to images weakly.
	u := k.upload(newImageVariant(src, image.Rect(0, 0, 8, 8), 8, 8))
	require.NotEmpty(t, u.data)
	for !u.pending {
		k.batch(time.Now())
	}
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id}, Payload: []byte("OK")})
	require.Empty(t, u.data, "the terminal has it now")
	require.Zero(t, k.bytes)

	// Lost by the terminal later: encoded again to be resent.
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id, PlacementID: 9}, Payload: []byte("ENOENT")})
	p := u.placement(kittyPlacementKey{cols: 1, rows: 1, cw: 8, ch: 8})
	p.pending = true
	k.handle(uv.KittyGraphicsEvent{Options: kitty.Options{ID: u.id, PlacementID: p.id}, Payload: []byte("ENOENT")})
	require.NotEmpty(t, u.data)
	require.False(t, u.failed)
}

func TestImageKittyBoundsHiddenImages(t *testing.T) {
	k := newKittyImages()
	src := testImage(t, 2, 2)
	var uploads []*kittyUpload
	for i := 1; i <= maxHiddenKittyImages+4; i++ {
		v := newImageVariant(src, image.Rect(0, 0, 2, 2), i, 1)
		k.begin(map[imageVariant]bool{v: true})
		uploads = append(uploads, k.upload(v))
		k.finish()
	}
	k.begin(nil)
	k.finish()
	k.finish()
	require.Len(t, k.uploads, maxHiddenKittyImages)
	for _, u := range uploads[:4] {
		require.Nil(t, k.ids[u.id], "the images longest off screen go first")
	}
	require.Len(t, k.garbage, 4)
}

func TestImageCachesDontKeepDroppedImages(t *testing.T) {
	k := newKittyImages()
	s := newSixelImages()
	func() {
		src := testImage(t, 4, 4)
		k.upload(newImageVariant(src, image.Rect(0, 0, 4, 4), 4, 4))
		require.NotNil(t, s.prepare(sixelKey{newImageVariant(src, image.Rect(0, 0, 4, 4), 4, 4), color.NRGBA{A: 255}}))
	}()
	runtime.GC()
	runtime.GC()
	k.begin(nil)
	require.Empty(t, k.uploads, "its source is gone")
	require.Contains(t, k.cleanup(), "a=d,d=I,i=1,q=2")
	s.prune()
	require.Empty(t, s.prepared)
	require.Zero(t, s.bytes)
}

func TestImageWorkRunsOffTheEventLoop(t *testing.T) {
	withRunningApp(t)
	k := newKittyImages()
	k.worker.async = true
	s := newSixelImages()
	s.worker.async = true
	src := testImage(t, 32, 32)

	u := k.upload(newImageVariant(src, image.Rect(0, 0, 32, 32), 32, 32))
	require.True(t, u.encoding)
	require.Empty(t, k.batch(time.Now()), "nothing to send until encoded")
	key := sixelKey{newImageVariant(src, image.Rect(0, 0, 32, 32), 32, 32), color.NRGBA{A: 255}}
	require.Nil(t, s.prepare(key), "not ready yet")

	// The results arrive through Dispatch, on the event loop.
	deadline := time.Now().Add(5 * time.Second)
	for (u.encoding || s.prepared[key].pixels == nil) && time.Now().Before(deadline) {
		drainPendingDispatches()
		time.Sleep(time.Millisecond)
	}
	require.False(t, u.encoding)
	require.NotEmpty(t, u.data)
	require.Contains(t, k.batch(time.Now()), "a=t,t=d,f=100,i=1")
	require.NotNil(t, s.prepare(key))
}

func TestImageSixelDrawnOnceWhileUnchanged(t *testing.T) {
	terminal := &imageTestTerminal{ScreenBuffer: uv.NewScreenBuffer(12, 6)}
	r := NewRenderer(terminal, 12, 6, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	label := NewSignal("a")
	root := Column{Children: []Widget{
		Image{Source: testImage(t, 4, 4), Fit: ImageStretch, Style: Style{Width: Cells(4), Height: Cells(3)}},
		Text{Content: label.Get()},
	}}
	r.Render(root)
	s := &imageSession{detector: &imageDetector{requested: "sixel", started: true, done: true}, pointer: &pixelPointer{cellWidth: 8, cellHeight: 16}, kitty: newKittyImages(), sixel: newSixelImages()}
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.Contains(t, terminal.output.String(), "\x1bP")

	// A frame that leaves the image alone neither clears the screen nor draws
	// the image again.
	terminal.calls = nil
	root.Children[1] = Text{Content: "b"}
	r.Render(root)
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.NotContains(t, terminal.calls, "ERASE")
	require.NotContains(t, strings.Join(terminal.calls, ""), "\x1bP")

	// Another image appears: only it is drawn.
	root.Children = append(root.Children, Image{Source: testImage(t, 2, 2), Fit: ImageStretch, Style: Style{Width: Cells(2), Height: Cells(1)}})
	r.Render(root)
	terminal.calls = nil
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.NotContains(t, terminal.calls, "ERASE")
	require.Equal(t, 1, strings.Count(strings.Join(terminal.calls, ""), "\x1bP"))

	// One goes: the screen is cleared and the other drawn again.
	root.Children = root.Children[:2]
	r.Render(root)
	terminal.calls = nil
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.Contains(t, terminal.calls, "ERASE")
	require.Equal(t, 1, strings.Count(strings.Join(terminal.calls, ""), "\x1bP"))

	// After the app clears the screen (a resize), it is drawn again.
	s.erased()
	terminal.calls = nil
	require.NoError(t, s.present(terminal, r, func() {}, 0))
	require.NotContains(t, terminal.calls, "ERASE")
	require.Equal(t, 1, strings.Count(strings.Join(terminal.calls, ""), "\x1bP"))
}

func TestImageSixelAvoidsLastRow(t *testing.T) {
	b := newImageBuffer(6, 4)
	ctx := NewRenderContext(b, 6, 4, nil, nil, BuildContext{}, nil)
	ctx.DrawImage(0, 0, 6, 4, testImage(t, 6, 8), ImageStretch)
	out := uv.NewBuffer(6, 4)
	copyCells(out, b, 6, 4)
	regions := sixelRegions(b, out, 0)
	require.NotEmpty(t, regions)
	for _, region := range regions {
		require.LessOrEqual(t, region.rect.Y+region.rect.Height, 3, "the last row stays half blocks")
	}
	require.Equal(t, "▀", out.CellAt(0, 3).Content)
}

func TestImageTransparencyShowsTerminalBackground(t *testing.T) {
	p := image.NewNRGBA(image.Rect(0, 0, 3, 2))
	red := color.NRGBA{255, 0, 0, 255}
	p.SetNRGBA(0, 0, red) // Column 0: opaque over transparent.
	p.SetNRGBA(1, 1, red) // Column 1: transparent over opaque.
	// Column 2: fully transparent.
	src, err := NewImageResource(p)
	require.NoError(t, err)

	b := newImageBuffer(3, 1)
	ctx := NewRenderContext(b, 3, 1, nil, nil, BuildContext{}, nil)
	ctx.DrawImage(0, 0, 3, 1, src, ImageStretch)
	require.Equal(t, uv.Cell{Content: "▀", Width: 1, Style: uv.Style{Fg: red}}, *b.CellAt(0, 0))
	require.Equal(t, uv.Cell{Content: "▄", Width: 1, Style: uv.Style{Fg: red}}, *b.CellAt(1, 0))
	require.Equal(t, uv.Cell{Content: " ", Width: 1}, *b.CellAt(2, 0))

	// Kitty placeholders leave the background to the terminal.
	k := newKittyImages()
	out := uv.NewBuffer(3, 1)
	k.paint(out, b, 1, 2)
	for _, u := range k.uploads {
		u.ready = true
		for _, p := range u.placements {
			p.ready = true
		}
	}
	k.paint(out, b, 1, 2)
	require.Nil(t, out.CellAt(0, 0).Style.Bg)

	// Sixel leaves transparent pixels undrawn.
	s := newSixelImages()
	prepared := s.prepare(sixelKey{variantFor(b.records[imageRecordKey{}], 1, 2), color.NRGBA{}})
	require.NotNil(t, prepared)
	require.Zero(t, prepared.pixels.ColorIndexAt(0, 1))
	require.Zero(t, prepared.pixels.ColorIndexAt(2, 0))
	require.NotZero(t, prepared.pixels.ColorIndexAt(0, 0))

	// Over a background of the app's own, the image is blended over it.
	withBg := newImageBuffer(3, 1)
	bgCtx := NewRenderContext(withBg, 3, 1, nil, nil, BuildContext{}, nil)
	bgCtx.inheritedBgAt = func(int, int) Color { return RGB(0, 0, 255) }
	bgCtx.DrawImage(0, 0, 3, 1, src, ImageStretch)
	require.Equal(t, "▀", withBg.CellAt(2, 0).Content)
	require.NotNil(t, withBg.CellAt(2, 0).Style.Bg)
}
