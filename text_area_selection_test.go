package terma

import "testing"

func TestTextAreaState_Selection_Basic(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Initial state should have no selection
	if state.HasSelection() {
		t.Error("expected no selection initially")
	}

	// Set anchor at 0, cursor at 5 ("hello" selected)
	state.SetSelectionAnchor(0)
	state.CursorIndex.Set(5)

	if !state.HasSelection() {
		t.Error("expected selection to be active")
	}

	start, end := state.GetSelectionBounds()
	if start != 0 || end != 5 {
		t.Errorf("expected selection bounds (0, 5), got (%d, %d)", start, end)
	}

	selected := state.GetSelectedText()
	if selected != "hello" {
		t.Errorf("expected selected text %q, got %q", "hello", selected)
	}
}

func TestTextAreaState_Selection_Reversed(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Set anchor at 11 (end), cursor at 6 ("world" selected, but cursor < anchor)
	state.SetSelectionAnchor(11)
	state.CursorIndex.Set(6)

	if !state.HasSelection() {
		t.Error("expected selection to be active")
	}

	start, end := state.GetSelectionBounds()
	if start != 6 || end != 11 {
		t.Errorf("expected selection bounds (6, 11), got (%d, %d)", start, end)
	}

	selected := state.GetSelectedText()
	if selected != "world" {
		t.Errorf("expected selected text %q, got %q", "world", selected)
	}
}

func TestTextAreaState_SelectWord(t *testing.T) {
	state := NewTextAreaState("hello world test")

	// Select word at position 7 (inside "world")
	state.SelectWord(7)

	selected := state.GetSelectedText()
	if selected != "world" {
		t.Errorf("expected selected text %q, got %q", "world", selected)
	}

	start, end := state.GetSelectionBounds()
	if start != 6 || end != 11 {
		t.Errorf("expected selection bounds (6, 11), got (%d, %d)", start, end)
	}
}

func TestTextAreaState_SelectWord_AtBoundary(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Select word at start of "hello"
	state.SelectWord(0)

	selected := state.GetSelectedText()
	if selected != "hello" {
		t.Errorf("expected selected text %q, got %q", "hello", selected)
	}
}

func TestTextAreaState_SelectLine(t *testing.T) {
	state := NewTextAreaState("first line\nsecond line\nthird line")

	// Select line at position 15 (inside "second line")
	state.SelectLine(15)

	selected := state.GetSelectedText()
	// Should include the newline at the end
	if selected != "second line\n" {
		t.Errorf("expected selected text %q, got %q", "second line\n", selected)
	}
}

func TestTextAreaState_SelectLine_LastLine(t *testing.T) {
	state := NewTextAreaState("first\nlast")

	// Select last line (no trailing newline)
	state.SelectLine(8)

	selected := state.GetSelectedText()
	if selected != "last" {
		t.Errorf("expected selected text %q, got %q", "last", selected)
	}
}

func TestTextAreaState_DeleteSelection(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Select "world" (positions 6-11)
	state.SetSelectionAnchor(6)
	state.CursorIndex.Set(11)

	// Delete selection
	deleted := state.DeleteSelection()
	if !deleted {
		t.Error("expected DeleteSelection to return true")
	}

	// Verify text
	text := state.GetText()
	if text != "hello " {
		t.Errorf("expected text %q, got %q", "hello ", text)
	}

	// Verify cursor position
	cursor := state.CursorIndex.Peek()
	if cursor != 6 {
		t.Errorf("expected cursor at 6, got %d", cursor)
	}

	// Verify selection is cleared
	if state.HasSelection() {
		t.Error("expected selection to be cleared after delete")
	}
}

func TestTextAreaState_DeleteSelection_Reversed(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Select "hello" with anchor at end (reversed)
	state.SetSelectionAnchor(5)
	state.CursorIndex.Set(0)

	// Delete selection
	deleted := state.DeleteSelection()
	if !deleted {
		t.Error("expected DeleteSelection to return true")
	}

	// Verify text
	text := state.GetText()
	if text != " world" {
		t.Errorf("expected text %q, got %q", " world", text)
	}

	// Verify cursor is at start of deleted region
	cursor := state.CursorIndex.Peek()
	if cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", cursor)
	}
}

func TestTextAreaState_DeleteSelection_NoSelection(t *testing.T) {
	state := NewTextAreaState("hello world")

	// No selection
	deleted := state.DeleteSelection()
	if deleted {
		t.Error("expected DeleteSelection to return false when no selection")
	}

	// Verify text unchanged
	text := state.GetText()
	if text != "hello world" {
		t.Errorf("expected text unchanged, got %q", text)
	}
}

func TestTextAreaState_ReplaceSelection(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Select "world"
	state.SetSelectionAnchor(6)
	state.CursorIndex.Set(11)

	// Replace with "there"
	state.ReplaceSelection("there")

	// Verify text
	text := state.GetText()
	if text != "hello there" {
		t.Errorf("expected text %q, got %q", "hello there", text)
	}

	// Verify cursor is after inserted text
	cursor := state.CursorIndex.Peek()
	if cursor != 11 { // "hello " (6) + "there" (5) = 11
		t.Errorf("expected cursor at 11, got %d", cursor)
	}
}

func TestTextAreaState_ReplaceSelection_NoSelection(t *testing.T) {
	state := NewTextAreaState("hello world")
	state.CursorIndex.Set(5) // After "hello"

	// No selection - should just insert
	state.ReplaceSelection(" there")

	// Verify text
	text := state.GetText()
	if text != "hello there world" {
		t.Errorf("expected text %q, got %q", "hello there world", text)
	}
}

func TestTextAreaState_SelectAll(t *testing.T) {
	state := NewTextAreaState("hello world")

	state.SelectAll()

	if !state.HasSelection() {
		t.Error("expected selection to be active")
	}

	start, end := state.GetSelectionBounds()
	if start != 0 || end != 11 {
		t.Errorf("expected selection bounds (0, 11), got (%d, %d)", start, end)
	}

	selected := state.GetSelectedText()
	if selected != "hello world" {
		t.Errorf("expected selected text %q, got %q", "hello world", selected)
	}
}

func TestTextAreaState_ClearSelection(t *testing.T) {
	state := NewTextAreaState("hello world")

	// Create a selection
	state.SetSelectionAnchor(0)
	state.CursorIndex.Set(5)

	if !state.HasSelection() {
		t.Error("expected selection to be active")
	}

	// Clear it
	state.ClearSelection()

	if state.HasSelection() {
		t.Error("expected selection to be cleared")
	}

	// Verify bounds return -1, -1
	start, end := state.GetSelectionBounds()
	if start != -1 || end != -1 {
		t.Errorf("expected selection bounds (-1, -1), got (%d, %d)", start, end)
	}
}

func TestTextAreaState_HasSelection_SamePosition(t *testing.T) {
	state := NewTextAreaState("hello")

	// Anchor and cursor at same position = no selection
	state.SetSelectionAnchor(2)
	state.CursorIndex.Set(2)

	if state.HasSelection() {
		t.Error("expected no selection when anchor equals cursor")
	}
}
