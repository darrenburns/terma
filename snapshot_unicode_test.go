package terma

import (
	"encoding/xml"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type svgTextElement struct {
	X       float64 `xml:"x,attr"`
	Anchor  string  `xml:"text-anchor,attr"`
	Class   string  `xml:"class,attr"`
	Content string  `xml:",chardata"`
}

func TestSVGDiffWideGraphemeUsesReservedCells(t *testing.T) {
	buf := RenderToBuffer(Text{Content: "中M"}, 3, 1)
	empty := uv.NewBuffer(3, 1)
	for name, svg := range map[string]string{
		"actual":   GenerateDiffSVG(empty, buf, 3, 1, DefaultSVGOptions()),
		"expected": GenerateDiffSVG(buf, uv.NewBuffer(0, 1), 3, 1, DefaultSVGOptions()),
	} {
		t.Run(name, func(t *testing.T) {
			var doc struct {
				Text []svgTextElement `xml:"text"`
			}
			require.NoError(t, xml.Unmarshal([]byte(svg), &doc))
			require.Equal(t, []svgTextElement{
				{X: 16.4, Anchor: "middle", Content: "中"},
				{X: 24.8, Content: "M"},
			}, doc.Text)
		})
	}
}

func TestSVGWideGraphemesReserveTerminalColumns(t *testing.T) {
	for _, wide := range []string{"中", "日", "한", "Ａ", "中\u0301", "👨‍👩‍👧‍👦"} {
		for _, reverse := range []bool{false, true} {
			t.Run(wide+map[bool]string{false: "/normal", true: "/reverse"}[reverse], func(t *testing.T) {
				widget := Text{Content: "A" + wide + "BC", Style: Style{Bold: true, Reverse: reverse}}
				var doc struct {
					Text []svgTextElement `xml:"text"`
				}
				require.NoError(t, xml.Unmarshal([]byte(Snapshot(widget, 5, 1)), &doc))
				require.Len(t, doc.Text, 3, "wide grapheme must not share a natural-width run with Latin text")
				require.Equal(t, svgTextElement{X: 8, Class: "bold", Content: "A"}, doc.Text[0])
				require.Equal(t, svgTextElement{X: 24.8, Anchor: "middle", Class: "bold", Content: wide}, doc.Text[1])
				require.Equal(t, svgTextElement{X: 33.2, Class: "bold", Content: "BC"}, doc.Text[2], "Latin text must start at column three")
			})
		}
	}
}
