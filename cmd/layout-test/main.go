package main

import "terma"

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Row{
		Spacing:    2,
		Width:      terma.Cells(21),
		Height:     terma.Cells(5),
		CrossAlign: terma.CrossAxisCenter,
		MainAlign:  terma.MainAxisCenter,
		Style: terma.Style{
			BackgroundColor: terma.Red,
			Border:          terma.RoundedBorder(terma.Blue),
		},
		Children: []terma.Widget{
			terma.Text{Wrap: terma.WrapHard, Content: "Hello", Style: terma.Style{BackgroundColor: terma.Black}, Width: terma.Cells(4)},
			terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
			terma.Text{Content: "!"},
		},
	}
}

func main() {
	terma.InitLogger()
	terma.SetDebugLogging(true)
	terma.Run(&App{})
}
