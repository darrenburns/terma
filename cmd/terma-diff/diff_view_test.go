package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLineNumberRolesForLine(tt *testing.T) {
	oldRole, newRole := lineNumberRolesForLine(RenderedLineContext)
	require.Equal(tt, TokenRoleOldLineNumber, oldRole)
	require.Equal(tt, TokenRoleNewLineNumber, newRole)

	oldRole, newRole = lineNumberRolesForLine(RenderedLineAdd)
	require.Equal(tt, TokenRoleLineNumberAdd, oldRole)
	require.Equal(tt, TokenRoleLineNumberAdd, newRole)

	oldRole, newRole = lineNumberRolesForLine(RenderedLineRemove)
	require.Equal(tt, TokenRoleLineNumberRemove, oldRole)
	require.Equal(tt, TokenRoleLineNumberRemove, newRole)
}

func TestWrappedLineRowCount(tt *testing.T) {
	line := RenderedDiffLine{
		Segments:     []RenderedSegment{{Text: "abcdefghi", Role: TokenRoleSyntaxPlain}},
		ContentWidth: 9,
	}
	require.Equal(tt, 3, wrappedLineRowCount(line, 4))
	require.Equal(tt, 1, wrappedLineRowCount(line, 20))

	empty := RenderedDiffLine{}
	require.Equal(tt, 1, wrappedLineRowCount(empty, 4))
}

func TestWrappedLineAtRow(tt *testing.T) {
	lines := []RenderedDiffLine{
		{
			Kind:         RenderedLineContext,
			OldLine:      1,
			NewLine:      1,
			Prefix:       " ",
			Segments:     []RenderedSegment{{Text: "abc", Role: TokenRoleSyntaxPlain}},
			ContentWidth: 3,
		},
		{
			Kind:         RenderedLineAdd,
			OldLine:      2,
			NewLine:      2,
			Prefix:       "+",
			Segments:     []RenderedSegment{{Text: "abcdefgh", Role: TokenRoleSyntaxPlain}},
			ContentWidth: 8,
		},
	}

	line, wrapRow, ok := wrappedLineAtRow(lines, 4, 0)
	require.True(tt, ok)
	require.Equal(tt, 1, line.NewLine)
	require.Equal(tt, 0, wrapRow)

	line, wrapRow, ok = wrappedLineAtRow(lines, 4, 1)
	require.True(tt, ok)
	require.Equal(tt, 2, line.NewLine)
	require.Equal(tt, 0, wrapRow)

	line, wrapRow, ok = wrappedLineAtRow(lines, 4, 2)
	require.True(tt, ok)
	require.Equal(tt, 2, line.NewLine)
	require.Equal(tt, 1, wrapRow)

	_, _, ok = wrappedLineAtRow(lines, 4, 3)
	require.False(tt, ok)
}

func TestWrappedContentHeight(tt *testing.T) {
	lines := []RenderedDiffLine{
		{ContentWidth: 3},
		{ContentWidth: 8},
	}
	require.Equal(tt, 3, wrappedContentHeight(lines, 4))
	require.Equal(tt, 2, wrappedContentHeight(lines, 10))
	require.Equal(tt, 1, wrappedContentHeight(nil, 4))
}
