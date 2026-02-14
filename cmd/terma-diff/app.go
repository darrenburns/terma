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
	diffFilesFilterID    = "terma-diff-files-filter"
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
	branch   string
	loadErr  string
	files    []*DiffFile

	activePath  string
	activeIsDir bool

	renderedByPath     map[string]*RenderedFile
	fileByPath         map[string]*DiffFile
	filePathToTreePath map[string][]int
	orderedFilePaths   []string

	treeState       *t.TreeState[DiffTreeNodeData]
	treeScrollState *t.ScrollState
	treeFilterState *t.FilterState
	treeFilterInput *t.TextInputState
	diffScrollState *t.ScrollState
	diffViewState   *DiffViewState
	splitState      *t.SplitPaneState
	commandPalette  *t.CommandPaletteState

	treeFilterVisible   bool
	treeFilterNoMatches bool
	diffHardWrap        bool
	focusedWidgetID     string
	sidebarVisible      bool

	dividerFocused        bool
	dividerFocusRequested bool
	lastNonDividerFocus   string
	focusReturnID         string
	themeCursorSynced     bool
	themePreviewBase      string
}

func NewDiffApp(provider DiffProvider, staged bool) *DiffApp {
	app := &DiffApp{
		provider:            provider,
		staged:              staged,
		renderedByPath:      map[string]*RenderedFile{},
		fileByPath:          map[string]*DiffFile{},
		filePathToTreePath:  map[string][]int{},
		orderedFilePaths:    []string{},
		treeState:           t.NewTreeState([]t.TreeNode[DiffTreeNodeData]{}),
		treeScrollState:     t.NewScrollState(),
		treeFilterState:     t.NewFilterState(),
		treeFilterInput:     t.NewTextInputState(""),
		diffScrollState:     t.NewScrollState(),
		diffViewState:       NewDiffViewState(buildMetaRenderedFile("Diff", []string{"Loading diff..."})),
		splitState:          t.NewSplitPaneState(0.30),
		sidebarVisible:      true,
		lastNonDividerFocus: diffViewerScrollID,
		focusReturnID:       diffViewerScrollID,
	}
	app.commandPalette = app.newCommandPalette()
	app.refreshDiff()
	t.RequestFocus(diffViewerScrollID)
	return app
}

func (a *DiffApp) Keybinds() []t.Keybind {
	showFilterFiles := a.focusedWidgetID == diffFilesTreeID
	return []t.Keybind{
		{Key: "n", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "]", Name: "Next file", Action: func() { a.moveFileCursor(1) }},
		{Key: "p", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "[", Name: "Prev file", Action: func() { a.moveFileCursor(-1) }},
		{Key: "/", Name: "Filter files", Action: a.openTreeFilter, Hidden: !showFilterFiles},
		{Key: "ctrl+b", Name: "Toggle sidebar", Action: a.toggleSidebar, Hidden: true},
		{Key: "escape", Name: "Clear filter", Action: a.handleEscape, Hidden: true},
		{Key: "r", Name: "Refresh", Action: a.refreshDiff, Hidden: true},
		{Key: "s", Name: "Toggle staged", Action: a.toggleMode, Hidden: true},
		{Key: "w", Name: "Toggle line wrap", Action: a.toggleDiffWrap, Hidden: true},
		{Key: "d", Name: "Focus divider", Action: a.focusDivider, Hidden: true},
		{Key: "ctrl+p", Name: "Command palette", Action: a.togglePalette},
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *DiffApp) Build(ctx t.BuildContext) t.Widget {
	a.syncFocusState(ctx)
	theme := ctx.Theme()
	body := a.buildRightPane(theme)
	if a.sidebarVisible {
		body = FocusAwareSplitPane{
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
					t.Row{
						Style: t.Style{
							Width:           t.Flex(1),
							BackgroundColor: theme.Background,
						},
						Children: []t.Widget{
							t.Spacer{Width: t.Flex(1)},
							t.KeybindBar{
								Style: t.Style{
									Width:           t.Auto,
									BackgroundColor: theme.Background,
									Padding:         t.EdgeInsetsXY(1, 0),
								},
							},
							t.Spacer{Width: t.Flex(1)},
						},
					},
				},
				Body: body,
			},
			t.CommandPalette{
				ID:             diffCommandPaletteID,
				State:          a.commandPalette,
				Position:       t.FloatPositionTopCenter,
				Offset:         t.Offset{Y: 1},
				BackdropColor:  t.Black.WithAlpha(0.05),
				OnCursorChange: a.handlePaletteCursorChange,
				OnDismiss:      a.handlePaletteDismiss,
			},
		},
	}
}

