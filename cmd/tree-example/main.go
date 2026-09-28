// Command tree-example demonstrates the Tree widget: lazy loading through
// OnExpand, filtering with match highlights, multi-select, and the OnSelect and
// OnCursorChange callbacks.
//
//	↑↓ jk      move the cursor (g/G or home/end jump to first/last)
//	←→ hl      collapse / expand (or move to parent / first child)
//	space      toggle the folder under the cursor
//	shift+↑↓   extend the selection
//	enter      open the node under the cursor (fires OnSelect)
//	/ ctrl+f   focus the filter
//	ctrl+t     toggle fuzzy / contains matching
//	e          load and expand every folder
//	c          collapse every folder
//	r          reset the tree
//	t          cycle theme
//	escape     clear the filter, or the selection when there is no filter
package main

import (
	"fmt"
	"log"
	"path"
	"strings"
	"sync/atomic"
	"time"

	t "github.com/darrenburns/terma"
)

// Theme names for cycling
var themeNames = []string{
	t.ThemeNameRosePine,
	t.ThemeNameDracula,
	t.ThemeNameTokyoNight,
	t.ThemeNameCatppuccin,
	t.ThemeNameGruvbox,
	t.ThemeNameNord,
	t.ThemeNameSolarized,
	t.ThemeNameKanagawa,
	t.ThemeNameMonokai,
}

const (
	sidebarWidth = 34
	loadDelay    = 400 * time.Millisecond
	maxEvents    = 3
)

// FileInfo is the data held by each tree node.
type FileInfo struct {
	Name  string
	Path  string
	IsDir bool
}

// event is one line in the Events panel.
type event struct {
	kind   string // short verb, e.g. "loaded"
	colour string // theme colour name for the verb, e.g. "$Success"
	detail string
}

// TreeExampleApp holds the tree state and everything the demo shows about it.
type TreeExampleApp struct {
	treeState    *t.TreeState[FileInfo]
	filterState  *t.FilterState
	filterInput  *t.TextInputState
	scrollState  *t.ScrollState
	lazyChildren map[string][]t.TreeNode[FileInfo]
	totalDirs    int

	loading    t.AnySignal[map[string]bool] // folders whose children are loading
	events     t.AnySignal[[]event]         // newest first
	themeIndex t.Signal[int]
	generation atomic.Int64 // bumped on reset so stale loads are dropped
}

func NewTreeExampleApp() *TreeExampleApp {
	roots, lazy := sampleTree()
	return &TreeExampleApp{
		treeState:    t.NewTreeState(roots),
		filterState:  t.NewFilterState(),
		filterInput:  t.NewTextInputState(""),
		scrollState:  t.NewScrollState(),
		lazyChildren: lazy,
		totalDirs:    countDirs(loadAll(roots, lazy)),
		loading:      t.NewAnySignal(map[string]bool{}),
		events:       t.NewAnySignal([]event{{kind: "ready", colour: "$TextMuted", detail: "folders load on first expand"}}),
		themeIndex:   t.NewSignal(0),
	}
}

func (a *TreeExampleApp) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "/", Name: "Filter", Action: a.focusFilter},
		{Key: "ctrl+f", Name: "Filter", Action: a.focusFilter, Hidden: true},
		{Key: "ctrl+t", Name: "Fuzzy", Action: a.toggleFuzzy},
		{Key: "e", Name: "Expand all", Action: a.expandAll},
		{Key: "c", Name: "Collapse all", Action: a.collapseAll},
		{Key: "r", Name: "Reset", Action: a.reset},
		{Key: "t", Name: "Theme", Action: a.cycleTheme},
		{Key: "escape", Name: "Clear", Action: a.clear, Hidden: true},
	}
}

func (a *TreeExampleApp) focusFilter() {
	t.RequestFocus("tree-filter")
}

func (a *TreeExampleApp) toggleFuzzy() {
	a.filterState.Mode.Update(func(mode t.FilterMode) t.FilterMode {
		if mode == t.FilterFuzzy {
			return t.FilterContains
		}
		return t.FilterFuzzy
	})
}

