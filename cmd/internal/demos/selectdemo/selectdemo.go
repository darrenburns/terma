// Package selectdemo exercises committed choice, search, cancellation and edge states.
package selectdemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

type demo struct {
	state    *t.SelectState[string]
	empty    *t.SelectState[string]
	disabled *t.SelectState[string]
	changes  t.Signal[int]
	scroll   *t.ScrollState
}

func (d *demo) Build(ctx t.BuildContext) t.Widget {
	value, set := d.state.Value()
	label := "<unset>"
	if set {
		label = value
	}
	focus := "none"
	if w, ok := ctx.Focused().(t.Identifiable); ok {
		focus = w.WidgetID()
	}
	options := []t.SelectOption[string]{
		{Label: "Development", Value: "dev"},
		{Label: "Production (locked)", Value: "prod", Disabled: true},
		{Label: "Staging", Value: "stage"},
		{Label: "日本語 environment", Value: "jp"},
		{Label: "Café Montréal", Value: "cafe"},
	}
	for i := 1; i <= 15; i++ {
		options = append(options, t.SelectOption[string]{Label: fmt.Sprintf("Preview environment %02d", i), Value: fmt.Sprintf("preview-%02d", i)})
	}
	body := t.Column{Style: t.Style{Padding: t.EdgeInsets{Left: 2, Top: 1}}, Spacing: 1, Children: []t.Widget{
		t.Text{Content: "SelectBox · typed choices", Style: t.Style{Bold: true}},
		t.Text{Content: "Enter opens/commits · arrows navigate · type filters · Esc/Tab cancels"},
		t.SelectBox[string]{ID: "environment", State: d.state, Options: options, Placeholder: "Choose environment", Searchable: true, MaxVisible: 5, Style: t.Style{Width: t.Cells(32)}, OnChange: func(string) { d.changes.Update(func(n int) int { return n + 1 }) }},
		t.Text{Content: fmt.Sprintf("Committed: %s   OnChange calls: %d   Focus: %s", label, d.changes.Get(), focus)},
		t.Button{ID: "clear", Label: "Clear selection", OnPress: d.state.Clear},
		t.SelectBox[string]{ID: "empty", State: d.empty, Placeholder: "Empty options", Style: t.Style{Width: t.Cells(32)}},
		t.SelectBox[string]{ID: "all-disabled", State: d.disabled, Placeholder: "All options disabled", Options: []t.SelectOption[string]{{Label: "Locked one", Value: "one", Disabled: true}, {Label: "Locked two", Value: "two", Disabled: true}}, Style: t.Style{Width: t.Cells(32)}},
		t.Text{Content: "Mouse: click a choice · wheel scrolls · outside click cancels"},
	}}
	return t.Dock{Style: t.Style{BackgroundColor: ctx.Theme().Background}, Top: []t.Widget{demokit.Header{Title: "SelectBox", Tagline: "Choose, filter and commit"}}, Bottom: []t.Widget{demokit.Footer(ctx.Theme())}, Body: t.Scrollable{State: d.scroll, Width: t.Flex(1), Height: t.Flex(1), Child: body}}
}

// Info describes this demo in the gallery.
var Info = demokit.Info{Key: "select", Title: "SelectBox", Description: "Typed choices, search, disabled options and cancellation"}

// New creates a gallery or standalone demo.
func New() demokit.Demo {
	return &demo{state: t.NewSelectState[string](), empty: t.NewSelectState[string](), disabled: t.NewSelectState[string](), changes: t.NewSignal(0), scroll: t.NewScrollState()}
}
func (d *demo) InitialFocus() string { return "environment" }
func (d *demo) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme}}
}

// NewProbe creates the nested-popup verification demo.
func NewProbe() demokit.Demo              { return newProbeDemo() }
func (d *probeDemo) InitialFocus() string { return "first" }
