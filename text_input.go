package terma

import (
	"strings"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"terma/layout"
)

// --- Grapheme Helper Functions (reusable for multi-line) ---

// splitGraphemes splits a string into grapheme clusters.
func splitGraphemes(s string) []string {
	if s == "" {
		return nil
	}
	var graphemes []string
	remaining := s
	for len(remaining) > 0 {
		grapheme, _ := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		if grapheme == "" {
			break
		}
		graphemes = append(graphemes, grapheme)
		remaining = remaining[len(grapheme):]
	}
	return graphemes
}

// joinGraphemes joins grapheme clusters into a string.
func joinGraphemes(graphemes []string) string {
	return strings.Join(graphemes, "")
}

// graphemeWidth returns the display width of a grapheme cluster.
func graphemeWidth(s string) int {
	_, width := ansi.FirstGraphemeCluster(s, ansi.GraphemeWidth)
	return width
}

// isWordChar returns true if the grapheme is a word character (letter, digit, underscore).
func isWordChar(g string) bool {
	for _, r := range g {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			return true
		}
	}
	return false
}

// --- TextInputState ---

// TextInputState holds the state for a TextInput widget.
// It is the source of truth for text content and cursor position,
// and must be provided to TextInput.
type TextInputState struct {
	Content         AnySignal[[]string] // Grapheme clusters for Unicode safety
	CursorIndex     Signal[int]         // Grapheme index (0 = before first char)
	SelectionAnchor Signal[int]         // -1 = no selection, else anchor grapheme index

	// scrollOffset is calculated during render to keep cursor visible.
	// Not a signal because it's derived state, not source of truth.
	scrollOffset int
}

// NewTextInputState creates a new TextInputState with optional initial text.
func NewTextInputState(initial string) *TextInputState {
	graphemes := splitGraphemes(initial)
	return &TextInputState{
		Content:         NewAnySignal(graphemes),
		CursorIndex:     NewSignal(len(graphemes)), // Cursor at end
		SelectionAnchor: NewSignal(-1),
	}
}

// GetText returns the content as a string.
func (s *TextInputState) GetText() string {
	return joinGraphemes(s.Content.Peek())
}

// SetText replaces the content and clamps the cursor.
func (s *TextInputState) SetText(text string) {
	graphemes := splitGraphemes(text)
	s.Content.Set(graphemes)
	s.clampCursor()
}

// Insert inserts text at the cursor position and advances the cursor.
func (s *TextInputState) Insert(text string) {
	if text == "" {
		return
	}
	newGraphemes := splitGraphemes(text)
	s.Content.Update(func(graphemes []string) []string {
		cursor := s.CursorIndex.Peek()
		// Insert new graphemes at cursor position
		result := make([]string, 0, len(graphemes)+len(newGraphemes))
		result = append(result, graphemes[:cursor]...)
		result = append(result, newGraphemes...)
		result = append(result, graphemes[cursor:]...)
		return result
	})
	s.CursorIndex.Update(func(cursor int) int {
		return cursor + len(newGraphemes)
	})
}

// DeleteBackward deletes the grapheme before the cursor.
func (s *TextInputState) DeleteBackward() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:cursor-1], graphemes[cursor:]...)
	})
	s.CursorIndex.Set(cursor - 1)
}

// DeleteForward deletes the grapheme at the cursor.
func (s *TextInputState) DeleteForward() {
	cursor := s.CursorIndex.Peek()
	graphemes := s.Content.Peek()
	if cursor >= len(graphemes) {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:cursor], graphemes[cursor+1:]...)
	})
}

// DeleteToBeginning deletes from cursor to beginning of line.
func (s *TextInputState) DeleteToBeginning() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return graphemes[cursor:]
	})
	s.CursorIndex.Set(0)
}

// DeleteToEnd deletes from cursor to end of line.
func (s *TextInputState) DeleteToEnd() {
	cursor := s.CursorIndex.Peek()
	graphemes := s.Content.Peek()
	if cursor >= len(graphemes) {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return graphemes[:cursor]
	})
}

