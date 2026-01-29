package main

import (
	"log"

	t "terma"
)

// TableColumnWidthDemo demonstrates different table column width behaviors.
type TableColumnWidthDemo struct {
	autoTable    *t.TableState[[]string]
	cellsTable   *t.TableState[[]string]
	flexTable    *t.TableState[[]string]
	percentTable *t.TableState[[]string]
	wrapTable    *t.TableState[[]string]
	scrollState  *t.ScrollState
}

func NewTableColumnWidthDemo() *TableColumnWidthDemo {
	return &TableColumnWidthDemo{
		// Auto: columns size to LARGEST child
		autoTable: t.NewTableState([][]string{
			{"A", "Short", "X"},
			{"Longer Name Here", "Medium text", "Extended content in column"},
			{"B", "Tiny", "Y"},
		}),
		// Fixed cells
		cellsTable: t.NewTableState([][]string{
			{"Fixed 10", "Fixed 15", "Fixed 20"},
			{"Truncated if too long", "Also truncated", "Content may be cut off"},
		}),
		// Flex proportional
		flexTable: t.NewTableState([][]string{
			{"Flex(1)", "Flex(2)", "Flex(1)"},
			{"25%", "50%", "25%"},
		}),
		// Percent
		percentTable: t.NewTableState([][]string{
			{"Percent(20)", "Percent(50)", "Percent(30)"},
			{"20% width", "50% width", "30% width"},
		}),
		// Wrapping text in Auto columns
		wrapTable: t.NewTableState([][]string{
			{"Short", "This is a longer piece of text that might wrap"},
			{"Also short", "Another cell with extended content to demonstrate wrapping behavior"},
		}),
		scrollState: t.NewScrollState(),
	}
}

func (d *TableColumnWidthDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Scrollable{
		State:     d.scrollState,
		Focusable: true,
		Child: t.Column{
			Spacing: 2,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				d.buildTitle("Table Column Width Demo", theme),

				// Auto columns demo
				d.buildSection("1. Auto Columns (default)", "Sizes to LARGEST child in each column", theme),
				d.buildAutoTable(theme),

				// Fixed cells demo
				d.buildSection("2. Cells(n) - Fixed Width", "Each column has exact fixed width (10, 15, 20 cells)", theme),
				d.buildCellsTable(theme),

				// Flex demo
				d.buildSection("3. Flex(n) - Proportional", "Flex(1):Flex(2):Flex(1) = 25%:50%:25% of available space", theme),
				d.buildFlexTable(theme),

				// Percent demo
				d.buildSection("4. Percent(n)", "Percent(20):Percent(50):Percent(30) of available width", theme),
				d.buildPercentTable(theme),

				// Wrapping demo
				d.buildSection("5. Auto with Wrapping Text", "Auto columns prefer single-line width; text wraps only when constrained", theme),
				d.buildWrapTable(theme),

				t.Text{
					Content: "Press Ctrl+C to quit",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
			},
		},
	}
}

func (d *TableColumnWidthDemo) buildTitle(text string, theme t.ThemeData) t.Widget {
	return t.Text{
		Content: text,
		Style: t.Style{
			ForegroundColor: theme.TextOnPrimary,
			BackgroundColor: theme.Primary,
			Padding:         t.EdgeInsetsXY(1, 0),
			Bold:            true,
		},
	}
}

func (d *TableColumnWidthDemo) buildSection(title, description string, theme t.ThemeData) t.Widget {
	return t.Column{
		Children: []t.Widget{
			t.Text{
				Content: title,
				Style: t.Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},
			t.Text{
				Content: description,
				Style:   t.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (d *TableColumnWidthDemo) headerStyle(theme t.ThemeData) t.Style {
	return t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
	}
}

func (d *TableColumnWidthDemo) buildAutoTable(theme t.ThemeData) t.Widget {
	hs := d.headerStyle(theme)
	return t.Table[[]string]{
		ID:    "auto-table",
		State: d.autoTable,
		Columns: []t.TableColumn{
			{Width: t.Auto, Header: t.Text{Content: "Col A (Auto)", Style: hs}},
			{Width: t.Auto, Header: t.Text{Content: "Col B (Auto)", Style: hs}},
			{Width: t.Auto, Header: t.Text{Content: "Col C (Auto)", Style: hs}},
		},
		ColumnSpacing: 1,
		Style: t.Style{
			Border: t.Border{Style: t.BorderRounded, Color: theme.Border},
		},
	}
}

func (d *TableColumnWidthDemo) buildCellsTable(theme t.ThemeData) t.Widget {
	hs := d.headerStyle(theme)
	return t.Table[[]string]{
		ID:    "cells-table",
		State: d.cellsTable,
		Columns: []t.TableColumn{
			{Width: t.Cells(10), Header: t.Text{Content: "10 cells", Style: hs}},
			{Width: t.Cells(15), Header: t.Text{Content: "15 cells", Style: hs}},
			{Width: t.Cells(20), Header: t.Text{Content: "20 cells", Style: hs}},
		},
		ColumnSpacing: 1,
		Style: t.Style{
			Border: t.Border{Style: t.BorderRounded, Color: theme.Border},
		},
	}
}

func (d *TableColumnWidthDemo) buildFlexTable(theme t.ThemeData) t.Widget {
	hs := d.headerStyle(theme)
	return t.Table[[]string]{
		ID:    "flex-table",
		State: d.flexTable,
		Columns: []t.TableColumn{
			{Width: t.Flex(1), Header: t.Text{Content: "Flex(1)", Style: hs}},
			{Width: t.Flex(2), Header: t.Text{Content: "Flex(2)", Style: hs}},
			{Width: t.Flex(1), Header: t.Text{Content: "Flex(1)", Style: hs}},
		},
		ColumnSpacing: 1,
		Style: t.Style{
			Border: t.Border{Style: t.BorderRounded, Color: theme.Border},
			Width:  t.Percent(80),
		},
	}
}

func (d *TableColumnWidthDemo) buildPercentTable(theme t.ThemeData) t.Widget {
	hs := d.headerStyle(theme)
	return t.Table[[]string]{
		ID:    "percent-table",
		State: d.percentTable,
		Columns: []t.TableColumn{
			{Width: t.Percent(20), Header: t.Text{Content: "20%", Style: hs}},
			{Width: t.Percent(50), Header: t.Text{Content: "50%", Style: hs}},
			{Width: t.Percent(30), Header: t.Text{Content: "30%", Style: hs}},
		},
		ColumnSpacing: 1,
		Style: t.Style{
			Border: t.Border{Style: t.BorderRounded, Color: theme.Border},
			Width:  t.Percent(80),
		},
	}
}

func (d *TableColumnWidthDemo) buildWrapTable(theme t.ThemeData) t.Widget {
	hs := d.headerStyle(theme)
	return t.Table[[]string]{
		ID:    "wrap-table",
		State: d.wrapTable,
		Columns: []t.TableColumn{
			{Width: t.Auto, Header: t.Text{Content: "Short Col (Auto)", Style: hs}},
			{Width: t.Flex(1), Header: t.Text{Content: "Long Content Col (Flex1)", Style: hs}},
		},
		ColumnSpacing: 1,
		Style: t.Style{
			Border: t.Border{Style: t.BorderRounded, Color: theme.Border},
		},
		RenderCell: func(row []string, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
			return t.Text{Content: row[colIndex], Wrap: t.WrapSoft}
		},
	}
}

func main() {
	app := NewTableColumnWidthDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
