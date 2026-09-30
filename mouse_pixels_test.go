package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pixelModeReport(value ansi.ModeSetting) uv.ModeReportEvent {
	return uv.ModeReportEvent{Mode: ansi.ModeMouseExtSgrPixel, Value: value}
}

func TestPixelPointer_EnablesOnceSupportedAndSized(t *testing.T) {
	for name, events := range map[string][]uv.Event{
		"report first":    {pixelModeReport(ansi.ModeReset), uv.CellSizeEvent{Width: 10, Height: 20}},
		"cell size first": {uv.CellSizeEvent{Width: 10, Height: 20}, pixelModeReport(ansi.ModeReset)},
	} {
		t.Run(name, func(t *testing.T) {
			p := &pixelPointer{}
			var sequences []string
			for _, event := range events {
				if seq := p.handle(event); seq != "" {
					sequences = append(sequences, seq)
				}
			}
			assert.Equal(t, []string{ansi.SetModeMouseExtSgrPixel}, sequences, "enabled exactly once")
			assert.True(t, p.enabled)

			// Later cell size reports keep it on without sending it again.
			assert.Empty(t, p.handle(uv.CellSizeEvent{Width: 12, Height: 24}))
		})
	}
}

// paddedPointer returns a pointer switched on for 82x25 cells of 10x20 pixels
// in an 827x504 pixel window, as Ghostty reports it: 4 pixels of padding
// across and down count in the window's size, besides what's left over from
// fitting whole cells.
func paddedPointer(t *testing.T) *pixelPointer {
	t.Helper()
	p := &pixelPointer{}
	p.query()
	p.handle(uv.WindowSizeEvent{Width: 82, Height: 25})
	p.handle(uv.WindowPixelSizeEvent{Width: 827, Height: 504})
	p.handle(pixelModeReport(ansi.ModeReset))
	require.Equal(t, ansi.SetModeMouseExtSgrPixel, p.handle(uv.CellSizeEvent{Width: 10, Height: 20}))
	return p
}

func TestPixelPointer_WindowResizeAsksNothing(t *testing.T) {
	p := paddedPointer(t)
	// Dragging the window's edge: 99x30 cells of the same size in 1003x604.
	// Reading the window afresh sees its new size in pixels as its size in
	// cells arrives, before the event with the size in pixels.
	p.readWindow = func() (windowGeometry, bool) { return windowGeometry{99, 30, 1003, 604}, true }
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 99, Height: 30}))
	assert.Empty(t, p.handle(uv.WindowPixelSizeEvent{Width: 1003, Height: 604}))
	assert.False(t, p.awaiting)

	event, _, _ := p.locateEvent(uv.MouseMotionEvent{X: 985, Y: 585})
	assert.Equal(t, uv.MouseMotionEvent{X: 98, Y: 29}, event)
}

func TestPixelPointer_FontSizeChangeAsksOnce(t *testing.T) {
	p := paddedPointer(t)
	// The cells grow to 12x25 in the same window, 68x20 of them now. Only the
	// size in cells changes, which the old cells can't explain.
	assert.Equal(t, requestCellSize, p.handle(uv.WindowSizeEvent{Width: 68, Height: 20}))

	// Until the answer arrives, the window's size over its cells stands in.
	event, _, _ := p.locateEvent(uv.MouseMotionEvent{X: 246, Y: 137})
	assert.Equal(t, uv.MouseMotionEvent{X: 20, Y: 5}, event, "the old cells would put this at 24,6")

	// One question at a time: another change (to 14x29 cells) asks nothing
	// while it's out...
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 58, Height: 17}))
	// ...but the answer that arrives may be for the size before, so once it
	// does the terminal is asked again.
	assert.Equal(t, requestCellSize, p.handle(uv.CellSizeEvent{Width: 12, Height: 25}))
	assert.Empty(t, p.handle(uv.CellSizeEvent{Width: 14, Height: 29}))
	assert.False(t, p.awaiting)

	event, _, _ = p.locateEvent(uv.MouseMotionEvent{X: 14*57 + 1, Y: 29*16 + 1})
	assert.Equal(t, uv.MouseMotionEvent{X: 57, Y: 16}, event)

	p.disabled = true
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 82, Height: 25}))
}

