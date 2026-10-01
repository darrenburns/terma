package terma

import (
	"slices"
	"sort"
	"strings"
)

// TextRange is a range of grapheme indices [Start, End).
type TextRange struct {
	Start int
	End   int
}

// textAreaSearchCache holds the matches derived from the content and the
// search inputs, so rendering doesn't rescan the text every frame.
type textAreaSearchCache struct {
	core          *anySignalCore[[]string]
	revision      uint64
	query         string
	caseSensitive bool
	matches       []TextRange
}

func (s *TextAreaState) matchesFor(graphemes []string, revision uint64, query string, caseSensitive bool) []TextRange {
	cache := &s.search
	if cache.core != s.Content.core || cache.revision != revision || cache.query != query || cache.caseSensitive != caseSensitive {
		cache.core = s.Content.core
		cache.revision = revision
		cache.query = query
		cache.caseSensitive = caseSensitive
		cache.matches = findMatches(graphemes, splitGraphemes(query), caseSensitive)
	}
	return cache.matches
}

// findMatches returns the non-overlapping occurrences of query in graphemes,
// left to right.
func findMatches(graphemes, query []string, caseSensitive bool) []TextRange {
	if len(query) == 0 {
		return nil
	}
	equal := strings.EqualFold
	if caseSensitive {
		equal = func(a, b string) bool { return a == b }
	}
	var matches []TextRange
	for i := 0; i+len(query) <= len(graphemes); {
		j := 0
		for j < len(query) && equal(graphemes[i+j], query[j]) {
			j++
		}
		if j < len(query) {
			i++
			continue
		}
		matches = append(matches, TextRange{Start: i, End: i + len(query)})
		i += len(query)
	}
	return matches
}

func (s *TextAreaState) peekMatches() []TextRange {
	graphemes, revision := s.Content.peekWithRevision()
	return s.matchesFor(graphemes, revision, s.SearchQuery.Peek(), s.SearchCaseSensitive.Peek())
}

// currentMatch returns the index of the match the selection covers exactly.
func currentMatch(matches []TextRange, selStart, selEnd int) (int, bool) {
	i := sort.Search(len(matches), func(i int) bool { return matches[i].Start >= selStart })
	if i < len(matches) && matches[i] == (TextRange{Start: selStart, End: selEnd}) {
		return i, true
	}
	return -1, false
}

// firstMatchFrom returns the index of the first match starting at or after
// index, wrapping to the first match.
func firstMatchFrom(matches []TextRange, index int) int {
	i := sort.Search(len(matches), func(i int) bool { return matches[i].Start >= index })
	if i == len(matches) {
		return 0
	}
	return i
}

func (s *TextAreaState) searchOrigin() int {
	if start, _ := s.GetSelectionBounds(); start >= 0 {
		return start
	}
	return s.CursorIndex.Peek()
}

func (s *TextAreaState) selectMatch(match TextRange) {
	s.SelectionAnchor.Set(match.Start)
	s.CursorIndex.Set(match.End)
	s.resetPreferredColumn()
	if s.revealCursor != nil {
		s.revealCursor()
	}
}

// SetSearch highlights matches of query and selects the first match at or
// after the selection (or the cursor), wrapping to the first match. Extending
// the query therefore keeps the current match while it still matches. When
// nothing matches (including an empty query), a selection that was the
// previous match collapses to its start, so a half-typed query leaves no stray
// selection; any other selection stays.
func (s *TextAreaState) SetSearch(query string) {
	start, end := s.GetSelectionBounds()
	_, onMatch := currentMatch(s.peekMatches(), start, end)
	s.SearchQuery.Set(query)
	matches := s.peekMatches()
	if len(matches) == 0 {
		if onMatch {
			s.ClearSelection()
			s.CursorIndex.Set(start)
			s.resetPreferredColumn()
		}
		return
	}
	s.selectMatch(matches[firstMatchFrom(matches, s.searchOrigin())])
}

// ClearSearch removes the search and its highlights. Unlike SetSearch(""),
// it leaves the current match selected, for closing a search bar on the result.
func (s *TextAreaState) ClearSearch() {
	s.SearchQuery.Set("")
}

// NextMatch selects the match after the current one, or the first match at or
// after the cursor, wrapping around.
func (s *TextAreaState) NextMatch() {
	matches := s.peekMatches()
	if len(matches) == 0 {
		return
	}
	start, end := s.GetSelectionBounds()
	if i, ok := currentMatch(matches, start, end); ok {
		s.selectMatch(matches[(i+1)%len(matches)])
		return
	}
	s.selectMatch(matches[firstMatchFrom(matches, s.CursorIndex.Peek())])
}

// PreviousMatch selects the match before the current one, or the last match
// before the selection (or the cursor), wrapping around.
func (s *TextAreaState) PreviousMatch() {
	matches := s.peekMatches()
	if len(matches) == 0 {
		return
	}
	start, end := s.GetSelectionBounds()
	if i, ok := currentMatch(matches, start, end); ok {
		s.selectMatch(matches[(i-1+len(matches))%len(matches)])
		return
	}
	origin := s.searchOrigin()
	i := sort.Search(len(matches), func(i int) bool { return matches[i].Start >= origin }) - 1
	if i < 0 {
		i = len(matches) - 1
	}
	s.selectMatch(matches[i])
}

// SearchMatches returns the ranges matching the search query. Called from
// Build, it subscribes to the content and the search inputs.
func (s *TextAreaState) SearchMatches() []TextRange {
	graphemes, revision := s.Content.getWithRevision()
	return slices.Clone(s.matchesFor(graphemes, revision, s.SearchQuery.Get(), s.SearchCaseSensitive.Get()))
}

// SearchPosition returns the 1-based index of the selected match (0 when the
// selection isn't a match) and the number of matches. Called from Build, it
// subscribes to the content, selection and search inputs.
func (s *TextAreaState) SearchPosition() (current, total int) {
	graphemes, revision := s.Content.getWithRevision()
	matches := s.matchesFor(graphemes, revision, s.SearchQuery.Get(), s.SearchCaseSensitive.Get())
	start, end := selectionBounds(s.SelectionAnchor.Get(), s.CursorIndex.Get(), len(graphemes))
	if i, ok := currentMatch(matches, start, end); ok {
		return i + 1, len(matches)
	}
	return 0, len(matches)
}
