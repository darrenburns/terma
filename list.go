package terma

import "fmt"

// ListState holds the state for a List widget.
// It is the source of truth for items and cursor position, and must be provided to List.
// Items is a reactive Signal - changes trigger automatic re-renders.
type ListState[T any] struct {
	Items       AnySignal[[]T]              // Reactive list data
	CursorIndex Signal[int]                 // Cursor position
	Selection   AnySignal[map[int]struct{}] // Selected item indices (for multi-select)

	anchorIndex *int // Anchor point for shift-selection (nil = no anchor)

	itemLayouts       []listItemLayout // Cached layout metrics (per item)
	viewIndices       []int            // View index -> source index for filtered views
	viewIndexBySource map[int]int      // Source index -> view index for filtered views
	cachedMatches     []MatchResult    // Cached match results from filtering
	cachedFilterQuery string           // Query used for cached filter results
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
	s.resetFilterCache()
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
	s.resetFilterCache()
}

// Prepend adds an item to the beginning of the list.
func (s *ListState[T]) Prepend(item T) {
	s.Items.Update(func(items []T) []T {
		return append([]T{item}, items...)
	})
	s.resetFilterCache()
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
	s.resetFilterCache()
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
	s.resetFilterCache()
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
	if removed > 0 {
		s.resetFilterCache()
	}
	s.clampCursor()
	return removed
}

// Clear removes all items from the list.
func (s *ListState[T]) Clear() {
	s.Items.Set([]T{})
	s.resetFilterCache()
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

func (s *ListState[T]) setViewIndices(indices []int) {
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

func (s *ListState[T]) resetFilterCache() {
	s.setViewIndices(nil)
	s.cachedMatches = nil
	s.cachedFilterQuery = ""
}

func (s *ListState[T]) viewIndexForSource(sourceIdx int) (int, bool) {
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

// ApplyFilter applies a filter to the items and caches the results.
// Returns the number of items that match the filter.
// The cached results are used by List.Build() to avoid re-filtering.
func (s *ListState[T]) ApplyFilter(filter *FilterState, matchItem func(item T, query string, options FilterOptions) MatchResult) int {
	items := s.Items.Peek()
	if len(items) == 0 {
		s.setViewIndices(nil)
		s.cachedMatches = nil
		s.cachedFilterQuery = ""
		return 0
	}

	query, options := filterStateValues(filter)
	if matchItem == nil {
		matchItem = defaultListMatchItem[T]
	}

	filtered := ApplyFilter(items, query, func(item T, q string) MatchResult {
		return matchItem(item, q, options)
	})

	s.setViewIndices(filtered.Indices)
	s.cachedMatches = filtered.Matches
	s.cachedFilterQuery = query

	return len(filtered.Items)
}

// FilteredCount returns the number of items after filtering.
// Returns total item count if no filter has been applied.
func (s *ListState[T]) FilteredCount() int {
	if s.viewIndices != nil {
		return len(s.viewIndices)
	}
	return len(s.Items.Peek())
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
	ID                  string                                                             // Optional unique identifier
	DisableFocus        bool                                                               // If true, prevent keyboard focus
	CursorStyle                                                                            // Embedded - CursorPrefix/SelectedPrefix fields for customizable indicators
	State               *ListState[T]                                                      // Required - holds items and cursor position
	OnSelect            func(item T)                                                       // Callback invoked when Enter is pressed on an item
	OnCursorChange      func(item T)                                                       // Callback invoked when cursor moves to a different item
	ScrollState         *ScrollState                                                       // Optional state for scroll-into-view
	RenderItem          func(item T, active bool, selected bool) Widget                    // Function to render each item (uses default if nil)
	RenderItemWithMatch func(item T, active bool, selected bool, match MatchResult) Widget // Optional render function with match data
	Filter              *FilterState                                                       // Optional filter state for matching items
	MatchItem           func(item T, query string, options FilterOptions) MatchResult      // Optional matcher for filtering/highlighting
	ItemHeight          int                                                                // Optional uniform item height override (default 0 = layout metrics / fallback 1)
	MultiSelect         bool                                                               // Enable multi-select mode (space to toggle, shift+move to extend)
	Width               Dimension                                                          // Deprecated: use Style.Width
	Height              Dimension                                                          // Deprecated: use Style.Height
	Style               Style                                                              // Optional styling
	Click               func(MouseEvent)                                                   // Optional callback invoked when clicked
	MouseDown           func(MouseEvent)                                                   // Optional callback invoked when mouse is pressed
	MouseUp             func(MouseEvent)                                                   // Optional callback invoked when mouse is released
	Hover               func(bool)                                                         // Optional callback invoked when hover state changes
	Blur                func()                                                             // Optional callback invoked when focus leaves this widget
}

type listItemLayout struct {
	y      int
	height int
}

type listContainer[T any] struct {
	Column
	list List[T]
}

func (c listContainer[T]) Build(ctx BuildContext) Widget {
	return c
}

func (c listContainer[T]) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.list.State == nil {
		return
	}

	count := metrics.ChildCount()
	if count == 0 {
		c.list.State.itemLayouts = nil
		return
	}

	layouts := make([]listItemLayout, count)
	for i := 0; i < count; i++ {
		bounds, ok := metrics.ChildBounds(i)
		if !ok {
			continue
		}
		layouts[i] = listItemLayout{y: bounds.Y, height: bounds.Height}
	}

	c.list.State.itemLayouts = layouts
	c.list.scrollCursorIntoView()
}

func (c listContainer[T]) ChildWidgets() []Widget {
	return c.Children
}

// WidgetID returns the widget's unique identifier.
// Implements the Identifiable interface.
func (l List[T]) WidgetID() string {
	return l.ID
}

// GetContentDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (l List[T]) GetContentDimensions() (width, height Dimension) {
	dims := l.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = l.Width
	}
	if height.IsUnset() {
		height = l.Height
	}
	return width, height
}

