package main

import (
	"fmt"
	"log"
	"time"

	t "github.com/darrenburns/terma"
)

const maxProgress = 100

// App is the root widget for this application.
type App struct {
	progress *t.AnimatedValue[float64]
}

// Build returns the widget tree that describes the UI.
func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	progress := a.progress.Get()

	// Use Success color when full, Primary otherwise
	fillColor := theme.Primary
	if progress >= maxProgress {
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
					t.Text{Content: fmt.Sprintf("Progress: %.0f/%d", progress, maxProgress)},
					t.ProgressBar{
						Progress:    progress / maxProgress,
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
			if a.progress.Target() < maxProgress {
				a.progress.Set(a.progress.Target() + 20)
			}
		}},
		{Key: "down", Name: "Decrease", Action: func() {
			if a.progress.Target() > 0 {
				a.progress.Set(a.progress.Target() - 20)
			}
		}},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func main() {
	t.SetTheme("catppuccin")

	app := &App{
		progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
			Initial:  0,
			Duration: 300 * time.Millisecond,
			Easing:   t.EaseOutCubic,
		}),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
