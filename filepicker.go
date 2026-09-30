package terma

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FilePickerMode determines the kind of path returned by a FilePicker.
type FilePickerMode int

const (
	FilePickerOpen FilePickerMode = iota
	FilePickerSave
	FilePickerDirectory
)

// FileFilter matches basenames using filepath.Match. Empty Patterns matches all.
type FileFilter struct {
	Label    string
	Patterns []string
}

// FilePickerFileSystem supplies directory listings and metadata. Implementations
// accept operating-system paths; the default delegates to os. It never writes.
type FilePickerFileSystem interface {
	ReadDir(string) ([]fs.DirEntry, error)
	Stat(string) (fs.FileInfo, error)
	Lstat(string) (fs.FileInfo, error)
}
type filePickerOS struct{}

func (filePickerOS) ReadDir(p string) ([]fs.DirEntry, error) { return os.ReadDir(p) }
func (filePickerOS) Stat(p string) (fs.FileInfo, error)      { return os.Stat(p) }
func (filePickerOS) Lstat(p string) (fs.FileInfo, error)     { return os.Lstat(p) }

// FilePickerEntry is one direct child of the current directory. Err marks an
// unavailable entry (for example a broken symlink). Paths remain lexical.
type FilePickerEntry struct {
	Name, Path       string
	IsDir, IsSymlink bool
	Err              error
	mode             fs.FileMode
}

// FilePickerState owns the current directory, editable path and filename, and
// selection. Construct once outside Build. Navigate and Refresh perform bounded,
// synchronous filesystem reads; Build never performs IO.
type FilePickerState struct {
	viewKey       string
	Directory     Signal[string]
	Error         Signal[string]
	ShowHidden    Signal[bool]
	FilterIndex   Signal[int]
	PathInput     *TextInputState
	FilenameInput *TextInputState
	filesystem    FilePickerFileSystem
	list          *ListState[FilePickerEntry]
	scroll        *ScrollState
	filter        *FilterState
	pending       Signal[string]
	completed     Signal[bool]
	pendingInfo   fs.FileInfo
}

func NewFilePickerState(startDir string) *FilePickerState {
	return NewFilePickerStateWithFileSystem(startDir, filePickerOS{})
}

// NewFilePickerStateWithFileSystem supports deterministic or virtual filesystems.
// A nil filesystem uses the host OS. Initialization errors are available in Error.
func NewFilePickerStateWithFileSystem(startDir string, filesystem FilePickerFileSystem) *FilePickerState {
	if filesystem == nil {
		filesystem = filePickerOS{}
	}
	s := &FilePickerState{Directory: NewSignal(""), Error: NewSignal(""), ShowHidden: NewSignal(false), FilterIndex: NewSignal(0), PathInput: NewTextInputState(startDir), FilenameInput: NewTextInputState(""), filesystem: filesystem, list: NewListState([]FilePickerEntry{}), scroll: NewScrollState(), filter: NewFilterState(), pending: NewSignal(""), completed: NewSignal(false)}
	s.filter.Query.Set("visible")
	_ = s.Navigate(startDir)
	return s
}

// Navigate reads one directory. A file path navigates to its parent and seeds
// filename/cursor. Failed navigation preserves the existing directory and rows.
func (s *FilePickerState) Navigate(path string) error {
	if path == "" {
		path = "."
	}
	if !filepath.IsAbs(path) && s.Directory.Peek() != "" {
		path = filepath.Join(s.Directory.Peek(), path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return s.fail("Resolve path", path, err)
	}
	info, err := s.filesystem.Stat(absolute)
	if err != nil {
		return s.fail("Inspect", absolute, err)
	}
	filename := ""
	if !info.IsDir() {
		if !info.Mode().IsRegular() {
			return s.fail("Navigate", absolute, fmt.Errorf("not a directory or regular file"))
		}
		filename = filepath.Base(absolute)
		absolute = filepath.Dir(absolute)
	}
	entries, err := s.filesystem.ReadDir(absolute)
	if err != nil {
		return s.fail("Read directory", absolute, err)
	}
	rows := make([]FilePickerEntry, 0, len(entries))
	for _, entry := range entries {
		// ReadDir implementations must return direct children, never arbitrary paths.
		name := entry.Name()
		if name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsRune(name, 0) {
			continue
		}
		row := FilePickerEntry{Name: name, Path: filepath.Join(absolute, name), IsDir: entry.IsDir(), IsSymlink: entry.Type()&fs.ModeSymlink != 0, mode: entry.Type()}
		if row.IsSymlink {
			target, statErr := s.filesystem.Stat(row.Path)
			row.Err = statErr
			if statErr == nil {
				row.IsDir = target.IsDir()
				row.mode = target.Mode()
			}
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].IsDir != rows[j].IsDir {
			return rows[i].IsDir
		}
		return rows[i].Name < rows[j].Name
	})
	s.list.ClearSelection()
	s.list.ClearAnchor()
	s.list.SetItems(rows)
	s.list.SelectIndex(0)
	s.scroll.SetOffset(0)
	s.Directory.Set(absolute)
	s.PathInput.SetText(absolute)
	s.FilenameInput.SetText(filename)
	if filename != "" {
		for i, row := range rows {
			if row.Name == filename {
				s.list.SelectIndex(i)
				break
			}
		}
	}
	s.Reset()
	return nil
}
func (s *FilePickerState) fail(action, path string, err error) error {
	s.Error.Set(fmt.Sprintf("%s %q: %v", action, path, err))
	return err
}

