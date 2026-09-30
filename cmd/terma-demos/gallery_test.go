package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
	"github.com/stretchr/testify/require"
)

// Exercise the same widget callbacks and render boundary used by the app.
// Real terminal input and focus routing are checked through terma-browser.
func renderGallery(g *Gallery) *t.Renderer {
	fm := t.NewFocusManager()
	fm.SetRootWidget(g)
	focused := t.NewAnySignal[t.Focusable](nil)
	r := t.NewRenderer(uv.NewBuffer(100, 30), 100, 30, fm, focused, t.NewAnySignal[t.Widget](nil))
	fm.SetFocusables(r.Render(g))
	focused.Set(fm.Focused())
	r.Render(g)
	return r
}

func galleryTheme(test *testing.T) {
	oldTheme, oldHint := t.CurrentThemeName(), demokit.SwitchHint
	t.SetTheme(demokit.Themes[0])
	demokit.SwitchHint = "[b $Accent]ctrl+g[/] [$TextMuted]demos[/]"
	test.Cleanup(func() { t.SetTheme(oldTheme); demokit.SwitchHint = oldHint })
}

func openDemo(test *testing.T, g *Gallery, title string) {
	test.Helper()
	g.Keybinds()[0].Action()
	entry := renderGallery(g).WidgetByID(paletteID + "-list")
	require.NotNil(test, entry)
	list := entry.EventWidget.(t.List[t.CommandPaletteItem])
	for index, item := range list.State.GetItems() {
		if item.Label == title {
			list.State.SelectIndex(index)
			list.OnSelect(item)
			return
		}
	}
	test.Fatalf("demo %q missing from switcher", title)
}

func TestGalleryNewWidgetSnapshots(test *testing.T) {
	galleryTheme(test)
	g := NewGallery(demos)
	test.Cleanup(func() { require.NoError(test, g.Close()) })
	for _, title := range []string{"SelectBox", "Forms", "Markdown", "Table controls", "File picker"} {
		require.Contains(test, renderGallery(g).ScreenText(), title)
	}
	t.AssertSnapshotNamed(test, "GalleryNewWidgetsHome", g, 100, 30)
	for _, title := range []string{"SelectBox", "Forms", "Markdown", "Table controls"} {
		test.Run(title, func(test *testing.T) {
			openDemo(test, g, title)
			require.Contains(test, renderGallery(g).ScreenText(), title)
			t.AssertSnapshot(test, g, 100, 30)
		})
	}
	// File picker uses a random temporary path: verify actual rendered fixtures
	// instead of embedding that nondeterministic path in a golden.
	openDemo(test, g, "File picker")
	screen := renderGallery(g).ScreenText()
	for _, text := range []string{"File picker", "alpha.go", "README.md", "Open F2", "Save F3", "Directory F4"} {
		require.Contains(test, screen, text)
	}
}

func TestGalleryKeepsSelectionAndFormStateWhenSwitching(test *testing.T) {
	galleryTheme(test)
	g := NewGallery(demos)
	test.Cleanup(func() { require.NoError(test, g.Close()) })
	openDemo(test, g, "SelectBox")
	selectBox := renderGallery(g).WidgetByID("environment").EventWidget.(t.SelectBox[string])
	selectBox.State.SetValue("stage")
	openDemo(test, g, "Forms")
	renderGallery(g).WidgetByID("load").EventWidget.(t.Button).OnPress()
	renderGallery(g).WidgetByID("save").EventWidget.(t.Button).OnPress()
	require.Contains(test, renderGallery(g).ScreenText(), "Submissions=1")
	openDemo(test, g, "SelectBox")
	require.Contains(test, renderGallery(g).ScreenText(), "Committed: stage")
	openDemo(test, g, "Forms")
	screen := renderGallery(g).ScreenText()
	require.Contains(test, screen, "Alice")
	require.Contains(test, screen, "Submissions=1")
	t.AssertSnapshotNamed(test, "GalleryFormAfterReturning", g, 100, 30)
	openDemo(test, g, "Home")
	require.Contains(test, renderGallery(g).ScreenText(), "open · state kept")
}

func TestGalleryKeepsFixturesUntilExit(test *testing.T) {
	galleryTheme(test)
	g := NewGallery(demos)
	test.Cleanup(func() { require.NoError(test, g.Close()) })
	openDemo(test, g, "File picker")
	picker := renderGallery(g).WidgetByID("files").EventWidget.(t.FilePicker)
	root := picker.State.Directory.Peek()
	require.DirExists(test, root)
	openDemo(test, g, "Markdown")
	require.DirExists(test, root)
	openDemo(test, g, "File picker")
	returned := renderGallery(g).WidgetByID("files").EventWidget.(t.FilePicker)
	require.Equal(test, root, returned.State.Directory.Peek())
	require.NoError(test, g.Close())
	_, err := os.Stat(root)
	require.ErrorIs(test, err, os.ErrNotExist)
}

type closingDemo struct {
	calls int
	err   error
}

func (*closingDemo) InitialFocus() string          { return "" }
func (*closingDemo) Build(t.BuildContext) t.Widget { return t.Text{Content: "Disposable demo"} }
func (d *closingDemo) Close() error                { d.calls++; return d.err }

var _ io.Closer = (*Gallery)(nil)

func TestGalleryClosesAllOpenedDemosEvenAfterAnError(test *testing.T) {
	failure := errors.New("cleanup failed")
	first, second := &closingDemo{err: failure}, &closingDemo{}
	g := NewGallery([]Entry{
		{Info: demokit.Info{Key: "one", Title: "One"}, New: func() demokit.Demo { return first }},
		{Info: demokit.Info{Key: "two", Title: "Two"}, New: func() demokit.Demo { return second }},
		{Info: demokit.Info{Key: "unused", Title: "Unused"}, New: func() demokit.Demo { test.Fatal("unopened demo allocated"); return nil }},
	})
	openDemo(test, g, "One")
	openDemo(test, g, "Two")
	openDemo(test, g, "One")
	require.Zero(test, first.calls)
	require.Zero(test, second.calls)
	require.ErrorIs(test, g.Close(), failure)
	require.Equal(test, 1, first.calls)
	require.Equal(test, 1, second.calls)
	require.NoError(test, g.Close())
	require.Equal(test, 1, first.calls)
	require.Equal(test, 1, second.calls)
}

func TestGalleryEntriesHaveUniqueKeys(test *testing.T) {
	seen := map[string]bool{}
	for _, entry := range demos {
		require.NotEmpty(test, entry.Key)
		require.False(test, seen[entry.Key], "duplicate gallery key %s", entry.Key)
		require.True(test, strings.HasPrefix(entry.Command, "./cmd/"))
		seen[entry.Key] = true
	}
}
