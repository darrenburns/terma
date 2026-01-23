package main

import (
	"log"

	t "terma"
)

// App is the root widget for this application.
type App struct{}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Text{Content: "Hello, Terma!"}
}

func main() {
	app := &App{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
