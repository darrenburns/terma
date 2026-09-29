package terma

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// Raster checks catch seams which aren't visible in SVG markup comparisons.
func rasterizeSnapshot(t *testing.T, svg string, scale string) image.Image {
	t.Helper()
	tool, err := exec.LookPath("rsvg-convert")
	if err != nil {
		t.Skip("rsvg-convert is required for SVG raster checks")
	}
	path := filepath.Join(t.TempDir(), "snapshot.svg")
	require.NoError(t, os.WriteFile(path, []byte(svg), 0600))
	data, err := exec.Command(tool, "-z", scale, path).Output()
	require.NoError(t, err)
	im, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	return im
}

func TestSVGRasterTextHasVerticalCellPadding(t *testing.T) {
	for _, lineHeight := range []float64{1.4, 2} {
		buf := uv.NewBuffer(2, 1)
		buf.SetCell(0, 0, &uv.Cell{Content: "M", Width: 1})
		buf.SetCell(1, 0, &uv.Cell{Content: "g", Width: 1})
		opts := DefaultSVGOptions()
		opts.LineHeight = lineHeight
		for name, svg := range map[string]string{
			"snapshot": BufferToSVG(buf, 2, 1, opts),
			"diff":     GenerateDiffSVG(uv.NewBuffer(2, 1), buf, 2, 1, opts),
		} {
			t.Run(fmt.Sprintf("%s/%.1f", name, lineHeight), func(t *testing.T) {
				im := rasterizeSnapshot(t, svg, "4")
				top, bottom := im.Bounds().Dy(), -1
				for y := 0; y < im.Bounds().Dy(); y++ {
					for x := 0; x < im.Bounds().Dx(); x++ {
						r, g, b, _ := im.At(x, y).RGBA()
						// Glyph pixels are neutral; both the page and diff background are not.
						if r > 0x2000 && r == g && g == b {
							top = min(top, y)
							bottom = max(bottom, y)
						}
					}
				}
				require.GreaterOrEqual(t, bottom, top, "no text rendered")
				cellTop := float64(opts.Padding * 4)
				cellBottom := cellTop + float64(opts.FontSize)*opts.LineHeight*4
				topGap := float64(top) - cellTop
				bottomGap := cellBottom - float64(bottom+1)
				require.Greater(t, topGap, 4.0, "text touches the cell top")
				require.Greater(t, bottomGap, 4.0, "descender touches the cell bottom")
				require.Less(t, math.Abs(topGap-bottomGap), 12.0, "text is not vertically balanced")
			})
		}
	}
}

func TestSVGRasterLatinAfterCJKStaysInItsCell(t *testing.T) {
	for _, text := range []string{"中M", "日M", "한M", "ＡM", "A中M"} {
		t.Run(text, func(t *testing.T) {
			width := 3
			if text == "A中M" {
				width = 4
			}
			widget := Column{Children: []Widget{
				Text{Content: strings.Repeat(" ", width-1) + "M"},
				Text{Content: text},
			}}
			opts := DefaultSVGOptions()
			opts.FontSize = 20
			opts.LineHeight = 1.5 // Integral row height makes pixel comparisons exact.
			im := rasterizeSnapshot(t, SnapshotWithOptions(widget, width, 2, opts), "2")
			// The Latin M after a wide glyph must match M placed at that column
			// directly, regardless of the fallback font's natural CJK advance.
			startX := 16 + (width-1)*24
			for y := 16; y < 76; y++ {
				for x := startX; x < startX+24; x++ {
					require.Equal(t, im.At(x, y), im.At(x, y+60), "Latin glyph shifted after CJK at (%d,%d)", x, y)
				}
			}
		})
	}
}

