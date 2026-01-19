# Tree

Display hierarchical data with keyboard navigation, expand/collapse, filtering, and optional multi-selection. Use `Tree` for file browsers, nested menus, organization charts, or any data with parent-child relationships.

```go
Tree[FileInfo]{
    State: treeState,
    RenderNode: func(info FileInfo, ctx TreeNodeContext) Widget {
        return Text{Content: info.Name}
    },
    OnSelect: func(info FileInfo) { /* handle selection */ },
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional unique identifier |
| `State` | `*TreeState[T]` | — | **Required** - holds nodes and cursor position |
| `NodeID` | `func(data T) string` | — | Stable identifier for each node (enables state persistence across data changes) |
| `RenderNode` | `func(node T, ctx TreeNodeContext) Widget` | — | Custom node renderer |
| `RenderNodeWithMatch` | `func(node T, ctx TreeNodeContext, match MatchResult) Widget` | — | Node renderer with filter match data |
| `HasChildren` | `func(node T) bool` | — | Determines expandability for lazy-loaded nodes |
| `OnExpand` | `func(node T, path []int, setChildren func([]TreeNode[T]))` | — | Callback for lazy loading children |
| `Filter` | `*FilterState` | `nil` | Optional filter state for matching nodes |
| `MatchNode` | `func(node T, query string, opts FilterOptions) MatchResult` | — | Custom matcher per node |
| `OnSelect` | `func(node T)` | — | Callback when Enter pressed |
| `OnCursorChange` | `func(node T)` | — | Callback when cursor moves |
| `ScrollState` | `*ScrollState` | `nil` | For scroll-into-view behavior |
| `MultiSelect` | `bool` | `false` | Enable multi-select |
| `Indent` | `int` | `2` | Spaces per nesting level |
| `ExpandIndicator` | `string` | `"▼"` | Icon for expanded nodes |
| `CollapseIndicator` | `string` | `"▶"` | Icon for collapsed nodes |
| `LeafIndicator` | `string` | `" "` | Icon for leaf nodes |
| `Width` | `Dimension` | `Auto` | Container width |
| `Height` | `Dimension` | `Auto` | Container height |
| `Style` | `Style` | — | Padding, margin, border |

## TreeNode

Nodes are represented using the `TreeNode[T]` struct:

```go
type TreeNode[T any] struct {
    Data     T              // Your data for this node
    Children []TreeNode[T]  // Child nodes (nil = not loaded, [] = leaf)
}
```

The `Children` field has special semantics:

- `nil` — Children not yet loaded (enables lazy loading)
- `[]TreeNode[T]{}` — Node is a leaf (no children)
- `[]TreeNode[T]{...}` — Node has children

## TreeNodeContext

When rendering nodes, you receive context about the node's state:

| Field | Type | Description |
|-------|------|-------------|
| `Path` | `[]int` | Path to this node (e.g., `[0, 2, 1]`) |
| `Depth` | `int` | Nesting level (0 = root) |
| `Expanded` | `bool` | Is this node currently expanded? |
| `Expandable` | `bool` | Can this node be expanded? |
| `Active` | `bool` | Is cursor on this node? |
| `Selected` | `bool` | Is this node selected? (MultiSelect mode) |
| `FilteredAncestor` | `bool` | Visible only because a descendant matches the filter |

## TreeState Methods

### Creating State

```go
// Create state with initial root nodes
treeState := NewTreeState([]TreeNode[FileInfo]{
    {Data: FileInfo{Name: "src"}, Children: nil},  // Lazy-loaded
    {Data: FileInfo{Name: "README.md"}, Children: []TreeNode[FileInfo]{}},  // Leaf
})
```

### Cursor Control

| Method | Description |
|--------|-------------|
| `CursorUp()` | Move cursor to previous visible node |
| `CursorDown()` | Move cursor to next visible node |
| `CursorToParent()` | Move cursor to parent node |
| `CursorToFirstChild()` | Move cursor to first child (if visible) |
| `CursorNode() (T, bool)` | Get data at cursor position |
| `NodeAtPath(path []int) (TreeNode[T], bool)` | Get node at specific path |

### Expand/Collapse

| Method | Description |
|--------|-------------|
| `Toggle(path []int)` | Toggle expand/collapse state |
| `Expand(path []int)` | Expand node at path |
| `Collapse(path []int)` | Collapse node at path |
| `ExpandAll()` | Expand all nodes |
| `CollapseAll()` | Collapse all nodes |
| `IsCollapsed(path []int) bool` | Check if node is collapsed |

### Selection (MultiSelect mode)

| Method | Description |
|--------|-------------|
| `ToggleSelection(path []int)` | Toggle node selection |
| `Select(path []int)` | Add node to selection |
| `Deselect(path []int)` | Remove node from selection |
| `IsSelected(path []int) bool` | Check if node is selected |
| `ClearSelection()` | Clear all selections |
| `SelectedPaths() [][]int` | Get all selected node paths |

### Data Manipulation

| Method | Description |
|--------|-------------|
| `SetChildren(path []int, children []TreeNode[T])` | Set children for a node (used with lazy loading) |

## Keyboard Navigation

| Keys | Action |
|------|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Collapse node, or move to parent |
| `→` / `l` | Expand node, or move to first child |
| `Home` / `g` | First node |
| `End` / `G` | Last visible node |
| `Space` | Toggle expand/collapse |
| `Enter` | Trigger OnSelect |
| `Shift+↑/↓` | Extend selection (MultiSelect) |
| `Shift+Home/End` | Extend selection to start/end (MultiSelect) |

## Basic Usage

### Static Tree

```go
type Item struct {
    Name string
}

