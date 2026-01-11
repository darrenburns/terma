package main

import (
	"fmt"
	"time"

	"terma"
)

type App struct {
	// Spinner demos
	spinnerDots    *terma.SpinnerState
	spinnerLine    *terma.SpinnerState
	spinnerCircle  *terma.SpinnerState
	spinnerBraille *terma.SpinnerState

	// Animated value demo
	progress *terma.AnimatedValue[float64]

	// Animation demo
	colorAnim *terma.Animation[terma.Color]

	// State
	running terma.Signal[bool]
}

func NewApp() *App {
	app := &App{
		spinnerDots:    terma.NewSpinnerState(terma.SpinnerDots),
		spinnerLine:    terma.NewSpinnerState(terma.SpinnerLine),
		spinnerCircle:  terma.NewSpinnerState(terma.SpinnerCircle),
		spinnerBraille: terma.NewSpinnerState(terma.SpinnerBraille),
		progress: terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
			Initial:  0,
			Duration: 500 * time.Millisecond,
			Easing:   terma.EaseOutCubic,
		}),
		colorAnim: terma.NewAnimation(terma.AnimationConfig[terma.Color]{
			From:     terma.RGB(50, 50, 200),
			To:       terma.RGB(200, 50, 50),
			Duration: 2 * time.Second,
			Easing:   terma.EaseInOutSine,
		}),
		running: terma.NewSignal(false),
	}
	return app
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()
	isRunning := a.running.Get()
	progress := a.progress.Get()
	currentColor := a.colorAnim.Value().Get()

	return terma.Dock{
		Bottom: []terma.Widget{
			terma.KeybindBar{
				Style: terma.Style{
					BackgroundColor: theme.Surface,
					Padding:         terma.EdgeInsetsXY(2, 0),
				},
			},
		},
		Body: terma.Column{
			Spacing: 2,
			Width:   terma.Flex(1),
			Height:  terma.Flex(1),
			Style: terma.Style{
				Padding:         terma.EdgeInsetsAll(2),
				BackgroundColor: theme.Background,
			},
			Children: []terma.Widget{
				// Title
				terma.Text{
					Content: "Animation System Demo",
					Style: terma.Style{
						Bold:            true,
						ForegroundColor: theme.Primary,
					},
				},

				// Spinner section
				terma.Column{
					Spacing: 1,
					Children: []terma.Widget{
						terma.Text{
							Content: "Spinners:",
							Style:   terma.Style{ForegroundColor: theme.TextMuted},
						},
						terma.Row{
							Spacing: 4,
							Children: []terma.Widget{
								spinnerWithLabel(a.spinnerDots, "Dots"),
								spinnerWithLabel(a.spinnerLine, "Line"),
								spinnerWithLabel(a.spinnerCircle, "Circle"),
								spinnerWithLabel(a.spinnerBraille, "Braille"),
							},
						},
					},
				},

				// Progress bar section
				terma.Column{
					Spacing: 1,
					Children: []terma.Widget{
						terma.Text{
							Content: "Animated Progress:",
							Style:   terma.Style{ForegroundColor: theme.TextMuted},
						},
						progressBar(progress, theme),
					},
				},

				// Color animation section
				terma.Column{
					Spacing: 1,
					Children: []terma.Widget{
						terma.Text{
							Content: "Color Animation:",
							Style:   terma.Style{ForegroundColor: theme.TextMuted},
						},
						terma.Text{
							Content: "████████████████████",
							Style: terma.Style{
								ForegroundColor: currentColor,
							},
						},
					},
				},

				// Status
				terma.ShowWhen(isRunning, terma.Text{
					Content: "Animations running...",
					Style:   terma.Style{ForegroundColor: theme.Success},
				}),
			},
		},
	}
}

func spinnerWithLabel(state *terma.SpinnerState, label string) terma.Widget {
	return terma.Row{
		Spacing: 1,
		Children: []terma.Widget{
			terma.Spinner{State: state},
			terma.Text{Content: label},
		},
	}
}

func progressBar(value float64, theme terma.ThemeData) terma.Widget {
	const width = 30
	filled := int(value / 100 * width)
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	return terma.Row{
		Spacing: 1,
		Children: []terma.Widget{
			terma.Text{
				Content: bar,
				Style:   terma.Style{ForegroundColor: theme.Primary},
			},
			terma.Text{
				Content: fmt.Sprintf("%.0f%%", value),
			},
		},
	}
}

func (a *App) Keybinds() []terma.Keybind {
	return []terma.Keybind{
		{
			Key:  "space",
			Name: "Toggle",
			Action: func() {
				if a.running.Get() {
					a.spinnerDots.Stop()
					a.spinnerLine.Stop()
					a.spinnerCircle.Stop()
					a.spinnerBraille.Stop()
					a.running.Set(false)
				} else {
					a.spinnerDots.Start()
					a.spinnerLine.Start()
					a.spinnerCircle.Start()
					a.spinnerBraille.Start()
					a.running.Set(true)
				}
			},
		},
		{
			Key:  "+",
			Name: "Increase",
			Action: func() {
				current := a.progress.Target()
				if current < 100 {
					a.progress.Set(current + 10)
				}
			},
		},
		{
			Key:  "-",
			Name: "Decrease",
			Action: func() {
				current := a.progress.Target()
				if current > 0 {
					a.progress.Set(current - 10)
				}
			},
		},
		{
			Key:  "c",
			Name: "Color",
			Action: func() {
				a.colorAnim.Reset()
				a.colorAnim.Start()
			},
		},
		{
			Key:    "q",
			Name:   "Quit",
			Action: terma.Quit,
		},
	}
}

func main() {
	terma.InitLogger()
	app := NewApp()
	if err := terma.Run(app); err != nil {
		panic(err)
	}
}