// clear drops the filter if there is one, otherwise the selection.
func (a *TreeExampleApp) clear() {
	if a.filterState.PeekQuery() != "" {
		a.clearFilter()
		return
	}
	a.treeState.ClearSelection()
}

func (a *TreeExampleApp) clearFilter() {
	a.filterInput.SetText("")
	a.filterState.Query.Set("")
	a.logEvent(event{kind: "filter", colour: "$TextMuted", detail: "cleared"})
	t.RequestFocus("file-tree")
}

// expandAll loads every lazy folder immediately and expands the whole tree.
func (a *TreeExampleApp) expandAll() {
	a.treeState.Nodes.Set(loadAll(a.treeState.Nodes.Peek(), a.lazyChildren))
	a.treeState.ExpandAll()
	a.logEvent(event{kind: "expand", colour: "$Success", detail: "all folders (lazy ones loaded)"})
}

func (a *TreeExampleApp) collapseAll() {
	// Keep the cursor visible by moving it to its top-level folder first.
	if cursor := a.treeState.CursorPath.Peek(); len(cursor) > 1 {
		a.treeState.CursorPath.Set([]int{cursor[0]})
	}
	a.treeState.CollapseAll()
	a.logEvent(event{kind: "collapse", colour: "$Warning", detail: "all folders"})
}

func (a *TreeExampleApp) reset() {
	a.generation.Add(1)
	roots, _ := sampleTree()
	a.treeState.Nodes.Set(roots)
	a.treeState.Collapsed.Set(map[string]bool{})
	a.treeState.ClearSelection()
	a.treeState.CursorPath.Set([]int{0})
	a.loading.Set(map[string]bool{})
	a.filterInput.SetText("")
	a.filterState.Query.Set("")
	a.logEvent(event{kind: "reset", colour: "$Warning", detail: "lazy folders unloaded"})
}

