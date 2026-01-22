package main

import (
	"log"

	t "terma"
)

type App struct {
	inputState *t.TextInputState
}

func NewApp() *App {
	return &App{inputState: t.NewTextInputState("")}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.TextInput{
		ID:          "username",
		State:       a.inputState,
		Placeholder: "Enter username...",
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