// GetStyle returns the style of the list widget.
// Implements the Styled interface.
func (l List[T]) GetStyle() Style {
	return l.Style
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (l List[T]) OnClick(event MouseEvent) {
	if l.Click != nil {
		l.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (l List[T]) OnMouseDown(event MouseEvent) {
	if l.MouseDown != nil {
		l.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (l List[T]) OnMouseUp(event MouseEvent) {
	if l.MouseUp != nil {
		l.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (l List[T]) OnHover(hovered bool) {
	if l.Hover != nil {
		l.Hover(hovered)
	}
}

// OnBlur is called when this widget loses keyboard focus.
// Implements the Blurrable interface.
func (l List[T]) OnBlur() {
	if l.Blur != nil {
		l.Blur()
	}
}

// IsFocusable returns true to allow keyboard navigation.
// Implements the Focusable interface.
func (l List[T]) IsFocusable() bool {
	return !l.DisableFocus
}

// Build returns a Column of widgets, each rendered via RenderItem.
func (l List[T]) Build(ctx BuildContext) Widget {
	if l.State == nil {
		return Column{}
	}

	// Get items (subscribes to changes via signal)
	items := l.State.Items.Get()
	if len(items) == 0 {
		l.State.itemLayouts = nil
		l.State.setViewIndices(nil)
		return Column{}
	}

	query, options := filterStateValues(l.Filter)

	// Check if we have cached filter results for this query
	var filtered FilteredView[T]
	useCached := l.State.cachedFilterQuery == query && l.State.viewIndices != nil
	if useCached {
		if len(l.State.cachedMatches) > 0 && len(l.State.cachedMatches) != len(l.State.viewIndices) {
			useCached = false
		} else {
			for _, idx := range l.State.viewIndices {
				if idx < 0 || idx >= len(items) {
					useCached = false
					break
				}
			}
		}
	}

	if useCached {
		// Use cached results
		filtered = FilteredView[T]{
			Items:   make([]T, len(l.State.viewIndices)),
			Indices: l.State.viewIndices,
			Matches: l.State.cachedMatches,
		}
		for i, idx := range l.State.viewIndices {
			filtered.Items[i] = items[idx]
		}
	} else {
		// Compute filter
		matchItem := l.MatchItem
		if matchItem == nil {
			matchItem = defaultListMatchItem[T]
		}
		filtered = ApplyFilter(items, query, func(item T, q string) MatchResult {
			return matchItem(item, q, options)
		})
		l.State.setViewIndices(filtered.Indices)
		l.State.cachedMatches = filtered.Matches
		l.State.cachedFilterQuery = query
	}

	if len(filtered.Items) == 0 {
		l.State.itemLayouts = nil
		return Column{}
	}

	// Get cursor position (subscribes to changes)
	cursorIdx := l.State.CursorIndex.Get()

	// Get selection state (subscribes to changes)
	var Selection map[int]struct{}
	if l.MultiSelect {
		Selection = l.State.Selection.Get()
	}

	// Clamp cursor to valid bounds for source items
	clamped := clampInt(cursorIdx, 0, len(items)-1)
	if clamped != cursorIdx {
		l.State.CursorIndex.Set(clamped)
		cursorIdx = clamped
	}

	if _, ok := l.State.viewIndexForSource(cursorIdx); !ok {
		cursorIdx = filtered.Indices[0]
		l.State.CursorIndex.Set(cursorIdx)
	}

	// Register scroll callbacks for mouse wheel support
	l.registerScrollCallbacks()

	// Use default render function if none provided
	renderItem := l.RenderItem
	renderItemWithMatch := l.RenderItemWithMatch
	if renderItemWithMatch == nil && renderItem == nil {
		renderItemWithMatch = l.themedDefaultRenderItem(ctx)
	}

	// Build children
	children := make([]Widget, len(filtered.Items))
	for viewIdx, item := range filtered.Items {
		sourceIdx := filtered.Indices[viewIdx]
		_, selected := Selection[sourceIdx]
		active := sourceIdx == cursorIdx
		match := MatchResult{}
		if len(filtered.Matches) > 0 {
			match = filtered.Matches[viewIdx]
		}
		if renderItemWithMatch != nil {
			children[viewIdx] = renderItemWithMatch(item, active, selected, match)
		} else {
			children[viewIdx] = renderItem(item, active, selected)
		}
	}

	// Ensure cursor item is visible whenever we rebuild
	style := l.Style
	if style.Width.IsUnset() {
		style.Width = l.Width
	}
	if style.Height.IsUnset() {
		style.Height = l.Height
	}
	return listContainer[T]{
		Column: Column{
			ID:         l.ID,
			CrossAlign: CrossAxisStretch,
			Style:      style,
			Children:   children,
			Click:      l.Click,
			Hover:      l.Hover,
		},
		list: l,
	}
}

// themedDefaultRenderItem returns a themed render function for list items.
// Captures theme colors and widget focus state from the context for use in the render function.
// Cursor highlighting is only shown when the widget has focus.
func (l List[T]) themedDefaultRenderItem(ctx BuildContext) func(item T, active bool, selected bool, match MatchResult) Widget {
	theme := ctx.Theme()
	widgetFocused := ctx.IsFocused(l)
	cursorPrefix := l.CursorPrefix
	selectedPrefix := l.SelectedPrefix

	highlight := MatchHighlightStyle(theme)
	return func(item T, active bool, selected bool, match MatchResult) Widget {
		content := fmt.Sprintf("%v", item)
		prefix := ""
		style := Style{ForegroundColor: theme.Text}

		// Only show cursor highlight when widget has focus
		showCursor := active && widgetFocused

		if showCursor {
			prefix = cursorPrefix
			style.BackgroundColor = theme.ActiveCursor
			style.ForegroundColor = theme.SelectionText
		}

		// ActiveCursor highlight shown regardless of focus (user's selection persists)
		// Uses Selection for a dimmer appearance than the active cursor
		if selected && !showCursor {
			prefix = selectedPrefix
			style.BackgroundColor = theme.Selection
		}

		if match.Matched && len(match.Ranges) > 0 {
			spans := make([]Span, 0, 1+len(match.Ranges)*2)
			if prefix != "" {
				spans = append(spans, Span{Text: prefix})
			}
			spans = append(spans, HighlightSpans(content, match.Ranges, highlight)...)
			style.Width = Flex(1)
			return Text{
				Spans: spans,
				Style: style,
			}
		}

		style.Width = Flex(1)
		return Text{
			Content: prefix + content,
			Style:   style,
		}
	}
}

func defaultListMatchItem[T any](item T, query string, options FilterOptions) MatchResult {
	return MatchString(fmt.Sprintf("%v", item), query, options)
}

// OnKey handles keys not covered by declarative keybindings.
// Implements the Focusable interface.
func (l List[T]) OnKey(event KeyEvent) bool {
	return false
}

// Keybinds returns the declarative keybindings for this list.
func (l List[T]) Keybinds() []Keybind {
	if l.State == nil {
		return nil
	}
	binds := []Keybind{
		{Key: "enter", Action: l.selectItem, Hidden: true},
		{Key: "up", Action: l.keyCursorUp, Hidden: true},
		{Key: "k", Action: l.keyCursorUp, Hidden: true},
		{Key: "down", Action: l.keyCursorDown, Hidden: true},
		{Key: "j", Action: l.keyCursorDown, Hidden: true},
		{Key: "home", Action: l.keyCursorToFirst, Hidden: true},
		{Key: "g", Action: l.keyCursorToFirst, Hidden: true},
		{Key: "end", Action: l.keyCursorToLast, Hidden: true},
		{Key: "G", Action: l.keyCursorToLast, Hidden: true},
		{Key: "pgup", Action: l.pageUp, Hidden: true},
		{Key: "ctrl+u", Action: l.pageUp, Hidden: true},
		{Key: "pgdown", Action: l.pageDown, Hidden: true},
		{Key: "ctrl+d", Action: l.pageDown, Hidden: true},
	}
	if l.MultiSelect {
		binds = append(binds,
			Keybind{Key: "shift+up", Action: l.shiftCursorUp, Hidden: true},
			Keybind{Key: "shift+k", Action: l.shiftCursorUp, Hidden: true},
			Keybind{Key: "shift+down", Action: l.shiftCursorDown, Hidden: true},
			Keybind{Key: "shift+j", Action: l.shiftCursorDown, Hidden: true},
			Keybind{Key: "shift+home", Action: l.shiftCursorToFirst, Hidden: true},
			Keybind{Key: "shift+end", Action: l.shiftCursorToLast, Hidden: true},
		)
	}
	return binds
}

func (l List[T]) selectItem() {
	if l.OnSelect != nil {
		if item, ok := l.State.SelectedItem(); ok {
			l.OnSelect(item)
		}
	}
}

func (l List[T]) keyCursorUp() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		return
	}
	if cursorViewIdx == 0 {
		return
	}
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(cursorViewIdx - 1)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) keyCursorDown() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		return
	}
	if cursorViewIdx >= len(view)-1 {
		return
	}
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(cursorViewIdx + 1)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) keyCursorToFirst() {
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(0)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) keyCursorToLast() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(len(view) - 1)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) pageUp() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(cursorViewIdx - 10)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) pageDown() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	if l.MultiSelect {
		l.State.ClearSelection()
		l.State.ClearAnchor()
	}
	l.setCursorToViewIndex(cursorViewIdx + 10)
	l.scrollCursorIntoView()
	l.notifyCursorChange()
}

