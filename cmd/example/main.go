package main

import (
	"fmt"
	"log"
	t "terma"
)

func init() {
	t.InitDebug()
}

// App is the root widget for this application.
type App struct {
	count t.Signal[int]
	name  t.Signal[string]
}

// Build returns the widget tree for this app.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	count := a.count.Get()

	// Width is based on count, minimum 1
	w := max(count, 1)

	return t.Column{
		Children: []t.Widget{
			t.Text{Content: fmt.Sprintf("Hello, %s!", a.name.Get())},
			t.Text{Content: ""},
			// Counter with dynamic width based on count value
			t.Row{
				Width: t.Flex(1),
				Children: []t.Widget{
					t.Text{
						Width:   t.Cells(w),
						Content: fmt.Sprintf("Count: %d (width: %d)", count, w),
						Style:   t.Style{BackgroundColor: t.Blue, ForegroundColor: t.Black},
					},
					t.Text{
						Width:   t.Flex(1),
						Content: "1fr",
						Style:   t.Style{BackgroundColor: t.BrightBlack, ForegroundColor: t.White},
					},
				},
			},
			t.Text{Content: ""},
			t.Text{Content: "Press up/right to grow, down/left to shrink, q to quit"},
		},
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "up", Name: "Grow", Action: func() {
			a.count.Update(func(c int) int { return c + 1 })
		}},
		{Key: "right", Name: "Grow", Action: func() {
			a.count.Update(func(c int) int { return c + 1 })
		}},
		{Key: "left", Name: "Shrink", Action: func() {
			a.count.Update(func(c int) int { return c - 1 })
		}},
		{Key: "down", Name: "Shrink", Action: func() {
			a.count.Update(func(c int) int { return c - 1 })
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	app := &App{
		count: t.NewSignal(30),
		name:  t.NewSignal("World"),
	}

	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
