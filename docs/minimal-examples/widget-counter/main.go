package main

import (
	"fmt"
	"log"

	"github.com/darrenburns/terma"
)

type App struct {
	count terma.Signal[int]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Column{
		Children: []terma.Widget{
			terma.Text{Content: fmt.Sprintf("Count: %d", a.count.Get())},
			terma.Button{
				ID:    "increment",
				Label: "Increment",
				OnPress: func() {
					a.count.Set(a.count.Get() + 1)
				},
			},
		},
	}
}

func main() {
	if err := terma.Run(&App{count: terma.NewSignal(0)}); err != nil {
		log.Fatal(err)
	}
}
