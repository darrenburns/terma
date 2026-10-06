package terma

import (
	uv "github.com/charmbracelet/ultraviolet"
	"sort"
)

// TableSortDirection specifies the ordering of a table view.
type TableSortDirection int

const (
	TableSortNone TableSortDirection = iota
	TableSortAscending
	TableSortDescending
)

// TableSort identifies a column and direction. None restores source/filter order.
type TableSort struct {
	ColumnID  string
	Direction TableSortDirection
}

// NewTableStateWithRowID preserves unique nonempty row identities across SetRows.
// The extractor must be pure and return stable IDs. Duplicate/empty identities
// cannot be preserved: their selection is dropped and their cursor is clamped.
func NewTableStateWithRowID[T any](rows []T, rowID func(T) string) *TableState[T] {
	s := NewTableState(rows)
	s.rowID = rowID
	return s
}
func uniqueTableRowIDs[T any](rows []T, id func(T) string) map[string]int {
	result := make(map[string]int, len(rows))
	for i, row := range rows {
		key := id(row)
		if key == "" {
			continue
		}
		if _, exists := result[key]; exists {
			result[key] = -1
		} else {
			result[key] = i
		}
	}
	return result
}
func (s *TableState[T]) replaceIdentifiedRows(rows []T) {
	old := s.Rows.Peek()
	oldIDs := uniqueTableRowIDs(old, s.rowID)
	newIDs := uniqueTableRowIDs(rows, s.rowID)
	mapRow := func(i int) (int, bool) {
		if i < 0 || i >= len(old) {
			return 0, false
		}
		id := s.rowID(old[i])
		oi, ok := oldIDs[id]
		if !ok || oi < 0 {
			return 0, false
		}
		ni, ok := newIDs[id]
		return ni, ok && ni >= 0
	}
	cursor, preserved := mapRow(s.CursorIndex.Peek())
	s.remapSelectedRows(mapRow)
	s.Rows.Set(rows)
	s.setViewIndices(nil)
	if preserved {
		s.CursorIndex.Set(cursor)
	} else {
		s.CursorIndex.Set(max(0, min(s.CursorIndex.Peek(), len(rows)-1)))
	}
	s.dragging = false
}
func (t Table[T]) validColumnID(col int) bool {
	if col < 0 || col >= len(t.Columns) || t.Columns[col].ID == "" {
		return false
	}
	for i, c := range t.Columns {
		if i != col && c.ID == t.Columns[col].ID {
			return false
		}
	}
	return true
}
func (t Table[T]) sortableColumn(col int) bool {
	return t.validColumnID(col) && t.Comparators[t.Columns[col].ID] != nil
}
func (t Table[T]) cycleSort(col int) {
	if t.State == nil || !t.sortableColumn(col) {
		return
	}
	id := t.Columns[col].ID
	current := t.State.Sort.Peek()
	next := TableSort{ColumnID: id, Direction: TableSortAscending}
	if current.ColumnID == id {
		if current.Direction == TableSortAscending {
			next.Direction = TableSortDescending
		}
		if current.Direction == TableSortDescending {
			next = TableSort{}
		}
	}
	t.State.Sort.Set(next)
}
func (t Table[T]) sortedRows(rows []T, indices []int, matches [][]MatchResult, order TableSort) ([]T, []int, [][]MatchResult) {
	if order.Direction != TableSortAscending && order.Direction != TableSortDescending {
		return rows, indices, matches
	}
	valid := false
	for i, col := range t.Columns {
		if col.ID == order.ColumnID && t.sortableColumn(i) {
			valid = true
			break
		}
	}
	if !valid {
		return rows, indices, matches
	}
	compare := t.Comparators[order.ColumnID]
	permutation := make([]int, len(rows))
	for i := range permutation {
		permutation[i] = i
	}
	sort.SliceStable(permutation, func(i, j int) bool {
		a, b := permutation[i], permutation[j]
		result := compare(rows[a], rows[b])
		if result == 0 {
			return indices[a] < indices[b]
		}
		if order.Direction == TableSortDescending {
			return result > 0
		}
		return result < 0
	})
	sortedRows := make([]T, len(rows))
	sortedIndices := make([]int, len(rows))
	var sortedMatches [][]MatchResult
	if len(matches) > 0 {
		sortedMatches = make([][]MatchResult, len(matches))
	}
	for i, p := range permutation {
		sortedRows[i] = rows[p]
		sortedIndices[i] = indices[p]
		if sortedMatches != nil {
			sortedMatches[i] = matches[p]
		}
	}
	return sortedRows, sortedIndices, sortedMatches
}
func (t Table[T]) clampColumnWidth(col, width int) int {
	c := t.Columns[col]
	minimum := max(1, c.MinWidth)
	width = max(minimum, width)
	if c.MaxWidth > 0 {
		width = min(width, max(minimum, c.MaxWidth))
	}
	return width
}
func (t Table[T]) resizeColumn(col, width int) {
	if t.State == nil || !t.validColumnID(col) || !t.Columns[col].Resizable {
		return
	}
	width = t.clampColumnWidth(col, width)
	id := t.Columns[col].ID
	if old, ok := t.State.ColumnWidths.Peek()[id]; ok && old == width {
		return
	}
	t.State.ColumnWidths.Update(func(old map[string]int) map[string]int {
		next := make(map[string]int, len(old)+1)
		for k, v := range old {
			next[k] = v
		}
		next[id] = width
		return next
	})
}
func (t Table[T]) resizeCurrent(delta int) {
	col := t.State.CursorColumn.Peek()
	if col >= 0 && col < len(t.State.columnLayouts) {
		t.resizeColumn(col, t.State.columnLayouts[col].width+delta)
	}
}
func (t Table[T]) resetCurrentWidth() {
	col := t.State.CursorColumn.Peek()
	if !t.validColumnID(col) || !t.Columns[col].Resizable {
		return
	}
	id := t.Columns[col].ID
	t.State.ColumnWidths.Update(func(old map[string]int) map[string]int {
		next := make(map[string]int, len(old))
		for k, v := range old {
			if k != id {
				next[k] = v
			}
		}
		return next
	})
}
func (t Table[T]) headerMouseDown(event MouseEvent) bool {
	if event.Button != uv.MouseLeft || !t.hasHeader() {
		return false
	}
	x := event.LocalX - t.Style.Border.Width() - t.Style.Padding.Left
	y := event.LocalY - t.Style.Border.Width() - t.Style.Padding.Top
	if t.ownsViewport() {
		m := t.State.viewportMetrics
		if x < 0 || x >= m.width || y < 0 || y >= m.height {
			return false
		}
		if !t.FrozenHeader {
			y += t.viewportState().GetOffset()
		}
		if x >= m.frozenWidth {
			x += t.viewportState().GetOffsetX()
		}
	}
	if y < 0 || y >= t.State.headerHeight {
		return false
	}
	for col, bounds := range t.State.columnLayouts {
		if x < bounds.x || x >= bounds.x+bounds.width {
			continue
		}
		if !t.sortableColumn(col) && !(t.Columns[col].Resizable && t.validColumnID(col)) {
			return true
		}
		t.State.CursorColumn.Set(col)
		if t.Columns[col].Resizable && t.validColumnID(col) && x == bounds.x+bounds.width-1 {
			t.State.resizing = true
			t.State.resizeColumn = col
			t.State.resizeStartX = event.X
			t.State.resizeStartWidth = bounds.width
			return true
		}
		t.cycleSort(col)
		return true
	}
	return true
}
func (t Table[T]) ownsViewport() bool {
	return t.State != nil && (t.FrozenHeader || t.FrozenColumns > 0)
}
func (t Table[T]) viewportState() *ScrollState {
	if t.ScrollState != nil {
		return t.ScrollState
	}
	return t.State.viewport
}
func (t Table[T]) viewportCells(children []Widget) []Widget {
	if t.ownsViewport() {
		for i, child := range children {
			children[i] = passThrough{child: child}
		}
	}
	return children
}
func (t Table[T]) layoutViewport() {
	m := t.State.viewportMetrics
	scroll := t.viewportState()
	scroll.viewportWidth = m.width
	scroll.viewportHeight = m.height
	scroll.contentWidth = m.contentWidth
	scroll.contentHeight = m.contentHeight
	// Clamp to the new bounds, leaving a running glide alone.
	if offset := scroll.GetOffset(); offset > scroll.maxOffset() {
		scroll.SetOffset(offset)
	}
	scroll.SetOffsetX(scroll.GetOffsetX())
	t.State.columnLayouts = m.columns
	t.State.rowLayouts = m.rows
	t.State.headerHeight = m.headerHeight
	t.revealViewportCursor(false, false)
}

