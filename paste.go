package terma

import "strings"

// PasteHandler is implemented by widgets that accept pasted text.
//
// Terma enables bracketed paste, so a paste arrives as one piece of text
// rather than as a burst of key presses (where each newline would be an
// Enter). It goes to the focused widget if that is a PasteHandler, and
// otherwise bubbles up through its ancestors that are. Return true if the
// paste was handled. A paste nothing handles is dropped.
type PasteHandler interface {
	HandlePaste(text string) bool
}

// dispatchPaste routes pasted text to the focused widget, then its ancestors,
// then the root widget.
func dispatchPaste(focusManager *FocusManager, root Widget, text string) bool {
	text = normalizePastedText(text)
	if text == "" {
		return false
	}
	if focusManager.HandlePaste(text) {
		return true
	}
	if handler, ok := root.(PasteHandler); ok {
		return handler.HandlePaste(text)
	}
	return false
}

// normalizePastedText converts the line endings in pasted text to "\n".
// Terminals send a pasted newline as "\r" (or "\r\n" from some sources).
func normalizePastedText(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

// HandlePaste passes pasted text to the focused widget, then bubbles it up
// through the focused widget's ancestors. It reports whether any handled it.
func (fm *FocusManager) HandlePaste(text string) bool {
	entry := fm.focusedEntry()
	if entry == nil {
		return false
	}
	if handler, ok := entry.Focusable.(PasteHandler); ok && handler.HandlePaste(text) {
		return true
	}
	for i := len(entry.Ancestors) - 1; i >= 0; i-- {
		if handler, ok := entry.Ancestors[i].(PasteHandler); ok && handler.HandlePaste(text) {
			return true
		}
	}
	return false
}
