package main

import (
	"log"

	t "terma"
)

// TreeItem represents an item in the request tree
type TreeItem struct {
	Name string
	Type string // "folder" or "request"
}

// HeaderRow represents a row in the headers table
type HeaderRow struct {
	Key   string
	Value string
}

type APIClientDemo struct {
	mainState            *t.SplitPaneState
	rightState           *t.SplitPaneState
	treeState            *t.TreeState[TreeItem]
	requestTab           t.Signal[string]
	responseTab          t.Signal[string]
	headersTable         *t.TableState[HeaderRow]
	bodyTextArea         *t.TextAreaState
	responseBodyTextArea *t.TextAreaState
}

func NewAPIClientDemo() *APIClientDemo {
	// Build the tree structure
	treeNodes := []t.TreeNode[TreeItem]{
		{
			Data: TreeItem{Name: "My API", Type: "folder"},
			Children: []t.TreeNode[TreeItem]{
				{Data: TreeItem{Name: "GET /users", Type: "request"}},
				{Data: TreeItem{Name: "POST /users", Type: "request"}},
				{
					Data: TreeItem{Name: "Auth", Type: "folder"},
					Children: []t.TreeNode[TreeItem]{
						{Data: TreeItem{Name: "POST /login", Type: "request"}},
						{Data: TreeItem{Name: "POST /logout", Type: "request"}},
					},
				},
			},
		},
		{
			Data: TreeItem{Name: "External APIs", Type: "folder"},
			Children: []t.TreeNode[TreeItem]{
				{Data: TreeItem{Name: "GET /weather", Type: "request"}},
				{Data: TreeItem{Name: "GET /news", Type: "request"}},
			},
		},
	}

	// Sample headers data
	headers := []HeaderRow{
		{Key: "Content-Type", Value: "application/json"},
		{Key: "Authorization", Value: "Bearer token123"},
		{Key: "Accept", Value: "application/json"},
		{Key: "User-Agent", Value: "APIClient/1.0"},
	}

	return &APIClientDemo{
		mainState:    t.NewSplitPaneState(0.25),
		rightState:   t.NewSplitPaneState(0.5),
		treeState:    t.NewTreeState(treeNodes),
		requestTab:   t.NewSignal("headers"),
		responseTab:  t.NewSignal("body"),
		headersTable: t.NewTableState(headers),
		bodyTextArea: t.NewTextAreaState(`{
  "name": "John Doe",
  "email": "john@example.com"
}`),
		responseBodyTextArea: t.NewTextAreaState(`{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z"
}`),
	}
}

func (d *APIClientDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (d *APIClientDemo) Build(ctx t.BuildContext) t.Widget {
	return t.Dock{
		Top: []t.Widget{
			t.Text{
				Content: " API Client Demo ",
				Width:   t.Flex(1),
				Style: t.Style{
					ForegroundColor: ctx.Theme().Background,
					BackgroundColor: ctx.Theme().Primary,
				},
			},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: ctx.Theme().Surface,
					Padding:         t.EdgeInsetsXY(2, 0),
				},
			},
		},
		Body: d.buildBody(ctx),
	}
}

func (d *APIClientDemo) buildBody(ctx t.BuildContext) t.Widget {
	return t.SplitPane{
		ID:                "main-split",
		State:             d.mainState,
		Orientation:       t.SplitHorizontal,
		DividerSize:       1,
		DividerBackground: ctx.Theme().Surface,
		DividerFocusForeground: t.NewGradient(
			ctx.Theme().Primary,
			ctx.Theme().Accent,
		).WithAngle(0),
		MinPaneSize: 15,
		First:       d.buildLeftPanel(ctx),
		Second:      d.buildRightPanel(ctx),
	}
}

func (d *APIClientDemo) buildLeftPanel(ctx t.BuildContext) t.Widget {
	return t.Column{
		Style: t.Style{
			BackgroundColor: ctx.Theme().Surface,
			Height:          t.Flex(1),
		},
		Children: []t.Widget{
			t.Text{
				Content: " Requests",
				Style: t.Style{
					ForegroundColor: ctx.Theme().TextMuted,
					BackgroundColor: ctx.Theme().Surface,
					Bold:            true,
					Padding:         t.EdgeInsetsXY(0, 1),
				},
			},
			t.Tree[TreeItem]{
				ID:    "request-tree",
				State: d.treeState,
				RenderNode: func(item TreeItem, nodeCtx t.TreeNodeContext) t.Widget {
					icon := ""
					if item.Type == "folder" {
						icon = ""
					} else {
						icon = ""
					}
					return t.Text{
						Content: icon + " " + item.Name,
						Style: t.Style{
							Width: t.Flex(1),
						},
					}
				},
				Style: t.Style{
					Padding: t.EdgeInsetsXY(1, 0),
					Height:  t.Flex(1),
				},
			},
		},
	}
}

func (d *APIClientDemo) buildRightPanel(ctx t.BuildContext) t.Widget {
	return t.SplitPane{
		ID:                "right-split",
		State:             d.rightState,
		Orientation:       t.SplitVertical,
		DividerSize:       1,
		DividerBackground: ctx.Theme().Surface,
		MinPaneSize:       5,
		First:             d.buildRequestPanel(ctx),
		Second:            d.buildResponsePanel(ctx),
	}
}

