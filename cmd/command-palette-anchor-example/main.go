// Command command-palette-anchor-example demonstrates a searchable tab picker
// attached to a changing tab bar, without measuring widget coordinates.
package main

import (
	"log"

	t "github.com/darrenburns/terma"
)

type app struct {
	palette *t.CommandPaletteState
	shifted t.Signal[bool]
	lowered t.Signal[bool]
	right   t.Signal[bool]
	chosen  t.Signal[string]
}

func newApp() *app {
	a := &app{
		palette: t.NewCommandPaletteState("Open tabs", []t.CommandPaletteItem{
			{Label: "Request one"}, {Label: "Request two"}, {Label: "Request three"},
		}),
		shifted: t.NewSignal(false), lowered: t.NewSignal(false), right: t.NewSignal(false),
		chosen: t.NewSignal("No tab selected"),
	}
	a.palette.Open()
	return a
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	lead, top := 4, 2
	if a.shifted.Get() {
		lead = 14
	}
	if a.lowered.Get() {
		top = 5
	}
	anchor := t.AnchorBottomLeft
	if a.right.Get() {
		anchor = t.AnchorBottomRight
	}
	return t.Column{Style: t.Style{Width: t.Flex(1)}, Children: []t.Widget{
		t.Text{Content: "F2 open/close | F3 move right | F4 move down | F5 left/right align | Ctrl+C quit"},
		t.Spacer{Height: t.Cells(top)},
		t.Row{Style: t.Style{Width: t.Flex(1)}, Children: []t.Widget{
			t.Spacer{Width: t.Cells(lead)},
			t.Text{ID: "tab-bar", Content: " Request one | Request two | Request three ", Style: t.Style{
				Width: t.Flex(1), BackgroundColor: theme.Primary, ForegroundColor: theme.Primary.AutoText(),
			}},
		}},
		t.Spacer{Height: t.Cells(12)},
		t.ComputedText("Selected: Request three", func() string { return "Selected: " + a.chosen.Get() }),
		t.CommandPalette{
			ID: "tab-picker", State: a.palette, AnchorID: "tab-bar", Anchor: anchor,
			Style: t.Style{Width: t.Cells(36)},
			OnSelect: func(item t.CommandPaletteItem) {
				a.chosen.Set(item.Label)
				a.palette.Close()
			},
		},
	}}
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "f2", Name: "Open/close", Action: func() {
			if a.palette.Visible.Get() {
				a.palette.Close()
			} else {
				a.palette.Open()
			}
		}},
		{Key: "f3", Name: "Move right", Action: func() { a.shifted.Set(!a.shifted.Get()) }},
		{Key: "f4", Name: "Move down", Action: func() { a.lowered.Set(!a.lowered.Get()) }},
		{Key: "f5", Name: "Align", Action: func() { a.right.Set(!a.right.Get()) }},
		{Key: "ctrl+c", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	if err := t.Run(newApp()); err != nil {
		log.Fatal(err)
	}
}
