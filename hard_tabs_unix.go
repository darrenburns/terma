//go:build darwin || freebsd || linux

package terma

import (
	"os"

	"golang.org/x/sys/unix"
)

// disableHardTabs stops the terminal renderer moving the cursor with
// horizontal tabs, and returns a function that undoes the change.
//
// Ultraviolet uses tabs for cursor movement when the tty doesn't expand them,
// deciding once as the terminal starts. tmux treats a tab over blank cells as a
// single tab cell, and writing into part of that cell later resets the rest to
// the default colours, leaving unpainted gaps in the UI. Marking the tty as
// expanding tabs (TAB3) while the terminal starts turns the optimization off.
// Raw mode disables output processing, so the flag has no other effect, and
// the caller restores the original flags as the terminal starts and exits.
func disableHardTabs(f *os.File) (restore func()) {
	noop := func() {}
	if f == nil {
		return noop
	}
	fd := int(f.Fd())
	termios, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
	if err != nil || termios.Oflag&unix.TABDLY == unix.TAB3 {
		return noop
	}
	original := termios.Oflag & unix.TABDLY
	termios.Oflag = termios.Oflag&^unix.TABDLY | unix.TAB3
	if err := unix.IoctlSetTermios(fd, ioctlSetTermios, termios); err != nil {
		return noop
	}
	return func() {
		current, err := unix.IoctlGetTermios(fd, ioctlGetTermios)
		if err != nil {
			return
		}
		current.Oflag = current.Oflag&^unix.TABDLY | original
		_ = unix.IoctlSetTermios(fd, ioctlSetTermios, current)
	}
}
