package main

import "terma"

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Row{
		Spacing:    2,
		Width:      terma.Cells(40),
		Height:     terma.Cells(6),
		CrossAlign: terma.CrossAxisCenter,
		MainAlign:  terma.MainAxisEnd,
		Style: terma.Style{
			BackgroundColor: terma.Red,
			Margin:          terma.EdgeInsetsXY(2, 1),
			Border:          terma.RoundedBorder(terma.Blue),
		},
		Children: []terma.Widget{
			terma.Text{Wrap: terma.WrapHard, Content: "Hello", Style: terma.Style{BackgroundColor: terma.BrightCyan}, Width: terma.Cells(4)},
			terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
			terma.Text{Content: "!"},
			terma.Row{
				Spacing:   1,
				MainAlign: terma.MainAxisCenter,
				Width:     terma.Cells(20),
				Style:     terma.Style{BackgroundColor: terma.Green, Padding: terma.EdgeInsetsXY(2, 1)},
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
