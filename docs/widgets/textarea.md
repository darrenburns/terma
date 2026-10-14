# TextArea

`TextArea` edits multiple lines with cursor movement, selection, and configurable wrapping.

![A text area with release notes and a prompt to add a summary.](../assets/widgets/textarea.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/inputs/examples.go:textarea"
```

## Behavior

- Create state once with `NewTextAreaState` and reuse it across builds.
- `NewTextAreaState` stores the initial text, places the cursor at its end, and enables soft wrapping.
- `Style.Width` and `Style.Height` set the content dimensions.
- `State.GetText()` reads the current text without subscribing to changes.
- `State.SetText()` replaces the text and clamps the cursor without calling `OnChange`.
- Enter inserts a newline while editing is enabled.
- Ctrl+Enter calls `OnSubmit` when that callback is set.
- `OnChange` receives the text after edits handled by the widget.

The cursor is steady by default. `terma.SetCursorBlink(true)` enables the same
[classic on/off blink](textinput.md#cursor-animation) as `TextInput`. Select
`terma.SetCursorBlinkStyle(terma.CursorBlinkPulse)` separately for a gentle pulse.
Run `go run ./cmd/text-area-demo` to try it in the editor and its search input.

## Navigation and selection

- Arrow keys move the cursor, and Shift with an arrow key extends the selection.
- Home and End move to the boundaries of the current newline-delimited line, while Ctrl+Left and Ctrl+Right move by word.
- Ctrl+A selects all text.
- A double click selects a word, and a triple click selects a line.
- `State.ReadOnly` blocks user edits while movement and selection remain available.
- With `RequireInsertMode: true`, Escape leaves insert mode, and I or Enter re-enters it.
- New state starts in insert mode. Set `State.InsertMode.Set(false)` during setup to start in navigation mode.

## Wrapping and search

- `State.WrapMode` supports `WrapNone`, `WrapSoft`, and `WrapHard`.
- `State.ToggleWrap()` changes `WrapNone` to `WrapSoft`, and either wrapping mode to `WrapNone`.
- `State.SearchQuery` highlights text matches, with case sensitivity controlled by `State.SearchCaseSensitive`.
- `ExtraKeybinds` precede default bindings, and `OnPaste` can consume a paste by returning `true`.

## Transforming typed input

`BeforeInput` receives the proposed `TextAreaEdit` before typed text or Enter is
inserted. Return it unchanged for normal typing, or change its replacement range,
text, and final caret to implement pairing, overtyping, or indentation:

```go
TextArea{
	State: state,
	BeforeInput: func(edit TextAreaEdit) TextAreaEdit {
		if edit.Text == "(" && edit.Start == edit.End {
			edit.Text = "()"
			// The proposed cursor already points just after the opening "(".
		}
		return edit
	},
}
```

- `Start` and `End` describe the range to replace in the current text; a selection
  is supplied as that range. Read surrounding text from `state.Content.Peek()`.
- All positions are **grapheme indexes**, not bytes or runes. `Cursor` and
  `SelectionAnchor` refer to the resulting document. The supplied anchor is `-1`
  (no selection). Positions are clamped to document bounds.
- A transformed edit is one undo step, including its final cursor and selection.
  `OnChange` fires once after the edit. Returning the proposed edit unchanged
  preserves normal typing groups.
- To overtype a closer, return an empty replacement with `Start == End` and
  advance `Cursor`. To reject input, also set `Cursor` back to the original
  position. Text-preserving edits do not notify `OnChange`, add an undo step, or
  discard redo history.
- The hook runs only while editing is allowed. Paste, deletion, undo/redo, and
  programmatic state edits bypass it. Use `OnPaste` for paste-specific behavior.
  `ExtraKeybinds` still take precedence, including custom Enter bindings.
- The callback must not mutate state; Terma applies the returned edit and keeps
  the cursor visible.

## Line numbers

Set `ShowLineNumbers: true` to add a left-hand gutter:

```go
TextArea{
	State:           state,
	ShowLineNumbers: true,
	Style:           Style{Width: Flex(1), Height: Flex(1)},
}
```

- Numbers start at 1 and count newline-delimited lines, including empty lines.
  Wrapped continuation rows have a blank gutter.
- The gutter grows to fit the document's line count and reserves space for text
  wrapping, cursor movement, and mouse selection. It is hidden when the widget
  is too narrow to fit the gutter plus text and cursor space.
- With `state.WrapMode.Set(WrapNone)`, moving the cursor beyond the visible
  columns scrolls the text horizontally. The gutter stays fixed at the left edge;
  text, selections, and highlights do not draw over it.
- `LineNumberStyle` optionally overrides the gutter's theme-derived muted text
  and background using a `SpanStyle`.
