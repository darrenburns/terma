package terma

import (
	"fmt"
	"strings"
	"testing"
)

var textPerfLines []string
var textPerfSpanLines []lineData

func BenchmarkTextPerfHardWrap(b *testing.B) {
	for _, size := range []int{80, 4096, 16384} {
		b.Run(fmt.Sprintf("bytes=%d", size), func(b *testing.B) {
			content := strings.Repeat("x", size)
			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()
			for b.Loop() {
				textPerfLines = wrapText(content, 80, WrapHard)
			}
		})
	}
}

func BenchmarkTextPerfCollectSpans(b *testing.B) {
	for _, scenario := range []struct {
		name          string
		content       string
		width, height int
		wrap          WrapMode
	}{
		{"StatusRow", strings.Repeat("status: OK  ", 7), 80, 1, WrapNone},
		{"Paragraph", strings.Repeat("The quick brown fox jumps over the lazy dog. ", 48), 80, 24, WrapSoft},
		{"LongWord", strings.Repeat("x", 4096), 80, 24, WrapSoft},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			widget := Text{Spans: []Span{{Text: scenario.content, Style: SpanStyle{Bold: true}}}, Wrap: scenario.wrap}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				textPerfSpanLines = widget.collectSpanLines(scenario.width, scenario.height)
			}
		})
	}
}

// These scenarios include signal notification and a real renderer update, so
// collection improvements can be distinguished from total frame cost.
func BenchmarkTextPerfReactiveSpans(b *testing.B) {
	for _, scenario := range []struct {
		name, content string
		height        int
		wrap          WrapMode
	}{
		{"StatusRow", strings.Repeat("status: OK  ", 7), 1, WrapNone},
		{"Paragraph", strings.Repeat("The quick brown fox jumps over the lazy dog. ", 48), 24, WrapSoft},
	} {
		b.Run(scenario.name, func(b *testing.B) {
			value := NewSignal(false)
			widget := ComputedSpans("A"+scenario.content, func(ThemeData) []Span {
				prefix := "A"
				if value.Get() {
					prefix = "B"
				}
				return []Span{{Text: prefix + scenario.content, Style: SpanStyle{Bold: true}}}
			})
			widget.Wrap = scenario.wrap
			widget.LayoutStyle = Style{Width: Cells(80), Height: Cells(scenario.height)}
			benchmarkReactiveUpdates(b, newReactivityBenchRenderer(), widget, 1, func(i int) {
				value.Set(i%2 == 0)
			})
		})
	}
}
