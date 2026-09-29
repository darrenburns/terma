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
| `ScrollbarThumbColor` | `Color` | Theme `ScrollbarThumb` (`Primary` when focused) | Scrollbar thumb color |
| `ScrollbarTrackColor` | `Color` | Theme `ScrollbarTrack` | Scrollbar track color |
| `Click` | `func(MouseEvent)` | — | Click callback |
| `MouseWheel` | `func(MouseEvent) bool` | — | Return true to consume before normal scrolling |
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

The thumb is placed to an eighth of a cell. Cells it partly covers use the
lower block characters `▁ ▂ ▃ ▄ ▅ ▆ ▇`; cells it fills are drawn as background
colour. Its length depends only on the viewport and content heights, so it
stays exactly the same size as it moves.

## Smooth Scrolling

Content can only be drawn at whole lines, but the thumb is drawn at the exact
scroll position, which can fall between lines. Dragging the thumb keeps it under
the point where it was grabbed while the content follows to the nearest line.
Moving the thumb within a line repaints only the scrollbar.

In terminals that support SGR-Pixels mouse reporting (such as kitty and
Ghostty), Terma reads the pointer in pixels, so the thumb follows the pointer to
an eighth of a cell rather than jumping a cell at a time. The mode is switched on
only when the terminal confirms it supports it and reports the size of its cells
in pixels (`CSI 16 t`), which is asked for again whenever the window resizes;
set `TERMA_DISABLE_PIXEL_MOUSE=1` to keep cell-based reporting. Widgets can use
the pointer's position within its cell from `MouseEvent.SubCellX` and
`MouseEvent.SubCellY`.

## Pinning to the Bottom

With `PinToBottom` set, a `ScrollState` scrolled to the bottom stays there as its
content grows or shrinks, showing the new end in the same frame. Scrolling up
releases the pin and scrolling back to the bottom restores it. Content that is
taller than the viewport when first shown starts at the top.

## Notes

- `State` is required - create with `NewScrollState()`
- Set an `ID` to enable keyboard focus and navigation
- The scrollbar occupies 1 cell on the right edge
- Content is clipped to the viewport bounds
- Inside `Scrollable`, `Height: Flex(...)` and `Height: Percent(...)` are resolved against the visible viewport height
- Scroll offset is automatically clamped to valid bounds
