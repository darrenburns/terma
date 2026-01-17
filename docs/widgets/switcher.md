# Switcher

Display one widget at a time from a keyed collection. Use `Switcher` for tabbed interfaces, multi-view applications, or any scenario where you need to swap between different content areas.

```go
package main

import t "terma"

type App struct {
    activeTab t.Signal[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Switcher{
        Active: a.activeTab.Get(),
        Children: map[string]t.Widget{
            "home":     t.Text{Content: "Home View"},
            "settings": t.Text{Content: "Settings View"},
        },
    }
}

func main() {
    t.Run(&App{activeTab: t.NewSignal("home")})
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Active` | `string` | `""` | Key of the currently visible child |
| `Children` | `map[string]Widget` | — | Map of keys to widgets |
| `Width` | `Dimension` | `Auto` | Container width |
| `Height` | `Dimension` | `Auto` | Container height |
| `Style` | `Style` | — | Padding, margin, border |

## Basic Usage

```go
// Simple two-view switcher
Switcher{
    Active: "view1",
    Children: map[string]Widget{
        "view1": ContentA{},
        "view2": ContentB{},
    },
}

// With dimensions to fill available space
Switcher{
    Active: a.currentView.Get(),
    Width:  Flex(1),
    Height: Flex(1),
    Children: map[string]Widget{
        "list":   ListView{},
        "detail": DetailView{},
    },
}
```

## How It Works

Only the active child is built and rendered. Inactive children don't exist in the widget tree—they consume no resources and don't receive events.

When `Active` changes (typically via a Signal), Switcher rebuilds with the new child. If `Active` doesn't match any key in `Children`, an empty widget is rendered.

## State Preservation

Widget state persists across switches because state lives in your App, not in the widgets themselves. Store Signals and State objects in your App struct:

```go
type App struct {
    activeTab t.Signal[string]

    // State for each view persists across switches
    listState   *t.ListState[Item]
    scrollState *t.ScrollState
    formData    t.Signal[FormData]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Switcher{
        Active: a.activeTab.Get(),
        Children: map[string]t.Widget{
            "list":     ListView{State: a.listState},
            "details":  DetailsView{ScrollState: a.scrollState},
            "settings": SettingsView{FormData: a.formData},
        },
    }
}
```

When a user switches from "list" to "details" and back, the list selection in `listState` is preserved because it lives in the App, not the ListView widget.

## Tab Navigation

### With Keybinds

Combine Switcher with keybinds for keyboard-driven navigation:

```go
func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "1", Name: "Home", Action: func() { a.activeTab.Set("home") }},
        {Key: "2", Name: "Settings", Action: func() { a.activeTab.Set("settings") }},
        {Key: "3", Name: "Profile", Action: func() { a.activeTab.Set("profile") }},
    }
}
```

### With Tab Buttons

Create a row of buttons to switch between views:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Height: t.Flex(1),
        Children: []t.Widget{
            // Tab bar
            t.Row{
                Spacing: 1,
                Style: t.Style{
                    BackgroundColor: theme.Surface,
                    Padding:         t.EdgeInsetsHV(1, 0),
                },
                Children: []t.Widget{
                    a.tabButton("home", "Home", ctx),
                    a.tabButton("settings", "Settings", ctx),
                    a.tabButton("profile", "Profile", ctx),
                },
            },
            // Content
            t.Switcher{
                Active: a.activeTab.Get(),
                Height: t.Flex(1),
                Children: map[string]t.Widget{
                    "home":     HomeView{},
                    "settings": SettingsView{},
                    "profile":  ProfileView{},
                },
            },
        },
    }
}

