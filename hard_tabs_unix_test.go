//go:build darwin || freebsd || linux

package terma

import (
	"bytes"
	"context"
	"image/color"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/creack/pty"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func openTestPTY(t *testing.T) (ptmx, tty *os.File) {
	t.Helper()
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Skipf("pty unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = tty.Close()
		_ = ptmx.Close()
	})
	require.NoError(t, pty.Setsize(tty, &pty.Winsize{Rows: 3, Cols: 40}))
	return ptmx, tty
}

func ttyTabDelay(t *testing.T, f *os.File) uint64 {
	t.Helper()
	termios, err := unix.IoctlGetTermios(int(f.Fd()), ioctlGetTermios)
	require.NoError(t, err)
	return uint64(termios.Oflag & unix.TABDLY)
}

func TestDisableHardTabsRestoresTabDelay(t *testing.T) {
	_, tty := openTestPTY(t)
	require.Equal(t, uint64(unix.TAB0), ttyTabDelay(t, tty))

	restore := disableHardTabs(tty)
	require.Equal(t, uint64(unix.TAB3), ttyTabDelay(t, tty))

	restore()
	require.Equal(t, uint64(unix.TAB0), ttyTabDelay(t, tty))
}

// tmux turns a tab over blank cells into one tab cell, and writing into part
// of it later resets the rest to the default colours. The renderer must move
// the cursor without tabs so styled blank space survives partial repaints.
func TestStartTerminalDoesNotMoveCursorWithTabs(t *testing.T) {
	ptmx, tty := openTestPTY(t)

	var mu sync.Mutex
	var out bytes.Buffer
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			mu.Lock()
			out.Write(buf[:n])
			mu.Unlock()
			if err != nil {
				return
			}
		}
	}()
	output := func() string {
		mu.Lock()
		defer mu.Unlock()
		return out.String()
	}

	terminal := uv.NewTerminal(tty, tty, []string{"TERM=xterm-256color", "COLORTERM=truecolor"})
	require.NoError(t, startTerminal(terminal, tty))
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = terminal.Shutdown(ctx)
	})
	require.Equal(t, uint64(unix.TAB0), ttyTabDelay(t, tty), "the tty's own tab setting is restored after start")

	// A label and a value separated by a wide run of background-coloured blanks,
	// like a stat row: the layout that made the renderer reach for tabs.
	style := uv.Style{Bg: color.RGBA{R: 25, G: 23, B: 36, A: 255}}
	for x := 0; x < 40; x++ {
		terminal.SetCell(x, 0, &uv.Cell{Content: " ", Width: 1, Style: style})
	}
	for i, r := range "Items" {
		terminal.SetCell(i, 0, &uv.Cell{Content: string(r), Width: 1, Style: style})
	}
	terminal.SetCell(32, 0, &uv.Cell{Content: "3", Width: 1, Style: style})
	require.NoError(t, terminal.Display())

	require.Eventually(t, func() bool { return strings.Contains(output(), "3") }, time.Second, 10*time.Millisecond)
	require.NotContains(t, output(), "\t", "cursor movement must not use horizontal tabs")
}
