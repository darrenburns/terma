package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dragTestApp struct {
	shown  Signal[bool]
	label  Signal[string]
	input  *TextInputState
	builds int
	drops  []string
}

func (a *dragTestApp) Build(BuildContext) Widget {
	children := []Widget{}
	if a.shown.Get() {
		children = append(children, Draggable[string]{ID: "source", Payload: "card", HandleID: "handle", Child: dragTestContent{app: a}})
	}
	children = append(children, DropTarget[string]{ID: "target", OnDrop: func(value string) { a.drops = append(a.drops, value) }, Child: Text{Content: "destination", Width: Cells(14), Height: Cells(5)}})
	return Row{Children: children}
}

type dragTestContent struct{ app *dragTestApp }

func (w dragTestContent) Build(BuildContext) Widget {
	w.app.builds++
	return Column{Width: Cells(8), Children: []Widget{
		Text{ID: "handle", Content: w.app.label.Get()},
		TextInput{ID: "editor", State: w.app.input, Width: Cells(8)},
	}}
}

func TestDragRetainsWidgetAndFocus(t *testing.T) {
	a := &dragTestApp{shown: NewSignal(true), label: NewSignal("handle"), input: NewTextInputState("hello")}
	m, r := renderForMouse(a, 40, 10)
	m.focusManager.FocusByID("editor")
	source, editor := r.WidgetByID("source").node, r.WidgetByID("editor").node
	initialBuilds := a.builds
	m.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, .25, .75, time.Now())
	m.motion(uv.MouseMotionEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .25, .75)
	r.Update(a)
	require.NotNil(t, r.drag)
	assert.Same(t, source, r.WidgetByID("source").node)
	assert.Same(t, editor, r.WidgetByID("editor").node)
	assert.Equal(t, initialBuilds, a.builds, "motion must not rebuild the content")
	assert.Equal(t, Rect{X: 11, Y: 3, Width: 8, Height: 2}, r.WidgetByID("source").Bounds)
	assert.Equal(t, "editor", m.focusManager.FocusedID())
	assert.Equal(t, "target", r.WidgetAt(12, 3).parentID, "the lifted widget does not hide the target")
	assert.NotContains(t, r.ScreenText()[:40], "handle", "the source slot is empty")
	a.label.Set("updated")
	r.Update(a)
	assert.Same(t, source, r.WidgetByID("source").node)
	assert.Contains(t, r.ScreenText(), "update", "live content still paints inside the frozen size")
	m.release(uv.MouseReleaseEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .25, .75)
	r.Update(a)
	assert.Equal(t, []string{"card"}, a.drops)
	assert.Nil(t, r.drag)
	assert.Same(t, source, r.WidgetByID("source").node)
	assert.Equal(t, 0, r.WidgetByID("source").Bounds.X)
	assert.Equal(t, "hello", a.input.GetText())
	counts := map[string]int{}
	for _, entry := range r.widgetRegistry.entries {
		counts[entry.ID]++
	}
	assert.Equal(t, 1, counts["source"])
	assert.Equal(t, 1, counts["editor"])
}

func TestDraggableTextAndCancellation(t *testing.T) {
	for _, cancel := range []string{"escape", "lost button", "removal", "resize"} {
		t.Run(cancel, func(t *testing.T) {
			a := &dragTestApp{shown: NewSignal(true), label: NewSignal("handle"), input: NewTextInputState("")}
			m, r := renderForMouse(a, 40, 10)
			m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
			m.motion(uv.MouseMotionEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .5, .5)
			r.Update(a)
			switch cancel {
			case "escape":
				dispatchKey(r, m.focusManager, a, KeyEvent{event: uv.KeyPressEvent{Code: 27}})
			case "lost button":
				m.motion(uv.MouseMotionEvent{X: 12, Y: 3}, .5, .5)
			case "removal":
				a.shown.Set(false)
			case "resize":
				r.Resize(39, 10)
			}
			r.Update(a)
			assert.Nil(t, r.drag)
			assert.Empty(t, a.drops)
		})
	}
	t.Run("passive text", func(t *testing.T) {
		root := Draggable[int]{ID: "source", Payload: 7, Child: Text{Content: "drag me"}}
		m, r := renderForMouse(root, 20, 5)
		m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
		m.motion(uv.MouseMotionEvent{X: 5, Y: 2, Button: uv.MouseLeft}, .5, .5)
		r.Update(root)
		require.NotNil(t, r.drag)
		assert.Equal(t, dragging, r.drag.phase)
	})
}