func (a *App) tabButton(key, label string, ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()
    isActive := a.activeTab.Get() == key

    style := t.Style{Padding: t.EdgeInsetsHV(2, 0)}
    if isActive {
        style.BackgroundColor = theme.Primary
        style.ForegroundColor = theme.Primary.AutoText()
    }

    return t.Button{
        ID:      key,
        Label:   label,
        Style:   style,
        OnPress: func() { a.activeTab.Set(key) },
    }
}
```

## Nested Switchers

Switchers can be nested for hierarchical navigation:

```go
type App struct {
    mainView t.Signal[string]
    subView  t.Signal[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Switcher{
        Active: a.mainView.Get(),
        Children: map[string]t.Widget{
            "dashboard": DashboardView{},
            "settings": t.Switcher{
                Active: a.subView.Get(),
                Children: map[string]t.Widget{
                    "general":  GeneralSettings{},
                    "privacy":  PrivacySettings{},
                    "advanced": AdvancedSettings{},
                },
            },
        },
    }
}
```

## Dynamic Children

Children can be computed dynamically:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    children := map[string]t.Widget{
        "home": HomeView{},
    }

    // Conditionally add views
    if a.isAdmin.Get() {
        children["admin"] = AdminPanel{}
    }

    if a.hasSubscription.Get() {
        children["premium"] = PremiumContent{}
    }

    return t.Switcher{
        Active:   a.activeTab.Get(),
        Children: children,
    }
}
```

## Complete Example

A file manager with list and preview views:

```go
package main

import (
    "fmt"
    t "terma"
)

type App struct {
    view       t.Signal[string]
    files      []string
    listState  *t.ListState[string]
    selected   t.Signal[string]
}

func NewApp() *App {
    files := []string{"README.md", "main.go", "config.yaml", "Makefile"}
    return &App{
        view:      t.NewSignal("list"),
        files:     files,
        listState: t.NewListState(files),
        selected:  t.NewSignal(""),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Height: t.Flex(1),
        Children: []t.Widget{
            // Header
            t.Row{
                Style: t.Style{
                    BackgroundColor: theme.Surface,
                    Padding:         t.EdgeInsetsHV(1, 0),
                },
                Children: []t.Widget{
                    t.Text{Content: "File Manager"},
                    t.Spacer{},
                    t.Text{
                        Content: fmt.Sprintf("View: %s", a.view.Get()),
                        Style:   t.Style{ForegroundColor: theme.TextMuted},
                    },
                },
            },
            // Content
            t.Switcher{
                Active: a.view.Get(),
                Height: t.Flex(1),
                Children: map[string]t.Widget{
                    "list":    a.buildListView(ctx),
                    "preview": a.buildPreviewView(ctx),
                },
            },
        },
    }
}

func (a *App) buildListView(ctx t.BuildContext) t.Widget {
    return t.List[string]{
        State: a.listState,
        OnSelect: func(file string) {
            a.selected.Set(file)
            a.view.Set("preview")
        },
        RenderItem: func(file string, ctx t.ListItemContext) t.Widget {
            return t.Text{Content: file}
        },
    }
}

func (a *App) buildPreviewView(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()
    return t.Column{
        Style: t.Style{Padding: t.EdgeInsetsAll(1)},
        Children: []t.Widget{
            t.Text{
                Content: a.selected.Get(),
                Style:   t.Style{ForegroundColor: theme.Primary},
            },
            t.Spacer{Height: t.Cells(1)},
            t.Text{Content: "(File preview would appear here)"},
        },
    }
}

func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "l", Name: "List", Action: func() { a.view.Set("list") }},
        {Key: "p", Name: "Preview", Action: func() { a.view.Set("preview") }},
        {Key: "escape", Name: "Back", Action: func() { a.view.Set("list") }},
    }
}

func main() {
    t.Run(NewApp())
}
```

## Notes

- If `Active` doesn't match any key in `Children`, an empty widget is rendered
- Only the active child is built—inactive children don't consume resources
- State preservation requires storing state in the App, not in child widgets
- Keys are arbitrary strings—use descriptive names for readability
- Consider combining with `KeybindBar` to show available navigation keys
