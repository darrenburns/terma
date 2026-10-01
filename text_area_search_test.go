package terma

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func searchMatchTexts(state *TextAreaState) []string {
	graphemes := state.Content.Peek()
	var texts []string
	for _, m := range state.SearchMatches() {
		texts = append(texts, joinGraphemes(graphemes[m.Start:m.End]))
	}
	return texts
}

func selectedRange(state *TextAreaState) TextRange {
	start, end := state.GetSelectionBounds()
	return TextRange{Start: start, End: end}
}

func TestTextAreaSearch_Matching(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		query         string
		caseSensitive bool
		want          []TextRange
	}{
		{name: "case-insensitive by default", text: "Fox fox FOX", query: "fox", want: []TextRange{{0, 3}, {4, 7}, {8, 11}}},
		{name: "case-sensitive", text: "Fox fox FOX", query: "fox", caseSensitive: true, want: []TextRange{{4, 7}}},
		{name: "non-overlapping", text: "aaaa", query: "aa", want: []TextRange{{0, 2}, {2, 4}}},
		{name: "odd run leaves the remainder", text: "aaa", query: "aa", want: []TextRange{{0, 2}}},
		// "é" is two runes but one grapheme, so indices count graphemes.
		{name: "multi-rune graphemes", text: "café 👍🏽 café", query: "é 👍🏽", want: []TextRange{{3, 6}}},
		{name: "emoji query", text: "a👨‍👩‍👧b👨‍👩‍👧", query: "👨‍👩‍👧", want: []TextRange{{1, 2}, {3, 4}}},
		{name: "a bare letter does not match a combined grapheme", text: "café", query: "e", want: nil},
		{name: "newline in query", text: "one\ntwo\none\nthree", query: "one\nt", want: []TextRange{{0, 5}, {8, 13}}},
		{name: "empty query", text: "anything", query: "", want: nil},
		{name: "no match", text: "anything", query: "zebra", want: nil},
		{name: "query longer than text", text: "ab", query: "abc", want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewTextAreaState(tt.text)
			state.SearchCaseSensitive.Set(tt.caseSensitive)
			state.SearchQuery.Set(tt.query)
			require.Equal(t, tt.want, state.SearchMatches())
			_, total := state.SearchPosition()
			require.Equal(t, len(tt.want), total)
		})
	}
}

func TestTextAreaSearch_SettingQueryDirectlyLeavesCursor(t *testing.T) {
	state := NewTextAreaState("fox and fox")
	state.CursorIndex.Set(5)
	state.SearchQuery.Set("fox")
	require.Equal(t, []string{"fox", "fox"}, searchMatchTexts(state))
	require.Equal(t, 5, state.CursorIndex.Peek())
	require.False(t, state.HasSelection())
	current, total := state.SearchPosition()
	require.Equal(t, 0, current)
	require.Equal(t, 2, total)
}

func TestTextAreaSearch_SetSearch(t *testing.T) {
	t.Run("selects the first match at or after the cursor", func(t *testing.T) {
		state := NewTextAreaState("fox one fox two fox")
		state.CursorIndex.Set(5)
		state.SetSearch("fox")
		require.Equal(t, TextRange{8, 11}, selectedRange(state))
		require.Equal(t, 11, state.CursorIndex.Peek(), "the cursor sits at the end of the match")
		current, total := state.SearchPosition()
		require.Equal(t, 2, current)
		require.Equal(t, 3, total)
	})

	t.Run("wraps to the first match", func(t *testing.T) {
		state := NewTextAreaState("fox one fox two")
		state.SetSearch("fox") // cursor starts at the end of the text
		require.Equal(t, TextRange{0, 3}, selectedRange(state))
	})

	t.Run("extending the query keeps the current match", func(t *testing.T) {
		state := NewTextAreaState("fog then fox then fog then fox")
		state.CursorIndex.Set(5)
		state.SetSearch("f")
		require.Equal(t, TextRange{9, 10}, selectedRange(state))
		state.SetSearch("fo")
		require.Equal(t, TextRange{9, 11}, selectedRange(state))
		state.SetSearch("fox")
		require.Equal(t, TextRange{9, 12}, selectedRange(state))
		state.SetSearch("fog")
		require.Equal(t, TextRange{18, 21}, selectedRange(state), "the next fog after the old match")
	})

	t.Run("no match leaves cursor and selection alone", func(t *testing.T) {
		state := NewTextAreaState("hello world")
		state.SetSelectionAnchor(1)
		state.CursorIndex.Set(4)
		state.SetSearch("zebra")
		require.Equal(t, TextRange{1, 4}, selectedRange(state))
		require.Equal(t, 4, state.CursorIndex.Peek())
		_, total := state.SearchPosition()
		require.Zero(t, total)
	})

	t.Run("a query that stops matching collapses the match it had selected", func(t *testing.T) {
		state := NewTextAreaState("five dozen jugs")
		state.SetSearch("ze")
		require.Equal(t, TextRange{7, 9}, selectedRange(state))
		state.SetSearch("zeb")
		require.False(t, state.HasSelection(), "no stray selection from the half-typed query")
		require.Equal(t, 7, state.CursorIndex.Peek(), "the cursor stays at the old match so backspacing finds it again")
		state.SetSearch("ze")
		require.Equal(t, TextRange{7, 9}, selectedRange(state))
	})

	t.Run("empty query clears the search and collapses its match", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.CursorIndex.Set(0)
		state.SetSearch("fox")
		state.SetSearch("")
		require.Empty(t, state.SearchMatches())
		require.False(t, state.HasSelection())
		require.Equal(t, 0, state.CursorIndex.Peek())
	})

	t.Run("ClearSearch keeps the current match selected", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.CursorIndex.Set(0)
		state.SetSearch("fox")
		state.ClearSearch()
		require.Empty(t, state.SearchMatches())
		require.Equal(t, TextRange{0, 3}, selectedRange(state))
		current, total := state.SearchPosition()
		require.Zero(t, current)
		require.Zero(t, total)
	})
}

