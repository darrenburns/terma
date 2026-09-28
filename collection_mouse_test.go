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

// press holds the left button down at (x, y) and redraws.
func (s *clickScene) press(x, y int, mod uv.KeyMod) {
	s.t.Helper()
	s.now = s.now.Add(50 * time.Millisecond)
	s.router.press(uv.MouseClickEvent{X: x, Y: y, Button: uv.MouseLeft, Mod: mod}, 0.5, 0.5, s.now)
	s.draw()
}

// move reports the pointer at (x, y) with the left button held, and redraws.
func (s *clickScene) move(x, y int) {
	s.t.Helper()
	s.router.motion(uv.MouseMotionEvent{X: x, Y: y, Button: uv.MouseLeft}, 0.5, 0.5)
	s.draw()
}

// release lets the left button up at (x, y) and redraws.
func (s *clickScene) release(x, y int) {
	s.t.Helper()
	s.router.release(uv.MouseReleaseEvent{X: x, Y: y, Button: uv.MouseLeft}, 0.5, 0.5)
	s.draw()
}

// drag presses at (x0, y0), moves to (x1, y1) and releases there.
func (s *clickScene) drag(x0, y0, x1, y1 int) {
	s.t.Helper()
	s.press(x0, y0, 0)
	s.move(x1, y1)
	s.release(x1, y1)
}

func selectionSet(indices ...int) map[int]struct{} {
	set := make(map[int]struct{}, len(indices))
	for _, i := range indices {
		set[i] = struct{}{}
	}
	return set
}

func TestListDrag_MultiSelectSelectsRange(t *testing.T) {
	state := NewListState(clickSceneItems(6))
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	scene := newClickScene(t, list, 20, 6)

	scene.drag(2, 1, 2, 4)

	assert.Equal(t, 4, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(1, 2, 3, 4), state.Selection.Peek())
	scene.snapshot("TestListDrag_MultiSelectSelectsRange", "Dragging from Item 01 to Item 04 selects Items 01-04, with the cursor on Item 04")

	scene.drag(2, 3, 2, 2)
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(2, 3), state.Selection.Peek(), "a new drag replaces the selection, and can run upwards")
}

func TestListDrag_ShiftPressExtendsFromAnchor(t *testing.T) {
	state := NewListState(clickSceneItems(6))
	list := List[string]{ID: "list", State: state, MultiSelect: true}
	scene := newClickScene(t, list, 20, 6)

	scene.click(2, 1, 0)
	scene.press(2, 3, uv.ModShift)
	scene.move(2, 5)
	scene.release(2, 5)

	assert.Equal(t, selectionSet(1, 2, 3, 4, 5), state.Selection.Peek())
}

func TestListDrag_MovesCursorWithoutMultiSelect(t *testing.T) {
	state := NewListState(clickSceneItems(6))
	var changes []string
	list := List[string]{ID: "list", State: state, OnCursorChange: func(item string) { changes = append(changes, item) }}
	scene := newClickScene(t, list, 20, 6)

	scene.press(2, 0, 0)
	scene.move(2, 2)
	scene.move(2, 2)
	scene.move(2, 3)
	scene.release(2, 3)

	assert.Equal(t, 3, state.CursorIndex.Peek())
	assert.Empty(t, state.Selection.Peek())
	assert.Equal(t, []string{"Item 02", "Item 03"}, changes, "each item the drag reaches is reported once")
}

func TestListDrag_PastViewportScrolls(t *testing.T) {
	scroll := NewScrollState()
	state := NewListState(clickSceneItems(20))
	list := List[string]{ID: "list", State: state, ScrollState: scroll, MultiSelect: true}
	root := Column{Children: []Widget{
		Scrollable{ID: "scroll", State: scroll, Height: Cells(5), Child: list},
		Text{Content: "Footer"},
	}}
	scene := newClickScene(t, root, 20, 7)

	scene.press(2, 2, 0)
	scene.move(2, 6) // Below the viewport, over the footer.
	assert.Equal(t, 6, state.CursorIndex.Peek())
	assert.Equal(t, 2, scroll.GetOffset(), "the cursor item is scrolled into view")

	scene.move(2, 6) // The list has scrolled under the pointer.
	scene.release(2, 6)
	assert.Equal(t, 8, state.CursorIndex.Peek())
	assert.Equal(t, 4, scroll.GetOffset())
	assert.Equal(t, selectionSet(2, 3, 4, 5, 6, 7, 8), state.Selection.Peek())
	scene.snapshot("TestListDrag_PastViewportScrolls", "Dragging from Item 02 to below the viewport scrolls the list, selecting Items 02-08")
}

func TestListDrag_ReleaseEndsDrag(t *testing.T) {
	state := NewListState(clickSceneItems(6))
	list := List[string]{ID: "list", State: state}
	scene := newClickScene(t, list, 20, 6)

	scene.drag(2, 1, 2, 2)
	scene.router.motion(uv.MouseMotionEvent{X: 2, Y: 4}, 0.5, 0.5)

	assert.Equal(t, 2, state.CursorIndex.Peek(), "moving without a button held doesn't move the cursor")
}

