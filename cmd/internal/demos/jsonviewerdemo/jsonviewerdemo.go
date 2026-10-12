// Package jsonviewerdemo demonstrates a standalone JSONViewer with live JSONPath filtering.
package jsonviewerdemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes this demo in the gallery.
var Info = demokit.Info{
	Key: "json-viewer", Title: "JSON Viewer",
	Description: "Explore a JSON tree and filter it with JSONPath",
}

const document = `{
	"store": {
		"books": [
			{"title": "The Hobbit", "author": "J. R. R. Tolkien", "price": 8.99, "inStock": true},
			{"title": "Dune", "author": "Frank Herbert", "price": 14.50, "inStock": false},
			{"title": "A Wizard of Earthsea", "author": "Ursula K. Le Guin", "price": 7.50, "inStock": true}
		],
		"location": {"city": "London", "online": true},
		"tags": ["fiction", "fantasy", "science fiction"]
	},
	"updated": null,
	"version": 9007199254740993
}`

type demo struct {
	state       *t.JSONViewerState
	input       *t.TextInputState
	queryStatus t.Signal[string]
	status      t.Signal[string]
}

// New creates a gallery or standalone demo.
func New() demokit.Demo {
	state, err := t.NewJSONViewerState([]byte(document))
	if err != nil {
		panic(err) // The built-in fixture must be valid JSON.
	}
	return &demo{
		state: state, input: t.NewTextInputState(""),
		queryStatus: t.NewSignal("Whole document · object keys sorted · original array indices preserved"),
		status:      t.NewSignal("Enter or double-click a value to inspect its JSONPath"),
	}
}

func (a *demo) InitialFocus() string { return "json-viewer" }

func (a *demo) setQuery(query string) {
	if err := a.state.SetQuery(query); err != nil {
		a.queryStatus.Set("Invalid JSONPath (keeping last valid view): " + err.Error())
		return
	}
	if a.state.Query() == "" {
		a.queryStatus.Set("Whole document · object keys sorted · original array indices preserved")
	} else {
		a.queryStatus.Set(fmt.Sprintf("%d matches · ancestors and matched subtrees retained", a.state.MatchCount()))
	}
}

func (a *demo) clear() {
	a.input.SetText("")
	a.setQuery("")
	t.RequestFocus("json-viewer")
}

func (a *demo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "/", Name: "JSONPath", Action: func() { t.RequestFocus("json-query") }},
		{Key: "ctrl+f", Name: "JSONPath", Hidden: true, Action: func() { t.RequestFocus("json-query") }},
		{Key: "escape", Name: "Clear", Action: a.clear},
		{Key: "e", Name: "Expand all", Action: a.state.Tree.ExpandAll},
		{Key: "c", Name: "Collapse all", Action: a.state.Tree.CollapseAll},
		{Key: "t", Name: "Theme", Action: demokit.NextTheme},
	}
}

func (a *demo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	viewerStyle := demokit.PanelStyle(theme, "Document", ctx.IsFocused(t.JSONViewer{ID: "json-viewer"}))
	viewerStyle.Width, viewerStyle.Height = t.Flex(1), t.Flex(1)
	queryStatus := t.SignalText(a.queryStatus, func(value string) string { return value })
	queryStatus.LayoutStyle = t.Style{Width: t.Flex(1), Height: t.Auto}
	queryStatus.Wrap = t.WrapSoft
	return t.Dock{
		Style:  t.Style{BackgroundColor: theme.Background},
		Top:    []t.Widget{demokit.Header{Title: "JSON Viewer", Tagline: "Tree navigation + JSONPath"}},
		Bottom: []t.Widget{demokit.Footer(theme)},
		Body: t.Column{
			Style: t.Style{Width: t.Flex(1), Height: t.Flex(1), Padding: t.EdgeInsetsAll(1)},
			Children: []t.Widget{
				t.Row{
					Style: demokit.PanelStyle(theme, "JSONPath", ctx.IsFocused(t.TextInput{ID: "json-query"})),
					Width: t.Flex(1),
					Children: []t.Widget{t.TextInput{
						ID: "json-query", State: a.input, Style: t.Style{Width: t.Flex(1)},
						Placeholder: "Try $..price or $.store.books[?@.price < 10]",
						OnChange:    a.setQuery,
						OnSubmit:    func(string) { t.RequestFocus("json-viewer") },
						ExtraKeybinds: []t.Keybind{
							{Key: "escape", Name: "Clear", Action: a.clear},
						},
					}},
				},
				queryStatus,
				t.JSONViewer{
					ID: "json-viewer", State: a.state, Style: viewerStyle, MultiSelect: true,
					OnSelect: func(node t.JSONNode, _ []t.JSONNode) { a.status.Set("Selected: " + node.Path) },
				},
				t.SignalText(a.status, func(value string) string { return value }),
			},
		},
	}
}
