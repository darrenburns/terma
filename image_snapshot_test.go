package terma

import "testing"

func TestSnapshot_ImageFits(t *testing.T) {
	source := testImage(t, 48, 48)
	widgets := []Widget{}
	for _, fit := range []ImageFit{ImageContain, ImageCover, ImageStretch} {
		widgets = append(widgets, Image{Source: source, Fit: fit, Style: Style{Width: Cells(12), Height: Cells(4), Padding: EdgeInsetsAll(1), Border: RoundedBorder(Blue), BackgroundColor: RGB(20, 30, 50)}})
	}
	AssertSnapshot(t, Column{Spacing: 1, Children: widgets}, 18, 26, "Static image logical previews: contain, cover and stretch, with alpha flattened over the effective background")
}
