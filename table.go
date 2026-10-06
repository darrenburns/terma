package terma

import (
	"fmt"
	"reflect"
	"sort"

	uv "github.com/charmbracelet/ultraviolet"

	"github.com/darrenburns/terma/layout"
)

// TableState holds the state for a Table widget.
// It is the source of truth for rows and cursor position, and must be provided to Table.
// Rows is a reactive Signal - changes trigger automatic re-renders.
type TableState[T any] struct {
	Rows         AnySignal[[]T] // Reactive source rows; use SetRows for identity reconciliation
	CursorIndex  Signal[int]    // Source row index
	CursorColumn Signal[int]
	Selection    AnySignal[map[int]struct{}] // Row/column/cell keys according to selection mode
	Sort         Signal[TableSort]           // View ordering; source rows remain unchanged
	ColumnWidths AnySignal[map[string]int]   // User width overrides by unique column ID

	rowID                          func(T) string
	viewport                       *ScrollState
	viewportMetrics                tableViewportMetrics
	headerHeight                   int
	resizeColumn                   int
	resizing                       bool
	resizeStartX, resizeStartWidth int

	anchorIndex *int                      // Anchor point for shift-selection (nil = no anchor)
	dragging    bool                      // A press on a cell is held, so pointer motion moves the cursor
	hover       itemHover[tableHoverCell] // The cell, row or column under the pointer

	lastSelectionMode TableSelectionMode
	hasSelectionMode  bool
	columnCount       int // Column count of the Table showing this state, for decoding cell selections

	rowLayouts        []tableRowLayout    // Cached layout metrics (per row)
	columnLayouts     []tableColumnLayout // Cached horizontal extent of each column
	revealed          cursorReveal[int]   // Where the cursor was last scrolled into view
	viewIndices       []int               // View index -> source index for filtered views
	viewIndexBySource map[int]int         // Source index -> view index for filtered views
}

// NewTableState creates a new TableState with the given initial rows.
func NewTableState[T any](initialRows []T) *TableState[T] {
	if initialRows == nil {
		initialRows = []T{}
	}
	return &TableState[T]{
		Sort:         NewSignal(TableSort{}),
		ColumnWidths: NewAnySignal(map[string]int{}),
		viewport:     NewScrollState(),
		Rows:         NewAnySignal(initialRows),
		CursorIndex:  NewSignal(0),
		CursorColumn: NewSignal(0),
		Selection:    NewAnySignal(make(map[int]struct{})),
		hover:        newItemHover[tableHoverCell](),
	}
}

// SetRows replaces rows. Identity-aware states preserve unique records by ID;
// positional states clamp the cursor and drop selections past the new end.
func (s *TableState[T]) SetRows(rows []T) {
	if rows == nil {
		rows = []T{}
	}
	if s.rowID != nil {
		s.replaceIdentifiedRows(rows)
		return
	}
	s.Rows.Set(rows)
	s.clampCursor()
	s.remapSelectedRows(func(i int) (int, bool) { return i, i < len(rows) })
}

// GetRows returns the current table rows (without subscribing to changes).
func (s *TableState[T]) GetRows() []T {
	return s.Rows.Peek()
}

// RowCount returns the number of rows.
func (s *TableState[T]) RowCount() int {
	return len(s.Rows.Peek())
}

// Append adds a row to the end of the table.
func (s *TableState[T]) Append(row T) {
	s.Rows.Update(func(rows []T) []T {
		return append(rows, row)
	})
}

// Prepend adds a row to the beginning of the table.
// The cursor and selection stay on the same rows.
func (s *TableState[T]) Prepend(row T) {
	s.InsertAt(0, row)
}

// InsertAt inserts a row at the specified index.
// If index is out of bounds, it's clamped to valid range.
// The cursor and selection stay on the same rows.
func (s *TableState[T]) InsertAt(index int, row T) {
	hadRows := len(s.Rows.Peek()) > 0
	s.Rows.Update(func(rows []T) []T {
		index = clampInt(index, 0, len(rows))
		rows = append(rows, row)
		copy(rows[index+1:], rows[index:])
		rows[index] = row
		return rows
	})
	// Adjust cursor if insertion was at or before cursor
	cursorIdx := s.CursorIndex.Peek()
	if hadRows && index <= cursorIdx {
		s.CursorIndex.Set(cursorIdx + 1)
	}
	s.remapSelectedRows(func(i int) (int, bool) {
		if i >= index {
			return i + 1, true
		}
		return i, true
	})
}

// RemoveAt removes the row at the specified index.
// Returns true if a row was removed, false if index was out of bounds.
// Selections on the removed row are dropped; other selections stay on their rows.
func (s *TableState[T]) RemoveAt(index int) bool {
	rows := s.Rows.Peek()
	if index < 0 || index >= len(rows) {
		return false
	}
	s.Rows.Update(func(rows []T) []T {
		return append(rows[:index], rows[index+1:]...)
	})
	s.clampCursor()
	s.remapSelectedRows(func(i int) (int, bool) {
		switch {
		case i == index:
			return 0, false
		case i > index:
			return i - 1, true
		default:
			return i, true
		}
	})
	return true
}

// RemoveWhere removes all rows matching the predicate.
// Returns the number of rows removed.
// Selections on removed rows are dropped; other selections stay on their rows.
func (s *TableState[T]) RemoveWhere(predicate func(T) bool) int {
	removed := 0
	var newIndex []int // Old index -> new index, or -1 if removed
	s.Rows.Update(func(rows []T) []T {
		newIndex = make([]int, len(rows))
		result := make([]T, 0, len(rows))
		for i, row := range rows {
			if !predicate(row) {
				newIndex[i] = len(result)
				result = append(result, row)
			} else {
				newIndex[i] = -1
				removed++
			}
		}
		return result
	})
	s.clampCursor()
	if removed > 0 {
		s.remapSelectedRows(func(i int) (int, bool) {
			if i < 0 || i >= len(newIndex) || newIndex[i] < 0 {
				return 0, false
			}
			return newIndex[i], true
		})
	}
	return removed
}

// Clear removes all rows from the table, along with the selection.
func (s *TableState[T]) Clear() {
	s.Rows.Set([]T{})
	s.CursorIndex.Set(0)
	s.CursorColumn.Set(0)
	s.ClearSelection()
	s.ClearAnchor()
}

// remapSelectedRows moves the selection and shift-select anchor after rows are
// inserted or removed, so they stay on the same rows. mapRow returns a row's
// new index, or false if the row is gone. Column selections don't depend on
// rows and are left alone.
func (s *TableState[T]) remapSelectedRows(mapRow func(int) (int, bool)) {
	mapKey := mapRow
	if s.hasSelectionMode {
		switch s.lastSelectionMode {
		case TableSelectionColumn:
			return
		case TableSelectionCursor:
			columnCount := s.columnCount
			if columnCount <= 0 {
				// Cell keys can't be decoded without the column count.
				s.ClearSelection()
				s.ClearAnchor()
				return
			}
			mapKey = func(key int) (int, bool) {
				row, col := cellIndexToRowCol(key, columnCount)
				newRow, ok := mapRow(row)
				return cellIndex(newRow, col, columnCount), ok
			}
		}
	}

	if s.anchorIndex != nil {
		if idx, ok := mapKey(*s.anchorIndex); ok {
			s.anchorIndex = &idx
		} else {
			s.anchorIndex = nil
		}
	}

	sel := s.Selection.Peek()
	changed := false
	next := make(map[int]struct{}, len(sel))
	for key := range sel {
		newKey, ok := mapKey(key)
		if !ok || newKey != key {
			changed = true
		}
		if ok {
			next[newKey] = struct{}{}
		}
	}
	if changed {
		s.Selection.Set(next)
	}
}

// SelectedRow returns the currently selected row (if any).
func (s *TableState[T]) SelectedRow() (T, bool) {
	rows := s.Rows.Peek()
	idx := s.CursorIndex.Peek()
	if idx >= 0 && idx < len(rows) {
		return rows[idx], true
	}
	var zero T
	return zero, false
}

// SelectNext moves cursor to the next row.
func (s *TableState[T]) SelectNext() {
	rows := s.Rows.Peek()
	s.CursorIndex.Update(func(i int) int {
		if i < len(rows)-1 {
			return i + 1
		}
		return i
	})
}

// SelectPrevious moves cursor to the previous row.
func (s *TableState[T]) SelectPrevious() {
	s.CursorIndex.Update(func(i int) int {
		if i > 0 {
			return i - 1
		}
		return i
	})
}

// SelectFirst moves cursor to the first row.
func (s *TableState[T]) SelectFirst() {
	s.CursorIndex.Set(0)
}

// SelectLast moves cursor to the last row.
func (s *TableState[T]) SelectLast() {
	rows := s.Rows.Peek()
	if len(rows) > 0 {
		s.CursorIndex.Set(len(rows) - 1)
	}
}

// SelectIndex sets cursor to a specific index, clamped to valid range.
func (s *TableState[T]) SelectIndex(index int) {
	rows := s.Rows.Peek()
	clamped := clampInt(index, 0, len(rows)-1)
	s.CursorIndex.Set(clamped)
}

