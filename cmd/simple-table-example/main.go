package main

import (
	"fmt"
	"log"

	t "terma"
)

// SimpleTableDemo demonstrates the most basic usage of Table with default rendering.
type SimpleTableDemo struct {
	tableState  *t.TableState[[]string]
	selectedMsg t.Signal[string]
}

func NewSimpleTableDemo() *SimpleTableDemo {
	return &SimpleTableDemo{
		tableState: t.NewTableState([][]string{
			{"Alpha", "Up", "OK"},
			{"Bravo", "Stable", "Warn"},
			{"Charlie", "Down", "Degraded"},
			{"Delta", "Up", "OK"},
		}),
		selectedMsg: t.NewSignal("No selection yet"),
	}
}

func (d *SimpleTableDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
	}

	columns := []t.TableColumn{
		{Width: t.Auto, Header: t.Text{Content: "Name", Style: headerStyle}},
		{Width: t.Auto, Header: t.Text{Content: "State", Style: headerStyle}},
		{Width: t.Auto, Header: t.Text{Content: "Health", Style: headerStyle}},
	}

	return t.Column{
		ID:      "simple-table-root",
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Simple Table Example",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Cyan,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.ParseMarkupToText("Use [b #00ffff]↑/↓[/] or [b #00ffff]j/k[/] to navigate • [b #00ffff]Enter[/] to select", theme),
			t.Table[[]string]{
				ID:          "simple-table",
				State:       d.tableState,
				MultiSelect: true,
				Columns:     columns,
				OnSelect: func(row []string) {
					if len(row) > 0 {
						d.selectedMsg.Set(fmt.Sprintf("Selected: %s", row[0]))
					} else {
						d.selectedMsg.Set("Selected: (empty row)")
					}
				},
			},
			t.Text{
				Content: d.selectedMsg.Get(),
				Style: t.Style{
					ForegroundColor: t.BrightYellow,
				},
			},
			t.Text{
				Spans: t.ParseMarkup("Press [b #ff5555]Ctrl+C[/] to quit", t.ThemeData{}),
			},
		},
	}
}

func main() {
	app := NewSimpleTableDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
