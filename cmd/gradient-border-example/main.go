package main

import (
	"time"

	t "github.com/darrenburns/terma"
)

// Colors for the demo
var (
	bgColor      = t.Hex("#0f172a") // Dark background
	surfaceColor = t.Hex("#0f172a") // Card surface
)

type App struct {
	// Animation for rotating border
	borderAngle *t.Animation[float64]
}

func NewApp() *App {
	app := &App{}

	// Create the angle animation for the rotating border
	app.borderAngle = t.NewAnimation(t.AnimationConfig[float64]{
		From:     0,
		To:       360,
		Duration: 3 * time.Second,
		Easing:   t.EaseLinear,
		OnComplete: func() {
			// Loop: reset and restart
			app.borderAngle.Reset()
			app.borderAngle.Start()
		},
	})

	app.borderAngle.Start()

	return app
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	angle := a.borderAngle.Value().Get()

	return t.Column{
		Width:      t.Flex(1),
		Height:     t.Flex(1),
		Spacing:    3,
		MainAlign:  t.MainAxisCenter,
		CrossAlign: t.CrossAxisCenter,
		Style: t.Style{
			BackgroundColor: bgColor,
			Padding:         t.EdgeInsetsAll(2),
		},
		Children: []t.Widget{
			// Title
			t.Text{
				Content: "GRADIENT BORDER EFFECTS",
				Style: t.Style{
					ForegroundColor: t.NewGradient(
						t.Hex("#8b5cf6"),
						t.Hex("#06b6d4"),
					).WithAngle(90),
					Bold: true,
				},
			},

			// Two cards side by side
			t.Row{
				Spacing:    4,
				CrossAlign: t.CrossAxisCenter,
				Children: []t.Widget{
					// Card 1: Light hitting the top (gradient fades into background)
					lightEffectCard(angle),

					// Card 2: Animated rotating gradient border
					rotatingBorderCard(angle),
				},
			},

			// Description
			t.Text{
				Content: "Left: Border gradient fades into background | Right: Animated rotating border",
				Style: t.Style{
					ForegroundColor: t.Hex("#64748b"),
				},
			},
		},
	}
}

// lightEffectCard creates a card where the border gradient fades from
// a highlight color at the top to the background color, creating the
// effect of light hitting the top of the panel.
func lightEffectCard(angle float64) t.Widget {
	return t.Column{
		//Width:      t.Cells(30),
		//Height:     t.Cells(12),
		MainAlign: t.MainAxisCenter,
		Style: t.Style{
			BackgroundColor: surfaceColor,
			Border: t.Border{
				Style: t.BorderRounded,
				// Subtle gradient with pink and green tints, fading to background
				Color: t.NewGradient(
					t.Hex("#5a4d6b"), // Subtle pink tint
					t.Hex("#4d6b5a"), // Subtle green tint
					bgColor,          // Fades to background
				).WithAngle(angle),
			},
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Pending",
				Style: t.Style{
					ForegroundColor: t.White,
					Bold:            true,
				},
			},
			t.Text{
				Content: "Waiting for a response...",
				Style:   t.Style{ForegroundColor: t.Hex("#94a3b8")},
			},
		},
	}
}

// rotatingBorderCard creates a card with an animated border gradient
// that rotates continuously.
func rotatingBorderCard(angle float64) t.Widget {
	return t.Column{
		Width:      t.Cells(30),
		Height:     t.Cells(12),
		MainAlign:  t.MainAxisCenter,
		CrossAlign: t.CrossAxisCenter,
		Style: t.Style{
			BackgroundColor: surfaceColor,
			Border: t.Border{
				Style: t.BorderRounded,
				// Multi-color gradient with animated angle
				Color: t.NewGradient(
					t.Hex("#8b5cf6"), // Purple
					t.Hex("#06b6d4"), // Cyan
					t.Hex("#10b981"), // Green
					t.Hex("#f59e0b"), // Amber
					t.Hex("#ef4444"), // Red
					t.Hex("#8b5cf6"), // Back to purple for smooth loop
				).WithAngle(angle),
			},
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Rotating Border",
				Style: t.Style{
					ForegroundColor: t.White,
					Bold:            true,
				},
			},
			t.Text{
				Wrap:    t.WrapSoft,
				Content: "The border gradient angle animates continuously",
				Style:   t.Style{ForegroundColor: t.Hex("#94a3b8")},
			},
		},
	}
}

func main() {
	_ = t.Run(NewApp())
}
