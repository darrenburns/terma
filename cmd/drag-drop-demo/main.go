package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type app struct {
	locations t.AnySignal[map[string]string]
	last      t.Signal[string]
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "q", Name: "Quit", Action: t.Quit}}
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	locations := a.locations.Get()
	columns := []t.Widget{}
	for _, lane := range []string{"Ready", "Working", "Done"} {
		cards := []t.Widget{t.Text{Content: lane, Style: t.Style{Bold: true, ForegroundColor: theme.Accent}}}
		for _, name := range []string{"Write docs", "Add examples", "Review tests"} {
			if locations[name] != lane {
				continue
			}
			cards = append(cards, t.Draggable[string]{ID: "card-" + name, Payload: name, Child: t.Column{
				Style:    t.Style{Width: t.Cells(23), Padding: t.EdgeInsetsXY(1, 1), Border: t.RoundedBorder(theme.Primary), BackgroundColor: theme.Surface},
				Children: []t.Widget{t.Text{Content: name}, t.Text{Content: "Drag me to a column", Style: t.Style{ForegroundColor: theme.TextMuted}}},
			}})
		}
		columns = append(columns, t.DropTarget[string]{ID: "lane-" + lane, Accept: func(name string) bool { return a.locations.Peek()[name] != lane },
			OnDrop: func(name string) {
				a.locations.Update(func(old map[string]string) map[string]string {
					next := make(map[string]string, len(old))
					for key, value := range old {
						next[key] = value
					}
					next[name] = lane
					return next
				})
				a.last.Set(fmt.Sprintf("Moved %s to %s", name, lane))
			}, Child: t.Column{Style: t.Style{Width: t.Cells(27), Height: t.Cells(19), Padding: t.EdgeInsetsAll(1), Border: t.RoundedBorder(theme.Border)}, Spacing: 1, Children: cards}})
	}
	return t.Column{Style: t.Style{Padding: t.EdgeInsetsAll(1), BackgroundColor: theme.Background}, Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Drag and drop", Style: t.Style{Bold: true, ForegroundColor: theme.Primary}},
		t.Text{Content: "Drag cards between columns. Escape cancels. Release outside a column to return."},
		t.Row{Spacing: 1, Children: columns},
		t.SignalText(a.last, func(value string) string { return value }),
		t.KeybindBar{},
	}}
}

func main() {
	a := &app{locations: t.NewAnySignal(map[string]string{"Write docs": "Ready", "Add examples": "Ready", "Review tests": "Working"}), last: t.NewSignal("Pick up a card by any part of its body.")}
	if err := t.Run(a); err != nil {
		log.Fatal(err)
	}
}
