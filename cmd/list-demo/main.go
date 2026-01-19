package main

import (
	"fmt"
	"log"
	"strings"

	t "terma"
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

// ListDemo demonstrates the List modification APIs.
// Different keys exercise different parts of the ListState API:
//
//	a - Append item to end
//	p - Prepend item to beginning
//	i - Insert item at cursor position
//	d - Delete item at cursor position
//	c - Clear all items
//	r - Reset to initial items
//	t - Cycle theme
//	q - Quit
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
		listState: t.NewListState([]string{
			"Apple",
			"Banana",
			"Cherry",
		}),
		scrollState:      t.NewScrollState(),
		filterState:      t.NewFilterState(),
		filterInputState: t.NewTextInputState(""),
		counter:          3, // Start after initial items
		themeIndex:       t.NewSignal(0),
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
		{Key: "a", Name: "Append", Action: func() {
			d.counter++
			d.listState.Append(fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "A", Name: "Append 10", Action: func() {
			for i := 0; i < 10; i++ {
				d.counter++
				d.listState.Append(fmt.Sprintf("Item %d", d.counter))
			}
		}},
		{Key: "!", Name: "Append 1000", Action: func() {
			for i := 0; i < 1000; i++ {
				d.counter++
				d.listState.Append(fmt.Sprintf("Item %d", d.counter))
			}
		}},
		{Key: "p", Name: "Prepend", Action: func() {
			d.counter++
			d.listState.Prepend(fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "i", Name: "Insert at cursor", Action: func() {
			d.counter++
			idx := d.listState.CursorIndex.Peek()
			d.listState.InsertAt(idx, fmt.Sprintf("Item %d", d.counter))
		}},
		{Key: "d", Name: "Delete at cursor", Action: func() {
			idx := d.listState.CursorIndex.Peek()
			d.listState.RemoveAt(idx)
		}},
		{Key: "c", Name: "Clear all", Action: func() {
			d.listState.Clear()
		}},
		{Key: "r", Name: "Reset", Action: func() {
			d.listState.SetItems([]string{"Apple", "Banana", "Cherry"})
			d.counter = 3
		}},
		{Key: "t", Name: "Next theme", Action: d.cycleTheme},
	}
}

func (d *ListDemo) buildSelectionSummary(theme t.ThemeData) t.Widget {
	// Subscribe to selection changes with .Get()
	selection := d.listState.Selection.Get()
	if len(selection) == 0 {
		return t.Text{
			Spans: t.ParseMarkup("[$TextMuted]No items selected[/]", theme),
		}
	}

	// Get the actual selected items (also subscribe to item changes)
	items := d.listState.Items.Get()
	var selected []string
	for i, item := range items {
		if _, ok := selection[i]; ok {
			selected = append(selected, item)
		}
	}

	summary := strings.Join(selected, ", ")
	if len(summary) > 50 {
		summary = summary[:47] + "..."
	}

	return t.Text{
		Spans: t.ParseMarkup(fmt.Sprintf("[b $Secondary]Selected (%d): [/]%s", len(selected), summary), theme),
	}
}

func (d *ListDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := d.themeIndex.Get()
	currentTheme := themeNames[themeIdx]

	return t.Column{
		ID:      "list-demo-root",
		Height:  t.Flex(1),
		Spacing: 1,
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			// Header
			t.Text{
				Content: "List Modification Demo",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},

			// Theme indicator
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]Theme: [/][$Accent]%s[/][$TextMuted] (press t to change)[/]", currentTheme), theme),
			},

			// Instructions - navigation
			t.Text{
				Spans: t.ParseMarkup("Navigate: [b $Info]↑/↓[/] or [b $Info]j/k[/] | Select: [b $Secondary]Shift+↑/↓[/] to extend", theme),
			},

			// Instructions - modifications
			t.Text{
				Spans: t.ParseMarkup("Modify: [b $Success]a[/]ppend [b $Success]A[/]+10 [b $Success]![/]+1000 [b $Success]p[/]repend [b $Success]i[/]nsert [b $Error]d[/]elete [b $Error]c[/]lear [b $Warning]r[/]eset", theme),
			},
			t.Row{
				Spacing:    1,
				CrossAlign: t.CrossAxisCenter,
				Children: []t.Widget{
					t.Text{
						Content: "Filter:",
						Style: t.Style{
							ForegroundColor: theme.TextMuted,
						},
					},
					t.TextInput{
						ID:          "list-filter-input",
						State:       d.filterInputState,
						Placeholder: "Type to filter...",
						Width:       t.Flex(1),
						Style: t.Style{
							BackgroundColor: theme.Surface,
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
					},
					t.Text{
						Spans: t.ParseMarkup("[$TextMuted]Tab to move focus[/]", theme),
					},
				},
			},

			// The list with scrolling
			t.Scrollable{
				ID:           "list-scroll",
				State:        d.scrollState,
				Height:       t.Cells(10),
				DisableFocus: true,
				Style: t.Style{
					Border:  t.RoundedBorder(theme.Primary, t.BorderTitle("Items")),
					Padding: t.EdgeInsetsXY(1, 0),
				},
				Child: t.List[string]{
					ID:          "demo-list",
					State:       d.listState,
					ScrollState: d.scrollState,
					Filter:      d.filterState,
					MultiSelect: true,
				},
			},

			// Status showing item count and cursor
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("Items: [b $Warning]%d[/] | Cursor: [b $Info]%d[/] | Press [b $Error]Ctrl+C[/] to quit", d.listState.ItemCount(), d.listState.CursorIndex.Get()+1), theme),
			},

			// ActiveCursor summary
			d.buildSelectionSummary(theme),
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	app := NewListDemo()
	//t.InitDebug()
	_ = t.InitLogger()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
