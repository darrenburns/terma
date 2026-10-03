package terma

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
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
			scene := newClickScene(t, list, 20, 4)
			t.Cleanup(func() { scene.renderer.rootNode.dispose() })

			tc.change(state, filter)
			scene.draw()

			screen := ansi.Strip(BufferToANSI(scene.buf, scene.width, scene.height))
			require.Contains(t, screen, tc.want)
			require.NotContains(t, screen, tc.absent)
			scene.snapshot(t.Name(), "The list uses the current filter settings and source items")
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
	scene := newClickScene(t, list, 20, 4)
	t.Cleanup(func() { scene.renderer.rootNode.dispose() })
	scene.draw()
	screen := func() string { return ansi.Strip(BufferToANSI(scene.buf, scene.width, scene.height)) }
	require.Contains(t, screen(), "> apple")

	filter.Query.Set("b")
	scene.draw()
	require.Contains(t, screen(), "> banana")
	require.Equal(t, 0, state.CursorIndex.Peek(), "rendering must not change stored cursor state")

	state.CursorIndex.Set(3)
	scene.draw()
	require.Contains(t, screen(), "> blueberry")

	state.CursorIndex.Set(2)
	scene.draw()
	require.Contains(t, screen(), "> banana")
	require.Equal(t, 2, state.CursorIndex.Peek())
	scene.snapshot(t.Name(), "The default renderer shows the cursor on the first visible item without mutating source state")

	press := func(key string) {
		t.Helper()
		for _, bind := range list.Keybinds() {
			if bind.Key == key {
				bind.Action()
				scene.draw()
				return
			}
		}
		t.Fatalf("missing key binding %q", key)
	}
	press("enter")
	require.Equal(t, []string{"banana"}, activated)
	require.Equal(t, 1, state.CursorIndex.Peek())
	press("shift+down")
	require.Equal(t, []string{"banana", "blueberry"}, state.SelectedItems())
	require.Contains(t, screen(), "> blueberry")
	press("up")
	require.Empty(t, state.SelectedItems())
	require.Contains(t, screen(), "> banana")

	state.SetItems([]string{"blackberry", "pear"})
	scene.draw()
	require.Contains(t, screen(), "> blackberry")
	require.NotContains(t, screen(), "pear")
	state.SetItems(nil)
	scene.draw()
	require.NotContains(t, screen(), "blackberry")
	state.Items.Set([]string{"pear", "boysenberry"})
	scene.draw()
	require.Contains(t, screen(), "> boysenberry")
	require.NotContains(t, screen(), "pear")
}
