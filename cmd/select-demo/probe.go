package main

import (
	"fmt"

	t "github.com/darrenburns/terma"
)

// probeDemo keeps the popup inside a modal and the hints far enough below it
// to inspect retained rendering and outside-click consumption in a browser.
type probeDemo struct {
	first, second   *t.SelectState[int]
	visible         t.Signal[bool]
	clicks, changes t.Signal[int]
}

func newProbeDemo() *probeDemo {
	return &probeDemo{first: t.NewSelectState[int](), second: t.NewSelectState[int](), visible: t.NewSignal(true), clicks: t.NewSignal(0), changes: t.NewSignal(0)}
}
func (d *probeDemo) Build(ctx t.BuildContext) t.Widget {
	value, set := d.first.Value()
	committed := "<unset>"
	if set {
		committed = fmt.Sprint(value)
	}
	return t.Column{Children: []t.Widget{
		t.Button{ID: "reopen", Label: "Open SelectBox probe", OnPress: func() { d.visible.Set(true) }},
		t.Dialog{ID: "select-probe", Visible: d.visible.Get(), Title: "SelectBox nested probe", OnDismiss: func() { d.visible.Set(false) }, Style: t.Style{Width: t.Percent(85)}, Content: t.Column{Children: []t.Widget{
			t.Text{Content: "Click here outside popup", Click: func(t.MouseEvent) { d.clicks.Update(func(n int) int { return n + 1 }) }},
			t.Text{Content: fmt.Sprintf("Committed: %s | Changes: %d | Underlying clicks: %d", committed, d.changes.Get(), d.clicks.Get())},
			t.Row{Spacing: 2, Children: []t.Widget{
				t.SelectBox[int]{ID: "first", State: d.first, Searchable: true, Options: []t.SelectOption[int]{{Label: "Alpha", Value: 1}, {Label: "Locked", Value: 2, Disabled: true}, {Label: "日本語", Value: 3}, {Label: "Café", Value: 4}}, Style: t.Style{Width: t.Cells(20)}, OnChange: func(int) { d.changes.Update(func(n int) int { return n + 1 }) }},
				t.SelectBox[int]{ID: "second", State: d.second, Options: []t.SelectOption[int]{{Label: "Second only", Value: 9}}, Style: t.Style{Width: t.Cells(18)}},
			}},
			t.Text{Content: "Enter opens; Escape cancels; Tab moves within dialog.", Height: t.Cells(9)},
			t.KeybindBar{},
		}}},
	}}
}