// DeleteWordBackward deletes the word before the cursor.
func (s *TextInputState) DeleteWordBackward() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	graphemes := s.Content.Peek()

	// Skip any whitespace/non-word chars immediately before cursor
	newCursor := cursor
	for newCursor > 0 && !isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}
	// Then delete word chars
	for newCursor > 0 && isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}

	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:newCursor], graphemes[cursor:]...)
	})
	s.CursorIndex.Set(newCursor)
}

// CursorLeft moves the cursor left by one grapheme.
func (s *TextInputState) CursorLeft() {
	s.CursorIndex.Update(func(cursor int) int {
		if cursor > 0 {
			return cursor - 1
		}
		return cursor
	})
}

// CursorRight moves the cursor right by one grapheme.
func (s *TextInputState) CursorRight() {
	graphemes := s.Content.Peek()
	s.CursorIndex.Update(func(cursor int) int {
		if cursor < len(graphemes) {
			return cursor + 1
		}
		return cursor
	})
}

// CursorHome moves the cursor to the beginning.
func (s *TextInputState) CursorHome() {
	s.CursorIndex.Set(0)
}

// CursorEnd moves the cursor to the end.
func (s *TextInputState) CursorEnd() {
	s.CursorIndex.Set(len(s.Content.Peek()))
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (s *TextInputState) CursorWordLeft() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	graphemes := s.Content.Peek()

	newCursor := cursor
	// Skip any whitespace/non-word chars immediately before cursor
	for newCursor > 0 && !isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}
	// Then skip word chars
	for newCursor > 0 && isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}

	s.CursorIndex.Set(newCursor)
}

// CursorWordRight moves the cursor to the next word boundary.
func (s *TextInputState) CursorWordRight() {
	graphemes := s.Content.Peek()
	cursor := s.CursorIndex.Peek()
	if cursor >= len(graphemes) {
		return
	}

	newCursor := cursor
	// Skip current word chars
	for newCursor < len(graphemes) && isWordChar(graphemes[newCursor]) {
		newCursor++
	}
	// Then skip whitespace/non-word chars
	for newCursor < len(graphemes) && !isWordChar(graphemes[newCursor]) {
		newCursor++
	}

	s.CursorIndex.Set(newCursor)
}

// cursorDisplayX returns the cursor position in display cells.
func (s *TextInputState) cursorDisplayX() int {
	graphemes := s.Content.Peek()
	cursor := s.CursorIndex.Peek()
	x := 0
	for i := 0; i < cursor && i < len(graphemes); i++ {
		x += graphemeWidth(graphemes[i])
	}
	return x
}

// contentWidth returns the total display width of the content.
func (s *TextInputState) contentWidth() int {
	graphemes := s.Content.Peek()
	width := 0
	for _, g := range graphemes {
		width += graphemeWidth(g)
	}
	return width
}

// clampCursor ensures cursor is within valid range.
func (s *TextInputState) clampCursor() {
	graphemes := s.Content.Peek()
	cursor := s.CursorIndex.Peek()
	if cursor < 0 {
		s.CursorIndex.Set(0)
	} else if cursor > len(graphemes) {
		s.CursorIndex.Set(len(graphemes))
	}
}

// --- Selection Methods ---

// HasSelection returns true if there is an active selection.
func (s *TextInputState) HasSelection() bool {
	anchor := s.SelectionAnchor.Peek()
	return anchor >= 0 && anchor != s.CursorIndex.Peek()
}

// GetSelectionBounds returns the normalized selection bounds (start, end).
// Returns (-1, -1) if there is no selection.
func (s *TextInputState) GetSelectionBounds() (start, end int) {
	anchor := s.SelectionAnchor.Peek()
	cursor := s.CursorIndex.Peek()
	if anchor < 0 || anchor == cursor {
		return -1, -1
	}
	if anchor < cursor {
		return anchor, cursor
	}
	return cursor, anchor
}

