package main

import (
	"fmt"
	"log"
	t "terma"
)

func init() {
	t.InitDebug()
}

type App struct {
	themeIndex  t.Signal[int]
	themeNames  []string
	scrollState *t.ScrollState
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Width:  t.Fr(1),
		Height: t.Fr(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Row{
				Children: []t.Widget{
					t.Text{Spans: t.ParseMarkup("[b $Primary]Markup Demo[/]", theme)},
					t.Text{Content: "  "},
					t.Text{Spans: t.ParseMarkup("[$TextMuted]theme: [/][$Accent]"+theme.Name+"[/]", theme)},
				},
			},
			t.Text{Content: ""},

			// Scrollable content
			t.Scrollable{
				ID:     "content",
				State:  a.scrollState,
				Width:  t.Fr(1),
				Height: t.Fr(1),
				Child: t.Column{
					Width: t.Fr(1),
					Children: []t.Widget{
						a.section("Style Modifiers", theme,
							`[bold]bold[/]`,
							`[italic]italic[/]`,
							`[underline]underline[/]`,
							`[b i u]all three[/]`,
						),

						a.section("Theme Colors", theme,
							`[$Primary]Primary[/]`,
							`[$Secondary]Secondary[/]`,
							`[$Accent]Accent[/]`,
							`[$Error]Error[/] [$Warning]Warning[/] [$Success]Success[/] [$Info]Info[/]`,
						),

						a.section("Backgrounds", theme,
							`[$Text on $Surface] on Surface [/]`,
							`[$TextOnPrimary on $Primary] on Primary [/]`,
							`[on #333333] hex background [/]`,
						),

						a.section("Combined Styles", theme,
							`[b $Accent]bold + color[/]`,
							`[i $TextMuted on $Surface] italic + muted + bg [/]`,
							`Press [b $Info]Enter[/] to continue`,
						),

						a.section("Nesting", theme,
							`[b]bold [i]bold+italic[/] bold[/]`,
							`[$Primary]A [$Secondary]B[/] A[/]`,
						),

						a.section("Escaping & Hex", theme,
							`Use [[brackets]] in text`,
							`[#ff6600]Hex[/] [#00cc99]colors[/] [#ff00ff]work[/]`,
						),

						a.section("Case Insensitive", theme,
							`[$primary]a[/] [$PRIMARY]b[/] [$text_muted]c[/]`,
						),
					},
				},
			},

			// Footer (always visible)
			t.Text{
				Spans: t.ParseMarkup("[$TextMuted]Press [/][b $Accent]t[/][$TextMuted] to cycle themes, [/][b $Accent]q[/][$TextMuted] to quit[/]", theme),
			},
		},
	}
}

func (a *App) section(title string, theme t.ThemeData, examples ...string) t.Widget {
	children := make([]t.Widget, len(examples))
	for i, markup := range examples {
		children[i] = a.example(markup, theme)
	}

	return t.Column{
		Width: t.Fr(1),
		Style: t.Style{
			Border: t.RoundedBorder(theme.Border,
				t.BorderDecoration{Text: title, Position: t.DecorationTopLeft, Color: theme.Text},
			),
			Padding: t.EdgeInsetsXY(1, 0),
			Margin:  t.EdgeInsetsTRBL(0, 0, 1, 0),
		},
		Children: children,
	}
}

func (a *App) example(markup string, theme t.ThemeData) t.Widget {
	return t.Row{
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("%-45s", markup),
				Width:   t.Cells(45),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.Text{Spans: t.ParseMarkup(markup, theme)},
		},
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "t", Name: "Next theme", Action: a.cycleTheme},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *App) cycleTheme() {
	a.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(a.themeNames)
		t.SetTheme(a.themeNames[next])
		return next
	})
}

func main() {
	themeNames := []string{
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

	app := &App{
		themeIndex:  t.NewSignal(0),
		themeNames:  themeNames,
		scrollState: t.NewScrollState(),
	}

	t.SetTheme(themeNames[0])

	fmt.Println("Starting markup demo...")
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
