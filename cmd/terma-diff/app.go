package main

import (
	"fmt"
	"path/filepath"
	"strings"

	t "terma"
)

// DiffApp is a read-only, syntax-highlighted git diff viewer.
type DiffApp struct {
	provider DiffProvider
	staged   bool

	repoRoot string
	loadErr  string
	files    []*DiffFile

	activePath  string
	activeIsDir bool

	renderedByPath     map[string]*RenderedFile
	filePathToTreePath map[string][]int
	orderedFilePaths   []string

	treeState       *t.TreeState[DiffTreeNodeData]
	treeScrollState *t.ScrollState
	diffScrollState *t.ScrollState
	diffViewState   *DiffViewState
	splitState      *t.SplitPaneState
	themeNames      []string
}

func NewDiffApp(provider DiffProvider, staged bool) *DiffApp {
	app := &DiffApp{
		provider:           provider,
		staged:             staged,
		renderedByPath:     map[string]*RenderedFile{},
		filePathToTreePath: map[string][]int{},
		orderedFilePaths:   []string{},
		treeState:          t.NewTreeState([]t.TreeNode[DiffTreeNodeData]{}),
		treeScrollState:    t.NewScrollState(),
		diffScrollState:    t.NewScrollState(),
		diffViewState:      NewDiffViewState(buildMetaRenderedFile("Diff", []string{"Loading diff..."})),
		splitState:         t.NewSplitPaneState(0.30),
		themeNames:         t.ThemeNames(),
	}
	app.refreshDiff()
	t.RequestFocus("terma-diff-viewer-scroll")
	return app
}

func (a *DiffApp) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "n", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "]", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "p", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "[", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "r", Name: "Refresh", Action: a.refreshDiff},
		{Key: "t", Name: "Theme", Action: a.cycleTheme},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *DiffApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Dock{
		Top: []t.Widget{a.buildHeader(theme)},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Body: t.SplitPane{
			ID:                "terma-diff-split",
			State:             a.splitState,
			Orientation:       t.SplitHorizontal,
			DividerSize:       1,
			MinPaneSize:       20,
			DisableFocus:      true,
			DividerBackground: theme.Surface2,
			First:             a.buildLeftPane(ctx, theme),
			Second:            a.buildRightPane(theme),
		},
	}
}

func (a *DiffApp) buildHeader(theme t.ThemeData) t.Widget {
	repoName := "(unknown repo)"
	if a.repoRoot != "" {
		repoName = filepath.Base(a.repoRoot)
	}

	status := fmt.Sprintf("Repo: %s  Mode: %s  Files: %d  Theme: %s", repoName, a.modeLabel(), len(a.orderedFilePaths), t.CurrentThemeName())
	statusStyle := t.Style{ForegroundColor: theme.TextMuted}
	if a.loadErr != "" {
		status = "Error loading diff. See viewer panel for details."
		statusStyle.ForegroundColor = theme.Error
	}

	return t.Column{
		Spacing: 0,
		Children: []t.Widget{
			t.Text{
				Content: " terma-diff ",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
					Bold:            true,
				},
			},
			t.Text{
				Content: status,
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: theme.Surface,
					ForegroundColor: statusStyle.ForegroundColor,
				},
			},
		},
	}
}

func (a *DiffApp) buildLeftPane(ctx t.BuildContext, theme t.ThemeData) t.Widget {
	treeWidget := SplitFriendlyTree{
		Tree: t.Tree[DiffTreeNodeData]{
			ID:          "terma-diff-files-tree",
			State:       a.treeState,
			ScrollState: a.treeScrollState,
			Style:       t.Style{Width: t.Flex(1)},
			NodeID: func(node DiffTreeNodeData) string {
				return node.Path
			},
			HasChildren: func(node DiffTreeNodeData) bool {
				return node.IsDir
			},
			OnCursorChange: a.onTreeCursorChange,
		},
	}

	sidebarFocused := ctx.IsFocused(treeWidget)
	treeWidget.RenderNodeWithMatch = a.renderTreeNode(theme, sidebarFocused)

	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.Scrollable{
				ID:    "terma-diff-files-scroll",
				State: a.treeScrollState,
				Style: t.Style{
					Width:  t.Flex(1),
					Height: t.Flex(1),
					Border: t.RoundedBorder(theme.Border, t.BorderTitle("Files")),
				},
				Child: treeWidget,
			},
		},
	}
}