// GetSelectedText returns the selected text as a string.
// Returns empty string if there is no selection.
func (s *TextInputState) GetSelectedText() string {
	start, end := s.GetSelectionBounds()
	if start < 0 {
		return ""
	}
	graphemes := s.Content.Peek()
	return joinGraphemes(graphemes[start:end])
}

// ClearSelection clears the selection anchor.
func (s *TextInputState) ClearSelection() {
	s.SelectionAnchor.Set(-1)
}

// SetSelectionAnchor sets the selection anchor to the given index.
func (s *TextInputState) SetSelectionAnchor(index int) {
	s.SelectionAnchor.Set(index)
}

// SelectAll selects all text (anchor=0, cursor=len).
func (s *TextInputState) SelectAll() {
	graphemes := s.Content.Peek()
	s.SelectionAnchor.Set(0)
	s.CursorIndex.Set(len(graphemes))
}

// SelectWord selects the word at the given grapheme index.
func (s *TextInputState) SelectWord(index int) {
	graphemes := s.Content.Peek()
	if len(graphemes) == 0 {
		return
	}
	if index < 0 {
		index = 0
	}
	if index >= len(graphemes) {
		index = len(graphemes) - 1
	}

	// Find word boundaries
	start := index
	end := index

	// If at a non-word char, select consecutive non-word chars
	if !isWordChar(graphemes[index]) {
		for start > 0 && !isWordChar(graphemes[start-1]) {
			start--
		}
		for end < len(graphemes) && !isWordChar(graphemes[end]) {
			end++
		}
	} else {
		// Select word characters
		for start > 0 && isWordChar(graphemes[start-1]) {
			start--
		}
		for end < len(graphemes) && isWordChar(graphemes[end]) {
			end++
		}
	}

	s.SelectionAnchor.Set(start)
	s.CursorIndex.Set(end)
}

// DeleteSelection deletes the selected text.
// Returns true if selection was deleted, false if there was no selection.
func (s *TextInputState) DeleteSelection() bool {
	start, end := s.GetSelectionBounds()
	if start < 0 {
		return false
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:start], graphemes[end:]...)
	})
	s.CursorIndex.Set(start)
	s.SelectionAnchor.Set(-1)
	return true
}

// ReplaceSelection deletes any selected text and inserts the given text.
func (s *TextInputState) ReplaceSelection(text string) {
	s.DeleteSelection()
	s.Insert(text)
}

// SetCursorFromLocalPosition moves the cursor to the given local X position.
// It accounts for scroll offset internally. This mirrors TextArea's method
// but is simplified for single-line input.
func (s *TextInputState) SetCursorFromLocalPosition(localX int) {
	// Account for horizontal scroll offset
	displayX := localX + s.scrollOffset

	// Convert display position to grapheme index
	graphemes := s.Content.Peek()
	if len(graphemes) == 0 || displayX <= 0 {
		s.CursorIndex.Set(0)
		return
	}

	x := 0
	for i, grapheme := range graphemes {
		gWidth := graphemeWidth(grapheme)
		// Click in first half of grapheme -> position before it
		if displayX < x+gWidth/2+1 {
			s.CursorIndex.Set(i)
			return
		}
		x += gWidth
	}
	// Click past end -> position at end
	s.CursorIndex.Set(len(graphemes))
}

// --- TextInput Widget ---

