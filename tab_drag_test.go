package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tabKeys(tabs []Tab) []string {
	keys := make([]string, len(tabs))
	for i, tab := range tabs {
		keys[i] = tab.Key
	}
	return keys
}

func TestTabDragThresholds(t *testing.T) {
	for a := 1; a <= 20; a++ {
		for b := 1; b <= 20; b++ {
			p := tabDragPreview{items: []tabDragItem{{key: "a", width: float64(a)}, {key: "b", width: float64(b)}}, index: 0}
			midpoint := float64(b) / 2
			p.advance(midpoint, 1)
			require.Equal(t, 0, p.index, "equal midpoint, widths %d/%d", a, b)
			p.advance(midpoint+.001, 1)
			require.Equal(t, 1, p.index)
			p.advance(midpoint, -1)
			require.Equal(t, 1, p.index, "reversal equality")
			p.advance(midpoint-.001, -1)
			require.Equal(t, 0, p.index)
		}
	}
	p := tabDragPreview{items: []tabDragItem{{"a", 3}, {"b", 13}, {"c", 2}, {"d", 7}}, index: 0}
	p.advance(40, 40)
	assert.Equal(t, 3, p.index)
	p.advance(-10, -50)
	assert.Equal(t, 0, p.index)
}

func TestTabDragPreviewsThenCommitsUsingOnlyFinalX(t *testing.T) {
	for _, releaseY := range []int{-3, 0, 8} {
		t.Run(string(rune('a'+releaseY+3)), func(t *testing.T) {
			state := NewTabState([]Tab{{Key: "a", Label: "A"}, {Key: "b", Label: "Longer"}, {Key: "c", Label: "CC"}, {Key: "d", Label: "End"}})
			bar := TabBar{ID: "tabs", State: state, AllowReorder: true, TabStyle: Style{Padding: EdgeInsetsXY(1, 0)}, ActiveTabStyle: Style{Padding: EdgeInsetsXY(1, 0)}}
			m, r := renderForMouse(bar, 60, 10)
			source := r.WidgetByID(tabHeaderID("tabs", "a")).node
			x := r.WidgetByID(bar.TabID("a")).Bounds.X
			m.press(uv.MouseClickEvent{X: x, Button: uv.MouseLeft}, .2, .5, time.Now())
			m.motion(uv.MouseMotionEvent{X: x + 5, Y: 6, Button: uv.MouseLeft}, .2, .5)
			r.Update(bar)
			require.NotNil(t, state.dragPreview)
			assert.Equal(t, 1, state.dragPreview.index)
			assert.Equal(t, []string{"a", "b", "c", "d"}, tabKeys(state.TabsPeek()))
			assert.Same(t, source, r.WidgetByID(tabHeaderID("tabs", "a")).node)
			_, before := state.tabs.peekWithRevision()
			m.release(uv.MouseReleaseEvent{X: 45, Y: releaseY, Button: uv.MouseLeft}, .2, .5)
			r.Update(bar)
			_, after := state.tabs.peekWithRevision()
			assert.Equal(t, before+1, after, "commit changes the model exactly once")
			assert.Equal(t, []string{"b", "c", "d", "a"}, tabKeys(state.TabsPeek()))
			assert.Equal(t, "a", state.ActiveKeyPeek())
			assert.Nil(t, state.dragPreview)
			assert.Nil(t, r.drag)
		})
	}
}

