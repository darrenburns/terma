# TextInput

A single-line focusable text entry widget with cursor navigation, editing commands, and automatic horizontal scrolling.

=== "Demo"

    <video autoplay loop muted playsinline src="../../assets/textinput-demo.mp4"></video>

=== "Code"

    ```go
    --8<-- "cmd/text-input-example/main.go"
    ```

## Overview

`TextInput` provides a text entry field with built-in keyboard handling for cursor movement, text deletion, and submission. State is managed externally via `TextInputState`, which holds the text content and cursor position.

```go
--8<-- "docs/minimal-examples/textinput-basic/main.go"
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | — | Required for focus management |
| `DisableFocus` | `bool` | `false` | Prevent keyboard focus |
| `State` | `*TextInputState` | — | Required - holds text and cursor position |
| `Placeholder` | `string` | `""` | Text shown when empty and unfocused |
| `Width` | `Dimension` | `Auto` | Optional width |
| `Height` | `Dimension` | — | Ignored - always single-line; use `Style.Padding` for visual spacing |
| `Style` | `Style` | — | Padding, margin, border, colors |
| `OnChange` | `func(string)` | — | Callback when text changes |
| `OnSubmit` | `func(string)` | — | Callback when Enter is pressed |
| `ExtraKeybinds` | `[]Keybind` | — | Additional keybinds (checked before defaults) |

## TextInputState

`TextInputState` holds the source of truth for text content and cursor position. Create it with `NewTextInputState`:

```go
// Empty input
state := t.NewTextInputState("")

// Pre-filled input (cursor at end)
state := t.NewTextInputState("initial value")
```

### Methods

| Method | Description |
|--------|-------------|
| `GetText()` | Returns the current text as a string |
| `SetText(string)` | Replaces the content and clamps cursor |
| `Insert(string)` | Inserts text at cursor position |
| `Clear()` | Clears all content |

## Keyboard Shortcuts

TextInput includes built-in keyboard handling:

### Cursor Movement

| Key | Action |
|-----|--------|
| `Left` | Move cursor left |
| `Right` | Move cursor right |
| `Home` / `Ctrl+A` | Move to beginning |
| `End` / `Ctrl+E` | Move to end |
| `Ctrl+Left` / `Alt+B` | Move to previous word |
| `Ctrl+Right` / `Alt+F` | Move to next word |

### Text Editing

| Key | Action |
|-----|--------|
| `Backspace` | Delete character before cursor |
| `Delete` / `Ctrl+D` | Delete character after cursor |
| `Ctrl+U` | Delete to beginning of line |
| `Ctrl+K` | Delete to end of line |
| `Ctrl+W` / `Alt+Backspace` | Delete word backward |

### Submission

| Key | Action |
|-----|--------|
| `Enter` | Submit (triggers `OnSubmit` callback) |

## Basic Usage

```go
// Simple text input
t.TextInput{
    ID:          "search",
    State:       a.searchState,
    Placeholder: "Search...",
}

// With change and submit handlers
t.TextInput{
    ID:          "email",
    State:       a.emailState,
    Placeholder: "Enter email",
    OnChange:    func(text string) { a.validateEmail(text) },
    OnSubmit:    func(text string) { a.submitForm() },
}

// Fixed width with styling
t.TextInput{
    ID:    "code",
    State: a.codeState,
    Width: t.Cells(20),
    Style: t.Style{
        BorderStyle:     t.BorderRounded,
        Padding:         t.EdgeInsetsXY(1, 0),
        BackgroundColor: theme.Surface,
    },
}
```

## Callbacks

Use `OnChange` to react to text changes in real-time, and `OnSubmit` to handle Enter key presses:

```go
--8<-- "docs/minimal-examples/textinput-callbacks/main.go"
```

## Custom Keybinds

Add custom keybinds that are checked before the defaults using `ExtraKeybinds`:

```go
t.TextInput{
    ID:    "input",
    State: a.inputState,
    ExtraKeybinds: []t.Keybind{
        {Key: "ctrl+enter", Name: "Submit Alt", Action: func() {
            a.submitAlternate(a.inputState.GetText())
        }},
        {Key: "escape", Name: "Clear", Action: func() {
            a.inputState.Clear()
        }},
    },
}
```

## Styling

Apply visual styling through the `Style` field:

```go
--8<-- "docs/minimal-examples/textinput-styling/main.go"
```

## Notes

- Content height is always 1 cell (single-line input)
- Use `Style.Padding` to add visual space around the text
- The widget automatically scrolls horizontally when content exceeds the visible width
- Unicode text is fully supported with proper grapheme cluster handling
- Printable characters are captured and inserted as text; modifier keys bubble to parent widgets
