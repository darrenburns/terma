package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	t "github.com/darrenburns/terma"
)

type DirectoryTreeDemo struct {
	treeState   *t.TreeState[t.DirectoryEntry]
	filterState *t.FilterState
	filterInput *t.TextInputState
	scrollState *t.ScrollState
	status      t.Signal[string]
	rootPath    string
}

func NewDirectoryTreeDemo() *DirectoryTreeDemo {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	abs, err := filepath.Abs(root)
	if err == nil {
		root = abs
	}

	return &DirectoryTreeDemo{
		treeState:   t.NewDirectoryTreeState(root),
		filterState: t.NewFilterState(),
		filterInput: t.NewTextInputState(""),
		scrollState: t.NewScrollState(),
		status:      t.NewSignal("Ready"),
		rootPath:    root,
	}
}

func (d *DirectoryTreeDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+f", Name: "Focus filter", Action: func() {
			t.RequestFocus("dir-filter")
		}},
		{Key: "esc", Name: "Clear filter", Action: d.clearFilter},
	}
}

func (d *DirectoryTreeDemo) clearFilter() {
	d.filterInput.SetText("")
	d.filterState.Query.Set("")
}

func (d *DirectoryTreeDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	tree := t.DirectoryTree{
		EagerLoad: true,
		Tree: t.Tree[t.DirectoryEntry]{
			ID:          "dir-tree",
			State:       d.treeState,
			Filter:      d.filterState,
			ScrollState: d.scrollState,
			MultiSelect: true,
			OnSelect: func(entry t.DirectoryEntry, selected []t.DirectoryEntry) {
				d.updateSelectStatus(entry, selected)
			},
			OnCursorChange: func(entry t.DirectoryEntry) {
				d.updateCursorStatus(entry)
			},
			Style: t.Style{
				Width: t.Flex(1),
			},
		},
	}

	return t.Column{
		ID:      "directory-tree-root",
		Spacing: 1,
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Directory Tree",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},
			t.Text{
				Content: fmt.Sprintf("Root: %s", d.rootPath),
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
			t.ParseMarkupToText("Use [b $Accent]Up/Down[/] or [b $Accent]j/k[/] to navigate | [b $Accent]Left/Right[/] to collapse/expand | [b $Accent]Shift+Up/Down[/] to multi-select", theme),
			t.Row{
				Spacing: 1,
				Children: []t.Widget{
					t.Text{Content: "Filter:", Style: t.Style{ForegroundColor: theme.TextMuted}},
					t.TextInput{
						ID:          "dir-filter",
						State:       d.filterInput,
						Placeholder: "Type to filter...",
						Width:       t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Surface,
							ForegroundColor: theme.Text,
						},
						OnChange: func(text string) {
							d.filterState.Query.Set(text)
						},
					},
					t.Text{Content: "Ctrl+F", Style: t.Style{ForegroundColor: theme.TextMuted}},
				},
			},
			t.Scrollable{
				ID:    "dir-tree-scroll",
				State: d.scrollState,
				Style: t.Style{
					Height:   t.Flex(1),
					Width:    t.Flex(1),
					Border:   t.SquareBorder(t.NewGradient(theme.Primary, theme.Background), t.BorderTitle("Files")),
					Padding:  t.EdgeInsetsAll(1),
					MinWidth: t.Cells(20),
				},
				ScrollbarThumbColor: theme.ScrollbarThumb,
				ScrollbarTrackColor: theme.ScrollbarTrack,
				Child:               tree,
			},
			t.Text{
				Content: d.status.Get(),
				Style: t.Style{
					ForegroundColor: theme.TextMuted,
				},
			},
			t.ParseMarkupToText("Press [b $Error]Ctrl+C[/] to quit", theme),
		},
	}
}

func (d *DirectoryTreeDemo) updateSelectStatus(cursor t.DirectoryEntry, selected []t.DirectoryEntry) {
	if len(selected) == 0 {
		d.updateCursorStatus(cursor)
		return
	}

	paths := make([]string, 0, len(selected))
	for _, entry := range selected {
		if entry.Path == "" {
			paths = append(paths, entry.Name)
			continue
		}
		paths = append(paths, entry.Path)
	}
	d.status.Set(fmt.Sprintf("Selected %d: %s", len(selected), strings.Join(paths, ", ")))
}

func (d *DirectoryTreeDemo) updateCursorStatus(entry t.DirectoryEntry) {
	if entry.Err != nil {
		d.status.Set(fmt.Sprintf("Error: %s", entry.Err))
		return
	}

	kind := "File"
	if entry.IsDir {
		kind = "Dir"
	}
	label := entry.Path
	if label == "" {
		label = entry.Name
	}
	if label == "" {
		label = "(unknown)"
	}
	d.status.Set(fmt.Sprintf("%s: %s", kind, label))
}

func main() {
	app := NewDirectoryTreeDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
