package terma

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	isolateClipboard()
	code := m.Run()
	SnapshotTestMain("testdata/snapshot_gallery.html")
	os.Exit(code)
}
