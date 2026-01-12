# Checkbox

A focusable widget that displays a toggleable checked/unchecked state with an optional label.

```go
package main

import t "terma"

type App struct {
    enabled t.Signal[bool]
}

func NewApp() *App {
    return &App{enabled: t.NewSignal(false)}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return &t.Checkbox{
        ID:       "enable-feature",
        Label:    "Enable notifications",
        Checked:  a.enabled.Get(),
        OnToggle: func(v bool) { a.enabled.Set(v) },
    }
}

func main() {
    t.Run(NewApp())
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | — | Required identifier for focus management |
| `Label` | `string` | `""` | Optional text displayed next to the checkbox |
| `Checked` | `bool` | `false` | Current checked state (pass value, not signal) |
| `OnToggle` | `func(bool)` | — | Called when toggled, receives the new state |
| `Width` | `Dimension` | `Auto` | Width constraint |
| `Height` | `Dimension` | `Auto` | Height constraint |
| `Style` | `Style` | — | Padding, colors, border |
| `Click` | `func()` | — | Optional additional click handler |
| `Hover` | `func(bool)` | — | Optional hover state handler |

## Visual Appearance

The checkbox displays as:

```
[ ] Label text     (unchecked)
[✓] Label text     (checked)
```

When focused, the checkbox is highlighted with theme colors (Primary background, TextOnPrimary foreground).

## Keyboard Interaction

| Key | Action |
|-----|--------|
| `Space` | Toggle checked state |
| `Enter` | Toggle checked state |
| `Tab` | Move focus to next widget |
| `Shift+Tab` | Move focus to previous widget |

The Space keybind is shown in the KeybindBar when the checkbox is focused.

## Basic Usage

```go
// Simple checkbox
&Checkbox{
    ID:       "agree-terms",
    Label:    "I agree to the terms",
    Checked:  agreed,
    OnToggle: func(v bool) { a.agreed.Set(v) },
}

// Checkbox without label
&Checkbox{
    ID:       "selected",
    Checked:  selected,
    OnToggle: func(v bool) { a.selected.Set(v) },
}
```

## Settings Panel Example

A common pattern is a settings panel with multiple checkboxes:

```go
type SettingsApp struct {
    darkMode      t.Signal[bool]
    notifications t.Signal[bool]
    autoSave      t.Signal[bool]
}

func NewSettingsApp() *SettingsApp {
    return &SettingsApp{
        darkMode:      t.NewSignal(false),
        notifications: t.NewSignal(true),
        autoSave:      t.NewSignal(false),
    }
}

func (a *SettingsApp) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Style:   t.Style{Padding: t.EdgeInsetsAll(1)},
        Spacing: 1,
        Children: []t.Widget{
            t.Text{
                Content: "Settings",
                Style:   t.Style{ForegroundColor: theme.TextMuted},
            },
            &t.Checkbox{
                ID:       "dark-mode",
                Label:    "Enable dark mode",
                Checked:  a.darkMode.Get(),
                OnToggle: func(v bool) { a.darkMode.Set(v) },
            },
            &t.Checkbox{
                ID:       "notifications",
                Label:    "Enable notifications",
                Checked:  a.notifications.Get(),
                OnToggle: func(v bool) { a.notifications.Set(v) },
            },
            &t.Checkbox{
                ID:       "auto-save",
                Label:    "Auto-save documents",
                Checked:  a.autoSave.Get(),
                OnToggle: func(v bool) { a.autoSave.Set(v) },
            },
        },
    }
}
```

## With Conditional Content

Show or hide content based on checkbox state:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    showAdvanced := a.showAdvanced.Get()

    return t.Column{
        Children: []t.Widget{
            &t.Checkbox{
                ID:       "show-advanced",
                Label:    "Show advanced options",
                Checked:  showAdvanced,
                OnToggle: func(v bool) { a.showAdvanced.Set(v) },
            },
            t.ShowWhen(showAdvanced, t.Column{
                Style: t.Style{Padding: t.EdgeInsetsLeft(2)},
                Children: []t.Widget{
                    &t.Checkbox{ID: "opt1", Label: "Advanced option 1", ...},
                    &t.Checkbox{ID: "opt2", Label: "Advanced option 2", ...},
                },
            }),
        },
    }
}
```

## Styling

Apply custom styles through the Style field:

```go
&Checkbox{
    ID:      "styled",
    Label:   "Custom styled checkbox",
    Checked: checked,
    Style: Style{
        Padding:         EdgeInsetsXY(1, 0),
        BackgroundColor: theme.Surface,
    },
    OnToggle: func(v bool) { a.checked.Set(v) },
}
```

## Notes

- The `ID` field is required for focus management
- Use the values-first pattern: pass `signal.Get()` to `Checked`, not the signal itself
- The `OnToggle` callback receives the new state (inverted from current)
- Checkboxes are focusable and will be included in Tab navigation
- Click anywhere on the checkbox (including the label) to toggle
