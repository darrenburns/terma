package main

import (
	"log"
	"strings"

	t "terma"
)

type CommandPaletteDemo struct {
	palette *t.CommandPaletteState
	status  t.Signal[string]
	preview t.Signal[string]
}

const themesPaletteTitle = "Themes"

func NewCommandPaletteDemo() *CommandPaletteDemo {
	app := &CommandPaletteDemo{
		status:  t.NewSignal("Ready"),
		preview: t.NewSignal(""),
	}

	app.palette = t.NewCommandPaletteState("Commands", []t.CommandPaletteItem{
		{Label: "New File", Hint: "Ctrl+N", Action: app.selectAction("New File")},
		{Label: "Open File", Hint: "Ctrl+O", Action: app.selectAction("Open File")},
		{Divider: "Edit"},
		{Label: "Cut", Hint: "Ctrl+X", Action: app.selectAction("Cut")},
		{Label: "Copy", Hint: "Ctrl+C", Action: app.selectAction("Copy")},
		{Label: "Paste", Hint: "Ctrl+V", Action: app.selectAction("Paste")},
		{
			Label:         themesPaletteTitle,
			ChildrenTitle: themesPaletteTitle,
			Children:      app.themeItems,
		},
		{
			Label:         "Recent",
			ChildrenTitle: "Recent Files",
			Children: func() []t.CommandPaletteItem {
				return []t.CommandPaletteItem{
					{Label: "app/main.go", Action: app.selectAction("Open app/main.go")},
					{Label: "internal/config.yaml", Action: app.selectAction("Open internal/config.yaml")},
					{Label: "README.md", Action: app.selectAction("Open README.md")},
				}
			},
		},
	})

	return app
}

func (a *CommandPaletteDemo) selectAction(label string) func() {
	return func() {
		a.status.Set("Selected: " + label)
		a.palette.Close(false)
	}
}

func (a *CommandPaletteDemo) selectThemeAction(themeName, label string) func() {
	return func() {
		t.SetTheme(themeName)
		a.status.Set("Selected: Theme " + label)
		a.palette.Close(false)
	}
}

func (a *CommandPaletteDemo) themeItems() []t.CommandPaletteItem {
	items := make([]t.CommandPaletteItem, 0, 24)
	addGroup := func(title string, names []string) {
		if len(names) == 0 {
			return
		}
		items = append(items, t.CommandPaletteItem{Divider: title})
		for _, name := range names {
			label := themeDisplayName(name)
			items = append(items, t.CommandPaletteItem{
				Label:      label,
				FilterText: label + " " + name,
				Data:       name,
				Action:     a.selectThemeAction(name, label),
			})
		}
	}
	addGroup("Dark Themes", t.DarkThemeNames())
	addGroup("Light Themes", t.LightThemeNames())
	return items
}

func themeDisplayName(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func (a *CommandPaletteDemo) togglePalette() {
	if a.palette.Visible.Peek() {
		a.palette.Close(false)
		return
	}
	a.palette.Open()
}

func (a *CommandPaletteDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+p", Name: "Command palette", Action: a.togglePalette},
	}
}

func (a *CommandPaletteDemo) handleCursorChange(item t.CommandPaletteItem) {
	a.preview.Set(item.Label)
	level := a.palette.CurrentLevel()
	if level == nil || level.Title != themesPaletteTitle {
		return
	}
	if themeName, ok := item.Data.(string); ok {
		t.SetTheme(themeName)
	}
}

func (a *CommandPaletteDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Stack{
		Style: t.Style{
			BackgroundColor: theme.Background,
			Width:           t.Flex(1),
			Height:          t.Flex(1),
		},
		Children: []t.Widget{
			t.Column{
				Style: t.Style{
					Padding: t.EdgeInsetsAll(2),
				},
				Spacing: 1,
				Children: []t.Widget{
					t.Text{
						Content: "Command Palette Demo",
						Style: t.Style{
							ForegroundColor: theme.TextOnPrimary,
							BackgroundColor: theme.Primary,
							Padding:         t.EdgeInsetsXY(2, 0),
						},
					},
					t.Text{
						Content: "Press Ctrl+P to open the command palette.",
						Style: t.Style{
							ForegroundColor: theme.TextMuted,
						},
					},
					t.Text{
						Content: "Preview: " + a.preview.Get(),
						Style: t.Style{
							ForegroundColor: theme.AccentText,
						},
					},
					t.Text{
						Content: a.status.Get(),
						Style: t.Style{
							ForegroundColor: theme.Text,
						},
					},
				},
			},
			t.CommandPalette{
				ID:             "command-palette",
				State:          a.palette,
				Position:       t.FloatPositionTopCenter,
				OnCursorChange: a.handleCursorChange,
			},
		},
	}
}

func main() {
	app := NewCommandPaletteDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
