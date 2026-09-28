// Package jumpdemo demonstrates jump mode: keyboard labels overlaid on the
// screen that move focus straight to what they label.
package jumpdemo

import (
	"fmt"
	"strings"

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
// the jump labels, then a label's key to move focus there.
//
//	ctrl+o    - Enter (or leave) jump mode
//	1-6       - The static jump map: collections, method, URL, send, body, response
//	a, s, d…  - Dynamic hints on everything else in view: tabs, buttons, options
//	ctrl+y    - Turn dynamic hints off (static keys only, like Posting) or on
//	backspace - Undo the last character of a two-character hint
//	escape    - Leave jump mode without jumping
type JumpDemo struct {
	jump        *t.JumpState
	dynamic     *t.CheckboxState
	follow      *t.CheckboxState
	verify      *t.CheckboxState
	collections *t.ListState[string]
	url         *t.TextInputState
	body        *t.TextAreaState
	response    *t.ListState[string]
	method      t.Signal[int]
	sent        t.Signal[int]
	requestTab  t.Signal[string]
	view        t.Signal[string]
	status      t.Signal[string]
}

var methods = []string{"GET", "POST", "PUT", "DELETE"}

// New creates the demo.
func New() demokit.Demo {
	return &JumpDemo{
		jump:    t.NewJumpState(),
		dynamic: t.NewCheckboxState(true),
		follow:  t.NewCheckboxState(true),
		verify:  t.NewCheckboxState(true),
		collections: t.NewListState([]string{
			"users / list", "users / get", "users / create", "users / delete",
			"posts / list", "posts / search", "comments / list", "health",
		}),
		url:      t.NewTextInputState("https://jsonplaceholder.typicode.com/users"),
		body:     t.NewTextAreaState("{\n  \"name\": \"Ada Lovelace\"\n}"),
		response: t.NewListState([]string{"Press 4 in jump mode to send."}),
		method:     t.NewSignal(0),
		sent:       t.NewSignal(0),
		requestTab: t.NewSignal("Body"),
		view:       t.NewSignal("Pretty"),
		status:     t.NewSignal(""),
	}
}

func (d *JumpDemo) InitialFocus() string { return "url" }

func (d *JumpDemo) send() {
	count := d.sent.Peek() + 1
	d.sent.Set(count)
	d.response.SetItems([]string{
		fmt.Sprintf("%s %s", methods[d.method.Peek()], d.url.GetText()),
		fmt.Sprintf("HTTP/1.1 200 OK   (request #%d)", count),
		"content-type: application/json",
		"",
		"[",
		"  { \"id\": 1, \"name\": \"Leanne Graham\" },",
		"  { \"id\": 2, \"name\": \"Ervin Howell\" }",
		"]",
	})
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
			{Key: "1", ID: "collections"},
			{Key: "2", ID: "method"},
			{Key: "3", ID: "url"},
			// An action runs instead of moving focus.
			{Key: "4", ID: "send", Action: d.send},
			{Key: "5", ID: "body"},
			{Key: "6", ID: "response"},
		},
		// Label everything else in view, Vimium style.
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
					d.collectionsPanel(ctx),
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
									d.bodyPanel(ctx),
									d.responsePanel(ctx),
								},
							},
							d.optionsPanel(ctx),
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

func (d *JumpDemo) collectionsPanel(ctx t.BuildContext) t.Widget {
	list := t.List[string]{ID: "collections", State: d.collections, Height: t.Flex(1)}
	return t.Column{
		Width:  t.Cells(26),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(ctx.Theme(), "1 Collections", ctx.IsFocused(list)),
		Children: []t.Widget{
			list,
			t.Button{ID: "new-request", Label: "+ New request", OnPress: func() {
				d.collections.SetItems(append(d.collections.GetItems(), "untitled"))
			}},
		},
	}
}

func (d *JumpDemo) urlBar(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	url := t.TextInput{ID: "url", State: d.url, Style: t.Style{Width: t.Flex(1)}, OnSubmit: func(string) { d.send() }}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "2 Method · 3 URL · 4 Send", ctx.IsFocused(url)),
		Children: []t.Widget{
			t.Button{ID: "method", Label: methods[d.method.Get()], OnPress: func() {
				d.method.Update(func(i int) int { return (i + 1) % len(methods) })
			}},
			url,
			t.Button{ID: "send", Label: "Send", Variant: t.ButtonPrimary, OnPress: d.send},
		},
	}
}

func (d *JumpDemo) bodyPanel(ctx t.BuildContext) t.Widget {
	body := t.TextArea{ID: "body", State: d.body, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)}}
	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(ctx.Theme(), "5 "+d.requestTab.Get(), ctx.IsFocused(body)),
		Children: []t.Widget{
			// No static keys here: dynamic hints reach these.
			buttonRow("tab", []string{"Body", "Headers", "Query"}, d.requestTab.Set),
			body,
		},
	}
}

func (d *JumpDemo) responsePanel(ctx t.BuildContext) t.Widget {
	response := t.List[string]{ID: "response", State: d.response, Height: t.Flex(1)}
	title := "6 Response · " + d.view.Get()
	if status := d.status.Get(); status != "" {
		title += " · " + status
	}
	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(ctx.Theme(), title, ctx.IsFocused(response)),
		Children: []t.Widget{
			buttonRow("view", []string{"Pretty", "Raw"}, d.view.Set),
			response,
			buttonRow("action", []string{"Copy", "Save"}, func(action string) {
				d.status.Set(map[string]string{"Copy": "copied", "Save": "saved"}[action])
			}),
		},
	}
}

// buttonRow is a row of buttons, each calling onPress with its label.
func buttonRow(idPrefix string, labels []string, onPress func(string)) t.Widget {
	buttons := make([]t.Widget, len(labels))
	for i, label := range labels {
		buttons[i] = t.Button{
			ID:      idPrefix + "-" + strings.ToLower(label),
			Label:   label,
			OnPress: func() { onPress(label) },
		}
	}
	return t.Row{Spacing: 1, Children: buttons}
}

func (d *JumpDemo) optionsPanel(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 2,
		Style:   demokit.PanelStyle(theme, "Options", false),
		Children: []t.Widget{
			&t.Checkbox{ID: "dynamic", State: d.dynamic, Label: "Dynamic hints"},
			&t.Checkbox{ID: "follow", State: d.follow, Label: "Follow redirects"},
			&t.Checkbox{ID: "verify", State: d.verify, Label: "Verify TLS"},
		},
	}
}
