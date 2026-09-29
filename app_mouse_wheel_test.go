package terma

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDispatchMouseWheel_HorizontalMovesScrollableOffsetX(t *testing.T) {
	state := NewScrollState()
	widget := Scrollable{
		ID:    "scrollable",
		State: state,
		Width: Cells(10),
		Child: Text{
			Content: strings.Repeat("x", 30),
			Width:   Cells(30),
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)
	state.updateHorizontalLayout(10, 30)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, state.GetOffsetX())

	handled = dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelLeft)
	require.True(t, handled)
	assert.Equal(t, 0, state.GetOffsetX())
}

func TestDispatchMouseWheel_HorizontalBubblesToOuterScrollable(t *testing.T) {
	outerState := NewScrollState()
	innerState := NewScrollState()
	widget := Scrollable{
		ID:    "outer",
		State: outerState,
		Width: Cells(10),
		Child: Scrollable{
			ID:    "inner",
			State: innerState,
			Width: Cells(20),
			Child: Text{
				Content: "short",
				Width:   Cells(5),
			},
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)
	innerState.updateHorizontalLayout(10, 10)
	outerState.updateHorizontalLayout(10, 30)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, outerState.GetOffsetX())
	assert.Equal(t, 0, innerState.GetOffsetX())
}

func TestDispatchMouseWheel_HorizontalUsesCallbacksWhenNoLayoutOverflow(t *testing.T) {
	state := NewScrollState()
	calls := 0
	state.OnScrollRight = func(cols int) bool {
		calls += cols
		return true
	}

	widget := Scrollable{
		ID:    "scrollable",
		State: state,
		Width: Cells(10),
		Child: Text{
			Content: "fits",
		},
	}

	buf := uv.NewBuffer(20, 4)
	focusManager := NewFocusManager()
	renderer := NewRenderer(buf, 20, 4, focusManager, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	renderer.Render(widget)

	handled := dispatchMouseWheel(renderer, 0, 0, uv.MouseWheelRight)
	require.True(t, handled)
	assert.Equal(t, 1, calls)
	assert.Equal(t, 0, state.GetOffsetX())
}

// wheelRecorder is a one-row tab bar with no scrollable content.
type wheelRecorder struct {
	id      string
	child   Widget
	events  []MouseEvent
	consume bool
	order   *[]string
}

func (w *wheelRecorder) WidgetID() string { return w.id }
func (w *wheelRecorder) Build(BuildContext) Widget {
	if w.child != nil {
		return w.child
	}
	return Text{Content: "One | Two | Three", Width: Cells(20), Height: Cells(1)}
}
func (w *wheelRecorder) OnMouseWheel(event MouseEvent) bool {
	w.events = append(w.events, event)
	if w.order != nil {
		*w.order = append(*w.order, w.id)
	}
	return w.consume
}

func TestMouseRouter_WheelOnPlainTabBar(t *testing.T) {
	tabs := &wheelRecorder{id: "tabs", consume: true}
	router, renderer := renderForMouse(Column{Children: []Widget{Text{Content: "header"}, Row{Children: []Widget{Text{Content: ".."}, tabs}}}}, 30, 4)
	require.Empty(t, renderer.ScrollablesAt(4, 1), "tab bar needs no artificial scroll extent")
	require.True(t, router.wheel(uv.MouseWheelEvent{X: 4, Y: 1, Button: uv.MouseWheelDown, Mod: uv.ModShift | uv.ModAlt}))
	require.Len(t, tabs.events, 1)
	assert.Equal(t, MouseEvent{X: 4, Y: 1, LocalX: 2, LocalY: 0, WidgetID: "tabs", Button: uv.MouseWheelDown, Mod: uv.ModShift | uv.ModAlt, SubCellX: 0.5, SubCellY: 0.5}, tabs.events[0])
}

func TestMouseRouter_WheelBubblesInnermostFirst(t *testing.T) {
	for _, consume := range []bool{false, true} {
		t.Run(fmt.Sprint(consume), func(t *testing.T) {
			var order []string
			parent := &wheelRecorder{id: "parent", consume: true, order: &order}
			child := &wheelRecorder{id: "child", consume: consume, order: &order}
			// The child declines or consumes; the parent should only run if declined.
			parent.child = Row{Children: []Widget{Text{Content: ".."}, child}}
			router, _ := renderForMouse(parent, 30, 4)
			require.True(t, router.wheelAt(uv.MouseWheelEvent{X: 4, Y: 0, Button: uv.MouseWheelLeft}, 0.25, 0.75))
			expected := []string{"child"}
			if !consume {
				expected = append(expected, "parent")
			}
			assert.Equal(t, expected, order)
			assert.Equal(t, 2, child.events[0].LocalX)
			assert.Equal(t, 0.25, child.events[0].SubCellX)
			assert.Equal(t, 0.75, child.events[0].SubCellY)
			if !consume {
				assert.Equal(t, 4, parent.events[0].LocalX)
			}
		})
	}
}

func TestMouseRouter_WheelDeclineFallsBackToScrolling(t *testing.T) {
	for _, consume := range []bool{false, true} {
		t.Run(fmt.Sprint(consume), func(t *testing.T) {
			state := NewScrollState()
			child := &wheelRecorder{id: "child", consume: consume, child: Text{Content: strings.Repeat("row\n", 10)}}
			outerCalls := 0
			root := Column{MouseWheel: func(MouseEvent) bool { outerCalls++; return true }, Children: []Widget{
				Scrollable{ID: "scroll", State: state, Height: Cells(3), Child: child},
			}}
			router, renderer := renderForMouse(root, 20, 4)
			for i := 0; i < 2; i++ {
				require.True(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
				renderer.Render(root) // Also exercises replayed registry entries.
			}
			assert.Len(t, child.events, 2)
			if consume {
				assert.Zero(t, state.GetOffset())
			} else {
				assert.Equal(t, 2, state.GetOffset())
			}
			assert.Zero(t, outerCalls, "a moving inner viewport consumes before ancestor handlers")
			state.SetOffset(100)
			renderer.Render(root)
			require.True(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 1, Button: uv.MouseWheelDown}))
			if !consume {
				assert.Equal(t, 1, outerCalls, "at the scroll limit the event bubbles out")
			}
		})
	}
}

