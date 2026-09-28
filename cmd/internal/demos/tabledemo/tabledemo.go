// Package tabledemo demonstrates Table: variable-height cells, filtering with
// match highlighting, three selection modes and multi-select.
package tabledemo

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "table",
	Title:       "Table",
	Description: "Cell, row and column selection, filtering and edits in a TableState",
}

// columnNames are the table's column headers, in display order.
var columnNames = []string{"Service", "Owner", "Status", "Notes"}

const sidebarWidth = 30

// selectionModes is the order the m key cycles through.
var selectionModes = []t.TableSelectionMode{
	t.TableSelectionCursor,
	t.TableSelectionRow,
	t.TableSelectionColumn,
}

type TableRow struct {
	Service string
	Owner   string
	Status  string
	Notes   string
}

// TableDemo showcases the Table widget: variable-height cells, filtering with
// match highlighting, three selection modes and multi-select.
//
//	↑↓ jk      - Move the cursor between rows (←→ hl between cells in cell mode)
//	shift+move - Extend the selection
//	space      - Toggle selection of the current cell, row or column
//	enter      - Open the current row
//	m          - Cycle selection mode (cell, row, column)
//	a          - Append a row
//	p          - Prepend a row
//	i          - Insert a row at the cursor
//	d          - Delete the row at the cursor
//	c          - Clear all rows
//	r          - Reset to the initial rows
//	escape     - Clear selection
//	/          - Focus the filter
//	t          - Cycle theme
type TableDemo struct {
	tableState       *t.TableState[TableRow]
	scrollState      *t.ScrollState
	filterState      *t.FilterState
	filterInputState *t.TextInputState
	counter          int // For generating unique service names
	selectionMode    t.Signal[t.TableSelectionMode]
	opened           t.Signal[string] // Service of the last row opened with enter
}

// New creates the demo.
func New() demokit.Demo {
	rows := defaultRows()
	return &TableDemo{
		tableState:       t.NewTableState(rows),
		scrollState:      t.NewScrollState(),
		filterState:      t.NewFilterState(),
		filterInputState: t.NewTextInputState(""),
		counter:          len(rows),
		selectionMode:    t.NewSignal(t.TableSelectionCursor),
		opened:           t.NewSignal(""),
	}
}

func (d *TableDemo) InitialFocus() string { return "demo-table" }

func (d *TableDemo) cycleSelectionMode() {
	d.selectionMode.Update(func(mode t.TableSelectionMode) t.TableSelectionMode {
		i := slices.Index(selectionModes, mode)
		return selectionModes[(i+1)%len(selectionModes)]
	})
}

func (d *TableDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "space", Name: "Toggle", Action: d.toggleSelection},
		{Key: "m", Name: "Mode", Action: d.cycleSelectionMode},
		{Key: "a", Name: "Append", Action: func() { d.tableState.Append(d.nextRow()) }},
		{Key: "p", Name: "Prepend", Action: func() { d.tableState.Prepend(d.nextRow()) }, Hidden: true},
		{Key: "i", Name: "Insert", Action: func() { d.tableState.InsertAt(d.cursorRow(), d.nextRow()) }},
		{Key: "d", Name: "Delete", Action: d.deleteRow},
		{Key: "c", Name: "Clear", Action: d.tableState.Clear, Hidden: true},
		{Key: "r", Name: "Reset", Action: d.reset},
		{Key: "escape", Name: "Deselect", Action: func() {
			d.tableState.ClearSelection()
			d.tableState.ClearAnchor()
		}, Hidden: true},
		{Key: "/", Name: "Filter", Action: func() { t.RequestFocus("table-filter-input") }},
		{Key: "t", Name: "Theme", Action: demokit.NextTheme},
	}
}

// cursorRow returns the row the cursor is drawn on, without subscribing.
func (d *TableDemo) cursorRow() int {
	rows := d.tableState.Rows.Peek()
	view := visibleRows(rows, d.filterState.PeekQuery(), d.filterState.PeekOptions())
	return displayedCursor(d.tableState.CursorIndex.Peek(), view)
}

