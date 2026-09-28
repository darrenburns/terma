// Package textinputdemo demonstrates the single-line TextInput widget.
package textinputdemo

import (
	"fmt"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "text-input",
	Title:       "TextInput",
	Description: "A form of single-line inputs with highlighting and read-only mode",
}

// TextInputDemo shows off the single-line TextInput widget: a small form with
// placeholders, live highlighting, change tracking, read-only mode and
// submission. The sidebar reads the focused field's TextInputState live.
//
//	tab / shift+tab - Move between fields and buttons
//	enter           - Submit the form (from any field)
//	←→ home end     - Move the cursor (ctrl+←→ or alt+b/f by word)
//	shift+←→        - Extend the selection (ctrl+a selects all)
//	ctrl+u ctrl+k   - Delete to start / end of the field
//	ctrl+w          - Delete the previous word
//	ctrl+r          - Toggle read-only on every field
//	ctrl+x          - Clear the form (ignored while read-only)
//	escape          - Clear the focused field (ignored while read-only)
type TextInputDemo struct {
	fields []*field

	readOnly    t.Signal[bool]
	edits       t.Signal[int]
	submissions t.Signal[int]
	submitted   t.AnySignal[[]string] // Field values from the last submit, nil if none
}

// field describes one labelled input in the form.
type field struct {
	id          string
	label       string
	icon        string
	placeholder string
	state       *t.TextInputState
	highlight   func(theme t.ThemeData) t.Highlighter
}

// New creates the demo.
func New() demokit.Demo {
	return &TextInputDemo{
		fields: []*field{
			{id: "name-input", label: "Name", icon: "◆", placeholder: "Ada Lovelace", state: t.NewTextInputState("")},
			{id: "email-input", label: "Email", icon: "@", placeholder: "ada@example.com", state: t.NewTextInputState(""), highlight: emailHighlighter},
			{id: "message-input", label: "Message", icon: "✎", placeholder: "Say hi — #tags and @mentions are highlighted", state: t.NewTextInputState(""), highlight: messageHighlighter},
		},
		readOnly:    t.NewSignal(false),
		edits:       t.NewSignal(0),
		submissions: t.NewSignal(0),
		submitted:   t.NewAnySignal[[]string](nil),
	}
}

func (d *TextInputDemo) InitialFocus() string { return d.fields[0].id }

func (d *TextInputDemo) submit() {
	values := make([]string, len(d.fields))
	for i, f := range d.fields {
		values[i] = f.state.GetText()
	}
	d.submitted.Set(values)
	d.submissions.Update(func(n int) int { return n + 1 })
}

func (d *TextInputDemo) clearForm() {
	if d.readOnly.Peek() {
		return
	}
	for _, f := range d.fields {
		f.state.SetText("")
	}
	t.RequestFocus(d.fields[0].id)
}

func (d *TextInputDemo) toggleReadOnly() {
	readOnly := !d.readOnly.Peek()
	d.readOnly.Set(readOnly)
	for _, f := range d.fields {
		f.state.ReadOnly.Set(readOnly)
	}
}

func (d *TextInputDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+r", Name: "Read-only", Action: d.toggleReadOnly},
		{Key: "ctrl+x", Name: "Clear form", Action: d.clearForm},
	}
}

func (d *TextInputDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	formChildren := make([]t.Widget, 0, len(d.fields)+2)
	for _, f := range d.fields {
		formChildren = append(formChildren, fieldPanel{demo: d, field: f})
	}
	formChildren = append(formChildren,
		buttonRow{demo: d},
		demokit.Fill(submittedPanel{demo: d}),
	)

	return t.Dock{
		ID: "text-input-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			header{demo: d},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: []t.Widget{
				t.Column{
					Width:    t.Flex(1),
					Height:   t.Flex(1),
					Spacing:  1,
					Children: formChildren,
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{demo: d},
						keysPanel{},
					},
				},
			},
		},
	}
}

const sidebarWidth = 32

