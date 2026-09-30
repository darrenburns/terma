package filepickerdemo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
)

// replaceProbeTarget simulates a competing writer only inside this demo's
// disposable fixture directory. FilePicker itself never writes selected files.
func (a *app) replaceProbeTarget() {
	path := filepath.Join(a.root, "alpha.go")
	info, err := os.Stat(path)
	if err != nil {
		a.message.Set(err.Error())
		return
	}
	replacement := filepath.Join(a.root, ".probe-replacement")
	body := bytes.Repeat([]byte{'X'}, int(info.Size()))
	if err = os.WriteFile(replacement, body, info.Mode().Perm()); err == nil {
		err = os.Chtimes(replacement, info.ModTime(), info.ModTime())
	}
	if err == nil {
		err = os.Rename(replacement, path)
	}
	if err != nil {
		a.message.Set(err.Error())
		return
	}
	a.message.Set(fmt.Sprintf("F7 replaced alpha.go identity with same metadata · selections %d · cancels %d", a.selections, a.cancels))
}
