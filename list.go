package terma

import "fmt"

// List is a generic focusable widget that displays a selectable list of items.
// It builds a Column of widgets, with the selected item highlighted.
// Use with Scrollable and a shared ScrollController to enable scroll-into-view.
//
// Example usage:
//
//	controller := terma.NewScrollController()
//	selected := terma.NewSignal(0)
//	items := []string{"Item 1", "Item 2", "Item 3"}
//
//	&terma.Scrollable{
//	    Controller: controller,
//	    Child: &terma.List[string]{
//	        ID:               "my-list",
//	        Items:            items,
//	        Selected:         selected,
//	        ScrollController: controller,
//	    },
//	}
type List[T any] struct {
	ID               string                             // Unique identifier (required for focus management)
	Items            []T                                // List items to display
	Selected         *Signal[int]                       // Signal tracking the selected index
	ScrollController *ScrollController                  // Optional controller for scroll-into-view
	Width            Dimension                          // Optional width (zero value = auto)
	Height           Dimension                          // Optional height (zero value = auto)
	RenderItem       func(item T, selected bool) Widget // Function to render each item (uses default if nil)

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

	// Get and clamp selection to valid bounds
	selected := 0
	if l.Selected != nil {
		selected = l.Selected.Get()
		clamped := clampInt(selected, 0, len(l.Items)-1)
		if clamped != selected {
			l.Selected.Set(clamped) // Update signal (won't loop since value changed)
			selected = clamped
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
		widget := renderItem(item, i == selected)
		children[i] = widget

		// Cache height: check if widget implements Dimensioned, otherwise default to 1
		l.itemHeights[i] = getWidgetHeight(widget)
	}

	// Ensure selected item is visible whenever we rebuild
	l.scrollSelectedIntoView()

	return Column{
		Width:    l.Width,
		Height:   l.Height,
		Children: children,
	}
}

// defaultRenderItem provides a default rendering for list items.
// Uses magenta foreground and "▶ " prefix for selected items.
func defaultRenderItem[T any](item T, selected bool) Widget {
	content := fmt.Sprintf("%v", item)
	prefix := "  "
	style := Style{}

	if selected {
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

// OnKey handles navigation keys, updating selection and scrolling into view.
// Implements the Focusable interface.
func (l *List[T]) OnKey(event KeyEvent) bool {
	if l.Selected == nil || len(l.Items) == 0 {
		return false
	}

	selected := l.Selected.Peek()
	itemCount := len(l.Items)

	switch {
	case event.MatchString("up", "k"):
		if selected > 0 {
			l.Selected.Set(selected - 1)
			l.scrollSelectedIntoView()
		}
		return true

	case event.MatchString("down", "j"):
		if selected < itemCount-1 {
			l.Selected.Set(selected + 1)
			l.scrollSelectedIntoView()
		}
		return true

	case event.MatchString("home", "g"):
		l.Selected.Set(0)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("end", "G"):
		l.Selected.Set(itemCount - 1)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("pgup", "ctrl+u"):
		newSelected := selected - 10
		if newSelected < 0 {
			newSelected = 0
		}
		l.Selected.Set(newSelected)
		l.scrollSelectedIntoView()
		return true

	case event.MatchString("pgdown", "ctrl+d"):
		newSelected := selected + 10
		if newSelected >= itemCount {
			newSelected = itemCount - 1
		}
		l.Selected.Set(newSelected)
		l.scrollSelectedIntoView()
		return true
	}

	return false
}

// scrollSelectedIntoView uses the ScrollController to ensure
// the selected item is visible in the viewport.
func (l *List[T]) scrollSelectedIntoView() {
	if l.ScrollController == nil || l.Selected == nil {
		return
	}
	selectedIdx := l.Selected.Peek()
	itemY := l.getItemY(selectedIdx)
	itemHeight := l.getItemHeight(selectedIdx)
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
// to update selection when mouse wheel scrolling occurs.
// The callbacks move selection first, then scroll only if needed.
func (l *List[T]) registerScrollCallbacks() {
	if l.ScrollController == nil {
		return
	}

	l.ScrollController.OnScrollUp = func(lines int) bool {
		l.moveSelectionUp(lines)
		l.scrollSelectedIntoView()
		return true // We handle scrolling via scrollSelectedIntoView
	}
	l.ScrollController.OnScrollDown = func(lines int) bool {
		l.moveSelectionDown(lines)
		l.scrollSelectedIntoView()
		return true // We handle scrolling via scrollSelectedIntoView
	}
}

// moveSelectionUp moves the selection up by the given number of items.
func (l *List[T]) moveSelectionUp(count int) {
	if l.Selected == nil || len(l.Items) == 0 {
		return
	}
	selected := l.Selected.Peek()
	newSelected := selected - count
	if newSelected < 0 {
		newSelected = 0
	}
	l.Selected.Set(newSelected)
}

// moveSelectionDown moves the selection down by the given number of items.
func (l *List[T]) moveSelectionDown(count int) {
	if l.Selected == nil || len(l.Items) == 0 {
		return
	}
	selected := l.Selected.Peek()
	itemCount := len(l.Items)
	newSelected := selected + count
	if newSelected >= itemCount {
		newSelected = itemCount - 1
	}
	l.Selected.Set(newSelected)
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

// SelectedItem returns the currently selected item.
// Returns the zero value of T if the list is empty or Selected is nil.
func (l *List[T]) SelectedItem() T {
	var zero T
	if l.Selected == nil || len(l.Items) == 0 {
		return zero
	}
	idx := l.Selected.Peek()
	if idx < 0 || idx >= len(l.Items) {
		return zero
	}
	return l.Items[idx]
}
