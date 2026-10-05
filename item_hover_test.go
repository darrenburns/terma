package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hover moves the pointer to (x, y) with no button held, then redraws and
// reconciles hover as the app loop does.
func (s *clickScene) hover(x, y int) {
	s.t.Helper()
	s.router.motion(uv.MouseMotionEvent{X: x, Y: y, Button: uv.MouseNone}, 0.5, 0.5)
	s.draw()
	if s.router.reconcileHover() {
		s.draw()
	}
}

func (s *clickScene) bgAt(x, y int) Color {
	s.t.Helper()
	cell := s.buf.CellAt(x, y)
	require.NotNil(s.t, cell)
	return FromANSI(cell.Style.Bg)
}

func hoveredListItem(state *ListState[string]) (int, bool) {
	hovered := state.hover.signal.Peek()
	return hovered.key, hovered.ok
}

// hoverScreen is a full-screen themed background, so the tint has a known
// base to blend over.
type hoverScreen struct{ child Widget }

func (s hoverScreen) Build(ctx BuildContext) Widget {
	return Column{
		Style:    Style{BackgroundColor: ctx.Theme().Background, Width: Flex(1), Height: Flex(1)},
		Children: []Widget{s.child},
	}
}

func TestListHover_FollowsPointer(t *testing.T) {
	state := NewListState(numberedItems(4))
	p := NewPilot(t, hoverScreen{List[string]{ID: "list", State: state, Style: Style{Width: Cells(12)}}}, 20, 6)
	theme := getTheme()
	base := theme.Background
	tinted := theme.Hover.BlendOver(base)

	p.MouseMove(3, 2)
	index, ok := hoveredListItem(state)
	require.True(t, ok)
	assert.Equal(t, 2, index)
	assert.Equal(t, tinted, bgAt(p, 0, 2), "the whole row is tinted, not just its text")
	assert.Equal(t, tinted, bgAt(p, 11, 2))
	assert.Equal(t, base, bgAt(p, 0, 1))
	assert.Equal(t, base, bgAt(p, 12, 2), "the tint stays within the list")

	p.MouseMove(3, 3)
	index, _ = hoveredListItem(state)
	assert.Equal(t, 3, index)
	assert.Equal(t, base, bgAt(p, 0, 2), "the row the pointer left loses its tint")
	assert.Equal(t, tinted, bgAt(p, 0, 3))

	p.MouseMove(3, 5)
	_, ok = hoveredListItem(state)
	assert.False(t, ok, "below the items nothing is hovered")
	assert.Equal(t, base, bgAt(p, 0, 3))
}

func TestListHover_CursorRowKeepsItsColour(t *testing.T) {
	state := NewListState(numberedItems(4))
	p := NewPilot(t, hoverScreen{List[string]{ID: "list", State: state, Style: Style{Width: Cells(12)}}}, 20, 6)
	p.ClickAt(3, 0)
	cursor := bgAt(p, 0, 0)
	require.Equal(t, getTheme().ActiveCursor, cursor)

	p.MouseMove(3, 0)
	assert.Equal(t, cursor, bgAt(p, 0, 0), "the opaque cursor highlight covers the hover underlay")
}

func TestListHover_SelectedRowCompositesOverTint(t *testing.T) {
	state := NewListState(numberedItems(4))
	state.Select(2)
	list := List[string]{ID: "list", State: state, MultiSelect: true, Style: Style{Width: Cells(12)}}
	p := NewPilot(t, hoverScreen{list}, 20, 6)
	theme := getTheme()
	selected := bgAt(p, 0, 2)

	p.MouseMove(3, 2)
	assert.Equal(t, theme.Selection.BlendOver(theme.Hover.BlendOver(theme.Background)), bgAt(p, 0, 2),
		"the translucent selection lies over the hover tint")
	assert.NotEqual(t, selected, bgAt(p, 0, 2))
}

