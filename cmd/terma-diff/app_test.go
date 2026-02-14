package main

import (
	"fmt"
	"strings"
	"testing"

	t "terma"

	"github.com/stretchr/testify/require"
)

func TestDiffApp_RefreshPreservesActiveFileWhenStillPresent(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("a.txt", "b.txt"),
			diffForPaths("b.txt", "c.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	require.True(t, app.selectFilePath("b.txt"))
	require.Equal(t, "b.txt", app.activePath)
	require.False(t, app.activeIsDir)

	app.refreshDiff()

	require.Equal(t, "b.txt", app.activePath)
	require.False(t, app.activeIsDir)
	require.Equal(t, app.filePathToTreePath["b.txt"], app.treeState.CursorPath.Peek())
}

func TestDiffApp_NextPrevCycleFilesAndSyncTreeCursor(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("pkg/b.go", "pkg/c.go", "a.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	require.GreaterOrEqual(t, len(app.orderedFilePaths), 3)

	first := app.orderedFilePaths[0]
	second := app.orderedFilePaths[1]
	last := app.orderedFilePaths[len(app.orderedFilePaths)-1]

	require.Equal(t, first, app.activePath)

	app.moveFileCursor(1)
	require.Equal(t, second, app.activePath)
	require.Equal(t, app.filePathToTreePath[second], app.treeState.CursorPath.Peek())

	app.moveFileCursor(-1)
	require.Equal(t, first, app.activePath)
	require.Equal(t, app.filePathToTreePath[first], app.treeState.CursorPath.Peek())

	app.moveFileCursor(-1)
	require.Equal(t, last, app.activePath)
	require.Equal(t, app.filePathToTreePath[last], app.treeState.CursorPath.Peek())
}

func TestDiffApp_NextPrevCycleOnlyFilteredFiles(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("a.go", "b.go", "c.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	app.onTreeFilterChange(".go")
	require.False(t, app.treeFilterNoMatches)
	require.Equal(t, "a.go", app.activePath)

	app.moveFileCursor(1)
	require.Equal(t, "b.go", app.activePath)

	app.moveFileCursor(1)
	require.Equal(t, "a.go", app.activePath)

	app.moveFileCursor(-1)
	require.Equal(t, "b.go", app.activePath)
}

func TestDiffApp_NextPrevStartsAtFilteredSetWhenActiveFileExcluded(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("a.go", "b.go", "c.txt"),
		},
	}

	app := NewDiffApp(provider, false)
	require.True(t, app.selectFilePath("c.txt"))
	require.Equal(t, "c.txt", app.activePath)

	app.onTreeFilterChange(".go")
	require.False(t, app.treeFilterNoMatches)
	require.Equal(t, "c.txt", app.activePath)

	app.moveFileCursor(1)
	require.Equal(t, "a.go", app.activePath)

	app.moveFileCursor(-1)
	require.Equal(t, "b.go", app.activePath)
}

func TestDiffApp_DirectoryCursorShowsSummaryInViewer(t *testing.T) {
	provider := &scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs: []string{
			diffForPaths("pkg/a.go", "pkg/b.go", "README.md"),
		},
	}

	app := NewDiffApp(provider, false)
	dirPath, ok := findTreePathByDataPath(app.treeState.Nodes.Peek(), "pkg")
	require.True(t, ok)

	node, ok := app.treeState.NodeAtPath(dirPath)
	require.True(t, ok)
	app.treeState.CursorPath.Set(clonePath(dirPath))
	app.onTreeCursorChange(node.Data)

	require.True(t, app.activeIsDir)
	require.Equal(t, "pkg", app.activePath)

	rendered := app.diffViewState.Rendered.Peek()
	require.NotNil(t, rendered)
	require.GreaterOrEqual(t, len(rendered.Lines), 4)
	require.True(t, strings.Contains(lineText(rendered.Lines[0]), "Directory: pkg"))
	require.True(t, strings.Contains(lineText(rendered.Lines[1]), "Touched files: 2"))
}