func TestMouseRouter_WheelDoesNotBubbleToCoveredSibling(t *testing.T) {
	under := &wheelRecorder{id: "under", consume: true}
	over := &wheelRecorder{id: "over", consume: false}
	router, _ := renderForMouse(Stack{Children: []Widget{under, over}}, 20, 2)
	assert.False(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
	assert.Len(t, over.events, 1)
	assert.Empty(t, under.events)
}

func TestMouseRouter_WheelHandlersRespectLayersClippingAndDisabled(t *testing.T) {
	under := &wheelRecorder{id: "under", consume: true}
	overlay := &wheelRecorder{id: "overlay", consume: false}
	root := Column{Children: []Widget{under, Floating{Visible: true, Config: FloatConfig{Position: FloatPositionTopLeft}, Child: overlay}}}
	router, _ := renderForMouse(root, 30, 4)
	assert.False(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
	assert.Len(t, overlay.events, 1)
	assert.Empty(t, under.events)

	disabled := &wheelRecorder{id: "disabled", consume: true}
	router, _ = renderForMouse(DisabledWhen(true, disabled), 30, 4)
	assert.False(t, router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown}))
	assert.Empty(t, disabled.events)

	clipped := &wheelRecorder{id: "clipped", consume: true}
	state := NewScrollState()
	root = Column{Children: []Widget{Text{Content: "header"}, Scrollable{State: state, Height: Cells(2), Child: Column{Children: []Widget{clipped, Text{Content: "a\nb\nc"}}}}}}
	router, renderer := renderForMouse(root, 30, 4)
	state.SetOffset(1)
	renderer.Render(root)
	router.wheel(uv.MouseWheelEvent{X: 1, Y: 0, Button: uv.MouseWheelDown})
	assert.Empty(t, clipped.events, "a clipped row cannot steal the header's wheel")
}

func TestMouseRouter_WheelCallbackOnTextAndScrollable(t *testing.T) {
	state := NewScrollState()
	calls := 0
	root := Scrollable{State: state, Height: Cells(2), MouseWheel: func(MouseEvent) bool { calls++; return true }, Child: Text{Content: "a\nb\nc", MouseWheel: func(MouseEvent) bool { calls++; return false }}}
	router, _ := renderForMouse(root, 20, 3)
	require.True(t, router.wheel(uv.MouseWheelEvent{X: 0, Y: 0, Button: uv.MouseWheelDown}))
	assert.Equal(t, 2, calls)
	assert.Zero(t, state.GetOffset())
}
