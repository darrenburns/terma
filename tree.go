package terma

import (
	"fmt"
	"strconv"
	"strings"
)

// TreeNode represents a node in a tree.
// Children == nil means not loaded (lazy), Children == [] means leaf.
type TreeNode[T any] struct {
	Data     T
	Children []TreeNode[T]
}

// TreeState holds state for a Tree widget.
type TreeState[T any] struct {
	Nodes      AnySignal[[]TreeNode[T]]       // Root nodes
	CursorPath AnySignal[[]int]               // Path to cursor, e.g. [0, 2, 1]
	Collapsed  AnySignal[map[string]bool]     // Collapsed node identifiers
	Selection  AnySignal[map[string]struct{}] // Selected node identifiers

	anchorPath      []int
	viewPaths       [][]int
	viewIndexByPath map[string]int
	rowLayouts      []treeRowLayout
	nodeID          func(T) string
}

// NewTreeState creates a new TreeState with the given root nodes.
func NewTreeState[T any](roots []TreeNode[T]) *TreeState[T] {
	if roots == nil {
		roots = []TreeNode[T]{}
	}
	cursor := []int{}
	if len(roots) > 0 {
		cursor = []int{0}
	}
	return &TreeState[T]{
		Nodes:      NewAnySignal(roots),
		CursorPath: NewAnySignal(cursor),
		Collapsed:  NewAnySignal(make(map[string]bool)),
		Selection:  NewAnySignal(make(map[string]struct{})),
	}
}

// CursorUp moves the cursor to the previous visible node.
func (s *TreeState[T]) CursorUp() {
	s.moveCursor(-1)
}

// CursorDown moves the cursor to the next visible node.
func (s *TreeState[T]) CursorDown() {
	s.moveCursor(1)
}

// CursorToParent moves the cursor to its parent if possible.
func (s *TreeState[T]) CursorToParent() {
	if s == nil || !s.CursorPath.IsValid() {
		return
	}
	path := s.CursorPath.Peek()
	if len(path) <= 1 {
		return
	}
	parent := clonePath(path[:len(path)-1])
	s.CursorPath.Set(parent)
}

// CursorToFirstChild moves the cursor to the first child if it is visible.
func (s *TreeState[T]) CursorToFirstChild() {
	if s == nil || !s.CursorPath.IsValid() {
		return
	}
	cursor := s.CursorPath.Peek()
	view := s.viewPaths
	if len(view) > 0 {
		idx, ok := indexForPath(view, cursor)
		if !ok || idx+1 >= len(view) {
			return
		}
		next := view[idx+1]
		if isDirectChild(cursor, next) {
			s.CursorPath.Set(clonePath(next))
		}
		return
	}

	node, ok := s.NodeAtPath(cursor)
	if !ok || s.IsCollapsed(cursor) || len(node.Children) == 0 {
		return
	}
	child := append(clonePath(cursor), 0)
	s.CursorPath.Set(child)
}

// Toggle toggles the collapsed state for the node at the given path.
func (s *TreeState[T]) Toggle(path []int) {
	if s.IsCollapsed(path) {
		s.Expand(path)
		return
	}
	s.Collapse(path)
}

// Expand marks the node at the given path as expanded.
func (s *TreeState[T]) Expand(path []int) {
	if s == nil || !s.Collapsed.IsValid() {
		return
	}
	id := s.idForPath(path)
	if id == "" {
		return
	}
	s.Collapsed.Update(func(collapsed map[string]bool) map[string]bool {
		next := make(map[string]bool, len(collapsed))
		for k, v := range collapsed {
			next[k] = v
		}
		delete(next, id)
		return next
	})
}

// Collapse marks the node at the given path as collapsed.
func (s *TreeState[T]) Collapse(path []int) {
	if s == nil || !s.Collapsed.IsValid() {
		return
	}
	id := s.idForPath(path)
	if id == "" {
		return
	}
	s.Collapsed.Update(func(collapsed map[string]bool) map[string]bool {
		next := make(map[string]bool, len(collapsed)+1)
		for k, v := range collapsed {
			next[k] = v
		}
		next[id] = true
		return next
	})
}

// ExpandAll clears all collapsed state.
func (s *TreeState[T]) ExpandAll() {
	if s == nil || !s.Collapsed.IsValid() {
		return
	}
	s.Collapsed.Set(make(map[string]bool))
}

