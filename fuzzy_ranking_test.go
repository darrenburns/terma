package terma

import (
	"reflect"
	"testing"
)

func fuzzyStringMatcher(item string, query string) MatchResult {
	return MatchString(item, query, FilterOptions{Mode: FilterFuzzy})
}

func newTestBuildContext() BuildContext {
	return NewBuildContext(
		NewFocusManager(),
		NewAnySignal[Focusable](nil),
		NewAnySignal[Widget](nil),
		NewFloatCollector(),
	)
}

func TestSortFilteredViewByFuzzyRank_PrefersEarlierStart(t *testing.T) {
	items := []string{"xa---b", "ab---", "zab"}
	view := ApplyFilter(items, "ab", fuzzyStringMatcher)

	sortFilteredViewByFuzzyRank(&view)

	wantIndices := []int{1, 2, 0}
	if !reflect.DeepEqual(view.Indices, wantIndices) {
		t.Fatalf("unexpected sort order: got %v, want %v", view.Indices, wantIndices)
	}
}

func TestSortFilteredViewByFuzzyRank_PrefersTighterWhenStartEqual(t *testing.T) {
	items := []string{"a---b", "ab---", "a--b"}
	view := ApplyFilter(items, "ab", fuzzyStringMatcher)

	sortFilteredViewByFuzzyRank(&view)

	wantIndices := []int{1, 2, 0}
	if !reflect.DeepEqual(view.Indices, wantIndices) {
		t.Fatalf("unexpected sort order: got %v, want %v", view.Indices, wantIndices)
	}
}

func TestSortFilteredViewByFuzzyRank_StableOnExactTie(t *testing.T) {
	opts := FilterOptions{Mode: FilterFuzzy}
	view := FilteredView[string]{
		Items:   []string{"ab2", "ab1", "ab3"},
		Indices: []int{2, 0, 1},
		Matches: []MatchResult{
			MatchString("ab2", "ab", opts),
			MatchString("ab1", "ab", opts),
			MatchString("ab3", "ab", opts),
		},
	}

	sortFilteredViewByFuzzyRank(&view)

	wantIndices := []int{2, 0, 1}
	if !reflect.DeepEqual(view.Indices, wantIndices) {
		t.Fatalf("unexpected sort order for equal ranks: got %v, want %v", view.Indices, wantIndices)
	}
}

func TestListStateApplyFilter_FuzzySortsByRank(t *testing.T) {
	state := NewListState([]string{"a---b", "ab---", "a--b"})
	filter := NewFilterState()
	filter.Query.Set("ab")
	filter.Mode.Set(FilterFuzzy)

	count := state.ApplyFilter(filter, nil)
	if count != 3 {
		t.Fatalf("expected 3 matches, got %d", count)
	}

	want := []int{1, 2, 0}
	if !reflect.DeepEqual(state.viewIndices, want) {
		t.Fatalf("unexpected list view order: got %v, want %v", state.viewIndices, want)
	}
}

func TestListBuild_FuzzyCachedAndUncachedOrderMatch(t *testing.T) {
	filter := NewFilterState()
	filter.Query.Set("ab")
	filter.Mode.Set(FilterFuzzy)

	cachedState := NewListState([]string{"a---b", "ab---", "a--b"})
	cachedState.ApplyFilter(filter, nil)
	cachedList := List[string]{State: cachedState, Filter: filter}
	_ = cachedList.Build(newTestBuildContext())
	cachedOrder := append([]int(nil), cachedState.viewIndices...)

	uncachedState := NewListState([]string{"a---b", "ab---", "a--b"})
	uncachedList := List[string]{State: uncachedState, Filter: filter}
	_ = uncachedList.Build(newTestBuildContext())
	uncachedOrder := append([]int(nil), uncachedState.viewIndices...)

	if !reflect.DeepEqual(cachedOrder, uncachedOrder) {
		t.Fatalf("cached and uncached list order differ: cached=%v uncached=%v", cachedOrder, uncachedOrder)
	}

	want := []int{1, 2, 0}
	if !reflect.DeepEqual(cachedOrder, want) {
		t.Fatalf("unexpected list order: got %v, want %v", cachedOrder, want)
	}
}

func TestTableFilteredRows_FuzzySortsByBestCellRank(t *testing.T) {
	rows := [][]string{
		{"zzab", "x"},
		{"a---b", "x"},
		{"x", "abxx"},
		{"x", "a--b"},
	}
	table := Table[[]string]{}

	_, indices, _ := table.filteredRows(rows, 2, "ab", FilterOptions{Mode: FilterFuzzy})

	want := []int{2, 3, 1, 0}
	if !reflect.DeepEqual(indices, want) {
		t.Fatalf("unexpected table row order: got %v, want %v", indices, want)
	}
}

func TestTableFilteredRows_FuzzyStableOnExactTie(t *testing.T) {
	rows := [][]string{
		{"ab1", "x"},
		{"ab2", "x"},
		{"ab3", "x"},
	}
	table := Table[[]string]{}

	_, indices, _ := table.filteredRows(rows, 2, "ab", FilterOptions{Mode: FilterFuzzy})

	want := []int{0, 1, 2}
	if !reflect.DeepEqual(indices, want) {
		t.Fatalf("unexpected table row order for equal ranks: got %v, want %v", indices, want)
	}
}

func TestCommandPaletteFilteredView_FuzzySortsByRank(t *testing.T) {
	items := []CommandPaletteItem{
		{Label: "a---b"},
		{Label: "ab---"},
		{Label: "a--b"},
	}
	filter := NewFilterState()
	filter.Query.Set("ab")
	filter.Mode.Set(FilterFuzzy)

	view := commandPaletteFilteredView(items, filter)

	want := []int{1, 2, 0}
	if !reflect.DeepEqual(view.Indices, want) {
		t.Fatalf("unexpected command palette order: got %v, want %v", view.Indices, want)
	}
}

func TestCommandPaletteFilteredView_ContainsKeepsInputOrder(t *testing.T) {
	items := []CommandPaletteItem{
		{Label: "xab"},
		{Label: "abx"},
		{Label: "zab"},
	}
	filter := NewFilterState()
	filter.Query.Set("ab")
	filter.Mode.Set(FilterContains)

	view := commandPaletteFilteredView(items, filter)

	want := []int{0, 1, 2}
	if !reflect.DeepEqual(view.Indices, want) {
		t.Fatalf("contains mode should preserve input order: got %v, want %v", view.Indices, want)
	}
}

func TestTreeBuildViewEntries_FuzzyKeepsSiblingOrder(t *testing.T) {
	nodes := []TreeNode[string]{
		{Data: "a---b"},
		{Data: "ab---"},
		{Data: "a--b"},
	}
	tree := Tree[string]{}

	entries := tree.buildViewEntries(nodes, "ab", FilterOptions{Mode: FilterFuzzy})

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	got := []string{entries[0].node.Data, entries[1].node.Data, entries[2].node.Data}
	want := []string{"a---b", "ab---", "a--b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tree sibling order changed under fuzzy filtering: got %v, want %v", got, want)
	}
}
