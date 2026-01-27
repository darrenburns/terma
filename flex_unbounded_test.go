package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlexHeightInTableCell_Panics(t *testing.T) {
	items := []string{"item1"}
	tableState := NewTableState(items)
	scrollState := NewScrollState()

	// Reproduces the real app layout: Scrollable wrapping a Table.
	// The Scrollable measures its child with unbounded height,
	// which makes Flex(1) inside a table cell meaningless.
	widget := Scrollable{
		State:  scrollState,
		Height: Flex(1),
		Child: Table[string]{
			State: tableState,
			Columns: []TableColumn{
				{Width: Cells(25), Header: Text{Content: "Progress"}},
			},
			RenderCell: func(row string, rowIndex int, colIndex int, active bool, selected bool) Widget {
				return ProgressBar{
					Progress: 0.5,
					Style: Style{
						Height: Flex(1),
					},
				}
			},
		},
	}

	assert.Panics(t, func() {
		RenderToBuffer(widget, 80, 24)
	}, "Flex height inside a table cell within a Scrollable should panic")
}
