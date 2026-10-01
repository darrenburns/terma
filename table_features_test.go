package terma

import (
	"cmp"
	"fmt"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"slices"
	"strings"
	"testing"
)

type featureRow struct {
	ID, Name string
	Age      int
}

func featureRows() []featureRow {
	return []featureRow{{"a", "Zoe", 30}, {"b", "Ada", 20}, {"c", "Mia", 20}, {"d", "Leo", 40}, {"e", "Uma", 25}, {"f", "Kai", 35}, {"g", "Eli", 28}, {"h", "Bea", 22}}
}
func featureTable() Table[featureRow] {
	return Table[featureRow]{ID: "features", State: NewTableStateWithRowID(featureRows(), func(r featureRow) string { return r.ID }),
		Columns:      []TableColumn{{ID: "name", Width: Cells(12), Header: Text{Content: "Name"}, Resizable: true, MinWidth: 4, MaxWidth: 20}, {ID: "age", Width: Cells(9), Header: Text{Content: "Age"}, Resizable: true}, {ID: "note", Width: Cells(18), Header: Text{Content: "Note"}}},
		Comparators:  map[string]func(featureRow, featureRow) int{"name": func(a, b featureRow) int { return strings.Compare(a.Name, b.Name) }, "age": func(a, b featureRow) int { return cmp.Compare(a.Age, b.Age) }},
		FrozenHeader: true, FrozenColumns: 1, MultiSelect: true, SelectionMode: TableSelectionRow, ColumnSpacing: 1,
		Style: Style{Width: Flex(1), Height: Flex(1)},
		RenderCell: func(r featureRow, _, col int, active, selected bool) Widget {
			value := r.Name
			if col == 1 {
				value = fmt.Sprint(r.Age)
			}
			if col == 2 {
				value = "record-" + r.ID + "-details"
			}
			style := Style{}
			if active {
				style.BackgroundColor = Blue
			}
			if selected {
				style.Bold = true
			}
			return Text{Content: value, Style: style}
		},
	}
}
func TestTableFeaturesSorting(t *testing.T) {
	table := featureTable()
	original := slices.Clone(table.State.GetRows())
	scene := newWheelScene(t, table, 40, 6)
	table.State.SelectIndex(0)
	table.State.Select(0)
	table.cycleSort(1)
	scene.draw()
	assert.Equal(t, []int{1, 2, 7, 4, 6, 0, 5, 3}, table.State.viewIndices)
	assert.Equal(t, original, table.State.GetRows())
	assert.Equal(t, 0, table.State.CursorIndex.Peek())
	assert.True(t, table.State.IsSelected(0))
	table.cycleSort(1)
	scene.draw()
	assert.Equal(t, []int{3, 5, 0, 6, 4, 7, 1, 2}, table.State.viewIndices)
	table.cycleSort(1)
	scene.draw()
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, table.State.viewIndices)
	table.Columns[1].ID = "name"
	assert.False(t, table.sortableColumn(0))
	table.cycleSort(0)
	assert.Equal(t, TableSort{}, table.State.Sort.Peek())
	table.Columns[0].ID = ""
	assert.False(t, table.sortableColumn(0))
}
func TestTableFeaturesRowIdentity(t *testing.T) {
	for _, mode := range []TableSelectionMode{TableSelectionRow, TableSelectionCursor, TableSelectionColumn} {
		t.Run(fmt.Sprint(mode), func(t *testing.T) {
			state := NewTableStateWithRowID(featureRows(), func(r featureRow) string { return r.ID })
			state.syncSelectionMode(mode, 3)
			state.SelectIndex(1)
			key := 1
			if mode == TableSelectionCursor {
				key = cellIndex(1, 2, 3)
			}
			state.Select(key)
			state.SetAnchor(key)
			replacement := slices.Clone(featureRows())
			slices.Reverse(replacement)
			state.SetRows(replacement)
			assert.Equal(t, 6, state.CursorIndex.Peek())
			expected := 6
			if mode == TableSelectionCursor {
				expected = cellIndex(6, 2, 3)
			}
			if mode == TableSelectionColumn {
				expected = 1
			}
			assert.Equal(t, []int{expected}, state.SelectedIndices())
			assert.Equal(t, expected, state.GetAnchor())
		})
	}
	t.Run("ambiguous identities", func(t *testing.T) {
		state := NewTableStateWithRowID([]featureRow{{ID: "a"}, {ID: "dup"}, {ID: "dup"}, {ID: ""}}, func(r featureRow) string { return r.ID })
		state.SelectIndex(1)
		state.SelectAll()
		state.SetAnchor(1)
		state.SetRows([]featureRow{{ID: "dup"}, {ID: "a"}, {ID: ""}})
		assert.Equal(t, 1, state.CursorIndex.Peek())
		assert.Equal(t, []int{1}, state.SelectedIndices())
		assert.False(t, state.HasAnchor())
		state.SetRows(nil)
		assert.Equal(t, 0, state.CursorIndex.Peek())
		assert.Empty(t, state.SelectedIndices())
	})
	t.Run("new duplicates", func(t *testing.T) {
		state := NewTableStateWithRowID(featureRows(), func(r featureRow) string { return r.ID })
		state.Select(0)
		state.SetRows([]featureRow{{ID: "a"}, {ID: "a"}})
		assert.Empty(t, state.SelectedIndices())
	})
}
func TestTableFeaturesFilterAndSelection(t *testing.T) {
	table := featureTable()
	table.Filter = NewFilterState()
	table.MatchCell = func(r featureRow, _, col int, q string, opts FilterOptions) MatchResult {
		if col == 0 {
			return MatchString(r.Name, q, opts)
		}
		return MatchResult{}
	}
	selected := 0
	table.OnSelect = func(featureRow) { selected++ }
	scene := newWheelScene(t, table, 40, 6)
	table.State.SelectIndex(0)
	table.State.Select(0)
	table.Filter.Query.Set("Ada")
	scene.draw()
	assert.Equal(t, 0, table.State.CursorIndex.Peek())
	assert.True(t, table.State.IsSelected(0))
	table.Filter.Query.Set("")
	scene.draw()
	assert.Equal(t, 0, table.State.CursorIndex.Peek())
	table.Filter.Query.Set("NO SUCH ROW")
	scene.draw()
	table.selectRow()
	assert.Zero(t, selected)
	table.Filter.Query.Set("")
	table.State.Sort.Set(TableSort{ColumnID: "age", Direction: TableSortAscending})
	scene.draw()
	table.keyCursorToFirst()
	table.shiftRowDown()
	assert.Equal(t, []int{1, 2}, table.State.SelectedIndices())
}
func TestTableFeaturesResizeAndMouse(t *testing.T) {
	table := featureTable()
	scene := newWheelScene(t, table, 30, 6)
	table.OnMouseDown(MouseEvent{X: 11, LocalX: 11, LocalY: 0, Button: uv.MouseLeft})
	require.True(t, table.State.resizing)
	table.OnMouseMove(MouseEvent{X: 100})
	table.OnMouseUp(MouseEvent{})
	scene.draw()
	assert.Equal(t, 20, table.State.columnLayouts[0].width)
	assert.Equal(t, TableSort{}, table.State.Sort.Peek())
	assert.Empty(t, table.State.SelectedIndices())
	table.resizeColumn(0, -30)
	scene.draw()
	assert.Equal(t, 4, table.State.columnLayouts[0].width)
	table.resetCurrentWidth()
	scene.draw()
	assert.Equal(t, 12, table.State.columnLayouts[0].width)
	table.OnMouseDown(MouseEvent{LocalX: 2, LocalY: 0, Button: uv.MouseLeft})
	table.OnMouseUp(MouseEvent{})
	scene.draw()
	assert.Equal(t, TableSortAscending, table.State.Sort.Peek().Direction)
	table.viewportState().SetOffset(0)
	scene.draw()
	table.OnMouseDown(MouseEvent{LocalX: 2, LocalY: 1, Button: uv.MouseLeft})
	table.OnMouseUp(MouseEvent{})
	assert.Equal(t, "Ada", table.CursorRow().Name)
}
func TestTableFeaturesFrozenSnapshots(t *testing.T) {
	table := featureTable()
	scene := newWheelScene(t, table, 30, 6)
	scene.snapshot("TestTableFeatures_initial", "Frozen Name column and header, bounded 30×6 viewport")
	scene.wheel(uv.MouseWheelDown, 3)
	assert.Equal(t, 3, table.viewportState().GetOffset())
	scene.snapshot("TestTableFeatures_vertical", "Header stays at top while body scrolls three rows")
	table.viewportState().ScrollRight(8)
	scene.draw()
	scene.snapshot("TestTableFeatures_both_axes", "Name remains fixed while age/note pane shifts horizontally; header stays fixed vertically")
	table.OnMouseDown(MouseEvent{LocalX: 2, LocalY: 1, Button: uv.MouseLeft})
	table.OnMouseUp(MouseEvent{})
	assert.Equal(t, "Leo", table.CursorRow().Name)
	table.State.Sort.Set(TableSort{ColumnID: "age", Direction: TableSortAscending})
	scene.draw()
	scene.snapshot("TestTableFeatures_ascending", "Age ascending indicator; cursor stays on Leo after sorting")
	table.State.Sort.Set(TableSort{ColumnID: "age", Direction: TableSortDescending})
	scene.draw()
	scene.snapshot("TestTableFeatures_descending", "Age descending with same selected logical record")
	table.resizeColumn(0, 6)
	scene.draw()
	scene.snapshot("TestTableFeatures_resized", "First column shrinks; scroll offsets are clamped and other columns use the released space")
	table.State.SetRows(nil)
	scene.draw()
	scene.snapshot("TestTableFeatures_empty", "Empty data still shows sortable headers")
	for _, size := range [][2]int{{4, 3}, {1, 1}, {15, 2}} {
		table := featureTable()
		table.FrozenColumns = 99
		AssertSnapshotNamed(t, fmt.Sprintf("TestTableFeatures_tiny_%dx%d", size[0], size[1]), table, size[0], size[1], "Oversized frozen count and narrow viewport safely clip cells")
	}
}
func TestTableFeaturesReactiveSequence(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 6, featureTable)
	sequence.frame("initial", nil)
	sequence.frame("sort ascending", func(t Table[featureRow]) { t.cycleSort(1) })
	sequence.frame("select and move", func(t Table[featureRow]) { t.keyCursorToFirst(); t.shiftRowDown() })
	sequence.frame("scroll vertical", func(t Table[featureRow]) { t.viewportState().ScrollDown(3) })
	sequence.frame("scroll horizontal", func(t Table[featureRow]) { t.viewportState().ScrollRight(8) })
	sequence.frame("resize frozen column", func(t Table[featureRow]) { t.resizeColumn(0, 6) })
	sequence.frame("identity replacement", func(t Table[featureRow]) {
		rows := slices.Clone(t.State.GetRows())
		slices.Reverse(rows)
		t.State.SetRows(rows)
	})
	sequence.frame("empty", func(t Table[featureRow]) { t.State.SetRows(nil) })
}

