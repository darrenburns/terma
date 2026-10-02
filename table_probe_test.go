package terma

import (
	"fmt"
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
	"slices"
	"strings"
	"testing"
)

func tableProbeText(buffer *uv.Buffer) string {
	var s strings.Builder
	for y := 0; y < buffer.Height(); y++ {
		for x := 0; x < buffer.Width(); x++ {
			s.WriteString(buffer.CellAt(x, y).Content)
		}
		s.WriteByte('\n')
	}
	return s.String()
}

func TestTableProbeTallRowCanScrollToItsEnd(t *testing.T) {
	scroll := NewScrollState()
	table := Table[[]string]{ID: "tall", State: NewTableState([][]string{{"A\nB\nC"}}), ScrollState: scroll,
		Columns: []TableColumn{{ID: "text", Header: Text{Content: "Header"}, Width: Cells(10)}}, FrozenHeader: true, Style: Style{Width: Flex(1), Height: Flex(1)}}
	scene := newClickScene(t, table, 14, 2)
	scroll.SetOffset(100)
	t.Logf("requested offset100, effective offset%d", scroll.GetOffset())
	scene.draw()
	require.Contains(t, tableProbeText(scene.buf), "C", "scrolling must reach the final line of a row taller than the viewport")
	require.Contains(t, tableProbeText(scene.buf), "Header", "header remains frozen")
	scene.snapshot("TableProbe_tall_row_end", "A row taller than the frozen viewport retains its last line for scrolling")
}

func TestTableProbeSortedFilteredNavigationPreservesIdentity(t *testing.T) {
	table := featureTable()
	table.SelectionMode = TableSelectionRow
	table.Filter = NewFilterState()
	table.Filter.Query.Set("a")
	table.MatchCell = func(r featureRow, _, col int, q string, o FilterOptions) MatchResult {
		if col == 0 {
			return MatchString(r.Name, q, o)
		}
		return MatchResult{}
	}
	table.State.Sort.Set(TableSort{ColumnID: "age", Direction: TableSortAscending})
	var opened []string
	table.OnSelect = func(r featureRow) { opened = append(opened, r.ID) }
	scene := newClickScene(t, table, 30, 5)
	key := func(code rune, mod uv.KeyMod) {
		require.True(t, scene.focus.HandleKey(makeKeyEvent(code, mod)))
		scene.draw()
	}
	key(uv.KeyHome, 0)
	key(uv.KeyEnter, 0)
	require.Equal(t, []string{"b"}, opened, "Ada sorts first among matching names")
	key(uv.KeyDown, uv.ModShift)
	var selected []string
	for _, r := range table.State.SelectedRows() {
		selected = append(selected, r.ID)
	}
	require.Equal(t, []string{"b", "c"}, selected, "shift selection follows the displayed tied rows")
	replacement := slices.Clone(table.State.GetRows())
	slices.Reverse(replacement)
	table.State.SetRows(replacement)
	scene.draw()
	require.Equal(t, "c", table.CursorRow().ID, "Mia cursor follows its stable ID")
	key(uv.KeyHome, 0)
	key(uv.KeyEnter, 0)
	require.Equal(t, []string{"b", "c"}, opened, "ties now follow reversed source order")
	table.Filter.Query.Set("no-match")
	scene.draw()
	key(uv.KeyEnter, 0)
	require.Len(t, opened, 2)
	require.Equal(t, "c", table.CursorRow().ID, "hidden cursor remains stable until view interaction can normalize it")
	table.Filter.Query.Set("")
	scene.draw()
	key(uv.KeyEnter, 0)
	require.Equal(t, []string{"b", "c", "c"}, opened)
}

func TestTableProbePaddedMouseResizeAndSort(t *testing.T) {
	table := featureTable()
	table.Style.Padding = EdgeInsetsAll(1)
	table.Style.Border = RoundedBorder(Blue)
	table.Style.Width, table.Style.Height = Cells(30), Cells(5)
	scene := newClickScene(t, Column{Style: Style{Padding: EdgeInsetsAll(1)}, Children: []Widget{Text{Content: "outside"}, table}}, 40, 12)
	// Outer padding + heading + table border/padding puts the header at y4, x3.
	scene.click(5, 4, 0)
	require.Equal(t, TableSort{ColumnID: "name", Direction: TableSortAscending}, table.State.Sort.Peek())
	table.State.Select(0)
	scene.press(14, 4, 0)
	scene.move(19, 4)
	scene.release(19, 4)
	require.Equal(t, 17, table.State.ColumnWidths.Peek()["name"])
	require.Equal(t, TableSortAscending, table.State.Sort.Peek().Direction, "resize does not re-sort")
	require.True(t, table.State.IsSelected(0), "resize preserves selection")
	scene.snapshot("TableProbe_padded_resize", "Actual mouse router sorts and resizes the intended padded header without changing selection")
}

type tableProbeRetainedApp struct {
	table   Table[featureRow]
	padding Signal[int]
	frozen  Signal[bool]
}

