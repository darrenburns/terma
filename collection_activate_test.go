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
	scene := newClickScene(t, tree, 30, 8)

	scene.click(8, 1, 0) // main.go
	assert.Equal(t, []string{"main.go"}, selected)

	scene.click(8, 1, 0)
	assert.Equal(t, []string{"main.go"}, selected, "the second click of a double-click doesn't activate again")

	scene.pause()
	scene.click(8, 2, 0) // util.go
	assert.Equal(t, []string{"main.go", "util.go"}, selected)

	scene.pause()
	scene.click(0, 0, 0) // The expand indicator on "src" toggles it without activating.
	assert.Equal(t, []string{"main.go", "util.go"}, selected)

	scene.rightClick(8, 0)
	assert.Equal(t, []string{"main.go", "util.go"}, selected, "a right click doesn't activate")
}

func TestListClick_ActivateOnClickSelectsOnSingleClick(t *testing.T) {
	state := NewListState(clickSceneItems(5))
	var selected []string
	list := List[string]{ID: "list", State: state, ActivateOnClick: true, MultiSelect: true,
		OnSelect: func(item string) { selected = append(selected, item) }}
	scene := newClickScene(t, list, 20, 5)

	scene.click(3, 2, 0)
	assert.Equal(t, []string{"Item 02"}, selected)

	scene.pause()
	scene.click(3, 4, uv.ModShift)
	assert.Equal(t, []string{"Item 02"}, selected, "shift+click extends the selection instead")
	assert.Equal(t, selectionSet(2, 3, 4), state.Selection.Peek())
}

func TestListClick_DefaultNeedsDoubleClick(t *testing.T) {
	state := NewListState(clickSceneItems(3))
	var selected []string
	list := List[string]{ID: "list", State: state, OnSelect: func(item string) { selected = append(selected, item) }}
	scene := newClickScene(t, list, 20, 3)

	scene.click(3, 1, 0)
	assert.Empty(t, selected)
	scene.click(3, 1, 0)
	assert.Equal(t, []string{"Item 01"}, selected)
}

func TestTableClick_ActivateOnClickSelectsOnSingleClick(t *testing.T) {
	state := NewTableState(mouseTableRows())
	var selected []string
	table := mouseTable(state, TableSelectionRow, false)
	table.ActivateOnClick = true
	table.OnSelect = func(row []string) { selected = append(selected, row[0]) }
	scene := newClickScene(t, table, 20, 7)

	scene.click(11, 4, 0)
	assert.Equal(t, []string{"r3"}, selected)
}