func TestDiffApp_CommandPaletteIncludesCommonActions(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	toggle := findPaletteItemByLabel(level.Items, "Toggle staged mode")
	require.True(tt, toggle.IsSelectable())
	require.Equal(tt, "[s]", toggle.Hint)

	refresh := findPaletteItemByLabel(level.Items, "Refresh")
	require.True(tt, refresh.IsSelectable())
	require.Equal(tt, "[r]", refresh.Hint)

	sidebar := findPaletteItemByLabel(level.Items, "Toggle sidebar")
	require.True(tt, sidebar.IsSelectable())
	require.Equal(tt, "[ctrl+b]", sidebar.Hint)

	wrap := findPaletteItemByLabel(level.Items, "Toggle line wrap")
	require.True(tt, wrap.IsSelectable())
	require.Equal(tt, "[w]", wrap.Hint)

	signs := findPaletteItemByLabel(level.Items, "Toggle +/- symbols")
	require.True(tt, signs.IsSelectable())

	divider := findPaletteItemByLabel(level.Items, "Focus divider")
	require.True(tt, divider.IsSelectable())
	require.Equal(tt, "[d]", divider.Hint)
}

func TestDiffApp_CommandPaletteUsesLayoutAndAppearanceSections(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	layoutDivider := -1
	appearanceDivider := -1
	sidebarIdx := -1
	dividerIdx := -1
	wrapIdx := -1
	signsIdx := -1
	themeIdx := -1

	for idx, item := range level.Items {
		switch {
		case item.Divider == "Layout":
			layoutDivider = idx
		case item.Divider == "Appearance":
			appearanceDivider = idx
		case item.Label == "Toggle sidebar":
			sidebarIdx = idx
		case item.Label == "Focus divider":
			dividerIdx = idx
		case item.Label == "Toggle line wrap":
			wrapIdx = idx
		case item.Label == "Toggle +/- symbols":
			signsIdx = idx
		case item.Label == "Theme":
			themeIdx = idx
		}
	}

	require.GreaterOrEqual(tt, layoutDivider, 0)
	require.GreaterOrEqual(tt, appearanceDivider, 0)
	require.Greater(tt, sidebarIdx, layoutDivider)
	require.Greater(tt, dividerIdx, layoutDivider)
	require.Greater(tt, appearanceDivider, dividerIdx)
	require.Greater(tt, wrapIdx, appearanceDivider)
	require.Greater(tt, signsIdx, wrapIdx)
	require.Greater(tt, themeIdx, signsIdx)
}

func TestDiffApp_KeybindsHideCommandsExposedInPalette(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	keybinds := app.Keybinds()

	require.True(tt, keybindIsHidden(keybinds, "s"))
	require.True(tt, keybindIsHidden(keybinds, "r"))
	require.True(tt, keybindIsHidden(keybinds, "d"))
	require.True(tt, keybindIsHidden(keybinds, "w"))
	require.True(tt, keybindIsHidden(keybinds, "ctrl+b"))
	require.False(tt, keybindIsHidden(keybinds, "ctrl+p"))
}

func TestDiffApp_KeybindsIncludeSidebarToggle(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	keybind, ok := findKeybindByKey(app.Keybinds(), "ctrl+b")
	require.True(tt, ok)
	require.Equal(tt, "Toggle sidebar", keybind.Name)
	require.True(tt, keybind.Hidden)
}

func TestDiffApp_KeybindsIncludeWrapToggle(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	keybind, ok := findKeybindByKey(app.Keybinds(), "w")
	require.True(tt, ok)
	require.Equal(tt, "Toggle line wrap", keybind.Name)
	require.True(tt, keybind.Hidden)
}

func TestDiffApp_FilterFilesKeybindVisibleOnlyWhenTreeFocused(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)

	app.focusedWidgetID = diffViewerScrollID
	keybind, ok := findKeybindByKey(app.Keybinds(), "/")
	require.True(tt, ok)
	require.True(tt, keybind.Hidden)

	app.focusedWidgetID = diffFilesTreeID
	keybind, ok = findKeybindByKey(app.Keybinds(), "/")
	require.True(tt, ok)
	require.False(tt, keybind.Hidden)

	app.focusedWidgetID = diffFilesFilterID
	keybind, ok = findKeybindByKey(app.Keybinds(), "/")
	require.True(tt, ok)
	require.True(tt, keybind.Hidden)
}