// SelectColumn sets cursor to a specific column index.
func (s *TableState[T]) SelectColumn(index int) {
	s.CursorColumn.Set(index)
}

// clampCursor ensures cursor is within valid bounds after rows change.
func (s *TableState[T]) clampCursor() {
	rows := s.Rows.Peek()
	idx := s.CursorIndex.Peek()
	if len(rows) == 0 {
		s.CursorIndex.Set(0)
	} else if idx >= len(rows) {
		s.CursorIndex.Set(len(rows) - 1)
	}
}

func (s *TableState[T]) setViewIndices(indices []int) {
	s.viewIndices = indices
	if indices == nil {
		s.viewIndexBySource = nil
		return
	}
	viewIndexBySource := make(map[int]int, len(indices))
	for viewIdx, sourceIdx := range indices {
		viewIndexBySource[sourceIdx] = viewIdx
	}
	s.viewIndexBySource = viewIndexBySource
}

func (s *TableState[T]) viewIndexForSource(sourceIdx int) (int, bool) {
	if s.viewIndexBySource != nil {
		viewIdx, ok := s.viewIndexBySource[sourceIdx]
		return viewIdx, ok
	}
	if s.viewIndices != nil {
		for i, idx := range s.viewIndices {
			if idx == sourceIdx {
				return i, true
			}
		}
	}
	return 0, false
}

// ToggleSelection toggles the selection state of the row at the given index.
func (s *TableState[T]) ToggleSelection(index int) {
	s.Selection.Update(func(sel map[int]struct{}) map[int]struct{} {
		newSel := make(map[int]struct{}, len(sel))
		for k := range sel {
			newSel[k] = struct{}{}
		}
		if _, exists := newSel[index]; exists {
			delete(newSel, index)
		} else {
			newSel[index] = struct{}{}
		}
		return newSel
	})
}

// Select adds the row at the given index to the selection.
func (s *TableState[T]) Select(index int) {
	s.Selection.Update(func(sel map[int]struct{}) map[int]struct{} {
		newSel := make(map[int]struct{}, len(sel)+1)
		for k := range sel {
			newSel[k] = struct{}{}
		}
		newSel[index] = struct{}{}
		return newSel
	})
}

// Deselect removes the row at the given index from the selection.
func (s *TableState[T]) Deselect(index int) {
	s.Selection.Update(func(sel map[int]struct{}) map[int]struct{} {
		newSel := make(map[int]struct{}, len(sel))
		for k := range sel {
			if k != index {
				newSel[k] = struct{}{}
			}
		}
		return newSel
	})
}

// IsSelected returns true if the row at the given index is selected.
func (s *TableState[T]) IsSelected(index int) bool {
	sel := s.Selection.Peek()
	_, exists := sel[index]
	return exists
}

// ClearSelection removes all rows from the selection.
func (s *TableState[T]) ClearSelection() {
	s.Selection.Set(make(map[int]struct{}))
}

// SelectAll selects all rows in the table.
func (s *TableState[T]) SelectAll() {
	rows := s.Rows.Peek()
	sel := make(map[int]struct{}, len(rows))
	for i := range rows {
		sel[i] = struct{}{}
	}
	s.Selection.Set(sel)
}

// SelectedRows returns all currently selected rows.
// Note: This assumes row-based selection.
func (s *TableState[T]) SelectedRows() []T {
	rows := s.Rows.Peek()
	sel := s.Selection.Peek()
	result := make([]T, 0, len(sel))
	for i := range rows {
		if _, exists := sel[i]; exists {
			result = append(result, rows[i])
		}
	}
	return result
}

