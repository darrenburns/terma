package terma

import (
	"strings"
	"unicode"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"terma/layout"
)

// TextAreaState holds the state for a TextArea widget.
// It is the source of truth for text content, cursor position,
// wrapping mode, and insert mode.
type TextAreaState struct {
	Content         AnySignal[[]string] // Grapheme clusters for Unicode safety
	CursorIndex     Signal[int]         // Grapheme index (0 = before first char)
	InsertMode      Signal[bool]        // True when edits are allowed
	WrapMode        Signal[WrapMode]    // WrapNone or WrapSoft/WrapHard
	SelectionAnchor Signal[int]         // -1 = no selection, else anchor grapheme index

	scrollOffsetX int
	scrollOffsetY int
	lastWidth     int
	lastHeight    int
	lastFocused   bool

	preferredColumn int
}

// NewTextAreaState creates a new TextAreaState with optional initial text.
func NewTextAreaState(initial string) *TextAreaState {
	graphemes := splitGraphemes(initial)
	return &TextAreaState{
		Content:         NewAnySignal(graphemes),
		CursorIndex:     NewSignal(len(graphemes)),
		InsertMode:      NewSignal(true),
		WrapMode:        NewSignal(WrapSoft),
		SelectionAnchor: NewSignal(-1),
		preferredColumn: -1,
	}
}

// GetText returns the content as a string.
func (s *TextAreaState) GetText() string {
	return joinGraphemes(s.Content.Peek())
}

// SetText replaces the content and clamps the cursor.
func (s *TextAreaState) SetText(text string) {
	graphemes := splitGraphemes(text)
	s.Content.Set(graphemes)
	s.clampCursor()
	s.resetPreferredColumn()
}

// Insert inserts text at the cursor position and advances the cursor.
func (s *TextAreaState) Insert(text string) {
	if text == "" {
		return
	}
	newGraphemes := splitGraphemes(text)
	s.Content.Update(func(graphemes []string) []string {
		cursor := s.CursorIndex.Peek()
		result := make([]string, 0, len(graphemes)+len(newGraphemes))
		result = append(result, graphemes[:cursor]...)
		result = append(result, newGraphemes...)
		result = append(result, graphemes[cursor:]...)
		return result
	})
	s.CursorIndex.Update(func(cursor int) int {
		return cursor + len(newGraphemes)
	})
	s.updatePreferredColumn()
}

// InsertNewline inserts a newline at the cursor position.
func (s *TextAreaState) InsertNewline() {
	s.Insert("\n")
}

// DeleteBackward deletes the grapheme before the cursor.
func (s *TextAreaState) DeleteBackward() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:cursor-1], graphemes[cursor:]...)
	})
	s.CursorIndex.Set(cursor - 1)
	s.updatePreferredColumn()
}

// DeleteForward deletes the grapheme at the cursor.
func (s *TextAreaState) DeleteForward() {
	cursor := s.CursorIndex.Peek()
	graphemes := s.Content.Peek()
	if cursor >= len(graphemes) {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:cursor], graphemes[cursor+1:]...)
	})
	s.updatePreferredColumn()
}

// DeleteToBeginning deletes from cursor to beginning of line.
func (s *TextAreaState) DeleteToBeginning() {
	cursor := s.CursorIndex.Peek()
	start, _ := lineBoundsForIndex(s.Content.Peek(), cursor)
	if cursor <= start {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:start], graphemes[cursor:]...)
	})
	s.CursorIndex.Set(start)
	s.updatePreferredColumn()
}

// DeleteToEnd deletes from cursor to end of line.
func (s *TextAreaState) DeleteToEnd() {
	cursor := s.CursorIndex.Peek()
	_, end := lineBoundsForIndex(s.Content.Peek(), cursor)
	if cursor >= end {
		return
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:cursor], graphemes[end:]...)
	})
	s.updatePreferredColumn()
}

// DeleteWordBackward deletes the word before the cursor.
func (s *TextAreaState) DeleteWordBackward() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	graphemes := s.Content.Peek()

	newCursor := cursor
	for newCursor > 0 && !isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}
	for newCursor > 0 && isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}

	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:newCursor], graphemes[cursor:]...)
	})
	s.CursorIndex.Set(newCursor)
	s.updatePreferredColumn()
}

// CursorLeft moves the cursor left by one grapheme.
func (s *TextAreaState) CursorLeft() {
	s.CursorIndex.Update(func(cursor int) int {
		if cursor > 0 {
			return cursor - 1
		}
		return cursor
	})
	s.updatePreferredColumn()
}

