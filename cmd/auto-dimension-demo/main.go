package main

import (
	"log"

	t "terma"
)

// App demonstrates the difference between Auto, unset, Flex, and Cells dimensions.
type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Helper to create a labeled box with background color
	makeBox := func(label string, bg t.Color) t.Widget {
		return t.Text{
			Content: label,
			Style: t.Style{
				BackgroundColor: bg,
				ForegroundColor: t.Black,
				Padding:         t.EdgeInsetsXY(1, 0),
			},
		}
	}

	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: "Width: Auto vs Unset vs Flex vs Cells",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Content: "Parent Column has CrossAxisStretch (default), which tries to stretch children horizontally.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},

			// Section 1: Row width comparisons
			t.Column{
				Spacing: 1,
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					// Row with unset width (default) - gets stretched
					t.Row{
						// Width not set - parent stretches it
						Style: t.Style{BackgroundColor: t.Red},
						Children: []t.Widget{
							makeBox("Row width: unset (stretched by parent)", t.BrightRed),
						},
					},

					// Row with explicit Auto - fits content
					t.Row{
						Width: t.Auto, // Explicit Auto - resists stretching
						Style: t.Style{BackgroundColor: t.Green},
						Children: []t.Widget{
							makeBox("Row width: Auto (fits content)", t.BrightGreen),
						},
					},

					// Row with Flex(1) - fills space
					t.Row{
						Width: t.Flex(1),
						Style: t.Style{BackgroundColor: t.Blue},
						Children: []t.Widget{
							makeBox("Row width: Flex(1) (fills space)", t.BrightBlue),
						},
					},

					// Row with Cells(40) - fixed size
					t.Row{
						Width: t.Cells(40),
						Style: t.Style{BackgroundColor: t.Yellow},
						Children: []t.Widget{
							makeBox("Row width: Cells(40) (fixed)", t.BrightYellow),
						},
					},
				},
			},

			t.Spacer{Height: t.Cells(1)},

			t.Text{
				Content: "Height: Auto vs Unset vs Flex vs Cells",
				Style:   t.Style{ForegroundColor: theme.Primary},
			},
			t.Text{
				Content: "Parent Row has CrossAxisStretch (default), which tries to stretch children vertically.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},

			// Section 2: Column height comparisons
			t.Row{
				Height:  t.Cells(8),
				Spacing: 1,
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					// Column with unset height - gets stretched
					t.Column{
						// Height not set - parent stretches it
						Style: t.Style{BackgroundColor: t.Red},
						Children: []t.Widget{
							t.Text{
								Content: "unset",
								Style: t.Style{
									BackgroundColor: t.BrightRed,
									ForegroundColor: t.Black,
								},
							},
							t.Text{
								Content: "(stretched)",
								Style: t.Style{
									ForegroundColor: t.BrightRed,
								},
							},
						},
					},

					// Column with explicit Auto - fits content
					t.Column{
						Height: t.Auto, // Explicit Auto - resists stretching
						Style:  t.Style{BackgroundColor: t.Green},
						Children: []t.Widget{
							t.Text{
								Content: "Auto",
								Style: t.Style{
									BackgroundColor: t.BrightGreen,
									ForegroundColor: t.Black,
								},
							},
							t.Text{
								Content: "(fits)",
								Style: t.Style{
									ForegroundColor: t.BrightGreen,
								},
							},
						},
					},

					// Column with Flex(1) - fills space
					t.Column{
						Height: t.Flex(1),
						Style:  t.Style{BackgroundColor: t.Blue},
						Children: []t.Widget{
							t.Text{
								Content: "Flex(1)",
								Style: t.Style{
									BackgroundColor: t.BrightBlue,
									ForegroundColor: t.Black,
								},
							},
							t.Text{
								Content: "(fills)",
								Style: t.Style{
									ForegroundColor: t.BrightBlue,
								},
							},
						},
					},

					// Column with Cells(4) - fixed size
					t.Column{
						Height: t.Cells(4),
						Style:  t.Style{BackgroundColor: t.Yellow},
						Children: []t.Widget{
							t.Text{
								Content: "Cells(4)",
								Style: t.Style{
									BackgroundColor: t.BrightYellow,
									ForegroundColor: t.Black,
								},
							},
							t.Text{
								Content: "(fixed)",
								Style: t.Style{
									ForegroundColor: t.BrightYellow,
								},
							},
						},
					},
				},
			},

			t.Spacer{Height: t.Flex(1)},

			t.Text{
				Content: "Press Ctrl+C to quit",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func main() {
	app := &App{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
