# Tree

A focusable, navigable tree widget for hierarchical data with expand/collapse, filtering, and optional lazy loading.

## Overview

The tree is built from `TreeNode` values:

```go
type TreeNode[T any] struct {
    Data     T
    Children []TreeNode[T] // nil = not loaded (lazy), [] = leaf
}
```

- `Children == nil` means the node has not loaded children yet (lazy).
- `Children == []` means the node is a leaf.

## TreeState

`TreeState` holds all tree state and is required by `Tree`.

```go
type TreeState[T any] struct {
    Nodes      AnySignal[[]TreeNode[T]]
    CursorPath AnySignal[[]int]
    Collapsed  AnySignal[map[string]bool]
    Selection  AnySignal[map[string]struct{}]
}
```

Create state with `NewTreeState(roots)`. The cursor starts at the first root node (if any).

Common methods:

- Navigation: `CursorUp`, `CursorDown`, `CursorToParent`, `CursorToFirstChild`
- Expand/collapse: `Toggle`, `Expand`, `Collapse`, `ExpandAll`, `CollapseAll`, `IsCollapsed`
- Lazy load: `SetChildren`
- Selection: `ToggleSelection`, `Select`, `Deselect`, `ClearSelection`, `IsSelected`, `SelectedPaths`
- Queries: `NodeAtPath`, `CursorNode`

## Tree Widget

```go
Tree[T]{
    ID:          "my-tree",
    State:       state,
    RenderNode:  func(node T, ctx TreeNodeContext) Widget { ... },
}
```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `State` | `*TreeState[T]` | — | Required state |
| `NodeID` | `func(T) string` | path-based | Stable ID for nodes when tree structure changes |
| `RenderNode` | `func(T, TreeNodeContext) Widget` | default | Node renderer |
| `RenderNodeWithMatch` | `func(T, TreeNodeContext, MatchResult) Widget` | `nil` | Optional renderer with match info |
| `HasChildren` | `func(T) bool` | `nil` | For lazy nodes with `Children == nil` |
| `OnExpand` | `func(T, []int, func([]TreeNode[T]))` | `nil` | Lazy load callback |
| `Filter` | `*FilterState` | `nil` | Filter state (query + options) |
| `MatchNode` | `func(T, string, FilterOptions) MatchResult` | `MatchString(fmt)` | Custom matcher |
| `OnSelect` | `func(T)` | `nil` | Invoked on Enter |
| `OnCursorChange` | `func(T)` | `nil` | Invoked when cursor moves |
| `ScrollState` | `*ScrollState` | `nil` | Share with Scrollable for scroll-into-view |
| `Width` | `Dimension` | auto | Width preference |
| `Height` | `Dimension` | auto | Height preference |
| `Style` | `Style` | — | Container styling |
| `MultiSelect` | `bool` | `false` | Enable multi-select |
| `CursorPrefix` | `string` | `""` | Optional cursor prefix (from `CursorStyle`) |
| `SelectedPrefix` | `string` | `""` | Optional selection prefix (from `CursorStyle`) |
| `Indent` | `int` | `2` | Indentation per depth level |
| `ShowGuideLines` | `*bool` | `true` | Display guide lines connecting tree levels |
| `GuideStyle` | `Style` | `theme.TextMuted` | Style for guide lines |
| `ExpandIndicator` | `string` | `"\u25BC"` | Indicator for expanded nodes |
| `CollapseIndicator` | `string` | `"\u25B6"` | Indicator for collapsed nodes |
| `LeafIndicator` | `string` | `" "` | Indicator for leaf nodes |

To disable guide lines:

```go
showGuides := false
tree := t.Tree[string]{
    State:          state,
    ShowGuideLines: &showGuides,
}
```

To customize guide line styling:

```go
tree := t.Tree[string]{
    State:      state,
    GuideStyle: t.Style{ForegroundColor: t.Hex("#6b7280")},
}
```

## Basic Usage

```go
state := t.NewTreeState([]t.TreeNode[string]{
    {Data: "src", Children: []t.TreeNode[string]{
        {Data: "main.go", Children: []t.TreeNode[string]{}},
    }},
    {Data: "README.md", Children: []t.TreeNode[string]{}},
})

tree := t.Tree[string]{
    ID:    "file-tree",
    State: state,
    RenderNode: func(name string, ctx t.TreeNodeContext) t.Widget {
        return t.Text{Content: name, Width: t.Flex(1)}
    },
}
```

## Filtering

Filtering uses `FilterState` and a match callback. When a query is active:

- A node is visible if it matches or has a matching descendant.
- Ancestors are rendered with a dimmed style (`FilteredAncestor`).
- Matches can be highlighted with `RenderNodeWithMatch`.
- The tree auto-expands to show matches.

```go
filter := t.NewFilterState()
filter.Query.Set("main")

tree := t.Tree[FileInfo]{
    State:  state,
    Filter: filter,
    MatchNode: func(info FileInfo, q string, opts t.FilterOptions) t.MatchResult {
        return t.MatchString(info.Name, q, opts)
    },
    RenderNodeWithMatch: func(info FileInfo, ctx t.TreeNodeContext, match t.MatchResult) t.Widget {
        // Use match.Ranges to highlight
        return t.Text{Content: info.Name}
    },
}
```

## Lazy Loading

For nodes with `Children == nil`, use `HasChildren` and `OnExpand`:

```go
tree := t.Tree[FileInfo]{
    State: state,
    HasChildren: func(info FileInfo) bool { return info.IsDir },
    OnExpand: func(info FileInfo, path []int, setChildren func([]t.TreeNode[FileInfo])) {
        go func() {
            children := loadDirectory(info.Path)
            setChildren(children)
        }()
    },
}
```

## Selection

Enable `MultiSelect` and use shift navigation to extend the selection. Selection state is stored in `TreeState.Selection`.

If you want a toggle key, add your own keybind and call:

```go
treeState.ToggleSelection(treeState.CursorPath.Peek())
```

You can render prefixes with `CursorPrefix` and `SelectedPrefix`. By default no prefixes are shown; selection uses theme colors (`theme.Selection` and `theme.SelectionText`).

## Keyboard

| Key | Action |
|-----|--------|
| Up / k | Previous visible node |
| Down / j | Next visible node |
| Left / h | Collapse node, or move to parent |
| Right / l | Expand node, or move to first child |
| Enter | Trigger `OnSelect` |
| Space | Toggle expand/collapse |
| Home / g | Jump to first visible node |
| End / G | Jump to last visible node |
| Shift + Up/Down | Extend selection (multi-select) |
| Shift + Home/End | Extend selection to start/end |

## Scroll Integration

Wrap the tree in a `Scrollable` and share a `ScrollState`:

```go
scroll := t.NewScrollState()
tree := t.Tree[FileInfo]{State: state, ScrollState: scroll}

t.Scrollable{
    State:  scroll,
    Height: t.Flex(1),
    Child:  tree,
}
```

The tree will keep the cursor in view and respond to mouse wheel scrolling through `ScrollState`.

## Example App

See `cmd/tree-example` for a complete demo with filtering, lazy loading, and multi-select reporting.
