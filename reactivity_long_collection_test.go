package terma

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// These scenes have enough rows that painting finds the visible ones by
// binary search (minStackedChildren), rows scrolled out of view are skipped,
// and rows that rebuild at the same size revalidate their parent's layout
// instead of laying it out again. Every frame is compared with a forced full
// render.

const longCollectionRows = 100

func longCollectionItems(prefix string) []string {
	items := make([]string, longCollectionRows)
	for i := range items {
		items[i] = fmt.Sprintf("%s %03d", prefix, i)
	}
	return items
}

type reactivityLongListScene struct {
	state  *ListState[string]
	scroll *ScrollState
	custom bool
	// tallActive renders the cursor row two lines tall, so moving the cursor
	// changes row sizes and the parent's layout can't be reused.
	tallActive bool
}

func (s *reactivityLongListScene) list() List[string] {
	list := List[string]{ID: "list", State: s.state, ScrollState: s.scroll}
	if s.custom {
		list.RenderItem = func(item string, active, selected bool) Widget {
			if active {
				content := "> " + item
				if s.tallActive {
					content += "\n  (details)"
				}
				return Text{Content: content, Style: Style{Width: Flex(1)}}
			}
			return Text{Content: "  " + item, Style: Style{Width: Flex(1)}}
		}
	}
	return list
}

func (s *reactivityLongListScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Text{Content: "header"},
		Scrollable{ID: "scroll", State: s.scroll, Height: Cells(6), Child: s.list()},
		Text{Content: "footer"},
	}}
}

func TestReactivityLongListScrolling(t *testing.T) {
	for _, variant := range []struct {
		name               string
		custom, tallActive bool
	}{{"default", false, false}, {"custom", true, false}, {"custom-tall-active", true, true}} {
		t.Run(variant.name, func(t *testing.T) {
			sequence := newReactivitySequence(t, 24, 9, func() *reactivityLongListScene {
				return &reactivityLongListScene{
					state: NewListState(longCollectionItems("item")), scroll: NewScrollState(),
					custom: variant.custom, tallActive: variant.tallActive,
				}
			})
			sequence.frame("Initial", nil)
			sequence.focus("list")
			sequence.frame("Focus list", nil)
			for i := 0; i < 10; i++ {
				work := sequence.frame(fmt.Sprintf("Cursor down %d", i+1), func(s *reactivityLongListScene) { s.list().keyCursorDown() })
				if variant.custom && !variant.tallActive && i > 6 {
					require.Less(t, work.LayoutCount, 40, "rebuilt rows of unchanged size don't lay out the whole list")
				}
			}
			for i := 0; i < 3; i++ {
				sequence.frame(fmt.Sprintf("Cursor up %d", i+1), func(s *reactivityLongListScene) { s.list().keyCursorUp() })
			}
			sequence.frame("Wheel down", func(s *reactivityLongListScene) { s.scroll.ScrollDown(3) })
			sequence.frame("Jump far down", func(s *reactivityLongListScene) { s.scroll.SetOffset(80) })
			sequence.frame("Jump back up", func(s *reactivityLongListScene) { s.scroll.SetOffset(2) })
			sequence.frame("Page down", func(s *reactivityLongListScene) { s.list().pageDown() })
			sequence.frame("Cursor to last", func(s *reactivityLongListScene) { s.list().keyCursorToLast() })
			sequence.frame("Cursor to first", func(s *reactivityLongListScene) { s.list().keyCursorToFirst() })
			sequence.frame("Programmatic cursor", func(s *reactivityLongListScene) { s.state.SelectIndex(55) })
			sequence.frame("Remove a visible row", func(s *reactivityLongListScene) { s.state.RemoveAt(53) })
			sequence.frame("Insert a visible row", func(s *reactivityLongListScene) { s.state.InsertAt(52, "inserted") })
			sequence.frame("Shorter items", func(s *reactivityLongListScene) { s.state.SetItems(longCollectionItems("short")[:40]) })
			sequence.frame("Cursor down after replace", func(s *reactivityLongListScene) { s.list().keyCursorDown() })
		})
	}
}

type reactivityLongTableScene struct {
	state  *TableState[[]string]
	scroll *ScrollState
	custom bool
}

func (s *reactivityLongTableScene) table() Table[[]string] {
	cols := []TableColumn{{Width: Cells(8)}, {Width: Cells(6)}, {Width: Cells(6)}}
	table := Table[[]string]{ID: "table", State: s.state, ScrollState: s.scroll, Columns: cols}
	if s.custom {
		table.RenderCell = func(row []string, rowIndex, colIndex int, active, selected bool) Widget {
			marker := " "
			if active {
				marker = ">"
			}
			return Text{Content: marker + row[colIndex]}
		}
	}
	return table
}

func (s *reactivityLongTableScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Text{Content: "header"},
		Scrollable{ID: "scroll", State: s.scroll, Height: Cells(6), Child: s.table()},
	}}
}

func TestReactivityLongTableScrolling(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 26, 8, func() *reactivityLongTableScene {
				rows := make([][]string, longCollectionRows)
				for i := range rows {
					rows[i] = []string{fmt.Sprintf("row %03d", i), "mid", "end"}
				}
				return &reactivityLongTableScene{state: NewTableState(rows), scroll: NewScrollState(), custom: custom}
			})
			sequence.frame("Initial", nil)
			sequence.focus("table")
			sequence.frame("Focus table", nil)
			for i := 0; i < 9; i++ {
				sequence.frame(fmt.Sprintf("Cursor down %d", i+1), func(s *reactivityLongTableScene) { s.table().keyCursorDown() })
			}
			sequence.frame("Cursor up", func(s *reactivityLongTableScene) { s.table().keyCursorUp() })
			sequence.frame("Wheel down", func(s *reactivityLongTableScene) { s.scroll.ScrollDown(4) })
			sequence.frame("Jump far down", func(s *reactivityLongTableScene) { s.scroll.SetOffset(70) })
			sequence.frame("Cursor to last", func(s *reactivityLongTableScene) { s.table().keyCursorToLast() })
			sequence.frame("Cursor to first", func(s *reactivityLongTableScene) { s.table().keyCursorToFirst() })
		})
	}
}

