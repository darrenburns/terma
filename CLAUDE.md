# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terma is a declarative terminal UI (TUI) framework for Go. It provides a reactive widget system with automatic dependency tracking, similar to React or Flutter but for the terminal.

This project is not currently in use by any developers, so maintaining backwards compatibility is not a requirement. Improvements should always be preferred over backwards compatibility.

## Build Commands

```bash
# Run an example.
# IMPORTANT: You (Claude) CANNOT run the examples in this way.
# Where you want to run an example, you should instead provide the user with the command you wish
# to run, and instructions you want them to follow. The user can supply any required log lines to you.
go run ./cmd/example/main.go
go run ./cmd/simple-list-example/main.go

# Build an example
# IMPORTANT: You (Claude) don't need to do this unless it seems essential.
# Instead of building to check if the build works, just run the tests.
go build ./cmd/example

# Fetch/tidy dependencies
go mod tidy
```

## Checking your changes / feedback loop

You cannot run examples, but you can run snapshot tests. This means you can add debug logging,
write a snapshot test which will exercise the logic and hit the logs, and then you can read the log file
yourself.

## Snapshot Testing

**Visual features require snapshot tests.** Any change that affects widget appearance, layout, or rendering must include snapshot tests to verify correctness.

### Running Snapshot Tests

```bash
# Run all tests (includes snapshot tests)
go test ./...

# Update golden files after intentional visual changes
UPDATE_SNAPSHOTS=1 go test ./...
```

### Writing Snapshot Tests

```go
func TestMyWidget_Layout(t *testing.T) {
    widget := Column{
        Spacing: 1,
        Children: []Widget{
            Text{Content: "First"},
            Text{Content: "Second"},
        },
    }
    AssertSnapshot(t, widget, 20, 5, "Two text items stacked vertically with 1-cell gap")
}

// Use AssertSnapshotNamed for multiple snapshots in one test
func TestMyWidget_States(t *testing.T) {
    AssertSnapshotNamed(t, "default", widget, 20, 5, "Default state")
    AssertSnapshotNamed(t, "focused", focusedWidget, 20, 5, "With focus ring")
}
```

### Key Functions

| Function | Purpose |
|----------|---------|
| `AssertSnapshot(t, widget, width, height, description...)` | Basic snapshot assertion |
| `AssertSnapshotNamed(t, name, widget, width, height, description...)` | Named snapshot (multiple per test) |
| `RenderToBuffer(widget, width, height)` | Render widget to buffer for inspection |

Golden files are stored in `testdata/<TestName>.svg`. The test framework generates an HTML gallery at `testdata/snapshot_gallery.html` for visual review.

## Architecture

### Core Concepts

**Widgets**: Everything is a widget. Apps implement `Widget` with a `Build(ctx BuildContext) Widget` method that returns composed widgets.

**Signals**: Reactive state via `Signal[T]`. Call `Get()` during `Build()` to auto-subscribe; call `Set()` to trigger rebuilds. Use `AnySignal[T]` for non-comparable types.

**Layout**: Two-pass constraint-based system. Dimensions can be `Auto`, `Cells(n)` (fixed), or `Flex(n)` (flexible/proportional).

### Key Files

| File | Purpose |
|------|---------|
| `app.go` | Main event loop, `Run()` entry point |
| `signal.go` | Reactive `Signal[T]` and `AnySignal[T]` |
| `widget.go` | Core `Widget`, `Layoutable`, `Renderable` interfaces |
| `layout.go` | `Column`, `Row` layout widgets |
| `stack.go` | `Stack` widget for z-order overlays |
| `context.go` | `BuildContext` for focus/hover state |
| `focus.go` | Focus management, `Focusable`, `KeyHandler` interfaces |
| `list.go` | Generic `List[T]` with keyboard navigation |
| `scroll.go` | `Scrollable` widget and `ScrollController` |
| `style.go` | Styling: colors, padding, margins |
| `keybind.go` | Declarative keybinding system |
| `conditional.go` | Visibility wrappers: `ShowWhen`, `HideWhen`, etc. |
| `switcher.go` | `Switcher` widget for content switching |

### Widget Pattern

