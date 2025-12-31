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

// ListViewDemo demonstrates the ScrollController pattern.
// A custom SelectableList widget shares a ScrollController with its
// parent Scrollable, allowing the list to scroll the selected item into view.
type ListViewDemo struct {
	controller *t.ScrollController
	selected   *t.Signal[int]
	items      []string
}

func NewListViewDemo() *ListViewDemo {
	// Generate items
	items := make([]string, 50)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d - This is a selectable list item", i+1)
	}

	return &ListViewDemo{
		controller: t.NewScrollController(),
		selected:   t.NewSignal(0),
		items:      items,
	}
}

func (d *ListViewDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "ScrollController Demo",
				Style: t.Style{
					ForegroundColor: t.BrightWhite,
					BackgroundColor: t.Blue,
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
					t.PlainSpan(" to navigate • Selection auto-scrolls into view"),
				},
			},

			// Scrollable wrapping a SelectableList, both share the same controller.
			// DisableFocus lets the child SelectableList receive focus directly,
			// while still showing the scrollbar for visual feedback.
			&t.Scrollable{
				ID:           "list-scroll",
				Controller:   d.controller,
				Height:       t.Cells(15),
				DisableFocus: true, // Let child handle focus, but keep scrollbar visible
				Style: t.Style{
					Border:  t.RoundedBorder(t.Cyan, t.BorderTitle("Selectable List")),
					Padding: t.EdgeInsetsAll(1),
				},
				Child: &SelectableList{
					ID:               "selectable-list",
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

// SelectableList is a custom widget that renders a list of items with selection.
// It uses a ScrollController to scroll the selected item into view.
type SelectableList struct {
	ID               string
	Items            []string
	Selected         *t.Signal[int]
	ScrollController *t.ScrollController
}

// Key returns the widget's unique identifier.
func (l *SelectableList) Key() string {
	return l.ID
}

// Build returns itself since it handles its own rendering.
func (l *SelectableList) Build(ctx t.BuildContext) t.Widget {
	return l
}

// IsFocusable returns true to allow this widget to receive focus.
func (l *SelectableList) IsFocusable() bool {
	return true
}

// OnKey handles navigation keys, updating selection and scrolling into view.
func (l *SelectableList) OnKey(event t.KeyEvent) bool {
	selected := l.Selected.Peek()
	itemCount := len(l.Items)

	switch {
	case event.MatchString("up", "k"):
		if selected > 0 {
			l.Selected.Set(selected - 1)
			l.scrollSelectedIntoView()
		}
		return true

	case event.MatchString("down", "j"):
		if selected < itemCount-1 {
			l.Selected.Set(selected + 1)
			l.scrollSelectedIntoView()
		}
		return true

	case event.MatchString("home", "g"):
		l.Selected.Set(0)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("end", "G"):
		l.Selected.Set(itemCount - 1)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newSelected := selected - 10
		if newSelected < 0 {
			newSelected = 0
		}
		l.Selected.Set(newSelected)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		newSelected := selected + 10
		if newSelected >= itemCount {
			newSelected = itemCount - 1
		}
		l.Selected.Set(newSelected)
		l.scrollSelectedIntoView()
		return true
	}

	return false
}

// scrollSelectedIntoView uses the ScrollController to ensure
// the selected item is visible in the viewport.
func (l *SelectableList) scrollSelectedIntoView() {
	if l.ScrollController == nil {
		return
	}
	// Each item is 1 line tall
	itemHeight := 1
	itemY := l.Selected.Peek() * itemHeight
	l.ScrollController.ScrollToView(itemY, itemHeight)
}

// Layout computes the size needed for all items.
func (l *SelectableList) Layout(constraints t.Constraints) t.Size {
	// Each item takes 1 line, width is constrained
	height := len(l.Items)
	if height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}
	return t.Size{Width: constraints.MaxWidth, Height: len(l.Items)}
}

// Render draws all items, highlighting the selected one.
func (l *SelectableList) Render(ctx *t.RenderContext) {
	selected := l.Selected.Peek()

	for i, item := range l.Items {
		style := t.Style{ForegroundColor: t.White}
		prefix := "  "

		if i == selected {
			style.ForegroundColor = t.BrightWhite
			style.BackgroundColor = t.Blue
			prefix = "▶ "
		} else if i%2 == 0 {
			style.ForegroundColor = t.BrightBlack
		}

		// Truncate item if needed to fit width
		text := prefix + item
		if len(text) > ctx.Width {
			text = text[:ctx.Width]
		}
		// Pad to full width for consistent background color
		for len(text) < ctx.Width {
			text += " "
		}

		ctx.DrawStyledText(0, i, text, style)
	}
}

func main() {
	app := NewListViewDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
