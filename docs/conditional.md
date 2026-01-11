# Conditional Rendering

Terma provides functions and widgets for controlling when and how widgets are displayed. Use these to toggle visibility, show placeholder space, or switch between multiple views.

## ShowWhen / HideWhen

Control whether a widget is present in the layout. When hidden, the widget takes no space.

```go
// Show the error message only when there's an error
ShowWhen(a.error.Get() != "", Text{Content: a.error.Get()})

// Hide the loading spinner when data has loaded
HideWhen(a.loaded.Get(), Spinner{})
```

`ShowWhen` renders the child when the condition is `true`, otherwise renders nothing. `HideWhen` is the inverse—it hides the child when the condition is `true`.

Think of these like CSS `display: none`. The widget is completely removed from the layout.

## VisibleWhen / InvisibleWhen

Control visibility while preserving layout space. The widget's space is always reserved, but content only renders when visible.

```go
// Reserve space for a chart, render only when data exists
VisibleWhen(a.hasData.Get(), Chart{Data: a.chartData.Get()})

// Always hold space for the preview, hide when editing
InvisibleWhen(a.isEditing.Get(), Preview{Content: a.content.Get()})
```

Think of these like CSS `visibility: hidden`. The widget remains in the layout, but is not drawn.

### When to Use Each

| Function | Space Reserved | Use Case |
|----------|---------------|----------|
| `ShowWhen` | No | Toggle elements on/off |
| `HideWhen` | No | Inverse of ShowWhen |
| `VisibleWhen` | Yes | Placeholder UI, avoid layout shift |
| `InvisibleWhen` | Yes | Inverse of VisibleWhen |

## Switcher

Display one widget at a time from a collection, selected by a string key. Use `Switcher` for tabbed interfaces or multi-view applications.

```go
Switcher{
    Active: a.activeTab.Get(),
    Children: map[string]Widget{
        "home":     HomeView{},
        "settings": SettingsView{},
        "profile":  ProfileView{},
    },
}
```

Only the active child is built and rendered. Inactive children don't exist in the widget tree.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Active` | `string` | Key of the currently visible child |
| `Children` | `map[string]Widget` | Map of keys to widgets |
| `Width` | `Dimension` | Container width (optional) |
| `Height` | `Dimension` | Container height (optional) |
| `Style` | `Style` | Container styling (optional) |

### State Preservation

Widget state persists across switches because state lives in the App, not in widgets. Store Signals and State objects in your App struct:

```go
type App struct {
    activeTab Signal[string]

    // State for each tab
    homeListState    *ListState[string]
    settingsForm     Signal[FormData]
    profileScrollPos *ScrollState
}

func (a *App) Build(ctx BuildContext) Widget {
    return Switcher{
        Active: a.activeTab.Get(),
        Children: map[string]Widget{
            "home":     HomeView{ListState: a.homeListState},
            "settings": SettingsView{Form: a.settingsForm},
            "profile":  ProfileView{ScrollState: a.profileScrollPos},
        },
    }
}
```

When a user switches from "home" to "settings" and back, the list selection in `homeListState` is preserved.

### Tab Navigation with Keybinds

Combine Switcher with keybinds for keyboard-driven tab switching:

```go
func (a *App) Keybinds() []Keybind {
    return []Keybind{
        {Key: "1", Name: "Home", Action: func() { a.activeTab.Set("home") }},
        {Key: "2", Name: "Settings", Action: func() { a.activeTab.Set("settings") }},
        {Key: "3", Name: "Profile", Action: func() { a.activeTab.Set("profile") }},
    }
}
```

## Complete Example

A multi-tab application with conditional status display:

```go
package main

import (
    "fmt"
    t "terma"
)

type App struct {
    activeTab    t.Signal[string]
    itemsLoaded  t.Signal[bool]
    selectedItem t.Signal[string]
    counter      t.Signal[int]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Height: t.Flex(1),
        Children: []t.Widget{
            // Tab bar
            t.Row{
                Spacing: 2,
                Style: t.Style{
                    BackgroundColor: theme.Surface,
                    Padding: t.EdgeInsetsAll(1),
                },
                Children: []t.Widget{
                    a.tabButton("items", "Items"),
                    a.tabButton("counter", "Counter"),
                },
            },

            // Content area
            t.Switcher{
                Active: a.activeTab.Get(),
                Height: t.Flex(1),
                Children: map[string]t.Widget{
                    "items":   a.buildItemsTab(ctx),
                    "counter": a.buildCounterTab(ctx),
                },
            },

            // Status bar - only shows when an item is selected
            t.ShowWhen(a.selectedItem.Get() != "", t.Text{
                Content: fmt.Sprintf("Selected: %s", a.selectedItem.Get()),
                Style: t.Style{
                    BackgroundColor: theme.Primary,
                    ForegroundColor: theme.Primary.AutoText(),
                    Padding: t.EdgeInsetsHV(1, 0),
                },
            }),
        },
    }
}

func (a *App) tabButton(key, label string) t.Widget {
    return t.Button{
        ID:      key,
        Label:   label,
        OnPress: func() { a.activeTab.Set(key) },
    }
}

func (a *App) buildItemsTab(ctx t.BuildContext) t.Widget {
    return t.Column{
        Children: []t.Widget{
            // Show loading message until items are loaded
            t.HideWhen(a.itemsLoaded.Get(), t.Text{Content: "Loading..."}),

            // Show items when loaded
            t.ShowWhen(a.itemsLoaded.Get(), t.Text{Content: "Items loaded!"}),
        },
    }
}

func (a *App) buildCounterTab(ctx t.BuildContext) t.Widget {
    return t.Column{
        Children: []t.Widget{
            t.Text{Content: fmt.Sprintf("Count: %d", a.counter.Get())},
            t.Button{
                ID:      "increment",
                Label:   "Increment",
                OnPress: func() { a.counter.Set(a.counter.Get() + 1) },
            },
        },
    }
}

func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "1", Name: "Items", Action: func() { a.activeTab.Set("items") }},
        {Key: "2", Name: "Counter", Action: func() { a.activeTab.Set("counter") }},
    }
}

func main() {
    app := &App{
        activeTab:    t.NewSignal("items"),
        itemsLoaded:  t.NewSignal(false),
        selectedItem: t.NewSignal(""),
        counter:      t.NewSignal(0),
    }
    t.Run(app)
}
```

## Comparison

| Approach | Best For |
|----------|----------|
| `ShowWhen`/`HideWhen` | Simple toggles, conditional elements |
| `VisibleWhen`/`InvisibleWhen` | Avoiding layout shift, placeholder UI |
| `Switcher` | Tabs, multi-view apps, complex state preservation |
