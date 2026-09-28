package main

import (
	"fmt"
	"log"
	"strings"

	t "github.com/darrenburns/terma"
)

// TextAreaDemo shows off the multi-line TextArea widget: a scrolling editor
// with soft wrapping and highlighting, plus a modal (vim-style) editor that
// needs insert mode before it accepts text. The sidebar reads the focused
// editor's TextAreaState live.
//
//	tab / shift+tab - Switch between the two editors
//	↑↓←→ pgup pgdn  - Move the cursor (ctrl+←→ or alt+b/f by word)
//	shift+arrows    - Extend the selection (ctrl+a selects all)
//	ctrl+u ctrl+k   - Delete to start / end of the line
//	ctrl+w          - Delete the previous word
//	alt+z           - Toggle soft wrapping
//	ctrl+r          - Toggle read-only
//	ctrl+x          - Reset both editors to their starting text
//	i / enter       - Modal editor: enter insert mode
//	escape          - Modal editor: back to normal mode
type TextAreaDemo struct {
	editorState *t.TextAreaState
	modalState  *t.TextAreaState
	scrollState *t.ScrollState
}

const (
	editorID     = "editor"
	modalID      = "modal"
	sidebarWidth = 32
)

const editorText = `# Terma TextArea

A multi-line editor with soft wrapping, selection and scroll syncing. This editor sits inside a Scrollable, and its ScrollState keeps the cursor in view as you move — watch the Scroll row in the sidebar.

## Things to try

- Hold shift with the arrows to select, or ctrl+a to select everything.
- Press alt+z to switch wrapping off; long lines then scroll sideways instead of folding onto the next row, which is handy for code and logs.
- Press ctrl+r to make both editors read-only. The cursor still moves.
- Use ctrl+w to delete a word and ctrl+u / ctrl+k to trim a line.

## Highlighting

A Highlighter colours ` + "`inline code`" + `, TODO markers, list markers and headings as you type. Try typing a new TODO or wrapping a word in ` + "`backticks`" + `.

TODO: write the rest of the novel.

## Scrolling

The lines below are here so there is something to scroll through. Page Up and Page Down move a screen at a time.

1. The quick brown fox
2. jumps over
3. the lazy dog.
4. Pack my box with
5. five dozen liquor jugs.`

const modalText = "Press i or enter to edit here.\nEscape returns to normal mode."

func NewTextAreaDemo() *TextAreaDemo {
	editor := t.NewTextAreaState(editorText)
	editor.CursorIndex.Set(0)
	modal := t.NewTextAreaState(modalText)
	// Modal editors start in normal mode, like vim.
	modal.InsertMode.Set(false)
	return &TextAreaDemo{
		editorState: editor,
		modalState:  modal,
		scrollState: t.NewScrollState(),
	}
}

func (d *TextAreaDemo) states() []*t.TextAreaState {
	return []*t.TextAreaState{d.editorState, d.modalState}
}

func (d *TextAreaDemo) toggleWrap() {
	for _, s := range d.states() {
		s.ToggleWrap()
	}
}

func (d *TextAreaDemo) toggleReadOnly() {
	readOnly := !d.editorState.ReadOnly.Peek()
	for _, s := range d.states() {
		s.ReadOnly.Set(readOnly)
	}
}

func (d *TextAreaDemo) reset() {
	for _, s := range d.states() {
		s.ClearSelection()
	}
	d.editorState.SetText(editorText)
	d.editorState.CursorIndex.Set(0)
	d.modalState.SetText(modalText)
	d.modalState.InsertMode.Set(false)
	d.scrollState.Offset.Set(0)
}

func (d *TextAreaDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "alt+z", Name: "Wrap", Action: d.toggleWrap},
		{Key: "ctrl+r", Name: "Read-only", Action: d.toggleReadOnly},
		{Key: "ctrl+x", Name: "Reset", Action: d.reset, Hidden: true},
	}
}

func (d *TextAreaDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "text-area-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			header{demo: d},
		},
		Bottom: []t.Widget{
			keybindBar{demo: d},
		},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: []t.Widget{
				t.Column{
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						fill(editorPanel{demo: d}),
						modalPanel{demo: d},
					},
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

// header is the title bar across the top of the screen.
type header struct {
	demo *TextAreaDemo
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ TextArea Playground[/]  [$TextMuted]Multi-line text editing[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]wrap[/] %s", onOff(h.demo.editorState.WrapMode.Get() != t.WrapNone, "$Success", "$Warning")), theme),
		},
	}
}

