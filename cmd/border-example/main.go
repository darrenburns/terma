package main

import (
	"fmt"
	"log"

	t "terma"
)

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

type BorderDemo struct {
	themeIndex t.Signal[int]
}

func (b *BorderDemo) cycleTheme() {
	b.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (b *BorderDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Next theme", Action: b.cycleTheme},
	}
}

func (b *BorderDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := b.themeIndex.Get()
	currentTheme := themeNames[themeIdx]

	// Helper to create a border demo box
	borderBox := func(name string, border t.Border) t.Widget {
		return t.Text{
			Content: name,
			Style: t.Style{
				Border:  border,
				Padding: t.EdgeInsetsXY(1, 0),
			},
		}
	}

	return t.Column{
		ID:      "root",
		Height:  t.Flex(1),
		Width:   t.Flex(1),
		Spacing: 1,
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Border Demo - All Border Types",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Theme indicator
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]Theme: [/][$Accent]%s[/][$TextMuted] (press t to change, Ctrl+C to quit)[/]", currentTheme), theme),
			},

			// Row 1: Basic borders
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					borderBox("Square", t.SquareBorder(theme.Primary)),
					borderBox("Rounded", t.RoundedBorder(theme.Primary)),
					borderBox("Double", t.DoubleBorder(theme.Primary)),
					borderBox("Heavy", t.HeavyBorder(theme.Primary)),
				},
			},

			// Row 2: More line-based borders
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					borderBox("Dashed", t.DashedBorder(theme.Secondary)),
					borderBox("Ascii", t.AsciiBorder(theme.Secondary)),
				},
			},

			// Row 3: Block element borders
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					borderBox("Inner", t.InnerBorder(theme.Info)),
					borderBox("Outer", t.OuterBorder(theme.Info)),
					borderBox("Thick", t.ThickBorder(theme.Info)),
				},
			},

			// Row 4: Key-cap borders
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					borderBox("HKey", t.HKeyBorder(theme.Warning)),
					borderBox("VKey", t.VKeyBorder(theme.Warning)),
				},
			},

			// Decorations demo with Double border
			t.Row{
				Style: t.Style{
					Border: t.DoubleBorder(theme.Accent,
						t.BorderTitle("Double with Title"),
						t.BorderSubtitleCenter("and subtitle"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					t.Text{Content: "Decorations work with all border types"},
				},
			},

			// Heavy border with decorations
			t.Row{
				Style: t.Style{
					Border: t.HeavyBorder(theme.Error,
						t.BorderTitleCenter("Heavy Border"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					t.Text{Content: "Heavy borders make a bold statement"},
				},
			},
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	app := &BorderDemo{
		themeIndex: t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
