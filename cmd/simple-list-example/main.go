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
// This uses the simple pattern where the widget manages its own state.
type SimpleListDemo struct {
	selectedMsg *t.Signal[string]
}

func NewSimpleListDemo() *SimpleListDemo {
	return &SimpleListDemo{
		selectedMsg: t.NewSignal("No selection yet"),
	}
}

func (d *SimpleListDemo) Build(ctx t.BuildContext) t.Widget {
	items := []string{
		"Apple",
		"Banana",
		"Cherry",
		"Date",
		"Elderberry",
	}

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

			// Simple List[string] - widget manages its own state internally
			// No State field needed - cursor is tracked automatically by widget ID
			&t.List[string]{
				ID:    "simple-string-list",
				Items: items, // Just pass items - widget handles cursor state
				OnSelect: func(item string) {
					d.selectedMsg.Set(fmt.Sprintf("You selected: %s", item))
				},
				// No RenderItem - uses default rendering
				// No ScrollController - not wrapped in Scrollable
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
