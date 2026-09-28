//go:build linux

package terma

import "golang.org/x/sys/unix"

const (
	ioctlGetTermios = unix.TCGETS
	ioctlSetTermios = unix.TCSETS
)
