package terma

import (
	"slices"
	"time"
	"unicode"
)

// DefaultUndoKeys undo the last edit in a TextInput or TextArea whose UndoKeys
// is nil.
var DefaultUndoKeys = []string{"ctrl+z"}

// DefaultRedoKeys redo the last undone edit in a TextInput or TextArea whose
// RedoKeys is nil. ctrl+shift+z is left out: terminals without the kitty
// keyboard protocol send it as ctrl+z, so it would undo instead.
var DefaultRedoKeys = []string{"ctrl+y"}

const (
	// undoGroupPause ends a run of typing: a keystroke after a longer pause
	// starts a new undo step.
	undoGroupPause = time.Second
	// maxUndoSteps and maxUndoGraphemes cap the history. The oldest steps are
	// dropped first; the newest step is always kept.
	maxUndoSteps     = 500
	maxUndoGraphemes = 1 << 18
)

// editKind says how an edit may merge with the step before it.
type editKind uint8

const (
	editSingle         editKind = iota // Its own undo step: paste, delete selection, programmatic edits
	editTyping                         // Typed text, merged with adjacent typing
	editDeleteBackward                 // Backspace, merged with adjacent backspaces
	editDeleteForward                  // Delete, merged with adjacent deletes
)

// textCaret is the cursor and selection anchor (-1 for none).
type textCaret struct {
	cursor int
	anchor int
}

// textEdit replaces removed, starting at grapheme start, with inserted.
type textEdit struct {
	start    int
	removed  []string
	inserted []string
	before   textCaret
	after    textCaret
	kind     editKind
	at       time.Time
}

func (e *textEdit) size() int {
	return len(e.removed) + len(e.inserted)
}

// textFields are the signals of a text state that edits read and write.
type textFields struct {
	content AnySignal[[]string]
	cursor  Signal[int]
	anchor  Signal[int]
}

func (f textFields) caret() textCaret {
	return textCaret{cursor: f.cursor.Peek(), anchor: f.anchor.Peek()}
}

func (f textFields) setCaret(c textCaret) {
	f.cursor.Set(c.cursor)
	f.anchor.Set(c.anchor)
}

// splice writes a new content slice, so a slice read from Content is never
// changed under its reader.
func (f textFields) splice(start, end int, inserted []string) {
	graphemes := f.content.Peek()
	result := make([]string, 0, len(graphemes)-(end-start)+len(inserted))
	result = append(result, graphemes[:start]...)
	result = append(result, inserted...)
	result = append(result, graphemes[end:]...)
	f.content.Set(result)
}

// textHistory is the undo and redo stacks of a text state.
type textHistory struct {
	undo []textEdit
	redo []textEdit
	// merge is false once something other than an edit (undo, redo, reset)
	// has happened since the newest undo step, so the next edit starts a new
	// step.
	merge bool
	// revision is the content revision the history last wrote. A different
	// revision means Content was set directly, and the steps no longer apply.
	revision uint64
	now      func() time.Time
}

func (h *textHistory) reset(f textFields) {
	h.undo = nil
	h.redo = nil
	h.merge = false
	_, h.revision = f.content.peekWithRevision()
}

func (h *textHistory) current(f textFields) bool {
	_, revision := f.content.peekWithRevision()
	if revision != h.revision {
		h.reset(f)
		return false
	}
	return true
}

func (h *textHistory) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

// edit replaces graphemes [start, end) with inserted, leaves the cursor at
// cursorAfter with no selection, and records the change as an undo step.
func (h *textHistory) edit(f textFields, start, end int, inserted []string, cursorAfter int, kind editKind) {
	h.current(f)
	before := f.caret()
	after := textCaret{cursor: cursorAfter, anchor: -1}
	if start == end && len(inserted) == 0 {
		f.setCaret(after)
		return
	}
	e := textEdit{
		start:    start,
		removed:  slices.Clone(f.content.Peek()[start:end]),
		inserted: inserted,
		before:   before,
		after:    after,
		kind:     kind,
		at:       h.clock(),
	}
	f.splice(start, end, inserted)
	f.setCaret(after)
	_, h.revision = f.content.peekWithRevision()
	h.record(e)
}

func (h *textHistory) record(e textEdit) {
	h.redo = nil
	if h.merge && len(h.undo) > 0 && h.mergeInto(&h.undo[len(h.undo)-1], e) {
		return
	}
	h.undo = append(h.undo, e)
	h.merge = true
	h.trim()
}