func (a *DiffApp) buildHeader(theme t.ThemeData) t.Widget {
	repoName := "(unknown repo)"
	if a.repoRoot != "" {
		repoName = filepath.Base(a.repoRoot)
	}

	rightWidget := t.Text{
		Content: themeDisplayName(t.CurrentThemeName()),
		Style: t.Style{
			Padding:         t.EdgeInsetsXY(1, 0),
			ForegroundColor: theme.SecondaryText,
		},
	}
	if a.loadErr != "" {
		rightWidget = t.Label("Error loading diff", t.LabelError, theme)
	}

	children := []t.Widget{
		t.Label(repoName, t.LabelPrimary, theme),
	}
	if a.branch != "" {
		children = append(children,
			t.Spacer{Width: t.Cells(1)},
			t.Text{
				Content: a.branch,
				Style: t.Style{
					ForegroundColor: theme.Accent,
				},
			},
		)
	}
	if a.loadErr != "" {
		children = append(children,
			t.Spacer{Width: t.Cells(1)},
			t.Label("Error", t.LabelError, theme),
		)
	}
	children = append(children,
		t.Spacer{Width: t.Flex(1)},
		rightWidget,
	)

	return t.Row{
		Style: t.Style{
			Width:   t.Flex(1),
			Padding: t.EdgeInsetsXY(1, 0),
			BackgroundColor: t.NewGradient(
				theme.Surface,
				theme.Surface,
				theme.Background,
				theme.Background,
				theme.Background,
				theme.SecondaryBg,
			).WithAngle(90),
		},
		Children: children,
	}
}

func (a *DiffApp) buildLeftPane(ctx t.BuildContext, theme t.ThemeData) t.Widget {
	treeWidget := SplitFriendlyTree{
		Tree: t.Tree[DiffTreeNodeData]{
			ID:                diffFilesTreeID,
			State:             a.treeState,
			Filter:            a.treeFilterState,
			ScrollState:       a.treeScrollState,
			Style:             t.Style{Width: t.Flex(1), Padding: t.EdgeInsets{Left: 1}},
			ExpandIndicator:   "▼ ",
			CollapseIndicator: "▶ ",
			LeafIndicator:     " ",
			NodeID: func(node DiffTreeNodeData) string {
				return node.Path
			},
			HasChildren: func(node DiffTreeNodeData) bool {
				return node.IsDir
			},
			MatchNode: func(node DiffTreeNodeData, query string, options t.FilterOptions) t.MatchResult {
				return t.MatchString(node.Name, query, options)
			},
			OnCursorChange: a.onTreeCursorChange,
		},
	}

	sidebarFocused := ctx.IsFocused(treeWidget)
	treeWidget.RenderNodeWithMatch = a.renderTreeNode(theme, sidebarFocused)

	children := []t.Widget{
		t.Row{
			Style: t.Style{
				Width:           t.Flex(1),
				Padding:         t.EdgeInsetsXY(1, 0),
				BackgroundColor: theme.Background,
			},
			Children: []t.Widget{
				t.Text{Spans: a.sidebarHeadingSpans(theme)},
				t.Spacer{Width: t.Flex(1)},
				t.Text{Spans: a.sidebarTotalsSpans(theme)},
			},
		},
	}

	if a.shouldShowTreeFilterInput() {
		children = append(children, t.TextInput{
			ID:          diffFilesFilterID,
			State:       a.treeFilterInput,
			Placeholder: "Filter files...",
			Width:       t.Flex(1),
			Style: t.Style{
				Padding:         t.EdgeInsetsXY(1, 0),
				BackgroundColor: theme.Background,
				ForegroundColor: theme.Text,
			},
			OnChange: a.onTreeFilterChange,
		})
	}

	treeContent := t.Widget(treeWidget)
	if a.treeFilterNoMatches {
		treeContent = a.buildTreeFilterEmptyState(theme)
	}

	children = append(children, t.Scrollable{
		ID:    diffFilesScrollID,
		State: a.treeScrollState,
		Style: t.Style{
			Width:           t.Flex(1),
			Height:          t.Flex(1),
			BackgroundColor: theme.Background,
		},
		Child: treeContent,
	})

	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Children: children,
	}
}

