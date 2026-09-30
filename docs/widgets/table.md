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
| `OnSelect` | `func(row T)` | — | Callback when Enter is pressed or a row is double-clicked |
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
| `Tab` / `Shift+Tab` | Leave the table for the next/previous widget |
| `Shift+↑/↓` | Extend selection (MultiSelect) |

## Mouse

| Action | Effect |
|--------|--------|
| Click | Focus the table and move the cursor to the cell (row, in row mode) |
| Double-click | Trigger OnSelect |
| Shift+click | Extend selection to the cell, row or column (MultiSelect) |
| Drag | Move the cursor with the pointer; with MultiSelect, select from the pressed cell to the pointer: a box of cells, a run of rows or a run of columns, by selection mode. Dragging past the top or bottom scrolls |

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


## Sorting, identity, resizing, and frozen panes: specification

The table owns view ordering and widths; callers supply typed comparators and stable identities. Source rows are never sorted to implement view sorting. State changes happen in setup or handlers, never Build.

```go
state := NewTableStateWithRowID(people, func(p Person) string { return p.ID })
Table[Person]{
    State: state,
    Columns: []TableColumn{
        {ID: "name", Header: Text{Content: "Name"}, Width: Cells(20), Resizable: true, MinWidth: 6, MaxWidth: 40},
        {ID: "age", Header: Text{Content: "Age"}, Width: Cells(8)},
    },
    Comparators: map[string]func(Person, Person) int{
        "name": func(a, b Person) int { return strings.Compare(a.Name, b.Name) },
        "age": func(a, b Person) int { return cmp.Compare(a.Age, b.Age) },
    },
    FrozenHeader: true,
    FrozenColumns: 1,
    Style: Style{Width: Flex(1), Height: Flex(1)},
}
```

- `state.Sort` holds `TableSort{ColumnID, Direction}`; directions are `TableSortNone`, `TableSortAscending`, `TableSortDescending`. Header click or Ctrl+S cycles ascending → descending → unsorted. Indicators are ↑ / ↓ / ↕. Comparators must be pure and deterministic and return negative/zero/positive; callers define nil, case, locale, NaN policy. Explicit sort overrides fuzzy ranking; ties use source order.
- Unique nonempty column IDs enable sorting/resizing. Missing/duplicate IDs disable these controls without hiding data or panicking. User widths persist by ID. Minimum width is at least one; maximum zero means unlimited, and a maximum below minimum becomes minimum. Drag the last cell of a resizable header, or Ctrl+Left/Right. Ctrl+R resets the width. Resizing preserves selection and clamps scrolling to current content.
- `NewTableStateWithRowID` makes `SetRows` preserve cursor, row/cell selection and anchor for unique nonempty IDs present in both versions. Deleted, empty or duplicate IDs drop selection/anchor; their cursor falls back to the nearest valid source index. `NewTableState` keeps positional behavior. Direct writes to `Rows` bypass reconciliation: use `SetRows`. Public indices and rendering callbacks remain source indices.
- Filtering hides selections without clearing them. A hidden cursor is retained until navigation or activation normalizes it into the view. Zero matches disables activation. Clearing the filter restores the cursor if navigation has not occurred. Shift-selection follows displayed order.
- `FrozenHeader` or positive `FrozenColumns` enables an internal bounded viewport; set width/height and do not wrap it in `Scrollable`. Pass `ScrollState` to observe/control it, or use the state's default. Nonfrozen tables keep external scrolling. Wheel scrolls vertically; Alt+Left/Right scroll horizontally. Navigation reveals the current cell. Frozen column counts clamp to column count. If frozen panes consume all available space, other panes clip to zero. Empty data retains headers.
- Cells clip to their pane for both painting and hit testing. Cursor/selection for default cells remains paint-only; sorting/filtering/rows rebuild structure; offset changes relayout. Width overrides rebuild the cells because their clipping and hit regions change; this keeps retained rendering equivalent to a full render.

### Edge cases and expected verification

| Condition | Expected result | Planned evidence |
|---|---|---|
| Sort ties / direction / clear | Stable source-order ties, source unchanged | Unit tests and SVGs |
| Replacement/reorder | Same unique cursor, selected rows/cells, anchor | State tests |
| Duplicate/empty row IDs | Drop ambiguous identity, deterministic cursor fallback | State tests |
| Duplicate/empty column IDs | Controls disabled safely | Unit tests |
| Empty/filter misses | Header stays, Enter selects nothing | SVG + unit |
| Hidden cursor | Build writes no signals; clearing filter restores identity | Unit tests |
| Resize beyond limits | Clamp; resize never sorts or selects rows | Mouse test + browser |
| Vertical/horizontal scroll | Frozen panes stable; correct mouse mapping | SVG sequence + browser |
| Tiny viewport | Clip safely with no negative dimensions | SVGs |
| Sorted Shift-selection | Range follows view order | Unit test |
| Reactive updates | Incremental frame equals forced full render | Sequence tests |

### Independent probe follow-up

Rows in a frozen viewport are measured at their natural height before pane clipping. Even when one multiline row exceeds the viewport height, vertical scrolling can reach every line. When a row or column is too large to fit its scrolling pane, automatic cursor reveal shows its leading edge; manual scrolling still reaches the remaining content. Table exposes one focus stop, so Tab reaches sibling inputs and other table instances, including inside Dialog.

Run `go run ./cmd/terma-browser -- go run ./cmd/table-features-demo -probe` for a tall-row regression demo. Navigate to Notes with Right, use `b`/`t` for bottom/top, and `m` to test the same viewport inside a Dialog.
