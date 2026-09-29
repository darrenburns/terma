// This demo keeps a keyboard-opened summary visible under a stationary pointer.
package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type app struct {
	visible   t.Signal[bool]
	lastEnter t.Signal[string]
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "s", Name: "Toggle summary", Action: func() { a.visible.Set(!a.visible.Peek()) }},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *app) onHover(event t.HoverEvent) {
	if event.Type != t.HoverEnter {
		return
	}
	source := "layout"
	if event.Source == t.HoverSourcePointer {
		source = "pointer"
		a.visible.Set(false)
	}
	a.lastEnter.Set(source)
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{Children: []t.Widget{
		t.Text{Content: "Hover cause demo: s toggles summary; q quits"},
		t.Text{Content: "Park mouse in the box, press s. Move away and back to dismiss."},
		t.ComputedText("summary-status", func() string {
			return fmt.Sprintf("Summary visible: %t | Last summary enter: %s", a.visible.Get(), a.lastEnter.Get())
		}),
		t.Text{Content: "", Height: t.Cells(1)},
		t.Text{
			ID: "base", Content: "Park mouse here (rows 5-8, columns 1-40)", Width: t.Cells(40), Height: t.Cells(4),
			Style: t.Style{BackgroundColor: theme.Surface, ForegroundColor: theme.Text},
		},
		t.Floating{
			Visible: a.visible.Get(), Config: t.FloatConfig{Offset: t.Offset{X: 0, Y: 4}},
			Child: t.Text{
				ID: "summary", Content: "SUMMARY: layout enter keeps me open", Width: t.Cells(40), Height: t.Cells(4), Hover: a.onHover,
				Style: t.Style{BackgroundColor: theme.Primary, ForegroundColor: theme.TextOnPrimary},
			},
		},
	}}
}

func main() {
	if err := t.Run(&app{visible: t.NewSignal(false), lastEnter: t.NewSignal("none")}); err != nil {
		log.Fatal(err)
	}
}
