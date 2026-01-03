package main

import (
	"fmt"

	t "terma"
)

type ColorDemo struct{}

func (d *ColorDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		Children: []t.Widget{
			// Header
			t.Text{
				Content: " Terma Color API Demo ",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Hex("#06B6D4"),
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.Text{Content: ""},

			// Section: Constructors
			sectionTitle("Constructors"),
			t.Row{
				Children: []t.Widget{
					colorSwatch("RGB(59, 130, 246)", t.RGB(59, 130, 246)),
					colorSwatch("Hex(\"#10B981\")", t.Hex("#10B981")),
					colorSwatch("HSL(280, 0.8, 0.6)", t.HSL(280, 0.8, 0.6)),
				},
			},
			t.Text{Content: ""},

			// Section: Lightness
			sectionTitle("Lighten / Darken"),
			lightnessDemo(),
			t.Text{Content: ""},

			// Section: Saturation
			sectionTitle("Saturate / Desaturate"),
			saturationDemo(),
			t.Text{Content: ""},

			// Section: Hue Rotation
			sectionTitle("Hue Rotation"),
			hueRotationDemo(),
			t.Text{Content: ""},

			// Section: Blending
			sectionTitle("Color Blending"),
			blendingDemo(),
			t.Text{Content: ""},

			// Section: Special Operations
			sectionTitle("Invert & Complement"),
			specialOpsDemo(),
			t.Text{Content: ""},

			// Section: Contrast Ratio
			sectionTitle("WCAG Contrast Ratios"),
			contrastDemo(),
		},
	}
}

func sectionTitle(title string) t.Widget {
	return t.Text{
		Content: " " + title + " ",
		Style: t.Style{
			ForegroundColor: t.Hex("#FCD34D"),
			BackgroundColor: t.Hex("#1E293B"),
		},
	}
}

func colorSwatch(label string, color t.Color) t.Widget {
	// Choose text color based on background luminosity
	textColor := t.White
	if color.IsLight() {
		textColor = t.Black
	}

	return t.Text{
		Content: fmt.Sprintf(" %s ", label),
		Style: t.Style{
			ForegroundColor: textColor,
			BackgroundColor: color,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
	}
}

func lightnessDemo() t.Widget {
	base := t.Hex("#3B82F6") // Blue
	return t.Row{
		Children: []t.Widget{
			colorBlock(base.Darken(0.3), "-.3"),
			colorBlock(base.Darken(0.2), "-.2"),
			colorBlock(base.Darken(0.1), "-.1"),
			colorBlock(base, "base"),
			colorBlock(base.Lighten(0.1), "+.1"),
			colorBlock(base.Lighten(0.2), "+.2"),
			colorBlock(base.Lighten(0.3), "+.3"),
		},
	}
}

func saturationDemo() t.Widget {
	base := t.Hex("#EF4444") // Red
	return t.Row{
		Children: []t.Widget{
			colorBlock(base.Desaturate(0.6), "-.6"),
			colorBlock(base.Desaturate(0.4), "-.4"),
			colorBlock(base.Desaturate(0.2), "-.2"),
			colorBlock(base, "base"),
			colorBlock(base.Saturate(0.1), "+.1"),
			colorBlock(base.Saturate(0.2), "+.2"),
		},
	}
}

func hueRotationDemo() t.Widget {
	base := t.Hex("#10B981") // Emerald
	return t.Row{
		Children: []t.Widget{
			colorBlock(base, "0"),
			colorBlock(base.Rotate(30), "30"),
			colorBlock(base.Rotate(60), "60"),
			colorBlock(base.Rotate(90), "90"),
			colorBlock(base.Rotate(120), "120"),
			colorBlock(base.Rotate(150), "150"),
			colorBlock(base.Rotate(180), "180"),
			colorBlock(base.Rotate(210), "210"),
			colorBlock(base.Rotate(240), "240"),
			colorBlock(base.Rotate(270), "270"),
			colorBlock(base.Rotate(300), "300"),
			colorBlock(base.Rotate(330), "330"),
		},
	}
}

func blendingDemo() t.Widget {
	color1 := t.Hex("#EC4899") // Pink
	color2 := t.Hex("#3B82F6") // Blue

	return t.Column{
		Children: []t.Widget{
			t.Row{
				Children: []t.Widget{
					colorBlock(color1, "0%"),
					colorBlock(color1.Blend(color2, 0.2), "20%"),
					colorBlock(color1.Blend(color2, 0.4), "40%"),
					colorBlock(color1.Blend(color2, 0.5), "50%"),
					colorBlock(color1.Blend(color2, 0.6), "60%"),
					colorBlock(color1.Blend(color2, 0.8), "80%"),
					colorBlock(color2, "100%"),
				},
			},
		},
	}
}

func specialOpsDemo() t.Widget {
	colors := []t.Color{
		t.Hex("#F59E0B"), // Amber
		t.Hex("#8B5CF6"), // Violet
		t.Hex("#06B6D4"), // Cyan
	}

	var children []t.Widget
	for _, c := range colors {
		children = append(children,
			colorBlock(c, "orig"),
			colorBlock(c.Invert(), "inv"),
			colorBlock(c.Complement(), "comp"),
			t.Text{Content: " "},
		)
	}

	return t.Row{Children: children}
}

func contrastDemo() t.Widget {
	backgrounds := []t.Color{
		t.Hex("#1E293B"), // Slate 800
		t.Hex("#64748B"), // Slate 500
		t.Hex("#E2E8F0"), // Slate 200
		t.Hex("#FFFFFF"), // White
	}

	textColor := t.Black
	var children []t.Widget

	for _, bg := range backgrounds {
		ratio := textColor.ContrastRatio(bg)
		label := fmt.Sprintf(" %.1f:1 ", ratio)

		// Indicate if it passes WCAG AA (4.5:1 for normal text)
		passLabel := "FAIL"
		if ratio >= 4.5 {
			passLabel = "AA"
		}
		if ratio >= 7.0 {
			passLabel = "AAA"
		}

		children = append(children, t.Text{
			Content: label + passLabel + " ",
			Style: t.Style{
				ForegroundColor: textColor,
				BackgroundColor: bg,
			},
		})
	}

	return t.Row{Children: children}
}

func colorBlock(color t.Color, label string) t.Widget {
	textColor := t.White
	if color.IsLight() {
		textColor = t.Black
	}

	// Pad label to 4 chars for consistent width
	for len(label) < 4 {
		label = " " + label
	}

	return t.Text{
		Content: " " + label + " ",
		Style: t.Style{
			ForegroundColor: textColor,
			BackgroundColor: color,
		},
	}
}

func main() {
	t.Run(&ColorDemo{})
}
