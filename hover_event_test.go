package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hoverEventRecorderWidget struct {
	id     string
	events []HoverEvent
}

func (w *hoverEventRecorderWidget) Build(ctx BuildContext) Widget { return EmptyWidget{} }
func (w *hoverEventRecorderWidget) WidgetID() string              { return w.id }
func (w *hoverEventRecorderWidget) OnHover(event HoverEvent) {
	w.events = append(w.events, event)
}

type hoverOrderWidget struct {
	id    string
	order *[]string
}

func (w *hoverOrderWidget) Build(ctx BuildContext) Widget { return EmptyWidget{} }
func (w *hoverOrderWidget) WidgetID() string              { return w.id }
func (w *hoverOrderWidget) OnHover(event HoverEvent) {
	label := "enter"
	if event.Type == HoverLeave {
		label = "leave"
	}
	*w.order = append(*w.order, label+":"+w.id)
}

type hoverNoopEventWidget struct {
	id string
}

func (w *hoverNoopEventWidget) Build(ctx BuildContext) Widget { return EmptyWidget{} }
func (w *hoverNoopEventWidget) WidgetID() string              { return w.id }
func (w *hoverNoopEventWidget) OnHover(event HoverEvent)      {}

func makeHoverEntry(id string, widget Widget, x, y, width, height int) *WidgetEntry {
	return &WidgetEntry{
		ID:          id,
		EventWidget: widget,
		Bounds: Rect{
			X:      x,
			Y:      y,
			Width:  width,
			Height: height,
		},
	}
}

func TestHoverEvent_NoneToWidget_EmitsEnter(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	widget := &hoverEventRecorderWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 1, 1, 5, 3)

	changed := tracker.UpdatePointer(2, 3, uv.ModShift, uv.MouseLeft, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)

	require.True(t, changed)
	require.Len(t, widget.events, 1)
	event := widget.events[0]
	assert.Equal(t, HoverEnter, event.Type)
	assert.Equal(t, "a", event.WidgetID)
	assert.Equal(t, "", event.PreviousWidgetID)
	assert.Equal(t, "a", event.NextWidgetID)
	assert.Equal(t, 1, event.LocalX)
	assert.Equal(t, 2, event.LocalY)
	assert.Equal(t, uv.ModShift, event.Mod)
	assert.Equal(t, uv.MouseLeft, event.Button)
	assert.Equal(t, Widget(widget), hoveredSignal.Get())
}

func TestHoverEvent_WidgetToNone_EmitsLeave(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	widget := &hoverEventRecorderWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 0, 0, 10, 2)

	tracker.UpdatePointer(1, 1, 0, uv.MouseLeft, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)
	changed := tracker.UpdatePointer(20, 20, 0, uv.MouseLeft, func(x, y int) *WidgetEntry {
		return nil
	}, hoveredSignal)

	require.True(t, changed)
	require.Len(t, widget.events, 2)
	event := widget.events[1]
	assert.Equal(t, HoverLeave, event.Type)
	assert.Equal(t, "a", event.WidgetID)
	assert.Equal(t, "a", event.PreviousWidgetID)
	assert.Equal(t, "", event.NextWidgetID)
	assert.Nil(t, hoveredSignal.Get())
}

func TestHoverEvent_WidgetToWidget_LeaveBeforeEnter(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	order := []string{}
	a := &hoverOrderWidget{id: "a", order: &order}
	b := &hoverOrderWidget{id: "b", order: &order}
	entryA := makeHoverEntry("a", a, 0, 0, 10, 2)
	entryB := makeHoverEntry("b", b, 10, 0, 10, 2)

	tracker.UpdatePointer(1, 0, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entryA
	}, hoveredSignal)
	order = order[:0]

	changed := tracker.UpdatePointer(11, 0, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entryB
	}, hoveredSignal)

	require.True(t, changed)
	assert.Equal(t, []string{"leave:a", "enter:b"}, order)
	assert.Equal(t, Widget(b), hoveredSignal.Get())
}

func TestHoverEvent_MoveInsideSameWidget_NoExtraEvents(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	widget := &hoverEventRecorderWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 0, 0, 20, 5)

	tracker.UpdatePointer(1, 1, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)
	changed := tracker.UpdatePointer(2, 2, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)

	assert.False(t, changed)
	require.Len(t, widget.events, 1)
	assert.Equal(t, HoverEnter, widget.events[0].Type)
}

