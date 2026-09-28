package layout

import (
	"fmt"
	"strings"
	"testing"
)

var textPerfWidth, textPerfHeight int

func BenchmarkTextPerfMeasureHardWrap(b *testing.B) {
	for _, size := range []int{80, 4096, 16384} {
		b.Run(fmt.Sprintf("bytes=%d", size), func(b *testing.B) {
			content := strings.Repeat("x", size)
			b.ReportAllocs()
			b.SetBytes(int64(size))
			b.ResetTimer()
			for b.Loop() {
				textPerfWidth, textPerfHeight = MeasureText(content, WrapChar, 80)
			}
		})
	}
}