// CollapseAll collapses all nodes in the tree.
func (s *TreeState[T]) CollapseAll() {
	if s == nil || !s.Collapsed.IsValid() {
		return
	}
	nodes := s.Nodes.Peek()
	collapsed := make(map[string]bool)
	var walk func(nodes []TreeNode[T], path []int)
	walk = func(nodes []TreeNode[T], path []int) {
		for i, node := range nodes {
			nextPath := appendPath(path, i)
			id := s.idForNode(nextPath, node)
			if id != "" {
				collapsed[id] = true
			}
			if len(node.Children) > 0 {
				walk(node.Children, nextPath)
			}
		}
	}
	walk(nodes, nil)
	s.Collapsed.Set(collapsed)
}

// IsCollapsed returns true if the node at the given path is collapsed.
func (s *TreeState[T]) IsCollapsed(path []int) bool {
	if s == nil || !s.Collapsed.IsValid() {
		return false
	}
	id := s.idForPath(path)
	if id == "" {
		return false
	}
	return s.Collapsed.Peek()[id]
}

// SetChildren sets the children for the node at the given path.
func (s *TreeState[T]) SetChildren(path []int, children []TreeNode[T]) {
	if s == nil || !s.Nodes.IsValid() || len(path) == 0 {
		return
	}
	s.Nodes.Update(func(nodes []TreeNode[T]) []TreeNode[T] {
		updated, ok := setChildrenAtPath(nodes, path, children)
		if !ok {
			return nodes
		}
		return updated
	})
}

// ToggleSelection toggles selection for the node at the given path.
func (s *TreeState[T]) ToggleSelection(path []int) {
	if s == nil || !s.Selection.IsValid() {
		return
	}
	id := s.idForPath(path)
	if id == "" {
		return
	}
	s.Selection.Update(func(sel map[string]struct{}) map[string]struct{} {
		next := make(map[string]struct{}, len(sel))
		for k := range sel {
			next[k] = struct{}{}
		}
		if _, exists := next[id]; exists {
			delete(next, id)
		} else {
			next[id] = struct{}{}
		}
		return next
	})
}

// Select adds the node at the given path to the selection.
func (s *TreeState[T]) Select(path []int) {
	if s == nil || !s.Selection.IsValid() {
		return
	}
	id := s.idForPath(path)
	if id == "" {
		return
	}
	s.Selection.Update(func(sel map[string]struct{}) map[string]struct{} {
		next := make(map[string]struct{}, len(sel)+1)
		for k := range sel {
			next[k] = struct{}{}
		}
		next[id] = struct{}{}
		return next
	})
}

// Deselect removes the node at the given path from the selection.
func (s *TreeState[T]) Deselect(path []int) {
	if s == nil || !s.Selection.IsValid() {
		return
	}
	id := s.idForPath(path)
	if id == "" {
		return
	}
	s.Selection.Update(func(sel map[string]struct{}) map[string]struct{} {
		next := make(map[string]struct{}, len(sel))
		for k := range sel {
			next[k] = struct{}{}
		}
		delete(next, id)
		return next
	})
}

// ClearSelection clears all selected nodes.
func (s *TreeState[T]) ClearSelection() {
	if s == nil || !s.Selection.IsValid() {
		return
	}
	s.Selection.Set(make(map[string]struct{}))
}

// IsSelected returns true if the node at the given path is selected.
func (s *TreeState[T]) IsSelected(path []int) bool {
	if s == nil || !s.Selection.IsValid() {
		return false
	}
	id := s.idForPath(path)
	if id == "" {
		return false
	}
	_, ok := s.Selection.Peek()[id]
	return ok
}

// SelectedPaths returns all selected node paths in pre-order.
func (s *TreeState[T]) SelectedPaths() [][]int {
	if s == nil || !s.Selection.IsValid() {
		return nil
	}
	selected := s.Selection.Peek()
	if len(selected) == 0 {
		return nil
	}
	nodes := s.Nodes.Peek()
	paths := make([][]int, 0, len(selected))
	var walk func(nodes []TreeNode[T], path []int)
	walk = func(nodes []TreeNode[T], path []int) {
		for i, node := range nodes {
			nextPath := appendPath(path, i)
			id := s.idForNode(nextPath, node)
			if _, ok := selected[id]; ok {
				paths = append(paths, clonePath(nextPath))
			}
			if len(node.Children) > 0 {
				walk(node.Children, nextPath)
			}
		}
	}
	walk(nodes, nil)
	return paths
}

