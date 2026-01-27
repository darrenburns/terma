package terma

import "testing"

func TestSnapshot_TextArea_WrapOn(t *testing.T) {
	state := NewTextAreaState("First line\nSecond line is long enough to wrap")
	state.WrapMode.Set(WrapSoft)
	state.CursorIndex.Set(0)

	widget := TextArea{
		ID:     "textarea-wrap-on",
		State:  state,
		Width:  Cells(12),
		Height: Cells(4),
	}

	AssertSnapshot(t, widget, 12, 4,
		"TextArea with wrapping enabled. First line on row 1, second line wraps to additional rows. Cursor at start.")
}

func TestSnapshot_TextArea_WrapOff(t *testing.T) {
	state := NewTextAreaState("0123456789ABCDEF")
	state.WrapMode.Set(WrapNone)
	state.CursorIndex.Set(len(splitGraphemes(state.GetText())))

	widget := TextArea{
		ID:     "textarea-wrap-off",
		State:  state,
		Width:  Cells(10),
		Height: Cells(2),
	}

	AssertSnapshot(t, widget, 10, 2,
		"TextArea with wrapping disabled. Long line scrolls horizontally so the cursor at the end is visible.")
}

func TestSnapshot_TextArea_Selection(t *testing.T) {
	state := NewTextAreaState("hello world")
	state.WrapMode.Set(WrapSoft)
	state.SetSelectionAnchor(0)
	state.CursorIndex.Set(5) // "hello" selected

	widget := TextArea{
		ID:     "textarea-selection",
		State:  state,
		Width:  Cells(15),
		Height: Cells(2),
	}

	AssertSnapshot(t, widget, 15, 2,
		"TextArea with 'hello' selected using theme Selection colors.")
}

func TestSnapshot_TextArea_Selection_MultiLine(t *testing.T) {
	state := NewTextAreaState("first line\nsecond line\nthird line")
	state.WrapMode.Set(WrapSoft)
	state.SetSelectionAnchor(6) // Start at "line" on first row
	state.CursorIndex.Set(18)   // End at "d li" on second row

	widget := TextArea{
		ID:     "textarea-selection-multiline",
		State:  state,
		Width:  Cells(15),
		Height: Cells(4),
	}

	AssertSnapshot(t, widget, 15, 4,
		"TextArea with multi-line selection spanning from 'line' on first row through part of second row.")
}
