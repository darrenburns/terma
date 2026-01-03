package main

import (
	"fmt"

	t "terma"
)

type ColorDemo struct{}

func (d *ColorDemo) Build(ctx t.BuildContext) t.Widget {
	return t.GradientBox{
		Gradient: t.NewGradient(
			t.Hex("#0F172A"), // Slate 900
			t.Hex("#1E293B"), // Slate 800
		),
		Width:  t.Fr(1),
		Height: t.Fr(1),
		Child: t.Column{
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				header(),
				t.Text{Content: ""},

				// Color constructors showcase
				constructorsSection(),
				t.Text{Content: ""},

				// Lightness manipulation
				lightnessSection(),
				t.Text{Content: ""},

				// Saturation manipulation
				saturationSection(),
				t.Text{Content: ""},

				// Color harmonies
				harmoniesSection(),
				t.Text{Content: ""},

				// Blending showcase
				blendingSection(),
				t.Text{Content: ""},

				// AutoText showcase
				autoTextSection(),
				t.Text{Content: ""},

				// Contrast & accessibility
				accessibilitySection(),
				t.Text{Content: ""},

				// Alpha transparency showcase
				transparencySection(),
			},
		},
	}
}

func header() t.Widget {
	title := "TERMA COLOR API"

	// Create a subtle gradient for the title
	gradient := t.NewGradient(
		t.Hex("#10B981"), // Emerald
		t.Hex("#3B82F6"), // Blue
	)

	// Apply gradient colors to each character
	colors := gradient.Steps(len(title))
	var spans []t.Span
	for i, ch := range title {
		spans = append(spans, t.StyledSpan(string(ch), t.SpanStyle{
			Foreground: colors[i],
			Bold:       true,
		}))
	}

	return t.Column{
		Children: []t.Widget{
			t.Text{Spans: spans},
			t.Text{
				Content: "A beautiful, fluent color manipulation API",
				Style:   t.Style{ForegroundColor: t.Hex("#94A3B8")},
			},
		},
	}
}

func constructorsSection() t.Widget {
	return t.Column{
		Children: []t.Widget{
			sectionHeader("Color Constructors", t.Hex("#60A5FA")),
			t.Row{
				Children: []t.Widget{
					constructorCard("RGB()", "RGB(251, 146, 60)", t.RGB(251, 146, 60)),
					t.Text{Content: " "},
					constructorCard("Hex()", "Hex(\"#A78BFA\")", t.Hex("#A78BFA")),
					t.Text{Content: " "},
					constructorCard("HSL()", "HSL(160, 0.84, 0.39)", t.HSL(160, 0.84, 0.39)),
				},
			},
		},
	}
}

func constructorCard(name, code string, color t.Color) t.Widget {
	title := t.BorderTitle(name)
	title.Color = color

	return t.Column{
		Style: t.Style{
			Border: t.Border{
				Style:       t.BorderRounded,
				Color:       color.Darken(0.2),
				Decorations: []t.BorderDecoration{title},
			},
		},
		Children: []t.Widget{
			t.Text{
				Content: " " + code + " ",
				Style:   t.Style{ForegroundColor: t.Hex("#CBD5E1")},
			},
		},
	}
}

func lightnessSection() t.Widget {
	base := t.Hex("#8B5CF6") // Violet

	var blocks []t.Widget
	steps := []float64{-0.35, -0.25, -0.15, -0.05, 0, 0.05, 0.15, 0.25, 0.35}

	for _, step := range steps {
		var color t.Color
		var label string
		if step < 0 {
			color = base.Darken(-step)
			label = fmt.Sprintf("%.0f%%", step*100)
		} else if step > 0 {
			color = base.Lighten(step)
			label = fmt.Sprintf("+%.0f%%", step*100)
		} else {
			color = base
			label = "base"
		}
		blocks = append(blocks, gradientBlock(color, label))
	}

	return t.Column{
		Children: []t.Widget{
			sectionHeader("Lighten() & Darken()", t.Hex("#8B5CF6")),
			t.Row{Children: blocks},
		},
	}
}

