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

// ListItem represents a single item in our list.
// This is pure data - rendering is handled by RenderItem.
type ListItem struct {
	Title       string
	Description string
}

// ListDemo demonstrates the List widget with custom data types.
// The List widget builds a Column of widgets internally,
// and integrates with ScrollController for scroll-into-view.
type ListDemo struct {
	controller *t.ScrollController
	selected   *t.Signal[int]
	items      []ListItem
}

func NewListDemo() *ListDemo {
	// Generate sample items
	items := make([]ListItem, 100)
	for i := range items {
		items[i] = ListItem{
			Title:       fmt.Sprintf("Item %d", i+1),
			Description: "Use arrow keys to navigate",
		}
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
				Child: &t.List[ListItem]{
					ID:               "demo-list",
					Items:            d.items,
					Selected:         d.selected,
					ScrollController: d.controller,
					// Describe how to render each item in the list as a widget
					RenderItem: func(item ListItem, selected bool) t.Widget {
						// Style the title based on selection state
						titleStyle := t.Style{ForegroundColor: t.White}
						if selected {
							titleStyle.ForegroundColor = t.Magenta
						}

						// Each item is 2 lines tall (title + description)
						return t.Column{
							Height: t.Cells(2),
							Children: []t.Widget{
								t.Text{
									Content: item.Title,
									Style:   titleStyle,
									Width:   t.Fr(1),
								},
								t.Text{
									Content: item.Description,
									Style:   t.Style{ForegroundColor: t.BrightBlack},
									Width:   t.Fr(1),
								},
							},
						}
					},
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