// TextInput is a single-line focusable text entry widget.
// Content height is always 1 cell (single line). Use Style.Padding to add
// visual space around the text - the framework automatically accounts for padding.
type TextInput struct {
	ID            string            // Required for focus management
	State         *TextInputState   // Required - holds text and cursor position
	Placeholder   string            // Text shown when empty and unfocused
	Width         Dimension         // Deprecated: use Style.Width
	Height        Dimension         // Deprecated: use Style.Height (ignored; content height is always 1)
	Style         Style             // Optional styling (padding adds to outer size automatically)
	OnChange      func(text string) // Callback when text changes
	OnSubmit      func(text string) // Callback when Enter pressed
	Click         func(MouseEvent)  // Optional click callback
	MouseDown     func(MouseEvent)  // Optional mouse down callback
	MouseUp       func(MouseEvent)  // Optional mouse up callback
	Hover         func(bool)        // Optional hover callback
	Blur          func()            // Optional blur callback
	ExtraKeybinds []Keybind         // Optional additional keybinds (checked before defaults)
}

// WidgetID returns the text input's unique identifier.
func (t TextInput) WidgetID() string {
	return t.ID
}

// IsFocusable returns true, indicating this widget can receive keyboard focus.
func (t TextInput) IsFocusable() bool {
	return true
}

// CapturesKey returns true if this key would be captured by the text input
// (i.e., typed as text rather than bubbling to ancestors). This is true for
// printable characters without modifiers.
func (t TextInput) CapturesKey(key string) bool {
	// Keys with modifiers are not captured (they may have special handling)
	if strings.Contains(key, "+") {
		return false
	}

	// Single printable rune is captured as typed text
	runes := []rune(key)
	if len(runes) == 1 && unicode.IsPrint(runes[0]) {
		return true
	}

	// Multi-character key names (like "escape", "enter") are not captured
	return false
}

// Keybinds returns the declarative keybindings for this text input.
// ExtraKeybinds are checked first, allowing custom behavior to override defaults.
func (t TextInput) Keybinds() []Keybind {
	defaults := []Keybind{
		{Key: "enter", Name: "Submit", Action: t.submit},
		// Cursor movement (clears selection)
		{Key: "left", Action: t.cursorLeft, Hidden: true},
		{Key: "right", Action: t.cursorRight, Hidden: true},
		{Key: "home", Action: t.cursorHome, Hidden: true},
		{Key: "end", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+e", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+left", Action: t.cursorWordLeft, Hidden: true},
		{Key: "alt+b", Action: t.cursorWordLeft, Hidden: true},
		{Key: "ctrl+right", Action: t.cursorWordRight, Hidden: true},
		{Key: "alt+f", Action: t.cursorWordRight, Hidden: true},
		// Selection movement (extends selection)
		{Key: "shift+left", Action: t.selectLeft, Hidden: true},
		{Key: "shift+right", Action: t.selectRight, Hidden: true},
		{Key: "shift+home", Action: t.selectHome, Hidden: true},
		{Key: "shift+end", Action: t.selectEnd, Hidden: true},
		{Key: "ctrl+shift+left", Action: t.selectWordLeft, Hidden: true},
		{Key: "ctrl+shift+right", Action: t.selectWordRight, Hidden: true},
		// Select all
		{Key: "ctrl+a", Action: t.selectAll, Hidden: true},
		// Deletion
		{Key: "backspace", Action: t.deleteBackward, Hidden: true},
		{Key: "delete", Action: t.deleteForward, Hidden: true},
		{Key: "ctrl+d", Action: t.deleteForward, Hidden: true},
		{Key: "ctrl+u", Action: t.deleteToBeginning, Hidden: true},
		{Key: "ctrl+k", Action: t.deleteToEnd, Hidden: true},
		{Key: "ctrl+w", Action: t.deleteWordBackward, Hidden: true},
		{Key: "alt+backspace", Action: t.deleteWordBackward, Hidden: true},
	}
	// Prepend extra keybinds so they're checked first
	if len(t.ExtraKeybinds) > 0 {
		return append(t.ExtraKeybinds, defaults...)
	}
	return defaults
}

// Keybind action methods

func (t TextInput) submit() {
	if t.OnSubmit != nil && t.State != nil {
		t.OnSubmit(t.State.GetText())
	}
}

func (t TextInput) cursorLeft() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorLeft()
	}
}

