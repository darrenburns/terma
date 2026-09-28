//go:build !(darwin || freebsd || linux)

package terma

import "os"

// disableHardTabs is a no-op where the tab expansion flag can't be set.
func disableHardTabs(*os.File) (restore func()) {
	return func() {}
}
