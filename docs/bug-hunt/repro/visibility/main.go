package main

import (
	"fmt"
	t "github.com/darrenburns/terma"
	"log"
)

type app struct {
	visible t.Signal[bool]
	presses t.Signal[int]
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "v", Name: "Toggle visibility", Action: func() { a.visible.Update(func(v bool) bool { return !v }) }}}
}
func (a *app) Build(ctx t.BuildContext) t.Widget {
	return t.Column{Spacing: 1, Style: t.Style{Padding: t.EdgeInsets{Top: 1, Left: 2}}, Children: []t.Widget{
		t.Text{Content: "R1: VisibleWhen loses content and reserved space"},
		t.Text{Content: fmt.Sprintf("Visible: %v | button presses: %d | press v to toggle", a.visible.Get(), a.presses.Get())},
		t.Text{Content: "The bordered panel should appear below. When hidden, its space remains."},
		t.VisibleWhen(a.visible.Get(), t.Column{Style: t.Style{Width: t.Cells(48), Padding: t.EdgeInsets{Left: 1, Right: 1}, Border: t.RoundedBorder(t.Cyan)}, Children: []t.Widget{
			t.Text{Content: "VISIBLE CONTENT"},
			t.Text{Content: "This panel must reserve three content rows."},
			t.Button{ID: "proof", Label: "Count click", OnPress: func() { a.presses.Update(func(n int) int { return n + 1 }) }},
		}}),
		t.Text{Content: "FOOTER: stays on the same row while toggling visibility"},
	}}
}
func main() {
	if err := t.Run(&app{visible: t.NewSignal(true), presses: t.NewSignal(0)}); err != nil {
		log.Fatal(err)
	}
}
