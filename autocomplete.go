package terma

import (
	"unicode"
	"unicode/utf8"
)

// Suggestion represents an autocomplete option.
type Suggestion struct {
	Label       string // Display text in dropdown
	Value       string // Value to insert when selected
	Description string // Optional secondary text (dimmed)
	Icon        string // Optional leading icon
	Data        any    // User data for custom handling
}

// InsertStrategy defines how a suggestion is inserted into the text.
// It receives the current text, cursor position, the selected suggestion,
// and the trigger position (or -1 if no trigger), and returns the new text
// and new cursor position.
type InsertStrategy func(text string, cursor int, suggestion Suggestion, triggerPos int) (newText string, newCursor int)

// InsertReplace replaces the entire text with the suggestion value.
var InsertReplace InsertStrategy = func(text string, cursor int, suggestion Suggestion, triggerPos int) (string, int) {
	value := suggestion.Value
	if value == "" {
		value = suggestion.Label
	}
	return value, utf8.RuneCountInString(value)
}

// InsertFromTrigger replaces text from the trigger position to cursor with the suggestion.
// If no trigger position, replaces from start.
var InsertFromTrigger InsertStrategy = func(text string, cursor int, suggestion Suggestion, triggerPos int) (string, int) {
	value := suggestion.Value
	if value == "" {
		value = suggestion.Label
	}

	runes := []rune(text)
	if triggerPos < 0 {
		triggerPos = 0
	}
	if triggerPos > len(runes) {
		triggerPos = len(runes)
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}

	// Build new text: before trigger + value + after cursor
	newRunes := append([]rune{}, runes[:triggerPos]...)
	newRunes = append(newRunes, []rune(value)...)
	newRunes = append(newRunes, runes[cursor:]...)

	return string(newRunes), triggerPos + utf8.RuneCountInString(value)
}

// InsertAtCursor inserts the suggestion value at the cursor position.
var InsertAtCursor InsertStrategy = func(text string, cursor int, suggestion Suggestion, triggerPos int) (string, int) {
	value := suggestion.Value
	if value == "" {
		value = suggestion.Label
	}

	runes := []rune(text)
	if cursor > len(runes) {
		cursor = len(runes)
	}
	if cursor < 0 {
		cursor = 0
	}

	newRunes := append([]rune{}, runes[:cursor]...)
	newRunes = append(newRunes, []rune(value)...)
	newRunes = append(newRunes, runes[cursor:]...)

	return string(newRunes), cursor + utf8.RuneCountInString(value)
}

// InsertReplaceWord replaces the current word (delimited by whitespace) with the suggestion.
var InsertReplaceWord InsertStrategy = func(text string, cursor int, suggestion Suggestion, triggerPos int) (string, int) {
	value := suggestion.Value
	if value == "" {
		value = suggestion.Label
	}

	runes := []rune(text)
	if cursor > len(runes) {
		cursor = len(runes)
	}
	if cursor < 0 {
		cursor = 0
	}

	// Find word start (scan backward to whitespace)
	wordStart := cursor
	for wordStart > 0 && !unicode.IsSpace(runes[wordStart-1]) {
		wordStart--
	}

	// Find word end (scan forward to whitespace)
	wordEnd := cursor
	for wordEnd < len(runes) && !unicode.IsSpace(runes[wordEnd]) {
		wordEnd++
	}

	// Build new text: before word + value + after word
	newRunes := append([]rune{}, runes[:wordStart]...)
	newRunes = append(newRunes, []rune(value)...)
	newRunes = append(newRunes, runes[wordEnd:]...)

	return string(newRunes), wordStart + utf8.RuneCountInString(value)
}

// AutocompleteState holds the state for an Autocomplete widget.
type AutocompleteState struct {
	Visible         Signal[bool]
	Suggestions     AnySignal[[]Suggestion]
	listState       *ListState[Suggestion]
	scrollState     *ScrollState
	filterState     *FilterState
	triggerPosition Signal[int]    // Where trigger char was typed (-1 if none)
	filterQuery     Signal[string] // Text after trigger (for filtering)
	dismissed       bool           // Tracks manual dismissal (e.g. Escape) until query changes
	anchorWidth     Signal[int]    // Border-box width of the input for anchored popups
}

