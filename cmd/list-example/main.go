package main

import (
	"fmt"
	"log"

	t "terma"
)


// ListItem represents a single item in our list.
// This is pure data - rendering is handled by RenderItem.
type ListItem struct {
	Title       string
	Description string
}

// ListDemo demonstrates the List widget with custom data types.
// The List widget builds a Column of widgets internally,
// and integrates with ScrollState for scroll-into-view.
type EditorSettingsMenu struct {
	scrollState *t.ScrollState
	listState   *t.ListState[ListItem] // ListState holds both items and cursor position
	message     t.Signal[string]
}

func NewEditorSettingsMenu() *EditorSettingsMenu {
	return &EditorSettingsMenu{
		scrollState: t.NewScrollState(),
		listState: t.NewListState([]ListItem{
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
		}),
		message: t.NewSignal(""),
	}
}

func (d *EditorSettingsMenu) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "r", Name: "Move 2 rows down", Action: func() {
			d.scrollState.ScrollDown(2)
			cursorIdx := d.listState.CursorIndex.Peek()
			d.listState.SelectIndex(cursorIdx + 2)
		}},
		{Key: "l", Name: "Move 2 rows up", Action: func() {
			d.scrollState.ScrollUp(2)
			cursorIdx := d.listState.CursorIndex.Peek()
			d.listState.SelectIndex(cursorIdx - 2)
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
				Spans: t.ParseMarkup("Use [b #00ffff]↑/↓[/] or [b #00ffff]j/k[/] to navigate • [b #00ffff]Enter[/] to select • [b #00ffff]PgUp/PgDn[/] for fast scroll", t.ThemeData{}),
			},

			t.Row{
				Height: t.Flex(1),
				Children: []t.Widget{
					// The List widget inside a Scrollable (left side)
					t.Scrollable{
						ID:     "list-scroll",
						State:  d.scrollState,
						Height: t.Cells(12),
						Width:  t.Flex(2),
						Style: t.Style{
							Border: t.RoundedBorder(t.Cyan, t.BorderTitle("Editor Settings"),
								t.BorderDecoration{Text: fmt.Sprintf("%d", d.scrollState.GetOffset()), Position: t.DecorationBottomRight}),
							Padding: t.EdgeInsetsAll(1),
						},
						Child: t.List[ListItem]{
							ID:          "demo-list",
							State:       d.listState,
							ScrollState: d.scrollState,
							MultiSelect: true,
							OnSelect: func(item ListItem) {
								d.message.Set(fmt.Sprintf("Selected: %s", item.Title))
							},
							// Describe how to render each item in the list as a widget
							RenderItem: func(item ListItem, active bool, selected bool) t.Widget {
								// Style the title based on cursor position and selection
								var titleStyle t.Style
								if active {
									titleStyle.ForegroundColor = t.Magenta
								}
								if selected {
									titleStyle.BackgroundColor = t.Red
								}

								// Each item is 2 lines tall (title + description)
								return t.Column{
									Height: t.Cells(2), // We need to declare the height of the rendered item
									Children: []t.Widget{
										t.Text{
											Content: item.Title,
											Style:   titleStyle,
											Width:   t.Flex(1),
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
					t.Text{Content: fmt.Sprintf("%d", d.scrollState.GetOffset()), Width: t.Flex(1), Height: t.Flex(1), Style: t.Style{ForegroundColor: t.Black, BackgroundColor: t.Magenta}},
				},
			},

			// Footer showing current cursor position and last selection
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("Cursor: [b #ffff00]%d[/] / %d • %s • Press [b #ff5555]Ctrl+C[/] to quit", d.listState.CursorIndex.Get()+1, d.listState.ItemCount(), d.message.Get()), t.ThemeData{}),
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
