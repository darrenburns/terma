# Custom Widgets

Terma now tracks signal reads by phase. As a widget author, the main job is to read each piece of state in the lowest phase that actually needs it.

## Stable Leaf Widgets

For most leaf widgets, `Build()` should stay structurally stable and return the widget itself.

```go
type StatusBadge struct {
	Label Signal[string]
}

func (b StatusBadge) Build(ctx terma.BuildContext) terma.Widget {
	return b
}

func (b StatusBadge) Render(ctx *terma.RenderContext) {
	label := b.Label.Get()
	ctx.WriteText(0, 0, label)
}
```

This keeps signal changes in the paint phase instead of forcing a rebuild.

Do not mutate signals inside `Build()`. Update state in handlers, effects, or setup code.

## Signal Phases

Signal reads map to invalidation like this:

- Read in `Render()` -> paint-only update
- Read in layout or measurement code -> full layout pass
- Read in `Build()` -> full rebuild

Use these rules when deciding where state belongs:

- Read fast-changing visual state in `Render()`: spinner frame, cursor position, selection, progress value, hover state, temporary highlighting.
- Read size-affecting state in layout or measurement code: wrapped line count, content width, content height, child spacing, scroll offsets that change child positions.
- Read structural state in `Build()`: whether a child exists, which child subtree is active, or which widget type should be returned.

If a signal only changes pixels inside the widget's existing bounds, read it in `Render()`.

## Auto-Sized Widgets

If your widget can be auto-sized and its content size may change without changing structure, implement `IntrinsicContentSizer`.

```go
type IntrinsicContentSizer interface {
	ContentWidthHint() int
	ContentHeightHint(width int) int
}
```

Terma uses these hints for widgets with `Auto` dimensions to decide whether a signal change can stay paint-only or must promote to layout.

Use this for widgets such as:

- text inputs whose content width can grow or shrink
- text areas whose wrapped height depends on content
- animated widgets whose frames have different widths

If the intrinsic width or height changes, Terma reruns layout because surrounding widgets may be affected.

## Custom Containers

For container widgets, prefer `ContainerLayoutBuilder`.

```go
type ContainerLayoutBuilder interface {
	BuildContainerLayoutNode(ctx BuildContext, children []layout.LayoutNode) layout.LayoutNode
}
```

The framework builds child layout nodes first, then passes them to the container. Your container only needs to describe how those children are arranged.

This avoids rebuilding child subtrees while assembling layout and is the preferred API for custom containers.

`BuildLayoutNode(ctx)` still works as a compatibility fallback, especially for older widgets and simple leaf-style layout nodes.

## Checklist

Before shipping a custom widget, check the following:

- `Build()` stays stable unless the widget's structure actually changes.
- Fast-changing visual state is read in `Render()`.
- Size-affecting state is read during layout or measurement.
- Auto-sized content widgets implement `IntrinsicContentSizer`.
- Custom containers implement `BuildContainerLayoutNode(...)`.
