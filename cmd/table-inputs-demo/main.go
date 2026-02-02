package main

import (
	"log"

	t "terma"
)

type InputRow struct {
	Field string
	Value *t.TextInputState
	Notes *t.TextInputState
}

// TableInputsDemo shows a non-focusable Table containing focusable TextInput widgets.
type TableInputsDemo struct {
	tableState *t.TableState[InputRow]
}

func NewTableInputsDemo() *TableInputsDemo {
	rows := []InputRow{
		{
			Field: "Host",
			Value: t.NewTextInputState("localhost"),
			Notes: t.NewTextInputState("dev"),
		},
		{
			Field: "Port",
			Value: t.NewTextInputState("5432"),
			Notes: t.NewTextInputState("postgres"),
		},
		{
			Field: "User",
			Value: t.NewTextInputState("admin"),
			Notes: t.NewTextInputState("readonly"),
		},
	}

	return &TableInputsDemo{
		tableState: t.NewTableState(rows),
	}
}

func (d *TableInputsDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	headerStyle := t.Style{
		ForegroundColor: theme.Text,
		BackgroundColor: theme.Surface,
		Bold:            true,
	}
	inputStyle := t.Style{
		Width:           t.Cells(14),
		Padding:         t.EdgeInsetsXY(1, 0),
		BackgroundColor: theme.Surface,
	}

	columns := []t.TableColumn{
		{Width: t.Cells(8), Header: t.Text{Content: "Field", Style: headerStyle}},
		{Width: t.Cells(16), Header: t.Text{Content: "Value", Style: headerStyle}},
		{Width: t.Cells(16), Header: t.Text{Content: "Notes", Style: headerStyle}},
	}

	return t.Column{
		Spacing: 1,
		Style: t.Style{
			Padding: t.EdgeInsetsXY(2, 1),
		},
		Children: []t.Widget{
			t.Text{
				Content: "Table Inputs Demo",
				Style: t.Style{
					ForegroundColor: t.Black,
					BackgroundColor: t.Cyan,
					Padding:         t.EdgeInsetsXY(1, 0),
				},
			},
			t.ParseMarkupToText("Tab to move between inputs. Table itself does not take focus.", theme),
			t.Table[InputRow]{
				ID:            "input-table",
				DisableFocus:  true,
				State:         d.tableState,
				Columns:       columns,
				ColumnSpacing: 1,
				RenderCell: func(row InputRow, rowIndex int, colIndex int, active bool, selected bool) t.Widget {
					switch colIndex {
					case 0:
						return t.Text{Content: row.Field}
					case 1:
						return t.TextInput{
							State: row.Value,
							Style: inputStyle,
						}
					case 2:
						return t.TextInput{
							State: row.Notes,
							Style: inputStyle,
						}
					default:
						return t.Text{Content: ""}
					}
				},
			},
			t.Text{
				Spans: t.ParseMarkup("Press [b #ff5555]Ctrl+C[/] to quit", t.ThemeData{}),
			},
		},
	}
}

func main() {
	app := NewTableInputsDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
