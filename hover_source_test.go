package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHoverSource_KeyboardSummaryUnderStationaryPointer(t *testing.T) {
	visible := NewSignal(false)
	summary := Text{
		ID: "summary", Content: "Summary", Width: Cells(12), Height: Cells(3),
		Hover: func(event HoverEvent) {
			if event.Type == HoverEnter && event.Source == HoverSourcePointer {
				visible.Set(false)
			}
		},
	}
	root := func() Widget {
		return Column{Children: []Widget{
			Text{ID: "base", Content: "Base", Width: Cells(20), Height: Cells(5)},
			Floating{Visible: visible.Peek(), Config: FloatConfig{Offset: Offset{X: 1, Y: 1}}, Child: summary},
		}}
	}
	router, renderer := renderForMouse(root(), 20, 5)
	router.motion(uv.MouseMotionEvent{X: 2, Y: 2}, 0.5, 0.5)
	visible.Set(true) // Keyboard navigation opens the summary.
	renderer.Render(root())
	require.True(t, router.reconcileHover())
	assert.Equal(t, "summary", router.hover.currentID)
	assert.True(t, visible.Peek(), "layout enter must be distinguishable so keyboard navigation keeps its summary visible")

	// Moving away and back still dismisses the summary on a real pointer enter.
	router.motion(uv.MouseMotionEvent{X: 18, Y: 2}, 0.5, 0.5)
	router.motion(uv.MouseMotionEvent{X: 2, Y: 2}, 0.5, 0.5)
	assert.False(t, visible.Peek())
}

func TestHoverSource_StationaryOverlayAppearsAndDisappears(t *testing.T) {
	var events []HoverEvent
	record := func(event HoverEvent) { events = append(events, event) }
	base := Text{ID: "base", Content: "Base", Width: Cells(20), Height: Cells(5), Hover: record}
	root := func(visible bool) Widget {
		return Column{Children: []Widget{base, Floating{
			Visible: visible, Config: FloatConfig{Offset: Offset{X: 1, Y: 1}},
			Child: Text{ID: "summary", Content: "Summary", Width: Cells(12), Height: Cells(3), Hover: record},
		}}}
	}
	router, renderer := renderForMouse(root(false), 20, 5)
	router.motion(uv.MouseMotionEvent{X: 2, Y: 2, Mod: uv.ModShift, Button: uv.MouseLeft}, 0.5, 0.5)
	require.Len(t, events, 1)
	assert.Equal(t, HoverSourcePointer, events[0].Source)
	events = nil

	for _, visible := range []bool{true, false} {
		renderer.Render(root(visible))
		assert.Empty(t, events, "render alone must not dispatch hover")
		require.True(t, router.reconcileHover())
		require.Len(t, events, 2)
		assert.Equal(t, HoverLeave, events[0].Type)
		assert.Equal(t, HoverEnter, events[1].Type)
		for _, event := range events {
			assert.Equal(t, HoverSourceLayout, event.Source)
			assert.Equal(t, 2, event.X)
			assert.Equal(t, 2, event.Y)
			assert.Equal(t, uv.ModShift, event.Mod)
			assert.Equal(t, uv.MouseLeft, event.Button)
		}
		target := "base"
		if visible {
			target = "summary"
			assert.Equal(t, 1, events[1].LocalX)
			assert.Equal(t, 1, events[1].LocalY)
		}
		assert.Equal(t, target, events[1].NextWidgetID)
		assert.Equal(t, target, router.hover.currentID)
		ctx := NewBuildContext(router.focusManager, NewAnySignal[Focusable](nil), router.hoveredSignal, nil)
		assert.Equal(t, target, ctx.HoveredID(), "source filtering must not suppress hover styling/state")
		events = nil
		assert.False(t, router.reconcileHover(), "unchanged identity emits no additional transitions")
		assert.Empty(t, events)
	}
}

func TestHoverSource_ModifierAndButtonMotionReports(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mod    uv.KeyMod
		button uv.MouseButton
	}{
		{name: "modifier", mod: uv.ModAlt},
		{name: "button", button: uv.MouseLeft},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var events []HoverEvent
			record := func(event HoverEvent) { events = append(events, event) }
			root := Text{ID: "old", Content: "Old", Width: Cells(5), Hover: record}
			router, renderer := renderForMouse(root, 10, 2)
			router.motion(uv.MouseMotionEvent{X: 1}, 0.5, 0.5)
			events = nil
			renderer.Render(Text{ID: "new", Content: "New", Width: Cells(5), Hover: record})
			// Cell position is unchanged, but this is new mouse input rather than
			// the post-render reconciliation path.
			require.True(t, router.motion(uv.MouseMotionEvent{X: 1, Mod: tc.mod, Button: tc.button}, 0.5, 0.5))
			require.Len(t, events, 2)
			for _, event := range events {
				assert.Equal(t, HoverSourcePointer, event.Source)
				assert.Equal(t, tc.mod, event.Mod)
				assert.Equal(t, tc.button, event.Button)
			}
			events = nil
			assert.False(t, router.motion(uv.MouseMotionEvent{X: 1, Mod: tc.mod, Button: tc.button}, 0.2, 0.3))
			assert.Empty(t, events, "duplicate/subcell motion does not create a transition")
			assert.False(t, router.reconcileHover())
		})
	}
}
