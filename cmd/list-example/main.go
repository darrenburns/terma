package main

import (
	"fmt"
	"log"

	t "terma"
)

func init() {
	if err := t.InitLogger(); err != nil {
		log.Printf("Warning: could not initialize logger: %v", err)
	}
}

// ListItem represents a single item in our list.
// This is pure data - rendering is handled by RenderItem.
type ListItem struct {
	Title       string
	Description string
}

// ListDemo demonstrates the List widget with custom data types.
// The List widget builds a Column of widgets internally,
// and integrates with ScrollController for scroll-into-view.
type EditorSettingsMenu struct {
	controller  *t.ScrollController
	cursorIndex *t.Signal[int]
	items       []ListItem
	message     *t.Signal[string]
}

func NewEditorSettingsMenu() *EditorSettingsMenu {
	items := []ListItem{
		{Title: "New Project", Description: "Create a new project from template"},
		{Title: "Open Folder", Description: "Open an existing folder"},
		{Title: "Clone Repository", Description: "Clone a Git repository from URL"},
		{Title: "Connect to Remote", Description: "SSH into a remote server"},
		{Title: "Import Settings", Description: "Import settings from another editor"},
		{Title: "Install Extensions", Description: "Browse and install extensions"},
		{Title: "Keyboard Shortcuts", Description: "Customize key bindings"},
		{Title: "Color Theme", Description: "Change the editor color theme"},
		{Title: "Font Settings", Description: "Configure font family and size"},
		{Title: "Auto Save", Description: "Configure automatic file saving"},
		{Title: "Format on Save", Description: "Run formatter when saving files"},
		{Title: "Line Numbers", Description: "Toggle line number visibility"},
		{Title: "Word Wrap", Description: "Configure text wrapping behavior"},
		{Title: "Terminal", Description: "Open integrated terminal"},
		{Title: "Check for Updates", Description: "Check for application updates"},
	}

	return &EditorSettingsMenu{
		controller:  t.NewScrollController(),
		cursorIndex: t.NewSignal(0),
		items:       items,
		message:     t.NewSignal(""),
	}
}

func (d *EditorSettingsMenu) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "r", Name: "Move 2 rows down", Action: func() {
			d.controller.ScrollDown(2)
			d.cursorIndex.Update(func(s int) int { return s + 2 })
		}},
		{Key: "l", Name: "Move 2 rows up", Action: func() {
			d.controller.ScrollUp(2)
			d.cursorIndex.Update(func(s int) int { return s - 2 })
		}},
	}
}

func (d *EditorSettingsMenu) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		ID:      "root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsAll(1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "Editor Settings Menu",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Magenta,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},

			// Instructions
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Use "),
					t.BoldSpan("↑/↓", t.BrightCyan),
					t.PlainSpan(" or "),
					t.BoldSpan("j/k", t.BrightCyan),
					t.PlainSpan(" to navigate • "),
					t.BoldSpan("Enter", t.BrightCyan),
					t.PlainSpan(" to select • "),
					t.BoldSpan("PgUp/PgDn", t.BrightCyan),
					t.PlainSpan(" for fast scroll"),
				},
			},

			t.Row{
				Height: t.Fr(1),
				Children: []t.Widget{
					// The List widget inside a Scrollable (left side)
					&t.Scrollable{
						ID:           "list-scroll",
						Controller:   d.controller,
						Height:       t.Cells(12),
						Width:        t.Fr(2),
						DisableFocus: true, // Let List handle focus
						Style: t.Style{
							Border: t.RoundedBorder(t.Cyan, t.BorderTitle("Editor Settings"),
								t.BorderDecoration{Text: fmt.Sprintf("%d", d.controller.Offset()), Position: t.DecorationBottomRight}),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: &t.List[ListItem]{
							ID:               "demo-list",
							Items:            d.items,
							CursorIndex:      d.cursorIndex,
							ScrollController: d.controller,
							OnSelect: func(item ListItem) {
								d.message.Set(fmt.Sprintf("Selected: %s", item.Title))
							},
							// Describe how to render each item in the list as a widget
							RenderItem: func(item ListItem, active bool) t.Widget {
								// Style the title based on cursor position
								titleStyle := t.Style{ForegroundColor: t.DefaultColor}
								if active {
									titleStyle.ForegroundColor = t.Magenta
								}

								// Each item is 2 lines tall (title + description)
								return t.Column{
									Height: t.Cells(2), // We need to declare the height of the rendered item
									Children: []t.Widget{
										t.Text{
											Content: item.Title,
											Style:   titleStyle,
											Width:   t.Fr(1),
										},
										t.Text{
											Content: item.Description,
											Style:   t.Style{ForegroundColor: t.BrightBlack},
										},
									},
								}
							},
						},
					},

					// Right side of the screen showing the current scroll offset
					t.Text{Content: fmt.Sprintf("%d", d.controller.Offset()), Width: t.Fr(1), Height: t.Fr(1), Style: t.Style{ForegroundColor: t.Black, BackgroundColor: t.Magenta}},
				},
			},

			// Footer showing current cursor position and last selection
			t.Text{
				Spans: []t.Span{
					t.PlainSpan("Cursor: "),
					t.BoldSpan(fmt.Sprintf("%d", d.cursorIndex.Get()+1), t.BrightYellow),
					t.PlainSpan(" / "),
					t.PlainSpan(fmt.Sprintf("%d", len(d.items))),
					t.PlainSpan(" • "),
					t.PlainSpan(d.message.Get()),
					t.PlainSpan(" • Press "),
					t.BoldSpan("Ctrl+C", t.BrightRed),
					t.PlainSpan(" to quit"),
				},
			},
		},
	}
}

func main() {
	app := NewEditorSettingsMenu()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
