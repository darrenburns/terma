package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

// App is the root widget for this application.
type App struct {
	count t.Signal[int]
}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{Content: fmt.Sprintf("Count: %d", a.count.Get())},
			t.ParseMarkupToText("Press [b $Accent]Up[/] to increment, [b $Accent]Down[/] to decrement, [b $Accent]q[/] to quit", theme),
		},
	}
}

// Keybinds returns the keyboard shortcuts for this widget.
func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "up", Name: "Increment", Action: func() {
			a.count.Set(a.count.Get() + 1)
		}},
		{Key: "down", Name: "Decrement", Action: func() {
			a.count.Set(a.count.Get() - 1)
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	app := &App{
		count: t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