func (a *DiffApp) renderTreeNode(theme t.ThemeData, widgetFocused bool) func(node DiffTreeNodeData, nodeCtx t.TreeNodeContext, match t.MatchResult) t.Widget {
	highlightStyle := t.MatchHighlightStyle(theme)
	return func(node DiffTreeNodeData, nodeCtx t.TreeNodeContext, match t.MatchResult) t.Widget {
		rowStyle := t.Style{
			Width:   t.Flex(1),
			Padding: t.EdgeInsets{Right: 1},
		}
		labelStyle := t.Style{ForegroundColor: theme.Text}
		addStyle := t.Style{ForegroundColor: theme.Success}
		delStyle := t.Style{ForegroundColor: theme.Error}

		if nodeCtx.FilteredAncestor {
			labelStyle.ForegroundColor = theme.TextMuted
		}

		if nodeCtx.Active {
			if widgetFocused {
				rowStyle.BackgroundColor = theme.ActiveCursor
				labelStyle.ForegroundColor = theme.SelectionText
				addStyle.ForegroundColor = theme.SelectionText
				delStyle.ForegroundColor = theme.SelectionText
			} else {
				rowStyle.BackgroundColor = unfocusedTreeCursorColor(theme)
			}
		}

		label := node.Name
		if node.IsDir {
			label += "/"
		}

		labelWidget := t.Text{Content: label, Style: labelStyle}
		if match.Matched && len(match.Ranges) > 0 {
			spans := t.HighlightSpans(node.Name, match.Ranges, highlightStyle)
			if node.IsDir {
				spans = append(spans, t.Span{Text: "/"})
			}
			labelWidget = t.Text{
				Spans: spans,
				Style: labelStyle,
			}
		}

		children := []t.Widget{
			labelWidget,
		}
		children = append(children, t.Spacer{Width: t.Flex(1)})
		if addText, delText := nonZeroChangeTexts(node.Additions, node.Deletions); addText != "" || delText != "" {
			if addText != "" {
				children = append(children, t.Text{Content: addText, Style: addStyle})
			}
			if delText != "" {
				if addText != "" {
					children = append(children, t.Text{Content: " "})
				}
				children = append(children, t.Text{Content: delText, Style: delStyle})
			}
		}

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
		HardWrap:       a.diffHardWrap,
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
		},
		Children: []t.Widget{
			a.buildViewerTitle(theme),
			t.Scrollable{
				ID:        diffViewerScrollID,
				State:     a.diffScrollState,
				Focusable: true,
				Style: t.Style{
					Width:           t.Flex(1),
					Height:          t.Flex(1),
					BackgroundColor: theme.Background,
				},
				Child: viewer,
			},
		},
	}
}

func (a *DiffApp) buildViewerTitle(theme t.ThemeData) t.Widget {
	style := t.Style{
		Padding:         t.EdgeInsetsXY(1, 0),
		BackgroundColor: theme.Background,
		ForegroundColor: theme.Text,
		Bold:            true,
	}

	title := a.viewerTitle()
	if a.activePath == "" || a.activeIsDir {
		return t.Text{
			Content: title,
			Style:   style,
		}
	}

	file, ok := a.fileByPath[a.activePath]
	if !ok || file == nil {
		return t.Text{
			Content: title,
			Style:   style,
		}
	}

	spans := []t.Span{t.BoldSpan(title)}
	if statSpans := nonZeroChangeStatSpans(file.Additions, file.Deletions, theme, true); len(statSpans) > 0 {
		spans = append(spans, t.BoldSpan(" "))
		spans = append(spans, statSpans...)
	}

	return t.Text{Spans: spans, Style: style}
}

