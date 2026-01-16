package terma

import "testing"

func TestListScrollIntoView_UsesLayoutMetrics(t *testing.T) {
	state := NewListState([]string{"A", "B", "C"})
	state.SelectIndex(1)

	scrollState := NewScrollState()
	scrollState.updateLayout(3, 10)

	state.itemLayouts = []listItemLayout{
		{y: 0, height: 2},
		{y: 2, height: 4},
		{y: 6, height: 1},
	}

	list := List[string]{
		State:       state,
		ScrollState: scrollState,
		ItemHeight:  1,
	}

	list.scrollCursorIntoView()

	if got := scrollState.GetOffset(); got != 3 {
		t.Fatalf("expected scroll offset 3, got %d", got)
	}
}