func (a *tableProbeRetainedApp) Build(BuildContext) Widget {
	table := a.table
	table.FrozenHeader = a.frozen.Get()
	table.Style.Padding = EdgeInsetsAll(a.padding.Get())
	return Column{Width: Flex(1), Height: Flex(1), Children: []Widget{Text{Content: "outside"}, table}}
}
func TestTableProbeRetainedResizeAndDynamicRows(t *testing.T) {
	seq := newReactivitySequence(t, 38, 10, func() *tableProbeRetainedApp {
		table := featureTable()
		table.ScrollState = NewScrollState()
		table.Filter = NewFilterState()
		return &tableProbeRetainedApp{table, NewSignal(0), NewSignal(true)}
	})
	seq.frame("initial", nil)
	seq.frame("padding", func(a *tableProbeRetainedApp) { a.padding.Set(1) })
	seq.frame("sorted", func(a *tableProbeRetainedApp) {
		a.table.State.Sort.Set(TableSort{ColumnID: "age", Direction: TableSortDescending})
	})
	seq.frame("width override", func(a *tableProbeRetainedApp) { a.table.State.ColumnWidths.Set(map[string]int{"name": 5}) })
	seq.frame("pan", func(a *tableProbeRetainedApp) { a.table.ScrollState.ScrollRight(8); a.table.ScrollState.ScrollDown(2) })
	seq.frame("reverse", func(a *tableProbeRetainedApp) {
		rows := slices.Clone(a.table.State.GetRows())
		slices.Reverse(rows)
		a.table.State.SetRows(rows)
	})
	seq.frame("empty", func(a *tableProbeRetainedApp) { a.table.State.SetRows(nil) })
	seq.frame("restore", func(a *tableProbeRetainedApp) { a.table.State.SetRows(featureRows()) })
	for _, size := range [][2]int{{1, 1}, {4, 3}, {12, 5}, {50, 14}} {
		seq.resize(size[0], size[1])
		seq.frame(fmt.Sprintf("resize%dx%d", size[0], size[1]), nil)
	}
	seq.frame("header scrolls", func(a *tableProbeRetainedApp) { a.frozen.Set(false) })
	seq.frame("header frozen", func(a *tableProbeRetainedApp) { a.frozen.Set(true) })
}

func FuzzTableProbeRowsAndViewport(f *testing.F) {
	for _, seed := range [][]byte{{0}, {1, 2, 3, 4}, {7, 255, 0, 32, 9}, {5, 1, 1, 1, 1, 8, 2}} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 40 {
			t.Skip()
		}
		rows := make([][]string, min(len(data), 8))
		for i := range rows {
			rows[i] = []string{fmt.Sprintf("row%d", i), strings.Repeat("line\n", int(data[i]%8)) + "tail"}
		}
		original := slices.Clone(rows)
		state := NewTableStateWithRowID(rows, func(r []string) string { return r[0] })
		table := Table[[]string]{ID: "fuzz", State: state, Columns: []TableColumn{{ID: "id", Width: Cells(4), Header: Text{Content: "ID"}}, {ID: "value", Width: Cells(9), Header: Text{Content: "Value"}}},
			Comparators: map[string]func([]string, []string) int{"id": func(a, b []string) int { return strings.Compare(a[0], b[0]) }}, SelectionMode: TableSelectionRow, FrozenHeader: data[0]%2 == 0, FrozenColumns: int(data[0] % 4), Style: Style{Width: Flex(1), Height: Flex(1)}}
		state.Sort.Set(TableSort{ColumnID: "id", Direction: TableSortDirection(data[0] % 3)})
		RenderToBuffer(table, int(data[0]%35), int(data[len(data)-1]%12))
		require.Equal(t, original, state.GetRows(), "view operations cannot reorder the source")
		cursor := int(data[0]) % len(rows)
		state.SelectIndex(cursor)
		state.Select(cursor)
		desired := rows[cursor][0]
		replacement := slices.Clone(rows)
		slices.Reverse(replacement)
		state.SetRows(replacement)
		got, ok := state.SelectedRow()
		require.True(t, ok)
		require.Equal(t, desired, got[0])
		require.Equal(t, []int{len(rows) - 1 - cursor}, state.SelectedIndices())
	})
}

