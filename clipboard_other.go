//go:build !unix

package terma

import "os/exec"

// killProcessGroupOnCancel leaves exec.CommandContext's default, which kills
// only the tool itself on timeout.
func killProcessGroupOnCancel(cmd *exec.Cmd) {}
