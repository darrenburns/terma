package terma

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/theory/jsonpath"
	"github.com/theory/jsonpath/spec"
)

// JSONNode is a value in a JSON document. Treat Value and its children as read-only.
type JSONNode struct {
	Path  string // RFC 9535 normalized JSONPath, e.g. $['users'][2]['name'].
	Label string // JSON-quoted object key, array index in brackets, or $ for the root.
	Value any    // map[string]any, []any, string, json.Number, bool, or nil.

	ancestor bool // Present only to provide context for a matched descendant.
}

// JSONViewerState owns a decoded document and its filtered tree. Create it once,
// outside Build. Use SetJSON and SetQuery from event handlers to change the view.
type JSONViewerState struct {
	// Tree exposes navigation, expansion, and selection. Its Nodes are the current
	// filtered view; use SetJSON rather than modifying them or their values.
	Tree *TreeState[JSONNode]

	root    TreeNode[JSONNode]
	path    *jsonpath.Path
	query   Signal[string]
	matches Signal[int]
	scroll  *ScrollState
}

// NewJSONViewerState decodes a single JSON value, preserving number lexemes.
// Invalid JSON returns an error and no state. Objects are sorted by key.
func NewJSONViewerState(data []byte) (*JSONViewerState, error) {
	s := &JSONViewerState{
		Tree:    NewTreeState[JSONNode](nil),
		query:   NewSignal(""),
		matches: NewSignal(0),
		scroll:  NewScrollState(),
	}
	s.Tree.nodeID = func(node JSONNode) string { return node.Path }
	if err := s.SetJSON(data); err != nil {
		return nil, err
	}
	return s, nil
}

// SetJSON replaces the document and reapplies the current query. Expansion,
// cursor, selection, and scrolling reset. Invalid JSON leaves the state intact.
func (s *JSONViewerState) SetJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return fmt.Errorf("JSONViewer: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("JSONViewer: expected a single JSON value")
	}
	s.root = newJSONTreeNode(value, "$", nil)
	s.Tree.ExpandAll()
	s.applyQuery()
	return nil
}

// Query returns the last valid query, subscribing to changes.
func (s *JSONViewerState) Query() string { return s.query.Get() }

// MatchCount returns the number of distinct nodes selected by the query,
// subscribing to changes. An empty query selects the root (one match).
func (s *JSONViewerState) MatchCount() int { return s.matches.Get() }

// SetQuery filters using RFC 9535 JSONPath. Empty or whitespace-only queries
// restore the whole document. Matches retain their ancestors and entire subtrees;
// array indices and paths always refer to the original document. Ancestors of
// matches expand, but can subsequently be collapsed using normal tree controls.
// Applying a different query resets cursor, selection, and scroll position.
// Invalid queries return an error without changing the last valid view.
func (s *JSONViewerState) SetQuery(query string) error {
	query = strings.TrimSpace(query)
	if query == s.query.Peek() {
		return nil
	}
	var path *jsonpath.Path
	if query != "" {
		var err error
		path, err = jsonpath.Parse(query)
		if err != nil {
			return fmt.Errorf("JSONViewer: %w", err)
		}
	}
	s.path = path
	s.query.Set(query)
	s.applyQuery()
	return nil
}

func (s *JSONViewerState) applyQuery() {
	nodes := []TreeNode[JSONNode]{s.root}
	count := 1
	if s.path != nil {
		matched := make(map[string]bool)
		ancestors := make(map[string]bool)
		for _, node := range s.path.SelectLocated(s.root.Data.Value) {
			matched[node.Path.String()] = true
			for i := 0; i < len(node.Path); i++ {
				ancestors[node.Path[:i].String()] = true
			}
		}
		count = len(matched)
		nodes = filterJSONNodes(nodes, matched, ancestors)
		collapsed := maps.Clone(s.Tree.Collapsed.Peek())
		for path := range ancestors {
			delete(collapsed, path)
		}
		s.Tree.Collapsed.Set(collapsed)
	}
	s.Tree.Nodes.Set(nodes)
	s.Tree.ClearSelection()
	s.Tree.ClearAnchor()
	cursor := []int{}
	if len(nodes) > 0 {
		cursor = []int{0}
	}
	s.Tree.CursorPath.Set(cursor)
	s.scroll.SetOffset(0)
	s.scroll.SetOffsetX(0)
	s.matches.Set(count)
}

