package terma

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DirectoryEntry represents a filesystem entry used by DirectoryTree.
type DirectoryEntry struct {
	Name  string
	Path  string
	IsDir bool
	Err   error
}

// DirectoryTree is a utility widget that renders a filesystem tree using Tree.
type DirectoryTree struct {
	Tree[DirectoryEntry]

	// ReadDir returns the direct children of path. Default uses os.ReadDir.
	ReadDir func(path string) ([]DirectoryEntry, error)
	// EagerLoad preloads all directory children (recursively) in the background.
	// Loaded nodes remain collapsed until expanded.
	EagerLoad bool
	// IncludeHidden controls whether entries starting with "." are included.
	IncludeHidden bool
	// Sort controls entry ordering; default sorts directories first by name.
	Sort func([]DirectoryEntry)
}

// NewDirectoryTreeState creates a TreeState for a single root path.
func NewDirectoryTreeState(root string) *TreeState[DirectoryEntry] {
	return NewDirectoryTreeStateWithRoots([]string{root})
}

// NewDirectoryTreeStateWithRoots creates a TreeState for multiple root paths.
func NewDirectoryTreeStateWithRoots(roots []string) *TreeState[DirectoryEntry] {
	nodes := make([]TreeNode[DirectoryEntry], 0, len(roots))
	for _, root := range roots {
		nodes = append(nodes, directoryTreeRootNode(root))
	}
	return NewTreeState(nodes)
}

// Build renders a Tree[DirectoryEntry] with filesystem-aware defaults.
func (d DirectoryTree) Build(ctx BuildContext) Widget {
	tree := d.resolvedTree()
	if tree.State == nil {
		return Column{}
	}

	if tree.RenderNode == nil && tree.RenderNodeWithMatch == nil {
		tree.RenderNodeWithMatch = d.defaultRenderNode(ctx, tree)
	}

	return tree.Build(ctx)
}

// WidgetID returns the directory tree's unique identifier.
func (d DirectoryTree) WidgetID() string {
	return d.Tree.ID
}

// IsFocusable returns true to allow keyboard navigation.
func (d DirectoryTree) IsFocusable() bool {
	return d.Tree.IsFocusable()
}

// OnKey handles keys not covered by declarative keybindings.
func (d DirectoryTree) OnKey(event KeyEvent) bool {
	return d.resolvedTree().OnKey(event)
}

// Keybinds returns the declarative keybindings for this directory tree.
func (d DirectoryTree) Keybinds() []Keybind {
	return d.resolvedTree().Keybinds()
}

func (d DirectoryTree) resolvedTree() Tree[DirectoryEntry] {
	tree := d.Tree
	if tree.NodeID == nil {
		tree.NodeID = func(entry DirectoryEntry) string {
			return entry.Path
		}
	}
	if tree.MatchNode == nil {
		tree.MatchNode = func(entry DirectoryEntry, query string, options FilterOptions) MatchResult {
			return MatchString(entry.Name, query, options)
		}
	}
	if tree.HasChildren == nil {
		tree.HasChildren = func(entry DirectoryEntry) bool {
			return entry.IsDir
		}
	}
	if tree.OnExpand == nil {
		tree.OnExpand = func(entry DirectoryEntry, _ []int, setChildren func([]TreeNode[DirectoryEntry])) {
			d.loadChildren(entry, setChildren)
		}
	}
	if tree.State != nil {
		tree.State.nodeID = tree.NodeID
		if d.EagerLoad {
			tree.State.eagerLoadOnce.Do(func() {
				d.eagerLoad(tree.State)
			})
		}
	}
	return tree
}

func (d DirectoryTree) defaultRenderNode(ctx BuildContext, tree Tree[DirectoryEntry]) func(DirectoryEntry, TreeNodeContext, MatchResult) Widget {
	theme := ctx.Theme()
	widgetFocused := ctx.IsFocused(tree)
	highlight := MatchHighlightStyle(theme)

	return func(entry DirectoryEntry, nodeCtx TreeNodeContext, match MatchResult) Widget {
		style := directoryTreeNodeStyle(theme, nodeCtx, widgetFocused, entry.Err != nil)
		style.Width = Flex(1)

		label := entry.Name
		if label == "" {
			label = entry.Path
		}
		if entry.Err != nil {
			if label == "" {
				label = entry.Err.Error()
			} else {
				label = fmt.Sprintf("%s (%s)", label, entry.Err.Error())
			}
		}

		icon := "F"
		if entry.Err != nil {
			icon = "!"
		} else if entry.IsDir {
			icon = "D"
		}

		spans := []Span{{Text: icon + " "}}
		if match.Matched && len(match.Ranges) > 0 {
			spans = append(spans, HighlightSpans(label, match.Ranges, highlight)...)
		} else {
			spans = append(spans, Span{Text: label})
		}

		return Text{Spans: spans, Style: style}
	}
}

