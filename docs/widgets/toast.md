# Toasts

`Toasts` shows short notifications stacked in a corner of the screen. A `ToastState` holds them, and any code with the state can add one.

![A success toast and an error toast in the bottom-right corner.](../assets/widgets/toasts.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/display/examples.go:toasts"
```

## Adding toasts

- `NewToastState(ToastOptions{})` creates the queue. Create it once, outside `Build()`.
- `Notify(Toast{...})` adds a toast and returns its `ToastID`.
- `Info`, `Success`, `Warning`, and `Error` add a toast with only a message.
- A `Toast` has a `Message`, an optional bold `Title`, a `Severity`, and a `Timeout`.
- The severity sets the icon and the theme color of the border and title: `Info`, `Success`, `Warning`, or `Error`.
- `Dismiss(id)` removes one toast. `Clear()` removes all of them.
- Every `ToastState` method is safe to call from any goroutine, including a [Task](../async.md) or a ticker. No `Dispatch` is needed.

## Timeouts and the queue

- `ToastOptions.MaxVisible` sets how many toasts are on screen at once. The default is 3.
- Later toasts wait in arrival order, and a muted `+N more` line counts them.
- A toast's timeout starts when it appears on screen, so a waiting toast still gets its full time.
- `ToastOptions.Timeout` sets the default timeout. The default is 4 seconds.
- `Toast.Timeout` overrides the default for one toast. A negative value keeps the toast until it is dismissed.
- Timeouts use one timer per toast on screen. They do not run the animation ticker, and an expiry rebuilds only the toast overlay.

## The Toasts widget

- Place `Toasts{State: state}` anywhere in the tree. It takes no layout space and registers a [Floating](../floating.md) overlay while toasts are on screen.
- `Position` selects the corner or edge: one of the top or bottom `FloatPosition` values. The default is `FloatPositionBottomRight`. The stack sits one cell in from the screen edges.
- `Width` sets each toast's width in cells, including its border. The default is 40. Messages wrap to fit.
- Default cards use `TallBorder` in the toast's severity color.
- Clicking a toast dismisses it.
- While the pointer is over a toast, its timeout is held and its background changes to `SurfaceHover`. The remaining time resumes when the pointer leaves.
- The overlay is not modal and takes no keyboard focus.

## Custom cards

Set `Toasts.RenderCard` to replace the entire card, including its border and padding. Leave it nil to use the default card.

The callback receives the build context, the `Toast`, the resolved width (40 when `Width` is unset), and whether the pointer is over the card. It runs during `Build()` and must be free of side effects. Use the context for theme or reactive reads, and size the returned widget to fit the supplied width.

```go
Toasts{
    State: toasts,
    RenderCard: func(ctx BuildContext, toast Toast, width int, hovered bool) Widget {
        theme := ctx.Theme()
        background := theme.Surface
        if hovered {
            background = theme.SurfaceHover
        }
        return Column{
            Style: Style{
                // Width is content width; subtract two border cells and two padding cells.
                Width:           Cells(max(width-4, 1)),
                Border:          RoundedBorder(theme.Accent),
                Padding:         EdgeInsetsXY(1, 0),
                BackgroundColor: background,
                ForegroundColor: theme.Text,
            },
            Children: []Widget{
                Text{Content: toast.Title, Wrap: WrapSoft, Style: Style{Bold: true}},
                Text{Content: toast.Message, Wrap: WrapSoft},
            },
        }
    },
}
```

Custom cards keep the normal queue, timeouts, click-to-dismiss, and hover-to-pause behavior. Clicks anywhere in the card dismiss it, including clicks on descendants; the callback customizes presentation, not the card's pointer behavior. The `+N more` queue indicator is unchanged.

## Related

- [Floating](../floating.md)
- [Async tasks](../async.md)
