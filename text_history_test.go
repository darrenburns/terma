package terma

import (
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeClock is a manually advanced clock for undo grouping.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time         { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

type editorWidget interface {
	KeybindProvider
	OnKey(KeyEvent) bool
	HandlePaste(string) bool
}

// editorDriver sends keys to a text widget the way dispatchKey does:
// keybinds first, then OnKey.
type editorDriver struct {
	t      *testing.T
	widget editorWidget
	text   func() string
}

func (d editorDriver) press(code rune, mod uv.KeyMod) bool {
	event := KeyEvent{event: uv.KeyPressEvent{Code: code, Mod: mod}}
	if matchKeybind(event, d.widget.Keybinds()) {
		return true
	}
	return d.widget.OnKey(event)
}

func (d editorDriver) typeText(text string) {
	for _, r := range text {
		event := KeyEvent{event: uv.KeyPressEvent{Code: r, Text: string(r)}}
		if r == '\n' {
			event = KeyEvent{event: uv.KeyPressEvent{Code: uv.KeyEnter}}
		}
		if !matchKeybind(event, d.widget.Keybinds()) {
			require.True(d.t, d.widget.OnKey(event), "typed %q", r)
		}
	}
}

// changed presses a key and reports whether it changed the text.
func (d editorDriver) changed(code rune, mod uv.KeyMod) bool {
	before := d.text()
	d.press(code, mod)
	return d.text() != before
}

func (d editorDriver) undo() bool { return d.changed('z', uv.ModCtrl) }
func (d editorDriver) redo() bool { return d.changed('z', uv.ModCtrl|uv.ModShift) }

func newTextAreaDriver(t *testing.T, initial string) (*TextAreaState, *fakeClock, editorDriver) {
	state := NewTextAreaState(initial)
	clock := &fakeClock{t: time.Unix(0, 0)}
	state.history.now = clock.now
	return state, clock, editorDriver{t: t, widget: TextArea{State: state}, text: state.GetText}
}

func newTextInputDriver(t *testing.T, initial string) (*TextInputState, *fakeClock, editorDriver) {
	state := NewTextInputState(initial)
	clock := &fakeClock{t: time.Unix(0, 0)}
	state.history.now = clock.now
	return state, clock, editorDriver{t: t, widget: TextInput{State: state}, text: state.GetText}
}

type caretWant struct {
	text     string
	cursor   int
	selected string
}

func assertTextArea(t *testing.T, state *TextAreaState, want caretWant) {
	t.Helper()
	assert.Equal(t, want, caretWant{state.GetText(), state.CursorIndex.Peek(), state.GetSelectedText()})
}

func assertTextInput(t *testing.T, state *TextInputState, want caretWant) {
	t.Helper()
	assert.Equal(t, want, caretWant{state.GetText(), state.CursorIndex.Peek(), state.GetSelectedText()})
}

func TestTextAreaUndo_TypingGroupsByWord(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("hello world")

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "hello ", cursor: 6})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "", cursor: 0})
	assert.False(t, state.CanUndo())

	require.True(t, keys.redo())
	assertTextArea(t, state, caretWant{text: "hello ", cursor: 6})
	require.True(t, keys.changed('y', uv.ModCtrl), "ctrl+y redoes too")
	assertTextArea(t, state, caretWant{text: "hello world", cursor: 11})
	assert.False(t, state.CanRedo())
}

func TestTextAreaUndo_NewlineEndsTheWordBeforeIt(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("{\n  \"a\": 1\n}")

	var steps []string
	for keys.undo() {
		steps = append(steps, state.GetText())
	}
	assert.Equal(t, []string{"{\n  \"a\": 1\n", "{\n  \"a\": ", "{\n  ", ""}, steps)
}

func TestTextAreaUndo_PauseStartsNewStep(t *testing.T) {
	state, clock, keys := newTextAreaDriver(t, "")
	keys.typeText("ab")
	clock.advance(500 * time.Millisecond)
	keys.typeText("c")
	clock.advance(2 * time.Second)
	keys.typeText("d")

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "abc", cursor: 3})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "", cursor: 0})
}

func TestTextAreaUndo_CursorMoveStartsNewStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("abc")
	keys.press(uv.KeyLeft, 0)
	keys.typeText("X")
	assertTextArea(t, state, caretWant{text: "abXc", cursor: 3})

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "abc", cursor: 2})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "", cursor: 0})
}

