package main

import (
	"fmt"
	"path/filepath"
	"strings"

	t "terma"
)

const (
	diffFilesTreeID      = "terma-diff-files-tree"
	diffFilesScrollID    = "terma-diff-files-scroll"
	diffViewerID         = "terma-diff-viewer"
	diffViewerScrollID   = "terma-diff-viewer-scroll"
	diffSplitPaneID      = "terma-diff-split"
	diffCommandPaletteID = "terma-diff-command-palette"
	diffThemesPalette    = "Themes"
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
	commandPalette  *t.CommandPaletteState

	dividerFocused        bool
	dividerFocusRequested bool
	lastNonDividerFocus   string
	focusReturnID         string
}

func NewDiffApp(provider DiffProvider, staged bool) *DiffApp {
	app := &DiffApp{
		provider:            provider,
		staged:              staged,
		renderedByPath:      map[string]*RenderedFile{},
		filePathToTreePath:  map[string][]int{},
		orderedFilePaths:    []string{},
		treeState:           t.NewTreeState([]t.TreeNode[DiffTreeNodeData]{}),
		treeScrollState:     t.NewScrollState(),
		diffScrollState:     t.NewScrollState(),
		diffViewState:       NewDiffViewState(buildMetaRenderedFile("Diff", []string{"Loading diff..."})),
		splitState:          t.NewSplitPaneState(0.30),
		lastNonDividerFocus: diffViewerScrollID,
		focusReturnID:       diffViewerScrollID,
	}
	app.commandPalette = app.newCommandPalette()
	app.refreshDiff()
	t.RequestFocus(diffViewerScrollID)
	return app
}

func (a *DiffApp) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "n", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "]", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "p", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "[", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "r", Name: "Refresh", Action: a.refreshDiff, Hidden: true},
		{Key: "s", Name: "Toggle staged", Action: a.toggleMode, Hidden: true},
		{Key: "d", Name: "Focus divider", Action: a.focusDivider, Hidden: true},
		{Key: "ctrl+p", Name: "Command palette", Action: a.togglePalette},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *DiffApp) Build(ctx t.BuildContext) t.Widget {
	a.syncFocusState(ctx)
	theme := ctx.Theme()
	splitPane := FocusAwareSplitPane{
		SplitPane: t.SplitPane{
			ID:                     diffSplitPaneID,
			State:                  a.splitState,
			Orientation:            t.SplitHorizontal,
			DividerSize:            1,
			MinPaneSize:            20,
			DividerBackground:      theme.Background,
			DividerForeground:      dividerForeground(theme),
			DividerFocusForeground: dividerFocusForeground(theme),
			OnExitFocus:            a.exitDividerFocus,
			Style: t.Style{
				Width:           t.Flex(1),
				Height:          t.Flex(1),
				BackgroundColor: theme.Background,
			},
			First:  a.buildLeftPane(ctx, theme),
			Second: a.buildRightPane(theme),
		},
		AllowFocus:     a.dividerFocused || a.dividerFocusRequested,
		EnableKeybinds: a.dividerFocused,
	}

	return t.Stack{
		Style: t.Style{
			Width:           t.Flex(1),
			Height:          t.Flex(1),
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			t.Dock{
				Style: t.Style{
					BackgroundColor: theme.Background,
				},
				Top: []t.Widget{a.buildHeader(theme)},
				Bottom: []t.Widget{
					t.KeybindBar{
						Style: t.Style{
							BackgroundColor: theme.Surface,
							Padding:         t.EdgeInsetsXY(1, 0),
						},
					},
				},
				Body: splitPane,
			},
			t.CommandPalette{
				ID:             diffCommandPaletteID,
				State:          a.commandPalette,
				Position:       t.FloatPositionTopCenter,
				Offset:         t.Offset{Y: 1},
				BackdropColor:  t.Black.WithAlpha(0.05),
				OnCursorChange: a.handlePaletteCursorChange,
			},
		},
	}
}

