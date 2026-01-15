package main

import (
	"log"

	t "terma"
)

// App demonstrates the Percent dimension type.
type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  t.Style{BackgroundColor: theme.Background},
		Children: []t.Widget{
			// Title
			t.Text{
				Content: "Percent Dimension Demo",
				Style: t.Style{
					ForegroundColor: theme.Primary,
					Padding:         t.EdgeInsets{Bottom: 1},
				},
			},

			// Section 1: Basic percentages
			section(ctx, "1. Basic Percentages",
				t.Column{
					Spacing: 1,
					Children: []t.Widget{
						labeledRow(ctx, "50%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "50%", t.Percent(50), theme.Primary),
								},
							},
						),
						labeledRow(ctx, "25% + 75%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "25%", t.Percent(25), theme.Primary),
									box(ctx, "75%", t.Percent(75), theme.Accent),
								},
							},
						),
						labeledRow(ctx, "33% + 33% + 34%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "33%", t.Percent(33), theme.Primary),
									box(ctx, "33%", t.Percent(33), theme.Accent),
									box(ctx, "34%", t.Percent(34), theme.Success),
								},
							},
						),
					},
				},
			),

			// Section 2: Mixed with Cells
			section(ctx, "2. Mixed with Fixed Cells",
				t.Column{
					Spacing: 1,
					Children: []t.Widget{
						labeledRow(ctx, "10 cells + 50%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "10 cells", t.Cells(10), theme.Warning),
									box(ctx, "50%", t.Percent(50), theme.Primary),
								},
							},
						),
						labeledRow(ctx, "50% + 15 cells + 50%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "50%", t.Percent(50), theme.Primary),
									box(ctx, "15 cells", t.Cells(15), theme.Warning),
									box(ctx, "50%", t.Percent(50), theme.Accent),
								},
							},
						),
					},
				},
			),

			// Section 3: Mixed with Flex
			section(ctx, "3. Mixed with Flex",
				t.Column{
					Spacing: 1,
					Children: []t.Widget{
						labeledRow(ctx, "30% + Flex(1)",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "30%", t.Percent(30), theme.Primary),
									box(ctx, "Flex(1)", t.Flex(1), theme.Success),
								},
							},
						),
						labeledRow(ctx, "20% + Flex(1) + 20%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "20%", t.Percent(20), theme.Primary),
									box(ctx, "Flex(1)", t.Flex(1), theme.Success),
									box(ctx, "20%", t.Percent(20), theme.Primary),
								},
							},
						),
					},
				},
			),

			// Section 4: Overflow (>100%)
			section(ctx, "4. Overflow (>100%)",
				t.Column{
					Spacing: 1,
					Children: []t.Widget{
						labeledRow(ctx, "60% + 60% = 120%",
							t.Row{
								Width: t.Flex(1),
								Style: t.Style{BackgroundColor: theme.Surface},
								Children: []t.Widget{
									box(ctx, "60%", t.Percent(60), theme.Error),
									box(ctx, "60%", t.Percent(60), theme.Warning),
								},
							},
						),
					},
				},
			),

			// Section 5: Vertical percentages
			section(ctx, "5. Vertical Percentages",
				t.Row{
					Width:   t.Flex(1),
					Height:  t.Cells(6),
					Spacing: 1,
					Children: []t.Widget{
						t.Column{
							Width:  t.Cells(15),
							Height: t.Flex(1),
							Style:  t.Style{BackgroundColor: theme.Surface},
							Children: []t.Widget{
								t.Column{
									Height: t.Percent(50),
									Style:  t.Style{BackgroundColor: theme.Primary},
									Children: []t.Widget{
										t.Text{Content: "50%"},
									},
								},
								t.Column{
									Height: t.Percent(50),
									Style:  t.Style{BackgroundColor: theme.Accent},
									Children: []t.Widget{
										t.Text{Content: "50%"},
									},
								},
							},
						},
						t.Column{
							Width:  t.Cells(15),
							Height: t.Flex(1),
							Style:  t.Style{BackgroundColor: theme.Surface},
							Children: []t.Widget{
								t.Column{
									Height: t.Percent(25),
									Style:  t.Style{BackgroundColor: theme.Primary},
									Children: []t.Widget{
										t.Text{Content: "25%"},
									},
								},
								t.Column{
									Height: t.Percent(75),
									Style:  t.Style{BackgroundColor: theme.Success},
									Children: []t.Widget{
										t.Text{Content: "75%"},
									},
								},
							},
						},
						t.Column{
							Width:  t.Cells(20),
							Height: t.Flex(1),
							Style:  t.Style{BackgroundColor: theme.Surface},
							Children: []t.Widget{
								t.Column{
									Height: t.Percent(33),
									Style:  t.Style{BackgroundColor: theme.Primary},
									Children: []t.Widget{
										t.Text{Content: "33%"},
									},
								},
								t.Column{
									Height: t.Percent(33),
									Style:  t.Style{BackgroundColor: theme.Accent},
									Children: []t.Widget{
										t.Text{Content: "33%"},
									},
								},
								t.Column{
									Height: t.Percent(34),
									Style:  t.Style{BackgroundColor: theme.Success},
									Children: []t.Widget{
										t.Text{Content: "34%"},
									},
								},
							},
						},
					},
				},
			),

			t.Spacer{},

			// Footer
			t.Text{
				Content: "Press Ctrl+C to quit",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

// section creates a labeled section with content.
func section(ctx t.BuildContext, title string, content t.Widget) t.Widget {
	theme := ctx.Theme()
	return t.Column{
		Style: t.Style{Padding: t.EdgeInsets{Bottom: 1}},
		Children: []t.Widget{
			t.Text{
				Content: title,
				Style: t.Style{
					ForegroundColor: theme.Accent,
					Padding:         t.EdgeInsets{Bottom: 1},
				},
			},
			content,
		},
	}
}

// labeledRow creates a row with a label prefix.
func labeledRow(ctx t.BuildContext, label string, content t.Widget) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: label,
				Width:   t.Cells(20),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			content,
		},
	}
}

// box creates a colored box with a label.
func box(ctx t.BuildContext, label string, width t.Dimension, bg t.Color) t.Widget {
	return t.Text{
		Content: label,
		Width:   width,
		Style: t.Style{
			BackgroundColor: bg,
			ForegroundColor: t.White,
			Padding:         t.EdgeInsets{Left: 1, Right: 1},
		},
	}
}

func main() {
	app := &App{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
