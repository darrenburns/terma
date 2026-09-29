# Focus & Keyboard

Terma provides focus management and a declarative keybinding system.
Focusable widgets can receive keyboard events, and keybindings can be
automatically displayed in a keybind bar.

## Keyboard Focus

Widgets that implement `Focusable` receive key events when focused:

```go
type Focusable interface {
    OnKey(event KeyEvent) bool
    IsFocusable() bool
}
```

When focus moves away from a widget, Terma calls `OnBlur()` for widgets that
implement `Blurrable`.

## Pointer Hover Events

Terma also supports first-class hover transition events with event payloads:

```go
type HoverEventType int

const (
    HoverEnter HoverEventType = iota
    HoverLeave
)

type HoverEvent struct {
    Type             HoverEventType
    Source           HoverEventSource
    X, Y             int
    LocalX, LocalY   int
    Button           uv.MouseButton
    Mod              uv.KeyMod
    WidgetID         string
    PreviousWidgetID string
    NextWidgetID     string
}

type Hoverable interface {
    OnHover(event HoverEvent)
}
```

Hover transitions are direct-target only (no bubbling).

`event.Source` is `HoverSourcePointer` for mouse motion input (including reports
that change only buttons or modifiers), or `HoverSourceLayout` when rendering
changes the target under the last known pointer position. Layout transitions still
update `ctx.HoveredID()` and dispatch leave before enter, so hover styling remains
correct when a widget appears, disappears, or moves under a stationary pointer.
`Button` and `Mod` on layout events retain the last recorded motion report's state.
Unchanged target identity emits no transition; moving within the same widget does
not generate additional enter events. `HoverSourcePointer` is the zero value.

To dismiss a keyboard-opened summary only on pointer input, filter its callback:

```go
Hover: func(event terma.HoverEvent) {
    if event.Type == terma.HoverEnter && event.Source == terma.HoverSourcePointer {
        showSummary.Set(false)
    }
},
```

See `go run ./cmd/hover-cause-demo` for a summary that remains visible when opened
under a stationary mouse and dismisses after the mouse leaves and re-enters it.

## Blur Semantics

`HoverLeave` is a pointer leave transition, not keyboard focus blur.
Keyboard focus blur remains `Blurrable.OnBlur()` and is unchanged.
