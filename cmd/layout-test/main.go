package main

import "terma"

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Row{
		Spacing:    2,
		Width:      terma.Cells(21),
		Height:     terma.Cells(5),
		CrossAlign: terma.CrossAxisCenter,
		MainAlign:  terma.MainAxisEnd,
		Style: terma.Style{
			BackgroundColor: terma.Red,
			Margin:          terma.EdgeInsetsXY(2, 1),
			Border:          terma.RoundedBorder(terma.Blue),
		},
		Children: []terma.Widget{
			terma.Text{Wrap: terma.WrapHard, Content: "Hello", Style: terma.Style{BackgroundColor: terma.Black}, Width: terma.Cells(4)},
			terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
			terma.Text{Content: "!"},
			terma.Row{
				Spacing: 1,
				Style:   terma.Style{BackgroundColor: terma.Green, Padding: terma.EdgeInsetsXY(2, 1)},
				Children: []terma.Widget{
					terma.Text{Wrap: terma.WrapHard, Content: "Hello", Style: terma.Style{BackgroundColor: terma.Black}, Width: terma.Cells(4)},
					terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
					terma.Text{Content: "!"},
				},
			},
		},
	}
}

func main() {
	terma.InitLogger()
	terma.SetDebugLogging(true)
	terma.Run(&App{})
}
