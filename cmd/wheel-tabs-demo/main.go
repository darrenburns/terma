// wheel-tabs-demo switches a one-row tab bar without a Scrollable wrapper.
package main

import (
	"fmt"
	"log"

	uv "github.com/charmbracelet/ultraviolet"
	t "github.com/darrenburns/terma"
)

type demo struct {
	tabs *t.TabState
	last t.Signal[string]
}

func (d *demo) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "q", Name: "Quit", Action: t.Quit}}
}

func (d *demo) wheel(event t.MouseEvent) bool {
	switch event.Button {
	case uv.MouseWheelDown, uv.MouseWheelRight:
		d.tabs.SelectNext()
	case uv.MouseWheelUp, uv.MouseWheelLeft:
		d.tabs.SelectPrevious()
	default:
		return false
	}
	d.last.Set(fmt.Sprintf("Wheel at screen %d,%d local %d,%d modifiers %v", event.X, event.Y, event.LocalX, event.LocalY, event.Mod))
	return true
}

func (d *demo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
		t.Text{Content: "Scroll over the tab bar to switch tabs. Click or use left/right too. q quits."},
		t.Row{Width: t.Flex(1), Children: []t.Widget{
			t.TabBar{ID: "tabs", State: d.tabs, Width: t.Flex(1), Height: t.Cells(1), MouseWheel: d.wheel},
			t.Text{Content: "|", Height: t.Cells(1)},
		}},
		t.ComputedText("Active tab: three", func() string { return "Active tab: " + d.tabs.ActiveKey() }),
		t.SignalText(d.last, func(value string) string { return value }),
		t.Text{Content: "The tab bar is exactly one row; the right-edge | needs no scrollbar column."},
	}}
}

func main() {
	app := &demo{tabs: t.NewTabState([]t.Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "three", Label: "Three"}}), last: t.NewSignal("No wheel input yet")}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
