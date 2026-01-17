package main

import (
	"fmt"

	. "terma"
)

func init() {
	//InitDebug()
}

type App struct {
	scrollState *ScrollState
}

func (a *App) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()

	return Scrollable{
		ID:     "style-demo-scroll",
		State:  a.scrollState,
		Width:  Flex(1),
		Height: Flex(1),
		Child: Column{
			Spacing: 1,
			Style: Style{
				Padding:         EdgeInsetsAll(2),
				BackgroundColor: theme.Background,
			},
			Children: []Widget{
				// Title
				Text{
					Spans: []Span{
						StyledSpan("Text Style Demo", SpanStyle{
							Foreground: theme.Primary,
							Bold:       true,
						}),
					},
				},

				// Basic text attributes
				a.section("Text Attributes", theme, []Widget{
					a.styleRow("Bold", Style{Bold: true}, theme),
					a.styleRow("Faint", Style{Faint: true}, theme),
					a.styleRow("Italic", Style{Italic: true}, theme),
					a.styleRow("Strikethrough", Style{Strikethrough: true}, theme),
					a.styleRow("Reverse", Style{Reverse: true}, theme),
					a.styleRow("Blink", Style{Blink: true}, theme),
					a.styleRow("Conceal (hidden)", Style{Conceal: true}, theme),
				}),

				// Underline styles
				a.section("Underline Styles", theme, []Widget{
					a.styleRow("Single", Style{Underline: UnderlineSingle}, theme),
					a.styleRow("Double", Style{Underline: UnderlineDouble}, theme),
					a.styleRow("Curly", Style{Underline: UnderlineCurly}, theme),
					a.styleRow("Dotted", Style{Underline: UnderlineDotted}, theme),
					a.styleRow("Dashed", Style{Underline: UnderlineDashed}, theme),
				}),

				// Underline with colors
				a.section("Colored Underlines", theme, []Widget{
					a.underlineColorRow("Red underline", theme.Error, theme),
					a.underlineColorRow("Green underline", theme.Success, theme),
					a.underlineColorRow("Yellow underline", theme.Warning, theme),
					a.underlineColorRow("Accent underline", theme.Accent, theme),
				}),

				// Styled underlines with colors
				a.section("Styled Colored Underlines", theme, []Widget{
					a.styledUnderlineRow("Curly red", UnderlineCurly, theme.Error, theme),
					a.styledUnderlineRow("Dotted green", UnderlineDotted, theme.Success, theme),
					a.styledUnderlineRow("Dashed yellow", UnderlineDashed, theme.Warning, theme),
					a.styledUnderlineRow("Double accent", UnderlineDouble, theme.Accent, theme),
				}),

				// Combined styles
				a.section("Combined Styles", theme, []Widget{
					a.styleRow("Bold + Italic", Style{Bold: true, Italic: true}, theme),
					a.styleRow("Bold + Underline", Style{Bold: true, Underline: UnderlineSingle}, theme),
					a.styleRow("Italic + Strikethrough", Style{Italic: true, Strikethrough: true}, theme),
					a.styleRow("Bold + Faint", Style{Bold: true, Faint: true}, theme),
					a.combinedRow("Bold + Curly + Color", Style{
						Bold:           true,
						Underline:      UnderlineCurly,
						UnderlineColor: theme.Error,
					}, theme),
				}),

				// SpanStyle examples using Text widget
				a.section("Rich Text (Spans)", theme, []Widget{
					Text{
						Spans: []Span{
							PlainSpan("Mix "),
							BoldSpan("bold", theme.Primary),
							PlainSpan(", "),
							ItalicSpan("italic", theme.Accent),
							PlainSpan(", and "),
							UnderlineSpan("underline", theme.Success),
							PlainSpan(" in one line"),
						},
					},
					Text{
						Spans: []Span{
							PlainSpan("Use "),
							FaintSpan("faint"),
							PlainSpan(" for subtle text and "),
							StrikethroughSpan("strikethrough", theme.Error),
							PlainSpan(" for deleted"),
						},
					},
				}),

				// Footer
				Text{
					Content: "Press Ctrl+C to exit | Scroll with arrow keys or mouse",
					Style:   Style{ForegroundColor: theme.TextMuted},
				},
			},
		},
	}
}

func (a *App) section(title string, theme ThemeData, children []Widget) Widget {
	return Column{
		Spacing: 0,
		Children: append([]Widget{
			Text{
				Content: title,
				Style: Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},
		}, children...),
	}
}

func (a *App) styleRow(label string, style Style, theme ThemeData) Widget {
	style.ForegroundColor = theme.Text
	return Row{
		Spacing: 2,
		Children: []Widget{
			Text{
				Content: fmt.Sprintf("%-20s", label+":"),
				Style:   Style{ForegroundColor: theme.TextMuted},
			},
			Text{
				Content: "The quick brown fox jumps over the lazy dog",
				Style:   style,
			},
		},
	}
}

func (a *App) underlineColorRow(label string, underlineColor Color, theme ThemeData) Widget {
	return Row{
		Spacing: 2,
		Children: []Widget{
			Text{
				Content: fmt.Sprintf("%-20s", label+":"),
				Style:   Style{ForegroundColor: theme.TextMuted},
			},
			Text{
				Content: "The quick brown fox jumps over the lazy dog",
				Style: Style{
					ForegroundColor: theme.Text,
					Underline:       UnderlineSingle,
					UnderlineColor:  underlineColor,
				},
			},
		},
	}
}

func (a *App) combinedRow(label string, style Style, theme ThemeData) Widget {
	style.ForegroundColor = theme.Text
	return Row{
		Spacing: 2,
		Children: []Widget{
			Text{
				Content: fmt.Sprintf("%-20s", label+":"),
				Style:   Style{ForegroundColor: theme.TextMuted},
			},
			Text{
				Content: "The quick brown fox jumps over the lazy dog",
				Style:   style,
			},
		},
	}
}

func (a *App) styledUnderlineRow(label string, underlineStyle UnderlineStyle, underlineColor Color, theme ThemeData) Widget {
	return Row{
		Spacing: 2,
		Children: []Widget{
			Text{
				Content: fmt.Sprintf("%-20s", label+":"),
				Style:   Style{ForegroundColor: theme.TextMuted},
			},
			Text{
				Content: "The quick brown fox jumps over the lazy dog",
				Style: Style{
					ForegroundColor: theme.Text,
					Underline:       underlineStyle,
					UnderlineColor:  underlineColor,
				},
			},
		},
	}
}

func main() {
	_ = Run(&App{
		scrollState: NewScrollState(),
	})
}