```go
type App struct {
    counter Signal[int]
}

func (a *App) Build(ctx BuildContext) Widget {
    return Column{
        Children: []Widget{
            Text{Content: fmt.Sprintf("Count: %d", a.counter.Get())},
            Button{Label: "Increment", OnPress: func() { a.counter.Update(func(c int) int { return c + 1 }) }},
        },
    }
}

func main() {
    Run(&App{counter: NewSignal(0)})
}
```

### Key Interfaces

- `Widget`: Has `Build(ctx BuildContext) Widget`
- `Focusable`: Receives keyboard events when focused
- `KeybindProvider`: Returns declarative `[]Keybind` for queryable key mappings
- `Clickable`/`Hoverable`: Mouse interaction handlers

### Dimensions

- `Auto` - fit content
- `Cells(n)` - fixed n terminal cells
- `Flex(n)` - flexible (proportional to siblings)

## Available Widgets

### Layout Widgets

| Widget | Purpose | Key Fields |
|--------|---------|------------|
| `Column` | Arranges children vertically | `Children`, `Spacing`, `MainAlign`, `CrossAlign` |
| `Row` | Arranges children horizontally | `Children`, `Spacing`, `MainAlign`, `CrossAlign` |
| `Stack` | Overlays children in z-order | `Children`, `Alignment` |
| `Dock` | Edge-docking layout (like WPF DockPanel) | `Top`, `Bottom`, `Left`, `Right`, `Body`, `DockOrder` |
| `Scrollable` | Scrolling container with scrollbar | `Child`, `State` (required), `DisableScroll` |
| `Floating` | Overlay/modal positioning | `Visible`, `Config`, `Child` |
| `Switcher` | Shows one keyed child at a time | `Active`, `Children` |

### Content Widgets

| Widget | Purpose | Key Fields |
|--------|---------|------------|
| `Text` | Display text (plain or rich with Spans) | `Content`, `Spans`, `Wrap` |
| `Button` | Focusable button with press handler | `ID` (required), `Label`, `OnPress` |
| `List[T]` | Generic navigable list | `State` (required), `OnSelect`, `RenderItem`, `MultiSelect` |

### Utility Widgets

| Widget | Purpose | Key Fields |
|--------|---------|------------|
| `KeybindBar` | Displays active keybinds from focused widget | `Style`, `FormatKey` |
| `Spacer` | Flexible empty space for layout control | `Width`, `Height` (default Flex(1)) |

### Spacing: Prefer `Spacing` Field Over `Spacer` Widget

For uniform gaps between children, use the `Spacing` field on `Column`/`Row` instead of inserting `Spacer` widgets:

```go
// ✓ Preferred: Use Spacing field for uniform gaps
Column{
    Spacing: 1,  // 1 cell gap between each child
    Children: []Widget{item1, item2, item3},
}

// ✗ Avoid: Manual Spacer widgets for uniform gaps
Column{
    Children: []Widget{
        item1,
        Spacer{Height: Cells(1)},
        item2,
        Spacer{Height: Cells(1)},
        item3,
    },
}
```

Use `Spacer` only for **flexible/proportional space** that pushes widgets apart:

```go
// ✓ Correct use of Spacer: push widgets to edges
Row{
    Children: []Widget{
        leftItem,
        Spacer{},      // Flex(1) - expands to fill available space
        rightItem,     // Pushed to right edge
    },
}
```

### Widget Examples

```go
// Column with spacing and alignment
Column{
    Spacing:    1,
    MainAlign:  MainAxisCenter,
    CrossAlign: CrossAxisStart,
    Children:   []Widget{...},
}

// Dock layout (edges consume space, body fills remainder)
Dock{
    Top:    []Widget{Header{}},
    Bottom: []Widget{KeybindBar{}},
    Body:   MainContent{},
}

// Scrollable list with state
scrollState := NewScrollState()
listState := NewListState(items)
Scrollable{
    State:  scrollState,
    Height: Flex(1),
    Child: List[string]{
        State:       listState,
        ScrollState: scrollState,
        OnSelect:    func(item string) { ... },
    },
}

// Floating modal dialog
Floating{
    Visible: showDialog.Get(),
    Config: FloatConfig{
        Position:  FloatPositionCenter,
        Modal:     true,
        OnDismiss: func() { showDialog.Set(false) },
    },
    Child: Dialog{...},
}

// KeybindBar at bottom of app
KeybindBar{
    Style: Style{BackgroundColor: theme.Surface},
}

// Stack with overlapping children (first child at bottom, last on top)
Stack{
    Children: []Widget{
        Card{},  // Base layer - determines Stack size with Auto dimensions
        Positioned{
            Top:   IntPtr(-1),  // Overflow above Stack bounds
            Right: IntPtr(-1),  // Overflow right of Stack bounds
            Child: Badge{Count: 3},
        },
    },
}
```