func TestTableProbeIndependentTablesInDialog(t *testing.T) {
	first, second := featureTable(), featureTable()
	first.ID, second.ID = "first-table", "second-table"
	first.SelectionMode, second.SelectionMode = TableSelectionCursor, TableSelectionCursor
	first.Style, second.Style = Style{Width: Cells(28), Height: Cells(4)}, Style{Width: Cells(28), Height: Cells(4)}
	var opened []string
	first.OnSelect = func(r featureRow) { opened = append(opened, "first:"+r.ID) }
	second.OnSelect = func(r featureRow) { opened = append(opened, "second:"+r.ID) }
	scene := newClickScene(t, Dialog{ID: "table-probe-modal", Visible: true, Style: Style{Width: Cells(36)}, Content: Column{Spacing: 1, Children: []Widget{first, second}}}, 50, 18)
	require.Equal(t, "first-table", scene.focus.FocusedID())
	require.True(t, scene.focus.HandleKey(makeKeyEvent('s', uv.ModCtrl)))
	scene.draw()
	require.Equal(t, TableSortAscending, first.State.Sort.Peek().Direction)
	require.Equal(t, TableSort{}, second.State.Sort.Peek())
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyTab, 0)))
	scene.draw()
	require.Equal(t, "second-table", scene.focus.FocusedID())
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyRight, uv.ModCtrl)))
	scene.draw()
	require.Equal(t, 13, second.State.ColumnWidths.Peek()["name"])
	require.Empty(t, first.State.ColumnWidths.Peek())
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyEnter, 0)))
	require.Equal(t, []string{"second:a"}, opened)
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyTab, 0)))
	scene.draw()
	require.Equal(t, "first-table", scene.focus.FocusedID(), "dialog traps focus across table instances")
	scene.snapshot("TableProbe_dialog_instances", "Independent sorting and width state in two frozen tables inside a modal")
}

func TestTableProbeSortedTableInExternalScrollable(t *testing.T) {
	table := featureTable()
	table.FrozenHeader, table.FrozenColumns = false, 0
	table.ScrollState = NewScrollState()
	table.Style = Style{Padding: EdgeInsetsXY(1, 0)}
	table.State.Sort.Set(TableSort{ColumnID: "name", Direction: TableSortAscending})
	scene := newClickScene(t, Column{Children: []Widget{Text{Content: "outside"}, Scrollable{State: table.ScrollState, Height: Cells(4), Style: Style{Padding: EdgeInsetsAll(1), Border: RoundedBorder(Blue)}, Child: table}}}, 44, 10)
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyEnd, 0)))
	scene.draw()
	require.Equal(t, "a", table.CursorRow().ID, "End follows sorted view, not source order")
	require.Contains(t, tableProbeText(scene.buf), "Zoe")
	require.Positive(t, table.ScrollState.GetOffset())
}

func TestTableProbeTabLeavesTable(t *testing.T) {
	table := Table[[]string]{ID: "before", State: NewTableState([][]string{{"cell"}}), Columns: []TableColumn{{Width: Cells(8)}}, FrozenHeader: true, Style: Style{Height: Cells(2)}}
	scene := newClickScene(t, Column{Children: []Widget{table, Button{ID: "after", Label: "After"}}}, 20, 5)
	require.Equal(t, "before", scene.focus.FocusedID())
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyTab, 0)))
	require.Equal(t, "after", scene.focus.FocusedID(), "one Tab must leave the Table for the next focusable widget")
	scene.draw()
	require.Equal(t, "after", scene.focus.FocusedID(), "rendering must retain the next widget's focus")
}

func TestTableProbeOversizedColumnRevealsLeadingText(t *testing.T) {
	scroll := NewScrollState()
	table := Table[[]string]{ID: "wide", State: NewTableState([][]string{{"id", "hello"}}), ScrollState: scroll,
		Columns: []TableColumn{{Width: Cells(3)}, {Width: Cells(12)}}, FrozenColumns: 1, Style: Style{Width: Flex(1), Height: Flex(1)}}
	scene := newClickScene(t, table, 8, 1)
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyRight, 0)))
	scene.draw()
	require.Equal(t, 1, table.State.CursorColumn.Peek())
	require.Contains(t, tableProbeText(scene.buf), "hello", "an oversized column should reveal its useful leading edge")
	scene.snapshot("TableProbe_oversized_column", "An oversized column reveals its leading content instead of an empty far edge")
	scroll.ScrollRight(2)
	scene.draw()
	require.Equal(t, 2, scroll.GetOffsetX(), "manual panning should remain where the user puts it")
}

func TestTableProbeOversizedRowRevealsLeadingLine(t *testing.T) {
	scroll := NewScrollState()
	table := Table[[]string]{ID: "tall-leading", State: NewTableState([][]string{{"A\nB\nC"}}), ScrollState: scroll,
		Columns: []TableColumn{{Header: Text{Content: "Header"}, Width: Cells(8)}}, FrozenHeader: true, Style: Style{Width: Flex(1), Height: Flex(1)}}
	scene := newClickScene(t, table, 10, 2)
	require.Equal(t, "A", scene.buf.CellAt(0, 1).Content, "initial auto-reveal should show the start of an oversized row")
	scene.snapshot("TableProbe_oversized_row_start", "Initial navigation shows the leading line of a row taller than its available body pane")
	scroll.SetOffset(100)
	scene.draw()
	require.Equal(t, "C", scene.buf.CellAt(0, 1).Content, "manual scroll can still reach the final line")
	require.True(t, scene.focus.HandleKey(makeKeyEvent(uv.KeyHome, 0)))
	scene.draw()
	require.Equal(t, "A", scene.buf.CellAt(0, 1).Content, "explicit navigation reveals the start again")
}
