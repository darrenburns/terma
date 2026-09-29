package terma

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenURL_PlatformCommands(t *testing.T) {
	// Shell metacharacters and URL delimiters must remain in a single argument.
	const rawURL = "https://example.com/a%20b?q=one&other=$(whoami);value=%22quoted%22#fragment"
	for _, tc := range []struct {
		goos string
		args []string
	}{
		{"darwin", []string{"open", rawURL}},
		{"linux", []string{"xdg-open", rawURL}},
		{"windows", []string{"rundll32.exe", "url.dll,FileProtocolHandler", rawURL}},
	} {
		t.Run(tc.goos, func(t *testing.T) {
			called := false
			err := openURL(rawURL, tc.goos, func(cmd *exec.Cmd) error {
				called = true
				assert.Equal(t, tc.args, cmd.Args)
				assert.Nil(t, cmd.Stdin)
				assert.Nil(t, cmd.Stdout)
				assert.Nil(t, cmd.Stderr)
				return nil
			})
			require.NoError(t, err)
			assert.True(t, called)
		})
	}
}

func TestOpenURL_Validation(t *testing.T) {
	for _, rawURL := range []string{
		"", "example.com", "/relative", "//example.com", "--help",
		"file:///tmp/file", "javascript:alert(1)", "mailto:user@example.com",
		"https:", "https:///path", "https://:443/path", "https:example.com",
		"https://example.com/%zz", "https://example.com:bad/path", "https://[::1",
		" https://example.com", "https://example.com/a b", "https://example.com/\n",
		"https://example.com/\x00", "https://example.com/\u00a0",
	} {
		t.Run(rawURL, func(t *testing.T) {
			err := openURL(rawURL, "linux", func(*exec.Cmd) error {
				t.Fatal("invalid URL must not invoke the launcher")
				return nil
			})
			require.Error(t, err)
		})
	}
	for _, rawURL := range []string{
		"http://localhost:8080", "https://example.com", "HTTPS://example.com/path",
		"https://[::1]:443/path", "https://example.com/a%20b?q=two%20words#section",
	} {
		t.Run(rawURL, func(t *testing.T) {
			cmd, err := openURLCommand(rawURL, "linux")
			require.NoError(t, err)
			assert.Equal(t, rawURL, cmd.Args[1])
		})
	}
}

func TestOpenURL_UnsupportedPlatform(t *testing.T) {
	err := openURL("https://example.com", "plan9", func(*exec.Cmd) error {
		t.Fatal("unsupported platform must not invoke the launcher")
		return nil
	})
	require.ErrorContains(t, err, "unsupported platform")
}

func TestOpenURL_LauncherErrors(t *testing.T) {
	for _, launcherErr := range []error{
		exec.ErrNotFound,
		&exec.ExitError{},
		errors.New("launcher failed"),
	} {
		err := openURL("https://example.com", "linux", func(*exec.Cmd) error {
			return launcherErr
		})
		require.ErrorIs(t, err, launcherErr)
		assert.ErrorContains(t, err, "open URL")
	}
}