func TestTableFeaturesControlsValidation(t *testing.T) {
	table := featureTable()
	scene := newWheelScene(t, table, 30, 6)
	table.State.Sort.Set(TableSort{ColumnID: "missing", Direction: TableSortAscending})
	scene.draw()
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7}, table.State.viewIndices)
	table.resizeColumn(0, 18)
	scene.draw()
	table.Columns[0], table.Columns[1] = table.Columns[1], table.Columns[0]
	scene.root = table
	scene.renderer.Render(table)
	assert.Equal(t, 18, table.State.columnLayouts[1].width)
	table.Columns[0].ID = table.Columns[1].ID
	table.resizeColumn(0, 99)
	assert.Equal(t, 18, table.State.ColumnWidths.Peek()["name"])
	table.Columns[0].ID = "age"
	table.Columns[0].MinWidth = 8
	table.Columns[0].MaxWidth = 2
	table.resizeColumn(0, 100)
	assert.Equal(t, 8, table.State.ColumnWidths.Peek()["age"])
}
func TestTableFeaturesDefaultRendererVariableHeight(t *testing.T) {
	table := Table[[]string]{ID: "variable", State: NewTableState([][]string{{"A", "one\ntwo"}, {"B", "three\nfour\nfive"}, {"C", "six"}}), Columns: []TableColumn{{ID: "name", Width: Cells(6), Header: Text{Content: "Name"}}, {ID: "value", Width: Cells(12), Header: Text{Content: "Description"}}}, FrozenHeader: true, FrozenColumns: 1, Style: Style{Width: Flex(1), Height: Flex(1)}, ColumnSpacing: 1}
	scene := newWheelScene(t, table, 17, 5)
	scene.snapshot("TestTableFeatures_variable_height", "Multiline default cells share row heights in frozen viewport")
	scene.wheel(uv.MouseWheelDown, 2)
	table.viewportState().ScrollRight(2)
	scene.draw()
	scene.snapshot("TestTableFeatures_variable_height_scrolled", "Partially visible multiline row clips under header and fixed Name column")
	table.OnMouseDown(MouseEvent{LocalX: 1, LocalY: 1, Button: uv.MouseLeft})
	table.OnMouseUp(MouseEvent{})
	assert.Equal(t, 1, table.State.CursorIndex.Peek())
	table.keyCursorToLast()
	scene.draw()
	assert.Equal(t, 2, table.State.CursorIndex.Peek())
	assert.Equal(t, 2, table.viewportState().GetOffset())
}
func TestTableFeaturesFrozenHeaderOptional(t *testing.T) {
	table := featureTable()
	table.FrozenHeader = false
	scene := newWheelScene(t, table, 30, 5)
	scene.wheel(uv.MouseWheelDown, 2)
	scene.snapshot("TestTableFeatures_columns_only", "Frozen column remains while header scrolls away with body")
	table.OnMouseDown(MouseEvent{LocalX: 1, LocalY: 0, Button: uv.MouseLeft})
	table.OnMouseUp(MouseEvent{})
	assert.Equal(t, "Ada", table.CursorRow().Name)
}