func TestDragReleaseRefreshesRemovedSource(t *testing.T) {
	a := &dragTestApp{shown: NewSignal(true), label: NewSignal("handle"), input: NewTextInputState("")}
	m, r := renderForMouse(a, 40, 10)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .5, .5)
	r.Update(a)
	a.shown.Set(false)
	m.release(uv.MouseReleaseEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .5, .5)
	assert.Empty(t, a.drops)
	assert.Nil(t, r.drag)
}

func TestDraggablePreservesPressReleaseAndInteractiveChildren(t *testing.T) {
	down, up, click := 0, 0, 0
	root := Draggable[int]{ID: "source", Child: Text{Content: "click", MouseDown: func(MouseEvent) { down++ }, MouseUp: func(MouseEvent) { up++ }, Click: func(MouseEvent) { click++ }}}
	m, r := renderForMouse(root, 20, 6)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	assert.Equal(t, 1, click, "click happens on press")
	m.release(uv.MouseReleaseEvent{X: 1, Button: uv.MouseLeft}, .5, .5)
	assert.Equal(t, 1, down)
	assert.Equal(t, 1, up)
	r.Update(root)
	input := NewTextInputState("edit me")
	other := Draggable[int]{ID: "source", Child: TextInput{ID: "input", State: input}}
	r.Render(other)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	assert.Nil(t, r.drag, "a focusable child keeps its editing gesture")
}

func TestDragTypedTargetsAndButtonMatching(t *testing.T) {
	for _, mode := range []string{"accepted", "wrong type", "rejected", "outside"} {
		t.Run(mode, func(t *testing.T) {
			calls := 0
			var target Widget = DropTarget[string]{ID: "target", Child: Text{Content: "target", Width: Cells(12), Height: Cells(4)}, Accept: func(string) bool { return mode != "rejected" }, OnDrop: func(string) { calls++ }}
			if mode == "wrong type" {
				target = DropTarget[int]{ID: "target", Child: Text{Content: "target", Width: Cells(12), Height: Cells(4)}, OnDrop: func(int) { calls++ }}
			}
			root := Row{Children: []Widget{Draggable[string]{ID: "source", Payload: "value", Child: Text{Content: "drag"}}, target}}
			m, r := renderForMouse(root, 30, 8)
			m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .25, .75, time.Now())
			m.motion(uv.MouseMotionEvent{X: 8, Y: 2, Button: uv.MouseLeft}, .25, .75)
			r.Update(root)
			m.release(uv.MouseReleaseEvent{X: 8, Y: 2, Button: uv.MouseRight}, .25, .75)
			require.NotNil(t, r.drag, "another button cannot complete the drag")
			y := 2
			if mode == "outside" {
				y = 7
			}
			m.release(uv.MouseReleaseEvent{X: 8, Y: y, Button: uv.MouseLeft}, .25, .75)
			m.release(uv.MouseReleaseEvent{X: 8, Y: y, Button: uv.MouseLeft}, .25, .75)
			want := 0
			if mode == "accepted" {
				want = 1
			}
			assert.Equal(t, want, calls)
			r.Update(root)
			assert.Equal(t, 0, r.WidgetByID("source").Bounds.X)
		})
	}
}

type changingDragContent struct{ label Signal[string] }

func (a *changingDragContent) Build(BuildContext) Widget { return Text{Content: a.label.Get()} }