func TestTabDragFreezesGeometryBeforeActivation(t *testing.T) {
	state := NewTabState([]Tab{{Key: "a", Label: "A"}, {Key: "b", Label: "BBBB"}, {Key: "c", Label: "C"}})
	bar := TabBar{ID: "tabs", State: state, AllowReorder: true, TabStyle: Style{Padding: EdgeInsetsXY(1, 0), Margin: EdgeInsetsXY(1, 0)}, ActiveTabStyle: Style{Padding: EdgeInsetsXY(5, 0), Margin: EdgeInsetsXY(2, 0)}}
	m, r := renderForMouse(bar, 80, 10)
	aWidth := r.WidgetByID(tabHeaderID("tabs", "a")).Bounds.Width
	bBounds := r.WidgetByID(tabHeaderID("tabs", "b")).Bounds
	x := r.WidgetByID(bar.TabID("b")).Bounds.X + 3
	m.press(uv.MouseClickEvent{X: x, Button: uv.MouseLeft}, .8, .2, time.Now())
	r.Update(bar)
	assert.Equal(t, "b", state.ActiveKeyPeek())
	assert.Equal(t, aWidth, r.WidgetByID(tabHeaderID("tabs", "a")).Bounds.Width)
	assert.Equal(t, bBounds, r.WidgetByID(tabHeaderID("tabs", "b")).Bounds)
	m.motion(uv.MouseMotionEvent{X: x + 2, Y: 3, Button: uv.MouseLeft}, .8, .2)
	r.Update(bar)
	assert.Equal(t, bBounds.Width, r.WidgetByID(tabHeaderID("tabs", "b")).Bounds.Width)
	assert.Equal(t, bBounds.X+2, r.WidgetByID(tabHeaderID("tabs", "b")).Bounds.X)
	r.cancelDrag()
	r.Update(bar)
	assert.Equal(t, []string{"a", "b", "c"}, tabKeys(state.TabsPeek()))
}

func TestTabDragClippedHeadersAndExternalChanges(t *testing.T) {
	state := NewTabState([]Tab{{Key: "a", Label: "A long first header"}, {Key: "b", Label: "Another long header"}, {Key: "c", Label: "Offscreen header"}})
	bar := TabBar{ID: "tabs", State: state, AllowReorder: true}
	m, r := renderForMouse(bar, 30, 4)
	require.Nil(t, r.WidgetByID(tabHeaderID("tabs", "c")))
	m.press(uv.MouseClickEvent{X: 2, Button: uv.MouseLeft}, .5, .5, time.Now())
	require.NotNil(t, r.drag, "an offscreen sibling must not disable visible handles")
	m.motion(uv.MouseMotionEvent{X: 26, Button: uv.MouseLeft}, .5, .5)
	r.Update(bar)
	state.RemoveTab("b")
	m.release(uv.MouseReleaseEvent{X: 29, Y: 9, Button: uv.MouseLeft}, .5, .5)
	r.Update(bar)
	assert.Nil(t, r.drag)
	assert.Equal(t, []string{"a", "c"}, tabKeys(state.TabsPeek()), "cancellation must not roll back external changes")
}

