package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListFilterRefreshesWhenInputsChange(t *testing.T) {
	for _, tc := range []struct {
		name   string
		items  []string
		query  string
		change func(*ListState[string], *FilterState)
		want   string
		absent string
	}{
		{
			name:  "case sensitivity",
			items: []string{"Apple", "apricot"}, query: "a",
			change: func(_ *ListState[string], filter *FilterState) { filter.CaseSensitive.Set(true) },
			want:   "apricot", absent: "Apple",
		},
		{
			name:  "matching mode",
			items: []string{"apple", "pear"}, query: "ae",
			change: func(_ *ListState[string], filter *FilterState) { filter.Mode.Set(FilterFuzzy) },
			want:   "apple", absent: "pear",
		},
		{
			name:  "replacement items",
			items: []string{"apple", "pear"}, query: "app",
			change: func(state *ListState[string], _ *FilterState) { state.Items.Set([]string{"pear", "apple"}) },
			want:   "apple", absent: "pear",
		},
		{
			name:  "in-place item update",
			items: []string{"apple", "pear"}, query: "app",
			change: func(state *ListState[string], _ *FilterState) {
				state.Items.Update(func(items []string) []string {
					items[0], items[1] = items[1], items[0]
					return items
				})
			},
			want: "apple", absent: "pear",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := NewListState(tc.items)
			filter := NewFilterState()
			filter.Query.Set(tc.query)
			list := List[string]{ID: "list", State: state, Filter: filter}
			p := NewPilot(t, list, 20, 4)
			t.Cleanup(func() { p.session.renderer.rootNode.dispose() })

			tc.change(state, filter)

			screen := p.ScreenText()
			require.Contains(t, screen, tc.want)
			require.NotContains(t, screen, tc.absent)
			p.AssertSnapshot("filtered", "The list uses the current filter settings and source items")
		})
	}
}

func TestDefaultListCursorProjectsOntoFilteredView(t *testing.T) {
	state := NewListState([]string{"apple", "banana", "cherry", "blueberry"})
	filter := NewFilterState()
	var activated []string
	list := List[string]{
		ID: "list", State: state, Filter: filter, CursorStyle: CursorStyle{CursorPrefix: "> "},
		MultiSelect: true, OnSelect: func(item string) { activated = append(activated, item) },
	}
	p := NewPilot(t, list, 20, 4)
	t.Cleanup(func() { p.session.renderer.rootNode.dispose() })
	screen := p.ScreenText
	require.Contains(t, screen(), "> apple")

	filter.Query.Set("b")
	require.Contains(t, screen(), "> banana")
	require.Equal(t, 0, state.CursorIndex.Peek(), "rendering must not change stored cursor state")

	state.CursorIndex.Set(3)
	require.Contains(t, screen(), "> blueberry")

	state.CursorIndex.Set(2)
	require.Contains(t, screen(), "> banana")
	require.Equal(t, 2, state.CursorIndex.Peek())
	p.AssertSnapshot("filtered", "The default renderer shows the cursor on the first visible item without mutating source state")

	p.session.focus.FocusByID("list")
	p.Press("enter")
	require.Equal(t, []string{"banana"}, activated)
	require.Equal(t, 1, state.CursorIndex.Peek())
	p.Press("shift+down")
	require.Equal(t, []string{"banana", "blueberry"}, state.SelectedItems())
	require.Contains(t, screen(), "> blueberry")
	p.Press("up")
	require.Empty(t, state.SelectedItems())
	require.Contains(t, screen(), "> banana")

	state.SetItems([]string{"blackberry", "pear"})
	require.Contains(t, screen(), "> blackberry")
	require.NotContains(t, screen(), "pear")
	state.SetItems(nil)
	require.NotContains(t, screen(), "blackberry")
	state.Items.Set([]string{"pear", "boysenberry"})
	require.Contains(t, screen(), "> boysenberry")
	require.NotContains(t, screen(), "pear")
}
