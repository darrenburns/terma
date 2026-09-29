package terma

import (
	"sync"

	"github.com/charmbracelet/x/ansi"
)

// ClipboardSelection identifies which terminal clipboard selection to target.
//
// Common values are [SystemClipboard] and [PrimaryClipboard].
type ClipboardSelection = byte

const (
	// SystemClipboard targets the terminal's system clipboard.
	SystemClipboard ClipboardSelection = ansi.SystemClipboard
	// PrimaryClipboard targets the terminal's primary selection.
	PrimaryClipboard ClipboardSelection = ansi.PrimaryClipboard
)

// SetClipboard asks the terminal to set the selected clipboard to content,
// using OSC 52. The request is written between frames; it is safe to call
// from any goroutine. Terminals that don't support OSC 52 ignore it.
func SetClipboard(selection ClipboardSelection, content string) {
	WriteTerminal(ansi.SetClipboard(selection, content))
}

// clipboardReads holds ReadClipboard callbacks waiting for the terminal's
// reply, oldest first. Terminals answer OSC 52 queries in order.
var clipboardReads struct {
	mu      sync.Mutex
	pending []func(string)
}

// ReadClipboard asks the terminal for the contents of the selected clipboard
// and calls fn with them on the event loop when the reply arrives.
//
// Reading the clipboard needs OSC 52 read support, which many terminals leave
// off or ask the user to allow, so fn may never be called. Don't block on it.
func ReadClipboard(selection ClipboardSelection, fn func(content string)) {
	if fn == nil || currentAppContext() == nil {
		return
	}
	clipboardReads.mu.Lock()
	clipboardReads.pending = append(clipboardReads.pending, fn)
	clipboardReads.mu.Unlock()
	WriteTerminal(ansi.RequestClipboard(selection))
}

// deliverClipboard passes a clipboard reply to the oldest waiting reader.
// It reports whether a reader was waiting.
func deliverClipboard(content string) bool {
	clipboardReads.mu.Lock()
	if len(clipboardReads.pending) == 0 {
		clipboardReads.mu.Unlock()
		return false
	}
	fn := clipboardReads.pending[0]
	clipboardReads.pending = clipboardReads.pending[1:]
	clipboardReads.mu.Unlock()
	fn(content)
	return true
}

// resetClipboardReads drops readers left waiting by a previous app run.
func resetClipboardReads() {
	clipboardReads.mu.Lock()
	clipboardReads.pending = nil
	clipboardReads.mu.Unlock()
}