// CursorRight moves the cursor right by one grapheme.
func (s *TextAreaState) CursorRight() {
	graphemes := s.Content.Peek()
	s.CursorIndex.Update(func(cursor int) int {
		if cursor < len(graphemes) {
			return cursor + 1
		}
		return cursor
	})
	s.updatePreferredColumn()
}

// CursorHome moves the cursor to the start of the current line.
func (s *TextAreaState) CursorHome() {
	start, _ := lineBoundsForIndex(s.Content.Peek(), s.CursorIndex.Peek())
	s.CursorIndex.Set(start)
	s.updatePreferredColumn()
}

// CursorEnd moves the cursor to the end of the current line.
func (s *TextAreaState) CursorEnd() {
	_, end := lineBoundsForIndex(s.Content.Peek(), s.CursorIndex.Peek())
	s.CursorIndex.Set(end)
	s.updatePreferredColumn()
}

// CursorWordLeft moves the cursor to the previous word boundary.
func (s *TextAreaState) CursorWordLeft() {
	cursor := s.CursorIndex.Peek()
	if cursor <= 0 {
		return
	}
	graphemes := s.Content.Peek()

	newCursor := cursor
	for newCursor > 0 && !isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}
	for newCursor > 0 && isWordChar(graphemes[newCursor-1]) {
		newCursor--
	}

	s.CursorIndex.Set(newCursor)
	s.updatePreferredColumn()
}

// CursorWordRight moves the cursor to the next word boundary.
func (s *TextAreaState) CursorWordRight() {
	graphemes := s.Content.Peek()
	cursor := s.CursorIndex.Peek()
	if cursor >= len(graphemes) {
		return
	}

	newCursor := cursor
	for newCursor < len(graphemes) && isWordChar(graphemes[newCursor]) {
		newCursor++
	}
	for newCursor < len(graphemes) && !isWordChar(graphemes[newCursor]) {
		newCursor++
	}

	s.CursorIndex.Set(newCursor)
	s.updatePreferredColumn()
}

// CursorUp moves the cursor up by one display line.
func (s *TextAreaState) CursorUp() {
	s.cursorVerticalMove(-1)
}

// CursorDown moves the cursor down by one display line.
func (s *TextAreaState) CursorDown() {
	s.cursorVerticalMove(1)
}

// CursorUpBy moves the cursor up by the given number of display lines.
func (s *TextAreaState) CursorUpBy(lines int) {
	s.cursorVerticalMove(-lines)
}

// CursorDownBy moves the cursor down by the given number of display lines.
func (s *TextAreaState) CursorDownBy(lines int) {
	s.cursorVerticalMove(lines)
}

// ToggleWrap toggles between WrapNone and WrapSoft.
func (s *TextAreaState) ToggleWrap() {
	if s.WrapMode.Peek() == WrapNone {
		s.WrapMode.Set(WrapSoft)
	} else {
		s.WrapMode.Set(WrapNone)
	}
	s.resetPreferredColumn()
}

// HasSelection returns true if there is an active selection.
func (s *TextAreaState) HasSelection() bool {
	anchor := s.SelectionAnchor.Peek()
	return anchor >= 0 && anchor != s.CursorIndex.Peek()
}

