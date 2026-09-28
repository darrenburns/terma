// Package switcherdemo demonstrates Switcher: one keyed child at a time, with
// each page's state kept outside it.
package switcherdemo

import (
	"fmt"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "switcher",
	Title:       "Switcher",
	Description: "Keyed pages shown one at a time, each keeping its state",
}

// SwitcherDemo demonstrates the Switcher widget: three keyed pages, only one of
// which is built at a time. Each page keeps its state (list cursor, picked
// item, counter) in the App, so switching away and back picks up where you
// left off.
//
//	1 2 3  - Show the Fruits, Colours or Counter page
//	[ ]    - Previous / next page
//	↑↓ jk  - Move the list cursor
//	enter  - Pick the item under the cursor
//	+ -    - Change the counter (Counter page)
//	0      - Reset the counter (Counter page)
//	t      - Cycle theme
//	tab    - Move focus
type SwitcherDemo struct {
	active t.Signal[string]

	fruits       *t.ListState[string]
	fruitsPicked t.Signal[string]

	colours       *t.ListState[string]
	coloursPicked t.Signal[string]

	counter t.Signal[int]
}

// page describes one Switcher child: its key, its label in the page strip, and
// the widget that should take focus when it is shown.
type page struct {
	key     string
	label   string
	focusID string
}

var pages = []page{
	{key: "fruits", label: "Fruits", focusID: "fruits-list"},
	{key: "colours", label: "Colours", focusID: "colours-list"},
	{key: "counter", label: "Counter", focusID: "counter-inc"},
}

// themeColours are the swatches on the Colours page. They are theme colour
// names, so markup such as "[$Primary]" draws them in the current theme.
var themeColours = []string{
	"Primary", "Secondary", "Accent", "Success", "Warning",
	"Error", "Info", "Text", "TextMuted", "Border",
}

const sidebarWidth = 32

// New creates the demo.
func New() demokit.Demo {
	return &SwitcherDemo{
		active: t.NewSignal(pages[0].key),
		fruits: t.NewListState([]string{
			"Apple", "Banana", "Cherry", "Date", "Elderberry",
			"Fig", "Grape", "Honeydew", "Kiwi", "Lemon",
		}),
		fruitsPicked:  t.NewSignal(""),
		colours:       t.NewListState(append([]string(nil), themeColours...)),
		coloursPicked: t.NewSignal(""),
		counter:       t.NewSignal(0),
	}
}

func (d *SwitcherDemo) InitialFocus() string { return pages[0].focusID }

// show switches the Switcher to the page with the given key and focuses it.
func (d *SwitcherDemo) show(key string) {
	for _, p := range pages {
		if p.key == key {
			d.active.Set(key)
			t.RequestFocus(p.focusID)
			return
		}
	}
}

// step moves delta pages forwards or backwards, wrapping at either end.
func (d *SwitcherDemo) step(delta int) {
	current := d.active.Peek()
	for i, p := range pages {
		if p.key == current {
			d.show(pages[(i+delta+len(pages))%len(pages)].key)
			return
		}
	}
}

func (d *SwitcherDemo) addToCounter(n int) {
	d.counter.Update(func(c int) int { return c + n })
}

func (d *SwitcherDemo) Keybinds() []t.Keybind {
	keybinds := make([]t.Keybind, 0, len(pages)+5)
	for i, p := range pages {
		key := p.key
		keybinds = append(keybinds, t.Keybind{
			Key:    fmt.Sprintf("%d", i+1),
			Name:   p.label,
			Action: func() { d.show(key) },
		})
	}
	keybinds = append(keybinds,
		t.Keybind{Key: "[", Name: "Prev page", Action: func() { d.step(-1) }, Hidden: true},
		t.Keybind{Key: "]", Name: "Next page", Action: func() { d.step(1) }, Hidden: true},
		t.Keybind{Key: "t", Name: "Theme", Action: demokit.NextTheme},
	)

	// Counter keys only exist while the Counter page is showing.
	if d.active.Peek() == "counter" {
		keybinds = append(keybinds,
			t.Keybind{Key: "+", Name: "Increment", Action: func() { d.addToCounter(1) }},
			t.Keybind{Key: "-", Name: "Decrement", Action: func() { d.addToCounter(-1) }},
			t.Keybind{Key: "0", Name: "Reset", Action: func() { d.counter.Set(0) }, Hidden: true},
		)
	}
	return keybinds
}

