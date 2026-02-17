package main

import (
	"fmt"
	"log"

	t "github.com/darrenburns/terma"
)

type App struct {
	tableState *t.TableState[[]string]
}

func NewApp() *App {
	return &App{
		tableState: t.NewTableState([][]string{
			{"Alice", "Engineer", "Active"},
			{"Bob", "Designer", "Away"},
			{"Carol", "Manager", "Active"},
			{"Dave", "Engineer", "Busy"},
		}),
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Column{
		Height:  t.Flex(1),
		Spacing: 1,
		Style:   t.Style{Padding: t.EdgeInsetsAll(1)},
		Children: []t.Widget{
			t.Text{
				Content: "Team Directory",
				Style:   t.Style{Bold: true, ForegroundColor: theme.Primary},
			},
			t.Table[[]string]{
				ID:    "team-table",
				State: a.tableState,
				Columns: []t.TableColumn{
					{Width: t.Cells(12), Header: t.Text{Content: "Name", Style: t.Style{Bold: true}}},
					{Width: t.Cells(12), Header: t.Text{Content: "Role", Style: t.Style{Bold: true}}},
					{Width: t.Cells(10), Header: t.Text{Content: "Status", Style: t.Style{Bold: true}}},
				},
				OnSelect: func(row []string) {
					fmt.Printf("Selected: %s\n", row[0])
				},
			},
		},
	}
}

func main() {
	if err := t.Run(NewApp()); err != nil {
		log.Fatal(err)
	}
}
