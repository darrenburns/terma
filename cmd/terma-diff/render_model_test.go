package main

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestBuildRenderedFile_StructuredLinesAndGutters(t *testing.T) {
	file := &DiffFile{
		NewPath:     "main.go",
		DisplayPath: "main.go",
		Headers:     []string{"diff --git a/main.go b/main.go", "--- a/main.go", "+++ b/main.go"},
		Hunks: []DiffHunk{
			{
				Header: "@@ -1,2 +1,2 @@",
				Lines: []DiffLine{
					{Kind: DiffLineContext, Content: "package main", OldLine: 1, NewLine: 1},
					{Kind: DiffLineRemove, Content: "fmt.Println(\"old\")", OldLine: 2},
					{Kind: DiffLineAdd, Content: "fmt.Println(\"new\")", NewLine: 2},
				},
			},
		},
	}

	rendered := buildRenderedFile(file)
	require.NotNil(t, rendered)
	require.Equal(t, "main.go", rendered.Title)
	require.Len(t, rendered.Lines, 4)
	require.Equal(t, 1, rendered.OldNumWidth)
	require.Equal(t, 1, rendered.NewNumWidth)

	hunk := rendered.Lines[0]
	require.Equal(t, RenderedLineHunkHeader, hunk.Kind)
	require.Equal(t, "@@ -1,2 +1,2 @@", lineText(hunk))

	context := rendered.Lines[1]
	require.Equal(t, RenderedLineContext, context.Kind)
	require.Equal(t, 1, context.OldLine)
	require.Equal(t, 1, context.NewLine)
	require.Equal(t, " ", context.Prefix)
	require.Equal(t, "package main", lineText(context))

	remove := rendered.Lines[2]
	require.Equal(t, RenderedLineRemove, remove.Kind)
	require.Equal(t, 2, remove.OldLine)
	require.Equal(t, 0, remove.NewLine)
	require.Equal(t, "-", remove.Prefix)
	require.Equal(t, "fmt.Println(\"old\")", lineText(remove))

	add := rendered.Lines[3]
	require.Equal(t, RenderedLineAdd, add.Kind)
	require.Equal(t, 0, add.OldLine)
	require.Equal(t, 2, add.NewLine)
	require.Equal(t, "+", add.Prefix)
	require.Equal(t, "fmt.Println(\"new\")", lineText(add))
}

func TestBuildRenderedFile_ExpandsTabsWithTabStops(t *testing.T) {
	file := &DiffFile{
		NewPath:     "main.go",
		DisplayPath: "main.go",
		Hunks: []DiffHunk{
			{
				Header: "@@ -1,1 +1,1 @@",
				Lines: []DiffLine{
					{Kind: DiffLineContext, Content: "\tfoo\tbar", OldLine: 1, NewLine: 1},
				},
			},
		},
	}

	rendered := buildRenderedFile(file)
	require.NotNil(t, rendered)
	require.Len(t, rendered.Lines, 2)

	line := rendered.Lines[1]
	require.Equal(t, RenderedLineContext, line.Kind)
	require.Equal(t, "    foo bar", lineText(line))
	require.Equal(t, ansi.StringWidth("    foo bar"), line.ContentWidth)
}

func TestBuildMetaRenderedFile_ProducesMetaLines(t *testing.T) {
	rendered := buildMetaRenderedFile("Summary", []string{"Line one", "Line two"})
	require.NotNil(t, rendered)
	require.Equal(t, "Summary", rendered.Title)
	require.Len(t, rendered.Lines, 2)
	require.Equal(t, RenderedLineMeta, rendered.Lines[0].Kind)
	require.Equal(t, "Line one", lineText(rendered.Lines[0]))
	require.Equal(t, "Line two", lineText(rendered.Lines[1]))
}
