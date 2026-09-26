package terma

import (
	"fmt"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// Keep the viewport constant as the retained tree grows. Leaves after the first
// 400 are below the viewport; the changing leaves are always visible.
const (
	reactivityBenchWidth  = 120
	reactivityBenchHeight = 40
)

type reactivityBuildText struct {
	value Signal[string]
}

func (w reactivityBuildText) Build(BuildContext) Widget {
	return Text{Content: w.value.Get(), Style: reactivityBenchStyle()}
}

func reactivityBenchStyle() Style {
	return Style{Width: Cells(12), Height: Cells(1)}
}

func reactivityBenchText(value Signal[string]) PresentedText {
	text := SignalText(value, func(value string) string { return value })
	text.LayoutStyle = reactivityBenchStyle()
	return text
}

func reactivityBenchTree(leaves []Widget) Widget {
	rows := make([]Widget, 0, (len(leaves)+9)/10)
	for start := 0; start < len(leaves); start += 10 {
		rows = append(rows, Row{
			Children: leaves[start:min(start+10, len(leaves))],
			Style:    Style{Width: Cells(reactivityBenchWidth), Height: Cells(1)},
		})
	}
	return Column{Children: rows}
}

func reactivityBenchLeaves(count int) []Widget {
	leaves := make([]Widget, count)
	for i := range leaves {
		leaves[i] = Text{Content: "steady", Style: reactivityBenchStyle()}
	}
	return leaves
}

func newReactivityBenchRenderer() *Renderer {
	return NewRenderer(
		uv.NewBuffer(reactivityBenchWidth, reactivityBenchHeight),
		reactivityBenchWidth, reactivityBenchHeight,
		NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil),
	)
}

// benchmarkReactiveUpdates measures mutation and Update together, including
// listener notification and intrinsic-size checks in Set. Setup, initial render,
// validation, metrics collection, and disposal are outside the timed loop.
//
// Scenarios alternate between two visible states. Sample both directions before
// timing to report representative work without instrumenting every timed frame.
func benchmarkReactiveUpdates(b *testing.B, renderer *Renderer, root Widget, sets int, mutate func(int)) {
	b.Helper()
	b.StopTimer()
	b.Cleanup(func() {
		if renderer.rootNode != nil {
			renderer.rootNode.dispose()
		}
		for _, floating := range renderer.retainedFloats {
			if floating.root != nil {
				floating.root.dispose()
			}
		}
	})
	renderer.Update(root)
	var builds, layouts, paints, damageCells, full, partial int
	for i := 0; i < 2; i++ {
		before := renderer.ScreenText()
		beforeFull, beforePartial := renderer.fullRenderCount, renderer.partialRenderCount
		mutate(i)
		renderer.Update(root)
		fullDelta := renderer.fullRenderCount - beforeFull
		partialDelta := renderer.partialRenderCount - beforePartial
		if fullDelta+partialDelta != 1 {
			b.Fatal("scenario must produce exactly one frame per update")
		}
		if renderer.ScreenText() == before {
			b.Fatal("scenario must change visible output on every update")
		}
		builds += renderer.lastBuildCount
		layouts += renderer.lastLayoutCount
		paints += renderer.lastPaintCount
		full += fullDelta
		partial += partialDelta
		if renderer.lastFrameMode == rendererFrameFull {
			damageCells += reactivityBenchWidth * reactivityBenchHeight
		} else {
			viewport := Rect{Width: reactivityBenchWidth, Height: reactivityBenchHeight}
			for _, rect := range renderer.lastDamagedRects {
				clipped := rect.Intersect(viewport)
				damageCells += clipped.Width * clipped.Height
			}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mutate(i)
		renderer.Update(root)
	}
	b.StopTimer()
	b.ReportMetric(float64(builds)/2, "builds/op")
	b.ReportMetric(float64(layouts)/2, "layouts/op")
	b.ReportMetric(float64(paints)/2, "paints/op")
	b.ReportMetric(float64(damageCells)/2, "damage-cells/op")
	b.ReportMetric(float64(full)/2, "full/op")
	b.ReportMetric(float64(partial)/2, "partial/op")
	b.ReportMetric(float64(sets), "sets/op")
}

func BenchmarkReactivityLeaf(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		for _, phase := range []string{"build", "paint"} {
			b.Run(fmt.Sprintf("leaves=%d/%s", count, phase), func(b *testing.B) {
				value := NewSignal("aaaa")
				leaves := reactivityBenchLeaves(count)
				if phase == "build" {
					leaves[0] = reactivityBuildText{value: value}
				} else {
					leaves[0] = reactivityBenchText(value)
				}
				values := [2]string{"bbbb", "aaaa"}
				benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), reactivityBenchTree(leaves), 1, func(i int) {
					value.Set(values[i%2])
				})
			})
		}
	}
}

