package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

const maxProgress = 10

// App is the root widget for this application.
type App struct {
	progress t.Signal[int]
}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	progress := a.progress.Get()

	// Use Success color when full, Primary otherwise
	fillColor := theme.Primary
	if progress == maxProgress {
		fillColor = theme.Success
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
					t.Text{Content: fmt.Sprintf("Progress: %d/%d", progress, maxProgress)},
					t.ProgressBar{
						Progress:    float64(progress) / float64(maxProgress),
						Width:       t.Percent(50),
						FilledColor: fillColor,
					},
					t.ParseMarkupToText("Press [b $Accent]Up[/] to increase, [b $Accent]Down[/] to decrease, [b $Accent]q[/] to quit", theme),
				},
			},
		},
	}
}

// Keybinds returns the keyboard shortcuts for this widget.
func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "up", Name: "Increase", Action: func() {
			if a.progress.Get() < maxProgress {
				a.progress.Set(a.progress.Get() + 1)
			}
		}},
		{Key: "down", Name: "Decrease", Action: func() {
			if a.progress.Get() > 0 {
				a.progress.Set(a.progress.Get() - 1)
			}
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	app := &App{
		progress: t.NewSignal(0),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