func (d *SwitcherDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "switcher-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			demokit.Header{
				Title:   "Switcher",
				Tagline: "One keyed child at a time, state kept in the App",
				Right:   fmt.Sprintf("[$TextMuted]pages[/] [b $Accent]%d[/]", len(pages)),
			},
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
				t.Column{
					Width:  t.Flex(1),
					Height: t.Flex(1),
					Children: []t.Widget{
						pageStrip{demo: d},
						demokit.Fill(pageSwitcher{demo: d}),
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statePanel{demo: d},
						demokit.Fill(aboutPanel{}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// footer shows the keybinds. The App's keybinds depend on the active page, but
// KeybindBar only refreshes on focus changes, so subscribing to the page here
// keeps the counter keys in the bar in step with the page.
type footer struct {
	demo *SwitcherDemo
}

func (f footer) Build(ctx t.BuildContext) t.Widget {
	_ = f.demo.active.Get()
	return demokit.Footer(ctx.Theme())
}

// pageStrip is a hand-built tab strip: the Switcher itself has no chrome, so
// the App draws its own labels and decides which key is Active.
type pageStrip struct {
	demo *SwitcherDemo
}

func (s pageStrip) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	active := d.active.Get()

	children := make([]t.Widget, 0, len(pages))
	for i, p := range pages {
		key := p.key
		style := t.Style{
			ForegroundColor: theme.TextMuted,
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(2, 0),
		}
		number := "$Accent"
		if key == active {
			style.ForegroundColor = theme.TextOnAccent
			style.BackgroundColor = theme.Accent
			number = "$TextOnAccent"
		}
		children = append(children, t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[b %s]%d[/] %s", number, i+1, p.label), theme),
			Style: style,
			Click: func(t.MouseEvent) { d.show(key) },
		})
	}
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
		},
		Children: children,
	}
}

// pageSwitcher holds the Switcher. Only the child under Active is built, so the
// other pages cost nothing until they are shown again.
type pageSwitcher struct {
	demo *SwitcherDemo
}

func (p pageSwitcher) Build(ctx t.BuildContext) t.Widget {
	d := p.demo
	return t.Switcher{
		Active: d.active.Get(),
		Style: t.Style{
			Width:  t.Flex(1),
			Height: t.Flex(1),
		},
		Children: map[string]t.Widget{
			"fruits":  demokit.Fill(fruitsPage{demo: d}),
			"colours": demokit.Fill(coloursPage{demo: d}),
			"counter": demokit.Fill(counterPage{demo: d}),
		},
	}
}

// fruitsPage is a plain List whose cursor lives in a ListState on the App.
type fruitsPage struct {
	demo *SwitcherDemo
}

func (f fruitsPage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := f.demo
	list := t.List[string]{
		ID:       "fruits-list",
		State:    d.fruits,
		OnSelect: func(item string) { d.fruitsPicked.Set(item) },
		Style:    t.Style{Width: t.Flex(1)},
	}
	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "Fruits", ctx.IsFocused(list)),
		Children: []t.Widget{
			hint(theme, "Move with [b $Info]↑↓[/] and pick with [b $Success]enter[/]. Switch away and back: the cursor stays put."),
			list,
		},
	}
}

// coloursPage renders each theme colour as a swatch with a custom RenderItem.
type coloursPage struct {
	demo *SwitcherDemo
}

func (c coloursPage) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := c.demo
	list := t.List[string]{
		ID:       "colours-list",
		State:    d.colours,
		OnSelect: func(item string) { d.coloursPicked.Set(item) },
		Style:    t.Style{Width: t.Flex(1)},
	}
	focused := ctx.IsFocused(list)
	list.RenderItem = func(name string, active, _ bool) t.Widget {
		style := t.Style{Width: t.Flex(1), ForegroundColor: theme.Text}
		if active && focused {
			style.BackgroundColor = theme.ActiveCursor
			style.ForegroundColor = theme.SelectionText
		}
		return t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[$%s]██[/] %s", name, name), theme),
			Style: style,
		}
	}
	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "Colours", focused),
		Children: []t.Widget{
			hint(theme, "The current theme's palette, drawn by a custom [b $Accent]RenderItem[/]. Pick one with [b $Success]enter[/]."),
			list,
		},
	}
}