func directoryTreeNodeStyle(theme ThemeData, nodeCtx TreeNodeContext, widgetFocused bool, isError bool) Style {
	style := Style{ForegroundColor: theme.Text}
	if nodeCtx.FilteredAncestor {
		style.ForegroundColor = theme.TextMuted
	}
	if isError {
		style.ForegroundColor = theme.Error
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

func (d DirectoryTree) loadChildren(entry DirectoryEntry, setChildren func([]TreeNode[DirectoryEntry])) {
	go func(entry DirectoryEntry) {
		setChildren(d.readChildren(entry))
	}(entry)
}

func (d DirectoryTree) eagerLoad(state *TreeState[DirectoryEntry]) {
	if state == nil {
		return
	}
	go func() {
		roots := state.Nodes.Peek()
		if len(roots) == 0 {
			return
		}
		type queueItem struct {
			path []int
			node TreeNode[DirectoryEntry]
		}
		queue := make([]queueItem, 0, len(roots))
		for i, node := range roots {
			queue = append(queue, queueItem{path: []int{i}, node: node})
		}
		for len(queue) > 0 {
			item := queue[0]
			queue = queue[1:]

			node := item.node
			if node.Data.Err != nil || !node.Data.IsDir {
				continue
			}
			if node.Children == nil {
				children := d.readChildren(node.Data)
				state.SetChildren(item.path, children)
				node.Children = children
				if len(children) > 0 {
					state.Collapse(item.path)
				}
			}
			for i, child := range node.Children {
				queue = append(queue, queueItem{
					path: appendPath(item.path, i),
					node: child,
				})
			}
		}
	}()
}

func (d DirectoryTree) readChildren(entry DirectoryEntry) []TreeNode[DirectoryEntry] {
	readDir := d.ReadDir
	if readDir == nil {
		readDir = defaultDirectoryReadDir
	}
	parentPath := entry.Path
	if parentPath == "" {
		parentPath = "."
	}

	children, err := readDir(parentPath)
	if err != nil {
		return []TreeNode[DirectoryEntry]{directoryTreeErrorNode(parentPath, err)}
	}

	normalized := make([]DirectoryEntry, 0, len(children))
	for _, child := range children {
		child = normalizeDirectoryEntry(parentPath, child)
		if !d.IncludeHidden && strings.HasPrefix(child.Name, ".") {
			continue
		}
		normalized = append(normalized, child)
	}

	if d.Sort != nil {
		d.Sort(normalized)
	} else {
		defaultDirectorySort(normalized)
	}

	nodes := make([]TreeNode[DirectoryEntry], 0, len(normalized))
	for _, child := range normalized {
		nodes = append(nodes, directoryTreeNode(child))
	}
	return nodes
}

func normalizeDirectoryEntry(parentPath string, entry DirectoryEntry) DirectoryEntry {
	if entry.Path == "" && entry.Name != "" {
		entry.Path = filepath.Join(parentPath, entry.Name)
	}
	if entry.Path != "" {
		entry.Path = filepath.Clean(entry.Path)
	}
	if entry.Name == "" && entry.Path != "" {
		entry.Name = filepath.Base(entry.Path)
	}
	return entry
}

func defaultDirectoryReadDir(path string) ([]DirectoryEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	result := make([]DirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, DirectoryEntry{
			Name:  entry.Name(),
			Path:  filepath.Join(path, entry.Name()),
			IsDir: entry.IsDir(),
		})
	}
	return result, nil
}

func defaultDirectorySort(entries []DirectoryEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		left := entries[i]
		right := entries[j]
		if left.IsDir != right.IsDir {
			return left.IsDir && !right.IsDir
		}
		leftKey := directoryEntrySortKey(left)
		rightKey := directoryEntrySortKey(right)
		if leftKey == rightKey {
			return left.Path < right.Path
		}
		return leftKey < rightKey
	})
}

func directoryEntrySortKey(entry DirectoryEntry) string {
	name := entry.Name
	if name == "" {
		name = entry.Path
	}
	return strings.ToLower(name)
}

func directoryTreeRootNode(path string) TreeNode[DirectoryEntry] {
	entry := directoryEntryFromPath(path)
	return directoryTreeNode(entry)
}

func directoryEntryFromPath(path string) DirectoryEntry {
	if path == "" {
		path = "."
	}
	clean := filepath.Clean(path)
	name := filepath.Base(clean)
	if name == "." {
		name = clean
	}

	entry := DirectoryEntry{
		Name: name,
		Path: clean,
	}

	info, err := os.Stat(clean)
	if err != nil {
		entry.Err = err
		return entry
	}
	entry.IsDir = info.IsDir()
	return entry
}

func directoryTreeNode(entry DirectoryEntry) TreeNode[DirectoryEntry] {
	children := []TreeNode[DirectoryEntry]{}
	if entry.IsDir && entry.Err == nil {
		children = nil
	}
	return TreeNode[DirectoryEntry]{
		Data:     entry,
		Children: children,
	}
}

func directoryTreeErrorNode(path string, err error) TreeNode[DirectoryEntry] {
	return TreeNode[DirectoryEntry]{
		Data: DirectoryEntry{
			Name: err.Error(),
			Path: path,
			Err:  err,
		},
		Children: []TreeNode[DirectoryEntry]{},
	}
}
