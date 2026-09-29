// float-passthrough-demo puts a preview over a button and scrollable rows.
// Run with -blocking to compare the default interactive overlay behavior.
package main

import (
	"flag"
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type app struct {
	blocking bool
	clicks   t.Signal[int]
	hovered  t.Signal[string]
	scroll   *t.ScrollState
}

func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "q", Name: "Quit", Action: t.Quit}}
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	rows := make([]t.Widget, 20)
	for i := range rows {
		rows[i] = t.Text{Content: fmt.Sprintf("row %02d                           visible row %02d", i, i), Width: t.Cells(60)}
	}
	mode := "pass-through"
	if a.blocking {
		mode = "blocking"
	}
	return t.Column{Children: []t.Widget{
		t.Text{Content: "Float pointer demo: " + mode},
		t.Text{Content: "Click preview to press button; scroll preview; q quits."},
		t.ComputedText("Clicks: 999 | Hover: button | Scroll: 999", func() string {
			return fmt.Sprintf("Clicks: %d | Hover: %s | Scroll: %d", a.clicks.Get(), a.hovered.Get(), a.scroll.Offset.Get())
		}),
		t.Button{
			ID: "under-button", Label: "Underlying button", Width: t.Cells(60),
			OnPress: func() { a.clicks.Set(a.clicks.Peek() + 1) },
			Hover: func(ev t.HoverEvent) {
				if ev.Type == t.HoverEnter {
					a.hovered.Set("button")
				} else {
					a.hovered.Set("none")
				}
			},
		},
		t.Scrollable{ID: "under-scroll", State: a.scroll, Width: t.Cells(60), Height: t.Cells(6), Child: t.Column{Children: rows}},
		t.Floating{
			Visible: true,
			Config:  t.FloatConfig{Position: t.FloatPositionAbsolute, Offset: t.Offset{Y: 3}, PointerPassthrough: !a.blocking},
			Child:   t.Text{ID: "preview", Content: "PREVIEW (covers button)\nScroll here reaches rows", Width: t.Cells(28), Height: t.Cells(6), Style: t.Style{BackgroundColor: theme.Surface, ForegroundColor: theme.Text}},
		},
	}}
}

func main() {
	blocking := flag.Bool("blocking", false, "use default blocking float behavior")
	flag.Parse()
	a := &app{blocking: *blocking, clicks: t.NewSignal(0), hovered: t.NewSignal("none"), scroll: t.NewScrollState()}
	if err := t.Run(a); err != nil {
		log.Fatal(err)
	}
}
