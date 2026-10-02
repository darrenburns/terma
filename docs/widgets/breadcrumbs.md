# Breadcrumbs

`Breadcrumbs` displays a path as text segments separated by a configurable separator.

![Projects, Terma, and Docs breadcrumbs above a selection message.](../assets/widgets/breadcrumbs.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/collections/examples.go:breadcrumbs"
```

## Behavior

- `Path` supplies the segment labels in display order.
- `Separator` defaults to `>` with one space on either side.
- `OnSelect` makes each segment clickable and receives its zero-based index, including the final segment.
- The callback determines what navigation occurs.
- Without `OnSelect`, the segments are plain text.
- An empty path renders an `EmptyWidget`.
- The default colors distinguish clickable ancestors, the final segment, and separators.
- `Style.ForegroundColor` overrides these default colors.
- When `ID` is set, clickable segments receive IDs formed as `<ID>-segment-<index>`.