// GetSelectionBounds returns the normalized selection bounds (start, end).
// Returns (-1, -1) if there is no selection.
func (s *TextAreaState) GetSelectionBounds() (start, end int) {
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
func (s *TextAreaState) GetSelectedText() string {
	start, end := s.GetSelectionBounds()
	if start < 0 {
		return ""
	}
	graphemes := s.Content.Peek()
	return joinGraphemes(graphemes[start:end])
}

// ClearSelection clears the selection anchor.
func (s *TextAreaState) ClearSelection() {
	s.SelectionAnchor.Set(-1)
}

// SetSelectionAnchor sets the selection anchor to the given index.
func (s *TextAreaState) SetSelectionAnchor(index int) {
	s.SelectionAnchor.Set(index)
}

// SelectAll selects all text (anchor=0, cursor=len).
func (s *TextAreaState) SelectAll() {
	graphemes := s.Content.Peek()
	s.SelectionAnchor.Set(0)
	s.CursorIndex.Set(len(graphemes))
	s.updatePreferredColumn()
}

// SelectWord selects the word at the given grapheme index.
func (s *TextAreaState) SelectWord(index int) {
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
	s.updatePreferredColumn()
}

// SelectLine selects the line at the given grapheme index.
func (s *TextAreaState) SelectLine(index int) {
	graphemes := s.Content.Peek()
	start, end := lineBoundsForIndex(graphemes, index)
	// Include the newline if present
	if end < len(graphemes) && graphemes[end] == "\n" {
		end++
	}
	s.SelectionAnchor.Set(start)
	s.CursorIndex.Set(end)
	s.updatePreferredColumn()
}

// DeleteSelection deletes the selected text.
// Returns true if selection was deleted, false if there was no selection.
func (s *TextAreaState) DeleteSelection() bool {
	start, end := s.GetSelectionBounds()
	if start < 0 {
		return false
	}
	s.Content.Update(func(graphemes []string) []string {
		return append(graphemes[:start], graphemes[end:]...)
	})
	s.CursorIndex.Set(start)
	s.SelectionAnchor.Set(-1)
	s.updatePreferredColumn()
	return true
}

// ReplaceSelection deletes any selected text and inserts the given text.
func (s *TextAreaState) ReplaceSelection(text string) {
	s.DeleteSelection()
	s.Insert(text)
}

func (s *TextAreaState) cursorVerticalMove(delta int) {
	graphemes := s.Content.Peek()
	if len(graphemes) == 0 {
		return
	}
	contentWidth := reservedContentWidth(s.lastWidth)
	layout := buildTextAreaLayout(graphemes, s.WrapMode.Peek(), contentWidth, s.CursorIndex.Peek())
	if len(layout.lines) == 0 {
		return
	}
	targetLine := clampInt(layout.cursorLine+delta, 0, len(layout.lines)-1)
	targetCol := s.preferredColumn
	if targetCol < 0 {
		targetCol = layout.cursorCol
	}
	newCursor := cursorIndexForLineColumn(layout.lines, graphemes, targetLine, targetCol)
	s.CursorIndex.Set(newCursor)
	s.preferredColumn = targetCol
}

func (s *TextAreaState) updatePreferredColumn() {
	graphemes := s.Content.Peek()
	contentWidth := reservedContentWidth(s.lastWidth)
	layout := buildTextAreaLayout(graphemes, s.WrapMode.Peek(), contentWidth, s.CursorIndex.Peek())
	s.preferredColumn = layout.cursorCol
}

func (s *TextAreaState) resetPreferredColumn() {
	s.preferredColumn = -1
}

// SetCursorFromLocalPosition moves the cursor to the given local position within the widget.
// It accounts for scroll offset internally. contentWidth is the width available for text
// content (excluding scrollbar space).
func (s *TextAreaState) SetCursorFromLocalPosition(localX, localY int, contentWidth int) {
	displayLine := localY + s.scrollOffsetY
	displayCol := localX + s.scrollOffsetX

	graphemes := s.Content.Peek()
	wrapMode := s.WrapMode.Peek()
	layout := buildTextAreaLayout(graphemes, wrapMode, contentWidth, s.CursorIndex.Peek())
	newIdx := cursorIndexForLineColumn(layout.lines, graphemes, displayLine, displayCol)
	s.CursorIndex.Set(newIdx)
	s.updatePreferredColumn()
}

func (s *TextAreaState) clampCursor() {
	graphemes := s.Content.Peek()
	cursor := s.CursorIndex.Peek()
	if cursor < 0 {
		s.CursorIndex.Set(0)
	} else if cursor > len(graphemes) {
		s.CursorIndex.Set(len(graphemes))
	}
}

type textAreaLine struct {
	start int
	end   int
	width int
}

type textAreaLayout struct {
	lines      []textAreaLine
	cursorLine int
	cursorCol  int
	maxWidth   int
}

func buildTextAreaLayout(graphemes []string, wrap WrapMode, maxWidth, cursorIdx int) textAreaLayout {
	if maxWidth <= 0 || wrap == WrapNone {
		wrap = WrapNone
	}

	lines := make([]textAreaLine, 0)
	lineStart := 0
	lineWidth := 0
	lineIndex := 0
	cursorLine := 0
	cursorCol := 0
	maxLineWidth := 0

	flushLine := func(end int) {
		lines = append(lines, textAreaLine{start: lineStart, end: end, width: lineWidth})
		if lineWidth > maxLineWidth {
			maxLineWidth = lineWidth
		}
	}

	for i, g := range graphemes {
		if cursorIdx == i {
			cursorLine = lineIndex
			cursorCol = lineWidth
		}

		if g == "\n" {
			flushLine(i)
			lineStart = i + 1
			lineWidth = 0
			lineIndex++
			continue
		}

		gWidth := graphemeWidth(g)
		if wrap != WrapNone && lineWidth+gWidth > maxWidth && lineWidth > 0 {
			flushLine(i)
			lineStart = i
			lineWidth = 0
			lineIndex++
			if cursorIdx == i {
				cursorLine = lineIndex
				cursorCol = 0
			}
		}

		lineWidth += gWidth
		if cursorIdx == i+1 {
			cursorLine = lineIndex
			cursorCol = lineWidth
		}
	}

	if cursorIdx == len(graphemes) {
		cursorLine = lineIndex
		cursorCol = lineWidth
	}

	flushLine(len(graphemes))

	if len(lines) == 0 {
		lines = append(lines, textAreaLine{start: 0, end: 0, width: 0})
	}

	return textAreaLayout{
		lines:      lines,
		cursorLine: cursorLine,
		cursorCol:  cursorCol,
		maxWidth:   maxLineWidth,
	}
}

func cursorIndexForLineColumn(lines []textAreaLine, graphemes []string, lineIdx, column int) int {
	if len(lines) == 0 {
		return 0
	}
	if lineIdx < 0 {
		lineIdx = 0
	} else if lineIdx >= len(lines) {
		lineIdx = len(lines) - 1
	}
	line := lines[lineIdx]
	if column <= 0 {
		return line.start
	}
	displayX := 0
	for i := line.start; i < line.end; i++ {
		gWidth := graphemeWidth(graphemes[i])
		if displayX+gWidth > column {
			return i
		}
		displayX += gWidth
	}
	return line.end
}

func lineBoundsForIndex(graphemes []string, cursorIdx int) (start, end int) {
	if cursorIdx < 0 {
		cursorIdx = 0
	}
	if cursorIdx > len(graphemes) {
		cursorIdx = len(graphemes)
	}
	start = cursorIdx
	for start > 0 && graphemes[start-1] != "\n" {
		start--
	}
	end = cursorIdx
	for end < len(graphemes) && graphemes[end] != "\n" {
		end++
	}
	return start, end
}

func maxLineWidth(graphemes []string) int {
	maxWidth := 0
	current := 0
	for _, g := range graphemes {
		if g == "\n" {
			maxWidth = max(maxWidth, current)
			current = 0
			continue
		}
		current += graphemeWidth(g)
	}
	maxWidth = max(maxWidth, current)
	return maxWidth
}

func maxLineWidthString(text string) int {
	lines := strings.Split(text, "\n")
	maxWidth := 0
	for _, line := range lines {
		maxWidth = max(maxWidth, ansi.StringWidth(line))
	}
	return maxWidth
}

func wrapLineCount(text string, width int, wrap WrapMode) int {
	if text == "" {
		return 1
	}
	lines := wrapText(text, width, wrap)
	if len(lines) == 0 {
		return 1
	}
	return len(lines)
}

func reservedContentWidth(viewportWidth int) int {
	if viewportWidth <= 1 {
		return viewportWidth
	}
	return viewportWidth - 1
}

// TextArea is a multi-line focusable text entry widget.
type TextArea struct {
	ID                string            // Required for focus management
	State             *TextAreaState    // Required - holds text and cursor position
	Placeholder       string            // Text shown when empty and unfocused
	Width             Dimension         // Deprecated: use Style.Width
	Height            Dimension         // Deprecated: use Style.Height
	Style             Style             // Optional styling
	RequireInsertMode bool              // If true, require entering insert mode to edit
	ScrollState       *ScrollState      // Optional state for scroll-into-view
	OnChange          func(text string) // Callback when text changes
	OnSubmit          func(text string) // Callback when submit key is pressed
	Click             func(MouseEvent)  // Optional click callback
	MouseDown         func(MouseEvent)  // Optional mouse down callback
	MouseUp           func(MouseEvent)  // Optional mouse up callback
	Hover             func(bool)        // Optional hover callback
	ExtraKeybinds     []Keybind         // Optional additional keybinds (checked before defaults)
}

// WidgetID returns the text area's unique identifier.
func (t TextArea) WidgetID() string {
	return t.ID
}

// IsFocusable returns true, indicating this widget can receive keyboard focus.
func (t TextArea) IsFocusable() bool {
	return true
}

// CapturesKey returns true if this key would be captured by the text area
// (i.e., typed as text rather than bubbling to ancestors). This is true for
// printable characters without modifiers when in insert mode.
func (t TextArea) CapturesKey(key string) bool {
	if !t.canInsert() {
		return false
	}
	if strings.Contains(key, "+") {
		return false
	}
	runes := []rune(key)
	if len(runes) == 1 && unicode.IsPrint(runes[0]) {
		return true
	}
	return false
}

// Keybinds returns the declarative keybindings for this text area.
// ExtraKeybinds are checked first, allowing custom behavior to override defaults.
func (t TextArea) Keybinds() []Keybind {
	if t.State == nil {
		return nil
	}

	keybinds := []Keybind{
		// Cursor movement (clears selection)
		{Key: "left", Action: t.cursorLeft, Hidden: true},
		{Key: "right", Action: t.cursorRight, Hidden: true},
		{Key: "up", Action: t.cursorUp, Hidden: true},
		{Key: "down", Action: t.cursorDown, Hidden: true},
		{Key: "home", Action: t.cursorHome, Hidden: true},
		{Key: "end", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+e", Action: t.cursorEnd, Hidden: true},
		{Key: "ctrl+left", Action: t.cursorWordLeft, Hidden: true},
		{Key: "alt+b", Action: t.cursorWordLeft, Hidden: true},
		{Key: "ctrl+right", Action: t.cursorWordRight, Hidden: true},
		{Key: "alt+f", Action: t.cursorWordRight, Hidden: true},
		{Key: "pgup", Action: t.cursorPageUp, Hidden: true},
		{Key: "pgdown", Action: t.cursorPageDown, Hidden: true},

		// Selection movement (extends selection)
		{Key: "shift+left", Action: t.selectLeft, Hidden: true},
		{Key: "shift+right", Action: t.selectRight, Hidden: true},
		{Key: "shift+up", Action: t.selectUp, Hidden: true},
		{Key: "shift+down", Action: t.selectDown, Hidden: true},
		{Key: "shift+home", Action: t.selectHome, Hidden: true},
		{Key: "shift+end", Action: t.selectEnd, Hidden: true},
		{Key: "ctrl+shift+left", Action: t.selectWordLeft, Hidden: true},
		{Key: "ctrl+shift+right", Action: t.selectWordRight, Hidden: true},
		{Key: "shift+pgup", Action: t.selectPageUp, Hidden: true},
		{Key: "shift+pgdown", Action: t.selectPageDown, Hidden: true},

		// Select all
		{Key: "ctrl+a", Action: t.selectAll, Hidden: true},
	}

	if t.RequireInsertMode {
		if t.State.InsertMode.Peek() {
			keybinds = append(keybinds, Keybind{Key: "escape", Name: "Normal", Action: t.exitInsertMode})
		} else {
			keybinds = append(keybinds,
				Keybind{Key: "i", Name: "Insert", Action: t.enterInsertMode},
				Keybind{Key: "enter", Name: "Insert", Action: t.enterInsertMode},
			)
		}
	}

	if t.canInsert() {
		keybinds = append(keybinds,
			Keybind{Key: "enter", Name: "Newline", Action: t.insertNewline},
			Keybind{Key: "backspace", Action: t.deleteBackward, Hidden: true},
			Keybind{Key: "delete", Action: t.deleteForward, Hidden: true},
			Keybind{Key: "ctrl+d", Action: t.deleteForward, Hidden: true},
			Keybind{Key: "ctrl+u", Action: t.deleteToBeginning, Hidden: true},
			Keybind{Key: "ctrl+k", Action: t.deleteToEnd, Hidden: true},
			Keybind{Key: "ctrl+w", Action: t.deleteWordBackward, Hidden: true},
			Keybind{Key: "alt+backspace", Action: t.deleteWordBackward, Hidden: true},
		)
	}

	if t.OnSubmit != nil {
		keybinds = append(keybinds, Keybind{Key: "ctrl+enter", Name: "Submit", Action: t.submit})
	}

	if len(t.ExtraKeybinds) > 0 {
		return append(t.ExtraKeybinds, keybinds...)
	}
	return keybinds
}

// Keybind action methods

func (t TextArea) submit() {
	if t.OnSubmit != nil && t.State != nil {
		t.OnSubmit(t.State.GetText())
	}
}

func (t TextArea) enterInsertMode() {
	if t.State != nil {
		t.State.InsertMode.Set(true)
	}
}

func (t TextArea) exitInsertMode() {
	if t.State != nil {
		t.State.InsertMode.Set(false)
	}
}

func (t TextArea) insertNewline() {
	if t.State != nil {
		t.State.ReplaceSelection("\n")
		t.notifyChange()
	}
}

func (t TextArea) cursorLeft() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorLeft()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorRight() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorRight()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorUp() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorUp()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorDown() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorDown()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorPageUp() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorUpBy(max(1, t.State.lastHeight-1))
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorPageDown() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorDownBy(max(1, t.State.lastHeight-1))
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorHome() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorHome()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorEnd() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorEnd()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorWordLeft() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorWordLeft()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) cursorWordRight() {
	if t.State != nil {
		t.State.ClearSelection()
		t.State.CursorWordRight()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) deleteBackward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteBackward()
		}
		t.notifyChange()
	}
}

