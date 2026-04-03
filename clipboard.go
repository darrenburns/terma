package terma

import (
	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
)

var clipboardWriteString = func(s string) (int, error) {
	return uv.DefaultTerminal().WriteString(s)
}

var clipboardFlush = func() error {
	return uv.DefaultTerminal().Flush()
}

// CopyToClipboard copies text to the system clipboard using OSC52.
func CopyToClipboard(text string) error {
	if _, err := clipboardWriteString(ansi.SetSystemClipboard(text)); err != nil {
		return err
	}
	return clipboardFlush()
}