func TestListHover_MovesWhenContentScrollsUnderPointer(t *testing.T) {
	state := NewListState(numberedItems(10))
	scroll := NewScrollState()
	root := hoverScreen{Scrollable{
		State: scroll, Height: Cells(4),
		Child: List[string]{ID: "list", State: state, ScrollState: scroll},
	}}
	p := NewPilot(t, root, 20, 6)

	p.MouseMove(2, 1)
	index, _ := hoveredListItem(state)
	require.Equal(t, 1, index)

	// The wheel scrolls the list beneath a pointer that doesn't move.
	p.Scroll(2, 1, 1)
	index, _ = hoveredListItem(state)
	assert.Equal(t, 2, index)
}

func TestListHover_SkipsItemsThePointerCannotTarget(t *testing.T) {
	state := NewListState([]string{"one", "---", "three"})
	list := List[string]{ID: "list", State: state, pointerTargetable: func(item string) bool { return item != "---" }}
	p := NewPilot(t, list, 20, 4)

	p.MouseMove(1, 0)
	index, ok := hoveredListItem(state)
	require.True(t, ok)
	require.Equal(t, 0, index)

	p.MouseMove(1, 1)
	_, ok = hoveredListItem(state)
	assert.False(t, ok, "a divider isn't highlighted, and the item left loses its highlight")
}

func TestListHover_BlockedByModal(t *testing.T) {
	state := NewListState(numberedItems(4))
	root := hoverModalScene{list: List[string]{ID: "list", State: state}}
	p := NewPilot(t, root, 30, 8)

	p.MouseMove(1, 0)
	_, ok := hoveredListItem(state)
	assert.False(t, ok, "a modal covers the list")
}

type hoverModalScene struct{ list List[string] }

func (s hoverModalScene) Build(BuildContext) Widget {
	return Stack{Children: []Widget{
		s.list,
		Floating{Visible: true, Config: FloatConfig{Modal: true, Position: FloatPositionCenter}, Child: Text{Content: "modal"}},
	}}
}

func TestTableHover_RowModeKeepsHoverAcrossCells(t *testing.T) {
	rows := [][]string{{"a0", "b0"}, {"a1", "b1"}, {"a2", "b2"}}
	state := NewTableState(rows)
	table := Table[[]string]{
		ID: "table", State: state, SelectionMode: TableSelectionRow,
		Columns: []TableColumn{{Width: Cells(4)}, {Width: Cells(4)}},
	}
	p := NewPilot(t, hoverScreen{table}, 12, 4)
	tinted := getTheme().Hover.BlendOver(getTheme().Background)

	p.MouseMove(1, 1)
	assert.Equal(t, tableHoverCell{row: 1, col: -1}, state.hover.signal.Peek().key)
	assert.Equal(t, tinted, bgAt(p, 1, 1))
	assert.Equal(t, tinted, bgAt(p, 5, 1), "every cell of the hovered row is tinted")

	assert.False(t, p.session.mouse.itemHover.update(p.session.renderer.hoverItemAt(5, 1)),
		"moving to another cell of the same row doesn't move the hover")
}

func TestTreeHover_FollowsPointer(t *testing.T) {
	state := NewTreeState([]TreeNode[string]{
		{Data: "alpha", Children: []TreeNode[string]{{Data: "a1"}}},
		{Data: "beta"},
	})
	p := NewPilot(t, hoverScreen{Tree[string]{ID: "tree", State: state, Style: Style{Width: Cells(12)}}}, 16, 4)
	tinted := getTheme().Hover.BlendOver(getTheme().Background)

	p.MouseMove(6, 1)
	assert.Equal(t, "0/0", state.hover.signal.Peek().key)
	assert.Equal(t, tinted, bgAt(p, 0, 1), "the prefix is tinted along with the node")
	assert.Equal(t, tinted, bgAt(p, 11, 1))
}

func TestTabBarHover_TintsInactiveTab(t *testing.T) {
	state := NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}})
	p := NewPilot(t, TabBar{ID: "tabs", State: state}, 30, 1)
	theme := getTheme()
	inactive := bgAt(p, 8, 0)
	require.Equal(t, theme.Surface, inactive)
	active := bgAt(p, 1, 0)

	p.MouseMove(8, 0)
	assert.Equal(t, theme.Hover.BlendOver(theme.Surface), bgAt(p, 8, 0))

	p.MouseMove(1, 0)
	assert.Equal(t, inactive, bgAt(p, 8, 0))
	assert.Equal(t, active, bgAt(p, 1, 0), "the active tab doesn't change")

	// Clicks still reach the tab's label.
	p.ClickAt(8, 0)
	assert.Equal(t, "two", state.ActiveKeyPeek())
}

