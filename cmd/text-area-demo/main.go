package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"
)

type TextAreaDemo struct {
	editorState *t.TextAreaState
	vimState    *t.TextAreaState
	scrollState *t.ScrollState
}

func NewTextAreaDemo() *TextAreaDemo {
	editor := t.NewTextAreaState(strings.TrimSpace(`
Terma TextArea supports multiple lines, wrapping, and scroll control.
Try moving the cursor with arrows and Page Up/Down.
Toggle wrapping with Alt+Z to see the difference.

This block is intentionally long so the Scrollable container can show off
scroll syncing while you move the cursor through the content.
`))
	vim := t.NewTextAreaState("Press i or Enter to edit here.\nEscape returns to normal mode.\n")
	return &TextAreaDemo{
		editorState: editor,
		vimState:    vim,
		scrollState: t.NewScrollState(),
	}
}

func (d *TextAreaDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "alt+z", Name: "Toggle Wrap", Action: func() { d.editorState.ToggleWrap() }},
	}
}

func (d *TextAreaDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	wrapLabel := "On"
	if d.editorState.WrapMode.Get() == t.WrapNone {
		wrapLabel = "Off"
	}

	return t.Dock{
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				t.Text{
					Content: " TextArea Demo ",
					Style: t.Style{
						ForegroundColor: theme.TextOnPrimary,
						BackgroundColor: theme.Primary,
					},
				},
				t.Text{
					Spans: t.ParseMarkup(
						"[b $Primary]Tab[/] to switch fields | [b $Primary]Alt+Z[/] to toggle wrap | [b $Primary]Ctrl+C[/] to quit",
						theme,
					),
				},
				t.Text{
					Content: fmt.Sprintf("Wrap: %s", wrapLabel),
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},
				t.Text{Content: "Scrollable editor:", Style: t.Style{ForegroundColor: theme.Text}},
				t.Scrollable{
					ID:     "editor-scroll",
					State:  d.scrollState,
					Height: t.Cells(8),
					Child: t.TextArea{
						ID:          "editor",
						State:       d.editorState,
						ScrollState: d.scrollState,
						Width:       t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
						},
					},
				},
				t.Text{Content: "Insert-mode example:", Style: t.Style{ForegroundColor: theme.Text}},
				t.TextArea{
					ID:                "vim",
					State:             d.vimState,
					RequireInsertMode: true,
					Width:             t.Flex(1),
					Height:            t.Cells(4),
					Style: t.Style{
						BackgroundColor: theme.Surface,
						ForegroundColor: theme.Text,
					},
				},
			},
		},
	}
}

func main() {
	app := NewTextAreaDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
