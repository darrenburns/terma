package terma

import "fmt"

// ListState holds the state for a List widget.
// It is the source of truth for items and cursor position, and must be provided to List.
// Items is a reactive Signal - changes trigger automatic re-renders.
type ListState[T any] struct {
	Items       *AnySignal[[]T] // Reactive list data
	CursorIndex *Signal[int]    // Cursor position
}

// NewListState creates a new ListState with the given initial items.
func NewListState[T any](initialItems []T) *ListState[T] {
	if initialItems == nil {
		initialItems = []T{}
	}
	return &ListState[T]{
		Items:       NewAnySignal(initialItems),
		CursorIndex: NewSignal(0),
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
	ID          string                           // Optional unique identifier
	State       *ListState[T]                    // Required - holds items and cursor position
	OnSelect    func(item T)                     // Callback invoked when Enter is pressed on an item
	ScrollState *ScrollState                     // Optional state for scroll-into-view
	Width       Dimension                        // Optional width (zero value = auto)
	Height      Dimension                        // Optional height (zero value = auto)
	RenderItem  func(item T, active bool) Widget // Function to render each item (uses default if nil)
	ItemHeight  int                              // Height of each item in cells (default 1, must be uniform)
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
		renderItem = defaultRenderItem[T]
	}

	// Build children
	children := make([]Widget, len(items))
	for i, item := range items {
		children[i] = renderItem(item, i == cursorIdx)
	}

	// Ensure cursor item is visible whenever we rebuild
	l.scrollCursorIntoView()

	return Column{
		Width:    l.Width,
		Height:   l.Height,
		Children: children,
	}
}

// defaultRenderItem provides a default rendering for list items.
// Uses magenta foreground and "▶ " prefix for the active (cursor) item.
func defaultRenderItem[T any](item T, active bool) Widget {
	content := fmt.Sprintf("%v", item)
	prefix := "  "
	style := Style{}

	if active {
		prefix = "▶ "
		style.ForegroundColor = Magenta
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
		l.State.SelectPrevious()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("down", "j"):
		l.State.SelectNext()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("home", "g"):
		l.State.SelectFirst()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("end", "G"):
		l.State.SelectLast()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		l.State.SelectIndex(newCursor)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
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
		widget := renderItem(items[0], false)
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