// A custom item with parts of its own: the badge keeps its colour and only
// the transparent parts of the row show the tint.
func TestItemHover_Snapshot(t *testing.T) {
	listState := NewListState([]hoverTask{
		{title: "Write docs", tag: "docs"},
		{title: "Fix hover", tag: "bug"},
		{title: "Ship it", tag: "release"},
	})
	tableState := NewTableState([][]string{{"alpha", "1"}, {"beta", "2"}, {"gamma", "3"}})
	treeState := NewTreeState([]TreeNode[string]{
		{Data: "src", Children: []TreeNode[string]{{Data: "main.go"}, {Data: "util.go"}}},
	})
	tabState := NewTabState([]Tab{{Key: "a", Label: "Files"}, {Key: "b", Label: "Search"}, {Key: "c", Label: "Git"}})

	root := snapshotHoverScene{list: listState, table: tableState, tree: treeState, tabs: tabState}
	p := NewPilot(t, root, 40, 14)
	// The pointer rests on the Save button. Only one thing can be under the
	// pointer, so the collections' hovers are set directly to show every
	// kind of highlight in a single picture.
	p.MouseMove(2, 7)
	listState.hover.set(1, true)
	tableState.hover.set(tableHoverCell{row: 1, col: -1}, true)
	treeState.hover.set("0/0", true)
	tabState.hover.set("b", true)
	p.AssertSnapshot("highlights",
		"Hover highlights: the second list row is lightly tinted behind its text while its red 'bug' badge keeps its colour; "+
			"the table's 'beta' row, the tree's 'main.go' row (including its guide prefix), the inactive 'Search' tab "+
			"and the 'Save' button (but not 'Cancel') are tinted too")
}

type hoverTask struct{ title, tag string }

type snapshotHoverScene struct {
	list  *ListState[hoverTask]
	table *TableState[[]string]
	tree  *TreeState[string]
	tabs  *TabState
}

func (s snapshotHoverScene) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	return Column{
		Style: Style{BackgroundColor: theme.Background, Width: Flex(1), Height: Flex(1)},
		Children: []Widget{
			TabBar{ID: "tabs", State: s.tabs},
			Row{Spacing: 2, Children: []Widget{
				List[hoverTask]{
					ID: "list", State: s.list, Style: Style{Width: Cells(20)},
					RenderItem: func(item hoverTask, active, selected bool) Widget {
						return Row{Style: Style{Width: Flex(1)}, Spacing: 1, Children: []Widget{
							Text{Content: item.title, Style: Style{Width: Flex(1)}},
							Text{Content: item.tag, Style: Style{BackgroundColor: theme.Error, ForegroundColor: theme.TextOnError}},
						}}
					},
				},
				Table[[]string]{
					ID: "table", State: s.table, SelectionMode: TableSelectionRow, ColumnSpacing: 1,
					Columns: []TableColumn{{Width: Cells(8)}, {Width: Cells(3)}},
				},
			}},
			Tree[string]{ID: "tree", State: s.tree, Style: Style{Width: Cells(20)}},
			Row{Spacing: 1, Children: []Widget{
				Button{ID: "save", Label: "Save"},
				Button{ID: "cancel", Label: "Cancel"},
			}},
		},
	}
}

// hoverButtons is a pair of buttons on a themed background, the second
// disabled when disableSecond is set.
type hoverButtons struct{ disableSecond bool }

func (s hoverButtons) Build(ctx BuildContext) Widget {
	return Column{
		Style: Style{BackgroundColor: ctx.Theme().Background, Width: Flex(1), Height: Flex(1)},
		Children: []Widget{Row{Spacing: 1, Children: []Widget{
			Button{ID: "one", Label: "One"},
			DisabledWhen(s.disableSecond, Button{ID: "two", Label: "Two", Variant: ButtonPrimary}),
		}}},
	}
}

