package main

import (
	"fmt"
	"log"

	t "terma"
)


// SimpleListDemo demonstrates the most basic usage of List[string]
// without custom rendering logic or scrollable wrapper.
type SimpleListDemo struct {
	listState   *t.ListState[string]
	selectedMsg t.Signal[string]
}

func NewSimpleListDemo() *SimpleListDemo {
	return &SimpleListDemo{
		listState: t.NewListState([]string{
			"Apple",
			"Banana",
			"Cherry",
			"Date",
			"Elderberry",
		}),
		selectedMsg: t.NewSignal("No selection yet"),
	}
}

func (d *SimpleListDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "simple-list-root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Simple String List Example",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Cyan,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},

			t.Text{
				Spans: t.ParseMarkup("Use [b #00ffff]↑/↓[/] or [b #00ffff]j/k[/] to navigate • [b #00ffff]Enter[/] to select", t.ThemeData{}),
			},

			// List with state - the state holds items and cursor position
			t.List[string]{
				ID:    "simple-string-list",
				State: d.listState,
				OnSelect: func(item string) {
					d.selectedMsg.Set(fmt.Sprintf("You selected: %s", item))
				},
				// No RenderItem - uses default rendering
				// No ScrollState - not wrapped in Scrollable
			},

			// Display the selection message
			t.Text{
				Content: d.selectedMsg.Get(),
				Style: t.Style{
					ForegroundColor: t.BrightYellow,
				},
			},

			t.Text{
				Spans: t.ParseMarkup("Press [b #ff5555]Ctrl+C[/] to quit", t.ThemeData{}),
			},
		},
	}
}

func main() {
	app := NewSimpleListDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
