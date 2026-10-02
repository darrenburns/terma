# Drag and drop

`Draggable[T]` lifts a widget under the mouse pointer. `DropTarget[T]` receives its payload when the left button is released over an accepting target.

```go
Draggable[string]{
	ID:      "card-write-docs",
	Payload: "write-docs",
	Child:   Text{Content: "Write docs"},
}

DropTarget[string]{
	ID:    "done-column",
	Child: Column{Width: Cells(24), Height: Cells(12)},
	Accept: func(key string) bool {
		return !isDone(key)
	},
	OnDrop: func(key string) {
		markDone(key)
	},
}
```

## Sources and targets

- A source's `ID` must be stable and unique in the widget tree.
- The source and target use the same type parameter. A target with a different payload type is skipped.
- `Accept` is optional. When present, it must have no side effects. The deepest eligible target receives the drop.
- `OnDrop` runs once on release and owns application changes, including moving or removing the source.
- A rejected drop or a release outside a target returns the widget to its original place.
- Dragging starts after the pointer moves at least one terminal cell horizontally or vertically while the left button is held.
- The initial grab position stays under the pointer. The widget keeps its measured size and reserves that space in its parent.
- The floating widget keeps its state and keyboard focus. Its visual does not block targets beneath it.
- Escape, a missing mouse release, source removal or disabling, a new blocking modal, or a terminal resize cancels the drag.
- Terminal cells determine visual placement. Terminals with pixel mouse reporting also provide fractional cell positions for drag thresholds and grab offsets.

Ordinary clicks still activate on press. Focusable children, such as text inputs and buttons, keep their own mouse gestures unless selected as an explicit handle.

## Handles

`HandleID` restricts dragging to a descendant and its children. This lets controls elsewhere in the source keep their normal behavior.

```go
Draggable[string]{
	ID:       "card",
	Payload:  "write-docs",
	HandleID: "card-title",
	Child: Column{Children: []Widget{
		Text{ID: "card-title", Content: "Write docs"},
		Button{ID: "card-open", Label: "Open", OnPress: openCard},
	}},
}
```

## Appearance

`Draggable.Shadow` uses the same `FloatShadow` type as [Floating](floating.md#shadows-and-glows). Nil selects the default drag shadow. An explicitly transparent color disables it.

```go
Shadow: &FloatShadow{
	Color:      RGB(60, 160, 255).WithAlpha(0.35),
	BlurRadius: 2,
},
```

Run `go run ./cmd/drag-drop-demo` to move cards between columns. [Tabs](widgets/tabs.md) use the same floating drag behavior to reorder their headers.
