# Tabs

`TabBar` displays horizontal tabs, and `TabView` adds the active tab's content below the bar.

![Home and Settings tabs with the Home content visible.](../assets/widgets/tabs.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/collections/examples.go:tabs"
```

## Behavior

- `TabState` stores the tabs and the active key.
- `NewTabState` activates the first tab when the list is nonempty.
- Each `Tab` has a unique `Key`, a visible `Label`, and optional `Content`.
- `TabBar` ignores `Content`, while `TabView` uses it for the selected view.
- Left and Right, or `h` and `l`, select adjacent tabs and wrap at the ends.
- `KeybindPattern` supports plain, Alt, or Ctrl number keys for the first nine tabs. The default, `TabKeybindNone`, disables number shortcuts.
- Clicking a tab activates it.
- `Closable` adds close buttons and Ctrl+W.
- If `OnTabClose` is set, it handles closure instead of the bar calling `RemoveTab`.
- `AllowReorder` enables dragging labels and Ctrl+H or Ctrl+L to move the active tab.
- `AddTab`, `InsertTab`, `MoveTab`, `RemoveTab`, and `SetLabel` update tabs through `TabState`.
- `TabStyle` and `ActiveTabStyle` customize the inactive and active tabs.
- `TabView.ContentStyle` styles the content area.

## Drag reordering

The dragged header floats under the pointer and leaves an empty slot. A neighbor moves into the slot only after the dragged header's leading edge passes strictly beyond that neighbor's midpoint. Exactly halfway does not move the neighbor.

Reordering uses horizontal position only. Moving above or below the bar does not change the rule, and releasing there commits the position from the final mouse X coordinate. The active tab and its content state stay associated with the tab's key.

The visible order is a preview until release. `TabState.Tabs()` keeps the committed order during the gesture. Escape or a lost release cancels the preview. Resizing, changing the tab list or labels, or disabling reordering also cancels it. Close buttons keep their click behavior.

`TabBar.DragShadow` and `TabView.DragShadow` customize the floating header with a [FloatShadow](../floating.md#shadows-and-glows). Nil selects the default drag shadow.

`MoveTab(key, index)` moves a tab to a zero-based position without changing the active key. It returns false for an unknown key, an invalid index, or an unchanged position.

## Related

- [Switcher](switcher.md) displays content selected by a string key.