func saturationSection() t.Widget {
	base := t.Hex("#F43F5E") // Rose

	var blocks []t.Widget
	steps := []float64{-0.8, -0.6, -0.4, -0.2, 0, 0.1, 0.2}

	for _, step := range steps {
		var color t.Color
		var label string
		if step < 0 {
			color = base.Desaturate(-step)
			label = fmt.Sprintf("%.0f%%", step*100)
		} else if step > 0 {
			color = base.Saturate(step)
			label = fmt.Sprintf("+%.0f%%", step*100)
		} else {
			color = base
			label = "base"
		}
		blocks = append(blocks, gradientBlock(color, label))
	}

	return t.Column{
		Children: []t.Widget{
			sectionHeader("Saturate() & Desaturate()", t.Hex("#F43F5E")),
			t.Row{Children: blocks},
		},
	}
}

func harmoniesSection() t.Widget {
	base := t.Hex("#06B6D4") // Cyan

	return t.Column{
		Children: []t.Widget{
			sectionHeader("Color Harmonies", t.Hex("#06B6D4")),

			// Complementary
			t.Row{
				Children: []t.Widget{
					harmonyLabel("Complement"),
					harmonyBlock(base, "base"),
					harmonyBlock(base.Complement(), "+180°"),
				},
			},

			// Triadic
			t.Row{
				Children: []t.Widget{
					harmonyLabel("Triadic"),
					harmonyBlock(base, "base"),
					harmonyBlock(base.Rotate(120), "+120°"),
					harmonyBlock(base.Rotate(240), "+240°"),
				},
			},

			// Analogous
			t.Row{
				Children: []t.Widget{
					harmonyLabel("Analogous"),
					harmonyBlock(base.Rotate(-30), "-30°"),
					harmonyBlock(base, "base"),
					harmonyBlock(base.Rotate(30), "+30°"),
				},
			},

			// Split-complementary
			t.Row{
				Children: []t.Widget{
					harmonyLabel("Split-Comp"),
					harmonyBlock(base, "base"),
					harmonyBlock(base.Rotate(150), "+150°"),
					harmonyBlock(base.Rotate(210), "+210°"),
				},
			},
		},
	}
}

func harmonyLabel(name string) t.Widget {
	// Pad to consistent width
	for len(name) < 12 {
		name = name + " "
	}
	return t.Text{
		Content: name,
		Style:   t.Style{ForegroundColor: t.Hex("#94A3B8")},
	}
}

