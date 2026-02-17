package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

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

func TestTreeOnMouseDownMovesCursorToClickedRow(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.setViewPaths([][]int{
		{0},
		{0, 0},
		{0, 0, 0},
		{0, 1},
		{1},
	})
	state.rowLayouts = []treeRowLayout{
		{y: 0, height: 1},
		{y: 1, height: 1},
		{y: 2, height: 1},
		{y: 3, height: 1},
		{y: 4, height: 1},
	}

	tree := Tree[string]{State: state}
	tree.OnMouseDown(MouseEvent{LocalY: 3})

	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0, 1}) {
		t.Fatalf("expected cursor [0 1], got %v", got)
	}
}

func TestTreeOnMouseDownShiftExtendsSelection(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.setViewPaths([][]int{
		{0},
		{0, 0},
		{0, 0, 0},
		{0, 1},
		{1},
	})
	state.rowLayouts = []treeRowLayout{
		{y: 0, height: 1},
		{y: 1, height: 1},
		{y: 2, height: 1},
		{y: 3, height: 1},
		{y: 4, height: 1},
	}

	tree := Tree[string]{
		State:       state,
		MultiSelect: true,
	}

	tree.OnMouseDown(MouseEvent{LocalY: 4, Mod: uv.ModShift})

	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{1}) {
		t.Fatalf("expected cursor [1], got %v", got)
	}
	if anchor := state.getAnchor(); !pathsEqual(anchor, []int{0}) {
		t.Fatalf("expected anchor [0], got %v", anchor)
	}

	selected := state.SelectedPaths()
	if len(selected) != 5 {
		t.Fatalf("expected 5 selected paths, got %d (%v)", len(selected), selected)
	}
	if !pathsEqual(selected[0], []int{0}) || !pathsEqual(selected[len(selected)-1], []int{1}) {
		t.Fatalf("unexpected selected range: %v", selected)
	}
}

func TestTreeOnMouseDownEmptyAreaDoesNotMoveCursor(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.setViewPaths([][]int{
		{0},
		{0, 0},
		{0, 0, 0},
		{0, 1},
		{1},
	})
	state.rowLayouts = []treeRowLayout{
		{y: 0, height: 1},
		{y: 1, height: 1},
		{y: 2, height: 1},
		{y: 3, height: 1},
		{y: 4, height: 1},
	}

	tree := Tree[string]{State: state}
	tree.OnMouseDown(MouseEvent{LocalY: 99})

	if got := state.CursorPath.Peek(); !pathsEqual(got, []int{0}) {
		t.Fatalf("expected cursor to remain [0], got %v", got)
	}
}

func TestTreeOnMouseDownIndicatorExpandsNode(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.Collapse([]int{0})
	state.setViewPaths([][]int{
		{0},
		{1},
	})
	state.rowLayouts = []treeRowLayout{
		{y: 0, height: 1},
		{y: 1, height: 1},
	}
	state.indicatorLayout = []treeIndicatorLayout{
		{x: 0, width: 2, expandable: true},
		{x: 0, width: 2, expandable: false},
	}

	tree := Tree[string]{State: state}
	tree.OnMouseDown(MouseEvent{LocalX: 0, LocalY: 0})

	if state.IsCollapsed([]int{0}) {
		t.Fatalf("expected node [0] to expand after clicking indicator")
	}
}

func TestTreeOnMouseDownOutsideIndicatorDoesNotToggleExpansion(t *testing.T) {
	state := NewTreeState(sampleTreeNodes())
	state.Collapse([]int{0})
	state.setViewPaths([][]int{
		{0},
		{1},
	})
	state.rowLayouts = []treeRowLayout{
		{y: 0, height: 1},
		{y: 1, height: 1},
	}
	state.indicatorLayout = []treeIndicatorLayout{
		{x: 0, width: 2, expandable: true},
		{x: 0, width: 2, expandable: false},
	}

	tree := Tree[string]{State: state}
	tree.OnMouseDown(MouseEvent{LocalX: 5, LocalY: 0})

	if !state.IsCollapsed([]int{0}) {
		t.Fatalf("expected node [0] to remain collapsed when clicking outside indicator")
	}
}