### Stack Widget

`Stack` overlays children on top of each other in z-order (first child at bottom, last on top). Children can be:

- **Regular widgets**: Positioned using the Stack's `Alignment` within the **content area** (inside padding/borders)
- **`Positioned` wrappers**: Positioned using edge offsets (`Top`, `Right`, `Bottom`, `Left`) relative to the **border-box** (can overlap padding/borders)

**Coordinate systems:** When a Stack has padding or borders, non-positioned children are laid out within the inner content area, while `Positioned` children use the Stack's outer border-box as their reference. This means `Positioned{Top: IntPtr(0), Left: IntPtr(0)}` places a child at the Stack's top-left corner, potentially overlapping any border or padding.

```go
// Positioned helper functions
Positioned{Top: IntPtr(0), Left: IntPtr(0), Child: widget}  // Top-left corner
PositionedAt(5, 10, widget)                                  // At row 5, col 10
PositionedFill(widget)                                       // Fill entire Stack
```

**Important caveat**: Stack sizes itself based on the largest **non-positioned** child only. `Positioned` children do not affect Stack's size and can overflow its bounds. If you only have `Positioned` children, the Stack will have zero size with Auto dimensions.

```go
// ✓ Correct: Non-positioned child defines size, positioned child overlays
Stack{
    Children: []Widget{
        Card{Width: Cells(20), Height: Cells(10)},  // Defines Stack size
        Positioned{Top: IntPtr(-1), Right: IntPtr(-1), Child: Badge{}},
    },
}

// ✗ Problem: Only positioned children = zero size with Auto
Stack{
    Children: []Widget{
        Positioned{Top: IntPtr(0), Left: IntPtr(0), Child: Card{}},
    },
}
// Fix: Use explicit dimensions
Stack{
    Width:  Cells(20),
    Height: Cells(10),
    Children: []Widget{
        Positioned{Top: IntPtr(0), Left: IntPtr(0), Child: Card{}},
    },
}
```

### Rich Text with Markup

Use `ParseMarkupToText` for styled text (preferred) or `ParseMarkup` when you need the raw `[]Span`:

```go
// ✓ Preferred: Returns a ready-to-use Text widget
ParseMarkupToText("Press [b $Accent]Enter[/] to continue", ctx.Theme())

// Alternative: Returns []Span for custom use
Text{Spans: ParseMarkup("Press [b $Accent]Enter[/] to continue", ctx.Theme())}

// Markup syntax: [style $ThemeColor on $Background]text[/]
// Styles: bold/b, italic/i, underline/u
// Theme colors: $Primary, $Secondary, $Accent, $Text, $TextMuted, $TextOnPrimary,
//               $Surface, $SurfaceHover, $Background, $Border, $FocusRing,
//               $Error, $Warning, $Success, $Info
// Hex colors: #rrggbb
```

## Widget Conventions

### Keyboard Handling: Prefer Keybinds()

Use declarative `Keybinds()` instead of imperative `OnKey()` for keyboard handling:

```go
// ✓ Preferred: Declarative keybindings (auto-displayed in KeybindBar)
func (m *MyWidget) Keybinds() []Keybind {
    return []Keybind{
        {Key: "enter", Name: "Confirm", Action: m.confirm},
        {Key: "escape", Name: "Cancel", Action: m.cancel},
        {Key: "d", Name: "Delete", Action: m.delete, Hidden: true}, // Hidden from KeybindBar
    }
}

// ✗ Avoid: Imperative key handling (not discoverable, not shown in KeybindBar)
func (m *MyWidget) OnKey(event KeyEvent) bool {
    if event.MatchString("enter") { ... }
    return false
}
```

`OnKey()` should only be used for:
- Complex key sequences not expressible as simple bindings
- Keys that need access to the KeyEvent details (modifiers, raw key data)
- Fallback handling after Keybinds() has been checked