func harmonyBlock(color t.Color, label string) t.Widget {
	// Use AutoText() for readable text that preserves color character
	textColor := color.AutoText()

	// Pad label
	for len(label) < 5 {
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

func blendingSection() t.Widget {
	// Multiple gradient examples using the Gradient API
	gradients := []struct {
		gradient t.Gradient
		name     string
	}{
		{t.NewGradient(t.Hex("#EC4899"), t.Hex("#8B5CF6")), "Pink → Violet"},
		{t.NewGradient(t.Hex("#F59E0B"), t.Hex("#EF4444")), "Amber → Red"},
		{t.NewGradient(t.Hex("#10B981"), t.Hex("#3B82F6")), "Emerald → Blue"},
		{t.NewGradient(t.Hex("#EC4899"), t.Hex("#F59E0B"), t.Hex("#22C55E")), "Pink → Amber → Green"},
	}

	var rows []t.Widget
	rows = append(rows, sectionHeader("Gradient API", t.Hex("#F472B6")))

	for _, g := range gradients {
		var blocks []t.Widget
		blocks = append(blocks, blendLabel(g.name))

		// Use Steps() to get evenly distributed colors
		for _, color := range g.gradient.Steps(11) {
			blocks = append(blocks, t.Text{
				Content: "██",
				Style:   t.Style{ForegroundColor: color},
			})
		}
		rows = append(rows, t.Row{Children: blocks})
	}

	return t.Column{Children: rows}
}

func blendLabel(name string) t.Widget {
	for len(name) < 22 {
		name = name + " "
	}
	return t.Text{
		Content: name,
		Style:   t.Style{ForegroundColor: t.Hex("#94A3B8")},
	}
}

func autoTextSection() t.Widget {
	// Show a variety of colors with AutoText
	colors := []struct {
		color t.Color
		name  string
	}{
		{t.Hex("#1E293B"), "Slate 800"},
		{t.Hex("#7C3AED"), "Violet"},
		{t.Hex("#059669"), "Emerald"},
		{t.Hex("#DC2626"), "Red"},
		{t.Hex("#F59E0B"), "Amber"},
		{t.Hex("#06B6D4"), "Cyan"},
		{t.Hex("#EC4899"), "Pink"},
		{t.Hex("#E2E8F0"), "Slate 200"},
	}

	var blocks []t.Widget
	for _, c := range colors {
		// Pad name
		name := c.name
		for len(name) < 10 {
			name = name + " "
		}

		blocks = append(blocks, t.Text{
			Content: " " + name + " ",
			Style: t.Style{
				ForegroundColor: c.color.AutoText(),
				BackgroundColor: c.color,
			},
		})
	}

	return t.Column{
		Children: []t.Widget{
			sectionHeader("AutoText() - Always Readable", t.Hex("#FBBF24")),
			t.Row{Children: blocks},
		},
	}
}

func accessibilitySection() t.Widget {
	backgrounds := []t.Color{
		t.Hex("#0F172A"), // Slate 900
		t.Hex("#334155"), // Slate 700
		t.Hex("#64748B"), // Slate 500
		t.Hex("#94A3B8"), // Slate 400
		t.Hex("#CBD5E1"), // Slate 300
		t.Hex("#F1F5F9"), // Slate 100
	}

	var blocks []t.Widget
	for _, bg := range backgrounds {
		// Use AutoText to get readable text
		textColor := bg.AutoText()
		ratio := textColor.ContrastRatio(bg)

		// WCAG level indicator
		level := "FAIL"
		if ratio >= 7.0 {
			level = "AAA"
		} else if ratio >= 4.5 {
			level = "AA"
		}

		blocks = append(blocks, t.Text{
			Content: fmt.Sprintf(" %4.1f:1 %s ", ratio, level),
			Style: t.Style{
				ForegroundColor: textColor,
				BackgroundColor: bg,
			},
		})
	}

	return t.Column{
		Children: []t.Widget{
			sectionHeader("ContrastRatio() + AutoText()", t.Hex("#22C55E")),
			t.Row{Children: blocks},
		},
	}
}

func sectionHeader(title string, accentColor t.Color) t.Widget {
	return t.Text{
		Spans: []t.Span{
			t.StyledSpan("▌", t.SpanStyle{Foreground: accentColor}),
			t.StyledSpan(" "+title, t.SpanStyle{Foreground: t.White, Bold: true}),
		},
	}
}

func gradientBlock(color t.Color, label string) t.Widget {
	// Pad label to consistent width
	for len(label) < 5 {
		label = " " + label
	}

	return t.Text{
		Content: label + " ",
		Style: t.Style{
			ForegroundColor: color.AutoText(),
			BackgroundColor: color,
		},
	}
}

func transparencySection() t.Widget {
	pink := t.Hex("#EC4899")

	return t.Column{
		Children: []t.Widget{
			sectionHeader("Alpha Transparency", t.Hex("#A78BFA")),

			// Single layer transparency
			t.Row{
				Children: []t.Widget{
					alphaLabel("Single Layer"),
					alphaBlock(pink.WithAlpha(1.0), "100%"),
					alphaBlock(pink.WithAlpha(0.75), "75%"),
					alphaBlock(pink.WithAlpha(0.5), "50%"),
					alphaBlock(pink.WithAlpha(0.25), "25%"),
					alphaBlock(pink.WithAlpha(0.1), "10%"),
				},
			},
		},
	}
}

func alphaLabel(name string) t.Widget {
	for len(name) < 14 {
		name = name + " "
	}
	return t.Text{
		Content: name,
		Style:   t.Style{ForegroundColor: t.Hex("#94A3B8")},
	}
}

func alphaBlock(color t.Color, label string) t.Widget {
	// Pad label
	for len(label) < 5 {
		label = " " + label
	}

	return t.Text{
		Content: " " + label + " ",
		Style: t.Style{
			ForegroundColor: t.White,
			BackgroundColor: color,
		},
	}
}

func main() {
	t.Run(&ColorDemo{})
}