func TestListClick_NonFocusableListTakesClicks(t *testing.T) {
	state := NewListState(clickSceneItems(4))
	var selected []string
	list := List[string]{
		ID: "list", DisableFocus: true, State: state,
		OnSelect: func(item string) { selected = append(selected, item) },
	}
	input := TextInput{ID: "input", State: NewTextInputState("")}
	scene := newClickScene(t, Column{Children: []Widget{input, list}}, 20, 5)
	scene.focus.FocusByID("input")
	scene.draw()

	scene.click(2, 3, 0) // Item 02, below the input.
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, "input", scene.focus.FocusedID(), "a list that can't be focused leaves focus where it was")

	scene.click(2, 3, 0)
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
	scene := newClickScene(t, tree, 30, 8)

	scene.click(8, 1, 0) // main.go
	assert.Empty(t, selected, "a single click only moves the cursor")
	node, _ := state.CursorNode()
	assert.Equal(t, "main.go", node)

	scene.click(8, 1, 0)
	assert.Equal(t, []string{"main.go"}, selected)
}

func TestTreeClick_IndicatorTogglesWithoutSelecting(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	var selected int
	tree := Tree[string]{ID: "tree", State: state, OnSelect: func(string, []string) { selected++ }}
	scene := newClickScene(t, tree, 30, 8)
	collapsed := state.IsCollapsed([]int{0})

	scene.click(0, 0, 0) // The expand indicator on "src".
	assert.NotEqual(t, collapsed, state.IsCollapsed([]int{0}))
	scene.click(0, 0, 0)
	assert.Equal(t, collapsed, state.IsCollapsed([]int{0}), "each click on the indicator toggles")
	assert.Zero(t, selected, "double-clicking the indicator doesn't select")
}

func TestTreeDrag_MultiSelectSelectsRange(t *testing.T) {
	state := NewTreeState(mouseTreeNodes())
	tree := Tree[string]{ID: "tree", State: state, MultiSelect: true}
	scene := newClickScene(t, tree, 30, 8)
	require.Len(t, state.viewPaths, 6, "every folder starts expanded")

	scene.drag(8, 1, 8, 3)

	assert.Equal(t, [][]int{{0, 0}, {0, 1}, {1}}, state.SelectedPaths())
	scene.snapshot("TestTreeDrag_MultiSelectSelectsRange", "Dragging from main.go down to docs selects main.go, util.go and docs")
}

func TestDirectoryTreeClick_IndicatorLoadsChildren(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
	state := NewDirectoryTreeState(root)
	tree := DirectoryTree{Tree: Tree[DirectoryEntry]{ID: "dirs", State: state}}
	scene := newClickScene(t, tree, 30, 5)
	require.Len(t, state.viewPaths, 1)

	scene.click(0, 0, 0) // The root's expand indicator.
	require.Eventually(t, func() bool {
		scene.draw()
		return len(state.viewPaths) == 2
	}, time.Second, 10*time.Millisecond, "clicking the indicator loads the directory")

	scene.click(4, 1, 0)
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
	scene := newClickScene(t, table, 20, 7)

	scene.click(6, 3, 0) // Row r2, column A (columns start at x 0, 5, 10; the header takes row 0).
	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, 1, state.CursorColumn.Peek())
	assert.Equal(t, []string{"r2"}, changes)
	assert.Equal(t, "table", scene.focus.FocusedID())
	scene.snapshot("TestTableClick_MovesCursorToClickedCell", "Clicking cell a2 puts the cursor on row r2, column A")

	scene.click(4, 1, 0) // The gap after the Id column belongs to it.
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 0, state.CursorColumn.Peek())

	scene.click(6, 0, 0) // The header row isn't a row of data.
	assert.Equal(t, 0, state.CursorIndex.Peek())
	assert.Equal(t, 0, state.CursorColumn.Peek())
}

func TestTableClick_DoubleClickSelectsRow(t *testing.T) {
	state := NewTableState(mouseTableRows())
	var selected []string
	table := mouseTable(state, TableSelectionRow, false)
	table.OnSelect = func(row []string) { selected = append(selected, row[0]) }
	scene := newClickScene(t, table, 20, 7)

	scene.click(11, 4, 0)
	assert.Empty(t, selected)
	scene.click(11, 4, 0)
	assert.Equal(t, []string{"r3"}, selected)
}

func TestTableClick_ShiftClickExtendsRowSelection(t *testing.T) {
	state := NewTableState(mouseTableRows())
	scene := newClickScene(t, mouseTable(state, TableSelectionRow, true), 20, 7)

	scene.click(1, 1, 0)
	scene.click(1, 3, uv.ModShift)

	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, selectionSet(0, 1, 2), state.Selection.Peek())
}

