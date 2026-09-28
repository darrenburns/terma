package terma

import (
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
		"report first": {pixelModeReport(ansi.ModeReset), uv.WindowSizeEvent{Width: 80, Height: 24}, uv.WindowPixelSizeEvent{Width: 800, Height: 480}},
		"sizes first":  {uv.WindowSizeEvent{Width: 80, Height: 24}, uv.WindowPixelSizeEvent{Width: 800, Height: 480}, pixelModeReport(ansi.ModeReset)},
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

			// Later resizes keep it on without sending it again.
			assert.Empty(t, p.handle(uv.WindowSizeEvent{Width: 100, Height: 30}))
			assert.Empty(t, p.handle(uv.WindowPixelSizeEvent{Width: 1000, Height: 600}))
		})
	}
}

func TestPixelPointer_StaysOffWhenUnsupported(t *testing.T) {
	sized := func(p *pixelPointer) {
		p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
		p.handle(uv.WindowPixelSizeEvent{Width: 800, Height: 480})
	}
	for name, setup := range map[string]func(*pixelPointer){
		"not recognized":    func(p *pixelPointer) { sized(p); p.handle(pixelModeReport(ansi.ModeNotRecognized)) },
		"permanently reset": func(p *pixelPointer) { sized(p); p.handle(pixelModeReport(ansi.ModePermanentlyReset)) },
		"no report":         sized,
		"no pixel size": func(p *pixelPointer) {
			p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
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
	assert.Equal(t, ansi.RequestModeMouseExtSgrPixel, (&pixelPointer{}).query())
	assert.Empty(t, (&pixelPointer{disabled: true}).query())
}

func TestPixelPointer_LocatesWithinCell(t *testing.T) {
	p := &pixelPointer{}
	p.handle(uv.WindowSizeEvent{Width: 80, Height: 24})
	p.handle(uv.WindowPixelSizeEvent{Width: 800, Height: 480}) // 10x20 pixel cells
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
