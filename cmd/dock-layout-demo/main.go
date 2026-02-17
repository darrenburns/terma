package main

import "github.com/darrenburns/terma"

var (
	layoutRed    = terma.RGB(180, 70, 70)
	layoutGreen  = terma.RGB(70, 140, 70)
	layoutBlue   = terma.RGB(70, 100, 180)
	layoutPurple = terma.RGB(140, 70, 140)
	layoutOrange = terma.RGB(180, 120, 50)
)

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Dock{
		Top:    []terma.Widget{terma.Text{Content: "Top", Style: terma.Style{BackgroundColor: layoutRed}}},
		Bottom: []terma.Widget{terma.Text{Content: "Bottom", Style: terma.Style{BackgroundColor: layoutOrange}}},
		Left:   []terma.Widget{terma.Text{Content: "Left", Style: terma.Style{BackgroundColor: layoutGreen}}},
		Right:  []terma.Widget{terma.Text{Content: "Right", Style: terma.Style{BackgroundColor: layoutPurple}}},
		Body:   terma.Text{Content: "Center", Style: terma.Style{BackgroundColor: layoutBlue}},
	}
}

func main() {
	terma.Run(&App{})
}
