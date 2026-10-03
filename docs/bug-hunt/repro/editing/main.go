package main

import (
	"fmt"
	"os"
	"strings"

	t "github.com/darrenburns/terma"
)

type app struct {
	mode  string
	input *t.TextInputState
	area  *t.TextAreaState
	popup *t.AutocompleteState
}

func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	style := t.Style{Border: t.Border{Style: t.BorderRounded, Color: theme.Primary}, Padding: t.EdgeInsets{Left: 1, Right: 1}}
	var child t.Widget
	var description, value string
	var cursor int
	var triggers []rune
	if a.mode == "selection" {
		child = t.TextArea{ID: "editor", State: a.area, Width: t.Cells(52), Height: t.Cells(5), Style: style}
		description = "Press Enter. The selected word should become a newline."
		value = strings.Join(a.area.Content.Get(), "")
		cursor = a.area.CursorIndex.Get()
		triggers = []rune{'@'}
	} else {
		child = t.TextInput{ID: "editor", State: a.input, Width: t.Cells(52), Style: style}
		value = strings.Join(a.input.Content.Get(), "")
		cursor = a.input.CursorIndex.Get()
		description = "Press Enter to accept accented e, then type !."
		if a.mode == "trigger" {
			description = "Press Enter. Expected text is 👩‍💻 @john without trailing jo."
			triggers = []rune{'@'}
		}
	}
	return t.Column{
		Style:   t.Style{Padding: t.EdgeInsets{Top: 1, Left: 2}, BackgroundColor: theme.Background, ForegroundColor: theme.Text},
		Spacing: 1,
		Children: []t.Widget{
			t.Text{Content: "Autocomplete regression / " + a.mode, Style: t.Style{ForegroundColor: theme.Primary}},
			t.Text{Content: description},
			t.Autocomplete{ID: "autocomplete", State: a.popup, Child: child, TriggerChars: triggers, DismissWhenEmpty: true},
			t.Text{Content: fmt.Sprintf("Text = %+q", value)},
			t.Text{Content: fmt.Sprintf("Cursor grapheme index = %d", cursor)},
		},
	}
}

func main() {
	mode := "unicode"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}
	a := &app{mode: mode, input: t.NewTextInputState(""), area: t.NewTextAreaState("hello world"), popup: t.NewAutocompleteState()}
	a.area.SelectionAnchor.Set(6)
	if mode == "trigger" {
		a.input.SetText("👩‍💻 @jo")
		a.input.CursorEnd()
		a.popup.SetSuggestions([]t.Suggestion{{Label: "john", Value: "@john"}})
	} else if mode == "unicode" {
		a.popup.SetSuggestions([]t.Suggestion{{Label: "accented e", Value: "e\u0301"}})
	}
	if err := t.Run(a); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
