// Package jumpdemo demonstrates jump mode: keyboard labels overlaid on the
// screen that move focus straight to what they label.
package jumpdemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "jump",
	Title:       "Jump mode",
	Description: "Overlay keys on the screen and jump focus straight to any widget",
}

// JumpDemo is a small API client laid out like Posting. Press ctrl+o to show
// the jump labels, then a label's key to jump there.
//
// The panels have static keys (1-6), which never change, so they can be
// learned. Everything that comes from data - each request in the
// collection, each row of the response, each tab - gets a dynamic hint
// (a, s, d...), because nobody could have assigned it a key in advance.
//
//	ctrl+o    - Enter (or leave) jump mode
//	1-6       - Static keys: collection, method, URL, send, request, response
//	a, s, d…  - Dynamic hints on the requests, response rows and tabs in view
//	ctrl+y    - Turn dynamic hints off (static keys only, like Posting) or on
//	backspace - Undo the last character of a two-character hint
//	escape    - Leave jump mode without jumping
type JumpDemo struct {
	jump           *t.JumpState
	dynamic        *t.CheckboxState
	requests       *t.ListState[request]
	requestsScroll *t.ScrollState
	url            *t.TextInputState
	method         t.Signal[string]
	tabs           *t.TabState
	body           *t.TextAreaState
	users          *t.TableState[user]
	usersScroll    *t.ScrollState
	sent           t.Signal[int]
}

type request struct {
	method, path string
}

func (r request) String() string { return fmt.Sprintf("%-6s %s", r.method, r.path) }

type user struct {
	id   int
	name string
}

var names = []string{
	"Leanne Graham", "Ervin Howell", "Clementine Bauch", "Patricia Lebsack", "Chelsey Dietrich",
	"Dennis Schulist", "Kurtis Weissnat", "Nicholas Runolfsdottir", "Glenna Reichert", "Clementina DuBuque",
	"Ada Lovelace", "Grace Hopper", "Alan Turing", "Edsger Dijkstra", "Barbara Liskov",
}

// New creates the demo.
func New() demokit.Demo {
	d := &JumpDemo{
		jump:    t.NewJumpState(),
		dynamic: t.NewCheckboxState(true),
		requests: t.NewListState([]request{
			{"GET", "/users"}, {"GET", "/users/1"}, {"POST", "/users"}, {"PUT", "/users/1"},
			{"DELETE", "/users/1"}, {"GET", "/posts"}, {"GET", "/posts?userId=1"}, {"POST", "/posts"},
			{"GET", "/comments"}, {"GET", "/albums"}, {"GET", "/photos"}, {"GET", "/todos"},
			{"GET", "/todos?completed=true"}, {"GET", "/health"}, {"GET", "/version"},
			{"GET", "/metrics"}, {"POST", "/login"}, {"POST", "/logout"},
		}),
		requestsScroll: t.NewScrollState(),
		url:            t.NewTextInputState(""),
		method:         t.NewSignal("GET"),
		tabs: t.NewTabState([]t.Tab{
			{Key: "body", Label: "Body"}, {Key: "headers", Label: "Headers"},
			{Key: "query", Label: "Query"}, {Key: "auth", Label: "Auth"},
		}),
		body:        t.NewTextAreaState("{\n  \"name\": \"Ada Lovelace\"\n}"),
		users:       t.NewTableState[user](nil),
		usersScroll: t.NewScrollState(),
		sent:        t.NewSignal(0),
	}
	d.load(d.requests.GetItems()[0])
	d.send()
	return d
}

func (d *JumpDemo) InitialFocus() string { return "url" }

// load puts a request from the collection into the URL bar.
func (d *JumpDemo) load(r request) {
	d.method.Set(r.method)
	d.url.SetText("https://jsonplaceholder.typicode.com" + r.path)
}

// send fills the response table with (made up) users, in a different order
// each time.
func (d *JumpDemo) send() {
	count := d.sent.Peek() + 1
	d.sent.Set(count)
	users := make([]user, len(names))
	for i, name := range names {
		users[(i+count-1)%len(names)] = user{id: i + 1, name: name}
	}
	d.users.SetRows(users)
}

func (d *JumpDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "ctrl+y", Name: "Dynamic hints", Action: d.dynamic.Toggle},
		{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme},
	}
}

