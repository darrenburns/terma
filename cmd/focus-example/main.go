package main

import (
	"fmt"
	"log"

	t "terma"
)

// FocusedLabel is a widget that displays the currently focused widget's key.
type FocusedLabel struct{}

func (f FocusedLabel) Build(ctx t.BuildContext) t.Widget {
	// Read the focused widget from context (reactive - triggers rebuild on change)
	focused := ctx.Focused()

	label := "(none)"
	if focused != nil {
		if identifiable, ok := focused.(t.Identifiable); ok {
			label = identifiable.WidgetID()
		}
	}

	return t.Text{Content: fmt.Sprintf("Focused: %q", label)}
}

// EditorPanel is a container widget with its own keybinds.
// When a child inside this panel is focused, the panel's keybinds
// bubble up and appear in the footer alongside the child's keybinds.
type EditorPanel struct {
	message t.Signal[string]
}

// Keybinds returns panel-level keybindings available to all children.
func (p *EditorPanel) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+s", Name: "Save", Action: func() {
			p.message.Set("Panel: Save triggered!")
		}},
		{Key: "ctrl+z", Name: "Undo", Action: func() {
			p.message.Set("Panel: Undo triggered!")
		}},
		{Key: "d", Name: "Delete", Action: func() {
			p.message.Set("Panel: Delete triggered!")
		}},
	}
}

func (p *EditorPanel) Build(ctx t.BuildContext) t.Widget {
	return t.Column{
		Style: t.Style{
			Padding:         t.EdgeInsets{Left: 2, Right: 2, Top: 1, Bottom: 1},
			BackgroundColor: t.Hex("#1a1a2e"),
		},
		Children: []t.Widget{
			t.Text{Content: "Editor Panel (has ctrl+s, ctrl+z, d keybinds):"},
			t.Text{Content: ""},
			// Regular button - inherits panel keybinds
			&t.Button{ID: "btn-new-file", Label: "New File", OnPress: func() {
				p.message.Set("New File button pressed!")
			}},
			// Button with conflicting "d" keybind - overrides panel's "d"
			&DeleteButton{message: p.message},
		},
	}
}

// DeleteButton is a focusable widget that overrides the parent's "d" keybind.
// When focused, pressing "d" triggers this button's action, not the panel's.
type DeleteButton struct {
	message t.Signal[string]
}

func (b *DeleteButton) WidgetID() string { return "btn-delete" }
func (b *DeleteButton) IsFocusable() bool { return true }
func (b *DeleteButton) OnKey(event t.KeyEvent) bool { return false }

// Keybinds overrides the parent panel's "d" keybind with a different action.
func (b *DeleteButton) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "enter", Name: "Press", Action: b.press},
		{Key: " ", Name: "Press", Action: b.press},
		// This "d" keybind takes precedence over the panel's "d" keybind
		{Key: "d", Name: "Delete Forever", Action: func() {
			b.message.Set("DeleteButton: DELETE FOREVER! (overrides panel)")
		}},
	}
}

func (b *DeleteButton) press() {
	b.message.Set("Delete button pressed!")
}

func (b *DeleteButton) Build(ctx t.BuildContext) t.Widget {
	style := t.Style{Padding: t.EdgeInsets{Left: 1, Right: 1}}
	if ctx.IsFocused(b) {
		theme := ctx.Theme()
		style.BackgroundColor = theme.Primary
		style.ForegroundColor = theme.TextOnPrimary
	}
	return t.Text{Content: "Delete", Style: style}
}

// App is the root widget for this application.
// It demonstrates app-level declarative keybindings that apply globally.
type App struct {
	message t.Signal[string]
}

// Keybinds returns app-level keybindings.
// These fire when no focused widget handles the key.
func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "?", Name: "Help", Action: func() {
			a.message.Set("Help requested!")
		}},
		{Key: "r", Name: "Refresh", Action: func() {
			a.message.Set("Refreshed!")
		}},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	msg := a.message.Get()

	return t.Column{
		Children: []t.Widget{
			t.Text{Content: "=== Focus Demo (Keybind Bubbling) ==="},
			t.Text{Content: ""},
			FocusedLabel{},
			t.Text{Content: ""},
			t.Text{Content: msg, Style: t.Style{ForegroundColor: t.Green}},
			t.Text{Content: ""},
			t.Text{Content: "Use Tab/Shift+Tab to navigate. Watch footer change!"},
			t.Text{Content: ""},
			// Standalone button - only has app-level keybinds (?, r)
			&t.Button{ID: "btn-standalone", Label: "Standalone Button", OnPress: func() {
				a.message.Set("Standalone button pressed!")
			}},
			t.Text{Content: ""},
			// EditorPanel has its own keybinds that bubble to children
			&EditorPanel{message: a.message},
			t.Text{Content: ""},
			t.Text{Content: "Try: Tab to 'New File' shows panel keybinds (ctrl+s, ctrl+z, d)"},
			t.Text{Content: "     Tab to 'Delete' shows 'd' as 'Delete Forever' (overrides panel)"},
			t.Text{Content: ""},
			t.KeybindBar{},
		},
	}
}

func main() {
	app := &App{
		message: t.NewSignal(""),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
