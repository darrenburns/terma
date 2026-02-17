package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	t "github.com/darrenburns/terma"
)

type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
}

type TreeExampleApp struct {
	treeState    *t.TreeState[FileInfo]
	filterState  *t.FilterState
	filterInput  *t.TextInputState
	scrollState  *t.ScrollState
	lazyChildren map[string][]t.TreeNode[FileInfo]
	status       t.Signal[string]
}

func NewTreeExampleApp() *TreeExampleApp {
	roots, lazy := sampleTree()
	return &TreeExampleApp{
		treeState:    t.NewTreeState(roots),
		filterState:  t.NewFilterState(),
		filterInput:  t.NewTextInputState(""),
		scrollState:  t.NewScrollState(),
		lazyChildren: lazy,
		status:       t.NewSignal("Ready"),
	}
}

func (a *TreeExampleApp) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+f", Name: "Focus filter", Action: func() {
			t.RequestFocus("tree-filter")
		}},
		{Key: "esc", Name: "Clear filter", Action: a.clearFilter},
	}
}

func (a *TreeExampleApp) clearFilter() {
	a.filterInput.SetText("")
	a.filterState.Query.Set("")
	a.status.Set("Filter cleared")
	t.RequestFocus("file-tree")
}

func (a *TreeExampleApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "tree-example-root",
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Style: t.Style{BackgroundColor: theme.Background},
		Body: t.Column{
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsAll(1),
			},
			Children: []t.Widget{
				t.Text{
					Content: " Tree Example ",
					Style: t.Style{
						ForegroundColor: theme.TextOnPrimary,
						BackgroundColor: theme.Primary,
						Padding:         t.EdgeInsetsXY(2, 0),
					},
				},
				t.ParseMarkupToText("Use [b $Accent]Up/Down[/] or [b $Accent]j/k[/] to navigate | [b $Accent]Left/Right[/] or [b $Accent]h/l[/] to collapse/expand | [b $Accent]Shift+Up/Down[/] to extend selection | [b $Accent]Space[/] to toggle | [b $Accent]Enter[/] to select", theme),
				t.Row{
					Spacing:    1,
					CrossAlign: t.CrossAxisCenter,
					Children: []t.Widget{
						t.TextInput{
							ID:          "tree-filter",
							State:       a.filterInput,
							Placeholder: "Type to filter files...",
							Width:       t.Flex(1),
							Style: t.Style{
								BackgroundColor: theme.Surface,
								ForegroundColor: theme.Text,
							},
							OnChange: func(text string) {
								a.filterState.Query.Set(text)
							},
						},
						t.Text{
							Content: "Ctrl+F",
							Style: t.Style{
								ForegroundColor: theme.TextMuted,
							},
						},
					},
				},
				t.Scrollable{
					ID:    "tree-scroll",
					State: a.scrollState,
					ScrollbarThumbColor: theme.ScrollbarThumb,
					ScrollbarTrackColor: theme.ScrollbarTrack,
					Style: t.Style{
						Height:   t.Flex(1),
						Width:    t.Auto,
						MinWidth: t.Cells(14),
						Border:   t.SquareBorder(t.NewGradient(theme.Primary, theme.Background), t.BorderTitle("Project Files")),
						Padding:  t.EdgeInsetsAll(1),
					},
					Child: func() t.Widget {
						treeWidget := t.Tree[FileInfo]{
							ID:          "file-tree",
							State:       a.treeState,
							Filter:      a.filterState,
							ScrollState: a.scrollState,
							HasChildren: func(info FileInfo) bool {
								return info.IsDir
							},
							MultiSelect: true,
							OnExpand:    a.handleExpand,
							MatchNode: func(info FileInfo, query string, options t.FilterOptions) t.MatchResult {
								return t.MatchString(info.Name, query, options)
							},
							OnSelect: func(info FileInfo, selected []FileInfo) {
								a.updateSelectStatus(info, selected)
							},
							OnCursorChange: func(info FileInfo) {
								a.updateCursorStatus(info)
							},
						}
						widgetFocused := ctx.IsFocused(treeWidget)
						treeWidget.RenderNodeWithMatch = func(info FileInfo, nodeCtx t.TreeNodeContext, match t.MatchResult) t.Widget {
							style := treeNodeStyle(theme, nodeCtx, widgetFocused)
							icon := "F"
							if info.IsDir {
								icon = "D"
							}
							spans := []t.Span{{Text: icon + " "}}
							if match.Matched && len(match.Ranges) > 0 {
								highlight := t.SpanStyle{
									Underline:      t.UnderlineSingle,
									UnderlineColor: theme.Accent,
									Background:     theme.Selection,
								}
								spans = append(spans, t.HighlightSpans(info.Name, match.Ranges, highlight)...)
							} else {
								spans = append(spans, t.Span{Text: info.Name})
							}
							return t.Text{
								Spans: spans,
								Style: style,
							}
						}
						return treeWidget
					}(),
				},
				t.Text{
					Content: a.status.Get(),
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},
				t.Text{
					Content: a.selectionSummaryText(),
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},
			},
		},
	}
}