func TestDiffApp_ThemePreviewOnCursorChange(tt *testing.T) {
	originalTheme := t.CurrentThemeName()
	defer t.SetTheme(originalTheme)

	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	app.commandPalette.PushLevel(diffThemesPalette, app.themeItems())
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	// First cursor-change in Themes should sync to current theme.
	initialItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	app.handlePaletteCursorChange(initialItem)

	preview := t.CommandPaletteItem{}
	previewIdx := -1
	for idx, item := range level.Items {
		themeName, ok := item.Data.(string)
		if !ok || themeName == "" || themeName == t.CurrentThemeName() {
			continue
		}
		preview = item
		previewIdx = idx
		break
	}
	require.NotEmpty(tt, preview.Label, "expected at least one theme item different from current theme")
	require.GreaterOrEqual(tt, previewIdx, 0)

	level.ListState.SelectIndex(previewIdx)
	app.handlePaletteCursorChange(preview)

	themeName, _ := preview.Data.(string)
	require.Equal(tt, themeName, t.CurrentThemeName())
}

func TestDiffApp_ThemesMenuSelectsCurrentThemeByDefault(tt *testing.T) {
	originalTheme := t.CurrentThemeName()
	defer t.SetTheme(originalTheme)

	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	items := app.themeItems()

	selectableThemes := make([]string, 0, len(items))
	for _, item := range items {
		themeName, ok := item.Data.(string)
		if ok && themeName != "" {
			selectableThemes = append(selectableThemes, themeName)
		}
	}
	require.GreaterOrEqual(tt, len(selectableThemes), 2)

	currentTheme := selectableThemes[len(selectableThemes)-1]
	t.SetTheme(currentTheme)

	themeItems := app.themeItems()
	app.commandPalette.PushLevel(diffThemesPalette, themeItems)
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	initialItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	app.handlePaletteCursorChange(initialItem)

	selectedItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	selectedTheme, ok := selectedItem.Data.(string)
	require.True(tt, ok)
	require.Equal(tt, currentTheme, selectedTheme)
	require.Equal(tt, currentTheme, t.CurrentThemeName())
}

func TestDiffApp_ThemePreviewRevertsWhenLeavingThemesMenu(tt *testing.T) {
	originalTheme := t.CurrentThemeName()
	defer t.SetTheme(originalTheme)

	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	themeNames := paletteThemeNames(app.themeItems())
	require.GreaterOrEqual(tt, len(themeNames), 2)

	baseTheme := themeNames[0]
	t.SetTheme(baseTheme)

	app.commandPalette.PushLevel(diffThemesPalette, app.themeItems())
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	initialItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	app.handlePaletteCursorChange(initialItem)

	previewItem := t.CommandPaletteItem{}
	previewIdx := -1
	for idx, item := range level.Items {
		themeName, ok := item.Data.(string)
		if !ok || themeName == "" || themeName == baseTheme {
			continue
		}
		previewItem = item
		previewIdx = idx
		break
	}
	require.GreaterOrEqual(tt, previewIdx, 0)

	level.ListState.SelectIndex(previewIdx)
	app.handlePaletteCursorChange(previewItem)
	previewTheme, _ := previewItem.Data.(string)
	require.Equal(tt, previewTheme, t.CurrentThemeName())

	popped := app.commandPalette.PopLevel()
	require.True(tt, popped)
	rootItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	app.handlePaletteCursorChange(rootItem)

	require.Equal(tt, baseTheme, t.CurrentThemeName())
	require.Equal(tt, "", app.themePreviewBase)
}

func TestDiffApp_ThemeSelectionPersistsOnEnter(tt *testing.T) {
	originalTheme := t.CurrentThemeName()
	defer t.SetTheme(originalTheme)

	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	themeNames := paletteThemeNames(app.themeItems())
	require.GreaterOrEqual(tt, len(themeNames), 2)

	baseTheme := themeNames[0]
	selectedTheme := themeNames[1]
	t.SetTheme(baseTheme)

	app.commandPalette.PushLevel(diffThemesPalette, app.themeItems())
	level := app.commandPalette.CurrentLevel()
	require.NotNil(tt, level)

	initialItem, ok := app.commandPalette.CurrentItem()
	require.True(tt, ok)
	app.handlePaletteCursorChange(initialItem)

	var selectedItem t.CommandPaletteItem
	selectedIdx := -1
	for idx, item := range level.Items {
		themeName, ok := item.Data.(string)
		if ok && themeName == selectedTheme {
			selectedItem = item
			selectedIdx = idx
			break
		}
	}
	require.GreaterOrEqual(tt, selectedIdx, 0)
	require.NotNil(tt, selectedItem.Action)

	level.ListState.SelectIndex(selectedIdx)
	app.handlePaletteCursorChange(selectedItem)
	require.Equal(tt, selectedTheme, t.CurrentThemeName())

	selectedItem.Action()

	require.Equal(tt, selectedTheme, t.CurrentThemeName())
	require.Equal(tt, "", app.themePreviewBase)
}