func TestTextAreaSearch_NextAndPrevious(t *testing.T) {
	state := NewTextAreaState("ab ab ab")
	state.CursorIndex.Set(0)
	state.SearchQuery.Set("ab")

	var visited []int
	for i := 0; i < 4; i++ {
		state.NextMatch()
		current, _ := state.SearchPosition()
		visited = append(visited, current)
	}
	require.Equal(t, []int{1, 2, 3, 1}, visited, "next wraps from the last match to the first")

	visited = nil
	for i := 0; i < 4; i++ {
		state.PreviousMatch()
		current, _ := state.SearchPosition()
		visited = append(visited, current)
	}
	require.Equal(t, []int{3, 2, 1, 3}, visited, "previous wraps from the first match to the last")
}

func TestTextAreaSearch_NextAndPreviousFromCursor(t *testing.T) {
	text := "ab ab ab"
	tests := []struct {
		name   string
		cursor int
		next   TextRange
		prev   TextRange
	}{
		{name: "between matches", cursor: 4, next: TextRange{6, 8}, prev: TextRange{3, 5}},
		{name: "at a match start", cursor: 3, next: TextRange{3, 5}, prev: TextRange{0, 2}},
		{name: "before the first", cursor: 0, next: TextRange{0, 2}, prev: TextRange{6, 8}},
		{name: "after the last", cursor: 8, next: TextRange{0, 2}, prev: TextRange{6, 8}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewTextAreaState(text)
			state.SearchQuery.Set("ab")

			state.CursorIndex.Set(tt.cursor)
			state.NextMatch()
			require.Equal(t, tt.next, selectedRange(state), "next")

			state.ClearSelection()
			state.CursorIndex.Set(tt.cursor)
			state.PreviousMatch()
			require.Equal(t, tt.prev, selectedRange(state), "previous")
		})
	}
}

func TestTextAreaSearch_PreviousFromSelectionStart(t *testing.T) {
	state := NewTextAreaState("ab ab ab")
	state.SearchQuery.Set("ab")
	// A selection that isn't a match: previous counts from its start.
	state.SetSelectionAnchor(4)
	state.CursorIndex.Set(7)
	state.PreviousMatch()
	require.Equal(t, TextRange{3, 5}, selectedRange(state), "counting from the cursor would pick 6-8")
}

func TestTextAreaSearch_NoMatchesIsNoOp(t *testing.T) {
	state := NewTextAreaState("hello")
	state.CursorIndex.Set(2)
	state.SearchQuery.Set("zebra")
	state.NextMatch()
	state.PreviousMatch()
	require.Equal(t, 2, state.CursorIndex.Peek())
	require.False(t, state.HasSelection())
}

