package terma

import "testing"

// TestSnapshot_ScrollableList_ContentWidth verifies that list items
// are not truncated by the scrollbar when scrolling is active.
func TestSnapshot_ScrollableList_ContentWidth(t *testing.T) {
	// Create a list with enough items to require scrolling
	items := []string{
		"Item 1 - Last char visible",
		"Item 2 - Last char visible",
		"Item 3 - Last char visible",
		"Item 4 - Last char visible",
		"Item 5 - Last char visible",
		"Item 6 - Last char visible",
		"Item 7 - Last char visible",
		"Item 8 - Last char visible",
		"Item 9 - Last char visible",
		"Item 10 - Last char visible",
		"Item 11 - Last char visible",
		"Item 12 - Last char visible",
	}

	listState := NewListState(items)
	scrollState := NewScrollState()

	widget := &scrollableListTestWidget{
		listState:   listState,
		scrollState: scrollState,
	}

	// Width 30 allows text to fit; height 5 forces scrolling
	AssertSnapshot(t, widget, 30, 5, "List items should not be truncated by scrollbar")
}

type scrollableListTestWidget struct {
	listState   *ListState[string]
	scrollState *ScrollState
}

func (w *scrollableListTestWidget) Build(ctx BuildContext) Widget {
	return Scrollable{
		State:  w.scrollState,
		Height: Cells(5),
		Width:  Cells(30),
		Child: List[string]{
			ID:          "test-list",
			State:       w.listState,
			ScrollState: w.scrollState,
		},
	}
}
