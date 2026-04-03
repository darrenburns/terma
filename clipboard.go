package terma

import "github.com/charmbracelet/x/ansi"

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

// SetClipboard returns an OSC 52 sequence that asks the terminal to set the
// selected clipboard contents.
func SetClipboard(selection ClipboardSelection, content string) string {
	return ansi.SetClipboard(selection, content)
}

// RequestClipboard returns an OSC 52 sequence that asks the terminal to report
// the current contents of the selected clipboard.
func RequestClipboard(selection ClipboardSelection) string {
	return ansi.RequestClipboard(selection)
}