// Refresh reloads the current directory and clears stale selections.
func (s *FilePickerState) Refresh() error {
	name := s.FilenameInput.GetText()
	err := s.Navigate(s.Directory.Peek())
	if err == nil {
		s.FilenameInput.SetText(name)
	}
	return err
}

// Reset starts a new selection attempt without changing the directory.
func (s *FilePickerState) Reset() {
	s.pending.Set("")
	s.pendingInfo = nil
	s.completed.Set(false)
	s.Error.Set("")
}

// Entries returns a copy of the currently loaded direct children, before filtering.
func (s *FilePickerState) Entries() []FilePickerEntry {
	return append([]FilePickerEntry(nil), s.list.GetItems()...)
}

// SelectPaths selects matching loaded entries. Submit still validates visibility
// and mode, so this cannot submit filtered or unavailable paths.
func (s *FilePickerState) SelectPaths(paths ...string) {
	wanted := make(map[string]bool, len(paths))
	for _, p := range paths {
		wanted[filepath.Clean(p)] = true
	}
	s.list.ClearSelection()
	s.list.ClearAnchor()
	for i, e := range s.list.GetItems() {
		if wanted[e.Path] {
			s.list.Select(i)
		}
	}
}

// SelectedPaths returns loaded selected paths in directory display order.
func (s *FilePickerState) SelectedPaths() []string {
	var paths []string
	for _, e := range s.list.SelectedItems() {
		paths = append(paths, e.Path)
	}
	return paths
}

// FilePicker selects paths; it never opens, writes, or deletes selected files.
// Give each mounted picker a stable ID. With a bounded height the list scrolls.
type FilePicker struct {
	ID          string
	State       *FilePickerState
	Mode        FilePickerMode
	Filters     []FileFilter
	MultiSelect bool
	OnSelect    func([]string)
	OnCancel    func()
	Style       Style
}