func (t TextArea) deleteForward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteForward()
		}
		t.notifyChange()
	}
}

func (t TextArea) deleteToBeginning() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteToBeginning()
		}
		t.notifyChange()
	}
}

func (t TextArea) deleteToEnd() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteToEnd()
		}
		t.notifyChange()
	}
}

func (t TextArea) deleteWordBackward() {
	if t.State != nil {
		if !t.State.DeleteSelection() {
			t.State.DeleteWordBackward()
		}
		t.notifyChange()
	}
}

// ensureAnchor sets the selection anchor to the current cursor position if not already set.
func (t TextArea) ensureAnchor() {
	if t.State != nil && t.State.SelectionAnchor.Peek() < 0 {
		t.State.SetSelectionAnchor(t.State.CursorIndex.Peek())
	}
}

// Selection action methods

func (t TextArea) selectLeft() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorLeft()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectRight() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorRight()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectUp() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorUp()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectDown() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorDown()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectPageUp() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorUpBy(max(1, t.State.lastHeight-1))
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectPageDown() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorDownBy(max(1, t.State.lastHeight-1))
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectHome() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorHome()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectEnd() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorEnd()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectWordLeft() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorWordLeft()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectWordRight() {
	if t.State != nil {
		t.ensureAnchor()
		t.State.CursorWordRight()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) selectAll() {
	if t.State != nil {
		t.State.SelectAll()
		t.scrollCursorIntoView()
	}
}

func (t TextArea) notifyChange() {
	if t.OnChange != nil && t.State != nil {
		t.OnChange(t.State.GetText())
	}
}

func (t TextArea) canInsert() bool {
	if t.State == nil {
		return false
	}
	if !t.RequireInsertMode {
		return true
	}
	return t.State.InsertMode.Peek()
}

// OnKey handles printable character input not covered by Keybinds().
func (t TextArea) OnKey(event KeyEvent) bool {
	if t.State == nil || !t.canInsert() {
		return false
	}

	text := event.Text()
	if text != "" {
		t.State.ReplaceSelection(text)
		t.notifyChange()
		return true
	}
	return false
}

// Build returns self since TextArea is a leaf widget with custom rendering.
func (t TextArea) Build(ctx BuildContext) Widget {
	t.registerScrollCallbacks()
	return t
}

// BuildLayoutNode builds a layout node for this TextArea widget.
// Implements the LayoutNodeBuilder interface so Scrollable can measure it.
func (t TextArea) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	style := t.Style

	padding := toLayoutEdgeInsets(style.Padding)
	border := borderToEdgeInsets(style.Border)
	dims := style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = t.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = t.Height
	}
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

