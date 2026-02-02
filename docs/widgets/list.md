# List

A generic navigable list widget with keyboard navigation, single or multi-select, filtering, and customizable item rendering.

```go
List[string]{
    State:    listState,
    OnSelect: func(item string) { /* handle selection */ },
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional unique identifier |
| `DisableFocus` | `bool` | `false` | Prevent keyboard focus |
| `State` | `*ListState[T]` | — | **Required** - holds items and cursor position |
| `RenderItem` | `func(item T, index int, active, selected bool) Widget` | — | Custom item renderer |
| `RenderItemWithMatch` | `func(item T, index int, active, selected bool, match MatchResult) Widget` | — | Item renderer with filter match data |
| `Filter` | `*FilterState` | `nil` | Optional filter state for matching items |
| `MatchItem` | `func(item T, query string, opts FilterOptions) MatchResult` | — | Custom matcher per item |
| `OnSelect` | `func(item T)` | — | Callback when Enter pressed |
| `OnCursorChange` | `func(item T)` | — | Callback when cursor moves |
| `ScrollState` | `*ScrollState` | `nil` | For scroll-into-view behavior |
| `ItemSpacing` | `int` | `0` | Space between items |
| `MultiSelect` | `bool` | `false` | Enable multi-select |
| `Width` | `Dimension` | `Auto` | Container width |
| `Height` | `Dimension` | `Auto` | Container height |
| `Style` | `Style` | — | Padding, margin, border |

## ListState Methods

### Item Operations

| Method | Description |
|--------|-------------|
| `NewListState(items []T)` | Create state with initial items |
| `SetItems(items []T)` | Replace all items |
| `GetItems() []T` | Get current items |
| `ItemCount() int` | Number of items |
| `Append(item T)` | Add item at end |
| `Prepend(item T)` | Add item at beginning |
| `InsertAt(index int, item T)` | Insert item at index |
| `RemoveAt(index int) bool` | Remove item at index |
| `RemoveWhere(predicate func(T) bool) int` | Remove matching items |
| `Clear()` | Remove all items |

### Cursor Control

| Method | Description |
|--------|-------------|
| `SelectNext()` | Move cursor down |
| `SelectPrevious()` | Move cursor up |
| `SelectFirst()` | Move to first item |
| `SelectLast()` | Move to last item |
| `SelectIndex(index int)` | Move to specific item |
| `SelectedItem() (T, bool)` | Get item at cursor |

### Multi-Select

| Method | Description |
|--------|-------------|
| `ToggleSelection(index int)` | Toggle item selection |
| `Select(index int)` | Add item to selection |
| `Deselect(index int)` | Remove item from selection |
| `IsSelected(index int) bool` | Check if item selected |
| `ClearSelection()` | Clear all selections |
| `SelectAll()` | Select all items |
| `SelectedItems() []T` | Get selected items |
| `SelectedIndices() []int` | Get selected indices |

### Filtering

| Method | Description |
|--------|-------------|
| `ApplyFilter(filter, matchItem) int` | Eagerly apply filter and return count |
| `FilteredCount() int` | Get number of items after filtering |

## Keyboard Navigation

| Keys | Action |
|------|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `Home` / `g` | First item |
| `End` / `G` | Last item |
| `PageUp` / `Ctrl+U` | Page up |
| `PageDown` / `Ctrl+D` | Page down |
| `Enter` | Trigger OnSelect |
| `Space` | Toggle selection (MultiSelect) |
| `Shift+↑/↓` | Extend selection (MultiSelect) |

## Basic Usage

### Simple List

```go
listState := NewListState([]string{"Apple", "Banana", "Cherry"})

List[string]{
    State: listState,
}
```

### With Selection Handler

```go
List[string]{
    State: listState,
    OnSelect: func(item string) {
        fmt.Println("Selected:", item)
    },
}
```

## Filtering

Add real-time filtering by connecting a `FilterState`:

```go
filterState := NewFilterState()

Column{
    Children: []Widget{
        TextInput{
            ID:    "search",
            State: searchInputState,
            OnChange: func(text string) {
                filterState.SetQuery(text)
            },
        },
        List[string]{
            State:  listState,
            Filter: filterState,
        },
    },
}
```

### Eager Filtering with ApplyFilter

In most cases, set `Filter` on the List and let it handle filtering during `Build()`. However, if you need to know the filtered count *before* the List builds (e.g., to decide whether to show a "no results" message or to conditionally render the list), use `ApplyFilter`:

```go
// Apply filter eagerly to get count
listState.ApplyFilter(filterState, matchFunc)
count := listState.FilteredCount()

// Decide what to do based on count
if count == 0 {
    return Text{Content: "No matching items"}
}

// Build list - it will reuse cached results, not re-filter
return List[T]{State: listState, Filter: filterState, ...}
```

`ApplyFilter` caches its results, so when the List builds, it reuses the cached filter results rather than filtering twice. This is used internally by the `Autocomplete` widget to determine whether to show the popup before building the suggestion list.

## Custom Item Rendering

For struct-based items or custom styling, provide a `RenderItem` function:

```go
type Task struct {
    Title    string
    Complete bool
}

List[Task]{
    State: taskListState,
    RenderItem: func(task Task, index int, active, selected bool) Widget {
        icon := "○"
        if task.Complete {
            icon = "●"
        }
        return Text{Content: icon + " " + task.Title}
    },
}
```

### Rendering with Match Highlights

When filtering, use `RenderItemWithMatch` to access match ranges for highlighting:

```go
List[string]{
    State:  listState,
    Filter: filterState,
    RenderItemWithMatch: func(item string, index int, active, selected bool, match MatchResult) Widget {
        // match.Ranges contains the matched character positions
        spans := HighlightSpans(item, match, MatchHighlightStyle(theme))
        return Text{Spans: spans}
    },
}
```

## Multi-Select

Enable item selection with Space and Shift+arrow keys:

```go
List[T]{
    State:       listState,
    MultiSelect: true,
    OnSelect: func(item T) {
        // Access all selected items
        selected := listState.SelectedItems()
        fmt.Printf("Selected %d items\n", len(selected))
    },
}
```

## With Scrolling

Combine with `Scrollable` for long lists:

```go
scrollState := NewScrollState()

Scrollable{
    State:  scrollState,
    Height: Flex(1),
    Child: List[T]{
        State:       listState,
        ScrollState: scrollState,  // Enables scroll-into-view
    },
}
```

## Notes

- `State` is required
- Default rendering converts items to strings via `fmt.Sprint`
- For struct items, provide `RenderItem` to control display
- Use `ScrollState` with `Scrollable` to enable automatic scroll-into-view
- Selection state persists in `ListState` across rebuilds
