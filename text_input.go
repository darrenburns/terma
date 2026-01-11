package terma

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/x/ansi"
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
	Content     AnySignal[[]string] // Grapheme clusters for Unicode safety
	CursorIndex Signal[int]         // Grapheme index (0 = before first char)

	// scrollOffset is calculated during render to keep cursor visible.
	// Not a signal because it's derived state, not source of truth.
	scrollOffset int
}

// NewTextInputState creates a new TextInputState with optional initial text.
func NewTextInputState(initial string) *TextInputState {
	graphemes := splitGraphemes(initial)
	return &TextInputState{
		Content:     NewAnySignal(graphemes),
		CursorIndex: NewSignal(len(graphemes)), // Cursor at end
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

// --- TextInput Widget ---

// TextInput is a single-line focusable text entry widget.
type TextInput struct {
	ID             string           // Required for focus management
	State          *TextInputState  // Required - holds text and cursor position
	Placeholder    string           // Text shown when empty and unfocused
	Width          Dimension        // Optional width
	Height         Dimension        // Ignored (always 1 for single-line)
	Style          Style            // Optional styling
	OnChange       func(text string) // Callback when text changes
	OnSubmit       func(text string) // Callback when Enter pressed
	Click          func()           // Optional click callback
	Hover          func(bool)       // Optional hover callback
	ExtraKeybinds  []Keybind        // Optional additional keybinds (checked before defaults)
}

// WidgetID returns the text input's unique identifier.
func (t TextInput) WidgetID() string {
	return t.ID
}

// IsFocusable returns true, indicating this widget can receive keyboard focus.
func (t TextInput) IsFocusable() bool {
	return true
}

// Keybinds returns the declarative keybindings for this text input.
// ExtraKeybinds are checked first, allowing custom behavior to override defaults.
func (t TextInput) Keybinds() []Keybind {
	defaults := []Keybind{
		{Key: "enter", Name: "Submit", Action: t.submit},
		// Cursor movement
		{Key: "left", Action: t.cursorLeft, Hidden: true},
		{Key: "right", Action: t.cursorRight, Hidden: true},
		{Key: "home", Action: t.cursorHome, Hidden: true},
		{Key: "ctrl+a", Action: t.cursorHome, Hidden: true},
		{Key: "end", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+e", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+left", Action: t.cursorWordLeft, Hidden: true},
		{Key: "alt+b", Action: t.cursorWordLeft, Hidden: true},
		{Key: "ctrl+right", Action: t.cursorWordRight, Hidden: true},
		{Key: "alt+f", Action: t.cursorWordRight, Hidden: true},
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
		t.State.CursorLeft()
	}
}

func (t TextInput) cursorRight() {
	if t.State != nil {
		t.State.CursorRight()
	}
}

func (t TextInput) cursorHome() {
	if t.State != nil {
		t.State.CursorHome()
	}
}

func (t TextInput) cursorEnd() {
	if t.State != nil {
		t.State.CursorEnd()
	}
}

func (t TextInput) cursorWordLeft() {
	if t.State != nil {
		t.State.CursorWordLeft()
	}
}

func (t TextInput) cursorWordRight() {
	if t.State != nil {
		t.State.CursorWordRight()
	}
}

func (t TextInput) deleteBackward() {
	if t.State != nil {
		t.State.DeleteBackward()
		t.notifyChange()
	}
}

func (t TextInput) deleteForward() {
	if t.State != nil {
		t.State.DeleteForward()
		t.notifyChange()
	}
}

func (t TextInput) deleteToBeginning() {
	if t.State != nil {
		t.State.DeleteToBeginning()
		t.notifyChange()
	}
}

func (t TextInput) deleteToEnd() {
	if t.State != nil {
		t.State.DeleteToEnd()
		t.notifyChange()
	}
}

func (t TextInput) deleteWordBackward() {
	if t.State != nil {
		t.State.DeleteWordBackward()
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
		t.State.Insert(text)
		t.notifyChange()
		return true
	}

	return false
}

// Build returns self since TextInput is a leaf widget with custom rendering.
func (t TextInput) Build(ctx BuildContext) Widget {
	return t
}

// GetDimensions returns the width and height dimension preferences.
func (t TextInput) GetDimensions() (width, height Dimension) {
	return t.Width, Cells(1) // Always 1 cell tall
}

// GetStyle returns the style.
func (t TextInput) GetStyle() Style {
	return t.Style
}

// Layout computes the size of the text input.
func (t TextInput) Layout(ctx BuildContext, constraints Constraints) Size {
	// Height is always 1 for single-line
	height := 1

	// Width depends on Dimension setting
	var width int
	switch {
	case t.Width.IsCells():
		width = t.Width.CellsValue()
	case t.Width.IsFlex():
		width = constraints.MaxWidth
	default: // Auto - use content or placeholder width, minimum 1
		contentWidth := 1
		if t.State != nil {
			contentWidth = t.State.contentWidth()
		}
		placeholderWidth := ansi.StringWidth(t.Placeholder)
		width = max(contentWidth, placeholderWidth, 1)
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

	// Show placeholder if empty and not focused
	if len(graphemes) == 0 && !focused {
		placeholderStyle := baseStyle
		placeholderStyle.ForegroundColor = theme.TextMuted
		text := t.Placeholder
		if ansi.StringWidth(text) > viewportWidth {
			text = ansi.Truncate(text, viewportWidth, "")
		}
		ctx.DrawStyledText(0, 0, text, placeholderStyle)
		return
	}

	// Update scroll offset to keep cursor visible
	t.updateScrollOffset(viewportWidth)
	scrollOffset := t.State.scrollOffset

	// Render text with cursor
	t.renderContent(ctx, graphemes, cursorIdx, scrollOffset, viewportWidth, focused, baseStyle)
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

// renderContent renders the text with cursor highlighting.
func (t TextInput) renderContent(ctx *RenderContext, graphemes []string, cursorIdx, scrollOffset, viewportWidth int, focused bool, baseStyle Style) {
	displayX := 0 // Position in content (display cells)

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

		// Determine style - reverse for cursor when focused
		style := baseStyle
		if focused && i == cursorIdx {
			style.Reverse = true
		}

		// Handle partial visibility at left edge
		if visibleX < 0 {
			// Grapheme starts before viewport but extends into it
			// Skip it for simplicity (could render partial but complex)
			displayX += gWidth
			continue
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
func (t TextInput) OnClick() {
	if t.Click != nil {
		t.Click()
	}
}

// OnHover is called when the hover state changes.
func (t TextInput) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
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
