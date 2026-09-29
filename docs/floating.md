# Floating Layouts

Floating widgets render as overlays on top of other content. Use them for dropdowns, modals, tooltips, and notifications.

## Basic Usage

A `Floating` widget renders its child above the main widget tree when `Visible` is true:

```go
Floating{
    Visible: a.showMenu.Get(),
    Config: FloatConfig{
        Position:  FloatPositionCenter,
        OnDismiss: func() { a.showMenu.Set(false) },
    },
    Child: Menu{Items: menuItems},
}
```

## Floating Fields

| Field | Type | Description |
|-------|------|-------------|
| `Visible` | `bool` | Whether the floating widget is shown |
| `Config` | `FloatConfig` | Positioning and behavior options |
| `Child` | `Widget` | The widget to render as an overlay |

## Positioning

Floating widgets support two positioning modes: absolute (screen-based) and anchor-based (relative to another widget).

### Absolute Positioning

Position the float at a fixed screen location:

```go
Floating{
    Visible: show.Get(),
    Config: FloatConfig{
        Position: FloatPositionCenter,  // Centered on screen
    },
    Child: Dialog{},
}
```

**Available positions:**

| Position | Description |
|----------|-------------|
| `FloatPositionCenter` | Center of the screen |
| `FloatPositionTopLeft` | Top left corner |
| `FloatPositionTopCenter` | Top center |
| `FloatPositionTopRight` | Top right corner |
| `FloatPositionBottomLeft` | Bottom left corner |
| `FloatPositionBottomCenter` | Bottom center |
| `FloatPositionBottomRight` | Bottom right corner |
| `FloatPositionAbsolute` | Uses `Offset` as absolute coordinates |

### Anchor-Based Positioning

Attach the float to another widget using its ID:

```go
// The button to anchor to
Button{
    ID:      "user-menu",
    Label:   "Menu",
    OnPress: func() { a.showMenu.Set(true) },
}

// The floating menu
Floating{
    Visible: a.showMenu.Get(),
    Config: FloatConfig{
        AnchorID:  "user-menu",      // ID of the anchor widget
        Anchor:    AnchorBottomLeft, // Attach below the button
        OnDismiss: func() { a.showMenu.Set(false) },
    },
    Child: DropdownMenu{},
}
```

**Anchor points:**

| Anchor | Description |
|--------|-------------|
| `AnchorUnset` | Use widget default (menus: below anchor, aligned left) |
| `AnchorTopLeft` | Above anchor, aligned left |
| `AnchorTopCenter` | Above anchor, centered |
| `AnchorTopRight` | Above anchor, aligned right |
| `AnchorBottomLeft` | Below anchor, aligned left |
| `AnchorBottomCenter` | Below anchor, centered |
| `AnchorBottomRight` | Below anchor, aligned right |
| `AnchorLeftTop` | Left of anchor, aligned top |
| `AnchorLeftCenter` | Left of anchor, centered |
| `AnchorLeftBottom` | Left of anchor, aligned bottom |
| `AnchorRightTop` | Right of anchor, aligned top |
| `AnchorRightCenter` | Right of anchor, centered |
| `AnchorRightBottom` | Right of anchor, aligned bottom |

### Offset Adjustment

Fine-tune position with an offset:

```go
Floating{
    Visible: show.Get(),
    Config: FloatConfig{
        Position: FloatPositionTopCenter,
        Offset:   Offset{X: 0, Y: 2},  // Move down 2 cells
    },
    Child: Notification{},
}
```

## FloatConfig Reference

```go
type FloatConfig struct {
    // Anchor-based positioning
    AnchorID string      // ID of widget to anchor to
    Anchor   AnchorPoint // Where on anchor to attach

    // Absolute positioning (when AnchorID is empty)
    Position FloatPosition

    // Position adjustment
    Offset Offset

    // Modal behavior
    Modal         bool  // Show backdrop, trap focus
    BackdropColor Color // Backdrop color (default: semi-transparent black)

    // Mouse input
    PointerPassthrough bool // Ignore pointer input; Modal takes precedence

    // Dismissal
    OnDismiss             func() // Called when float should close
    DismissOnEsc          *bool  // Dismiss on Escape (default: true if OnDismiss set)
    DismissOnClickOutside *bool  // Dismiss on outside click (default: true for non-modal)
}
```

## Pointer-Pass-Through Previews

Set `PointerPassthrough: true` for a visual preview or tooltip that should allow
mouse input to reach the widgets beneath it:

```go
Floating{
    Visible: showPreview.Get(),
    Config: FloatConfig{
        AnchorID: "preview-trigger",
        Anchor: AnchorBottomLeft,
        PointerPassthrough: true,
    },
    Child: Text{Content: "Preview"},
}
```

The entire float subtree is skipped for clicks, dragging, wheel scrolling and
hover, including focus changes caused by a click. Input reaches the next
interactive float beneath it, or the main widget tree. Clicks and wheel events
use their own coordinates, so no preceding mouse movement is needed. Showing
the preview also preserves hover beneath a stationary pointer.

