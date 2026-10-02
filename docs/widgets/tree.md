# Tree

`Tree` displays hierarchical nodes with navigation and expansion.

![An expanded Project node containing main.go and README.md, followed by Archive.](../assets/widgets/tree.svg)

## [Example](index.md#run-an-example)

```go
--8<-- "docs/widget-examples/collections/examples.go:tree"
```

## Behavior

- `NewTreeState` stores root nodes, the cursor path, collapsed nodes, and selection.
- Each `TreeNode` holds `Data` and optional `Children`.
- Loaded nodes with children start expanded unless their state marks them collapsed.
- Up and Down move between visible nodes.
- Left collapses a node or moves to its parent, while Right expands a node or moves to its first child.
- Space toggles expansion.
- Enter and a double-click on a node's label invoke `OnSelect(node T, selected []T)` with the cursor node and the selected nodes.
- `ActivateOnClick` enables single-click activation on a node label. A click on its expansion indicator toggles expansion instead.
- `MultiSelect` enables range selection with Shift navigation, Shift-click, and dragging.
- `NodeID` supplies stable identifiers for collapsed and selected nodes. Without it, identifiers are based on node paths.
- `CursorPath` and `TreeNodeContext.Path` contain source child indices, starting at the root, even when a filter hides other nodes.
- `ClearSelection` clears selected nodes. `ClearAnchor` resets the starting point for the next Shift selection without clearing the current selection.
- `RenderNode` receives the data and a `TreeNodeContext` containing depth, path, expansion, and selection information.
- `Filter` keeps matching loaded nodes and their ancestors, including matches under collapsed parents. It does not load children to search them.
- `RenderNodeWithMatch` also receives the match result.
- For lazy loading, a nil `Children` slice uses `HasChildren` to determine whether the node can expand.
- `OnExpand` receives a callback that installs the loaded children.
- `ShowGuideLines`, `Indent`, and the indicator fields control the hierarchy markers. Guide lines are enabled by default, and the default indent is two cells.
- A shared `ScrollState` connects the tree to a surrounding `Scrollable` for cursor visibility.

## Related

- [DirectoryTree](directorytree.md) adds filesystem defaults to `Tree`.
