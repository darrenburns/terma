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
	bodyState       *t.ScrollState
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
				Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]Theme: [/][$Accent]%s[/][$TextMuted] (press t to change)[/]", currentTheme), theme),
			},

			// Instructions
			t.Text{
				Spans: t.ParseMarkup("Use [b $Info]↑/↓[/] or [b $Info]j/k[/] to scroll • [b $Info]PgUp/PgDn[/] or [b $Info]Ctrl+U/D[/] for half-page • [b $Info]Home/End[/] or [b $Info]g/G[/] for top/bottom", theme),
			},

			// Body: scrollable container with all the panels
			&t.Scrollable{
				ID:     "body",
				State:  s.bodyState,
				Height: t.Fr(1),
				Child: t.Column{
					Spacing: 1,
					Children: []t.Widget{
						// Side by side: scrollable list and long text
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
									Child: t.Text{
										Content: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.\n\nDuis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.\n\nSed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.\n\nNemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.",
										Width:   t.Fr(1),
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
									Child: t.Text{
										Content: "This panel has scrolling disabled. Content that overflows is hidden. This demonstrates what happens when you have more text than fits in the available space but scrolling is not enabled.",
										Width:   t.Fr(1),
									},
								},
							},
						},
					},
				},
			},

			// Footer
			t.Text{
				Spans: t.ParseMarkup("Press [b $Warning]Tab[/] to switch focus between scrollable panels • [b $Error]Ctrl+C[/] to quit", theme),
			},
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	app := &ScrollDemo{
		bodyState:       t.NewScrollState(),
		scrollListState: t.NewScrollState(),
		scrollTextState: t.NewScrollState(),
		noScrollState:   t.NewScrollState(),
		themeIndex:      t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
