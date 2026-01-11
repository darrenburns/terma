package main

import (
	"fmt"

	t "terma"
)

type GradientDemo struct {
	scrollState *t.ScrollState
}

func (d *GradientDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Scrollable{
		ID:     "gradient-demo-scroll",
		State:  d.scrollState,
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: t.NewGradient(
				t.Hex("#1e1b4b"), // Indigo 950
				t.Hex("#0f172a"), // Slate 900
			).WithAngle(90),
		},
		ScrollbarThumbColor: t.Hex("#6366f1"),
		ScrollbarTrackColor: t.Hex("#1e1b4b"),
		Child: t.Column{
			Spacing: 2,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				header(),
				angleSection(),
				multiColorSection(),
				transparencySection(),
				useCasesSection(),
			},
		},
	}
}

func header() t.Widget {
	return t.Column{
		Children: []t.Widget{
			// Gradient text using ForegroundColor as a gradient
			t.Text{
				Content: "GRADIENT BACKGROUNDS",
				Style: t.Style{
					ForegroundColor: t.NewGradient(
						t.Hex("#ff0080"), // Hot pink
						t.Hex("#00ffff"), // Cyan
					).WithAngle(90),
					Bold: true,
				},
				Wrap: t.WrapNone,
			},
			t.Text{
				Content: "Arbitrary-angle gradients with transparency support",
				Style:   t.Style{ForegroundColor: t.Hex("#94a3b8")},
				Wrap:    t.WrapNone,
			},
		},
	}
}

func sectionHeader(title string, color t.Color) t.Widget {
	return t.Text{
		Spans: []t.Span{
			t.StyledSpan("▌", t.SpanStyle{Foreground: color}),
			t.StyledSpan(" "+title, t.SpanStyle{Foreground: t.White, Bold: true}),
		},
		Wrap: t.WrapNone,
	}
}

func angleSection() t.Widget {
	angles := []struct {
		angle float64
		label string
	}{
		{0, "0° (vertical)"},
		{45, "45° (diagonal)"},
		{90, "90° (horizontal)"},
		{135, "135°"},
		{180, "180°"},
		{270, "270°"},
	}

	var boxes []t.Widget
	for _, a := range angles {
		boxes = append(boxes, gradientAngleBox(a.angle, a.label))
	}

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			sectionHeader("Gradient Angles", t.Hex("#f472b6")),
			t.Row{
				Spacing:  1,
				Children: boxes,
			},
		},
	}
}

func gradientAngleBox(angle float64, label string) t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Column{
				Width:     t.Cells(12),
				Height:    t.Cells(5),
				MainAlign: t.MainAxisCenter,
				Style: t.Style{
					BackgroundColor: t.NewGradient(
						t.Hex("#7c3aed"), // Violet
						t.Hex("#1e1b4b"), // Indigo 950
					).WithAngle(angle),
				},
				Children: []t.Widget{
					t.Text{
						Content: fmt.Sprintf("%.0f°", angle),
						Style:   t.Style{ForegroundColor: t.White},
						Wrap:    t.WrapNone,
					},
				},
			},
			t.Text{
				Content: label,
				Style:   t.Style{ForegroundColor: t.Hex("#a5b4fc")},
				Wrap:    t.WrapNone,
			},
		},
	}
}

func multiColorSection() t.Widget {
	gradients := []struct {
		colors []t.Color
		name   string
	}{
		{
			colors: []t.Color{t.Hex("#ef4444"), t.Hex("#f97316"), t.Hex("#eab308")},
			name:   "Sunset",
		},
		{
			colors: []t.Color{t.Hex("#06b6d4"), t.Hex("#3b82f6"), t.Hex("#8b5cf6")},
			name:   "Ocean",
		},
		{
			colors: []t.Color{t.Hex("#22c55e"), t.Hex("#14b8a6"), t.Hex("#0ea5e9")},
			name:   "Forest",
		},
		{
			colors: []t.Color{t.Hex("#ec4899"), t.Hex("#a855f7"), t.Hex("#6366f1")},
			name:   "Neon",
		},
		{
			colors: []t.Color{t.Hex("#f43f5e"), t.Hex("#fb7185"), t.Hex("#fda4af"), t.Hex("#fecdd3")},
			name:   "Rose Fade",
		},
	}

	var boxes []t.Widget
	for _, g := range gradients {
		boxes = append(boxes, multiColorBox(g.colors, g.name))
	}

	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			sectionHeader("Multi-Color Gradients", t.Hex("#c084fc")),
			t.Row{
				Spacing:  1,
				Children: boxes,
			},
		},
	}
}

func multiColorBox(colors []t.Color, name string) t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Column{
				Width:     t.Cells(14),
				Height:    t.Cells(3),
				MainAlign: t.MainAxisCenter,
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: t.NewGradient(colors...).WithAngle(90),
				},
				Children: []t.Widget{
					t.Text{
						Content: name,
						Style:   t.Style{ForegroundColor: colors[0].AutoText()},
						Wrap:    t.WrapNone,
					},
				},
			},
		},
	}
}