func (p FilePicker) WidgetID() string { return p.ID }
func (p FilePicker) GetStyle() Style  { return p.Style }
func (p FilePicker) GetContentDimensions() (Dimension, Dimension) {
	return p.Style.Width, p.Style.Height
}
func (p FilePicker) id(part string) string {
	base := p.ID
	if base == "" {
		base = fmt.Sprintf("filepicker-%p", p.State)
	}
	return base + "-" + part
}
func (p FilePicker) OnKey(KeyEvent) bool { return false }
func (p FilePicker) Keybinds() []Keybind {
	if p.State == nil {
		return nil
	}
	return []Keybind{
		{Key: "escape", Name: "Cancel", Action: p.Cancel},
		{Key: "ctrl+l", Name: "Path", Action: func() { RequestFocus(p.id("path")) }},
		{Key: "alt+up", Name: "Parent", Action: p.parent},
		{Key: "ctrl+f", Name: "Filter", Action: p.nextFilter},
		{Key: "ctrl+enter", Name: "Choose", Action: func() { p.Submit() }},
	}
}
func (p FilePicker) filterIndex() int {
	if len(p.Filters) == 0 {
		return 0
	}
	return clampInt(p.State.FilterIndex.Peek(), 0, len(p.Filters)-1)
}
func (p FilePicker) filterError() error {
	if len(p.Filters) == 0 {
		return nil
	}
	for _, pattern := range p.Filters[p.filterIndex()].Patterns {
		if _, err := filepath.Match(pattern, ""); err != nil {
			return fmt.Errorf("invalid file filter %q: %w", pattern, err)
		}
	}
	return nil
}
func (p FilePicker) matches(name string) bool {
	if len(p.Filters) == 0 {
		return true
	}
	patterns := p.Filters[p.filterIndex()].Patterns
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		if ok, _ := filepath.Match(pattern, name); ok {
			return true
		}
	}
	return false
}
func (p FilePicker) visible(e FilePickerEntry) bool {
	if !p.State.ShowHidden.Peek() && strings.HasPrefix(e.Name, ".") {
		return false
	}
	if e.IsDir {
		return true
	}
	return p.Mode != FilePickerDirectory && p.matches(e.Name)
}
func (p FilePicker) clearSelection() {
	p.State.list.ClearSelection()
	p.State.list.ClearAnchor()
	p.State.pending.Set("")
	p.State.Error.Set("")
}
func (p FilePicker) parent() {
	if p.State == nil {
		return
	}
	if p.State.Navigate(filepath.Dir(p.State.Directory.Peek())) == nil {
		RequestFocus(p.id("list"))
	}
}
func (p FilePicker) nextFilter() {
	if p.State == nil || len(p.Filters) == 0 {
		return
	}
	p.clearSelection()
	p.State.FilterIndex.Set((p.filterIndex() + 1) % len(p.Filters))
	p.State.list.SelectIndex(0)
	p.State.scroll.SetOffset(0)
}
func (p FilePicker) toggleHidden() {
	p.clearSelection()
	p.State.ShowHidden.Update(func(v bool) bool { return !v })
	p.State.list.SelectIndex(0)
	p.State.scroll.SetOffset(0)
}
func (p FilePicker) activate(e FilePickerEntry) {
	if p.State.completed.Peek() {
		return
	}
	if !p.visible(e) {
		return
	}
	if e.Err != nil {
		p.State.fail("Inspect", e.Path, e.Err)
		return
	}
	if e.IsDir {
		if p.State.Navigate(e.Path) == nil {
			RequestFocus(p.id("list"))
		}
		return
	}
	if p.Mode == FilePickerSave {
		p.State.FilenameInput.SetText(e.Name)
		RequestFocus(p.id("filename"))
		return
	}
	if p.Mode == FilePickerOpen {
		p.submitOpen([]string{e.Path})
	}
}
func (p FilePicker) fail(message string) bool { p.State.Error.Set(message); return false }
func (p FilePicker) complete(paths []string) bool {
	if p.State.completed.Peek() {
		return false
	}
	p.State.pending.Set("")
	p.State.Error.Set("")
	p.State.completed.Set(true)
	if p.OnSelect != nil {
		p.OnSelect(append([]string(nil), paths...))
	}
	return true
}

// Cancel dismisses pending overwrite confirmation, or ends the current attempt
// once when no confirmation is open.
func (p FilePicker) Cancel() {
	if p.State == nil || p.State.completed.Peek() {
		return
	}
	if p.State.pending.Peek() != "" {
		p.cancelOverwrite()
		return
	}
	p.State.completed.Set(true)
	p.State.Error.Set("")
	if p.OnCancel != nil {
		p.OnCancel()
	}
}