func (l List[T]) shiftCursorUp() {
	l.handleShiftMove(-1)
}

func (l List[T]) shiftCursorDown() {
	l.handleShiftMove(1)
}

func (l List[T]) shiftCursorToFirst() {
	l.handleShiftMoveTo(0)
}

func (l List[T]) shiftCursorToLast() {
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	l.handleShiftMoveTo(len(view) - 1)
}

// handleShiftMove extends selection by moving cursor by delta and selecting the range.
func (l List[T]) handleShiftMove(delta int) {
	if l.State == nil {
		return
	}

	view := l.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		cursorIdx = view[0]
		l.State.CursorIndex.Set(cursorIdx)
		cursorViewIdx = 0
	}

	// Set anchor if not already set
	if !l.State.HasAnchor() {
		l.State.SetAnchor(cursorIdx)
	}

	newViewIdx := clampInt(cursorViewIdx+delta, 0, len(view)-1)
	newCursor := view[newViewIdx]
	l.State.CursorIndex.Set(newCursor)
	l.selectViewRange(l.State.GetAnchor(), newCursor)
	l.scrollCursorIntoView()
}

// handleShiftMoveTo extends selection to a specific index.
func (l List[T]) handleShiftMoveTo(targetIdx int) {
	if l.State == nil {
		return
	}

	view := l.viewIndices()
	if len(view) == 0 {
		return
	}

	cursorIdx := l.State.CursorIndex.Peek()
	if !l.State.HasAnchor() {
		l.State.SetAnchor(cursorIdx)
	}

	targetViewIdx := clampInt(targetIdx, 0, len(view)-1)
	newCursor := view[targetViewIdx]
	l.State.CursorIndex.Set(newCursor)
	l.selectViewRange(l.State.GetAnchor(), newCursor)
	l.scrollCursorIntoView()
}

