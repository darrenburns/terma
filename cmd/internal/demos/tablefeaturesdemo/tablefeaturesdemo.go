// Package tablefeaturesdemo demonstrates sorting, resizing and frozen table panes.
package tablefeaturesdemo

import (
	"cmp"
	"fmt"
	"slices"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
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
		{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme},
		{Key: "m", Name: "Modal", Action: func() { a.modal.Set(!a.modal.Get()); t.RequestFocus("people") }},
		{Key: "b", Name: "Bottom", Action: func() {
			a.scroll.SetOffset(10000)
			a.message.Set("Scrolled to bottom; tall row tail must remain reachable")
		}},
		{Key: "t", Name: "Top", Action: func() { a.scroll.SetOffset(0); a.message.Set("Scrolled to top") }},
		{Key: "/", Name: "Filter", Action: func() { t.RequestFocus("filter") }},
		{Key: "escape", Name: "Table", Action: func() { t.RequestFocus("people") }},
		{Key: "space", Name: "Select cell", Action: func() { a.rows.ToggleSelection(a.rows.CursorIndex.Peek()*4 + a.rows.CursorColumn.Peek()) }},
		{Key: "r", Name: "Reverse data", Action: func() {
			rows := slices.Clone(a.rows.GetRows())
			slices.Reverse(rows)
			a.rows.SetRows(rows)
			message := "Data order reversed; cursor and selections follow row IDs"
			if a.rows.Sort.Peek().Direction != t.TableSortNone {
				message = "Data reversed; column sort still applies. Ties follow data order; selections keep their row IDs."
			}
			a.message.Set(message)
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
		t.Row{Spacing: 1, Children: []t.Widget{t.Text{Content: "Filter:"}, t.TextInput{ID: "filter", State: a.input, Placeholder: "Try Ada or no-match", OnChange: func(s string) { a.filter.Query.Set(s) }, Style: t.Style{Width: t.Cells(28)}}, t.Text{Content: "/ filter · Esc table · r reverse data · x empty"}}},
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
				// Custom cell renderers own their colours. Match Table's default
				// styling: a focused cursor wins over the persistent selection.
				style := t.Style{ForegroundColor: theme.Text}
				if selected {
					style.BackgroundColor = theme.Selection
				}
				if active && ctx.IsFocused(t.Table[person]{ID: "people"}) {
					style.BackgroundColor = theme.ActiveCursor
					style.ForegroundColor = theme.SelectionText
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
			order := "data order"
			if sort.Direction == t.TableSortAscending {
				order = sort.ColumnID + " ascending"
			} else if sort.Direction == t.TableSortDescending {
				order = sort.ColumnID + " descending"
			}
			return fmt.Sprintf("Cursor: %s (%s) · selected cells: %d · view: %s · scroll %d,%d", p.Name, p.ID, len(a.rows.Selection.Peek()), order, a.scroll.OffsetX.Get(), a.scroll.Offset.Get())
		}),
		t.ComputedText(strings.Repeat(" ", 120), func() string { return a.message.Get() }),
		t.Text{Content: "Arrows navigate · Shift+arrows select · Ctrl+S sort column · Ctrl+←/→ resize · Ctrl+R reset · Alt+←/→ pan"},
	}}
	if a.modal.Get() {
		return t.Dialog{ID: "table-probe-dialog", Visible: true, Title: "Table probe", Style: t.Style{Width: t.Percent(94)}, Content: body, OnDismiss: func() { a.modal.Set(false) }}
	}
	return t.Dock{Style: t.Style{BackgroundColor: theme.Background}, Top: []t.Widget{demokit.Header{Title: "Table controls", Tagline: "Sort, resize and freeze"}}, Bottom: []t.Widget{demokit.Footer(theme)}, Body: body}
}

// Info describes this demo in the gallery.
var Info = demokit.Info{Key: "tablefeatures", Title: "Table controls", Description: "Sorting, column resizing, frozen panes and stable row identity"}

// New creates a gallery or standalone demo.
func New() demokit.Demo { return newApp() }

// NewProbe creates the tall-row viewport verification demo.
func NewProbe() demokit.Demo {
	a := newApp()
	a.probe = true
	a.rows.SetRows([]person{{ID: "tall", Name: "Tall row", Team: "Platform", Score: 10, Notes: "line1\nline2\nline3\nline4\nline5\nline6\nline7\nTAIL"}})
	a.message.Set("Probe: Right×3 reveals Notes; b bottom / t top; m modal")
	return a
}