// GetContentDimensions returns the width and height dimension preferences.
func (t TextArea) GetContentDimensions() (width, height Dimension) {
	dims := t.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = t.Width
	}
	if height.IsUnset() {
		height = t.Height
	}
	return width, height
}

// GetStyle returns the style.
func (t TextArea) GetStyle() Style {
	return t.Style
}

// Layout computes the size of the text area.
func (t TextArea) Layout(ctx BuildContext, constraints Constraints) Size {
	dims := t.Style.GetDimensions()
	widthDim := dims.Width
	heightDim := dims.Height
	if widthDim.IsUnset() {
		widthDim = t.Width
	}
	if heightDim.IsUnset() {
		heightDim = t.Height
	}
	var width int
	switch {
	case widthDim.IsCells():
		width = widthDim.CellsValue()
	case widthDim.IsFlex():
		width = constraints.MaxWidth
	default:
		contentWidth := 1
		if t.State != nil {
			contentWidth = maxLineWidth(t.State.Content.Peek())
		}
		placeholderWidth := maxLineWidthString(t.Placeholder)
		width = max(contentWidth, placeholderWidth, 1)
		if width > 1 {
			width++
		}
	}
	width = clampInt(width, constraints.MinWidth, constraints.MaxWidth)

	var height int
	switch {
	case heightDim.IsCells():
		height = heightDim.CellsValue()
	case heightDim.IsFlex():
		height = constraints.MaxHeight
	default:
		contentLines := 1
		wrapMode := WrapSoft
		if t.State != nil {
			wrapMode = t.State.WrapMode.Peek()
			contentWidth := reservedContentWidth(width)
			layout := buildTextAreaLayout(t.State.Content.Peek(), wrapMode, contentWidth, t.State.CursorIndex.Peek())
			contentLines = max(1, len(layout.lines))
		}
		placeholderLines := wrapLineCount(t.Placeholder, reservedContentWidth(width), wrapMode)
		height = max(contentLines, placeholderLines, 1)
	}
	height = clampInt(height, constraints.MinHeight, constraints.MaxHeight)

	return Size{Width: width, Height: height}
}

