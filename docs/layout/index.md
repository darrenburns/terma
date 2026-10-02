# Layout

Layout widgets control arrangement and spacing.

## Widgets

- [Row and Column](row-column.md)
- [Dock](dock.md)
- [Scrollable](scrollable.md)
- [Spacer](spacer.md)
- [SplitPane](splitpane.md)
- [Stack and Positioned](stack.md)
- [Floating](../floating.md)

## Dimensions

`Style.Width` and `Style.Height` accept `Auto`, `Cells(n)`, `Flex(n)`, or `Percent(n)`.
`Auto` sizes from content, and `Cells(n)` requests a fixed number of cells.
`Flex(n)` assigns a share of the remaining space along a container axis.
`Percent(n)` requests a percentage of the parent's available space, before fixed-size siblings consume their share.
`Style.MinWidth`, `Style.MaxWidth`, `Style.MinHeight`, and `Style.MaxHeight` constrain dimensions.
These dimensions describe the content box. Padding, border, and margin add to the outer space a widget occupies.
Leaving a dimension unset allows widget-specific defaults, such as `Dock` filling available space. Setting `Auto` explicitly requests content sizing.

## Spacing

`Row.Spacing` and `Column.Spacing` set gaps between children.
`Style.Padding` reserves space inside the border.
`Style.Margin` reserves space outside the border.

The [widget examples](../widgets/index.md#run-an-example) include runnable layouts.
