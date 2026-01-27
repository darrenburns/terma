package terma

import (
	"fmt"
	"reflect"

	"terma/layout"
)

// TableState holds the state for a Table widget.
// It is the source of truth for rows and cursor position, and must be provided to Table.
// Rows is a reactive Signal - changes trigger automatic re-renders.
type TableState[T any] struct {
	Rows         AnySignal[[]T]              // Reactive table rows
	CursorIndex  Signal[int]                 // Cursor position (row index)
	CursorColumn Signal[int]                 // Cursor position (column index)
	Selection    AnySignal[map[int]struct{}] // Selected indices (row/column/cell based on selection mode)

	anchorIndex *int // Anchor point for shift-selection (nil = no anchor)

	lastSelectionMode TableSelectionMode
	hasSelectionMode  bool

	rowLayouts        []tableRowLayout // Cached layout metrics (per row)
	viewIndices       []int            // View index -> source index for filtered views
	viewIndexBySource map[int]int      // Source index -> view index for filtered views
}

// NewTableState creates a new TableState with the given initial rows.
func NewTableState[T any](initialRows []T) *TableState[T] {
	if initialRows == nil {
		initialRows = []T{}
	}
	return &TableState[T]{
		Rows:         NewAnySignal(initialRows),
		CursorIndex:  NewSignal(0),
		CursorColumn: NewSignal(0),
		Selection:    NewAnySignal(make(map[int]struct{})),
	}
}

