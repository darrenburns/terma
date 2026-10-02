package terma

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pickerTestFS struct {
	files fstest.MapFS
	calls int
	deny  string
}

func newPickerTestFS() *pickerTestFS {
	return &pickerTestFS{files: fstest.MapFS{
		"docs": {Mode: fs.ModeDir | 0755}, "empty": {Mode: fs.ModeDir | 0755}, "docs/guide.md": {Data: []byte("guide"), Mode: 0644},
		"alpha.go": {Data: []byte("package alpha"), Mode: 0644, ModTime: time.Unix(10, 0)}, "beta.txt": {Data: []byte("beta"), Mode: 0644}, "space name.go": {Data: []byte("space"), Mode: 0644}, "日本語.go": {Data: []byte("unicode"), Mode: 0644}, ".secret.go": {Data: []byte("hidden"), Mode: 0600},
	}}
}
func (f *pickerTestFS) name(p string) (string, error) {
	if p == "/fixture" {
		return ".", nil
	}
	if !strings.HasPrefix(p, "/fixture/") {
		return "", fs.ErrNotExist
	}
	return strings.TrimPrefix(p, "/fixture/"), nil
}
func (f *pickerTestFS) ReadDir(p string) ([]fs.DirEntry, error) {
	f.calls++
	if p == f.deny {
		return nil, fs.ErrPermission
	}
	name, err := f.name(p)
	if err != nil {
		return nil, err
	}
	return fs.ReadDir(f.files, name)
}
func (f *pickerTestFS) Stat(p string) (fs.FileInfo, error) {
	f.calls++
	name, err := f.name(p)
	if err != nil {
		return nil, err
	}
	return fs.Stat(f.files, name)
}
func (f *pickerTestFS) Lstat(p string) (fs.FileInfo, error) { return f.Stat(p) }
func pickerFixture() FilePicker {
	return FilePicker{ID: "picker", State: NewFilePickerStateWithFileSystem("/fixture", newPickerTestFS()), Filters: []FileFilter{{Label: "All"}, {Label: "Go", Patterns: []string{"*.go"}}}, Style: Style{Width: Flex(1), Height: Flex(1)}}
}

