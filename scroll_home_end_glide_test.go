package terma

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type frozenTableGlideScene struct {
	table *TableState[[]string]
	mode  TableSelectionMode
}

func (s *frozenTableGlideScene) Build(BuildContext) Widget {
	return s.widget()
}

func (s *frozenTableGlideScene) widget() Table[[]string] {
	return Table[[]string]{
		ID:            "table",
		State:         s.table,
		Columns:       []TableColumn{{Header: Text{Content: "Name"}}},
		FrozenHeader:  true,
		SelectionMode: s.mode,
		Height:        Cells(6),
	}
}

// A table with a frozen header scrolls its own viewport. End and home glide
// it, and every frame of the glide matches a full render.
func TestReactivityFrozenTableStartAndEndGlide(t *testing.T) {
	for name, mode := range map[string]TableSelectionMode{"rows": TableSelectionRow, "columns": TableSelectionColumn} {
		t.Run(name, func(t *testing.T) {
			advance := installScrollAnimationClock(t)
			sequence := newReactivitySequence(t, 20, 6, func() *frozenTableGlideScene {
				rows := make([][]string, 60)
				for i := range rows {
					rows[i] = []string{fmt.Sprintf("item %d", i)}
				}
				return &frozenTableGlideScene{table: NewTableState(rows), mode: mode}
			})
			sequence.frame("Initial", nil)
			sequence.focus("table")
			sequence.frame("Focus table", nil)
			actual := sequence.actual.root
			scroll := actual.table.viewport
			maxOffset := scroll.maxOffset()
			require.Positive(t, maxOffset)

			frames := func(label string) []int {
				var offsets []int
				for i := 0; scroll.animation != nil; i++ {
					require.Less(t, i, 100)
					advance(testFrame)
					sequence.frame(fmt.Sprintf("%s frame %d", label, i+1), nil)
					offsets = append(offsets, scroll.GetOffset())
				}
				return offsets
			}

			sequence.frame("End", func(s *frozenTableGlideScene) { s.widget().keyCursorToLast() })
			require.Equal(t, 0, scroll.GetOffset(), "the viewport hasn't moved yet")
			requireGlide(t, frames("End"), 0, maxOffset)
			require.Contains(t, sequence.actual.renderer.ScreenText(), "item 59")

			sequence.frame("Home", func(s *frozenTableGlideScene) { s.widget().keyCursorToFirst() })
			requireGlide(t, frames("Home"), maxOffset, 0)
			require.Contains(t, sequence.actual.renderer.ScreenText(), "item 0")
		})
	}
}

// Moving the cursor during a frozen table's glide retargets it rather than
// cutting it short, as it does for a table in a Scrollable.
func TestFrozenTableCursorMoveDuringGlideRetargets(t *testing.T) {
	rows := make([][]string, 60)
	for i := range rows {
		rows[i] = []string{fmt.Sprintf("item %d", i)}
	}
	scene := &frozenTableGlideScene{table: NewTableState(rows)}
	p := NewPilot(t, scene.widget(), 20, 6)
	scroll := scene.table.viewport
	maxOffset := scroll.maxOffset()

	scene.widget().keyCursorToLast()
	p.Advance(testFrame)
	require.NotNil(t, scroll.animation)
	scene.widget().keyCursorUp()
	p.settle()
	require.NotNil(t, scroll.animation, "the glide keeps going")
	require.Equal(t, maxOffset, scroll.scrollTarget(), "the row above the last is still in view at the end")
}

// The palette's input keeps focus, so its home and end move the list cursor.
// The list glides to it, as a focused List does.
func TestCommandPaletteStartAndEndGlide(t *testing.T) {
	items := make([]CommandPaletteItem, 30)
	for i := range items {
		items[i] = CommandPaletteItem{Label: fmt.Sprintf("Item %d", i+1)}
	}
	state := NewCommandPaletteState("Commands", items)
	state.Open()
	palette := CommandPalette{ID: "palette", State: state}
	p := NewPilot(t, Stack{Children: []Widget{Text{Content: "app"}, palette}}, 60, 20)
	level := state.CurrentLevel()
	scroll := level.ScrollState
	maxOffset := scroll.maxOffset()
	require.Positive(t, maxOffset)
	frames := func() []int {
		var offsets []int
		for i := 0; scroll.animation != nil; i++ {
			require.Less(t, i, 100)
			p.Advance(testFrame)
			offsets = append(offsets, scroll.GetOffset())
		}
		return offsets
	}

	palette.moveCursorToEnd()
	require.Equal(t, len(items)-1, level.ListState.CursorIndex.Peek(), "the cursor moves at once")
	p.settle()
	require.Equal(t, 0, scroll.GetOffset(), "the viewport hasn't moved yet")
	requireGlide(t, frames(), 0, maxOffset)

	palette.moveCursorToStart()
	require.Equal(t, 0, level.ListState.CursorIndex.Peek())
	p.settle()
	requireGlide(t, frames(), maxOffset, 0)
}
