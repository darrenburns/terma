package main

import (
	"fmt"
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// ListDemo demonstrates the List widget.
// The List widget builds a Column of Text objects internally,
// and integrates with ScrollController for scroll-into-view.
type ListDemo struct {
	controller *t.ScrollController
	selected   *t.Signal[int]
	items      []string
}

func NewListDemo() *ListDemo {
	// Generate items
	items := make([]string, 10000)
	for i := range items {
		items[i] = fmt.Sprintf("List item %d\nClick or use arrow keys to navigate", i+1)
	}

	return &ListDemo{
		controller: t.NewScrollController(),
		selected:   t.NewSignal(0),
		items:      items,
	}
}

func (d *ListDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "r", Name: "Move 2 rows down", Action: func() {
			d.controller.ScrollDown(2)
			d.selected.Update(func(s int) int { return s + 2 })
		}},
		{Key: "l", Name: "Move 2 rows up", Action: func() {
			d.controller.ScrollUp(2)
			d.selected.Update(func(s int) int { return s - 2 })
		}},
	}
}

func (d *ListDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "List Widget Demo",
				Style: t.Style{
					ForegroundColor: t.BrightWhite,
					BackgroundColor: t.Magenta,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Instructions
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Use "),
					t.BoldSpan("↑/↓", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", t.BrightCyan),
					t.PlainSpan(" to navigate • "),
					t.BoldSpan("PgUp/PgDn", t.BrightCyan),
					t.PlainSpan(" for fast scroll • "),
					t.BoldSpan("Home/End", t.BrightCyan),
					t.PlainSpan(" for top/bottom"),
				},
			},

			// The List widget inside a Scrollable
			&t.Scrollable{
				ID:           "list-scroll",
				Controller:   d.controller,
				Height:       t.Cells(12),
				DisableFocus: true, // Let List handle focus
				Style: t.Style{
					Border:  t.RoundedBorder(t.Cyan, t.BorderTitle("Selectable List")),
					Padding: t.EdgeInsetsAll(1),
				},
				Child: &t.List{
					ID:               "demo-list",
					Items:            d.items,
					Selected:         d.selected,
					ScrollController: d.controller,
				},
			},

			// Footer showing current selection
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Selected: "),
					t.BoldSpan(fmt.Sprintf("%d", d.selected.Get()+1), t.BrightYellow),
					t.PlainSpan(" / "),
					t.PlainSpan(fmt.Sprintf("%d", len(d.items))),
					t.PlainSpan(" • Press "),
					t.BoldSpan("Ctrl+C", t.BrightRed),
					t.PlainSpan(" to quit"),
				},
			},
		},
	}
}

func main() {
	app := NewListDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
