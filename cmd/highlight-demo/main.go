package main

import (
	"regexp"

	"terma"
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

// Sample suggestions for autocomplete
var hashtagSuggestions = []terma.Suggestion{
	{Label: "terma", Value: "#terma", Description: "Terma framework"},
	{Label: "golang", Value: "#golang", Description: "Go programming"},
	{Label: "tui", Value: "#tui", Description: "Terminal UI"},
	{Label: "AI", Value: "#AI", Description: "Artificial Intelligence"},
	{Label: "world", Value: "#world", Description: "Hello world"},
	{Label: "topics", Value: "#topics", Description: "Discussion topics"},
	{Label: "demo", Value: "#demo", Description: "Demonstration"},
	{Label: "highlight", Value: "#highlight", Description: "Syntax highlighting"},
}

var mentionSuggestions = []terma.Suggestion{
	{Label: "claude", Value: "@claude", Description: "Claude AI"},
	{Label: "alice", Value: "@alice", Description: "Alice Smith"},
	{Label: "bob", Value: "@bob", Description: "Bob Johnson"},
	{Label: "charlie", Value: "@charlie", Description: "Charlie Brown"},
	{Label: "mentions", Value: "@mentions", Description: "Mention system"},
	{Label: "users", Value: "@users", Description: "All users"},
}

type App struct {
	inputState    *terma.TextInputState
	areaState     *terma.TextAreaState
	codeAreaState *terma.TextAreaState

	// Autocomplete states - one per input since each tracks its own trigger/query
	inputHashtagAc *terma.AutocompleteState
	inputMentionAc *terma.AutocompleteState
	areaHashtagAc  *terma.AutocompleteState
	areaMentionAc  *terma.AutocompleteState
}

func NewApp() *App {
	// Create autocomplete states
	inputHashtagAc := terma.NewAutocompleteState()
	inputHashtagAc.SetSuggestions(hashtagSuggestions)

	inputMentionAc := terma.NewAutocompleteState()
	inputMentionAc.SetSuggestions(mentionSuggestions)

	areaHashtagAc := terma.NewAutocompleteState()
	areaHashtagAc.SetSuggestions(hashtagSuggestions)

	areaMentionAc := terma.NewAutocompleteState()
	areaMentionAc.SetSuggestions(mentionSuggestions)

	inputState := terma.NewTextInputState("Hello #world! Check out @claude for #AI updates.")
	areaState := terma.NewTextAreaState(
		"Welcome to #terma!\n\n" +
			"This demo shows @mentions and #hashtags.\n" +
			"Try typing more #topics or @users.")
	codeAreaState := terma.NewTextAreaState(
		"func main() {\n" +
			"    x := undefined // Error: undefined\n" +
			"    fmt.Println(x)\n" +
			"    _ = unusedVar  // Warning: unused\n" +
			"    return\n" +
			"}")

	// Position cursors at start
	inputState.CursorIndex.Set(0)
	areaState.CursorIndex.Set(0)
	codeAreaState.CursorIndex.Set(0)

	return &App{
		inputState:     inputState,
		areaState:      areaState,
		codeAreaState:  codeAreaState,
		inputHashtagAc: inputHashtagAc,
		inputMentionAc: inputMentionAc,
		areaHashtagAc:  areaHashtagAc,
		areaMentionAc:  areaMentionAc,
	}
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	// Create highlighter for hashtags and mentions
	hashtagColor := terma.RGB(100, 200, 255) // Light blue
	mentionColor := terma.RGB(255, 180, 100) // Orange
	socialHighlighter := NewSocialHighlighter(hashtagColor, mentionColor)

	return terma.Column{
		Spacing: 2,
		Style: terma.Style{
			Padding: terma.EdgeInsetsAll(2),
		},
		Children: []terma.Widget{
			// Title
			terma.Text{
				Content: "Highlighting + Autocomplete Demo",
				Style: terma.Style{
					Bold:            true,
					ForegroundColor: theme.Primary,
				},
			},

			// Section 1: TextInput with highlighting and autocomplete
			terma.Column{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{
						Content: "TextInput with #hashtag and @mention (type # or @ for suggestions):",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					},
					a.buildTextInputWithAutocomplete(ctx, socialHighlighter, hashtagColor, mentionColor),
				},
			},

			// Section 2: TextArea with highlighting and autocomplete
			terma.Column{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{
						Content: "TextArea with #hashtag and @mention (type # or @ for suggestions):",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					},
					a.buildTextAreaWithAutocomplete(ctx, socialHighlighter, hashtagColor, mentionColor),
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
									ForegroundColor: hashtagColor,
								},
							},
							terma.Text{
								Content: "@mention",
								Style: terma.Style{
									Italic:          true,
									ForegroundColor: mentionColor,
								},
							},
						},
					},
					terma.Text{
						Content: "Press Tab/Enter to select, Escape to dismiss, Up/Down to navigate",
						Style:   terma.Style{ForegroundColor: theme.TextMuted, Italic: true},
					},
				},
			},
		},
	}
}