A pass-through float does not consume outside clicks or run outside-click
dismissal, even when `OnDismiss` is set. Escape dismissal and keyboard focus
behavior are unchanged. Use display-only children for previews; focusable
children still participate in keyboard navigation. `Modal: true` overrides
`PointerPassthrough`, so modal content and its backdrop still receive or block
mouse input. The default `false` preserves existing interactive float behavior.

Try `go run ./cmd/float-passthrough-demo`; add `-blocking` to compare the default
behavior. For browser verification, launch:

```sh
go run ./cmd/terma-browser -- go run ./cmd/float-passthrough-demo
```

Click the preview heading to increment the underlying button's click count;
hover over it to see `Hover: button`. Scroll over the second preview line to
move the rows (the right-hand column remains visible). The automated regression
also sends clicks and wheel events without preceding mouse movement.

## Modal Dialogs

Set `Modal: true` for dialog behavior with a backdrop:

```go
Floating{
    Visible: a.showConfirm.Get(),
    Config: FloatConfig{
        Position:  FloatPositionCenter,
        Modal:     true,
        OnDismiss: func() { a.showConfirm.Set(false) },
    },
    Child: Column{
        Style: Style{
            BackgroundColor: theme.Surface,
            Padding:         EdgeInsetsAll(2),
            Border:          BorderRounded,
            BorderColor:     theme.Primary,
        },
        Children: []Widget{
            Text{Content: "Delete this item?"},
            Spacer{Height: Cells(1)},
            Row{
                Spacing: 2,
                Children: []Widget{
                    Button{ID: "cancel", Label: "Cancel", OnPress: func() {
                        a.showConfirm.Set(false)
                    }},
                    Button{ID: "confirm", Label: "Delete", OnPress: func() {
                        a.deleteItem()
                        a.showConfirm.Set(false)
                    }},
                },
            },
        },
    },
}
```

### Custom Backdrop Color

```go
Config: FloatConfig{
    Position:      FloatPositionCenter,
    Modal:         true,
    BackdropColor: theme.Background.WithAlpha(0.7),  // More opaque
}
```

## Dropdown Menus

Combine anchor positioning with a list for dropdown menus:

```go
type App struct {
    showDropdown t.Signal[bool]
    menuItems    *t.ListState[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Children: []t.Widget{
            // Trigger button
            t.Button{
                ID:      "dropdown-trigger",
                Label:   "Options",
                OnPress: func() { a.showDropdown.Set(true) },
            },

            // Dropdown menu
            t.Floating{
                Visible: a.showDropdown.Get(),
                Config: t.FloatConfig{
                    AnchorID:  "dropdown-trigger",
                    Anchor:    t.AnchorBottomLeft,
                    OnDismiss: func() { a.showDropdown.Set(false) },
                },
                Child: t.Column{
                    Width: t.Cells(20),
                    Style: t.Style{
                        BackgroundColor: theme.Surface,
                        Border:          t.BorderRounded,
                        BorderColor:     theme.TextMuted,
                    },
                    Children: []t.Widget{
                        t.List[string]{
                            State: a.menuItems,
                            RenderItem: func(item string, idx int, isFocused bool) t.Widget {
                                style := t.Style{Padding: t.EdgeInsetsHV(1, 0)}
                                if isFocused {
                                    style.BackgroundColor = theme.Primary
                                    style.ForegroundColor = theme.Primary.AutoText()
                                }
                                return t.Text{Content: item, Style: style}
                            },
                            OnSelect: func(item string) {
                                a.handleSelection(item)
                                a.showDropdown.Set(false)
                            },
                        },
                    },
                },
            },
        },
    }
}
```

## Tooltips

Position tooltips relative to their trigger:

```go
type App struct {
    hoveredButton t.Signal[string]
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Children: []t.Widget{
            t.Button{
                ID:    "help-btn",
                Label: "?",
                Hover: func(event t.HoverEvent) {
                    if event.Type == t.HoverEnter {
                        a.hoveredButton.Set("help-btn")
                    } else {
                        a.hoveredButton.Set("")
                    }
                },
            },

            t.Floating{
                Visible: a.hoveredButton.Get() == "help-btn",
                Config: t.FloatConfig{
                    AnchorID: "help-btn",
                    Anchor:   t.AnchorRightCenter,
                    Offset:   t.Offset{X: 1, Y: 0},
                },
                Child: t.Text{
                    Content: "Click for help",
                    Style: t.Style{
                        BackgroundColor: theme.Surface,
                        Padding:         t.EdgeInsetsHV(1, 0),
                    },
                },
            },
        },
    }
}
```

## Notifications

Display notifications at screen edges:

