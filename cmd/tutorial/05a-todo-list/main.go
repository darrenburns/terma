package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

// App is the root widget for this application.
type App struct {
	listState *t.ListState[string]
	taskCount int
}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

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
					t.List[string]{
						ID:    "todo-list",
						State: a.listState,
					},
					t.ParseMarkupToText("[b $Accent]a[/] add  [b $Accent]d[/] delete  [b $Accent]q[/] quit", theme),
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
			a.listState.RemoveAt(a.listState.CursorIndex.Peek())
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
		}),
		taskCount: 3,
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
