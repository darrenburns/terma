package main

import t "github.com/darrenburns/terma"

type App struct {
	scrollState *t.ScrollState
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	return t.Scrollable{
		ID:     "main-scroll",
		State:  a.scrollState,
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: t.NewGradient(
				t.Hex("#0a0a20"),
				t.Hex("#1a0530"),
				t.Hex("#300510"),
			).WithAngle(135),
		},
		Child: t.Column{
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsAll(2),
			},
			Children: []t.Widget{
				// Header
				t.Text{
					Content: "GRADIENT TEST",
					Style: t.Style{
						ForegroundColor: t.NewGradient(
							t.Hex("#ff0080"),
							t.Hex("#00ffff"),
						).WithAngle(90),
						Bold: true,
					},
				},

				// Row 1: Short texts with different angles
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						box("SHORT", 90, "#ff0000", "#00ff00", "#ff0000", "#00ff00"),
						box("HELLO", 0, "#ffff00", "#ff00ff", "#ffff00", "#ff00ff"),
						box("WORLD", 45, "#00ffff", "#ff8800", "#00ffff", "#ff8800"),
					},
				},

				// Row 2: Medium texts
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						box("MEDIUM TEXT", 90, "#ff0088", "#8800ff", "#ff0088", "#8800ff"),
						box("GRADIENT!", 135, "#00ff88", "#0088ff", "#00ff88", "#0088ff"),
					},
				},

				// Row 3: Different angles and different colors
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						// Different angles: border horizontal, text vertical
						t.Column{
							Width:  t.Cells(22),
							Height: t.Cells(3),
							Style: t.Style{
								Border: t.Border{
									Style: t.BorderRounded,
									Color: t.NewGradient(
										t.Hex("#ff0000"),
										t.Hex("#0000ff"),
									).WithAngle(90), // Horizontal border
								},
							},
							MainAlign: t.MainAxisCenter,
							Children: []t.Widget{
								t.Text{
									Content: "CROSSED ANGLES",
									Style: t.Style{
										ForegroundColor: t.NewGradient(
											t.Hex("#ff0000"),
											t.Hex("#0000ff"),
										).WithAngle(0), // Vertical text
										Bold: true,
									},
								},
							},
						},
						// Different colors entirely
						t.Column{
							Width:  t.Cells(22),
							Height: t.Cells(3),
							Style: t.Style{
								Border: t.Border{
									Style: t.BorderRounded,
									Color: t.NewGradient(
										t.Hex("#ff8800"), // Orange
										t.Hex("#ffff00"), // Yellow
									).WithAngle(90),
								},
							},
							MainAlign: t.MainAxisCenter,
							Children: []t.Widget{
								t.Text{
									Content: "DIFFERENT COLORS",
									Style: t.Style{
										ForegroundColor: t.NewGradient(
											t.Hex("#00ffff"), // Cyan
											t.Hex("#ff00ff"), // Magenta
										).WithAngle(90),
										Bold: true,
									},
								},
							},
						},
					},
				},

				// Rainbow box
				t.Column{
					Width:  t.Cells(50),
					Height: t.Cells(5),
					Style: t.Style{
						Border: t.Border{
							Style: t.BorderRounded,
							Color: t.NewGradient(
								t.Hex("#ff0000"),
								t.Hex("#ff8800"),
								t.Hex("#ffff00"),
								t.Hex("#00ff00"),
								t.Hex("#0088ff"),
								t.Hex("#8800ff"),
							).WithAngle(90),
						},
						Padding: t.EdgeInsetsAll(1),
					},
					Children: []t.Widget{
						t.Text{
							Content: "RAINBOW BORDER WITH RAINBOW TEXT!",
							Style: t.Style{
								ForegroundColor: t.NewGradient(
									t.Hex("#ff0000"),
									t.Hex("#ff8800"),
									t.Hex("#ffff00"),
									t.Hex("#00ff00"),
									t.Hex("#0088ff"),
									t.Hex("#8800ff"),
								).WithAngle(90),
								Bold: true,
							},
						},
						t.Text{
							Content: "Both border and text use 6-color rainbow gradient",
							Style:   t.Style{ForegroundColor: t.Hex("#888888")},
						},
					},
				},

				// Multi-line wrapped text in a box
				t.Column{
					Width:  t.Cells(45),
					Height: t.Cells(8),
					Style: t.Style{
						Border: t.Border{
							Style: t.BorderRounded,
							Color: t.NewGradient(
								t.Hex("#ff00ff"),
								t.Hex("#00ffff"),
							).WithAngle(0),
						},
						Padding: t.EdgeInsetsAll(1),
					},
					Children: []t.Widget{
						t.Text{
							Content: "This is wrapped text with a vertical gradient. The text flows from magenta at the top to cyan at the bottom, matching the border gradient direction.",
							Style: t.Style{
								ForegroundColor: t.NewGradient(
									t.Hex("#ff00ff"),
									t.Hex("#00ffff"),
								).WithAngle(0),
							},
							Wrap: t.WrapSoft,
						},
					},
				},

				// Diagonal everything
				t.Column{
					Width:  t.Cells(45),
					Height: t.Cells(6),
					Style: t.Style{
						Border: t.Border{
							Style: t.BorderSquare,
							Color: t.NewGradient(
								t.Hex("#ffff00"),
								t.Hex("#ff0000"),
							).WithAngle(45),
						},
						BackgroundColor: t.NewGradient(
							t.Hex("#1a1a00"),
							t.Hex("#1a0000"),
						).WithAngle(45),
						Padding: t.EdgeInsetsAll(1),
					},
					Children: []t.Widget{
						t.Text{
							Content: "DIAGONAL EVERYTHING (45°)",
							Style: t.Style{
								ForegroundColor: t.NewGradient(
									t.Hex("#ffff00"),
									t.Hex("#ff0000"),
								).WithAngle(45),
								Bold: true,
							},
						},
						t.Text{
							Content: "Border, background, and text all diagonal",
							Style:   t.Style{ForegroundColor: t.Hex("#888888")},
						},
					},
				},
			},
		},
	}
}

// box creates a small box with gradient border and gradient text
func box(text string, angle float64, fgStart, fgEnd, borderStart, borderEnd string) t.Widget {
	return t.Column{
		Width:  t.Cells(15),
		Height: t.Cells(3),
		Style: t.Style{
			Border: t.Border{
				Style: t.BorderRounded,
				Color: t.NewGradient(
					t.Hex(borderStart),
					t.Hex(borderEnd),
				).WithAngle(angle),
			},
		},
		MainAlign: t.MainAxisCenter,
		Children: []t.Widget{
			t.Text{
				Content: text,
				Style: t.Style{
					ForegroundColor: t.NewGradient(
						t.Hex(fgStart),
						t.Hex(fgEnd),
					).WithAngle(angle),
					Bold: true,
				},
			},
		},
	}
}

func main() {
	_ = t.Run(&App{
		scrollState: t.NewScrollState(),
	})
}
