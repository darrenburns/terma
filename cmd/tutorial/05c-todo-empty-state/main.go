package main

import (
	"fmt"
	"log"

	t "terma"
)

// App is the root widget for this application.
type App struct {
	listState   *t.ListState[string]
	scrollState *t.ScrollState
	taskCount   int
}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	isEmpty := a.listState.ItemCount() == 0

	emptyMessage := t.Text{
		Content:   "No tasks yet. Press 'a' to add one.",
		Style:     t.Style{ForegroundColor: theme.TextMuted},
		TextAlign: t.TextAlignCenter,
		Width:     t.Flex(1),
	}

	scrollableList := t.Scrollable{
		State:  a.scrollState,
		Height: t.Cells(10),
		Child: t.List[string]{
			ID:          "todo-list",
			State:       a.listState,
			ScrollState: a.scrollState,
		},
	}

	return t.Column{
		Width:      t.Flex(1),
		Height:     t.Flex(1),
		MainAlign:  t.MainAxisCenter,
		CrossAlign: t.CrossAxisCenter,
		Style:      t.Style{BackgroundColor: theme.Background},
		Children: []t.Widget{
			t.Column{
				Width:   t.Cells(40),
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "TODO List", Style: t.Style{Bold: true}},
					t.ShowWhen(isEmpty, emptyMessage),
					t.HideWhen(isEmpty, scrollableList),
					t.ParseMarkupToText("[b $Accent]a[/] add  [b $Accent]d[/] delete  [b $Accent]c[/] clear  [b $Accent]q[/] quit", theme),
				},
			},
		},
	}
}

// Keybinds returns the keyboard shortcuts for this widget.
func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "a", Name: "Add", Action: func() {
			a.taskCount++
			a.listState.Append(fmt.Sprintf("Task %d", a.taskCount))
		}},
		{Key: "d", Name: "Delete", Action: func() {
			if a.listState.ItemCount() > 0 {
				a.listState.RemoveAt(a.listState.CursorIndex.Peek())
			}
		}},
		{Key: "c", Name: "Clear", Action: func() {
			a.listState.Clear()
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	t.SetTheme("catppuccin")

	app := &App{
		listState:   t.NewListState([]string{}),
		scrollState: t.NewScrollState(),
		taskCount:   0,
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