func TestTableDrag_CellModeSelectsBox(t *testing.T) {
	state := NewTableState(mouseTableRows())
	scene := newClickScene(t, mouseTable(state, TableSelectionCursor, true), 20, 7)

	scene.drag(6, 1, 11, 3) // From a0 to b2.

	assert.Equal(t, 2, state.CursorIndex.Peek())
	assert.Equal(t, 2, state.CursorColumn.Peek())
	assert.Equal(t, selectionSet(
		cellIndex(0, 1, 3), cellIndex(0, 2, 3),
		cellIndex(1, 1, 3), cellIndex(1, 2, 3),
		cellIndex(2, 1, 3), cellIndex(2, 2, 3),
	), state.Selection.Peek())
	scene.snapshot("TestTableDrag_CellModeSelectsBox", "Dragging from a0 to b2 selects the box of cells between them")
}

func TestTableDrag_ColumnModeSelectsColumns(t *testing.T) {
	state := NewTableState(mouseTableRows())
	scene := newClickScene(t, mouseTable(state, TableSelectionColumn, true), 20, 7)

	scene.drag(1, 2, 11, 2)

	assert.Equal(t, 2, state.CursorColumn.Peek())
	assert.Equal(t, selectionSet(0, 1, 2), state.Selection.Peek())
}

func TestTableDrag_RowModeMovesCursorWithoutMultiSelect(t *testing.T) {
	state := NewTableState(mouseTableRows())
	scene := newClickScene(t, mouseTable(state, TableSelectionRow, false), 20, 7)

	scene.drag(1, 1, 1, 9) // Past the last row.

	assert.Equal(t, 4, state.CursorIndex.Peek())
	assert.Empty(t, state.Selection.Peek())
}

func mouseMenuScene(t *testing.T, onSelect func(MenuItem)) (*clickScene, *MenuState) {
	state := NewMenuState([]MenuItem{
		{Label: "Open"},
		{Label: "Save"},
		{Divider: "More"},
		{Label: "Export", Disabled: true},
		{Label: "Close"},
	})
	menu := Menu{ID: "menu", State: state, Position: FloatPositionTopLeft, OnSelect: onSelect}
	scene := newClickScene(t, menu, 20, 7)
	scene.focus.FocusByID("menu")
	scene.draw()
	return scene, state
}

func TestMenuClick_HighlightsAndDoubleClickSelects(t *testing.T) {
	var selected []string
	scene, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	scene.click(2, 1, 0)
	assert.Equal(t, 1, state.CursorIndex())
	assert.Empty(t, selected, "a single click only moves the cursor")
	scene.snapshot("TestMenuClick_HighlightsAndDoubleClickSelects", "Clicking Save highlights it")

	scene.click(2, 1, 0)
	assert.Equal(t, []string{"Save"}, selected)
}

func TestMenuClick_IgnoresDividersAndDisabledItems(t *testing.T) {
	var selected []string
	scene, state := mouseMenuScene(t, func(item MenuItem) { selected = append(selected, item.Label) })

	scene.click(2, 2, 0)
	scene.click(2, 3, 0)
	scene.click(2, 3, 0)

	assert.Equal(t, 0, state.CursorIndex())
	assert.Empty(t, selected)
}

func TestMenuDrag_MovesCursor(t *testing.T) {
	scene, state := mouseMenuScene(t, nil)

	scene.press(2, 0, 0)
	scene.move(2, 3) // Disabled: the cursor stays put.
	assert.Equal(t, 0, state.CursorIndex())
	scene.move(2, 4)
	scene.release(2, 4)
	assert.Equal(t, 4, state.CursorIndex())
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
	scene := newClickScene(t, palette, 50, 12)
	scene.draw()
	inputID := palette.inputID()
	require.Equal(t, inputID, scene.focus.FocusedID())

	listEntry := scene.renderer.WidgetByID(palette.listID())
	require.NotNil(t, listEntry)
	x, y := listEntry.Bounds.X+2, listEntry.Bounds.Y

	scene.click(x, y+2, 0) // Cut.
	level := state.CurrentLevel()
	assert.Equal(t, 2, level.ListState.CursorIndex.Peek())
	assert.Equal(t, inputID, scene.focus.FocusedID(), "clicking a result leaves focus in the input")
	scene.snapshot("TestCommandPaletteClick_KeepsInputFocusAndDoubleClickRuns", "Clicking Cut highlights it while the input keeps focus")

	scene.click(x, y+1, 0) // The divider can't take the cursor.
	scene.click(x, y+3, 0) // Nor can a disabled item.
	scene.click(x, y+3, 0)
	assert.Equal(t, 2, level.ListState.CursorIndex.Peek())
	assert.Empty(t, ran)

	scene.click(x, y+2, 0)
	scene.click(x, y+2, 0)
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
	scene := newClickScene(t, ac, 30, 8)
	scene.focus.FocusByID("input")
	acState.Show()
	scene.draw()

	listEntry := scene.renderer.WidgetByID("ac-list")
	require.NotNil(t, listEntry)
	x, y := listEntry.Bounds.X+1, listEntry.Bounds.Y

	scene.click(x, y+1, 0)
	assert.Equal(t, "input", scene.focus.FocusedID())
	assert.True(t, acState.IsVisible(), "clicking a suggestion keeps the popup open")
	suggestion, _ := acState.SelectedSuggestion()
	assert.Equal(t, "apricot", suggestion.Label)

	scene.click(x, y+1, 0)
	assert.Equal(t, "apricot", inputState.GetText())
}
