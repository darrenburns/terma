package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMenuHover_IndependentOfCursorAndSkipsUnselectableRows(t *testing.T) {
	scene, state := mouseMenuScene(t, nil)
	theme := getTheme()

	scene.hover(2, 1)
	assert.Equal(t, 0, state.CursorIndex(), "hover does not move the keyboard cursor")
	assert.Equal(t, hoveredKey[int]{key: 1, ok: true}, state.hover.signal.Peek())
	assert.Equal(t, theme.Hover.BlendOver(theme.Surface), scene.bgAt(2, 1))

	scene.hover(2, 0)
	assert.Equal(t, theme.ActiveCursor, scene.bgAt(2, 0), "cursor colour covers the hover tint")
	assert.Equal(t, theme.Surface, scene.bgAt(2, 1), "leaving clears the previous row")

	for _, y := range []int{2, 3, 6} {
		scene.hover(2, y)
		assert.False(t, state.hover.signal.Peek().ok, "divider, disabled row and outside have no hover")
		assert.Equal(t, 0, state.CursorIndex())
	}
}

func TestMenuHover_SubmenuRowsUseTheirOwnState(t *testing.T) {
	state := NewMenuState([]MenuItem{{Label: "Recent", Children: []MenuItem{{Label: "First"}, {Label: "Second"}}}})
	state.OpenSubmenu(0)
	scene := newClickScene(t, Menu{ID: "menu", State: state, Position: FloatPositionTopLeft}, 40, 8)
	scene.focus.FocusByID("menu-sub")
	scene.draw()
	entry := scene.renderer.WidgetByID("menu-sub-item-1")
	require.NotNil(t, entry)
	x, y := entry.Bounds.X+1, entry.Bounds.Y
	scene.hover(x, y)
	assert.Equal(t, hoveredKey[int]{key: 1, ok: true}, state.submenuState.hover.signal.Peek())
	assert.False(t, state.hover.signal.Peek().ok)
	assert.Equal(t, 0, state.submenuState.CursorIndex())
	assert.Equal(t, getTheme().Hover.BlendOver(getTheme().Surface), scene.bgAt(x, y))
}
