package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/x/ansi"
	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// ListViewDemo demonstrates the ScrollController pattern.
// A custom SelectableList widget shares a ScrollController with its
// parent Scrollable, allowing the list to scroll the cursor item into view.
type ListViewDemo struct {
	controller  *t.ScrollController
	cursorIndex *t.Signal[int]
	items       []string
	message     *t.Signal[string]
}

func NewListViewDemo() *ListViewDemo {
	// Generate items
	items := make([]string, 50)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d - This is a selectable list item", i+1)
	}

	return &ListViewDemo{
		controller:  t.NewScrollController(),
		cursorIndex: t.NewSignal(0),
		items:       items,
		message:     t.NewSignal(""),
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
					t.PlainSpan(" to navigate • "),
					t.BoldSpan("Enter", t.BrightCyan),
					t.PlainSpan(" to select • Cursor auto-scrolls into view"),
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
					CursorIndex:      d.cursorIndex,
					ScrollController: d.controller,
					OnSelect: func(item string) {
						d.message.Set(fmt.Sprintf("Selected: %s", item))
					},
				},
			},

			// Footer showing current cursor position and last selection
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Cursor: "),
					t.BoldSpan(fmt.Sprintf("%d", d.cursorIndex.Get()+1), t.BrightYellow),
					t.PlainSpan(" / "),
					t.PlainSpan(fmt.Sprintf("%d", len(d.items))),
					t.PlainSpan(" • "),
					t.PlainSpan(d.message.Get()),
					t.PlainSpan(" • Press "),
					t.BoldSpan("Ctrl+C", t.BrightRed),
					t.PlainSpan(" to quit"),
				},
			},
		},
	}
}

// SelectableList is a custom widget that renders a list of items with cursor navigation.
// It uses a ScrollController to scroll the cursor item into view.
type SelectableList struct {
	ID               string
	Items            []string
	CursorIndex      *t.Signal[int]
	OnSelect         func(item string)
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

// OnKey handles navigation keys and selection, updating cursor position and scrolling into view.
func (l *SelectableList) OnKey(event t.KeyEvent) bool {
	cursorIdx := l.CursorIndex.Peek()
	itemCount := len(l.Items)

	switch {
	case event.MatchString("enter"):
		// Handle selection (Enter key press)
		if l.OnSelect != nil && cursorIdx >= 0 && cursorIdx < len(l.Items) {
			l.OnSelect(l.Items[cursorIdx])
		}
		return true

	case event.MatchString("up", "k"):
		if cursorIdx > 0 {
			l.CursorIndex.Set(cursorIdx - 1)
			l.scrollCursorIntoView()
		}
		return true

	case event.MatchString("down", "j"):
		if cursorIdx < itemCount-1 {
			l.CursorIndex.Set(cursorIdx + 1)
			l.scrollCursorIntoView()
		}
		return true

	case event.MatchString("home", "g"):
		l.CursorIndex.Set(0)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("end", "G"):
		l.CursorIndex.Set(itemCount - 1)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		l.CursorIndex.Set(newCursor)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		newCursor := cursorIdx + 10
		if newCursor >= itemCount {
			newCursor = itemCount - 1
		}
		l.CursorIndex.Set(newCursor)
		l.scrollCursorIntoView()
		return true
	}

	return false
}

// scrollCursorIntoView uses the ScrollController to ensure
// the cursor item is visible in the viewport.
func (l *SelectableList) scrollCursorIntoView() {
	if l.ScrollController == nil {
		return
	}
	// Each item is 1 line tall
	itemHeight := 1
	itemY := l.CursorIndex.Peek() * itemHeight
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

// Render draws all items, highlighting the cursor item.
func (l *SelectableList) Render(ctx *t.RenderContext) {
	cursorIdx := l.CursorIndex.Peek()

	for i, item := range l.Items {
		style := t.Style{ForegroundColor: t.White}
		prefix := "  "

		if i == cursorIdx {
			style.ForegroundColor = t.BrightWhite
			style.BackgroundColor = t.Blue
			prefix = "▶ "
		} else if i%2 == 0 {
			style.ForegroundColor = t.BrightBlack
		}

		// Truncate item if needed to fit width (using display width)
		text := prefix + item
		textWidth := ansi.StringWidth(text)
		if textWidth > ctx.Width {
			text = ansi.Truncate(text, ctx.Width, "")
			textWidth = ctx.Width
		}
		// Pad to full width for consistent background color
		if textWidth < ctx.Width {
			text += strings.Repeat(" ", ctx.Width-textWidth)
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