func TestPixelPointer_KeepsCellSizeOnEmptyReply(t *testing.T) {
	p := paddedPointer(t)
	// A minimised window, say, is reported to have cells of no size.
	assert.Empty(t, p.handle(decodeAll(t, "\x1b[6;0;0t")[0]))
	assert.True(t, p.enabled)

	event, _, _ := p.locateEvent(uv.MouseClickEvent{X: 400, Y: 300})
	assert.Equal(t, uv.MouseClickEvent{X: 40, Y: 15}, event)
}

func TestPixelPointer_FallsBackToEvenWindowSize(t *testing.T) {
	// A terminal that supports mode 1016 but doesn't answer CSI 16 t. Its
	// window's size in pixels divides evenly into its cells, so the cells'
	// size comes from that.
	p := &pixelPointer{}
	p.query()
	p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
	p.handle(uv.WindowPixelSizeEvent{Width: 800, Height: 480})
	require.Equal(t, ansi.SetModeMouseExtSgrPixel, p.handle(pixelModeReport(ansi.ModeReset)))

	event, subX, _ := p.locateEvent(uv.MouseMotionEvent{X: 125, Y: 47})
	assert.Equal(t, uv.MouseMotionEvent{X: 12, Y: 2}, event)
	assert.InDelta(t, 0.5, subX, 1e-9)

	// It isn't asked again after a resize (here to 8x15 cells), and the
	// window gives the new size.
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 100, Height: 32}))
	event, _, _ = p.locateEvent(uv.MouseMotionEvent{X: 125, Y: 47})
	assert.Equal(t, uv.MouseMotionEvent{X: 15, Y: 3}, event)
	assert.False(t, p.replyDue(), "nothing to wait for before handing the terminal over")

	// Taking the terminal back switches it straight back on.
	assert.Equal(t, ansi.SetModeMouseExtSgrPixel, p.resume())
}

func TestPixelPointer_AsksOnEveryResizeWithoutWindowPixels(t *testing.T) {
	// A terminal that answers CSI 16 t but doesn't give the window's size in
	// pixels: a font size change can't be told from any other resize.
	p := &pixelPointer{}
	p.query()
	p.handle(pixelModeReport(ansi.ModeReset))
	p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
	require.Equal(t, ansi.SetModeMouseExtSgrPixel, p.handle(uv.CellSizeEvent{Width: 10, Height: 20}))

	assert.Equal(t, requestCellSize, p.handle(uv.WindowSizeEvent{Width: 90, Height: 24}))
	assert.True(t, p.replyDue())
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 100, Height: 24}), "one question at a time")
	assert.Equal(t, requestCellSize, p.handle(uv.CellSizeEvent{Width: 10, Height: 20}), "asked for another window")
	assert.Empty(t, p.handle(uv.CellSizeEvent{Width: 10, Height: 20}))
	assert.False(t, p.replyDue())
}

func TestPixelPointer_ResumeAsksBeforeSwitchingOn(t *testing.T) {
	p := paddedPointer(t)
	// The font size changed to 12x25 cells while the terminal was handed over.
	p.readWindow = func() (windowGeometry, bool) { return windowGeometry{68, 20, 827, 504}, true }
	assert.Equal(t, requestCellSize, p.resume(), "not switched on with the cells' old size")
	assert.False(t, p.enabled)
	assert.True(t, p.replyDue())

	// Mode 1016 is off, so positions are in cells meanwhile.
	assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 68, Height: 20}))
	m, _, _ := p.locate(uv.Mouse{X: 5, Y: 3})
	assert.Equal(t, uv.Mouse{X: 5, Y: 3}, m)

	assert.Equal(t, ansi.SetModeMouseExtSgrPixel, p.handle(uv.CellSizeEvent{Width: 12, Height: 25}))
	event, _, _ := p.locateEvent(uv.MouseMotionEvent{X: 246, Y: 137})
	assert.Equal(t, uv.MouseMotionEvent{X: 20, Y: 5}, event)
}

