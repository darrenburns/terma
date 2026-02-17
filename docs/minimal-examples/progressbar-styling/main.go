package main

import (
	"log"

	t "github.com/darrenburns/terma"
)

type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.ProgressBar{
		Progress: 0.6,
		Height:   t.Cells(2),
		Width:    t.Cells(30),
		Style: t.Style{
			Border:  t.RoundedBorder(theme.TextMuted),
			Padding: t.EdgeInsetsXY(1, 0),
		},
		FilledColor:   theme.Accent,
		UnfilledColor: theme.Background,
	}
}

func main() {
	if err := t.Run(&App{}); err != nil {
		log.Fatal(err)
	}
}
