package terma

import (
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
)

// rightClick presses and releases the right button at (x, y).
func (s *clickScene) rightClick(x, y int) {
	s.t.Helper()
	s.now = s.now.Add(time.Second)
	s.router.press(uv.MouseClickEvent{X: x, Y: y, Button: uv.MouseRight}, 0.5, 0.5, s.now)
	s.router.release(uv.MouseReleaseEvent{X: x, Y: y, Button: uv.MouseRight}, 0.5, 0.5)
	s.draw()
}

// pause lets enough time pass that the next click starts a new click chain.
func (s *clickScene) pause() {
	s.now = s.now.Add(time.Second)
}

func TestTreeClick_ActivateOnClickSelectsOnSingleClick(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	var selected []string
	tree := Tree[string]{ID: "tree", State: state, ActivateOnClick: true,
		OnSelect: func(node string, _ []string) { selected = append(selected, node) }}
	p := NewPilot(t, tree, 30, 8)

	p.ClickAt(8, 1) // main.go
	assert.Equal(t, []string{"main.go"}, selected)

	p.ClickAt(8, 1)
	assert.Equal(t, []string{"main.go"}, selected, "the second click of a double-click doesn't activate again")

	p.Advance(time.Second)
	p.ClickAt(8, 2) // util.go
	assert.Equal(t, []string{"main.go", "util.go"}, selected)

	p.Advance(time.Second)
	p.ClickAt(0, 0) // The expand indicator on "src" toggles it without activating.
	assert.Equal(t, []string{"main.go", "util.go"}, selected)

	p.Advance(time.Second)
	p.MouseDown(8, 0, uv.MouseRight, 0)
	p.MouseUp(8, 0, uv.MouseRight, 0)
	assert.Equal(t, []string{"main.go", "util.go"}, selected, "a right click doesn't activate")
}

func TestListClick_ActivateOnClickSelectsOnSingleClick(t *testing.T) {
	state := NewListState(numberedItems(5))
	var selected []string
	list := List[string]{ID: "list", State: state, ActivateOnClick: true, MultiSelect: true,
		OnSelect: func(item string) { selected = append(selected, item) }}
	p := NewPilot(t, list, 20, 5)

	p.ClickAt(3, 2)
	assert.Equal(t, []string{"Item 02"}, selected)

	p.Advance(time.Second)
	p.MouseDown(3, 4, uv.MouseLeft, uv.ModShift)
	p.MouseUp(3, 4, uv.MouseLeft, uv.ModShift)
	assert.Equal(t, []string{"Item 02"}, selected, "shift+click extends the selection instead")
	assert.Equal(t, selectionSet(2, 3, 4), state.Selection.Peek())
}

func TestListClick_DefaultNeedsDoubleClick(t *testing.T) {
	state := NewListState(numberedItems(3))
	var selected []string
	list := List[string]{ID: "list", State: state, OnSelect: func(item string) { selected = append(selected, item) }}
	p := NewPilot(t, list, 20, 3)

	p.ClickAt(3, 1)
	assert.Empty(t, selected)
	p.ClickAt(3, 1)
	assert.Equal(t, []string{"Item 01"}, selected)
}

func TestTableClick_ActivateOnClickSelectsOnSingleClick(t *testing.T) {
	state := NewTableState(mouseTableRows())
	var selected []string
	table := mouseTable(state, TableSelectionRow, false)
	table.ActivateOnClick = true
	table.OnSelect = func(row []string) { selected = append(selected, row[0]) }
	p := NewPilot(t, table, 20, 7)

	p.ClickAt(11, 4)
	assert.Equal(t, []string{"r3"}, selected)
}