func TestPixelPointer_IgnoresWindowPixelSize(t *testing.T) {
	// Ghostty counts its window padding in the window's size in pixels, so
	// dividing that by the columns and rows gives cells that are too big.
	// Here 80x24 cells of 10x20 pixels sit in an 820x500 pixel window.
	p := &pixelPointer{}
	p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
	p.handle(uv.WindowPixelSizeEvent{Width: 820, Height: 500})
	assert.Empty(t, p.handle(pixelModeReport(ansi.ModeReset)), "no cell size reported yet")
	reply := decodeAll(t, "\x1b[6;20;10t") // The reply to CSI 16 t: height, then width.
	require.Equal(t, []uv.Event{uv.CellSizeEvent{Width: 10, Height: 20}}, reply)
	assert.Equal(t, ansi.SetModeMouseExtSgrPixel, p.handle(reply[0]))

	// The pointer at the left edge of the bottom-right cell is in that cell.
	event, subX, subY := p.locateEvent(uv.MouseMotionEvent{X: 790, Y: 460})
	assert.Equal(t, uv.MouseMotionEvent{X: 79, Y: 23}, event)
	assert.InDelta(t, 0, subX, 1e-9)
	assert.InDelta(t, 0, subY, 1e-9)
}

func TestPixelPointer_StaysOffWhenUnsupported(t *testing.T) {
	sized := func(p *pixelPointer) {
		p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
		p.handle(uv.CellSizeEvent{Width: 10, Height: 20})
	}
	for name, setup := range map[string]func(*pixelPointer){
		"not recognized":    func(p *pixelPointer) { sized(p); p.handle(pixelModeReport(ansi.ModeNotRecognized)) },
		"permanently reset": func(p *pixelPointer) { sized(p); p.handle(pixelModeReport(ansi.ModePermanentlyReset)) },
		"no report":         sized,
		"window pixels divide unevenly": func(p *pixelPointer) {
			p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
			p.handle(uv.WindowPixelSizeEvent{Width: 820, Height: 500})
			p.handle(pixelModeReport(ansi.ModeSet))
		},
		"zero cell size": func(p *pixelPointer) {
			p.handle(uv.CellSizeEvent{})
			p.handle(pixelModeReport(ansi.ModeSet))
		},
		"other mode": func(p *pixelPointer) {
			sized(p)
			p.handle(uv.ModeReportEvent{Mode: ansi.ModeMouseExtSgr, Value: ansi.ModeSet})
		},
		"opted out": func(p *pixelPointer) {
			p.disabled = true
			sized(p)
			p.handle(pixelModeReport(ansi.ModeSet))
		},
	} {
		t.Run(name, func(t *testing.T) {
			p := &pixelPointer{}
			setup(p)
			assert.False(t, p.enabled)
			m, subX, subY := p.locate(uv.Mouse{X: 12, Y: 3})
			assert.Equal(t, uv.Mouse{X: 12, Y: 3}, m, "cell positions pass through")
			assert.Equal(t, 0.5, subX)
			assert.Equal(t, 0.5, subY)
		})
	}
}

func TestPixelPointer_OptOutSkipsQuery(t *testing.T) {
	assert.Equal(t, ansi.RequestModeMouseExtSgrPixel+"\x1b[16t", (&pixelPointer{}).query())
	assert.Empty(t, (&pixelPointer{disabled: true}).query())
}

func TestPixelPointer_LocatesWithinCell(t *testing.T) {
	p := &pixelPointer{}
	p.handle(uv.CellSizeEvent{Width: 10, Height: 20})
	p.handle(pixelModeReport(ansi.ModeReset))
	require.True(t, p.enabled)

	event, subX, subY := p.locateEvent(uv.MouseMotionEvent{X: 125, Y: 47, Button: uv.MouseLeft})
	assert.Equal(t, uv.MouseMotionEvent{X: 12, Y: 2, Button: uv.MouseLeft}, event)
	assert.InDelta(t, 0.5, subX, 1e-9)
	assert.InDelta(t, 0.35, subY, 1e-9)

	// Positions in the window padding are reported as negative by some terminals.
	event, subX, subY = p.locateEvent(uv.MouseClickEvent{X: -3, Y: -1})
	assert.Equal(t, uv.MouseClickEvent{X: 0, Y: 0}, event)
	assert.Zero(t, subX)
	assert.Zero(t, subY)

	// Other events are untouched.
	key := uv.KeyPressEvent{Code: 'a'}
	other, _, _ := p.locateEvent(key)
	assert.Equal(t, uv.Event(key), other)
}

// decodeAll decodes input as the terminal reader does.
func decodeAll(t *testing.T, input string) []uv.Event {
	t.Helper()
	var decoder uv.EventDecoder
	var events []uv.Event
	for buf := []byte(input); len(buf) > 0; {
		n, event := decoder.Decode(buf)
		require.Positive(t, n, "decoder made no progress on %q", buf)
		events = append(events, event)
		buf = buf[n:]
	}
	return events
}