// SelectedIndices returns the indices of all selected rows in ascending order.
// Note: This assumes row-based selection.
func (s *TableState[T]) SelectedIndices() []int {
	sel := s.Selection.Peek()
	result := make([]int, 0, len(sel))
	for i := range sel {
		result = append(result, i)
	}
	// Sort for consistent ordering
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// SetAnchor sets the anchor point for shift-selection.
func (s *TableState[T]) SetAnchor(index int) {
	s.anchorIndex = &index
}

// ClearAnchor removes the anchor point.
func (s *TableState[T]) ClearAnchor() {
	s.anchorIndex = nil
}

// syncSelectionMode records how the Table showing this state interprets
// selection keys. A change of mode or column count invalidates the existing
// keys, so the selection is cleared.
func (s *TableState[T]) syncSelectionMode(mode TableSelectionMode, columnCount int) {
	if !s.hasSelectionMode {
		s.lastSelectionMode = mode
		s.columnCount = columnCount
		s.hasSelectionMode = true
		return
	}
	if s.lastSelectionMode == mode && s.columnCount == columnCount {
		return
	}
	s.ClearSelection()
	s.ClearAnchor()
	s.lastSelectionMode = mode
	s.columnCount = columnCount
}

// HasAnchor returns true if an anchor point is set.
func (s *TableState[T]) HasAnchor() bool {
	return s.anchorIndex != nil
}

// GetAnchor returns the anchor index, or -1 if no anchor is set.
func (s *TableState[T]) GetAnchor() int {
	if s.anchorIndex == nil {
		return -1
	}
	return *s.anchorIndex
}

// SelectRange selects all rows between from and to (inclusive).
func (s *TableState[T]) SelectRange(from, to int) {
	if from > to {
		from, to = to, from
	}
	rows := s.Rows.Peek()
	if from < 0 {
		from = 0
	}
	if to >= len(rows) {
		to = len(rows) - 1
	}
	sel := make(map[int]struct{}, to-from+1)
	for i := from; i <= to; i++ {
		sel[i] = struct{}{}
	}
	s.Selection.Set(sel)
}

// TableColumn defines layout properties for a table column.
type TableColumn struct {
	ID        string // Unique nonempty ID enables sorting and resizing
	Resizable bool
	MinWidth  int       // At least one cell for user resizing
	MaxWidth  int       // Zero means unlimited
	Width     Dimension // Optional width (Cells, Percent, Flex, Auto)
	Header    Widget    // Optional header widget for this column
}

// TableSelectionMode controls how cursor and selection highlights are applied.
type TableSelectionMode int

const (
	// TableSelectionCursor highlights only the cursor cell (default).
	TableSelectionCursor TableSelectionMode = iota
	// TableSelectionRow highlights the entire row.
	TableSelectionRow
	// TableSelectionColumn highlights the entire column.
	TableSelectionColumn
)

// Table is a generic focusable widget that displays a navigable table of rows.
// Use with Scrollable and a shared ScrollState to enable scroll-into-view.
type Table[T any] struct {
	Comparators         map[string]func(T, T) int                                                                     // Typed comparisons by unique column ID
	FrozenHeader        bool                                                                                          // Own a bounded viewport and keep headers visible
	FrozenColumns       int                                                                                           // Keep the first N columns visible during horizontal scrolling
	ID                  string                                                                                        // Optional unique identifier
	DisableFocus        bool                                                                                          // If true, prevent keyboard focus
	CursorStyle                                                                                                       // Embedded - CursorPrefix/SelectedPrefix fields for customizable indicators
	State               *TableState[T]                                                                                // Required - holds rows and cursor position
	Columns             []TableColumn                                                                                 // Required - defines column count and widths
	RenderCell          func(row T, rowIndex int, colIndex int, active bool, selected bool) Widget                    // Cell renderer (default uses fmt)
	RenderCellWithMatch func(row T, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget // Optional cell renderer with match data
	Filter              *FilterState                                                                                  // Optional filter state for matching rows
	MatchCell           func(row T, rowIndex int, colIndex int, query string, options FilterOptions) MatchResult      // Optional matcher per cell
	RenderHeader        func(colIndex int) Widget                                                                     // Optional header renderer (takes precedence over column headers)
	OnSelect            func(row T)                                                                                   // Callback invoked when Enter is pressed on a row or a row is double-clicked
	ActivateOnClick     bool                                                                                          // Invoke OnSelect on a single left click instead of a double-click
	OnCursorChange      func(row T)                                                                                   // Callback invoked when cursor moves to a different row
	ScrollState         *ScrollState                                                                                  // Optional state for scroll-into-view
	RowHeight           int                                                                                           // Optional uniform row height override (default 0 = layout metrics / fallback 1)
	ColumnSpacing       int                                                                                           // Space between columns
	RowSpacing          int                                                                                           // Space between rows
	SelectionMode       TableSelectionMode                                                                            // Cursor/selection highlight mode (row/column/cursor)
	MultiSelect         bool                                                                                          // Enable multi-select mode (shift+move, shift+click or drag to extend)
	Width               Dimension                                                                                     // Deprecated: use Style.Width
	Height              Dimension                                                                                     // Deprecated: use Style.Height
	Style               Style                                                                                         // Optional styling
	Click               func(MouseEvent)                                                                              // Optional callback invoked when clicked
	MouseDown           func(MouseEvent)                                                                              // Optional callback invoked when mouse is pressed
	MouseUp             func(MouseEvent)                                                                              // Optional callback invoked when mouse is released
	Hover               func(HoverEvent)                                                                              // Optional callback invoked when hover state changes
}

type tableRowLayout struct {
	y      int
	height int
}

type tableColumnLayout struct {
	x     int
	width int
}

type tableContainer[T any] struct {
	Table[T]
	children    []Widget
	rowCount    int
	columnCount int
	headerRows  int
}

type defaultTableCellWidget[T any] struct {
	table        Table[T]
	focusID      string
	focusManager *FocusManager
	theme        ThemeData
	row          T
	sourceRow    int
	colIndex     int
	match        MatchResult
	prefixWidth  int
}

// The public Table owns focus; its layout container must not register the same
// focus identity a second time and prevent Tab from advancing.
func (c tableContainer[T]) IsFocusable() bool { return false }

func (c tableContainer[T]) Build(ctx BuildContext) Widget {
	return c
}

func (c tableContainer[T]) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.ownsViewport() && c.State != nil {
		c.layoutViewport()
		// Layout observers run after final parent constraints are known. Apply
		// cursor reveal to the shared computed child slice before it is assigned
		// to child render trees, so the first frame uses the revealed viewport.
		if reclip := c.State.viewportMetrics.reclip; reclip != nil {
			scroll := c.viewportState()
			copy(metrics.layout.Children, reclip(scroll.GetOffsetX(), scroll.GetOffset()))
		}
		return
	}
	if c.State == nil || c.columnCount == 0 {
		if c.State != nil {
			c.State.rowLayouts = nil
			c.State.columnLayouts = nil
		}
		return
	}

	count := metrics.ChildCount()
	if count == 0 {
		c.State.rowLayouts = nil
		c.State.columnLayouts = nil
		return
	}

	if c.headerRows > 0 {
		if b, ok := metrics.ChildBounds(0); ok {
			c.State.headerHeight = b.Height
		}
	} else {
		c.State.headerHeight = 0
	}
	rowLayouts := make([]tableRowLayout, c.rowCount)
	seen := make([]bool, c.rowCount)
	columnLayouts := make([]tableColumnLayout, c.columnCount)

	for i := 0; i < count; i++ {
		bounds, ok := metrics.ChildBounds(i)
		if !ok {
			continue
		}
		if col := i % c.columnCount; i < c.columnCount || columnLayouts[col].width == 0 {
			columnLayouts[col] = tableColumnLayout{x: bounds.X, width: bounds.Width}
		}
		row := i / c.columnCount
		dataRow := row - c.headerRows
		if dataRow < 0 || dataRow >= c.rowCount {
			continue
		}
		if !seen[dataRow] {
			rowLayouts[dataRow] = tableRowLayout{y: bounds.Y, height: bounds.Height}
			seen[dataRow] = true
			continue
		}

		layout := rowLayouts[dataRow]
		top := layout.y
		bottom := layout.y + layout.height
		if bounds.Y < top {
			top = bounds.Y
		}
		if bounds.Y+bounds.Height > bottom {
			bottom = bounds.Y + bounds.Height
		}
		rowLayouts[dataRow] = tableRowLayout{y: top, height: bottom - top}
	}

	c.State.rowLayouts = rowLayouts
	c.State.columnLayouts = columnLayouts
	if c.selectionMode() != TableSelectionColumn {
		c.revealMovedCursor()
	}
}

func (c tableContainer[T]) ChildWidgets() []Widget {
	return c.children
}

func (w defaultTableCellWidget[T]) Build(ctx BuildContext) Widget {
	hover := newHoverTint(w.theme, func() bool { return w.table.cellHovered(w.sourceRow, w.colIndex) })
	content, ok := tableDefaultCellContent(w.row, w.colIndex)
	if !ok {
		if w.colIndex != 0 {
			return withHoverUnderlay(PresentStyledText("", Style{}, w.currentStyle, w.paintStyle), hover)
		} else {
			content = fmt.Sprintf("%v", w.row)
			return withHoverUnderlay(PresentPrefixedText(
				padCursorPrefix("", w.prefixWidth)+content,
				content,
				Style{},
				w.currentStyle,
				w.paintStyle,
				w.paintPrefix,
				func(*RenderContext) MatchResult { return w.match },
			), hover)
		}
	}
	return withHoverUnderlay(PresentHighlightedText(
		content,
		Style{},
		w.currentStyle,
		w.paintStyle,
		func(*RenderContext) MatchResult { return w.match },
	), hover)
}

func (w defaultTableCellWidget[T]) hoverKey() any {
	return w.table.cellHoverKey(w.sourceRow, w.colIndex)
}

func (w defaultTableCellWidget[T]) setHovered(hovered bool) {
	w.table.setCellHovered(w.sourceRow, w.colIndex, hovered)
}

func (w defaultTableCellWidget[T]) currentStyle() Style {
	mode := w.table.selectionMode()
	cursorRow := 0
	cursorCol := 0
	if w.table.State != nil {
		cursorRow = w.table.State.CursorIndex.Peek()
		cursorCol = w.table.State.CursorColumn.Peek()
	}

	selection := map[int]struct{}{}
	if w.table.MultiSelect && w.table.State != nil {
		if current := w.table.State.Selection.Peek(); current != nil {
			selection = current
		}
	}

	focused := w.focusManager != nil && w.focusID != "" && w.focusManager.FocusedID() == w.focusID
	active := tableCellActive(mode, w.sourceRow, w.colIndex, cursorRow, cursorCol)
	selected := false
	if w.table.MultiSelect {
		selected = tableCellSelected(mode, selection, w.sourceRow, w.colIndex, len(w.table.Columns))
	}
	return tableDefaultCellStyle(w.theme, active, selected, focused)
}

// paintState reports whether this cell shows the cursor or selection,
// subscribing only to changes in those answers.
func (w defaultTableCellWidget[T]) paintState(ctx *RenderContext) (active, selected, focused bool) {
	if w.table.State != nil {
		identity := func(i int) int { return i }
		mode := w.table.selectionMode()
		active = w.table.cellActiveSelect(mode, w.sourceRow, w.colIndex, identity, identity)
		if w.table.MultiSelect {
			selected = w.table.cellSelectedSelect(mode, w.sourceRow, w.colIndex)
		}
	} else {
		active = tableCellActive(w.table.selectionMode(), w.sourceRow, w.colIndex, 0, 0)
	}
	return active, selected, ctx.IsFocusedID(w.focusID)
}

func (w defaultTableCellWidget[T]) paintPrefix(ctx *RenderContext) string {
	active, selected, focused := w.paintState(ctx)
	showCursor := active && focused
	if showCursor {
		return w.table.CursorPrefix
	}
	if selected {
		return w.table.SelectedPrefix
	}
	return ""
}

func (w defaultTableCellWidget[T]) paintStyle(ctx *RenderContext) Style {
	active, selected, focused := w.paintState(ctx)
	return tableDefaultCellStyle(w.theme, active, selected, focused)
}

// cellActiveSelect reports whether a cell shows the cursor, subscribing only
// to changes in that answer. rowFor and colFor map stored cursor values to the
// rendered ones. In cursor mode a cell only watches the column while the cursor
// is in its row, so a horizontal move notifies just two cells.
func (t Table[T]) cellActiveSelect(mode TableSelectionMode, row, col int, rowFor, colFor func(int) int) bool {
	inRow := func() bool { return Select(t.State.CursorIndex, func(r int) bool { return rowFor(r) == row }) }
	inCol := func() bool { return Select(t.State.CursorColumn, func(c int) bool { return colFor(c) == col }) }
	switch mode {
	case TableSelectionColumn:
		return inCol()
	case TableSelectionCursor:
		return inRow() && inCol()
	default:
		return inRow()
	}
}

// cellSelectedSelect reports whether a cell is selected, subscribing only to
// changes in that answer.
func (t Table[T]) cellSelectedSelect(mode TableSelectionMode, row, col int) bool {
	columnCount := len(t.Columns)
	return SelectAny(t.State.Selection, func(selection map[int]struct{}) bool {
		return tableCellSelected(mode, selection, row, col, columnCount)
	})
}

// tableHoverCell is the part of a Table under the pointer: a cell, or with
// row or column highlighting, a row (col -1) or column (row -1).
type tableHoverCell struct {
	row, col int
}

// hoverCell returns the part of the table that hovering a cell highlights,
// following the selection mode as the cursor does.
func (t Table[T]) hoverCell(row, col int) tableHoverCell {
	switch t.selectionMode() {
	case TableSelectionRow:
		return tableHoverCell{row: row, col: -1}
	case TableSelectionColumn:
		return tableHoverCell{row: -1, col: col}
	}
	return tableHoverCell{row: row, col: col}
}

// cellHovered reports whether the cell shows the hover highlight,
// subscribing only to changes in that answer.
func (t Table[T]) cellHovered(row, col int) bool {
	return t.State != nil && t.State.hover.is(t.hoverCell(row, col))
}

func (t Table[T]) cellHoverKey(row, col int) any {
	return hoverItemKey{owner: t.State, item: t.hoverCell(row, col)}
}

func (t Table[T]) setCellHovered(row, col int, hovered bool) {
	if t.State != nil {
		t.State.hover.set(t.hoverCell(row, col), hovered)
	}
}

// WidgetID returns the table's unique identifier.
// Implements the Identifiable interface.
func (t Table[T]) WidgetID() string {
	return t.ID
}

// GetContentDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (t Table[T]) GetContentDimensions() (width, height Dimension) {
	dims := t.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = t.Width
	}
	if height.IsUnset() {
		height = t.Height
	}
	return width, height
}