roots := []TreeNode[Item]{
    {
        Data: Item{Name: "Fruits"},
        Children: []TreeNode[Item]{
            {Data: Item{Name: "Apple"}, Children: []TreeNode[Item]{}},
            {Data: Item{Name: "Banana"}, Children: []TreeNode[Item]{}},
        },
    },
    {
        Data: Item{Name: "Vegetables"},
        Children: []TreeNode[Item]{
            {Data: Item{Name: "Carrot"}, Children: []TreeNode[Item]{}},
        },
    },
}

treeState := NewTreeState(roots)

Tree[Item]{
    State: treeState,
    RenderNode: func(item Item, ctx TreeNodeContext) Widget {
        return Text{Content: item.Name}
    },
}
```

### Custom Node Rendering

```go
Tree[FileInfo]{
    State: treeState,
    RenderNode: func(info FileInfo, ctx TreeNodeContext) Widget {
        theme := ctx.Theme()
        icon := "📄"
        if info.IsDir {
            if ctx.Expanded {
                icon = "📂"
            } else {
                icon = "📁"
            }
        }

        style := Style{ForegroundColor: theme.Text}
        if ctx.Active {
            style.BackgroundColor = theme.Selection
        }

        return Text{
            Content: icon + " " + info.Name,
            Style:   style,
        }
    },
}
```

## Lazy Loading

Load children on-demand when a node is expanded. This is useful for large trees or when children are fetched from an API.

```go
Tree[FileInfo]{
    State: treeState,

    // Determine if a node can be expanded (before children are loaded)
    HasChildren: func(info FileInfo) bool {
        return info.IsDir
    },

    // Called when user expands a node with nil Children
    OnExpand: func(info FileInfo, path []int, setChildren func([]TreeNode[FileInfo])) {
        // Fetch children asynchronously
        go func() {
            children := fetchDirectoryContents(info.Path)
            setChildren(children)  // Update the tree
        }()
    },
}
```

When `Children` is `nil` and `HasChildren` returns `true`, the node shows as expandable. When the user expands it, `OnExpand` is called with a `setChildren` callback to populate the children.

## Filtering

Combine with `FilterState` to filter visible nodes:

```go
type App struct {
    treeState   *TreeState[FileInfo]
    filterState *FilterState
    filterInput *TextInputState
}

func (a *App) Build(ctx BuildContext) Widget {
    return Column{
        Children: []Widget{
            TextInput{
                State:       a.filterInput,
                Placeholder: "Filter...",
                OnChange: func(text string) {
                    a.filterState.Query.Set(text)
                },
            },
            Tree[FileInfo]{
                State:  a.treeState,
                Filter: a.filterState,

                // Optional: custom matching logic
                MatchNode: func(info FileInfo, query string, opts FilterOptions) MatchResult {
                    return MatchString(info.Name, query, opts)
                },
            },
        },
    }
}
```

When filtering:

- Matching nodes are shown with highlighted matches
- Ancestor nodes of matches are shown (with `FilteredAncestor: true` in context)
- Non-matching nodes without matching descendants are hidden

### Rendering with Match Highlights

```go
Tree[FileInfo]{
    State:  treeState,
    Filter: filterState,
    RenderNodeWithMatch: func(info FileInfo, ctx TreeNodeContext, match MatchResult) Widget {
        if match.Matched && len(match.Ranges) > 0 {
            // Highlight matching portions
            highlight := SpanStyle{
                Underline:      UnderlineSingle,
                UnderlineColor: theme.Accent,
            }
            return Text{Spans: HighlightSpans(info.Name, match.Ranges, highlight)}
        }
        return Text{Content: info.Name}
    },
}
```

## With Scrolling

Combine with `Scrollable` for large trees:

```go
scrollState := NewScrollState()

Scrollable{
    State:  scrollState,
    Height: Flex(1),
    Child: Tree[FileInfo]{
        State:       treeState,
        ScrollState: scrollState,  // Enables scroll-into-view
    },
}
```

## Multi-Select

Enable selection of multiple nodes:

```go
Tree[FileInfo]{
    State:       treeState,
    MultiSelect: true,
    OnSelect: func(info FileInfo) {
        // Access all selected paths
        paths := treeState.SelectedPaths()
        fmt.Printf("Selected %d nodes\n", len(paths))
    },
}
```

Use `Shift+Up/Down` to extend selection, or programmatically with `treeState.Select(path)` and `treeState.ToggleSelection(path)`.

## Complete Example

Run this example with:

```bash
go run ./cmd/tree-example
```

```go
--8<-- "cmd/tree-example/main.go"
```

## Notes

- `State` is the only required field—default rendering uses `fmt.Sprintf("%v", node)`
- Paths are represented as `[]int` slices (e.g., `[0, 2, 1]` means first root's third child's second child)
- Use `NodeID` for stable identity when tree data changes (otherwise paths are used)
- Ancestor nodes matching due to descendant matches have `FilteredAncestor: true` and are shown muted by default
- For lazy loading, set `Children: nil` and provide `HasChildren` and `OnExpand`
- Selection state persists in `TreeState` across rebuilds
