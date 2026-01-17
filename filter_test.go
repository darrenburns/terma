package terma

import (
	"fmt"
	"testing"
)

// =============================================================================
// List Filter Tests
// =============================================================================

func TestSnapshot_List_Filter_Contains(t *testing.T) {
	state := NewListState([]string{"Apple", "Banana", "Cherry", "Apricot", "Blueberry"})
	filter := NewFilterState()
	filter.Query.Set("ap") // Should match "Apple" and "Apricot" (case-insensitive)

	widget := List[string]{
		ID:     "list_filter_contains",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 8, "Two items visible: 'Apple' and 'Apricot' with 'ap' highlighted in accent color")
}

func TestSnapshot_List_Filter_CaseSensitive(t *testing.T) {
	state := NewListState([]string{"Apple", "apple", "APPLE", "Apricot"})
	filter := NewFilterState()
	filter.Query.Set("Apple")
	filter.CaseSensitive.Set(true)

	widget := List[string]{
		ID:     "list_filter_case_sensitive",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "Only 'Apple' visible (exact case match), lowercase 'apple' and uppercase 'APPLE' filtered out")
}

func TestSnapshot_List_Filter_Fuzzy(t *testing.T) {
	state := NewListState([]string{"JavaScript", "TypeScript", "CoffeeScript", "Java", "Python"})
	filter := NewFilterState()
	filter.Query.Set("jsc") // Should match "JavaScript" and "CoffeeScript" (fuzzy)
	filter.Mode.Set(FilterFuzzy)

	widget := List[string]{
		ID:     "list_filter_fuzzy",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 8, "One item visible: 'JavaScript' with fuzzy-matched characters 'J', 'S', 'c' highlighted")
}

func TestSnapshot_List_Filter_NoMatches(t *testing.T) {
	state := NewListState([]string{"Apple", "Banana", "Cherry"})
	filter := NewFilterState()
	filter.Query.Set("xyz") // No matches

	widget := List[string]{
		ID:     "list_filter_no_matches",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 5, "Empty list - no items match the query 'xyz'")
}

func TestSnapshot_List_Filter_EmptyQuery(t *testing.T) {
	state := NewListState([]string{"Apple", "Banana", "Cherry"})
	filter := NewFilterState()
	// Empty query should show all items

	widget := List[string]{
		ID:     "list_filter_empty_query",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "All three items visible - empty query shows unfiltered list")
}

func TestSnapshot_List_Filter_WithSelection(t *testing.T) {
	state := NewListState([]string{"Apple", "Apricot", "Banana", "Cherry"})
	filter := NewFilterState()
	filter.Query.Set("ap")
	state.SelectIndex(0) // Cursor should be on first filtered item

	widget := List[string]{
		ID:     "list_filter_with_selection",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "Two filtered items with cursor (▶) on 'Apple', both showing 'ap' highlighted")
}

func TestSnapshot_List_Filter_CustomMatcher(t *testing.T) {
	type Item struct {
		Name  string
		Value int
	}

	items := []Item{
		{Name: "First", Value: 100},
		{Name: "Second", Value: 200},
		{Name: "Third", Value: 150},
	}

	state := NewListState(items)
	filter := NewFilterState()
	filter.Query.Set("first")

	// Custom matcher that only matches on the Name field
	matchItem := func(item Item, query string, options FilterOptions) MatchResult {
		return MatchString(item.Name, query, options)
	}

	// Custom renderer
	renderItem := func(item Item, active bool, selected bool, match MatchResult) Widget {
		content := fmt.Sprintf("%s (%d)", item.Name, item.Value)
		prefix := "  "
		if active {
			prefix = "▶ "
		}
		return Text{Content: prefix + content}
	}

	widget := List[Item]{
		ID:                  "list_filter_custom_matcher",
		State:               state,
		Filter:              filter,
		MatchItem:           matchItem,
		RenderItemWithMatch: renderItem,
	}
	AssertSnapshot(t, widget, 40, 6, "Only 'First (100)' visible - custom matcher filters by Name field only")
}

func TestSnapshot_List_Filter_Highlighting(t *testing.T) {
	state := NewListState([]string{"Apple Pie", "Apricot Tart", "Banana Bread"})
	filter := NewFilterState()
	filter.Query.Set("ap")

	// The default renderer should highlight matching portions
	widget := List[string]{
		ID:     "list_filter_highlighting",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "Two items with 'ap' highlighted in accent color with underline on both 'Apple' and 'Apricot'")
}

// =============================================================================
// Table Filter Tests
// =============================================================================

func TestSnapshot_Table_Filter_Contains(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer", "NYC"},
		{"Bob", "Designer", "LA"},
		{"Charlie", "Engineer", "SF"},
		{"Diana", "Manager", "NYC"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("engineer") // Should match Alice and Charlie

	widget := Table[[]string]{
		ID:     "table_filter_contains",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(12)},
			{Width: Cells(8)},
		},
	}
	AssertSnapshot(t, widget, 50, 8, "Two rows visible: Alice and Charlie, both with 'Engineer' highlighted in the role column")
}

func TestSnapshot_Table_Filter_CaseSensitive(t *testing.T) {
	rows := [][]string{
		{"Alice", "alice@example.com"},
		{"ALICE", "ALICE@EXAMPLE.COM"},
		{"Bob", "bob@example.com"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("Alice")
	filter.CaseSensitive.Set(true)

	widget := Table[[]string]{
		ID:     "table_filter_case_sensitive",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(20)},
		},
	}
	AssertSnapshot(t, widget, 50, 8, "Only first row visible with 'Alice' highlighted - case-sensitive match excludes 'ALICE' row")
}

