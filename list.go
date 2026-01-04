package terma

import "fmt"

// ListState holds the state for a List widget.
// It is the source of truth for items and cursor position, and must be provided to List.
// Items is a reactive Signal - changes trigger automatic re-renders.
type ListState[T any] struct {
	Items       *AnySignal[[]T]              // Reactive list data
	CursorIndex *Signal[int]                 // Cursor position
	Selection   *AnySignal[map[int]struct{}] // Selected item indices (for multi-select)

	anchorIndex *int // Anchor point for shift-selection (nil = no anchor)
}

// NewListState creates a new ListState with the given initial items.
func NewListState[T any](initialItems []T) *ListState[T] {
	if initialItems == nil {
		initialItems = []T{}
	}
	return &ListState[T]{
		Items:       NewAnySignal(initialItems),
		CursorIndex: NewSignal(0),
		Selection:   NewAnySignal(make(map[int]struct{})),
	}
}

// SetItems replaces all items and clamps cursor to valid range.
func (s *ListState[T]) SetItems(items []T) {
	if items == nil {
		items = []T{}
	}
	s.Items.Set(items)
	s.clampCursor()
}

// GetItems returns the current list data (without subscribing to changes).
func (s *ListState[T]) GetItems() []T {
	return s.Items.Peek()
}

// ItemCount returns the number of items.
func (s *ListState[T]) ItemCount() int {
	return len(s.Items.Peek())
}

// Append adds an item to the end of the list.
func (s *ListState[T]) Append(item T) {
	s.Items.Update(func(items []T) []T {
		return append(items, item)
	})
}

// Prepend adds an item to the beginning of the list.
func (s *ListState[T]) Prepend(item T) {
	s.Items.Update(func(items []T) []T {
		return append([]T{item}, items...)
	})
	// Adjust cursor to keep same item selected
	s.CursorIndex.Update(func(i int) int {
		return i + 1
	})
}

// InsertAt inserts an item at the specified index.
// If index is out of bounds, it's clamped to valid range.
func (s *ListState[T]) InsertAt(index int, item T) {
	s.Items.Update(func(items []T) []T {
		if index < 0 {
			index = 0
		}
		if index > len(items) {
			index = len(items)
		}
		// Make room for new item
		items = append(items, item) // Extend slice
		copy(items[index+1:], items[index:])
		items[index] = item
		return items
	})
	// Adjust cursor if insertion was at or before cursor
	cursorIdx := s.CursorIndex.Peek()
	if index <= cursorIdx {
		s.CursorIndex.Set(cursorIdx + 1)
	}
}

// RemoveAt removes the item at the specified index.
// Returns true if an item was removed, false if index was out of bounds.
func (s *ListState[T]) RemoveAt(index int) bool {
	items := s.Items.Peek()
	if index < 0 || index >= len(items) {
		return false
	}
	s.Items.Update(func(items []T) []T {
		return append(items[:index], items[index+1:]...)
	})
	s.clampCursor()
	return true
}

// RemoveWhere removes all items matching the predicate.
// Returns the number of items removed.
func (s *ListState[T]) RemoveWhere(predicate func(T) bool) int {
	removed := 0
	s.Items.Update(func(items []T) []T {
		result := make([]T, 0, len(items))
		for _, item := range items {
			if !predicate(item) {
				result = append(result, item)
			} else {
				removed++
			}
		}
		return result
	})
	s.clampCursor()
	return removed
}

// Clear removes all items from the list.
func (s *ListState[T]) Clear() {
	s.Items.Set([]T{})
	s.CursorIndex.Set(0)
}

// SelectedItem returns the currently selected item (if any).
func (s *ListState[T]) SelectedItem() (T, bool) {
	items := s.Items.Peek()
	idx := s.CursorIndex.Peek()
	if idx >= 0 && idx < len(items) {
		return items[idx], true
	}
	var zero T
	return zero, false
}

// SelectNext moves cursor to the next item.
func (s *ListState[T]) SelectNext() {
	items := s.Items.Peek()
	s.CursorIndex.Update(func(i int) int {
		if i < len(items)-1 {
			return i + 1
		}
		return i
	})
}

// SelectPrevious moves cursor to the previous item.
func (s *ListState[T]) SelectPrevious() {
	s.CursorIndex.Update(func(i int) int {
		if i > 0 {
			return i - 1
		}
		return i
	})
}

// SelectFirst moves cursor to the first item.
func (s *ListState[T]) SelectFirst() {
	s.CursorIndex.Set(0)
}

// SelectLast moves cursor to the last item.
func (s *ListState[T]) SelectLast() {
	items := s.Items.Peek()
	if len(items) > 0 {
		s.CursorIndex.Set(len(items) - 1)
	}
}

