// Command long-list-example shows a List of 10,000 items in a Scrollable.
// Scrolling costs what is on screen, not the length of the list.
//
// Run with TERMA_DEBUG_OVERLAY=1 to see per-frame timings.
package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

const itemCount = 10_000

type LongListDemo struct {
	listState *t.ListState[string]
	scroll    *t.ScrollState
	custom    t.Signal[bool]
}

func NewLongListDemo() *LongListDemo {
	items := make([]string, itemCount)
	for i := range items {
		items[i] = fmt.Sprintf("Item %05d", i+1)
	}
	return &LongListDemo{
		listState: t.NewListState(items),
		scroll:    t.NewScrollState(),
		custom:    t.NewSignal(false),
	}
}

func (d *LongListDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "c", Name: "Toggle custom rows", Action: func() { d.custom.Update(func(v bool) bool { return !v }) }},
	}
}

func (d *LongListDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	list := t.List[string]{
		ID:          "long-list",
		State:       d.listState,
		ScrollState: d.scroll,
	}
	mode := "default rows"
	if d.custom.Get() {
		mode = "custom RenderItem rows"
		list.RenderItem = func(item string, active, selected bool) t.Widget {
			style := t.Style{ForegroundColor: theme.Text, Width: t.Flex(1)}
			prefix := "  "
			if active {
				prefix = "▶ "
				style.BackgroundColor = theme.ActiveCursor
				style.ForegroundColor = theme.SelectionText
			}
			return t.Text{Content: prefix + item, Style: style}
		}
	}
	return t.Dock{
		Top: []t.Widget{
			t.ParseMarkupToText(fmt.Sprintf("[b $Primary]%d items[/] · %s · [b $Accent]↑/↓ PgUp/PgDn Home/End[/] to move", itemCount, mode), theme),
		},
		Bottom: []t.Widget{t.KeybindBar{}},
		Body: t.Scrollable{
			State:  d.scroll,
			Height: t.Flex(1),
			Style:  t.Style{Border: t.RoundedBorder(theme.Border)},
			Child:  list,
		},
	}
}

func main() {
	if err := t.Run(NewLongListDemo()); err != nil {
		log.Fatal(err)
	}
}
