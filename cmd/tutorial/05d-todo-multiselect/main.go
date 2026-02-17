package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
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
			MultiSelect: true,
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
				Width:   t.Cells(50),
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "TODO List (Multi-Select)", Style: t.Style{Bold: true}},
					t.ShowWhen(isEmpty, emptyMessage),
					t.HideWhen(isEmpty, scrollableList),
					t.ParseMarkupToText("[b $Accent]space[/] select  [b $Accent]a[/] add  [b $Accent]d[/] delete  [b $Accent]q[/] quit", theme),
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
			indices := a.listState.SelectedIndices()
			if len(indices) > 0 {
				// Delete selected items (in reverse order to preserve indices)
				for i := len(indices) - 1; i >= 0; i-- {
					a.listState.RemoveAt(indices[i])
				}
				a.listState.ClearSelection()
			} else if a.listState.ItemCount() > 0 {
				// No selection - delete item at cursor
				a.listState.RemoveAt(a.listState.CursorIndex.Peek())
			}
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	t.SetTheme("catppuccin")

	app := &App{
		listState: t.NewListState([]string{
			"Task 1",
			"Task 2",
			"Task 3",
			"Task 4",
			"Task 5",
		}),
		scrollState: t.NewScrollState(),
		taskCount:   5,
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