func (l List[T]) setCursorToViewIndex(viewIdx int) {
	if l.State == nil {
		return
	}
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	viewIdx = clampInt(viewIdx, 0, len(view)-1)
	l.State.SelectIndex(view[viewIdx])
}

func (l List[T]) selectViewRange(anchorSource, cursorSource int) {
	if l.State == nil {
		return
	}
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}

	anchorView, ok := l.viewIndexForSource(anchorSource)
	if !ok {
		anchorView = 0
	}
	cursorView, ok := l.viewIndexForSource(cursorSource)
	if !ok {
		cursorView = anchorView
	}

	if anchorView > cursorView {
		anchorView, cursorView = cursorView, anchorView
	}

	sel := make(map[int]struct{}, cursorView-anchorView+1)
	for i := anchorView; i <= cursorView; i++ {
		sel[view[i]] = struct{}{}
	}
	l.State.Selection.Set(sel)
}

// scrollCursorIntoView uses the ScrollState to ensure
// the cursor item is visible in the viewport.
func (l List[T]) scrollCursorIntoView() {
	if l.ScrollState == nil || l.State == nil {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	viewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		return
	}
	itemY, itemHeight, ok := l.getItemLayout(cursorIdx)
	if !ok {
		itemHeight = l.getItemHeight()
		itemY = viewIdx * itemHeight
	}
	l.ScrollState.ScrollToView(itemY, itemHeight)
}

