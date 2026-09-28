package terma

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mouseRecorder is a sized widget that records the mouse events it receives.
type mouseRecorder struct {
	id    string
	width int
	downs []MouseEvent
	moves []MouseEvent
	ups   []MouseEvent
}

func (m *mouseRecorder) WidgetID() string { return m.id }
func (m *mouseRecorder) Build(BuildContext) Widget {
	return Text{Content: m.id, Width: Cells(m.width)}
}
func (m *mouseRecorder) OnMouseDown(event MouseEvent) { m.downs = append(m.downs, event) }
func (m *mouseRecorder) OnMouseMove(event MouseEvent) { m.moves = append(m.moves, event) }
func (m *mouseRecorder) OnMouseUp(event MouseEvent)   { m.ups = append(m.ups, event) }

// renderForMouse renders root and returns a router over the frame.
func renderForMouse(root Widget, width, height int) (*mouseRouter, *Renderer) {
	focusManager := NewFocusManager()
	focusManager.SetRootWidget(root)
	hovered := NewAnySignal[Widget](nil)
	renderer := NewRenderer(uv.NewBuffer(width, height), width, height, focusManager, NewAnySignal[Focusable](nil), hovered)
	focusManager.SetFocusables(renderer.Render(root))
	return newMouseRouter(renderer, focusManager, hovered), renderer
}

func hitID(renderer *Renderer, x, y int) string {
	if entry := renderer.WidgetAt(x, y); entry != nil {
		return entry.ID
	}
	return ""
}

func TestMouseHitTest_WidgetsScrolledOutOfViewAreNotHit(t *testing.T) {
	scroll := NewScrollState()
	rows := make([]Widget, 10)
	for i := range rows {
		rows[i] = Text{ID: fmt.Sprintf("row-%d", i), Content: fmt.Sprintf("row %d", i)}
	}
	root := Column{Children: []Widget{
		Text{ID: "header", Content: "HEADER", Width: Flex(1)},
		Scrollable{ID: "scroller", State: scroll, Height: Cells(3), Child: Column{Children: rows}},
		Text{ID: "footer", Content: "FOOTER", Width: Flex(1)},
	}}
	_, renderer := renderForMouse(root, 20, 6)
	scroll.SetOffset(5)
	renderer.Render(root)

	// Rows 0-4 now sit above the viewport, over the header; rows 8-9 below
	// it, over the footer and the empty line. Neither may take the pointer.
	assert.Equal(t, "header", hitID(renderer, 1, 0))
	assert.Equal(t, "row-5", hitID(renderer, 1, 1))
	assert.Equal(t, "row-7", hitID(renderer, 1, 3))
	assert.Equal(t, "footer", hitID(renderer, 1, 4))
	assert.Equal(t, "", hitID(renderer, 1, 5))
}

func TestMouseHitTest_PartlyScrolledWidgetUsesFullBoundsForLocalCoordinates(t *testing.T) {
	scroll := NewScrollState()
	area := &tallRecorder{id: "area", height: 3}
	root := Scrollable{State: scroll, Height: Cells(2), Child: Column{Children: []Widget{area, Text{Content: "after"}}}}
	router, renderer := renderForMouse(root, 20, 2)
	scroll.SetOffset(1)
	renderer.Render(root)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)
	require.Equal(t, "area", entry.ID)
	assert.Equal(t, Rect{X: 0, Y: -1, Width: 6, Height: 3}, entry.Bounds)
	assert.Equal(t, Rect{X: 0, Y: 0, Width: 6, Height: 2}, entry.Visible)

	router.press(uv.MouseClickEvent{X: 2, Y: 0, Button: uv.MouseLeft}, time.Now())
	require.Len(t, area.downs, 1)
	assert.Equal(t, 2, area.downs[0].LocalX)
	assert.Equal(t, 1, area.downs[0].LocalY, "local coordinates count the scrolled-away row")
}

// tallRecorder is a multi-row widget that records presses.
type tallRecorder struct {
	id     string
	height int
	downs  []MouseEvent
}

func (m *tallRecorder) WidgetID() string { return m.id }
func (m *tallRecorder) Build(BuildContext) Widget {
	return Text{Content: m.id, Width: Cells(6), Height: Cells(m.height)}
}
func (m *tallRecorder) OnMouseDown(event MouseEvent) { m.downs = append(m.downs, event) }

