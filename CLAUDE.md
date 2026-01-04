# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Terma is a declarative terminal UI (TUI) framework for Go. It provides a reactive widget system with automatic dependency tracking, similar to React or Flutter but for the terminal.

This project is not currently in use by any developers, so maintaining backwards compatibility is not a requirement. Improvements should always be preferred over backwards compatibility.

## Build Commands

```bash
# Run an example
go run ./cmd/example/main.go
go run ./cmd/simple-list-example/main.go

# Build an example (cleanup the artifact afterwards)
go build ./cmd/example

# Fetch/tidy dependencies
go mod tidy
```

## Architecture

### Core Concepts

**Widgets**: Everything is a widget. Apps implement `Widget` with a `Build(ctx BuildContext) Widget` method that returns composed widgets.

**Signals**: Reactive state via `Signal[T]`. Call `Get()` during `Build()` to auto-subscribe; call `Set()` to trigger rebuilds. Use `AnySignal[T]` for non-comparable types.

**Layout**: Two-pass constraint-based system. Dimensions can be `Auto`, `Cells(n)` (fixed), or `Fr(n)` (fractional/proportional).

### Key Files

| File | Purpose |
|------|---------|
| `app.go` | Main event loop, `Run()` entry point |
| `signal.go` | Reactive `Signal[T]` and `AnySignal[T]` |
| `widget.go` | Core `Widget`, `Layoutable`, `Renderable` interfaces |
| `layout.go` | `Column`, `Row` layout widgets |
| `context.go` | `BuildContext` for focus/hover state |
| `focus.go` | Focus management, `Focusable`, `KeyHandler` interfaces |
| `list.go` | Generic `List[T]` with keyboard navigation |
| `scroll.go` | `Scrollable` widget and `ScrollController` |
| `style.go` | Styling: colors, padding, margins |
| `keybind.go` | Declarative keybinding system |

### Widget Pattern

```go
type App struct {
    counter *Signal[int]
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
- `Fr(n)` - fractional (proportional to siblings)

## Widget Conventions

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

### Future: Wrapper Functions for Common Properties

When visibility/disabled/opacity are needed, use wrapper functions (not yet implemented):

```go
// Future API
ShowWhen(isLoggedIn, AdminPanel{})
HideWhen(isLoading, Content{})
Disable(Button{Label: "Submit"})
```

## Examples

Working examples in `cmd/*/main.go`. Start with `cmd/simple-list-example/TUTORIAL.md` for a comprehensive walkthrough.

## Debugging

Initialize logging with `InitLogger()`, then use `Log(format, args...)`. Logs write to `terma.log`.
