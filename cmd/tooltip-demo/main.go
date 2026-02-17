package main

import "github.com/darrenburns/terma"

type App struct{}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	return terma.Column{
		Spacing: 2,
		Style: terma.Style{
			Padding:         terma.EdgeInsetsAll(2),
			BackgroundColor: theme.Background,
		},
		Children: []terma.Widget{
			// Title
			terma.Text{
				Content: "Tooltip Demo (Tab to focus buttons, tooltips show on focus)",
				Style: terma.Style{
					Bold:            true,
					ForegroundColor: theme.Primary,
				},
			},

			// Position demos
			terma.Text{
				Content: "Positions:",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
			terma.Row{
				Spacing: 4,
				Children: []terma.Widget{
					terma.Tooltip{
						Content:  "I appear above!",
						Position: terma.TooltipTop,
						Child: terma.Button{
							ID:    "top-btn",
							Label: "Top",
						},
					},
					terma.Tooltip{
						Content:  "I appear below!",
						Position: terma.TooltipBottom,
						Child: terma.Button{
							ID:    "bottom-btn",
							Label: "Bottom",
						},
					},
					terma.Tooltip{
						Content:  "Left!",
						Position: terma.TooltipLeft,
						Child: terma.Button{
							ID:    "left-btn",
							Label: "Left",
						},
					},
					terma.Tooltip{
						Content:  "Right!",
						Position: terma.TooltipRight,
						Child: terma.Button{
							ID:    "right-btn",
							Label: "Right",
						},
					},
				},
			},

			// Rich text tooltip
			terma.Text{
				Content: "Rich text:",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
			terma.Tooltip{
				Spans: []terma.Span{
					terma.BoldSpan("Ctrl+S"),
					terma.PlainSpan(" to save, "),
					terma.BoldSpan("Ctrl+Q"),
					terma.PlainSpan(" to quit"),
				},
				Child: terma.Button{
					ID:    "shortcuts-btn",
					Label: "Keyboard shortcuts",
				},
			},

			// Custom styled tooltip
			terma.Text{
				Content: "Custom styling:",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
			terma.Row{
				Spacing: 3,
				Children: []terma.Widget{
					terma.Tooltip{
						Content: "Warning!",
						Style: terma.Style{
							BackgroundColor: theme.Warning,
							ForegroundColor: terma.RGB(0, 0, 0),
							Padding:         terma.EdgeInsetsXY(2, 0),
						},
						Child: terma.Button{
							ID:    "warning-btn",
							Label: "Warning style",
						},
					},
					terma.Tooltip{
						Content: "Error occurred",
						Style: terma.Style{
							BackgroundColor: theme.Error,
							ForegroundColor: terma.RGB(255, 255, 255),
							Padding:         terma.EdgeInsetsXY(2, 0),
						},
						Child: terma.Button{
							ID:    "error-btn",
							Label: "Error style",
						},
					},
					terma.Tooltip{
						Content: "Success!",
						Style: terma.Style{
							BackgroundColor: theme.Success,
							ForegroundColor: terma.RGB(0, 0, 0),
							Padding:         terma.EdgeInsetsXY(2, 0),
						},
						Child: terma.Button{
							ID:    "success-btn",
							Label: "Success style",
						},
					},
				},
			},

			// Custom offset
			terma.Text{
				Content: "Custom offset (2 cells gap):",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
			terma.Tooltip{
				Content: "Tooltip with gap",
				Offset:  2,
				Child: terma.Button{
					ID:    "offset-btn",
					Label: "With gap",
				},
			},

			// Footer
			terma.Spacer{},
			terma.Text{
				Content: "Press Ctrl+C to exit",
				Style:   terma.Style{ForegroundColor: theme.TextMuted, Faint: true},
			},
		},
	}
}

func main() {
	terma.Run(&App{})
}