func (a *TreeExampleApp) cycleTheme() {
	a.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

// handleExpand is the Tree's OnExpand callback. It runs when a folder whose
// children are nil is expanded, and loads them after a simulated delay.
func (a *TreeExampleApp) handleExpand(info FileInfo, _ []int, setChildren func([]t.TreeNode[FileInfo])) {
	children, ok := a.lazyChildren[info.Path]
	if !ok {
		setChildren([]t.TreeNode[FileInfo]{})
		return
	}
	a.setLoading(info.Path, true)
	a.logEvent(event{kind: "loading", colour: "$Warning", detail: info.Path})
	generation := a.generation.Load()
	go func() {
		time.Sleep(loadDelay)
		if a.generation.Load() != generation {
			return // the tree was reset while this folder was loading
		}
		setChildren(children)
		a.setLoading(info.Path, false)
		a.logEvent(event{kind: "loaded", colour: "$Success", detail: fmt.Sprintf("%s (%s)", info.Path, plural(len(children), "item"))})
	}()
}

func (a *TreeExampleApp) setLoading(path string, loading bool) {
	a.loading.Update(func(current map[string]bool) map[string]bool {
		next := make(map[string]bool, len(current)+1)
		for k, v := range current {
			next[k] = v
		}
		if loading {
			next[path] = true
		} else {
			delete(next, path)
		}
		return next
	})
}

// handleSelect is the Tree's OnSelect callback, fired by enter.
func (a *TreeExampleApp) handleSelect(info FileInfo, selected []FileInfo) {
	if len(selected) > 1 {
		a.logEvent(event{kind: "open", colour: "$Accent", detail: plural(len(selected), "selected node")})
		return
	}
	a.logEvent(event{kind: "open", colour: "$Accent", detail: info.Path})
}

// handleCursorChange is the Tree's OnCursorChange callback. Consecutive cursor
// events replace each other so they don't push everything else out of the log.
func (a *TreeExampleApp) handleCursorChange(info FileInfo) {
	a.logEvent(event{kind: "cursor", colour: "$Info", detail: info.Path})
}

func (a *TreeExampleApp) logEvent(e event) {
	a.events.Update(func(events []event) []event {
		if len(events) > 0 && e.kind == "cursor" && events[0].kind == "cursor" {
			events = events[1:]
		}
		next := append([]event{e}, events...)
		if len(next) > maxEvents {
			next = next[:maxEvents]
		}
		return next
	})
}

func (a *TreeExampleApp) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "tree-example-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			header{themeName: themeNames[a.themeIndex.Get()]},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
		},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: []t.Widget{
				t.Column{
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						filterPanel{app: a},
						fill(treePanel{app: a}),
						eventsPanel{app: a},
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{app: a},
						fill(selectionPanel{app: a}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// header is the title bar across the top of the screen.
type header struct {
	themeName string
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ Tree Explorer[/]  [$TextMuted]Lazy loading, filtering and multi-select[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", h.themeName), theme),
		},
	}
}

// fill gives a component the remaining space in its Column. Rows and Columns
// read Flex from their direct children, so a component's own Flex(1) needs a
// plain wrapper to take effect.
func fill(child t.Widget) t.Widget {
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{child},
	}
}

// panelStyle is the bordered look shared by every panel in the demo.
func panelStyle(theme t.ThemeData, title string, focused bool) t.Style {
	color := theme.Border
	titleColor := "$TextMuted"
	if focused {
		color = theme.FocusRing
		titleColor = "$FocusRing"
	}
	return t.Style{
		BackgroundColor: theme.Background,
		Border:          t.RoundedBorder(color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] %s [/]", titleColor, title))),
		Padding:         t.EdgeInsetsXY(1, 0),
	}
}

// filterPanel holds the text input that narrows the tree, plus the match mode.
type filterPanel struct {
	app *TreeExampleApp
}

func (f filterPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := f.app
	input := t.TextInput{
		ID:          "tree-filter",
		State:       a.filterInput,
		Placeholder: "Type to filter loaded files…",
		Width:       t.Flex(1),
		Style: t.Style{
			ForegroundColor: theme.Text,
		},
		OnChange: func(text string) {
			a.filterState.Query.Set(text)
		},
		OnSubmit: func(string) {
			t.RequestFocus("file-tree")
		},
		ExtraKeybinds: []t.Keybind{
			{Key: "escape", Name: "Clear filter", Action: a.clearFilter},
		},
	}

	mode := "[$TextMuted]contains[/]"
	if a.filterState.Mode.Get() == t.FilterFuzzy {
		mode = "[b $Accent]fuzzy[/]"
	}

	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   panelStyle(theme, "Filter", ctx.IsFocused(input)),
		Children: []t.Widget{
			t.ParseMarkupToText("[$Accent]⌕[/]", theme),
			input,
			t.ParseMarkupToText(mode, theme),
		},
	}
}

// treePanel shows the scrolling tree, or a hint when the filter hides
// everything.
type treePanel struct {
	app *TreeExampleApp
}

func (p treePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := p.app
	tree := t.Tree[FileInfo]{
		ID:          "file-tree",
		State:       a.treeState,
		NodeID:      func(info FileInfo) string { return info.Path },
		Filter:      a.filterState,
		ScrollState: a.scrollState,
		MultiSelect: true,
		HasChildren: func(info FileInfo) bool {
			return info.IsDir
		},
		MatchNode: func(info FileInfo, query string, options t.FilterOptions) t.MatchResult {
			return t.MatchString(info.Name, query, options)
		},
		OnExpand:       a.handleExpand,
		OnSelect:       a.handleSelect,
		OnCursorChange: a.handleCursorChange,
		Style: t.Style{
			Width: t.Flex(1),
		},
	}
	focused := ctx.IsFocused(tree)
	tree.RenderNodeWithMatch = a.renderNode(theme, focused)

	var empty t.Widget
	if query := a.filterState.QueryText(); query != "" {
		if countMatches(a.treeState.Nodes.Get(), query, a.filterState.Options()) == 0 {
			empty = t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]No loaded file matches [/][b $Warning]%q[/][$TextMuted]. Press [b]esc[/] to clear the filter, or [b $Success]e[/] to load every folder.[/]", query), theme),
				Wrap:  t.WrapSoft,
			}
		}
	}

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  panelStyle(theme, "Files", focused),
		Children: []t.Widget{
			t.ShowWhen(empty != nil, empty),
			t.Scrollable{
				ID:                  "tree-scroll",
				State:               a.scrollState,
				ScrollbarThumbColor: theme.ScrollbarThumb,
				ScrollbarTrackColor: theme.ScrollbarTrack,
				Style: t.Style{
					Width:  t.Flex(1),
					Height: t.Flex(1),
				},
				Child: tree,
			},
		},
	}
}

