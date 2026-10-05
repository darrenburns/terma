package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tabClosePositions(t *testing.T, p *Pilot) []int {
	t.Helper()
	buf := p.Buffer()
	var positions []int
	for x := 0; x < buf.Width(); x++ {
		if cell := buf.CellAt(x, 0); cell != nil && cell.Content == "×" {
			positions = append(positions, x)
		}
	}
	return positions
}

func TestTabCloseHover_IndependentFeedback(t *testing.T) {
	for _, id := range []string{"tabs", ""} {
		t.Run("id="+id, func(t *testing.T) {
			state := NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})
			p := NewPilot(t, TabBar{ID: id, State: state, Closable: true}, 40, 2)
			positions := tabClosePositions(t, p)
			require.Len(t, positions, 2)
			activeX, inactiveX := positions[0], positions[1]
			activeBase, inactiveBase := bgAt(p, activeX, 0), bgAt(p, inactiveX, 0)
			activeForeground := FromANSI(p.Buffer().CellAt(activeX, 0).Style.Fg)
			inactiveForeground := FromANSI(p.Buffer().CellAt(inactiveX, 0).Style.Fg)
			activeAttrs := p.Buffer().CellAt(activeX, 0).Style.Attrs
			inactiveAttrs := p.Buffer().CellAt(inactiveX, 0).Style.Attrs
			theme := getTheme()

			p.MouseMove(activeX, 0)
			assert.Equal(t, activeBase, bgAt(p, activeX, 0), "close hover preserves the active background")
			assert.Equal(t, theme.Text, FromANSI(p.Buffer().CellAt(activeX, 0).Style.Fg))
			assert.Equal(t, activeAttrs, p.Buffer().CellAt(activeX, 0).Style.Attrs)
			assert.Equal(t, activeBase, bgAt(p, activeX-2, 0), "active label retains its selection color")
			assert.Equal(t, "one", state.ActiveKeyPeek(), "hover does not select a tab")

			p.MouseMove(activeX-2, 0)
			assert.Equal(t, activeBase, bgAt(p, activeX, 0), "leaving the close control clears its tint")
			assert.Equal(t, activeForeground, FromANSI(p.Buffer().CellAt(activeX, 0).Style.Fg))
			assert.Equal(t, activeAttrs, p.Buffer().CellAt(activeX, 0).Style.Attrs)

			p.MouseMove(inactiveX-2, 0)
			wholeTabTint := theme.Hover.BlendOver(inactiveBase)
			assert.Equal(t, wholeTabTint, bgAt(p, inactiveX, 0), "existing whole-tab hover is preserved")
			p.MouseMove(inactiveX, 0)
			assert.Equal(t, wholeTabTint, bgAt(p, inactiveX-2, 0))
			assert.Equal(t, wholeTabTint, bgAt(p, inactiveX, 0), "close hover preserves the whole-tab background")
			assert.Equal(t, theme.Text, FromANSI(p.Buffer().CellAt(inactiveX, 0).Style.Fg))
			assert.Equal(t, inactiveAttrs, p.Buffer().CellAt(inactiveX, 0).Style.Attrs)

			p.MouseMove(39, 1)
			assert.Equal(t, inactiveBase, bgAt(p, inactiveX, 0))
			assert.Equal(t, inactiveForeground, FromANSI(p.Buffer().CellAt(inactiveX, 0).Style.Fg))
			assert.Equal(t, inactiveAttrs, p.Buffer().CellAt(inactiveX, 0).Style.Attrs)
			p.ClickAt(activeX, 0)
			assert.Equal(t, "two", state.ActiveKeyPeek())
			assert.Equal(t, 1, state.TabCount(), "click still closes the active tab")
		})
	}
}

func TestTabCloseHover_PreservesCloseCallbackAndKeybinds(t *testing.T) {
	state := NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})
	var closed []string
	bar := TabBar{ID: "tabs", State: state, Closable: true, OnTabClose: func(key string) { closed = append(closed, key) }}
	p := NewPilot(t, bar, 40, 2)
	positions := tabClosePositions(t, p)
	require.Len(t, positions, 2)
	p.MouseMove(positions[1], 0)
	p.ClickAt(positions[1], 0)
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
	p := NewPilot(t, DisabledWhen(true, bar), 40, 2)
	positions := tabClosePositions(t, p)
	require.Len(t, positions, 1)
	x := positions[0]
	before := bgAt(p, x, 0)
	p.MouseMove(x, 0)
	assert.Equal(t, before, bgAt(p, x, 0))
	p.ClickAt(x, 0)
	assert.Equal(t, 1, state.TabCount())
}