// SetRows replaces all rows and clamps cursor to valid range.
func (s *TableState[T]) SetRows(rows []T) {
	if rows == nil {
		rows = []T{}
	}
	s.Rows.Set(rows)
	s.clampCursor()
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
func (s *TableState[T]) Prepend(row T) {
	s.Rows.Update(func(rows []T) []T {
		return append([]T{row}, rows...)
	})
	// Adjust cursor to keep same row selected
	s.CursorIndex.Update(func(i int) int {
		return i + 1
	})
}

// InsertAt inserts a row at the specified index.
// If index is out of bounds, it's clamped to valid range.
func (s *TableState[T]) InsertAt(index int, row T) {
	s.Rows.Update(func(rows []T) []T {
		if index < 0 {
			index = 0
		}
		if index > len(rows) {
			index = len(rows)
		}
		rows = append(rows, row)
		copy(rows[index+1:], rows[index:])
		rows[index] = row
		return rows
	})
	// Adjust cursor if insertion was at or before cursor
	cursorIdx := s.CursorIndex.Peek()
	if index <= cursorIdx {
		s.CursorIndex.Set(cursorIdx + 1)
	}
}

// RemoveAt removes the row at the specified index.
// Returns true if a row was removed, false if index was out of bounds.
func (s *TableState[T]) RemoveAt(index int) bool {
	rows := s.Rows.Peek()
	if index < 0 || index >= len(rows) {
		return false
	}
	s.Rows.Update(func(rows []T) []T {
		return append(rows[:index], rows[index+1:]...)
	})
	s.clampCursor()
	return true
}

// RemoveWhere removes all rows matching the predicate.
// Returns the number of rows removed.
func (s *TableState[T]) RemoveWhere(predicate func(T) bool) int {
	removed := 0
	s.Rows.Update(func(rows []T) []T {
		result := make([]T, 0, len(rows))
		for _, row := range rows {
			if !predicate(row) {
				result = append(result, row)
			} else {
				removed++
			}
		}
		return result
	})
	s.clampCursor()
	return removed
}

// Clear removes all rows from the table.
func (s *TableState[T]) Clear() {
	s.Rows.Set([]T{})
	s.CursorIndex.Set(0)
	s.CursorColumn.Set(0)
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

func (s *TableState[T]) syncSelectionMode(mode TableSelectionMode) {
	if !s.hasSelectionMode {
		s.lastSelectionMode = mode
		s.hasSelectionMode = true
		return
	}
	if s.lastSelectionMode == mode {
		return
	}
	s.ClearSelection()
	s.ClearAnchor()
	s.lastSelectionMode = mode
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
	Width  Dimension // Optional width (Cells, Percent, Flex, Auto)
	Header Widget    // Optional header widget for this column
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
	ID                  string                                                                                        // Optional unique identifier
	CursorStyle                                                                                                       // Embedded - CursorPrefix/SelectedPrefix fields for customizable indicators
	State               *TableState[T]                                                                                // Required - holds rows and cursor position
	Columns             []TableColumn                                                                                 // Required - defines column count and widths
	RenderCell          func(row T, rowIndex int, colIndex int, active bool, selected bool) Widget                    // Cell renderer (default uses fmt)
	RenderCellWithMatch func(row T, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget // Optional cell renderer with match data
	Filter              *FilterState                                                                                  // Optional filter state for matching rows
	MatchCell           func(row T, rowIndex int, colIndex int, query string, options FilterOptions) MatchResult      // Optional matcher per cell
	RenderHeader        func(colIndex int) Widget                                                                     // Optional header renderer (takes precedence over column headers)
	OnSelect            func(row T)                                                                                   // Callback invoked when Enter is pressed on a row
	OnCursorChange      func(row T)                                                                                   // Callback invoked when cursor moves to a different row
	ScrollState         *ScrollState                                                                                  // Optional state for scroll-into-view
	RowHeight           int                                                                                           // Optional uniform row height override (default 0 = layout metrics / fallback 1)
	ColumnSpacing       int                                                                                           // Space between columns
	RowSpacing          int                                                                                           // Space between rows
	SelectionMode       TableSelectionMode                                                                            // Cursor/selection highlight mode (row/column/cursor)
	MultiSelect         bool                                                                                          // Enable multi-select mode (shift+move to extend)
	Width               Dimension                                                                                     // Deprecated: use Style.Width
	Height              Dimension                                                                                     // Deprecated: use Style.Height
	Style               Style                                                                                         // Optional styling
	Click               func(MouseEvent)                                                                              // Optional callback invoked when clicked
	MouseDown           func(MouseEvent)                                                                              // Optional callback invoked when mouse is pressed
	MouseUp             func(MouseEvent)                                                                              // Optional callback invoked when mouse is released
	Hover               func(bool)                                                                                    // Optional callback invoked when hover state changes
}

type tableRowLayout struct {
	y      int
	height int
}

type tableContainer[T any] struct {
	Table[T]
	children    []Widget
	rowCount    int
	columnCount int
	headerRows  int
}

func (c tableContainer[T]) Build(ctx BuildContext) Widget {
	return c
}

func (c tableContainer[T]) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.State == nil || c.columnCount == 0 || c.rowCount == 0 {
		if c.State != nil {
			c.State.rowLayouts = nil
		}
		return
	}

	count := metrics.ChildCount()
	if count == 0 {
		c.State.rowLayouts = nil
		return
	}

	rowLayouts := make([]tableRowLayout, c.rowCount)
	seen := make([]bool, c.rowCount)

	for i := 0; i < count; i++ {
		bounds, ok := metrics.ChildBounds(i)
		if !ok {
			continue
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
	if c.selectionMode() != TableSelectionColumn {
		c.scrollCursorIntoView()
	}
}

func (c tableContainer[T]) ChildWidgets() []Widget {
	return c.children
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

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (t Table[T]) OnMouseDown(event MouseEvent) {
	if t.MouseDown != nil {
		t.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (t Table[T]) OnMouseUp(event MouseEvent) {
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (t Table[T]) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (t Table[T]) IsFocusable() bool {
	return true
}

// Build returns a table container that arranges the rendered cells.
func (t Table[T]) Build(ctx BuildContext) Widget {
	if t.State == nil || len(t.Columns) == 0 {
		return Column{}
	}

	renderCell := t.RenderCell
	renderCellWithMatch := t.RenderCellWithMatch
	if renderCellWithMatch == nil && renderCell == nil {
		renderCellWithMatch = t.themedDefaultRenderCell(ctx)
	}

	rows := t.State.Rows.Get()
	columnCount := len(t.Columns)
	mode := t.selectionMode()
	query, options := filterStateValues(t.Filter)
	viewRows, viewIndices, viewMatches := t.filteredRows(rows, columnCount, query, options)
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
			headerCells[colIdx] = header
		}
	}

	if len(viewRows) == 0 && headerRows == 0 {
		t.State.rowLayouts = nil
		return Column{}
	}

	children := make([]Widget, 0, (len(viewRows)+headerRows)*columnCount)
	if headerRows > 0 {
		children = append(children, headerCells...)
	}

	cursorRow := 0
	cursorCol := 0
	selection := map[int]struct{}{}
	if len(viewRows) > 0 {
		cursorRow = t.State.CursorIndex.Get()
		cursorCol = t.State.CursorColumn.Get()
		if t.MultiSelect {
			selection = t.State.Selection.Get()
		}

		if len(rows) > 0 {
			clampedRow := clampInt(cursorRow, 0, len(rows)-1)
			if clampedRow != cursorRow {
				t.State.CursorIndex.Set(clampedRow)
				cursorRow = clampedRow
			}
		}

		clampedCol := clampInt(cursorCol, 0, columnCount-1)
		if clampedCol != cursorCol {
			t.State.CursorColumn.Set(clampedCol)
			cursorCol = clampedCol
		}

		if _, ok := t.State.viewIndexForSource(cursorRow); !ok {
			cursorRow = viewIndices[0]
			t.State.CursorIndex.Set(cursorRow)
		}

		t.registerScrollCallbacks(mode, hasHeader)
	}

	for viewRowIdx, row := range viewRows {
		sourceRowIdx := viewIndices[viewRowIdx]
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			active := tableCellActive(mode, sourceRowIdx, colIdx, cursorRow, cursorCol)
			selected := false
			if t.MultiSelect {
				selected = tableCellSelected(mode, selection, sourceRowIdx, colIdx, columnCount)
			}
			match := MatchResult{}
			if len(viewMatches) > 0 {
				match = viewMatches[viewRowIdx][colIdx]
			}
			var cell Widget
			if renderCellWithMatch != nil {
				cell = renderCellWithMatch(row, sourceRowIdx, colIdx, active, selected, match)
			} else {
				cell = renderCell(row, sourceRowIdx, colIdx, active, selected)
			}
			if cell == nil {
				cell = Text{}
			}
			children = append(children, cell)
		}
	}

	return tableContainer[T]{
		Table:       t,
		children:    children,
		rowCount:    len(viewRows),
		columnCount: columnCount,
		headerRows:  headerRows,
	}
}

// themedDefaultRenderCell returns a themed render function for table cells.
// Captures theme colors and widget focus state from the context for use in the render function.
// Cursor highlighting is only shown when the widget has focus.
func (t Table[T]) themedDefaultRenderCell(ctx BuildContext) func(row T, rowIndex int, colIndex int, active bool, selected bool, match MatchResult) Widget {
	theme := ctx.Theme()
	widgetFocused := ctx.IsFocused(t)
	cursorPrefix := t.CursorPrefix
	selectedPrefix := t.SelectedPrefix

	highlight := SpanStyle{
		Underline:      UnderlineSingle,
		UnderlineColor: theme.Accent,
		Background:     theme.Selection,
	}
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

	for rowIdx, row := range rows {
		cellMatches := make([]MatchResult, columnCount)
		rowMatched := false
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			match := matchCell(row, rowIdx, colIdx, query, options)
			cellMatches[colIdx] = match
			if match.Matched {
				rowMatched = true
			}
		}
		if rowMatched {
			viewRows = append(viewRows, row)
			viewIndices = append(viewIndices, rowIdx)
			viewMatches = append(viewMatches, cellMatches)
		}
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

	// Left/right only in Cursor mode (not Row, not Column)
	if mode == TableSelectionCursor {
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
	if t.OnSelect != nil {
		if row, ok := t.State.SelectedRow(); ok {
			t.OnSelect(row)
		}
	}
}

func (t Table[T]) keyCursorUp() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(-1)
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorRow := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
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
		t.scrollBy(1)
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorRow := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
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
		if t.ScrollState != nil {
			t.ScrollState.SetOffset(0)
		}
		return
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(0)
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorToLast() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		if t.ScrollState != nil {
			t.ScrollState.SetOffset(maxTableInt())
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
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) pageUp() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(-10)
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorRow := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
	if !ok {
		cursorViewIdx = 0
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx - 10)
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) pageDown() {
	mode := t.selectionMode()
	if mode == TableSelectionColumn {
		t.scrollBy(10)
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorRow := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorRow)
	if !ok {
		cursorViewIdx = 0
	}
	if t.MultiSelect {
		t.State.ClearSelection()
		t.State.ClearAnchor()
	}
	t.setCursorToViewIndex(cursorViewIdx + 10)
	t.scrollCursorIntoView()
	t.notifyCursorChange()
}

func (t Table[T]) keyCursorLeft() {
	columnCount := len(t.Columns)
	if columnCount == 0 {
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
}

func (t Table[T]) keyCursorRight() {
	columnCount := len(t.Columns)
	if columnCount == 0 {
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
	t.scrollCursorIntoView()
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
	t.scrollCursorIntoView()
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
	t.scrollCursorIntoView()
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
	t.scrollCursorIntoView()
}

func (t Table[T]) scrollBy(lines int) bool {
	if t.ScrollState == nil {
		return false
	}
	if lines < 0 {
		return t.ScrollState.ScrollUp(-lines)
	}
	return t.ScrollState.ScrollDown(lines)
}

// scrollCursorIntoView uses the ScrollState to ensure
// the cursor row is visible in the viewport.
func (t Table[T]) scrollCursorIntoView() {
	if t.ScrollState == nil || t.State == nil {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	viewIdx, ok := t.viewIndexForSource(cursorIdx)
	if !ok {
		return
	}
	rowY, rowHeight, ok := t.getRowLayout(cursorIdx)
	if !ok {
		rowHeight = t.getRowHeight()
		rowY = viewIdx * rowHeight
	}
	t.ScrollState.ScrollToView(rowY, rowHeight)
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

// registerScrollCallbacks sets up callbacks on the ScrollState
// to update cursor position when mouse wheel scrolling occurs.
// The callbacks move cursor first, then scroll only if needed.
func (t Table[T]) registerScrollCallbacks(mode TableSelectionMode, hasHeader bool) {
	if t.ScrollState == nil {
		return
	}

	if mode == TableSelectionColumn {
		t.ScrollState.OnScrollUp = nil
		t.ScrollState.OnScrollDown = nil
		return
	}

	t.ScrollState.OnScrollUp = func(lines int) bool {
		if hasHeader && t.State != nil {
			if viewIdx, ok := t.viewIndexForSource(t.State.CursorIndex.Peek()); ok && viewIdx == 0 {
				return false
			}
		}
		t.moveCursorUp(lines)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true
	}
	t.ScrollState.OnScrollDown = func(lines int) bool {
		t.moveCursorDown(lines)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true
	}
}

// moveCursorUp moves the cursor up by the given number of rows.
func (t Table[T]) moveCursorUp(count int) {
	if t.State == nil {
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	newCursor := clampInt(cursorViewIdx-count, 0, len(view)-1)
	t.State.SelectIndex(view[newCursor])
}

// moveCursorDown moves the cursor down by the given number of rows.
func (t Table[T]) moveCursorDown(count int) {
	if t.State == nil {
		return
	}
	view := t.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	cursorViewIdx, ok := t.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	newCursor := clampInt(cursorViewIdx+count, 0, len(view)-1)
	t.State.SelectIndex(view[newCursor])
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

	padding := toLayoutEdgeInsets(c.Style.Padding)
	border := borderToEdgeInsets(c.Style.Border)
	dims := GetWidgetDimensionSet(c)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	columnWidths := make([]Dimension, len(c.Columns))
	for i, col := range c.Columns {
		columnWidths[i] = col.Width
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
	mode := TableSelectionCursor
	switch t.SelectionMode {
	case TableSelectionRow, TableSelectionColumn:
		mode = t.SelectionMode
	default:
		mode = TableSelectionCursor
	}
	if t.State != nil {
		t.State.syncSelectionMode(mode)
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