func (t TextInput) cursorRight() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorRight()
	}
}

func (t TextInput) cursorHome() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorHome()
	}
}

func (t TextInput) cursorEnd() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorEnd()
	}
}

func (t TextInput) cursorWordLeft() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorWordLeft()
	}
}

func (t TextInput) cursorWordRight() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorWordRight()
	}
}

// ensureAnchor sets the selection anchor to the current cursor position if not already set.
func (t TextInput) ensureAnchor() {
	if t.State != nil && t.State.SelectionAnchor.Peek() < 0 {
		t.State.SetSelectionAnchor(t.State.CursorIndex.Peek())
	}
}

// Selection action methods

func (t TextInput) selectLeft() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorLeft()
	}
}

func (t TextInput) selectRight() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorRight()
	}
}

func (t TextInput) selectHome() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorHome()
	}
}

func (t TextInput) selectEnd() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorEnd()
	}
}

func (t TextInput) selectWordLeft() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorWordLeft()
	}
}

func (t TextInput) selectWordRight() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorWordRight()
	}
}

func (t TextInput) selectAll() {
	if t.State != nil {
		t.State.SelectAll()
	}
}

func (t TextInput) deleteBackward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteBackward()
		}
		t.notifyChange()
	}
}

func (t TextInput) deleteForward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteForward()
		}
		t.notifyChange()
	}
}

func (t TextInput) deleteToBeginning() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteToBeginning()
		}
		t.notifyChange()
	}
}

func (t TextInput) deleteToEnd() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteToEnd()
		}
		t.notifyChange()
	}
}

func (t TextInput) deleteWordBackward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteWordBackward()
		}
		t.notifyChange()
	}
}

func (t TextInput) notifyChange() {
	if t.OnChange != nil && t.State != nil {
		t.OnChange(t.State.GetText())
	}
}

// OnKey handles printable character input not covered by Keybinds().
func (t TextInput) OnKey(event KeyEvent) bool {
	if t.State == nil {
		return false
	}

	// Use Text() to get the actual typed characters (including space as " ")
	// Text() returns empty string for non-text keys like arrows, function keys, etc.
	text := event.Text()
	if text != "" {
		t.State.ReplaceSelection(text)
		t.notifyChange()
		return true
	}

	return false
}

// Build returns self since TextInput is a leaf widget with custom rendering.
func (t TextInput) Build(ctx BuildContext) Widget {
	return t
}

// GetContentDimensions returns the content-box dimensions.
// Height is always 1 cell (single line of text). Padding/border from Style
// are automatically added by the framework to compute the final outer size.
func (t TextInput) GetContentDimensions() (width, height Dimension) {
	width = t.Style.GetDimensions().Width
	if width.IsUnset() {
		width = t.Width
	}
	return width, Cells(1)
}

// GetStyle returns the style.
func (t TextInput) GetStyle() Style {
	return t.Style
}

// BuildLayoutNode builds a layout node for this TextInput widget.
// Implements the LayoutNodeBuilder interface.
func (t TextInput) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	style := t.Style

	padding := toLayoutEdgeInsets(style.Padding)
	border := borderToEdgeInsets(style.Border)
	dims := style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = t.Width
	}
	dims.Height = Cells(1)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.BoxNode{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		Padding:   padding,
		Border:    border,
		Margin:    toLayoutEdgeInsets(style.Margin),
		MeasureFunc: func(constraints layout.Constraints) (int, int) {
			size := t.Layout(ctx, Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			})
			return size.Width, size.Height
		},
	})

	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{
			child:     node,
			minWidth:  dims.MinWidth,
			maxWidth:  dims.MaxWidth,
			minHeight: dims.MinHeight,
			maxHeight: dims.MaxHeight,
			padding:   padding,
			border:    border,
		}
	}

	return node
}

