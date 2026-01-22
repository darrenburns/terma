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
	theme := ctx.Theme()

	return t.TextInput{
		ID:          "styled-input",
		State:       a.inputState,
		Placeholder: "Enter text...",
		Width:       t.Cells(30),
		Style: t.Style{
			Border:          t.RoundedBorder(theme.Primary),
			Padding:         t.EdgeInsetsXY(1, 0),
			BackgroundColor: theme.Surface,
			ForegroundColor: theme.Text,
		},
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
