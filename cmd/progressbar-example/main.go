package main

import (
	"fmt"
	"log"

	t "terma"
)

// App demonstrates the ProgressBar widget with smooth Unicode rendering.
type App struct {
	progress t.Signal[float64]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	p := a.progress.Get()

	return t.Column{
		Spacing: 1,
		Style:   t.Style{Padding: t.EdgeInsetsAll(1)},
		Children: []t.Widget{
			t.Text{Content: "=== ProgressBar Widget Demo ===", Style: t.Style{ForegroundColor: theme.Primary, Bold: true}},
			t.Text{Content: "Uses Unicode block characters for smooth sub-cell rendering", Style: t.Style{ForegroundColor: theme.TextMuted}},

			t.Spacer{Height: t.Cells(1)},

			// Interactive progress bar
			t.Text{Content: fmt.Sprintf("Progress: %.0f%%", p*100), Style: t.Style{ForegroundColor: theme.Text}},
			t.ProgressBar{
				Progress:    p,
				Width:       t.Cells(40),
				FilledColor: theme.Primary,
			},

			// Control buttons
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Button{
						ID:    "decrement",
						Label: "- 5%",
						OnPress: func() {
							a.progress.Update(func(v float64) float64 {
								return max(0, v-0.05)
							})
						},
					},
					t.Button{
						ID:    "increment",
						Label: "+ 5%",
						OnPress: func() {
							a.progress.Update(func(v float64) float64 {
								return min(1, v+0.05)
							})
						},
					},
					t.Button{
						ID:    "reset",
						Label: "Reset",
						OnPress: func() {
							a.progress.Set(0)
						},
					},
				},
			},

			t.Spacer{Height: t.Cells(1)},

			// Static examples showing different progress values
			t.Text{Content: "Static Examples:", Style: t.Style{ForegroundColor: theme.Primary}},

			// 0%
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "  0%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.0, Width: t.Cells(30), FilledColor: theme.Primary},
				},
			},

			// 12.5% (tests 1/8 partial)
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 12%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.125, Width: t.Cells(30), FilledColor: theme.Primary},
				},
			},

			// 33%
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 33%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.33, Width: t.Cells(30), FilledColor: theme.Accent},
				},
			},

			// 50%
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 50%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.5, Width: t.Cells(30), FilledColor: theme.Success},
				},
			},

			// 75%
			t.Row{
				Children: []t.Widget{
					t.Text{Content: " 75%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 0.75, Width: t.Cells(30), FilledColor: theme.Warning},
				},
			},

			// 100%
			t.Row{
				Children: []t.Widget{
					t.Text{Content: "100%: ", Width: t.Cells(6)},
					t.ProgressBar{Progress: 1.0, Width: t.Cells(30), FilledColor: theme.Error},
				},
			},

			t.Spacer{Height: t.Cells(1)},

			// Full-width progress bar
			t.Text{Content: "Full-width (Flex):", Style: t.Style{ForegroundColor: theme.Primary}},
			t.ProgressBar{
				Progress:    0.65,
				FilledColor: theme.Accent,
			},

			t.Spacer{},
			t.Text{Content: "Press Tab to navigate, Enter/Space to activate buttons, Ctrl+C to quit", Style: t.Style{ForegroundColor: theme.TextMuted}},
		},
	}
}

func main() {
	app := &App{
		progress: t.NewSignal(0.35),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
