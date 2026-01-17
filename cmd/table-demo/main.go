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

type TableRow struct {
	Service string
	Owner   string
	Status  string
	Notes   string
}

// TableDemo showcases the Table widget with variable-height cells and multi-select.
type TableDemo struct {
	tableState       *t.TableState[TableRow]
	scrollState      *t.ScrollState
	filterState      *t.FilterState
	filterInputState *t.TextInputState
	counter          int
	themeIndex       t.Signal[int]
	selectionMode    t.Signal[int]
}

func NewTableDemo() *TableDemo {
	rows := defaultRows()

	return &TableDemo{
		tableState:       t.NewTableState(rows),
		scrollState:      t.NewScrollState(),
		filterState:      t.NewFilterState(),
		filterInputState: t.NewTextInputState(""),
		counter:          len(rows),
		themeIndex:       t.NewSignal(0),
		selectionMode:    t.NewSignal(int(t.TableSelectionCursor)),
	}
}

func (d *TableDemo) cycleTheme() {
	d.themeIndex.Update(func(i int) int {
		next := (i + 1) % len(themeNames)
		t.SetTheme(themeNames[next])
		return next
	})
}

func (d *TableDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "space", Name: "Toggle selection", Action: func() {
			idx := d.selectionIndex(d.columnCount())
			d.tableState.ToggleSelection(idx)
		}},
		{Key: "m", Name: "Selection mode", Action: d.cycleSelectionMode},
		{Key: "a", Name: "Append row", Action: func() {
			d.counter++
			d.tableState.Append(d.makeRow(d.counter))
		}},
		{Key: "p", Name: "Prepend row", Action: func() {
			d.counter++
			d.tableState.Prepend(d.makeRow(d.counter))
		}},
		{Key: "d", Name: "Delete row", Action: func() {
			idx := d.tableState.CursorIndex.Peek()
			d.tableState.RemoveAt(idx)
		}},
		{Key: "r", Name: "Reset", Action: func() {
			rows := defaultRows()
			d.tableState.SetRows(rows)
			d.tableState.ClearSelection()
			d.tableState.ClearAnchor()
			d.tableState.SelectFirst()
			d.scrollState.SetOffset(0)
			d.counter = len(rows)
		}},
		{Key: "t", Name: "Next theme", Action: d.cycleTheme},
	}
}

func (d *TableDemo) makeRow(index int) TableRow {
	owners := []string{"Ingest", "Search", "Storage", "Stream", "Gateway", "Billing"}
	statuses := []string{"OK", "Warn", "Degraded"}
	notes := []string{
		"Batch complete; next run scheduled for 06:00 UTC.",
		"Hot partition observed; routing traffic to standby.",
		"Cache hit rate recovering after cold start.",
		"Backlog climbing; scaling workers to compensate.",
		"Disk cleanup in progress.\nETA 20 minutes.",
	}

	owner := owners[index%len(owners)]
	status := statuses[index%len(statuses)]
	note := notes[index%len(notes)]

	return TableRow{
		Service: fmt.Sprintf("Service-%02d", index),
		Owner:   owner,
		Status:  status,
		Notes:   note,
	}
}

func (d *TableDemo) cycleSelectionMode() {
	d.selectionMode.Update(func(i int) int {
		next := (i + 1) % 3
		return next
	})
}

func (d *TableDemo) selectionIndex(columnCount int) int {
	mode := t.TableSelectionMode(d.selectionMode.Get())
	rowIdx := d.tableState.CursorIndex.Peek()
	colIdx := d.tableState.CursorColumn.Peek()

	switch mode {
	case t.TableSelectionColumn:
		return colIdx
	case t.TableSelectionCursor:
		if columnCount <= 0 {
			return 0
		}
		return rowIdx*columnCount + colIdx
	default:
		return rowIdx
	}
}

func (d *TableDemo) columnCount() int {
	return 4
}

