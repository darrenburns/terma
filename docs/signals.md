# Signals

Signals provide reactive state management in Terma.

Terma now tracks where a signal was read:

- Read during `Build()` -> the widget is marked for rebuild
- Read during layout or measurement -> the widget is marked for layout
- Read during `Render()` -> the widget is marked for paint only

This means not every signal change refreshes the whole app. Widgets that keep fast-changing state in `Render()` can repaint only the damaged region.

## Choosing The Right Phase

As a rule:

- Use `Build()` for structural decisions such as which children exist.
- Use layout or measurement code for state that affects size or child placement.
- Use `Render()` for state that only changes what is painted inside the widget's current bounds.

Examples:

- A spinner frame should usually be read in `Render()`.
- A text input's cursor position should usually be read in `Render()`.
- A text area's wrapped line count belongs in layout or intrinsic sizing.
- A `Switcher` choosing a different child belongs in `Build()`.

## Auto-Sized Widgets

If a widget can use `Auto` width or height and its content size can change without changing structure, implement `IntrinsicContentSizer`:

```go
type IntrinsicContentSizer interface {
	ContentWidthHint() int
	ContentHeightHint(width int) int
}
```

Terma compares these hints across updates. If the intrinsic size changes, the widget is promoted from paint-only to layout so the rest of the tree can reflow correctly.

See [Custom Widgets](widgets/custom-widgets.md) for authoring guidance.