// keybindBar is the footer. TextArea's keybinds depend on its insert mode and
// read-only state, but KeybindBar only rebuilds when focus changes, so this
// wrapper subscribes to those signals to keep the bar current.
type keybindBar struct {
	demo *TextAreaDemo
}

func (k keybindBar) Build(ctx t.BuildContext) t.Widget {
	_ = k.demo.modalState.InsertMode.Get()
	_ = k.demo.editorState.ReadOnly.Get()
	return t.KeybindBar{
		Style: t.Style{
			BackgroundColor: ctx.Theme().Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
	}
}

// fill gives a component the remaining space in its Column. Rows and Columns
// read Flex from their direct children, so a component's own Flex(1) needs a
// plain wrapper to take effect.
func fill(child t.Widget) t.Widget {
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{child},
	}
}

// panelStyle is the bordered look shared by every panel in the demo.
func panelStyle(theme t.ThemeData, title string, focused bool) t.Style {
	color := theme.Border
	titleColor := "$TextMuted"
	if focused {
		color = theme.FocusRing
		titleColor = "$FocusRing"
	}
	return t.Style{
		BackgroundColor: theme.Background,
		Border:          t.RoundedBorder(color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] %s [/]", titleColor, title))),
		Padding:         t.EdgeInsetsXY(1, 0),
	}
}

// editorPanel is the main editor: a TextArea inside a Scrollable, with the
// ScrollState shared so moving the cursor scrolls it into view.
type editorPanel struct {
	demo *TextAreaDemo
}

func (p editorPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := p.demo
	editor := t.TextArea{
		ID:          editorID,
		State:       d.editorState,
		ScrollState: d.scrollState,
		Placeholder: "Start typing…",
		Highlighter: markdownHighlighter(theme),
		Style: t.Style{
			Width:           t.Flex(1),
			BackgroundColor: theme.Background,
			ForegroundColor: theme.Text,
		},
	}

	title := "Editor"
	if d.editorState.ReadOnly.Get() {
		title += " · read-only"
	}
	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  panelStyle(theme, title, ctx.IsFocused(editor)),
		Children: []t.Widget{
			t.Scrollable{
				ID:    "editor-scroll",
				State: d.scrollState,
				Style: t.Style{
					Width:  t.Flex(1),
					Height: t.Flex(1),
				},
				Child: editor,
			},
		},
	}
}

// modalPanel is a TextArea with RequireInsertMode: keys move the cursor until
// you press i or enter, and escape leaves insert mode again.
type modalPanel struct {
	demo *TextAreaDemo
}

func (p modalPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := p.demo
	editor := t.TextArea{
		ID:                modalID,
		State:             d.modalState,
		RequireInsertMode: true,
		Style: t.Style{
			Width:           t.Flex(1),
			Height:          t.Cells(3),
			BackgroundColor: theme.Background,
			ForegroundColor: theme.Text,
		},
		// Leave insert mode when focus moves away. TextArea also drops to
		// normal mode when it gains focus, but it does so while painting,
		// which leaves the KeybindBar and sidebar showing the old mode.
		Blur: func() { d.modalState.InsertMode.Set(false) },
	}

	mode := "[b $Info]NORMAL[/]"
	if d.modalState.InsertMode.Get() {
		mode = "[b $Success]INSERT[/]"
	}
	focused := ctx.IsFocused(editor)
	titleColor := "$TextMuted"
	if focused {
		titleColor = "$FocusRing"
	}
	style := panelStyle(theme, "", focused)
	style.Border = t.RoundedBorder(style.Border.Color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] Modal editor ·[/] %s ", titleColor, mode)))

	return t.Column{
		Width:    t.Flex(1),
		Style:    style,
		Children: []t.Widget{editor},
	}
}

