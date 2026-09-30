// Package filepickerdemo demonstrates path selection using disposable fixtures.
package filepickerdemo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

type app struct {
	probe               bool
	root                string
	state               *t.FilePickerState
	mode                t.Signal[t.FilePickerMode]
	message             t.Signal[string]
	selections, cancels int
	modal               t.Signal[bool]
}

func (a *app) picker() t.FilePicker {
	return t.FilePicker{ID: "files", State: a.state, Mode: a.mode.Get(), MultiSelect: true, Filters: []t.FileFilter{{Label: "All"}, {Label: "Go", Patterns: []string{"*.go"}}, {Label: "Markdown", Patterns: []string{"*.md"}}}, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)}, OnSelect: func(paths []string) {
		a.selections++
		var names []string
		for _, p := range paths {
			names = append(names, strings.TrimPrefix(p, a.root+string(filepath.Separator)))
		}
		a.message.Set(fmt.Sprintf("Selected #%d: %s · no file written", a.selections, strings.Join(names, ", ")))
		a.modal.Set(false)
	}, OnCancel: func() {
		a.cancels++
		a.message.Set(fmt.Sprintf("Cancelled #%d · no selection callback", a.cancels))
		a.modal.Set(false)
	}}
}
func (a *app) setMode(mode t.FilePickerMode) {
	a.state.Reset()
	a.mode.Set(mode)
	t.RequestFocus("files-list")
}
func (a *app) Keybinds() []t.Keybind {
	keys := []t.Keybind{{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme}, {Key: "f2", Name: "Open", Action: func() { a.setMode(t.FilePickerOpen) }}, {Key: "f3", Name: "Save", Action: func() { a.setMode(t.FilePickerSave) }}, {Key: "f4", Name: "Directory", Action: func() { a.setMode(t.FilePickerDirectory) }}, {Key: "f5", Name: "Reset", Action: func() {
		_ = a.state.Navigate(a.root)
		a.message.Set("Fresh temporary fixtures · F2 open / F3 save / F4 directory")
	}}, {Key: "f6", Name: "Modal", Action: func() { a.state.Reset(); a.modal.Set(true) }}}
	if a.probe {
		keys = append(keys, t.Keybind{Key: "f7", Name: "Simulate replacement", Action: a.replaceProbeTarget})
	}
	return keys
}
func (a *app) Build(ctx t.BuildContext) t.Widget {
	picker := a.picker()
	body := t.Widget(picker)
	if a.modal.Get() {
		body = t.Text{Content: "Picker is open in a modal. Overwrite confirmation nests inside it."}
	}
	status := t.SignalText(a.message, func(s string) string { return s })
	status.LayoutStyle.Width = t.Flex(1)
	content := t.Column{Spacing: 1, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1), Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
		t.Text{Content: "FILE PICKER LAB · temporary fixtures only", Style: t.Style{Bold: true, ForegroundColor: ctx.Theme().Primary}},
		t.Row{Spacing: 1, Children: []t.Widget{t.Button{ID: "mode-open", Label: "Open F2", OnPress: func() { a.setMode(t.FilePickerOpen) }}, t.Button{ID: "mode-save", Label: "Save F3", OnPress: func() { a.setMode(t.FilePickerSave) }}, t.Button{ID: "mode-directory", Label: "Directory F4", OnPress: func() { a.setMode(t.FilePickerDirectory) }}}}, body, status,
		t.Dialog{ID: "outer-picker", Visible: a.modal.Get(), Title: "Choose a path", Content: picker, Style: t.Style{Width: t.Percent(90), Height: t.Percent(90)}, OnDismiss: func() { picker.Cancel(); a.modal.Set(false) }},
	}}
	return t.Dock{Style: t.Style{BackgroundColor: ctx.Theme().Background}, Top: []t.Widget{demokit.Header{Title: "File picker", Tagline: "Choose paths without writing files"}}, Bottom: []t.Widget{demokit.Footer(ctx.Theme())}, Body: content}
}

// Info describes this demo in the gallery.
var Info = demokit.Info{Key: "filepicker", Title: "File picker", Description: "Open, save and directory selection with temporary fixtures"}

// New creates a demo backed by its own disposable fixture directory.
func New() demokit.Demo { return createDemo(false) }

// NewProbe creates the nested-save verification demo.
func NewProbe() demokit.Demo { return createDemo(true) }
func createDemo(probe bool) demokit.Demo {
	root, err := createFixtures()
	if err != nil {
		return setupError{err: err}
	}
	a := &app{root: root, state: t.NewFilePickerState(root), mode: t.NewSignal(t.FilePickerOpen), message: t.NewSignal("Temporary fixtures · F2 open / F3 save / F4 directory · F6 modal"), modal: t.NewSignal(false), probe: probe}
	if probe {
		a.mode.Set(t.FilePickerSave)
		a.modal.Set(true)
		a.state.FilenameInput.SetText("alpha.go")
		a.message.Set("Probe: Save alpha.go, then F7 simulates same-metadata replacement before Overwrite")
	}
	return a
}
func (a *app) InitialFocus() string {
	if a.probe {
		return "files-filename"
	}
	return "files-list"
}

// Close removes only this demo's temporary fixtures. Switching demos keeps them alive.
func (a *app) Close() error {
	if a.root == "" {
		return nil
	}
	err := os.RemoveAll(a.root)
	if err == nil {
		a.root = ""
	}
	return err
}

type setupError struct{ err error }

func (s setupError) InitialFocus() string { return "" }
func (s setupError) Build(ctx t.BuildContext) t.Widget {
	return t.Text{Content: "Cannot create File picker demo fixtures: " + s.err.Error(), Style: t.Style{ForegroundColor: ctx.Theme().Error}}
}
func createFixtures() (root string, err error) {
	root, err = os.MkdirTemp("", "terma-filepicker-demo-")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(root)
		}
	}()
	for _, dir := range []string{"docs", "empty", "folder with spaces"} {
		if err = os.Mkdir(filepath.Join(root, dir), 0755); err != nil {
			return root, err
		}
	}
	for name, content := range map[string]string{"alpha.go": "package alpha\n", "beta.txt": "beta\n", "README.md": "# Demo\n", "space name.go": "package space\n", "日本語.go": "package unicode\n", ".hidden.go": "package hidden\n", "docs/guide.md": "# Guide\n"} {
		if err = os.WriteFile(filepath.Join(root, name), []byte(content), 0644); err != nil {
			return root, err
		}
	}
	_ = os.Symlink("docs", filepath.Join(root, "link-to-docs"))
	_ = os.Symlink("missing", filepath.Join(root, "broken-link"))
	return root, nil
}