// NodeAtPath returns the node at the given path.
func (s *TreeState[T]) NodeAtPath(path []int) (TreeNode[T], bool) {
	nodes := s.Nodes.Peek()
	return nodeAtPath(nodes, path)
}

// CursorNode returns the data at the cursor position.
func (s *TreeState[T]) CursorNode() (T, bool) {
	var zero T
	if s == nil || !s.CursorPath.IsValid() {
		return zero, false
	}
	node, ok := s.NodeAtPath(s.CursorPath.Peek())
	if !ok {
		return zero, false
	}
	return node.Data, true
}

func (s *TreeState[T]) moveCursor(delta int) {
	if s == nil || !s.CursorPath.IsValid() {
		return
	}
	view := s.viewPaths
	if view == nil {
		view = s.visiblePaths()
	}
	if len(view) == 0 {
		return
	}
	cursor := s.CursorPath.Peek()
	idx, ok := indexForPath(view, cursor)
	if !ok {
		idx = 0
	}
	newIdx := clampInt(idx+delta, 0, len(view)-1)
	if newIdx == idx && ok {
		return
	}
	s.CursorPath.Set(clonePath(view[newIdx]))
}

func (s *TreeState[T]) visiblePaths() [][]int {
	nodes := s.Nodes.Peek()
	paths := make([][]int, 0)
	var walk func(nodes []TreeNode[T], path []int)
	walk = func(nodes []TreeNode[T], path []int) {
		for i, node := range nodes {
			nextPath := appendPath(path, i)
			paths = append(paths, clonePath(nextPath))
			if s.isCollapsedForNode(nextPath, node) {
				continue
			}
			if len(node.Children) > 0 {
				walk(node.Children, nextPath)
			}
		}
	}
	walk(nodes, nil)
	return paths
}

func (s *TreeState[T]) isCollapsedForNode(path []int, node TreeNode[T]) bool {
	if s == nil || !s.Collapsed.IsValid() {
		return false
	}
	id := s.idForNode(path, node)
	if id == "" {
		return false
	}
	return s.Collapsed.Peek()[id]
}

func (s *TreeState[T]) idForPath(path []int) string {
	if s == nil || len(path) == 0 {
		return ""
	}
	if s.nodeID != nil {
		if node, ok := s.NodeAtPath(path); ok {
			if id := s.nodeID(node.Data); id != "" {
				return id
			}
		}
	}
	return pathKey(path)
}

func (s *TreeState[T]) idForNode(path []int, node TreeNode[T]) string {
	if s == nil {
		return ""
	}
	if s.nodeID != nil {
		if id := s.nodeID(node.Data); id != "" {
			return id
		}
	}
	return pathKey(path)
}

func (s *TreeState[T]) setViewPaths(paths [][]int) {
	if s == nil {
		return
	}
	if paths == nil {
		s.viewPaths = nil
		s.viewIndexByPath = nil
		return
	}
	view := make([][]int, len(paths))
	indexByPath := make(map[string]int, len(paths))
	for i, path := range paths {
		clone := clonePath(path)
		view[i] = clone
		indexByPath[pathKey(clone)] = i
	}
	s.viewPaths = view
	s.viewIndexByPath = indexByPath
}

func (s *TreeState[T]) viewIndexForPath(path []int) (int, bool) {
	if s == nil || len(path) == 0 {
		return 0, false
	}
	if s.viewIndexByPath != nil {
		idx, ok := s.viewIndexByPath[pathKey(path)]
		return idx, ok
	}
	if s.viewPaths != nil {
		return indexForPath(s.viewPaths, path)
	}
	return 0, false
}

func (s *TreeState[T]) setAnchor(path []int) {
	s.anchorPath = clonePath(path)
}

func (s *TreeState[T]) clearAnchor() {
	s.anchorPath = nil
}

func (s *TreeState[T]) hasAnchor() bool {
	return len(s.anchorPath) > 0
}

func (s *TreeState[T]) getAnchor() []int {
	return clonePath(s.anchorPath)
}

// TreeNodeContext contains rendering metadata for a node.
type TreeNodeContext struct {
	Path             []int // Path to this node
	Depth            int   // Nesting level (0 = root)
	Expanded         bool  // Is this node currently expanded?
	Expandable       bool  // Can this node be expanded?
	Active           bool  // Is cursor on this node?
	Selected         bool  // Is this node selected?
	FilteredAncestor bool  // Visible due to descendant match
}