// Render draws the text area with cursor.
func (t TextArea) Render(ctx *RenderContext) {
	if t.State == nil {
		return
	}

	focused := ctx.IsFocused(t)
	if t.RequireInsertMode {
		if focused && !t.State.lastFocused {
			t.State.InsertMode.Set(false)
		}
	}
	t.State.lastFocused = focused
	t.State.lastWidth = ctx.Width
	t.State.lastHeight = ctx.Height

	theme := ctx.buildContext.Theme()
	graphemes := t.State.Content.Get()
	cursorIdx := t.State.CursorIndex.Get()
	wrapMode := t.State.WrapMode.Get()
	contentWidth := reservedContentWidth(ctx.Width)

	baseStyle := t.Style
	if baseStyle.ForegroundColor == nil || !baseStyle.ForegroundColor.IsSet() {
		baseStyle.ForegroundColor = theme.Text
	}
	if baseStyle.BackgroundColor == nil || !baseStyle.BackgroundColor.IsSet() {
		baseStyle.BackgroundColor = theme.Surface
	}

	bgColor := baseStyle.BackgroundColor.ColorAt(ctx.Width, ctx.Height, 0, 0)
	ctx.FillRect(0, 0, ctx.Width, ctx.Height, bgColor)

	if len(graphemes) == 0 && !focused {
		placeholderStyle := baseStyle
		placeholderStyle.ForegroundColor = theme.TextMuted
		lines := wrapText(t.Placeholder, contentWidth, wrapMode)
		for i := 0; i < ctx.Height && i < len(lines); i++ {
			line := lines[i]
			if ansi.StringWidth(line) > contentWidth {
				line = ansi.Truncate(line, contentWidth, "")
			}
			ctx.DrawStyledText(0, i, line, placeholderStyle)
		}
		return
	}

	layout := buildTextAreaLayout(graphemes, wrapMode, contentWidth, cursorIdx)
	t.updateScrollOffsets(layout, contentWidth, ctx.Height)
	t.scrollCursorIntoViewWithLayout(layout)

	selStart, selEnd := t.State.GetSelectionBounds()
	t.renderContent(ctx, graphemes, layout, cursorIdx, focused, baseStyle, contentWidth, selStart, selEnd, theme)
}

