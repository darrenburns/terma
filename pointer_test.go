package terma

import (
	"testing"

	uv "github.com/darrenburns/ultraviolet"
	"github.com/stretchr/testify/require"
)

func TestPointerStyles(t *testing.T) {
	shape := NewSignal(PointerText)
	root := pilotBuilder(func(BuildContext) Widget {
		return Column{Style: Style{Pointer: PointerPointer}, Children: []Widget{
			Text{ID: "inherited", Content: "inherit"},
			Text{ID: "default", Content: "default", Style: Style{Pointer: PointerDefault}},
			Button{ID: "changing", Label: "changing", Style: Style{Pointer: shape.Get()}},
			DisabledWhen(true, Text{ID: "disabled", Content: "disabled", Style: Style{Pointer: PointerWait}}),
		}}
	})
	p := NewPilot(t, root, 20, 8)
	for _, tc := range []struct {
		id   string
		want PointerShape
	}{
		{"inherited", PointerPointer},
		{"default", PointerDefault},
		{"changing", PointerText},
		{"disabled", PointerDefault},
	} {
		p.Hover(tc.id)
		require.Equal(t, tc.want, p.PointerShape(), tc.id)
	}
	p.Hover("changing")
	shape.Set(PointerWait)
	p.Advance(0)
	require.Equal(t, PointerWait, p.PointerShape(), "style changes under a stationary mouse")
	shape.Set(PointerShape("text\x07injected"))
	p.Advance(0)
	require.Equal(t, PointerDefault, p.PointerShape(), "invalid names cannot inject terminal output")
	p.MouseMove(19, 7)
	require.Equal(t, PointerDefault, p.PointerShape(), "empty space resets the shape")
	require.False(t, (Style{Pointer: PointerDefault}).IsZero())
}

func TestPointerOverlayLayers(t *testing.T) {
	for _, tc := range []struct {
		name        string
		passthrough bool
		modal       bool
	}{
		{name: "overlay"},
		{name: "modal", modal: true},
		{name: "passthrough", passthrough: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			visible := NewSignal(false)
			p := NewPilot(t, pilotBuilder(func(BuildContext) Widget {
				return Column{Style: Style{Pointer: PointerText}, Children: []Widget{
					Text{Content: "under", Width: Cells(20), Height: Cells(6)},
					Floating{
						Visible: visible.Get(),
						Config:  FloatConfig{Position: FloatPositionTopLeft, Modal: tc.modal, PointerPassthrough: tc.passthrough},
						Child:   Text{Content: "over", Width: Cells(5), Height: Cells(2)},
					},
				}}
			}), 20, 6)
			p.MouseMove(1, 1)
			require.Equal(t, PointerText, p.PointerShape())
			visible.Set(true)
			p.Advance(0)
			want := PointerDefault
			if tc.passthrough {
				want = PointerText
			}
			require.Equal(t, want, p.PointerShape(), "overlay does not inherit a covered widget's pointer")
			p.MouseMove(15, 4)
			want = PointerText
			if tc.modal {
				want = PointerDefault
			}
			require.Equal(t, want, p.PointerShape(), "modal backdrop blocks underlying pointer styles")
			visible.Set(false)
			p.Advance(0)
			require.Equal(t, PointerText, p.PointerShape())
		})
	}
}

func TestPointerCaptureAndRemoval(t *testing.T) {
	visible := NewSignal(true)
	p := NewPilot(t, pilotBuilder(func(BuildContext) Widget {
		return Column{Children: []Widget{
			ShowWhen(visible.Get(), Text{ID: "grab", Content: "grab", Style: Style{Pointer: PointerGrab}}),
			Text{Content: "text", Style: Style{Pointer: PointerText}},
		}}
	}), 20, 6)
	// A press is enough to locate the pointer, even without prior motion.
	p.MouseDown(1, 0, uv.MouseLeft, 0)
	require.Equal(t, PointerGrab, p.PointerShape())
	p.MouseMove(1, 1)
	require.Equal(t, PointerGrab, p.PointerShape(), "capture keeps the pressed widget's shape")
	p.MouseUp(1, 1, uv.MouseLeft, 0)
	require.Equal(t, PointerText, p.PointerShape(), "release uses the widget now under the mouse")
	p.MouseDown(1, 0, uv.MouseLeft, 0)
	visible.Set(false)
	p.Advance(0)
	require.Equal(t, PointerText, p.PointerShape(), "removed capture cannot keep a stale shape")
}

func TestPointerBuiltSplitPaneOverride(t *testing.T) {
	shape := NewSignal(PointerInherit)
	state := NewSplitPaneState(0.5)
	p := NewPilot(t, pilotBuilder(func(BuildContext) Widget {
		return SplitPane{
			State: state,
			Style: Style{Width: Cells(20), Height: Cells(8), Padding: EdgeInsetsAll(1), Pointer: shape.Get()},
		}
	}), 24, 10)
	p.MouseMove(10, 3)
	require.Equal(t, PointerEWResize, p.PointerShape(), "built-in behavior works when returned by a custom Build")
	p.MouseMove(10, 0)
	require.Equal(t, PointerDefault, p.PointerShape(), "padding is not a divider even with the same hover identity")
	p.MouseMove(10, 3)
	shape.Set(PointerGrab)
	p.Advance(0)
	require.Equal(t, PointerGrab, p.PointerShape(), "explicit style overrides the automatic resize shape")
	shape.Set(PointerDefault)
	p.Advance(0)
	require.Equal(t, PointerDefault, p.PointerShape())
}