// header is the title bar across the top of the screen. It reads the
// read-only signal itself, so toggling it only rebuilds the header.
type header struct {
	demo *TextInputDemo
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	mode := "[$TextMuted]mode[/] [b $Success]editable[/]"
	if h.demo.readOnly.Get() {
		mode = "[$TextMuted]mode[/] [b $Warning]read-only[/]"
	}
	return demokit.Header{Title: "TextInput Playground", Tagline: "Single-line text entry", Right: mode}
}

// fieldPanel is a bordered TextInput. The border lights up while it is focused.
type fieldPanel struct {
	demo  *TextInputDemo
	field *field
}

func (p fieldPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d, f := p.demo, p.field

	input := t.TextInput{
		ID:          f.id,
		State:       f.state,
		Placeholder: f.placeholder,
		Style: t.Style{
			Width:           t.Flex(1),
			ForegroundColor: theme.Text,
		},
		OnChange: func(string) {
			d.edits.Update(func(n int) int { return n + 1 })
		},
		OnSubmit: func(string) { d.submit() },
		ExtraKeybinds: []t.Keybind{
			{Key: "escape", Name: "Clear field", Action: func() {
				if !f.state.ReadOnly.Peek() {
					f.state.SetText("")
				}
			}, Hidden: true},
		},
	}
	if f.highlight != nil {
		input.Highlighter = f.highlight(theme)
	}

	title := f.label
	if d.readOnly.Get() {
		title += " · read-only"
	}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, title, ctx.IsFocused(input)),
		Children: []t.Widget{
			t.ParseMarkupToText(fmt.Sprintf("[$Accent]%s[/]", f.icon), theme),
			input,
		},
	}
}

// buttonRow holds the form's buttons.
type buttonRow struct {
	demo *TextInputDemo
}

func (b buttonRow) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	idle := t.Style{BackgroundColor: theme.Surface, ForegroundColor: theme.Text}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 2,
		Style:   t.Style{Padding: t.EdgeInsetsXY(1, 0)},
		Children: []t.Widget{
			// Button.Style applies only while unfocused, so the variant colours
			// double as a clear focus indicator.
			t.Button{ID: "submit-btn", Label: " Submit ", Variant: t.ButtonPrimary, Style: idle, OnPress: b.demo.submit},
			t.Button{ID: "clear-btn", Label: " Clear ", Variant: t.ButtonWarning, Style: idle, OnPress: b.demo.clearForm},
		},
	}
}

// submittedPanel shows the values captured by the last submit.
type submittedPanel struct {
	demo *TextInputDemo
}

func (s submittedPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	values := d.submitted.Get()

	var children []t.Widget
	if values == nil {
		children = []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[$TextMuted]Nothing submitted yet. Fill in the form and press [b $Primary]enter[/].[/]", theme),
				Wrap:  t.WrapSoft,
			},
		}
	} else {
		for i, f := range d.fields {
			children = append(children, submittedRow(theme, f.label, values[i]))
		}
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "Last submission", false),
		Children: children,
	}
}

func submittedRow(theme t.ThemeData, label, value string) t.Widget {
	valueText := t.Text{
		Content: value,
		Wrap:    t.WrapSoft,
		Width:   t.Flex(1),
		Style:   t.Style{ForegroundColor: theme.Success},
	}
	if value == "" {
		valueText.Content = "(empty)"
		valueText.Style.ForegroundColor = theme.TextMuted
	}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.Text{Content: label, Width: t.Cells(8), Style: t.Style{ForegroundColor: theme.TextMuted}},
			valueText,
		},
	}
}