// counterPage is a counter whose value lives in a Signal on the App.
type counterPage struct {
	demo *SwitcherDemo
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

	dec := t.Button{ID: "counter-dec", Label: " − ", OnPress: func() { d.addToCounter(-1) }}
	inc := t.Button{ID: "counter-inc", Label: " + ", Variant: t.ButtonPrimary, OnPress: func() { d.addToCounter(1) }}
	reset := t.Button{ID: "counter-reset", Label: "Reset", OnPress: func() { d.counter.Set(0) }}
	focused := ctx.IsFocused(dec) || ctx.IsFocused(inc) || ctx.IsFocused(reset)

	return t.Column{
		Width:   t.Flex(1),
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "Counter", focused),
		Children: []t.Widget{
			hint(theme, "Press [b $Success]+[/] / [b $Error]-[/] or use the buttons. [b $Warning]0[/] resets. The count survives page switches."),
			t.Row{
				Spacing:    2,
				CrossAlign: t.CrossAxisCenter,
				Children: []t.Widget{
					dec,
					t.ParseMarkupToText(fmt.Sprintf("[b %s]%4d[/]", countColour, count), theme),
					inc,
					reset,
				},
			},
		},
	}
}

// hint is a muted, wrapping line of help text at the top of a page.
func hint(theme t.ThemeData, markup string) t.Widget {
	return t.Text{
		Spans: t.ParseMarkup("[$TextMuted]"+markup+"[/]", theme),
		Wrap:  t.WrapSoft,
		Style: t.Style{Width: t.Flex(1)},
	}
}

// statePanel shows the state of every page, including the ones the Switcher
// is not currently building. It reads the signals itself, so only it rebuilds
// when they change.
type statePanel struct {
	demo *SwitcherDemo
}

func (s statePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo

	picked := func(v, colour string) string {
		if v == "" {
			return "[$TextMuted]—[/]"
		}
		return fmt.Sprintf("[b %s]%s[/]", colour, v)
	}
	colourPicked := "[$TextMuted]—[/]"
	if name := d.coloursPicked.Get(); name != "" {
		colourPicked = fmt.Sprintf("[$%s]██[/] [b]%s[/]", name, name)
	}

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "State", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Active", fmt.Sprintf("[b $Accent]%q[/]", d.active.Get())),
			demokit.StatRow(theme, "Fruit cursor", fmt.Sprintf("[b $Info]%s[/]", cursorItem(d.fruits))),
			demokit.StatRow(theme, "Fruit picked", picked(d.fruitsPicked.Get(), "$Success")),
			demokit.StatRow(theme, "Colour cursor", fmt.Sprintf("[b $Info]%s[/]", cursorItem(d.colours))),
			demokit.StatRow(theme, "Colour picked", colourPicked),
			demokit.StatRow(theme, "Counter", fmt.Sprintf("[b $Primary]%d[/]", d.counter.Get())),
		},
	}
}

// cursorItem returns the item under a list's cursor, subscribing to both the
// items and the cursor.
func cursorItem(state *t.ListState[string]) string {
	items := state.Items.Get()
	cursor := state.CursorIndex.Get()
	if cursor < 0 || cursor >= len(items) {
		return "—"
	}
	return items[cursor]
}

// aboutPanel explains what the Switcher is doing.
type aboutPanel struct{}

func (aboutPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, "How it works", false),
		Children: []t.Widget{
			t.Text{
				Spans: t.ParseMarkup("[$TextMuted]Only the child under [/][b $Accent]Active[/][$TextMuted] is built. Hidden pages keep nothing themselves: their state lives in Signals and a ListState on the App.[/]", theme),
				Wrap:  t.WrapSoft,
			},
		},
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
				t.Text{Content: keys, Width: t.Cells(7), Style: t.Style{ForegroundColor: colour, Bold: true}},
				t.Text{Content: desc, Style: t.Style{ForegroundColor: theme.TextMuted}},
			},
		}
	}
	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("1 2 3", theme.Accent, "show page"),
			key("[ ]", theme.Accent, "previous / next page"),
			key("↑↓ jk", theme.Info, "move cursor"),
			key("enter", theme.Success, "pick item"),
			key("+ - 0", theme.Warning, "count up/down/reset"),
			key("tab", theme.Info, "switch focus"),
		},
	}
}
