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
				Content: "Border Demo",
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

			// Row with square and rounded borders side by side
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Text{
						//Width:   t.Cells(10),
						Content: "Square Border",
						Style: t.Style{
							Border:  t.SquareBorder(theme.Info),
							Padding: t.EdgeInsetsAll(1),
						},
					},
					t.Text{
						//Width:   t.Cells(10),
						Content: "Rounded Border",
						Style: t.Style{
							Border:  t.RoundedBorder(theme.Secondary),
							Padding: t.EdgeInsetsAll(1),
						},
					},
				},
			},

			// Border decorations - titles and subtitles
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Column{
						Style: t.Style{
							Border: t.RoundedBorder(theme.Info,
								t.BorderTitle("Settings"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{Content: "Option 1: Enabled"},
							t.Text{Content: "Option 2: Disabled"},
						},
					},
					t.Column{
						Style: t.Style{
							Border: t.SquareBorder(theme.Warning,
								t.BorderTitle("Info"),
								t.BorderSubtitle("Press q to quit"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{Content: "Title and subtitle"},
							t.Text{Content: "on the same border"},
						},
					},
				},
			},

			// Centered decorations
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Text{
						Content: "Center title",
						Style: t.Style{
							Border: t.RoundedBorder(theme.Success,
								t.BorderTitleCenter("Centered"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
					},
					t.Text{
						Content: "Right aligned",
						Style: t.Style{
							Border: t.SquareBorder(theme.Secondary,
								t.BorderTitleRight("Right"),
								t.BorderSubtitleRight("Also Right"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
					},
				},
			},

			// Colored decorations
			t.Row{
				Style: t.Style{
					Border: t.Border{
						Style: t.BorderRounded,
						Color: theme.Border,
						Decorations: []t.BorderDecoration{
							{Text: "Status", Position: t.DecorationTopLeft, Color: theme.Info},
							{Text: "Online", Position: t.DecorationTopRight, Color: theme.Success},
							{Text: "v1.0.0", Position: t.DecorationBottomRight, Color: theme.Warning},
						},
					},
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					t.Text{Content: "Decorations can have"},
					t.Text{Content: "individual colors!"},
				},
			},

			// Nested borders with titles
			t.Row{
				ID: "outer-box",
				Style: t.Style{
					Border: t.RoundedBorder(theme.Primary,
						t.BorderTitle("Outer"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Spacing: 2,
				Children: []t.Widget{
					t.Text{Content: "Outer container"},
					t.Column{
						ID: "inner-box",
						Style: t.Style{
							Border: t.SquareBorder(theme.Error,
								t.BorderTitle("Inner"),
								t.BorderSubtitleCenter("Nested!"),
							),
							BackgroundColor: theme.Surface,
							//Padding:         t.EdgeInsetsAll(1),
							//Margin:          t.EdgeInsetsTRBL(1, 0, 0, 0),
						},
						Children: []t.Widget{
							t.Text{Content: "Nested border with title"},
						},
					},
				},
			},

			// Rich text with spans
			t.Row{
				Style: t.Style{
					Border: t.RoundedBorder(theme.Info,
						t.BorderTitle("Rich Text"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					// Status line with multiple colored spans
					t.Text{
						Spans: t.ParseMarkup("Status: [$Success]Online[/] | Errors: [$Error]3[/]", theme),
					},
					// Text with formatting attributes
					t.Text{
						Spans: t.ParseMarkup("This is [b $Text]bold[/], [i $Info]italic[/], and [u $Warning]underlined[/] text.", theme),
					},
					// Fully styled span
					t.Text{
						//Wrap:  t.WrapSoft,
						Spans: t.ParseMarkup("Mixed: [b $Secondary]Bold+Color[/] and [i u $Primary]Italic+Underline[/]", theme),
					},
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