// getItemHeight returns the fallback uniform height of list items.
func (l List[T]) getItemHeight() int {
	if l.ItemHeight > 0 {
		return l.ItemHeight
	}
	return 1
}

// getItemLayout returns the cached item layout for the given index.
func (l List[T]) getItemLayout(index int) (y, height int, ok bool) {
	if l.State == nil {
		return 0, 0, false
	}
	viewIdx, ok := l.viewIndexForSource(index)
	if !ok {
		return 0, 0, false
	}
	if viewIdx < 0 || viewIdx >= len(l.State.itemLayouts) {
		return 0, 0, false
	}
	layout := l.State.itemLayouts[viewIdx]
	if layout.height <= 0 {
		return 0, 0, false
	}
	return layout.y, layout.height, true
}

// registerScrollCallbacks sets up callbacks on the ScrollState
// to update cursor position when mouse wheel scrolling occurs.
// The callbacks move cursor first, then scroll only if needed.
func (l List[T]) registerScrollCallbacks() {
	if l.ScrollState == nil {
		return
	}

	l.ScrollState.OnScrollUp = func(lines int) bool {
		if l.State == nil {
			return false
		}
		before := l.State.CursorIndex.Peek()
		l.moveCursorUp(lines)
		after := l.State.CursorIndex.Peek()
		if after == before {
			return false
		}
		l.scrollCursorIntoView()
		l.notifyCursorChange()
		return true // We handled scrolling via cursor movement
	}
	l.ScrollState.OnScrollDown = func(lines int) bool {
		if l.State == nil {
			return false
		}
		before := l.State.CursorIndex.Peek()
		l.moveCursorDown(lines)
		after := l.State.CursorIndex.Peek()
		if after == before {
			return false
		}
		l.scrollCursorIntoView()
		l.notifyCursorChange()
		return true // We handled scrolling via cursor movement
	}
}

// moveCursorUp moves the cursor up by the given number of items.
func (l List[T]) moveCursorUp(count int) {
	if l.State == nil || l.State.ItemCount() == 0 {
		return
	}
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	newCursor := clampInt(cursorViewIdx-count, 0, len(view)-1)
	l.State.SelectIndex(view[newCursor])
}

// moveCursorDown moves the cursor down by the given number of items.
func (l List[T]) moveCursorDown(count int) {
	if l.State == nil || l.State.ItemCount() == 0 {
		return
	}
	view := l.viewIndices()
	if len(view) == 0 {
		return
	}
	cursorIdx := l.State.CursorIndex.Peek()
	cursorViewIdx, ok := l.viewIndexForSource(cursorIdx)
	if !ok {
		cursorViewIdx = 0
	}
	newCursor := clampInt(cursorViewIdx+count, 0, len(view)-1)
	l.State.SelectIndex(view[newCursor])
}

// notifyCursorChange calls OnCursorChange with the current item if the callback is set.
func (l List[T]) notifyCursorChange() {
	if l.OnCursorChange == nil || l.State == nil {
		return
	}
	if item, ok := l.State.SelectedItem(); ok {
		l.OnCursorChange(item)
	}
}

func (l List[T]) viewIndices() []int {
	if l.State == nil {
		return nil
	}
	if l.State.viewIndices != nil {
		return l.State.viewIndices
	}
	count := l.State.ItemCount()
	indices := make([]int, count)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

func (l List[T]) viewIndexForSource(sourceIdx int) (int, bool) {
	if l.State == nil {
		return 0, false
	}
	if l.State.viewIndices == nil {
		if sourceIdx >= 0 && sourceIdx < l.State.ItemCount() {
			return sourceIdx, true
		}
		return 0, false
	}
	return l.State.viewIndexForSource(sourceIdx)
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