func TestTextAreaUndo_ClickStartsNewStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	area := keys.widget.(TextArea)
	keys.typeText("abc")
	state.lastWidth = 20
	area.OnMouseDown(MouseEvent{LocalX: 3, LocalY: 0, ClickCount: 1})
	require.Equal(t, 3, state.CursorIndex.Peek(), "clicked where typing left off")
	keys.typeText("d")

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "abc", cursor: 3})
}

func TestTextAreaUndo_SwitchingBetweenTypingAndDeletingStartsNewStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("abc")
	keys.press(uv.KeyBackspace, 0)
	keys.press(uv.KeyBackspace, 0)
	keys.typeText("d")
	assertTextArea(t, state, caretWant{text: "ad", cursor: 2})

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "a", cursor: 1})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "abc", cursor: 3})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "", cursor: 0})
}

func TestTextAreaUndo_BackspaceGroupsByWord(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "one two")
	for range 7 {
		keys.press(uv.KeyBackspace, 0)
	}
	require.Equal(t, "", state.GetText())

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "one", cursor: 3})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "one two", cursor: 7})
	assert.False(t, state.CanUndo(), "the initial text is not an undo step")
}

func TestTextAreaUndo_ForwardDeleteGroups(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "abc")
	state.CursorIndex.Set(0)
	keys.press(uv.KeyDelete, 0)
	keys.press(uv.KeyDelete, 0)
	assertTextArea(t, state, caretWant{text: "c", cursor: 0})

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "abc", cursor: 0})
	assert.False(t, state.CanUndo())
}

func TestTextAreaUndo_PasteIsOneStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("a")
	require.True(t, keys.widget.HandlePaste("one two\nthree"))
	keys.typeText("b")

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "aone two\nthree", cursor: 14})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "a", cursor: 1})
}

func TestTextAreaUndo_DeleteSelectionRestoresSelection(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "hello world")
	for range 5 {
		keys.press(uv.KeyLeft, uv.ModShift)
	}
	require.Equal(t, "world", state.GetSelectedText())
	keys.press(uv.KeyBackspace, 0)
	assertTextArea(t, state, caretWant{text: "hello ", cursor: 6})

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "hello world", cursor: 6, selected: "world"})
	require.True(t, keys.redo())
	assertTextArea(t, state, caretWant{text: "hello ", cursor: 6})
}

func TestTextAreaUndo_TypingOverSelectionIsOneStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "hello world")
	for range 5 {
		keys.press(uv.KeyLeft, uv.ModShift)
	}
	keys.typeText("there")
	assertTextArea(t, state, caretWant{text: "hello there", cursor: 11})

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "hello world", cursor: 6, selected: "world"})
}

func TestTextAreaUndo_ProgrammaticEditsAreSingleSteps(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	state.Insert("a")
	state.Insert("b")
	state.ReplaceText("ab = 1", 2)

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "ab", cursor: 2})
	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "a", cursor: 1})
	require.True(t, keys.redo())
	require.True(t, keys.redo())
	assertTextArea(t, state, caretWant{text: "ab = 1", cursor: 2})
}

func TestTextAreaUndo_NewEditClearsRedo(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("one")
	require.True(t, keys.undo())
	keys.typeText("two")

	assert.False(t, keys.redo())
	assert.False(t, state.CanRedo())
	assertTextArea(t, state, caretWant{text: "two", cursor: 3})
}

func TestTextAreaUndo_EditAfterUndoStartsNewStep(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("ab")
	require.True(t, keys.undo())
	require.True(t, keys.redo())
	keys.typeText("c")

	require.True(t, keys.undo())
	assertTextArea(t, state, caretWant{text: "ab", cursor: 2})
}

func TestTextAreaUndo_SetTextClearsHistory(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("request one")
	state.SetText("request two")

	assert.False(t, state.CanUndo())
	assert.False(t, keys.undo())
	assert.Equal(t, "request two", state.GetText())

	keys.typeText("!")
	require.True(t, keys.undo())
	assert.Equal(t, "request two", state.GetText())
	assert.False(t, keys.undo(), "undo stops at the loaded text")
}

func TestTextAreaUndo_DirectContentWriteClearsHistory(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("abc")
	state.Content.Set(splitGraphemes("x"))

	assert.False(t, keys.undo())
	assert.Equal(t, "x", state.GetText())
}