// Tree is a generic focusable widget that displays a navigable tree.
type Tree[T any] struct {
	ID                  string
	State               *TreeState[T]
	NodeID              func(data T) string
	RenderNode          func(node T, ctx TreeNodeContext) Widget
	RenderNodeWithMatch func(node T, ctx TreeNodeContext, match MatchResult) Widget
	HasChildren         func(node T) bool
	OnExpand            func(node T, path []int, setChildren func([]TreeNode[T]))
	Filter              *FilterState
	MatchNode           func(node T, query string, options FilterOptions) MatchResult
	OnSelect            func(node T)
	OnCursorChange      func(node T)
	ScrollState         *ScrollState
	Width               Dimension
	Height              Dimension
	Style               Style
	MultiSelect         bool
	Indent              int
	ExpandIndicator     string
	CollapseIndicator   string
	LeafIndicator       string
}

type treeRowLayout struct {
	y      int
	height int
}

type treeViewEntry[T any] struct {
	node       TreeNode[T]
	path       []int
	depth      int
	match      MatchResult
	ancestor   bool
	expandable bool
	expanded   bool
}

type treeContainer[T any] struct {
	Column
	tree Tree[T]
}

func (c treeContainer[T]) Build(ctx BuildContext) Widget {
	return c
}

func (c treeContainer[T]) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.tree.State == nil {
		return
	}
	count := metrics.ChildCount()
	if count == 0 {
		c.tree.State.rowLayouts = nil
		return
	}
	layouts := make([]treeRowLayout, count)
	for i := 0; i < count; i++ {
		bounds, ok := metrics.ChildBounds(i)
		if !ok {
			continue
		}
		layouts[i] = treeRowLayout{y: bounds.Y, height: bounds.Height}
	}
	c.tree.State.rowLayouts = layouts
	c.tree.scrollCursorIntoView()
}

func (c treeContainer[T]) ChildWidgets() []Widget {
	return c.Children
}

// WidgetID returns the tree widget's unique identifier.
func (t Tree[T]) WidgetID() string {
	return t.ID
}

// GetDimensions returns the width and height dimension preferences.
func (t Tree[T]) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the tree widget's style.
func (t Tree[T]) GetStyle() Style {
	return t.Style
}

// IsFocusable returns true to allow keyboard navigation.
func (t Tree[T]) IsFocusable() bool {
	return true
}

// Build builds the tree into a column of rendered nodes.
func (t Tree[T]) Build(ctx BuildContext) Widget {
	if t.State == nil {
		return Column{}
	}

	t.State.nodeID = t.NodeID

	nodes := t.State.Nodes.Get()
	query, options := filterStateValues(t.Filter)
	entries := t.buildViewEntries(nodes, query, options)

	viewPaths := make([][]int, len(entries))
	for i, entry := range entries {
		viewPaths[i] = entry.path
	}
	t.State.setViewPaths(viewPaths)

	if len(entries) == 0 {
		t.State.rowLayouts = nil
		return Column{}
	}

	cursorPath := t.State.CursorPath.Get()
	cursorPath = t.ensureCursor(viewPaths, cursorPath)

	var selection map[string]struct{}
	if t.MultiSelect {
		selection = t.State.Selection.Get()
	}

	renderNode := t.RenderNode
	renderNodeWithMatch := t.RenderNodeWithMatch
	if renderNodeWithMatch == nil && renderNode == nil {
		renderNodeWithMatch = t.themedDefaultRenderNode(ctx)
	}

	indent := t.Indent
	if indent <= 0 {
		indent = 2
	}
	expandIndicator := t.ExpandIndicator
	if expandIndicator == "" {
		expandIndicator = "▼"
	}
	collapseIndicator := t.CollapseIndicator
	if collapseIndicator == "" {
		collapseIndicator = "▶"
	}
	leafIndicator := t.LeafIndicator
	if leafIndicator == "" {
		leafIndicator = " "
	}

	children := make([]Widget, len(entries))
	for i, entry := range entries {
		active := pathsEqual(entry.path, cursorPath)
		selected := false
		if t.MultiSelect {
			if _, ok := selection[t.State.idForPath(entry.path)]; ok {
				selected = true
			}
		}
		nodeCtx := TreeNodeContext{
			Path:             clonePath(entry.path),
			Depth:            entry.depth,
			Expanded:         entry.expanded,
			Expandable:       entry.expandable,
			Active:           active,
			Selected:         selected,
			FilteredAncestor: entry.ancestor,
		}

		var nodeWidget Widget
		if renderNodeWithMatch != nil {
			nodeWidget = renderNodeWithMatch(entry.node.Data, nodeCtx, entry.match)
		} else {
			nodeWidget = renderNode(entry.node.Data, nodeCtx)
		}

		prefix := t.prefixForEntry(entry, indent, expandIndicator, collapseIndicator, leafIndicator)
		prefixStyle := t.styleForContext(ctx, nodeCtx)

		children[i] = Row{
			Spacing: 0,
			Children: []Widget{
				Text{Content: prefix, Style: prefixStyle},
				nodeWidget,
			},
		}
	}

	t.registerScrollCallbacks()

	return treeContainer[T]{
		Column: Column{
			ID:       t.ID,
			Width:    t.Width,
			Height:   t.Height,
			Style:    t.Style,
			Children: children,
		},
		tree: t,
	}
}

