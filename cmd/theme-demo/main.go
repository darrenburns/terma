package main

import (
	"fmt"
	"log"
	t "terma"
)


// App demonstrates the theme system with interactive theme switching.
type App struct {
	listState  *t.ListState[string]
	themeIndex t.Signal[int]
	themeNames []string
}

// Build returns the widget tree for the theme demo.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := a.themeIndex.Get()
	currentTheme := a.themeNames[themeIdx]

	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Theme Demo",
				Style: t.Style{
					ForegroundColor: theme.Primary,
				},
			},
			t.Text{Content: ""},

			// Current theme display
			t.Row{
				Children: []t.Widget{
					t.Text{
						Content: "Current theme: ",
						Style:   t.Style{ForegroundColor: theme.TextMuted},
					},
					t.Text{
						Content: currentTheme,
						Style: t.Style{
							ForegroundColor: theme.Accent,
						},
					},
				},
			},
			t.Text{Content: ""},

			// Color swatches
			t.Text{
				Content: "Color Palette:",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			a.buildColorSwatches(theme),
			t.Text{Content: ""},

			// Sample button
			t.Text{
				Content: "Button (auto-themed):",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			&t.Button{
				ID:    "sample-button",
				Label: " Click Me ",
			},
			t.Text{Content: ""},

			// Sample list
			t.Text{
				Content: "List (auto-themed):",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.List[string]{
				ID:          "sample-list",
				State:       a.listState,
				MultiSelect: true,
			},
			t.Text{Content: ""},

			// Feedback colors
			t.Text{
				Content: "Feedback Colors:",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{
						Content: " Error ",
						Style:   t.Style{BackgroundColor: theme.Error, ForegroundColor: theme.TextOnError},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Warning ",
						Style:   t.Style{BackgroundColor: theme.Warning, ForegroundColor: theme.TextOnWarning},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Success ",
						Style:   t.Style{BackgroundColor: theme.Success, ForegroundColor: theme.TextOnSuccess},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Info ",
						Style:   t.Style{BackgroundColor: theme.Info, ForegroundColor: theme.TextOnInfo},
					},
				},
			},
			t.Text{Content: ""},

			// Additional colors (new)
			t.Text{
				Content: "Additional Colors:",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			a.buildAdditionalSwatches(theme),
			t.Text{Content: ""},

			// Instructions
			t.Text{
				Content: "Press 't' to cycle themes, 'q' to quit",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

// buildColorSwatches creates a row of color swatches showing theme colors.
func (a *App) buildColorSwatches(theme t.ThemeData) t.Widget {
	return t.Row{
		Children: []t.Widget{
			t.Text{
				Content: " Primary ",
				Style:   t.Style{BackgroundColor: theme.Primary, ForegroundColor: theme.TextOnPrimary},
			},
			t.Text{Content: " "},
			t.Text{
				Content: " Secondary ",
				Style:   t.Style{BackgroundColor: theme.Secondary, ForegroundColor: theme.TextOnSecondary},
			},
			t.Text{Content: " "},
			t.Text{
				Content: " Accent ",
				Style:   t.Style{BackgroundColor: theme.Accent, ForegroundColor: theme.TextOnAccent},
			},
			t.Text{Content: " "},
			t.Text{
				Content: " Surface ",
				Style:   t.Style{BackgroundColor: theme.Surface, ForegroundColor: theme.Text},
			},
		},
	}
}

// buildAdditionalSwatches creates a row of additional color swatches showing new theme colors.
func (a *App) buildAdditionalSwatches(theme t.ThemeData) t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Row{
				Children: []t.Widget{
					t.Text{
						Content: " Surface2 ",
						Style:   t.Style{BackgroundColor: theme.Surface2, ForegroundColor: theme.Text},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Surface3 ",
						Style:   t.Style{BackgroundColor: theme.Surface3, ForegroundColor: theme.Text},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Selection ",
						Style:   t.Style{BackgroundColor: theme.Selection, ForegroundColor: theme.SelectionText},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Link ",
						Style:   t.Style{ForegroundColor: theme.Link},
					},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{
						Content: " Scrollbar ",
						Style:   t.Style{BackgroundColor: theme.ScrollbarTrack, ForegroundColor: theme.ScrollbarThumb},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Disabled ",
						Style:   t.Style{ForegroundColor: theme.TextDisabled},
					},
					t.Text{Content: " "},
					t.Text{
						Content: " Placeholder ",
						Style:   t.Style{ForegroundColor: theme.Placeholder},
					},
				},
			},
		},
	}
}

// Keybinds returns the keybindings for the demo.
func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Next theme", Action: a.cycleTheme},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

// cycleTheme switches to the next theme in the list.
func (a *App) cycleTheme() {
	a.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(a.themeNames)
		t.SetTheme(a.themeNames[next])
		return next
	})
}

func main() {
	// Get available theme names (dark themes followed by light themes)
	themeNames := []string{
		// Dark themes
		t.ThemeNameRosePine,
		t.ThemeNameDracula,
		t.ThemeNameTokyoNight,
		t.ThemeNameCatppuccin,
		t.ThemeNameGruvbox,
		t.ThemeNameNord,
		t.ThemeNameSolarized,
		t.ThemeNameKanagawa,
		t.ThemeNameMonokai,
		// Light themes
		t.ThemeNameRosePineDawn,
		t.ThemeNameDraculaLight,
		t.ThemeNameTokyoNightDay,
		t.ThemeNameCatppuccinLatte,
		t.ThemeNameGruvboxLight,
		t.ThemeNameNordLight,
		t.ThemeNameSolarizedLight,
		t.ThemeNameKanagawaLotus,
		t.ThemeNameMonokaiLight,
	}

	// Create list state with sample items
	listState := t.NewListState([]string{
		"First item",
		"Second item",
		"Third item",
		"Fourth item",
	})

	app := &App{
		listState:  listState,
		themeIndex: t.NewSignal(0),
		themeNames: themeNames,
	}

	// Start with first theme
	t.SetTheme(themeNames[0])

	fmt.Println("Starting theme demo...")
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