func TestTextAreaSearch_CurrentFollowsSelection(t *testing.T) {
	t.Run("moving the cursor leaves the match", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.CursorIndex.Set(0)
		state.SetSearch("fox")
		current, _ := state.SearchPosition()
		require.Equal(t, 1, current)

		state.ClearSelection()
		state.CursorLeft()
		current, total := state.SearchPosition()
		require.Equal(t, 0, current)
		require.Equal(t, 2, total)
	})

	t.Run("editing the text leaves the match", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.CursorIndex.Set(0)
		state.SetSearch("fox")
		state.Insert("!") // typing clears the selection
		current, total := state.SearchPosition()
		require.Equal(t, 0, current)
		require.Equal(t, 2, total)
	})

	t.Run("selecting a match by hand makes it current", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.SearchQuery.Set("fox")
		state.SetSelectionAnchor(7)
		state.CursorIndex.Set(4) // reversed selection over the second match
		current, _ := state.SearchPosition()
		require.Equal(t, 2, current)
	})

	t.Run("a selection wider than a match is not current", func(t *testing.T) {
		state := NewTextAreaState("fox fox")
		state.SearchQuery.Set("fox")
		state.SetSelectionAnchor(0)
		state.CursorIndex.Set(4)
		current, _ := state.SearchPosition()
		require.Equal(t, 0, current)
	})
}

func TestTextAreaSearch_MatchesFollowContentAndInputs(t *testing.T) {
	state := NewTextAreaState("cat")
	state.SearchQuery.Set("cat")
	require.Equal(t, []TextRange{{0, 3}}, state.SearchMatches())

	state.CursorIndex.Set(3)
	state.Insert(" CAT")
	require.Equal(t, []TextRange{{0, 3}, {4, 7}}, state.SearchMatches(), "after an edit")

	state.SearchCaseSensitive.Set(true)
	require.Equal(t, []TextRange{{0, 3}}, state.SearchMatches(), "after turning on case sensitivity")

	state.Content.Update(func(g []string) []string {
		g[0] = "b" // in-place edits published through Update count too
		return g
	})
	require.Empty(t, state.SearchMatches(), "after an in-place content update")

	state.SetText("cat cat cat")
	require.Len(t, state.SearchMatches(), 3, "after SetText")

	state.ClearSearch()
	require.Empty(t, state.SearchMatches(), "after ClearSearch")
}

func TestTextAreaSearch_ReturnedMatchesAreACopy(t *testing.T) {
	state := NewTextAreaState("fox fox")
	state.SearchQuery.Set("fox")
	state.SearchMatches()[0] = TextRange{5, 6}
	require.Equal(t, []TextRange{{0, 3}, {4, 7}}, state.SearchMatches())
}

func TestSnapshot_TextArea_SearchMatches(t *testing.T) {
	state := NewTextAreaState("The fox saw a fox.\nA Fox ran past the\nfoxglove by the fox.")
	state.CursorIndex.Set(20)
	state.SetSearch("fox")

	textArea := TextArea{ID: "search", State: state, Width: Cells(22), Height: Cells(3)}
	AssertSnapshotNamed(t, "TextArea_Search_Unfocused", Column{Children: []Widget{Button{ID: "focus-stealer", Label: ""}, textArea}}, 22, 4,
		"Five 'fox' matches (any case) highlighted. The 'Fox' on row 2 (after the button row) is the current match in the accent colour; the others use the plain match style.")
	AssertSnapshotNamed(t, "TextArea_Search_Focused", textArea, 22, 3,
		"Same search, focused: the cursor sits just after the current 'Fox' on row 2.")

	state.ClearSelection()
	state.SetSelectionAnchor(0)
	state.CursorIndex.Set(3)
	AssertSnapshotNamed(t, "TextArea_Search_SelectionElsewhere", Column{Children: []Widget{Button{ID: "focus-stealer", Label: ""}, textArea}}, 22, 4,
		"A real selection over 'The' with no current match. The selection must look different from the five plain matches.")
}

func searchScrollDocument() string {
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, fmt.Sprintf("line %02d some filler words", i))
	}
	lines[30] = "line 30 has the needle in it"
	return strings.Join(lines, "\n")
}

func TestSnapshot_TextArea_SearchScrollsToMatch(t *testing.T) {
	for _, wrap := range []WrapMode{WrapSoft, WrapNone} {
		name := "wrap"
		if wrap == WrapNone {
			name = "nowrap"
		}
		t.Run(name, func(t *testing.T) {
			state := NewTextAreaState(searchScrollDocument())
			state.WrapMode.Set(wrap)
			state.CursorIndex.Set(0)
			state.SetSearch("needle")

			widget := TextArea{ID: "search-scroll", State: state, Width: Cells(18), Height: Cells(4)}
			AssertSnapshotNamed(t, "TextArea_SearchScrollsToMatch_"+name, widget, 18, 4,
				"The current 'needle' match on line 30 is scrolled into view in a 40-line document with a 4-row viewport.")
		})
	}
}