func TestTabDragSurvivesUnchangedLabelWrites(t *testing.T) {
	state := NewTabState([]Tab{{Key: "a", Label: "A"}, {Key: "b", Label: "B"}, {Key: "c", Label: "C"}})
	bar := TabBar{ID: "tabs", State: state, AllowReorder: true}
	m, r := renderForMouse(bar, 60, 4)
	x := r.WidgetByID(bar.TabID("a")).Bounds.X
	m.press(uv.MouseClickEvent{X: x, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.motion(uv.MouseMotionEvent{X: x + 20, Button: uv.MouseLeft}, .5, .5)
	r.Update(bar)
	for _, tab := range state.TabsPeek() {
		state.SetLabel(tab.Key, tab.Label)
	}
	state.SetLabel("missing", "Label")
	m.release(uv.MouseReleaseEvent{X: x + 20, Button: uv.MouseLeft}, .5, .5)
	r.Update(bar)
	assert.Equal(t, []string{"b", "c", "a"}, tabKeys(state.TabsPeek()))
}

type toggledTabDrag struct {
	state   *TabState
	enabled Signal[bool]
}

func (a *toggledTabDrag) Build(BuildContext) Widget {
	return TabBar{ID: "tabs", State: a.state, AllowReorder: a.enabled.Get()}
}

func TestTabDragKeysAndCapabilityChange(t *testing.T) {
	a := &toggledTabDrag{state: NewTabState([]Tab{{Key: "a", Label: "A"}, {Key: "a-body", Label: "B"}, {Key: "a-label", Label: "C"}}), enabled: NewSignal(true)}
	m, r := renderForMouse(a, 40, 8)
	x := r.WidgetByID((TabBar{ID: "tabs"}).TabID("a-body")).Bounds.X
	m.press(uv.MouseClickEvent{X: x, Button: uv.MouseLeft}, .5, .5, time.Now())
	require.NotNil(t, r.drag)
	assert.Equal(t, tabHeaderID("tabs", "a-body"), r.drag.source.eventID)
	m.motion(uv.MouseMotionEvent{X: 30, Button: uv.MouseLeft}, .5, .5)
	r.Update(a)
	a.enabled.Set(false)
	m.release(uv.MouseReleaseEvent{X: 30, Y: 5, Button: uv.MouseLeft}, .5, .5)
	r.Update(a)
	assert.Nil(t, r.drag)
	assert.Equal(t, []string{"a", "a-body", "a-label"}, tabKeys(a.state.TabsPeek()))
}

func TestTabCloseImmediatelyAfterDrag(t *testing.T) {
	state := NewTabState([]Tab{{Key: "home", Label: "Home"}, {Key: "counter", Label: "Counter"}, {Key: "list", Label: "List"}, {Key: "info", Label: "Info"}})
	bar := TabBar{ID: "tabs", State: state, AllowReorder: true, Closable: true, TabStyle: Style{Padding: EdgeInsetsXY(1, 0)}, ActiveTabStyle: Style{Padding: EdgeInsetsXY(1, 0)}}
	m, r := renderForMouse(bar, 80, 12)
	frame := func() {
		r.focusedSignal.Set(m.focusManager.Focused())
		m.focusManager.SetFocusables(r.Update(bar))
		if m.reconcileHover() {
			r.Update(bar)
		}
	}
	x := r.WidgetByID(bar.TabID("counter")).Bounds.X + 2
	m.motion(uv.MouseMotionEvent{X: x, Button: uv.MouseNone}, .5, .5)
	frame()
	m.press(uv.MouseClickEvent{X: x, Button: uv.MouseLeft}, .5, .5, time.Now())
	frame()
	m.motion(uv.MouseMotionEvent{X: 25, Y: 3, Button: uv.MouseLeft}, .5, .5)
	frame()
	m.motion(uv.MouseMotionEvent{X: 55, Y: 9, Button: uv.MouseLeft}, .5, .5)
	frame()
	m.release(uv.MouseReleaseEvent{X: 55, Y: 9, Button: uv.MouseLeft}, .5, .5)
	frame()
	require.Equal(t, []string{"home", "list", "info", "counter"}, tabKeys(state.TabsPeek()))
	header := r.WidgetByID(tabHeaderID("tabs", "counter")).Bounds
	closeX := header.X + header.Width - 2
	m.motion(uv.MouseMotionEvent{X: closeX, Button: uv.MouseNone}, .5, .5)
	frame()
	require.Equal(t, "×", r.terminal.CellAt(closeX, 0).Content)
	entry := r.WidgetAt(closeX, 0)
	require.NotNil(t, entry)
	closeLabel, ok := entry.EventWidget.(Text)
	require.True(t, ok)
	assert.Equal(t, "×", closeLabel.Content)
	m.press(uv.MouseClickEvent{X: closeX, Button: uv.MouseLeft}, .5, .5, time.Now())
	m.release(uv.MouseReleaseEvent{X: closeX, Button: uv.MouseLeft}, .5, .5)
	r.Update(bar)
	assert.Equal(t, []string{"home", "list", "info"}, tabKeys(state.TabsPeek()))
}