func TestDiffApp_OpenTreeFilterRequiresFocusedTree(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)

	app.focusedWidgetID = diffViewerScrollID
	app.openTreeFilter()
	require.False(tt, app.treeFilterVisible)

	app.focusedWidgetID = diffFilesTreeID
	app.openTreeFilter()
	require.True(tt, app.treeFilterVisible)
}

func TestDiffApp_HandleEscapeClearsActiveTreeFilter(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	app.treeFilterVisible = true
	app.onTreeFilterChange("a")

	require.Equal(tt, "a", app.treeFilterState.PeekQuery())
	require.Equal(tt, "", app.treeFilterInput.GetText())

	app.treeFilterInput.SetText("a")
	app.handleEscape()

	require.Equal(tt, "", app.treeFilterState.PeekQuery())
	require.Equal(tt, "", app.treeFilterInput.GetText())
	require.False(tt, app.treeFilterVisible)
}

func TestDiffApp_FilterNoMatchesSetsExplicitState(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt", "b.txt")}}, false)

	app.onTreeFilterChange("zzz")

	require.True(tt, app.treeFilterNoMatches)
	require.Equal(tt, "", app.activePath)
	require.False(tt, app.activeIsDir)
	require.Equal(tt, "No matches", app.viewerTitle())
	require.Equal(tt, "0/2 unstaged", app.sidebarSummaryLabel())

	rendered := app.diffViewState.Rendered.Peek()
	require.NotNil(tt, rendered)
	require.Equal(tt, "No matches", rendered.Title)
	require.GreaterOrEqual(tt, len(rendered.Lines), 1)
	require.Contains(tt, lineText(rendered.Lines[0]), `No files match "zzz".`)
}

func TestDiffApp_ClearTreeFilterResetsNoMatchesState(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt", "b.txt")}}, false)
	app.treeFilterInput.SetText("zzz")
	app.onTreeFilterChange("zzz")
	require.True(tt, app.treeFilterNoMatches)

	cleared := app.clearTreeFilter()

	require.True(tt, cleared)
	require.False(tt, app.treeFilterNoMatches)
	require.Equal(tt, "", app.treeFilterState.PeekQuery())
	require.Equal(tt, "", app.treeFilterInput.GetText())
	require.False(tt, app.treeFilterVisible)
	require.Equal(tt, app.orderedFilePaths[0], app.activePath)
}

func TestDiffApp_RenderTreeNodeHighlightsMatchWithDefaultStyle(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("server.go")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	render := app.renderTreeNode(theme, false)
	rowWidget := render(
		DiffTreeNodeData{Name: "server.go", Path: "server.go", Additions: 1, Deletions: 1},
		t.TreeNodeContext{},
		t.MatchResult{
			Matched: true,
			Ranges:  []t.MatchRange{{Start: 0, End: len("server")}},
		},
	)

	row, ok := rowWidget.(t.Row)
	require.True(tt, ok)
	require.NotEmpty(tt, row.Children)

	label, ok := row.Children[0].(t.Text)
	require.True(tt, ok)
	require.NotEmpty(tt, label.Spans)

	highlight := t.MatchHighlightStyle(theme)
	found := false
	for _, span := range label.Spans {
		if span.Style == highlight {
			found = true
			break
		}
	}
	require.True(tt, found, "expected at least one highlighted span")
}

func TestDiffApp_RenderTreeNodeOmitsZeroStats(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("server.go")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	render := app.renderTreeNode(theme, false)

	addOnlyWidget := render(
		DiffTreeNodeData{Name: "added.go", Path: "added.go", Additions: 2, Deletions: 0},
		t.TreeNodeContext{},
		t.MatchResult{},
	)
	addOnlyRow, ok := addOnlyWidget.(t.Row)
	require.True(tt, ok)
	addOnlyText := strings.Join(rowTextContents(addOnlyRow), "|")
	require.Contains(tt, addOnlyText, "+2")
	require.NotContains(tt, addOnlyText, "-0")

	delOnlyWidget := render(
		DiffTreeNodeData{Name: "removed.go", Path: "removed.go", Additions: 0, Deletions: 3},
		t.TreeNodeContext{},
		t.MatchResult{},
	)
	delOnlyRow, ok := delOnlyWidget.(t.Row)
	require.True(tt, ok)
	delOnlyText := strings.Join(rowTextContents(delOnlyRow), "|")
	require.Contains(tt, delOnlyText, "-3")
	require.NotContains(tt, delOnlyText, "+0")
}

