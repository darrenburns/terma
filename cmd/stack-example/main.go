package main

import (
	"fmt"
	"log"

	t "terma"
)

type App struct {
	showOverlay t.Signal[bool]
	badgeCount  t.Signal[int]
}

func NewApp() *App {
	return &App{
		showOverlay: t.NewSignal(false),
		badgeCount:  t.NewSignal(3),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		Top: []t.Widget{
			t.Text{
				Content: " Stack Widget Examples ",
				Style: t.Style{
					ForegroundColor: theme.Background,
					BackgroundColor: theme.Primary,
				},
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 2,
			Style: t.Style{
				Padding:         t.EdgeInsetsAll(1),
				BackgroundColor: theme.Background,
			},
			Children: []t.Widget{
				// Example 1: Overlapping cards
				a.buildOverlappingCards(ctx),

				// Example 2: Badge with partial overlap
				a.buildBadgeExample(ctx),

				// Example 3: Loading overlay
				a.buildLoadingOverlay(ctx),
			},
		},
	}
}

func (a *App) buildOverlappingCards(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Children: []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[b $Primary]1. Overlapping Cards[/] - Cards stacked with offset", theme),
			},
			t.Stack{
				Width:  t.Cells(50),
				Height: t.Cells(8),
				Style:  t.Style{BackgroundColor: theme.Surface},
				Children: []t.Widget{
					// Back card (bottom layer)
					t.Positioned{
						Top:  t.IntPtr(0),
						Left: t.IntPtr(0),
						Child: t.Column{
							Width:  t.Cells(30),
							Height: t.Cells(6),
							Style: t.Style{
								BackgroundColor: t.NewGradient(theme.Secondary.WithAlpha(0.1), theme.Surface).WithAngle(45),
								Padding:         t.EdgeInsetsXY(2, 1),
							},
							Children: []t.Widget{
								t.Text{Content: "Back Card", Style: t.Style{ForegroundColor: theme.Secondary.AutoText(), Bold: true}},
								t.Text{Content: "I'm behind!", Style: t.Style{ForegroundColor: theme.Secondary.AutoText().WithAlpha(0.6)}},
							},
						},
					},
					// Front card (top layer, offset to show overlap)
					t.Positioned{
						Bottom: t.IntPtr(0),
						Right:  t.IntPtr(0),
						Child: t.Column{
							Width:  t.Cells(30),
							Height: t.Cells(6),
							Style: t.Style{
								BackgroundColor: t.RGB(80, 60, 100).WithAlpha(0.8),
								Padding:         t.EdgeInsetsAll(1),
							},
							Children: []t.Widget{
								t.Text{Content: "Front Card", Style: t.Style{ForegroundColor: t.White}},
								t.Text{Content: "I'm on top!", Style: t.Style{ForegroundColor: t.RGB(180, 180, 180)}},
							},
						},
					},
				},
			},
		},
	}
}

func (a *App) buildBadgeExample(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	count := a.badgeCount.Get()

	return t.Column{
		Children: []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[b $Primary]2. Badge with Overflow[/] - Badge extends beyond card", theme),
			},
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					// Card with notification badge
					t.Stack{
						Height: t.Auto,
						Style:  t.Style{Margin: t.EdgeInsetsXY(0, 1)},
						Children: []t.Widget{
							// The card
							t.Column{
								Style: t.Style{
									Padding:         t.EdgeInsetsXY(2, 1),
									BackgroundColor: t.NewGradient(theme.Surface.Lighten(0.2), theme.Surface).WithAngle(90),
								},
								Children: []t.Widget{
									t.Text{Content: "Inbox", Style: t.Style{ForegroundColor: theme.Text}},
									t.Text{Content: "You have messages", Style: t.Style{ForegroundColor: theme.TextMuted}},
								},
							},
							t.ShowWhen(count > 0, t.Positioned{
								Top:   t.IntPtr(0),  // Above the card border
								Right: t.IntPtr(-1), // Beyond the card border
								Child: t.Text{
									Content: fmt.Sprintf(" %d ", count),
									Style: t.Style{
										ForegroundColor: theme.Success.AutoText(),
										BackgroundColor: theme.Success,
									},
								},
							}),
						},
					},
					// Controls
					t.Column{
						Spacing: 1,
						Children: []t.Widget{
							&t.Button{ID: "add", Label: "+", OnPress: func() {
								a.badgeCount.Set(a.badgeCount.Get() + 1)
							}},
							&t.Button{ID: "sub", Label: "-", OnPress: func() {
								if a.badgeCount.Get() > 0 {
									a.badgeCount.Set(a.badgeCount.Get() - 1)
								}
							}},
						},
					},
				},
			},
		},
	}
}

func (a *App) buildLoadingOverlay(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	showOverlay := a.showOverlay.Get()

	// Build children list
	children := []t.Widget{
		// Content underneath
		t.Column{
			Width:  t.Cells(38),
			Height: t.Cells(5),
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Padding:         t.EdgeInsetsAll(1),
			},
			Children: []t.Widget{
				t.Text{Content: "Data Panel", Style: t.Style{ForegroundColor: theme.Text}},
				t.Text{Content: "User: Alice", Style: t.Style{ForegroundColor: theme.TextMuted}},
				t.Text{Content: "Status: Active", Style: t.Style{ForegroundColor: theme.TextMuted}},
			},
		},
	}

	// Add overlay if enabled
	if showOverlay {
		children = append(children, t.PositionedFill(
			t.Stack{
				Alignment: t.AlignCenter,
				Width:     t.Flex(1),
				Height:    t.Flex(1),
				Style: t.Style{
					BackgroundColor: t.Black.WithAlpha(0.6),
				},
				Children: []t.Widget{
					t.Text{
						Content: " Loading... ",
						Style: t.Style{
							ForegroundColor: theme.Background,
							BackgroundColor: theme.Warning,
						},
					},
				},
			},
		))
	}

	return t.Column{
		Children: []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[b $Primary]3. Loading Overlay[/] - Semi-transparent overlay", theme),
			},
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Stack{
						Width:  t.Cells(40),
						Height: t.Cells(7),
						Style: t.Style{
							Border: t.Border{Style: t.BorderRounded, Color: theme.Primary},
						},
						Children: children,
					},
					&t.Button{
						ID:    "toggle",
						Label: toggleLabel(showOverlay),
						OnPress: func() {
							a.showOverlay.Set(!a.showOverlay.Get())
						},
					},
				},
			},
		},
	}
}

func toggleLabel(show bool) string {
	if show {
		return "Hide"
	}
	return "Show"
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "l", Name: "Toggle Overlay", Action: func() {
			a.showOverlay.Set(!a.showOverlay.Get())
		}},
		{Key: "=", Name: "Add Badge", Action: func() {
			a.badgeCount.Set(a.badgeCount.Get() + 1)
		}},
		{Key: "-", Name: "Remove Badge", Action: func() {
			if a.badgeCount.Get() > 0 {
				a.badgeCount.Set(a.badgeCount.Get() - 1)
			}
		}},
	}
}

func main() {
	t.InitLogger()
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