func TestMouseHitTest_OverlayHidesFocusablesBeneathIt(t *testing.T) {
	input := NewTextInputState("")
	root := Column{Children: []Widget{
		TextInput{ID: "input", State: input, Width: Cells(20)},
		Button{ID: "other", Label: "other"},
		Floating{
			Visible: true,
			Config:  FloatConfig{Position: FloatPositionTopLeft},
			Child:   Text{ID: "tooltip", Content: "a tooltip"},
		},
	}}
	router, renderer := renderForMouse(root, 30, 4)
	router.focusManager.FocusByID("other")

	assert.Equal(t, "tooltip", hitID(renderer, 1, 0))
	assert.Nil(t, renderer.FocusableAt(1, 0), "the input is under the overlay")
	require.NotNil(t, renderer.FocusableAt(15, 0), "beside the overlay the input is reachable")

	router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, time.Now())
	assert.Equal(t, "other", router.focusManager.FocusedID())
}

func TestMouseRouter_ReleaseGoesToThePressedWidget(t *testing.T) {
	a := &mouseRecorder{id: "a", width: 5}
	b := &mouseRecorder{id: "b", width: 5}
	router, _ := renderForMouse(Row{Children: []Widget{a, b}}, 20, 1)

	router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, time.Now())
	router.motion(uv.MouseMotionEvent{X: 7, Y: 0, Button: uv.MouseLeft})
	router.release(uv.MouseReleaseEvent{X: 8, Y: 0, Button: uv.MouseLeft})

	require.Len(t, a.moves, 1)
	assert.Equal(t, 7, a.moves[0].LocalX)
	require.Len(t, a.ups, 1)
	assert.Equal(t, 8, a.ups[0].X)
	assert.Equal(t, 8, a.ups[0].LocalX, "local to the pressed widget, even outside it")
	assert.Empty(t, b.moves)
	assert.Empty(t, b.ups)
}

func TestMouseRouter_MotionWithoutButtonsEndsALostDrag(t *testing.T) {
	a := &mouseRecorder{id: "a", width: 5}
	router, _ := renderForMouse(Row{Children: []Widget{a}}, 20, 1)

	router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, time.Now())
	router.motion(uv.MouseMotionEvent{X: 2, Y: 0, Button: uv.MouseLeft})
	// The release happened outside the window and was never reported.
	assert.True(t, router.motion(uv.MouseMotionEvent{X: 3, Y: 0}))
	require.Len(t, a.ups, 1)
	assert.Equal(t, uv.MouseLeft, a.ups[0].Button)

	router.motion(uv.MouseMotionEvent{X: 4, Y: 0, Button: uv.MouseLeft})
	assert.Len(t, a.moves, 1, "the drag has ended")
}

func TestMouseRouter_SplitPaneDragReleasedOverAPaneEnds(t *testing.T) {
	state := NewSplitPaneState(0.5)
	pane := SplitPane{
		ID:     "split",
		State:  state,
		First:  Text{ID: "left", Content: "left"},
		Second: Text{ID: "right", Content: "right", Width: Flex(1)},
	}
	router, _ := renderForMouse(pane, 21, 3)
	dividerX := computeSplitPaneMetrics(21, pane.dividerSize(), pane.minPaneSize(), state.GetPosition()).offset

	router.press(uv.MouseClickEvent{X: dividerX, Y: 1, Button: uv.MouseLeft}, time.Now())
	require.True(t, state.dragging)
	router.motion(uv.MouseMotionEvent{X: dividerX + 4, Y: 1, Button: uv.MouseLeft})
	router.release(uv.MouseReleaseEvent{X: dividerX + 4, Y: 1, Button: uv.MouseLeft})

	assert.False(t, state.dragging, "the divider stays highlighted if its release is lost")
	assert.Greater(t, state.GetPosition(), 0.5)
}

func TestMouseRouter_ModalBlocksWheelAndBackdropPresses(t *testing.T) {
	scroll := NewScrollState()
	under := &mouseRecorder{id: "under", width: 10}
	rows := make([]Widget, 10)
	for i := range rows {
		rows[i] = Text{Content: fmt.Sprintf("row %d", i)}
	}
	root := Column{Children: []Widget{
		under,
		Scrollable{State: scroll, Height: Cells(4), Child: Column{Children: rows}},
		Floating{
			Visible: true,
			Config:  FloatConfig{Modal: true, Position: FloatPositionBottomRight},
			Child:   Text{Content: "modal"},
		},
	}}
	router, _ := renderForMouse(root, 20, 6)

	assert.False(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 2, Button: uv.MouseWheelDown}))
	assert.Equal(t, 0, scroll.GetOffset())

	router.press(uv.MouseClickEvent{X: 1, Y: 0, Button: uv.MouseLeft}, time.Now())
	router.release(uv.MouseReleaseEvent{X: 1, Y: 0, Button: uv.MouseLeft})
	assert.Empty(t, under.downs)
	assert.Empty(t, under.ups, "a release after a blocked press goes nowhere")
}