// NewAutocompleteState creates a new AutocompleteState.
func NewAutocompleteState() *AutocompleteState {
	return &AutocompleteState{
		Visible:         NewSignal(false),
		Suggestions:     NewAnySignal([]Suggestion(nil)),
		listState:       NewListState([]Suggestion{}),
		scrollState:     NewScrollState(),
		filterState:     NewFilterState(),
		triggerPosition: NewSignal(-1),
		filterQuery:     NewSignal(""),
		anchorWidth:     NewSignal(0),
	}
}

// SetSuggestions sets the available suggestions.
func (s *AutocompleteState) SetSuggestions(suggestions []Suggestion) {
	s.Suggestions.Set(suggestions)
	s.listState.SetItems(suggestions)
}

// Show makes the popup visible.
func (s *AutocompleteState) Show() {
	s.dismissed = false
	s.Visible.Set(true)
}

// Hide hides the popup.
func (s *AutocompleteState) Hide() {
	s.dismissed = true
	s.Visible.Set(false)
}

// IsVisible returns whether the popup is currently visible.
func (s *AutocompleteState) IsVisible() bool {
	return s.Visible.Peek()
}

// SelectedSuggestion returns the currently selected suggestion, if any.
func (s *AutocompleteState) SelectedSuggestion() (Suggestion, bool) {
	return s.listState.SelectedItem()
}

// Autocomplete is a widget that wraps TextInput or TextArea to provide
// dropdown suggestions. The input keeps focus while navigating the popup.
type Autocomplete struct {
	ID    string             // Optional unique identifier
	State *AutocompleteState // Required - holds popup and suggestion state
	Child Widget             // TextInput or TextArea

	// Trigger behavior
	TriggerChars          []rune // e.g., {'@', '#'} - empty = always on
	TriggerAtWordBoundary bool   // Only trigger at word start (default: true)
	MinChars              int    // Min chars after trigger to show popup (default 0)

	// Selection & matching
	MaxVisible int            // Max visible items (default 8)
	Insert     InsertStrategy // Default: InsertFromTrigger if TriggerChars set, else InsertReplace
	MatchMode  FilterMode     // FilterContains (default) or FilterFuzzy

	// Dismissal behavior
	DismissOnBlur    *bool // Dismiss when input loses focus (default: true)
	DismissWhenEmpty bool  // Dismiss when no matches (default: false)
	// If true, only attach popup keybinds when the popup is visible.
	// This allows parent widgets to handle keys like Up/Down when hidden.
	DisableKeysWhenHidden bool

	// Callbacks
	OnSelect      func(Suggestion) // Called when a suggestion is selected
	OnDismiss     func()           // Called when popup is dismissed
	OnQueryChange func(string)     // For async loading - called when query changes

	// Dimensions
	Width         Dimension // Widget width
	Height        Dimension // Widget height
	PopupWidth    Dimension // Popup width (default: fit content)
	AnchorToInput bool      // Anchor popup to input content area and match its width

	// Styling
	Style      Style // Widget styling
	PopupStyle Style // Popup styling

	// Custom rendering
	RenderSuggestion func(Suggestion, bool, MatchResult, BuildContext) Widget
}

type autocompleteContainer struct {
	Column
	state         *AutocompleteState
	contentInsets EdgeInsets
}

func (c autocompleteContainer) Build(ctx BuildContext) Widget {
	return c
}

func (c autocompleteContainer) ChildWidgets() []Widget {
	return c.Children
}

func (c autocompleteContainer) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.state == nil {
		return
	}
	bounds, ok := metrics.ChildBounds(0)
	if !ok {
		return
	}
	contentWidth := bounds.Width - c.contentInsets.Horizontal()
	if contentWidth < 1 {
		contentWidth = 1
	}
	c.state.anchorWidth.Set(contentWidth)
}

