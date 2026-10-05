package terma

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func selectionSet(indices ...int) map[int]struct{} {
	set := make(map[int]struct{}, len(indices))
	for _, i := range indices {
		set[i] = struct{}{}
	}
	return set
}

func TestListDrag_MultiSelectSelectsRange(t *testing.T) {
	state := NewListState(numberedItems(6))
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	p := NewPilot(t, list, 20, 6)

	drag(p, 2, 1, 2, 4)

	assert.Equal(t, 4, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(1, 2, 3, 4), state.Selection.Peek())
	p.AssertSnapshot("dragged", "Dragging from Item 01 to Item 04 selects Items 01-04, with the cursor on Item 04")

	drag(p, 2, 3, 2, 2)
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(2, 3), state.Selection.Peek(), "a new drag replaces the selection, and can run upwards")
}

func TestListDrag_ShiftPressExtendsFromAnchor(t *testing.T) {
	state := NewListState(numberedItems(6))
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	p := NewPilot(t, list, 20, 6)

	p.ClickAt(2, 1)
	p.MouseDown(2, 3, uv.MouseLeft, uv.ModShift)
	p.MouseMove(2, 5)
	p.MouseUp(2, 5, uv.MouseLeft, 0)

	assert.Equal(t, selectionSet(1, 2, 3, 4, 5), state.Selection.Peek())
}

func TestListDrag_MovesCursorWithoutMultiSelect(t *testing.T) {
	state := NewListState(numberedItems(6))
	var changes []string
	list := List[string]{ID: "list", State: state, OnCursorChange: func(item string) { changes = append(changes, item) }}
	p := NewPilot(t, list, 20, 6)

	p.MouseDown(2, 0, uv.MouseLeft, 0)
	p.MouseMove(2, 2)
	p.MouseMove(2, 2)
	p.MouseMove(2, 3)
	p.MouseUp(2, 3, uv.MouseLeft, 0)

	assert.Equal(t, 3, state.CursorIndex.Peek())
	assert.Empty(t, state.Selection.Peek())
	assert.Equal(t, []string{"Item 02", "Item 03"}, changes, "each item the drag reaches is reported once")
}

func TestListDrag_PastViewportScrolls(t *testing.T) {
	scroll := NewScrollState()
	state := NewListState(numberedItems(20))
	list := List[string]{ID: "list", State: state, ScrollState: scroll, MultiSelect: true}
	root := Column{Children: []Widget{
		Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list},
		Text{Content: "Footer"},
	}}
	p := NewPilot(t, root, 20, 7)

	p.MouseDown(2, 2, uv.MouseLeft, 0)
	p.MouseMove(2, 6) // Below the viewport, over the footer.
	assert.Equal(t, 6, state.CursorIndex.Peek())
	assert.Equal(t, 2, scroll.GetOffset(), "the cursor item is scrolled into view")

	p.MouseMove(2, 6) // The list has scrolled under the pointer.
	p.MouseUp(2, 6, uv.MouseLeft, 0)
	assert.Equal(t, 8, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())
	assert.Equal(t, selectionSet(2, 3, 4, 5, 6, 7, 8), state.Selection.Peek())
	p.AssertSnapshot("dragged", "Dragging from Item 02 to below the viewport scrolls the list, selecting Items 02-08")
}

func TestListDrag_ReleaseEndsDrag(t *testing.T) {
	state := NewListState(numberedItems(6))
	list := List[string]{ID: "list", State: state}
	p := NewPilot(t, list, 20, 6)

	drag(p, 2, 1, 2, 2)
	p.MouseMove(2, 4)

	assert.Equal(t, 2, state.CursorIndex.Peek(), "moving without a button held doesn't move the cursor")
}

func TestListClick_NonFocusableListTakesClicks(t *testing.T) {
	state := NewListState(numberedItems(4))
	var selected []string
	list := List[string]{
		ID: "list", DisableFocus: true, State: state,
		OnSelect: func(item string) { selected = append(selected, item) },
	}
	input := TextInput{ID: "input", State: NewTextInputState("")}
	p := NewPilot(t, Column{Children: []Widget{input, list}}, 20, 5)
	p.Click("input")

	p.ClickAt(2, 3) // Item 02, below the input.
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, "input", p.FocusedID(), "a list that can't be focused leaves focus where it was")

	p.ClickAt(2, 3)
	assert.Equal(t, []string{"Item 02"}, selected)
}

