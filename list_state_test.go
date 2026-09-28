package terma

import (
	"reflect"
	"testing"
)

func newSelectedListState(items []string, selected ...int) *ListState[string] {
	state := NewListState(items)
	for _, idx := range selected {
		state.Select(idx)
	}
	return state
}

func assertSelectedItems(t *testing.T, state *ListState[string], want ...string) {
	t.Helper()
	got := state.SelectedItems()
	if len(want) == 0 {
		want = []string{}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selected items = %v, want %v", got, want)
	}
}

func TestListState_InsertAtKeepsSelectionOnSameItems(t *testing.T) {
	state := newSelectedListState([]string{"a", "b", "c"}, 0, 2)
	state.SetAnchor(2)

	state.InsertAt(1, "new")

	assertSelectedItems(t, state, "a", "c")
	if got := state.GetAnchor(); got != 3 {
		t.Fatalf("anchor = %d, want 3", got)
	}
}

func TestListState_PrependKeepsCursorAndSelectionOnSameItems(t *testing.T) {
	state := newSelectedListState([]string{"a", "b"}, 1)
	state.SelectIndex(1)

	state.Prepend("new")

	assertSelectedItems(t, state, "b")
	if item, _ := state.SelectedItem(); item != "b" {
		t.Fatalf("cursor item = %q, want %q", item, "b")
	}
}

func TestListState_InsertIntoEmptyListKeepsCursorInRange(t *testing.T) {
	for name, insert := range map[string]func(*ListState[string]){
		"prepend":  func(s *ListState[string]) { s.Prepend("only") },
		"insertAt": func(s *ListState[string]) { s.InsertAt(0, "only") },
	} {
		t.Run(name, func(t *testing.T) {
			state := NewListState([]string{})
			insert(state)
			if got := state.CursorIndex.Peek(); got != 0 {
				t.Fatalf("cursor = %d, want 0", got)
			}
		})
	}
}

func TestListState_RemoveAtDropsRemovedItemFromSelection(t *testing.T) {
	state := newSelectedListState([]string{"a", "b", "c", "d"}, 0, 1, 3)
	state.SetAnchor(1)

	state.RemoveAt(1)

	assertSelectedItems(t, state, "a", "d")
	if state.HasAnchor() {
		t.Fatalf("anchor = %d, want none after its item was removed", state.GetAnchor())
	}
}

func TestListState_RemoveWhereKeepsSelectionOnRemainingItems(t *testing.T) {
	state := newSelectedListState([]string{"a", "x", "b", "x", "c"}, 1, 4)

	removed := state.RemoveWhere(func(item string) bool { return item == "x" })

	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}
	assertSelectedItems(t, state, "c")
}

func TestListState_ClearResetsSelection(t *testing.T) {
	state := newSelectedListState([]string{"a", "b"}, 0, 1)
	state.SetAnchor(0)

	state.Clear()

	if got := len(state.Selection.Peek()); got != 0 {
		t.Fatalf("selection size = %d, want 0", got)
	}
	if state.HasAnchor() {
		t.Fatal("anchor should be cleared")
	}
}

func TestListState_SetItemsDropsSelectionPastEnd(t *testing.T) {
	state := newSelectedListState([]string{"a", "b", "c"}, 0, 2)

	state.SetItems([]string{"x", "y"})

	if got := state.SelectedIndices(); !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("selected indices = %v, want [0]", got)
	}
}