// mergeInto folds e into the step top when e continues it: the same kind of
// keystroke, at the caret top left, soon after, and not starting a new word.
func (h *textHistory) mergeInto(top *textEdit, e textEdit) bool {
	if e.kind == editSingle || e.kind != top.kind || e.before != top.after || e.at.Sub(top.at) > undoGroupPause {
		return false
	}
	switch e.kind {
	case editTyping:
		if len(e.removed) > 0 || e.start != top.start+len(top.inserted) || startsWord(top.inserted[len(top.inserted)-1], e.inserted[0]) {
			return false
		}
		top.inserted = append(top.inserted, e.inserted...)
	case editDeleteBackward:
		if len(e.inserted) > 0 || len(top.inserted) > 0 || e.start+len(e.removed) != top.start || startsWord(top.removed[0], e.removed[len(e.removed)-1]) {
			return false
		}
		top.removed = append(slices.Clone(e.removed), top.removed...)
		top.start = e.start
	case editDeleteForward:
		if len(e.inserted) > 0 || len(top.inserted) > 0 || e.start != top.start || startsWord(top.removed[len(top.removed)-1], e.removed[0]) {
			return false
		}
		top.removed = append(top.removed, e.removed...)
	}
	top.after = e.after
	top.at = e.at
	return true
}

// startsWord reports whether next, typed or deleted after prev, begins a new
// word, so "one two" undoes as "two" then "one ".
func startsWord(prev, next string) bool {
	return isSpaceGrapheme(prev) && !isSpaceGrapheme(next)
}

func isSpaceGrapheme(g string) bool {
	for _, r := range g {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return g != ""
}

func (h *textHistory) trim() {
	total := 0
	for _, e := range h.undo {
		total += e.size()
	}
	drop := 0
	for len(h.undo)-drop > 1 && (len(h.undo)-drop > maxUndoSteps || total > maxUndoGraphemes) {
		total -= h.undo[drop].size()
		drop++
	}
	if drop > 0 {
		h.undo = slices.Delete(h.undo, 0, drop)
	}
}

func (h *textHistory) canUndo(f textFields) bool {
	return h.current(f) && len(h.undo) > 0
}

func (h *textHistory) canRedo(f textFields) bool {
	return h.current(f) && len(h.redo) > 0
}

// stepBack reverts the newest undo step and restores the caret it started from.
func (h *textHistory) stepBack(f textFields) bool {
	if !h.canUndo(f) {
		return false
	}
	e := h.undo[len(h.undo)-1]
	h.undo = h.undo[:len(h.undo)-1]
	f.splice(e.start, e.start+len(e.inserted), slices.Clone(e.removed))
	f.setCaret(e.before)
	h.redo = append(h.redo, e)
	h.merge = false
	_, h.revision = f.content.peekWithRevision()
	return true
}

// stepForward reapplies the newest undone step and restores the caret it left.
func (h *textHistory) stepForward(f textFields) bool {
	if !h.canRedo(f) {
		return false
	}
	e := h.redo[len(h.redo)-1]
	h.redo = h.redo[:len(h.redo)-1]
	f.splice(e.start, e.start+len(e.removed), slices.Clone(e.inserted))
	f.setCaret(e.after)
	h.undo = append(h.undo, e)
	h.merge = false
	_, h.revision = f.content.peekWithRevision()
	return true
}

// diffGraphemes trims the common prefix and suffix of old and new, so a
// whole-text replacement is recorded as the span that changed.
func diffGraphemes(old, new []string) (start, oldEnd, newEnd int) {
	for start < len(old) && start < len(new) && old[start] == new[start] {
		start++
	}
	oldEnd, newEnd = len(old), len(new)
	for oldEnd > start && newEnd > start && old[oldEnd-1] == new[newEnd-1] {
		oldEnd--
		newEnd--
	}
	return start, oldEnd, newEnd
}

// historyKeybinds binds the undo and redo keys, defaulting nil key lists.
func historyKeybinds(undoKeys, redoKeys []string, undo, redo func()) []Keybind {
	var keybinds []Keybind
	for _, key := range undoKeysOr(undoKeys) {
		keybinds = append(keybinds, Keybind{Key: key, Action: undo, Hidden: true})
	}
	for _, key := range redoKeysOr(redoKeys) {
		keybinds = append(keybinds, Keybind{Key: key, Action: redo, Hidden: true})
	}
	return keybinds
}

func undoKeysOr(keys []string) []string {
	if keys == nil {
		return DefaultUndoKeys
	}
	return keys
}

func redoKeysOr(keys []string) []string {
	if keys == nil {
		return DefaultRedoKeys
	}
	return keys
}
