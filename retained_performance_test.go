package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/darrenburns/terma/layout"
)

// Alternate flex rows and columns as in an application shell with a header,
// sidebar, and nested detail panes. Each level has one stable label sibling.
func retainedPerformancePanels(leaf Widget, depth int) Widget {
	for i := 0; i < depth; i++ {
		children := []Widget{Text{Content: "panel"}, leaf}
		style := Style{Width: Flex(1), Height: Flex(1)}
		if i%2 == 0 {
			leaf = Row{Children: children, Style: style, CrossAlign: CrossAxisStretch}
		} else {
			leaf = Column{Children: children, Style: style, CrossAlign: CrossAxisStretch}
		}
	}
	return leaf
}

func BenchmarkRetainedNestedLayout(b *testing.B) {
	for _, depth := range []int{3, 6} {
		b.Run(fmt.Sprintf("depth=%d", depth), func(b *testing.B) {
			value := NewSignal("aaaa")
			root := retainedPerformancePanels(reactivityBuildText{value: value}, depth)
			values := [2]string{"bbbb", "aaaa"}
			benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), root, 1, func(i int) {
				value.Set(values[i%2])
			})
		})
	}
}

type retainedPerformanceLayoutLeaf struct {
	width Signal[int]
	calls map[layout.Constraints]int
}

func (w retainedPerformanceLayoutLeaf) Build(BuildContext) Widget { return w }

func (w retainedPerformanceLayoutLeaf) BuildLayoutNode(BuildContext) layout.LayoutNode { return w }

func (w retainedPerformanceLayoutLeaf) ComputeLayout(c layout.Constraints) layout.ComputedLayout {
	w.calls[c]++
	node := layout.BoxNode{Width: w.width.Get(), Height: 1}
	return node.ComputeLayout(c)
}

func TestRetainedLayoutMeasuresEachConstraintOncePerFrame(t *testing.T) {
	width := NewSignal(3)
	calls := make(map[layout.Constraints]int)
	root := retainedPerformancePanels(retainedPerformanceLayoutLeaf{width: width, calls: calls}, 3)
	renderer := newTestRenderer(uv.NewBuffer(120, 40), 120, 40)
	t.Cleanup(func() { renderer.rootNode.dispose() })
	for _, frame := range []string{"initial", "layout change", "forced"} {
		clear(calls)
		switch frame {
		case "initial":
			renderer.Update(root)
		case "layout change":
			width.Set(7)
			renderer.Update(root)
		case "forced":
			renderer.Render(root)
		}
		if len(calls) == 0 {
			t.Fatalf("%s: expected a fresh measurement", frame)
		}
		leaf := renderer.rootNode
		for len(leaf.children) != 0 {
			leaf = leaf.children[1]
		}
		if got := leaf.layout.Box.Width; got != width.Peek() {
			t.Errorf("%s: leaf width = %d, want %d", frame, got, width.Peek())
		}
		for constraint, count := range calls {
			if count != 1 {
				t.Errorf("%s: constraint %+v computed %d times", frame, constraint, count)
			}
		}
	}
}

type retainedPerformancePaintParent struct {
	Column
	value Signal[string]
}

func (w retainedPerformancePaintParent) Build(BuildContext) Widget { return w }
func (w retainedPerformancePaintParent) ChildWidgets() []Widget    { return w.Children }
func (w retainedPerformancePaintParent) Render(ctx *RenderContext) {
	ctx.DrawText(0, 0, w.value.Get())
}

func retainedPerformanceOverlappingPaint(value Signal[string]) Widget {
	return retainedPerformancePaintParent{
		Column: Column{Children: []Widget{reactivityBenchText(value)}, Style: reactivityBenchStyle()},
		value:  value,
	}
}

func BenchmarkRetainedPartialDamage(b *testing.B) {
	b.Run("overlap", func(b *testing.B) {
		value := NewSignal("aaaa")
		root := retainedPerformanceOverlappingPaint(value)
		values := [2]string{"bbbb", "aaaa"}
		benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), root, 1, func(i int) {
			value.Set(values[i%2])
		})
	})
	for _, distribution := range []string{"contiguous", "sparse"} {
		for _, count := range []int{2, 10, 100} {
			b.Run(fmt.Sprintf("%s/leaves=%d", distribution, count), func(b *testing.B) {
				leaves := reactivityBenchLeaves(400)
				values := make([]Signal[string], count)
				for i := range values {
					values[i] = NewSignal("aaaa")
					index := i
					if distribution == "sparse" {
						index = i * (len(leaves) - 1) / (count - 1)
					}
					leaves[index] = reactivityBenchText(values[i])
				}
				states := [2]string{"bbbb", "aaaa"}
				benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), reactivityBenchTree(leaves), count, func(i int) {
					for _, value := range values {
						value.Set(states[i%2])
					}
				})
			})
		}
	}
}

func TestRetainedPartialDamageDoesNotRepaintOverlaps(t *testing.T) {
	value := NewSignal("aaaa")
	root := retainedPerformanceOverlappingPaint(value)
	renderer := newReactivityBenchRenderer()
	t.Cleanup(func() { renderer.rootNode.dispose() })
	renderer.Update(root)
	value.Set("bbbb")
	renderer.Update(root)
	if got := renderer.lastPaintCount; got != 2 {
		t.Errorf("parent and child must each paint once, got %d paint visits", got)
	}
	if got := len(renderer.lastDamagedRects); got != 1 {
		t.Errorf("overlapping parent/child damage must merge, got %d rectangles", got)
	}
	incremental := renderedCells(renderer)
	renderer.Render(root)
	if incremental != renderedCells(renderer) {
		t.Fatal("incremental output differs from a forced full render")
	}
}

func TestRetainedPartialDamageKeepsContentAlongsideScrollbar(t *testing.T) {
	value := NewSignal("aaaa")
	state := NewScrollState()
	root := Scrollable{
		State: state,
		Style: Style{Width: Cells(20), Height: Cells(3)},
		Child: Column{Children: []Widget{
			reactivityBenchText(value),
			Text{Content: "one\ntwo\nthree\nfour\nfive"},
		}},
	}
	renderer := newReactivityBenchRenderer()
	t.Cleanup(func() { renderer.rootNode.dispose() })
	renderer.Update(root)
	state.setPosition(0.4)
	value.Set("bbbb")
	renderer.Update(root)
	if renderer.lastFrameMode != rendererFramePartial {
		t.Fatalf("expected partial repaint, got %s", renderer.lastFrameMode)
	}
	if len(renderer.lastDamagedRects) != 2 {
		t.Errorf("expected separate scrollbar and content damage, got %v", renderer.lastDamagedRects)
	}
	incremental := renderedCells(renderer)
	renderer.Render(root)
	if incremental != renderedCells(renderer) {
		t.Fatal("incremental output differs from a forced full render")
	}
}
