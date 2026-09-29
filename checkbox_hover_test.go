package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckboxHover_TintsAndRepaints(t *testing.T) {
	for _, checked := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchecked", true: "checked"}[checked], func(t *testing.T) {
			checkbox := &Checkbox{ID: "checkbox", DisableFocus: true, State: NewCheckboxState(checked), Label: "Option", Style: Style{Width: Cells(12)}}
			scene := newClickScene(t, hoverScreen{checkbox}, 20, 3)
			theme := getTheme()
			require.Equal(t, theme.Surface, scene.bgAt(0, 0))
			require.True(t, scene.router.motion(uv.MouseMotionEvent{X: 3, Y: 0}, 0.5, 0.5))
			scene.renderer.Update(scene.root)
			assert.Zero(t, scene.renderer.Stats().BuildCount, "hover only repaints")
			assert.Equal(t, theme.Hover.BlendOver(theme.Surface), scene.bgAt(0, 0))
			assert.Equal(t, theme.Hover.BlendOver(theme.Surface), scene.bgAt(11, 0), "extra width is tinted too")
			scene.hover(15, 2)
			assert.Equal(t, theme.Surface, scene.bgAt(0, 0))
			assert.Equal(t, checked, checkbox.State.IsChecked(), "hover does not toggle")
		})
	}
}

func TestCheckboxHover_PreservesFocusAndDisabled(t *testing.T) {
	checkbox := &Checkbox{ID: "checkbox", State: NewCheckboxState(false), Label: "Option"}
	scene := newClickScene(t, hoverScreen{checkbox}, 20, 3)
	scene.focus.FocusByID("checkbox")
	scene.draw()
	theme := getTheme()
	scene.hover(3, 0)
	assert.Equal(t, theme.Hover.BlendOver(theme.ActiveCursor), scene.bgAt(0, 0))
	assert.Equal(t, theme.SelectionText, FromANSI(scene.buf.CellAt(0, 0).Style.Fg))
	scene.hover(15, 2)
	assert.Equal(t, theme.ActiveCursor, scene.bgAt(0, 0))

	disabled := newClickScene(t, hoverScreen{DisabledWhen(true, checkbox)}, 20, 3)
	before := disabled.bgAt(0, 0)
	disabled.hover(3, 0)
	assert.Equal(t, before, disabled.bgAt(0, 0))
	assert.Equal(t, theme.TextDisabled, FromANSI(disabled.buf.CellAt(0, 0).Style.Fg))
}

func TestCheckboxHover_CustomBackground(t *testing.T) {
	custom := RGB(20, 50, 80)
	checkbox := &Checkbox{ID: "checkbox", DisableFocus: true, State: NewCheckboxState(false), Label: "Option", Style: Style{BackgroundColor: custom}}
	scene := newClickScene(t, hoverScreen{checkbox}, 20, 3)
	scene.hover(3, 0)
	assert.Equal(t, getTheme().Hover.BlendOver(custom), scene.bgAt(0, 0))
}
