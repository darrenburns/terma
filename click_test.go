package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Click Event Coordinate Tests
// =============================================================================

// clickRecorder records mouse events for testing.
type clickRecorder struct {
	clicks    []MouseEvent
	mouseDown []MouseEvent
	mouseUp   []MouseEvent
}

func (r *clickRecorder) onClick(event MouseEvent) {
	r.clicks = append(r.clicks, event)
}

func (r *clickRecorder) onMouseDown(event MouseEvent) {
	r.mouseDown = append(r.mouseDown, event)
}

func (r *clickRecorder) onMouseUp(event MouseEvent) {
	r.mouseUp = append(r.mouseUp, event)
}

func TestClick_EventContainsCorrectCoordinates(t *testing.T) {
	recorder := &clickRecorder{}

	widget := Text{
		ID:      "coord-test",
		Content: "Click here",
		Click:   recorder.onClick,
		Width:   Cells(20),
		Height:  Cells(3),
	}

	// Render the widget
	buf := uv.NewBuffer(30, 10)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 30, 10, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	// Find widget and simulate click at specific coordinates
	entry := renderer.WidgetAt(5, 1)
	require.NotNil(t, entry, "Widget should be found at (5, 1)")

	// Simulate click event
	event := MouseEvent{
		X:          5,
		Y:          1,
		Button:     uv.MouseLeft,
		ClickCount: 1,
		WidgetID:   entry.ID,
	}

	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}

	// Verify click was recorded with correct coordinates
	require.Len(t, recorder.clicks, 1)
	assert.Equal(t, 5, recorder.clicks[0].X)
	assert.Equal(t, 1, recorder.clicks[0].Y)
	assert.Equal(t, "coord-test", recorder.clicks[0].WidgetID)
}

func TestClick_EventContainsWidgetID(t *testing.T) {
	recorder := &clickRecorder{}

	widget := Text{
		ID:      "my-widget-id",
		Content: "Test",
		Click:   recorder.onClick,
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)
	assert.Equal(t, "my-widget-id", entry.ID)

	event := MouseEvent{
		X:        0,
		Y:        0,
		WidgetID: entry.ID,
	}

	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}

	require.Len(t, recorder.clicks, 1)
	assert.Equal(t, "my-widget-id", recorder.clicks[0].WidgetID)
}

// =============================================================================
// Widget Hit Detection Tests
// =============================================================================

func TestClick_WidgetAtFindsCorrectWidget(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{ID: "left", Content: "LEFT", Width: Cells(10)},
			Text{ID: "right", Content: "RIGHT", Width: Cells(10)},
		},
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	// Click in left widget area
	leftEntry := renderer.WidgetAt(2, 0)
	require.NotNil(t, leftEntry)
	assert.Equal(t, "left", leftEntry.ID)

	// Click in right widget area
	rightEntry := renderer.WidgetAt(12, 0)
	require.NotNil(t, rightEntry)
	assert.Equal(t, "right", rightEntry.ID)
}

func TestClick_WidgetAtReturnsNilOutsideBounds(t *testing.T) {
	widget := Text{
		ID:      "small",
		Content: "Hi",
		Width:   Cells(5),
		Height:  Cells(2),
	}

	buf := uv.NewBuffer(20, 10)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 10, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	// Click outside widget bounds
	entry := renderer.WidgetAt(15, 8)
	assert.Nil(t, entry, "Should return nil when clicking outside all widgets")
}

func TestClick_NestedWidgetsReturnsInnermost(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Row{
				Children: []Widget{
					Text{ID: "inner", Content: "Inner", Width: Cells(10)},
				},
			},
		},
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(2, 0)
	require.NotNil(t, entry)
	assert.Equal(t, "inner", entry.ID, "Should return innermost widget")
}

// =============================================================================
// Click Chain Tests (Single/Double/Triple Click)
// =============================================================================

func TestClick_ChainTracker_SingleClick(t *testing.T) {
	tracker := &mouseClickTracker{}

	count := tracker.nextClick("widget1", uv.MouseLeft, 0, 0, time.Now())
	assert.Equal(t, 1, count, "First click should have count 1")
}

func TestClick_ChainTracker_DoubleClick(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	count := tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(100*time.Millisecond))

	assert.Equal(t, 2, count, "Second quick click should have count 2")
}

func TestClick_ChainTracker_TripleClick(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(100*time.Millisecond))
	count := tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(200*time.Millisecond))

	assert.Equal(t, 3, count, "Third quick click should have count 3")
}

func TestClick_ChainTracker_TimeoutResetsChain(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	// Wait longer than clickChainTimeout (500ms)
	count := tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(600*time.Millisecond))

	assert.Equal(t, 1, count, "Click after timeout should reset to count 1")
}

func TestClick_ChainTracker_DifferentWidgetResetsChain(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	count := tracker.nextClick("widget2", uv.MouseLeft, 0, 0, now.Add(100*time.Millisecond))

	assert.Equal(t, 1, count, "Click on different widget should reset to count 1")
}

func TestClick_ChainTracker_DifferentButtonResetsChain(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	count := tracker.nextClick("widget1", uv.MouseRight, 0, 0, now.Add(100*time.Millisecond))

	assert.Equal(t, 1, count, "Click with different button should reset to count 1")
}

