package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// chordApp is an editor-like app: chords on the root, single keys and a chord
// sharing their start on a focused panel, and a text input to type into.
type chordApp struct {
	log     *[]string
	input   *TextInputState
	focusOn string
}

func (a *chordApp) record(name string) func() {
	return func() { *a.log = append(*a.log, name) }
}

func (a *chordApp) Build(ctx BuildContext) Widget {
	return Column{Children: []Widget{
		chordPanel{app: a},
		TextInput{ID: "input", State: a.input},
		KeybindBar{},
	}}
}

func (a *chordApp) Keybinds() []Keybind {
	return []Keybind{
		{Key: "ctrl+s", Name: "Save", Action: a.record("save")},
		{Key: "ctrl+k", Name: "Window"},
		{Key: "ctrl+k ctrl+s", Name: "Save all", Action: a.record("save all")},
		{Key: "ctrl+k w", Name: "Close all", Action: a.record("close all")},
		{Key: "ctrl+k ctrl+c", Name: "Comment", Action: a.record("comment")},
		{Key: "space f", Name: "Find", Action: a.record("find")},
		{Key: "ctrl+g x", Name: "Run", Action: a.record("ctrl+g x"), Hidden: true},
		{Key: "q", Name: "Quit", Action: a.record("quit")},
	}
}

type chordPanel struct{ app *chordApp }

func (p chordPanel) Build(BuildContext) Widget {
	return Text{Content: "panel"}
}

func (p chordPanel) WidgetID() string          { return "panel" }
func (p chordPanel) IsFocusable() bool         { return true }
func (p chordPanel) OnKey(event KeyEvent) bool { return false }

func (p chordPanel) Keybinds() []Keybind {
	return []Keybind{
		{Key: "d", Name: "Delete", Action: p.app.record("d")},
		{Key: "D", Name: "Delete all", Action: p.app.record("D")},
		{Key: "g", Name: "Go", Action: p.app.record("g")},
		{Key: "g g", Name: "Top", Action: p.app.record("g g")},
		{Key: "g e", Name: "End", Action: p.app.record("g e")},
	}
}

func newChordPilot(t *testing.T) (*Pilot, *[]string, *chordApp) {
	log := &[]string{}
	app := &chordApp{log: log, input: NewTextInputState("")}
	p := NewPilot(t, app, 70, 4)
	assert.Equal(t, "panel", p.FocusedID())
	return p, log, app
}

func TestKeybind_CharacterCaseMatters(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("d")
	p.Press("shift+d")
	p.Type("D")
	assert.Equal(t, []string{"d", "D", "D"}, *log)
}

func TestChord_RunsAfterItsLastKey(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("ctrl+k")
	assert.Empty(t, *log, "the first key of a chord only waits")
	assert.True(t, p.session.focus.chordPending())
	p.Press("w")
	assert.Equal(t, []string{"close all"}, *log)
	assert.False(t, p.session.focus.chordPending())

	p.Press("ctrl+k", "ctrl+s", "ctrl+s")
	assert.Equal(t, []string{"close all", "save all", "save"}, *log)
}

func TestChord_WinsOverItsOwnStart(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("g")
	assert.Empty(t, *log)
	p.Press("g")
	p.Press("g", "e")
	assert.Equal(t, []string{"g g", "g e"}, *log)
}

func TestChord_EscapeCancels(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("ctrl+k", "escape")
	assert.False(t, p.session.focus.chordPending())
	p.Press("q")
	assert.Equal(t, []string{"quit"}, *log)
}

func TestChord_UnboundKeyCancelsAndIsDropped(t *testing.T) {
	p, log, _ := newChordPilot(t)
	// q is bound on its own, but after ctrl+k it continues no chord.
	p.Press("ctrl+k", "q")
	assert.Empty(t, *log)
	assert.False(t, p.session.focus.chordPending())
	p.Press("q")
	assert.Equal(t, []string{"quit"}, *log)
}

