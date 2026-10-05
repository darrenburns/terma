package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListClick_MovesCursorToClickedItem(t *testing.T) {
	state := NewListState(numberedItems(5))
	var changes []string
	list := List[string]{ID: "list", State: state, OnCursorChange: func(item string) { changes = append(changes, item) }}
	root := Column{Children: []Widget{Text{Content: "Header"}, list}}
	p := NewPilot(t, root, 20, 6)

	p.ClickAt(3, 4) // Item 03 sits on screen row 4, below the header.

	assert.Equal(t, 3, state.CursorIndex.Peek())
	assert.Equal(t, []string{"Item 03"}, changes)
	assert.Equal(t, "list", p.FocusedID(), "clicking an item focuses the list")
	p.AssertSnapshot("clicked", "Clicking Item 03 focuses the list and moves the cursor onto it")

	p.ClickAt(3, 4)
	assert.Equal(t, []string{"Item 03"}, changes, "clicking the cursor item again doesn't report a cursor change")
}

func TestListClick_BorderAndPadding(t *testing.T) {
	state := NewListState(numberedItems(4))
	list := List[string]{
		ID: "list", State: state,
		Style: Style{Padding: EdgeInsetsAll(1), Border: SquareBorder(RGB(200, 200, 200))},
	}
	p := NewPilot(t, list, 20, 8)

	p.ClickAt(4, 1) // Top padding row: no item.
	assert.Equal(t, 0, state.CursorIndex.Peek())

	p.ClickAt(4, 4) // Border and padding take rows 0-1, so row 4 is Item 02.
	assert.Equal(t, 2, state.CursorIndex.Peek())
	p.AssertSnapshot("clicked", "In a bordered, padded List, clicking screen row 4 puts the cursor on Item 02")
}

func TestListClick_Scrolled(t *testing.T) {
	scroll := NewScrollState()
	state := NewListState(numberedItems(20))
	list := List[string]{ID: "list", State: state, ScrollState: scroll}
	p := NewPilot(t, Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list}, 24, 5)

	scroll.ScrollDown(10)
	p.settle()
	p.ClickAt(4, 2) // Row 2 of a view scrolled 10 rows is Item 12.

	assert.Equal(t, 12, state.CursorIndex.Peek())
	assert.Equal(t, 10, scroll.GetOffset(), "clicking a visible item doesn't scroll")
	p.AssertSnapshot("clicked", "In a List scrolled 10 rows, clicking the third visible row puts the cursor on Item 12 without scrolling")
}

func TestListClick_ShiftClickExtendsSelection(t *testing.T) {
	state := NewListState(numberedItems(6))
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	p := NewPilot(t, list, 20, 6)

	p.ClickAt(2, 1)
	p.MouseDown(2, 4, uv.MouseLeft, uv.ModShift)
	p.MouseUp(2, 4, uv.MouseLeft, uv.ModShift)

	assert.Equal(t, 4, state.CursorIndex.Peek())
	assert.Equal(t, map[int]struct{}{1: {}, 2: {}, 3: {}, 4: {}}, state.Selection.Peek())
	p.AssertSnapshot("extended", "Clicking Item 01 then shift+clicking Item 04 selects Items 01-04")

	p.ClickAt(2, 2)
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Empty(t, state.Selection.Peek(), "a plain click clears the selection")
}

func TestListClick_DoubleClickSelectsItem(t *testing.T) {
	state := NewListState(numberedItems(4))
	var selected []string
	list := List[string]{ID: "list", State: state, OnSelect: func(item string) { selected = append(selected, item) }}
	p := NewPilot(t, list, 20, 4)

	p.ClickAt(1, 2)
	assert.Empty(t, selected, "a single click only moves the cursor")
	p.ClickAt(1, 2)
	assert.Equal(t, []string{"Item 02"}, selected)
}

func TestListClick_BelowLastItemDoesNothing(t *testing.T) {
	state := NewListState(numberedItems(3))
	state.CursorIndex.Set(1)
	var changes int
	list := List[string]{ID: "list", State: state, Style: Style{Height: Cells(6)}, OnCursorChange: func(string) { changes++ }}
	p := NewPilot(t, list, 20, 6)

	p.ClickAt(1, 5)

	assert.Equal(t, 1, state.CursorIndex.Peek())
	assert.Zero(t, changes)
}

func TestListClick_FilteredViewMapsToSourceIndex(t *testing.T) {
	state := NewListState([]string{"apple", "banana", "cherry", "blueberry"})
	filter := NewFilterState()
	filter.Query.Set("b")
	list := List[string]{ID: "list", State: state, Filter: filter}
	p := NewPilot(t, list, 20, 4)

	p.ClickAt(1, 1) // Second visible row: "blueberry".

	require.Equal(t, 3, state.CursorIndex.Peek())
	item, _ := state.SelectedItem()
	assert.Equal(t, "blueberry", item)
}
