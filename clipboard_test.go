package terma

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

func TestClipboardSelectionConstants(t *testing.T) {
	assert.EqualValues(t, ansi.SystemClipboard, SystemClipboard)
	assert.EqualValues(t, ansi.PrimaryClipboard, PrimaryClipboard)
}

func TestSetClipboard(t *testing.T) {
	assert.Equal(t, ansi.SetClipboard(SystemClipboard, "hello"), SetClipboard(SystemClipboard, "hello"))
	assert.Equal(t, ansi.SetClipboard(PrimaryClipboard, "world"), SetClipboard(PrimaryClipboard, "world"))
}

func TestRequestClipboard(t *testing.T) {
	assert.Equal(t, ansi.RequestClipboard(SystemClipboard), RequestClipboard(SystemClipboard))
	assert.Equal(t, ansi.RequestClipboard(PrimaryClipboard), RequestClipboard(PrimaryClipboard))
}
