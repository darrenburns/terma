package main

import (
	"fmt"

	"terma"
)

// Sample data for demonstrations
var commands = []terma.Suggestion{
	{Label: "Open File", Value: "open", Icon: "O", Description: "Ctrl+O"},
	{Label: "Save File", Value: "save", Icon: "S", Description: "Ctrl+S"},
	{Label: "Save As", Value: "saveas", Icon: "S", Description: "Ctrl+Shift+S"},
	{Label: "Close File", Value: "close", Icon: "X", Description: "Ctrl+W"},
	{Label: "Exit", Value: "exit", Icon: "Q", Description: "Ctrl+Q"},
	{Label: "Find", Value: "find", Icon: "?", Description: "Ctrl+F"},
	{Label: "Find and Replace", Value: "replace", Icon: "R", Description: "Ctrl+H"},
	{Label: "Go to Line", Value: "goto", Icon: "G", Description: "Ctrl+G"},
	{Label: "Toggle Comment", Value: "comment", Icon: "#", Description: "Ctrl+/"},
	{Label: "Format Document", Value: "format", Icon: "F", Description: "Alt+Shift+F"},
}

var users = []terma.Suggestion{
	{Label: "alice", Value: "@alice", Description: "Alice Smith"},
	{Label: "bob", Value: "@bob", Description: "Bob Johnson"},
	{Label: "charlie", Value: "@charlie", Description: "Charlie Brown"},
	{Label: "diana", Value: "@diana", Description: "Diana Prince"},
	{Label: "eve", Value: "@eve", Description: "Eve Wilson"},
	{Label: "frank", Value: "@frank", Description: "Frank Miller"},
}

var tags = []terma.Suggestion{
	{Label: "bug", Value: "#bug"},
	{Label: "feature", Value: "#feature"},
	{Label: "enhancement", Value: "#enhancement"},
	{Label: "documentation", Value: "#documentation"},
	{Label: "help-wanted", Value: "#help-wanted"},
	{Label: "good-first-issue", Value: "#good-first-issue"},
}

type App struct {
	// Command palette (always-on, fuzzy matching)
	cmdInputState *terma.TextInputState
	cmdAcState    *terma.AutocompleteState
	lastCommand   terma.Signal[string]

	// @mention input (trigger-based)
	mentionInputState *terma.TextAreaState
	mentionAcState    *terma.AutocompleteState
	lastMention       terma.Signal[string]

	// #tag input (trigger-based)
	tagInputState *terma.TextInputState
	tagAcState    *terma.AutocompleteState
	lastTag       terma.Signal[string]
}

func NewApp() *App {
	cmdAc := terma.NewAutocompleteState()
	cmdAc.SetSuggestions(commands)

	mentionAc := terma.NewAutocompleteState()
	mentionAc.SetSuggestions(users)

	tagAc := terma.NewAutocompleteState()
	tagAc.SetSuggestions(tags)

	return &App{
		cmdInputState:     terma.NewTextInputState(""),
		cmdAcState:        cmdAc,
		lastCommand:       terma.NewSignal("(none)"),
		mentionInputState: terma.NewTextAreaState(""),
		mentionAcState:    mentionAc,
		lastMention:       terma.NewSignal("(none)"),
		tagInputState:     terma.NewTextInputState(""),
		tagAcState:        tagAc,
		lastTag:           terma.NewSignal("(none)"),
	}
}

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
				Content: "Autocomplete Widget Demo",
				Style: terma.Style{
					ForegroundColor: theme.Primary,
					Bold:            true,
				},
			},

			// Command Palette Example
			a.buildCommandPaletteSection(ctx),

			// @Mention Example
			a.buildMentionSection(ctx),

			// #Tag Example
			a.buildTagSection(ctx),

			// Help
			terma.Text{
				Content: "Use Tab/Enter to select, Escape to dismiss, Up/Down to navigate",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (a *App) buildCommandPaletteSection(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	return terma.Column{
		Spacing: 1,
		Children: []terma.Widget{
			terma.Text{
				Content: "Command Palette (fuzzy matching, always-on)",
				Style:   terma.Style{ForegroundColor: theme.Accent, Bold: true},
			},
			terma.Autocomplete{
				ID:            "cmd-palette",
				State:         a.cmdAcState,
				MatchMode:     terma.FilterFuzzy,
				Insert:        terma.InsertReplace,
				Style:         terma.Style{Height: terma.Auto},
				AnchorToInput: true,
				Child: terma.TextInput{
					ID:          "cmd-input",
					State:       a.cmdInputState,
					Placeholder: "Type a command...",
					Style: terma.Style{
						Width:           terma.Cells(40),
						Padding:         terma.EdgeInsetsXY(1, 0),
						BackgroundColor: theme.Background,
						Border:          terma.Border{Style: terma.BorderRounded, Color: theme.Border},
					},
				},
				DismissOnBlur: terma.BoolPtr(true),
				OnSelect: func(s terma.Suggestion) {
					a.lastCommand.Set(fmt.Sprintf("%s (%s)", s.Label, s.Value))
					a.cmdInputState.SetText("")
				},
			},
			terma.Text{
				Content: "Last selected: " + a.lastCommand.Get(),
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (a *App) buildMentionSection(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	return terma.Column{
		Spacing: 1,
		Children: []terma.Widget{
			terma.Text{
				Content: "@Mention (trigger-based in TextArea)",
				Style:   terma.Style{ForegroundColor: theme.Accent, Bold: true},
			},
			terma.Autocomplete{
				ID:           "mention-ac",
				State:        a.mentionAcState,
				TriggerChars: []rune{'@'},
				MinChars:     0,
				PopupWidth:   terma.Cells(26),
				Child: terma.TextArea{
					ID:          "mention-input",
					State:       a.mentionInputState,
					Placeholder: "Type @ to mention someone...",
					Style: terma.Style{
						Width:           terma.Cells(40),
						Height:          terma.Cells(3),
						Padding:         terma.EdgeInsetsXY(1, 0),
						BackgroundColor: theme.Background,
						Border:          terma.Border{Style: terma.BorderRounded, Color: theme.Border},
					},
				},
				OnSelect: func(s terma.Suggestion) {
					a.lastMention.Set(s.Label + " (" + s.Description + ")")
				},
			},
			terma.Text{
				Content: "Last mentioned: " + a.lastMention.Get(),
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (a *App) buildTagSection(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()

	return terma.Column{
		Spacing: 1,
		Children: []terma.Widget{
			terma.Text{
				Content: "#Tag (trigger-based)",
				Style:   terma.Style{ForegroundColor: theme.Accent, Bold: true},
			},
			terma.Autocomplete{
				ID:           "tag-ac",
				State:        a.tagAcState,
				TriggerChars: []rune{'#'},
				MinChars:     0,
				Child: terma.TextInput{
					ID:          "tag-input",
					State:       a.tagInputState,
					Placeholder: "Type # to add a tag...",
					Width:       terma.Cells(40),
					Style: terma.Style{
						Padding:         terma.EdgeInsetsXY(1, 0),
						BackgroundColor: theme.Background,
						Border:          terma.Border{Style: terma.BorderRounded, Color: theme.Border},
					},
				},
				OnSelect: func(s terma.Suggestion) {
					a.lastTag.Set(s.Value)
				},
			},
			terma.Text{
				Content: "Last tag: " + a.lastTag.Get(),
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func main() {
	terma.Run(NewApp())
}
