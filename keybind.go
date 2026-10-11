package terma

import (
	"slices"
	"strings"
)

// Keybind represents a single declarative keybinding.
// It associates a key pattern with a display name and action callback.
type Keybind struct {
	// Key is the key pattern to match, e.g., "ctrl+s", "enter", "shift+tab".
	// A character key matches its case: "D" is shift+d, "d" is d.
	//
	// A chord is several key presses separated by spaces, pressed one after
	// the other: "ctrl+k ctrl+s", "g g", "space f". Inside a chord, write the
	// space key as "space"; on its own, " " works too.
	//
	// A chord wins over a binding to its own start: with "g" and "g g" in the
	// same widget, pressing g waits for the next key and "g" never runs.
	Key string
	// Name is a short display name for the binding, e.g., "Save", "Submit"
	Name string
	// Action is the callback to execute when the keybinding is triggered.
	//
	// A keybind with no Action whose Key starts other chords labels them as
	// a group: {Key: "ctrl+k", Name: "Window"} collapses every "ctrl+k …"
	// chord into one "ctrl+k Window…" hint in KeybindBar. Pressing ctrl+k then
	// shows the group's chords.
	Action func()
	// Hidden prevents this keybind from appearing in KeybindBar.
	// Use for internal bindings that shouldn't be displayed to users.
	// A hidden group label hides its whole group.
	Hidden bool
}

// KeybindProvider is implemented by widgets that declare keybindings.
// Widgets implementing this interface can define their keybindings declaratively,
// allowing the framework to query and display them (e.g., in a footer).
type KeybindProvider interface {
	Keybinds() []Keybind
}

// keySteps splits a Keybind.Key into the key presses of its chord.
func keySteps(key string) []string {
	if key == " " {
		return []string{" "}
	}
	return strings.Fields(key)
}

// startsChord reports whether key is the start of a longer chord among
// keybinds.
func startsChord(key string, keybinds []Keybind) bool {
	steps := keySteps(key)
	for _, kb := range keybinds {
		other := keySteps(kb.Key)
		if len(other) > len(steps) && slices.Equal(other[:len(steps)], steps) {
			return true
		}
	}
	return false
}

// keyStepMatches reports whether event is the key press step names.
func keyStepMatches(event KeyEvent, step string) bool {
	// WORKAROUND: ultraviolet's MatchString has a bug with "+" character.
	// It uses "+" as a delimiter for modifiers (e.g., "ctrl+a"), so a literal
	// "+" gets split into empty strings and fails to match.
	// Fall back to direct string comparison when MatchString fails.
	return event.MatchString(step) || event.Key() == step
}

