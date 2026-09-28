package terma

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Each level has ten siblings, like expanded directories in a file browser.
func performanceTreeNodes(depth int) []TreeNode[string] {
	nodes := make([]TreeNode[string], 10)
	for i := range nodes {
		nodes[i].Data = fmt.Sprintf("Node %d", i)
		if depth > 1 {
			nodes[i].Children = performanceTreeNodes(depth - 1)
		} else {
			nodes[i].Children = []TreeNode[string]{}
		}
	}
	return nodes
}

func BenchmarkTreeFlattenVisible(b *testing.B) {
	for _, depth := range []int{2, 3} {
		nodes := performanceTreeNodes(depth)
		tree := Tree[string]{State: NewTreeState(nodes)}
		count := len(tree.flattenVisible(nodes, nil, 0))
		b.Run(fmt.Sprintf("nodes=%d", count), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if got := tree.flattenVisible(nodes, nil, 0); len(got) != count {
					b.Fatalf("flattened %d nodes, want %d", len(got), count)
				}
			}
		})
	}
}

func TestTreeFlattenVisibleRetainsPreorderAndIndependentPaths(t *testing.T) {
	nodes := sampleTreeNodes()
	state := NewTreeState(nodes)
	tree := Tree[string]{State: state}
	entries := tree.flattenVisible(nodes, nil, 0)
	var data []string
	var paths [][]int
	var depths []int
	for _, entry := range entries {
		data = append(data, entry.node.Data)
		paths = append(paths, entry.path)
		depths = append(depths, entry.depth)
	}
	require.Equal(t, []string{"A", "A1", "A1a", "A2", "B"}, data)
	require.Equal(t, [][]int{{0}, {0, 0}, {0, 0, 0}, {0, 1}, {1}}, paths)
	require.Equal(t, []int{0, 1, 2, 1, 0}, depths)
	entries[2].path[0] = 9
	require.Equal(t, []int{0}, entries[0].path)
	require.Equal(t, []int{0, 0}, entries[1].path)
	require.Equal(t, []int{0, 1}, entries[3].path)

	state.Collapse([]int{0, 0})
	entries = tree.flattenVisible(nodes, nil, 0)
	data = nil
	for _, entry := range entries {
		data = append(data, entry.node.Data)
	}
	require.Equal(t, []string{"A", "A1", "A2", "B"}, data)
	require.True(t, entries[1].expandable)
	require.False(t, entries[1].expanded)
}