func (d *APIClientDemo) buildRequestPanel(ctx t.BuildContext) t.Widget {
	active := d.requestTab.Get()
	return t.Column{
		Style: t.Style{
			Height: t.Flex(1),
		},
		Children: []t.Widget{
			d.buildTabBar(ctx, []string{"headers", "body", "query", "auth", "info", "options"},
				[]string{"Headers", "Body", "Query", "Auth", "Info", "Options"},
				active, func(key string) { d.requestTab.Set(key) }),
			t.Switcher{
				Active: active,
				Children: map[string]t.Widget{
					"headers": d.buildHeadersTable(ctx),
					"body":    d.buildBodyTextArea(ctx),
					"query":   d.buildPlaceholder(ctx, "Query parameters will appear here"),
					"auth":    d.buildPlaceholder(ctx, "Authentication settings will appear here"),
					"info":    d.buildPlaceholder(ctx, "Request info will appear here"),
					"options": d.buildPlaceholder(ctx, "Request options will appear here"),
				},
				Style: t.Style{
					Height:          t.Flex(1),
					BackgroundColor: ctx.Theme().Surface,
				},
			},
		},
	}
}

func (d *APIClientDemo) buildResponsePanel(ctx t.BuildContext) t.Widget {
	active := d.responseTab.Get()
	return t.Column{
		Style: t.Style{
			Height: t.Flex(1),
		},
		Children: []t.Widget{
			d.buildTabBar(ctx, []string{"body", "headers", "cookie", "trace"},
				[]string{"Body", "Headers", "Cookie", "Trace"},
				active, func(key string) { d.responseTab.Set(key) }),
			t.Switcher{
				Active: active,
				Children: map[string]t.Widget{
					"body":    d.buildResponseBodyTextArea(ctx),
					"headers": d.buildPlaceholder(ctx, "Response headers will appear here"),
					"cookie":  d.buildPlaceholder(ctx, "Cookies will appear here"),
					"trace":   d.buildPlaceholder(ctx, "Request trace will appear here"),
				},
				Style: t.Style{
					Height:          t.Flex(1),
					BackgroundColor: ctx.Theme().Surface,
				},
			},
		},
	}
}

func (d *APIClientDemo) buildTabBar(ctx t.BuildContext, keys []string, labels []string, active string, onSelect func(string)) t.Widget {
	children := make([]t.Widget, len(keys))
	for i, key := range keys {
		k := key // capture for closure
		style := t.Style{
			ForegroundColor: ctx.Theme().TextMuted,
			BackgroundColor: ctx.Theme().Surface,
			Padding:         t.EdgeInsetsXY(2, 0),
		}
		if key == active {
			style.ForegroundColor = ctx.Theme().Background
			style.BackgroundColor = ctx.Theme().Accent
		}
		children[i] = t.Text{
			Content: labels[i],
			Style:   style,
			Click: func(t.MouseEvent) {
				onSelect(k)
			},
		}
	}
	return t.Row{
		Style: t.Style{
			BackgroundColor: ctx.Theme().Surface,
		},
		Children: children,
	}
}

func (d *APIClientDemo) buildHeadersTable(ctx t.BuildContext) t.Widget {
	return t.Table[HeaderRow]{
		ID:    "headers-table",
		State: d.headersTable,
		Columns: []t.TableColumn{
			{Width: t.Flex(1), Header: t.Text{Content: "Header", Style: t.Style{Bold: true, Padding: t.EdgeInsetsXY(1, 0)}}},
			{Width: t.Flex(2), Header: t.Text{Content: "Value", Style: t.Style{Bold: true, Padding: t.EdgeInsetsXY(1, 0)}}},
		},
		RenderCell: func(row HeaderRow, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
			content := row.Key
			if colIndex == 1 {
				content = row.Value
			}
			style := t.Style{
				ForegroundColor: ctx.Theme().Text,
				Padding:         t.EdgeInsetsXY(1, 0),
			}
			if active {
				style.BackgroundColor = ctx.Theme().ActiveCursor
				style.ForegroundColor = ctx.Theme().SelectionText
			}
			return t.Text{Content: content, Style: style}
		},
		SelectionMode: t.TableSelectionRow,
		Style: t.Style{
			Height: t.Cells(6),
		},
	}
}

func (d *APIClientDemo) buildBodyTextArea(ctx t.BuildContext) t.Widget {
	return t.TextArea{
		ID:    "body-textarea",
		State: d.bodyTextArea,
		Style: t.Style{
			Height:          t.Flex(1),
			Width:           t.Flex(1),
			BackgroundColor: ctx.Theme().Surface,
			ForegroundColor: ctx.Theme().Text,
			Padding:         t.EdgeInsetsXY(1, 1),
		},
	}
}

func (d *APIClientDemo) buildResponseBodyTextArea(ctx t.BuildContext) t.Widget {
	return t.TextArea{
		ID:    "response-body-textarea",
		State: d.responseBodyTextArea,
		Style: t.Style{
			Height:          t.Flex(1),
			Width:           t.Flex(1),
			BackgroundColor: ctx.Theme().Surface,
			ForegroundColor: ctx.Theme().Text,
			Padding:         t.EdgeInsetsXY(1, 1),
		},
	}
}

func (d *APIClientDemo) buildPlaceholder(ctx t.BuildContext, text string) t.Widget {
	return t.Text{
		Content: text,
		Style: t.Style{
			ForegroundColor: ctx.Theme().TextMuted,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
	}
}

func main() {
	app := NewAPIClientDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
