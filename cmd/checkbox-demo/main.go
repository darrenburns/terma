package main

import (
	"fmt"

	t "terma"
)

// CheckboxDemo demonstrates the Checkbox widget with various settings.
type CheckboxDemo struct {
	darkMode      t.Signal[bool]
	notifications t.Signal[bool]
	autoSave      t.Signal[bool]
	compactView   t.Signal[bool]
}

// NewCheckboxDemo creates a new checkbox demo application.
func NewCheckboxDemo() *CheckboxDemo {
	return &CheckboxDemo{
		darkMode:      t.NewSignal(false),
		notifications: t.NewSignal(true),
		autoSave:      t.NewSignal(false),
		compactView:   t.NewSignal(false),
	}
}

// Keybinds returns the app-level keybindings.
func (d *CheckboxDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

// Build returns the widget tree for this app.
func (d *CheckboxDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	// Read all signal values
	darkMode := d.darkMode.Get()
	notifications := d.notifications.Get()
	autoSave := d.autoSave.Get()
	compactView := d.compactView.Get()

	// Count checked items
	count := 0
	if darkMode {
		count++
	}
	if notifications {
		count++
	}
	if autoSave {
		count++
	}
	if compactView {
		count++
	}

	return t.Dock{
		Top: []t.Widget{
			t.Text{
				Content: " Checkbox Demo ",
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: theme.Primary,
					ForegroundColor: theme.TextOnPrimary,
				},
			},
		},
		Bottom: []t.Widget{t.KeybindBar{}},
		Body: t.Column{
			Style: t.Style{Padding: t.EdgeInsetsAll(1)},
			Children: []t.Widget{
				t.Text{
					Content: "Settings",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Spacer{Height: t.Cells(1)},
				&t.Checkbox{
					ID:       "dark-mode",
					Label:    "Enable dark mode",
					Checked:  darkMode,
					OnToggle: func(v bool) { d.darkMode.Set(v) },
				},
				&t.Checkbox{
					ID:       "notifications",
					Label:    "Enable notifications",
					Checked:  notifications,
					OnToggle: func(v bool) { d.notifications.Set(v) },
				},
				&t.Checkbox{
					ID:       "auto-save",
					Label:    "Auto-save documents",
					Checked:  autoSave,
					OnToggle: func(v bool) { d.autoSave.Set(v) },
				},
				&t.Checkbox{
					ID:       "compact-view",
					Label:    "Use compact view",
					Checked:  compactView,
					OnToggle: func(v bool) { d.compactView.Set(v) },
				},
				t.Spacer{Height: t.Cells(1)},
				t.Text{
					Content: fmt.Sprintf("Selected: %d/4", count),
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Spacer{Height: t.Cells(1)},
				t.Text{
					Content: "Use Tab to navigate, Space or Enter to toggle",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
			},
		},
	}
}

func main() {
	app := NewCheckboxDemo()
	if err := t.Run(app); err != nil {
		panic(err)
	}
}