func (a Autocomplete) anchorContentInsets() EdgeInsets {
	var style Style
	switch child := a.Child.(type) {
	case TextInput:
		style = child.Style
	case TextArea:
		style = child.Style
	default:
		if styled, ok := a.Child.(Styled); ok {
			style = styled.GetStyle()
		} else {
			return EdgeInsets{}
		}
	}
	border := style.Border.Width()
	padding := style.Padding
	return EdgeInsets{
		Top:    padding.Top + border,
		Right:  padding.Right + border,
		Bottom: padding.Bottom + border,
		Left:   padding.Left + border,
	}
}

// WidgetID returns the autocomplete's unique identifier.
func (a Autocomplete) WidgetID() string {
	return a.ID
}

// GetContentDimensions returns the width and height dimension preferences.
func (a Autocomplete) GetContentDimensions() (width, height Dimension) {
	return a.Width, a.Height
}

// GetStyle returns the style.
func (a Autocomplete) GetStyle() Style {
	return a.Style
}

// Build builds the autocomplete widget with popup overlay.
func (a Autocomplete) Build(ctx BuildContext) Widget {
	if a.State == nil || a.Child == nil {
		return a.Child
	}

	// Get child text and cursor for trigger detection
	text, cursorPos := a.getChildTextAndCursor()

	// Update trigger and query based on current text/cursor
	a.updateTriggerAndQuery(text, cursorPos)

	// Apply filter to get filtered count (List.Build will reuse cached results)
	hasItems := a.filteredSuggestionCount() > 0

	// Determine visibility
	visible := a.State.Visible.Get()

	// Auto-dismiss when empty if configured
	if a.DismissWhenEmpty && visible && !hasItems {
		a.State.Visible.Set(false)
		visible = false
	}
	if a.dismissOnBlurEnabled() && !a.isChildFocused(ctx) {
		if visible {
			a.State.Visible.Set(false)
		}
		visible = false
	}

	// Wrap child to inject keybinds and intercept changes
	enablePopupKeys := true
	if a.DisableKeysWhenHidden && (!visible || !hasItems) {
		enablePopupKeys = false
	}
	wrappedChild := a.wrapChild(ctx, enablePopupKeys)

	// Build popup
	popup := a.buildPopup(ctx, visible && hasItems)

	column := Column{
		ID:     a.ID,
		Width:  a.Width,
		Height: a.Height,
		Style:  a.Style,
		Children: []Widget{
			wrappedChild,
			popup,
		},
	}

	if a.AnchorToInput {
		return autocompleteContainer{
			Column:        column,
			state:         a.State,
			contentInsets: a.anchorContentInsets(),
		}
	}

	return column
}

// wrapChild wraps the child widget with autocomplete keybinds and change handlers.
func (a Autocomplete) wrapChild(ctx BuildContext, enablePopupKeys bool) Widget {
	switch child := a.Child.(type) {
	case TextInput:
		return a.wrapTextInput(child, ctx, enablePopupKeys)
	case TextArea:
		return a.wrapTextArea(child, ctx, enablePopupKeys)
	default:
		return a.Child
	}
}

// wrapTextInput wraps a TextInput with autocomplete behavior.
func (a Autocomplete) wrapTextInput(child TextInput, ctx BuildContext, enablePopupKeys bool) Widget {
	if enablePopupKeys {
		// Inject extra keybinds for popup navigation
		extraKeybinds := a.popupKeybindsForTextInput()
		child.ExtraKeybinds = append(extraKeybinds, child.ExtraKeybinds...)
	}

	// Wrap OnChange to process text changes
	originalOnChange := child.OnChange
	child.OnChange = func(text string) {
		cursorPos := 0
		if child.State != nil {
			cursorPos = child.State.CursorIndex.Peek()
		}
		a.handleTextChange(text, cursorPos)
		if originalOnChange != nil {
			originalOnChange(text)
		}
	}

	originalOnBlur := child.Blur
	child.Blur = func() {
		if a.dismissOnBlurEnabled() {
			a.dismissOnBlur()
		}
		if originalOnBlur != nil {
			originalOnBlur()
		}
	}

	return child
}