func (d *TableDemo) matchCell(row TableRow, rowIndex int, colIndex int, query string, options t.FilterOptions) t.MatchResult {
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

func defaultRows() []TableRow {
	return []TableRow{
		{
			Service: "Atlas",
			Owner:   "Ingest",
			Status:  "OK",
			Notes:   "Backfills running; next window starts 02:00 UTC.",
		},
		{
			Service: "Borealis",
			Owner:   "Search",
			Status:  "Warn",
			Notes:   "Queue depth elevated after deploy; monitoring retries.",
		},
		{
			Service: "Caldera",
			Owner:   "Storage",
			Status:  "Degraded",
			Notes:   "Compaction stalled on shard 11.\nEscalated to infra.",
		},
		{
			Service: "Drift",
			Owner:   "Stream",
			Status:  "OK",
			Notes:   "Index rebuilt; latency back to baseline.",
		},
		{
			Service: "Echo",
			Owner:   "Pipeline",
			Status:  "OK",
			Notes:   "Daily rollup complete.\nRetention set to 90d.",
		},
		{
			Service: "Flux",
			Owner:   "Gateway",
			Status:  "Warn",
			Notes:   "Spike in 429s; consider rate limit bump.",
		},
		{
			Service: "Glide",
			Owner:   "Billing",
			Status:  "OK",
			Notes:   "Reconciliation job caught up.",
		},
	}
}

func (d *TableDemo) buildSelectionSummary(theme t.ThemeData, mode t.TableSelectionMode, columnCount int) t.Widget {
	selection := d.tableState.Selection.Get()
	if len(selection) == 0 {
		emptyLabel := "No selection"
		switch mode {
		case t.TableSelectionColumn:
			emptyLabel = "No columns selected"
		case t.TableSelectionRow:
			emptyLabel = "No rows selected"
		default:
			emptyLabel = "No cells selected"
		}
		return t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]%s[/]", emptyLabel), theme),
		}
	}

	switch mode {
	case t.TableSelectionColumn:
		cols := sortedKeys(selection)
		label := "Columns"
		if len(cols) == 1 {
			label = "Column"
		}
		return t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[b $Secondary]%s (%d): [/]%s", label, len(cols), joinInts(cols)), theme),
		}
	case t.TableSelectionCursor:
		return t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[b $Secondary]Cells selected: [/]%d", len(selection)), theme),
		}
	default:
		rows := d.tableState.Rows.Get()
		var selected []string
		for i, row := range rows {
			if _, ok := selection[i]; ok {
				selected = append(selected, row.Service)
			}
		}

		summary := strings.Join(selected, ", ")
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}

		return t.Text{
			Spans: t.ParseMarkup(fmt.Sprintf("[b $Secondary]Selected (%d): [/]%s", len(selected), summary), theme),
		}
	}
}

