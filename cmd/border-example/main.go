package main

import (
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

type BorderDemo struct{}

func (b *BorderDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Border Demo",
				Style: t.Style{
					ForegroundColor: t.BrightWhite,
					BackgroundColor: t.Blue,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Row with square and rounded borders side by side
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Text{
						Content: "Square Border",
						Style: t.Style{
							Border:  t.SquareBorder(t.Cyan),
							Padding: t.EdgeInsetsAll(1),
						},
					},
					t.Text{
						Content: "Rounded Border",
						Style: t.Style{
							Border:  t.RoundedBorder(t.Magenta),
							Padding: t.EdgeInsetsAll(1),
						},
					},
				},
			},

			// Border decorations - titles and subtitles
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Column{
						Style: t.Style{
							Border: t.RoundedBorder(t.Cyan,
								t.BorderTitle("Settings"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{Content: "Option 1: Enabled"},
							t.Text{Content: "Option 2: Disabled"},
						},
					},
					t.Column{
						Style: t.Style{
							Border: t.SquareBorder(t.Yellow,
								t.BorderTitle("Info"),
								t.BorderSubtitle("Press q to quit"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
						Children: []t.Widget{
							t.Text{Content: "Title and subtitle"},
							t.Text{Content: "on the same border"},
						},
					},
				},
			},

			// Centered decorations
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Text{
						Content: "Center title",
						Style: t.Style{
							Border: t.RoundedBorder(t.Green,
								t.BorderTitleCenter("Centered"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
					},
					t.Text{
						Content: "Right aligned",
						Style: t.Style{
							Border: t.SquareBorder(t.Magenta,
								t.BorderTitleRight("Right"),
								t.BorderSubtitleRight("Also Right"),
							),
							Padding: t.EdgeInsetsAll(1),
						},
					},
				},
			},

			// Colored decorations
			t.Column{
				Style: t.Style{
					Border: t.Border{
						Style: t.BorderRounded,
						Color: t.BrightBlack,
						Decorations: []t.BorderDecoration{
							{Text: "Status", Position: t.DecorationTopLeft, Color: t.BrightCyan},
							{Text: "Online", Position: t.DecorationTopRight, Color: t.BrightGreen},
							{Text: "v1.0.0", Position: t.DecorationBottomRight, Color: t.BrightYellow},
						},
					},
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					t.Text{Content: "Decorations can have"},
					t.Text{Content: "individual colors!"},
				},
			},

			// Nested borders with titles
			t.Column{
				ID: "outer-box",
				Style: t.Style{
					Border: t.RoundedBorder(t.BrightBlue,
						t.BorderTitle("Outer"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					t.Text{Content: "Outer container"},
					t.Column{
						ID: "inner-box",
						Style: t.Style{
							Border: t.SquareBorder(t.Red,
								t.BorderTitle("Inner"),
								t.BorderSubtitleCenter("Nested!"),
							),
							BackgroundColor: t.BrightBlack,
							Padding:         t.EdgeInsetsAll(1),
							Margin:          t.EdgeInsetsTRBL(1, 0, 0, 0),
						},
						Children: []t.Widget{
							t.Text{
								Content: "Nested border with title",
								Style:   t.Style{ForegroundColor: t.BrightWhite},
							},
						},
					},
				},
			},

			// Rich text with spans
			t.Column{
				Style: t.Style{
					Border: t.RoundedBorder(t.Cyan,
						t.BorderTitle("Rich Text"),
					),
					Padding: t.EdgeInsetsAll(1),
				},
				Children: []t.Widget{
					// Status line with multiple colored spans
					t.Text{
						Spans: []t.Span{
							t.PlainSpan("Status: "),
							t.ColorSpan("Online", t.Green),
							t.PlainSpan(" | Errors: "),
							t.ColorSpan("3", t.Red),
						},
					},
					// Text with formatting attributes
					t.Text{
						Spans: []t.Span{
							t.PlainSpan("This is "),
							t.BoldSpan("bold", t.BrightWhite),
							t.PlainSpan(", "),
							t.ItalicSpan("italic", t.BrightCyan),
							t.PlainSpan(", and "),
							t.UnderlineSpan("underlined", t.BrightYellow),
							t.PlainSpan(" text."),
						},
					},
					// Fully styled span
					t.Text{
						Spans: []t.Span{
							t.PlainSpan("Mixed: "),
							t.StyledSpan("Bold+Color", t.SpanStyle{
								Foreground: t.Magenta,
								Bold:       true,
							}),
							t.PlainSpan(" and "),
							t.StyledSpan("Italic+Underline", t.SpanStyle{
								Foreground: t.Blue,
								Italic:     true,
								Underline:  true,
							}),
						},
					},
				},
			},
		},
	}
}

func main() {
	app := &BorderDemo{}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
