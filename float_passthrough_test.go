package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func passiveFloat(child Widget) Floating {
	return Floating{
		Visible: true,
		Config:  FloatConfig{Position: FloatPositionTopLeft, PointerPassthrough: true},
		Child:   child,
	}
}

func TestFloatPointerPassthrough(t *testing.T) {
	t.Run("press without motion", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		overlay := &mouseRecorder{id: "preview", width: 10}
		router, _ := renderForMouse(Column{Children: []Widget{under, passiveFloat(overlay)}}, 20, 4)
		router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		router.release(uv.MouseReleaseEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5)
		assert.Len(t, under.downs, 1, "preview must not steal the press")
		assert.Len(t, under.ups, 1, "release must reach the pressed widget")
		assert.Empty(t, overlay.downs)
		assert.Empty(t, overlay.ups)
	})
	t.Run("wheel without motion", func(t *testing.T) {
		state := NewScrollState()
		rows := []Widget{Text{Content: "row 0"}, Text{Content: "row 1"}, Text{Content: "row 2"}}
		root := Column{Children: []Widget{
			Scrollable{State: state, Height: Cells(2), Child: Column{Children: rows}},
			passiveFloat(Text{Content: "preview"}),
		}}
		router, _ := renderForMouse(root, 20, 4)
		assert.True(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
		assert.Equal(t, 1, state.GetOffset())
	})
	t.Run("hover", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		router, renderer := renderForMouse(Column{Children: []Widget{under, passiveFloat(Text{Content: "preview"})}}, 20, 4)
		router.motion(uv.MouseMotionEvent{X: 1, Y: 0}, 0.5, 0.5)
		hovered, ok := renderer.hoveredSignal.Peek().(Identifiable)
		require.True(t, ok, "underlying widget should own hover")
		assert.Equal(t, "under", hovered.WidgetID())
	})
	t.Run("focus and pointer owner without motion", func(t *testing.T) {
		state := NewTextInputState("abcd")
		root := Column{Children: []Widget{
			TextInput{ID: "input", State: state, Width: Cells(10)},
			Button{ID: "other", Label: "other"},
			passiveFloat(Text{ID: "preview", Content: "preview"}),
		}}
		router, renderer := renderForMouse(root, 20, 4)
		router.focusManager.FocusByID("other")
		require.NotNil(t, renderer.FloatAt(1, 0), "visual float queries still include the preview")
		assert.Equal(t, "preview", renderer.FloatAt(1, 0).Child.(Text).ID)
		require.NotNil(t, renderer.PointerOwnerAt(1, 0))
		assert.Equal(t, "input", renderer.PointerOwnerAt(1, 0).ID)
		router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.Equal(t, "input", router.focusManager.FocusedID())
		assert.Equal(t, 1, state.CursorIndex.Peek(), "input owner receives the press and places its cursor")
	})
	t.Run("drag reaches underlying widget", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		overlay := &mouseRecorder{id: "preview", width: 10}
		router, _ := renderForMouse(Column{Children: []Widget{under, passiveFloat(overlay)}}, 20, 4)
		router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		router.motion(uv.MouseMotionEvent{X: 2, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5)
		router.release(uv.MouseReleaseEvent{X: 2, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5)
		assert.Len(t, under.moves, 1)
		assert.Empty(t, overlay.moves)
	})
	t.Run("collection item hover", func(t *testing.T) {
		state := NewListState([]string{"first", "second"})
		list := List[string]{ID: "list", State: state}
		router, _ := renderForMouse(Column{Children: []Widget{list, passiveFloat(Text{Content: "preview"})}}, 20, 4)
		router.motion(uv.MouseMotionEvent{X: 1, Y: 0}, 0.5, 0.5)
		assert.True(t, list.itemHovered(0))
		assert.False(t, list.itemHovered(1))
	})
}

type passiveFloatHoverApp struct {
	visible Signal[bool]
	under   Widget
}

func (a *passiveFloatHoverApp) Build(BuildContext) Widget {
	preview := passiveFloat(Text{Content: "preview"})
	preview.Visible = a.visible.Get()
	return Column{Children: []Widget{a.under, preview}}
}

func TestFloatPointerPassthrough_StationaryHover(t *testing.T) {
	under := &mouseRecorder{id: "under", width: 10}
	app := &passiveFloatHoverApp{visible: NewSignal(false), under: under}
	router, renderer := renderForMouse(app, 20, 4)
	router.motion(uv.MouseMotionEvent{X: 1, Y: 0}, 0.5, 0.5)
	app.visible.Set(true)
	renderer.Update(app)
	router.reconcileHover()
	assert.Equal(t, under, renderer.hoveredSignal.Peek(), "showing a preview must preserve hover without another mouse movement")
	assert.Equal(t, "under", hitID(renderer, 1, 0))
}

func TestFloatPointerPassthrough_StackingAndDismissal(t *testing.T) {
	t.Run("interactive float beneath preview still receives press", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		menu := &mouseRecorder{id: "menu", width: 10}
		previewDismissed, menuDismissed := false, false
		preview := passiveFloat(Text{Content: "preview"})
		preview.Config.OnDismiss = func() { previewDismissed = true }
		root := Column{Children: []Widget{
			under,
			Floating{Visible: true, Config: FloatConfig{OnDismiss: func() { menuDismissed = true }}, Child: menu},
			preview,
		}}
		router, renderer := renderForMouse(root, 20, 4)
		router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.Len(t, menu.downs, 1)
		assert.Empty(t, under.downs)
		assert.False(t, previewDismissed)
		assert.False(t, menuDismissed)
		assert.Equal(t, preview.Child, renderer.TopFloat().Child, "visual ordering is unchanged")
		// A click outside the menu still dismisses it despite the passive float above it.
		router.press(uv.MouseClickEvent{X: 15, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.True(t, menuDismissed)
		assert.False(t, previewDismissed)
	})
	t.Run("passive float does not consume outside click", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		dismissed := false
		preview := passiveFloat(Text{Content: "tip"})
		preview.Config.OnDismiss = func() { dismissed = true }
		router, _ := renderForMouse(Column{Children: []Widget{under, preview}}, 20, 4)
		router.press(uv.MouseClickEvent{X: 5, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.Len(t, under.downs, 1)
		assert.False(t, dismissed)
	})
	t.Run("modal takes precedence", func(t *testing.T) {
		under := &mouseRecorder{id: "under", width: 10}
		modal := &mouseRecorder{id: "modal", width: 5}
		preview := passiveFloat(modal)
		preview.Config.Modal = true
		router, _ := renderForMouse(Column{Children: []Widget{under, preview}}, 20, 4)
		router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.Len(t, modal.downs, 1)
		router.press(uv.MouseClickEvent{X: 8, Y: 0, Button: uv.MouseLeft}, 0.5, 0.5, time.Now())
		assert.Empty(t, under.downs, "the modal backdrop still blocks the underlying widget")
	})
}