func (d *TableDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := d.themeIndex.Get()
	currentTheme := themeNames[themeIdx]
	mode := t.TableSelectionMode(d.selectionMode.Get())

	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
		Padding:         t.EdgeInsetsXY(1, 0),
	}

	columns := []t.TableColumn{
		{Width: t.Cells(12), Header: t.Text{Content: "Service", Style: headerStyle}},
		{Width: t.Cells(10), Header: t.Text{Content: "Owner", Style: headerStyle}},
		{Width: t.Cells(10), Header: t.Text{Content: "Status", Style: headerStyle}},
		{Width: t.Flex(1), Header: t.Text{Content: "Notes", Style: headerStyle}},
	}

	cellPadding := t.EdgeInsetsXY(1, 0)
	highlight := t.SpanStyle{
		Underline:      t.UnderlineSingle,
		UnderlineColor: theme.Accent,
		Background:     theme.Accent.WithAlpha(0.25),
	}
	renderCellText := func(content string, style t.Style, match t.MatchResult, wrap t.WrapMode) t.Widget {
		if match.Matched && len(match.Ranges) > 0 {
			return t.Text{
				Spans: t.HighlightSpans(content, match.Ranges, highlight),
				Style: style,
				Wrap:  wrap,
			}
		}
		return t.Text{Content: content, Style: style, Wrap: wrap}
	}

	return t.Column{
		ID:      "table-demo-root",
		Height:  t.Flex(1),
		Spacing: 1,
		Style: t.Style{
			BackgroundColor: theme.Background,
			Padding:         t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Table Widget Demo",
				Style: t.Style{
					ForegroundColor: theme.TextOnPrimary,
					BackgroundColor: theme.Primary,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("[$TextMuted]Theme: [/][$Accent]%s[/][$TextMuted] (press t to change)[/]", currentTheme), theme),
			},
			t.Text{
				Spans: t.ParseMarkup("Navigate: [b $Info]↑/↓[/] or [b $Info]j/k[/] | Select range: [b $Secondary]Shift+Arrow[/] | Toggle: [b $Secondary]Space[/] | Mode: [b $Secondary]m[/]", theme),
			},
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("Modify: [b $Success]a[/]ppend [b $Success]p[/]repend [b $Error]d[/]elete [b $Warning]r[/]eset  •  Selection: [b $Accent]%s[/]", selectionModeLabel(mode)), theme),
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
						ID:          "table-filter-input",
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
					},
					t.Text{
						Spans: t.ParseMarkup("[$TextMuted]Tab to move focus[/]", theme),
					},
				},
			},
			t.Scrollable{
				ID:           "table-scroll",
				State:        d.scrollState,
				Height:       t.Cells(12),
				DisableFocus: true,
				Style: t.Style{
					Border:  t.RoundedBorder(theme.Primary, t.BorderTitle("Services")),
					Padding: t.EdgeInsetsXY(1, 0),
				},
				Child: t.Table[TableRow]{
					ID:            "demo-table",
					State:         d.tableState,
					ScrollState:   d.scrollState,
					Columns:       columns,
					SelectionMode: mode,
					MultiSelect:   true,
					Filter:        d.filterState,
					MatchCell:     d.matchCell,
					RenderCellWithMatch: func(row TableRow, rowIndex int, colIndex int, active bool, selected bool, match t.MatchResult) t.Widget {
						style := t.Style{ForegroundColor: theme.Text}
						if selected {
							style.BackgroundColor = theme.Accent.WithAlpha(0.2)
						}
						if active {
							style.ForegroundColor = theme.Accent.AutoText().WithAlpha(0.8)
							style.BackgroundColor = theme.Accent
						}
						style.Padding = cellPadding

						switch colIndex {
						case 0:
							return renderCellText(row.Service, style, match, t.WrapNone)
						case 1:
							return renderCellText(row.Owner, style, match, t.WrapNone)
						case 2:
							statusStyle := style
							if !active {
								switch row.Status {
								case "Warn":
									statusStyle.ForegroundColor = theme.Warning
								case "Degraded":
									statusStyle.ForegroundColor = theme.Error
								default:
									statusStyle.ForegroundColor = theme.Success
								}
							}
							return renderCellText(row.Status, statusStyle, match, t.WrapNone)
						case 3:
							notesStyle := style
							if !active && !selected {
								notesStyle.ForegroundColor = theme.TextMuted
							}
							return renderCellText(row.Notes, notesStyle, match, t.WrapSoft)
						default:
							return t.Text{Content: "", Style: style}
						}
					},
				},
			},
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("Rows: [b $Warning]%d[/] | Cursor: [b $Info]%d[/] | Press [b $Error]Ctrl+C[/] to quit", d.tableState.RowCount(), d.tableState.CursorIndex.Get()+1), theme),
			},
			d.buildSelectionSummary(theme, mode, d.columnCount()),
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

func sortedKeys(m map[int]struct{}) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%d", v+1))
	}
	return strings.Join(parts, ", ")
}

func main() {
	t.SetTheme(themeNames[0])
	app := NewTableDemo()
	_ = t.InitLogger()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
