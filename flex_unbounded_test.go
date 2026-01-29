package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlexHeightInTableCell_GracefullyHandled(t *testing.T) {
	items := []string{"item1"}
	tableState := NewTableState(items)
	scrollState := NewScrollState()

	// Reproduces the real app layout: Scrollable wrapping a Table.
	// The Scrollable measures its child with unbounded height.
	// Flex(1) inside a table cell has no natural size in this context,
	// so it gracefully returns 0 instead of panicking.
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

	// Should not panic - Flex content in unbounded context gracefully returns zero size
	assert.NotPanics(t, func() {
		RenderToBuffer(widget, 80, 24)
	}, "Flex height inside a table cell within a Scrollable should not panic")
}

func TestScrollableWithFixedAndFlexChildren(t *testing.T) {
	scrollState := NewScrollState()

	red := RGB(180, 70, 70)
	blue := RGB(70, 100, 180)
	gray := RGB(80, 80, 80)

	// A Scrollable containing a Column with both fixed and flex children.
	// The Scrollable measures with unbounded height, so the Flex child
	// should gracefully collapse to zero while the fixed child retains its size.
	widget := Scrollable{
		State: scrollState,
		Child: Column{
			Style: Style{BackgroundColor: gray, Width: Flex(1)},
			Children: []Widget{
				Text{Content: "Fixed header", Style: Style{BackgroundColor: red, Width: Flex(1)}},
				Spacer{Height: Flex(1)}, // Flex - should collapse to 0 (not visible)
				Text{Content: "Fixed footer", Style: Style{BackgroundColor: blue, Width: Flex(1)}},
			},
		},
	}

	AssertSnapshot(t, widget, 40, 10,
		"Red 'Fixed header' and blue 'Fixed footer' on adjacent lines (no gap). "+
			"Flex Spacer between them collapses to 0 height in unbounded context. "+
			"Gray column background extends full width.")
}

func TestScrollableWithNestedFlexInRow(t *testing.T) {
	scrollState := NewScrollState()

	red := RGB(180, 70, 70)
	green := RGB(70, 140, 70)
	blue := RGB(70, 100, 180)
	gray := RGB(80, 80, 80)

	// A Scrollable containing a Column with a Row that has flex children.
	// The Row's width is bounded (by Scrollable), but Column's height measurement
	// is unbounded. The Row itself should lay out correctly (bounded width),
	// while any Flex height in the Column should collapse.
	widget := Scrollable{
		State: scrollState,
		Child: Column{
			Style: Style{BackgroundColor: gray, Width: Flex(1)},
			Children: []Widget{
				Row{
					Style: Style{Width: Flex(1), BackgroundColor: green},
					Children: []Widget{
						Text{Content: "Left", Style: Style{BackgroundColor: red}},
						Spacer{Width: Flex(1)}, // Flex width - bounded, Row bg shows expansion
						Text{Content: "Right", Style: Style{BackgroundColor: blue}},
					},
				},
				Spacer{Height: Flex(1)}, // Flex height - unbounded, should collapse (not visible)
			},
		},
	}

	AssertSnapshot(t, widget, 40, 10,
		"Single row: red 'Left' on left edge, blue 'Right' on right edge, green fills between. "+
			"Flex width works (Row has bounded width). "+
			"Flex height Spacer below Row collapses to 0 (Column has unbounded height).")
}