func (d *TableDemo) nextRow() TableRow {
	d.counter++
	return makeRow(d.counter)
}

// toggleSelection toggles the cell, row or column under the cursor. Selection
// keys depend on the mode: row indices, column indices, or
// row*columnCount+column for cells.
func (d *TableDemo) toggleSelection() {
	row := d.cursorRow()
	col := d.tableState.CursorColumn.Peek()
	switch d.selectionMode.Peek() {
	case t.TableSelectionColumn:
		d.tableState.ToggleSelection(col)
	case t.TableSelectionRow:
		d.tableState.ToggleSelection(row)
	default:
		d.tableState.ToggleSelection(row*len(columnNames) + col)
	}
}

func (d *TableDemo) deleteRow() {
	d.tableState.RemoveAt(d.cursorRow())
}

func (d *TableDemo) reset() {
	rows := defaultRows()
	d.tableState.SetRows(rows)
	d.tableState.ClearSelection()
	d.tableState.ClearAnchor()
	d.tableState.SelectFirst()
	d.tableState.SelectColumn(0)
	d.scrollState.SetOffset(0)
	d.opened.Set("")
	d.counter = len(rows)
}

func (d *TableDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "table-demo-root",
		Style: t.Style{
			BackgroundColor: theme.Background,
		},
		Top: []t.Widget{
			demokit.Header{Title: "Table Playground", Tagline: "Cells, rows and columns from a TableState"},
		},
		Bottom: []t.Widget{demokit.Footer(theme)},
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
						demokit.Fill(tablePanel{demo: d}),
					},
				},
				t.Column{
					Width:   t.Cells(sidebarWidth),
					Height:  t.Flex(1),
					Spacing: 1,
					Children: []t.Widget{
						statsPanel{demo: d},
						demokit.Fill(selectionPanel{demo: d}),
						keysPanel{},
					},
				},
			},
		},
	}
}

// filterPanel holds the text input that narrows the table.
type filterPanel struct {
	demo *TableDemo
}

func (f filterPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := f.demo
	input := t.TextInput{
		ID:          "table-filter-input",
		State:       d.filterInputState,
		Placeholder: "Type to filter any column…",
		Width:       t.Flex(1),
		Style: t.Style{
			ForegroundColor: theme.Text,
		},
		OnChange: func(text string) {
			d.filterState.Query.Set(text)
		},
		OnSubmit: func(text string) {
			t.RequestFocus("demo-table")
		},
		ExtraKeybinds: []t.Keybind{
			{
				Key:  "escape",
				Name: "Clear filter",
				Action: func() {
					d.filterInputState.SetText("")
					d.filterState.Query.Set("")
					t.RequestFocus("demo-table")
				},
			},
		},
	}
	return t.Row{
		Width:   t.Flex(1),
		Spacing: 1,
		Style:   demokit.PanelStyle(theme, "Filter", ctx.IsFocused(input)),
		Children: []t.Widget{
			t.ParseMarkupToText("[$Accent]⌕[/]", theme),
			input,
		},
	}
}

// tablePanel shows the scrolling table, or a hint when there is nothing to show.
type tablePanel struct {
	demo *TableDemo
}

