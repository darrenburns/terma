package terma

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

func TestFilePickerProbeRejectsReplacedOverwriteTarget(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target.txt")
	replacement := filepath.Join(dir, "replacement.txt")
	stamp := time.Unix(1000, 0)
	for name, body := range map[string]string{path: "old", replacement: "new"} {
		require.NoError(t, os.WriteFile(name, []byte(body), 0600))
		require.NoError(t, os.Chtimes(name, stamp, stamp))
	}
	state := NewFilePickerState(dir)
	state.FilenameInput.SetText("target.txt")
	selections := 0
	picker := FilePicker{ID: "identity-picker", State: state, Mode: FilePickerSave, OnSelect: func([]string) { selections++ }}
	require.False(t, picker.Submit(), "existing target requires confirmation")
	scene := newClickScene(t, picker, 80, 28)
	require.Contains(t, scene.renderer.ScreenText(), "Confirm overwrite")
	require.NoError(t, os.Rename(replacement, path), "different inode, same size, mode and mtime")
	scene.focus.FocusByID("identity-picker-overwrite-btn-1")
	require.Equal(t, "identity-picker-overwrite-btn-1", scene.focus.FocusedID())
	dispatchKey(scene.renderer, scene.focus, picker, makeKeyEvent(uv.KeyEnter, 0))
	scene.draw()
	require.Zero(t, selections, "an overwrite decision for the previous file must not approve its replacement")
	require.Contains(t, state.Error.Get(), "changed")
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new", string(content), "picker must never write the selected target")
}

type filePickerModalProbe struct {
	state               *FilePickerState
	visible             Signal[bool]
	selections, cancels int
}

func newFilePickerModalProbe() *filePickerModalProbe {
	return &filePickerModalProbe{state: NewFilePickerStateWithFileSystem("/fixture", newPickerTestFS()), visible: NewSignal(true)}
}
func (s *filePickerModalProbe) picker() FilePicker {
	return FilePicker{ID: "modal-picker", State: s.state, Mode: FilePickerSave, Style: Style{Width: Flex(1), Height: Cells(25)}, OnSelect: func([]string) { s.selections++ }, OnCancel: func() { s.cancels++ }}
}
func (s *filePickerModalProbe) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Button{ID: "outside-picker-modal", Label: "Outside"},
		Dialog{ID: "probe-picker-outer", Visible: s.visible.Get(), Title: "Parent picker", Style: Style{Width: Cells(80)}, Content: s.picker(), OnDismiss: func() { s.visible.Set(false) }},
	}}
}

// Use the live retained render path, preserving focus requests from handlers.
// The separate retained/full comparison also verifies nested backdrop output.
func filePickerProbeDraw(scene *clickScene) {
	for pass := 0; pass < 4; pass++ {
		before := scene.focus.FocusedID()
		scene.focused.Set(scene.focus.Focused())
		scene.focus.SetFocusables(scene.renderer.Update(scene.root))
		if pendingFocusID != "" {
			scene.focus.FocusByID(pendingFocusID)
			pendingFocusID = ""
		}
		if before == scene.focus.FocusedID() {
			return
		}
	}
	scene.t.Fatal("FilePicker focus did not settle after four frames")
}
func TestFilePickerProbeNestedOverwriteCancellation(t *testing.T) {
	previous := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previous })
	root := newFilePickerModalProbe()
	scene := newClickScene(t, root, 110, 36)
	filePickerProbeDraw(scene)
	scene.focus.FocusByID("modal-picker-filename")
	filePickerProbeDraw(scene)
	root.state.FilenameInput.SetText("alpha.go")
	require.False(t, root.picker().Submit())
	filePickerProbeDraw(scene)
	require.Equal(t, "modal-picker-overwrite-btn-0", scene.focus.FocusedID(), "safe Keep existing action gets focus")
	scene.snapshot("FilePicker_Probe_NestedConfirmation", "Nested overwrite confirmation keeps keyboard focus on Keep existing")
	key := func(event KeyEvent) {
		dispatchKey(scene.renderer, scene.focus, root, event)
		filePickerProbeDraw(scene)
	}
	key(makeKeyEvent(uv.KeyEscape, 0))
	require.Contains(t, scene.renderer.ScreenText(), "Parent picker")
	require.NotContains(t, scene.renderer.ScreenText(), "Confirm overwrite")
	require.Equal(t, "modal-picker-filename", scene.focus.FocusedID())
	require.Zero(t, root.cancels)
	require.Zero(t, root.selections)
	require.False(t, root.picker().Submit())
	filePickerProbeDraw(scene)
	require.Equal(t, "modal-picker-overwrite-btn-0", scene.focus.FocusedID())
	key(makeKeyEvent(uv.KeyEnter, 0))
	require.NotContains(t, scene.renderer.ScreenText(), "Confirm overwrite")
	require.Equal(t, "modal-picker-filename", scene.focus.FocusedID())
	require.Zero(t, root.cancels)
	require.Zero(t, root.selections)
	require.False(t, root.picker().Submit())
	filePickerProbeDraw(scene)
	key(makeKeyEvent(uv.KeyTab, 0))
	require.Equal(t, "modal-picker-overwrite-btn-1", scene.focus.FocusedID())
	key(makeKeyEvent(uv.KeyEnter, 0))
	require.Equal(t, 1, root.selections)
	require.False(t, root.picker().Submit())
	root.picker().Cancel()
	require.Equal(t, 1, root.selections)
	require.Zero(t, root.cancels)
}

