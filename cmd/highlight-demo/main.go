package main

import (
	"log"
	"regexp"

	"github.com/darrenburns/terma"
)

// PatternHighlighter matches regex patterns and applies styles.
type PatternHighlighter struct {
	Rules []PatternRule
}

type PatternRule struct {
	Pattern *regexp.Regexp
	Style   terma.SpanStyle
}

func (h *PatternHighlighter) Highlight(text string, graphemes []string) []terma.TextHighlight {
	var highlights []terma.TextHighlight

	for _, rule := range h.Rules {
		matches := rule.Pattern.FindAllStringIndex(text, -1)
		for _, match := range matches {
			startByte := match[0]
			endByte := match[1]

			// Convert byte indices to grapheme indices
			bytePos := 0
			startGrapheme := -1
			endGrapheme := -1
			for i, g := range graphemes {
				if bytePos == startByte {
					startGrapheme = i
				}
				bytePos += len(g)
				if bytePos == endByte {
					endGrapheme = i + 1
					break
				}
			}
			if startGrapheme >= 0 && endGrapheme > startGrapheme {
				highlights = append(highlights, terma.TextHighlight{
					Start: startGrapheme,
					End:   endGrapheme,
					Style: rule.Style,
				})
			}
		}
	}
	return highlights
}

// NewSocialHighlighter creates a highlighter for #hashtags and @mentions.
func NewSocialHighlighter(hashtagColor, mentionColor terma.Color) *PatternHighlighter {
	return &PatternHighlighter{
		Rules: []PatternRule{
			{
				Pattern: regexp.MustCompile(`#\w+`),
				Style: terma.SpanStyle{
					Foreground: hashtagColor,
					Bold:       true,
				},
			},
			{
				Pattern: regexp.MustCompile(`@\w+`),
				Style: terma.SpanStyle{
					Foreground: mentionColor,
					Italic:     true,
				},
			},
		},
	}
}

type App struct {
	inputState    *terma.TextInputState
	areaState     *terma.TextAreaState
	codeAreaState *terma.TextAreaState
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	// Create highlighter for hashtags and mentions
	socialHighlighter := NewSocialHighlighter(
		terma.RGB(100, 200, 255), // Light blue for hashtags
		terma.RGB(255, 180, 100), // Orange for mentions
	)

	return terma.Column{
		Spacing: 2,
		Style: terma.Style{
			Padding: terma.EdgeInsetsAll(2),
		},
		Children: []terma.Widget{
			// Title
			terma.Text{
				Content: "Highlighting Demo",
				Style: terma.Style{
					Bold:            true,
					ForegroundColor: theme.Primary,
				},
			},

			// Section 1: TextInput with highlighting
			terma.Column{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{
						Content: "TextInput with #hashtag and @mention highlighting:",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					},
					terma.TextInput{
						ID:          "social-input",
						State:       a.inputState,
						Placeholder: "Type #topics and @users...",
						Highlighter: socialHighlighter,
						Style: terma.Style{
							Width:  terma.Cells(50),
							Border: terma.RoundedBorder(theme.Border),
						},
					},
				},
			},

			// Section 2: TextArea with highlighting
			terma.Column{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{
						Content: "TextArea with #hashtag and @mention highlighting:",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					},
					terma.TextArea{
						ID:          "social-area",
						State:       a.areaState,
						Placeholder: "Write a post with #topics and @mentions...",
						Highlighter: socialHighlighter,
						Style: terma.Style{
							Width:  terma.Cells(50),
							Height: terma.Cells(5),
							Border: terma.RoundedBorder(theme.Border),
						},
					},
				},
			},

			// Section 3: TextArea with line highlighting (like error lines)
			terma.Column{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{
						Content: "TextArea with line highlighting (simulated error/warning):",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					},
					terma.TextArea{
						ID:    "code-area",
						State: a.codeAreaState,
						Style: terma.Style{
							Width:  terma.Cells(50),
							Height: terma.Cells(6),
							Border: terma.RoundedBorder(theme.Border),
						},
						LineHighlights: []terma.LineHighlight{
							// Error on line 2 (0-indexed: line 1)
							{
								StartLine: 1,
								EndLine:   2,
								Style:     terma.Style{BackgroundColor: terma.RGB(80, 30, 30)},
							},
							// Warning on line 4 (0-indexed: line 3)
							{
								StartLine: 3,
								EndLine:   4,
								Style:     terma.Style{BackgroundColor: terma.RGB(80, 70, 20)},
							},
						},
					},
					terma.Text{
						Content: "Line 2 = error (red), Line 4 = warning (yellow)",
						Style: terma.Style{
							ForegroundColor: theme.TextMuted,
							Italic:          true,
						},
					},
				},
			},

			// Legend
			terma.Column{
				Spacing: 0,
				Children: []terma.Widget{
					terma.Text{
						Content: "Legend:",
						Style:   terma.Style{Bold: true, ForegroundColor: theme.Text},
					},
					terma.Row{
						Spacing: 2,
						Children: []terma.Widget{
							terma.Text{
								Content: "#hashtag",
								Style: terma.Style{
									Bold:            true,
									ForegroundColor: terma.RGB(100, 200, 255),
								},
							},
							terma.Text{
								Content: "@mention",
								Style: terma.Style{
									Italic:          true,
									ForegroundColor: terma.RGB(255, 180, 100),
								},
							},
						},
					},
				},
			},
		},
	}
}

func main() {
	app := &App{
		inputState: terma.NewTextInputState("Hello #world! Check out @darrenburns for #Terma updates."),
		areaState: terma.NewTextAreaState(
			"Welcome to #terma!\n\n" +
				"This demo shows @mentions and #hashtags.\n" +
				"Try typing more #topics or @users."),
		codeAreaState: terma.NewTextAreaState(
			"func main() {\n" +
				"    x := undefined // Error: undefined\n" +
				"    fmt.Println(x)\n" +
				"    _ = unusedVar  // Warning: unused\n" +
				"    return\n" +
				"}"),
	}

	// Position cursors at start
	app.inputState.CursorIndex.Set(0)
	app.areaState.CursorIndex.Set(0)
	app.codeAreaState.CursorIndex.Set(0)

	if err := terma.Run(app); err != nil {
		log.Fatal(err)
	}
}
