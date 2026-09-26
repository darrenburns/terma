package terma

import "testing"

func sampleTreeSnapshotNodes() []TreeNode[string] {
	return []TreeNode[string]{
		{
			Data: "Project",
			Children: []TreeNode[string]{
				{Data: "README.md", Children: []TreeNode[string]{}},
				{
					Data: "cmd",
					Children: []TreeNode[string]{
						{Data: "main.go", Children: []TreeNode[string]{}},
					},
				},
			},
		},
		{Data: "LICENSE", Children: []TreeNode[string]{}},
	}
}

func TestSnapshot_Tree_Basic(t *testing.T) {
	state := NewTreeState(sampleTreeSnapshotNodes())
	widget := Tree[string]{
		ID:    "tree_basic",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 8, "Expanded tree with indicators and indentation for nested nodes")
}

func TestSnapshot_Tree_Collapsed(t *testing.T) {
	state := NewTreeState(sampleTreeSnapshotNodes())
	state.Collapse([]int{0})
	widget := Tree[string]{
		ID:    "tree_collapsed",
		State: state,
	}
	AssertSnapshot(t, widget, 40, 6, "Root node collapsed with collapse indicator and only top-level nodes visible")
}

func TestSnapshot_Tree_Filter(t *testing.T) {
	state := NewTreeState(sampleTreeSnapshotNodes())
	filter := NewFilterState()
	filter.Query.Set("main")
	widget := Tree[string]{
		ID:     "tree_filter",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "Filtered view showing Project -> cmd -> main.go with ancestors dimmed and match highlighted")
}

func TestSnapshot_Tree_InAutoWidthScrollable(t *testing.T) {
	roots := []TreeNode[string]{
		{Data: "Fruits", Children: []TreeNode[string]{{Data: "Apple"}, {Data: "Banana"}, {Data: "Cherry"}}},
		{Data: "Vegetables", Children: []TreeNode[string]{{Data: "Carrot"}}},
	}
	widget := Column{Children: []Widget{
		Text{Content: "Scrollable with no width set:"},
		Scrollable{ID: "scroll", State: NewScrollState(), Height: Cells(5), Child: Tree[string]{ID: "tree", State: NewTreeState(roots)}},
	}}
	AssertSnapshot(t, widget, 32, 6, "Tree sizes to its widest row inside an auto-width Scrollable, with a scrollbar beside it")
}