func (t TextArea) updateScrollOffsets(layout textAreaLayout, contentWidth, viewportHeight int) {
	if viewportHeight <= 0 {
		return
	}

	if layout.cursorLine < t.State.scrollOffsetY {
		t.State.scrollOffsetY = layout.cursorLine
	} else if layout.cursorLine >= t.State.scrollOffsetY+viewportHeight {
		t.State.scrollOffsetY = layout.cursorLine - viewportHeight + 1
	}

	maxY := max(0, len(layout.lines)-viewportHeight)
	t.State.scrollOffsetY = clampInt(t.State.scrollOffsetY, 0, maxY)

	if t.State.WrapMode.Peek() == WrapNone {
		if layout.cursorCol < t.State.scrollOffsetX {
			t.State.scrollOffsetX = layout.cursorCol
		} else if layout.cursorCol > t.State.scrollOffsetX+contentWidth {
			t.State.scrollOffsetX = layout.cursorCol - contentWidth
		}
		maxX := max(0, layout.maxWidth-contentWidth)
		t.State.scrollOffsetX = clampInt(t.State.scrollOffsetX, 0, maxX)
	} else {
		t.State.scrollOffsetX = 0
	}
}

func (t TextArea) renderContent(ctx *RenderContext, graphemes []string, layout textAreaLayout, cursorIdx int, focused bool, baseStyle Style, contentWidth int, selStart, selEnd int, theme ThemeData) {
	scrollY := t.State.scrollOffsetY
	scrollX := t.State.scrollOffsetX
	hasSelection := selStart >= 0

	for lineIdx := scrollY; lineIdx < len(layout.lines) && lineIdx < scrollY+ctx.Height; lineIdx++ {
		line := layout.lines[lineIdx]
		row := lineIdx - scrollY

		displayX := 0
		for i := line.start; i < line.end; i++ {
			grapheme := graphemes[i]
			gWidth := graphemeWidth(grapheme)

			if t.State.WrapMode.Peek() == WrapNone {
				if displayX+gWidth <= scrollX {
					displayX += gWidth
					continue
				}
			}

			visibleX := displayX - scrollX
			if visibleX >= contentWidth {
				break
			}
			if visibleX < 0 {
				displayX += gWidth
				continue
			}

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

			ctx.DrawStyledText(visibleX, row, grapheme, style)
			displayX += gWidth
		}

		if focused && cursorIdx == line.end && layout.cursorLine == lineIdx {
			cursorX := layout.cursorCol - scrollX
			if cursorX >= 0 && cursorX < ctx.Width {
				cursorStyle := baseStyle
				cursorStyle.Reverse = true
				ctx.DrawStyledText(cursorX, row, " ", cursorStyle)
			}
		}
	}
}