func TestFilePickerNavigationAndErrors(t *testing.T) {
	p := pickerFixture()
	s := p.State
	assert.Equal(t, "/fixture", s.Directory.Peek())
	assert.Equal(t, []string{"docs", "empty", ".secret.go", "alpha.go", "beta.txt", "space name.go", "日本語.go"}, func() []string {
		var names []string
		for _, e := range s.Entries() {
			names = append(names, e.Name)
		}
		return names
	}())
	s.SelectPaths("/fixture/alpha.go")
	require.Error(t, s.Navigate("missing"))
	assert.Equal(t, "/fixture", s.Directory.Peek())
	assert.Equal(t, []string{"/fixture/alpha.go"}, s.SelectedPaths())
	assert.Contains(t, s.Error.Peek(), "not exist")
	require.NoError(t, s.Navigate("docs"))
	assert.Empty(t, s.SelectedPaths())
	assert.Equal(t, "/fixture/docs", s.Directory.Peek())
	p.parent()
	assert.Equal(t, "/fixture", s.Directory.Peek())
	require.NoError(t, s.Navigate("space name.go"))
	assert.Equal(t, "space name.go", s.FilenameInput.GetText())
	assert.Equal(t, "space name.go", s.list.GetItems()[s.list.CursorIndex.Peek()].Name)
	require.NoError(t, s.Navigate("日本語.go"))
	assert.Equal(t, "日本語.go", s.FilenameInput.GetText())
	s.filesystem.(*pickerTestFS).deny = "/fixture/empty"
	require.ErrorIs(t, s.Navigate("empty"), fs.ErrPermission)
	assert.Contains(t, s.Error.Peek(), "permission denied")
	missing := NewFilePickerStateWithFileSystem("/missing", newPickerTestFS())
	assert.Empty(t, missing.Entries())
	assert.NotEmpty(t, missing.Error.Peek())
}
func TestFilePickerOpenMultiSelectionAndVisibility(t *testing.T) {
	p := pickerFixture()
	p.MultiSelect = true
	var got [][]string
	p.OnSelect = func(paths []string) { got = append(got, paths) }
	p.State.SelectPaths("/fixture/日本語.go", "/fixture/alpha.go", "/fixture/docs", "/fixture/.secret.go")
	require.True(t, p.Submit())
	assert.Equal(t, [][]string{{"/fixture/alpha.go", "/fixture/日本語.go"}}, got)
	assert.False(t, p.Submit())
	p.activate(p.State.Entries()[3])
	assert.Len(t, got, 1)
	p.State.Reset()
	p.State.SelectPaths("/fixture/beta.txt")
	p.nextFilter()
	assert.Empty(t, p.State.SelectedPaths())
	assert.True(t, p.visible(p.State.Entries()[0]))
	assert.False(t, p.visible(p.State.Entries()[4]))
	p.State.SelectPaths("/fixture/alpha.go")
	p.toggleHidden()
	assert.Empty(t, p.State.SelectedPaths())
	assert.True(t, p.visible(p.State.Entries()[2]))
	p.Filters = []FileFilter{{Patterns: []string{"["}}}
	assert.False(t, p.Submit())
	assert.Contains(t, p.State.Error.Peek(), "invalid file filter")
}
func TestFilePickerOpenRevalidates(t *testing.T) {
	p := pickerFixture()
	p.MultiSelect = true
	calls := 0
	p.OnSelect = func([]string) { calls++ }
	p.State.SelectPaths("/fixture/alpha.go", "/fixture/beta.txt")
	delete(p.State.filesystem.(*pickerTestFS).files, "alpha.go")
	assert.False(t, p.Submit())
	assert.Zero(t, calls)
	assert.NotEmpty(t, p.State.Error.Peek())
}
func TestFilePickerDirectoryMode(t *testing.T) {
	p := pickerFixture()
	p.Mode = FilePickerDirectory
	p.MultiSelect = true
	var result []string
	p.OnSelect = func(paths []string) { result = paths }
	require.NoError(t, p.State.Navigate("empty"))
	assert.Empty(t, p.State.Entries())
	require.True(t, p.Submit())
	assert.Equal(t, []string{"/fixture/empty"}, result)
}
func TestFilePickerSaveValidationAndOverwrite(t *testing.T) {
	p := pickerFixture()
	p.Mode = FilePickerSave
	var got [][]string
	p.OnSelect = func(paths []string) { got = append(got, paths) }
	for _, name := range []string{"", ".", "..", "../escape", "one/two", "one\\two", "nul\x00name"} {
		p.State.FilenameInput.SetText(name)
		assert.False(t, p.Submit(), name)
		assert.Empty(t, got)
	}
	p.State.FilterIndex.Set(1)
	p.State.FilenameInput.SetText("file.txt")
	assert.False(t, p.Submit())
	assert.Contains(t, p.State.Error.Peek(), "filter")
	p.State.FilenameInput.SetText("new name 日本語.go")
	require.True(t, p.Submit())
	assert.Equal(t, []string{"/fixture/new name 日本語.go"}, got[0])
	_, err := p.State.filesystem.Stat(got[0][0])
	require.ErrorIs(t, err, fs.ErrNotExist)
	p.State.Reset()
	p.State.FilenameInput.SetText("alpha.go")
	assert.False(t, p.Submit())
	assert.Equal(t, "/fixture/alpha.go", p.State.pending.Peek())
	assert.Len(t, got, 1)
	cancelled := 0
	p.OnCancel = func() { cancelled++ }
	p.Cancel()
	assert.Empty(t, p.State.pending.Peek())
	assert.Zero(t, cancelled)
	assert.False(t, p.State.completed.Peek())
	p.Submit()
	p.confirmOverwrite()
	p.confirmOverwrite()
	assert.Len(t, got, 2)
	assert.Equal(t, []string{"/fixture/alpha.go"}, got[1])
	p.State.Reset()
	p.State.FilterIndex.Set(0)
	p.State.FilenameInput.SetText("docs")
	assert.False(t, p.Submit())
	assert.Contains(t, p.State.Error.Peek(), "directory")
}
func TestFilePickerOverwriteChanges(t *testing.T) {
	for _, change := range []string{"filename", "type", "content", "filter", "mode", "vanished"} {
		t.Run(change, func(t *testing.T) {
			p := pickerFixture()
			p.Mode = FilePickerSave
			p.State.FilenameInput.SetText("alpha.go")
			calls := 0
			p.OnSelect = func([]string) { calls++ }
			p.Submit()
			require.NotEmpty(t, p.State.pending.Peek())
			f := p.State.filesystem.(*pickerTestFS)
			switch change {
			case "filename":
				p.State.FilenameInput.SetText("beta.txt")
			case "type":
				f.files["alpha.go"] = &fstest.MapFile{Mode: fs.ModeDir | 0755}
			case "content":
				f.files["alpha.go"] = &fstest.MapFile{Data: []byte("changed larger contents"), Mode: 0644}
			case "filter":
				p.Filters = []FileFilter{{Patterns: []string{"*.txt"}}}
			case "mode":
				p.Mode = FilePickerOpen
			case "vanished":
				delete(f.files, "alpha.go")
			}
			p.confirmOverwrite()
			if change == "vanished" {
				assert.Equal(t, 1, calls)
			} else {
				assert.Zero(t, calls)
				assert.NotEmpty(t, p.State.Error.Peek())
			}
			assert.Empty(t, p.State.pending.Peek())
		})
	}
}
func TestFilePickerCancelCardinality(t *testing.T) {
	p := pickerFixture()
	cancelled, selected := 0, 0
	p.OnCancel = func() { cancelled++ }
	p.OnSelect = func([]string) { selected++ }
	p.Cancel()
	p.Cancel()
	assert.False(t, p.Submit())
	assert.Equal(t, 1, cancelled)
	assert.Zero(t, selected)
	p.State.Reset()
	p.Cancel()
	assert.Equal(t, 2, cancelled)
}
func TestFilePickerRealPathsAndSymlinks(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "child"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "space 日本語.go"), []byte("content"), 0644))
	require.NoError(t, os.Symlink("child", filepath.Join(dir, "directory-link")))
	require.NoError(t, os.Symlink("space 日本語.go", filepath.Join(dir, "file-link")))
	require.NoError(t, os.Symlink("missing", filepath.Join(dir, "broken-link")))
	s := NewFilePickerState(dir)
	p := FilePicker{State: s, Mode: FilePickerOpen}
	var result []string
	p.OnSelect = func(paths []string) { result = paths }
	s.SelectPaths(filepath.Join(dir, "file-link"))
	p.MultiSelect = true
	assert.True(t, p.Submit())
	assert.Equal(t, []string{filepath.Join(dir, "file-link")}, result)
	s.Reset()
	s.SelectPaths(filepath.Join(dir, "broken-link"))
	for _, e := range s.Entries() {
		if e.Name == "broken-link" {
			require.Error(t, e.Err)
			p.activate(e)
			assert.False(t, s.completed.Peek())
		}
	}
	require.NoError(t, s.Navigate("directory-link"))
	assert.Equal(t, filepath.Join(dir, "directory-link"), s.Directory.Peek())
	assert.Empty(t, s.Entries())
	p.parent()
	p.Mode = FilePickerSave
	s.FilenameInput.SetText("broken-link")
	assert.False(t, p.Submit())
	assert.NotEmpty(t, s.Error.Peek())
	assert.Empty(t, s.pending.Peek())
}
func TestFilePickerBuildPurity(t *testing.T) {
	p := pickerFixture()
	f := p.State.filesystem.(*pickerTestFS)
	before := f.calls
	_, revision := p.State.list.Items.peekWithRevision()
	p.State.SelectPaths("/fixture/alpha.go")
	selection := p.State.SelectedPaths()
	for i := 0; i < 3; i++ {
		RenderToBuffer(p, 64, 22)
	}
	assert.Equal(t, before, f.calls)
	_, after := p.State.list.Items.peekWithRevision()
	assert.Equal(t, revision, after)
	assert.Equal(t, selection, p.State.SelectedPaths())
	assert.Empty(t, p.State.Error.Peek())
}
func TestFilePickerSnapshots(t *testing.T) {
	p := pickerFixture()
	p.MultiSelect = true
	AssertSnapshotNamed(t, "FilePicker_open", p, 70, 22, "Open picker with directories, regular files, Unicode and spaces; hidden files excluded")
	p.State.SelectPaths("/fixture/alpha.go", "/fixture/日本語.go")
	AssertSnapshotNamed(t, "FilePicker_multiselect", p, 70, 22, "Two files selected in one directory")
	p.nextFilter()
	AssertSnapshotNamed(t, "FilePicker_filter", p, 70, 22, "Go filter excludes beta.txt but directories remain navigable")
	p.toggleHidden()
	AssertSnapshotNamed(t, "FilePicker_hidden", p, 70, 22, "Hidden file becomes visible; selection cleared")
	p.Mode = FilePickerSave
	p.State.FilenameInput.SetText("alpha.go")
	AssertSnapshotNamed(t, "FilePicker_save", p, 70, 24, "Save mode shows filename and save action")
	p.Submit()
	AssertSnapshotNamed(t, "FilePicker_overwrite", p, 70, 24, "Explicit overwrite modal defaults to Keep existing")
	p.cancelOverwrite()
	p.State.FilenameInput.SetText("docs")
	p.State.FilterIndex.Set(0)
	p.Submit()
	AssertSnapshotNamed(t, "FilePicker_save_error", p, 70, 24, "Existing directory cannot be a save target")
	p.State.Reset()
	p.Mode = FilePickerDirectory
	require.NoError(t, p.State.Navigate("empty"))
	AssertSnapshotNamed(t, "FilePicker_empty_directory", p, 70, 22, "Empty directory still supports Choose")
	p.State.filesystem.(*pickerTestFS).deny = "/fixture"
	p.State.Navigate("..")
	AssertSnapshotNamed(t, "FilePicker_permission_error", p, 70, 22, "Injected permission error preserves the previous empty directory")
	p = pickerFixture()
	p.Filters = []FileFilter{{Label: "Broken", Patterns: []string{"["}}}
	AssertSnapshotNamed(t, "FilePicker_invalid_filter", p, 70, 22, "Invalid glob produces a clear error")
	for _, size := range [][2]int{{36, 18}, {12, 8}, {1, 1}} {
		p = pickerFixture()
		AssertSnapshotNamed(t, fmtPickerSize(size), p, size[0], size[1], "Narrow viewport clips safely")
	}
}
func fmtPickerSize(size [2]int) string {
	return fmt.Sprintf("FilePicker_narrow_%dx%d", size[0], size[1])
}
func TestFilePickerReactiveSequence(t *testing.T) {
	seq := newReactivitySequence(t, 70, 24, pickerFixture)
	seq.frame("initial", nil)
	seq.frame("filter", func(p FilePicker) { p.nextFilter() })
	seq.frame("hidden", func(p FilePicker) { p.toggleHidden() })
	seq.frame("select", func(p FilePicker) { p.State.SelectPaths("/fixture/alpha.go") })
	seq.frame("navigate", func(p FilePicker) { _ = p.State.Navigate("docs") })
	seq.frame("parent", func(p FilePicker) { p.parent() })
	seq.frame("error", func(p FilePicker) { _ = p.State.Navigate("missing") })
	seq.frame("cancel", func(p FilePicker) { p.Cancel() })
}
func TestFilePickerKeybinds(t *testing.T) {
	p := pickerFixture()
	require.True(t, matchKeybind(makeKeyEvent('f', uv.ModCtrl), p.Keybinds()))
	assert.Equal(t, 1, p.State.FilterIndex.Peek())
	require.NoError(t, p.State.Navigate("docs"))
	require.True(t, matchKeybind(makeKeyEvent(uv.KeyUp, uv.ModAlt), p.Keybinds()))
	assert.Equal(t, "/fixture", p.State.Directory.Peek())
	require.True(t, matchKeybind(makeKeyEvent(uv.KeyEscape, 0), p.Keybinds()))
	assert.True(t, p.State.completed.Peek())
}

