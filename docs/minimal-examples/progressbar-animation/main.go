package main

import (
	"log"
	"time"

	t "terma"
)

type App struct {
	anim *t.Animation[float64]
}

func NewApp() *App {
	anim := t.NewAnimation(t.AnimationConfig[float64]{
		From:       0,
		To:         1,
		Duration:   3 * time.Second,
		Easing:     t.EaseInOutSine,
		OnComplete: func() { /* called when animation finishes */ },
	})
	anim.Start()

	return &App{anim: anim}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.ProgressBar{
		Progress: a.anim.Value().Get(),
		Width:    t.Cells(40),
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