func (p tablePanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := p.demo

	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
		Padding:         t.EdgeInsetsXY(1, 0),
	}
	widths := []t.Dimension{t.Cells(12), t.Cells(10), t.Cells(10), t.Flex(1)}
	columns := make([]t.TableColumn, len(columnNames))
	for i, name := range columnNames {
		columns[i] = t.TableColumn{Width: widths[i], Header: t.Text{Content: name, Style: headerStyle}}
	}

	table := t.Table[TableRow]{
		ID:                  "demo-table",
		State:               d.tableState,
		ScrollState:         d.scrollState,
		Columns:             columns,
		SelectionMode:       d.selectionMode.Get(),
		MultiSelect:         true,
		Filter:              d.filterState,
		MatchCell:           matchCell,
		RenderCellWithMatch: cellRenderer(theme),
		OnSelect:            func(row TableRow) { d.opened.Set(row.Service) },
		Style: t.Style{
			Width: t.Flex(1),
		},
	}

	rows := d.tableState.Rows.Get()
	var empty t.Widget
	switch {
	case len(rows) == 0:
		empty = t.ParseMarkupToText("[$TextMuted]The table is empty. Press [b $Success]a[/] to add a row or [b $Warning]r[/] to reset.[/]", theme)
	case len(visibleRows(rows, d.filterState.QueryText(), d.filterState.Options())) == 0:
		empty = t.ParseMarkupToText(fmt.Sprintf("[$TextMuted]Nothing matches [/][b $Warning]%q[/][$TextMuted]. Press [b]esc[/] in the filter to clear it.[/]", d.filterState.QueryText()), theme)
	}

	return t.Column{
		Width:  t.Flex(1),
		Height: t.Flex(1),
		Style:  demokit.PanelStyle(theme, "Services", ctx.IsFocused(table)),
		Children: []t.Widget{
			t.ShowWhen(empty != nil, empty),
			t.Scrollable{
				ID:    "table-scroll",
				State: d.scrollState,
				Style: t.Style{
					Width:  t.Flex(1),
					Height: t.Flex(1),
				},
				Child: table,
			},
		},
	}
}

// cellRenderer draws each cell, colouring statuses, dimming notes and
// highlighting the parts of a cell that match the filter.
func cellRenderer(theme t.ThemeData) func(TableRow, int, int, bool, bool, t.MatchResult) t.Widget {
	highlight := t.SpanStyle{
		Underline:      t.UnderlineSingle,
		UnderlineColor: theme.Accent,
		Background:     theme.Accent.WithAlpha(0.25),
	}
	cellText := func(content string, style t.Style, match t.MatchResult, wrap t.WrapMode) t.Widget {
		if match.Matched && len(match.Ranges) > 0 {
			return t.Text{Spans: t.HighlightSpans(content, match.Ranges, highlight), Style: style, Wrap: wrap}
		}
		return t.Text{Content: content, Style: style, Wrap: wrap}
	}

	return func(row TableRow, rowIndex int, colIndex int, active bool, selected bool, match t.MatchResult) t.Widget {
		style := t.Style{ForegroundColor: theme.Text, Padding: t.EdgeInsetsXY(1, 0)}
		if selected {
			style.BackgroundColor = theme.Secondary.WithAlpha(0.25)
		}
		if active {
			style.ForegroundColor = theme.Accent.AutoText()
			style.BackgroundColor = theme.Accent
		}

		switch colIndex {
		case 0:
			return cellText(row.Service, style, match, t.WrapNone)
		case 1:
			return cellText(row.Owner, style, match, t.WrapNone)
		case 2:
			if !active {
				style.ForegroundColor = statusColor(theme, row.Status)
			}
			return cellText(row.Status, style, match, t.WrapNone)
		case 3:
			if !active && !selected {
				style.ForegroundColor = theme.TextMuted
			}
			return cellText(row.Notes, style, match, t.WrapSoft)
		default:
			return t.Text{Style: style}
		}
	}
}

func statusColor(theme t.ThemeData, status string) t.Color {
	switch status {
	case "Warn":
		return theme.Warning
	case "Degraded":
		return theme.Error
	default:
		return theme.Success
	}
}

// statsPanel shows live table state. It subscribes to the table's signals
// itself, so moving the cursor only rebuilds this panel.
type statsPanel struct {
	demo *TableDemo
}

