package main

import (
	"log"

	t "terma"
)

type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.ProgressBar{
		Progress: 0.7,
		Width:    t.Cells(30),
	}
}

func main() {
	if err := t.Run(&App{}); err != nil {
		log.Fatal(err)
	}
}
