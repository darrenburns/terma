package terma

import (
	"os"
	"os/exec"
)

// appSuspender hands the terminal back to the shell, runs fn, then takes the
// terminal back and redraws. It is set while an app runs.
var appSuspender func(fn func() error) error

func currentSuspender() func(fn func() error) error {
	appRuntimeMu.RLock()
	defer appRuntimeMu.RUnlock()
	return appSuspender
}

func setSuspender(suspend func(fn func() error) error) {
	appRuntimeMu.Lock()
	appSuspender = suspend
	appRuntimeMu.Unlock()
}

// RunExternal runs an interactive program, such as $EDITOR or $PAGER, in the
// terminal. It suspends the UI (leaving the alternate screen and handing the
// terminal's input back), runs cmd until it exits, then restores the UI and
// redraws it. It returns cmd's error.
//
// A nil Stdin, Stdout or Stderr on cmd is connected to the terminal.
//
// RunExternal blocks the event loop, so call it from a key, click or other
// event handler, or from a Dispatch callback, not from a background goroutine.
// With no app running it just runs cmd.
func RunExternal(cmd *exec.Cmd) error {
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	suspend := currentSuspender()
	if suspend == nil {
		return cmd.Run()
	}
	return suspend(cmd.Run)
}
