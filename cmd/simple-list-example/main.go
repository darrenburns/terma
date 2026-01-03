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

// SimpleListDemo demonstrates the most basic usage of List[string]
// without custom rendering logic or scrollable wrapper.
type SimpleListDemo struct {
	listState   *t.ListState[string]
	selectedMsg *t.Signal[string]
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
				Spans: []t.Span{
					t.PlainSpan("Use "),
					t.BoldSpan("↑/↓", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", t.BrightCyan),
					t.PlainSpan(" to navigate • "),
					t.BoldSpan("Enter", t.BrightCyan),
					t.PlainSpan(" to select"),
				},
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
				Spans: []t.Span{
					t.PlainSpan("Press "),
					t.BoldSpan("Ctrl+C", t.BrightRed),
					t.PlainSpan(" to quit"),
				},
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