// GetStyle returns the style of the table widget.
// Implements the Styled interface.
func (t Table[T]) GetStyle() Style {
	return t.Style
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (t Table[T]) OnClick(event MouseEvent) {
	if t.Click != nil {
		t.Click(event)
	}
}

func (t Table[T]) ownsDescendantPointer() {}

// OnMouseDown moves the cursor to the clicked cell, extends the selection on
// shift+click in multi-select mode, and selects the row on double-click.
// Implements the MouseDownHandler interface.
func (t Table[T]) OnMouseDown(event MouseEvent) {
	t.handleMouseDown(event)
	if t.MouseDown != nil {
		t.MouseDown(event)
	}
}

func (t Table[T]) handleMouseDown(event MouseEvent) {
	if t.State == nil {
		return
	}
	t.State.dragging = false
	if t.headerMouseDown(event) {
		return
	}
	viewRow, col, ok := t.cellFromMouse(event, false)
	if !ok {
		return
	}
	previous := t.State.CursorIndex.Peek()

	if t.MultiSelect && event.Mod.Contains(uv.ModShift) {
		t.extendSelectionTo(viewRow, col)
	} else {
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		t.moveCursorTo(viewRow, col)
	}
	t.State.dragging = event.Button == uv.MouseLeft

	if t.State.CursorIndex.Peek() != previous {
		t.notifyCursorChange()
	}
	if clickActivates(event, t.ActivateOnClick) {
		t.selectRow()
	}
}

// OnMouseMove drags the cursor to the cell under the pointer while a press on
// a cell is held, extending the selection from the pressed cell in
// multi-select mode. Dragging past the top or bottom scrolls.
// Implements the MouseMoveHandler interface.
func (t Table[T]) OnMouseMove(event MouseEvent) {
	if t.State != nil && t.State.resizing {
		t.resizeColumn(t.State.resizeColumn, t.State.resizeStartWidth+event.X-t.State.resizeStartX)
		return
	}
	if t.State == nil || !t.State.dragging {
		return
	}
	viewRow, col, ok := t.cellFromMouse(event, true)
	if !ok {
		return
	}
	previous := t.State.CursorIndex.Peek()
	if t.MultiSelect {
		t.extendSelectionTo(viewRow, col)
	} else {
		t.moveCursorTo(viewRow, col)
	}
	if t.State.CursorIndex.Peek() != previous {
		t.notifyCursorChange()
	}
}

// moveCursorTo puts the cursor on a cell, given by view row and column.
func (t Table[T]) moveCursorTo(viewRow, col int) {
	t.setCursorToViewIndex(viewRow)
	if t.selectionMode() != TableSelectionRow {
		t.State.CursorColumn.Set(col)
	}
	t.jumpCursorIntoView()
}

// extendSelectionTo moves the cursor to a cell, selecting the rows, columns
// or cells between it and the anchor, as the selection mode dictates.
func (t Table[T]) extendSelectionTo(viewRow, col int) {
	columnCount := len(t.Columns)
	switch t.selectionMode() {
	case TableSelectionRow:
		t.handleShiftMoveRowTo(viewRow)
	case TableSelectionColumn:
		t.handleShiftMoveColumnTo(col, columnCount)
	default:
		t.handleShiftMoveCellTo(viewRow, col, columnCount)
	}
}

// cellFromMouse returns the view row and column of the cell under the
// pointer. With clamp, a pointer outside the cells gives the nearest one.
func (t Table[T]) cellFromMouse(event MouseEvent, clamp bool) (viewRow, col int, ok bool) {
	rows, columns := t.State.rowLayouts, t.State.columnLayouts
	rowCount := min(len(rows), len(t.viewIndices()))
	if rowCount == 0 || len(columns) != len(t.Columns) {
		return 0, 0, false
	}
	inset := t.Style.Border.Width()
	y := event.LocalY - inset - t.Style.Padding.Top
	x := event.LocalX - inset - t.Style.Padding.Left
	if t.ownsViewport() {
		m := t.State.viewportMetrics
		if !clamp && (x < 0 || y < 0 || x >= m.width || y >= m.height || y < m.frozenHeight) {
			return 0, 0, false
		}
		if y >= m.frozenHeight {
			y += t.viewportState().GetOffset()
		}
		if x >= m.frozenWidth {
			x += t.viewportState().GetOffsetX()
		}
	}
	viewRow, ok = spanAt(rowCount, func(i int) (int, int) { return rows[i].y, rows[i].height }, y, clamp)
	if !ok {
		return 0, 0, false
	}
	// Clamped, so a click in the gap between columns picks the column before it.
	col, ok = spanAt(len(columns), func(i int) (int, int) { return columns[i].x, columns[i].width }, x, true)
	return viewRow, col, ok
}

// OnMouseUp ends a drag begun on a cell.
// Implements the MouseUpHandler interface.
func (t Table[T]) OnMouseUp(event MouseEvent) {
	if t.State != nil {
		t.State.dragging = false
		t.State.resizing = false
	}
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called on hover enter/leave transitions.
// Implements the Hoverable interface.
func (t Table[T]) OnHover(event HoverEvent) {
	if t.Hover != nil {
		t.Hover(event)
	}
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (t Table[T]) IsFocusable() bool {
	return !t.DisableFocus
}

// Build returns a table container that arranges the rendered cells.
func (t Table[T]) Build(ctx BuildContext) Widget {
	if t.State == nil || len(t.Columns) == 0 {
		return Column{}
	}

	// Width overrides change clipping and hit regions for every row. Rebuild
	// the table cells so retained paint caches cannot retain old clip extents.
	_ = t.State.ColumnWidths.Get()
	renderCell := t.RenderCell
	renderCellWithMatch := t.RenderCellWithMatch
	useDefaultRenderer := renderCellWithMatch == nil && renderCell == nil

	rows := t.State.Rows.Get()
	columnCount := len(t.Columns)
	mode := t.selectionMode()
	query, options := filterStateValues(t.Filter)
	viewRows, viewIndices, viewMatches := t.filteredRows(rows, columnCount, query, options)
	viewRows, viewIndices, viewMatches = t.sortedRows(viewRows, viewIndices, viewMatches, t.State.Sort.Get())
	t.State.setViewIndices(viewIndices)

	hasHeader := t.hasHeader()
	headerRows := 0
	var headerCells []Widget
	if hasHeader {
		headerRows = 1
		headerCells = make([]Widget, columnCount)
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			var header Widget
			if t.RenderHeader != nil {
				header = t.RenderHeader(colIdx)
			}
			if header == nil {
				header = t.Columns[colIdx].Header
			}
			if header == nil {
				header = Text{}
			}
			if t.sortableColumn(colIdx) {
				indicator := "↕"
				order := t.State.Sort.Get()
				if order.ColumnID == t.Columns[colIdx].ID {
					if order.Direction == TableSortAscending {
						indicator = "↑"
					}
					if order.Direction == TableSortDescending {
						indicator = "↓"
					}
				}
				header = Row{Children: []Widget{header, Text{Content: " " + indicator}}}
			}
			headerCells[colIdx] = header
		}
	}

	if len(viewRows) == 0 && headerRows == 0 && !t.ownsViewport() {
		t.State.rowLayouts = nil
		return Column{}
	}

	children := make([]Widget, 0, (len(viewRows)+headerRows)*columnCount)
	if headerRows > 0 {
		children = append(children, headerCells...)
	}

	if useDefaultRenderer {
		focusID := widgetIdentity(t, ctx)
		theme := ctx.Theme()
		prefixWidth := cursorPrefixSlotWidth(t.CursorPrefix, t.SelectedPrefix)
		for viewRowIdx, row := range viewRows {
			sourceRowIdx := viewIndices[viewRowIdx]
			for colIdx := 0; colIdx < columnCount; colIdx++ {
				match := MatchResult{}
				if len(viewMatches) > 0 {
					match = viewMatches[viewRowIdx][colIdx]
				}
				children = append(children, defaultTableCellWidget[T]{
					table:        t,
					focusID:      focusID,
					focusManager: ctx.focusManager,
					theme:        theme,
					row:          row,
					sourceRow:    sourceRowIdx,
					colIndex:     colIdx,
					match:        match,
					prefixWidth:  prefixWidth,
				})
			}
		}

		children = t.viewportCells(children)
		return tableContainer[T]{
			Table:       t,
			children:    children,
			rowCount:    len(viewRows),
			columnCount: columnCount,
			headerRows:  headerRows,
		}
	}

	if renderCellWithMatch == nil && renderCell == nil {
		renderCellWithMatch = t.themedDefaultRenderCell(ctx)
	}

	if renderCellWithMatch == nil {
		renderCellWithMatch = func(row T, rowIndex, colIndex int, active, selected bool, _ MatchResult) Widget {
			return renderCell(row, rowIndex, colIndex, active, selected)
		}
	}

	// Cells read the cursor and selection in their own Build, so moving the
	// cursor rebuilds only the cells whose state changes. The rendered cursor
	// is clamped to the rows and columns, and falls back to the first visible
	// row when the stored row is filtered out.
	rowCount, firstRow := len(rows), 0
	if len(viewIndices) > 0 {
		firstRow = viewIndices[0]
	}
	rowFor := func(r int) int {
		if rowCount > 0 {
			r = clampInt(r, 0, rowCount-1)
		}
		if _, ok := t.State.viewIndexForSource(r); !ok {
			return firstRow
		}
		return r
	}
	colFor := func(c int) int { return clampInt(c, 0, columnCount-1) }
	for viewRowIdx, row := range viewRows {
		sourceRowIdx := viewIndices[viewRowIdx]
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			match := MatchResult{}
			if len(viewMatches) > 0 {
				match = viewMatches[viewRowIdx][colIdx]
			}
			children = append(children, tableCell[T]{
				table: t, mode: mode, row: row, sourceRow: sourceRowIdx, col: colIdx,
				match: match, rowFor: rowFor, colFor: colFor, render: renderCellWithMatch,
			})
		}
	}

	children = t.viewportCells(children)
	return tableContainer[T]{
		Table:       t,
		children:    children,
		rowCount:    len(viewRows),
		columnCount: columnCount,
		headerRows:  headerRows,
	}
}

// tableCell renders one cell of a Table with a custom RenderCell.
type tableCell[T any] struct {
	table          Table[T]
	mode           TableSelectionMode
	row            T
	sourceRow, col int
	match          MatchResult
	rowFor, colFor func(int) int
	render         func(row T, rowIndex, colIndex int, active, selected bool, match MatchResult) Widget
}

func (c tableCell[T]) Build(ctx BuildContext) Widget {
	active := c.table.cellActiveSelect(c.mode, c.sourceRow, c.col, c.rowFor, c.colFor)
	selected := c.table.MultiSelect && c.table.cellSelectedSelect(c.mode, c.sourceRow, c.col)
	cell := c.render(c.row, c.sourceRow, c.col, active, selected, c.match)
	if cell == nil {
		cell = Text{}
	}
	// Keep the rendered cell a child so its own Build still runs.
	return hoverUnderlay{
		passThrough: passThrough{child: cell},
		hoverTint:   newHoverTint(ctx.Theme(), func() bool { return c.table.cellHovered(c.sourceRow, c.col) }),
	}
}

func (c tableCell[T]) hoverKey() any { return c.table.cellHoverKey(c.sourceRow, c.col) }

func (c tableCell[T]) setHovered(hovered bool) { c.table.setCellHovered(c.sourceRow, c.col, hovered) }

// themedDefaultRenderCell returns a themed render function for table cells.
// Captures theme colors and widget focus state from the context for use in the render function.
// Cursor highlighting is only shown when the widget has focus.
func (t Table[T]) themedDefaultRenderCell(ctx BuildContext) func(row T, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget {
	theme := ctx.Theme()
	widgetFocused := ctx.IsFocused(t)
	cursorPrefix := t.CursorPrefix
	selectedPrefix := t.SelectedPrefix

	highlight := MatchHighlightStyle(theme)
	return func(row T, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget {
		style := tableDefaultCellStyle(theme, active, selected, widgetFocused)
		if content, ok := tableDefaultCellContent(row, colIndex); ok {
			if match.Matched && len(match.Ranges) > 0 {
				return Text{
					Spans: HighlightSpans(content, match.Ranges, highlight),
					Style: style,
				}
			}
			return Text{
				Content: content,
				Style:   style,
			}
		}

		if colIndex != 0 {
			return Text{Content: "", Style: style}
		}

		content := fmt.Sprintf("%v", row)
		prefix := ""

		// Only show cursor prefix when widget has focus
		showCursor := active && widgetFocused

		if showCursor {
			prefix = cursorPrefix
		} else if selected {
			prefix = selectedPrefix
		}

		if match.Matched && len(match.Ranges) > 0 {
			spans := make([]Span, 0, 1+len(match.Ranges)*2)
			if prefix != "" {
				spans = append(spans, Span{Text: prefix})
			}
			spans = append(spans, HighlightSpans(content, match.Ranges, highlight)...)
			return Text{
				Spans: spans,
				Style: style,
			}
		}

		return Text{Content: prefix + content, Style: style}
	}
}

func (t Table[T]) filteredRows(rows []T, columnCount int, query string, options FilterOptions) ([]T, []int, [][]MatchResult) {
	if query == "" || columnCount == 0 {
		viewIndices := make([]int, len(rows))
		for i := range rows {
			viewIndices[i] = i
		}
		return rows, viewIndices, nil
	}

	matchCell := t.MatchCell
	if matchCell == nil {
		matchCell = defaultTableMatchCell[T]
	}

	viewRows := make([]T, 0, len(rows))
	viewIndices := make([]int, 0, len(rows))
	viewMatches := make([][]MatchResult, 0, len(rows))
	rowBest := make([]MatchResult, 0, len(rows))

	for rowIdx, row := range rows {
		cellMatches := make([]MatchResult, columnCount)
		// A row ranks by its best-matching cell.
		var best MatchResult
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			match := matchCell(row, rowIdx, colIdx, query, options)
			cellMatches[colIdx] = match
			if matchRanksAhead(match, best) {
				best = match
			}
		}
		if best.Matched {
			viewRows = append(viewRows, row)
			viewIndices = append(viewIndices, rowIdx)
			viewMatches = append(viewMatches, cellMatches)
			rowBest = append(rowBest, best)
		}
	}

	if options.Mode == FilterFuzzy && len(viewRows) > 1 {
		order := make([]int, len(viewRows))
		for i := range order {
			order[i] = i
		}
		sort.SliceStable(order, func(i, j int) bool {
			return matchRanksAhead(rowBest[order[i]], rowBest[order[j]])
		})

		sortedRows := make([]T, len(viewRows))
		sortedIndices := make([]int, len(viewIndices))
		sortedMatches := make([][]MatchResult, len(viewMatches))
		for i, originalIdx := range order {
			sortedRows[i] = viewRows[originalIdx]
			sortedIndices[i] = viewIndices[originalIdx]
			sortedMatches[i] = viewMatches[originalIdx]
		}
		viewRows = sortedRows
		viewIndices = sortedIndices
		viewMatches = sortedMatches
	}

	return viewRows, viewIndices, viewMatches
}

func defaultTableMatchCell[T any](row T, rowIndex int, colIndex int, query string, options FilterOptions) MatchResult {
	if content, ok := tableDefaultCellContent(row, colIndex); ok {
		return MatchString(content, query, options)
	}
	if colIndex != 0 {
		return MatchResult{}
	}
	return MatchString(fmt.Sprintf("%v", row), query, options)
}

// OnKey handles keys not covered by declarative keybindings.
// Implements the Focusable interface.
func (t Table[T]) OnKey(event KeyEvent) bool {
	return false
}

// Keybinds returns the declarative keybindings for this table.
func (t Table[T]) Keybinds() []Keybind {
	if t.State == nil {
		return nil
	}
	mode := t.selectionMode()

	binds := []Keybind{
		{Key: "enter", Action: t.selectRow, Hidden: true},
		{Key: "up", Action: t.keyCursorUp, Hidden: true},
		{Key: "k", Action: t.keyCursorUp, Hidden: true},
		{Key: "down", Action: t.keyCursorDown, Hidden: true},
		{Key: "j", Action: t.keyCursorDown, Hidden: true},
		{Key: "home", Action: t.keyCursorToFirst, Hidden: true},
		{Key: "g", Action: t.keyCursorToFirst, Hidden: true},
		{Key: "end", Action: t.keyCursorToLast, Hidden: true},
		{Key: "G", Action: t.keyCursorToLast, Hidden: true},
		{Key: "pgup", Action: t.pageUp, Hidden: true},
		{Key: "ctrl+u", Action: t.pageUp, Hidden: true},
		{Key: "pgdown", Action: t.pageDown, Hidden: true},
		{Key: "ctrl+d", Action: t.pageDown, Hidden: true},
	}

	for col := range t.Columns {
		if t.sortableColumn(col) {
			binds = append(binds, Keybind{Key: "ctrl+s", Name: "Sort", Action: func() { t.cycleSort(t.State.CursorColumn.Peek()) }})
			break
		}
	}
	binds = append(binds,
		Keybind{Key: "ctrl+left", Action: func() { t.resizeCurrent(-1) }, Hidden: true},
		Keybind{Key: "ctrl+right", Action: func() { t.resizeCurrent(1) }, Hidden: true},
		Keybind{Key: "ctrl+r", Action: t.resetCurrentWidth, Hidden: true},
	)
	if t.ownsViewport() {
		binds = append(binds,
			Keybind{Key: "alt+left", Action: func() { t.viewportState().ScrollLeft(3) }, Hidden: true},
			Keybind{Key: "alt+right", Action: func() { t.viewportState().ScrollRight(3) }, Hidden: true},
		)
	}
	// Left/right move between cells or columns (rows have no horizontal cursor)
	if mode == TableSelectionCursor || mode == TableSelectionColumn {
		binds = append(binds,
			Keybind{Key: "left", Action: t.keyCursorLeft, Hidden: true},
			Keybind{Key: "h", Action: t.keyCursorLeft, Hidden: true},
			Keybind{Key: "right", Action: t.keyCursorRight, Hidden: true},
			Keybind{Key: "l", Action: t.keyCursorRight, Hidden: true},
		)
	}

	// Shift keybinds conditional on MultiSelect and mode
	if t.MultiSelect {
		switch mode {
		case TableSelectionRow:
			binds = append(binds,
				Keybind{Key: "shift+up", Action: t.shiftRowUp, Hidden: true},
				Keybind{Key: "shift+k", Action: t.shiftRowUp, Hidden: true},
				Keybind{Key: "shift+down", Action: t.shiftRowDown, Hidden: true},
				Keybind{Key: "shift+j", Action: t.shiftRowDown, Hidden: true},
				Keybind{Key: "shift+home", Action: t.shiftRowToFirst, Hidden: true},
				Keybind{Key: "shift+end", Action: t.shiftRowToLast, Hidden: true},
			)
		case TableSelectionColumn:
			binds = append(binds,
				Keybind{Key: "shift+left", Action: t.shiftColumnLeft, Hidden: true},
				Keybind{Key: "shift+h", Action: t.shiftColumnLeft, Hidden: true},
				Keybind{Key: "shift+right", Action: t.shiftColumnRight, Hidden: true},
				Keybind{Key: "shift+l", Action: t.shiftColumnRight, Hidden: true},
				Keybind{Key: "shift+home", Action: t.shiftColumnToFirst, Hidden: true},
				Keybind{Key: "shift+end", Action: t.shiftColumnToLast, Hidden: true},
			)
		case TableSelectionCursor:
			binds = append(binds,
				Keybind{Key: "shift+up", Action: t.shiftCellUp, Hidden: true},
				Keybind{Key: "shift+k", Action: t.shiftCellUp, Hidden: true},
				Keybind{Key: "shift+down", Action: t.shiftCellDown, Hidden: true},
				Keybind{Key: "shift+j", Action: t.shiftCellDown, Hidden: true},
				Keybind{Key: "shift+left", Action: t.shiftCellLeft, Hidden: true},
				Keybind{Key: "shift+h", Action: t.shiftCellLeft, Hidden: true},
				Keybind{Key: "shift+right", Action: t.shiftCellRight, Hidden: true},
				Keybind{Key: "shift+l", Action: t.shiftCellRight, Hidden: true},
				Keybind{Key: "shift+home", Action: t.shiftCellToFirst, Hidden: true},
				Keybind{Key: "shift+end", Action: t.shiftCellToLast, Hidden: true},
			)
		}
	}

	return binds
}

func (t Table[T]) selectRow() {
	if _, _, ok := t.normalizeRowCursorForInteraction(); !ok {
		return
	}
	if t.OnSelect != nil {
		if row, ok := t.State.SelectedRow(); ok {
			t.OnSelect(row)
		}
	}
}

func (t Table[T]) keyCursorUp() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(-1, false)
		return
	}
	_, cursorViewIdx, ok := t.normalizeRowCursorForInteraction()
	if !ok {
		return
	}
	if cursorViewIdx == 0 {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx - 1)
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorDown() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(1, false)
		return
	}
	view, cursorViewIdx, ok := t.normalizeRowCursorForInteraction()
	if !ok {
		return
	}
	if cursorViewIdx >= len(view)-1 {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx + 1)
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorToFirst() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		if t.ownsViewport() || t.ScrollState != nil {
			t.viewportState().animateOffset(0)
		}
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(0)
	t.glideCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorToLast() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		if t.ownsViewport() || t.ScrollState != nil {
			t.viewportState().animateOffset(maxTableInt())
		}
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(len(view) - 1)
	t.glideCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) pageUp() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(-10, true)
		return
	}
	_, cursorViewIdx, ok := t.normalizeRowCursorForInteraction()
	if !ok {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx - 10)
	t.glideCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) pageDown() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(10, true)
		return
	}
	_, cursorViewIdx, ok := t.normalizeRowCursorForInteraction()
	if !ok {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx + 10)
	t.glideCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorLeft() {
	columnCount := len(t.Columns)
	if !t.normalizeColumnCursorForInteraction(columnCount) {
		return
	}
	cursorCol := t.State.CursorColumn.Peek()
	if cursorCol <= 0 {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.State.CursorColumn.Set(cursorCol - 1)
	if t.ownsViewport() {
		t.scrollCursorIntoView()
	}
}

func (t Table[T]) keyCursorRight() {
	columnCount := len(t.Columns)
	if !t.normalizeColumnCursorForInteraction(columnCount) {
		return
	}
	cursorCol := t.State.CursorColumn.Peek()
	if cursorCol >= columnCount-1 {
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.State.CursorColumn.Set(cursorCol + 1)
	if t.ownsViewport() {
		t.scrollCursorIntoView()
	}
}

func (t Table[T]) shiftRowUp() {
	t.handleShiftMoveRow(-1)
}

func (t Table[T]) shiftRowDown() {
	t.handleShiftMoveRow(1)
}

func (t Table[T]) shiftRowToFirst() {
	t.handleShiftMoveRowTo(0)
}

func (t Table[T]) shiftRowToLast() {
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	t.handleShiftMoveRowTo(len(view) - 1)
}

func (t Table[T]) shiftColumnLeft() {
	t.handleShiftMoveColumn(-1, len(t.Columns))
}

func (t Table[T]) shiftColumnRight() {
	t.handleShiftMoveColumn(1, len(t.Columns))
}

func (t Table[T]) shiftColumnToFirst() {
	t.handleShiftMoveColumnTo(0, len(t.Columns))
}

func (t Table[T]) shiftColumnToLast() {
	columnCount := len(t.Columns)
	if columnCount == 0 {
		return
	}
	t.handleShiftMoveColumnTo(columnCount-1, columnCount)
}

func (t Table[T]) shiftCellUp() {
	t.handleShiftMoveCell(-1, 0, len(t.Columns))
}

func (t Table[T]) shiftCellDown() {
	t.handleShiftMoveCell(1, 0, len(t.Columns))
}

func (t Table[T]) shiftCellLeft() {
	t.handleShiftMoveCell(0, -1, len(t.Columns))
}

func (t Table[T]) shiftCellRight() {
	t.handleShiftMoveCell(0, 1, len(t.Columns))
}

func (t Table[T]) shiftCellToFirst() {
	t.handleShiftMoveCellTo(0, 0, len(t.Columns))
}

func (t Table[T]) shiftCellToLast() {
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	columnCount := len(t.Columns)
	if columnCount == 0 {
		return
	}
	t.handleShiftMoveCellTo(len(view)-1, columnCount-1, columnCount)
}

// handleShiftMoveRow extends row selection by moving cursor by delta.
func (t Table[T]) handleShiftMoveRow(delta int) {
	if t.State == nil {
		return
	}

	view := t.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorRow := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
	if !ok {
		cursorRow = view[0]
		t.State.CursorIndex.Set(cursorRow)
		cursorViewIdx = 0
	}

	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorRow)
	}

	newViewIdx := clampInt(cursorViewIdx+delta, 0, len(view)-1)
	newCursor := view[newViewIdx]
	t.State.CursorIndex.Set(newCursor)
	t.setSelectionRangeFromView(view, t.State.GetAnchor(), newCursor)
	t.jumpCursorIntoView()
}