// OnKey handles navigation keys and selection.
func (t Tree[T]) OnKey(event KeyEvent) bool {
	if t.State == nil {
		return false
	}

	view := t.viewPaths()
	if len(view) == 0 {
		return false
	}

	cursor := t.ensureCursor(view, t.State.CursorPath.Peek())
	cursorViewIdx, _ := t.viewIndexForPath(cursor)
	lastIdx := len(view) - 1

	if t.MultiSelect {
		switch {
		case event.MatchString("shift+up", "shift+k"):
			t.handleShiftMove(-1)
			return true
		case event.MatchString("shift+down", "shift+j"):
			t.handleShiftMove(1)
			return true
		case event.MatchString("shift+home"):
			t.handleShiftMoveTo(0)
			return true
		case event.MatchString("shift+end"):
			t.handleShiftMoveTo(lastIdx)
			return true
		}
	}

	switch {
	case event.MatchString("enter"):
		if t.OnSelect != nil {
			if node, ok := t.State.CursorNode(); ok {
				t.OnSelect(node)
			}
		}
		return true

	case event.MatchString("up", "k"):
		if cursorViewIdx == 0 {
			return false
		}
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.clearAnchor()
		}
		t.setCursorToViewIndex(cursorViewIdx - 1)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("down", "j"):
		if cursorViewIdx >= lastIdx {
			return false
		}
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.clearAnchor()
		}
		t.setCursorToViewIndex(cursorViewIdx + 1)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("home", "g"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.clearAnchor()
		}
		t.setCursorToViewIndex(0)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("end", "G"):
		if t.MultiSelect {
			t.State.ClearSelection()
			t.State.clearAnchor()
		}
		t.setCursorToViewIndex(lastIdx)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true

	case event.MatchString("left", "h"):
		node, ok := t.State.NodeAtPath(cursor)
		if !ok {
			return false
		}
		if t.nodeExpanded(node, cursor) {
			t.State.Collapse(cursor)
			return true
		}
		if len(cursor) > 1 {
			t.State.CursorPath.Set(clonePath(cursor[:len(cursor)-1]))
			t.scrollCursorIntoView()
			t.notifyCursorChange()
			return true
		}
		return false

	case event.MatchString("right", "l"):
		node, ok := t.State.NodeAtPath(cursor)
		if !ok {
			return false
		}
		if t.nodeExpandable(node) && !t.nodeExpanded(node, cursor) {
			t.expandNode(node, cursor)
			return true
		}
		if child, ok := t.firstChildPath(cursor, view); ok {
			t.State.CursorPath.Set(clonePath(child))
			t.scrollCursorIntoView()
			t.notifyCursorChange()
			return true
		}
		return false

	case event.MatchString("space", " "):
		node, ok := t.State.NodeAtPath(cursor)
		if !ok {
			return false
		}
		if t.nodeExpandable(node) {
			if t.nodeExpanded(node, cursor) {
				t.State.Collapse(cursor)
			} else {
				t.expandNode(node, cursor)
			}
			return true
		}
		return false
	}

	return false
}

