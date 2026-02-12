package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildRenderedFile_IncludesLineNumbersAndPrefixes(t *testing.T) {
	file := &DiffFile{
		Headers: []string{"diff --git a/main.go b/main.go", "--- a/main.go", "+++ b/main.go"},
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
	require.NotEmpty(t, rendered.Text)

	lines := strings.Split(rendered.Text, "\n")
	require.GreaterOrEqual(t, len(lines), 7)
	require.Contains(t, rendered.Text, "1 1   package main")
	require.Contains(t, rendered.Text, "2   - fmt.Println(\"old\")")
	require.Contains(t, rendered.Text, " 2 + fmt.Println(\"new\")")

	var hasAddPrefix bool
	var hasRemovePrefix bool
	for _, tok := range rendered.Tokens {
		if tok.Role == TokenRoleDiffPrefixAdd {
			hasAddPrefix = true
		}
		if tok.Role == TokenRoleDiffPrefixRemove {
			hasRemovePrefix = true
		}
	}
	require.True(t, hasAddPrefix)
	require.True(t, hasRemovePrefix)
}