func TestSnapshot_Table_Filter_Fuzzy(t *testing.T) {
	rows := [][]string{
		{"JavaScript", "Frontend"},
		{"TypeScript", "Frontend"},
		{"Python", "Backend"},
		{"Java", "Backend"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("jsc") // Should fuzzy match JavaScript
	filter.Mode.Set(FilterFuzzy)

	widget := Table[[]string]{
		ID:     "table_filter_fuzzy",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(15)},
			{Width: Cells(12)},
		},
	}
	AssertSnapshot(t, widget, 50, 8, "One row visible: 'JavaScript' with fuzzy-matched characters 'J', 'S', 'c' highlighted")
}

func TestSnapshot_Table_Filter_NoMatches(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer"},
		{"Bob", "Designer"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("xyz")

	widget := Table[[]string]{
		ID:     "table_filter_no_matches",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(12)},
		},
	}
	AssertSnapshot(t, widget, 50, 6, "Empty table - no rows match the query 'xyz'")
}

func TestSnapshot_Table_Filter_EmptyQuery(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer"},
		{"Bob", "Designer"},
		{"Charlie", "Manager"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	// Empty query should show all rows

	widget := Table[[]string]{
		ID:     "table_filter_empty_query",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(12)},
		},
	}
	AssertSnapshot(t, widget, 50, 7, "All three rows visible - empty query shows unfiltered table")
}

func TestSnapshot_Table_Filter_WithSelection(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer"},
		{"Bob", "Designer"},
		{"Charlie", "Engineer"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("engineer")
	state.SelectIndex(0) // Cursor should be on first filtered row

	widget := Table[[]string]{
		ID:     "table_filter_with_selection",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(12)},
		},
	}
	AssertSnapshot(t, widget, 50, 7, "Two filtered rows with Alice row highlighted (active), both showing 'Engineer' with match highlighting")
}

func TestSnapshot_Table_Filter_WithHeaders(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer", "NYC"},
		{"Bob", "Designer", "LA"},
		{"Charlie", "Engineer", "SF"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("engineer")

	widget := Table[[]string]{
		ID:     "table_filter_with_headers",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(10), Header: Text{Content: "Name"}},
			{Width: Cells(12), Header: Text{Content: "Role"}},
			{Width: Cells(8), Header: Text{Content: "City"}},
		},
	}
	AssertSnapshot(t, widget, 50, 8, "Header row visible (Name, Role, City) followed by two filtered data rows with 'Engineer' highlighted")
}

func TestSnapshot_Table_Filter_CustomMatcher(t *testing.T) {
	type Employee struct {
		Name string
		Age  int
	}

	rows := []Employee{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
		{Name: "Charlie", Age: 35},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("alice")

	// Custom matcher that only matches the Name field in column 0
	matchCell := func(row Employee, rowIndex int, colIndex int, query string, options FilterOptions) MatchResult {
		if colIndex == 0 {
			return MatchString(row.Name, query, options)
		}
		return MatchResult{Matched: false}
	}

	// Custom renderer
	renderCell := func(row Employee, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget {
		var content string
		if colIndex == 0 {
			content = row.Name
		} else {
			content = fmt.Sprintf("%d", row.Age)
		}
		return Text{Content: content}
	}

	widget := Table[Employee]{
		ID:                  "table_filter_custom_matcher",
		State:               state,
		Filter:              filter,
		MatchCell:           matchCell,
		RenderCellWithMatch: renderCell,
		Columns: []TableColumn{
			{Width: Cells(10)},
			{Width: Cells(8)},
		},
	}
	AssertSnapshot(t, widget, 50, 6, "Only Alice row visible - custom renderer shows plain text without highlighting")
}

func TestSnapshot_Table_Filter_Highlighting(t *testing.T) {
	rows := [][]string{
		{"Apple Inc", "Technology"},
		{"Apricot Co", "Food"},
		{"Banana Corp", "Agriculture"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("ap")

	// The default renderer should highlight matching portions
	widget := Table[[]string]{
		ID:     "table_filter_highlighting",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(15)},
			{Width: Cells(15)},
		},
	}
	AssertSnapshot(t, widget, 50, 7, "Two rows visible with 'ap' highlighted in accent color in first column of each row")
}

// =============================================================================
// Filter Utility Tests
// =============================================================================

func TestSnapshot_List_Filter_MultipleMatches(t *testing.T) {
	state := NewListState([]string{
		"apple apple apple",
		"banana",
		"apple pie",
	})
	filter := NewFilterState()
	filter.Query.Set("apple")

	widget := List[string]{
		ID:     "list_filter_multiple_matches",
		State:  state,
		Filter: filter,
	}
	AssertSnapshot(t, widget, 40, 6, "Two items visible with all occurrences of 'apple' highlighted (3x in first item, 1x in second)")
}

func TestSnapshot_Table_Filter_MatchAcrossCells(t *testing.T) {
	rows := [][]string{
		{"Alice", "Engineer", "NYC"},
		{"Bob", "Designer", "LA"},
		{"NYC Admin", "Manager", "NYC"},
	}

	state := NewTableState(rows)
	filter := NewFilterState()
	filter.Query.Set("nyc") // Should match row 0 (city) and row 2 (name and city)

	widget := Table[[]string]{
		ID:     "table_filter_match_across_cells",
		State:  state,
		Filter: filter,
		Columns: []TableColumn{
			{Width: Cells(12)},
			{Width: Cells(12)},
			{Width: Cells(8)},
		},
	}
	AssertSnapshot(t, widget, 50, 7, "Two rows visible: Alice (NYC in city column) and NYC Admin (NYC in both name and city columns)")
}
