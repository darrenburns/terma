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
	tableState  *t.TableState[TableRow]
	scrollState *t.ScrollState
	counter     int
	themeIndex  t.Signal[int]
}

func NewTableDemo() *TableDemo {
	rows := defaultRows()

	return &TableDemo{
		tableState:  t.NewTableState(rows),
		scrollState: t.NewScrollState(),
		counter:     len(rows),
		themeIndex:  t.NewSignal(0),
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
			idx := d.tableState.CursorIndex.Peek()
			d.tableState.ToggleSelection(idx)
		}},
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

func (d *TableDemo) buildSelectionSummary(theme t.ThemeData) t.Widget {
	selection := d.tableState.Selection.Get()
	if len(selection) == 0 {
		return t.Text{
			Spans: t.ParseMarkup("[$TextMuted]No rows selected[/]", theme),
		}
	}

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

func (d *TableDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	themeIdx := d.themeIndex.Get()
	currentTheme := themeNames[themeIdx]

	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
		Padding:         t.EdgeInsetsXY(1, 0),
	}

	cellPadding := t.EdgeInsetsXY(1, 0)
	columns := []t.TableColumn{
		{Width: t.Cells(12), Header: t.Text{Content: "Service", Style: headerStyle}},
		{Width: t.Cells(10), Header: t.Text{Content: "Owner", Style: headerStyle}},
		{Width: t.Cells(10), Header: t.Text{Content: "Status", Style: headerStyle}},
		{Width: t.Flex(1), Header: t.Text{Content: "Notes", Style: headerStyle}},
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
				Spans: t.ParseMarkup("Navigate: [b $Info]↑/↓[/] or [b $Info]j/k[/] | Select range: [b $Secondary]Shift+↑/↓[/] | Toggle: [b $Secondary]Space[/]", theme),
			},
			t.Text{
				Spans: t.ParseMarkup("Modify: [b $Success]a[/]ppend [b $Success]p[/]repend [b $Error]d[/]elete [b $Warning]r[/]eset", theme),
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
					ID:          "demo-table",
					State:       d.tableState,
					ScrollState: d.scrollState,
					Columns:     columns,
					MultiSelect: true,
					RenderCell: func(row TableRow, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
						style := t.Style{ForegroundColor: theme.Text}
						if selected {
							style.BackgroundColor = theme.SurfaceHover
						}
						if active {
							style.ForegroundColor = theme.Accent
						}
						style.Padding = cellPadding

						switch colIndex {
						case 0:
							return t.Text{Content: row.Service, Style: style}
						case 1:
							return t.Text{Content: row.Owner, Style: style}
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
							return t.Text{Content: row.Status, Style: statusStyle}
						case 3:
							notesStyle := style
							if !active && !selected {
								notesStyle.ForegroundColor = theme.TextMuted
							}
							return t.Text{Content: row.Notes, Style: notesStyle, Wrap: t.WrapSoft}
						default:
							return t.Text{Content: "", Style: style}
						}
					},
				},
			},
			t.Text{
				Spans: t.ParseMarkup(fmt.Sprintf("Rows: [b $Warning]%d[/] | Cursor: [b $Info]%d[/] | Press [b $Error]Ctrl+C[/] to quit", d.tableState.RowCount(), d.tableState.CursorIndex.Get()+1), theme),
			},
			d.buildSelectionSummary(theme),
		},
	}
}

func main() {
	t.SetTheme(themeNames[0])
	app := NewTableDemo()
	_ = t.InitLogger()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