func mouseTreeNodes() []TreeNode[string] {
	return []TreeNode[string]{
		{Data: "src", Children: []TreeNode[string]{{Data: "main.go"}, {Data: "util.go"}}},
		{Data: "docs", Children: []TreeNode[string]{{Data: "guide.md"}}},
		{Data: "README.md"},
	}
}

func TestTreeClick_DoubleClickSelectsNode(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	var selected []string
	tree := Tree[string]{ID: "tree", State: state, OnSelect: func(node string, _ []string) { selected = append(selected, node) }}
	p := NewPilot(t, tree, 30, 8)

	p.ClickAt(8, 1) // main.go
	assert.Empty(t, selected, "a single click only moves the cursor")
	node, _ := state.CursorNode()
	assert.Equal(t, "main.go", node)

	p.ClickAt(8, 1)
	assert.Equal(t, []string{"main.go"}, selected)
}

func TestTreeClick_IndicatorTogglesWithoutSelecting(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	var selected int
	tree := Tree[string]{ID: "tree", State: state, OnSelect: func(string, []string) { selected++ }}
	p := NewPilot(t, tree, 30, 8)
	collapsed := state.IsCollapsed([]int{0})

	p.ClickAt(0, 0) // The expand indicator on "src".
	assert.NotEqual(t, collapsed, state.IsCollapsed([]int{0}))
	p.ClickAt(0, 0)
	assert.Equal(t, collapsed, state.IsCollapsed([]int{0}), "each click on the indicator toggles")
	assert.Zero(t, selected, "double-clicking the indicator doesn't select")
}

func TestTreeDrag_MultiSelectSelectsRange(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	tree := Tree[string]{ID: "tree", State: state, MultiSelect: true}
	p := NewPilot(t, tree, 30, 8)
	require.Len(t, state.viewPaths, 6, "every folder starts expanded")

	drag(p, 8, 1, 8, 3)

	assert.Equal(t, [][]int{{0, 0}, {0, 1}, {1}}, state.SelectedPaths())
	p.AssertSnapshot("dragged", "Dragging from main.go down to docs selects main.go, util.go and docs")
}

func TestDirectoryTreeClick_IndicatorLoadsChildren(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
	state := NewDirectoryTreeState(root)
	tree := DirectoryTree{Tree: Tree[DirectoryEntry]{ID: "dirs", State: state}}
	p := NewPilot(t, tree, 30, 5)
	require.Len(t, state.viewPaths, 1)

	p.ClickAt(0, 0) // The root's expand indicator.
	p.WaitUntil(func() bool { return len(state.viewPaths) == 2 }, time.Second)

	p.ClickAt(4, 1)
	node, _ := state.CursorNode()
	assert.Equal(t, "sub", node.Name)
}

func mouseTableRows() [][]string {
	rows := make([][]string, 5)
	for i := range rows {
		rows[i] = []string{fmt.Sprintf("r%d", i), fmt.Sprintf("a%d", i), fmt.Sprintf("b%d", i)}
	}
	return rows
}

func mouseTable(state *TableState[[]string], mode TableSelectionMode, multi bool) Table[[]string] {
	return Table[[]string]{
		ID:            "table",
		State:         state,
		SelectionMode: mode,
		MultiSelect:   multi,
		ColumnSpacing: 1,
		Columns: []TableColumn{
			{Width: Cells(4), Header: Text{Content: "Id"}},
			{Width: Cells(4), Header: Text{Content: "A"}},
			{Width: Cells(4), Header: Text{Content: "B"}},
		},
	}
}