// renderNode draws one row of the tree: folders in bold with a trailing slash,
// files coloured by extension, filter matches highlighted, and a loading hint
// while a folder's children are on their way.
func (a *TreeExampleApp) renderNode(theme t.ThemeData, treeFocused bool) func(FileInfo, t.TreeNodeContext, t.MatchResult) t.Widget {
	highlight := t.MatchHighlightStyle(theme)
	return func(info FileInfo, nodeCtx t.TreeNodeContext, match t.MatchResult) t.Widget {
		style := treeNodeStyle(theme, nodeCtx, treeFocused)
		style.Width = t.Flex(1)

		// The cursor row and folders shown only because a descendant matches
		// take their colour from the row style; everything else is coloured by
		// kind.
		var base t.SpanStyle
		plain := (nodeCtx.Active && treeFocused) || nodeCtx.FilteredAncestor
		if !plain {
			base.Foreground = nameColour(theme, info)
		}
		base.Bold = info.IsDir

		name := info.Name
		var spans []t.Span
		if match.Matched && len(match.Ranges) > 0 {
			spans = t.HighlightSpans(name, match.Ranges, highlight)
		} else {
			spans = []t.Span{{Text: name}}
		}
		for i := range spans {
			if !spans[i].Style.Foreground.IsSet() {
				spans[i].Style.Foreground = base.Foreground
			}
			spans[i].Style.Bold = base.Bold
		}
		if info.IsDir {
			spans = append(spans, t.Span{Text: "/", Style: base})
		}
		if a.loading.Get()[info.Path] {
			loadingStyle := t.SpanStyle{Italic: true}
			if !plain {
				loadingStyle.Foreground = theme.Warning
			}
			spans = append(spans, t.Span{Text: "  loading…", Style: loadingStyle})
		}
		return t.Text{Spans: spans, Style: style}
	}
}

func nameColour(theme t.ThemeData, info FileInfo) t.Color {
	if info.IsDir {
		return theme.Primary
	}
	switch path.Ext(info.Name) {
	case ".go", ".mod":
		return theme.Info
	case ".md":
		return theme.Secondary
	}
	return theme.Text
}

func treeNodeStyle(theme t.ThemeData, nodeCtx t.TreeNodeContext, treeFocused bool) t.Style {
	style := t.Style{ForegroundColor: theme.Text}
	if nodeCtx.FilteredAncestor {
		style.ForegroundColor = theme.TextMuted
	}
	if nodeCtx.Active && treeFocused {
		style.BackgroundColor = theme.ActiveCursor
		style.ForegroundColor = theme.SelectionText
		return style
	}
	if nodeCtx.Selected {
		style.BackgroundColor = theme.Selection
	}
	return style
}

// eventsPanel logs the Tree's callbacks as they fire.
type eventsPanel struct {
	app *TreeExampleApp
}

func (e eventsPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	events := e.app.events.Get()
	rows := make([]t.Widget, 0, maxEvents)
	for i, ev := range events {
		detail := "$Text"
		if i > 0 {
			detail = "$TextMuted"
		}
		rows = append(rows, t.ParseMarkupToText(fmt.Sprintf("[b %s]%-8s[/] [%s]%s[/]", ev.colour, ev.kind, detail, ev.detail), theme))
	}
	// Keep the panel a constant height so the tree above doesn't jump.
	for len(rows) < maxEvents {
		rows = append(rows, t.Text{Content: " "})
	}
	return t.Column{
		Width:    t.Flex(1),
		Style:    panelStyle(theme, "Events", false),
		Children: rows,
	}
}

