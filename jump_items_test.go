package terma

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// jumpItemsScene holds data-driven content: tabs, a scrolled list, a tree and
// a table. Only the list has a static key; everything else in view is
// reached through dynamic hints on its items.
type jumpItemsScene struct {
	jump   *JumpState
	tabs   *TabState
	scroll *ScrollState
	list   *ListState[string]
	tree   *TreeState[string]
	table  *TableState[string]
}

func newJumpItemsScene() *jumpItemsScene {
	items := make([]string, 10)
	for i := range items {
		items[i] = fmt.Sprintf("item %d", i)
	}
	return &jumpItemsScene{
		jump:   NewJumpState(),
		tabs:   NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "three", Label: "Three"}}),
		scroll: NewScrollState(),
		list:   NewListState(items),
		tree:   NewTreeState([]TreeNode[string]{{Data: "alpha"}, {Data: "beta"}, {Data: "gamma"}}),
		table:  NewTableState([]string{"r0", "r1", "r2"}),
	}
}

func (s *jumpItemsScene) Build(ctx BuildContext) Widget {
	return Jumper{
		State:   s.jump,
		Targets: []JumpTarget{{Key: "1", ID: "list"}},
		Dynamic: true,
		Child: Column{Children: []Widget{
			TabBar{ID: "tabs", State: s.tabs},
			Row{Spacing: 2, Children: []Widget{
				Scrollable{State: s.scroll, Height: Cells(4), Width: Cells(12), Child: List[string]{ID: "list", State: s.list, ScrollState: s.scroll}},
				Tree[string]{ID: "tree", State: s.tree, Width: Cells(12)},
				Table[string]{
					ID:            "table",
					State:         s.table,
					SelectionMode: TableSelectionRow,
					Columns:       []TableColumn{{Width: Cells(5)}, {Width: Cells(5)}},
					RenderCell: func(row string, r, c int, active, selected bool) Widget {
						return Text{Content: fmt.Sprintf("%s.%d", row, c)}
					},
				},
			}},
		}},
	}
}

// jumpItemKeys returns the keys of the item hints a widget holds, in reading
// order, leaving out static keys.
func jumpItemKeys(state *JumpState, holder string) []string {
	var labels []jumpLabel
	for _, label := range state.labels {
		if label.focusID == holder && label.key != "1" {
			labels = append(labels, label)
		}
	}
	slices.SortStableFunc(labels, func(a, b jumpLabel) int {
		if a.at.Y != b.at.Y {
			return a.at.Y - b.at.Y
		}
		return a.at.X - b.at.X
	})
	keys := make([]string, len(labels))
	for i, label := range labels {
		keys[i] = label.key
	}
	return keys
}

// typeKey types a hint one character at a time.
func (s *reactivitySequence[T]) typeKey(name, key string) {
	s.t.Helper()
	var keys []string
	for _, r := range key {
		keys = append(keys, string(r))
	}
	s.press(name, keys...)
}

func TestJumpDynamicHintsLabelItems(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 8, newJumpItemsScene)
	scene := sequence.actual.root
	sequence.frame("Initial", nil)
	sequence.press("Jump mode", "ctrl+o")

	// One hint per item in view, not per widget: four of the ten list rows
	// fit, and the table labels rows (their first cells), not every cell.
	require.Len(t, jumpItemKeys(scene.jump, "list"), 4)
	require.Len(t, jumpItemKeys(scene.jump, "tabs"), 3)
	require.Len(t, jumpItemKeys(scene.jump, "tree"), 3)
	require.Len(t, jumpItemKeys(scene.jump, "table"), 3)
	// The list's static key still reaches the list itself.
	require.True(t, slices.ContainsFunc(scene.jump.labels, func(label jumpLabel) bool {
		return label.key == "1" && label.focusID == "list" && label.action == nil
	}))

	sequence.typeKey("Jump to the third list row", jumpItemKeys(scene.jump, "list")[2])
	require.Equal(t, 2, scene.list.CursorIndex.Peek())
	require.Equal(t, "list", sequence.actual.focus.FocusedID())

	sequence.press("Jump mode", "ctrl+o")
	sequence.typeKey("Jump to the second tab", jumpItemKeys(scene.jump, "tabs")[1])
	require.Equal(t, "two", scene.tabs.ActiveKeyPeek())

	sequence.press("Jump mode", "ctrl+o")
	sequence.typeKey("Jump to the last tree node", jumpItemKeys(scene.jump, "tree")[2])
	require.Equal(t, []int{2}, scene.tree.CursorPath.Peek())
	require.Equal(t, "tree", sequence.actual.focus.FocusedID())

	sequence.press("Jump mode", "ctrl+o")
	sequence.typeKey("Jump to the second table row", jumpItemKeys(scene.jump, "table")[1])
	require.Equal(t, 1, scene.table.CursorIndex.Peek())
	require.Equal(t, "table", sequence.actual.focus.FocusedID())
}

func TestJumpDynamicHintsFollowScrolling(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 8, newJumpItemsScene)
	scene := sequence.actual.root
	sequence.frame("Initial", nil)
	sequence.frame("Wheel-scroll the list", func(s *jumpItemsScene) { s.scroll.ScrollDown(5) })
	sequence.press("Jump mode", "ctrl+o")

	// Only the rows scrolled into view are labelled.
	require.Len(t, jumpItemKeys(scene.jump, "list"), 4)
	sequence.typeKey("Jump to the first row in view", jumpItemKeys(scene.jump, "list")[0])
	require.Equal(t, 5, scene.list.CursorIndex.Peek())
}

func TestJumpItemsSnapshot(t *testing.T) {
	scene := newJumpItemsScene()
	scene.jump.Activate()
	AssertSnapshot(t, scene, 60, 8,
		"Dynamic hints on data: a letter on each tab (One, Two, Three), on each of the four list rows in view, on each tree node (alpha, beta, gamma) and on each table row. The list also shows its static key '1' at its top-left, next to its first row's hint.")
}
