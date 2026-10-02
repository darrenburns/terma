package collections

import (
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/docs/widget-examples/demo"
)

// --8<-- [start:list]
func List() t.Widget {
	state := t.NewListState([]string{"Inbox", "Today", "Upcoming"})
	return t.List[string]{ID: "folders", State: state}
}

// --8<-- [end:list]

// --8<-- [start:table]
func Table() t.Widget {
	state := t.NewTableState([][]string{
		{"Mina", "Engineer"},
		{"Leon", "Designer"},
		{"Ada", "Writer"},
	})
	return t.Table[[]string]{
		ID: "people", State: state, SelectionMode: t.TableSelectionRow,
		Columns: []t.TableColumn{
			{Width: t.Cells(14), Header: t.Text{Content: "Name"}},
			{Width: t.Cells(20), Header: t.Text{Content: "Role"}},
		},
	}
}

// --8<-- [end:table]

// --8<-- [start:table-controls]
func TableControls() t.Widget {
	rows := [][]string{{"Mina", "Engineer"}, {"Leon", "Designer"}, {"Ada", "Writer"}}
	state := t.NewTableStateWithRowID(rows, func(row []string) string { return row[0] })
	return t.Table[[]string]{
		ID: "sortable-people", State: state,
		Columns: []t.TableColumn{
			{ID: "name", Header: t.Text{Content: "Name"}, Width: t.Cells(14),
				Resizable: true, MinWidth: 8, MaxWidth: 24},
			{ID: "role", Header: t.Text{Content: "Role"}, Width: t.Cells(20)},
		},
		Comparators: map[string]func([]string, []string) int{
			"name": func(a, b []string) int { return strings.Compare(a[0], b[0]) },
		},
		FrozenHeader: true,
		Style:        t.Style{Width: t.Cells(36), Height: t.Cells(6)},
	}
}

// --8<-- [end:table-controls]

// --8<-- [start:tree]
func Tree() t.Widget {
	state := t.NewTreeState([]t.TreeNode[string]{
		{Data: "Project", Children: []t.TreeNode[string]{
			{Data: "main.go"},
			{Data: "README.md"},
		}},
		{Data: "Archive"},
	})
	return t.Tree[string]{ID: "project-tree", State: state}
}

// --8<-- [end:tree]

// --8<-- [start:directorytree]
func DirectoryTree() t.Widget {
	state := t.NewDirectoryTreeState(".")
	return t.DirectoryTree{
		Tree: t.Tree[t.DirectoryEntry]{ID: "files", State: state},
	}
}

// --8<-- [end:directorytree]

// --8<-- [start:directorytree-preview]
func DirectoryTreePreview() t.Widget {
	state := t.NewTreeState([]t.TreeNode[t.DirectoryEntry]{
		{Data: t.DirectoryEntry{Name: "project", Path: "project", IsDir: true},
			Children: []t.TreeNode[t.DirectoryEntry]{
				{Data: t.DirectoryEntry{Name: "main.go", Path: "project/main.go"}},
				{Data: t.DirectoryEntry{Name: "README.md", Path: "project/README.md"}},
			}},
	})
	return t.DirectoryTree{Tree: t.Tree[t.DirectoryEntry]{ID: "files", State: state}}
}

// --8<-- [end:directorytree-preview]

// --8<-- [start:tabs]
func Tabs() t.Widget {
	state := t.NewTabState([]t.Tab{
		{Key: "home", Label: "Home", Content: t.Text{Content: "Welcome home."}},
		{Key: "settings", Label: "Settings", Content: t.Text{Content: "Your preferences."}},
	})
	return t.TabView{ID: "views", State: state, KeybindPattern: t.TabKeybindNumbers}
}

// --8<-- [end:tabs]

// --8<-- [start:breadcrumbs]
func Breadcrumbs() t.Widget {
	selected := t.NewSignal("Choose a location")
	path := []string{"Projects", "Terma", "Docs"}
	return t.Column{Children: []t.Widget{
		t.Breadcrumbs{
			Path:     path,
			OnSelect: func(index int) { selected.Set(path[index]) },
		},
		t.SignalText(selected, func(s string) string { return s }),
	}}
}

// --8<-- [end:breadcrumbs]

// --8<-- [start:keybindbar]
func KeybindBar() t.Widget {
	status := t.NewSignal("Ready")
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Button{ID: "save", Label: "Save", OnPress: func() { status.Set("Saved") }},
		t.SignalText(status, func(s string) string { return s }),
		t.KeybindBar{},
	}}
}

// --8<-- [end:keybindbar]

// --8<-- [start:jumper]
func Jumper() t.Widget {
	state := t.NewJumpState()
	folders := t.NewListState([]string{"Inbox", "Today", "Upcoming"})
	return t.Jumper{
		State: state, Dynamic: true,
		Targets: []t.JumpTarget{{Key: "1", ID: "folders"}},
		Child: t.Column{Spacing: 1, Children: []t.Widget{
			t.Text{Content: "Press Ctrl+O to jump"},
			t.List[string]{ID: "folders", State: folders},
		}},
	}
}

// --8<-- [end:jumper]

// --8<-- [start:jumper-preview]
func JumperPreview() t.Widget {
	widget := Jumper().(t.Jumper)
	widget.State.Activate()
	return widget
}

// --8<-- [end:jumper-preview]

func Examples() []demo.Example {
	return []demo.Example{
		{Name: "list", Widget: List, Width: 36, Height: 5},
		{Name: "table", Widget: Table, Width: 38, Height: 6},
		{Name: "table-controls", Widget: TableControls, Width: 36, Height: 6},
		{Name: "tree", Widget: Tree, Width: 36, Height: 6},
		{Name: "directorytree", Widget: DirectoryTreePreview, Width: 36, Height: 5},
		{Name: "tabs", Widget: Tabs, Width: 44, Height: 5},
		{Name: "breadcrumbs", Widget: Breadcrumbs, Width: 38, Height: 4},
		{Name: "keybindbar", Widget: KeybindBar, Width: 38, Height: 7},
		{Name: "jumper", Widget: JumperPreview, Width: 38, Height: 7},
	}
}
