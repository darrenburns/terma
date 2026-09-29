package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePastedText(t *testing.T) {
	assert.Equal(t, "a\nb\nc\n", normalizePastedText("a\r\nb\rc\n"))
}

func TestTextInput_HandlePaste_InsertsAsOneEdit(t *testing.T) {
	state := NewTextInputState("hello world")
	state.SetSelectionAnchor(6)
	state.CursorIndex.Set(11) // "world" selected
	var changes []string
	input := TextInput{State: state, OnChange: func(text string) { changes = append(changes, text) }}

	assert.True(t, input.HandlePaste("there"))

	assert.Equal(t, "hello there", state.GetText())
	assert.Equal(t, 11, state.CursorIndex.Peek())
	assert.Equal(t, []string{"hello there"}, changes, "OnChange fires once for the whole paste")
}

func TestTextInput_HandlePaste_FlattensLinesWithoutSubmitting(t *testing.T) {
	state := NewTextInputState("")
	submitted := false
	input := TextInput{State: state, OnSubmit: func(string) { submitted = true }}

	assert.True(t, input.HandlePaste("curl https://example.com \\\n  -H 'a:\tb'\n"))

	assert.Equal(t, "curl https://example.com \\   -H 'a: b'", state.GetText())
	assert.False(t, submitted, "a pasted newline is not an Enter")
}

func TestTextInput_HandlePaste_OnPasteCanConsume(t *testing.T) {
	state := NewTextInputState("keep")
	var seen string
	input := TextInput{State: state, OnPaste: func(text string) bool {
		seen = text
		return true
	}}

	assert.True(t, input.HandlePaste("curl -X POST\nhttps://x"))

	assert.Equal(t, "curl -X POST\nhttps://x", seen, "OnPaste sees the text with its newlines")
	assert.Equal(t, "keep", state.GetText())
}

func TestTextInput_HandlePaste_OnPasteCanDecline(t *testing.T) {
	state := NewTextInputState("")
	input := TextInput{State: state, OnPaste: func(string) bool { return false }}

	assert.True(t, input.HandlePaste("abc"))
	assert.Equal(t, "abc", state.GetText())
}

func TestTextInput_HandlePaste_ReadOnlyIgnoresPaste(t *testing.T) {
	state := NewTextInputState("fixed")
	state.ReadOnly.Set(true)
	assert.False(t, TextInput{State: state}.HandlePaste("x"))
	assert.Equal(t, "fixed", state.GetText())
}

func TestTextArea_HandlePaste_KeepsNewlines(t *testing.T) {
	state := NewTextAreaState("{}")
	state.CursorIndex.Set(1)
	changes := 0
	area := TextArea{State: state, OnChange: func(string) { changes++ }}

	assert.True(t, area.HandlePaste("\n\t\"a\": 1\n"))

	assert.Equal(t, "{\n    \"a\": 1\n}", state.GetText())
	assert.Equal(t, 1, changes)
}

func TestTextArea_HandlePaste_OnPasteCanConsume(t *testing.T) {
	state := NewTextAreaState("")
	area := TextArea{State: state, OnPaste: func(string) bool { return true }}
	assert.True(t, area.HandlePaste("x"))
	assert.Equal(t, "", state.GetText())
}

func TestTextArea_HandlePaste_NormalModeIgnoresPaste(t *testing.T) {
	state := NewTextAreaState("")
	state.InsertMode.Set(false)
	area := TextArea{State: state, RequireInsertMode: true}
	assert.False(t, area.HandlePaste("x"))
	assert.Equal(t, "", state.GetText())
}

type pasteRecorder struct {
	got    *[]string
	handle bool
}

func (p pasteRecorder) Build(ctx BuildContext) Widget { return p }
func (p pasteRecorder) HandlePaste(text string) bool {
	*p.got = append(*p.got, text)
	return p.handle
}

func TestDispatchPaste_GoesToFocusedWidget(t *testing.T) {
	state := NewTextInputState("")
	fm := NewFocusManager()
	fm.SetFocusables([]FocusableEntry{{ID: "input", Focusable: TextInput{ID: "input", State: state}}})
	fm.focusedID = "input"

	assert.True(t, dispatchPaste(fm, nil, "a\r\nb"))
	assert.Equal(t, "a b", state.GetText())
}

func TestDispatchPaste_BubblesToAncestorsThenRoot(t *testing.T) {
	var ancestorGot, rootGot []string
	fm := NewFocusManager()
	// A focused widget that doesn't take pastes (a List) under an ancestor
	// that does.
	fm.SetFocusables([]FocusableEntry{{
		ID:        "list",
		Focusable: List[string]{ID: "list", State: NewListState([]string{"a"})},
		Ancestors: []Widget{pasteRecorder{got: &ancestorGot, handle: false}},
	}})
	fm.focusedID = "list"
	root := pasteRecorder{got: &rootGot, handle: true}

	assert.True(t, dispatchPaste(fm, root, "text\r"))
	assert.Equal(t, []string{"text\n"}, ancestorGot)
	assert.Equal(t, []string{"text\n"}, rootGot)
}

func TestDispatchPaste_UnhandledPasteIsDropped(t *testing.T) {
	fm := NewFocusManager()
	assert.False(t, dispatchPaste(fm, Text{Content: "root"}, "q"))
}

func TestFocusCollector_TracksPasteHandlerAncestors(t *testing.T) {
	fc := NewFocusCollector()
	var got []string
	assert.True(t, fc.ShouldTrackAncestor(pasteRecorder{got: &got}))
}

func TestSnapshot_TextArea_AfterMultilinePaste(t *testing.T) {
	state := NewTextAreaState("")
	TextArea{State: state}.HandlePaste("{\r\n\t\"name\": \"posting\",\r\n\t\"id\": 3\r\n}")
	widget := TextArea{ID: "body", State: state, Style: Style{Width: Cells(24), Height: Cells(5)}}
	AssertSnapshot(t, widget, 24, 5,
		"A pasted four-line JSON object: CRLF line endings became newlines and tabs became four spaces. Cursor after the closing brace.")
}
