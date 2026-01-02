package terma

import "fmt"

// ListState holds the state for a List widget.
// When provided to a List, it becomes the source of truth for items and cursor position.
// This allows external control over the list's state.
type ListState[T any] struct {
	items       []T          // The list data
	CursorIndex *Signal[int] // Selection state (exported for direct access)
}

// NewListState creates a new ListState with default values.
func NewListState[T any]() *ListState[T] {
	return &ListState[T]{
		CursorIndex: NewSignal(0),
	}
}

// SetItems updates the list data and clamps cursor to valid range.
func (s *ListState[T]) SetItems(items []T) {
	s.items = items
	// Clamp cursor to valid range
	if idx := s.CursorIndex.Peek(); idx >= len(items) && len(items) > 0 {
		s.CursorIndex.Set(len(items) - 1)
	} else if len(items) == 0 {
		s.CursorIndex.Set(0)
	}
}

// Items returns the current list data.
func (s *ListState[T]) Items() []T {
	return s.items
}

// ItemCount returns the number of items.
func (s *ListState[T]) ItemCount() int {
	return len(s.items)
}

// SelectedItem returns the currently selected item (if any).
func (s *ListState[T]) SelectedItem() (T, bool) {
	idx := s.CursorIndex.Peek()
	if idx >= 0 && idx < len(s.items) {
		return s.items[idx], true
	}
	var zero T
	return zero, false
}