// wrapTextArea wraps a TextArea with autocomplete behavior.
func (a Autocomplete) wrapTextArea(child TextArea, ctx BuildContext, enablePopupKeys bool) Widget {
	if enablePopupKeys {
		// Inject extra keybinds for popup navigation
		extraKeybinds := a.popupKeybindsForTextArea()
		child.ExtraKeybinds = append(extraKeybinds, child.ExtraKeybinds...)
	}

	// Wrap OnChange to process text changes
	originalOnChange := child.OnChange
	child.OnChange = func(text string) {
		cursorPos := 0
		if child.State != nil {
			cursorPos = child.State.CursorIndex.Peek()
		}
		a.handleTextChange(text, cursorPos)
		if originalOnChange != nil {
			originalOnChange(text)
		}
	}

	originalOnBlur := child.Blur
	child.Blur = func() {
		if a.dismissOnBlurEnabled() {
			a.dismissOnBlur()
		}
		if originalOnBlur != nil {
			originalOnBlur()
		}
	}

	return child
}

// popupKeybindsForTextInput returns keybinds for TextInput popup navigation.
func (a Autocomplete) popupKeybindsForTextInput() []Keybind {
	return []Keybind{
		{Key: "down", Action: a.onDown, Hidden: true},
		{Key: "up", Action: a.onUp, Hidden: true},
		{Key: "escape", Action: a.onEscape, Hidden: true},
		{Key: "enter", Action: a.onEnterTextInput, Hidden: true},
		{Key: "tab", Action: a.onTabTextInput, Hidden: true},
	}
}

// popupKeybindsForTextArea returns keybinds for TextArea popup navigation.
func (a Autocomplete) popupKeybindsForTextArea() []Keybind {
	return []Keybind{
		{Key: "down", Action: a.onDownTextArea, Hidden: true},
		{Key: "up", Action: a.onUpTextArea, Hidden: true},
		{Key: "escape", Action: a.onEscape, Hidden: true},
		{Key: "enter", Action: a.onEnterTextArea, Hidden: true},
		{Key: "tab", Action: a.onTabTextArea, Hidden: true},
	}
}

// Keybind action handlers

func (a Autocomplete) onDown() {
	if a.State == nil {
		return
	}
	if !a.State.Visible.Peek() {
		// Popup hidden - show it if we have suggestions
		if a.filteredSuggestionCount() > 0 {
			a.State.dismissed = false
			a.State.Visible.Set(true)
		}
		return
	}
	if a.filteredSuggestionCount() == 0 {
		return
	}
	a.State.listState.SelectNext()
	a.scrollCursorIntoView()
}

func (a Autocomplete) onUp() {
	if a.State == nil {
		return
	}
	if !a.State.Visible.Peek() {
		return
	}
	if a.filteredSuggestionCount() == 0 {
		return
	}
	a.State.listState.SelectPrevious()
	a.scrollCursorIntoView()
}

func (a Autocomplete) onDownTextArea() {
	if a.State == nil {
		return
	}
	if !a.State.Visible.Peek() {
		// Popup hidden - invoke TextArea's default cursor movement
		if state := a.textAreaState(); state != nil {
			state.CursorDown()
		}
		return
	}
	if a.filteredSuggestionCount() == 0 {
		if state := a.textAreaState(); state != nil {
			state.CursorDown()
		}
		return
	}
	a.State.listState.SelectNext()
	a.scrollCursorIntoView()
}