func TestClick_ChainTracker_DifferentPositionResetsChain(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	count := tracker.nextClick("widget1", uv.MouseLeft, 5, 0, now.Add(100*time.Millisecond))

	assert.Equal(t, 1, count, "Click at different position should reset to count 1")
}

func TestClick_ChainTracker_ReleaseCountMatchesDown(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	// Double click
	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(100*time.Millisecond))

	// Release should have same count as the down
	releaseCount := tracker.releaseCount("widget1", uv.MouseLeft)
	assert.Equal(t, 2, releaseCount, "Release count should match the click count")
}

func TestClick_ChainTracker_ReleaseOnDifferentWidgetReturns1(t *testing.T) {
	tracker := &mouseClickTracker{}
	now := time.Now()

	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now)
	tracker.nextClick("widget1", uv.MouseLeft, 0, 0, now.Add(100*time.Millisecond))

	// Release on different widget
	releaseCount := tracker.releaseCount("widget2", uv.MouseLeft)
	assert.Equal(t, 1, releaseCount, "Release on different widget should return 1")
}

// =============================================================================
// MouseDown/MouseUp Tests
// =============================================================================

func TestClick_MouseDownCalled(t *testing.T) {
	recorder := &clickRecorder{}

	widget := Text{
		ID:        "down-test",
		Content:   "Test",
		MouseDown: recorder.onMouseDown,
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)

	event := MouseEvent{X: 0, Y: 0, WidgetID: entry.ID}

	if handler, ok := entry.EventWidget.(MouseDownHandler); ok {
		handler.OnMouseDown(event)
	}

	require.Len(t, recorder.mouseDown, 1)
	assert.Equal(t, "down-test", recorder.mouseDown[0].WidgetID)
}

func TestClick_MouseUpCalled(t *testing.T) {
	recorder := &clickRecorder{}

	widget := Text{
		ID:      "up-test",
		Content: "Test",
		MouseUp: recorder.onMouseUp,
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)

	event := MouseEvent{X: 0, Y: 0, WidgetID: entry.ID}

	if handler, ok := entry.EventWidget.(MouseUpHandler); ok {
		handler.OnMouseUp(event)
	}

	require.Len(t, recorder.mouseUp, 1)
	assert.Equal(t, "up-test", recorder.mouseUp[0].WidgetID)
}

func TestClick_AllHandlersCalledInOrder(t *testing.T) {
	var order []string

	widget := Text{
		ID:        "order-test",
		Content:   "Test",
		MouseDown: func(e MouseEvent) { order = append(order, "down") },
		Click:     func(e MouseEvent) { order = append(order, "click") },
		MouseUp:   func(e MouseEvent) { order = append(order, "up") },
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)

	event := MouseEvent{X: 0, Y: 0, WidgetID: entry.ID}

	// Simulate click sequence as done in app.go
	if handler, ok := entry.EventWidget.(MouseDownHandler); ok {
		handler.OnMouseDown(event)
	}
	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}
	if handler, ok := entry.EventWidget.(MouseUpHandler); ok {
		handler.OnMouseUp(event)
	}

	assert.Equal(t, []string{"down", "click", "up"}, order)
}

// =============================================================================
// Button Click Tests
// =============================================================================

func TestClick_ButtonReceivesClickEvent(t *testing.T) {
	var received *MouseEvent

	widget := Button{
		ID:    "btn",
		Label: "Test",
		Click: func(e MouseEvent) {
			received = &e
		},
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetByID("btn")
	require.NotNil(t, entry)

	event := MouseEvent{
		X:          2,
		Y:          0,
		Button:     uv.MouseLeft,
		ClickCount: 1,
		WidgetID:   "btn",
	}

	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}

	require.NotNil(t, received)
	assert.Equal(t, 2, received.X)
	assert.Equal(t, uv.MouseLeft, received.Button)
	assert.Equal(t, 1, received.ClickCount)
}

// =============================================================================
// Click Count in Event Tests
// =============================================================================

func TestClick_EventClickCountPassedCorrectly(t *testing.T) {
	var receivedCount int

	widget := Text{
		ID:      "count-test",
		Content: "Test",
		Click: func(e MouseEvent) {
			receivedCount = e.ClickCount
		},
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)

	// Simulate triple click
	event := MouseEvent{
		X:          0,
		Y:          0,
		ClickCount: 3,
		WidgetID:   entry.ID,
	}

	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}

	assert.Equal(t, 3, receivedCount, "Click count should be passed to handler")
}

// =============================================================================
// Mouse Button Tests
// =============================================================================

func TestClick_MouseButtonPassedCorrectly(t *testing.T) {
	var receivedButton uv.MouseButton

	widget := Text{
		ID:      "button-test",
		Content: "Test",
		Click: func(e MouseEvent) {
			receivedButton = e.Button
		},
	}

	buf := uv.NewBuffer(20, 5)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	renderer := NewRenderer(buf, 20, 5, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	entry := renderer.WidgetAt(0, 0)
	require.NotNil(t, entry)

	event := MouseEvent{
		X:      0,
		Y:      0,
		Button: uv.MouseRight,
	}

	if clickable, ok := entry.EventWidget.(Clickable); ok {
		clickable.OnClick(event)
	}

	assert.Equal(t, uv.MouseRight, receivedButton)
}
