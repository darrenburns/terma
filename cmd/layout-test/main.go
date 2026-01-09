package main

import "terma"

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Row{
		Spacing: 2,
		Width:   terma.Cells(3),
		Style: terma.Style{
			BackgroundColor: terma.Red,
		},
		Children: []terma.Widget{
			terma.Text{Content: "Hello", Style: terma.Style{BackgroundColor: terma.Black}},
			terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
			terma.Text{Content: "!"},
		},
	}
}

func main() {
	terma.SetDebugLogging(true)
	terma.Run(&App{})
}