func newJSONTreeNode(value any, label string, path spec.NormalizedPath) TreeNode[JSONNode] {
	node := TreeNode[JSONNode]{
		Data:     JSONNode{Path: path.String(), Label: label, Value: value},
		Children: []TreeNode[JSONNode]{},
	}
	switch value := value.(type) {
	case map[string]any:
		for _, key := range slices.Sorted(maps.Keys(value)) {
			name, _ := json.Marshal(key)
			node.Children = append(node.Children, newJSONTreeNode(value[key], string(name), append(path, spec.Name(key))))
		}
	case []any:
		for i, child := range value {
			node.Children = append(node.Children, newJSONTreeNode(child, fmt.Sprintf("[%d]", i), append(path, spec.Index(i))))
		}
	}
	return node
}

func filterJSONNodes(nodes []TreeNode[JSONNode], matched, ancestors map[string]bool) []TreeNode[JSONNode] {
	filtered := make([]TreeNode[JSONNode], 0, len(nodes))
	for _, node := range nodes {
		if matched[node.Data.Path] {
			filtered = append(filtered, node)
		} else if ancestors[node.Data.Path] {
			node.Data.ancestor = true
			node.Children = filterJSONNodes(node.Children, matched, ancestors)
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// JSONViewer displays a JSON document as a syntax-colored, scrollable Tree.
// Supply a State created with NewJSONViewerState. JSONPath input is deliberately
// separate: connect a TextInput's OnChange to State.SetQuery and display its error.
// Standard Tree keyboard and mouse navigation, expansion, and activation apply.
type JSONViewer struct {
	ID              string
	State           *JSONViewerState
	Style           Style
	DisableFocus    bool
	MultiSelect     bool
	ActivateOnClick bool
	OnSelect        func(node JSONNode, selected []JSONNode)
	OnCursorChange  func(node JSONNode)
}

// WidgetID returns the ID used to focus the viewer's tree.
func (v JSONViewer) WidgetID() string { return v.ID }

// GetStyle exposes the viewport dimensions and styling to parent layouts.
func (v JSONViewer) GetStyle() Style { return v.Style }

// Build composes the viewer from Tree and Scrollable.
func (v JSONViewer) Build(ctx BuildContext) Widget {
	if v.State == nil {
		return EmptyWidget{}
	}
	theme := ctx.Theme()
	tree := Tree[JSONNode]{
		ID:              widgetIdentity(v, ctx),
		State:           v.State.Tree,
		NodeID:          func(node JSONNode) string { return node.Path },
		ScrollState:     v.State.scroll,
		Style:           Style{Width: Flex(1)},
		DisableFocus:    v.DisableFocus,
		MultiSelect:     v.MultiSelect,
		ActivateOnClick: v.ActivateOnClick,
		OnSelect:        v.OnSelect,
		OnCursorChange:  v.OnCursorChange,
	}
	focused := ctx.IsFocused(tree)
	tree.RenderNode = func(node JSONNode, nodeCtx TreeNodeContext) Widget {
		nodeCtx.FilteredAncestor = node.ancestor
		style := styleForTreeNodeContext(theme, nodeCtx, focused)
		style.Width = Flex(1)
		value, color := jsonNodeText(node.Value, theme)
		keyColor := theme.Primary
		if node.ancestor {
			keyColor, color = theme.TextMuted, theme.TextMuted
		}
		if nodeCtx.Active && focused {
			keyColor, color = theme.SelectionText, theme.SelectionText
		}
		return Text{Style: style, Spans: []Span{
			{Text: node.Label + ": ", Style: SpanStyle{Foreground: keyColor}},
			{Text: value, Style: SpanStyle{Foreground: color}},
		}}
	}
	var child Widget = tree
	if len(v.State.Tree.Nodes.Get()) == 0 {
		// Keep the empty tree focusable, so clearing the query restores navigation.
		child = Column{Style: Style{Width: Flex(1)}, Children: []Widget{tree, Text{
			Content: "No JSONPath matches",
			Style:   Style{ForegroundColor: theme.TextMuted},
		}}}
	}
	return Scrollable{
		State:               v.State.scroll,
		Style:               v.Style,
		ScrollbarThumbColor: theme.ScrollbarThumb,
		ScrollbarTrackColor: theme.ScrollbarTrack,
		Child:               child,
	}
}

func jsonNodeText(value any, theme ThemeData) (string, Color) {
	color := theme.TextMuted
	switch value := value.(type) {
	case map[string]any:
		if len(value) == 0 {
			return "{}", color
		}
		return fmt.Sprintf("{%d keys}", len(value)), color
	case []any:
		if len(value) == 0 {
			return "[]", color
		}
		return fmt.Sprintf("[%d items]", len(value)), color
	case string:
		color = theme.Success
	case json.Number:
		color = theme.Warning
	case bool:
		color = theme.Accent
	}
	text, _ := json.Marshal(value)
	return string(text), color
}