func (a Autocomplete) onUpTextArea() {
	if a.State == nil {
		return
	}
	if !a.State.Visible.Peek() {
		// Popup hidden - invoke TextArea's default cursor movement
		if state := a.textAreaState(); state != nil {
			state.CursorUp()
		}
		return
	}
	if a.filteredSuggestionCount() == 0 {
		if state := a.textAreaState(); state != nil {
			state.CursorUp()
		}
		return
	}
	a.State.listState.SelectPrevious()
	a.scrollCursorIntoView()
}

func (a Autocomplete) onEscape() {
	if a.State == nil {
		return
	}
	if a.State.Visible.Peek() {
		a.dismiss()
	}
}

func (a Autocomplete) onEnterTextInput() {
	if a.State == nil {
		return
	}
	if a.State.Visible.Peek() && a.filteredSuggestionCount() > 0 {
		a.selectCurrentSuggestion()
		return
	}
	if ti, ok := a.Child.(TextInput); ok && ti.OnSubmit != nil && ti.State != nil {
		ti.OnSubmit(ti.State.GetText())
	}
}

func (a Autocomplete) onTabTextInput() {
	if a.State == nil {
		return
	}
	if a.State.Visible.Peek() && a.filteredSuggestionCount() > 0 {
		a.selectCurrentSuggestion()
	}
	// If popup not visible, don't consume - let focus manager handle tab
}

func (a Autocomplete) onEnterTextArea() {
	if a.State == nil {
		return
	}
	if a.State.Visible.Peek() && a.filteredSuggestionCount() > 0 {
		a.selectCurrentSuggestion()
		return
	}
	if ta, ok := a.Child.(TextArea); ok && ta.State != nil && ta.canInsert() {
		ta.State.InsertNewline()
		if ta.OnChange != nil {
			ta.OnChange(ta.State.GetText())
		}
	}
}

func (a Autocomplete) onTabTextArea() {
	if a.State == nil {
		return
	}
	if a.State.Visible.Peek() && a.filteredSuggestionCount() > 0 {
		a.selectCurrentSuggestion()
	}
	// If popup not visible, let TextArea handle it (insert tab or focus)
}

// handleTextChange processes text changes to update trigger and visibility.
func (a Autocomplete) handleTextChange(text string, cursorPos int) {
	if a.State != nil {
		a.State.dismissed = false
	}
	a.updateTriggerAndQuery(text, cursorPos)
}

// updateTriggerAndQuery detects trigger characters and extracts the query.
func (a Autocomplete) updateTriggerAndQuery(text string, cursorPos int) {
	if a.State == nil {
		return
	}

	// Find trigger position by searching backwards from cursor
	triggerPos := a.findTriggerPosition(text, cursorPos)
	query := a.extractQuery(text, cursorPos, triggerPos)

	a.State.triggerPosition.Set(triggerPos)
	a.State.filterQuery.Set(query)

	// Determine if we should show the popup
	queryRuneCount := utf8.RuneCountInString(query)
	shouldShow := (len(a.TriggerChars) == 0 || triggerPos >= 0) && queryRuneCount >= a.MinChars

	// Also need text input to have something (unless MinChars is 0)
	if len(a.TriggerChars) == 0 && text == "" && a.MinChars == 0 {
		// Always-on mode with empty input - still show suggestions
		shouldShow = true
	}

	if a.State.dismissed && shouldShow {
		shouldShow = false
	}
	a.State.Visible.Set(shouldShow)

	if a.OnQueryChange != nil && shouldShow {
		a.OnQueryChange(query)
	}
}