### Values-First Pattern

Pass values to widgets, not Signals. The App reads from Signals and passes values to widgets:

```go
// ✓ Correct: Parent reads signal, passes value
Text{Content: a.message.Get()}
Button{Label: "Submit", Disabled: !a.isValid.Get()}  // (future)

// Exception: State objects for complex interactive state
List[string]{State: a.listState}
Scrollable{State: a.scrollState}
```

State objects (`ListState`, `ScrollState`) are used when the widget needs to manage complex internal state like cursor position, selection, or scroll offset.

### Use Theme Variables in Demo Apps

When creating demo apps or examples, always use theme variables from `ctx.Theme()` instead of hardcoding colors:

```go
// ✓ Correct: Use theme variables
Style{
    BackgroundColor: ctx.Theme().Surface,
    ForegroundColor: ctx.Theme().Text,
    BorderColor:     ctx.Theme().Primary,
}

// ✗ Avoid: Hardcoded colors
Style{
    BackgroundColor: color.RGB(30, 30, 30),
    ForegroundColor: color.RGB(255, 255, 255),
}
```

Available theme colors: `Primary`, `Secondary`, `Accent`, `Text`, `TextMuted`, `TextOnPrimary`, `Surface`, `SurfaceHover`, `Background`, `Border`, `FocusRing`, `Error`, `Warning`, `Success`, `Info`.

### Standard Widget Field Order

All widgets should follow this consistent field ordering:

```go
type WidgetName struct {
    ID       string      // 1. Identity (optional)
    // ... widget-specific fields ...  // 2. Core fields (State, Child, Content, etc.)
    Width    Dimension   // 3. Dimensions (optional)
    Height   Dimension
    Style    Style       // 4. Styling (optional)
    Click    func()      // 5. Callbacks (optional)
    Hover    func(bool)
}
```

### Required Interfaces by Field

| Field | Interface | Methods |
|-------|-----------|---------|
| ID | `Identifiable` | `WidgetID() string` |
| Width/Height | `Dimensioned` | `GetDimensions() (Dimension, Dimension)` |
| Style | `Styled` | `GetStyle() Style` |
| Click | `Clickable` | `OnClick()` |
| Hover | `Hoverable` | `OnHover(bool)` |

### Conditional Rendering

Use wrapper functions for visibility control:

| Function | Behavior |
|----------|----------|
| `ShowWhen(cond, child)` | Shows child when true, gone when false (no space) |
| `HideWhen(cond, child)` | Gone when true (no space), shows child when false |
| `VisibleWhen(cond, child)` | Always reserves space, renders only when true |
| `InvisibleWhen(cond, child)` | Always reserves space, renders only when false |

```go
// Conditional presence (like CSS display: none)
ShowWhen(user.IsAdmin(), AdminPanel{})
HideWhen(isLoading.Get(), Content{})

// Invisible but reserves space (like CSS visibility: hidden)
VisibleWhen(hasData.Get(), Chart{})  // placeholder space when no data
```

### Content Switching

Use `Switcher` to show one widget at a time based on a string key:

```go
Switcher{
    Active: a.activeTab.Get(),
    Children: map[string]Widget{
        "home":     HomeTab{listState: a.homeListState},
        "settings": SettingsTab{},
    },
}
```

State is preserved across switches via Signals and State objects held by the App.

## Examples

Working examples in `cmd/*/main.go`. Start with `cmd/simple-list-example/TUTORIAL.md` for a comprehensive walkthrough.

## Debugging

### Logging

Initialize logging with `InitLogger()`, then use `Log(format, args...)`. Logs write to `terma.log`.

### Debug Widget

Terma provides a debug overlay for development that shows real-time performance metrics.

**Enabling Debug Mode:**

```go
func main() {
    terma.InitDebug()  // Enable debug overlay
    terma.Run(&MyApp{})
}
```

**Using the Debug Overlay:**

- **Toggle**: Press `ctrl+backtick` to show/hide the debug overlay
- **Metrics displayed**:
  - Build time (milliseconds per render)
  - Widget count (total widgets in current render)
  - Focused widget (ID and type)
  - Frame rate (FPS)

The debug overlay is non-modal and doesn't interfere with app interaction. It appears as a floating panel at the top-center of the screen when toggled on.