func (s statsPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	rows := d.tableState.Rows.Get()
	cursor := d.tableState.CursorIndex.Get()
	column := d.tableState.CursorColumn.Get()
	selected := len(d.tableState.Selection.Get())
	mode := d.selectionMode.Get()
	opened := d.opened.Get()
	view := visibleRows(rows, d.filterState.QueryText(), d.filterState.Options())

	shownColor := "$Text"
	if len(view) < len(rows) {
		shownColor = "$Warning"
	}

	// The cursor is a row, a column or a cell depending on the mode.
	var parts []string
	if mode != t.TableSelectionColumn {
		if i := slices.Index(view, displayedCursor(cursor, view)); i >= 0 {
			parts = append(parts, fmt.Sprintf("%d of %d", i+1, len(view)))
		}
	}
	if mode != t.TableSelectionRow && column >= 0 && column < len(columnNames) {
		parts = append(parts, columnNames[column])
	}
	position := "—"
	if len(parts) > 0 {
		position = strings.Join(parts, " · ")
	}
	if opened == "" {
		opened = "—"
	}

	counts := map[string]int{}
	for _, row := range rows {
		counts[row.Status]++
	}
	health := fmt.Sprintf("[b $Success]%d[/] [$TextMuted]ok[/] [b $Warning]%d[/] [$TextMuted]warn[/] [b $Error]%d[/] [$TextMuted]down[/]",
		counts["OK"], counts["Warn"], counts["Degraded"])

	return t.Column{
		Width: t.Flex(1),
		Style: demokit.PanelStyle(theme, "Overview", false),
		Children: []t.Widget{
			demokit.StatRow(theme, "Rows", fmt.Sprintf("[b $Primary]%d[/]", len(rows))),
			demokit.StatRow(theme, "Showing", fmt.Sprintf("[b %s]%d[/]", shownColor, len(view))),
			demokit.StatRow(theme, "Health", health),
			demokit.StatRow(theme, "Mode", fmt.Sprintf("[b $Accent]%s[/]", selectionModeLabel(mode))),
			demokit.StatRow(theme, "Cursor", fmt.Sprintf("[b $Info]%s[/]", position)),
			demokit.StatRow(theme, "Selected", fmt.Sprintf("[b $Secondary]%d[/]", selected)),
			demokit.StatRow(theme, "Opened", fmt.Sprintf("[b $Success]%s[/]", opened)),
		},
	}
}

// selectionPanel lists what is selected, described in terms of the current
// selection mode.
type selectionPanel struct {
	demo *TableDemo
}

func (s selectionPanel) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	d := s.demo
	selection := d.tableState.Selection.Get()
	rows := d.tableState.Rows.Get()
	mode := d.selectionMode.Get()

	var content t.Widget
	if len(selection) == 0 {
		content = t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]No %ss selected.\nPress [b $Secondary]space[/] or [b $Secondary]shift+move[/].[/]", strings.ToLower(selectionModeLabel(mode))), theme),
			Wrap:  t.WrapSoft,
		}
	} else {
		content = t.Text{
			Content: strings.Join(describeSelection(selection, rows, mode), ", "),
			Wrap:    t.WrapSoft,
			Style:   t.Style{ForegroundColor: theme.Secondary},
		}
	}

	return t.Column{
		Width:    t.Flex(1),
		Height:   t.Flex(1),
		Style:    demokit.PanelStyle(theme, "Selection", false),
		Children: []t.Widget{content},
	}
}