// findTriggerPosition searches backwards from cursor to find a trigger character.
func (a Autocomplete) findTriggerPosition(text string, cursorPos int) int {
	if len(a.TriggerChars) == 0 {
		return -1 // No triggers means always-on mode
	}

	runes := []rune(text)
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	// Search backwards from cursor
	for i := cursorPos - 1; i >= 0; i-- {
		r := runes[i]

		// Stop at whitespace if looking for word-boundary triggers
		triggerAtWordBoundary := a.TriggerAtWordBoundary
		// Default to true if not explicitly set (zero value bool is false)
		if !a.TriggerAtWordBoundary && len(a.TriggerChars) > 0 {
			// Check if this is actually the default (field not set) vs explicitly false
			// Since we can't distinguish, we'll default to true behavior for triggers
			triggerAtWordBoundary = true
		}

		if triggerAtWordBoundary && unicode.IsSpace(r) {
			break
		}

		// Check if this is a trigger char
		if a.isTriggerChar(r) {
			if !triggerAtWordBoundary {
				return i
			}
			// Word boundary check: must be at start or preceded by whitespace
			if i == 0 || unicode.IsSpace(runes[i-1]) {
				return i
			}
		}
	}

	return -1
}

// isTriggerChar checks if a rune is in the trigger character list.
func (a Autocomplete) isTriggerChar(r rune) bool {
	for _, trigger := range a.TriggerChars {
		if r == trigger {
			return true
		}
	}
	return false
}

// extractQuery extracts the text between trigger and cursor.
func (a Autocomplete) extractQuery(text string, cursorPos int, triggerPos int) string {
	runes := []rune(text)
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	if triggerPos < 0 {
		// No trigger - in always-on mode, the entire text up to cursor is the query
		if len(a.TriggerChars) == 0 {
			return string(runes[:cursorPos])
		}
		return ""
	}

	// Extract text after trigger (excluding trigger char itself)
	queryStart := triggerPos + 1
	if queryStart > cursorPos {
		return ""
	}
	return string(runes[queryStart:cursorPos])
}

// matchMode returns the configured match mode.
func (a Autocomplete) matchMode() FilterMode {
	return a.MatchMode // defaults to FilterContains (0)
}

func (a Autocomplete) filteredSuggestionCount() int {
	if a.State == nil {
		return 0
	}
	a.State.filterState.Query.Set(a.State.filterQuery.Peek())
	a.State.filterState.Mode.Set(a.matchMode())
	return a.State.listState.ApplyFilter(a.State.filterState, suggestionMatchItem)
}

func (a Autocomplete) dismissOnBlurEnabled() bool {
	if a.DismissOnBlur != nil {
		return *a.DismissOnBlur
	}
	return true
}

// selectCurrentSuggestion selects the currently highlighted suggestion.
func (a Autocomplete) selectCurrentSuggestion() {
	if a.State == nil {
		return
	}

	suggestion, ok := a.State.listState.SelectedItem()
	if !ok {
		return
	}

	a.selectSuggestion(suggestion)
}

// selectSuggestion applies the selected suggestion to the input.
func (a Autocomplete) selectSuggestion(suggestion Suggestion) {
	text, cursor := a.getChildTextAndCursor()

	strategy := a.Insert
	if strategy == nil {
		if len(a.TriggerChars) > 0 {
			strategy = InsertFromTrigger
		} else {
			strategy = InsertReplace
		}
	}

	triggerPos := -1
	if a.State != nil {
		triggerPos = a.State.triggerPosition.Peek()
	}

	newText, newCursor := strategy(text, cursor, suggestion, triggerPos)
	a.setChildTextAndCursor(newText, newCursor)
	a.dismiss()

	if a.OnSelect != nil {
		a.OnSelect(suggestion)
	}
}

// dismiss hides the popup and calls OnDismiss if set.
func (a Autocomplete) dismiss() {
	if a.State != nil {
		a.State.dismissed = true
		a.State.Visible.Set(false)
	}
	if a.OnDismiss != nil {
		a.OnDismiss()
	}
}

// dismissOnBlur hides the popup without marking it as manually dismissed.
func (a Autocomplete) dismissOnBlur() {
	if a.State != nil && a.State.Visible.Peek() {
		a.State.Visible.Set(false)
		if a.OnDismiss != nil {
			a.OnDismiss()
		}
	}
}

