package main

import (
	"fmt"
	"terma"
)

var mainAlignNames = []string{"Start", "Center", "End"}
var crossAlignNames = []string{"Stretch", "Start", "Center", "End"}

type App struct {
	width      terma.Signal[int]
	height     terma.Signal[int]
	mainAlign  terma.Signal[terma.MainAxisAlign]
	crossAlign terma.Signal[terma.CrossAxisAlign]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Dock{
		Style:  terma.Style{BackgroundColor: terma.Blue},
		Bottom: []terma.Widget{terma.KeybindBar{}},
		Body: terma.Column{
			Children: []terma.Widget{
				terma.Text{Content: fmt.Sprintf("Main: %s  Cross: %s  Size: %dx%d",
					mainAlignNames[a.mainAlign.Get()],
					crossAlignNames[a.crossAlign.Get()],
					a.width.Get(),
					a.height.Get(),
				)},
				terma.Row{
					ID:         "root",
					Spacing:    2,
					Width:      terma.Cells(a.width.Get()),
					Height:     terma.Cells(a.height.Get()),
					CrossAlign: a.crossAlign.Get(),
					MainAlign:  a.mainAlign.Get(),
					Style: terma.Style{
						BackgroundColor: terma.Red,
					},
					Children: []terma.Widget{
						terma.Text{Content: "World", Style: terma.Style{BackgroundColor: terma.Blue}},
						terma.Text{Content: "!", Style: terma.Style{BackgroundColor: terma.Cyan}},
						terma.Column{
							Width:  terma.Cells(10),
							Height: terma.Cells(6),
							Style:  terma.Style{BackgroundColor: terma.Green},
							Children: []terma.Widget{
								terma.Text{Wrap: terma.WrapHard, Content: "Hello", Width: terma.Cells(4), Style: terma.Style{BackgroundColor: terma.Black}},
							},
						},
					},
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
		{Key: "m", Name: "MainAlign", Action: func() {
			a.mainAlign.Set((a.mainAlign.Get() + 1) % 3)
		}},
		{Key: "c", Name: "CrossAlign", Action: func() {
			a.crossAlign.Set((a.crossAlign.Get() + 1) % 4)
		}},
	}
}

func main() {
	_ = terma.InitLogger()
	terma.SetDebugLogging(true)
	_ = terma.Run(&App{
		width:      terma.NewSignal(40),
		height:     terma.NewSignal(10),
		mainAlign:  terma.NewSignal(terma.MainAxisEnd),
		crossAlign: terma.NewSignal(terma.CrossAxisCenter),
	})
}