func TestMouseRouter_WheelOverAnOverlayDoesNotScrollBeneathIt(t *testing.T) {
	scroll := NewScrollState()
	rows := make([]Widget, 10)
	for i := range rows {
		rows[i] = Text{Content: fmt.Sprintf("row %d", i)}
	}
	root := Column{Children: []Widget{
		Scrollable{State: scroll, Height: Cells(4), Child: Column{Children: rows}},
		Floating{
			Visible: true,
			Config:  FloatConfig{Position: FloatPositionTopLeft},
			Child:   Text{Content: "menu"},
		},
	}}
	router, _ := renderForMouse(root, 20, 4)

	assert.False(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
	assert.Equal(t, 0, scroll.GetOffset())

	assert.True(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 2, Button: uv.MouseWheelDown}))
	assert.Equal(t, 1, scroll.GetOffset())
}

func TestCoalesceMouseMotion(t *testing.T) {
	events := make(chan uv.Event, 8)
	events <- uv.MouseMotionEvent{X: 2, Button: uv.MouseLeft}
	events <- uv.MouseMotionEvent{X: 3, Button: uv.MouseLeft}
	events <- uv.MouseReleaseEvent{X: 3, Button: uv.MouseLeft}
	events <- uv.MouseMotionEvent{X: 4}

	latest, next := coalesceMouseMotion(uv.MouseMotionEvent{X: 1, Button: uv.MouseLeft}, events)
	assert.Equal(t, uv.MouseMotionEvent{X: 3, Button: uv.MouseLeft}, latest)
	assert.Equal(t, uv.MouseReleaseEvent{X: 3, Button: uv.MouseLeft}, next)

	// Motion with different buttons isn't folded.
	latest, next = coalesceMouseMotion(uv.MouseMotionEvent{X: 1, Button: uv.MouseLeft}, events)
	assert.Equal(t, uv.MouseMotionEvent{X: 1, Button: uv.MouseLeft}, latest)
	assert.Equal(t, uv.MouseMotionEvent{X: 4}, next)

	latest, next = coalesceMouseMotion(uv.MouseMotionEvent{X: 5}, events)
	assert.Equal(t, uv.MouseMotionEvent{X: 5}, latest)
	assert.Nil(t, next)
}

type hoverReader struct{ id string }

func (h hoverReader) WidgetID() string { return h.id }
func (h hoverReader) Build(ctx BuildContext) Widget {
	label := h.id
	if ctx.IsHovered(h) {
		label += "*"
	}
	return Text{Content: label}
}

func TestIsHovered_OnlyRebuildsWidgetsWhoseHoverChanged(t *testing.T) {
	readers := make([]Widget, 10)
	for i := range readers {
		readers[i] = hoverReader{id: fmt.Sprintf("r%d", i)}
	}
	root := Column{Children: readers}
	_, renderer := renderForMouse(root, 10, 10)

	renderer.hoveredSignal.Set(readers[3])
	renderer.Update(root)
	assert.Equal(t, 1, renderer.Stats().BuildCount)

	renderer.hoveredSignal.Set(readers[4])
	renderer.Update(root)
	assert.Equal(t, 2, renderer.Stats().BuildCount)
	assert.True(t, strings.Contains(renderer.ScreenText(), "r4*"))
	assert.False(t, strings.Contains(renderer.ScreenText(), "r3*"))
}

func BenchmarkMouseHitTest_LargeScrolledContent(b *testing.B) {
	for _, rows := range []int{100, 5000} {
		scroll := NewScrollState()
		children := make([]Widget, rows)
		for i := range children {
			children[i] = Row{Children: []Widget{
				Text{Content: fmt.Sprintf("row %d", i)},
				Button{ID: fmt.Sprintf("btn-%d", i), Label: "x"},
			}}
		}
		root := Scrollable{State: scroll, Height: Flex(1), Child: Column{Children: children}}
		_, renderer := renderForMouse(root, 120, 40)
		scroll.SetOffset(rows / 2)
		renderer.Render(root)

		b.Run(fmt.Sprintf("rows=%d/WidgetAt", rows), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = renderer.WidgetAt(i%120, i%40)
			}
		})
		b.Run(fmt.Sprintf("rows=%d/FirstHitAfterLayout", rows), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				renderer.widgetRegistry.rowsValid = false
				_ = renderer.WidgetAt(i%120, i%40)
			}
		})
	}
}
