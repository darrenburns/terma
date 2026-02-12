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

	activePath     string
	renderedByPath map[string]*RenderedFile

	listState       *t.ListState[*DiffFile]
	listScrollState *t.ScrollState
	viewerState     *t.TextAreaState
	splitState      *t.SplitPaneState
	themeNames      []string
}

func NewDiffApp(provider DiffProvider, staged bool) *DiffApp {
	viewerState := t.NewTextAreaState("")
	viewerState.ReadOnly.Set(true)
	viewerState.WrapMode.Set(t.WrapNone)
	viewerState.CursorIndex.Set(0)

	app := &DiffApp{
		provider:        provider,
		staged:          staged,
		listState:       t.NewListState([]*DiffFile{}),
		listScrollState: t.NewScrollState(),
		viewerState:     viewerState,
		splitState:      t.NewSplitPaneState(0.30),
		renderedByPath:  map[string]*RenderedFile{},
		themeNames:      t.ThemeNames(),
	}

	app.refreshDiff()
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
	palette := NewThemePalette(theme)

	var highlighter t.Highlighter
	lineHighlights := []t.LineHighlight{}
	if rendered := a.currentRendered(); rendered != nil {
		highlighter = DiffHighlighter{Tokens: rendered.Tokens, Palette: palette}
		lineHighlights = buildLineHighlights(rendered, palette)
	}

	split := t.SplitPane{
		ID:                "terma-diff-split",
		State:             a.splitState,
		Orientation:       t.SplitHorizontal,
		DividerSize:       1,
		MinPaneSize:       20,
		DisableFocus:      true,
		DividerBackground: theme.Surface2,
		First:             a.buildLeftPane(theme),
		Second:            a.buildRightPane(theme, highlighter, lineHighlights),
	}

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
		Body: split,
	}
}

func (a *DiffApp) buildHeader(theme t.ThemeData) t.Widget {
	repoName := "(unknown repo)"
	if a.repoRoot != "" {
		repoName = filepath.Base(a.repoRoot)
	}

	status := fmt.Sprintf("Repo: %s  Mode: %s  Files: %d  Theme: %s", repoName, a.modeLabel(), len(a.files), t.CurrentThemeName())
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

func (a *DiffApp) buildLeftPane(theme t.ThemeData) t.Widget {
	return t.Column{
		Height: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.Scrollable{
				ID:    "terma-diff-files-scroll",
				State: a.listScrollState,
				Style: t.Style{
					Height: t.Flex(1),
					Border: t.RoundedBorder(theme.Border, t.BorderTitle("Files")),
				},
				Child: t.List[*DiffFile]{
					ID:             "terma-diff-files-list",
					State:          a.listState,
					ScrollState:    a.listScrollState,
					OnCursorChange: a.onFileCursorChange,
					RenderItem:     a.renderFileItem(theme),
				},
			},
		},
	}
}

func (a *DiffApp) buildRightPane(theme t.ThemeData, highlighter t.Highlighter, lineHighlights []t.LineHighlight) t.Widget {
	title := "Diff"
	if a.activePath != "" {
		title = a.activePath
	}

	return t.TextArea{
		ID:             "terma-diff-viewer",
		State:          a.viewerState,
		Highlighter:    highlighter,
		LineHighlights: lineHighlights,
		Style: t.Style{
			Width:           t.Flex(1),
			Height:          t.Flex(1),
			Padding:         t.EdgeInsetsXY(1, 0),
			BackgroundColor: theme.Surface,
			Border:          t.RoundedBorder(theme.Border, t.BorderTitle(title)),
		},
		ExtraKeybinds: []t.Keybind{
			{Key: "j", Hidden: true, Action: func() { a.viewerState.CursorDown() }},
			{Key: "k", Hidden: true, Action: func() { a.viewerState.CursorUp() }},
			{Key: "h", Hidden: true, Action: func() { a.viewerState.CursorLeft() }},
			{Key: "l", Hidden: true, Action: func() { a.viewerState.CursorRight() }},
		},
	}
}

