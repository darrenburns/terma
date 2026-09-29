package terma

import (
	"context"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// withRunningApp makes the package behave as if an app were running, so
// terminal writes are queued, and clears the queue afterwards.
func withRunningApp(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	setAppRuntimeState(ctx, newDispatchQueue())
	takeTerminalWrites()
	resetClipboardReads()
	t.Cleanup(func() {
		cancel()
		clearAppRuntimeState()
		takeTerminalWrites()
		resetClipboardReads()
	})
}

func TestClipboardSelectionConstants(t *testing.T) {
	assert.EqualValues(t, ansi.SystemClipboard, SystemClipboard)
	assert.EqualValues(t, ansi.PrimaryClipboard, PrimaryClipboard)
}

func TestSetClipboard_QueuesOSC52WriteForNextFrame(t *testing.T) {
	withRunningApp(t)

	SetClipboard(SystemClipboard, "hello")
	SetClipboard(PrimaryClipboard, "world")

	assert.Equal(t, []string{
		ansi.SetClipboard(SystemClipboard, "hello"),
		ansi.SetClipboard(PrimaryClipboard, "world"),
	}, takeTerminalWrites())
	assert.Empty(t, takeTerminalWrites(), "writes are taken once")
}

func TestSetClipboard_WithoutRunningAppWritesNothing(t *testing.T) {
	takeTerminalWrites()
	SetClipboard(SystemClipboard, "hello")
	assert.Empty(t, takeTerminalWrites())
}

func TestWriteTerminal_IgnoresEmptySequence(t *testing.T) {
	withRunningApp(t)
	WriteTerminal("")
	assert.Empty(t, takeTerminalWrites())
}

func TestReadClipboard_RequestsAndDeliversRepliesInOrder(t *testing.T) {
	withRunningApp(t)

	var got []string
	ReadClipboard(SystemClipboard, func(s string) { got = append(got, "first:"+s) })
	ReadClipboard(PrimaryClipboard, func(s string) { got = append(got, "second:"+s) })

	assert.Equal(t, []string{
		ansi.RequestClipboard(SystemClipboard),
		ansi.RequestClipboard(PrimaryClipboard),
	}, takeTerminalWrites())

	assert.True(t, deliverClipboard("a"))
	assert.True(t, deliverClipboard("b"))
	assert.False(t, deliverClipboard("unrequested"), "a reply nobody asked for is dropped")
	assert.Equal(t, []string{"first:a", "second:b"}, got)
}

func TestReadClipboard_WithoutRunningAppDoesNothing(t *testing.T) {
	resetClipboardReads()
	called := false
	ReadClipboard(SystemClipboard, func(string) { called = true })
	assert.False(t, deliverClipboard("x"))
	assert.False(t, called)
}