// getChildTextAndCursor returns the child's text content and cursor position.
func (a Autocomplete) getChildTextAndCursor() (string, int) {
	switch child := a.Child.(type) {
	case TextInput:
		if child.State != nil {
			return child.State.GetText(), child.State.CursorIndex.Peek()
		}
	case TextArea:
		if child.State != nil {
			return child.State.GetText(), child.State.CursorIndex.Peek()
		}
	}
	return "", 0
}

// setChildTextAndCursor updates the child's text and cursor position.
func (a Autocomplete) setChildTextAndCursor(text string, cursor int) {
	switch child := a.Child.(type) {
	case TextInput:
		if child.State != nil {
			child.State.SetText(text)
			child.State.CursorIndex.Set(cursor)
		}
	case TextArea:
		if child.State != nil {
			child.State.SetText(text)
			child.State.CursorIndex.Set(cursor)
		}
	}
}

// textAreaState returns the TextArea state if child is a TextArea.
func (a Autocomplete) textAreaState() *TextAreaState {
	if ta, ok := a.Child.(TextArea); ok {
		return ta.State
	}
	return nil
}

// textInputState returns the TextInput state if child is a TextInput.
func (a Autocomplete) textInputState() *TextInputState {
	if ti, ok := a.Child.(TextInput); ok {
		return ti.State
	}
	return nil
}

// scrollCursorIntoView ensures the selected item is visible.
func (a Autocomplete) scrollCursorIntoView() {
	if a.State == nil || a.State.scrollState == nil {
		return
	}
	cursorIdx := a.State.listState.CursorIndex.Peek()
	// Simple approach: scroll to show item at index
	a.State.scrollState.ScrollToView(cursorIdx, 1)
}

// buildPopup builds the floating popup with suggestions list.
func (a Autocomplete) buildPopup(ctx BuildContext, visible bool) Widget {
	if a.State == nil {
		return EmptyWidget{}
	}

	maxVisible := a.MaxVisible
	if maxVisible <= 0 {
		maxVisible = 8
	}

	// Get anchor ID from child
	anchorID := a.getChildID()

	floatConfig := a.buildFloatConfig(anchorID)

	// Create the suggestion list
	list := List[Suggestion]{
		ID:          a.ID + "-list",
		State:       a.State.listState,
		ScrollState: a.State.scrollState,
		Filter:      a.State.filterState,
		MatchItem:   suggestionMatchItem,
		OnSelect:    a.selectSuggestion,
		RenderItemWithMatch: func(item Suggestion, active bool, selected bool, match MatchResult) Widget {
			if a.RenderSuggestion != nil {
				return a.RenderSuggestion(item, active, match, ctx)
			}
			return a.defaultRenderSuggestion(item, active, match, ctx)
		},
	}

	popupStyle := a.PopupStyle
	if popupStyle.BackgroundColor == nil {
		popupStyle.BackgroundColor = ctx.Theme().Surface
	}
	if popupStyle.MaxHeight.IsUnset() {
		popupStyle.MaxHeight = Cells(maxVisible)
	}
	popupWidth := a.PopupWidth
	if a.AnchorToInput && a.State != nil {
		anchorWidth := a.State.anchorWidth.Get()
		if anchorWidth > 0 {
			popupWidth = Cells(anchorWidth)
		}
	}

	return Floating{
		Visible: visible,
		Config:  floatConfig,
		Child: Scrollable{
			State: a.State.scrollState,
			Width: popupWidth,
			Style: popupStyle,
			Child: list,
		},
	}
}

// buildFloatConfig creates the floating configuration for the popup.
func (a Autocomplete) buildFloatConfig(anchorID string) FloatConfig {
	config := FloatConfig{
		OnDismiss:             a.dismiss,
		DismissOnClickOutside: BoolPtr(true),
		DismissOnEsc:          BoolPtr(true),
	}

	if anchorID != "" {
		config.AnchorID = anchorID
		config.Anchor = AnchorBottomLeft
	}

	if a.AnchorToInput {
		inset := a.anchorContentInsets()
		config.Offset = Offset{
			X: inset.Left,
			Y: -inset.Bottom,
		}
	} else if offset, ok := a.cursorPopupOffset(); ok {
		config.Offset = offset
	}

	return config
}

