// Package tabsdemo demonstrates TabView and the TabState behind it.
package tabsdemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "tabs",
	Title:       "Tabs",
	Description: "Switch, reorder, close and reopen tabs in a TabView",
}

// TabDemo demonstrates TabView (a TabBar above a Switcher) and the TabState
// that drives it: switching, position keys, reordering, closing, and adding
// tabs at runtime. Each tab's state lives on the App, so it survives switches.
//
//	←→ hl    - Previous / next tab (tab bar focused)
//	1-9      - Jump to a tab by position (tab bar focused)
//	ctrl+h/l - Move the active tab left / right
//	ctrl+w   - Close the active tab (or click its ×)
//	[ ]      - Previous / next tab from anywhere
//	n        - Open a new tab
//	u        - Reopen the last closed tab
//	+ -      - Change the counter (Counter tab)
//	t        - Cycle theme
//	tab      - Move focus between the tab bar and the content
type TabDemo struct {
	tabs *t.TabState

	counter t.Signal[int]
	fruits  *t.ListState[string]

	closed    t.AnySignal[[]t.Tab] // closed tabs, most recent last, for reopening
	lastEvent t.Signal[string]
	notes     int // for naming new tabs
}

// tabBarID is the ID TabView gives its TabBar: the TabView's ID + "-tabbar".
const (
	tabViewID    = "tabs"
	tabBarID     = tabViewID + "-tabbar"
	sidebarWidth = 32
)

// New creates the demo.
func New() demokit.Demo {
	d := &TabDemo{
		counter: t.NewSignal(0),
		fruits: t.NewListState([]string{
			"Apple", "Banana", "Cherry", "Date", "Elderberry",
		}),
		closed:    t.NewAnySignal([]t.Tab{}),
		lastEvent: t.NewSignal(""),
	}
	d.tabs = t.NewTabState([]t.Tab{
		{Key: "home", Label: "Home", Content: homePage{}},
		{Key: "counter", Label: "Counter", Content: counterPage{demo: d}},
		{Key: "list", Label: "List", Content: listPage{demo: d}},
		{Key: "info", Label: "Info", Content: infoPage{}},
	})
	return d
}

func (d *TabDemo) InitialFocus() string { return tabBarID }

// label returns the label of the tab with the given key.
func (d *TabDemo) label(key string) string {
	for _, tab := range d.tabs.TabsPeek() {
		if tab.Key == key {
			return tab.Label
		}
	}
	return key
}

// closeTab removes a tab, remembering it so it can be reopened.
func (d *TabDemo) closeTab(key string) {
	for _, tab := range d.tabs.TabsPeek() {
		if tab.Key == key {
			d.closed.Update(func(closed []t.Tab) []t.Tab { return append(closed, tab) })
			break
		}
	}
	d.lastEvent.Set("closed " + d.label(key))
	d.tabs.RemoveTab(key)
	// The closed tab may have held focus.
	t.RequestFocus(tabBarID)
}

// open adds a tab at the end and makes it active.
func (d *TabDemo) open(tab t.Tab, event string) {
	d.tabs.AddTab(tab)
	d.tabs.SetActiveKey(tab.Key)
	d.lastEvent.Set(event + " " + tab.Label)
	t.RequestFocus(tabBarID)
}

func (d *TabDemo) newTab() {
	d.notes++
	d.open(t.Tab{
		Key:     fmt.Sprintf("note-%d", d.notes),
		Label:   fmt.Sprintf("Note %d", d.notes),
		Content: notePage{number: d.notes},
	}, "opened")
}

func (d *TabDemo) reopen() {
	closed := d.closed.Peek()
	if len(closed) == 0 {
		return
	}
	tab := closed[len(closed)-1]
	d.closed.Set(closed[:len(closed)-1])
	d.open(tab, "reopened")
}

// step switches tabs from anywhere, for when focus is in the content.
func (d *TabDemo) step(forward bool) {
	if forward {
		d.tabs.SelectNext()
	} else {
		d.tabs.SelectPrevious()
	}
	d.onTabChange(d.tabs.ActiveKeyPeek())
	t.RequestFocus(tabBarID)
}