func (t TextArea) scrollCursorIntoView() {
	if t.ScrollState == nil || t.State == nil || t.State.lastWidth <= 0 {
		return
	}
	contentWidth := reservedContentWidth(t.State.lastWidth)
	layout := buildTextAreaLayout(t.State.Content.Peek(), t.State.WrapMode.Peek(), contentWidth, t.State.CursorIndex.Peek())
	t.scrollCursorIntoViewWithLayout(layout)
}

func (t TextArea) scrollCursorIntoViewWithLayout(layout textAreaLayout) {
	if t.ScrollState == nil {
		return
	}
	t.ScrollState.ScrollToView(layout.cursorLine, 1)
}

func (t TextArea) registerScrollCallbacks() {
	if t.ScrollState == nil {
		return
	}
	t.ScrollState.OnScrollUp = func(lines int) bool {
		if t.State == nil {
			return false
		}
		t.State.CursorUpBy(lines)
		t.scrollCursorIntoView()
		return true
	}
	t.ScrollState.OnScrollDown = func(lines int) bool {
		if t.State == nil {
			return false
		}
		t.State.CursorDownBy(lines)
		t.scrollCursorIntoView()
		return true
	}
}

// OnClick is called when the widget is clicked.
func (t TextArea) OnClick(event MouseEvent) {
	if t.Click != nil {
		t.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (t TextArea) OnMouseDown(event MouseEvent) {
	if t.State == nil || t.State.lastWidth <= 0 {
		if t.MouseDown != nil {
			t.MouseDown(event)
		}
		return
	}

	contentWidth := reservedContentWidth(t.State.lastWidth)

	// Shift+click: extend selection from current position
	if event.Mod.Contains(uv.ModShift) {
		t.ensureAnchor()
		t.State.SetCursorFromLocalPosition(event.LocalX, event.LocalY, contentWidth)
		if t.MouseDown != nil {
			t.MouseDown(event)
		}
		return
	}

	// Clear any existing selection
	t.State.ClearSelection()

	// Position cursor at click location
	t.State.SetCursorFromLocalPosition(event.LocalX, event.LocalY, contentWidth)
	cursor := t.State.CursorIndex.Peek()

	// Handle multi-click (click chain is reset at app level when position changes)
	switch event.ClickCount {
	case 2:
		// Double-click: select word
		t.State.SelectWord(cursor)
	case 3:
		// Triple-click: select line
		t.State.SelectLine(cursor)
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
func (t TextArea) OnMouseMove(event MouseEvent) {
	if t.State == nil || t.State.lastWidth <= 0 {
		return
	}

	// Update cursor position; selection extends from anchor
	contentWidth := reservedContentWidth(t.State.lastWidth)
	t.State.SetCursorFromLocalPosition(event.LocalX, event.LocalY, contentWidth)
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (t TextArea) OnMouseUp(event MouseEvent) {
	if t.MouseUp != nil {
		t.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
func (t TextArea) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// OnBlur is called when the widget loses focus.
func (t TextArea) OnBlur() {
	if t.RequireInsertMode && t.State != nil {
		t.State.InsertMode.Set(false)
	}
	if t.State != nil {
		t.State.lastFocused = false
	}
}