// SelectNext moves cursor to the next item.
func (s *ListState[T]) SelectNext() {
	s.CursorIndex.Update(func(i int) int {
		if i < len(s.items)-1 {
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
	if len(s.items) > 0 {
		s.CursorIndex.Set(len(s.items) - 1)
	}
}

// SelectIndex sets cursor to a specific index, clamped to valid range.
func (s *ListState[T]) SelectIndex(index int) {
	clamped := clampInt(index, 0, len(s.items)-1)
	s.CursorIndex.Set(clamped)
}

// List is a generic focusable widget that displays a navigable list of items.
// It builds a Column of widgets, with the active item (cursor position) highlighted.
// Use with Scrollable and a shared ScrollController to enable scroll-into-view.
//
// Example usage (simple - widget manages state):
//
//	list := &terma.List[string]{
//	    Items: []string{"Item 1", "Item 2", "Item 3"},
//	    OnSelect: func(item string) {
//	        // Handle selection
//	    },
//	}
//
// Example usage (controlled - external state):
//
//	state := terma.NewListState[string]()
//	state.SetItems(items)
//	list := &terma.List[string]{
//	    State: state,
//	    OnSelect: func(item string) {
//	        // Handle selection
//	    },
//	}
//	state.SelectLast() // Programmatic control
type List[T any] struct {
	ID               string                           // Optional unique identifier (auto-generated from tree position if empty)
	Items            []T                              // List items (used when State is nil)
	State            *ListState[T]                    // Optional state - if provided, is source of truth
	OnSelect         func(item T)                     // Callback invoked when Enter is pressed on an item
	ScrollController *ScrollController                // Optional controller for scroll-into-view
	Width            Dimension                        // Optional width (zero value = auto)
	Height           Dimension                        // Optional height (zero value = auto)
	RenderItem       func(item T, active bool) Widget // Function to render each item (uses default if nil)
	ItemHeight       int                              // Height of each item in cells (default 1, must be uniform)

	resolvedState *ListState[T] // Cached during Build() for OnKey access
}

// WidgetID returns the widget's unique identifier.
// Implements the Identifiable interface.
func (l *List[T]) WidgetID() string {
	return l.ID
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (l *List[T]) GetDimensions() (width, height Dimension) {
	return l.Width, l.Height
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (l *List[T]) IsFocusable() bool {
	return true
}

// getState returns provided state or creates/retrieves internal state.
func (l *List[T]) getState(ctx BuildContext) *ListState[T] {
	if l.State != nil {
		return l.State
	}

	// Use explicit ID or auto-ID for internal state
	id := l.ID
	if id == "" {
		id = ctx.AutoID()
	}

	state := GetOrCreateState(id, NewListState[T])

	// Sync Items field to internal state
	// (only when State not provided - widget's Items field is source)
	state.SetItems(l.Items)

	return state
}

// Build returns a Column of widgets, each rendered via RenderItem.
func (l *List[T]) Build(ctx BuildContext) Widget {
	// Resolve state (provided or internal)
	l.resolvedState = l.getState(ctx)

	// Get items and cursor from state
	items := l.resolvedState.Items()
	if len(items) == 0 {
		return Column{}
	}

	// Get cursor position (subscribes to changes)
	cursorIdx := l.resolvedState.CursorIndex.Get()

	// Clamp cursor to valid bounds
	clamped := clampInt(cursorIdx, 0, len(items)-1)
	if clamped != cursorIdx {
		l.resolvedState.CursorIndex.Set(clamped)
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
func (l *List[T]) OnKey(event KeyEvent) bool {
	state := l.resolvedState
	if state == nil || state.ItemCount() == 0 {
		return false
	}

	cursorIdx := state.CursorIndex.Peek()
	itemCount := state.ItemCount()

	switch {
	case event.MatchString("enter"):
		// Handle selection (Enter key press)
		if l.OnSelect != nil {
			if item, ok := state.SelectedItem(); ok {
				l.OnSelect(item)
			}
		}
		return true

	case event.MatchString("up", "k"):
		state.SelectPrevious()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("down", "j"):
		state.SelectNext()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("home", "g"):
		state.SelectFirst()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("end", "G"):
		state.SelectLast()
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		state.SelectIndex(newCursor)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		newCursor := cursorIdx + 10
		if newCursor >= itemCount {
			newCursor = itemCount - 1
		}
		state.SelectIndex(newCursor)
		l.scrollCursorIntoView()
		return true
	}

	return false
}

// scrollCursorIntoView uses the ScrollController to ensure
// the cursor item is visible in the viewport.
func (l *List[T]) scrollCursorIntoView() {
	if l.ScrollController == nil || l.resolvedState == nil {
		return
	}
	cursorIdx := l.resolvedState.CursorIndex.Peek()
	itemY := l.getItemY(cursorIdx)
	l.ScrollController.ScrollToView(itemY, l.getItemHeight())
}

// getItemHeight returns the uniform height of list items.
// Returns ItemHeight if set, otherwise defaults to 1.
func (l *List[T]) getItemHeight() int {
	if l.ItemHeight > 0 {
		return l.ItemHeight
	}
	return 1
}

// getItemY returns the Y position of the item at the given index.
func (l *List[T]) getItemY(index int) int {
	return index * l.getItemHeight()
}

// registerScrollCallbacks sets up callbacks on the ScrollController
// to update cursor position when mouse wheel scrolling occurs.
// The callbacks move cursor first, then scroll only if needed.
func (l *List[T]) registerScrollCallbacks() {
	if l.ScrollController == nil {
		return
	}

	l.ScrollController.OnScrollUp = func(lines int) bool {
		l.moveCursorUp(lines)
		l.scrollCursorIntoView()
		return true // We handle scrolling via scrollCursorIntoView
	}
	l.ScrollController.OnScrollDown = func(lines int) bool {
		l.moveCursorDown(lines)
		l.scrollCursorIntoView()
		return true // We handle scrolling via scrollCursorIntoView
	}
}

// moveCursorUp moves the cursor up by the given number of items.
func (l *List[T]) moveCursorUp(count int) {
	state := l.resolvedState
	if state == nil || state.ItemCount() == 0 {
		return
	}
	cursorIdx := state.CursorIndex.Peek()
	newCursor := cursorIdx - count
	if newCursor < 0 {
		newCursor = 0
	}
	state.SelectIndex(newCursor)
}

// moveCursorDown moves the cursor down by the given number of items.
func (l *List[T]) moveCursorDown(count int) {
	state := l.resolvedState
	if state == nil || state.ItemCount() == 0 {
		return
	}
	cursorIdx := state.CursorIndex.Peek()
	itemCount := state.ItemCount()
	newCursor := cursorIdx + count
	if newCursor >= itemCount {
		newCursor = itemCount - 1
	}
	state.SelectIndex(newCursor)
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
// Deprecated: Use state.SelectedItem() instead when using controlled mode.
func (l *List[T]) CursorItem() T {
	var zero T
	state := l.resolvedState
	if state == nil || state.ItemCount() == 0 {
		return zero
	}
	if item, ok := state.SelectedItem(); ok {
		return item
	}
	return zero
}
