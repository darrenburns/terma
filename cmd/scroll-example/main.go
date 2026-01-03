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

// Theme names for cycling
var themeNames = []string{
	t.ThemeNameRosePine,
	t.ThemeNameDracula,
	t.ThemeNameTokyoNight,
	t.ThemeNameCatppuccin,
	t.ThemeNameGruvbox,
	t.ThemeNameNord,
	t.ThemeNameOneDark,
	t.ThemeNameSolarized,
	t.ThemeNameKanagawa,
	t.ThemeNameMonokai,
}

type ScrollDemo struct {
	scrollListState *t.ScrollState
	scrollTextState *t.ScrollState
	noScrollState   *t.ScrollState
	themeIndex      *t.Signal[int]
}

func (s *ScrollDemo) cycleTheme() {
	s.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (s *ScrollDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Next theme", Action: s.cycleTheme},
	}
}

func (s *ScrollDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := s.themeIndex.Get()
	currentTheme := themeNames[themeIdx]

	// Generate a list of items that will exceed the viewport
	var items []t.Widget
	for i := 1; i <= 50; i++ {
		color := theme.Text
		if i%2 == 0 {
			color = theme.TextMuted
		}
		items = append(items, t.Text{
			Content: fmt.Sprintf("Item %d - This is a scrollable list item", i),
			Style:   t.Style{ForegroundColor: color},
		})
	}

	return t.Column{
		ID:      "root",
		Height:  t.Fr(1),
		Spacing: 1,
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Scroll Demo",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Theme indicator
			t.Text{
				Spans: []t.Span{
					t.ColorSpan("Theme: ", theme.TextMuted),
					t.ColorSpan(currentTheme, theme.Accent),
					t.ColorSpan(" (press t to change)", theme.TextMuted),
				},
			},

			// Instructions
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Use "),
					t.BoldSpan("↑/↓", theme.Info),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", theme.Info),
					t.PlainSpan(" to scroll • "),
					t.BoldSpan("PgUp/PgDn", theme.Info),
					t.PlainSpan(" or "),
					t.BoldSpan("Ctrl+U/D", theme.Info),
					t.PlainSpan(" for half-page • "),
					t.BoldSpan("Home/End", theme.Info),
					t.PlainSpan(" or "),
					t.BoldSpan("g/G", theme.Info),
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
						State:  s.scrollListState,
						Height: t.Cells(15),
						Width:  t.Fr(1),
						Style: t.Style{
							Border:  t.RoundedBorder(theme.Info, t.BorderTitle("Scrollable List")),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: t.Column{
							Children: items,
						},
					},

					// Second scrollable panel with different content
					&t.Scrollable{
						ID:     "scroll-text",
						State:  s.scrollTextState,
						Height: t.Cells(15),
						Width:  t.Fr(1),
						Style: t.Style{
							Border:  t.RoundedBorder(theme.Secondary, t.BorderTitle("Long Text")),
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
						State:         s.noScrollState,
						Height:        t.Cells(5),
						DisableScroll: true,
						Style: t.Style{
							Border:  t.SquareBorder(theme.Warning, t.BorderTitle("Scrolling Disabled")),
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
					t.BoldSpan("Tab", theme.Warning),
					t.PlainSpan(" to switch focus between scrollable panels • "),
					t.BoldSpan("Ctrl+C", theme.Error),
					t.PlainSpan(" to quit"),
				},
			},
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	app := &ScrollDemo{
		scrollListState: t.NewScrollState(),
		scrollTextState: t.NewScrollState(),
		noScrollState:   t.NewScrollState(),
		themeIndex:      t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
