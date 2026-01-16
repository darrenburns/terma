package terma

import (
	"fmt"

	"terma/layout"
)

// TableState holds the state for a Table widget.
// It is the source of truth for rows and cursor position, and must be provided to Table.
// Rows is a reactive Signal - changes trigger automatic re-renders.
type TableState[T any] struct {
	Rows        AnySignal[[]T]              // Reactive table rows
	CursorIndex Signal[int]                 // Cursor position (row index)
	Selection   AnySignal[map[int]struct{}] // Selected row indices (for multi-select)

	anchorIndex *int // Anchor point for shift-selection (nil = no anchor)

	rowLayouts []tableRowLayout // Cached layout metrics (per row)
}

// NewTableState creates a new TableState with the given initial rows.
func NewTableState[T any](initialRows []T) *TableState[T] {
	if initialRows == nil {
		initialRows = []T{}
	}
	return &TableState[T]{
		Rows:        NewAnySignal(initialRows),
		CursorIndex: NewSignal(0),
		Selection:   NewAnySignal(make(map[int]struct{})),
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

// Table is a generic focusable widget that displays a navigable table of rows.
// Use with Scrollable and a shared ScrollState to enable scroll-into-view.
type Table[T any] struct {
	ID             string                                                                     // Optional unique identifier
	State          *TableState[T]                                                             // Required - holds rows and cursor position
	Columns        []TableColumn                                                              // Required - defines column count and widths
	RenderCell     func(row T, rowIndex int, colIndex int, active bool, selected bool) Widget // Cell renderer (default uses fmt)
	RenderHeader   func(colIndex int) Widget                                                  // Optional header renderer (takes precedence over column headers)
	OnSelect       func(row T)                                                                // Callback invoked when Enter is pressed on a row
	OnCursorChange func(row T)                                                                // Callback invoked when cursor moves to a different row
	ScrollState    *ScrollState                                                               // Optional state for scroll-into-view
	RowHeight      int                                                                        // Optional uniform row height override (default 0 = layout metrics / fallback 1)
	ColumnSpacing  int                                                                        // Space between columns
	RowSpacing     int                                                                        // Space between rows
	MultiSelect    bool                                                                       // Enable multi-select mode (shift+move to extend)
	Width          Dimension                                                                  // Optional width (zero value = auto)
	Height         Dimension                                                                  // Optional height (zero value = auto)
	Style          Style                                                                      // Optional styling
	Click          func()                                                                     // Optional callback invoked when clicked
	Hover          func(bool)                                                                 // Optional callback invoked when hover state changes
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
	c.scrollCursorIntoView()
}

func (c tableContainer[T]) ChildWidgets() []Widget {
	return c.children
}

// WidgetID returns the table's unique identifier.
// Implements the Identifiable interface.
func (t Table[T]) WidgetID() string {
	return t.ID
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (t Table[T]) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the style of the table widget.
// Implements the Styled interface.
func (t Table[T]) GetStyle() Style {
	return t.Style
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (t Table[T]) OnClick() {
	if t.Click != nil {
		t.Click()
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
	if renderCell == nil {
		renderCell = t.themedDefaultRenderCell(ctx)
	}

	rows := t.State.Rows.Get()
	columnCount := len(t.Columns)

	headerRows := 0
	var headerCells []Widget
	if t.RenderHeader != nil || tableHasColumnHeaders(t.Columns) {
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

	if len(rows) == 0 && headerRows == 0 {
		t.State.rowLayouts = nil
		return Column{}
	}

	children := make([]Widget, 0, (len(rows)+headerRows)*columnCount)
	if headerRows > 0 {
		children = append(children, headerCells...)
	}

	cursorIdx := 0
	selection := map[int]struct{}{}
	if len(rows) > 0 {
		cursorIdx = t.State.CursorIndex.Get()
		if t.MultiSelect {
			selection = t.State.Selection.Get()
		}

		clamped := clampInt(cursorIdx, 0, len(rows)-1)
		if clamped != cursorIdx {
			t.State.CursorIndex.Set(clamped)
			cursorIdx = clamped
		}

		t.registerScrollCallbacks()
	}

	for rowIdx, row := range rows {
		active := rowIdx == cursorIdx
		selected := false
		if t.MultiSelect {
			_, selected = selection[rowIdx]
		}
		for colIdx := 0; colIdx < columnCount; colIdx++ {
			cell := renderCell(row, rowIdx, colIdx, active, selected)
			if cell == nil {
				cell = Text{}
			}
			children = append(children, cell)
		}
	}

	return tableContainer[T]{
		Table:       t,
		children:    children,
		rowCount:    len(rows),
		columnCount: columnCount,
		headerRows:  headerRows,
	}
}

// themedDefaultRenderCell returns a themed render function for table cells.
// Captures theme colors from the context for use in the render function.
func (t Table[T]) themedDefaultRenderCell(ctx BuildContext) func(row T, rowIndex int, colIndex int, active bool, selected bool) Widget {
	theme := ctx.Theme()
	return func(row T, rowIndex int, colIndex int, active bool, selected bool) Widget {
		if colIndex != 0 {
			return Text{Content: "", Style: Style{ForegroundColor: theme.Text}}
		}

		content := fmt.Sprintf("%v", row)
		prefix := "  "
		style := Style{ForegroundColor: theme.Text}

		if selected && active {
			prefix = "▶*"
			style.ForegroundColor = theme.Accent
		} else if active {
			prefix = "▶ "
			style.ForegroundColor = theme.Accent
		} else if selected {
			prefix = " *"
		}

		return Text{
			Content: prefix + content,
			Style:   style,
		}
	}
}

// OnKey handles navigation keys and selection, updating cursor position and scrolling into view.
// Implements the Focusable interface.
func (t Table[T]) OnKey(event KeyEvent) bool {
	if t.State == nil || t.State.RowCount() == 0 {
		return false
	}

	cursorIdx := t.State.CursorIndex.Peek()
	rowCount := t.State.RowCount()

	// Handle multi-select specific keys (shift+movement to extend selection)
	if t.MultiSelect {
		switch {
		case event.MatchString("shift+up", "shift+k"):
			t.handleShiftMove(-1)
			return true

		case event.MatchString("shift+down", "shift+j"):
			t.handleShiftMove(1)
			return true

		case event.MatchString("shift+home"):
			t.handleShiftMoveTo(0)
			return true

		case event.MatchString("shift+end"):
			t.handleShiftMoveTo(rowCount - 1)
			return true
		}
	}

	switch {
	case event.MatchString("enter"):
		if t.OnSelect != nil {
			if row, ok := t.State.SelectedRow(); ok {
				t.OnSelect(row)
			}
		}
		return true

	case event.MatchString("up", "k"):
		if cursorIdx == 0 {
			return false
		}
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		t.State.SelectPrevious()
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("down", "j"):
		if cursorIdx >= rowCount-1 {
			return false
		}
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		t.State.SelectNext()
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("home", "g"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		t.State.SelectFirst()
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("end", "G"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		t.State.SelectLast()
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		t.State.SelectIndex(newCursor)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.ClearAnchor()
		}
		newCursor := cursorIdx + 10
		if newCursor >= rowCount {
			newCursor = rowCount - 1
		}
		t.State.SelectIndex(newCursor)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true
	}

	return false
}

// handleShiftMove extends selection by moving cursor by delta and selecting the range.
func (t Table[T]) handleShiftMove(delta int) {
	cursorIdx := t.State.CursorIndex.Peek()

	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorIdx)
	}

	if delta > 0 {
		t.State.SelectNext()
	} else {
		t.State.SelectPrevious()
	}

	newCursor := t.State.CursorIndex.Peek()
	t.State.SelectRange(t.State.GetAnchor(), newCursor)

	t.scrollCursorIntoView()
}

// handleShiftMoveTo extends selection to a specific index.
func (t Table[T]) handleShiftMoveTo(targetIdx int) {
	cursorIdx := t.State.CursorIndex.Peek()

	if !t.State.HasAnchor() {
		t.State.SetAnchor(cursorIdx)
	}

	t.State.SelectIndex(targetIdx)

	newCursor := t.State.CursorIndex.Peek()
	t.State.SelectRange(t.State.GetAnchor(), newCursor)

	t.scrollCursorIntoView()
}

// scrollCursorIntoView uses the ScrollState to ensure
// the cursor row is visible in the viewport.
func (t Table[T]) scrollCursorIntoView() {
	if t.ScrollState == nil || t.State == nil {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	rowY, rowHeight, ok := t.getRowLayout(cursorIdx)
	if !ok {
		rowHeight = t.getRowHeight()
		rowY = cursorIdx * rowHeight
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
	if index < 0 || index >= len(t.State.rowLayouts) {
		return 0, 0, false
	}
	layout := t.State.rowLayouts[index]
	if layout.height <= 0 {
		return 0, 0, false
	}
	return layout.y, layout.height, true
}

// registerScrollCallbacks sets up callbacks on the ScrollState
// to update cursor position when mouse wheel scrolling occurs.
// The callbacks move cursor first, then scroll only if needed.
func (t Table[T]) registerScrollCallbacks() {
	if t.ScrollState == nil {
		return
	}

	t.ScrollState.OnScrollUp = func(lines int) bool {
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
	if t.State == nil || t.State.RowCount() == 0 {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	newCursor := cursorIdx - count
	if newCursor < 0 {
		newCursor = 0
	}
	t.State.SelectIndex(newCursor)
}

// moveCursorDown moves the cursor down by the given number of rows.
func (t Table[T]) moveCursorDown(count int) {
	if t.State == nil || t.State.RowCount() == 0 {
		return
	}
	cursorIdx := t.State.CursorIndex.Peek()
	rowCount := t.State.RowCount()
	newCursor := cursorIdx + count
	if newCursor >= rowCount {
		newCursor = rowCount - 1
	}
	t.State.SelectIndex(newCursor)
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
		built := child.Build(ctx.PushChild(i))

		var childNode layout.LayoutNode
		if builder, ok := built.(LayoutNodeBuilder); ok {
			childNode = builder.BuildLayoutNode(ctx.PushChild(i))
		} else {
			childNode = buildFallbackLayoutNode(built, ctx.PushChild(i))
		}

		children[i] = childNode
	}

	minWidth, maxWidth := dimensionToMinMax(c.Width)
	minHeight, maxHeight := dimensionToMinMax(c.Height)

	preserveWidth := c.Width.IsAuto() && !c.Width.IsUnset()
	preserveHeight := c.Height.IsAuto() && !c.Height.IsUnset()

	columnWidths := make([]Dimension, len(c.Columns))
	for i, col := range c.Columns {
		columnWidths[i] = col.Width
	}

	return &tableNode{
		Columns:        c.columnCount,
		Rows:           c.rowCount + c.headerRows,
		ColumnWidths:   columnWidths,
		ColumnSpacing:  c.ColumnSpacing,
		RowSpacing:     c.RowSpacing,
		Children:       children,
		Padding:        toLayoutEdgeInsets(c.Style.Padding),
		Border:         borderToEdgeInsets(c.Style.Border),
		Margin:         toLayoutEdgeInsets(c.Style.Margin),
		MinWidth:       minWidth,
		MaxWidth:       maxWidth,
		MinHeight:      minHeight,
		MaxHeight:      maxHeight,
		ExpandWidth:    c.Width.IsFlex(),
		ExpandHeight:   c.Height.IsFlex(),
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
	}
}

func tableHasColumnHeaders(cols []TableColumn) bool {
	for _, col := range cols {
		if col.Header != nil {
			return true
		}
	}
	return false
}
