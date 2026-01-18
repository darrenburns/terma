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
