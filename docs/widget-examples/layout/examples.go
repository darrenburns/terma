package layout

import (
	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/docs/widget-examples/demo"
)

// --8<-- [start:row-column]
func RowColumn() t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Project files"},
		t.Row{Spacing: 3, Children: []t.Widget{
			t.Text{Content: "main.go"},
			t.Text{Content: "app.go"},
			t.Text{Content: "README.md"},
		}},
		t.Text{Content: "3 files"},
	}}
}

// --8<-- [end:row-column]

// --8<-- [start:dock]
func Dock() t.Widget {
	return t.Dock{
		Top: []t.Widget{t.Text{Content: "Files", Style: t.Style{
			BackgroundColor: t.RGB(46, 65, 91),
		}}},
		Bottom: []t.Widget{t.Text{Content: "Ready"}},
		Left: []t.Widget{t.Text{Content: "src\ndocs\ntests", Style: t.Style{
			Width: t.Cells(12), BackgroundColor: t.RGB(33, 43, 61),
		}}},
		Body: t.Text{Content: "Select a file to preview."},
	}
}

// --8<-- [end:dock]

// --8<-- [start:scrollable]
func Scrollable() t.Widget {
	return t.Scrollable{
		ID: "log", State: t.NewScrollState(), Focusable: true,
		Style: t.Style{Height: t.Cells(4), Border: t.Border{Style: t.BorderRounded}},
		Child: t.Text{Content: "Connecting\nConnected\nReading files\nChecking changes\nBuilding\nTesting\nComplete"},
	}
}

// --8<-- [end:scrollable]

// --8<-- [start:spacer]
func Spacer() t.Widget {
	return t.Row{
		Style: t.Style{Width: t.Flex(1)},
		Children: []t.Widget{
			t.Text{Content: "Project files"},
			t.Spacer{},
			t.Text{Content: "3 files"},
		},
	}
}

// --8<-- [end:spacer]

// --8<-- [start:splitpane]
func SplitPane() t.Widget {
	return t.SplitPane{
		ID: "editor", State: t.NewSplitPaneState(0.35),
		First:  t.Text{Content: "Files\nmain.go\napp.go"},
		Second: t.Text{Content: "Preview\nSelect a file."},
	}
}

// --8<-- [end:splitpane]

// --8<-- [start:stack]
func Stack() t.Widget {
	return t.Stack{
		Style: t.Style{Width: t.Cells(34), Height: t.Cells(5)},
		Children: []t.Widget{
			t.PositionedFill(t.Text{Content: "Preview", Style: t.Style{
				BackgroundColor: t.RGB(33, 43, 61),
			}}),
			t.Positioned{
				Top: t.IntPtr(1), Right: t.IntPtr(1),
				Child: t.Text{Content: "Draft", Style: t.Style{
					BackgroundColor: t.RGB(77, 64, 30),
				}},
			},
		},
	}
}

// --8<-- [end:stack]

// --8<-- [start:floating]
func Floating() t.Widget {
	return t.Column{Children: []t.Widget{
		t.Text{Content: "Project files\nmain.go\napp.go\nREADME.md"},
		t.Floating{
			Visible: true,
			Config:  t.FloatConfig{Position: t.FloatPositionCenter},
			Child: t.Text{Content: "All changes saved", Style: t.Style{
				Border: t.Border{Style: t.BorderRounded}, Padding: t.EdgeInsetsXY(2, 1),
				BackgroundColor: t.RGB(33, 43, 61),
			}},
		},
	}}
}

// --8<-- [end:floating]

// --8<-- [start:floating-geometry]
func FloatingGeometry() t.Widget {
	return t.Column{Children: []t.Widget{
		t.Text{ID: "file", Content: "main.go", Style: t.Style{Width: t.Cells(24)}},
		t.Floating{
			Visible: true,
			Config:  t.FloatConfig{AnchorID: "file", Anchor: t.AnchorBottomLeft},
			BuildChild: func(_ t.BuildContext, g t.FloatGeometry) t.Widget {
				if !g.AnchorFound || g.AnchorVisibleBounds.Height == 0 {
					return nil
				}
				return t.Text{Content: "Updated today", Style: t.Style{
					Width: t.Cells(g.AnchorBounds.Width),
				}}
			},
		},
	}}
}

// --8<-- [end:floating-geometry]

// --8<-- [start:focustrap]
func FocusTrap() t.Widget {
	return t.FocusTrap{
		ID: "profile", Active: true,
		Child: t.Column{Spacing: 1, Children: []t.Widget{
			t.Text{Content: "Edit profile"},
			t.TextInput{ID: "name", State: t.NewTextInputState("Ada"), Style: t.Style{Width: t.Cells(24)}},
			t.TextInput{ID: "role", State: t.NewTextInputState("Developer"), Style: t.Style{Width: t.Cells(24)}},
		}},
	}
}

// --8<-- [end:focustrap]

// --8<-- [start:switcher]
func Switcher() t.Widget {
	return &switcherExample{active: t.NewSignal("files")}
}

type switcherExample struct{ active t.Signal[string] }

func (s *switcherExample) Build(t.BuildContext) t.Widget {
	return t.Column{Spacing: 1, Children: []t.Widget{
		t.Row{Spacing: 1, Children: []t.Widget{
			t.Button{Label: "Files", OnPress: func() { s.active.Set("files") }},
			t.Button{Label: "Activity", OnPress: func() { s.active.Set("activity") }},
		}},
		t.Switcher{Active: s.active.Get(), Children: map[string]t.Widget{
			"files":    t.Text{Content: "main.go\napp.go"},
			"activity": t.Text{Content: "All changes saved"},
		}},
	}}
}

// --8<-- [end:switcher]

// --8<-- [start:tooltip]
func Tooltip() t.Widget {
	return t.Column{Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
		t.Tooltip{
			Content: "Save your changes", Position: t.TooltipBottom,
			Child: t.Button{ID: "save", Label: "Save"},
		},
	}}
}

// --8<-- [end:tooltip]

// --8<-- [start:empty]
func Empty() t.Widget {
	return t.Column{Children: []t.Widget{
		t.Text{Content: "Build complete"},
		t.EmptyWidget{},
		t.ShowWhen(false, t.Text{Content: "Errors found"}),
		t.Text{Content: "0 errors"},
	}}
}

// --8<-- [end:empty]

func Examples() []demo.Example {
	return []demo.Example{
		{Name: "row-column", Widget: RowColumn, Width: 38, Height: 5},
		{Name: "dock", Widget: Dock, Width: 44, Height: 7},
		{Name: "scrollable", Widget: Scrollable, Width: 36, Height: 6},
		{Name: "spacer", Widget: Spacer, Width: 38, Height: 1},
		{Name: "splitpane", Widget: SplitPane, Width: 44, Height: 6},
		{Name: "stack", Widget: Stack, Width: 34, Height: 5},
		{Name: "floating", Widget: Floating, Width: 44, Height: 8},
		{Name: "floating-geometry", Widget: FloatingGeometry, Width: 30, Height: 4},
		{Name: "focustrap", Widget: FocusTrap, Width: 36, Height: 7},
		{Name: "switcher", Widget: Switcher, Width: 34, Height: 7},
		{Name: "tooltip", Widget: Tooltip, Width: 28, Height: 7},
		{Name: "empty", Widget: Empty, Width: 24, Height: 2},
	}
}