func TestFilePickerProbeRejectsRetargetedSymlink(t *testing.T) {
	dir := t.TempDir()
	stamp := time.Unix(1000, 0)
	for _, name := range []string{"first.txt", "second.txt"} {
		path := filepath.Join(dir, name)
		require.NoError(t, os.WriteFile(path, []byte("same"), 0600))
		require.NoError(t, os.Chtimes(path, stamp, stamp))
	}
	link := filepath.Join(dir, "target.txt")
	require.NoError(t, os.Symlink("first.txt", link))
	state := NewFilePickerState(dir)
	state.FilenameInput.SetText("target.txt")
	selections := 0
	picker := FilePicker{ID: "symlink-picker", State: state, Mode: FilePickerSave, OnSelect: func([]string) { selections++ }}
	require.False(t, picker.Submit())
	scene := newClickScene(t, picker, 80, 28)
	require.NoError(t, os.Remove(link))
	require.NoError(t, os.Symlink("second.txt", link))
	scene.focus.FocusByID("symlink-picker-overwrite-btn-1")
	dispatchKey(scene.renderer, scene.focus, picker, makeKeyEvent(uv.KeyEnter, 0))
	require.Zero(t, selections)
	require.Contains(t, state.Error.Get(), "changed")
	for _, name := range []string{"first.txt", "second.txt"} {
		content, err := os.ReadFile(filepath.Join(dir, name))
		require.NoError(t, err)
		require.Equal(t, "same", string(content))
	}
}

func TestFilePickerProbeDirectVisibilityChangesRevalidate(t *testing.T) {
	picker := pickerFixture()
	picker.MultiSelect = true
	picker.State.ShowHidden.Set(true)
	picker.State.SelectPaths("/fixture/.secret.go", "/fixture/beta.txt", "/fixture/日本語.go")
	// Applications can change exposed signals without using the widget buttons.
	picker.State.ShowHidden.Set(false)
	picker.State.FilterIndex.Set(1)
	var selected []string
	picker.OnSelect = func(paths []string) { selected = paths }
	require.True(t, picker.Submit())
	require.Equal(t, []string{"/fixture/日本語.go"}, selected)
	picker.State.Reset()
	picker.State.SelectPaths("/fixture/alpha.go")
	require.NoError(t, picker.State.Refresh())
	require.Empty(t, picker.State.SelectedPaths(), "refresh cannot transfer selection to a different row")
	require.Error(t, picker.State.Navigate("missing"))
	require.Equal(t, "/fixture", picker.State.Directory.Get())
}