// feedAll passes events through repair, returning those left to handle.
func feedAll(repair *sgrMouseRepair, events []uv.Event) []uv.Event {
	var out []uv.Event
	for _, event := range events {
		if event = repair.feed(event); event != nil {
			out = append(out, event)
		}
	}
	return out
}

func TestSgrMouseRepair_NegativeCoordinates(t *testing.T) {
	for _, tc := range []struct {
		input    string
		repaired string // The same report with negative coordinates as 0.
	}{
		{"\x1b[<0;-3;5m", "\x1b[<0;0;5m"},       // Release outside the window.
		{"\x1b[<35;-3;5M", "\x1b[<35;0;5M"},     // Motion off the left edge.
		{"\x1b[<35;10;-2M", "\x1b[<35;10;0M"},   // Motion off the top edge.
		{"\x1b[<35;-1;-1M", "\x1b[<35;0;0M"},    // Motion off the top-left corner.
		{"\x1b[<32;-40;-7M", "\x1b[<32;0;0M"},   // Drag outside the window.
		{"\x1b[<0;-123;456M", "\x1b[<0;0;456M"}, // Press (reported while outside).
	} {
		t.Run(tc.input, func(t *testing.T) {
			split := decodeAll(t, tc.input)
			require.IsType(t, uv.UnknownEvent(""), split[0], "the decoder still splits this report")

			var repair sgrMouseRepair
			out := feedAll(&repair, split)
			for _, event := range out {
				_, isKey := event.(uv.KeyPressEvent)
				assert.False(t, isKey, "stray key press %v", event)
			}
			assert.Equal(t, decodeAll(t, tc.repaired), out)
			assert.Nil(t, repair.held)
		})
	}
}

func TestSgrMouseRepair_ReleaseOutsideWindow(t *testing.T) {
	var repair sgrMouseRepair
	out := feedAll(&repair, decodeAll(t, "\x1b[<0;-3;5m"))
	require.Len(t, out, 1)
	release, ok := out[0].(uv.MouseReleaseEvent)
	require.True(t, ok, "got %T", out[0])
	assert.Equal(t, uv.MouseLeft, release.Button)
	assert.Equal(t, 4, release.Y) // 1-based 5.
}

func TestSgrMouseRepair_EventsAfterReportPassThrough(t *testing.T) {
	var repair sgrMouseRepair
	out := feedAll(&repair, decodeAll(t, "\x1b[<35;-3;5M\x1b[<35;-1;-1Mq\x1b[I"))
	require.Len(t, out, 4)
	assert.IsType(t, uv.MouseMotionEvent{}, out[0])
	assert.IsType(t, uv.MouseMotionEvent{}, out[1])
	assert.Equal(t, uv.Event(uv.KeyPressEvent{Code: 'q', Text: "q"}), out[2])
	assert.IsType(t, uv.FocusEvent{}, out[3])
}

func TestSgrMouseRepair_OtherEventEndsReport(t *testing.T) {
	var repair sgrMouseRepair
	events := decodeAll(t, "\x1b[<0;-3")
	events = append(events, uv.KeyPressEvent{Code: 'x', Text: "x"})
	events = append(events, decodeAll(t, "5m")...)
	out := feedAll(&repair, events)
	// The held prefix is dropped; what follows the interruption is ordinary input.
	assert.Equal(t, []uv.Event{
		uv.KeyPressEvent{Code: 'x', Text: "x"},
		uv.KeyPressEvent{Code: '5', Text: "5"},
		uv.KeyPressEvent{Code: 'm', Text: "m"},
	}, out)
}

func TestSgrMouseRepair_LeavesOtherInputAlone(t *testing.T) {
	var repair sgrMouseRepair
	events := decodeAll(t, "12;-mM\x1b[<0;3;5m")
	events = append(events, uv.UnknownEvent("\x1b[?999z"))
	assert.Equal(t, events, feedAll(&repair, events))
}

func TestSgrMouseRepair_MalformedReportIsDropped(t *testing.T) {
	var repair sgrMouseRepair
	// Too many parameters to be a mouse report: nothing comes out.
	assert.Empty(t, feedAll(&repair, decodeAll(t, "\x1b[<0;-3;5;7m")))
	// A report that never ends stops being held once too long to be one.
	out := feedAll(&repair, decodeAll(t, "\x1b[<0;-"+strings.Repeat("1", 40)))
	assert.NotEmpty(t, out)
	assert.Nil(t, repair.held)
}
