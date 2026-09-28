package terma

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// FilterMode controls how text matching is performed.
type FilterMode int

const (
	// FilterContains matches contiguous substrings (default).
	FilterContains FilterMode = iota
	// FilterFuzzy matches characters in order (subsequence). Whitespace
	// separates terms, which may match in any order. Each match carries a
	// Score, and ranked consumers list the best matches first.
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
	// Score ranks fuzzy matches; higher is better. Scores only compare between
	// matches for the same query, and are zero outside FilterFuzzy. Between
	// matches of equal quality, the one on shorter text scores higher.
	Score int
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

// matchRanksAhead reports whether match a should be listed before match b.
func matchRanksAhead(a, b MatchResult) bool {
	if a.Matched != b.Matched {
		return a.Matched
	}
	return a.Score > b.Score
}

// sortFilteredViewByScore orders a view best match first, keeping the input
// order of matches with equal scores.
func sortFilteredViewByScore[T any](view *FilteredView[T]) {
	if view == nil || len(view.Items) < 2 {
		return
	}

	n := len(view.Items)
	if len(view.Indices) != n || len(view.Matches) != n {
		return
	}

	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		return matchRanksAhead(view.Matches[order[i]], view.Matches[order[j]])
	})

	items := make([]T, n)
	indices := make([]int, n)
	matches := make([]MatchResult, n)
	for i, originalIdx := range order {
		items[i] = view.Items[originalIdx]
		indices[i] = view.Indices[originalIdx]
		matches[i] = view.Matches[originalIdx]
	}

	view.Items = items
	view.Indices = indices
	view.Matches = matches
}

// MatchString matches query against text using the provided options.
func MatchString(text string, query string, options FilterOptions) MatchResult {
	if query == "" {
		return MatchResult{Matched: true}
	}

	if options.Mode == FilterFuzzy {
		return matchFuzzy(text, query, options.CaseSensitive)
	}

	haystack := text
	needle := query
	if !options.CaseSensitive {
		haystack = strings.ToLower(haystack)
		needle = strings.ToLower(needle)
	}
	return matchContains(haystack, needle)
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

// Fuzzy scoring follows fzf: every matched character scores, characters that
// start a word (or continue a run that started one) score extra, and skipped
// characters between matches cost a little. Characters before the first match
// are free, so "file" scores the same in "New File" and "Open File".
const (
	fuzzyScoreMatch               = 16
	fuzzyScoreGapStart            = -3
	fuzzyScoreGapExtension        = -1
	fuzzyBonusBoundaryWhite       = 10 // word start at the beginning or after whitespace
	fuzzyBonusBoundaryDelimiter   = 9  // word start after a delimiter such as / - _ .
	fuzzyBonusBoundary            = 8  // word start after other punctuation
	fuzzyBonusNonWord             = 8  // punctuation matched literally
	fuzzyBonusCamel               = 7  // camelCase hump or a letter-to-digit step
	fuzzyBonusConsecutive         = -(fuzzyScoreGapStart + fuzzyScoreGapExtension)
	fuzzyBonusFirstCharMultiplier = 2
	// fuzzyLengthScale leaves room below each point of match quality so that,
	// between matches of equal quality, the shorter text scores higher.
	fuzzyLengthScale = 256
)

type fuzzyCharClass int

const (
	fuzzyClassWhite fuzzyCharClass = iota
	fuzzyClassNonWord
	fuzzyClassDelimiter
	fuzzyClassLower
	fuzzyClassUpper
	fuzzyClassLetter
	fuzzyClassNumber
)

func fuzzyClassOf(r rune) fuzzyCharClass {
	switch {
	case unicode.IsSpace(r):
		return fuzzyClassWhite
	case unicode.IsLower(r):
		return fuzzyClassLower
	case unicode.IsUpper(r):
		return fuzzyClassUpper
	case unicode.IsLetter(r):
		return fuzzyClassLetter
	case unicode.IsNumber(r):
		return fuzzyClassNumber
	case strings.ContainsRune(`/\,:;|-_.`, r):
		return fuzzyClassDelimiter
	default:
		return fuzzyClassNonWord
	}
}

// fuzzyBonus is the bonus for matching a character of class cur that follows
// a character of class prev.
func fuzzyBonus(prev, cur fuzzyCharClass) int {
	if cur > fuzzyClassDelimiter {
		switch prev {
		case fuzzyClassWhite:
			return fuzzyBonusBoundaryWhite
		case fuzzyClassDelimiter:
			return fuzzyBonusBoundaryDelimiter
		case fuzzyClassNonWord:
			return fuzzyBonusBoundary
		}
	}
	if prev == fuzzyClassLower && cur == fuzzyClassUpper || prev != fuzzyClassNumber && cur == fuzzyClassNumber {
		return fuzzyBonusCamel
	}
	switch cur {
	case fuzzyClassNonWord, fuzzyClassDelimiter:
		return fuzzyBonusNonWord
	case fuzzyClassWhite:
		return fuzzyBonusBoundaryWhite
	}
	return 0
}

// matchFuzzy matches each whitespace-separated term of query as a subsequence
// of text, in any order, placing each term where it scores highest.
func matchFuzzy(text, query string, caseSensitive bool) MatchResult {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return MatchResult{Matched: true}
	}

	runes := make([]rune, 0, len(text))
	offsets := make([]int, 0, len(text)+1)
	bonuses := make([]int, 0, len(text))
	prevClass := fuzzyClassWhite
	for offset, r := range text {
		class := fuzzyClassOf(r)
		bonuses = append(bonuses, fuzzyBonus(prevClass, class))
		prevClass = class
		if !caseSensitive {
			r = unicode.ToLower(r)
		}
		runes = append(runes, r)
		offsets = append(offsets, offset)
	}
	offsets = append(offsets, len(text))

	total := 0
	var ranges []MatchRange
	for _, term := range terms {
		pattern := []rune(term)
		if !caseSensitive {
			for i, r := range pattern {
				pattern[i] = unicode.ToLower(r)
			}
		}
		score, positions, ok := fuzzyAlign(runes, bonuses, pattern)
		if !ok {
			return MatchResult{}
		}
		total += score
		for _, pos := range positions {
			ranges = append(ranges, MatchRange{Start: offsets[pos], End: offsets[pos+1]})
		}
	}

	return MatchResult{
		Matched: true,
		Ranges:  normalizeMatchRanges(ranges, len(text)),
		Score:   total*fuzzyLengthScale - min(len(runes), fuzzyLengthScale-1),
	}
}

