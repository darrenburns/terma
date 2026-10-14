package terma

import (
	"testing"

	uv "github.com/darrenburns/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextAreaBeforeInputEditsAndCaretAreOneUndoStep(t *testing.T) {
	for _, tc := range []struct {
		name, initial, input string
		cursor, anchor       int
		transform            func(TextAreaEdit) TextAreaEdit
		want                 caretWant
		wantAnchor           int
	}{
		{"pair after graphemes", "猫e\u0301", `"`, 2, -1, func(e TextAreaEdit) TextAreaEdit {
			e.Text = `""`
			return e
		}, caretWant{text: "猫e\u0301\"\"", cursor: 3}, -1},
		{"wrap reverse selection", "a猫e\u0301z", `"`, 1, 3, func(e TextAreaEdit) TextAreaEdit {
			e.Text = "\"猫e\u0301\""
			e.Cursor, e.SelectionAnchor = 2, 4
			return e
		}, caretWant{text: "a\"猫e\u0301\"z", cursor: 2, selected: "猫e\u0301"}, 4},
		{"newline and indentation", "{}", "\n", 1, -1, func(e TextAreaEdit) TextAreaEdit {
			e.Text, e.Cursor = "\n  \n", 4
			return e
		}, caretWant{text: "{\n  \n}", cursor: 4}, -1},
		{"clamp returned positions", "abc", "x", 1, -1, func(e TextAreaEdit) TextAreaEdit {
			e.Start, e.End, e.Cursor, e.SelectionAnchor = -2, 99, 100, 90
			return e
		}, caretWant{text: "x", cursor: 1}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewTextAreaState(tc.initial)
			s.CursorIndex.Set(tc.cursor)
			s.SelectionAnchor.Set(tc.anchor)
			before := caretWant{s.GetText(), tc.cursor, s.GetSelectedText()}
			var changes []string
			area := TextArea{State: s, BeforeInput: tc.transform, OnChange: func(text string) { changes = append(changes, text) }}
			keys := editorDriver{t: t, widget: area, text: s.GetText}
			keys.typeText(tc.input)
			assertTextArea(t, s, tc.want)
			assert.Equal(t, tc.wantAnchor, s.SelectionAnchor.Peek())
			assert.Equal(t, []string{tc.want.text}, changes)
			require.True(t, keys.undo())
			assertTextArea(t, s, before)
			assert.Equal(t, tc.anchor, s.SelectionAnchor.Peek())
			assert.False(t, s.CanUndo())
			require.True(t, keys.redo())
			assertTextArea(t, s, tc.want)
			assert.Equal(t, tc.wantAnchor, s.SelectionAnchor.Peek())
			assert.Equal(t, []string{tc.want.text, tc.initial, tc.want.text}, changes)
		})
	}
}

func TestTextAreaBeforeInputOvertypeAndSuppressionPreserveHistory(t *testing.T) {
	s := NewTextAreaState(`""`)
	s.CursorIndex.Set(1)
	var changes []string
	area := TextArea{State: s, OnChange: func(text string) { changes = append(changes, text) },
		BeforeInput: func(e TextAreaEdit) TextAreaEdit {
			switch e.Text {
			case `"`: // Overtype the existing closer.
				e.Text = ""
			case "!": // Reject input without moving the cursor.
				e.Text, e.Cursor = "", e.Start
			}
			return e
		}}
	keys := editorDriver{t: t, widget: area, text: s.GetText}
	keys.typeText(`"!`)
	assertTextArea(t, s, caretWant{text: `""`, cursor: 2})
	assert.Empty(t, changes)
	assert.False(t, s.CanUndo())
	keys.typeText("abc")
	require.True(t, keys.undo(), "unchanged input still groups normal typing")
	assertTextArea(t, s, caretWant{text: `""`, cursor: 2})
	keys.press(uv.KeyLeft, 0)
	keys.typeText(`"!`)
	require.True(t, keys.redo(), "cursor-only input must not discard redo")
	assertTextArea(t, s, caretWant{text: `""abc`, cursor: 5})
}

func TestTextAreaBeforeInputDoesNotTransformPasteOrDisabledInput(t *testing.T) {
	s := NewTextAreaState("")
	area := TextArea{State: s, RequireInsertMode: true, BeforeInput: func(e TextAreaEdit) TextAreaEdit {
		e.Text = "transformed"
		return e
	}}
	keys := editorDriver{t: t, widget: area, text: s.GetText}
	require.True(t, area.HandlePaste("{\n  \"x\": 1\n}"))
	assert.Equal(t, "{\n  \"x\": 1\n}", s.GetText())
	require.True(t, keys.undo())
	require.True(t, keys.redo())
	assert.Equal(t, "{\n  \"x\": 1\n}", s.GetText())
	keys.press(uv.KeyBackspace, 0)
	assert.Equal(t, "{\n  \"x\": 1\n", s.GetText(), "deletion bypasses the input hook")
	for _, readOnly := range []bool{false, true} {
		s.SetText("")
		s.ReadOnly.Set(readOnly)
		s.InsertMode.Set(false)
		assert.False(t, area.OnKey(KeyEvent{event: uv.KeyPressEvent{Code: 'x', Text: "x"}}))
		keys.press(uv.KeyEnter, 0) // May enter insert mode; must not insert yet.
		assert.Equal(t, "", s.GetText())
		assert.False(t, s.CanUndo())
	}
}