// handleShiftMoveRowTo extends row selection to a specific index.
func (t Table[T]) handleShiftMoveRowTo(targetIdx int) {
	if t.State == nil {
		return
	}

	view := t.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorRow := t.State.CursorIndex.Peek()
	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorRow)
	}

	targetViewIdx := clampInt(targetIdx, 0, len(view)-1)
	newCursor := view[targetViewIdx]
	t.State.CursorIndex.Set(newCursor)
	t.setSelectionRangeFromView(view, t.State.GetAnchor(), newCursor)
	t.jumpCursorIntoView()
}

// handleShiftMoveColumn extends column selection by moving cursor column by delta.
func (t Table[T]) handleShiftMoveColumn(delta int, columnCount int) {
	if columnCount == 0 {
		return
	}
	cursorCol := t.State.CursorColumn.Peek()
	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorCol)
	}

	target := clampInt(cursorCol+delta, 0, columnCount-1)
	t.State.CursorColumn.Set(target)
	t.setSelectionRange(t.State.GetAnchor(), target, columnCount)
	if t.ownsViewport() {
		t.scrollCursorIntoView()
	}
}

// handleShiftMoveColumnTo extends column selection to a specific index.
func (t Table[T]) handleShiftMoveColumnTo(targetIdx int, columnCount int) {
	if columnCount == 0 {
		return
	}
	cursorCol := t.State.CursorColumn.Peek()
	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorCol)
	}

	target := clampInt(targetIdx, 0, columnCount-1)
	t.State.CursorColumn.Set(target)
	t.setSelectionRange(t.State.GetAnchor(), target, columnCount)
	if t.ownsViewport() {
		t.scrollCursorIntoView()
	}
}