// fuzzyAlign finds the highest-scoring placement of pattern as a subsequence
// of text, returning its score and the matched positions in text.
func fuzzyAlign(text []rune, bonuses []int, pattern []rune) (int, []int, bool) {
	m := len(pattern)
	if m == 0 {
		return 0, nil, true
	}

	// Greedy scans from each end reject non-matches cheaply and bound the
	// window any placement can use.
	start, pi := 0, 0
	for j := 0; j < len(text) && pi < m; j++ {
		if text[j] == pattern[pi] {
			if pi == 0 {
				start = j
			}
			pi++
		}
	}
	if pi < m {
		return 0, nil, false
	}
	end := len(text)
	pi = m - 1
	for j := len(text) - 1; j >= start && pi >= 0; j-- {
		if text[j] == pattern[pi] {
			if pi == m-1 {
				end = j + 1
			}
			pi--
		}
	}

	// Row i scores placements of pattern[:i+1] whose last character lands on
	// each window column. from records where the previous character landed on
	// each best path, for recovering the positions afterwards.
	const none = math.MinInt32
	w := end - start
	from := make([]int32, m*w)
	prevScore, curScore := make([]int32, w), make([]int32, w)
	prevRun, curRun := make([]int32, w), make([]int32, w)
	for i := 0; i < m; i++ {
		// Best placement of pattern[:i] ending two or more columns back, less
		// the penalty for the columns skipped since.
		gapScore, gapFrom := int32(none), int32(-1)
		for j := 0; j < w; j++ {
			if i > 0 && j >= 2 {
				if gapScore != none {
					gapScore += fuzzyScoreGapExtension
				}
				if s := prevScore[j-2]; s != none && s+fuzzyScoreGapStart > gapScore {
					gapScore, gapFrom = s+fuzzyScoreGapStart, int32(j-2)
				}
			}

			curScore[j], curRun[j] = none, 0
			pos := start + j
			if text[pos] != pattern[i] {
				continue
			}
			bonus := int32(bonuses[pos])
			if i == 0 {
				curScore[j] = fuzzyScoreMatch + bonus*fuzzyBonusFirstCharMultiplier
				curRun[j] = 1
				from[j] = -1
				continue
			}

			best, bestRun, bestFrom := int32(none), int32(0), int32(-1)
			if j > 0 && prevScore[j-1] != none {
				// A run keeps the bonus of the character that started it, so
				// matching a whole word scores like matching its start. A new
				// word start part-way through a run starts a new run.
				run := prevRun[j-1] + 1
				runBonus := int32(bonuses[pos-int(run)+1])
				if bonus >= fuzzyBonusBoundary && bonus > runBonus {
					run, runBonus = 1, bonus
				}
				best = prevScore[j-1] + fuzzyScoreMatch + max(bonus, runBonus, fuzzyBonusConsecutive)
				bestRun, bestFrom = run, int32(j-1)
			}
			if gapScore != none {
				if s := gapScore + fuzzyScoreMatch + bonus; s > best {
					best, bestRun, bestFrom = s, 1, gapFrom
				}
			}
			curScore[j], curRun[j], from[i*w+j] = best, bestRun, bestFrom
		}
		prevScore, curScore = curScore, prevScore
		prevRun, curRun = curRun, prevRun
	}

	bestEnd := -1
	for j := 0; j < w; j++ {
		if prevScore[j] != none && (bestEnd < 0 || prevScore[j] > prevScore[bestEnd]) {
			bestEnd = j
		}
	}
	if bestEnd < 0 {
		return 0, nil, false
	}
	positions := make([]int, m)
	for i, j := m-1, bestEnd; i >= 0; i-- {
		positions[i] = start + j
		j = int(from[i*w+j])
	}
	return int(prevScore[bestEnd]), positions, true
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

// MatchHighlightStyle returns the standard SpanStyle used to highlight
// matched text in filtered lists, tables, trees, and similar widgets.
func MatchHighlightStyle(theme ThemeData) SpanStyle {
	return SpanStyle{
		Underline:      UnderlineSingle,
		UnderlineColor: theme.Accent,
		Background:     theme.Selection,
	}
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