// Submit validates the current mode and paths, or requests explicit overwrite
// confirmation. True means OnSelect was emitted (or completion with nil callback).
func (p FilePicker) Submit() bool {
	if p.State == nil || p.State.completed.Peek() || p.State.pending.Peek() != "" {
		return false
	}
	if p.Mode < FilePickerOpen || p.Mode > FilePickerDirectory {
		return p.fail("Unknown file picker mode")
	}
	if err := p.filterError(); err != nil {
		return p.fail(err.Error())
	}
	if p.State.Directory.Peek() == "" {
		return p.fail("Choose an accessible directory first")
	}
	switch p.Mode {
	case FilePickerDirectory:
		path := p.State.Directory.Peek()
		info, err := p.State.filesystem.Stat(path)
		if err != nil {
			p.State.fail("Inspect directory", path, err)
			return false
		}
		if !info.IsDir() {
			return p.fail("The current path is no longer a directory")
		}
		return p.complete([]string{path})
	case FilePickerSave:
		return p.submitSave()
	default:
		var paths []string
		if p.MultiSelect {
			for _, e := range p.State.list.SelectedItems() {
				if p.visible(e) && !e.IsDir && e.Err == nil {
					paths = append(paths, e.Path)
				}
			}
		}
		if len(paths) == 0 {
			e, ok := p.State.list.SelectedItem()
			if !ok || !p.visible(e) {
				return p.fail("Select a visible file")
			}
			if e.IsDir {
				p.activate(e)
				return false
			}
			paths = []string{e.Path}
		}
		return p.submitOpen(paths)
	}
}
func (p FilePicker) submitOpen(paths []string) bool {
	if err := p.filterError(); err != nil {
		return p.fail(err.Error())
	}
	for _, path := range paths {
		if filepath.Dir(path) != p.State.Directory.Peek() {
			return p.fail("Directory changed; navigate again before selecting files")
		}
		info, err := p.State.filesystem.Stat(path)
		if err != nil {
			p.State.fail("Inspect file", path, err)
			return false
		}
		if !info.Mode().IsRegular() {
			return p.fail("Only regular files can be opened: " + path)
		}
	}
	return p.complete(paths)
}
func (p FilePicker) saveCandidate() (string, error) {
	name := p.State.FilenameInput.GetText()
	if name == "" || name == "." || name == ".." || strings.ContainsRune(name, 0) || filepath.Base(name) != name || strings.ContainsAny(name, "/\\") {
		return "", fmt.Errorf("Enter a filename without directory separators")
	}
	if !p.matches(name) {
		return "", fmt.Errorf("Filename does not match the selected file filter")
	}
	return filepath.Join(p.State.Directory.Peek(), name), nil
}
func (p FilePicker) submitSave() bool {
	path, err := p.saveCandidate()
	if err != nil {
		return p.fail(err.Error())
	}
	info, err := p.State.filesystem.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			parent, parentErr := p.State.filesystem.Stat(filepath.Dir(path))
			if parentErr != nil {
				p.State.fail("Inspect parent directory", filepath.Dir(path), parentErr)
				return false
			}
			if !parent.IsDir() {
				return p.fail("Save parent is no longer a directory")
			}
			return p.complete([]string{path})
		}
		p.State.fail("Inspect save target", path, err)
		return false
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		info, err = p.State.filesystem.Stat(path)
		if err != nil {
			p.State.fail("Inspect symlink target", path, err)
			return false
		}
	}
	if !info.Mode().IsRegular() {
		return p.fail("Save target must be a regular file, not a directory or special file")
	}
	p.State.pendingInfo = info
	p.State.pending.Set(path)
	p.State.Error.Set("")
	return false
}
func (p FilePicker) cancelOverwrite() {
	p.State.pending.Set("")
	p.State.pendingInfo = nil
	RequestFocus(p.id("filename"))
}
func (p FilePicker) confirmOverwrite() {
	path := p.State.pending.Peek()
	if path == "" || p.State.completed.Peek() {
		return
	}
	candidate, err := p.saveCandidate()
	if err != nil || candidate != path || p.Mode != FilePickerSave {
		p.cancelOverwrite()
		p.fail("Save target changed; submit again to review it")
		return
	}
	if err := p.filterError(); err != nil {
		p.cancelOverwrite()
		p.fail(err.Error())
		return
	}
	info, err := p.State.filesystem.Lstat(path)
	if err != nil && !os.IsNotExist(err) {
		p.cancelOverwrite()
		p.State.fail("Inspect save target", path, err)
		return
	}
	if err == nil && info.Mode()&fs.ModeSymlink != 0 {
		info, err = p.State.filesystem.Stat(path)
		if err != nil {
			p.cancelOverwrite()
			p.State.fail("Inspect symlink target", path, err)
			return
		}
	}
	if err == nil && !info.Mode().IsRegular() {
		p.cancelOverwrite()
		p.fail("Save target changed to a directory or special file")
		return
	}
	// OS FileInfo carries identity as well as metadata. A replacement (or a
	// retargeted symlink) can have identical size, mode and modification time.
	// SameFile returns false for virtual FileInfo, so preserve its metadata
	// fallback by first checking whether the pending identity is supported.
	identityChanged := err == nil && p.State.pendingInfo != nil && os.SameFile(p.State.pendingInfo, p.State.pendingInfo) && !os.SameFile(p.State.pendingInfo, info)
	if err == nil && p.State.pendingInfo != nil && (identityChanged || info.Size() != p.State.pendingInfo.Size() || !info.ModTime().Equal(p.State.pendingInfo.ModTime()) || info.Mode() != p.State.pendingInfo.Mode()) {
		p.cancelOverwrite()
		p.fail("Save target changed; submit again to review it")
		return
	}
	parent, parentErr := p.State.filesystem.Stat(filepath.Dir(path))
	if parentErr != nil || !parent.IsDir() {
		p.cancelOverwrite()
		if parentErr != nil {
			p.State.fail("Inspect parent directory", filepath.Dir(path), parentErr)
		} else {
			p.fail("Save parent is no longer a directory")
		}
		return
	}
	p.complete([]string{path})
}