func (d *JumpDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Jumper{
		State: d.jump,
		// The static jump map: the same key always reaches the same place.
		Targets: []t.JumpTarget{
			{Key: "1", ID: "requests"},
			{Key: "2", ID: "method"},
			{Key: "3", ID: "url"},
			// An action runs instead of moving focus.
			{Key: "4", ID: "send", Action: d.send},
			{Key: "5", ID: "body"},
			{Key: "6", ID: "users"},
		},
		// Label the requests, response rows and tabs in view, Vimium style.
		Dynamic: d.dynamic.Checked.Get(),
		Child: t.Dock{
			Style:  t.Style{BackgroundColor: theme.Background},
			Top:    []t.Widget{header{demo: d}},
			Bottom: []t.Widget{demokit.Footer(theme)},
			Body: t.Row{
				Width:   t.Flex(1),
				Height:  t.Flex(1),
				Spacing: 1,
				Style:   t.Style{Padding: t.EdgeInsetsXY(1, 1)},
				Children: []t.Widget{
					d.requestsPanel(ctx),
					t.Column{
						Width:   t.Flex(1),
						Height:  t.Flex(1),
						Spacing: 1,
						Children: []t.Widget{
							d.urlBar(ctx),
							t.Row{
								Width:   t.Flex(1),
								Height:  t.Flex(1),
								Spacing: 1,
								Children: []t.Widget{
									d.requestPanel(ctx),
									d.responsePanel(ctx),
								},
							},
						},
					},
				},
			},
		},
	}
}

// header reads the jump state itself, so entering jump mode only rebuilds it.
type header struct {
	demo *JumpDemo
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	hints := "static + dynamic"
	if !h.demo.dynamic.Checked.Get() {
		hints = "static only"
	}
	mode := fmt.Sprintf("[$TextMuted]press[/] [b $Accent]ctrl+o[/] [$TextMuted]to jump · hints:[/] [b]%s[/]", hints)
	if h.demo.jump.IsActive() {
		mode = "[b $Accent]jump mode[/] [$TextMuted]type a label · esc to cancel[/]"
		if typed := h.demo.jump.Typed(); typed != "" {
			mode = fmt.Sprintf("[b $Accent]jump mode[/] [$TextMuted]typed[/] [b]%s[/]", typed)
		}
	}
	return demokit.Header{Title: "Jump Mode", Tagline: "Inspired by Posting and Vimium", Right: mode}
}

func (d *JumpDemo) requestsPanel(ctx t.BuildContext) t.Widget {
	list := t.List[request]{
		ID:             "requests",
		State:          d.requests,
		ScrollState:    d.requestsScroll,
		OnCursorChange: d.load,
	}
	return t.Scrollable{
		State:  d.requestsScroll,
		Width:  t.Cells(30),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(ctx.Theme(), "1 Collection", ctx.IsFocused(list)),
		Child:  list,
	}
}

func (d *JumpDemo) urlBar(ctx t.BuildContext) t.Widget {
	url := t.TextInput{ID: "url", State: d.url, Style: t.Style{Width: t.Flex(1)}, OnSubmit: func(string) { d.send() }}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(ctx.Theme(), "2 Method · 3 URL · 4 Send", ctx.IsFocused(url)),
		Children: []t.Widget{
			t.Button{ID: "method", Label: d.method.Get(), OnPress: d.cycleMethod},
			url,
			t.Button{ID: "send", Label: "Send", Variant: t.ButtonPrimary, OnPress: d.send},
		},
	}
}

func (d *JumpDemo) cycleMethod() {
	methods := []string{"GET", "POST", "PUT", "DELETE"}
	for i, m := range methods {
		if m == d.method.Peek() {
			d.method.Set(methods[(i+1)%len(methods)])
			return
		}
	}
	d.method.Set(methods[0])
}

func (d *JumpDemo) requestPanel(ctx t.BuildContext) t.Widget {
	body := t.TextArea{ID: "body", State: d.body, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)}}
	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(ctx.Theme(), "5 Request", ctx.IsFocused(body)),
		Children: []t.Widget{
			t.TabBar{ID: "request-tabs", State: d.tabs},
			body,
		},
	}
}

func (d *JumpDemo) responsePanel(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	table := t.Table[user]{
		ID:            "users",
		State:         d.users,
		ScrollState:   d.usersScroll,
		SelectionMode: t.TableSelectionRow,
		Columns:       []t.TableColumn{{Width: t.Cells(4)}, {Width: t.Flex(1)}},
		RenderCell: func(u user, row, col int, active, selected bool) t.Widget {
			style := t.Style{Width: t.Flex(1)}
			if active {
				style.BackgroundColor = theme.ActiveCursor
				style.ForegroundColor = theme.SelectionText
			}
			if col == 0 {
				if !active {
					style.ForegroundColor = theme.TextMuted
				}
				return t.Text{Content: fmt.Sprint(u.id), Style: style}
			}
			return t.Text{Content: u.name, Style: style}
		},
	}
	return t.Scrollable{
		State:  d.usersScroll,
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, fmt.Sprintf("6 Response · 200 OK · #%d", d.sent.Get()), ctx.IsFocused(table)),
		Child:  table,
	}
}