func (a *DiffApp) renderFileItem(theme t.ThemeData) func(file *DiffFile, active bool, selected bool) t.Widget {
	return func(file *DiffFile, active bool, _ bool) t.Widget {
		if file == nil {
			return t.Text{Content: ""}
		}

		rowStyle := t.Style{
			Width:   t.Flex(1),
			Padding: t.EdgeInsetsXY(1, 0),
		}
		pathStyle := t.Style{ForegroundColor: theme.Text}
		addStyle := t.Style{ForegroundColor: theme.Success}
		delStyle := t.Style{ForegroundColor: theme.Error}
		if active {
			rowStyle.BackgroundColor = theme.ActiveCursor
			pathStyle.ForegroundColor = theme.SelectionText
			addStyle.ForegroundColor = theme.SelectionText
			delStyle.ForegroundColor = theme.SelectionText
		}

		return t.Row{
			Style: rowStyle,
			Children: []t.Widget{
				t.Text{Content: file.DisplayPath, Style: pathStyle},
				t.Spacer{Width: t.Flex(1)},
				t.Text{Content: fmt.Sprintf("+%d", file.Additions), Style: addStyle},
				t.Text{Content: " "},
				t.Text{Content: fmt.Sprintf("-%d", file.Deletions), Style: delStyle},
			},
		}
	}
}

func (a *DiffApp) refreshDiff() {
	if repoRoot, err := a.provider.RepoRoot(); err == nil {
		a.repoRoot = repoRoot
	}

	raw, err := a.provider.LoadDiff(a.staged)
	if err != nil {
		a.loadErr = err.Error()
		a.files = nil
		a.activePath = ""
		a.renderedByPath = map[string]*RenderedFile{}
		a.listState.SetItems([]*DiffFile{})
		a.setViewerText(a.errorMessage())
		return
	}

	doc, err := parseUnifiedDiff(raw)
	if err != nil {
		a.loadErr = fmt.Sprintf("parse error: %v", err)
		a.files = nil
		a.activePath = ""
		a.renderedByPath = map[string]*RenderedFile{}
		a.listState.SetItems([]*DiffFile{})
		a.setViewerText(a.errorMessage())
		return
	}

	a.loadErr = ""
	a.files = doc.Files
	a.renderedByPath = make(map[string]*RenderedFile, len(a.files))
	a.listState.SetItems(a.files)

	if len(a.files) == 0 {
		a.activePath = ""
		a.setViewerText(a.emptyMessage())
		return
	}

	targetIdx := a.indexOfPath(a.activePath)
	if targetIdx < 0 {
		targetIdx = 0
	}
	a.listState.SelectIndex(targetIdx)
	a.updateActiveFromCursor()
}

func (a *DiffApp) moveFileCursor(delta int) {
	if len(a.files) == 0 {
		return
	}
	current := a.listState.CursorIndex.Peek()
	next := current + delta
	if next < 0 {
		next = len(a.files) - 1
	}
	if next >= len(a.files) {
		next = 0
	}
	a.listState.SelectIndex(next)
	a.updateActiveFromCursor()
}

func (a *DiffApp) onFileCursorChange(file *DiffFile) {
	a.setActiveFile(file)
}

func (a *DiffApp) updateActiveFromCursor() {
	file, ok := a.listState.SelectedItem()
	if !ok || file == nil {
		a.activePath = ""
		a.setViewerText(a.emptyMessage())
		return
	}
	a.setActiveFile(file)
}

func (a *DiffApp) setActiveFile(file *DiffFile) {
	if file == nil {
		return
	}
	a.activePath = file.DisplayPath
	rendered, ok := a.renderedByPath[file.DisplayPath]
	if !ok {
		rendered = buildRenderedFile(file)
		a.renderedByPath[file.DisplayPath] = rendered
	}
	a.setViewerText(rendered.Text)
}

func (a *DiffApp) setViewerText(text string) {
	a.viewerState.SetText(text)
	a.viewerState.CursorIndex.Set(0)
	a.viewerState.SelectionAnchor.Set(-1)
}

func (a *DiffApp) currentRendered() *RenderedFile {
	if a.activePath == "" {
		return nil
	}
	return a.renderedByPath[a.activePath]
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

func (a *DiffApp) indexOfPath(path string) int {
	if path == "" {
		return -1
	}
	for i, file := range a.files {
		if file != nil && file.DisplayPath == path {
			return i
		}
	}
	return -1
}

func buildLineHighlights(rendered *RenderedFile, palette ThemePalette) []t.LineHighlight {
	if rendered == nil || len(rendered.LineKinds) == 0 {
		return nil
	}
	highlights := make([]t.LineHighlight, 0, len(rendered.LineKinds))
	for lineIdx, kind := range rendered.LineKinds {
		if style, ok := palette.LineStyleForKind(kind); ok {
			highlights = append(highlights, t.LineHighlight{
				StartLine: lineIdx,
				EndLine:   lineIdx + 1,
				Style:     style,
			})
		}
	}
	return highlights
}
