package main

import (
	"cmp"
	"flag"
	"fmt"
	t "github.com/darrenburns/terma"
	"log"
	"slices"
	"strings"
)

type person struct {
	ID, Name, Team, Notes string
	Score                 int
}
type app struct {
	rows    *t.TableState[person]
	scroll  *t.ScrollState
	filter  *t.FilterState
	input   *t.TextInputState
	message t.Signal[string]
	probe   bool
	modal   t.Signal[bool]
}

func newApp() *app {
	names := []string{"Zoe", "Ada", "Mia", "Leo", "Uma", "Kai", "Eli", "Bea", "Noah", "Ivy", "Finn", "Luna", "Theo", "Aria", "Otis", "Ruby", "Max", "Nora", "Hugo", "Rose", "Luca", "Esme", "Alex", "Milo", "Orla", "Seth", "Alma", "Cleo", "Ezra", "Tess", "Owen", "Isla"}
	rows := make([]person, len(names))
	for i, name := range names {
		rows[i] = person{ID: fmt.Sprintf("p%02d", i), Name: name, Team: []string{"Platform", "Design", "Research"}[i%3], Score: []int{30, 20, 20, 40, 25, 35, 28, 22}[i%8], Notes: fmt.Sprintf("Stable record p%02d; detail column extends beyond narrow screens", i)}
	}
	return &app{rows: t.NewTableStateWithRowID(rows, func(p person) string { return p.ID }), scroll: t.NewScrollState(), filter: t.NewFilterState(), input: t.NewTextInputState(""), message: t.NewSignal("Click headers to sort; drag their last cell to resize"), modal: t.NewSignal(false)}
}
func (a *app) InitialFocus() string { return "people" }
func (a *app) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "m", Name: "Modal", Action: func() { a.modal.Set(!a.modal.Get()); t.RequestFocus("people") }},
		{Key: "b", Name: "Bottom", Action: func() {
			a.scroll.SetOffset(10000)
			a.message.Set("Scrolled to bottom; tall row tail must remain reachable")
		}},
		{Key: "t", Name: "Top", Action: func() { a.scroll.SetOffset(0); a.message.Set("Scrolled to top") }},
		{Key: "/", Name: "Filter", Action: func() { t.RequestFocus("filter") }},
		{Key: "escape", Name: "Table", Action: func() { t.RequestFocus("people") }},
		{Key: "space", Name: "Select cell", Action: func() { a.rows.ToggleSelection(a.rows.CursorIndex.Peek()*4 + a.rows.CursorColumn.Peek()) }},
		{Key: "r", Name: "Reverse source", Action: func() {
			rows := slices.Clone(a.rows.GetRows())
			slices.Reverse(rows)
			a.rows.SetRows(rows)
			a.message.Set("Source reversed: cursor and selected records retained by ID")
		}},
		{Key: "x", Name: "Empty/reset", Action: func() {
			if a.rows.RowCount() == 0 {
				a.rows.SetRows(newApp().rows.GetRows())
			} else {
				a.rows.SetRows(nil)
			}
		}},
	}
}
func (a *app) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	tableHeight := t.Flex(1)
	if a.probe {
		tableHeight = t.Cells(5)
	}
	body := t.Column{Spacing: 1, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1), Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
		t.Text{Content: "TABLE LAB · sorting, stable identity, resizing & frozen panes", Style: t.Style{Bold: true, ForegroundColor: theme.Primary}},
		t.Row{Spacing: 1, Children: []t.Widget{t.Text{Content: "Filter:"}, t.TextInput{ID: "filter", State: a.input, Placeholder: "Try Ada or no-match", OnChange: func(s string) { a.filter.Query.Set(s) }, Style: t.Style{Width: t.Cells(28)}}, t.Text{Content: "/ filter · Esc table · r reverse · x empty"}}},
		t.Table[person]{ID: "people", State: a.rows, ScrollState: a.scroll, Filter: a.filter, FrozenHeader: true, FrozenColumns: 1, MultiSelect: true, SelectionMode: t.TableSelectionCursor, ColumnSpacing: 1, Style: t.Style{Width: t.Flex(1), Height: tableHeight},
			Columns:     []t.TableColumn{{ID: "name", Header: t.Text{Content: "Name"}, Width: t.Cells(16), Resizable: true, MinWidth: 6, MaxWidth: 30}, {ID: "score", Header: t.Text{Content: "Score"}, Width: t.Cells(10), Resizable: true, MinWidth: 6}, {ID: "team", Header: t.Text{Content: "Team"}, Width: t.Cells(18), Resizable: true, MinWidth: 8}, {ID: "notes", Header: t.Text{Content: "Notes"}, Width: t.Cells(65), Resizable: true, MinWidth: 12}},
			Comparators: map[string]func(person, person) int{"name": func(a, b person) int { return strings.Compare(a.Name, b.Name) }, "score": func(a, b person) int { return cmp.Compare(a.Score, b.Score) }, "team": func(a, b person) int { return strings.Compare(a.Team, b.Team) }},
			MatchCell: func(p person, _, col int, q string, o t.FilterOptions) t.MatchResult {
				value := p.Name
				if col == 1 {
					value = fmt.Sprint(p.Score)
				}
				if col == 2 {
					value = p.Team
				}
				if col == 3 {
					value = p.Notes
				}
				return t.MatchString(value, q, o)
			},
			RenderCell: func(p person, _, col int, active, selected bool) t.Widget {
				value := p.Name
				if col == 1 {
					value = fmt.Sprint(p.Score)
				}
				if col == 2 {
					value = p.Team
				}
				if col == 3 {
					value = p.Notes
				}
				style := t.Style{}
				if active {
					style.BackgroundColor = theme.Primary
					style.ForegroundColor = theme.Background
				}
				if selected {
					style.Bold = true
					style.BackgroundColor = theme.Surface
				}
				return t.Text{Content: value, Style: style}
			},
			OnSelect: func(p person) { a.message.Set("Opened " + p.Name + " (" + p.ID + ")") },
		},
		t.ComputedText(strings.Repeat(" ", 120), func() string {
			a.rows.Rows.Get()
			a.rows.CursorIndex.Get()
			a.rows.Selection.Get()
			sort := a.rows.Sort.Get()
			p, _ := a.rows.SelectedRow()
			return fmt.Sprintf("Cursor: %s (%s) · selected cells: %d · sort: %s/%d · scroll %d,%d", p.Name, p.ID, len(a.rows.Selection.Peek()), sort.ColumnID, sort.Direction, a.scroll.OffsetX.Get(), a.scroll.Offset.Get())
		}),
		t.SignalText(a.message, func(s string) string { return s }),
		t.Text{Content: "Arrows navigate · Shift+arrows select · Ctrl+S sort column · Ctrl+←/→ resize · Ctrl+R reset · Alt+←/→ pan"},
	}}
	if a.modal.Get() {
		return t.Dialog{ID: "table-probe-dialog", Visible: true, Title: "Table probe", Style: t.Style{Width: t.Percent(94)}, Content: body, OnDismiss: func() { a.modal.Set(false) }}
	}
	return body
}
func main() {
	probe := flag.Bool("probe", false, "Run tall-row viewport verification")
	flag.Parse()
	demo := newApp()
	if *probe {
		demo.probe = true
		demo.rows.SetRows([]person{{ID: "tall", Name: "Tall row", Team: "Platform", Score: 10, Notes: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nTAIL"}})
		demo.message.Set("Probe: Right×3 reveals Notes; b bottom / t top; m modal")
	}
	t.RequestFocus("people")
	if err := t.Run(demo); err != nil {
		log.Fatal(err)
	}
}
