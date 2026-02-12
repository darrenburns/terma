package main

import (
	"strings"
	"testing"

	t "terma"

	"github.com/stretchr/testify/require"
)

func TestDiffApp_RefreshPreservesActiveFileWhenStillPresent(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("a.txt", "b.txt"),
			diffForPaths("b.txt", "c.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	require.True(t, app.selectFilePath("b.txt"))
	require.Equal(t, "b.txt", app.activePath)
	require.False(t, app.activeIsDir)

	app.refreshDiff()

	require.Equal(t, "b.txt", app.activePath)
	require.False(t, app.activeIsDir)
	require.Equal(t, app.filePathToTreePath["b.txt"], app.treeState.CursorPath.Peek())
}

func TestDiffApp_NextPrevCycleFilesAndSyncTreeCursor(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("pkg/b.go", "pkg/c.go", "a.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	require.GreaterOrEqual(t, len(app.orderedFilePaths), 3)

	first := app.orderedFilePaths[0]
	second := app.orderedFilePaths[1]
	last := app.orderedFilePaths[len(app.orderedFilePaths)-1]

	require.Equal(t, first, app.activePath)

	app.moveFileCursor(1)
	require.Equal(t, second, app.activePath)
	require.Equal(t, app.filePathToTreePath[second], app.treeState.CursorPath.Peek())

	app.moveFileCursor(-1)
	require.Equal(t, first, app.activePath)
	require.Equal(t, app.filePathToTreePath[first], app.treeState.CursorPath.Peek())

	app.moveFileCursor(-1)
	require.Equal(t, last, app.activePath)
	require.Equal(t, app.filePathToTreePath[last], app.treeState.CursorPath.Peek())
}

func TestDiffApp_DirectoryCursorShowsSummaryInViewer(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("pkg/a.go", "pkg/b.go", "README.md"),
		},
	}

	app := NewDiffApp(provider, false)
	dirPath, ok := findTreePathByDataPath(app.treeState.Nodes.Peek(), "pkg")
	require.True(t, ok)

	node, ok := app.treeState.NodeAtPath(dirPath)
	require.True(t, ok)
	app.treeState.CursorPath.Set(clonePath(dirPath))
	app.onTreeCursorChange(node.Data)

	require.True(t, app.activeIsDir)
	require.Equal(t, "pkg", app.activePath)

	rendered := app.diffViewState.Rendered.Peek()
	require.NotNil(t, rendered)
	require.GreaterOrEqual(t, len(rendered.Lines), 4)
	require.True(t, strings.Contains(lineText(rendered.Lines[0]), "Directory: pkg"))
	require.True(t, strings.Contains(lineText(rendered.Lines[1]), "Touched files: 2"))
}

type scriptedDiffProvider struct {
	repoRoot string
	diffs    []string
	index    int
}

func (p *scriptedDiffProvider) LoadDiff(staged bool) (string, error) {
	if len(p.diffs) == 0 {
		return "", nil
	}
	if p.index >= len(p.diffs) {
		return p.diffs[len(p.diffs)-1], nil
	}
	value := p.diffs[p.index]
	p.index++
	return value, nil
}

func (p *scriptedDiffProvider) RepoRoot() (string, error) {
	return p.repoRoot, nil
}

func diffForPaths(paths ...string) string {
	var builder strings.Builder
	for _, path := range paths {
		builder.WriteString("diff --git a/")
		builder.WriteString(path)
		builder.WriteString(" b/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("index 1111111..2222222 100644\n")
		builder.WriteString("--- a/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("+++ b/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("@@ -1 +1 @@\n")
		builder.WriteString("-old\n")
		builder.WriteString("+new\n")
	}
	return builder.String()
}

func findTreePathByDataPath(nodes []t.TreeNode[DiffTreeNodeData], target string) ([]int, bool) {
	var walk func(items []t.TreeNode[DiffTreeNodeData], prefix []int) ([]int, bool)
	walk = func(items []t.TreeNode[DiffTreeNodeData], prefix []int) ([]int, bool) {
		for idx, node := range items {
			next := append(clonePath(prefix), idx)
			if node.Data.Path == target {
				return next, true
			}
			if path, ok := walk(node.Children, next); ok {
				return path, true
			}
		}
		return nil, false
	}
	return walk(nodes, nil)
}