// statePanel shows the live TreeState. It subscribes to the state's signals
// itself, so moving the cursor only rebuilds this panel rather than the app.
type statePanel struct {
	app *TreeExampleApp
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	a := s.app
	state := a.treeState
	nodes := state.Nodes.Get()
	cursor := state.CursorPath.Get()
	collapsed := state.Collapsed.Get()
	selected := len(state.Selection.Get())
	query := a.filterState.QueryText()
	options := a.filterState.Options()

	cursorLabel, depth := "—", "—"
	if node, ok := state.NodeAtPath(cursor); ok {
		cursorLabel = truncateLeft(node.Data.Path, sidebarWidth-14)
		depth = fmt.Sprintf("%d", len(cursor)-1)
	}

	filter := "[$TextMuted]off[/]"
	if query != "" {
		matches := countMatches(nodes, query, options)
		colour := "$Success"
		if matches == 0 {
			colour = "$Error"
		}
		filter = fmt.Sprintf("[b %s]%s[/]", colour, plural(matches, "match"))
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "State", false),
		Children: []t.Widget{
			statRow(theme, "Cursor", fmt.Sprintf("[b $Info]%s[/]", cursorLabel)),
			statRow(theme, "Depth", fmt.Sprintf("[b $Info]%s[/]", depth)),
			statRow(theme, "Expanded", fmt.Sprintf("[b $Primary]%s[/]", plural(countExpanded(nodes, collapsed), "folder"))),
			statRow(theme, "Loaded", fmt.Sprintf("[b $Primary]%d[/][$TextMuted] of %d folders[/]", countLoaded(nodes), a.totalDirs)),
			statRow(theme, "Selected", fmt.Sprintf("[b $Secondary]%d[/]", selected)),
			statRow(theme, "Filter", filter),
		},
	}
}

func statRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}

// selectionPanel lists the paths of the selected nodes.
type selectionPanel struct {
	app *TreeExampleApp
}

func (s selectionPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	state := s.app.treeState
	state.Selection.Get()
	state.Nodes.Get()

	var content t.Widget
	paths := state.SelectedPaths()
	if len(paths) == 0 {
		content = t.Text{
			Spans: t.ParseMarkup("[$TextMuted]Nothing selected.\nHold [b $Secondary]shift[/] and move to select.[/]", theme),
			Wrap:  t.WrapSoft,
		}
	} else {
		lines := make([]string, 0, len(paths))
		for _, p := range paths {
			if node, ok := state.NodeAtPath(p); ok {
				lines = append(lines, node.Data.Path)
			}
		}
		content = t.Text{
			Content: strings.Join(lines, "\n"),
			Style:   t.Style{ForegroundColor: theme.Secondary},
		}
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    panelStyle(theme, "Selection", false),
		Children: []t.Widget{content},
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("↑↓ jk", "$Info", "move cursor"),
			key("←→ hl", "$Info", "collapse / expand"),
			key("space", "$Info", "toggle folder"),
			key("⇧↑↓", "$Secondary", "extend selection"),
			key("enter", "$Accent", "open (OnSelect)"),
			key("/ ^t", "$Accent", "filter / fuzzy mode"),
			key("e c", "$Success", "expand / collapse all"),
			key("esc", "$Warning", "clear filter/selection"),
		},
	}
}

// loadAll returns a copy of nodes with every lazy folder's children filled in.
func loadAll(nodes []t.TreeNode[FileInfo], lazy map[string][]t.TreeNode[FileInfo]) []t.TreeNode[FileInfo] {
	out := make([]t.TreeNode[FileInfo], len(nodes))
	for i, node := range nodes {
		children := node.Children
		if children == nil && node.Data.IsDir {
			children = lazy[node.Data.Path]
			if children == nil {
				children = []t.TreeNode[FileInfo]{}
			}
		}
		out[i] = t.TreeNode[FileInfo]{Data: node.Data, Children: loadAll(children, lazy)}
	}
	return out
}