// getChildID returns the ID of the child widget.
func (a Autocomplete) getChildID() string {
	switch child := a.Child.(type) {
	case TextInput:
		return child.ID
	case TextArea:
		return child.ID
	}
	return ""
}

func (a Autocomplete) isChildFocused(ctx BuildContext) bool {
	childID := a.getChildID()
	if childID == "" {
		return false
	}
	focused := ctx.Focused()
	if focused == nil {
		return false
	}
	if identifiable, ok := focused.(Identifiable); ok {
		return identifiable.WidgetID() == childID
	}
	return false
}

// defaultRenderSuggestion renders a suggestion with the default style.
func (a Autocomplete) defaultRenderSuggestion(item Suggestion, active bool, match MatchResult, ctx BuildContext) Widget {
	theme := ctx.Theme()
	widgetFocused := a.isChildFocused(ctx)
	showCursor := active && widgetFocused
	textColor := theme.Text
	if showCursor {
		textColor = theme.SelectionText
	}

	style := Style{
		Padding: EdgeInsets{Left: 1, Right: 1},
	}
	if showCursor {
		style.BackgroundColor = theme.ActiveCursor
	} else {
		style.ForegroundColor = theme.Text
	}

	// Build content with optional icon and description
	var children []Widget

	if item.Icon != "" {
		children = append(children, Text{
			Content: item.Icon + " ",
			Style:   Style{ForegroundColor: textColor},
		})
	}

	// Main label with match highlighting
	if match.Matched && len(match.Ranges) > 0 {
		children = append(children, Text{
			Spans: HighlightSpans(item.Label, match.Ranges, MatchHighlightStyle(theme)),
			Style: Style{ForegroundColor: textColor},
		})
	} else {
		children = append(children, Text{
			Content: item.Label,
			Style:   Style{ForegroundColor: textColor},
		})
	}

	// Add description if present
	if item.Description != "" {
		children = append(children, Spacer{Width: Flex(1)})
		descColor := theme.TextMuted
		if showCursor {
			descColor = theme.SelectionText
		}
		children = append(children, Text{
			Content: item.Description,
			Style:   Style{ForegroundColor: descColor},
		})
	}

	return Row{
		Style:    style,
		Children: children,
	}
}

// suggestionMatchItem matches a Suggestion's Label against the query.
func suggestionMatchItem(item Suggestion, query string, options FilterOptions) MatchResult {
	return MatchString(item.Label, query, options)
}

// cursorPopupOffset returns the popup offset to align with the cursor position.
func (a Autocomplete) cursorPopupOffset() (Offset, bool) {
	switch child := a.Child.(type) {
	case TextInput:
		if child.State == nil {
			return Offset{}, false
		}
		cursorX := child.State.cursorDisplayX() - child.State.scrollOffset
		padding := child.Style.Padding
		border := child.Style.Border.Width()
		contentHeight := 1
		return Offset{
			X: cursorX + padding.Left + border,
			Y: 1 - contentHeight - padding.Bottom - border,
		}, true
	case TextArea:
		if child.State == nil {
			return Offset{}, false
		}
		cursorX, cursorY := child.State.CursorScreenPosition(0, 0)
		padding := child.Style.Padding
		border := child.Style.Border.Width()
		contentHeight := child.State.lastHeight
		if contentHeight <= 0 {
			contentHeight = textAreaContentHeight(child)
		}
		return Offset{
			X: cursorX + padding.Left + border,
			Y: cursorY + 1 - contentHeight - padding.Bottom - border,
		}, true
	default:
		return Offset{}, false
	}
}

func textAreaContentHeight(area TextArea) int {
	heightDim := area.Style.Height
	if heightDim.IsUnset() {
		heightDim = area.Height
	}
	if heightDim.IsCells() {
		return heightDim.CellsValue()
	}
	return 1
}