func (a *DiffApp) renderTreeNode(theme t.ThemeData, widgetFocused bool) func(node DiffTreeNodeData, nodeCtx t.TreeNodeContext, match t.MatchResult) t.Widget {
	return func(node DiffTreeNodeData, nodeCtx t.TreeNodeContext, _ t.MatchResult) t.Widget {
		rowStyle := t.Style{
			Width:   t.Flex(1),
			Padding: t.EdgeInsetsXY(1, 0),
		}
		labelStyle := t.Style{ForegroundColor: theme.Text}
		addStyle := t.Style{ForegroundColor: theme.Success}
		delStyle := t.Style{ForegroundColor: theme.Error}
		metaStyle := t.Style{ForegroundColor: theme.TextMuted}

		showCursor := nodeCtx.Active && widgetFocused
		if showCursor {
			rowStyle.BackgroundColor = theme.ActiveCursor
			labelStyle.ForegroundColor = theme.SelectionText
			addStyle.ForegroundColor = theme.SelectionText
			delStyle.ForegroundColor = theme.SelectionText
			metaStyle.ForegroundColor = theme.SelectionText
		}

		label := node.Name
		meta := ""
		if node.IsDir {
			label += "/"
			meta = fmt.Sprintf("%d files", node.TouchedFiles)
		}

		children := []t.Widget{
			t.Text{Content: label, Style: labelStyle},
		}
		if meta != "" {
			children = append(children, t.Text{Content: " " + meta, Style: metaStyle})
		}
		children = append(children,
			t.Spacer{Width: t.Flex(1)},
			t.Text{Content: fmt.Sprintf("+%d", node.Additions), Style: addStyle},
			t.Text{Content: " "},
			t.Text{Content: fmt.Sprintf("-%d", node.Deletions), Style: delStyle},
		)

		return t.Row{
			Style:    rowStyle,
			Children: children,
		}
	}
}

func (a *DiffApp) buildRightPane(theme t.ThemeData) t.Widget {
	viewer := DiffView{
		ID:             "terma-diff-viewer",
		DisableFocus:   true,
		State:          a.diffViewState,
		VerticalScroll: a.diffScrollState,
		Palette:        NewThemePalette(theme),
		Style: t.Style{
			Width:           t.Flex(1),
			Padding:         t.EdgeInsets{},
			BackgroundColor: theme.Surface,
		},
	}

	return t.Scrollable{
		ID:        "terma-diff-viewer-scroll",
		State:     a.diffScrollState,
		Focusable: true,
		Style: t.Style{
			Width:           t.Flex(1),
			Height:          t.Flex(1),
			Padding:         t.EdgeInsetsXY(1, 0),
			BackgroundColor: theme.Surface,
			Border:          t.RoundedBorder(theme.Border, t.BorderTitle(a.viewerTitle())),
		},
		Child: viewer,
	}
}

func (a *DiffApp) refreshDiff() {
	if repoRoot, err := a.provider.RepoRoot(); err == nil {
		a.repoRoot = repoRoot
	}

	previousActiveFile := ""
	if !a.activeIsDir {
		previousActiveFile = a.activePath
	}

	raw, err := a.provider.LoadDiff(a.staged)
	if err != nil {
		a.loadErr = err.Error()
		a.files = nil
		a.activePath = ""
		a.activeIsDir = false
		a.renderedByPath = map[string]*RenderedFile{}
		a.filePathToTreePath = map[string][]int{}
		a.orderedFilePaths = nil
		a.treeState.Nodes.Set([]t.TreeNode[DiffTreeNodeData]{})
		a.treeState.CursorPath.Set(nil)
		a.treeState.Collapsed.Set(map[string]bool{})
		a.diffViewState.SetRendered(messageToRendered("Error", a.errorMessage()))
		a.diffScrollState.SetOffset(0)
		return
	}

	doc, err := parseUnifiedDiff(raw)
	if err != nil {
		a.loadErr = fmt.Sprintf("parse error: %v", err)
		a.files = nil
		a.activePath = ""
		a.activeIsDir = false
		a.renderedByPath = map[string]*RenderedFile{}
		a.filePathToTreePath = map[string][]int{}
		a.orderedFilePaths = nil
		a.treeState.Nodes.Set([]t.TreeNode[DiffTreeNodeData]{})
		a.treeState.CursorPath.Set(nil)
		a.treeState.Collapsed.Set(map[string]bool{})
		a.diffViewState.SetRendered(messageToRendered("Error", a.errorMessage()))
		a.diffScrollState.SetOffset(0)
		return
	}

	a.loadErr = ""
	a.files = doc.Files
	a.renderedByPath = make(map[string]*RenderedFile, len(a.files))
	for _, file := range a.files {
		if file == nil {
			continue
		}
		a.renderedByPath[file.DisplayPath] = buildRenderedFile(file)
	}

	roots, filePathToTreePath, orderedFilePaths := buildDiffTree(a.files)
	a.filePathToTreePath = filePathToTreePath
	a.orderedFilePaths = orderedFilePaths
	a.treeState.Nodes.Set(roots)
	a.treeState.Collapsed.Set(map[string]bool{})

	if len(a.orderedFilePaths) == 0 {
		a.activePath = ""
		a.activeIsDir = false
		a.treeState.CursorPath.Set(nil)
		a.diffViewState.SetRendered(messageToRendered("Diff", a.emptyMessage()))
		a.diffScrollState.SetOffset(0)
		return
	}

	targetPath := previousActiveFile
	if _, ok := a.filePathToTreePath[targetPath]; !ok {
		targetPath = a.orderedFilePaths[0]
	}
	if !a.selectFilePath(targetPath) {
		a.selectFilePath(a.orderedFilePaths[0])
	}
}