func TestDiffApp_LeftPaneTreeHasOneCellLeftPadding(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	ctx := t.NewBuildContext(nil, t.AnySignal[t.Focusable]{}, t.AnySignal[t.Widget]{}, nil)
	widget := app.buildLeftPane(ctx, theme)

	column, ok := widget.(t.Column)
	require.True(tt, ok)

	foundScrollable := false
	for _, child := range column.Children {
		scrollable, isScrollable := child.(t.Scrollable)
		if !isScrollable {
			continue
		}
		foundScrollable = true
		treeWidget, isTree := scrollable.Child.(SplitFriendlyTree)
		require.True(tt, isTree)
		require.Equal(tt, 1, treeWidget.Tree.Style.Padding.Left)
		break
	}
	require.True(tt, foundScrollable)
}

func TestDiffApp_LeftPaneHeaderRightAlignsTotals(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt", "b.txt")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	ctx := t.NewBuildContext(nil, t.AnySignal[t.Focusable]{}, t.AnySignal[t.Widget]{}, nil)
	widget := app.buildLeftPane(ctx, theme)

	column, ok := widget.(t.Column)
	require.True(tt, ok)
	require.NotEmpty(tt, column.Children)

	header, ok := column.Children[0].(t.Row)
	require.True(tt, ok)
	require.Len(tt, header.Children, 3)

	left, ok := header.Children[0].(t.Text)
	require.True(tt, ok)
	require.NotEmpty(tt, left.Spans)

	spacer, ok := header.Children[1].(t.Spacer)
	require.True(tt, ok)
	require.True(tt, spacer.Width.IsFlex())

	right, ok := header.Children[2].(t.Text)
	require.True(tt, ok)
	require.Len(tt, right.Spans, 3)
	require.Equal(tt, "+2", right.Spans[0].Text)
	require.Equal(tt, "-2", right.Spans[2].Text)
}

func TestDiffApp_SidebarSummaryLabelIncludesMode(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	require.Equal(tt, "1 unstaged", app.sidebarSummaryLabel())

	app.toggleMode()
	require.Equal(tt, "1 staged", app.sidebarSummaryLabel())
}

func TestDiffApp_SidebarHeadingIncludesStagedHint(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	spans := app.sidebarHeadingSpans(theme)
	require.Len(tt, spans, 5)
	require.Equal(tt, "1", spans[0].Text)
	require.True(tt, spans[0].Style.Bold)
	require.False(tt, spans[0].Style.Faint)
	require.Equal(tt, theme.Error, spans[0].Style.Foreground)
	require.Equal(tt, "unstaged", spans[2].Text)
	require.True(tt, spans[2].Style.Bold)
	require.False(tt, spans[2].Style.Faint)
	require.Equal(tt, theme.Error, spans[2].Style.Foreground)
	require.Equal(tt, "[s]", spans[4].Text)
	require.True(tt, spans[4].Style.Faint)
	require.Equal(tt, theme.TextMuted, spans[4].Style.Foreground)

	app.toggleMode()
	spans = app.sidebarHeadingSpans(theme)
	require.Equal(tt, "1", spans[0].Text)
	require.True(tt, spans[0].Style.Bold)
	require.False(tt, spans[0].Style.Faint)
	require.Equal(tt, theme.Success, spans[0].Style.Foreground)
	require.Equal(tt, "staged", spans[2].Text)
	require.True(tt, spans[2].Style.Bold)
	require.False(tt, spans[2].Style.Faint)
	require.Equal(tt, theme.Success, spans[2].Style.Foreground)
}

