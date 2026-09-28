package terma

import (
	"reflect"
	"testing"
)

func newModeTableState(mode TableSelectionMode, columnCount int, rows []string, selected ...int) *TableState[string] {
	state := NewTableState(rows)
	state.syncSelectionMode(mode, columnCount)
	for _, key := range selected {
		state.Select(key)
	}
	return state
}

func assertSelectedKeys(t *testing.T, state *TableState[string], want ...int) {
	t.Helper()
	got := state.SelectedIndices()
	if len(want) == 0 {
		want = []int{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selected keys = %v, want %v", got, want)
	}
}

func TestTableState_RowSelectionFollowsInsertAndRemove(t *testing.T) {
	state := newModeTableState(TableSelectionRow, 3, []string{"a", "b", "c", "d"}, 1, 3)
	state.SetAnchor(3)

	state.Prepend("new")
	assertSelectedKeys(t, state, 2, 4) // b, d
	if got := state.GetAnchor(); got != 4 {
		t.Fatalf("anchor = %d, want 4", got)
	}

	state.RemoveAt(2)               // b
	assertSelectedKeys(t, state, 3) // d

	state.InsertAt(3, "x")
	assertSelectedKeys(t, state, 4) // d
}

func TestTableState_CellSelectionFollowsRowEdits(t *testing.T) {
	const columns = 3
	// Select row 1 col 2, and row 2 col 0.
	state := newModeTableState(TableSelectionCursor, columns, []string{"a", "b", "c"},
		cellIndex(1, 2, columns), cellIndex(2, 0, columns))

	state.Prepend("new")
	assertSelectedKeys(t, state, cellIndex(2, 2, columns), cellIndex(3, 0, columns))

	state.RemoveAt(2) // b
	assertSelectedKeys(t, state, cellIndex(2, 0, columns))
}

func TestTableState_ColumnSelectionIgnoresRowEdits(t *testing.T) {
	state := newModeTableState(TableSelectionColumn, 3, []string{"a", "b"}, 0, 2)

	state.Prepend("new")
	state.RemoveAt(1)

	assertSelectedKeys(t, state, 0, 2)
}

func TestTableState_RemoveWhereKeepsSelectionOnRemainingRows(t *testing.T) {
	state := newModeTableState(TableSelectionRow, 1, []string{"a", "x", "b", "x", "c"}, 1, 4)

	removed := state.RemoveWhere(func(row string) bool { return row == "x" })

	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}
	assertSelectedKeys(t, state, 2) // c
}

func TestTableState_InsertIntoEmptyTableKeepsCursorInRange(t *testing.T) {
	for name, insert := range map[string]func(*TableState[string]){
		"prepend":  func(s *TableState[string]) { s.Prepend("only") },
		"insertAt": func(s *TableState[string]) { s.InsertAt(0, "only") },
	} {
		t.Run(name, func(t *testing.T) {
			state := NewTableState([]string{})
			insert(state)
			if got := state.CursorIndex.Peek(); got != 0 {
				t.Fatalf("cursor = %d, want 0", got)
			}
		})
	}
}

func TestTableState_ClearResetsSelection(t *testing.T) {
	state := newModeTableState(TableSelectionRow, 1, []string{"a", "b"}, 0, 1)
	state.SetAnchor(0)

	state.Clear()

	assertSelectedKeys(t, state)
	if state.HasAnchor() {
		t.Fatal("anchor should be cleared")
	}
}

func TestTableState_SetRowsDropsSelectionPastEnd(t *testing.T) {
	state := newModeTableState(TableSelectionRow, 1, []string{"a", "b", "c"}, 0, 2)

	state.SetRows([]string{"x", "y"})

	assertSelectedKeys(t, state, 0)
}

func TestTable_ColumnModeMovesBetweenColumnsWithArrows(t *testing.T) {
	state := NewTableState([]string{"a"})
	table := Table[string]{
		State:         state,
		Columns:       []TableColumn{{}, {}, {}},
		SelectionMode: TableSelectionColumn,
	}

	press := func(key string) {
		t.Helper()
		for _, bind := range table.Keybinds() {
			if bind.Key == key {
				bind.Action()
				return
			}
		}
		t.Fatalf("no %q keybind in column mode", key)
	}

	press("right")
	press("l")
	if got := state.CursorColumn.Peek(); got != 2 {
		t.Fatalf("cursor column = %d, want 2", got)
	}
	press("left")
	if got := state.CursorColumn.Peek(); got != 1 {
		t.Fatalf("cursor column = %d, want 1", got)
	}
}

func TestTableState_ColumnCountChangeClearsSelection(t *testing.T) {
	state := newModeTableState(TableSelectionCursor, 3, []string{"a", "b"}, cellIndex(1, 1, 3))

	state.syncSelectionMode(TableSelectionCursor, 4)

	assertSelectedKeys(t, state)
}