```go
type App struct {
    notification t.Signal[string]
}

func (a *App) showNotification(msg string) {
    a.notification.Set(msg)
    // Auto-dismiss after delay (would need a timer)
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()
    msg := a.notification.Get()

    return t.Column{
        Children: []t.Widget{
            // Main content...

            t.Floating{
                Visible: msg != "",
                Config: t.FloatConfig{
                    Position:  t.FloatPositionBottomCenter,
                    Offset:    t.Offset{X: 0, Y: -2},
                    OnDismiss: func() { a.notification.Set("") },
                },
                Child: t.Text{
                    Content: msg,
                    Style: t.Style{
                        BackgroundColor: theme.Success,
                        ForegroundColor: theme.Success.AutoText(),
                        Padding:         t.EdgeInsetsHV(2, 0),
                    },
                },
            },
        },
    }
}
```

## Complete Example

An app with a dropdown menu and a modal dialog:

```go
package main

import (
    t "terma"
)

type App struct {
    showDropdown t.Signal[bool]
    showModal    t.Signal[bool]
    selected     t.Signal[string]
    menuState    *t.ListState[string]
}

func NewApp() *App {
    items := []string{"View", "Edit", "Delete", "Share"}
    return &App{
        showDropdown: t.NewSignal(false),
        showModal:    t.NewSignal(false),
        selected:     t.NewSignal(""),
        menuState:    t.NewListState(items),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Height: t.Flex(1),
        Style:  t.Style{Padding: t.EdgeInsetsAll(2)},
        Children: []t.Widget{
            // Header with dropdown
            t.Row{
                Children: []t.Widget{
                    t.Text{Content: "My App"},
                    t.Spacer{},
                    t.Button{
                        ID:      "menu-btn",
                        Label:   "Actions",
                        OnPress: func() { a.showDropdown.Set(true) },
                    },
                },
            },

            t.Spacer{Height: t.Cells(2)},

            // Selected action display
            t.ShowWhen(a.selected.Get() != "", t.Text{
                Content: "Selected: " + a.selected.Get(),
            }),

            // Dropdown menu
            t.Floating{
                Visible: a.showDropdown.Get(),
                Config: t.FloatConfig{
                    AnchorID:  "menu-btn",
                    Anchor:    t.AnchorBottomRight,
                    OnDismiss: func() { a.showDropdown.Set(false) },
                },
                Child: t.Column{
                    Width: t.Cells(15),
                    Style: t.Style{
                        BackgroundColor: theme.Surface,
                        Border:          t.BorderRounded,
                        BorderColor:     theme.TextMuted,
                    },
                    Children: []t.Widget{
                        t.List[string]{
                            State: a.menuState,
                            RenderItem: func(item string, idx int, focused bool) t.Widget {
                                style := t.Style{Padding: t.EdgeInsetsHV(1, 0)}
                                if focused {
                                    style.BackgroundColor = theme.Primary
                                    style.ForegroundColor = theme.Primary.AutoText()
                                }
                                return t.Text{Content: item, Style: style}
                            },
                            OnSelect: func(item string) {
                                a.showDropdown.Set(false)
                                if item == "Delete" {
                                    a.showModal.Set(true)
                                } else {
                                    a.selected.Set(item)
                                }
                            },
                        },
                    },
                },
            },

            // Confirmation modal
            t.Floating{
                Visible: a.showModal.Get(),
                Config: t.FloatConfig{
                    Position:  t.FloatPositionCenter,
                    Modal:     true,
                    OnDismiss: func() { a.showModal.Set(false) },
                },
                Child: t.Column{
                    Width: t.Cells(30),
                    Style: t.Style{
                        BackgroundColor: theme.Surface,
                        Padding:         t.EdgeInsetsAll(2),
                        Border:          t.BorderRounded,
                        BorderColor:     theme.Primary,
                    },
                    Children: []t.Widget{
                        t.Text{
                            Content: "Confirm Delete",
                            Style:   t.Style{ForegroundColor: theme.Primary},
                        },
                        t.Spacer{Height: t.Cells(1)},
                        t.Text{Content: "Are you sure you want to delete?"},
                        t.Spacer{Height: t.Cells(1)},
                        t.Row{
                            Spacing:    2,
                            MainAlign:  t.MainAxisEnd,
                            Children: []t.Widget{
                                t.Button{
                                    ID:    "cancel",
                                    Label: "Cancel",
                                    OnPress: func() {
                                        a.showModal.Set(false)
                                    },
                                },
                                t.Button{
                                    ID:    "delete",
                                    Label: "Delete",
                                    OnPress: func() {
                                        a.selected.Set("Deleted!")
                                        a.showModal.Set(false)
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "m", Name: "Menu", Action: func() {
            a.showDropdown.Set(!a.showDropdown.Get())
        }},
    }
}

func main() {
    t.Run(NewApp())
}
```

## Behavior Notes

- Floats are rendered after the main widget tree, ensuring they appear on top
- Screen bounds clamping keeps floats visible even near edges
- Multiple floats can be visible simultaneously (later ones render on top)
- Modal floats capture clicks to prevent interaction with widgets behind them
- Escape key dismisses floats by default when `OnDismiss` is set
