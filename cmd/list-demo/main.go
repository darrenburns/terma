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

// ListDemo demonstrates the List modification APIs.
// Different keys exercise different parts of the ListState API:
//   a - Append item to end
//   p - Prepend item to beginning
//   i - Insert item at cursor position
//   d - Delete item at cursor position
//   c - Clear all items
//   r - Reset to initial items
//   q - Quit
type ListDemo struct {
	listState   *t.ListState[string]
	scrollState *t.ScrollState
	counter     int // For generating unique item names
}

func NewListDemo() *ListDemo {
	return &ListDemo{
		listState: t.NewListState([]string{
			"Apple",
			"Banana",
			"Cherry",
		}),
		scrollState: t.NewScrollState(),
		counter:     3, // Start after initial items
	}
}

func (d *ListDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "a", Name: "Append", Action: func() {
			d.counter++
			d.listState.Append(fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "A", Name: "Append 10", Action: func() {
			for i := 0; i < 10; i++ {
				d.counter++
				d.listState.Append(fmt.Sprintf("Item %d", d.counter))
			}
		}},
		{Key: "p", Name: "Prepend", Action: func() {
			d.counter++
			d.listState.Prepend(fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "i", Name: "Insert at cursor", Action: func() {
			d.counter++
			idx := d.listState.CursorIndex.Peek()
			d.listState.InsertAt(idx, fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "d", Name: "Delete at cursor", Action: func() {
			idx := d.listState.CursorIndex.Peek()
			d.listState.RemoveAt(idx)
		}},
		{Key: "c", Name: "Clear all", Action: func() {
			d.listState.Clear()
		}},
		{Key: "r", Name: "Reset", Action: func() {
			d.listState.SetItems([]string{"Apple", "Banana", "Cherry"})
			d.counter = 3
		}},
	}
}

func (d *ListDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "list-demo-root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "List Modification Demo",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Green,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},

			// Instructions - navigation
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Navigate: "),
					t.BoldSpan("↑/↓", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", t.BrightCyan),
				},
			},

			// Instructions - modifications
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Modify: "),
					t.BoldSpan("a", t.BrightGreen),
					t.PlainSpan("ppend "),
					t.BoldSpan("A", t.BrightGreen),
					t.PlainSpan("ppend10 "),
					t.BoldSpan("p", t.BrightGreen),
					t.PlainSpan("repend "),
					t.BoldSpan("i", t.BrightGreen),
					t.PlainSpan("nsert "),
					t.BoldSpan("d", t.BrightRed),
					t.PlainSpan("elete "),
					t.BoldSpan("c", t.BrightRed),
					t.PlainSpan("lear "),
					t.BoldSpan("r", t.BrightYellow),
					t.PlainSpan("eset"),
				},
			},

			// The list with scrolling
			t.Scrollable{
				ID:           "list-scroll",
				State:        d.scrollState,
				Height:       t.Cells(10),
				DisableFocus: true,
				Style: t.Style{
					Border:  t.RoundedBorder(t.Green, t.BorderTitle("Items")),
					Padding: t.EdgeInsetsXY(1, 0),
				},
				Child: t.List[string]{
					ID:          "demo-list",
					State:       d.listState,
					ScrollState: d.scrollState,
				},
			},

			// Status showing item count
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Items: "),
					t.BoldSpan(fmt.Sprintf("%d", d.listState.ItemCount()), t.BrightYellow),
					t.PlainSpan(" | Cursor: "),
					t.BoldSpan(fmt.Sprintf("%d", d.listState.CursorIndex.Get()+1), t.BrightCyan),
					t.PlainSpan(" | Press "),
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
