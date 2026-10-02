# List

`List` displays a navigable sequence of items.

![A list containing Inbox, Today, and Upcoming.](../assets/widgets/list.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/collections/examples.go:list"
```

## Behavior

- `NewListState` stores the items, cursor position, and selection.
- `SetItems`, `Append`, `InsertAt`, and `RemoveAt` update the items.
- `SelectedItem` returns the item at the cursor, while `SelectedItems` returns the multi-selection.
- Up and Down, or `k` and `j`, move the cursor.
- Home and End jump to the first and last visible items.
- Enter and a double-click call `OnSelect`, or `ActivateOnClick` enables single-click activation.
- With `MultiSelect`, Shift navigation, Shift-click, and dragging extend the selection.
- The default renderer formats each item with `fmt.Sprintf("%v", item)`.
- `RenderItem` receives the item and its active and selected flags.
- `Filter` and `MatchItem` control filtering, while `RenderItemWithMatch` also receives the match result.
- `List` computes the filtered view during rendering. `State.ApplyFilter` can compute it earlier, and `FilteredCount` reports the last computed view size.
- `CursorIndex` and selection indices refer to the source items even when filtering changes the visible order.
- `OnSelect` receives one item. Multi-selection is available separately through `State.SelectedItems()`.
- A shared `ScrollState` connects the list to a surrounding `Scrollable` and keeps cursor movements in view.

## Related

- [Table](table.md) displays rows in columns.