func (d *TabDemo) onTabChange(key string) {
	d.lastEvent.Set("switched to " + d.label(key))
}

func (d *TabDemo) addToCounter(n int) {
	d.counter.Update(func(c int) int { return c + n })
}

func (d *TabDemo) Keybinds() []t.Keybind {
	keybinds := []t.Keybind{
		{Key: "n", Name: "New tab", Action: d.newTab},
		{Key: "u", Name: "Reopen", Action: d.reopen},
		{Key: "[", Name: "Prev tab", Action: func() { d.step(false) }, Hidden: true},
		{Key: "]", Name: "Next tab", Action: func() { d.step(true) }, Hidden: true},
		{Key: "t", Name: "Theme", Action: demokit.NextTheme},
	}
	// Counter keys only exist while the Counter tab is active.
	if d.tabs.ActiveKeyPeek() == "counter" {
		keybinds = append(keybinds,
			t.Keybind{Key: "+", Name: "Count", Action: func() { d.addToCounter(1) }},
			t.Keybind{Key: "-", Name: "Count", Action: func() { d.addToCounter(-1) }},
		)
	}
	return keybinds
}

func (d *TabDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "tab-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			demokit.Header{Title: "Tabs", Tagline: "A TabView driven by a TabState"},
		},
		Bottom: []t.Widget{
			footer{demo: d},
		},
		Body: t.Row{
			Width:   t.Flex(1),
			Height:  t.Flex(1),
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(1, 1),
			},
			Children: []t.Widget{
				demokit.Fill(tabsArea{demo: d}),
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{demo: d},
						demokit.Fill(orderPanel{demo: d}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// footer shows the keybinds. The App's keybinds depend on the active tab, but
// KeybindBar only refreshes on focus changes, so subscribing to the active key
// here keeps the counter keys in the bar in step with the tab.
type footer struct {
	demo *TabDemo
}

func (f footer) Build(ctx t.BuildContext) t.Widget {
	_ = f.demo.tabs.ActiveKey()
	return demokit.Footer(ctx.Theme())
}

// focusedID returns the ID of the focused widget, or "" if nothing is focused.
func focusedID(ctx t.BuildContext) string {
	if id, ok := ctx.Focused().(t.Identifiable); ok {
		return id.WidgetID()
	}
	return ""
}

// tabsArea holds the TabView. The content panel is titled with the active tab
// and lights up when focus is inside it rather than on the tab bar.
type tabsArea struct {
	demo *TabDemo
}

func (a tabsArea) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := a.demo
	tabs := d.tabs.Tabs()

	if len(tabs) == 0 {
		return t.Column{
			Width:  t.Flex(1),
			Height: t.Flex(1),
			Style:  demokit.PanelStyle(theme, "No tabs", false),
			Children: []t.Widget{
				hint(theme, "Every tab is closed. Press [b $Success]n[/] to open a new one or [b $Warning]u[/] to reopen the last one you closed."),
			},
		}
	}

	activeKey := d.tabs.ActiveKey()
	title := activeKey
	for _, tab := range tabs {
		if tab.Key == activeKey {
			title = tab.Label
		}
	}
	focused := focusedID(ctx)
	contentFocused := focused != "" && focused != tabBarID

	content := demokit.PanelStyle(theme, title, contentFocused)
	content.Padding = t.EdgeInsetsXY(1, 1)

	return t.TabView{
		ID:             tabViewID,
		State:          d.tabs,
		KeybindPattern: t.TabKeybindNumbers,
		AllowReorder:   true,
		Closable:       true,
		OnTabClose:     d.closeTab,
		OnTabChange:    d.onTabChange,
		Style: t.Style{
			Width:  t.Flex(1),
			Height: t.Flex(1),
		},
		TabBarStyle: t.Style{
			Width:           t.Flex(1),
			BackgroundColor: theme.Surface,
		},
		// Tighter than the default padding so more tabs fit; colours are
		// left unset so TabBar fills in its theme defaults.
		TabStyle:       t.Style{Padding: t.EdgeInsetsXY(1, 0)},
		ActiveTabStyle: t.Style{Padding: t.EdgeInsetsXY(1, 0)},
		ContentStyle:   content,
	}
}

// hint is a muted, wrapping line of help text.
func hint(theme t.ThemeData, markup string) t.Widget {
	return t.Text{
		Spans: t.ParseMarkup("[$TextMuted]"+markup+"[/]", theme),
		Wrap:  t.WrapSoft,
		Style: t.Style{Width: t.Flex(1)},
	}
}

// homePage is the first tab's content.
type homePage struct{}

func (homePage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	line := func(markup string) t.Widget { return t.Text{Spans: t.ParseMarkup(markup, theme), Wrap: t.WrapSoft} }
	return t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[b $Primary]Welcome![/] [$TextMuted]This area is a [/][b $Accent]TabView[/][$TextMuted]: a TabBar above a Switcher that shows the active tab's Content.[/]", theme),
				Wrap:  t.WrapSoft,
			},
			t.Column{
				Children: []t.Widget{
					line("[b $Info]h l[/] [$TextMuted]or[/] [b $Info]←→[/]  [$TextMuted]switch tabs[/]"),
					line("[b $Accent]1-9[/]        [$TextMuted]jump to a tab by position[/]"),
					line("[b $Secondary]ctrl+h/l[/]   [$TextMuted]move the active tab[/]"),
					line("[b $Error]ctrl+w[/]     [$TextMuted]close the active tab, or click its ×[/]"),
					line("[b $Success]n[/] [$TextMuted]/[/] [b $Warning]u[/]      [$TextMuted]open a new tab / reopen a closed one[/]"),
				},
			},
			hint(theme, "The tab bar has focus at startup. Press [b $Info]tab[/] to move into a tab's content, and [b $Accent][[ ]][/] to switch tabs from there."),
		},
	}
}