func TestChord_CanContinueWithCtrlC(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("ctrl+k", "ctrl+c")
	assert.False(t, p.Exited())
	assert.Equal(t, []string{"comment"}, *log)
}

func TestChord_ClickCancels(t *testing.T) {
	p, log, _ := newChordPilot(t)
	p.Press("ctrl+k")
	p.Click("input")
	assert.False(t, p.session.focus.chordPending())
	p.Press("w")
	assert.Empty(t, *log)
}

func TestChord_TextInputTypesKeysAncestorChordsStartWith(t *testing.T) {
	p, log, app := newChordPilot(t)
	p.Click("input")
	// space starts "space f" on the root, but the input types it.
	p.Type(" f")
	assert.Equal(t, " f", app.input.GetText())
	assert.Empty(t, *log)

	// The input binds ctrl+k itself (delete to end), so it takes ctrl+k
	// before the root's chords see it.
	p.Press("ctrl+k", "w")
	assert.Equal(t, " fw", app.input.GetText())
	assert.Empty(t, *log)

	// A chord started with a key the input doesn't bind finishes with one it
	// would type: the key completes the chord instead of being typed.
	p.Press("ctrl+g", "x")
	assert.Equal(t, " fw", app.input.GetText())
	assert.Equal(t, []string{"ctrl+g x"}, *log)
}

func TestChord_ActiveKeybindsListContinuations(t *testing.T) {
	p, _, _ := newChordPilot(t)
	p.Press("ctrl+k")
	var keys []string
	for _, kb := range p.Keybinds() {
		keys = append(keys, kb.Key)
	}
	assert.Equal(t, []string{"ctrl+s", "w", "ctrl+c"}, keys)
}

func TestChord_ActiveKeybindsLeaveOutUnreachable(t *testing.T) {
	p, _, _ := newChordPilot(t)
	keys := func() []string {
		var keys []string
		for _, kb := range p.Keybinds() {
			keys = append(keys, kb.Key)
		}
		return keys
	}
	// Plain g never runs: the panel's g chords win.
	assert.NotContains(t, keys(), "g")
	assert.Contains(t, keys(), "g g")
	assert.Contains(t, keys(), "ctrl+k", "a group label stays")

	// The input binds ctrl+k, so the root's ctrl+k chords can't start, and it
	// types space, so "space f" can't either.
	p.Click("input")
	assert.NotContains(t, keys(), "ctrl+k w")
	assert.NotContains(t, keys(), "space f")
	assert.Contains(t, keys(), "ctrl+g x")
	assert.Contains(t, keys(), "ctrl+s")
}

func TestChord_KeybindBar(t *testing.T) {
	p, _, _ := newChordPilot(t)
	p.AssertSnapshot("idle", "Chords under the ctrl+k group label collapse into 'ctrl+k Window…'; 'g g' and 'g e' show in full; plain g is left out because the chords starting with g win; d and D keep their case")
	p.Press("ctrl+k")
	p.AssertSnapshot("pending", "During ctrl+k the bar shows 'ctrl+k …', then the group's chords by their last key, then 'escape Cancel'")
	p.Press("escape")
	p.Press("g")
	p.AssertSnapshot("pending-g", "During g the bar shows 'g …', 'g Top' and 'e End'")
}

// emacsChordApp shows chordApp's bar with Emacs key names.
type emacsChordApp struct{ chordApp }

func (a *emacsChordApp) Build(ctx BuildContext) Widget {
	return Column{Children: []Widget{
		chordPanel{app: &a.chordApp},
		KeybindBar{FormatKey: FormatKeyEmacs},
	}}
}

func TestChord_KeybindBarFormatsEachKey(t *testing.T) {
	log := &[]string{}
	p := NewPilot(t, &emacsChordApp{chordApp{log: log}}, 70, 3)
	p.AssertSnapshot("idle", "Emacs names: 'g g Top', 'C-k Window…', 'SPC f Find'")
	p.Press("ctrl+k")
	p.AssertSnapshot("pending", "Emacs names during a chord: 'C-k …', 'C-s Save all', 'C-c Comment', 'ESC Cancel'")
}