// describeSelection names each selected row, column or cell in order.
func describeSelection(selection map[int]struct{}, rows []TableRow, mode t.TableSelectionMode) []string {
	cols := len(columnNames)
	var names []string
	for _, key := range slices.Sorted(maps.Keys(selection)) {
		switch mode {
		case t.TableSelectionColumn:
			if key < cols {
				names = append(names, columnNames[key])
			}
		case t.TableSelectionRow:
			if key < len(rows) {
				names = append(names, rows[key].Service)
			}
		default:
			if row := key / cols; row < len(rows) {
				names = append(names, rows[row].Service+"."+strings.ToLower(columnNames[key%cols]))
			}
		}
	}
	return names
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
		Style: demokit.PanelStyle(theme, "Keys", false),
		Children: []t.Widget{
			key("↑↓ jk", "$Info", "move row"),
			key("←→ hl", "$Info", "move cell/column"),
			key("⇧+move", "$Secondary", "extend selection"),
			key("space", "$Secondary", "toggle selection"),
			key("m", "$Accent", "cell/row/column"),
			key("enter", "$Success", "open row"),
			key("a p i", "$Success", "add end/top/cursor"),
			key("d c", "$Error", "delete / clear all"),
			key("r", "$Warning", "reset rows"),
			key("/ tab", "$Accent", "filter / focus"),
		},
	}
}

func selectionModeLabel(mode t.TableSelectionMode) string {
	switch mode {
	case t.TableSelectionColumn:
		return "Column"
	case t.TableSelectionRow:
		return "Row"
	default:
		return "Cell"
	}
}

func matchCell(row TableRow, rowIndex int, colIndex int, query string, options t.FilterOptions) t.MatchResult {
	switch colIndex {
	case 0:
		return t.MatchString(row.Service, query, options)
	case 1:
		return t.MatchString(row.Owner, query, options)
	case 2:
		return t.MatchString(row.Status, query, options)
	case 3:
		return t.MatchString(row.Notes, query, options)
	default:
		return t.MatchResult{}
	}
}

// visibleRows returns the source indices of the rows that pass the filter, in
// the order the Table displays them (a row is shown if any cell matches).
func visibleRows(rows []TableRow, query string, options t.FilterOptions) []int {
	return t.ApplyFilter(rows, query, func(row TableRow, q string) t.MatchResult {
		for col := range columnNames {
			if match := matchCell(row, 0, col, q, options); match.Matched {
				return match
			}
		}
		return t.MatchResult{}
	}).Indices
}

// displayedCursor maps the stored cursor row to the row the Table draws it on:
// when the stored row is filtered out, the Table shows the cursor on the first
// visible row until it next moves.
func displayedCursor(cursor int, view []int) int {
	if len(view) == 0 || slices.Contains(view, cursor) {
		return cursor
	}
	return view[0]
}

func makeRow(index int) TableRow {
	owners := []string{"Ingest", "Search", "Storage", "Stream", "Gateway", "Billing"}
	statuses := []string{"OK", "Warn", "Degraded"}
	notes := []string{
		"Batch complete; next run scheduled for 06:00 UTC.",
		"Hot partition observed; routing traffic to standby.",
		"Cache hit rate recovering after cold start.",
		"Backlog climbing; scaling workers to compensate.",
		"Disk cleanup in progress.\nETA 20 minutes.",
	}
	return TableRow{
		Service: fmt.Sprintf("Service-%02d", index),
		Owner:   owners[index%len(owners)],
		Status:  statuses[index%len(statuses)],
		Notes:   notes[index%len(notes)],
	}
}

func defaultRows() []TableRow {
	return []TableRow{
		{Service: "Atlas", Owner: "Ingest", Status: "OK", Notes: "Backfills running; next window starts 02:00 UTC."},
		{Service: "Borealis", Owner: "Search", Status: "Warn", Notes: "Queue depth elevated after deploy; monitoring retries."},
		{Service: "Caldera", Owner: "Storage", Status: "Degraded", Notes: "Compaction stalled on shard 11.\nEscalated to infra."},
		{Service: "Drift", Owner: "Stream", Status: "OK", Notes: "Index rebuilt; latency back to baseline."},
		{Service: "Echo", Owner: "Pipeline", Status: "OK", Notes: "Daily rollup complete.\nRetention set to 90d."},
		{Service: "Flux", Owner: "Gateway", Status: "Warn", Notes: "Spike in 429s; consider rate limit bump."},
		{Service: "Glide", Owner: "Billing", Status: "OK", Notes: "Reconciliation job caught up."},
	}
}
