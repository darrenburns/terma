# Tabs

A horizontal tab bar for switching between views with keyboard navigation, reordering, and closable tabs.

=== "Demo"

    <video autoplay loop muted playsinline src="../../assets/tabs-demo.mp4"></video>

=== "Code"

    ```go
    --8<-- "cmd/tab-example/main.go"
    ```

## Overview

The tab system consists of three main components:

- **`Tab`**: A struct representing a single tab (key, label, optional content)
- **`TabState`**: Manages the list of tabs and tracks the active tab
- **`TabBar`**: A focusable widget that renders tabs horizontally
- **`TabView`**: A convenience widget combining TabBar with content switching

```go
--8<-- "docs/minimal-examples/tabs-basic/main.go"
```

## Tab Struct

| Field | Type | Description |
|-------|------|-------------|
| `Key` | `string` | Unique identifier for switching |
| `Label` | `string` | Display text (can differ from Key) |
| `Content` | `Widget` | Optional content widget (used by TabView) |

## TabBar Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Required for focus management |
| `DisableFocus` | `bool` | `false` | Prevent keyboard focus |
| `State` | `*TabState` | — | Required - holds tabs and active key |
| `KeybindPattern` | `TabKeybindPattern` | `TabKeybindNone` | Position-based keybind style |
| `OnTabChange` | `func(key string)` | `nil` | Called when active tab changes |
| `OnTabClose` | `func(key string)` | `nil` | Called when close button is clicked |
| `Closable` | `bool` | `false` | Show close buttons on tabs |
| `AllowReorder` | `bool` | `false` | Enable ctrl+h/l tab reordering |
| `Width` | `Dimension` | `Auto` | Bar width |
| `Height` | `Dimension` | `Auto` | Bar height |
| `Style` | `Style` | — | Container style |
| `TabStyle` | `Style` | — | Inactive tab style |
| `ActiveTabStyle` | `Style` | — | Active tab style |

## TabView Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `State` | `*TabState` | — | Required - holds tabs and active key |
| `KeybindPattern` | `TabKeybindPattern` | `TabKeybindNone` | Position-based keybind style |
| `OnTabChange` | `func(key string)` | `nil` | Called when active tab changes |
| `OnTabClose` | `func(key string)` | `nil` | Called when close button is clicked |
| `Closable` | `bool` | `false` | Show close buttons on tabs |
| `AllowReorder` | `bool` | `false` | Enable ctrl+h/l tab reordering |
| `Width` | `Dimension` | `Auto` | View width |
| `Height` | `Dimension` | `Auto` | View height |
| `Style` | `Style` | — | Container style |
| `TabBarStyle` | `Style` | — | Style for the tab bar row |
| `TabStyle` | `Style` | — | Inactive tab style |
| `ActiveTabStyle` | `Style` | — | Active tab style |
| `ContentStyle` | `Style` | — | Style for the content area |

## TabState Methods

| Method | Description |
|--------|-------------|
| `NewTabState(tabs []Tab)` | Create state with tabs (first becomes active) |
| `NewTabStateWithActive(tabs []Tab, key string)` | Create state with specific active tab |
| `Tabs()` | Get all tabs (subscribes to changes) |
| `TabsPeek()` | Get all tabs without subscribing |
| `ActiveKey()` | Get active tab key (subscribes to changes) |
| `ActiveKeyPeek()` | Get active tab key without subscribing |
| `SetActiveKey(key string)` | Set active tab by key |
| `ActiveIndex()` | Get index of active tab (-1 if not found) |
| `ActiveTab()` | Get active tab pointer (nil if not found) |
| `SelectNext()` | Move to next tab (wraps around) |
| `SelectPrevious()` | Move to previous tab (wraps around) |
| `SelectIndex(index int)` | Set active tab by index |
| `AddTab(tab Tab)` | Add tab to end of list |
| `InsertTab(index int, tab Tab)` | Insert tab at specific index |
| `RemoveTab(key string)` | Remove tab by key (returns success) |
| `MoveTabLeft(key string)` | Move tab one position left |
| `MoveTabRight(key string)` | Move tab one position right |
| `SetLabel(key, label string)` | Update a tab's label |
| `TabCount()` | Get number of tabs |

## Basic Usage

Create a `TabState` with your tabs and pass it to `TabBar`:

```go
type App struct {
    tabState *t.TabState
}

func NewApp() *App {
    return &App{
        tabState: t.NewTabState([]t.Tab{
            {Key: "home", Label: "Home"},
            {Key: "settings", Label: "Settings"},
            {Key: "help", Label: "Help"},
        }),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Column{
        Children: []t.Widget{
            t.TabBar{
                ID:    "tabs",
                State: a.tabState,
            },
            // Your content area here
        },
    }
}
```

## Keyboard Navigation

TabBar is focusable and provides built-in keyboard navigation:

| Key | Action |
|-----|--------|
| ++h++ / ++left++ | Previous tab |
| ++l++ / ++right++ | Next tab |
| ++ctrl+h++ | Move active tab left (if `AllowReorder` enabled) |
| ++ctrl+l++ | Move active tab right (if `AllowReorder` enabled) |
| ++ctrl+w++ | Close active tab (if `Closable` enabled) |

## Position-Based Keybindings

Enable direct tab jumping with number keys using `KeybindPattern`:

```go
t.TabBar{
    ID:             "tabs",
    State:          a.tabState,
    KeybindPattern: t.TabKeybindNumbers,  // 1, 2, 3... 9
}
```

Three patterns are available:

| Pattern | Keys | Example |
|---------|------|---------|
| `TabKeybindNone` | Disabled | — |
| `TabKeybindNumbers` | 1-9 | Press ++1++ for first tab |
| `TabKeybindAltNumbers` | Alt+1-9 | Press ++alt+1++ for first tab |
| `TabKeybindCtrlNumbers` | Ctrl+1-9 | Press ++ctrl+1++ for first tab |

Position keybinds are hidden from KeybindBar to avoid clutter.

## Closable Tabs

Enable close buttons with the `Closable` field:

```go
t.TabBar{
    ID:       "tabs",
    State:    a.tabState,
    Closable: true,
    OnTabClose: func(key string) {
        // Custom close handling
        a.tabState.RemoveTab(key)
    },
}
```

If `OnTabClose` is not provided, `RemoveTab` is called automatically. When the active tab is closed, focus moves to an adjacent tab.

## Tab Reordering

Allow users to reorder tabs with `AllowReorder`:

```go
t.TabBar{
    ID:           "tabs",
    State:        a.tabState,
    AllowReorder: true,  // Enables ctrl+h/l reordering
}
```

Programmatic reordering is also available via `TabState`:

```go
a.tabState.MoveTabLeft("settings")
a.tabState.MoveTabRight("home")
```

## TabView

`TabView` combines a `TabBar` with automatic content switching based on each tab's `Content` field:

```go
type App struct {
    tabState *t.TabState
}

func NewApp() *App {
    return &App{
        tabState: t.NewTabState([]t.Tab{
            {Key: "home", Label: "Home", Content: HomeView{}},
            {Key: "settings", Label: "Settings", Content: SettingsView{}},
        }),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.TabView{
        ID:             "app-tabs",
        State:          a.tabState,
        KeybindPattern: t.TabKeybindNumbers,
        Height:         t.Flex(1),
    }
}
```

## Manual Content Switching

For more control over content rendering, use `TabBar` with a `Switcher` or manual switch:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Column{
        Children: []t.Widget{
            t.TabBar{
                ID:    "tabs",
                State: a.tabState,
            },
            t.Switcher{
                Active: a.tabState.ActiveKey(),
                Children: map[string]t.Widget{
                    "home":     HomeView{},
                    "settings": SettingsView{},
                },
                Height: t.Flex(1),
            },
        },
    }
}
```

## Styling

Customize tab appearance with `TabStyle` and `ActiveTabStyle`:

```go
t.TabBar{
    ID:    "tabs",
    State: a.tabState,
    Style: t.Style{
        BackgroundColor: theme.Surface,
    },
    TabStyle: t.Style{
        ForegroundColor: theme.TextMuted,
        BackgroundColor: theme.Surface,
        Padding:         t.EdgeInsetsXY(2, 0),
    },
    ActiveTabStyle: t.Style{
        ForegroundColor: theme.Background,
        BackgroundColor: theme.Accent,
        Padding:         t.EdgeInsetsXY(2, 0),
    },
}
```

Default styling uses theme colors:

- Active tab: `theme.Background` foreground on `theme.Accent` background
- Inactive tab: `theme.TextMuted` foreground on `theme.Surface` background

## Dynamic Tabs

Add and remove tabs programmatically:

```go
// Add a new tab
a.tabState.AddTab(t.Tab{
    Key:   "new-tab",
    Label: "New Tab",
})

// Insert at specific position
a.tabState.InsertTab(1, t.Tab{
    Key:   "inserted",
    Label: "Inserted",
})

// Remove a tab
a.tabState.RemoveTab("settings")

// Update a tab's label
a.tabState.SetLabel("home", "Dashboard")
```

## Tab Change Callback

React to tab changes with `OnTabChange`:

```go
t.TabBar{
    ID:    "tabs",
    State: a.tabState,
    OnTabChange: func(key string) {
        // Load data for the new tab
        t.Log("Switched to tab: %s", key)
    },
}
```

## Notes

- TabBar requires an `ID` for focus management
- Tab keys must be unique within a TabState
- Navigation wraps around (after last tab comes first)
- When removing the active tab, focus moves to an adjacent tab
- Position keybinds support tabs 1-9 only
- State is preserved across tab switches when using Signals and State objects held by the App