// Layout computes the size of the text input.
func (t TextInput) Layout(ctx BuildContext, constraints Constraints) Size {
	// Height is always 1 for single-line
	height := 1

	// Width depends on Dimension setting
	widthDim := t.Style.GetDimensions().Width
	if widthDim.IsUnset() {
		widthDim = t.Width
	}
	var width int
	switch {
	case widthDim.IsCells():
		width = widthDim.CellsValue()
	case widthDim.IsFlex():
		width = constraints.MaxWidth
	default: // Auto - use content or placeholder width, minimum 1
		contentWidth := 1
		if t.State != nil {
			contentWidth = t.State.contentWidth()
		}
		placeholderWidth := ansi.StringWidth(t.Placeholder)
		width = max(contentWidth, placeholderWidth)
	}

	// Clamp to constraints
	width = clampInt(width, constraints.MinWidth, constraints.MaxWidth)

	return Size{Width: width, Height: height}
}

// Render draws the text input with cursor.
func (t TextInput) Render(ctx *RenderContext) {
	if t.State == nil {
		return
	}

	focused := ctx.IsFocused(t)
	theme := ctx.buildContext.Theme()

	// Subscribe to state changes by calling Get()
	graphemes := t.State.Content.Get()
	cursorIdx := t.State.CursorIndex.Get()
	_ = t.State.SelectionAnchor.Get() // Subscribe to selection changes

	viewportWidth := ctx.Width

	// Determine base style
	baseStyle := t.Style
	if baseStyle.ForegroundColor == nil || !baseStyle.ForegroundColor.IsSet() {
		baseStyle.ForegroundColor = theme.Text
	}
	if baseStyle.BackgroundColor == nil || !baseStyle.BackgroundColor.IsSet() {
		baseStyle.BackgroundColor = theme.Surface
	}

	// Fill background - sample from ColorProvider
	bgColor := baseStyle.BackgroundColor.ColorAt(viewportWidth, 1, 0, 0)
	ctx.FillRect(0, 0, viewportWidth, 1, bgColor)

	// Show placeholder if empty
	if len(graphemes) == 0 {
		placeholderStyle := baseStyle
		placeholderStyle.ForegroundColor = theme.TextMuted
		text := t.Placeholder
		if ansi.StringWidth(text) > viewportWidth {
			text = ansi.Truncate(text, viewportWidth, "")
		}
		ctx.DrawStyledText(0, 0, text, placeholderStyle)
		// Draw cursor at position 0 if focused
		if focused {
			cursorStyle := baseStyle
			cursorStyle.Reverse = true
			// Show the first placeholder character under the cursor, or space if no placeholder
			cursorChar := " "
			if len(text) > 0 {
				firstGrapheme, _ := ansi.FirstGraphemeCluster(text, ansi.GraphemeWidth)
				if firstGrapheme != "" {
					cursorChar = firstGrapheme
				}
			}
			ctx.DrawStyledText(0, 0, cursorChar, cursorStyle)
		}
		return
	}

	// Update scroll offset to keep cursor visible
	t.updateScrollOffset(viewportWidth)
	scrollOffset := t.State.scrollOffset

	// Get selection bounds
	selStart, selEnd := t.State.GetSelectionBounds()

	// Render text with cursor and selection
	t.renderContent(ctx, graphemes, cursorIdx, scrollOffset, viewportWidth, focused, baseStyle, selStart, selEnd, theme)
}

// updateScrollOffset ensures the cursor is visible within the viewport.
func (t TextInput) updateScrollOffset(viewportWidth int) {
	cursorX := t.State.cursorDisplayX()
	scrollOffset := t.State.scrollOffset

	// Cursor left of viewport - scroll left
	if cursorX < scrollOffset {
		t.State.scrollOffset = cursorX
		return
	}

	// Cursor right of viewport - scroll right
	// We need at least 1 cell for the cursor
	if cursorX >= scrollOffset+viewportWidth {
		t.State.scrollOffset = cursorX - viewportWidth + 1
	}
}