// buildTextInputWithAutocomplete wraps the TextInput with autocomplete for both # and @
func (a *App) buildTextInputWithAutocomplete(ctx terma.BuildContext, highlighter *PatternHighlighter, hashtagColor, mentionColor terma.Color) terma.Widget {
	theme := ctx.Theme()

	return terma.Autocomplete{
		ID:           "social-input-ac",
		State:        a.inputHashtagAc,
		TriggerChars: []rune{'#', '@'},
		MinChars:     0,
		PopupWidth:   terma.Cells(30),
		Child: terma.TextInput{
			ID:          "social-input",
			State:       a.inputState,
			Placeholder: "Type #topics and @users...",
			Highlighter: highlighter,
			Style: terma.Style{
				Width:  terma.Cells(50),
				Border: terma.RoundedBorder(theme.Border),
			},
		},
		OnQueryChange: func(query string) {
			// Filter suggestions based on which trigger character was typed
			text := a.inputState.GetText()
			triggerPos := a.inputHashtagAc.TriggerPosition()
			if triggerPos >= 0 && triggerPos < len([]rune(text)) {
				triggerChar := []rune(text)[triggerPos]
				if triggerChar == '#' {
					a.inputHashtagAc.SetSuggestions(hashtagSuggestions)
				} else if triggerChar == '@' {
					a.inputHashtagAc.SetSuggestions(mentionSuggestions)
				}
			}
		},
		RenderSuggestion: func(s terma.Suggestion, active bool, match terma.MatchResult, ctx terma.BuildContext) terma.Widget {
			return renderSocialSuggestion(s, active, match, ctx, hashtagColor, mentionColor)
		},
	}
}

// buildTextAreaWithAutocomplete wraps the TextArea with autocomplete for both # and @
func (a *App) buildTextAreaWithAutocomplete(ctx terma.BuildContext, highlighter *PatternHighlighter, hashtagColor, mentionColor terma.Color) terma.Widget {
	theme := ctx.Theme()

	return terma.Autocomplete{
		ID:           "social-area-ac",
		State:        a.areaHashtagAc,
		TriggerChars: []rune{'#', '@'},
		MinChars:     0,
		PopupWidth:   terma.Cells(30),
		Child: terma.TextArea{
			ID:          "social-area",
			State:       a.areaState,
			Placeholder: "Write a post with #topics and @mentions...",
			Highlighter: highlighter,
			Style: terma.Style{
				Width:  terma.Cells(50),
				Height: terma.Cells(5),
				Border: terma.RoundedBorder(theme.Border),
			},
		},
		OnQueryChange: func(query string) {
			// Filter suggestions based on which trigger character was typed
			text := a.areaState.GetText()
			triggerPos := a.areaHashtagAc.TriggerPosition()
			if triggerPos >= 0 && triggerPos < len([]rune(text)) {
				triggerChar := []rune(text)[triggerPos]
				if triggerChar == '#' {
					a.areaHashtagAc.SetSuggestions(hashtagSuggestions)
				} else if triggerChar == '@' {
					a.areaHashtagAc.SetSuggestions(mentionSuggestions)
				}
			}
		},
		RenderSuggestion: func(s terma.Suggestion, active bool, match terma.MatchResult, ctx terma.BuildContext) terma.Widget {
			return renderSocialSuggestion(s, active, match, ctx, hashtagColor, mentionColor)
		},
	}
}

// renderSocialSuggestion renders a hashtag or mention suggestion with appropriate styling
func renderSocialSuggestion(s terma.Suggestion, active bool, match terma.MatchResult, ctx terma.BuildContext, hashtagColor, mentionColor terma.Color) terma.Widget {
	theme := ctx.Theme()

	style := terma.Style{Padding: terma.EdgeInsets{Left: 1, Right: 1}}
	textColor := theme.Text
	descColor := theme.TextMuted

	if active {
		style.BackgroundColor = theme.ActiveCursor
		textColor = theme.SelectionText
		descColor = theme.SelectionText
	}

	// Determine if this is a hashtag or mention based on the value
	isHashtag := len(s.Value) > 0 && s.Value[0] == '#'
	isMention := len(s.Value) > 0 && s.Value[0] == '@'

	// Color the prefix character
	var prefixColor terma.Color
	prefix := ""
	if isHashtag {
		prefix = "#"
		prefixColor = hashtagColor
	} else if isMention {
		prefix = "@"
		prefixColor = mentionColor
	}

	if active {
		prefixColor = theme.SelectionText
	}

	children := []terma.Widget{}

	if prefix != "" {
		children = append(children, terma.Text{
			Content: prefix,
			Style:   terma.Style{ForegroundColor: prefixColor, Bold: isHashtag, Italic: isMention},
		})
	}

	children = append(children, terma.Text{
		Content: s.Label,
		Style:   terma.Style{ForegroundColor: textColor},
	})

	if s.Description != "" {
		children = append(children,
			terma.Spacer{Width: terma.Flex(1)},
			terma.Text{
				Content: s.Description,
				Style:   terma.Style{ForegroundColor: descColor},
			},
		)
	}

	return terma.Row{
		Style:    style,
		Children: children,
	}
}

func main() {
	terma.Run(NewApp())
}
