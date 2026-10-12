package terma_test

import (
	"encoding/json"
	"testing"
	"time"

	t "github.com/darrenburns/terma"
	"github.com/stretchr/testify/require"
)

const viewerJSON = `{
	"items": [
		{"name": "same", "price": 8},
		{"name": "other", "price": 14},
		{"name": "same", "price": 9.5}
	],
	"nothing": null,
	"a.b'\\\n": {"": true}
}`

func jsonPaths(nodes []t.TreeNode[t.JSONNode]) []string {
	paths := []string{}
	for _, node := range nodes {
		paths = append(paths, node.Data.Path)
		paths = append(paths, jsonPaths(node.Children)...)
	}
	return paths
}

func TestJSONViewerJSONPath(tst *testing.T) {
	for _, tc := range []struct {
		query string
		count int
		paths []string
	}{
		{`$.items[?@.price < 10].name`, 2, []string{
			`$`, `$['items']`, `$['items'][0]`, `$['items'][0]['name']`, `$['items'][2]`, `$['items'][2]['name']`,
		}},
		{`$.items[1]`, 1, []string{
			`$`, `$['items']`, `$['items'][1]`, `$['items'][1]['name']`, `$['items'][1]['price']`,
		}},
		{`$..name`, 3, []string{
			`$`, `$['items']`, `$['items'][0]`, `$['items'][0]['name']`, `$['items'][1]`, `$['items'][1]['name']`, `$['items'][2]`, `$['items'][2]['name']`,
		}},
		{`$.items[1:].price`, 2, []string{
			`$`, `$['items']`, `$['items'][1]`, `$['items'][1]['price']`, `$['items'][2]`, `$['items'][2]['price']`,
		}},
		{`$.items[-1,-1].name`, 1, []string{
			`$`, `$['items']`, `$['items'][2]`, `$['items'][2]['name']`,
		}},
		{`$.items[*].price`, 3, []string{
			`$`, `$['items']`, `$['items'][0]`, `$['items'][0]['price']`, `$['items'][1]`, `$['items'][1]['price']`, `$['items'][2]`, `$['items'][2]['price']`,
		}},
		{`$.nothing`, 1, []string{`$`, `$['nothing']`}},
		{`$["a.b'\\\n"][""]`, 1, []string{`$`, `$['a.b\'\\\n']`, `$['a.b\'\\\n']['']`}},
		{`$.missing`, 0, []string{}},
	} {
		tst.Run(tc.query, func(tst *testing.T) {
			state, err := t.NewJSONViewerState([]byte(viewerJSON))
			require.NoError(tst, err)
			require.NoError(tst, state.SetQuery(tc.query))
			require.Equal(tst, tc.paths, jsonPaths(state.Tree.Nodes.Peek()))
			require.Equal(tst, tc.count, state.MatchCount())
		})
	}
}

func TestJSONViewerValues(tst *testing.T) {
	for _, source := range []string{`null`, `true`, `false`, `"line\n\t\u001b[31m"`, `9007199254740993`, `1.2300e+42`, `[]`, `{}`} {
		tst.Run(source, func(tst *testing.T) {
			state, err := t.NewJSONViewerState([]byte(source))
			require.NoError(tst, err)
			p := t.NewPilot(tst, t.JSONViewer{State: state}, 60, 4)
			require.NotEmpty(tst, p.FocusedID(), "an explicit ID is optional")
			require.Contains(tst, p.ScreenText(), source)
			require.NoError(tst, state.SetQuery("$"))
			require.Equal(tst, []string{"$"}, jsonPaths(state.Tree.Nodes.Peek()))
			require.Equal(tst, 1, state.MatchCount())
		})
	}
	state, err := t.NewJSONViewerState([]byte(`{"z": 1, "a": 9007199254740993}`))
	require.NoError(tst, err)
	require.Equal(tst, []string{`$`, `$['a']`, `$['z']`}, jsonPaths(state.Tree.Nodes.Peek()))
	node, ok := state.Tree.NodeAtPath([]int{0, 0})
	require.True(tst, ok)
	require.Equal(tst, json.Number("9007199254740993"), node.Data.Value)
}

