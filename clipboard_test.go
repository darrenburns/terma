package terma

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestCopyToClipboard_WritesOsc52SequenceAndFlushes(t *testing.T) {
	origWrite := clipboardWriteString
	origFlush := clipboardFlush
	t.Cleanup(func() {
		clipboardWriteString = origWrite
		clipboardFlush = origFlush
	})

	var written string
	flushCalls := 0
	clipboardWriteString = func(s string) (int, error) {
		written = s
		return len(s), nil
	}
	clipboardFlush = func() error {
		flushCalls++
		return nil
	}

	err := CopyToClipboard("hello")
	require.NoError(t, err)
	require.Equal(t, ansi.SetSystemClipboard("hello"), written)
	require.Equal(t, 1, flushCalls)
}
