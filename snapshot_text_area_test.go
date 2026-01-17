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