// handleShiftMoveCell extends cell selection by moving cursor by row/col deltas.
func (t Table[T]) handleShiftMoveCell(deltaRow, deltaCol, columnCount int) {
	if t.State == nil || columnCount == 0 {
		return
	}

	view := t.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorRow := t.State.CursorIndex.Peek()
	cursorCol := t.State.CursorColumn.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
	if !ok {
		cursorRow = view[0]
		t.State.CursorIndex.Set(cursorRow)
		cursorViewIdx = 0
	}

	if !t.State.HasAnchor() {
		t.State.SetAnchor(cellIndex(cursorRow, cursorCol, columnCount))
	}

	newViewRow := clampInt(cursorViewIdx+deltaRow, 0, len(view)-1)
	newCol := clampInt(cursorCol+deltaCol, 0, columnCount-1)
	newRow := view[newViewRow]
	t.State.CursorIndex.Set(newRow)
	t.State.CursorColumn.Set(newCol)

	anchorRow, anchorCol := cellIndexToRowCol(t.State.GetAnchor(), columnCount)
	anchorViewRow, ok := t.viewIndexForSource(anchorRow)
	if !ok {
		anchorViewRow = newViewRow
		t.State.SetAnchor(cellIndex(newRow, newCol, columnCount))
		anchorCol = newCol
	}
	t.setSelectionBox(view, anchorViewRow, anchorCol, newViewRow, newCol, columnCount)
	t.jumpCursorIntoView()
}

