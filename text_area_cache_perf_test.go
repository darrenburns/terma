package terma

import (
	"fmt"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// Each navigation operation moves left and right, ending at the same position.
// Paint measurements draw into a fixed 80x24 buffer without a highlighter.
func BenchmarkTextAreaGeometryCache(b *testing.B) {
	for _, lines := range []int{64, 1000} {
		text := strings.Repeat("a paragraph with words, 界 and 🙂 Unicode, wrapping across the available width.\n", lines)
		b.Run(fmt.Sprintf("lines=%d", lines), func(b *testing.B) {
			newState := func() *TextAreaState {
				state := NewTextAreaState(text)
				state.lastWidth = 80
				state.CursorIndex.Set(len(state.Content.Peek()) / 2)
				return state
			}
			b.Run("CursorLeftRight", func(b *testing.B) {
				state := newState()
				state.CursorScreenPosition(0, 0)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					state.CursorLeft()
					state.CursorRight()
				}
			})
			b.Run("CursorUpDown", func(b *testing.B) {
				state := newState()
				state.CursorScreenPosition(0, 0)
				b.ReportAllocs()
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					state.CursorUp()
					state.CursorDown()
				}
			})
			for _, changed := range []bool{false, true} {
				name := "CursorPaint"
				if changed {
					name = "ContentChangePaint"
				}
				b.Run(name, func(b *testing.B) {
					state := newState()
					widget := TextArea{ID: "benchmark-editor", State: state, Width: Cells(80), Height: Cells(24)}
					buffer := uv.NewBuffer(80, 24)
					ctx := NewRenderContext(buffer, 80, 24, nil, nil, BuildContext{}, nil)
					widget.Render(ctx)
					b.ReportAllocs()
					b.ResetTimer()
					for i := 0; i < b.N; i++ {
						if changed {
							state.Content.Update(func(content []string) []string {
								if content[0] == "a" {
									content[0] = "b"
								} else {
									content[0] = "a"
								}
								return content
							})
						}
						state.CursorIndex.Set(len(state.Content.Peek())/2 + i%2)
						widget.Render(ctx)
					}
				})
			}
		})
	}
}
