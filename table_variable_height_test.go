package terma

import "testing"

func TestTableScrollIntoView_UsesLayoutMetrics(t *testing.T) {
	state := NewTableState([]string{"A", "B", "C"})
	state.SelectIndex(1)

	scrollState := NewScrollState()
	scrollState.updateLayout(3, 10)

	state.rowLayouts = []tableRowLayout{
		{y: 0, height: 2},
		{y: 2, height: 4},
		{y: 6, height: 1},
	}

	table := Table[string]{
		State:       state,
		ScrollState: scrollState,
		RowHeight:   1,
		Columns:     []TableColumn{{}},
	}

	table.scrollCursorIntoView()

	if got := scrollState.GetOffset(); got != 3 {
		t.Fatalf("expected scroll offset 3, got %d", got)
	}
}