func (t Tree[T]) themedDefaultRenderNode(ctx BuildContext) func(node T, nodeCtx TreeNodeContext, match MatchResult) Widget {
	theme := ctx.Theme()
	highlight := SpanStyle{
		Underline:      UnderlineSingle,
		UnderlineColor: theme.Accent,
		Background:     theme.Accent.WithAlpha(0.25),
	}
	return func(node T, nodeCtx TreeNodeContext, match MatchResult) Widget {
		content := fmt.Sprintf("%v", node)
		style := t.styleForContext(ctx, nodeCtx)
		if match.Matched && len(match.Ranges) > 0 {
			spans := HighlightSpans(content, match.Ranges, highlight)
			return Text{
				Spans: spans,
				Style: style,
				Width: Flex(1),
			}
		}
		return Text{
			Content: content,
			Style:   style,
			Width:   Flex(1),
		}
	}
}

func (t Tree[T]) styleForContext(ctx BuildContext, nodeCtx TreeNodeContext) Style {
	theme := ctx.Theme()
	style := Style{ForegroundColor: theme.Text}
	if nodeCtx.FilteredAncestor {
		style.ForegroundColor = theme.TextMuted
	}
	if nodeCtx.Selected {
		style.ForegroundColor = theme.Secondary
	}
	if nodeCtx.Active {
		style.ForegroundColor = theme.Accent
	}
	return style
}

func (t Tree[T]) prefixForEntry(entry treeViewEntry[T], indent int, expandedIndicator, collapsedIndicator, leafIndicator string) string {
	indentation := ""
	if indent > 0 && entry.depth > 0 {
		indentation = strings.Repeat(" ", indent*entry.depth)
	}
	indicator := leafIndicator
	if entry.expandable {
		if entry.expanded {
			indicator = expandedIndicator
		} else {
			indicator = collapsedIndicator
		}
	}
	return indentation + indicator + " "
}

func (t Tree[T]) buildViewEntries(nodes []TreeNode[T], query string, options FilterOptions) []treeViewEntry[T] {
	matchNode := t.MatchNode
	if matchNode == nil {
		matchNode = defaultTreeMatchNode[T]
	}
	if query == "" {
		return t.flattenVisible(nodes, nil, 0)
	}
	entries, _ := t.filterVisible(nodes, nil, 0, query, options, matchNode)
	return entries
}

func (t Tree[T]) flattenVisible(nodes []TreeNode[T], path []int, depth int) []treeViewEntry[T] {
	entries := make([]treeViewEntry[T], 0)
	for i, node := range nodes {
		nextPath := appendPath(path, i)
		expandable := t.nodeExpandable(node)
		expanded := t.nodeExpanded(node, nextPath)
		entries = append(entries, treeViewEntry[T]{
			node:       node,
			path:       nextPath,
			depth:      depth,
			match:      MatchResult{Matched: true},
			expandable: expandable,
			expanded:   expanded,
		})
		if expanded && len(node.Children) > 0 {
			entries = append(entries, t.flattenVisible(node.Children, nextPath, depth+1)...)
		}
	}
	return entries
}

func (t Tree[T]) filterVisible(nodes []TreeNode[T], path []int, depth int, query string, options FilterOptions, matchNode func(node T, query string, options FilterOptions) MatchResult) ([]treeViewEntry[T], bool) {
	entries := make([]treeViewEntry[T], 0)
	hasMatch := false
	for i, node := range nodes {
		nextPath := appendPath(path, i)
		match := matchNode(node.Data, query, options)
		childEntries, childHasMatch := t.filterVisible(node.Children, nextPath, depth+1, query, options, matchNode)
		if match.Matched || childHasMatch {
			expandable := t.nodeExpandable(node)
			entries = append(entries, treeViewEntry[T]{
				node:       node,
				path:       nextPath,
				depth:      depth,
				match:      match,
				ancestor:   childHasMatch && !match.Matched,
				expandable: expandable,
				expanded:   childHasMatch,
			})
			if childHasMatch {
				entries = append(entries, childEntries...)
			}
			hasMatch = true
		}
	}
	return entries, hasMatch
}

func (t Tree[T]) nodeExpandable(node TreeNode[T]) bool {
	if node.Children != nil {
		return len(node.Children) > 0
	}
	if t.HasChildren != nil {
		return t.HasChildren(node.Data)
	}
	return false
}

func (t Tree[T]) nodeExpanded(node TreeNode[T], path []int) bool {
	if !t.nodeExpandable(node) {
		return false
	}
	if node.Children == nil {
		return false
	}
	if t.State != nil && t.State.IsCollapsed(path) {
		return false
	}
	return true
}

