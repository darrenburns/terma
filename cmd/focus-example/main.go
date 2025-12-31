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
		if keyed, ok := focused.(t.Keyed); ok {
			label = keyed.Key()
		}
	}

	return t.Text{Content: fmt.Sprintf("Focused: %q", label)}
}

// Footer displays the currently active keybindings.
type Footer struct{}

func (f Footer) Build(ctx t.BuildContext) t.Widget {
	keybinds := ctx.ActiveKeybinds()

	if len(keybinds) == 0 {
		return t.Text{Content: ""}
	}

	// Build the keybinding display, deduplicating by key
	seen := make(map[string]bool)
	var children []t.Widget
	for _, kb := range keybinds {
		if seen[kb.Key] {
			continue
		}
		seen[kb.Key] = true

		// Add separator between keybindings
		if len(children) > 0 {
			children = append(children, t.Text{Content: "  "})
		}

		children = append(children, t.Text{
			Content: fmt.Sprintf("[%s] %s", kb.Key, kb.Name),
			Style:   t.Style{ForegroundColor: t.Cyan},
		})
	}

	return t.Row{Children: children}
}

// App is the root widget for this application.
// It demonstrates app-level declarative keybindings that apply globally.
type App struct {
	message *t.Signal[string]
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
			t.Text{Content: "=== Focus Demo (Declarative Keybindings) ==="},
			t.Text{Content: ""},
			FocusedLabel{}, // Uses ctx.Focused() to get current focus
			t.Text{Content: ""},
			// Show message if set
			t.Text{Content: msg, Style: t.Style{ForegroundColor: t.Green}},
			t.Text{Content: ""},
			t.Text{Content: "Use Tab/Shift+Tab to navigate:"},
			t.Text{Content: ""},
			t.Row{
				Children: []t.Widget{
					&t.Button{ID: "btn-save", Label: "Save", OnPress: func() {
						a.message.Set("Save button pressed!")
					}},
					t.Text{Content: "  "},
					&t.Button{ID: "btn-cancel", Label: "Cancel", OnPress: func() {
						a.message.Set("Cancel button pressed!")
					}},
					t.Text{Content: "  "},
					&t.Button{ID: "btn-help", Label: "Help", OnPress: func() {
						a.message.Set("Help button pressed!")
					}},
				},
			},
			t.Text{Content: ""},
			t.Column{
				Children: []t.Widget{
					&t.Button{ID: "btn-option-1", Label: "Option 1", OnPress: func() {
						a.message.Set("Option 1 selected!")
					}},
					&t.Button{ID: "btn-option-2", Label: "Option 2", OnPress: func() {
						a.message.Set("Option 2 selected!")
					}},
					&t.Button{ID: "btn-option-3", Label: "Option 3", OnPress: func() {
						a.message.Set("Option 3 selected!")
					}},
				},
			},
			t.Text{Content: ""},
			t.Text{Content: "Press Ctrl+C to quit"},
			t.Text{Content: ""},
			Footer{}, // Displays active keybindings from focused widget + ancestors
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
