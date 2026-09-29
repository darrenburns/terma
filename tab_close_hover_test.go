package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tabClosePositions(t *testing.T, scene *clickScene) []int {
	t.Helper()
	var positions []int
	for x := 0; x < scene.width; x++ {
		if cell := scene.buf.CellAt(x, 0); cell != nil && cell.Content == "×" {
			positions = append(positions, x)
		}
	}
	return positions
}

func TestTabCloseHover_IndependentFeedback(t *testing.T) {
	for _, id := range []string{"tabs", ""} {
		t.Run("id="+id, func(t *testing.T) {
			state := NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})
			scene := newClickScene(t, TabBar{ID: id, State: state, Closable: true}, 40, 2)
			positions := tabClosePositions(t, scene)
			require.Len(t, positions, 2)
			activeX, inactiveX := positions[0], positions[1]
			activeBase, inactiveBase := scene.bgAt(activeX, 0), scene.bgAt(inactiveX, 0)
			activeForeground := FromANSI(scene.buf.CellAt(activeX, 0).Style.Fg)
			inactiveForeground := FromANSI(scene.buf.CellAt(inactiveX, 0).Style.Fg)
			activeAttrs := scene.buf.CellAt(activeX, 0).Style.Attrs
			inactiveAttrs := scene.buf.CellAt(inactiveX, 0).Style.Attrs
			theme := getTheme()

			scene.hover(activeX, 0)
			assert.Equal(t, activeBase, scene.bgAt(activeX, 0), "close hover preserves the active background")
			assert.Equal(t, theme.Text, FromANSI(scene.buf.CellAt(activeX, 0).Style.Fg))
			assert.Equal(t, activeAttrs, scene.buf.CellAt(activeX, 0).Style.Attrs)
			assert.Equal(t, activeBase, scene.bgAt(activeX-2, 0), "active label retains its selection color")
			assert.Equal(t, "one", state.ActiveKeyPeek(), "hover does not select a tab")

			scene.hover(activeX-2, 0)
			assert.Equal(t, activeBase, scene.bgAt(activeX, 0), "leaving the close control clears its tint")
			assert.Equal(t, activeForeground, FromANSI(scene.buf.CellAt(activeX, 0).Style.Fg))
			assert.Equal(t, activeAttrs, scene.buf.CellAt(activeX, 0).Style.Attrs)

			scene.hover(inactiveX-2, 0)
			wholeTabTint := theme.Hover.BlendOver(inactiveBase)
			assert.Equal(t, wholeTabTint, scene.bgAt(inactiveX, 0), "existing whole-tab hover is preserved")
			scene.hover(inactiveX, 0)
			assert.Equal(t, wholeTabTint, scene.bgAt(inactiveX-2, 0))
			assert.Equal(t, wholeTabTint, scene.bgAt(inactiveX, 0), "close hover preserves the whole-tab background")
			assert.Equal(t, theme.Text, FromANSI(scene.buf.CellAt(inactiveX, 0).Style.Fg))
			assert.Equal(t, inactiveAttrs, scene.buf.CellAt(inactiveX, 0).Style.Attrs)

			scene.hover(39, 1)
			assert.Equal(t, inactiveBase, scene.bgAt(inactiveX, 0))
			assert.Equal(t, inactiveForeground, FromANSI(scene.buf.CellAt(inactiveX, 0).Style.Fg))
			assert.Equal(t, inactiveAttrs, scene.buf.CellAt(inactiveX, 0).Style.Attrs)
			scene.click(activeX, 0, 0)
			assert.Equal(t, "two", state.ActiveKeyPeek())
			assert.Equal(t, 1, state.TabCount(), "click still closes the active tab")
		})
	}
}

func TestTabCloseHover_PreservesCloseCallbackAndKeybinds(t *testing.T) {
	state := NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})
	var closed []string
	bar := TabBar{ID: "tabs", State: state, Closable: true, OnTabClose: func(key string) { closed = append(closed, key) }}
	scene := newClickScene(t, bar, 40, 2)
	positions := tabClosePositions(t, scene)
	require.Len(t, positions, 2)
	scene.hover(positions[1], 0)
	scene.click(positions[1], 0, 0)
	assert.Equal(t, []string{"two"}, closed)
	assert.Equal(t, "one", state.ActiveKeyPeek(), "closing an inactive tab does not select it")
	assert.Equal(t, 2, state.TabCount(), "custom close handler controls removal")
	for _, keybind := range bar.Keybinds() {
		if keybind.Key == "right" {
			keybind.Action()
		}
	}
	assert.Equal(t, "two", state.ActiveKeyPeek(), "keyboard tab navigation still selects")
	for _, keybind := range bar.Keybinds() {
		if keybind.Key == "ctrl+w" {
			keybind.Action()
		}
	}
	assert.Equal(t, []string{"two", "two"}, closed, "keyboard close keeps its original behavior")
}

func TestTabCloseHover_DisabledTabDoesNotTintOrClose(t *testing.T) {
	state := NewTabState([]Tab{{Key: "one", Label: "One"}})
	bar := TabBar{ID: "tabs", State: state, Closable: true}
	scene := newClickScene(t, DisabledWhen(true, bar), 40, 2)
	positions := tabClosePositions(t, scene)
	require.Len(t, positions, 1)
	x := positions[0]
	before := scene.bgAt(x, 0)
	scene.hover(x, 0)
	assert.Equal(t, before, scene.bgAt(x, 0))
	scene.click(x, 0, 0)
	assert.Equal(t, 1, state.TabCount())
}