type reactivityLongTreeScene struct {
	state  *TreeState[string]
	scroll *ScrollState
}

func (s *reactivityLongTreeScene) tree() Tree[string] {
	return Tree[string]{ID: "tree", State: s.state, ScrollState: s.scroll}
}

func (s *reactivityLongTreeScene) Build(BuildContext) Widget {
	return Scrollable{ID: "scroll", State: s.scroll, Height: Cells(6), Child: s.tree()}
}

func TestReactivityLongTreeScrolling(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 8, func() *reactivityLongTreeScene {
		roots := make([]TreeNode[string], longCollectionRows)
		for i := range roots {
			roots[i] = TreeNode[string]{Data: fmt.Sprintf("node %03d", i)}
		}
		roots[8].Children = []TreeNode[string]{{Data: "child a"}, {Data: "child b"}}
		return &reactivityLongTreeScene{state: NewTreeState(roots), scroll: NewScrollState()}
	})
	sequence.frame("Initial", nil)
	sequence.focus("tree")
	sequence.frame("Focus tree", nil)
	for i := 0; i < 8; i++ {
		sequence.frame(fmt.Sprintf("Cursor down %d", i+1), func(s *reactivityLongTreeScene) { s.tree().keyCursorDown() })
	}
	sequence.frame("Expand", func(s *reactivityLongTreeScene) { s.tree().expandOrMoveToChild() })
	sequence.frame("Into child", func(s *reactivityLongTreeScene) { s.tree().expandOrMoveToChild() })
	sequence.frame("Back to parent", func(s *reactivityLongTreeScene) { s.tree().collapseOrMoveToParent() })
	sequence.frame("Collapse", func(s *reactivityLongTreeScene) { s.tree().collapseOrMoveToParent() })
	sequence.frame("Jump far down", func(s *reactivityLongTreeScene) { s.scroll.SetOffset(60) })
	sequence.frame("Cursor to last", func(s *reactivityLongTreeScene) { s.tree().keyCursorToLast() })
}

// A long column of rows in which one row reads a signal while painting, one
// row can turn into a Stack whose overlay overflows its box, and rows can
// change size.
type reactivityLongColumnScene struct {
	scroll   *ScrollState
	painted  Signal[string]
	stackRow Signal[bool]
	tallRow  Signal[bool]
}

func (s *reactivityLongColumnScene) Build(BuildContext) Widget {
	rows := make([]Widget, longCollectionRows)
	for i := range rows {
		rows[i] = Text{Content: fmt.Sprintf("row %03d", i)}
	}
	rows[3] = reactivityBench(s.painted)
	rows[20] = reactivityBuilder{ID: "stack-row", build: func(BuildContext) Widget {
		if !s.stackRow.Get() {
			return Text{Content: "plain row 20"}
		}
		return Stack{Children: []Widget{
			Text{Content: "stack row 20"},
			Positioned{Top: IntPtr(-2), Left: IntPtr(10), Child: Text{Content: "OVERFLOW"}},
		}}
	}}
	rows[40] = reactivityBuilder{ID: "tall-row", build: func(BuildContext) Widget {
		if s.tallRow.Get() {
			return Text{Content: "tall row 40\nsecond line"}
		}
		return Text{Content: "row 040"}
	}}
	return Column{Children: []Widget{
		Text{Content: "header"},
		Scrollable{ID: "scroll", State: s.scroll, Height: Cells(6), Child: Column{Children: rows}},
		Text{Content: "footer"},
	}}
}

func TestReactivityLongColumnOffscreenChanges(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 9, func() *reactivityLongColumnScene {
		return &reactivityLongColumnScene{
			scroll: NewScrollState(), painted: NewSignal("row 003"),
			stackRow: NewSignal(false), tallRow: NewSignal(false),
		}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Scroll row 3 out of view", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(10) })
	stats := sequence.frame("Change row 3 out of view", func(s *reactivityLongColumnScene) { s.painted.Set("row XYZ") })
	require.Empty(t, stats.DamagedRects, "a change out of view repaints nothing")
	sequence.frame("Scroll row 3 back", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(0) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "row XYZ")

	sequence.frame("Row 20 becomes a Stack out of view", func(s *reactivityLongColumnScene) { s.stackRow.Set(true) })
	// Row 20 sits just below the viewport; its overlay reaches up into it.
	sequence.frame("Scroll so row 20 is just below the viewport", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(14) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "OV", "a Stack out of view still draws what overflows into view")
	sequence.frame("Scroll row 20 into view", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(17) })
	sequence.frame("Row 20 becomes plain", func(s *reactivityLongColumnScene) { s.stackRow.Set(false) })

	sequence.frame("Row 40 grows out of view", func(s *reactivityLongColumnScene) { s.tallRow.Set(true) })
	sequence.frame("Scroll to row 40", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(37) })
	sequence.frame("Row 40 shrinks in view", func(s *reactivityLongColumnScene) { s.tallRow.Set(false) })
	sequence.frame("Row 40 grows in view", func(s *reactivityLongColumnScene) { s.tallRow.Set(true) })
	sequence.frame("Scroll to the end", func(s *reactivityLongColumnScene) { s.scroll.SetOffset(1000) })
}