func TestButtonHover_TintsBackground(t *testing.T) {
	p := NewPilot(t, hoverButtons{}, 20, 2)
	theme := getTheme()
	require.Equal(t, theme.Surface, bgAt(p, 1, 0))
	require.Equal(t, theme.Primary, bgAt(p, 7, 0))

	p.MouseMove(1, 0)
	assert.Equal(t, 1, p.session.renderer.Stats().BuildCount, "only the hovered button rebuilds")
	assert.Equal(t, theme.Hover.BlendOver(theme.Surface), bgAt(p, 0, 0), "the whole button is tinted")
	assert.Equal(t, theme.Hover.BlendOver(theme.Surface), bgAt(p, 4, 0))
	assert.Equal(t, theme.Primary, bgAt(p, 7, 0), "the other button is unchanged")

	p.MouseMove(7, 0)
	assert.Equal(t, theme.Surface, bgAt(p, 1, 0))
	assert.Equal(t, theme.Hover.BlendOver(theme.Primary), bgAt(p, 7, 0), "variant buttons are tinted too")

	p.MouseMove(15, 1)
	assert.Equal(t, theme.Primary, bgAt(p, 7, 0))
}

func TestButtonHover_FocusedButtonIsTinted(t *testing.T) {
	p := NewPilot(t, hoverButtons{}, 20, 2)
	theme := getTheme()
	p.session.focus.FocusByID("two")
	p.settle()

	p.MouseMove(7, 0)
	assert.Equal(t, theme.Hover.BlendOver(theme.Primary), bgAt(p, 7, 0))
}

func TestButtonHover_DisabledButtonIsNotTinted(t *testing.T) {
	p := NewPilot(t, hoverButtons{disableSecond: true}, 20, 2)
	before := bgAt(p, 7, 0)

	p.MouseMove(7, 0)
	assert.Equal(t, before, bgAt(p, 7, 0))
}

func TestHoverTint_Background(t *testing.T) {
	hovered := true
	tint := hoverTint{hovered: func() bool { return hovered }, color: RGBA(255, 255, 255, 0.1)}
	screen := RGB(20, 20, 40)

	// Over a translucent background the result stays translucent, and lands
	// on the screen as the tint over that background over the screen.
	base := RGBA(200, 0, 0, 0.4)
	got := tint.background(base).(Color)
	assert.Less(t, got.Alpha(), 1.0)
	want := tint.color.BlendOver(base.BlendOver(screen))
	assertColorClose(t, want, got.BlendOver(screen))

	// Gradients are tinted wherever they are sampled.
	gradient := NewGradient(RGB(0, 0, 0), RGB(255, 255, 255))
	tinted := tint.background(gradient)
	for _, x := range []int{0, 5, 9} {
		assert.Equal(t, tint.color.BlendOver(gradient.ColorAt(10, 1, x, 0)), tinted.ColorAt(10, 1, x, 0))
	}

	hovered = false
	assert.Equal(t, ColorProvider(base), tint.background(base), "no tint when not hovered")
	assert.Nil(t, tint.background(nil))
}

func assertColorClose(t *testing.T, want, got Color) {
	t.Helper()
	wr, wg, wb := want.RGB()
	gr, gg, gb := got.RGB()
	for i, pair := range [][2]uint8{{wr, gr}, {wg, gg}, {wb, gb}} {
		assert.InDelta(t, pair[0], pair[1], 1, "channel %d of %v vs %v", i, want, got)
	}
}