func (a *DiffApp) refreshDiff() {
	if repoRoot, err := a.provider.RepoRoot(); err == nil {
		a.repoRoot = repoRoot
	}
	if branch, err := a.provider.CurrentBranch(); err == nil {
		a.branch = branch
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
		a.fileByPath = map[string]*DiffFile{}
		a.filePathToTreePath = map[string][]int{}
		a.orderedFilePaths = nil
		a.treeState.Nodes.Set([]t.TreeNode[DiffTreeNodeData]{})
		a.treeState.CursorPath.Set(nil)
		a.treeState.Collapsed.Set(map[string]bool{})
		a.treeFilterNoMatches = false
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
		a.fileByPath = map[string]*DiffFile{}
		a.filePathToTreePath = map[string][]int{}
		a.orderedFilePaths = nil
		a.treeState.Nodes.Set([]t.TreeNode[DiffTreeNodeData]{})
		a.treeState.CursorPath.Set(nil)
		a.treeState.Collapsed.Set(map[string]bool{})
		a.treeFilterNoMatches = false
		a.diffViewState.SetRendered(messageToRendered("Error", a.errorMessage()))
		a.diffScrollState.SetOffset(0)
		return
	}

	a.loadErr = ""
	a.files = doc.Files
	a.renderedByPath = make(map[string]*RenderedFile, len(a.files))
	a.fileByPath = make(map[string]*DiffFile, len(a.files))
	for _, file := range a.files {
		if file == nil {
			continue
		}
		a.fileByPath[file.DisplayPath] = file
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
		a.treeFilterNoMatches = false
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
	a.syncTreeFilterSelection()
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

func (a *DiffApp) toggleDiffWrap() {
	a.diffHardWrap = !a.diffHardWrap
	if a.diffViewState != nil {
		a.diffViewState.ScrollX.Set(0)
	}
}

func (a *DiffApp) toggleSidebar() {
	a.sidebarVisible = !a.sidebarVisible
	if a.sidebarVisible {
		return
	}

	a.dividerFocusRequested = false
	a.dividerFocused = false

	switch a.focusedWidgetID {
	case diffSplitPaneID, diffFilesTreeID, diffFilesFilterID, diffFilesScrollID:
		t.RequestFocus(diffViewerScrollID)
	}
}

func (a *DiffApp) openTreeFilter() {
	if a.focusedWidgetID != diffFilesTreeID {
		return
	}
	a.treeFilterVisible = true
	if a.treeFilterInput != nil {
		a.treeFilterInput.ClearSelection()
		a.treeFilterInput.CursorEnd()
	}
	t.RequestFocus(diffFilesFilterID)
}

func (a *DiffApp) handleEscape() {
	if a.clearTreeFilter() {
		return
	}
	if a.focusedWidgetID == diffFilesFilterID && a.treeFilterVisible {
		a.treeFilterVisible = false
		t.RequestFocus(diffFilesTreeID)
	}
}

func (a *DiffApp) onTreeFilterChange(text string) {
	a.treeFilterVisible = true
	if a.treeFilterState != nil {
		a.treeFilterState.Query.Set(text)
	}
	a.syncTreeFilterSelection()
}

func (a *DiffApp) clearTreeFilter() bool {
	if a.treeFilterState == nil {
		return false
	}
	if a.treeFilterState.PeekQuery() == "" {
		return false
	}
	if a.treeFilterInput != nil {
		a.treeFilterInput.SetText("")
	}
	a.treeFilterState.Query.Set("")
	a.treeFilterVisible = false
	a.syncTreeFilterSelection()
	t.RequestFocus(diffFilesTreeID)
	return true
}

func (a *DiffApp) shouldShowTreeFilterInput() bool {
	if a.treeFilterVisible {
		return true
	}
	if a.focusedWidgetID == diffFilesFilterID {
		return true
	}
	if a.treeFilterState == nil {
		return false
	}
	return a.treeFilterState.PeekQuery() != ""
}

func (a *DiffApp) syncTreeFilterSelection() {
	query := ""
	options := t.FilterOptions{}
	if a.treeFilterState != nil {
		query = a.treeFilterState.PeekQuery()
		options = a.treeFilterState.PeekOptions()
	}
	if query == "" {
		a.treeFilterNoMatches = false
		if a.activePath == "" && !a.activeIsDir && len(a.orderedFilePaths) > 0 {
			a.selectFilePath(a.orderedFilePaths[0])
		}
		return
	}

	path, node, ok := findFirstTreeFilterMatch(a.treeState.Nodes.Peek(), nil, query, options)
	if !ok {
		a.setTreeFilterNoMatches(query)
		return
	}

	a.treeFilterNoMatches = false
	if a.activePath != "" || a.activeIsDir {
		return
	}
	a.treeState.CursorPath.Set(clonePath(path))
	a.onTreeCursorChange(node)
}

func (a *DiffApp) setTreeFilterNoMatches(query string) {
	a.treeFilterNoMatches = true
	a.treeState.CursorPath.Set(nil)
	a.activePath = ""
	a.activeIsDir = false
	a.diffViewState.SetRendered(messageToRendered("No matches", a.noFilterMatchesMessage(query)))
	a.diffScrollState.SetOffset(0)
}

func (a *DiffApp) noFilterMatchesMessage(query string) string {
	if query == "" {
		return "No files match the current filter.\n\nPress escape to clear the filter."
	}
	return fmt.Sprintf("No files match %q.\n\nPress escape to clear the filter.", query)
}

func (a *DiffApp) buildTreeFilterEmptyState(theme t.ThemeData) t.Widget {
	query := ""
	if a.treeFilterState != nil {
		query = a.treeFilterState.PeekQuery()
	}

	message := "No files match the current filter."
	if query != "" {
		message = fmt.Sprintf("No files match %q.", query)
	}

	return t.Column{
		Style: t.Style{
			Width:           t.Flex(1),
			Padding:         t.EdgeInsets{Top: 1, Left: 1, Right: 1},
			BackgroundColor: theme.Background,
		},
		Children: []t.Widget{
			t.Text{
				Content: message,
				Wrap:    t.WrapSoft,
				Style: t.Style{
					ForegroundColor: theme.TextMuted,
					Bold:            true,
				},
			},
			t.Text{
				Content: "Press escape to clear the filter.",
				Wrap:    t.WrapSoft,
				Style: t.Style{
					ForegroundColor: theme.TextMuted,
				},
			},
		},
	}
}

func (a *DiffApp) focusDivider() {
	if !a.sidebarVisible {
		return
	}
	target := a.dividerReturnTarget()
	a.dividerFocusRequested = true
	a.focusReturnID = target
	t.RequestFocus(diffSplitPaneID)
}

func (a *DiffApp) focusDividerFromPalette() {
	if !a.sidebarVisible {
		return
	}
	a.dividerFocusRequested = true
	a.focusReturnID = a.dividerReturnTarget()
	if a.commandPalette != nil {
		a.cancelThemePreview()
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
		a.cancelThemePreview()
		a.commandPalette.Close(false)
		return
	}
	a.themePreviewBase = ""
	a.themeCursorSynced = false
	a.commandPalette.Open()
}

func (a *DiffApp) syncFocusState(ctx t.BuildContext) {
	wasDividerFocused := a.dividerFocused
	focusedID := focusedWidgetID(ctx)
	a.focusedWidgetID = focusedID
	a.dividerFocused = a.sidebarVisible && focusedID == diffSplitPaneID
	if wasDividerFocused && !a.dividerFocused {
		a.dividerFocusRequested = false
	}
	if !a.sidebarVisible {
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
		{Divider: "Layout"},
		{
			Label:      "Toggle sidebar",
			FilterText: "Toggle sidebar layout panel",
			Hint:       "[ctrl+b]",
			Action:     a.paletteAction(a.toggleSidebar),
		},
		{
			Label:      "Focus divider",
			FilterText: "Focus divider split resize",
			Hint:       "[d]",
			Action:     a.focusDividerFromPalette,
		},
		{Divider: "Appearance"},
		{
			Label:      "Toggle line wrap",
			FilterText: "Toggle line wrap hard wrap soft wrap",
			Hint:       "[w]",
			Action:     a.paletteAction(a.toggleDiffWrap),
		},
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
		a.commitThemePreview()
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
		a.cancelThemePreview()
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
		a.cancelThemePreview()
		return
	}
	if a.themePreviewBase == "" {
		a.themePreviewBase = t.CurrentThemeName()
	}
	themeName, ok := item.Data.(string)
	if !ok || themeName == "" {
		return
	}
	if !a.themeCursorSynced {
		currentItem, hasCurrent := a.commandPalette.CurrentItem()
		if hasCurrent {
			currentThemeName, _ := currentItem.Data.(string)
			if currentThemeName == themeName {
				a.themeCursorSynced = true
				if selectPaletteTheme(level, t.CurrentThemeName()) {
					return
				}
			}
		}
	}
	t.SetTheme(themeName)
}

func (a *DiffApp) handlePaletteDismiss() {
	a.cancelThemePreview()
}

func (a *DiffApp) commitThemePreview() {
	a.finishThemePreview(true)
}

func (a *DiffApp) cancelThemePreview() {
	a.finishThemePreview(false)
}

func (a *DiffApp) finishThemePreview(commit bool) {
	if !commit && a.themePreviewBase != "" && t.CurrentThemeName() != a.themePreviewBase {
		t.SetTheme(a.themePreviewBase)
	}
	a.themePreviewBase = ""
	a.themeCursorSynced = false
}

func selectPaletteTheme(level *t.CommandPaletteLevel, themeName string) bool {
	if level == nil || level.ListState == nil || themeName == "" {
		return false
	}
	for idx, item := range level.Items {
		name, ok := item.Data.(string)
		if !ok || name != themeName {
			continue
		}
		if level.ListState.CursorIndex.Peek() == idx {
			return false
		}
		level.ListState.SelectIndex(idx)
		return true
	}
	return false
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

func (a *DiffApp) sidebarSummaryLabel() string {
	return fmt.Sprintf("%s %s", a.sidebarSummaryCountLabel(), a.modeLabel())
}

func (a *DiffApp) sidebarSummaryCountLabel() string {
	if a.treeFilterNoMatches {
		return fmt.Sprintf("0/%d", len(a.orderedFilePaths))
	}
	return fmt.Sprintf("%d", len(a.orderedFilePaths))
}

func (a *DiffApp) sidebarHeadingSpans(theme t.ThemeData) []t.Span {
	modeColor := theme.Error
	if a.staged {
		modeColor = theme.Success
	}
	modeStyle := t.SpanStyle{
		Foreground: modeColor,
		Bold:       true,
	}

	return []t.Span{
		t.StyledSpan(a.sidebarSummaryCountLabel(), modeStyle),
		t.StyledSpan(" ", modeStyle),
		t.StyledSpan(a.modeLabel(), modeStyle),
		t.BoldSpan(" ", theme.TextMuted),
		t.StyledSpan("[s]", t.SpanStyle{
			Foreground: theme.TextMuted,
			Faint:      true,
		}),
	}
}

func (a *DiffApp) sidebarTotals() (additions int, deletions int) {
	for _, file := range a.files {
		if file == nil {
			continue
		}
		additions += file.Additions
		deletions += file.Deletions
	}
	return additions, deletions
}

func (a *DiffApp) sidebarTotalsSpans(theme t.ThemeData) []t.Span {
	additions, deletions := a.sidebarTotals()
	return nonZeroChangeStatSpans(additions, deletions, theme, true)
}

func (a *DiffApp) viewerTitle() string {
	if a.activePath == "" {
		if a.loadErr != "" {
			return "Error"
		}
		if a.treeFilterNoMatches {
			return "No matches"
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

func findFirstTreeFilterMatch(nodes []t.TreeNode[DiffTreeNodeData], parentPath []int, query string, options t.FilterOptions) ([]int, DiffTreeNodeData, bool) {
	for idx, node := range nodes {
		path := append(clonePath(parentPath), idx)
		if t.MatchString(node.Data.Name, query, options).Matched {
			return path, node.Data, true
		}
		if _, _, ok := findFirstTreeFilterMatch(node.Children, path, query, options); ok {
			return path, node.Data, true
		}
	}
	return nil, DiffTreeNodeData{}, false
}

func nonZeroChangeTexts(additions int, deletions int) (addText string, delText string) {
	if additions > 0 {
		addText = fmt.Sprintf("+%d", additions)
	}
	if deletions > 0 {
		delText = fmt.Sprintf("-%d", deletions)
	}
	return addText, delText
}

func nonZeroChangeStatSpans(additions int, deletions int, theme t.ThemeData, bold bool) []t.Span {
	addText, delText := nonZeroChangeTexts(additions, deletions)
	if addText == "" && delText == "" {
		return nil
	}

	spans := make([]t.Span, 0, 3)
	if addText != "" {
		if bold {
			spans = append(spans, t.BoldSpan(addText, theme.Success))
		} else {
			spans = append(spans, t.ColorSpan(addText, theme.Success))
		}
	}
	if delText != "" {
		if len(spans) > 0 {
			spans = append(spans, t.PlainSpan(" "))
		}
		if bold {
			spans = append(spans, t.BoldSpan(delText, theme.Error))
		} else {
			spans = append(spans, t.ColorSpan(delText, theme.Error))
		}
	}
	return spans
}
