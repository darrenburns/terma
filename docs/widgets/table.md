# Table

Display tabular data with keyboard navigation, selection, and optional filtering. Use `Table` for data grids, file browsers, or any multi-column navigable list.

```go
Table[[]string]{
    State: tableState,
    Columns: []TableColumn{
        {Width: Cells(12), Header: Text{Content: "Name"}},
        {Width: Cells(10), Header: Text{Content: "Status"}},
    },
    OnSelect: func(row []string) { /* handle selection */ },
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional unique identifier |
| `DisableFocus` | `bool` | `false` | Prevent keyboard focus |
| `State` | `*TableState[T]` | — | **Required** - holds rows and cursor position |
| `Columns` | `[]TableColumn` | — | **Required** - defines column count and widths |
| `RenderCell` | `func(row T, rowIdx, colIdx int, active, selected bool) Widget` | — | Custom cell renderer |
| `RenderCellWithMatch` | `func(..., match MatchResult) Widget` | — | Cell renderer with filter match data |
| `Filter` | `*FilterState` | `nil` | Optional filter state for matching rows |
| `MatchCell` | `func(row T, rowIdx, colIdx int, query string, opts FilterOptions) MatchResult` | — | Custom matcher per cell |
| `RenderHeader` | `func(colIndex int) Widget` | — | Header renderer (overrides column headers) |
| `OnSelect` | `func(row T)` | — | Callback when Enter pressed |
| `OnCursorChange` | `func(row T)` | — | Callback when cursor moves |
| `ScrollState` | `*ScrollState` | `nil` | For scroll-into-view behavior |
| `RowHeight` | `int` | `0` | Uniform row height override |
| `ColumnSpacing` | `int` | `0` | Space between columns |
| `RowSpacing` | `int` | `0` | Space between rows |
| `SelectionMode` | `TableSelectionMode` | `TableSelectionCursor` | Highlight mode |
| `MultiSelect` | `bool` | `false` | Enable multi-select |
| `Width` | `Dimension` | `Auto` | Container width |
| `Height` | `Dimension` | `Auto` | Container height |
| `Style` | `Style` | — | Padding, margin, border |

## TableColumn

| Field | Type | Description |
|-------|------|-------------|
| `Width` | `Dimension` | Column width (`Cells`, `Flex`, `Auto`) |
| `Header` | `Widget` | Header widget for this column |

## TableState Methods

### Row Operations

| Method | Description |
|--------|-------------|
| `NewTableState(rows []T)` | Create state with initial rows |
| `SetRows(rows []T)` | Replace all rows |
| `GetRows() []T` | Get current rows |
| `RowCount() int` | Number of rows |
| `Append(row T)` | Add row at end |
| `Prepend(row T)` | Add row at beginning |
| `InsertAt(index int, row T)` | Insert row at index |
| `RemoveAt(index int) bool` | Remove row at index |
| `RemoveWhere(predicate func(T) bool) int` | Remove matching rows |
| `Clear()` | Remove all rows |

### Cursor Control

| Method | Description |
|--------|-------------|
| `SelectNext()` | Move cursor down |
| `SelectPrevious()` | Move cursor up |
| `SelectFirst()` | Move to first row |
| `SelectLast()` | Move to last row |
| `SelectIndex(index int)` | Move to specific row |
| `SelectColumn(index int)` | Move to specific column |
| `SelectedRow() (T, bool)` | Get row at cursor |

### Multi-Select

| Method | Description |
|--------|-------------|
| `ToggleSelection(index int)` | Toggle row selection |
| `Select(index int)` | Add row to selection |
| `Deselect(index int)` | Remove row from selection |
| `IsSelected(index int) bool` | Check if row selected |
| `ClearSelection()` | Clear all selections |
| `SelectAll()` | Select all rows |
| `SelectedRows() []T` | Get selected rows |
| `SelectedIndices() []int` | Get selected indices |
| `SelectRange(from, to int)` | Select range of rows |

## Selection Modes

Control how the cursor and selection are highlighted:

```go
// Highlight only the cursor cell (default)
Table[T]{SelectionMode: TableSelectionCursor, ...}

// Highlight the entire row
Table[T]{SelectionMode: TableSelectionRow, ...}

// Highlight the entire column
Table[T]{SelectionMode: TableSelectionColumn, ...}
```

## Keyboard Navigation

| Keys | Action |
|------|--------|
| `↑` / `k` | Move up |
| `↓` / `j` | Move down |
| `←` / `h` | Move left (column) |
| `→` / `l` | Move right (column) |
| `Home` / `g` | First row |
| `End` / `G` | Last row |
| `PageUp` / `Ctrl+U` | Page up |
| `PageDown` / `Ctrl+D` | Page down |
| `Enter` | Trigger OnSelect |
| `Space` | Toggle selection (MultiSelect) |
| `Shift+↑/↓` | Extend selection (MultiSelect) |

## Basic Usage

### Simple Table

```go
tableState := NewTableState([][]string{
    {"Alice", "Engineer"},
    {"Bob", "Designer"},
})

Table[[]string]{
    State: tableState,
    Columns: []TableColumn{
        {Width: Cells(15)},
        {Width: Cells(15)},
    },
}
```

### With Headers

```go
Table[[]string]{
    State: tableState,
    Columns: []TableColumn{
        {Width: Cells(15), Header: Text{Content: "Name", Style: Style{Bold: true}}},
        {Width: Cells(15), Header: Text{Content: "Role", Style: Style{Bold: true}}},
    },
}
```

### Flexible Column Widths

```go
Columns: []TableColumn{
    {Width: Flex(1)},  // Takes 1/3 of available space
    {Width: Flex(2)},  // Takes 2/3 of available space
}
```

## Custom Cell Rendering

For struct-based rows or custom styling, provide a `RenderCell` function:

```go
type Person struct {
    Name   string
    Role   string
    Active bool
}

Table[Person]{
    State: personTableState,
    Columns: []TableColumn{
        {Width: Cells(15)},
        {Width: Cells(15)},
        {Width: Cells(8)},
    },
    RenderCell: func(p Person, rowIdx, colIdx int, active, selected bool) Widget {
        var content string
        switch colIdx {
        case 0:
            content = p.Name
        case 1:
            content = p.Role
        case 2:
            if p.Active {
                content = "Active"
            } else {
                content = "Away"
            }
        }
        return Text{Content: content}
    },
}
```

## Multi-Select

Enable row selection with Space and Shift+arrow keys:

```go
Table[T]{
    State:       tableState,
    MultiSelect: true,
    OnSelect: func(row T) {
        // Access all selected rows
        selected := tableState.SelectedRows()
        fmt.Printf("Selected %d rows\n", len(selected))
    },
}
```

## With Scrolling

Combine with `Scrollable` for long tables:

```go
scrollState := NewScrollState()

Scrollable{
    State:  scrollState,
    Height: Flex(1),
    Child: Table[T]{
        State:       tableState,
        ScrollState: scrollState,  // Enables scroll-into-view
        Columns:     columns,
    },
}
```

## Complete Example

Run this example with:

```bash
go run ./cmd/table-example
```

```go
--8<-- "cmd/table-example/main.go"
```

## Notes

- `State` and `Columns` are required fields
- Default rendering works with slice/array row types (e.g., `[]string`)
- For struct rows, provide `RenderCell` to extract column values
- Use `ScrollState` with `Scrollable` to enable automatic scroll-into-view
- Selection state persists in `TableState` across rebuilds