func TestTableClick_MovesCursorToClickedCell(t *testing.T) {
	state := NewTableState(mouseTableRows())
	var changes []string
	table := mouseTable(state, TableSelectionCursor, false)
	table.OnCursorChange = func(row []string) { changes = append(changes, row[0]) }
	p := NewPilot(t, table, 20, 7)

	p.ClickAt(6, 3) // Row r2, column A (columns start at x 0, 5, 10; the header takes row 0).
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, 1, state.CursorColumn.Peek())
	assert.Equal(t, []string{"r2"}, changes)
	assert.Equal(t, "table", p.FocusedID())
	p.AssertSnapshot("clicked", "Clicking cell a2 puts the cursor on row r2, column A")

	p.ClickAt(4, 1) // The gap after the Id column belongs to it.
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 0, state.CursorColumn.Peek())

	p.ClickAt(6, 0) // The header row isn't a row of data.
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 0, state.CursorColumn.Peek())
}

func TestTableClick_DoubleClickSelectsRow(t *testing.T) {
	state := NewTableState(mouseTableRows())
	var selected []string
	table := mouseTable(state, TableSelectionRow, false)
	table.OnSelect = func(row []string) { selected = append(selected, row[0]) }
	p := NewPilot(t, table, 20, 7)

	p.ClickAt(11, 4)
	assert.Empty(t, selected)
	p.ClickAt(11, 4)
	assert.Equal(t, []string{"r3"}, selected)
}

func TestTableClick_ShiftClickExtendsRowSelection(t *testing.T) {
	state := NewTableState(mouseTableRows())
	p := NewPilot(t, mouseTable(state, TableSelectionRow, true), 20, 7)

	p.ClickAt(1, 1)
	p.MouseDown(1, 3, uv.MouseLeft, uv.ModShift)
	p.MouseUp(1, 3, uv.MouseLeft, uv.ModShift)

	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(0, 1, 2), state.Selection.Peek())
}

func TestTableDrag_CellModeSelectsBox(t *testing.T) {
	state := NewTableState(mouseTableRows())
	p := NewPilot(t, mouseTable(state, TableSelectionCursor, true), 20, 7)

	drag(p, 6, 1, 11, 3) // From a0 to b2.

	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, 2, state.CursorColumn.Peek())
	assert.Equal(t, selectionSet(
		cellIndex(0, 1, 3), cellIndex(0, 2, 3),
		cellIndex(1, 1, 3), cellIndex(1, 2, 3),
		cellIndex(2, 1, 3), cellIndex(2, 2, 3),
	), state.Selection.Peek())
	p.AssertSnapshot("dragged", "Dragging from a0 to b2 selects the box of cells between them")
}

func TestTableDrag_ColumnModeSelectsColumns(t *testing.T) {
	state := NewTableState(mouseTableRows())
	p := NewPilot(t, mouseTable(state, TableSelectionColumn, true), 20, 7)

	drag(p, 1, 2, 11, 2)

	assert.Equal(t, 2, state.CursorColumn.Peek())
	assert.Equal(t, selectionSet(0, 1, 2), state.Selection.Peek())
}

func TestTableDrag_RowModeMovesCursorWithoutMultiSelect(t *testing.T) {
	state := NewTableState(mouseTableRows())
	p := NewPilot(t, mouseTable(state, TableSelectionRow, false), 20, 7)

	drag(p, 1, 1, 1, 9) // Past the last row.

	assert.Equal(t, 4, state.CursorIndex.Peek())
	assert.Empty(t, state.Selection.Peek())
}

func mouseMenuScene(t *testing.T, onSelect func(MenuItem)) (*Pilot, *MenuState) {
	state := NewMenuState([]MenuItem{
		{Label: "Open"},
		{Label: "Save"},
		{Divider: "More"},
		{Label: "Export", Disabled: true},
		{Label: "Close"},
	})
	menu := Menu{ID: "menu", State: state, Position: FloatPositionTopLeft, OnSelect: onSelect}
	p := NewPilot(t, menu, 20, 7)
	p.session.focus.FocusByID("menu")
	p.settle()
	return p, state
}

func TestMenuClick_ChoosesItem(t *testing.T) {
	var selected []string
	p, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	p.MouseDown(2, 1, uv.MouseLeft, 0)
	assert.Equal(t, 1, state.CursorIndex())
	assert.Empty(t, selected, "pressing only moves the cursor")
	p.AssertSnapshot("pressed", "Pressing Save highlights it")

	p.MouseUp(2, 1, uv.MouseLeft, 0)
	assert.Equal(t, []string{"Save"}, selected, "releasing over the item chooses it")
}

