package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckboxHover_TintsAndRepaints(t *testing.T) {
	for _, checked := range []bool{false, true} {
		t.Run(map[bool]string{false: "unchecked", true: "checked"}[checked], func(t *testing.T) {
			checkbox := &Checkbox{ID: "checkbox", DisableFocus: true, State: NewCheckboxState(checked), Label: "Option", Style: Style{Width: Cells(12)}}
			p := NewPilot(t, hoverScreen{checkbox}, 20, 3)
			theme := getTheme()
			require.Equal(t, theme.Surface, bgAt(p, 0, 0))
			p.MouseMove(3, 0)
			assert.Zero(t, p.session.renderer.Stats().BuildCount, "hover only repaints")
			assert.Equal(t, theme.Hover.BlendOver(theme.Surface), bgAt(p, 0, 0))
			assert.Equal(t, theme.Hover.BlendOver(theme.Surface), bgAt(p, 11, 0), "extra width is tinted too")
			p.MouseMove(15, 2)
			assert.Equal(t, theme.Surface, bgAt(p, 0, 0))
			assert.Equal(t, checked, checkbox.State.IsChecked(), "hover does not toggle")
		})
	}
}

func TestCheckboxHover_PreservesFocusAndDisabled(t *testing.T) {
	checkbox := &Checkbox{ID: "checkbox", State: NewCheckboxState(false), Label: "Option"}
	theme := getTheme()
	t.Run("focused", func(t *testing.T) {
		p := NewPilot(t, hoverScreen{checkbox}, 20, 3)
		p.session.focus.FocusByID("checkbox")
		p.settle()
		p.MouseMove(3, 0)
		assert.Equal(t, theme.Hover.BlendOver(theme.ActiveCursor), bgAt(p, 0, 0))
		assert.Equal(t, theme.SelectionText, FromANSI(p.Buffer().CellAt(0, 0).Style.Fg))
		p.MouseMove(15, 2)
		assert.Equal(t, theme.ActiveCursor, bgAt(p, 0, 0))
	})
	t.Run("disabled", func(t *testing.T) {
		p := NewPilot(t, hoverScreen{DisabledWhen(true, checkbox)}, 20, 3)
		before := bgAt(p, 0, 0)
		p.MouseMove(3, 0)
		assert.Equal(t, before, bgAt(p, 0, 0))
		assert.Equal(t, theme.TextDisabled, FromANSI(p.Buffer().CellAt(0, 0).Style.Fg))
	})
}

func TestCheckboxHover_CustomBackground(t *testing.T) {
	custom := RGB(20, 50, 80)
	checkbox := &Checkbox{ID: "checkbox", DisableFocus: true, State: NewCheckboxState(false), Label: "Option", Style: Style{BackgroundColor: custom}}
	p := NewPilot(t, hoverScreen{checkbox}, 20, 3)
	p.MouseMove(3, 0)
	assert.Equal(t, getTheme().Hover.BlendOver(custom), bgAt(p, 0, 0))
}