func TestDiffApp_SidebarTotalsSpansAggregatesAllFiles(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt", "b.txt")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	spans := app.sidebarTotalsSpans(theme)
	require.Len(tt, spans, 3)
	require.Equal(tt, "+2", spans[0].Text)
	require.True(tt, spans[0].Style.Bold)
	require.Equal(tt, theme.Success, spans[0].Style.Foreground)
	require.Equal(tt, "-2", spans[2].Text)
	require.True(tt, spans[2].Style.Bold)
	require.Equal(tt, theme.Error, spans[2].Style.Foreground)

	app.onTreeFilterChange("zzz")
	require.True(tt, app.treeFilterNoMatches)

	spans = app.sidebarTotalsSpans(theme)
	require.Len(tt, spans, 3)
	require.Equal(tt, "+2", spans[0].Text)
	require.Equal(tt, "-2", spans[2].Text)
}

func TestDiffApp_SidebarTotalsSpansOmitsZeroValues(tt *testing.T) {
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	addOnlyApp := NewDiffApp(&scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs:    []string{diffForPathWithStats("added.go", 4, 0)},
	}, false)
	addOnlySpans := addOnlyApp.sidebarTotalsSpans(theme)
	require.Len(tt, addOnlySpans, 1)
	require.Equal(tt, "+4", addOnlySpans[0].Text)
	require.Equal(tt, theme.Success, addOnlySpans[0].Style.Foreground)

	delOnlyApp := NewDiffApp(&scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs:    []string{diffForPathWithStats("removed.go", 0, 5)},
	}, false)
	delOnlySpans := delOnlyApp.sidebarTotalsSpans(theme)
	require.Len(tt, delOnlySpans, 1)
	require.Equal(tt, "-5", delOnlySpans[0].Text)
	require.Equal(tt, theme.Error, delOnlySpans[0].Style.Foreground)
}

func TestDiffApp_ViewerTitleIncludesLineStats(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	require.Equal(tt, "a.txt", app.viewerTitle())

	widget := app.buildViewerTitle(theme)
	text, ok := widget.(t.Text)
	require.True(tt, ok)
	require.True(tt, text.Style.Bold)
	require.Len(tt, text.Spans, 5)
	require.Equal(tt, "a.txt", text.Spans[0].Text)
	require.True(tt, text.Spans[0].Style.Bold)
	require.Equal(tt, "+1", text.Spans[2].Text)
	require.True(tt, text.Spans[2].Style.Bold)
	require.Equal(tt, theme.Success, text.Spans[2].Style.Foreground)
	require.Equal(tt, "-1", text.Spans[4].Text)
	require.True(tt, text.Spans[4].Style.Bold)
	require.Equal(tt, theme.Error, text.Spans[4].Style.Foreground)
}

func TestDiffApp_ViewerTitleOmitsZeroStats(tt *testing.T) {
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	addOnlyApp := NewDiffApp(&scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs:    []string{diffForPathWithStats("added.go", 3, 0)},
	}, false)
	addOnlyWidget := addOnlyApp.buildViewerTitle(theme)
	addOnlyText, ok := addOnlyWidget.(t.Text)
	require.True(tt, ok)
	addOnlySpanTexts := spanTexts(addOnlyText.Spans)
	require.Equal(tt, []string{"added.go", " ", "+3"}, addOnlySpanTexts)
	require.NotContains(tt, strings.Join(addOnlySpanTexts, ""), "-0")

	delOnlyApp := NewDiffApp(&scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		diffs:    []string{diffForPathWithStats("removed.go", 0, 2)},
	}, false)
	delOnlyWidget := delOnlyApp.buildViewerTitle(theme)
	delOnlyText, ok := delOnlyWidget.(t.Text)
	require.True(tt, ok)
	delOnlySpanTexts := spanTexts(delOnlyText.Spans)
	require.Equal(tt, []string{"removed.go", " ", "-2"}, delOnlySpanTexts)
	require.NotContains(tt, strings.Join(delOnlySpanTexts, ""), "+0")
}

