package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiffViewState_ClampScrollXAndY(t *testing.T) {
	rendered := buildTestRenderedFile(40, 80)
	state := NewDiffViewState(rendered)
	gutterWidth := renderedGutterWidth(rendered)

	state.SetViewport(30, 10, gutterWidth)
	state.ScrollY.Set(999)
	state.ScrollX.Set(999)

	state.Clamp(gutterWidth)

	require.Equal(t, 30, state.MaxScrollY())
	require.Equal(t, 58, state.MaxScrollX(gutterWidth))
	require.Equal(t, 30, state.ScrollY.Peek())
	require.Equal(t, 58, state.ScrollX.Peek())
}

func TestDiffViewState_PageAndHalfPageSteps(t *testing.T) {
	rendered := buildTestRenderedFile(100, 20)
	state := NewDiffViewState(rendered)
	gutterWidth := renderedGutterWidth(rendered)
	state.SetViewport(40, 12, gutterWidth)

	state.PageDown(gutterWidth)
	require.Equal(t, 11, state.ScrollY.Peek())

	state.HalfPageDown(gutterWidth)
	require.Equal(t, 17, state.ScrollY.Peek())

	state.PageUp(gutterWidth)
	require.Equal(t, 6, state.ScrollY.Peek())

	state.HalfPageUp(gutterWidth)
	require.Equal(t, 0, state.ScrollY.Peek())
}

func TestDiffViewState_GoTopAndGoBottom(t *testing.T) {
	rendered := buildTestRenderedFile(25, 10)
	state := NewDiffViewState(rendered)
	gutterWidth := renderedGutterWidth(rendered)
	state.SetViewport(20, 5, gutterWidth)

	state.GoBottom(gutterWidth)
	require.Equal(t, 20, state.ScrollY.Peek())

	state.GoTop(gutterWidth)
	require.Equal(t, 0, state.ScrollY.Peek())
}

func buildTestRenderedFile(lineCount int, contentWidth int) *RenderedFile {
	lines := make([]RenderedDiffLine, 0, lineCount)
	for i := 0; i < lineCount; i++ {
		lines = append(lines, RenderedDiffLine{
			Kind:         RenderedLineContext,
			OldLine:      i + 1,
			NewLine:      i + 1,
			Prefix:       " ",
			Segments:     []RenderedSegment{{Text: "x", Role: TokenRoleSyntaxPlain}},
			ContentWidth: contentWidth,
		})
	}
	return &RenderedFile{
		Title:           "test",
		Lines:           lines,
		OldNumWidth:     2,
		NewNumWidth:     2,
		MaxContentWidth: contentWidth,
	}
}