func BenchmarkReactivityTextSize(b *testing.B) {
	for _, sizing := range []string{"fixed", "auto-stable", "auto-changing"} {
		b.Run(sizing, func(b *testing.B) {
			value := NewSignal("aaaa")
			text := reactivityBenchText(value)
			values := [2]string{"bbbb", "aaaa"}
			if sizing != "fixed" {
				text.LayoutStyle.Width = Auto
			}
			if sizing != "auto-stable" {
				values[0] = "longer text"
			}
			leaves := reactivityBenchLeaves(100)
			leaves[0] = text
			benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), reactivityBenchTree(leaves), 1, func(i int) {
				value.Set(values[i%2])
			})
		})
	}
}

func BenchmarkReactivityListCursor(b *testing.B) {
	for _, count := range []int{10, 100, 1000} {
		for _, rendering := range []string{"default", "custom"} {
			b.Run(fmt.Sprintf("items=%d/%s", count, rendering), func(b *testing.B) {
				items := make([]string, count)
				for i := range items {
					items[i] = fmt.Sprintf("Item %04d", i)
				}
				state := NewListState(items)
				list := List[string]{
					ID: "bench-list", State: state,
					Style:       Style{Width: Cells(reactivityBenchWidth)},
					CursorStyle: CursorStyle{CursorPrefix: "> "},
				}
				if rendering == "custom" {
					list.RenderItem = func(item string, active, selected bool) Widget {
						prefix := "  "
						if active {
							prefix = "> "
						}
						return Text{Content: prefix + item, Style: Style{Width: Flex(1), Height: Cells(1)}}
					}
				}
				renderer := newReactivityBenchRenderer()
				renderer.focusManager.focusedID = list.ID
				renderer.focusedSignal.Set(list)
				benchmarkReactiveUpdates(b, renderer, list, 1, func(i int) {
					if i%2 == 0 {
						state.SelectNext()
					} else {
						state.SelectPrevious()
					}
				})
			})
		}
	}
}

func BenchmarkReactivityBatch(b *testing.B) {
	for _, target := range []string{"same-leaf", "distinct-leaves"} {
		for _, sets := range []int{1, 10, 100} {
			b.Run(fmt.Sprintf("%s/sets=%d", target, sets), func(b *testing.B) {
				signals := make([]Signal[string], 1000)
				leaves := make([]Widget, len(signals))
				for i := range signals {
					signals[i] = NewSignal("aaaa")
					leaves[i] = reactivityBenchText(signals[i])
				}
				values := [2]string{"bbbb", "aaaa"}
				intermediate := [2]string{"xxxx", "yyyy"}
				benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), reactivityBenchTree(leaves), sets, func(i int) {
					if target == "same-leaf" {
						for j := 0; j < sets-1; j++ {
							signals[0].Set(intermediate[j%2])
						}
						signals[0].Set(values[i%2])
						return
					}
					for j := 0; j < sets; j++ {
						signals[j].Set(values[i%2])
					}
				})
			})
		}
	}
}
