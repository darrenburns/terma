package terma

import "fmt"

// List is a generic focusable widget that displays a navigable list of items.
// It builds a Column of widgets, with the active item (cursor position) highlighted.
// Use with Scrollable and a shared ScrollController to enable scroll-into-view.
//
// Example usage:
//
//	controller := terma.NewScrollController()
//	cursorIndex := terma.NewSignal(0)
//	items := []string{"Item 1", "Item 2", "Item 3"}
//
//	&terma.Scrollable{
//	    Controller: controller,
//	    Child: &terma.List[string]{
//	        ID:               "my-list",
//	        Items:            items,
//	        CursorIndex:      cursorIndex,
//	        ScrollController: controller,
//	        OnSelect: func(item string) {
//	            // Handle selection (Enter key)
//	        },
//	    },
//	}
type List[T any] struct {
	ID               string                           // Optional unique identifier (auto-generated from tree position if empty)
	Items            []T                              // List items to display
	CursorIndex      *Signal[int]                     // Signal tracking the cursor/highlighted index
	OnSelect         func(item T)                     // Callback invoked when Enter is pressed on an item
	ScrollController *ScrollController                // Optional controller for scroll-into-view
	Width            Dimension                        // Optional width (zero value = auto)
	Height           Dimension                        // Optional height (zero value = auto)
	RenderItem       func(item T, active bool) Widget // Function to render each item (uses default if nil)
	ItemHeight       int                              // Height of each item in cells (default 1, must be uniform)
}

// Key returns the widget's unique identifier.
// Implements the Keyed interface.
func (l *List[T]) Key() string {
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

// Build returns a Column of widgets, each rendered via RenderItem.
func (l *List[T]) Build(ctx BuildContext) Widget {
	if len(l.Items) == 0 {
		return Column{}
	}

	// Get and clamp cursor position to valid bounds
	cursorIdx := 0
	if l.CursorIndex != nil {
		cursorIdx = l.CursorIndex.Get()
		clamped := clampInt(cursorIdx, 0, len(l.Items)-1)
		if clamped != cursorIdx {
			l.CursorIndex.Set(clamped) // Update signal (won't loop since value changed)
			cursorIdx = clamped
		}
	}

	// Register scroll callbacks for mouse wheel support
	l.registerScrollCallbacks()

	// Use default render function if none provided
	renderItem := l.RenderItem
	if renderItem == nil {
		renderItem = defaultRenderItem[T]
	}

	// Build children
	children := make([]Widget, len(l.Items))
	for i, item := range l.Items {
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
	if l.CursorIndex == nil || len(l.Items) == 0 {
		return false
	}

	cursorIdx := l.CursorIndex.Peek()
	itemCount := len(l.Items)

	switch {
	case event.MatchString("enter"):
		// Handle selection (Enter key press)
		if l.OnSelect != nil {
			l.OnSelect(l.CursorItem())
		}
		return true

	case event.MatchString("up", "k"):
		if cursorIdx > 0 {
			l.CursorIndex.Set(cursorIdx - 1)
			l.scrollCursorIntoView()
		}
		return true

	case event.MatchString("down", "j"):
		if cursorIdx < itemCount-1 {
			l.CursorIndex.Set(cursorIdx + 1)
			l.scrollCursorIntoView()
		}
		return true

	case event.MatchString("home", "g"):
		l.CursorIndex.Set(0)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("end", "G"):
		l.CursorIndex.Set(itemCount - 1)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newCursor := cursorIdx - 10
		if newCursor < 0 {
			newCursor = 0
		}
		l.CursorIndex.Set(newCursor)
		l.scrollCursorIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		newCursor := cursorIdx + 10
		if newCursor >= itemCount {
			newCursor = itemCount - 1
		}
		l.CursorIndex.Set(newCursor)
		l.scrollCursorIntoView()
		return true
	}

	return false
}

// scrollCursorIntoView uses the ScrollController to ensure
// the cursor item is visible in the viewport.
func (l *List[T]) scrollCursorIntoView() {
	if l.ScrollController == nil || l.CursorIndex == nil {
		return
	}
	cursorIdx := l.CursorIndex.Peek()
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
	if l.CursorIndex == nil || len(l.Items) == 0 {
		return
	}
	cursorIdx := l.CursorIndex.Peek()
	newCursor := cursorIdx - count
	if newCursor < 0 {
		newCursor = 0
	}
	l.CursorIndex.Set(newCursor)
}

// moveCursorDown moves the cursor down by the given number of items.
func (l *List[T]) moveCursorDown(count int) {
	if l.CursorIndex == nil || len(l.Items) == 0 {
		return
	}
	cursorIdx := l.CursorIndex.Peek()
	itemCount := len(l.Items)
	newCursor := cursorIdx + count
	if newCursor >= itemCount {
		newCursor = itemCount - 1
	}
	l.CursorIndex.Set(newCursor)
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
// Returns the zero value of T if the list is empty or CursorIndex is nil.
func (l *List[T]) CursorItem() T {
	var zero T
	if l.CursorIndex == nil || len(l.Items) == 0 {
		return zero
	}
	idx := l.CursorIndex.Peek()
	if idx < 0 || idx >= len(l.Items) {
		return zero
	}
	return l.Items[idx]
}