// revealViewportCursor scrolls the table's own viewport to the cursor. A glide
// (animate) is for long moves; any other move during a glide retargets it.
func (t Table[T]) revealViewportCursor(force, animate bool) {
	if t.State == nil {
		return
	}
	row := t.State.CursorIndex.Peek()
	y, height, ok := t.cursorRegion(row)
	if !ok {
		return
	}
	scroll := t.viewportState()
	m := t.State.viewportMetrics
	if !force && !t.State.revealed.needed(row, y, height, scroll) {
		return
	}
	t.State.revealed.record(row, y, height, scroll)
	offset := scroll.scrollTarget()
	if m.height > m.frozenHeight {
		if height > m.height-m.frozenHeight {
			// Oversized cells cannot fit: reveal their useful leading edge.
			offset = y - m.frozenHeight
		} else if y < offset+m.frozenHeight {
			offset = y - m.frozenHeight
		} else if y+height > offset+m.height {
			offset = y + height - m.height
		}
		if animate || scroll.animation != nil {
			scroll.animateOffset(offset)
		} else {
			scroll.SetOffset(offset)
		}
	}
	col := t.State.CursorColumn.Peek()
	if col >= max(0, t.FrozenColumns) && col < len(m.columns) && m.width > m.frozenWidth {
		bounds := m.columns[col]
		x := scroll.GetOffsetX()
		if bounds.width > m.width-m.frozenWidth {
			x = bounds.x - m.frozenWidth
		} else if bounds.x < x+m.frozenWidth {
			x = bounds.x - m.frozenWidth
		} else if bounds.x+bounds.width > x+m.width {
			x = bounds.x + bounds.width - m.width
		}
		scroll.SetOffsetX(x)
	}
}

// OnMouseWheel scrolls the table's own viewport when frozen panes are enabled.
func (t Table[T]) OnMouseWheel(event MouseEvent) bool {
	if !t.ownsViewport() {
		return false
	}
	s := t.viewportState()
	switch event.Button {
	case uv.MouseWheelUp:
		return s.ScrollUp(1)
	case uv.MouseWheelDown:
		return s.ScrollDown(1)
	case uv.MouseWheelLeft:
		return s.ScrollLeft(1)
	case uv.MouseWheelRight:
		return s.ScrollRight(1)
	}
	return false
}
