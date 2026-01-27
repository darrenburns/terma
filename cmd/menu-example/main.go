package main

import (
	"fmt"
	"log"

	t "terma"
)

type MenuDemo struct {
	showMenu  t.Signal[bool]
	status    t.Signal[string]
	menuState *t.MenuState
}

func NewMenuDemo() *MenuDemo {
	demo := &MenuDemo{
		showMenu: t.NewSignal(false),
		status:   t.NewSignal("Ready"),
	}
	demo.menuState = t.NewMenuState([]t.MenuItem{
		{Label: "New", Shortcut: "Ctrl+N", Action: demo.action("New")},
		{Label: "Open", Shortcut: "Ctrl+O", Action: demo.action("Open")},
		{
			Label: "Open Recent",
			Children: []t.MenuItem{
				{Label: "alpha.txt", Action: demo.action("Open alpha.txt")},
				{Label: "bravo.txt", Action: demo.action("Open bravo.txt")},
			},
		},
		{Divider: "Settings"},
		{
			Label: "Settings",
			Children: []t.MenuItem{
				{Label: "Editor", Action: demo.action("Settings: Editor")},
				{
					Label: "Theme",
					Children: []t.MenuItem{
						{Label: "Light", Action: demo.action("Theme: Light")},
						{Label: "Dark", Action: demo.action("Theme: Dark")},
					},
				},
			},
		},
		{Label: "Exit", Action: demo.action("Exit")},
	})
	return demo
}

func (d *MenuDemo) action(label string) func() {
	return func() {
		d.status.Set(fmt.Sprintf("Selected: %s", label))
	}
}

func (d *MenuDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		ID:      "menu-demo-root",
		Spacing: 1,
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Style: t.Style{
			Padding:         t.EdgeInsetsXY(2, 1),
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			t.Text{
				Content: "Menu Widget Demo",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Cyan,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.Text{
				Content: "Use arrow keys or j/k to navigate, Enter to select, right/left to open/close submenus.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.Row{
				Spacing: 2,
				Children: []t.Widget{
					t.Button{
						ID:    "file-btn",
						Label: "File",
						OnPress: func() {
							d.showMenu.Set(true)
							t.RequestFocus("file-menu")
						},
					},
				},
			},
			t.ShowWhen(d.showMenu.Get(), t.Menu{
				ID:       "file-menu",
				State:    d.menuState,
				AnchorID: "file-btn",
				OnSelect: func(item t.MenuItem) {
					if item.Action != nil {
						item.Action()
					}
					d.menuState.CloseSubmenu()
					d.showMenu.Set(false)
					t.RequestFocus("file-btn")
				},
				OnDismiss: func() {
					d.menuState.CloseSubmenu()
					d.showMenu.Set(false)
					t.RequestFocus("file-btn")
				},
			}),
			t.Text{
				Content: d.status.Get(),
				Style:   t.Style{ForegroundColor: theme.Accent},
			},
			t.Text{
				Content: "Press Ctrl+C to quit.",
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func main() {
	app := NewMenuDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
