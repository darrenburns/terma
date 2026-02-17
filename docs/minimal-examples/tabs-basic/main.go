package main

import (
	"log"

	t "github.com/darrenburns/terma"
)

type App struct {
	tabState *t.TabState
}

func NewApp() *App {
	return &App{
		tabState: t.NewTabState([]t.Tab{
			{Key: "home", Label: "Home"},
			{Key: "settings", Label: "Settings"},
			{Key: "help", Label: "Help"},
		}),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.TabBar{
				ID:    "tabs",
				State: a.tabState,
			},
			t.Text{Content: "Active: " + a.tabState.ActiveKey()},
		},
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
