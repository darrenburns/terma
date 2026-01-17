package terma

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// FilterMode controls how text matching is performed.
type FilterMode int

const (
	// FilterContains matches contiguous substrings (default).
	FilterContains FilterMode = iota
	// FilterFuzzy matches characters in order (subsequence).
	FilterFuzzy
)

// FilterOptions configures text matching behavior.
type FilterOptions struct {
	Mode          FilterMode
	CaseSensitive bool
}

// FilterState holds reactive filter input and matching options.
type FilterState struct {
	Query         Signal[string]
	Mode          Signal[FilterMode]
	CaseSensitive Signal[bool]
}

// NewFilterState creates a FilterState with default options.
func NewFilterState() *FilterState {
	return &FilterState{
		Query:         NewSignal(""),
		Mode:          NewSignal(FilterContains),
		CaseSensitive: NewSignal(false),
	}
}

// QueryText returns the current query text (subscribes to changes).
func (s *FilterState) QueryText() string {
	if s == nil {
		return ""
	}
	return s.Query.Get()
}

// PeekQuery returns the current query text without subscribing.
func (s *FilterState) PeekQuery() string {
	if s == nil {
		return ""
	}
	return s.Query.Peek()
}

// Options returns the current filter options (subscribes to changes).
func (s *FilterState) Options() FilterOptions {
	if s == nil {
		return FilterOptions{}
	}
	return FilterOptions{
		Mode:          s.Mode.Get(),
		CaseSensitive: s.CaseSensitive.Get(),
	}
}

// PeekOptions returns the current filter options without subscribing.
func (s *FilterState) PeekOptions() FilterOptions {
	if s == nil {
		return FilterOptions{}
	}
	return FilterOptions{
		Mode:          s.Mode.Peek(),
		CaseSensitive: s.CaseSensitive.Peek(),
	}
}

func filterStateValues(filter *FilterState) (string, FilterOptions) {
	if filter == nil {
		return "", FilterOptions{}
	}
	return filter.Query.Get(), filter.Options()
}

func filterStateValuesPeek(filter *FilterState) (string, FilterOptions) {
	if filter == nil {
		return "", FilterOptions{}
	}
	return filter.Query.Peek(), filter.PeekOptions()
}

// MatchRange defines a matched substring range [Start, End) in bytes.
type MatchRange struct {
	Start int
	End   int
}

// MatchResult represents match status and highlight ranges.
type MatchResult struct {
	Matched bool
	Ranges  []MatchRange
}

// FilteredView contains the filtered slice, source indices, and match data.
type FilteredView[T any] struct {
	Items   []T
	Indices []int
	Matches []MatchResult
}

// ApplyFilter filters items using the matcher and returns the view with match data.
func ApplyFilter[T any](items []T, query string, match func(item T, query string) MatchResult) FilteredView[T] {
	if match == nil {
		view := FilteredView[T]{
			Items:   items,
			Indices: make([]int, len(items)),
			Matches: make([]MatchResult, len(items)),
		}
		for i := range items {
			view.Indices[i] = i
			view.Matches[i] = MatchResult{Matched: true}
		}
		return view
	}

	if query == "" {
		view := FilteredView[T]{
			Items:   items,
			Indices: make([]int, len(items)),
			Matches: make([]MatchResult, len(items)),
		}
		for i := range items {
			view.Indices[i] = i
			view.Matches[i] = MatchResult{Matched: true}
		}
		return view
	}

	view := FilteredView[T]{
		Items:   make([]T, 0, len(items)),
		Indices: make([]int, 0, len(items)),
		Matches: make([]MatchResult, 0, len(items)),
	}
	for i, item := range items {
		result := match(item, query)
		if result.Matched {
			view.Items = append(view.Items, item)
			view.Indices = append(view.Indices, i)
			view.Matches = append(view.Matches, result)
		}
	}
	return view
}

// MatchString matches query against text using the provided options.
func MatchString(text string, query string, options FilterOptions) MatchResult {
	if query == "" {
		return MatchResult{Matched: true}
	}

	haystack := text
	needle := query
	if !options.CaseSensitive {
		haystack = strings.ToLower(haystack)
		needle = strings.ToLower(needle)
	}

	switch options.Mode {
	case FilterFuzzy:
		return matchFuzzy(text, haystack, needle)
	default:
		return matchContains(haystack, needle)
	}
}

func matchContains(haystack, needle string) MatchResult {
	if needle == "" {
		return MatchResult{Matched: true}
	}
	var ranges []MatchRange
	offset := 0
	for {
		idx := strings.Index(haystack[offset:], needle)
		if idx == -1 {
			break
		}
		start := offset + idx
		end := start + len(needle)
		ranges = append(ranges, MatchRange{Start: start, End: end})
		offset = end
	}
	return MatchResult{Matched: len(ranges) > 0, Ranges: ranges}
}

func matchFuzzy(original, haystack, needle string) MatchResult {
	if needle == "" {
		return MatchResult{Matched: true}
	}
	needleRunes := []rune(needle)
	var ranges []MatchRange
	needleIdx := 0

	for idx, r := range haystack {
		if needleIdx >= len(needleRunes) {
			break
		}
		if r == needleRunes[needleIdx] {
			size := utf8.RuneLen(r)
			if size <= 0 {
				size = 1
			}
			ranges = append(ranges, MatchRange{Start: idx, End: idx + size})
			needleIdx++
		}
	}

	if needleIdx < len(needleRunes) {
		return MatchResult{}
	}

	ranges = normalizeMatchRanges(ranges, len(original))
	return MatchResult{Matched: len(ranges) > 0, Ranges: ranges}
}

// HighlightSpans builds spans with highlight style applied to matched ranges.
func HighlightSpans(text string, ranges []MatchRange, highlight SpanStyle) []Span {
	if text == "" {
		return []Span{{Text: ""}}
	}

	normalized := normalizeMatchRanges(ranges, len(text))
	if len(normalized) == 0 {
		return []Span{{Text: text}}
	}

	spans := make([]Span, 0, len(normalized)*2+1)
	cursor := 0

	for _, r := range normalized {
		if r.Start > cursor {
			spans = append(spans, Span{Text: text[cursor:r.Start]})
		}
		if r.End > r.Start {
			spans = append(spans, Span{Text: text[r.Start:r.End], Style: highlight})
		}
		cursor = r.End
	}
	if cursor < len(text) {
		spans = append(spans, Span{Text: text[cursor:]})
	}

	return spans
}

func normalizeMatchRanges(ranges []MatchRange, textLen int) []MatchRange {
	if len(ranges) == 0 || textLen <= 0 {
		return nil
	}

	trimmed := make([]MatchRange, 0, len(ranges))
	for _, r := range ranges {
		start := clampInt(r.Start, 0, textLen)
		end := clampInt(r.End, 0, textLen)
		if end <= start {
			continue
		}
		trimmed = append(trimmed, MatchRange{Start: start, End: end})
	}

	if len(trimmed) == 0 {
		return nil
	}

	sort.Slice(trimmed, func(i, j int) bool {
		if trimmed[i].Start == trimmed[j].Start {
			return trimmed[i].End < trimmed[j].End
		}
		return trimmed[i].Start < trimmed[j].Start
	})

	merged := trimmed[:0]
	for _, r := range trimmed {
		if len(merged) == 0 {
			merged = append(merged, r)
			continue
		}
		last := &merged[len(merged)-1]
		if r.Start <= last.End {
			if r.End > last.End {
				last.End = r.End
			}
			continue
		}
		merged = append(merged, r)
	}
	return merged
}
