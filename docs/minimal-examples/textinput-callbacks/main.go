package main

import (
	"fmt"
	"log"

	t "terma"
)

type App struct {
	inputState *t.TextInputState
	charCount  t.Signal[int]
}

func NewApp() *App {
	return &App{
		inputState: t.NewTextInputState(""),
		charCount:  t.NewSignal(0),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.TextInput{
				ID:          "message",
				State:       a.inputState,
				Placeholder: "Type something...",
				OnChange:    func(text string) { a.charCount.Set(len([]rune(text))) },
				OnSubmit:    func(text string) { fmt.Println("Submitted:", text) },
			},
			t.Text{Content: fmt.Sprintf("Characters: %d", a.charCount.Get())},
		},
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
