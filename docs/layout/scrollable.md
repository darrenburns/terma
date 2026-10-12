# Scrollable

`Scrollable` displays a child through a viewport with vertical and horizontal scroll offsets.

![The first four log lines inside a bordered scrollable viewport.](../assets/widgets/scrollable.svg)

## [Example](../widgets/index.md#run-an-example)

```go
--8<-- "docs/widget-examples/layout/examples.go:scrollable"
```

## Behavior

- `State` holds the scroll position and is created with `NewScrollState()`.
- Retain the same state across `Build()` calls so scrolling does not reset when another signal changes.
- The example keeps seven log lines inside a four-line content area.
- `Focusable: true` enables keyboard focus unless `DisableFocus` or `DisableScroll` is set.
- Arrow keys scroll one cell, and `h`, `j`, `k`, and `l` provide the corresponding left, down, up, and right controls.
- Page Up and Page Down scroll by the viewport height.
- Home and End move to the top and bottom.
- The vertical scrollbar supports dragging its thumb.
- `DisableScroll` disables scrolling and hides the scrollbar.
- With `State.PinToBottom` enabled, content growth follows the bottom while the state remains pinned.
- Scrolling up breaks that pin, and scrolling back to the bottom restores it.
- `ScrollToView(y, height)` adjusts the offset to reveal a vertical region of the child.

## Fading scrollbars

Set a positive `ScrollbarFadeDelay` to fade the thumb and track after a period
of inactivity. The scrollbar still reserves its column, so content never shifts
or rewraps as it disappears. By default, scrollbars stay visible.

```go
terma.Scrollable{
    State:                 scrollState,
    Child:                 content,
    Style:                 terma.Style{Height: terma.Cells(10)},
    ScrollbarFadeDelay:     time.Second,
    ScrollbarFadeDuration:  250 * time.Millisecond, // Optional; this is the default.
}
```

The scrollbar starts visible. Mouse-wheel and keyboard scrolling reveal it
immediately and restart the delay, including attempts to scroll past an edge.
Programmatic position changes (including `ScrollToView`) also reveal it.
The reserved column remains clickable while hidden, and dragging keeps the
scrollbar visible until release. Hover or keyboard focus alone does not prevent
fading. Both colors fade into the background behind the scrollbar.

Set `ScrollbarFadeDelay` to zero or a negative duration to keep the scrollbar
visible. `ScrollbarFadeDuration` values of zero or less use the 250ms default.
The [scroll example](../widgets/index.md#run-an-example) demonstrates idle fading:
`go run ./cmd/scroll-example`.

## Scrollbars without a track

Set an explicitly transparent track color to show only the thumb:

```go
ScrollbarTrackColor: terma.RGBA(0, 0, 0, 0),
```

Unlike an unset `Color{}` (which uses the theme's track color), alpha zero shows
the background beneath the scrollbar. This works with solid colors and gradients,
including gradients inherited from a parent. The Unicode partial-cell thumb ends
also reveal the correct background for that row. You can combine a transparent
track with idle fading; the thumb fades into the gradient without changing its
size or the reserved space. Intermediate alpha values give a translucent track.

## Related

- [Column](row-column.md)
- [List](../widgets/list.md)
