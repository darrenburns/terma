package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type commandPaletteAnchorScene struct {
	palette CommandPalette
	lead    Signal[int]
	top     Signal[int]
}

func newCommandPaletteAnchorScene() *commandPaletteAnchorScene {
	state := NewCommandPaletteState("Tabs", []CommandPaletteItem{{Label: "Request one"}, {Label: "Request two"}})
	state.Open()
	return &commandPaletteAnchorScene{
		palette: CommandPalette{ID: "tab-picker", State: state, AnchorID: "tab-bar", Style: Style{Width: Cells(20)}},
		lead:    NewSignal(8), top: NewSignal(3),
	}
}

func (s *commandPaletteAnchorScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "top", build: func(BuildContext) Widget { return Spacer{Height: Cells(s.top.Get())} }},
		Row{Style: Style{Width: Flex(1)}, Children: []Widget{
			reactivityBuilder{ID: "lead", build: func(BuildContext) Widget { return Spacer{Width: Cells(s.lead.Get())} }},
			Text{ID: "tab-bar", Content: "Request tabs", Style: Style{Width: Flex(1)}},
		}},
		s.palette,
	}}
}

func requirePaletteBelowAnchor(t *testing.T, renderer *Renderer) Rect {
	t.Helper()
	anchor := renderer.WidgetByID("tab-bar")
	palette := renderer.WidgetByID("tab-picker-content")
	require.NotNil(t, anchor)
	require.NotNil(t, palette)
	require.Equal(t, anchor.Bounds.X, palette.Bounds.X, "palette should align with the tab bar's current left edge")
	require.Equal(t, anchor.Bounds.Y+anchor.Bounds.Height, palette.Bounds.Y, "palette should touch the bottom of the tab bar without a screen-top inset")
	return palette.Bounds
}

func TestCommandPaletteAnchorFirstFrame(t *testing.T) {
	scene := newCommandPaletteAnchorScene()
	buffer := uv.NewBuffer(80, 30)
	focus := NewFocusManager()
	renderer := NewRenderer(reactivityScreen{buffer}, 80, 30, focus, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	previousPendingFocus := pendingFocusID
	t.Cleanup(func() {
		pendingFocusID = previousPendingFocus
		if renderer.rootNode != nil {
			renderer.rootNode.dispose()
		}
		for _, float := range renderer.retainedFloats {
			float.root.dispose()
		}
	})
	// A single render: the anchor hasn't been measured or cached by an earlier frame.
	renderer.Render(scene)
	requirePaletteBelowAnchor(t, renderer)
	require.Contains(t, renderer.ScreenText(), "Request one")
}

func TestCommandPaletteAnchorFollowsLayoutAndResize(t *testing.T) {
	sequence := newReactivitySequence(t, 80, 30, newCommandPaletteAnchorScene)
	sequence.frame("Initially anchored", nil)
	first := requirePaletteBelowAnchor(t, sequence.actual.renderer)
	sequence.frame("Tab bar moves horizontally", func(s *commandPaletteAnchorScene) { s.lead.Set(18) })
	moved := requirePaletteBelowAnchor(t, sequence.actual.renderer)
	require.NotEqual(t, first.X, moved.X)
	sequence.frame("Tab bar moves vertically", func(s *commandPaletteAnchorScene) { s.top.Set(6) })
	vertical := requirePaletteBelowAnchor(t, sequence.actual.renderer)
	require.NotEqual(t, moved.Y, vertical.Y)
	sequence.resize(60, 24)
	sequence.frame("Terminal resizes", nil)
	requirePaletteBelowAnchor(t, sequence.actual.renderer)
}

func TestCommandPaletteAnchorRightFollowsResize(t *testing.T) {
	sequence := newReactivitySequence(t, 80, 30, func() *commandPaletteAnchorScene {
		scene := newCommandPaletteAnchorScene()
		scene.palette.Anchor = AnchorBottomRight
		scene.palette.Offset = Offset{X: -1, Y: 1}
		// AnchorID takes precedence over screen positioning.
		scene.palette.Position = FloatPositionCenter
		return scene
	})
	check := func() Rect {
		anchor := sequence.actual.renderer.WidgetByID("tab-bar")
		palette := sequence.actual.renderer.WidgetByID("tab-picker-content")
		require.NotNil(t, anchor)
		require.NotNil(t, palette)
		require.Equal(t, anchor.Bounds.X+anchor.Bounds.Width-palette.Bounds.Width-1, palette.Bounds.X)
		require.Equal(t, anchor.Bounds.Y+anchor.Bounds.Height+1, palette.Bounds.Y)
		return palette.Bounds
	}
	sequence.frame("Right aligned with explicit offset", nil)
	before := check()
	sequence.resize(60, 24)
	sequence.frame("Right aligned after terminal resize", nil)
	after := check()
	require.Equal(t, before.X-20, after.X, "palette follows the resized tab bar's right edge")
}

func TestCommandPalettePositionCompatibility(t *testing.T) {
	tests := []struct {
		name     string
		position FloatPosition
		offset   Offset
		x, y     int
	}{
		{name: "default top center", x: 30, y: 2},
		{name: "top left keeps default inset", position: FloatPositionTopLeft, offset: Offset{X: 3}, x: 3, y: 2},
		{name: "explicit absolute coordinates", position: FloatPositionAbsolute, offset: Offset{X: 7, Y: 9}, x: 7, y: 9},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sequence := newReactivitySequence(t, 80, 30, func() *commandPaletteAnchorScene {
				scene := newCommandPaletteAnchorScene()
				scene.palette.AnchorID = ""
				scene.palette.Position, scene.palette.Offset = tt.position, tt.offset
				return scene
			})
			sequence.frame("Screen positioned palette", nil)
			palette := sequence.actual.renderer.WidgetByID("tab-picker-content")
			require.NotNil(t, palette)
			require.Equal(t, tt.x, palette.Bounds.X)
			require.Equal(t, tt.y, palette.Bounds.Y)
		})
	}
}

func TestCommandPaletteAnchorTopUsesZeroOffset(t *testing.T) {
	sequence := newReactivitySequence(t, 80, 30, func() *commandPaletteAnchorScene {
		scene := newCommandPaletteAnchorScene()
		scene.top.Set(15)
		scene.palette.Anchor = AnchorTopLeft
		scene.palette.Position = FloatPositionTopCenter
		return scene
	})
	sequence.frame("Above anchor without screen inset", nil)
	anchor := sequence.actual.renderer.WidgetByID("tab-bar")
	palette := sequence.actual.renderer.WidgetByID("tab-picker-content")
	require.NotNil(t, anchor)
	require.NotNil(t, palette)
	require.Equal(t, anchor.Bounds.X, palette.Bounds.X)
	require.Equal(t, anchor.Bounds.Y-palette.Bounds.Height, palette.Bounds.Y)
}
