package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	t "github.com/darrenburns/terma"
)

// Theme names for cycling
var themeNames = []string{
	t.ThemeNameRosePine,
	t.ThemeNameDracula,
	t.ThemeNameTokyoNight,
	t.ThemeNameCatppuccin,
	t.ThemeNameGruvbox,
	t.ThemeNameNord,
	t.ThemeNameSolarized,
	t.ThemeNameKanagawa,
	t.ThemeNameMonokai,
}

var initialItems = []string{"Apple", "Banana", "Cherry"}

const sidebarWidth = 32

// ListDemo demonstrates the List modification APIs.
// Different keys exercise different parts of the ListState API:
//
//	a - Append item to end (A: +10, !: +1000)
//	p - Prepend item to beginning
//	i - Insert item at cursor position
//	d - Delete item at cursor position
//	c - Clear all items
//	r - Reset to initial items
//	t - Cycle theme
//	/ - Focus the filter
//	escape - Clear selection
type ListDemo struct {
	listState        *t.ListState[string]
	scrollState      *t.ScrollState
	filterState      *t.FilterState
	filterInputState *t.TextInputState
	counter          int // For generating unique item names
	themeIndex       t.Signal[int]
}

func NewListDemo() *ListDemo {
	return &ListDemo{
		listState:        t.NewListState(append([]string(nil), initialItems...)),
		scrollState:      t.NewScrollState(),
		filterState:      t.NewFilterState(),
		filterInputState: t.NewTextInputState(""),
		counter:          len(initialItems), // Start after initial items
		themeIndex:       t.NewSignal(0),
	}
}

func (d *ListDemo) nextItem() string {
	d.counter++
	return fmt.Sprintf("Item %d", d.counter)
}

func (d *ListDemo) appendItems(n int) {
	for i := 0; i < n; i++ {
		d.listState.Append(d.nextItem())
	}
}

func (d *ListDemo) cycleTheme() {
	d.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (d *ListDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "a", Name: "Append", Action: func() { d.appendItems(1) }},
		{Key: "A", Name: "Append 10", Action: func() { d.appendItems(10) }, Hidden: true},
		{Key: "!", Name: "Append 1000", Action: func() { d.appendItems(1000) }, Hidden: true},
		{Key: "p", Name: "Prepend", Action: func() { d.listState.Prepend(d.nextItem()) }},
		{Key: "i", Name: "Insert", Action: func() {
			d.listState.InsertAt(d.listState.CursorIndex.Peek(), d.nextItem())
		}},
		{Key: "d", Name: "Delete", Action: func() {
			d.listState.RemoveAt(d.listState.CursorIndex.Peek())
		}},
		{Key: "c", Name: "Clear", Action: d.listState.Clear},
		{Key: "r", Name: "Reset", Action: func() {
			d.listState.SetItems(append([]string(nil), initialItems...))
			d.listState.ClearSelection()
			d.counter = len(initialItems)
		}},
		{Key: "escape", Name: "Deselect", Action: d.listState.ClearSelection, Hidden: true},
		{Key: "/", Name: "Filter", Action: func() { t.RequestFocus("list-filter-input") }},
		{Key: "t", Name: "Theme", Action: d.cycleTheme},
	}
}

func (d *ListDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "list-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			header{themeName: themeNames[d.themeIndex.Get()]},
		},
		Bottom: []t.Widget{
			t.KeybindBar{
				Style: t.Style{
					BackgroundColor: theme.Surface,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
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
					Width:   t.Flex(1),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						filterPanel{demo: d},
						fill(listPanel{demo: d}),
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statsPanel{demo: d},
						fill(selectionPanel{demo: d}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// header is the title bar across the top of the screen.
type header struct {
	themeName string
}

func (h header) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Row{
		Width: t.Flex(1),
		Style: t.Style{
			BackgroundColor: theme.Surface,
			Padding:         t.EdgeInsetsXY(1, 0),
		},
		Children: []t.Widget{
			t.ParseMarkupToText("[b $Primary]≡ List Playground[/]  [$TextMuted]Live edits to a ListState[/]", theme),
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]theme[/] [b $Accent]%s[/]", h.themeName), theme),
		},
	}
}

// fill gives a component the remaining space in its Column. Rows and Columns
// read Flex from their direct children, so a component's own Flex(1) needs a
// plain wrapper to take effect.
func fill(child t.Widget) t.Widget {
	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Children: []t.Widget{child},
	}
}

// panelStyle is the bordered look shared by every panel in the demo.
func panelStyle(theme t.ThemeData, title string, focused bool) t.Style {
	color := theme.Border
	titleColor := "$TextMuted"
	if focused {
		color = theme.FocusRing
		titleColor = "$FocusRing"
	}
	return t.Style{
		BackgroundColor: theme.Background,
		Border:          t.RoundedBorder(color, t.BorderTitleMarkup(fmt.Sprintf("[b %s] %s [/]", titleColor, title))),
		Padding:         t.EdgeInsetsXY(1, 0),
	}
}

// filterPanel holds the text input that narrows the list.
type filterPanel struct {
	demo *ListDemo
}

func (f filterPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := f.demo
	input := t.TextInput{
		ID:          "list-filter-input",
		State:       d.filterInputState,
		Placeholder: "Type to filter items…",
		Width:       t.Flex(1),
		Style: t.Style{
			ForegroundColor: theme.Text,
		},
		OnChange: func(text string) {
			d.filterState.Query.Set(text)
		},
		OnSubmit: func(text string) {
			t.RequestFocus("demo-list")
		},
		ExtraKeybinds: []t.Keybind{
			{
				Key:  "escape",
				Name: "Clear filter",
				Action: func() {
					d.filterInputState.SetText("")
					d.filterState.Query.Set("")
					t.RequestFocus("demo-list")
				},
			},
		},
	}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   panelStyle(theme, "Filter", ctx.IsFocused(input)),
		Children: []t.Widget{
			t.ParseMarkupToText("[$Accent]⌕[/]", theme),
			input,
		},
	}
}

