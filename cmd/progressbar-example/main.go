package main

import (
	"fmt"
	"log"
	"time"

	t "github.com/darrenburns/terma"
)

// App demonstrates the ProgressBar widget with smooth Unicode rendering and animations.
type App struct {
	// Interactive progress with smooth animation
	progress *t.AnimatedValue[float64]

	// Auto-cycling progress bar
	autoProgress *t.Animation[float64]
}

func NewApp() *App {
	app := &App{
		progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
			Initial:  0.35,
			Duration: 300 * time.Millisecond,
			Easing:   t.EaseOutCubic,
		}),
	}

	// Create auto-cycling animation that loops
	app.autoProgress = t.NewAnimation(t.AnimationConfig[float64]{
		From:     0,
		To:       1,
		Duration: 3 * time.Second,
		Easing:   t.EaseInOutSine,
		OnComplete: func() {
			// Loop the animation
			app.autoProgress.Reset()
			app.autoProgress.Start()
		},
	})
	app.autoProgress.Start() // Safe to call before app runs

	return app
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	p := a.progress.Get()
	autoP := a.autoProgress.Value().Get()

	return t.Column{
		Spacing: 1,
		Width:   t.Flex(1),
		Style:   t.Style{Padding: t.EdgeInsetsAll(1), BackgroundColor: theme.Background},
		Children: []t.Widget{
			t.Text{Content: "ProgressBar Widget Demo", Style: t.Style{ForegroundColor: theme.Primary, Bold: true}},
			t.Text{Content: "Smooth Unicode rendering with animation support", Style: t.Style{ForegroundColor: theme.TextMuted}},

			// Auto-cycling animated progress bar
			t.Text{Content: "Auto-cycling (loops continuously):", Style: t.Style{ForegroundColor: theme.Accent}},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: fmt.Sprintf(" %.0f%%", autoP*100), Width: t.Cells(6)},
					t.ProgressBar{
						Progress:    autoP,
						Width:       t.Flex(1),
						FilledColor: theme.Accent,
					},
				},
			},

			// Interactive progress bar with smooth animation
			t.Text{Content: "Interactive (use buttons or +/- keys):", Style: t.Style{ForegroundColor: theme.Secondary}},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: fmt.Sprintf(" %.0f%%", p*100), Width: t.Cells(6)},
					t.ProgressBar{
						Progress:    p,
						Width:       t.Flex(1),
						FilledColor: theme.Secondary,
					},
				},
			},

			// Control buttons
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Button{
						ID:    "decrement",
						Label: "-10%",
						OnPress: func() {
							current := a.progress.Target()
							a.progress.Set(max(0, current-0.1))
						},
					},
					t.Button{
						ID:    "increment",
						Label: "+10%",
						OnPress: func() {
							current := a.progress.Target()
							a.progress.Set(min(1, current+0.1))
						},
					},
					t.Button{
						ID:    "reset",
						Label: "Reset",
						OnPress: func() {
							a.progress.Set(0)
						},
					},
					t.Button{
						ID:    "fill",
						Label: "Fill",
						OnPress: func() {
							a.progress.Set(1)
						},
					},
				},
			},

			// Static examples showing different progress values and colors
			t.Text{Content: "Static examples:", Style: t.Style{ForegroundColor: theme.TextMuted}},

			t.Row{
				Children: []t.Widget{
					t.Text{Content: "  0%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.0, Width: t.Cells(30), FilledColor: theme.Primary},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 12%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.125, Width: t.Cells(30), FilledColor: theme.Primary},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 50%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.5, Width: t.Cells(30), FilledColor: theme.Success},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 75%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.75, Width: t.Cells(30), FilledColor: theme.Warning},
				},
			},
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "100%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 1.0, Width: t.Cells(30), FilledColor: theme.Error},
				},
			},
		},
	}
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{
			Key:  "+",
			Name: "+10%",
			Action: func() {
				current := a.progress.Target()
				a.progress.Set(min(1, current+0.1))
			},
		},
		{
			Key:  "-",
			Name: "-10%",
			Action: func() {
				current := a.progress.Target()
				a.progress.Set(max(0, current-0.1))
			},
		},
		{
			Key:    "q",
			Name:   "Quit",
			Action: t.Quit,
		},
	}
}

func main() {
	t.SetTheme(t.ThemeNameRosePine)
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