// renderContent renders the text with cursor and selection highlighting.
func (t TextInput) renderContent(ctx *RenderContext, graphemes []string, cursorIdx, scrollOffset, viewportWidth int, focused bool, baseStyle Style, selStart, selEnd int, theme ThemeData) {
	displayX := 0 // Position in content (display cells)
	hasSelection := selStart >= 0

	for i, grapheme := range graphemes {
		gWidth := graphemeWidth(grapheme)

		// Skip graphemes before scroll offset
		if displayX+gWidth <= scrollOffset {
			displayX += gWidth
			continue
		}

		// Calculate visible X position
		visibleX := displayX - scrollOffset

		// Stop if past viewport
		if visibleX >= viewportWidth {
			break
		}

		// Handle partial visibility at left edge
		if visibleX < 0 {
			// Grapheme starts before viewport but extends into it
			// Skip it for simplicity (could render partial but complex)
			displayX += gWidth
			continue
		}

		// Determine style
		style := baseStyle
		isSelected := hasSelection && i >= selStart && i < selEnd
		isCursor := focused && i == cursorIdx

		// Cursor style (reverse) takes precedence over selection
		if isCursor {
			style.Reverse = true
		} else if isSelected {
			// Match List/Table selection styling - just background, no foreground change
			style.BackgroundColor = theme.Selection
		}

		ctx.DrawStyledText(visibleX, 0, grapheme, style)
		displayX += gWidth
	}

	// Draw cursor at end if focused and cursor is at end
	if focused && cursorIdx >= len(graphemes) {
		cursorX := t.State.cursorDisplayX() - scrollOffset
		if cursorX >= 0 && cursorX < viewportWidth {
			cursorStyle := baseStyle
			cursorStyle.Reverse = true
			ctx.DrawStyledText(cursorX, 0, " ", cursorStyle)
		}
	}
}

// OnClick is called when the widget is clicked.
func (t TextInput) OnClick(event MouseEvent) {
	if t.Click != nil {
		t.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (t TextInput) OnMouseDown(event MouseEvent) {
	if t.State == nil {
		if t.MouseDown != nil {
			t.MouseDown(event)
		}
		return
	}

	// Shift+click: extend selection from current position
	if event.Mod.Contains(uv.ModShift) {
		t.ensureAnchor()
		t.State.SetCursorFromLocalPosition(event.LocalX)
		if t.MouseDown != nil {
			t.MouseDown(event)
		}
		return
	}

	// Clear any existing selection
	t.State.ClearSelection()

	// Position cursor at click location
	t.State.SetCursorFromLocalPosition(event.LocalX)
	cursor := t.State.CursorIndex.Peek()

	// Handle multi-click
	switch event.ClickCount {
	case 2:
		// Double-click: select word
		t.State.SelectWord(cursor)
	default:
		// Single click: set anchor to prepare for drag
		t.State.SetSelectionAnchor(cursor)
	}

	if t.MouseDown != nil {
		t.MouseDown(event)
	}
}

// OnMouseMove is called when the mouse is moved while dragging.
// Implements the MouseMoveHandler interface.
func (t TextInput) OnMouseMove(event MouseEvent) {
	if t.State == nil {
		return
	}

	// Update cursor position; selection extends from anchor
	t.State.SetCursorFromLocalPosition(event.LocalX)
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (t TextInput) OnMouseUp(event MouseEvent) {
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
func (t TextInput) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// OnBlur is called when the widget loses focus.
func (t TextInput) OnBlur() {
	if t.Blur != nil {
		t.Blur()
	}
}

// CursorScreenPosition returns the screen X position of the cursor
// given the widget's screen X position. Used for IME positioning.
func (t TextInput) CursorScreenPosition(widgetX int) int {
	if t.State == nil {
		return widgetX
	}
	cursorX := t.State.cursorDisplayX()
	scrollOffset := t.State.scrollOffset
	return widgetX + cursorX - scrollOffset
}