func (t Tree[T]) expandNode(node TreeNode[T], path []int) {
	if t.State == nil {
		return
	}
	if !t.nodeExpandable(node) {
		return
	}
	t.State.Expand(path)
	if node.Children == nil && t.OnExpand != nil {
		pathCopy := clonePath(path)
		t.OnExpand(node.Data, pathCopy, func(children []TreeNode[T]) {
			t.State.SetChildren(pathCopy, children)
		})
	}
}

func (t Tree[T]) ensureCursor(view [][]int, cursor []int) []int {
	if len(view) == 0 {
		return cursor
	}
	if len(cursor) == 0 {
		next := clonePath(view[0])
		t.State.CursorPath.Set(next)
		return next
	}
	if _, ok := t.viewIndexForPath(cursor); !ok {
		next := clonePath(view[0])
		t.State.CursorPath.Set(next)
		return next
	}
	return cursor
}

func (t Tree[T]) viewPaths() [][]int {
	if t.State == nil {
		return nil
	}
	if t.State.viewPaths != nil {
		return t.State.viewPaths
	}
	entries := t.flattenVisible(t.State.Nodes.Peek(), nil, 0)
	view := make([][]int, len(entries))
	for i, entry := range entries {
		view[i] = entry.path
	}
	return view
}

func (t Tree[T]) viewIndexForPath(path []int) (int, bool) {
	if t.State == nil {
		return 0, false
	}
	if t.State.viewPaths != nil {
		return t.State.viewIndexForPath(path)
	}
	return indexForPath(t.viewPaths(), path)
}

func (t Tree[T]) setCursorToViewIndex(viewIdx int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	viewIdx = clampInt(viewIdx, 0, len(view)-1)
	t.State.CursorPath.Set(clonePath(view[viewIdx]))
}

func (t Tree[T]) firstChildPath(parent []int, view [][]int) ([]int, bool) {
	idx, ok := t.viewIndexForPath(parent)
	if !ok || idx+1 >= len(view) {
		return nil, false
	}
	next := view[idx+1]
	if isDirectChild(parent, next) {
		return next, true
	}
	return nil, false
}

func (t Tree[T]) handleShiftMove(delta int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	cursor := t.ensureCursor(view, t.State.CursorPath.Peek())
	cursorIdx, ok := t.viewIndexForPath(cursor)
	if !ok {
		cursorIdx = 0
	}
	if !t.State.hasAnchor() {
		t.State.setAnchor(cursor)
	}
	newIdx := clampInt(cursorIdx+delta, 0, len(view)-1)
	newCursor := view[newIdx]
	t.State.CursorPath.Set(clonePath(newCursor))
	t.selectViewRange(t.State.getAnchor(), newCursor)
	t.scrollCursorIntoView()
}

func (t Tree[T]) handleShiftMoveTo(targetIdx int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	cursor := t.ensureCursor(view, t.State.CursorPath.Peek())
	if !t.State.hasAnchor() {
		t.State.setAnchor(cursor)
	}
	targetIdx = clampInt(targetIdx, 0, len(view)-1)
	newCursor := view[targetIdx]
	t.State.CursorPath.Set(clonePath(newCursor))
	t.selectViewRange(t.State.getAnchor(), newCursor)
	t.scrollCursorIntoView()
}

func (t Tree[T]) selectViewRange(anchor, cursor []int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	anchorIdx, ok := t.viewIndexForPath(anchor)
	if !ok {
		anchorIdx = 0
	}
	cursorIdx, ok := t.viewIndexForPath(cursor)
	if !ok {
		cursorIdx = anchorIdx
	}
	if anchorIdx > cursorIdx {
		anchorIdx, cursorIdx = cursorIdx, anchorIdx
	}
	sel := make(map[string]struct{}, cursorIdx-anchorIdx+1)
	for i := anchorIdx; i <= cursorIdx; i++ {
		id := t.State.idForPath(view[i])
		if id != "" {
			sel[id] = struct{}{}
		}
	}
	t.State.Selection.Set(sel)
}

func (t Tree[T]) scrollCursorIntoView() {
	if t.ScrollState == nil || t.State == nil {
		return
	}
	cursor := t.State.CursorPath.Peek()
	rowY, rowHeight, ok := t.getRowLayout(cursor)
	if !ok {
		if viewIdx, ok := t.viewIndexForPath(cursor); ok {
			rowHeight = 1
			rowY = viewIdx * rowHeight
		} else {
			return
		}
	}
	t.ScrollState.ScrollToView(rowY, rowHeight)
}

