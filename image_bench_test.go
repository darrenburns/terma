package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

func BenchmarkImagePresentation(b *testing.B) {
	for _, protocol := range []string{"copy", "kitty", "sixel"} {
		b.Run(protocol, func(b *testing.B) {
			buffer := newImageBuffer(100, 30)
			ctx := NewRenderContext(buffer, 100, 30, nil, nil, BuildContext{}, nil)
			ctx.DrawImage(0, 0, 40, 15, testImage(b, 320, 240), ImageStretch)
			out := uv.NewBuffer(100, 30)
			copyImageCells(out, buffer, 100, 30)
			kitty := newKittyImages()
			sixel := newSixelImages()
			kitty.paint(out, buffer, 8, 16)
			for _, u := range kitty.uploads {
				u.ready = true
				for _, p := range u.placements {
					p.ready = true
				}
			}
			copyImageCells(out, buffer, 100, 30)
			sixel.output(sixel.draws(buffer, out, 8, 16, 0))
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				copyImageCells(out, buffer, 100, 30)
				switch protocol {
				case "kitty":
					kitty.paint(out, buffer, 8, 16)
				case "sixel":
					sixel.output(sixel.draws(buffer, out, 8, 16, 0))
				}
			}
			b.StopTimer()
			b.ReportMetric(float64(sixel.preparations), "preparations")
			b.ReportMetric(float64(len(kitty.uploads)), "cached-uploads")
			b.ReportMetric(float64(sixel.bytes+kitty.bytes), "cache-bytes")
		})
	}
}