func TestDiffApp_RightPaneUsesPaddedEmptyStateWhenNoDiffs(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	theme, ok := t.GetTheme(t.CurrentThemeName())
	require.True(tt, ok)

	widget := app.buildRightPane(theme)
	column, ok := widget.(t.Column)
	require.True(tt, ok)
	require.Len(tt, column.Children, 2)

	scrollable, ok := column.Children[1].(t.Scrollable)
	require.True(tt, ok)

	emptyState, ok := scrollable.Child.(t.Column)
	require.True(tt, ok, "expected a padded empty-state widget when no diff files exist")
	require.Equal(tt, 1, emptyState.Style.Padding.Top)
	require.Equal(tt, 2, emptyState.Style.Padding.Left)
	require.Equal(tt, 2, emptyState.Style.Padding.Right)
	require.Len(tt, emptyState.Children, 3)

	heading, ok := emptyState.Children[0].(t.Text)
	require.True(tt, ok)
	require.Equal(tt, "No unstaged changes.", heading.Content)
	require.True(tt, heading.Style.Bold)

	details, ok := emptyState.Children[2].(t.Text)
	require.True(tt, ok)
	require.Equal(tt, "Make edits in this repo and press r to refresh.", details.Content)

	app.toggleMode()
	widget = app.buildRightPane(theme)
	column, ok = widget.(t.Column)
	require.True(tt, ok)
	scrollable, ok = column.Children[1].(t.Scrollable)
	require.True(tt, ok)
	emptyState, ok = scrollable.Child.(t.Column)
	require.True(tt, ok)

	heading, ok = emptyState.Children[0].(t.Text)
	require.True(tt, ok)
	require.Equal(tt, "No staged changes.", heading.Content)
}

func TestDiffApp_ToggleSidebarVisibility(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	require.True(tt, app.sidebarVisible)

	app.toggleSidebar()
	require.False(tt, app.sidebarVisible)

	app.toggleSidebar()
	require.True(tt, app.sidebarVisible)
}

func TestDiffApp_ToggleDiffWrap(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	require.False(tt, app.diffHardWrap)

	app.toggleDiffWrap()
	require.True(tt, app.diffHardWrap)

	app.toggleDiffWrap()
	require.False(tt, app.diffHardWrap)
}

func TestDiffApp_DiffScrollStateHorizontalCallbacksMoveAndClamp(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)

	rendered := buildTestRenderedFile(20, 120)
	app.diffViewState.SetRendered(rendered)
	gutterWidth := renderedGutterWidth(rendered, app.diffHideChangeSigns)
	app.diffViewState.SetViewport(40, 10, gutterWidth)
	require.NotNil(tt, app.diffScrollState.OnScrollRight)
	require.NotNil(tt, app.diffScrollState.OnScrollLeft)

	handled := app.diffScrollState.ScrollRight(5)
	require.True(tt, handled)
	require.Equal(tt, 5, app.diffViewState.ScrollX.Peek())

	handled = app.diffScrollState.ScrollRight(1000)
	require.True(tt, handled)
	require.Equal(tt, app.diffViewState.MaxScrollX(gutterWidth), app.diffViewState.ScrollX.Peek())

	handled = app.diffScrollState.ScrollLeft(1000)
	require.True(tt, handled)
	require.Equal(tt, 0, app.diffViewState.ScrollX.Peek())
}

func TestDiffApp_DiffScrollStateHorizontalCallbacksNoopWhenWrapped(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)

	rendered := buildTestRenderedFile(20, 120)
	app.diffViewState.SetRendered(rendered)
	app.diffViewState.ScrollX.Set(9)

	app.diffHardWrap = true
	handled := app.diffScrollState.ScrollRight(1)
	require.False(tt, handled)
	require.Equal(tt, 9, app.diffViewState.ScrollX.Peek())
}

func TestDiffApp_ToggleDiffChangeSigns(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo", diffs: []string{diffForPaths("a.txt")}}, false)
	require.False(tt, app.diffHideChangeSigns)

	app.toggleDiffChangeSigns()
	require.True(tt, app.diffHideChangeSigns)

	app.toggleDiffChangeSigns()
	require.False(tt, app.diffHideChangeSigns)
}

func TestDiffApp_FocusDividerNoopWhenSidebarHidden(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{repoRoot: "/tmp/repo"}, false)
	app.sidebarVisible = false
	app.dividerFocusRequested = false

	app.focusDivider()
	require.False(tt, app.dividerFocusRequested)

	app.focusDividerFromPalette()
	require.False(tt, app.dividerFocusRequested)
}

func TestDiffApp_RefreshLoadsCurrentBranch(tt *testing.T) {
	app := NewDiffApp(&scriptedDiffProvider{
		repoRoot: "/tmp/repo",
		branch:   "feature/header-branch",
		diffs:    []string{diffForPaths("a.txt")},
	}, false)

	require.Equal(tt, "feature/header-branch", app.branch)
}

type scriptedDiffProvider struct {
	repoRoot string
	branch   string
	diffs    []string
	index    int
}

