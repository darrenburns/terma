# Scrollable

A container that enables vertical scrolling when content exceeds the viewport. Displays a scrollbar and supports keyboard navigation.

```go
scrollState := NewScrollState()

Scrollable{
    ID:     "content",
    State:  scrollState,
    Height: Cells(15),
    Child:  LongContent{},
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Identifier (required for focus) |
| `Child` | `Widget` | — | The content to scroll |
| `State` | `*ScrollState` | — | Required scroll state |
| `DisableScroll` | `bool` | `false` | Disable scrolling and hide scrollbar |
| `Focusable` | `bool` | `false` | Allow keyboard focus for scroll navigation |
| `DisableFocus` | `bool` | `false` | Prevent keyboard focus |
| `Width` | `Dimension` | — | Container width |
| `Height` | `Dimension` | — | Container height |
| `Style` | `Style` | — | Padding, border, colors |
| `ScrollbarThumbColor` | `Color` | White/BrightCyan | Scrollbar thumb color |
| `ScrollbarTrackColor` | `Color` | BrightBlack | Scrollbar track color |
| `Click` | `func(MouseEvent)` | — | Click callback |
| `MouseDown` | `func(MouseEvent)` | — | Mouse down callback |
| `MouseUp` | `func(MouseEvent)` | — | Mouse up callback |
| `Hover` | `func(HoverEvent)` | — | Hover transition callback |

## ScrollState

Manages scroll position and provides scroll control methods.

```go
state := NewScrollState()
```

### Methods

| Method | Description |
|--------|-------------|
| `GetOffset()` | Get current scroll offset |
| `SetOffset(n)` | Set scroll offset (auto-clamps to bounds) |
| `ScrollUp(n)` | Scroll up by n lines |
| `ScrollDown(n)` | Scroll down by n lines |
| `ScrollToView(y, height)` | Ensure a region is visible |

### Reactive Offset

The scroll offset is a Signal, so reading it in `Build()` subscribes to changes:

```go
func (a *App) Build(ctx BuildContext) Widget {
    offset := a.scrollState.Offset.Get()  // Subscribes to changes

    return Column{
        Children: []Widget{
            Text{Content: fmt.Sprintf("Scroll: %d", offset)},
            Scrollable{
                State: a.scrollState,
                Child: Content{},
            },
        },
    }
}
```

## Keyboard Navigation

When focused (requires an ID), Scrollable responds to these keys:

| Key | Action |
|-----|--------|
| `↑` / `k` | Scroll up 1 line |
| `↓` / `j` | Scroll down 1 line |
| `PageUp` / `Ctrl+U` | Scroll up half viewport |
| `PageDown` / `Ctrl+D` | Scroll down half viewport |
| `Home` / `g` | Scroll to top |
| `End` / `G` | Scroll to bottom |

## Examples

### Basic Scrollable Content

```go
type App struct {
    scrollState *ScrollState
}

func NewApp() *App {
    return &App{
        scrollState: NewScrollState(),
    }
}

func (a *App) Build(ctx BuildContext) Widget {
    // Generate many items
    var items []Widget
    for i := 0; i < 50; i++ {
        items = append(items, Text{
            Content: fmt.Sprintf("Item %d", i+1),
        })
    }

    return Scrollable{
        ID:     "list",
        State:  a.scrollState,
        Height: Cells(15),
        Style: Style{
            Border:      BorderRounded,
            BorderColor: ctx.Theme().TextMuted,
            Padding:     EdgeInsetsAll(1),
        },
        Child: Column{Children: items},
    }
}
```

### With Custom Scrollbar Colors

```go
Scrollable{
    ID:                  "content",
    State:               scrollState,
    Height:              Flex(1),
    ScrollbarThumbColor: theme.Primary,
    ScrollbarTrackColor: theme.Background,
    Child:               Content{},
}
```

### Disabled Scrolling

Use `DisableScroll` to show content without scrolling capability:

```go
Scrollable{
    State:         scrollState,
    Height:        Cells(5),
    DisableScroll: true,  // No scrollbar, no scroll
    Child:         Preview{},
}
```

### Scrollable with List

When using with `List`, share the `ScrollState` for coordinated scrolling:

```go
type App struct {
    scrollState *ScrollState
    listState   *ListState[string]
}

func (a *App) Build(ctx BuildContext) Widget {
    return Scrollable{
        ID:     "scroll-list",
        State:  a.scrollState,
        Height: Flex(1),
        Child: List[string]{
            State:       a.listState,
            ScrollState: a.scrollState,  // Share state
            RenderItem: func(item string, idx int, focused bool) Widget {
                return Text{Content: item}
            },
        },
    }
}
```

### Programmatic Scrolling

Control scroll position from code:

```go
func (a *App) Keybinds() []Keybind {
    return []Keybind{
        {Key: "t", Name: "Top", Action: func() {
            a.scrollState.SetOffset(0)
        }},
        {Key: "b", Name: "Bottom", Action: func() {
            a.scrollState.SetOffset(9999)  // Clamps to max
        }},
        {Key: "m", Name: "Middle", Action: func() {
            // ScrollToView ensures a region is visible
            a.scrollState.ScrollToView(25, 1)
        }},
    }
}
```

### In a Dock Layout

```go
Dock{
    Top: []Widget{Header{}},
    Bottom: []Widget{KeybindBar{}},
    Body: Scrollable{
        ID:     "main",
        State:  scrollState,
        Height: Flex(1),
        Child:  MainContent{},
    },
}
```

## Scrollbar Rendering

The scrollbar uses Unicode block characters for smooth sub-cell precision:

```
▁ ▂ ▃ ▄ ▅ ▆ ▇ █
```

This allows the scrollbar thumb to smoothly track scroll position even with small viewports.

## Notes

- `State` is required - create with `NewScrollState()`
- Set an `ID` to enable keyboard focus and navigation
- The scrollbar occupies 1 cell on the right edge
- Content is clipped to the viewport bounds
- Scroll offset is automatically clamped to valid bounds
