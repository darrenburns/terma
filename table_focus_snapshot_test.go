package terma

import "testing"

type tableInputRow struct {
	Label string
	Value *TextInputState
}

func buildTableInputsSnapshot(disableFocus bool) Widget {
	rows := []tableInputRow{
		{Label: "Host", Value: NewTextInputState("localhost")},
		{Label: "Port", Value: NewTextInputState("5432")},
	}

	state := NewTableState(rows)

	return Table[tableInputRow]{
		ID:           "input-table",
		DisableFocus: disableFocus,
		State:        state,
		Columns: []TableColumn{
			{Width: Cells(8), Header: Text{Content: "Field"}},
			{Width: Cells(14), Header: Text{Content: "Value"}},
		},
		ColumnSpacing: 1,
		RenderCell: func(row tableInputRow, rowIndex int, colIndex int, active bool, selected bool) Widget {
			switch colIndex {
			case 0:
				return Text{Content: row.Label}
			case 1:
				return TextInput{
					State: row.Value,
					Style: Style{Width: Cells(12)},
				}
			default:
				return Text{Content: ""}
			}
		},
	}
}

func TestSnapshot_TableInputs_TableFocused(t *testing.T) {
	widget := buildTableInputsSnapshot(false)
	AssertSnapshot(t, widget, 32, 4,
		"Table focused by default: header row and two data rows visible; inputs are unfocused (no cursor).")
}

func TestSnapshot_TableInputs_TableFocusDisabled(t *testing.T) {
	widget := buildTableInputsSnapshot(true)
	AssertSnapshot(t, widget, 32, 4,
		"Table focus disabled: first text input receives focus and shows a cursor.")
}
