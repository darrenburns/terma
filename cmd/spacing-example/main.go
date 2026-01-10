package main

import (
	"log"

	t "terma"
)


// App demonstrates padding, margin, and spacing.
type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		Children: []t.Widget{
			// Title
			t.Text{Content: "=== Padding, Margin & Spacing Demo ==="},

			t.Column{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "1. PADDING (space inside the widget, before content):"},
					t.Row{
						Spacing: 1,
						Children: []t.Widget{
							t.Text{
								Content: "none",
								Style: t.Style{
									BackgroundColor: t.Blue,
									ForegroundColor: t.White,
								},
							},
							t.Text{
								Content: "pad=1",
								Style: t.Style{
									BackgroundColor: t.Blue,
									ForegroundColor: t.White,
									Padding:         t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "pad=2",
								Style: t.Style{
									BackgroundColor: t.Blue,
									ForegroundColor: t.White,
									Padding:         t.EdgeInsetsAll(2),
								},
							},
						},
					},
				},
			},

			// Section 2: Margin demo
			t.Column{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "2. MARGIN (space outside the widget, gray=container):"},
					t.Row{
						Style: t.Style{BackgroundColor: t.BrightBlack},
						Children: []t.Widget{
							t.Text{
								Content: "none",
								Style: t.Style{
									BackgroundColor: t.Green,
									ForegroundColor: t.Black,
								},
							},
							t.Text{
								Content: "margin=1",
								Style: t.Style{
									BackgroundColor: t.Green,
									ForegroundColor: t.Black,
									Margin:          t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "margin=2",
								Style: t.Style{
									BackgroundColor: t.Green,
									ForegroundColor: t.Black,
									Margin:          t.EdgeInsetsAll(2),
								},
							},
						},
					},
				},
			},

			// Section 3: Spacing demo
			t.Column{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "3. SPACING (uniform gap between container children):"},
					t.Text{Content: "(spacing: 0, 1, 2 from left to right)", Style: t.Style{ForegroundColor: t.BrightBlack}},
					t.Row{
						Style: t.Style{
							BackgroundColor: t.Green,
							Padding:         t.EdgeInsetsAll(1),
						},
						Spacing: 4,
						Children: []t.Widget{
							t.Column{
								Children: []t.Widget{
									t.Text{Content: "A", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "B", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "C", Style: t.Style{BackgroundColor: t.Red}},
								},
							},
							t.Column{
								Spacing: 1,
								Children: []t.Widget{
									t.Text{Content: "A", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "B", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "C", Style: t.Style{BackgroundColor: t.Red}},
								},
							},
							t.Column{
								Spacing: 2,
								Children: []t.Widget{
									t.Text{Content: "A", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "B", Style: t.Style{BackgroundColor: t.Red}},
									t.Text{Content: "C", Style: t.Style{BackgroundColor: t.Red}},
								},
							},
						},
					},
				},
			},

			// Section 4: Combined
			t.Column{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "4. COMBINED (container has padding, children have margin):"},
					t.Column{
						Spacing: 1,
						Style: t.Style{
							BackgroundColor: t.BrightBlack,
							Padding:         t.EdgeInsetsAll(2),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Child with margin=1",
								Style: t.Style{
									BackgroundColor: t.Red,
									ForegroundColor: t.White,
									Margin:          t.EdgeInsetsAll(1),
								},
							},
							t.Text{
								Content: "Child with no margin",
								Style: t.Style{
									BackgroundColor: t.Cyan,
									ForegroundColor: t.Black,
								},
							},
						},
					},
				},
			},

			t.Text{Content: "Press Ctrl+C to quit"},
		},
	}
}

func main() {
	app := &App{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