// statePanel reads the focused field's TextInputState live. It subscribes to
// the signals itself, so typing only rebuilds this panel and the input.
type statePanel struct {
	demo *TextInputDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	focusName := "—"
	var focused *field
	if w := ctx.Focused(); w != nil {
		if id, ok := w.(t.Identifiable); ok {
			focusName = id.WidgetID()
			for _, f := range d.fields {
				if f.id == focusName {
					focused = f
					focusName = f.label
				}
			}
			switch focusName {
			case "submit-btn":
				focusName = "Submit"
			case "clear-btn":
				focusName = "Clear"
			}
		}
	}

	length, cursor, selection := "—", "—", "—"
	if focused != nil {
		content := focused.state.Content.Get()
		pos := focused.state.CursorIndex.Get()
		anchor := focused.state.SelectionAnchor.Get()
		length = fmt.Sprintf("%d", len(content))
		cursor = fmt.Sprintf("%d of %d", pos, len(content))
		selection = "none"
		if anchor >= 0 && anchor != pos {
			selection = fmt.Sprintf("%d chars", abs(pos-anchor))
		}
	}

	email := strings.TrimSpace(joinGraphemes(d.fields[1].state.Content.Get()))
	emailStatus := "[b $TextMuted]empty[/]"
	if email != "" {
		if validEmail(email) {
			emailStatus = "[b $Success]valid[/]"
		} else {
			emailStatus = "[b $Error]invalid[/]"
		}
	}

	readOnly := "[b $Success]off[/]"
	if d.readOnly.Get() {
		readOnly = "[b $Warning]on[/]"
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Focus", fmt.Sprintf("[b $Primary]%s[/]", focusName)),
			demokit.StatRow(theme, "Length", fmt.Sprintf("[b $Info]%s[/]", length)),
			demokit.StatRow(theme, "Cursor", fmt.Sprintf("[b $Info]%s[/]", cursor)),
			demokit.StatRow(theme, "Selection", fmt.Sprintf("[b $Secondary]%s[/]", selection)),
			demokit.StatRow(theme, "Email", emailStatus),
			demokit.StatRow(theme, "Read-only", readOnly),
			demokit.StatRow(theme, "Edits", fmt.Sprintf("[b $Accent]%d[/]", d.edits.Get())),
			demokit.StatRow(theme, "Submitted", fmt.Sprintf("[b $Accent]%d[/]", d.submissions.Get())),
		},
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("tab", "$Info", "next field / button"),
			key("←→", "$Info", "move (ctrl: by word)"),
			key("⇧←→ ^a", "$Secondary", "select / select all"),
			key("ctrl+w", "$Error", "delete word"),
			key("ctrl+uk", "$Error", "delete to start/end"),
			key("esc", "$Error", "clear field"),
			key("enter", "$Success", "submit form"),
			key("ctrl+r", "$Warning", "toggle read-only"),
			key("ctrl+x", "$Warning", "clear form"),
		},
	}
}

// emailHighlighter colours the domain of an email address, and the whole
// address once it looks valid.
func emailHighlighter(theme t.ThemeData) t.Highlighter {
	return t.HighlighterFunc(func(text string, graphemes []string) []t.TextHighlight {
		at := -1
		for i, g := range graphemes {
			if g == "@" {
				at = i
				break
			}
		}
		if at < 0 {
			return nil
		}
		domain := t.SpanStyle{Foreground: theme.Info}
		if !validEmail(strings.TrimSpace(text)) {
			domain.Underline = t.UnderlineCurly
			domain.UnderlineColor = theme.Error
		}
		return []t.TextHighlight{
			{Start: at, End: at + 1, Style: t.SpanStyle{Foreground: theme.Accent, Bold: true}},
			{Start: at + 1, End: len(graphemes), Style: domain},
		}
	})
}

// messageHighlighter colours #tags and @mentions.
func messageHighlighter(theme t.ThemeData) t.Highlighter {
	return t.HighlighterFunc(func(_ string, graphemes []string) []t.TextHighlight {
		var highlights []t.TextHighlight
		for i := 0; i < len(graphemes); i++ {
			g := graphemes[i]
			if (g != "#" && g != "@") || (i > 0 && graphemes[i-1] != " ") {
				continue
			}
			end := i + 1
			for end < len(graphemes) && graphemes[end] != " " {
				end++
			}
			if end == i+1 {
				continue
			}
			colour := theme.Accent
			if g == "@" {
				colour = theme.Info
			}
			highlights = append(highlights, t.TextHighlight{Start: i, End: end, Style: t.SpanStyle{Foreground: colour, Bold: true}})
			i = end
		}
		return highlights
	})
}

// validEmail is a deliberately simple check: something@something.something.
func validEmail(s string) bool {
	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || strings.ContainsAny(s, " ") {
		return false
	}
	dot := strings.LastIndex(domain, ".")
	return dot > 0 && dot < len(domain)-1
}

func joinGraphemes(graphemes []string) string {
	return strings.Join(graphemes, "")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