// handleShiftMoveCellTo extends cell selection to a specific cell.
func (t Table[T]) handleShiftMoveCellTo(targetRow, targetCol, columnCount int) {
	if t.State == nil || columnCount == 0 {
		return
	}

	view := t.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorRow := t.State.CursorIndex.Peek()
	cursorCol := t.State.CursorColumn.Peek()
	if !t.State.HasAnchor() {
		t.State.SetAnchor(cellIndex(cursorRow, cursorCol, columnCount))
	}

	newViewRow := clampInt(targetRow, 0, len(view)-1)
	newCol := clampInt(targetCol, 0, columnCount-1)
	newRow := view[newViewRow]
	t.State.CursorIndex.Set(newRow)
	t.State.CursorColumn.Set(newCol)

	anchorRow, anchorCol := cellIndexToRowCol(t.State.GetAnchor(), columnCount)
	anchorViewRow, ok := t.viewIndexForSource(anchorRow)
	if !ok {
		anchorViewRow = newViewRow
		t.State.SetAnchor(cellIndex(newRow, newCol, columnCount))
		anchorCol = newCol
	}
	t.setSelectionBox(view, anchorViewRow, anchorCol, newViewRow, newCol, columnCount)
	t.jumpCursorIntoView()
}

// scrollBy scrolls the viewport by lines, gliding there if animate. A table
// that owns its viewport (frozen header or columns) always jumps.
func (t Table[T]) scrollBy(lines int, animate bool) bool {
	if t.ownsViewport() {
		if lines < 0 {
			return t.viewportState().ScrollUp(-lines)
		}
		return t.viewportState().ScrollDown(lines)
	}
	if t.ScrollState == nil {
		return false
	}
	if lines < 0 {
		return t.ScrollState.scrollUp(-lines, animate)
	}
	return t.ScrollState.scrollDown(lines, animate)
}

// jumpCursorIntoView interrupts a glide for range selection and pointer moves.
func (t Table[T]) jumpCursorIntoView() {
	if t.ownsViewport() || t.ScrollState != nil {
		t.viewportState().stopAnimation()
	}
	t.scrollCursorIntoView()
}

// scrollCursorIntoView uses the ScrollState to ensure
// the cursor row is visible in the viewport.
func (t Table[T]) scrollCursorIntoView() {
	t.revealCursor(false)
}

// glideCursorIntoView is scrollCursorIntoView after a long move (a page, or to
// the start or end): the viewport glides to the cursor so the eye can follow
// the content.
func (t Table[T]) glideCursorIntoView() {
	t.revealCursor(true)
}

func (t Table[T]) revealCursor(animate bool) {
	if t.ownsViewport() {
		t.revealViewportCursor(true, animate)
		return
	}
	if t.ScrollState == nil || t.State == nil {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	rowY, rowHeight, ok := t.cursorRegion(cursorIdx)
	if !ok {
		return
	}
	t.State.revealed.record(cursorIdx, rowY, rowHeight, t.ScrollState)
	if animate {
		t.ScrollState.glideToView(rowY, rowHeight)
	} else {
		t.ScrollState.ScrollToView(rowY, rowHeight)
	}
}

// revealMovedCursor scrolls the cursor into view after layout, unless it was
// already revealed at its current position. Mouse wheel scrolling moves only
// the viewport, so it must not be undone by the next layout.
func (t Table[T]) revealMovedCursor() {
	if t.ownsViewport() {
		t.revealViewportCursor(false, false)
		return
	}
	if t.ScrollState == nil || t.State == nil {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	rowY, rowHeight, ok := t.cursorRegion(cursorIdx)
	if !ok || !t.State.revealed.needed(cursorIdx, rowY, rowHeight, t.ScrollState) {
		return
	}
	t.ScrollState.ScrollToView(rowY, rowHeight)
}

// cursorRegion returns the content rows occupied by the cursor row.
func (t Table[T]) cursorRegion(cursorIdx int) (y, height int, ok bool) {
	viewIdx, ok := t.viewIndexForSource(cursorIdx)
	if !ok {
		return 0, 0, false
	}
	y, height, ok = t.getRowLayout(cursorIdx)
	if !ok {
		height = t.getRowHeight()
		y = viewIdx * height
	}
	return y, height, true
}

// getRowHeight returns the fallback uniform height of table rows.
func (t Table[T]) getRowHeight() int {
	if t.RowHeight > 0 {
		return t.RowHeight
	}
	return 1
}

// getRowLayout returns the cached row layout for the given index.
func (t Table[T]) getRowLayout(index int) (y, height int, ok bool) {
	if t.State == nil {
		return 0, 0, false
	}
	viewIdx, ok := t.viewIndexForSource(index)
	if !ok {
		return 0, 0, false
	}
	if viewIdx < 0 || viewIdx >= len(t.State.rowLayouts) {
		return 0, 0, false
	}
	layout := t.State.rowLayouts[viewIdx]
	if layout.height <= 0 {
		return 0, 0, false
	}
	return layout.y, layout.height, true
}

// notifyCursorChange calls OnCursorChange with the current row if the callback is set.
func (t Table[T]) notifyCursorChange() {
	if t.OnCursorChange == nil || t.State == nil {
		return
	}
	if row, ok := t.State.SelectedRow(); ok {
		t.OnCursorChange(row)
	}
}

func (t Table[T]) setCursorToViewIndex(viewIdx int) {
	if t.State == nil {
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	viewIdx = clampInt(viewIdx, 0, len(view)-1)
	t.State.SelectIndex(view[viewIdx])
}

// normalizeRowCursorForInteraction clamps the source row cursor to bounds and
// ensures it points at a visible row when filtering is active.
func (t Table[T]) normalizeRowCursorForInteraction() (view []int, cursorViewIdx int, ok bool) {
	if t.State == nil {
		return nil, 0, false
	}

	view = t.viewIndices()
	if len(view) == 0 {
		return nil, 0, false
	}

	rows := t.State.Rows.Peek()
	if len(rows) == 0 {
		return nil, 0, false
	}

	cursorRow := t.State.CursorIndex.Peek()
	clamped := clampInt(cursorRow, 0, len(rows)-1)
	if clamped != cursorRow {
		cursorRow = clamped
		t.State.CursorIndex.Set(cursorRow)
	}

	cursorViewIdx, ok = t.viewIndexForSource(cursorRow)
	if ok {
		return view, cursorViewIdx, true
	}

	t.State.CursorIndex.Set(view[0])
	return view, 0, true
}

func (t Table[T]) normalizeColumnCursorForInteraction(columnCount int) bool {
	if t.State == nil || columnCount == 0 {
		return false
	}

	cursorCol := t.State.CursorColumn.Peek()
	clamped := clampInt(cursorCol, 0, columnCount-1)
	if clamped != cursorCol {
		t.State.CursorColumn.Set(clamped)
	}
	return true
}

func (t Table[T]) viewIndices() []int {
	if t.State == nil {
		return nil
	}
	if t.State.viewIndices != nil {
		return t.State.viewIndices
	}
	count := t.State.RowCount()
	indices := make([]int, count)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

func (t Table[T]) viewIndexForSource(sourceIdx int) (int, bool) {
	if t.State == nil {
		return 0, false
	}
	if t.State.viewIndices == nil {
		if sourceIdx >= 0 && sourceIdx < t.State.RowCount() {
			return sourceIdx, true
		}
		return 0, false
	}
	return t.State.viewIndexForSource(sourceIdx)
}

// CursorRow returns the row at the current cursor position.
// Returns the zero value of T if the table is empty or state is nil.
func (t Table[T]) CursorRow() T {
	var zero T
	if t.State == nil || t.State.RowCount() == 0 {
		return zero
	}
	if row, ok := t.State.SelectedRow(); ok {
		return row
	}
	return zero
}

// BuildLayoutNode builds a layout node for this table widget.
// Implements the LayoutNodeBuilder interface.
func (c tableContainer[T]) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	children := make([]layout.LayoutNode, len(c.children))
	for i, child := range c.children {
		childCtx := ctx.PushChild(i)
		built := child.Build(childCtx)

		var childNode layout.LayoutNode
		if builder, ok := built.(LayoutNodeBuilder); ok {
			childNode = builder.BuildLayoutNode(childCtx)
		} else {
			childNode = buildFallbackLayoutNode(built, childCtx)
		}

		children[i] = childNode
	}
	return c.BuildContainerLayoutNode(ctx, children)
}

func (c tableContainer[T]) BuildContainerLayoutNode(ctx BuildContext, children []layout.LayoutNode) layout.LayoutNode {
	padding := toLayoutEdgeInsets(c.Style.Padding)
	border := borderToEdgeInsets(c.Style.Border)
	dims := GetWidgetDimensionSet(c)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)
	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	columnWidths := make([]Dimension, len(c.Columns))
	overrides := c.State.ColumnWidths.Get()
	for i, col := range c.Columns {
		columnWidths[i] = col.Width
		if width, ok := overrides[col.ID]; ok && c.validColumnID(i) {
			columnWidths[i] = Cells(c.clampColumnWidth(i, width))
		}
	}

	node := layout.LayoutNode(&tableNode{
		Columns:        c.columnCount,
		Rows:           c.rowCount + c.headerRows,
		ColumnWidths:   columnWidths,
		ColumnSpacing:  c.ColumnSpacing,
		RowSpacing:     c.RowSpacing,
		Children:       children,
		Padding:        padding,
		Border:         border,
		Margin:         toLayoutEdgeInsets(c.Style.Margin),
		MinWidth:       minWidth,
		MaxWidth:       maxWidth,
		MinHeight:      minHeight,
		MaxHeight:      maxHeight,
		ExpandWidth:    dims.Width.IsFlex(),
		ExpandHeight:   dims.Height.IsFlex(),
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
		Viewport:       c.ownsViewport(),
		Header:         c.headerRows > 0,
		FrozenRows: func() int {
			if c.FrozenHeader {
				return c.headerRows
			}
			return 0
		}(),
		FrozenColumns: c.FrozenColumns,
		OffsetX: func() int {
			if c.ownsViewport() {
				return c.viewportState().OffsetX.Get()
			}
			return 0
		}(),
		OffsetY: func() int {
			if c.ownsViewport() {
				return c.viewportState().Offset.Get()
			}
			return 0
		}(),
		ViewportMetrics: &c.State.viewportMetrics,
	})

	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{
			child:     node,
			minWidth:  dims.MinWidth,
			maxWidth:  dims.MaxWidth,
			minHeight: dims.MinHeight,
			maxHeight: dims.MaxHeight,
			padding:   padding,
			border:    border,
		}
	}

	return node
}

func tableHasColumnHeaders(cols []TableColumn) bool {
	for _, col := range cols {
		if col.Header != nil {
			return true
		}
	}
	return false
}

func (t Table[T]) hasHeader() bool {
	return t.RenderHeader != nil || tableHasColumnHeaders(t.Columns)
}

func (t Table[T]) selectionMode() TableSelectionMode {
	var mode TableSelectionMode
	switch t.SelectionMode {
	case TableSelectionRow, TableSelectionColumn:
		mode = t.SelectionMode
	default:
		mode = TableSelectionCursor
	}
	if t.State != nil {
		t.State.syncSelectionMode(mode, len(t.Columns))
	}
	return mode
}

func tableCellActive(mode TableSelectionMode, rowIdx, colIdx, cursorRow, cursorCol int) bool {
	switch mode {
	case TableSelectionColumn:
		return colIdx == cursorCol
	case TableSelectionCursor:
		return rowIdx == cursorRow && colIdx == cursorCol
	default:
		return rowIdx == cursorRow
	}
}

func tableCellSelected(mode TableSelectionMode, selection map[int]struct{}, rowIdx, colIdx, columnCount int) bool {
	if len(selection) == 0 {
		return false
	}
	switch mode {
	case TableSelectionColumn:
		_, ok := selection[colIdx]
		return ok
	case TableSelectionCursor:
		_, ok := selection[cellIndex(rowIdx, colIdx, columnCount)]
		return ok
	default:
		_, ok := selection[rowIdx]
		return ok
	}
}

func cellIndex(rowIdx, colIdx, columnCount int) int {
	return rowIdx*columnCount + colIdx
}

func tableDefaultCellContent[T any](row T, colIndex int) (string, bool) {
	value := reflect.ValueOf(row)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return "", false
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		if colIndex < 0 || colIndex >= value.Len() {
			return "", true
		}
		return fmt.Sprintf("%v", value.Index(colIndex).Interface()), true
	default:
		return "", false
	}
}

