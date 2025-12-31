package terma

import "strings"

// List is a focusable widget that displays a selectable list of items.
// It builds a Column of Text widgets, with the selected item highlighted.
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
//	    Child: &terma.List{
//	        ID:               "my-list",
//	        Items:            items,
//	        Selected:         selected,
//	        ScrollController: controller,
//	    },
//	}
type List struct {
	ID               string            // Unique identifier (required for focus management)
	Items            []string          // List items to display
	Selected         *Signal[int]      // Signal tracking the selected index
	ScrollController *ScrollController // Optional controller for scroll-into-view
	Width            Dimension         // Optional width (zero value = auto)
	Height           Dimension         // Optional height (zero value = auto)
	Style            Style             // Base style for unselected items
	SelectedStyle    Style             // Style for selected item (defaults to white on blue)
	Prefix           string            // Prefix for unselected items (defaults to "  ")
	SelectedPrefix   string            // Prefix for selected item (defaults to "▶ ")
}

func NewList(
	id string,
	items []string,
	selected *Signal[int],
	scrollController *ScrollController,
	width Dimension,
	height Dimension,
	style Style,
	selectedStyle Style,
	prefix string,
	selectedPrefix string,
) *List {
	return &List{
		ID:               id,
		Items:            items,
		Selected:         selected,
		ScrollController: scrollController,
		Width:            width,
		Height:           height,
		Style:            style,
		SelectedStyle:    selectedStyle,
		Prefix:           prefix,
		SelectedPrefix:   selectedPrefix,
	}
}

// Key returns the widget's unique identifier.
// Implements the Keyed interface.
func (l *List) Key() string {
	return l.ID
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (l *List) GetDimensions() (width, height Dimension) {
	return l.Width, l.Height
}

// GetStyle returns the base style of the list.
// Implements the Styled interface.
func (l *List) GetStyle() Style {
	return l.Style
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (l *List) IsFocusable() bool {
	return true
}

// Build returns a Column of Text widgets, each styled based on selection state.
func (l *List) Build(ctx BuildContext) Widget {
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

	// Determine prefixes (use defaults if not set)
	prefix := l.Prefix
	selectedPrefix := l.SelectedPrefix

	// Determine selected style (use default if not set)
	selectedStyle := l.SelectedStyle
	if selectedStyle.ForegroundColor == DefaultColor && selectedStyle.BackgroundColor == DefaultColor {
		selectedStyle = Style{
			ForegroundColor: Magenta,
		}
	}

	children := make([]Widget, len(l.Items))
	for i, item := range l.Items {
		var style Style
		var itemPrefix string

		if i == selected {
			style = selectedStyle
			itemPrefix = selectedPrefix
		} else {
			style = l.Style
			itemPrefix = prefix
		}

		// Get item height (default 1 if not specified)
		itemHeight := l.getItemHeight(i)
		var heightDim Dimension
		if itemHeight > 1 {
			heightDim = Cells(itemHeight)
		}

		children[i] = Text{
			Content: itemPrefix + item,
			Style:   style,
			Width:   Fr(1), // Fill available width for consistent background
			Height:  heightDim,
		}
	}

	// Ensure selected item is visible whenever we rebuild
	l.scrollSelectedIntoView()

	return Column{
		Width:    l.Width,
		Height:   l.Height,
		Children: children,
	}
}

// OnKey handles navigation keys, updating selection and scrolling into view.
// Implements the Focusable interface.
func (l *List) OnKey(event KeyEvent) bool {
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
func (l *List) scrollSelectedIntoView() {
	if l.ScrollController == nil || l.Selected == nil {
		return
	}
	selectedIdx := l.Selected.Peek()
	itemY := l.getItemY(selectedIdx)
	itemHeight := l.getItemHeight(selectedIdx)
	l.ScrollController.ScrollToView(itemY, itemHeight)
}

// getItemHeight returns the height of the item at the given index.
func (l *List) getItemHeight(index int) int {
	// Auto-calculate from content by counting lines
	if index < len(l.Items) {
		return countLines(l.Items[index])
	}
	return 1
}

// getItemY returns the Y position of the item at the given index,
// calculated by summing the heights of all preceding items.
func (l *List) getItemY(index int) int {
	y := 0
	for i := 0; i < index && i < len(l.Items); i++ {
		y += l.getItemHeight(i)
	}
	return y
}

// countLines returns the number of lines in a string (1 + number of newlines).
func countLines(s string) int {
	if s == "" {
		return 1
	}
	return strings.Count(s, "\n") + 1
}

// registerScrollCallbacks sets up callbacks on the ScrollController
// to update selection when mouse wheel scrolling occurs.
// The callbacks move selection first, then scroll only if needed.
func (l *List) registerScrollCallbacks() {
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
func (l *List) moveSelectionUp(count int) {
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
func (l *List) moveSelectionDown(count int) {
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