func countDirs(nodes []t.TreeNode[FileInfo]) int {
	n := 0
	for _, node := range nodes {
		if node.Data.IsDir {
			n++
		}
		n += countDirs(node.Children)
	}
	return n
}

// countLoaded counts folders whose children have been loaded.
func countLoaded(nodes []t.TreeNode[FileInfo]) int {
	n := 0
	for _, node := range nodes {
		if node.Data.IsDir && node.Children != nil {
			n++
		}
		n += countLoaded(node.Children)
	}
	return n
}

// countExpanded counts folders that are open and not hidden inside a closed one.
func countExpanded(nodes []t.TreeNode[FileInfo], collapsed map[string]bool) int {
	n := 0
	for _, node := range nodes {
		if !node.Data.IsDir || len(node.Children) == 0 || collapsed[node.Data.Path] {
			continue
		}
		n += 1 + countExpanded(node.Children, collapsed)
	}
	return n
}

// countMatches counts loaded nodes whose name matches the query, as the tree's
// MatchNode does.
func countMatches(nodes []t.TreeNode[FileInfo], query string, options t.FilterOptions) int {
	n := 0
	for _, node := range nodes {
		if t.MatchString(node.Data.Name, query, options).Matched {
			n++
		}
		n += countMatches(node.Children, query, options)
	}
	return n
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	if strings.HasSuffix(noun, "ch") {
		return fmt.Sprintf("%d %ses", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// truncateLeft shortens s to at most n cells, keeping the end.
func truncateLeft(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return "…" + string(r[len(r)-n+1:])
}

func dir(name, p string, children ...t.TreeNode[FileInfo]) t.TreeNode[FileInfo] {
	return t.TreeNode[FileInfo]{Data: FileInfo{Name: name, Path: p, IsDir: true}, Children: children}
}

// lazyDir is a folder whose children are loaded by OnExpand (nil Children).
func lazyDir(name, p string) t.TreeNode[FileInfo] {
	return t.TreeNode[FileInfo]{Data: FileInfo{Name: name, Path: p, IsDir: true}}
}

func file(name, p string) t.TreeNode[FileInfo] {
	return t.TreeNode[FileInfo]{Data: FileInfo{Name: name, Path: p}, Children: []t.TreeNode[FileInfo]{}}
}

// sampleTree returns the initial roots and the children of each lazy folder.
func sampleTree() ([]t.TreeNode[FileInfo], map[string][]t.TreeNode[FileInfo]) {
	roots := []t.TreeNode[FileInfo]{
		lazyDir("cmd", "/cmd"),
		dir("docs", "/docs",
			file("getting-started.md", "/docs/getting-started.md"),
			file("themes.md", "/docs/themes.md"),
			dir("widgets", "/docs/widgets",
				file("list.md", "/docs/widgets/list.md"),
				file("tree.md", "/docs/widgets/tree.md"),
			),
		),
		lazyDir("internal", "/internal"),
		file("go.mod", "/go.mod"),
		file("README.md", "/README.md"),
	}

	lazy := map[string][]t.TreeNode[FileInfo]{
		"/cmd": {
			dir("tree-example", "/cmd/tree-example",
				file("main.go", "/cmd/tree-example/main.go"),
			),
			lazyDir("list-example", "/cmd/list-example"),
		},
		"/cmd/list-example": {
			file("main.go", "/cmd/list-example/main.go"),
		},
		"/internal": {
			lazyDir("ui", "/internal/ui"),
			file("signals.go", "/internal/signals.go"),
			file("layout.go", "/internal/layout.go"),
		},
		"/internal/ui": {
			file("list.go", "/internal/ui/list.go"),
			file("scroll.go", "/internal/ui/scroll.go"),
			file("tree.go", "/internal/ui/tree.go"),
		},
	}

	return roots, lazy
}

func main() {
	t.SetTheme(themeNames[0])
	app := NewTreeExampleApp()
	t.RequestFocus("file-tree")
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
