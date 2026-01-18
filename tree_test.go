package terma

import "testing"

func sampleTreeNodes() []TreeNode[string] {
	return []TreeNode[string]{
		{
			Data: "A",
			Children: []TreeNode[string]{
				{
					Data: "A1",
					Children: []TreeNode[string]{
						{Data: "A1a", Children: []TreeNode[string]{}},
					},
				},
				{Data: "A2", Children: []TreeNode[string]{}},
			},
		},
		{Data: "B", Children: []TreeNode[string]{}},
	}
}

func TestTreeStateCursorNavigation(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0}) {
		t.Fatalf("expected initial cursor [0], got %v", got)
	}

	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 0}) {
		t.Fatalf("expected cursor [0 0], got %v", got)
	}

	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 0, 0}) {
		t.Fatalf("expected cursor [0 0 0], got %v", got)
	}

	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 1}) {
		t.Fatalf("expected cursor [0 1], got %v", got)
	}

	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{1}) {
		t.Fatalf("expected cursor [1], got %v", got)
	}

	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{1}) {
		t.Fatalf("expected cursor to stay [1], got %v", got)
	}

	state.CursorUp()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 1}) {
		t.Fatalf("expected cursor [0 1], got %v", got)
	}

	state.CursorToParent()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0}) {
		t.Fatalf("expected cursor [0], got %v", got)
	}

	state.CursorToFirstChild()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 0}) {
		t.Fatalf("expected cursor [0 0], got %v", got)
	}
}

func TestTreeStateCollapseAndExpand(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.Collapse([]int{0})
	if !state.IsCollapsed([]int{0}) {
		t.Fatalf("expected node [0] to be collapsed")
	}

	state.CursorPath.Set([]int{0})
	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{1}) {
		t.Fatalf("expected cursor to skip collapsed children, got %v", got)
	}

	state.Expand([]int{0})
	if state.IsCollapsed([]int{0}) {
		t.Fatalf("expected node [0] to be expanded")
	}

	state.CursorPath.Set([]int{0})
	state.CursorDown()
	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 0}) {
		t.Fatalf("expected cursor to move into expanded children, got %v", got)
	}
}

func TestTreeStateSelection(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.ToggleSelection([]int{0})
	if !state.IsSelected([]int{0}) {
		t.Fatalf("expected node [0] to be selected")
	}

	state.Select([]int{1})
	if !state.IsSelected([]int{1}) {
		t.Fatalf("expected node [1] to be selected")
	}

	state.Deselect([]int{0})
	if state.IsSelected([]int{0}) {
		t.Fatalf("expected node [0] to be deselected")
	}

	selected := state.SelectedPaths()
	if len(selected) != 1 || !pathsEqual(selected[0], []int{1}) {
		t.Fatalf("expected selected paths [[1]], got %v", selected)
	}

	state.ClearSelection()
	if len(state.SelectedPaths()) != 0 {
		t.Fatalf("expected no selected paths after ClearSelection")
	}
}

func TestTreeStateSetChildren(t *testing.T) {
	state := NewTreeState([]TreeNode[string]{
		{Data: "Root", Children: nil},
	})

	state.SetChildren([]int{0}, []TreeNode[string]{
		{Data: "Child", Children: []TreeNode[string]{}},
	})

	if _, ok := state.NodeAtPath([]int{0, 0}); !ok {
		t.Fatalf("expected child node to be set")
	}
}