func tableDefaultCellStyle(theme ThemeData, active, selected, widgetFocused bool) Style {
	style := Style{ForegroundColor: theme.Text}

	// Only show cursor highlight when widget has focus
	showCursor := active && widgetFocused

	if showCursor {
		style.BackgroundColor = theme.ActiveCursor
		style.ForegroundColor = theme.SelectionText
	} else if selected {
		// ActiveCursor highlight shown regardless of focus (user's selection persists)
		// Uses Selection for a dimmer appearance than the active cursor
		style.BackgroundColor = theme.Selection
	}
	return style
}

func cellIndexToRowCol(index, columnCount int) (row, col int) {
	if columnCount <= 0 {
		return 0, 0
	}
	if index < 0 {
		index = 0
	}
	return index / columnCount, index % columnCount
}

func (t Table[T]) setSelectionRange(from, to, count int) {
	if t.State == nil || count <= 0 {
		return
	}
	if from > to {
		from, to = to, from
	}
	if from < 0 {
		from = 0
	}
	if to >= count {
		to = count - 1
	}
	sel := make(map[int]struct{}, to-from+1)
	for i := from; i <= to; i++ {
		sel[i] = struct{}{}
	}
	t.State.Selection.Set(sel)
}

func (t Table[T]) setSelectionRangeFromView(viewIndices []int, anchorSource, cursorSource int) {
	if t.State == nil || len(viewIndices) == 0 {
		return
	}

	anchorView, ok := t.viewIndexForSource(anchorSource)
	if !ok {
		anchorView = 0
	}
	cursorView, ok := t.viewIndexForSource(cursorSource)
	if !ok {
		cursorView = anchorView
	}

	if anchorView > cursorView {
		anchorView, cursorView = cursorView, anchorView
	}

	sel := make(map[int]struct{}, cursorView-anchorView+1)
	for i := anchorView; i <= cursorView; i++ {
		sel[viewIndices[i]] = struct{}{}
	}
	t.State.Selection.Set(sel)
}

func (t Table[T]) setSelectionBox(viewIndices []int, anchorRow, anchorCol, rowIdx, colIdx, columnCount int) {
	if t.State == nil || len(viewIndices) == 0 || columnCount <= 0 {
		return
	}
	rowCount := len(viewIndices)
	anchorRow = clampInt(anchorRow, 0, rowCount-1)
	anchorCol = clampInt(anchorCol, 0, columnCount-1)
	rowIdx = clampInt(rowIdx, 0, rowCount-1)
	colIdx = clampInt(colIdx, 0, columnCount-1)

	minRow, maxRow := anchorRow, rowIdx
	if minRow > maxRow {
		minRow, maxRow = maxRow, minRow
	}
	minCol, maxCol := anchorCol, colIdx
	if minCol > maxCol {
		minCol, maxCol = maxCol, minCol
	}

	sel := make(map[int]struct{}, (maxRow-minRow+1)*(maxCol-minCol+1))
	for viewRow := minRow; viewRow <= maxRow; viewRow++ {
		sourceRow := viewIndices[viewRow]
		for col := minCol; col <= maxCol; col++ {
			sel[cellIndex(sourceRow, col, columnCount)] = struct{}{}
		}
	}
	t.State.Selection.Set(sel)
}

// Jump moves the cursor to this cell's row (and column, in cursor selection
// mode), for jump mode (see Jumpable).
func (w defaultTableCellWidget[T]) Jump()          { w.table.jumpTo(w.sourceRow, w.colIndex) }
func (w defaultTableCellWidget[T]) jumpable() bool { return w.table.cellJumpable(w.colIndex) }

// Jump moves the cursor to this cell's row (and column, in cursor selection
// mode), for jump mode (see Jumpable).
func (c tableCell[T]) Jump()          { c.table.jumpTo(c.sourceRow, c.col) }
func (c tableCell[T]) jumpable() bool { return c.table.cellJumpable(c.col) }

// cellJumpable reports whether jump mode labels the cell in column col: every
// cell when the cursor is a cell, the first cell of each row when it is a
// row, and none when it is a column.
func (t Table[T]) cellJumpable(col int) bool {
	switch t.selectionMode() {
	case TableSelectionCursor:
		return true
	case TableSelectionRow:
		return col == 0
	}
	return false
}

// jumpTo moves the cursor to a cell for jump mode.
func (t Table[T]) jumpTo(sourceRow, col int) {
	if t.State == nil {
		return
	}
	previous := t.State.CursorIndex.Peek()
	t.State.SelectIndex(sourceRow)
	if t.selectionMode() == TableSelectionCursor {
		t.State.SelectColumn(col)
	}
	t.scrollCursorIntoView()
	if t.State.CursorIndex.Peek() != previous {
		t.notifyCursorChange()
	}
}
