package main

import "terma"

type App struct {
	width  terma.Signal[int]
	height terma.Signal[int]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Row{
		ID:         "root",
		Spacing:    2,
		Width:      terma.Cells(a.width.Get()),
		Height:     terma.Cells(a.height.Get()),
		CrossAlign: terma.CrossAxisCenter,
		MainAlign:  terma.MainAxisEnd,
		Style: terma.Style{
			BackgroundColor: terma.Red,
		},
		Children: []terma.Widget{
			terma.ParseMarkupToText("[on #f42e41]World[/]", ctx.Theme()),
			terma.Text{Content: "!"},
			terma.Column{
				Width:  terma.Cells(10),
				Height: terma.Cells(6),
				Style:  terma.Style{BackgroundColor: terma.Green},
				Children: []terma.Widget{
					terma.Text{Wrap: terma.WrapHard, Content: "Hello", Width: terma.Cells(4), Style: terma.Style{BackgroundColor: terma.Black}},
				},
			},
		},
	}
}

func (a *App) Keybinds() []terma.Keybind {
	return []terma.Keybind{
		{Key: "s", Name: "Height+", Action: func() { a.height.Set(a.height.Get() + 1) }},
		{Key: "w", Name: "Height-", Action: func() { a.height.Set(max(1, a.height.Get()-1)) }},
		{Key: "d", Name: "Width+", Action: func() { a.width.Set(a.width.Get() + 1) }},
		{Key: "a", Name: "Width-", Action: func() { a.width.Set(max(1, a.width.Get()-1)) }},
	}
}

func main() {
	terma.InitLogger()
	terma.SetDebugLogging(true)
	terma.Run(&App{
		width:  terma.NewSignal(40),
		height: terma.NewSignal(10),
	})
}
