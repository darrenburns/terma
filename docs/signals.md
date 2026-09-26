# Signals

Signals provide reactive state management in Terma.

Terma now tracks where a signal was read:

- Read during `Build()` -> the widget is marked for rebuild
- Read during layout or measurement -> the widget is marked for layout
- Read during `Render()` -> the widget is marked for paint only

This means not every signal change refreshes the whole app. Widgets that keep fast-changing state in `Render()` can repaint only the damaged region.

A build-phase change rebuilds only the subscribed widget and the widgets it returns. Parents and siblings keep their previous build output. Layout and paint still run for the whole tree after any build change.

Because clean widgets are not rebuilt, `Build()` must read changing state through signals (or state objects built on signals). A plain struct field that changes without a signal write will not be picked up, even if some other part of the app rebuilds in the same frame. `ctx.IsFocused()`, `ctx.Focused()`, `ctx.IsHovered()`, `ctx.ActiveKeybinds()` and `ctx.Theme()` are all reactive.

A widget returned directly from `Build()` is used for layout and painting but its own `Build()` is not called. Place composite widgets such as `Dialog` or `Floating` inside a container (for example as a child of a `Column`) so they are built.

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