// SelectIndex sets cursor to a specific index, clamped to valid range.
func (s *ListState[T]) SelectIndex(index int) {
	items := s.Items.Peek()
	clamped := clampInt(index, 0, len(items)-1)
	s.CursorIndex.Set(clamped)
}

// clampCursor ensures cursor is within valid bounds after items change.
func (s *ListState[T]) clampCursor() {
	items := s.Items.Peek()
	idx := s.CursorIndex.Peek()
	if len(items) == 0 {
		s.CursorIndex.Set(0)
	} else if idx >= len(items) {
		s.CursorIndex.Set(len(items) - 1)
	}
}

// ToggleSelection toggles the selection state of the item at the given index.
func (s *ListState[T]) ToggleSelection(index int) {
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

// Select adds the item at the given index to the selection.
func (s *ListState[T]) Select(index int) {
	s.Selection.Update(func(sel map[int]struct{}) map[int]struct{} {
		newSel := make(map[int]struct{}, len(sel)+1)
		for k := range sel {
			newSel[k] = struct{}{}
		}
		newSel[index] = struct{}{}
		return newSel
	})
}

// Deselect removes the item at the given index from the selection.
func (s *ListState[T]) Deselect(index int) {
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

// IsSelected returns true if the item at the given index is selected.
func (s *ListState[T]) IsSelected(index int) bool {
	sel := s.Selection.Peek()
	_, exists := sel[index]
	return exists
}

// ClearSelection removes all items from the selection.
func (s *ListState[T]) ClearSelection() {
	s.Selection.Set(make(map[int]struct{}))
}

// SelectAll selects all items in the list.
func (s *ListState[T]) SelectAll() {
	items := s.Items.Peek()
	sel := make(map[int]struct{}, len(items))
	for i := range items {
		sel[i] = struct{}{}
	}
	s.Selection.Set(sel)
}

// SelectedItems returns all currently selected items.
func (s *ListState[T]) SelectedItems() []T {
	items := s.Items.Peek()
	sel := s.Selection.Peek()
	result := make([]T, 0, len(sel))
	for i := range items {
		if _, exists := sel[i]; exists {
			result = append(result, items[i])
		}
	}
	return result
}

// SelectedIndices returns the indices of all selected items in ascending order.
func (s *ListState[T]) SelectedIndices() []int {
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
func (s *ListState[T]) SetAnchor(index int) {
	s.anchorIndex = &index
}

// ClearAnchor removes the anchor point.
func (s *ListState[T]) ClearAnchor() {
	s.anchorIndex = nil
}

// HasAnchor returns true if an anchor point is set.
func (s *ListState[T]) HasAnchor() bool {
	return s.anchorIndex != nil
}

// GetAnchor returns the anchor index, or -1 if no anchor is set.
func (s *ListState[T]) GetAnchor() int {
	if s.anchorIndex == nil {
		return -1
	}
	return *s.anchorIndex
}

// SelectRange selects all items between from and to (inclusive).
func (s *ListState[T]) SelectRange(from, to int) {
	if from > to {
		from, to = to, from
	}
	items := s.Items.Peek()
	if from < 0 {
		from = 0
	}
	if to >= len(items) {
		to = len(items) - 1
	}
	sel := make(map[int]struct{}, to-from+1)
	for i := from; i <= to; i++ {
		sel[i] = struct{}{}
	}
	s.Selection.Set(sel)
}

// List is a generic focusable widget that displays a navigable list of items.
// It builds a Column of widgets, with the active item (cursor position) highlighted.
// Use with Scrollable and a shared ScrollState to enable scroll-into-view.
//
// Example usage:
//
//	state := terma.NewListState([]string{"Item 1", "Item 2", "Item 3"})
//	list := terma.List[string]{
//	    State: state,
//	    OnSelect: func(item string) {
//	        // Handle selection
//	    },
//	}
//
//	// Add item at runtime:
//	state.Append("Item 4")
//
//	// Remove item at runtime:
//	state.RemoveAt(0)
type List[T any] struct {
	ID          string                                          // Optional unique identifier
	State       *ListState[T]                                   // Required - holds items and cursor position
	OnSelect    func(item T)                                    // Callback invoked when Enter is pressed on an item
	ScrollState *ScrollState                                    // Optional state for scroll-into-view
	RenderItem  func(item T, active bool, selected bool) Widget // Function to render each item (uses default if nil)
	ItemHeight  int                                             // Height of each item in cells (default 1, must be uniform)
	MultiSelect bool                                            // Enable multi-select mode (space to toggle, shift+move to extend)
	Width       Dimension                                       // Optional width (zero value = auto)
	Height      Dimension                                       // Optional height (zero value = auto)
	Style       Style                                           // Optional styling
	Click       func()                                          // Optional callback invoked when clicked
	Hover       func(bool)                                      // Optional callback invoked when hover state changes
}

// WidgetID returns the widget's unique identifier.
// Implements the Identifiable interface.
func (l List[T]) WidgetID() string {
	return l.ID
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (l List[T]) GetDimensions() (width, height Dimension) {
	return l.Width, l.Height
}

// GetStyle returns the style of the list widget.
// Implements the Styled interface.
func (l List[T]) GetStyle() Style {
	return l.Style
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (l List[T]) OnClick() {
	if l.Click != nil {
		l.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (l List[T]) OnHover(hovered bool) {
	if l.Hover != nil {
		l.Hover(hovered)
	}
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (l List[T]) IsFocusable() bool {
	return true
}

// Build returns a Column of widgets, each rendered via RenderItem.
func (l List[T]) Build(ctx BuildContext) Widget {
	if l.State == nil {
		return Column{}
	}

	// Get items (subscribes to changes via signal)
	items := l.State.Items.Get()
	if len(items) == 0 {
		return Column{}
	}

	// Get cursor position (subscribes to changes)
	cursorIdx := l.State.CursorIndex.Get()

	// Get selection state (subscribes to changes)
	var Selection map[int]struct{}
	if l.MultiSelect {
		Selection = l.State.Selection.Get()
	}

	// Clamp cursor to valid bounds
	clamped := clampInt(cursorIdx, 0, len(items)-1)
	if clamped != cursorIdx {
		l.State.CursorIndex.Set(clamped)
		cursorIdx = clamped
	}

	// Register scroll callbacks for mouse wheel support
	l.registerScrollCallbacks()

	// Use default render function if none provided
	renderItem := l.RenderItem
	if renderItem == nil {
		renderItem = l.themedDefaultRenderItem(ctx)
	}

	// Build children
	children := make([]Widget, len(items))
	for i, item := range items {
		_, selected := Selection[i]
		children[i] = renderItem(item, i == cursorIdx, selected)
	}

	// Ensure cursor item is visible whenever we rebuild
	l.scrollCursorIntoView()

	return Column{
		Width:    l.Width,
		Height:   l.Height,
		Children: children,
	}
}

// themedDefaultRenderItem returns a themed render function for list items.
// Captures theme colors from the context for use in the render function.
func (l List[T]) themedDefaultRenderItem(ctx BuildContext) func(item T, active bool, selected bool) Widget {
	theme := ctx.Theme()
	return func(item T, active bool, selected bool) Widget {
		content := fmt.Sprintf("%v", item)
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
			Width:   Fr(1), // Fill available width for consistent background
		}
	}
}

// defaultRenderItem provides a non-themed default rendering for list items.
// Deprecated: Use themedDefaultRenderItem instead, which applies theme colors.
func defaultRenderItem[T any](item T, active bool, selected bool) Widget {
	content := fmt.Sprintf("%v", item)
	prefix := "  "
	style := Style{}

	if selected && active {
		prefix = "▶*"
		style.ForegroundColor = Magenta
	} else if active {
		prefix = "▶ "
		style.ForegroundColor = Magenta
	} else if selected {
		prefix = " *"
	}

	return Text{
		Content: prefix + content,
		Style:   style,
		Width:   Fr(1), // Fill available width for consistent background
	}
}

// OnKey handles navigation keys and selection, updating cursor position and scrolling into view.
// Implements the Focusable interface.
func (l List[T]) OnKey(event KeyEvent) bool {
	if l.State == nil || l.State.ItemCount() == 0 {
		return false
	}

	cursorIdx := l.State.CursorIndex.Peek()
	itemCount := l.State.ItemCount()

	// Handle multi-select specific keys (shift+movement to extend selection)
	if l.MultiSelect {
		switch {
		case event.MatchString("shift+up", "shift+k"):
			l.handleShiftMove(-1)
			return true

		case event.MatchString("shift+down", "shift+j"):
			l.handleShiftMove(1)
			return true

		case event.MatchString("shift+home"):
			l.handleShiftMoveTo(0)
			return true

		case event.MatchString("shift+end"):
			l.handleShiftMoveTo(itemCount - 1)
			return true
		}
	}

	switch {
	case event.MatchString("enter"):
		// Handle selection (Enter key press)
		if l.OnSelect != nil {
			if item, ok := l.State.SelectedItem(); ok {
				l.OnSelect(item)
			}
		}
		return true

	case event.MatchString("up", "k"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		l.State.SelectPrevious()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("down", "j"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		l.State.SelectNext()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("home", "g"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		l.State.SelectFirst()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("end", "G"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		l.State.SelectLast()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		l.State.SelectIndex(newCursor)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		if l.MultiSelect {
			l.State.ClearSelection()
			l.State.ClearAnchor()
		}
		newCursor := cursorIdx + 10
		if newCursor >= itemCount {
			newCursor = itemCount - 1
		}
		l.State.SelectIndex(newCursor)
		l.scrollCursorIntoView()
		return true
	}

	return false
}

// handleShiftMove extends selection by moving cursor by delta and selecting the range.
func (l List[T]) handleShiftMove(delta int) {
	cursorIdx := l.State.CursorIndex.Peek()

	// Set anchor if not already set
	if !l.State.HasAnchor() {
		l.State.SetAnchor(cursorIdx)
	}

	// Move cursor
	if delta > 0 {
		l.State.SelectNext()
	} else {
		l.State.SelectPrevious()
	}

	// Update selection range from anchor to new cursor
	newCursor := l.State.CursorIndex.Peek()
	l.State.SelectRange(l.State.GetAnchor(), newCursor)

	l.scrollCursorIntoView()
}

// handleShiftMoveTo extends selection to a specific index.
func (l List[T]) handleShiftMoveTo(targetIdx int) {
	cursorIdx := l.State.CursorIndex.Peek()

	// Set anchor if not already set
	if !l.State.HasAnchor() {
		l.State.SetAnchor(cursorIdx)
	}

	// Move cursor to target
	l.State.SelectIndex(targetIdx)

	// Update selection range from anchor to new cursor
	newCursor := l.State.CursorIndex.Peek()
	l.State.SelectRange(l.State.GetAnchor(), newCursor)

	l.scrollCursorIntoView()
}

// scrollCursorIntoView uses the ScrollState to ensure
// the cursor item is visible in the viewport.
func (l List[T]) scrollCursorIntoView() {
	if l.ScrollState == nil || l.State == nil {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	itemY := l.getItemY(cursorIdx)
	l.ScrollState.ScrollToView(itemY, l.getItemHeight())
}

// getItemHeight returns the uniform height of list items.
// If ItemHeight is explicitly set, uses that value.
// Otherwise, attempts to infer height from RenderItem by checking
// if the returned widget has an explicit Cells height dimension.
// Falls back to 1 if height cannot be determined.
func (l List[T]) getItemHeight() int {
	if l.ItemHeight > 0 {
		return l.ItemHeight
	}

	// Try to infer from RenderItem
	if l.State != nil && l.State.ItemCount() > 0 {
		items := l.State.Items.Peek()
		renderItem := l.RenderItem
		if renderItem == nil {
			renderItem = defaultRenderItem[T]
		}

		// Render the first item and check its dimensions
		widget := renderItem(items[0], false, false)
		if dimensioned, ok := widget.(Dimensioned); ok {
			_, height := dimensioned.GetDimensions()
			if height.IsCells() {
				return height.CellsValue()
			}
		}
	}

	return 1
}

// getItemY returns the Y position of the item at the given index.
func (l List[T]) getItemY(index int) int {
	return index * l.getItemHeight()
}

// registerScrollCallbacks sets up callbacks on the ScrollState
// to update cursor position when mouse wheel scrolling occurs.
// The callbacks move cursor first, then scroll only if needed.
func (l List[T]) registerScrollCallbacks() {
	if l.ScrollState == nil {
		return
	}

	l.ScrollState.OnScrollUp = func(lines int) bool {
		l.moveCursorUp(lines)
		l.scrollCursorIntoView()
		return true // We handle scrolling via scrollCursorIntoView
	}
	l.ScrollState.OnScrollDown = func(lines int) bool {
		l.moveCursorDown(lines)
		l.scrollCursorIntoView()
		return true // We handle scrolling via scrollCursorIntoView
	}
}

// moveCursorUp moves the cursor up by the given number of items.
func (l List[T]) moveCursorUp(count int) {
	if l.State == nil || l.State.ItemCount() == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	newCursor := cursorIdx - count
	if newCursor < 0 {
		newCursor = 0
	}
	l.State.SelectIndex(newCursor)
}

// moveCursorDown moves the cursor down by the given number of items.
func (l List[T]) moveCursorDown(count int) {
	if l.State == nil || l.State.ItemCount() == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	itemCount := l.State.ItemCount()
	newCursor := cursorIdx + count
	if newCursor >= itemCount {
		newCursor = itemCount - 1
	}
	l.State.SelectIndex(newCursor)
}

// clampInt clamps value to the range [min, max].
func clampInt(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// CursorItem returns the item at the current cursor position.
// Returns the zero value of T if the list is empty or state is nil.
// Deprecated: Use State.SelectedItem() instead.
func (l List[T]) CursorItem() T {
	var zero T
	if l.State == nil || l.State.ItemCount() == 0 {
		return zero
	}
	if item, ok := l.State.SelectedItem(); ok {
		return item
	}
	return zero
}
