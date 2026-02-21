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

## Blur Semantics

`HoverLeave` is a pointer leave transition, not keyboard focus blur.
Keyboard focus blur remains `Blurrable.OnBlur()` and is unchanged.