func (t Tree[T]) getRowLayout(path []int) (y, height int, ok bool) {
	if t.State == nil {
		return 0, 0, false
	}
	viewIdx, ok := t.viewIndexForPath(path)
	if !ok {
		return 0, 0, false
	}
	if viewIdx < 0 || viewIdx >= len(t.State.rowLayouts) {
		return 0, 0, false
	}
	layout := t.State.rowLayouts[viewIdx]
	if layout.height <= 0 {
		return 0, 0, false
	}
	return layout.y, layout.height, true
}

func (t Tree[T]) registerScrollCallbacks() {
	if t.ScrollState == nil {
		return
	}
	t.ScrollState.OnScrollUp = func(lines int) bool {
		t.moveCursorUp(lines)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true
	}
	t.ScrollState.OnScrollDown = func(lines int) bool {
		t.moveCursorDown(lines)
		t.scrollCursorIntoView()
		t.notifyCursorChange()
		return true
	}
}

func (t Tree[T]) moveCursorUp(count int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	cursor := t.State.CursorPath.Peek()
	cursorIdx, ok := t.viewIndexForPath(cursor)
	if !ok {
		cursorIdx = 0
	}
	newIdx := clampInt(cursorIdx-count, 0, len(view)-1)
	t.State.CursorPath.Set(clonePath(view[newIdx]))
}

func (t Tree[T]) moveCursorDown(count int) {
	if t.State == nil {
		return
	}
	view := t.viewPaths()
	if len(view) == 0 {
		return
	}
	cursor := t.State.CursorPath.Peek()
	cursorIdx, ok := t.viewIndexForPath(cursor)
	if !ok {
		cursorIdx = 0
	}
	newIdx := clampInt(cursorIdx+count, 0, len(view)-1)
	t.State.CursorPath.Set(clonePath(view[newIdx]))
}

func (t Tree[T]) notifyCursorChange() {
	if t.OnCursorChange == nil || t.State == nil {
		return
	}
	if node, ok := t.State.CursorNode(); ok {
		t.OnCursorChange(node)
	}
}

func defaultTreeMatchNode[T any](node T, query string, options FilterOptions) MatchResult {
	return MatchString(fmt.Sprintf("%v", node), query, options)
}

func nodeAtPath[T any](nodes []TreeNode[T], path []int) (TreeNode[T], bool) {
	var zero TreeNode[T]
	if len(path) == 0 {
		return zero, false
	}
	idx := path[0]
	if idx < 0 || idx >= len(nodes) {
		return zero, false
	}
	node := nodes[idx]
	if len(path) == 1 {
		return node, true
	}
	if len(node.Children) == 0 {
		return zero, false
	}
	return nodeAtPath(node.Children, path[1:])
}

func setChildrenAtPath[T any](nodes []TreeNode[T], path []int, children []TreeNode[T]) ([]TreeNode[T], bool) {
	if len(path) == 0 {
		return nodes, false
	}
	idx := path[0]
	if idx < 0 || idx >= len(nodes) {
		return nodes, false
	}
	updated := make([]TreeNode[T], len(nodes))
	copy(updated, nodes)
	node := updated[idx]
	if len(path) == 1 {
		node.Children = children
		updated[idx] = node
		return updated, true
	}
	if len(node.Children) == 0 {
		return nodes, false
	}
	childNodes, ok := setChildrenAtPath(node.Children, path[1:], children)
	if !ok {
		return nodes, false
	}
	node.Children = childNodes
	updated[idx] = node
	return updated, true
}

func clonePath(path []int) []int {
	if len(path) == 0 {
		return nil
	}
	clone := make([]int, len(path))
	copy(clone, path)
	return clone
}

func appendPath(path []int, idx int) []int {
	next := make([]int, len(path)+1)
	copy(next, path)
	next[len(path)] = idx
	return next
}

func pathKey(path []int) string {
	if len(path) == 0 {
		return ""
	}
	var b strings.Builder
	for i, idx := range path {
		if i > 0 {
			b.WriteByte('/')
		}
		b.WriteString(strconv.Itoa(idx))
	}
	return b.String()
}

func pathsEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func isDirectChild(parent, child []int) bool {
	if len(child) != len(parent)+1 {
		return false
	}
	for i := range parent {
		if parent[i] != child[i] {
			return false
		}
	}
	return true
}

func indexForPath(view [][]int, path []int) (int, bool) {
	for i, candidate := range view {
		if pathsEqual(candidate, path) {
			return i, true
		}
	}
	return 0, false
}
