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
	ID               string                           // Unique identifier (required for focus management)
	Items            []T                              // List items to display
	CursorIndex      *Signal[int]                     // Signal tracking the cursor/highlighted index
	OnSelect         func(item T)                     // Callback invoked when Enter is pressed on an item
	ScrollController *ScrollController                // Optional controller for scroll-into-view
	Width            Dimension                        // Optional width (zero value = auto)
	Height           Dimension                        // Optional height (zero value = auto)
	RenderItem       func(item T, active bool) Widget // Function to render each item (uses default if nil)

	// Cached item heights computed during Build, used for scroll calculations
	itemHeights []int
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

	// Build children and cache their heights
	children := make([]Widget, len(l.Items))
	l.itemHeights = make([]int, len(l.Items))

	for i, item := range l.Items {
		widget := renderItem(item, i == cursorIdx)
		children[i] = widget

		// Cache height: check if widget implements Dimensioned, otherwise default to 1
		l.itemHeights[i] = getWidgetHeight(widget)
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

// getWidgetHeight extracts the height from a widget if it implements Dimensioned,
// otherwise returns 1 as the default height.
// Panics if the widget uses Fr dimensions, as fractional heights are not supported
// for list items (scroll calculations require known cell heights).
func getWidgetHeight(widget Widget) int {
	if dimensioned, ok := widget.(Dimensioned); ok {
		_, height := dimensioned.GetDimensions()
		if height.IsFr() {
			panic("List item widgets cannot use Fr height dimensions. Use Cells(n) for multi-line items or omit Height for single-line items.")
		}
		if height.IsCells() && height.CellsValue() > 0 {
			return height.CellsValue()
		}
	}
	return 1
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
	itemHeight := l.getItemHeight(cursorIdx)
	l.ScrollController.ScrollToView(itemY, itemHeight)
}

// getItemHeight returns the height of the item at the given index.
// Uses cached heights from Build if available, otherwise returns 1.
func (l *List[T]) getItemHeight(index int) int {
	if index < len(l.itemHeights) {
		return l.itemHeights[index]
	}
	return 1
}

// getItemY returns the Y position of the item at the given index,
// calculated by summing the heights of all preceding items.
func (l *List[T]) getItemY(index int) int {
	y := 0
	for i := 0; i < index && i < len(l.Items); i++ {
		y += l.getItemHeight(i)
	}
	return y
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
