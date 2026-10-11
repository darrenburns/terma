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

## Keybindings

Widgets declare keybindings by implementing `KeybindProvider`. A key press goes
to the focused widget first, then bubbles up through its ancestors to the root;
the first widget with a matching binding handles it.

```go
func (e *Editor) Keybinds() []Keybind {
    return []Keybind{
        {Key: "ctrl+s", Name: "Save", Action: e.save},
        {Key: "d", Name: "Delete", Action: e.delete},
        {Key: "D", Name: "Delete all", Action: e.deleteAll}, // shift+d
    }
}
```

A character key matches its case: `"D"` is shift+d and `"d"` is d. The
keybind bar shows each one as written.

### Chords

A chord is several key presses, one after the other. Write its keys separated
by spaces. Inside a chord, the space key is written `space`:

```go
{Key: "ctrl+k ctrl+s", Name: "Save all", Action: e.saveAll},
{Key: "ctrl+k w", Name: "Close all", Action: e.closeAll},
{Key: "g g", Name: "Top", Action: e.top},
{Key: "space f", Name: "Find file", Action: e.find},
```

When a key press starts a chord, Terma waits for the next key, with no timeout.
That key goes only to the bindings that can continue the chord, even if a text
input is focused. `escape` cancels the chord. So does a key that continues no
chord, and that key is then dropped. A mouse click also cancels it.

A chord wins over a binding to its own start: if one widget binds both `g` and
`g g`, pressing g waits for the next key, and `g` never runs. Precedence across
widgets is unchanged. If the focused widget binds `ctrl+k` (as `TextInput`
does), an ancestor's `ctrl+k …` chords can't start while it's focused.

### Group labels

A binding with no `Action` whose key starts other chords labels them as a
group. The keybind bar shows the group as one hint instead of every chord in it:

```go
{Key: "ctrl+k", Name: "Window"}, // shown as "ctrl+k Window…"
{Key: "ctrl+k ctrl+s", Name: "Save all", Action: e.saveAll},
{Key: "ctrl+k w", Name: "Close all", Action: e.closeAll},
```

Making the label `Hidden` hides the whole group.

### Keybind bar

`KeybindBar` lists the bindings the focused widget can reach. Bindings that can
never run are left out: those shadowed by an earlier widget, those whose key
the focused widget types as text, and those whose key starts a chord in the
same widget. While a chord is in progress, the bar shows it, then the keys that
can come next, then how to cancel:

```
ctrl+k … ctrl+s Save all w Close all escape Cancel
```

`FormatKey` changes how keys are written. The presets keep a character's case
and handle chords: `FormatKeyVerbose` (`Ctrl+k Ctrl+s`), `FormatKeyCaret`
(`^k ^s`, and `^D` only for an uppercase D), `FormatKeyEmacs` (`C-k C-s`) and
`FormatKeyVim` (`<C-k><C-s>`, `gg`). A custom `FormatKey` receives the whole
key, chords included. Widgets other than the bar can read the same information
with `ctx.ActiveKeybinds()` and `ctx.PendingChord()`.

## Focus in snapshots

Outside a running app, call `RequestFocus` with a stable widget ID immediately
before `RenderToBuffer`, `Snapshot`, or `AssertSnapshot` to capture its focused
appearance. A wrapper that requests focus from `Build` is unnecessary:

```go
terma.RequestFocus("summary-menu")
terma.AssertSnapshot(t, widget, 80, 24)
```

Each headless render consumes its request, including an empty or invalid ID;
without a valid request the first focusable widget receives focus. Open modals
still pull focus inside, and a request for a widget within the modal wins over
its default first focusable. Requests made during rendering do not carry into
the next snapshot. When an app is running, its pending request is preserved and
is not used by a headless snapshot.

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

## Mouse Wheel Events

Any custom widget can implement a consumable wheel handler:

```go
type MouseWheelHandler interface {
    OnMouseWheel(event MouseEvent) bool
}
```

Wheel input starts at the topmost visible widget under the pointer and bubbles
through its ancestors, innermost first. A handler returning `true` consumes the
event. If it returns `false`, normal `Scrollable` behavior at that widget runs
before the event continues to its parent. Scrollable viewports that reach their
limit still allow an outer viewport or handler to consume the event. Covered
siblings, clipped widgets, disabled handlers, and widgets beneath an overlay do
not receive the event. Wheel handling does not require keyboard focus.

`MouseEvent.Button` identifies `uv.MouseWheelUp`, `uv.MouseWheelDown`,
`uv.MouseWheelLeft`, or `uv.MouseWheelRight`. The event includes absolute and
widget-local coordinates, modifiers, widget ID, and sub-cell pointer position;
`ClickCount` is zero.

`Text`, `Row`, `Column`, `Stack`, `TabBar`, and `Scrollable` also expose a
`MouseWheel func(MouseEvent) bool` callback. For example, a one-row tab bar can
switch tabs directly, without artificial scroll overflow or a scrollbar:

```go
TabBar{
    ID: "tabs",
    State: tabs,
    MouseWheel: func(event MouseEvent) bool {
        switch event.Button {
        case uv.MouseWheelDown, uv.MouseWheelRight:
            tabs.SelectNext()
        case uv.MouseWheelUp, uv.MouseWheelLeft:
            tabs.SelectPrevious()
        default:
            return false
        }
        return true
    },
}
```

Run `go run ./cmd/wheel-tabs-demo` for a working example.
