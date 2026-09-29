package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// Large collections inside a Scrollable, as apps use them. Each step moves the
// cursor past the edge of the viewport (so every frame scrolls), or scrolls the
// viewport with the mouse wheel.

const collectionScrollBenchViewport = 30

func collectionScrollBenchRenderer(focusedID string, focused Focusable) *Renderer {
	renderer := NewRenderer(
		uv.NewBuffer(reactivityBenchWidth, reactivityBenchHeight),
		reactivityBenchWidth, reactivityBenchHeight,
		NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil),
	)
	renderer.focusManager.focusedID = focusedID
	renderer.focusedSignal.Set(focused)
	return renderer
}

// runCollectionScrollBench renders the first frame outside the timer, moves the
// cursor to the bottom of the viewport, then times step+Update. Steps
// alternate direction in long runs so the viewport scrolls on every step.
func runCollectionScrollBench(b *testing.B, renderer *Renderer, root Widget, step func(down bool)) {
	b.Helper()
	b.Cleanup(func() {
		if renderer.rootNode != nil {
			renderer.rootNode.dispose()
		}
	})
	renderer.Update(root)
	// Park the cursor at the viewport's bottom edge.
	for i := 0; i < collectionScrollBenchViewport; i++ {
		step(true)
		renderer.Update(root)
	}
	b.ReportAllocs()
	b.ResetTimer()
	const run = 200
	for i := 0; i < b.N; i++ {
		down := (i/run)%2 == 0
		step(down)
		renderer.Update(root)
		// Mirror the app loop: a frame that left more work queued (for example
		// a scroll made by OnLayout) renders again.
		renderer.Update(root)
	}
}

func BenchmarkCollectionScroll(b *testing.B) {
	for _, count := range []int{1000, 10000} {
		b.Run(fmt.Sprintf("list/items=%d/cursor", count), func(b *testing.B) {
			items := make([]string, count)
			for i := range items {
				items[i] = fmt.Sprintf("Item %05d", i)
			}
			state := NewListState(items)
			scroll := NewScrollState()
			list := List[string]{ID: "bench-list", State: state, ScrollState: scroll}
			root := Scrollable{State: scroll, Height: Cells(collectionScrollBenchViewport), Child: list}
			runCollectionScrollBench(b, collectionScrollBenchRenderer(list.ID, list), root, func(down bool) {
				if down {
					list.keyCursorDown()
				} else {
					list.keyCursorUp()
				}
			})
		})
		b.Run(fmt.Sprintf("list/items=%d/custom-cursor", count), func(b *testing.B) {
			items := make([]string, count)
			for i := range items {
				items[i] = fmt.Sprintf("Item %05d", i)
			}
			state := NewListState(items)
			scroll := NewScrollState()
			list := List[string]{ID: "bench-list", State: state, ScrollState: scroll,
				RenderItem: func(item string, active, selected bool) Widget {
					prefix := "  "
					if active {
						prefix = "> "
					}
					return Text{Content: prefix + item, Style: Style{Width: Flex(1)}}
				}}
			root := Scrollable{State: scroll, Height: Cells(collectionScrollBenchViewport), Child: list}
			runCollectionScrollBench(b, collectionScrollBenchRenderer(list.ID, list), root, func(down bool) {
				if down {
					list.keyCursorDown()
				} else {
					list.keyCursorUp()
				}
			})
		})
		b.Run(fmt.Sprintf("list/items=%d/wheel", count), func(b *testing.B) {
			items := make([]string, count)
			for i := range items {
				items[i] = fmt.Sprintf("Item %05d", i)
			}
			state := NewListState(items)
			scroll := NewScrollState()
			list := List[string]{ID: "bench-list", State: state, ScrollState: scroll}
			root := Scrollable{State: scroll, Height: Cells(collectionScrollBenchViewport), Child: list}
			runCollectionScrollBench(b, collectionScrollBenchRenderer(list.ID, list), root, func(down bool) {
				if down {
					scroll.ScrollDown(1)
				} else {
					scroll.ScrollUp(1)
				}
			})
		})
		b.Run(fmt.Sprintf("table/rows=%d/cursor", count), func(b *testing.B) {
			rows := make([][]string, count)
			for i := range rows {
				rows[i] = []string{fmt.Sprintf("Row %05d", i), "middle", "last"}
			}
			state := NewTableState(rows)
			scroll := NewScrollState()
			table := Table[[]string]{ID: "bench-table", State: state, ScrollState: scroll,
				Columns: []TableColumn{{Width: Cells(20)}, {Width: Cells(20)}, {Width: Cells(20)}}}
			root := Scrollable{State: scroll, Height: Cells(collectionScrollBenchViewport), Child: table}
			runCollectionScrollBench(b, collectionScrollBenchRenderer(table.ID, table), root, func(down bool) {
				if down {
					table.keyCursorDown()
				} else {
					table.keyCursorUp()
				}
			})
		})
		b.Run(fmt.Sprintf("tree/nodes=%d/cursor", count), func(b *testing.B) {
			roots := make([]TreeNode[string], count)
			for i := range roots {
				roots[i] = TreeNode[string]{Data: fmt.Sprintf("Node %05d", i)}
			}
			state := NewTreeState(roots)
			scroll := NewScrollState()
			tree := Tree[string]{ID: "bench-tree", State: state, ScrollState: scroll}
			root := Scrollable{State: scroll, Height: Cells(collectionScrollBenchViewport), Child: tree}
			runCollectionScrollBench(b, collectionScrollBenchRenderer(tree.ID, tree), root, func(down bool) {
				if down {
					tree.keyCursorDown()
				} else {
					tree.keyCursorUp()
				}
			})
		})
	}
}
