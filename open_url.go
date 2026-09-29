package terma

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"unicode"
)

// OpenURL asks the operating system to open an absolute HTTP or HTTPS URL in
// the user's default browser. It supports macOS, Linux (with xdg-open), and
// Windows. The URL is passed directly to the launcher without using a shell.
//
// OpenURL waits for the launcher to exit and returns validation, unsupported
// platform, or launcher errors. Success means the launcher accepted the request;
// it does not mean the browser loaded the page. It does not suspend the UI or
// attach the launcher to the terminal. Use a background task to avoid blocking
// the UI while the launcher runs.
func OpenURL(rawURL string) error {
	return openURL(rawURL, runtime.GOOS, (*exec.Cmd).Run)
}

func openURL(rawURL, goos string, run func(*exec.Cmd) error) error {
	cmd, err := openURLCommand(rawURL, goos)
	if err != nil {
		return err
	}
	if err := run(cmd); err != nil {
		return fmt.Errorf("open URL: %w", err)
	}
	return nil
}

func openURLCommand(rawURL, goos string) (*exec.Cmd, error) {
	if strings.IndexFunc(rawURL, func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsControl(r)
	}) >= 0 {
		return nil, fmt.Errorf("open URL: URL must not contain unescaped whitespace or control characters")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Hostname() == "" || parsed.Opaque != "" ||
		(!strings.EqualFold(parsed.Scheme, "http") && !strings.EqualFold(parsed.Scheme, "https")) {
		return nil, fmt.Errorf("open URL: expected an absolute HTTP or HTTPS URL with a host")
	}

	switch goos {
	case "darwin":
		return exec.Command("open", rawURL), nil
	case "linux":
		return exec.Command("xdg-open", rawURL), nil
	case "windows":
		return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", rawURL), nil
	default:
		return nil, fmt.Errorf("open URL: unsupported platform %q", goos)
	}
}