// statePanel reads the focused editor's TextAreaState live. It subscribes to
// the signals itself, so typing only rebuilds this panel and the editor.
type statePanel struct {
	demo *TextAreaDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	focusName := "—"
	state := d.editorState
	if w := ctx.Focused(); w != nil {
		if id, ok := w.(t.Identifiable); ok {
			switch id.WidgetID() {
			case editorID:
				focusName = "Editor"
			case modalID:
				focusName = "Modal"
				state = d.modalState
			}
		}
	}

	content := state.Content.Get()
	cursor := state.CursorIndex.Get()
	anchor := state.SelectionAnchor.Get()
	line, col := lineCol(content, cursor)

	selection := "none"
	if anchor >= 0 && anchor != cursor {
		selection = fmt.Sprintf("%d chars", abs(cursor-anchor))
	}

	mode := "[b $Info]normal[/]"
	if d.modalState.InsertMode.Get() {
		mode = "[b $Success]insert[/]"
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "State", false),
		Children: []t.Widget{
			statRow(theme, "Focus", fmt.Sprintf("[b $Primary]%s[/]", focusName)),
			statRow(theme, "Lines", fmt.Sprintf("[b $Info]%d[/]", strings.Count(strings.Join(content, ""), "\n")+1)),
			statRow(theme, "Chars", fmt.Sprintf("[b $Info]%d[/]", len(content))),
			statRow(theme, "Cursor", fmt.Sprintf("[b $Info]Ln %d, Col %d[/]", line, col)),
			statRow(theme, "Selection", fmt.Sprintf("[b $Secondary]%s[/]", selection)),
			statRow(theme, "Scroll", fmt.Sprintf("[b $Accent]%d[/]", d.scrollState.Offset.Get())),
			statRow(theme, "Wrap", onOff(d.editorState.WrapMode.Get() != t.WrapNone, "$Success", "$Warning")),
			statRow(theme, "Read-only", onOff(d.editorState.ReadOnly.Get(), "$Warning", "$Success")),
			statRow(theme, "Modal mode", mode),
		},
	}
}

func statRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}

// onOff renders a boolean as coloured "on"/"off" markup.
func onOff(on bool, onColour, offColour string) string {
	if on {
		return fmt.Sprintf("[b %s]on[/]", onColour)
	}
	return fmt.Sprintf("[b %s]off[/]", offColour)
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
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("tab", "$Info", "switch editor"),
			key("arrows", "$Info", "move (pgup/pgdn)"),
			key("⇧ ^a", "$Secondary", "select / select all"),
			key("^w ^u^k", "$Error", "delete word / line"),
			key("alt+z", "$Accent", "toggle wrap"),
			key("ctrl+r", "$Warning", "toggle read-only"),
			key("ctrl+x", "$Warning", "reset text"),
			key("i esc", "$Success", "insert / normal mode"),
		},
	}
}

// markdownHighlighter colours headings, list markers, `inline code` and TODOs.
func markdownHighlighter(theme t.ThemeData) t.Highlighter {
	return t.HighlighterFunc(func(_ string, graphemes []string) []t.TextHighlight {
		var highlights []t.TextHighlight
		lineStart := 0
		for i := 0; i <= len(graphemes); i++ {
			if i < len(graphemes) && graphemes[i] != "\n" {
				continue
			}
			highlights = append(highlights, highlightLine(theme, graphemes, lineStart, i)...)
			lineStart = i + 1
		}
		return highlights
	})
}

// highlightLine returns the highlights for graphemes[start:end], one line.
func highlightLine(theme t.ThemeData, g []string, start, end int) []t.TextHighlight {
	if start < end && g[start] == "#" {
		return []t.TextHighlight{{Start: start, End: end, Style: t.SpanStyle{Foreground: theme.Primary, Bold: true}}}
	}

	var highlights []t.TextHighlight
	if start+1 < end && (g[start] == "-" || isDigit(g[start])) {
		marker := start
		for marker < end && g[marker] != " " {
			marker++
		}
		highlights = append(highlights, t.TextHighlight{Start: start, End: marker, Style: t.SpanStyle{Foreground: theme.Secondary, Bold: true}})
	}

	for i := start; i < end; i++ {
		switch {
		case g[i] == "`":
			close := i + 1
			for close < end && g[close] != "`" {
				close++
			}
			if close < end {
				highlights = append(highlights, t.TextHighlight{Start: i, End: close + 1, Style: t.SpanStyle{Foreground: theme.Accent}})
				i = close
			}
		case i+4 <= end && strings.Join(g[i:i+4], "") == "TODO":
			highlights = append(highlights, t.TextHighlight{Start: i, End: i + 4, Style: t.SpanStyle{Foreground: theme.Warning, Bold: true}})
			i += 3
		}
	}
	return highlights
}

func isDigit(g string) bool {
	return len(g) == 1 && g[0] >= '0' && g[0] <= '9'
}

// lineCol returns the 1-based line and column of a grapheme index.
func lineCol(graphemes []string, cursor int) (line, col int) {
	line, col = 1, 1
	for i := 0; i < cursor && i < len(graphemes); i++ {
		if graphemes[i] == "\n" {
			line++
			col = 1
		} else {
			col++
		}
	}
	return line, col
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func main() {
	app := NewTextAreaDemo()
	t.RequestFocus(editorID)
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