// listPanel shows the scrolling list, or a hint when there is nothing to show.
type listPanel struct {
	demo *ListDemo
}

func (l listPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := l.demo
	list := t.List[string]{
		ID:          "demo-list",
		State:       d.listState,
		ScrollState: d.scrollState,
		Filter:      d.filterState,
		MultiSelect: true,
		Style: t.Style{
			Width: t.Flex(1),
		},
	}

	items := d.listState.Items.Get()
	var empty t.Widget
	switch {
	case len(items) == 0:
		empty = t.ParseMarkupToText("[$TextMuted]The list is empty. Press [b $Success]a[/] to add an item or [b $Warning]r[/] to reset.[/]", theme)
	case len(visibleIndices(items, d.filterState)) == 0:
		empty = t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]Nothing matches [/][b $Warning]%q[/][$TextMuted]. Press [b]esc[/] in the filter to clear it.[/]", d.filterState.QueryText()), theme)
	}

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  panelStyle(theme, "Items", ctx.IsFocused(list)),
		Children: []t.Widget{
			t.ShowWhen(empty != nil, empty),
			t.Scrollable{
				ID:    "list-scroll",
				State: d.scrollState,
				Style: t.Style{
					Width:  t.Flex(1),
					Height: t.Flex(1),
				},
				Child: list,
			},
		},
	}
}

// statsPanel shows live counts. It subscribes to the list's signals itself, so
// moving the cursor only rebuilds this panel rather than the whole app.
type statsPanel struct {
	demo *ListDemo
}

func (s statsPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	items := d.listState.Items.Get()
	cursor := d.listState.CursorIndex.Get()
	selected := len(d.listState.Selection.Get())
	view := visibleIndices(items, d.filterState)

	position := "—"
	for i, idx := range view {
		if idx == cursor {
			position = fmt.Sprintf("%d of %d", i+1, len(view))
			break
		}
	}

	shownColor := "$Text"
	if len(view) < len(items) {
		shownColor = "$Warning"
	}

	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Overview", false),
		Children: []t.Widget{
			statRow(theme, "Items", fmt.Sprintf("[b $Primary]%d[/]", len(items))),
			statRow(theme, "Showing", fmt.Sprintf("[b %s]%d[/]", shownColor, len(view))),
			statRow(theme, "Selected", fmt.Sprintf("[b $Secondary]%d[/]", selected)),
			statRow(theme, "Cursor", fmt.Sprintf("[b $Info]%s[/]", position)),
		},
	}
}

func statRow(theme t.ThemeData, label, valueMarkup string) t.Widget {
	return t.Row{
		Width: t.Flex(1),
		Children: []t.Widget{
			t.Text{Content: label, Style: t.Style{ForegroundColor: theme.TextMuted}},
			t.Spacer{Width: t.Flex(1)},
			t.ParseMarkupToText(valueMarkup, theme),
		},
	}
}

// selectionPanel lists the selected items.
type selectionPanel struct {
	demo *ListDemo
}

func (s selectionPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	selection := s.demo.listState.Selection.Get()
	items := s.demo.listState.Items.Get()

	var content t.Widget
	if len(selection) == 0 {
		content = t.Text{
			Spans: t.ParseMarkup("[$TextMuted]Nothing selected.\nHold [b $Secondary]shift[/] and move to select.[/]", theme),
			Wrap:  t.WrapSoft,
		}
	} else {
		var selected []string
		for i, item := range items {
			if _, ok := selection[i]; ok {
				selected = append(selected, item)
			}
		}
		content = t.Text{
			Content: strings.Join(selected, ", "),
			Wrap:    t.WrapSoft,
			Style:   t.Style{ForegroundColor: theme.Secondary},
		}
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    panelStyle(theme, "Selection", false),
		Children: []t.Widget{content},
	}
}

// keysPanel is a quick reference for the keys the demo responds to.
type keysPanel struct{}

func (keysPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	key := func(keys, colour, desc string) t.Widget {
		return t.ParseMarkupToText(fmt.Sprintf("[b %s]%-7s[/] [$TextMuted]%s[/]", colour, keys, desc), theme)
	}
	return t.Column{
		Width: t.Flex(1),
		Style: panelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("↑↓ jk", "$Info", "move cursor"),
			key("⇧↑↓", "$Secondary", "extend selection"),
			key("a A !", "$Success", "append 1 / 10 / 1000"),
			key("p i", "$Success", "prepend / insert"),
			key("d c", "$Error", "delete / clear all"),
			key("r", "$Warning", "reset items"),
			key("/", "$Accent", "filter items"),
			key("tab", "$Info", "switch focus"),
		},
	}
}

// visibleIndices returns the source indices of the items that pass the filter,
// matching the order the List displays them in.
func visibleIndices(items []string, filter *t.FilterState) []int {
	query := filter.QueryText()
	options := filter.Options()
	return t.ApplyFilter(items, query, func(item string, q string) t.MatchResult {
		return t.MatchString(item, q, options)
	}).Indices
}

func main() {
	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	t.SetTheme(themeNames[0])
	app := NewListDemo()
	t.RequestFocus("demo-list")
	_ = t.InitLogger()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
