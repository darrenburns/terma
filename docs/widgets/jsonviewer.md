# JSONViewer

`JSONViewer` displays JSON as a syntax-colored, scrollable tree that can be filtered with JSONPath. It composes `Tree` and `Scrollable`; no application-specific tree building is needed.

## Example

Create the state once, outside `Build()`:

```go
state, err := t.NewJSONViewerState([]byte(`{
    "users": [{"name": "Ada", "active": true}, {"name": "Leon", "active": false}]
}`))
if err != nil {
    return err
}

// Optional: select only active users, retaining their full objects and ancestors.
if err := state.SetQuery("$.users[?@.active == true]"); err != nil {
    return err
}

viewer := t.JSONViewer{
    ID:    "json",
    State: state,
    Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)},
    OnSelect: func(node t.JSONNode, selected []t.JSONNode) {
        // node.Path is a normalized JSONPath, such as $['users'][0]['name'].
        // node.Value is the original decoded value, not a filtered copy.
    },
}
```

Run the live-filtering demo:

```sh
go run ./cmd/json-viewer-demo
```

It is also available in `go run ./cmd/terma-demos`.

## Filtering

Connect a separate `TextInput` to `state.SetQuery` in its `OnChange` or `OnSubmit` handler. The viewer does not impose an input field or keybinding on its host application. `SetQuery` returns parsing errors without changing the last valid query or view, so the application can display errors while the user edits. `state.Query()` and `state.MatchCount()` subscribe to changes for reactive status text.

JSONPath follows [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html), using `github.com/theory/jsonpath`:

| Query | Selects |
| --- | --- |
| `$` | The whole document |
| `$.users[0]` | The first user and its complete subtree |
| `$.users[*].name` | Every user's name |
| `$..name` | Every `name` member at any depth |
| `$.users[1:3]` | Users at indices 1 and 2 |
| `$.users[-1]` | The last user |
| `$.users[?@.active == true]` | Users whose `active` member is true |
| `$['key.with.dots']` | A member whose name contains dots |

- Empty or whitespace-only queries restore the whole document.
- Matches retain their ancestors for context. Matched objects and arrays retain their complete subtrees.
- Ancestors of matches expand when a query is applied. All visible branches can still be collapsed or expanded normally.
- Results remain in document-tree order, with object keys sorted lexicographically and arrays in their original order. Duplicate matches appear once; equal values at different paths remain distinct.
- Array labels and `JSONNode.Path` always refer to the original document. `Tree.CursorPath` refers to child indices in the **filtered tree**, so use `JSONNode.Path` to identify source values.
- `MatchCount()` counts distinct matches, excluding context ancestors and descendants included with a matched container. An empty query counts as one root match.
- A different valid query resets the cursor to the root, clears selection, and resets scrolling. Existing collapse state is retained except for ancestors expanded to reveal matches.
- No results displays **No JSONPath matches**, not `null`. JSON `null` is an ordinary selectable value.

Numbers retain their original text through `json.Number`, including large integers and exponent notation. **Numeric comparisons in JSONPath use `float64` precision and range**, as provided by the JSONPath library; very large integers and high-precision decimals may not compare exactly.

## Navigation and data

- Up/Down (or j/k) moves between visible nodes. Left/Right (or h/l) collapses/expands or moves to a parent/child. Space toggles expansion; Home/End moves to the first/last visible node.
- Enter or a double-click invokes `OnSelect`. `ActivateOnClick` enables single-click activation. Clicking an expansion indicator toggles the branch instead.
- `MultiSelect` enables Tree's Shift navigation, Shift-click, and drag selection. `OnCursorChange` receives the current `JSONNode` after navigation.
- `state.Tree` exposes Tree's navigation, expansion, and selection methods, including `ExpandAll()` and `CollapseAll()`.
- `state.SetJSON(data)` replaces the document, reapplies the current query, and resets expansion, selection, cursor, and scroll position. Invalid JSON leaves the previous state intact.
- Any single JSON value is accepted, including scalar roots and empty containers. Trailing non-whitespace content is rejected.
- Object keys and strings are JSON-escaped before display, so embedded control characters remain visible text.
- Container summaries show the original number of keys or items, even when filtering hides children.
- Treat `JSONNode.Value` and `state.Tree.Nodes` as read-only. Use `SetJSON` to change the data; call setters outside `Build()`.
- Set `Style.Width` and `Style.Height` to constrain the built-in viewport. `DisableFocus` disables tree focus.

## Related

- [Tree](tree.md) provides the underlying navigation and expansion behavior.
- [TextInput](textinput.md) can supply an editable JSONPath query.