func TestHoverEvent_Reconcile_EmitsLeaveWhenTargetRemoved(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	widget := &hoverEventRecorderWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 0, 0, 8, 2)

	tracker.UpdatePointer(1, 1, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)
	changed := tracker.Reconcile(func(x, y int) *WidgetEntry {
		return nil
	}, hoveredSignal)

	require.True(t, changed)
	require.Len(t, widget.events, 2)
	assert.Equal(t, HoverLeave, widget.events[1].Type)
	assert.Nil(t, hoveredSignal.Get())
}

func TestHoverEvent_Reconcile_EmitsTransitionWhenTargetChanges(t *testing.T) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	order := []string{}
	a := &hoverOrderWidget{id: "a", order: &order}
	b := &hoverOrderWidget{id: "b", order: &order}
	entryA := makeHoverEntry("a", a, 0, 0, 10, 2)
	entryB := makeHoverEntry("b", b, 0, 0, 10, 2)

	tracker.UpdatePointer(1, 0, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entryA
	}, hoveredSignal)
	order = order[:0]

	changed := tracker.Reconcile(func(x, y int) *WidgetEntry {
		return entryB
	}, hoveredSignal)

	require.True(t, changed)
	assert.Equal(t, []string{"leave:a", "enter:b"}, order)
	assert.Equal(t, Widget(b), hoveredSignal.Get())
}

func TestHoverEvent_BuiltInButtonHoverCallbackReceivesEvents(t *testing.T) {
	var events []HoverEvent
	button := Button{
		Hover: func(event HoverEvent) {
			events = append(events, event)
		},
	}

	dispatchHoverEvent(button, HoverEvent{Type: HoverEnter, WidgetID: "btn", NextWidgetID: "btn"})
	dispatchHoverEvent(button, HoverEvent{Type: HoverLeave, WidgetID: "btn", PreviousWidgetID: "btn"})

	require.Len(t, events, 2)
	assert.Equal(t, HoverEnter, events[0].Type)
	assert.Equal(t, HoverLeave, events[1].Type)
}

func TestHoverEvent_BuildContextHoveredSignalTracksTransitions(t *testing.T) {
	tracker := &hoverTracker{}
	fm := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	ctx := NewBuildContext(fm, focusedSignal, hoveredSignal, nil)

	widget := &hoverEventRecorderWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 0, 0, 3, 1)

	tracker.UpdatePointer(1, 0, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return entry
	}, hoveredSignal)
	assert.Equal(t, Widget(widget), ctx.Hovered())
	assert.Equal(t, "a", ctx.HoveredID())

	tracker.UpdatePointer(10, 0, 0, uv.MouseNone, func(x, y int) *WidgetEntry {
		return nil
	}, hoveredSignal)
	assert.Nil(t, ctx.Hovered())
	assert.Equal(t, "", ctx.HoveredID())
}

func BenchmarkHoverTracker_NoTransition(b *testing.B) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	widget := &hoverNoopEventWidget{id: "a"}
	entry := makeHoverEntry("a", widget, 0, 0, 10, 2)
	resolve := func(x, y int) *WidgetEntry { return entry }

	tracker.UpdatePointer(1, 1, 0, uv.MouseNone, resolve, hoveredSignal)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.UpdatePointer(1+(i%3), 1, 0, uv.MouseNone, resolve, hoveredSignal)
	}
}

func BenchmarkHoverTracker_RapidTransitions(b *testing.B) {
	tracker := &hoverTracker{}
	hoveredSignal := NewAnySignal[Widget](nil)
	a := &hoverNoopEventWidget{id: "a"}
	c := &hoverNoopEventWidget{id: "b"}
	entryA := makeHoverEntry("a", a, 0, 0, 1, 1)
	entryB := makeHoverEntry("b", c, 1, 0, 1, 1)
	resolve := func(x, y int) *WidgetEntry {
		if x%2 == 0 {
			return entryA
		}
		return entryB
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tracker.UpdatePointer(i%2, 0, 0, uv.MouseNone, resolve, hoveredSignal)
	}
}