func FuzzFilePickerSaveNamesAndLifecycle(f *testing.F) {
	for _, name := range []string{"new.go", "space 日本語.go", " ", "../escape", "nul\x00name", "alpha.go", "docs", ".secret.go", "[bad"} {
		f.Add(name, []byte{0, 0, 1, 1, 2, 6, 0, 3, 6, 0, 5, 9, 0})
	}
	f.Fuzz(func(t *testing.T, name string, actions []byte) {
		if len(name) > 128 {
			name = name[:128]
		}
		name = strings.ToValidUTF8(name, "�")
		if len(actions) > 32 {
			actions = actions[:32]
		}
		filesystem := newPickerTestFS()
		state := NewFilePickerStateWithFileSystem("/fixture", filesystem)
		picker := FilePicker{ID: "fuzz-picker", State: state, Mode: FilePickerSave, Filters: []FileFilter{{Label: "All"}, {Label: "Go", Patterns: []string{"*.go"}}}}
		state.FilenameInput.SetText(name)
		completion := ""
		picker.OnSelect = func(paths []string) {
			require.Empty(t, completion, "at most one callback per attempt")
			completion = "selected"
			require.Len(t, paths, 1)
			candidateName := state.FilenameInput.GetText()
			require.NotEmpty(t, candidateName)
			require.NotEqual(t, ".", candidateName)
			require.NotEqual(t, "..", candidateName)
			require.False(t, strings.ContainsAny(candidateName, "/\\\x00"))
			if state.FilterIndex.Get() == 1 {
				require.True(t, strings.HasSuffix(candidateName, ".go"))
			}
			require.Equal(t, filepath.Join(state.Directory.Get(), candidateName), paths[0])
		}
		picker.OnCancel = func() {
			require.Empty(t, completion, "cancel cannot repeat or follow a completed selection")
			completion = "cancelled"
		}
		for i, action := range actions {
			switch action % 10 {
			case 0:
				picker.Submit()
			case 1:
				picker.Cancel()
			case 2:
				state.Reset()
				completion = ""
			case 3:
				if state.Navigate("docs") == nil {
					completion = ""
				}
			case 4:
				if state.Navigate("..") == nil {
					completion = ""
				}
			case 5:
				previous := state.Directory.Get()
				require.Error(t, state.Navigate("definitely-missing"))
				require.Equal(t, previous, state.Directory.Get())
			case 6:
				state.FilenameInput.SetText(name)
			case 7:
				state.FilterIndex.Set(int(action) % 2)
			case 8:
				state.ShowHidden.Set(!state.ShowHidden.Get())
			case 9:
				if state.Refresh() == nil {
					completion = ""
				}
			}
			if i%8 == 0 {
				RenderToBuffer(picker, 2+int(action)%60, 2+int(action)%24)
			}
		}
		require.Equal(t, newPickerTestFS().files, filesystem.files, "picker never writes its filesystem")
	})
}

type filePickerPairProbe struct {
	left, right FilePicker
	scroll      *ScrollState
}

func (s *filePickerPairProbe) Build(BuildContext) Widget {
	return Scrollable{State: s.scroll, Height: Cells(32), Child: Column{Style: Style{Height: Cells(45), Padding: EdgeInsets{Top: 2, Left: 2}}, Children: []Widget{
		Row{Style: Style{Height: Cells(28)}, Spacing: 2, Children: []Widget{s.left, s.right}},
		Text{Content: "Below both pickers", Height: Cells(12)},
	}}}
}
func TestFilePickerProbePaddedClippedInstancesStayIndependent(t *testing.T) {
	root := &filePickerPairProbe{left: pickerFixture(), right: pickerFixture(), scroll: NewScrollState()}
	root.left.ID = "left-picker"
	root.right.ID = "right-picker"
	root.left.MultiSelect = true
	root.right.MultiSelect = true
	root.left.State.SelectPaths("/fixture/alpha.go")
	root.right.State.SelectPaths("/fixture/beta.txt")
	scene := newClickScene(t, root, 128, 36)
	button := scene.renderer.WidgetByID("right-picker-hidden")
	require.NotNil(t, button)
	x, y := button.Visible.X, button.Visible.Y
	require.False(t, button.Visible.IsEmpty())
	scene.click(x, y, 0)
	require.True(t, root.right.State.ShowHidden.Get())
	require.False(t, root.left.State.ShowHidden.Get())
	require.Empty(t, root.right.State.SelectedPaths())
	scene.snapshot("FilePicker_Probe_IndependentPickers", "Two FilePickers in a padded Row and clipped Scrollable keep visibility and selections independent")
	require.Equal(t, []string{"/fixture/alpha.go"}, root.left.State.SelectedPaths())
	root.scroll.SetOffset(20)
	scene.draw()
	clipped := scene.renderer.WidgetByID("right-picker-hidden")
	require.True(t, clipped == nil || clipped.Visible.IsEmpty())
	scene.click(x, y, 0)
	require.True(t, root.right.State.ShowHidden.Get(), "old coordinates cannot toggle a clipped control")
	require.Equal(t, []string{"/fixture/alpha.go"}, root.left.State.SelectedPaths())
}

func TestFilePickerProbeRetainedNestedConfirmation(t *testing.T) {
	seq := newReactivitySequence(t, 110, 36, newFilePickerModalProbe)
	seq.frame("parent open", nil)
	seq.focus("modal-picker-filename")
	seq.frame("filename focus", nil)
	seq.frame("existing target", func(s *filePickerModalProbe) { s.state.FilenameInput.SetText("alpha.go"); s.picker().Submit() })
	require.Equal(t, "modal-picker-overwrite-btn-0", seq.actual.focus.FocusedID())
}