func TestJSONViewerInvalidInputAndReplacement(tst *testing.T) {
	state, err := t.NewJSONViewerState([]byte(viewerJSON))
	require.NoError(tst, err)
	require.NoError(tst, state.SetQuery("$.items[1].name"))
	before := state.Tree.Nodes.Peek()
	for _, source := range []string{"", "{", `{"x":}`, `{} {}`, `null false`, `[] garbage`} {
		_, err := t.NewJSONViewerState([]byte(source))
		require.Error(tst, err)
		require.Error(tst, state.SetJSON([]byte(source)))
		require.Equal(tst, before, state.Tree.Nodes.Peek())
	}
	for _, query := range []string{"$[", "$.items[?", "not a path"} {
		require.Error(tst, state.SetQuery(query))
		require.Equal(tst, "$.items[1].name", state.Query())
		require.Equal(tst, before, state.Tree.Nodes.Peek())
		require.Equal(tst, 1, state.MatchCount())
	}
	require.NoError(tst, state.SetJSON([]byte(`{"items": [{"name":"first"}, {"name":"replacement"}]}`)))
	node, ok := state.Tree.NodeAtPath([]int{0, 0, 0, 0})
	require.True(tst, ok)
	require.Equal(tst, "replacement", node.Data.Value)
	require.Equal(tst, `$['items'][1]['name']`, node.Data.Path)
	require.NoError(tst, state.SetQuery(" \n\t"))
	require.Equal(tst, "", state.Query())
	require.Equal(tst, []string{`$`, `$['items']`, `$['items'][0]`, `$['items'][0]['name']`, `$['items'][1]`, `$['items'][1]['name']`}, jsonPaths(state.Tree.Nodes.Peek()))
}

func TestJSONViewerNavigationAndFiltering(tst *testing.T) {
	state, err := t.NewJSONViewerState([]byte(viewerJSON))
	require.NoError(tst, err)
	var activated t.JSONNode
	var selected []t.JSONNode
	viewer := t.JSONViewer{
		ID: "json", State: state, MultiSelect: true,
		Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)},
		OnSelect: func(node t.JSONNode, selection []t.JSONNode) {
			activated, selected = node, selection
		},
	}
	p := t.NewPilot(tst, t.Column{Height: t.Flex(1), Children: []t.Widget{
		viewer, t.Text{ID: "footer", Content: "Footer"},
	}}, 45, 5)
	require.Equal(tst, "json", p.FocusedID())
	p.Press("left")
	require.NotContains(tst, p.ScreenText(), `"items"`)
	// A query reveals a result beneath the collapsed root.
	require.NoError(tst, state.SetQuery("$.items[2]"))
	require.Contains(tst, p.ScreenText(), `"same"`)
	p.Press("right", "right", "left")
	require.NotContains(tst, p.ScreenText(), `"same"`, "filtered branches still collapse")
	p.Press("right", "end", "enter")
	require.Equal(tst, `$['items'][2]['price']`, activated.Path)
	require.Equal(tst, json.Number("9.5"), activated.Value)
	p.Advance(time.Second)
	require.Contains(tst, p.ScreenText(), `"price": 9.5`, "cursor is scrolled into view")
	p.Press("shift+up", "enter")
	require.Len(tst, selected, 2)
	require.Equal(tst, `$['items'][2]['name']`, selected[0].Path)
	require.Equal(tst, `$['items'][2]['price']`, selected[1].Path)
	p.Press("home")
	p.Advance(time.Second)
	p.ClickAt(0, 0)
	require.NotContains(tst, p.ScreenText(), `"same"`, "mouse clicks toggle the root indicator")
	p.ClickAt(0, 0)
	require.Contains(tst, p.ScreenText(), `"same"`)

	require.NoError(tst, state.SetQuery("$.missing"))
	require.Contains(tst, p.ScreenText(), "No JSONPath matches")
	footer, ok := p.Bounds("footer")
	require.True(tst, ok)
	require.Equal(tst, 4, footer.Y, "a flex-height viewport must not shrink when filtering")
	viewport, ok := p.Bounds("json")
	require.True(tst, ok)
	require.Equal(tst, 45, viewport.Width, "a flex-width viewport must not shrink when filtering")
	p.Press("down", "enter")
	require.Equal(tst, `$['items'][2]['name']`, activated.Path, "no stale activation in an empty view")
	require.NoError(tst, state.SetQuery(""))
	p.Press("home", "enter")
	require.Equal(tst, "$", activated.Path)
	require.Empty(tst, selected, "changing queries clears selection")
	require.Equal(tst, "json", p.FocusedID())
}