func TestMenuClick_IgnoresDividersAndDisabledItems(t *testing.T) {
	var selected []string
	p, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	p.ClickAt(2, 2)
	p.ClickAt(2, 3)

	assert.Equal(t, 0, state.CursorIndex())
	assert.Empty(t, selected)
}

func TestMenuDrag_ReleaseChoosesItemUnderPointer(t *testing.T) {
	var selected []string
	p, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	p.MouseDown(2, 0, uv.MouseLeft, 0)
	p.MouseMove(2, 3) // Disabled: the cursor stays put.
	assert.Equal(t, 0, state.CursorIndex())
	p.MouseMove(2, 4)
	assert.Equal(t, 4, state.CursorIndex())
	assert.Empty(t, selected)

	p.MouseUp(2, 4, uv.MouseLeft, 0)
	assert.Equal(t, []string{"Close"}, selected)
}

func TestMenuDrag_ReleaseOffItemsChoosesNothing(t *testing.T) {
	var selected []string
	p, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	p.MouseDown(2, 1, uv.MouseLeft, 0)
	p.MouseMove(18, 1) // Beside the menu.
	p.MouseUp(18, 1, uv.MouseLeft, 0)
	p.MouseDown(2, 1, uv.MouseLeft, 0)
	p.MouseMove(2, 3)
	p.MouseUp(2, 3, uv.MouseLeft, 0) // Over the disabled item.

	assert.Equal(t, 1, state.CursorIndex())
	assert.Empty(t, selected)
}

func TestCommandPaletteClick_KeepsInputFocusAndDoubleClickRuns(t *testing.T) {
	var ran []string
	items := []CommandPaletteItem{
		{Label: "New File", Action: func() { ran = append(ran, "new") }},
		{Divider: "Edit"},
		{Label: "Cut", Action: func() { ran = append(ran, "cut") }},
		{Label: "Copy", Disabled: true, Action: func() { ran = append(ran, "copy") }},
	}
	state := NewCommandPaletteState("Commands", items)
	state.Visible.Set(true)
	palette := CommandPalette{ID: "palette", State: state, Position: FloatPositionTopLeft}
	p := NewPilot(t, palette, 50, 12)
	inputID := palette.inputID()
	require.Equal(t, inputID, p.FocusedID())

	listBounds, ok := p.Bounds(palette.listID())
	require.True(t, ok)
	x, y := listBounds.X+2, listBounds.Y

	p.ClickAt(x, y+2) // Cut.
	level := state.CurrentLevel()
	assert.Equal(t, 2, level.ListState.CursorIndex.Peek())
	assert.Equal(t, inputID, p.FocusedID(), "clicking a result leaves focus in the input")
	p.AssertSnapshot("clicked", "Clicking Cut highlights it while the input keeps focus")

	p.ClickAt(x, y+1) // The divider can't take the cursor.
	p.ClickAt(x, y+3) // Nor can a disabled item.
	p.ClickAt(x, y+3)
	assert.Equal(t, 2, level.ListState.CursorIndex.Peek())
	assert.Empty(t, ran)

	p.ClickAt(x, y+2)
	p.ClickAt(x, y+2)
	assert.Equal(t, []string{"cut"}, ran)
}

func TestAutocompleteClick_DoubleClickInsertsSuggestion(t *testing.T) {
	inputState := NewTextInputState("")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{{Label: "apple"}, {Label: "apricot"}, {Label: "banana"}})
	ac := Autocomplete{
		ID:    "ac",
		State: acState,
		Child: TextInput{ID: "input", State: inputState},
	}
	p := NewPilot(t, ac, 30, 8)
	p.Click("input")
	acState.Show()

	listBounds, ok := p.Bounds("ac-list")
	require.True(t, ok)
	x, y := listBounds.X+1, listBounds.Y

	p.ClickAt(x, y+1)
	assert.Equal(t, "input", p.FocusedID())
	assert.True(t, acState.IsVisible(), "clicking a suggestion keeps the popup open")
	suggestion, _ := acState.SelectedSuggestion()
	assert.Equal(t, "apricot", suggestion.Label)

	p.ClickAt(x, y+1)
	assert.Equal(t, "apricot", inputState.GetText())
}
