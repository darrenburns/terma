package main

import (
	"fmt"
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

type ScrollDemo struct{}

func (s *ScrollDemo) Build(ctx t.BuildContext) t.Widget {
	// Generate a list of items that will exceed the viewport
	var items []t.Widget
	for i := 1; i <= 50; i++ {
		color := t.White
		if i%2 == 0 {
			color = t.BrightBlack
		}
		items = append(items, t.Text{
			Content: fmt.Sprintf("Item %d - This is a scrollable list item", i),
			Style:   t.Style{ForegroundColor: color},
		})
	}

	return t.Column{
		ID:      "root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Scroll Demo",
				Style: t.Style{
					ForegroundColor: t.BrightWhite,
					BackgroundColor: t.Blue,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Instructions
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Use "),
					t.BoldSpan("↑/↓", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", t.BrightCyan),
					t.PlainSpan(" to scroll • "),
					t.BoldSpan("PgUp/PgDn", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("Ctrl+U/D", t.BrightCyan),
					t.PlainSpan(" for half-page • "),
					t.BoldSpan("Home/End", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("g/G", t.BrightCyan),
					t.PlainSpan(" for top/bottom"),
				},
			},

			// Side by side: scrollable list and non-scrollable content
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					// Scrollable list with fixed height
					&t.Scrollable{
						ID:     "scroll-list",
						Height: t.Cells(15),
						Style: t.Style{
							Border:  t.RoundedBorder(t.Cyan, t.BorderTitle("Scrollable List")),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: t.Column{
							Children: items,
						},
					},

					// Second scrollable panel with different content
					&t.Scrollable{
						ID:     "scroll-text",
						Height: t.Cells(15),
						Width:  t.Cells(40),
						Style: t.Style{
							Border:  t.RoundedBorder(t.Magenta, t.BorderTitle("Long Text")),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: t.Column{
							Children: []t.Widget{
								t.Text{Content: "Lorem ipsum dolor sit amet, consectetur"},
								t.Text{Content: "adipiscing elit. Sed do eiusmod tempor"},
								t.Text{Content: "incididunt ut labore et dolore magna"},
								t.Text{Content: "aliqua. Ut enim ad minim veniam, quis"},
								t.Text{Content: "nostrud exercitation ullamco laboris"},
								t.Text{Content: "nisi ut aliquip ex ea commodo consequat."},
								t.Text{Content: ""},
								t.Text{Content: "Duis aute irure dolor in reprehenderit"},
								t.Text{Content: "in voluptate velit esse cillum dolore"},
								t.Text{Content: "eu fugiat nulla pariatur. Excepteur sint"},
								t.Text{Content: "occaecat cupidatat non proident, sunt in"},
								t.Text{Content: "culpa qui officia deserunt mollit anim"},
								t.Text{Content: "id est laborum."},
								t.Text{Content: ""},
								t.Text{Content: "Sed ut perspiciatis unde omnis iste"},
								t.Text{Content: "natus error sit voluptatem accusantium"},
								t.Text{Content: "doloremque laudantium, totam rem aperiam"},
								t.Text{Content: "eaque ipsa quae ab illo inventore"},
								t.Text{Content: "veritatis et quasi architecto beatae"},
								t.Text{Content: "vitae dicta sunt explicabo."},
								t.Text{Content: ""},
								t.Text{Content: "Nemo enim ipsam voluptatem quia voluptas"},
								t.Text{Content: "sit aspernatur aut odit aut fugit, sed"},
								t.Text{Content: "quia consequuntur magni dolores eos qui"},
								t.Text{Content: "ratione voluptatem sequi nesciunt."},
							},
						},
					},
				},
			},

			// Example of disabled scrolling
			t.Row{
				Children: []t.Widget{
					&t.Scrollable{
						ID:            "no-scroll",
						Height:        t.Cells(5),
						DisableScroll: true,
						Style: t.Style{
							Border:  t.SquareBorder(t.Yellow, t.BorderTitle("Scrolling Disabled")),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: t.Column{
							Children: []t.Widget{
								t.Text{Content: "This panel has scrolling disabled."},
								t.Text{Content: "Content that overflows is hidden."},
								t.Text{Content: "Line 3 - might be visible"},
								t.Text{Content: "Line 4 - might be cut off"},
								t.Text{Content: "Line 5 - probably hidden"},
								t.Text{Content: "Line 6 - definitely hidden"},
								t.Text{Content: "Line 7 - not visible"},
							},
						},
					},
				},
			},

			// Footer
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Press "),
					t.BoldSpan("Tab", t.BrightYellow),
					t.PlainSpan(" to switch focus between scrollable panels • "),
					t.BoldSpan("Ctrl+C", t.BrightRed),
					t.PlainSpan(" to quit"),
				},
			},
		},
	}
}

func main() {
	app := &ScrollDemo{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