func (a *DiffApp) buildHeader(theme t.ThemeData) t.Widget {
	repoName := "(unknown repo)"
	if a.repoRoot != "" {
		repoName = filepath.Base(a.repoRoot)
	}

	status := fmt.Sprintf("Repo: %s  Mode: %s  Theme: %s", repoName, a.modeLabel(), t.CurrentThemeName())
	statusStyle := t.Style{ForegroundColor: theme.TextMuted}
	if a.loadErr != "" {
		status = "Error loading diff. See viewer panel for details."
		statusStyle.ForegroundColor = theme.Error
	}

	return t.Text{
		Content: status,
		Style: t.Style{
			Padding:         t.EdgeInsetsXY(1, 0),
			BackgroundColor: theme.Surface,
			ForegroundColor: statusStyle.ForegroundColor,
		},
	}
}

func (a *DiffApp) buildLeftPane(ctx t.BuildContext, theme t.ThemeData) t.Widget {
	treeWidget := SplitFriendlyTree{
		Tree: t.Tree[DiffTreeNodeData]{
			ID:          diffFilesTreeID,
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
			t.Text{
				Content: fmt.Sprintf("Files: %d", len(a.orderedFilePaths)),
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: theme.Background,
					ForegroundColor: theme.TextMuted,
					Bold:            true,
				},
			},
			t.Scrollable{
				ID:    diffFilesScrollID,
				State: a.treeScrollState,
				Style: t.Style{
					Width:           t.Flex(1),
					Height:          t.Flex(1),
					BackgroundColor: theme.Background,
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

		if nodeCtx.Active {
			if widgetFocused {
				rowStyle.BackgroundColor = theme.ActiveCursor
				labelStyle.ForegroundColor = theme.SelectionText
				addStyle.ForegroundColor = theme.SelectionText
				delStyle.ForegroundColor = theme.SelectionText
				metaStyle.ForegroundColor = theme.SelectionText
			} else {
				rowStyle.BackgroundColor = unfocusedTreeCursorColor(theme)
			}
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
		ID:             diffViewerID,
		DisableFocus:   true,
		State:          a.diffViewState,
		VerticalScroll: a.diffScrollState,
		Palette:        NewThemePalette(theme),
		Style: t.Style{
			Width:           t.Flex(1),
			Padding:         t.EdgeInsets{},
			BackgroundColor: theme.Background,
		},
	}

	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.Text{
				Content: a.viewerTitle(),
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: theme.Background,
					ForegroundColor: theme.Text,
					Bold:            true,
				},
			},
			t.Scrollable{
				ID:        diffViewerScrollID,
				State:     a.diffScrollState,
				Focusable: true,
				Style: t.Style{
					Width:           t.Flex(1),
					Height:          t.Flex(1),
					Padding:         t.EdgeInsetsXY(1, 0),
					BackgroundColor: theme.Background,
				},
				Child: viewer,
			},
		},
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

func (a *DiffApp) toggleMode() {
	a.staged = !a.staged
	a.refreshDiff()
}

func (a *DiffApp) focusDivider() {
	target := a.dividerReturnTarget()
	a.dividerFocusRequested = true
	a.focusReturnID = target
	t.RequestFocus(diffSplitPaneID)
}

func (a *DiffApp) focusDividerFromPalette() {
	a.dividerFocusRequested = true
	a.focusReturnID = a.dividerReturnTarget()
	if a.commandPalette != nil {
		a.commandPalette.SetNextFocusIDOnClose(diffSplitPaneID)
		a.commandPalette.Close(false)
	}
}

func (a *DiffApp) exitDividerFocus() {
	a.dividerFocusRequested = false
	target := a.focusReturnID
	if target == "" || target == diffSplitPaneID {
		target = diffViewerScrollID
	}
	t.RequestFocus(target)
}

func (a *DiffApp) togglePalette() {
	if a.commandPalette == nil {
		return
	}
	if a.commandPalette.Visible.Peek() {
		a.commandPalette.Close(false)
		return
	}
	a.commandPalette.Open()
}

func (a *DiffApp) syncFocusState(ctx t.BuildContext) {
	wasDividerFocused := a.dividerFocused
	focusedID := focusedWidgetID(ctx)
	a.dividerFocused = focusedID == diffSplitPaneID
	if wasDividerFocused && !a.dividerFocused {
		a.dividerFocusRequested = false
	}
	if focusedID != "" && focusedID != diffSplitPaneID {
		a.lastNonDividerFocus = focusedID
	}
}

func (a *DiffApp) dividerReturnTarget() string {
	target := a.lastNonDividerFocus
	if target == "" || target == diffSplitPaneID {
		target = diffViewerScrollID
	}
	return target
}

func dividerFocusForeground(theme t.ThemeData) t.ColorProvider {
	return dividerGradient(theme, theme.Accent)
}

func dividerForeground(theme t.ThemeData) t.ColorProvider {
	return dividerGradient(theme, theme.Border)
}

func dividerGradient(theme t.ThemeData, center t.Color) t.ColorProvider {
	return t.NewGradient(theme.Background, center, theme.Background).WithAngle(0)
}

func unfocusedTreeCursorColor(theme t.ThemeData) t.Color {
	alpha := theme.ActiveCursor.Alpha()
	if alpha <= 0 {
		alpha = 1.0
	}
	alpha = alpha * 0.35
	if alpha < 0.12 {
		alpha = 0.12
	}
	if alpha > 0.35 {
		alpha = 0.35
	}
	return theme.ActiveCursor.WithAlpha(alpha)
}

func focusedWidgetID(ctx t.BuildContext) string {
	focused := ctx.Focused()
	if focused == nil {
		return ""
	}
	if identifiable, ok := focused.(t.Identifiable); ok {
		return identifiable.WidgetID()
	}
	return ""
}

func (a *DiffApp) newCommandPalette() *t.CommandPaletteState {
	return t.NewCommandPaletteState("Commands", []t.CommandPaletteItem{
		{
			Label:      "Toggle staged mode",
			FilterText: "Toggle staged mode staged unstaged",
			Hint:       "[s]",
			Action:     a.paletteAction(a.toggleMode),
		},
		{
			Label:      "Refresh",
			FilterText: "Refresh reload diff",
			Hint:       "[r]",
			Action:     a.paletteAction(a.refreshDiff),
		},
		{
			Label:      "Focus divider",
			FilterText: "Focus divider split resize",
			Hint:       "[d]",
			Action:     a.focusDividerFromPalette,
		},
		{Divider: "Appearance"},
		{
			Label:         "Theme",
			ChildrenTitle: diffThemesPalette,
			Children:      a.themeItems,
		},
	})
}

func (a *DiffApp) themeItems() []t.CommandPaletteItem {
	items := make([]t.CommandPaletteItem, 0, len(t.ThemeNames())+2)
	addGroup := func(title string, names []string) {
		if len(names) == 0 {
			return
		}
		items = append(items, t.CommandPaletteItem{Divider: title})
		for _, name := range names {
			label := themeDisplayName(name)
			hint := ""
			if name == t.CurrentThemeName() {
				hint = "current"
			}
			themeName := name
			items = append(items, t.CommandPaletteItem{
				Label:      label,
				FilterText: label + " " + themeName,
				Hint:       hint,
				Data:       themeName,
				Action:     a.setThemeAction(themeName),
			})
		}
	}

	addGroup("Dark themes", t.DarkThemeNames())
	addGroup("Light themes", t.LightThemeNames())

	return items
}

func (a *DiffApp) setThemeAction(themeName string) func() {
	return func() {
		t.SetTheme(themeName)
		if a.commandPalette != nil {
			a.commandPalette.Close(false)
		}
	}
}

func (a *DiffApp) paletteAction(action func()) func() {
	return func() {
		if action != nil {
			action()
		}
		if a.commandPalette != nil {
			a.commandPalette.Close(false)
		}
	}
}

func (a *DiffApp) handlePaletteCursorChange(item t.CommandPaletteItem) {
	if a.commandPalette == nil {
		return
	}
	level := a.commandPalette.CurrentLevel()
	if level == nil || level.Title != diffThemesPalette {
		return
	}
	themeName, ok := item.Data.(string)
	if !ok || themeName == "" {
		return
	}
	t.SetTheme(themeName)
}

func themeDisplayName(name string) string {
	parts := strings.Split(name, "-")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
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