// counterPage is a counter whose value lives in a Signal on the App.
type counterPage struct {
	demo *TabDemo
}

func (c counterPage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := c.demo
	count := d.counter.Get()

	countColour := "$Text"
	switch {
	case count > 0:
		countColour = "$Success"
	case count < 0:
		countColour = "$Error"
	}

	return t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			hint(theme, "Press [b $Success]+[/] / [b $Error]-[/] or use the buttons. Switch tabs and come back: the count is still here."),
			t.Row{
				Spacing:    2,
				CrossAlign: t.CrossAxisCenter,
				Children: []t.Widget{
					t.Button{ID: "counter-dec", Label: " − ", OnPress: func() { d.addToCounter(-1) }},
					t.ParseMarkupToText(fmt.Sprintf("[b %s]%4d[/]", countColour, count), theme),
					t.Button{ID: "counter-inc", Label: " + ", Variant: t.ButtonPrimary, OnPress: func() { d.addToCounter(1) }},
				},
			},
		},
	}
}

// listPage is a List whose cursor lives in a ListState on the App.
type listPage struct {
	demo *TabDemo
}

func (l listPage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			hint(theme, "Press [b $Info]tab[/] to focus the list, then move with [b $Info]↑↓[/] or [b $Info]jk[/]. The cursor survives tab switches."),
			t.List[string]{
				ID:    "fruit-list",
				State: l.demo.fruits,
				Style: t.Style{Width: t.Flex(1)},
			},
		},
	}
}

// infoPage lists what TabBar supports.
type infoPage struct{}

func (infoPage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	item := func(text string) t.Widget {
		return t.Text{Spans: t.ParseMarkup("[$Accent]•[/] "+text, theme), Wrap: t.WrapSoft}
	}
	return t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			hint(theme, "TabBar is a focusable row of tabs. TabView pairs one with a content area. Both read and write a TabState."),
			t.Column{
				Children: []t.Widget{
					item("Keyboard navigation (←/→, h/l)"),
					item("Position keybinds (1-9, alt+1-9 or ctrl+1-9)"),
					item("Reordering with AllowReorder (ctrl+h/l)"),
					item("Closable tabs with a × button and OnTabClose"),
					item("Click a tab to select it"),
					item("Add, insert and remove tabs through TabState"),
				},
			},
		},
	}
}

// notePage is the content of a tab opened at runtime.
type notePage struct {
	number int
}