func TestFilePickerOverwriteParentRemoved(t *testing.T) {
	p := pickerFixture()
	p.Mode = FilePickerSave
	require.NoError(t, p.State.Navigate("docs"))
	p.State.FilenameInput.SetText("guide.md")
	calls := 0
	p.OnSelect = func([]string) { calls++ }
	p.Submit()
	require.NotEmpty(t, p.State.pending.Peek())
	f := p.State.filesystem.(*pickerTestFS)
	delete(f.files, "docs/guide.md")
	delete(f.files, "docs")
	p.confirmOverwrite()
	assert.Zero(t, calls)
	assert.Empty(t, p.State.pending.Peek())
	assert.Contains(t, p.State.Error.Peek(), "parent directory")
}
func TestFilePickerSpaceSelectionBinding(t *testing.T) {
	p := pickerFixture()
	p.MultiSelect = true
	l := filePickerList{picker: p, List: List[FilePickerEntry]{State: p.State.list, MultiSelect: true}}
	require.True(t, matchKeybind(makeKeyEvent(' ', 0), l.Keybinds()))
	assert.Empty(t, p.State.SelectedPaths(), "directory is not toggled")
	p.State.list.CursorIndex.Set(3)
	require.True(t, matchKeybind(makeKeyEvent(' ', 0), l.Keybinds()))
	assert.Equal(t, []string{"/fixture/alpha.go"}, p.State.SelectedPaths())
	require.True(t, matchKeybind(makeKeyEvent(' ', 0), l.Keybinds()))
	assert.Empty(t, p.State.SelectedPaths())
}