func (a *DiffApp) moveFileCursor(delta int) {
	if len(a.orderedFilePaths) == 0 {
		return
	}

	currentIdx := -1
	if !a.activeIsDir {
		currentIdx = a.indexOfOrderedPath(a.activePath)
	}

	nextIdx := 0
	if currentIdx < 0 {
		if delta < 0 {
			nextIdx = len(a.orderedFilePaths) - 1
		}
	} else {
		nextIdx = currentIdx + delta
		for nextIdx < 0 {
			nextIdx += len(a.orderedFilePaths)
		}
		nextIdx = nextIdx % len(a.orderedFilePaths)
	}

	a.selectFilePath(a.orderedFilePaths[nextIdx])
}

func (a *DiffApp) selectFilePath(filePath string) bool {
	treePath, ok := a.filePathToTreePath[filePath]
	if !ok {
		return false
	}
	a.treeState.CursorPath.Set(clonePath(treePath))
	node, ok := a.treeState.NodeAtPath(treePath)
	if !ok {
		return false
	}
	a.onTreeCursorChange(node.Data)
	return true
}

func (a *DiffApp) onTreeCursorChange(node DiffTreeNodeData) {
	if node.IsDir {
		a.setActiveDirectory(node)
		return
	}
	if node.File != nil {
		a.setActiveFile(node.File)
		return
	}
	if rendered, ok := a.renderedByPath[node.Path]; ok {
		a.activePath = node.Path
		a.activeIsDir = false
		a.diffViewState.SetRendered(rendered)
		a.diffScrollState.SetOffset(0)
	}
}

func (a *DiffApp) setActiveFile(file *DiffFile) {
	if file == nil {
		return
	}
	a.activePath = file.DisplayPath
	a.activeIsDir = false
	rendered, ok := a.renderedByPath[file.DisplayPath]
	if !ok {
		rendered = buildRenderedFile(file)
		a.renderedByPath[file.DisplayPath] = rendered
	}
	a.diffViewState.SetRendered(rendered)
	a.diffScrollState.SetOffset(0)
}

func (a *DiffApp) setActiveDirectory(node DiffTreeNodeData) {
	a.activePath = node.Path
	a.activeIsDir = true
	a.diffViewState.SetRendered(buildDirectorySummaryRenderedFile(node))
	a.diffScrollState.SetOffset(0)
}

func (a *DiffApp) cycleTheme() {
	if len(a.themeNames) == 0 {
		return
	}
	current := t.CurrentThemeName()
	idx := 0
	for i, name := range a.themeNames {
		if name == current {
			idx = i
			break
		}
	}
	next := (idx + 1) % len(a.themeNames)
	t.SetTheme(a.themeNames[next])
}

func (a *DiffApp) modeLabel() string {
	if a.staged {
		return "staged"
	}
	return "unstaged"
}

func (a *DiffApp) viewerTitle() string {
	if a.activePath == "" {
		if a.loadErr != "" {
			return "Error"
		}
		return "Diff"
	}
	if a.activeIsDir {
		return a.activePath + " (directory)"
	}
	return a.activePath
}

func (a *DiffApp) emptyMessage() string {
	if a.staged {
		return "No staged changes.\n\nRun git add <file> and press r to refresh."
	}
	return "No unstaged changes.\n\nMake edits in this repo and press r to refresh."
}

func (a *DiffApp) errorMessage() string {
	msg := strings.TrimSpace(a.loadErr)
	if msg == "" {
		msg = "Unknown error"
	}
	return "Failed to load git diff:\n\n" + msg + "\n\nPress r to retry."
}

func (a *DiffApp) indexOfOrderedPath(path string) int {
	if path == "" {
		return -1
	}
	for idx, filePath := range a.orderedFilePaths {
		if filePath == path {
			return idx
		}
	}
	return -1
}

func messageToRendered(title string, text string) *RenderedFile {
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	return buildMetaRenderedFile(title, strings.Split(normalized, "\n"))
}

func buildDirectorySummaryRenderedFile(node DiffTreeNodeData) *RenderedFile {
	path := node.Path
	if path == "" {
		path = node.Name
	}
	if path == "" {
		path = "(root)"
	}
	return buildMetaRenderedFile(path, []string{
		fmt.Sprintf("Directory: %s", path),
		fmt.Sprintf("Touched files: %d", node.TouchedFiles),
		fmt.Sprintf("Additions: +%d", node.Additions),
		fmt.Sprintf("Deletions: -%d", node.Deletions),
		"",
		"Use n/p to jump between changed files.",
	})
}