func TestReactivityListHover(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 20, 6, func() *reactivityListScene {
				items := []string{"one", "two", "three", "four", "five"}
				return &reactivityListScene{state: NewListState(items), scroll: NewScrollState(), custom: custom, multi: true, header: NewSignal("list")}
			})
			sequence.frame("Initial", nil)
			sequence.focus("list")
			sequence.frame("Focus list", nil)
			move := func(from, to int) func(*reactivityListScene) {
				return func(s *reactivityListScene) {
					if to >= 0 {
						s.state.hover.set(to, true)
					}
					if from >= 0 {
						s.state.hover.set(from, false)
					}
				}
			}
			enter := sequence.frame("Hover row", move(-1, 1))
			require.LessOrEqual(t, damagedRows(enter), 1, "hovering repaints one row: %v", enter.DamagedRects)
			require.Zero(t, enter.BuildCount, "hover is paint-only")
			next := sequence.frame("Hover next row", move(1, 2))
			require.LessOrEqual(t, damagedRows(next), 2, "moving repaints the rows left and entered")
			require.Zero(t, next.BuildCount)
			sequence.frame("Select hovered row", func(s *reactivityListScene) { s.state.Select(2) })
			sequence.frame("Hover cursor row", move(2, 0))
			sequence.frame("Cursor onto hovered row", func(s *reactivityListScene) { s.list().keyCursorDown() })
			sequence.frame("Leave", move(0, -1))
			// The last item starts out of view, so it only subscribes to hover
			// once scrolling brings it into view.
			sequence.frame("Scroll to last", func(s *reactivityListScene) { s.list().keyCursorToLast() })
			sequence.frame("Hover row scrolled into view", move(-1, 3))
			sequence.frame("Hover last row", move(3, 4))
		})
	}
}

func TestReactivityTreeHover(t *testing.T) {
	for _, custom := range []bool{false, true} {
		t.Run(fmt.Sprintf("custom=%v", custom), func(t *testing.T) {
			sequence := newReactivitySequence(t, 24, 9, func() *reactivityTreeScene {
				roots := []TreeNode[string]{
					{Data: "alpha", Children: []TreeNode[string]{{Data: "a1"}, {Data: "a2"}}},
					{Data: "beta"},
				}
				return &reactivityTreeScene{state: NewTreeState(roots), custom: custom}
			})
			sequence.frame("Initial", nil)
			sequence.focus("tree")
			sequence.frame("Focus tree", nil)
			enter := sequence.frame("Hover row", func(s *reactivityTreeScene) { s.state.hover.set("0/1", true) })
			require.LessOrEqual(t, damagedRows(enter), 1)
			require.Zero(t, enter.BuildCount, "hover is paint-only")
			next := sequence.frame("Hover next row", func(s *reactivityTreeScene) {
				s.state.hover.set("1", true)
				s.state.hover.set("0/1", false)
			})
			require.LessOrEqual(t, damagedRows(next), 2)
			require.Zero(t, next.BuildCount)
			sequence.frame("Collapse under hover", func(s *reactivityTreeScene) { s.state.Collapse([]int{0}) })
			sequence.frame("Leave", func(s *reactivityTreeScene) { s.state.hover.set("1", false) })
		})
	}
}

func TestReactivityTableHover(t *testing.T) {
	modes := map[string]TableSelectionMode{"cursor": TableSelectionCursor, "row": TableSelectionRow, "column": TableSelectionColumn}
	for name, mode := range modes {
		for _, custom := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/custom=%v", name, custom), func(t *testing.T) {
				sequence := newReactivitySequence(t, 24, 9, func() *reactivityTableScene {
					rows := make([][]string, 5)
					for i := range rows {
						rows[i] = []string{fmt.Sprintf("a%d", i), fmt.Sprintf("b%d", i), fmt.Sprintf("c%d", i)}
					}
					return &reactivityTableScene{state: NewTableState(rows), mode: mode, custom: custom, header: NewSignal("table")}
				})
				sequence.frame("Initial", nil)
				sequence.focus("table")
				sequence.frame("Focus table", nil)
				hover := func(row, col int, hovered bool) func(*reactivityTableScene) {
					return func(s *reactivityTableScene) { s.table().setCellHovered(row, col, hovered) }
				}
				enter := sequence.frame("Hover cell", hover(2, 1, true))
				require.Zero(t, enter.BuildCount, "hover is paint-only")
				if mode != TableSelectionColumn {
					require.LessOrEqual(t, damagedRows(enter), 1)
				}
				sequence.frame("Hover another cell", func(s *reactivityTableScene) {
					s.table().setCellHovered(3, 2, true)
					s.table().setCellHovered(2, 1, false)
				})
				sequence.frame("Leave", hover(3, 2, false))
			})
		}
	}
}