func TestDragRestoresFreshIntrinsicLayout(t *testing.T) {
	child := &changingDragContent{label: NewSignal("short")}
	root := Draggable[int]{ID: "source", Child: child}
	m, r := renderForMouse(root, 60, 8)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: 10, Y: 3, Button: uv.MouseLeft}, .5, .5)
	r.Update(root)
	child.label.Set("a considerably wider label")
	r.Update(root)
	assert.Equal(t, 5, r.WidgetByID("source").Bounds.Width)
	m.release(uv.MouseReleaseEvent{X: 10, Y: 3, Button: uv.MouseLeft}, .5, .5)
	r.Update(root)
	assert.Equal(t, len(child.label.Peek()), r.WidgetByID("source").Bounds.Width)
	assert.Contains(t, r.ScreenText(), child.label.Peek())
}

func TestDragFloatsAboveTargetAndKeepsNestedFloat(t *testing.T) {
	root := Column{Children: []Widget{
		Draggable[string]{ID: "source", Payload: "v", HandleID: "handle", Child: Column{Children: []Widget{
			Text{ID: "handle", Content: "DRAG"},
			Floating{Visible: true, Config: FloatConfig{AnchorID: "handle", Anchor: AnchorBottomLeft, PointerPassthrough: true}, Child: Text{ID: "nested", Content: "TIP"}},
		}}},
		Floating{Visible: true, Config: FloatConfig{Offset: Offset{X: 10, Y: 3}}, Child: DropTarget[string]{ID: "target", OnDrop: func(string) {}, Child: Text{Content: "DESTINATION", Width: Cells(12), Height: Cells(4)}}},
	}}
	m, r := renderForMouse(root, 35, 10)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: 12, Y: 3, Button: uv.MouseLeft}, .5, .5)
	r.Update(root)
	assert.Equal(t, "D", r.terminal.CellAt(11, 3).Content, "drag content is above the destination float")
	assert.Equal(t, Rect{X: 11, Y: 4, Width: 3, Height: 1}, r.WidgetByID("nested").Bounds)
	assert.Equal(t, "T", r.terminal.CellAt(11, 4).Content, "source-owned overlays remain above the source")
}

type dragModalApp struct{ second Signal[bool] }

func (a *dragModalApp) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Floating{Visible: true, Config: FloatConfig{Modal: true}, Child: Draggable[int]{ID: "source", Child: Text{Content: "drag"}}},
		Floating{Visible: a.second.Get(), Config: FloatConfig{Modal: true}, BuildChild: func(BuildContext, FloatGeometry) Widget { return Text{Content: "second modal"} }},
	}}
}

func TestDragNewModalCancelsUnnamedScope(t *testing.T) {
	a := &dragModalApp{second: NewSignal(false)}
	m, r := renderForMouse(a, 30, 8)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: 5, Y: 2, Button: uv.MouseLeft}, .5, .5)
	r.Update(a)
	require.NotNil(t, r.drag)
	a.second.Set(true)
	r.Update(a)
	assert.Nil(t, r.drag)
}

func TestDragRetainsImageRecord(t *testing.T) {
	root := Draggable[int]{ID: "source", Child: Image{ID: "image", Source: testImage(t, 16, 16), Fit: ImageStretch, Style: Style{Width: Cells(4), Height: Cells(3)}}}
	m, r := renderForMouse(root, 25, 10)
	node := r.WidgetByID("image").node
	key := imageRecordKey{node: node, slot: 0}
	record := r.images.records[key]
	require.NotNil(t, record)
	m.press(uv.MouseClickEvent{X: 1, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: 10, Y: 4, Button: uv.MouseLeft}, .5, .5)
	r.Update(root)
	assert.Same(t, node, r.WidgetByID("image").node)
	assert.Same(t, record, r.images.records[key])
	assert.False(t, r.images.owns(1, 1, record))
	assert.True(t, r.images.owns(10, 5, record))
	r.cancelDrag()
	r.Update(root)
	assert.Same(t, record, r.images.records[key])
	assert.True(t, r.images.owns(1, 1, record))
	assert.False(t, r.images.owns(10, 5, record))
}
