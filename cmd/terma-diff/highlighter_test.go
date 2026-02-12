package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	terma "terma"
)

func TestDiffHighlighter_ByteRangesToGraphemeRanges(t *testing.T) {
	theme, ok := terma.GetTheme(terma.ThemeNameRosePine)
	require.True(t, ok)

	h := DiffHighlighter{
		Tokens: []HighlightToken{
			{StartByte: 2, EndByte: 5, Role: TokenRoleDiffMeta},
		},
		Palette: NewThemePalette(theme),
	}

	graphemes := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	highlights := h.Highlight("0123456789", graphemes)

	require.Len(t, highlights, 1)
	require.Equal(t, 2, highlights[0].Start)
	require.Equal(t, 5, highlights[0].End)
}
