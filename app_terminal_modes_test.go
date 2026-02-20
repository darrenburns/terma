package terma

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func captureTerminalSequences(t *testing.T, fn func(writeString func(string) (int, error))) string {
	t.Helper()

	var out strings.Builder
	fn(func(s string) (int, error) {
		n, err := out.WriteString(s)
		return n, err
	})

	return out.String()
}

func TestEnableTerminalInputModes(t *testing.T) {
	t.Run("enables Kitty keyboard when requested", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			enableTerminalInputModes(writeString, true, false)
		})

		expected := strings.Join(terminalEnableSequences, "") + ansi.PushKittyKeyboard(ansi.KittyAllFlags)
		require.Equal(t, expected, got)
	})

	t.Run("forces Kitty keyboard off when explicitly disabled", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			enableTerminalInputModes(writeString, false, true)
		})

		expected := strings.Join(terminalEnableSequences, "") + ansi.PushKittyKeyboard(0)
		require.Equal(t, expected, got)
	})

	t.Run("leaves Kitty keyboard unchanged when neither enabled nor forced off", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			enableTerminalInputModes(writeString, false, false)
		})

		expected := strings.Join(terminalEnableSequences, "")
		require.Equal(t, expected, got)
	})
}

func TestDisableTerminalInputModes(t *testing.T) {
	t.Run("pops Kitty keyboard stack when enabled", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			disableTerminalInputModes(writeString, true, false, false)
		})

		expected := strings.Join(terminalDisableSequences, "") + ansi.PopKittyKeyboard(1)
		require.Equal(t, expected, got)
	})

	t.Run("pops Kitty keyboard stack when force-disabled", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			disableTerminalInputModes(writeString, false, true, false)
		})

		expected := strings.Join(terminalDisableSequences, "") + ansi.PopKittyKeyboard(1)
		require.Equal(t, expected, got)
	})

	t.Run("does not touch Kitty keyboard when unchanged", func(t *testing.T) {
		got := captureTerminalSequences(t, func(writeString func(string) (int, error)) {
			disableTerminalInputModes(writeString, false, false, false)
		})

		expected := strings.Join(terminalDisableSequences, "")
		require.Equal(t, expected, got)
	})
}

func TestResolveKittyKeyboardMode(t *testing.T) {
	t.Run("defaults to force-disabled", func(t *testing.T) {
		t.Setenv("TERMA_ENABLE_KITTY_KEYBOARD", "")
		t.Setenv("TERMA_DISABLE_KITTY_KEYBOARD", "")

		enable, forceDisable := resolveKittyKeyboardMode()
		require.False(t, enable)
		require.True(t, forceDisable)
	})

	t.Run("explicit enable opts in", func(t *testing.T) {
		t.Setenv("TERMA_ENABLE_KITTY_KEYBOARD", "1")
		t.Setenv("TERMA_DISABLE_KITTY_KEYBOARD", "")

		enable, forceDisable := resolveKittyKeyboardMode()
		require.True(t, enable)
		require.False(t, forceDisable)
	})

	t.Run("explicit disable keeps force-disabled", func(t *testing.T) {
		t.Setenv("TERMA_ENABLE_KITTY_KEYBOARD", "")
		t.Setenv("TERMA_DISABLE_KITTY_KEYBOARD", "1")

		enable, forceDisable := resolveKittyKeyboardMode()
		require.False(t, enable)
		require.True(t, forceDisable)
	})

	t.Run("enable takes precedence when both are set", func(t *testing.T) {
		t.Setenv("TERMA_ENABLE_KITTY_KEYBOARD", "1")
		t.Setenv("TERMA_DISABLE_KITTY_KEYBOARD", "1")

		enable, forceDisable := resolveKittyKeyboardMode()
		require.True(t, enable)
		require.False(t, forceDisable)
	})
}