func transparencySection() t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			sectionHeader("Transparency Blending", t.Hex("#60a5fa")),
			t.Text{
				Content: "Semi-transparent elements blend smoothly over gradients:",
				Style:   t.Style{ForegroundColor: t.Hex("#94a3b8")},
				Wrap:    t.WrapNone,
			},
			t.Column{
				Width:     t.Flex(1),
				Height:    t.Cells(7),
				Spacing:   1,
				MainAlign: t.MainAxisCenter,
				Style: t.Style{
					BackgroundColor: t.NewGradient(
						t.Hex("#7c3aed"),
						t.Hex("#2563eb"),
						t.Hex("#0891b2"),
					).WithAngle(90),
				},
				Children: []t.Widget{
					t.Text{
						Content: " Opaque text on gradient ",
						Style:   t.Style{ForegroundColor: t.White},
						Wrap:    t.WrapNone,
					},
					t.Text{
						Content: " Red with 50% transparent background ",
						Style: t.Style{
							ForegroundColor: t.White,
							BackgroundColor: t.Red.WithAlpha(0.5),
						},
						Wrap: t.WrapNone,
					},
					t.Text{
						Content: " Red with 25% transparent background ",
						Style: t.Style{
							ForegroundColor: t.White,
							BackgroundColor: t.Red.WithAlpha(0.25),
						},
						Wrap: t.WrapNone,
					},
				},
			},
		},
	}
}

func useCasesSection() t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			sectionHeader("Use Cases", t.Hex("#34d399")),
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					cardExample(),
					buttonExample(),
					headerExample(),
				},
			},
		},
	}
}

func cardExample() t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Column{
				Width:     t.Cells(20),
				Height:    t.Cells(8),
				MainAlign: t.MainAxisCenter,
				Style: t.Style{
					BackgroundColor: t.NewGradient(
						t.Hex("#1e293b"),
						t.Hex("#0f172a"),
					).WithAngle(90),
					Padding: t.EdgeInsetsXY(1, 0),
				},
				Children: []t.Widget{
					t.Text{
						Content: "A subtle gradient adds depth to cards",
						Style:   t.Style{ForegroundColor: t.White},
						Wrap:    t.WrapSoft,
					},
				},
			},
			t.Text{
				Content: "Subtle card gradient",
				Style:   t.Style{ForegroundColor: t.Hex("#64748b")},
				Wrap:    t.WrapNone,
			},
		},
	}
}

func buttonExample() t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Column{
				Spacing: 1,
				Width:   t.Cells(20),
				Height:  t.Cells(8),
				Children: []t.Widget{
					t.Row{
						Style: t.Style{
							BackgroundColor: t.NewGradient(
								t.Hex("#8b5cf6"),
								t.Hex("#6366f1"),
							).WithAngle(90),
							Padding: t.EdgeInsetsXY(2, 1),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Primary Button",
								Style:   t.Style{ForegroundColor: t.Black.WithAlpha(0.63)},
								Wrap:    t.WrapNone,
							},
						},
					},
					t.Row{
						Style: t.Style{
							BackgroundColor: t.NewGradient(
								t.Hex("#f43f5e"),
								t.Hex("#e11d48"),
							).WithAngle(90),
							Padding: t.EdgeInsetsXY(2, 0),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Danger Button",
								Style:   t.Style{ForegroundColor: t.Hex("#f43f5e").AutoText()},
								Wrap:    t.WrapNone,
							},
						},
					},
					t.Row{
						Style: t.Style{
							BackgroundColor: t.NewGradient(
								t.Hex("#22c55e"),
								t.Hex("#16a34a"),
							).WithAngle(90),
							Padding: t.EdgeInsetsXY(2, 0),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Success Button",
								Style:   t.Style{ForegroundColor: t.Hex("#22c55e").AutoText()},
								Wrap:    t.WrapNone,
							},
						},
					},
				},
			},
			t.Text{
				Content: "Gradient buttons",
				Style:   t.Style{ForegroundColor: t.Hex("#64748b")},
				Wrap:    t.WrapNone,
			},
		},
	}
}

func headerExample() t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Column{
				Width:     t.Cells(20),
				Height:    t.Cells(8),
				MainAlign: t.MainAxisCenter,
				Style: t.Style{
					BackgroundColor: t.NewGradient(
						t.Hex("#7c3aed"),
						t.Hex("#2563eb"),
					).WithAngle(45),
					Padding: t.EdgeInsetsXY(1, 0),
				},
				Children: []t.Widget{
					t.Text{
						Content: "Welcome!",
						Style:   t.Style{ForegroundColor: t.White, Bold: true},
						Wrap:    t.WrapNone,
					},
					t.Text{
						Content: "Diagonal gradients",
						Style:   t.Style{ForegroundColor: t.Hex("#c4b5fd")},
						Wrap:    t.WrapNone,
					},
					t.Text{
						Content: "make great headers",
						Style:   t.Style{ForegroundColor: t.Hex("#a5b4fc")},
						Wrap:    t.WrapNone,
					},
				},
			},
			t.Text{
				Content: "Header with diagonal",
				Style:   t.Style{ForegroundColor: t.Hex("#64748b")},
				Wrap:    t.WrapNone,
			},
		},
	}
}

func main() {
	t.Run(&GradientDemo{
		scrollState: t.NewScrollState(),
	})
}