func TestTextAreaUndo_OnChangeSeesUndoAndRedo(t *testing.T) {
	state := NewTextAreaState("")
	var changes []string
	keys := editorDriver{t: t, widget: TextArea{State: state, OnChange: func(text string) { changes = append(changes, text) }}, text: state.GetText}
	keys.typeText("hi")
	keys.undo()
	keys.redo()
	keys.undo()
	keys.undo()

	assert.Equal(t, []string{"h", "hi", "", "hi", ""}, changes, "an undo with nothing to undo doesn't report a change")
}

func TestTextAreaUndo_ReadOnlyIgnoresUndo(t *testing.T) {
	state, _, keys := newTextAreaDriver(t, "")
	keys.typeText("abc")
	state.ReadOnly.Set(true)

	assert.False(t, keys.undo())
	assert.Equal(t, "abc", state.GetText())
	assert.False(t, keys.widget.(TextArea).CapturesKey("ctrl+z"), "ctrl+z reaches the app when the text can't change")
}

func TestTextAreaUndo_CapturesUndoKeysSoTheAppDoesNotSuspend(t *testing.T) {
	area := TextArea{State: NewTextAreaState("")}
	for _, key := range []string{"ctrl+z", "ctrl+shift+z", "ctrl+y"} {
		assert.True(t, area.CapturesKey(key), key)
	}
	assert.Equal(t, "ctrl+z", KeyEvent{event: uv.KeyPressEvent{Code: 'z', Mod: uv.ModCtrl}}.Key())
	assert.Equal(t, "ctrl+shift+z", KeyEvent{event: uv.KeyPressEvent{Code: 'z', Mod: uv.ModCtrl | uv.ModShift}}.Key())
}

func TestTextAreaUndo_ConfigurableKeys(t *testing.T) {
	state := NewTextAreaState("")
	keys := editorDriver{t: t, widget: TextArea{State: state, UndoKeys: []string{"ctrl+u"}, RedoKeys: []string{}}, text: state.GetText}
	keys.typeText("abc")

	assert.False(t, keys.undo(), "ctrl+z is no longer bound")
	keys.press(uv.KeyLeft, 0)
	require.True(t, keys.changed('u', uv.ModCtrl))
	assert.Equal(t, "", state.GetText(), "ctrl+u undoes instead of deleting to line start")
	assert.False(t, keys.changed('y', uv.ModCtrl), "an empty RedoKeys turns redo off")
	assert.Equal(t, "", state.GetText())

	area := keys.widget.(TextArea)
	assert.True(t, area.CapturesKey("ctrl+u"))
	assert.False(t, area.CapturesKey("ctrl+z"))
}

func TestTextAreaUndo_HistoryIsCapped(t *testing.T) {
	state, _, _ := newTextAreaDriver(t, "")
	for range maxUndoSteps + 20 {
		state.Insert("x")
	}
	undone := 0
	for state.Undo() {
		undone++
	}
	assert.Equal(t, maxUndoSteps, undone)
	assert.Equal(t, strings.Repeat("x", 20), state.GetText())

	state.SetText("")
	state.Insert("a")
	state.Insert(strings.Repeat("b", maxUndoGraphemes))
	require.True(t, state.Undo(), "a step larger than the cap is still kept")
	assert.False(t, state.Undo(), "older steps are dropped to make room")
	assert.Equal(t, "a", state.GetText())
}

func TestTextInputUndo_TypingAndRedo(t *testing.T) {
	state, _, keys := newTextInputDriver(t, "")
	keys.typeText("hello world")
	keys.press(uv.KeyBackspace, 0)

	require.True(t, keys.undo())
	assertTextInput(t, state, caretWant{text: "hello world", cursor: 11})
	require.True(t, keys.undo())
	assertTextInput(t, state, caretWant{text: "hello ", cursor: 6})
	require.True(t, keys.redo())
	assertTextInput(t, state, caretWant{text: "hello world", cursor: 11})
	assert.True(t, keys.widget.(TextInput).CapturesKey("ctrl+z"))
}

func TestTextInputUndo_DeleteSelectionRestoresSelection(t *testing.T) {
	state, _, keys := newTextInputDriver(t, "key=value")
	state.SelectAll()
	keys.press(uv.KeyBackspace, 0)
	require.Equal(t, "", state.GetText())

	require.True(t, keys.undo())
	assertTextInput(t, state, caretWant{text: "key=value", cursor: 9, selected: "key=value"})
}

func TestTextInputUndo_SetTextClearsHistory(t *testing.T) {
	state, _, keys := newTextInputDriver(t, "")
	keys.typeText("abc")
	state.SetText("other")
	assert.False(t, keys.undo())
	assert.Equal(t, "other", state.GetText())
}