// typedKeybinds looks up the key presses typed (a chord's earlier presses
// followed by the latest) in keybinds. complete is the first binding they
// complete; started is the first chord they begin but don't finish.
func typedKeybinds(typed []KeyEvent, keybinds []Keybind) (complete, started *Keybind) {
	for i := range keybinds {
		steps := keySteps(keybinds[i].Key)
		if len(steps) < len(typed) {
			continue
		}
		matched := true
		for j, event := range typed {
			if !keyStepMatches(event, steps[j]) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		if len(steps) == len(typed) {
			if complete == nil {
				complete = &keybinds[i]
			}
		} else if started == nil {
			started = &keybinds[i]
		}
	}
	return complete, started
}

// matchKeybind runs the single-key binding in keybinds that event matches.
// Returns true if a keybind was matched and executed. It keeps no chord
// state; key dispatch goes through FocusManager.matchKeybinds instead.
func matchKeybind(event KeyEvent, keybinds []Keybind) bool {
	complete, started := typedKeybinds([]KeyEvent{event}, keybinds)
	if complete == nil || started != nil {
		return false
	}
	if complete.Action != nil {
		complete.Action()
	}
	return true
}

// keyChord is a chord in progress: the key presses typed so far, and the
// steps they matched, spelled as the binding that started the chord writes
// them.
type keyChord struct {
	events []KeyEvent
	steps  []string
}

// chordPending reports whether a chord is in progress.
func (fm *FocusManager) chordPending() bool {
	return len(fm.chord.events) > 0
}

// setChord records the chord in progress and notifies widgets showing
// keybinds, which list the chord's continuations while it is pending.
func (fm *FocusManager) setChord(chord keyChord) {
	fm.chord = chord
	fm.keybindsVersion.Update(func(v int) int { return v + 1 })
}

// cancelChord abandons the chord in progress, if any.
func (fm *FocusManager) cancelChord() {
	if fm.chordPending() {
		Log("Chord %q cancelled", strings.Join(fm.chord.steps, " "))
		fm.setChord(keyChord{})
	}
}

// matchKeybinds handles event with keybinds, after any chord in progress:
// it runs the binding the key presses complete, or starts or continues a
// chord that begins with them. A chord wins over a binding to its own start.
// Returns true if keybinds handled the key.
func (fm *FocusManager) matchKeybinds(event KeyEvent, keybinds []Keybind) bool {
	typed := append(slices.Clone(fm.chord.events), event)
	complete, started := typedKeybinds(typed, keybinds)
	switch {
	case started != nil:
		steps := keySteps(started.Key)[:len(typed)]
		Log("Chord %q started, waiting for the next key", strings.Join(steps, " "))
		fm.setChord(keyChord{events: typed, steps: steps})
	case complete != nil:
		if fm.chordPending() {
			fm.setChord(keyChord{})
		}
		if complete.Action != nil {
			complete.Action()
		}
	default:
		return false
	}
	return true
}

// continueChord sends the next key press of a chord in progress to the
// keybinds that could continue it, in dispatch order: the focused widget,
// its ancestors, global keybinds, then the root. Escape, or a key no chord
// continues with, cancels the chord, and the key is dropped.
func (fm *FocusManager) continueChord(event KeyEvent, global []Keybind) {
	if event.MatchString("escape") {
		fm.cancelChord()
		return
	}
	for _, keybinds := range fm.keybindChain(global) {
		if fm.matchKeybinds(event, keybinds) {
			return
		}
	}
	Log("Chord %q has no continuation %q", strings.Join(fm.chord.steps, " "), event.Key())
	fm.cancelChord()
}

// keybindChain returns the keybinds a key press is checked against, in
// order: the focused widget's, its ancestors' (innermost first), global
// keybinds, then the root widget's.
func (fm *FocusManager) keybindChain(global []Keybind) [][]Keybind {
	var chain [][]Keybind
	if entry := fm.focusedEntry(); entry != nil {
		if provider, ok := entry.Focusable.(KeybindProvider); ok {
			chain = append(chain, provider.Keybinds())
		}
		for i := len(entry.Ancestors) - 1; i >= 0; i-- {
			if fm.isRootWidget(entry.Ancestors[i]) {
				continue
			}
			if provider, ok := entry.Ancestors[i].(KeybindProvider); ok {
				chain = append(chain, provider.Keybinds())
			}
		}
	}
	chain = append(chain, global)
	if provider, ok := fm.rootWidget.(KeybindProvider); ok {
		chain = append(chain, provider.Keybinds())
	}
	return chain
}

// chordContinuations returns the keybinds that continue the chord in
// progress, keyed by the presses still to come: during "ctrl+k", the binding
// "ctrl+k ctrl+s" is returned with Key "ctrl+s".
func (fm *FocusManager) chordContinuations(keybinds []Keybind) []Keybind {
	n := len(fm.chord.events)
	var continuations []Keybind
	for _, kb := range keybinds {
		steps := keySteps(kb.Key)
		if len(steps) <= n {
			continue
		}
		matched := true
		for i, event := range fm.chord.events {
			if !keyStepMatches(event, steps[i]) {
				matched = false
				break
			}
		}
		if matched {
			kb.Key = strings.Join(steps[n:], " ")
			continuations = append(continuations, kb)
		}
	}
	return continuations
}
