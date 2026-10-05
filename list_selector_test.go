package terma

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCustomListCursorProjection(t *testing.T) {
	state := NewListState([]string{"apple", "banana", "cherry", "blueberry"})
	filter := NewFilterState()
	builds := map[string]int{}
	list := List[string]{ID: "list", State: state, Filter: filter,
		RenderItem: func(item string, active, selected bool) Widget {
			builds[item]++
			prefix := "  "
			if active {
				prefix = "> "
			}
			return Text{Content: prefix + item}
		}}
	p := NewPilot(t, list, 20, 4)
	t.Cleanup(func() { p.session.renderer.rootNode.dispose() })
	screen := p.ScreenText
	require.Contains(t, screen(), "> apple")

	clear(builds)
	state.CursorIndex.Set(2)
	p.settle()
	require.Equal(t, map[string]int{"apple": 1, "cherry": 1}, builds, "direct writes rebuild only the old and new cursor rows")
	require.Contains(t, screen(), "> cherry")

	for _, tc := range []struct {
		cursor int
		item   string
	}{{-5, "apple"}, {99, "blueberry"}} {
		state.CursorIndex.Set(tc.cursor)
		p.settle()
		require.Equal(t, tc.cursor, state.CursorIndex.Peek(), "painting clamps the projection without changing stored state")
		require.Contains(t, screen(), "> "+tc.item)
	}

	filter.Query.Set("b")
	p.settle()
	state.CursorIndex.Set(2) // Cherry is absent: show the first visible source index.
	p.settle()
	require.Contains(t, screen(), "> banana")
	require.Equal(t, 2, state.CursorIndex.Peek())
	state.CursorIndex.Set(3)
	p.settle()
	require.Contains(t, screen(), "> blueberry")

	filter.Query.Set("a") // Rebuild with a different source/view mapping.
	p.settle()
	require.Contains(t, screen(), "> apple")
	state.CursorIndex.Set(1)
	p.settle()
	require.Contains(t, screen(), "> banana", "selectors consult the current filtered view")
}

// Isolate signal selector fanout from layout and painting. The collection scroll
// benchmark covers the complete cursor step and frame with the same item counts.
func BenchmarkCustomListCursorNotification(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("items=%d", count), func(b *testing.B) {
			items := make([]string, count)
			for i := range items {
				items[i] = fmt.Sprintf("Item %05d", i)
			}
			state := NewListState(items)
			list := List[string]{ID: "bench-list", State: state,
				RenderItem: func(item string, active, selected bool) Widget {
					return Text{Content: item, Style: Style{Width: Flex(1)}}
				}}
			renderer := collectionScrollBenchRenderer(list.ID, list)
			renderer.Update(list)
			b.Cleanup(func() { renderer.rootNode.dispose() })
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				state.CursorIndex.Set(1 + i%2)
			}
		})
	}
}