func TestSVGRasterDiffDoesNotCoverWideGlyph(t *testing.T) {
	buf := RenderToBuffer(Text{Content: "中", Style: Style{ForegroundColor: RGB(255, 255, 255)}}, 2, 1)
	opts := DefaultSVGOptions()
	opts.FontSize = 20
	opts.LineHeight = 1.5
	opts.Background = RGB(255, 0, 255)
	want := rasterizeSnapshot(t, BufferToSVG(buf, 2, 1, opts), "2")
	got := rasterizeSnapshot(t, GenerateDiffSVG(uv.NewBuffer(2, 1), buf, 2, 1, opts), "2")
	// Compare glyph pixels in both allocated cells. Continuation-cell
	// highlighting must not paint over any part of the glyph.
	for y := 16; y < 76; y++ {
		for x := 16; x < 64; x++ {
			_, green, _, _ := want.At(x, y).RGBA()
			if green > 0 { // White glyph pixels on a magenta background.
				require.Equal(t, want.At(x, y), got.At(x, y), "wide diff glyph obscured at (%d,%d)", x, y)
			}
		}
	}
}

func TestSVGRasterBackgroundHasNoCellSeams(t *testing.T) {
	for _, scale := range []string{"1", "1.25", "2"} {
		t.Run(scale, func(t *testing.T) {
			buf := uv.NewBuffer(5, 3)
			bg := color.RGBA{R: 90, G: 70, B: 130, A: 255}
			for y := 0; y < 3; y++ {
				for x := 0; x < 5; x++ {
					buf.SetCell(x, y, &uv.Cell{Content: " ", Width: 1, Style: uv.Style{Bg: bg}})
				}
			}
			im := rasterizeSnapshot(t, BufferToSVG(buf, 5, 3, DefaultSVGOptions()), scale)
			// Stay inside the panel's outer edge; every interior pixel must be solid.
			b := im.Bounds()
			for y := b.Dy() / 3; y < b.Dy()*2/3; y++ {
				for x := b.Dx() / 3; x < b.Dx()*2/3; x++ {
					actual := color.RGBAModel.Convert(im.At(x, y))
					require.Equal(t, bg, actual, "cell seam at (%d,%d)", x, y)
				}
			}
		})
	}
}

func TestSVGRasterBoxLinesJoinBetweenRows(t *testing.T) {
	for _, char := range []string{"│", "┃", "║"} {
		for _, scale := range []string{"1", "1.25", "2"} {
			t.Run(char+"/"+scale, func(t *testing.T) {
				buf := uv.NewBuffer(1, 3)
				for y := 0; y < 3; y++ {
					buf.SetCell(0, y, &uv.Cell{Content: char, Width: 1})
				}
				im := rasterizeSnapshot(t, BufferToSVG(buf, 1, 3, DefaultSVGOptions()), scale)
				// Interior rows must have the same stroke pixels, even where cells meet.
				b := im.Bounds()
				for y := b.Dy() / 3; y < b.Dy()*2/3; y++ {
					for x := 0; x < b.Dx(); x++ {
						require.Equal(t, im.At(x, b.Dy()/2), im.At(x, y), "border seam at (%d,%d)", x, y)
					}
				}
			})
		}
	}
}

func TestSVGBoxDrawingCoversBorderStyles(t *testing.T) {
	for _, style := range []BorderStyle{BorderSquare, BorderRounded, BorderDouble, BorderHeavy, BorderDashed} {
		chars := GetBorderCharSet(style)
		for _, char := range []string{chars.TopLeft, chars.TopRight, chars.BottomLeft, chars.BottomRight, chars.Top, chars.Bottom, chars.Left, chars.Right} {
			buf := uv.NewBuffer(1, 1)
			buf.SetCell(0, 0, &uv.Cell{Content: char, Width: 1})
			svg := BufferToSVG(buf, 1, 1, DefaultSVGOptions())
			require.Contains(t, svg, "<path ", "border character %s must use cell geometry", char)
			require.NotContains(t, svg, "<text ")
		}
	}
	// Ordinary text and multi-rune graphemes must continue through text rendering.
	for _, content := range []string{"A", "│\u0301", ""} {
		_, ok := boxDrawingShape(content)
		require.False(t, ok, "unexpected box drawing shape for %q", content)
	}
}