func (a *TreeExampleApp) handleExpand(info FileInfo, path []int, setChildren func([]t.TreeNode[FileInfo])) {
	children, ok := a.lazyChildren[info.Path]
	if !ok {
		setChildren([]t.TreeNode[FileInfo]{})
		return
	}
	a.status.Set(fmt.Sprintf("Loading %s...", info.Name))
	go func() {
		time.Sleep(200 * time.Millisecond)
		setChildren(children)
		a.status.Set(fmt.Sprintf("Loaded %s", info.Name))
	}()
}

func treeNodeStyle(theme t.ThemeData, nodeCtx t.TreeNodeContext, widgetFocused bool) t.Style {
	style := t.Style{ForegroundColor: theme.Text}
	if nodeCtx.FilteredAncestor {
		style.ForegroundColor = theme.TextMuted
	}
	showCursor := nodeCtx.Active && widgetFocused
	if showCursor {
		style.BackgroundColor = theme.ActiveCursor
		style.ForegroundColor = theme.SelectionText
		return style
	}
	if nodeCtx.Selected {
		style.BackgroundColor = theme.Selection
		style.ForegroundColor = theme.SelectionText
	}
	return style
}

func (a *TreeExampleApp) updateSelectStatus(info FileInfo, selected []FileInfo) {
	if len(selected) > 0 {
		paths := make([]string, len(selected))
		for i, s := range selected {
			paths[i] = s.Path
		}
		a.status.Set(fmt.Sprintf("Selected: %s", strings.Join(paths, ", ")))
		return
	}
	a.status.Set(fmt.Sprintf("Selected: %s", info.Path))
}

func (a *TreeExampleApp) updateCursorStatus(info FileInfo) {
	if summary, ok := a.selectionSummary(); ok {
		a.status.Set(fmt.Sprintf("Cursor: %s | %s", info.Path, summary))
		return
	}
	a.status.Set(fmt.Sprintf("Cursor: %s", info.Path))
}

func (a *TreeExampleApp) selectionSummaryText() string {
	if summary, ok := a.selectionSummary(); ok {
		return summary
	}
	return "Selected: (none)"
}

func (a *TreeExampleApp) selectionSummary() (string, bool) {
	if a.treeState == nil || !a.treeState.Selection.IsValid() {
		return "", false
	}
	a.treeState.Selection.Get()
	paths := a.treeState.SelectedPaths()
	if len(paths) == 0 {
		return "", false
	}
	labels := make([]string, 0, len(paths))
	for _, path := range paths {
		if node, ok := a.treeState.NodeAtPath(path); ok {
			labels = append(labels, node.Data.Path)
		} else {
			labels = append(labels, fmt.Sprintf("%v", path))
		}
	}
	if len(labels) == 0 {
		return "", false
	}
	return fmt.Sprintf("Selected: %s", strings.Join(labels, ", ")), true
}

func sampleTree() ([]t.TreeNode[FileInfo], map[string][]t.TreeNode[FileInfo]) {
	roots := []t.TreeNode[FileInfo]{
		{
			Data:     FileInfo{Name: "cmd", Path: "/cmd", IsDir: true},
			Children: nil, // lazy
		},
		{
			Data: FileInfo{Name: "docs", Path: "/docs", IsDir: true},
			Children: []t.TreeNode[FileInfo]{
				{Data: FileInfo{Name: "getting-started.md", Path: "/docs/getting-started.md", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
				{Data: FileInfo{Name: "themes.md", Path: "/docs/themes.md", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
			},
		},
		{
			Data:     FileInfo{Name: "internal", Path: "/internal", IsDir: true},
			Children: nil, // lazy
		},
		{
			Data:     FileInfo{Name: "README.md", Path: "/README.md", IsDir: false},
			Children: []t.TreeNode[FileInfo]{},
		},
	}

	lazy := map[string][]t.TreeNode[FileInfo]{
		"/cmd": {
			{
				Data: FileInfo{Name: "tree-example", Path: "/cmd/tree-example", IsDir: true},
				Children: []t.TreeNode[FileInfo]{
					{Data: FileInfo{Name: "main.go", Path: "/cmd/tree-example/main.go", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
				},
			},
			{Data: FileInfo{Name: "list-example", Path: "/cmd/list-example", IsDir: true}, Children: nil},
		},
		"/cmd/list-example": {
			{Data: FileInfo{Name: "main.go", Path: "/cmd/list-example/main.go", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
		},
		"/internal": {
			{Data: FileInfo{Name: "ui", Path: "/internal/ui", IsDir: true}, Children: nil},
			{Data: FileInfo{Name: "signals.go", Path: "/internal/signals.go", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
		},
		"/internal/ui": {
			{Data: FileInfo{Name: "tree.go", Path: "/internal/ui/tree.go", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
			{Data: FileInfo{Name: "list.go", Path: "/internal/ui/list.go", IsDir: false}, Children: []t.TreeNode[FileInfo]{}},
		},
	}

	return roots, lazy
}

func main() {
	app := NewTreeExampleApp()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