func TestTableFeaturesKeyboardRevealsColumns(t *testing.T) {
	table := featureTable()
	table.SelectionMode = TableSelectionCursor
	scene := newWheelScene(t, table, 24, 5)
	table.keyCursorRight()
	scene.draw()
	assert.Equal(t, 0, table.viewportState().GetOffsetX())
	table.keyCursorRight()
	scene.draw()
	assert.Positive(t, table.viewportState().GetOffsetX())
	table.keyCursorLeft()
	scene.draw()
	assert.Equal(t, 0, table.viewportState().GetOffsetX())
	table.FrozenColumns = 0
	table.FrozenHeader = true
	table.State.SetRows(nil)
	table.Columns[0].Header = nil
	table.Columns[1].Header = nil
	table.Columns[2].Header = nil
	scene.root = table
	scene.renderer.Render(table)
	assert.Zero(t, table.viewportState().GetOffset())
	assert.Zero(t, table.viewportState().GetOffsetX())
}

func TestTableFeaturesSortRevealSameFrame(t *testing.T) {
	table := featureTable()
	scene := newWheelScene(t, Column{Style: Style{Height: Flex(1)}, Children: []Widget{table}}, 30, 4)
	table.cycleSort(0)
	scene.draw()
	require.Positive(t, table.viewportState().GetOffset())
	// The cursor's highlight must be in the viewport on this very frame,
	// rather than waiting for another key or wheel event.
	found := false
	for y := 1; y < 4; y++ {
		for x := 0; x < 12; x++ {
			cell := scene.buf.CellAt(x, y)
			if cell != nil && cell.Content == "Z" {
				found = true
			}
		}
	}
	assert.True(t, found, "Zoe must be visible immediately after sorting by Name")
}

func TestTableFeaturesPanBindingsAndWheel(t *testing.T) {
	table := featureTable()
	scene := newWheelScene(t, table, 24, 5)
	require.True(t, matchKeybind(makeKeyEvent(uv.KeyRight, uv.ModAlt), table.Keybinds()))
	scene.draw()
	assert.Equal(t, 3, table.viewportState().GetOffsetX())
	require.True(t, table.OnMouseWheel(MouseEvent{Button: uv.MouseWheelRight}))
	scene.draw()
	assert.Equal(t, 4, table.viewportState().GetOffsetX())
	require.True(t, matchKeybind(makeKeyEvent(uv.KeyLeft, uv.ModAlt), table.Keybinds()))
	scene.draw()
	assert.Equal(t, 1, table.viewportState().GetOffsetX())
	require.True(t, table.OnMouseWheel(MouseEvent{Button: uv.MouseWheelLeft}))
	scene.draw()
	assert.Zero(t, table.viewportState().GetOffsetX())
}