func (p *scriptedDiffProvider) LoadDiff(staged bool) (string, error) {
	if len(p.diffs) == 0 {
		return "", nil
	}
	if p.index >= len(p.diffs) {
		return p.diffs[len(p.diffs)-1], nil
	}
	value := p.diffs[p.index]
	p.index++
	return value, nil
}

func (p *scriptedDiffProvider) RepoRoot() (string, error) {
	return p.repoRoot, nil
}

func (p *scriptedDiffProvider) CurrentBranch() (string, error) {
	return p.branch, nil
}

func diffForPaths(paths ...string) string {
	var builder strings.Builder
	for _, path := range paths {
		builder.WriteString("diff --git a/")
		builder.WriteString(path)
		builder.WriteString(" b/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("index 1111111..2222222 100644\n")
		builder.WriteString("--- a/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("+++ b/")
		builder.WriteString(path)
		builder.WriteString("\n")
		builder.WriteString("@@ -1 +1 @@\n")
		builder.WriteString("-old\n")
		builder.WriteString("+new\n")
	}
	return builder.String()
}

func diffForPathWithStats(path string, additions int, deletions int) string {
	var builder strings.Builder
	builder.WriteString("diff --git a/")
	builder.WriteString(path)
	builder.WriteString(" b/")
	builder.WriteString(path)
	builder.WriteString("\n")
	builder.WriteString("index 1111111..2222222 100644\n")
	builder.WriteString("--- a/")
	builder.WriteString(path)
	builder.WriteString("\n")
	builder.WriteString("+++ b/")
	builder.WriteString(path)
	builder.WriteString("\n")

	switch {
	case additions > 0 && deletions == 0:
		builder.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", additions))
		for i := 0; i < additions; i++ {
			builder.WriteString(fmt.Sprintf("+new%d\n", i+1))
		}
	case additions == 0 && deletions > 0:
		builder.WriteString(fmt.Sprintf("@@ -1,%d +0,0 @@\n", deletions))
		for i := 0; i < deletions; i++ {
			builder.WriteString(fmt.Sprintf("-old%d\n", i+1))
		}
	default:
		builder.WriteString("@@ -1 +1 @@\n")
		builder.WriteString("-old\n")
		builder.WriteString("+new\n")
	}

	return builder.String()
}

func findTreePathByDataPath(nodes []t.TreeNode[DiffTreeNodeData], target string) ([]int, bool) {
	var walk func(items []t.TreeNode[DiffTreeNodeData], prefix []int) ([]int, bool)
	walk = func(items []t.TreeNode[DiffTreeNodeData], prefix []int) ([]int, bool) {
		for idx, node := range items {
			next := append(clonePath(prefix), idx)
			if node.Data.Path == target {
				return next, true
			}
			if path, ok := walk(node.Children, next); ok {
				return path, true
			}
		}
		return nil, false
	}
	return walk(nodes, nil)
}

func findPaletteItemByLabel(items []t.CommandPaletteItem, label string) t.CommandPaletteItem {
	for _, item := range items {
		if item.Label == label {
			return item
		}
	}
	return t.CommandPaletteItem{}
}

func keybindIsHidden(keybinds []t.Keybind, key string) bool {
	for _, keybind := range keybinds {
		if keybind.Key == key {
			return keybind.Hidden
		}
	}
	return false
}

func findKeybindByKey(keybinds []t.Keybind, key string) (t.Keybind, bool) {
	for _, keybind := range keybinds {
		if keybind.Key == key {
			return keybind, true
		}
	}
	return t.Keybind{}, false
}

func paletteThemeNames(items []t.CommandPaletteItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		themeName, ok := item.Data.(string)
		if ok && themeName != "" {
			names = append(names, themeName)
		}
	}
	return names
}

func spanTexts(spans []t.Span) []string {
	texts := make([]string, 0, len(spans))
	for _, span := range spans {
		texts = append(texts, span.Text)
	}
	return texts
}

func rowTextContents(row t.Row) []string {
	texts := []string{}
	for _, child := range row.Children {
		text, ok := child.(t.Text)
		if !ok {
			continue
		}
		if len(text.Spans) > 0 {
			var builder strings.Builder
			for _, span := range text.Spans {
				builder.WriteString(span.Text)
			}
			texts = append(texts, builder.String())
			continue
		}
		texts = append(texts, text.Content)
	}
	return texts
}
