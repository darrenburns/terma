# CommandPalette

A filterable command palette with fuzzy search, nested levels, and keyboard-driven selection. It opens as a modal float over the app.

```go
palette := NewCommandPaletteState("Commands", []CommandPaletteItem{
    {Label: "New File", Hint: "Ctrl+N", Action: newFile},
    {Divider: "View"},
    {Label: "Toggle Word Wrap", FilterText: "Toggle Word Wrap soft wrap", Action: toggleWrap},
    {Label: "Theme", Children: themeItems},
})

CommandPalette{
    ID:             "palette",
    State:          palette,
    OnCursorChange: previewItem,
}
```

Call `palette.Open()` to show it. Actions don't close the palette on their own, so call `palette.Close()` in them when they should.

## Positioning

By default, the palette sits at the top center of the screen with a two-row inset.
Use `Position` and `Offset` for screen positioning, or `AnchorID` to attach it to
another widget's current layout bounds:

```go
Text{ID: "tab-bar", Content: "Request one | Request two"}

CommandPalette{
    ID:       "tab-picker",
    State:    palette,
    AnchorID: "tab-bar",
    // Anchor defaults to AnchorBottomLeft: directly below, aligned left.
    Style:    Style{Width: Cells(40)},
}
```

The anchor widget needs a stable ID. The palette follows its position and size on
the first frame and on later layout changes and terminal resizes; the application
doesn't need to measure or cache coordinates. `AnchorID` takes precedence over
`Position`. Set `Anchor` to any [floating anchor point](../floating.md#anchor-based-positioning),
such as `AnchorBottomRight` to align with the tab bar's right edge.

For anchored palettes, `Offset` is used exactly as given, with no screen-top
inset. For example, `Offset{Y: 1}` leaves one row below a bottom anchor. Placement
uses the usual floating-widget screen clamping. If the anchor ID cannot be found,
the existing floating-widget fallback places it at `Offset` coordinates, clamped
to the screen. Palette focus, filtering, selection, and dismissal behavior are
the same with either positioning mode.

Run `go run ./cmd/command-palette-anchor-example` to try moving the anchor and
switching between left and right alignment while the palette is open.

## Searching

Typing filters the current level with fuzzy matching: the query's characters must appear in order, but not next to each other, so `tw` finds "Toggle Word Wrap". Whitespace separates terms, which can match in any order (`wrap toggle` also finds it).

Results are ranked best match first:

- Matches at the start of words score highest, so `file` lists "New File" before "Profile Settings". Initials count too: `ps` finds "Profile Settings" first.
- Consecutive characters score higher than scattered ones.
- Between matches of equal quality, shorter labels come first (`copy` lists "Copy" before "Copy Path"), then the original order.
- `FilterText` adds hidden search keywords. An item that matches only through its keywords ranks below a comparable match on a visible label, and has nothing highlighted.

While a query is entered, results are one ranked list and dividers are hidden. They return when the query is cleared.

For substring matching in the original item order instead, set `level.FilterState.Mode` to `FilterContains`.

## The cursor

- Every change to the query moves the cursor to the top result.
- With an empty query the cursor rests on the level's `Current` item if it has one, otherwise on the first selectable item. Mark the item that represents the current value (the active theme, say) with `Current: true` so a level opens on it rather than on its first item. This matters when `OnCursorChange` previews the item under the cursor.
- Up/Down (or Ctrl+P/Ctrl+N) move between selectable items, skipping dividers and disabled items. Home/End jump to the first and last.
- Clicking an item moves the cursor to it, and double-clicking chooses it, as Enter does. Focus stays in the search input throughout.

## Nested levels

An item with `Children` shows a `▸` and opens a nested level on Enter. `Children` is called at that point, so it can build items from current state. The breadcrumb shows the path, using `ChildrenTitle` (or the label).

Each level has its own query, cursor and scroll position. A nested level always opens with an empty query. Going back returns to the parent as it was left, with its query and the cursor on the item that opened the level. To go back, use Escape, Backspace in an empty query (unless `DisableBackspaceToPop` is set), or click a breadcrumb.

## Closing

Escape at the root level or a click outside the palette dismisses it and calls `OnDismiss`. `Close()` hides it. Either way the next `Open()` starts fresh: back at the root level with an empty query and the cursor reset. Use `Close(true)` to keep the level stack, queries and cursors for the next open.

Focus returns to the widget that had it before the palette opened, or to the ID given with `SetNextFocusIDOnClose`.

=== "Code"

    ```go
    --8<-- "cmd/command-palette-example/main.go"
    ```