func (n notePage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{
		Width:   t.Flex(1),
		Spacing: 1,
		Children: []t.Widget{
			t.ParseMarkupToText(fmt.Sprintf("[b $Primary]Note %d[/]", n.number), theme),
			hint(theme, "This tab was added at runtime with [b $Accent]TabState.AddTab[/]. Close it with [b $Error]ctrl+w[/] or its ×, then bring it back with [b $Warning]u[/]."),
		},
	}
}

// statePanel shows the live TabState and each tab's state. It reads the
// signals itself, so only it rebuilds when they change.
type statePanel struct {
	demo *TabDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	tabs := d.tabs.Tabs()
	activeKey := d.tabs.ActiveKey()

	position := "—"
	for i, tab := range tabs {
		if tab.Key == activeKey {
			position = fmt.Sprintf("%d of %d", i+1, len(tabs))
		}
	}
	active := "[$TextMuted]—[/]"
	if activeKey != "" {
		active = fmt.Sprintf("[b $Accent]%q[/]", activeKey)
	}

	focus := "[$TextMuted]—[/]"
	switch id := focusedID(ctx); id {
	case "":
	case tabBarID:
		focus = "[b $Info]tab bar[/]"
	default:
		focus = "[b $Info]content[/]"
	}

	fruit := "—"
	if items, cursor := d.fruits.Items.Get(), d.fruits.CursorIndex.Get(); cursor >= 0 && cursor < len(items) {
		fruit = items[cursor]
	}

	event := "[$TextMuted]—[/]"
	if e := d.lastEvent.Get(); e != "" {
		event = fmt.Sprintf("[b $Warning]%s[/]", e)
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Active", active),
			demokit.StatRow(theme, "Position", fmt.Sprintf("[b $Primary]%s[/]", position)),
			demokit.StatRow(theme, "Focus", focus),
			demokit.StatRow(theme, "Can reopen", fmt.Sprintf("[b $Warning]%d[/]", len(d.closed.Get()))),
			demokit.StatRow(theme, "Counter", fmt.Sprintf("[b $Secondary]%d[/]", d.counter.Get())),
			demokit.StatRow(theme, "Fruit", fmt.Sprintf("[b $Info]%s[/]", fruit)),
			demokit.StatRow(theme, "Last", event),
		},
	}
}

// orderPanel lists TabState's tabs in order with their keys, so reordering,
// closing and opening tabs show up here as they happen.
type orderPanel struct {
	demo *TabDemo
}

func (o orderPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	tabs := o.demo.tabs.Tabs()
	activeKey := o.demo.tabs.ActiveKey()

	rows := make([]t.Widget, 0, len(tabs))
	for i, tab := range tabs {
		marker, labelColour := " ", "$Text"
		if tab.Key == activeKey {
			marker, labelColour = "▸", "$Accent"
		}
		rows = append(rows, demokit.StatRow(theme,
			fmt.Sprintf("%s %d %s", marker, i+1, tab.Label),
			fmt.Sprintf("[%s]%s[/]", labelColour, tab.Key),
		))
	}
	if len(rows) == 0 {
		rows = append(rows, t.ParseMarkupToText("[$TextMuted]No tabs.[/]", theme))
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "TabState", false),
		Children: rows,
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	// Keys are plain text rather than markup so "[" needs no escaping.
	key := func(keys string, colour t.Color, desc string) t.Widget {
		return t.Row{
			Spacing: 1,
			Children: []t.Widget{
				t.Text{Content: keys, Width: t.Cells(8), Style: t.Style{ForegroundColor: colour, Bold: true}},
				t.Text{Content: desc, Style: t.Style{ForegroundColor: theme.TextMuted}},
			},
		}
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("←→ 1-9", theme.Info, "select tab"),
			key("ctrl+h/l", theme.Secondary, "move tab"),
			key("ctrl+w", theme.Error, "close tab"),
			key("n u", theme.Success, "new / reopen tab"),
			key("[ ]", theme.Accent, "switch from content"),
			key("+ -", theme.Warning, "counter"),
			key("tab", theme.Info, "switch focus"),
		},
	}
}
