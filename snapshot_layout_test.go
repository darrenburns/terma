package terma

import (
	"testing"
)

// Layout test colors - distinct colors to visualize space allocation
var (
	layoutRed    = RGB(180, 70, 70)
	layoutGreen  = RGB(70, 140, 70)
	layoutBlue   = RGB(70, 100, 180)
	layoutPurple = RGB(140, 70, 140)
	layoutOrange = RGB(180, 120, 50)
	layoutTeal   = RGB(70, 140, 140)
	layoutGray   = RGB(100, 100, 100)
)

// =============================================================================
// Column Widget Tests
// =============================================================================

func TestSnapshot_Column_BasicVerticalLayout(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Text{Content: "First", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Second", Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "Third", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Three text items stacked vertically at top-left. Red 'First' on row 1, green 'Second' on row 2, blue 'Third' on row 3. Each sized to content width.")
}

func TestSnapshot_Column_MainAlignStart(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisStart,
		Children: []Widget{
			Text{Content: "Top", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"Blue 'Top' text at row 1 (top edge). Column is 10 rows tall with empty space below the text.")
}

func TestSnapshot_Column_MainAlignCenter(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisCenter,
		Children: []Widget{
			Text{Content: "Centered", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"Blue 'Centered' text vertically centered around row 5. Equal empty space above and below.")
}

func TestSnapshot_Column_MainAlignEnd(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisEnd,
		Children: []Widget{
			Text{Content: "Bottom", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"Blue 'Bottom' text at row 10 (bottom edge). Column is 10 rows tall with empty space above the text.")
}

func TestSnapshot_Column_CrossAlignStretch(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisStretch,
		Children: []Widget{
			Text{Content: "Stretched", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue text background stretches full 20-cell width. Text 'Stretched' left-aligned within the stretched area.")
}

func TestSnapshot_Column_CrossAlignStart(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisStart,
		Children: []Widget{
			Text{Content: "Left", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'Left' text at left edge (column 1). Text width matches content, not stretched. Column is 20 cells wide.")
}

func TestSnapshot_Column_CrossAlignCenter(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisCenter,
		Children: []Widget{
			Text{Content: "CenterH", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'CenterH' text horizontally centered in 20-cell column. Equal empty space on left and right.")
}

func TestSnapshot_Column_CrossAlignEnd(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisEnd,
		Children: []Widget{
			Text{Content: "Right", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'Right' text at right edge of 20-cell column. Empty space on left, text aligned to column 20.")
}

func TestSnapshot_Column_WithSpacing(t *testing.T) {
	widget := Column{
		Spacing: 2,
		Children: []Widget{
			Text{Content: "A", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "B", Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "C", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"Three items stacked vertically with 2-row gaps. Red 'A' at row 1, green 'B' at row 4, blue 'C' at row 7.")
}

func TestSnapshot_Column_NestedColumns(t *testing.T) {
	widget := Column{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Column{
				Style: Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Nested1"},
				},
			},
			Column{
				Style: Style{BackgroundColor: layoutBlue},
				Children: []Widget{
					Text{Content: "Nested2"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Gray outer column with two nested columns stacked. Red column with 'Nested1' on row 1, blue column with 'Nested2' on row 2.")
}

func TestSnapshot_Column_MixedDimensions(t *testing.T) {
	widget := Column{
		Height: Cells(10),
		Children: []Widget{
			Text{Content: "Fixed", Height: Cells(2), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Flex", Height: Flex(1), Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "Auto", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"10-row column with mixed heights. Red 'Fixed' takes 2 rows, green 'Flex' expands to fill 7 rows, blue 'Auto' takes 1 row at bottom.")
}

// =============================================================================
// Row Widget Tests
// =============================================================================

func TestSnapshot_Row_BasicHorizontalLayout(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{Content: "A", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "B", Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "C", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Three single-char items in a horizontal row. Red 'A' at column 1, green 'B' at column 2, blue 'C' at column 3. All on row 1.")
}

func TestSnapshot_Row_MainAlignStart(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisStart,
		Children: []Widget{
			Text{Content: "Left", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Blue 'Left' text at left edge (column 1). Row is 20 cells wide with empty space to the right.")
}

func TestSnapshot_Row_MainAlignCenter(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisCenter,
		Children: []Widget{
			Text{Content: "Mid", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Blue 'Mid' text horizontally centered in 20-cell row. Equal empty space on left and right.")
}

func TestSnapshot_Row_MainAlignEnd(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisEnd,
		Children: []Widget{
			Text{Content: "Right", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Blue 'Right' text at right edge of 20-cell row. Empty space on left, text ends at column 20.")
}

func TestSnapshot_Row_CrossAlignStretch(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisStretch,
		Children: []Widget{
			Text{Content: "Tall", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue background stretches full 5-row height. Text 'Tall' at top within the stretched area.")
}

func TestSnapshot_Row_CrossAlignStart(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisStart,
		Children: []Widget{
			Text{Content: "Top", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'Top' text at row 1 (top edge). Row is 5 rows tall, text height matches content.")
}

func TestSnapshot_Row_CrossAlignCenter(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisCenter,
		Children: []Widget{
			Text{Content: "CenterV", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'CenterV' text vertically centered in 5-row container. Equal empty space above and below.")
}

func TestSnapshot_Row_CrossAlignEnd(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisEnd,
		Children: []Widget{
			Text{Content: "Bottom", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Blue 'Bottom' text at row 5 (bottom edge). Empty space above the text.")
}

func TestSnapshot_Row_WithSpacing(t *testing.T) {
	widget := Row{
		Spacing: 2,
		Children: []Widget{
			Text{Content: "X", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Y", Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "Z", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Three items in horizontal row with 2-column gaps. Red 'X' at column 1, green 'Y' at column 4, blue 'Z' at column 7.")
}

func TestSnapshot_Row_NestedRows(t *testing.T) {
	widget := Row{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Style: Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Inner1"},
				},
			},
			Row{
				Style: Style{BackgroundColor: layoutBlue},
				Children: []Widget{
					Text{Content: "Inner2"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"Gray outer row with two nested rows side by side. Red row with 'Inner1' on left, blue row with 'Inner2' on right.")
}

func TestSnapshot_Row_MixedDimensions(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "Fixed", Width: Cells(5), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Flex", Width: Flex(1), Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "Auto", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"30-cell row with mixed widths. Red 'Fixed' takes 5 columns, green 'Flex' expands to fill remaining space, blue 'Auto' sized to content.")
}

// =============================================================================
// Dock Widget Tests
// =============================================================================

func TestSnapshot_Dock_TopOnly(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Text{Content: "Header", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Body", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Red 'Header' docked at top (row 1, full width). Blue 'Body' fills remaining space below.")
}

func TestSnapshot_Dock_BottomOnly(t *testing.T) {
	widget := Dock{
		Bottom: []Widget{
			Text{Content: "Footer", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Body", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Red 'Footer' docked at bottom (row 10, full width). Blue 'Body' fills remaining space above.")
}

func TestSnapshot_Dock_LeftOnly(t *testing.T) {
	widget := Dock{
		Left: []Widget{
			Text{Content: "Side", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Main", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Red 'Side' docked at left (full height). Blue 'Main' fills remaining space to the right.")
}

func TestSnapshot_Dock_RightOnly(t *testing.T) {
	widget := Dock{
		Right: []Widget{
			Text{Content: "Aside", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Main", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Red 'Aside' docked at right (full height). Blue 'Main' fills remaining space to the left.")
}

func TestSnapshot_Dock_AllEdges(t *testing.T) {
	widget := Dock{
		Top:    []Widget{Text{Content: "Broken", Style: Style{BackgroundColor: layoutRed}}},
		Bottom: []Widget{Text{Content: "Bottom", Style: Style{BackgroundColor: layoutOrange}}},
		Left:   []Widget{Text{Content: "Left", Style: Style{BackgroundColor: layoutGreen}}},
		Right:  []Widget{Text{Content: "Right", Style: Style{BackgroundColor: layoutPurple}}},
		Body:   Text{Content: "Center", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 40, 10,
		"All edges docked. Red top, orange bottom, green left, purple right. Blue 'Center' fills middle area.")
}

func TestSnapshot_Dock_BodyFillsRemainder(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Text{Content: "Header", Height: Cells(2), Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Content fills the rest", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Red 'Header' takes 2 rows at top. Blue body with text fills remaining 8 rows.")
}

func TestSnapshot_Dock_MultipleTop(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Text{Content: "Toolbar1", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Toolbar2", Style: Style{BackgroundColor: layoutOrange}},
		},
		Body: Text{Content: "Content", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Two widgets docked at top: red 'Toolbar1' on row 1, orange 'Toolbar2' on row 2. Blue 'Content' below.")
}

// =============================================================================
// Dimension Tests
// =============================================================================

func TestSnapshot_Dimension_AutoWidth(t *testing.T) {
	widget := Text{Content: "Auto sized", Style: Style{BackgroundColor: layoutBlue}}
	AssertSnapshot(t, widget, 20, 3,
		"Blue text 'Auto sized' at top-left. Width automatically sized to 10 characters (content width).")
}

func TestSnapshot_Dimension_CellsFixed(t *testing.T) {
	widget := Column{
		Width:  Cells(10),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutBlue},
		Children: []Widget{
			Text{Content: "Fixed"},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"Blue column exactly 10 cells wide by 5 cells tall. 'Fixed' text at top-left of column.")
}

func TestSnapshot_Dimension_FlexProportional(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "1", Width: Flex(1), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "2", Width: Flex(2), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"30-cell row split proportionally. Red '1' takes 10 cells (1/3), green '2' takes 20 cells (2/3).")
}

func TestSnapshot_Dimension_FlexVsCells(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "Fixed", Width: Cells(10), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Flex", Width: Flex(1), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"30-cell row. Red 'Fixed' takes exactly 10 cells. Green 'Flex' expands to fill remaining 20 cells.")
}

func TestSnapshot_Dimension_NestedFlex(t *testing.T) {
	widget := Column{
		Height: Cells(10),
		Children: []Widget{
			Row{
				Height: Flex(1),
				Style:  Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Nested Flex"},
				},
			},
			Row{
				Height: Flex(1),
				Style:  Style{BackgroundColor: layoutBlue},
				Children: []Widget{
					Text{Content: "Another Flex"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"10-row column split equally. Red row with 'Nested Flex' takes top 5 rows, blue row with 'Another Flex' takes bottom 5 rows.")
}

// =============================================================================
// Combined Layout Tests
// =============================================================================

func TestSnapshot_Layout_RowInColumn(t *testing.T) {
	widget := Column{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Style: Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Left"},
					Text{Content: "Right"},
				},
			},
			Text{Content: "Below", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Gray column containing red row on top, blue text below. Red row has 'Left' and 'Right' side by side.")
}

func TestSnapshot_Layout_ColumnInRow(t *testing.T) {
	widget := Row{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Column{
				Style: Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Top"},
					Text{Content: "Bottom"},
				},
			},
			Text{Content: "Beside", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Gray row containing red column on left, blue 'Beside' on right. Red column has 'Top' and 'Bottom' stacked.")
}

func TestSnapshot_Layout_DockWithRowColumn(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Row{
				Style: Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "Logo"},
					Text{Content: "Menu"},
				},
			},
		},
		Body: Column{
			Style: Style{BackgroundColor: layoutBlue},
			Children: []Widget{
				Text{Content: "Section1"},
				Text{Content: "Section2"},
			},
		},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Dock with red header row containing 'Logo' and 'Menu'. Blue column body with 'Section1' and 'Section2' stacked below.")
}

// =============================================================================
// Stack Widget Tests
// =============================================================================

func TestSnapshot_Stack_BasicOverlay(t *testing.T) {
	// Two children stacked - red on bottom, green on top (partially overlapping)
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Bottom", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Top", Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 25, 7,
		"Gray 20x5 stack. Green 'Top' overlays red 'Bottom', both at top-left. Green fully covers red since they overlap at same position.")
}

func TestSnapshot_Stack_ThreeLayersZOrder(t *testing.T) {
	// Three children stacked - verifies z-order (last child on top)
	widget := Stack{
		Width:  Cells(25),
		Height: Cells(6),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Layer1-Back", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Layer2-Mid", Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "Layer3-Top", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 30, 8,
		"Gray 25x6 stack with three overlapping layers. Blue 'Layer3-Top' visible on top, covering green and red beneath.")
}

func TestSnapshot_Stack_SizesFromLargestChild(t *testing.T) {
	// Stack should size based on largest non-positioned child
	widget := Stack{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Short", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "This is much longer", Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"Stack auto-sized to fit longest child. Green 'This is much longer' visible, red 'Short' hidden beneath. Stack width matches green text.")
}

// Alignment tests for non-positioned children

func TestSnapshot_Stack_AlignTopStart(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:    Cells(6),
		Alignment: AlignTopStart,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "TopLeft", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'TopLeft' at top-left corner (row 1, column 1).")
}

func TestSnapshot_Stack_AlignCenter(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:    Cells(6),
		Alignment: AlignCenter,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Center", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'Center' at center of stack, both horizontally and vertically.")
}

func TestSnapshot_Stack_AlignBottomEnd(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:    Cells(6),
		Alignment: AlignBottomEnd,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "BotRight", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'BotRight' at bottom-right corner (row 6, right-aligned).")
}

func TestSnapshot_Stack_AlignBottomCenter(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:    Cells(6),
		Alignment: AlignBottomCenter,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "BotMid", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'BotMid' at bottom, horizontally centered (row 6).")
}

// Positioned children tests

func TestSnapshot_Stack_PositionedTopLeft(t *testing.T) {
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(6),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:   IntPtr(1),
				Left:  IntPtr(2),
				Child: Text{Content: "At 2,1", Style: Style{BackgroundColor: layoutBlue}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'At 2,1' positioned at row 2 (1 from top), column 3 (2 from left).")
}

func TestSnapshot_Stack_PositionedBottomRight(t *testing.T) {
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(6),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Bottom: IntPtr(1),
				Right:  IntPtr(2),
				Child:  Text{Content: "BotRight", Style: Style{BackgroundColor: layoutBlue}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'BotRight' positioned 1 row from bottom, 2 columns from right edge.")
}

func TestSnapshot_Stack_PositionedFill(t *testing.T) {
	// PositionedFill should stretch to fill entire stack
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			PositionedFill(Text{Content: "Fills", Style: Style{BackgroundColor: layoutBlue}}),
		},
	}
	AssertSnapshot(t, widget, 25, 7,
		"Gray 20x5 stack completely filled by blue background. Text 'Fills' at top-left, blue covers entire stack area.")
}

func TestSnapshot_Stack_PositionedStretchHorizontal(t *testing.T) {
	// Both Left and Right set = child width is computed
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:   IntPtr(1),
				Left:  IntPtr(2),
				Right: IntPtr(2),
				Child: Text{Content: "Stretched H", Style: Style{BackgroundColor: layoutBlue}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 7,
		"Gray 20x5 stack. Blue text at row 2, stretched horizontally with 2-cell margins on left and right (16 cells wide).")
}

func TestSnapshot_Stack_PositionedStretchVertical(t *testing.T) {
	// Both Top and Bottom set = child height is computed
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(6),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:    IntPtr(1),
				Bottom: IntPtr(1),
				Left:   IntPtr(2),
				Child:  Text{Content: "V", Style: Style{BackgroundColor: layoutBlue}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack. Blue 'V' stretched vertically from row 2 to row 5 (4 rows), starting at column 3.")
}

// Overflow tests

func TestSnapshot_Stack_PositionedOverflowNegativeOffset(t *testing.T) {
	// Positioned child extends beyond stack bounds with negative offset
	widget := Stack{
		Width:  Cells(15),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Base", Style: Style{BackgroundColor: layoutBlue}},
			Positioned{
				Top:   IntPtr(-1),
				Right: IntPtr(-2),
				Child: Text{Content: "Badge", Style: Style{BackgroundColor: layoutRed}},
			},
		},
	}
	AssertSnapshot(t, widget, 20, 7,
		"Gray 15x5 stack with blue 'Base' at top-left. Red 'Badge' overflows stack bounds: 1 row above top, 2 columns right of stack edge.")
}

func TestSnapshot_Stack_ChildLargerThanStack(t *testing.T) {
	// Child widget larger than stack container
	widget := Stack{
		Width:  Cells(10),
		Height: Cells(3),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "This text is too long for container", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 15, 5,
		"Gray 10x3 stack. Blue text exceeds stack width, content overflows and is clipped at stack boundary.")
}

// Transparency / layering interaction tests

func TestSnapshot_Stack_OverlappingWithTransparency(t *testing.T) {
	// Demonstrates layering - top widget partially covers bottom
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:  IntPtr(0),
				Left: IntPtr(0),
				Child: Column{
					Width:  Cells(15),
					Height: Cells(4),
					Style:  Style{BackgroundColor: layoutRed},
					Children: []Widget{
						Text{Content: "Background"},
						Text{Content: "Content"},
					},
				},
			},
			Positioned{
				Top:  IntPtr(1),
				Left: IntPtr(5),
				Child: Column{
					Width:  Cells(12),
					Height: Cells(3),
					Style:  Style{BackgroundColor: layoutBlue},
					Children: []Widget{
						Text{Content: "Overlay"},
					},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 7,
		"Gray 20x5 stack with layered content. Red 15x4 column at origin with 'Background' and 'Content'. Blue 12x3 'Overlay' partially covers red, offset to row 2, column 6.")
}

func TestSnapshot_Stack_MultipleOverlappingPositioned(t *testing.T) {
	// Multiple positioned children creating complex overlap
	widget := Stack{
		Width:  Cells(25),
		Height: Cells(8),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:  IntPtr(0),
				Left: IntPtr(0),
				Child: Column{
					Width:  Cells(12),
					Height: Cells(5),
					Style:  Style{BackgroundColor: layoutRed},
					Children: []Widget{
						Text{Content: "Card 1"},
					},
				},
			},
			Positioned{
				Top:  IntPtr(2),
				Left: IntPtr(6),
				Child: Column{
					Width:  Cells(12),
					Height: Cells(5),
					Style:  Style{BackgroundColor: layoutGreen},
					Children: []Widget{
						Text{Content: "Card 2"},
					},
				},
			},
			Positioned{
				Top:  IntPtr(4),
				Left: IntPtr(12),
				Child: Column{
					Width:  Cells(12),
					Height: Cells(4),
					Style:  Style{BackgroundColor: layoutBlue},
					Children: []Widget{
						Text{Content: "Card 3"},
					},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Gray 25x8 stack with three cascading cards. Red 'Card 1' at top-left, green 'Card 2' overlaps at row 3/col 7, blue 'Card 3' overlaps at row 5/col 13. Each card partially visible.")
}

// Styling tests

func TestSnapshot_Stack_WithBorder(t *testing.T) {
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style: Style{
			Border:          Border{Style: BorderRounded, Color: layoutPurple},
			BackgroundColor: layoutGray,
		},
		Children: []Widget{
			Text{Content: "Bordered Stack", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 7,
		"Gray 20x5 stack with purple rounded border. Blue 'Bordered Stack' at top-left inside border. Border adds 1 cell each side.")
}

func TestSnapshot_Stack_WithPadding(t *testing.T) {
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(6),
		Style: Style{
			Padding:         EdgeInsetsAll(1),
			BackgroundColor: layoutGray,
		},
		Children: []Widget{
			Text{Content: "Padded", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack with 1-cell padding. Blue 'Padded' inset by 1 cell from all edges, gray padding visible around content.")
}

func TestSnapshot_Stack_WithBorderAndPadding(t *testing.T) {
	widget := Stack{
		Width:  Cells(22),
		Height: Cells(7),
		Style: Style{
			Border:          Border{Style: BorderSquare, Color: layoutPurple},
			Padding:         EdgeInsetsAll(1),
			BackgroundColor: layoutGray,
		},
		Children: []Widget{
			Positioned{
				Top:  IntPtr(0),
				Left: IntPtr(0),
				Child: Text{Content: "At origin", Style: Style{BackgroundColor: layoutBlue}},
			},
		},
	}
	AssertSnapshot(t, widget, 27, 9,
		"Gray 22x7 stack with purple square border and 1-cell padding. Blue 'At origin' positioned at border-box origin, overlapping the border/padding area.")
}

// Combined with other layouts

func TestSnapshot_Stack_InsideColumn(t *testing.T) {
	widget := Column{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Header", Style: Style{BackgroundColor: layoutRed}},
			Stack{
				Width:  Cells(20),
				Height: Cells(4),
				Style:  Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "Stacked", Style: Style{BackgroundColor: layoutBlue}},
				},
			},
			Text{Content: "Footer", Style: Style{BackgroundColor: layoutOrange}},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray column with three children stacked vertically. Red 'Header' at top, teal 20x4 stack with blue 'Stacked' in middle, orange 'Footer' at bottom.")
}

func TestSnapshot_Stack_InsideRow(t *testing.T) {
	widget := Row{
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Left", Style: Style{BackgroundColor: layoutRed}},
			Stack{
				Width:  Cells(12),
				Height: Cells(4),
				Style:  Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "Stack", Style: Style{BackgroundColor: layoutBlue}},
				},
			},
			Text{Content: "Right", Style: Style{BackgroundColor: layoutOrange}},
		},
	}
	AssertSnapshot(t, widget, 30, 6,
		"Gray row with three children side by side. Red 'Left' on left, teal 12x4 stack with blue 'Stack' in middle, orange 'Right' on right.")
}

func TestSnapshot_Stack_NestedStacks(t *testing.T) {
	widget := Stack{
		Width:  Cells(25),
		Height: Cells(8),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:  IntPtr(0),
				Left: IntPtr(0),
				Child: Stack{
					Width:  Cells(15),
					Height: Cells(5),
					Style:  Style{BackgroundColor: layoutRed},
					Children: []Widget{
						Text{Content: "Inner Stack", Style: Style{BackgroundColor: layoutBlue}},
					},
				},
			},
			Positioned{
				Bottom: IntPtr(0),
				Right:  IntPtr(0),
				Child:  Text{Content: "Outer", Style: Style{BackgroundColor: layoutGreen}},
			},
		},
	}
	AssertSnapshot(t, widget, 30, 10,
		"Gray 25x8 outer stack. Red 15x5 inner stack at top-left with blue 'Inner Stack'. Green 'Outer' at bottom-right of outer stack.")
}

// Mixed positioned and non-positioned children

func TestSnapshot_Stack_MixedPositionedAndAligned(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:    Cells(6),
		Alignment: AlignCenter,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			// Non-positioned child uses alignment
			Text{Content: "Centered", Style: Style{BackgroundColor: layoutBlue}},
			// Positioned child ignores alignment
			Positioned{
				Top:   IntPtr(0),
				Right: IntPtr(0),
				Child: Text{Content: "TR", Style: Style{BackgroundColor: layoutRed}},
			},
			Positioned{
				Bottom: IntPtr(0),
				Left:   IntPtr(0),
				Child:  Text{Content: "BL", Style: Style{BackgroundColor: layoutGreen}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 8,
		"Gray 20x6 stack with center alignment. Blue 'Centered' at center (uses alignment). Red 'TR' at top-right corner, green 'BL' at bottom-left (positioned, ignore alignment).")
}

// =============================================================================
// Percentage Dimension Tests
// =============================================================================

// --- Basic Percentage Tests (in Cells containers) ---

func TestSnapshot_Dimension_PercentWidth50(t *testing.T) {
	// 50% of 20 = 10 cells
	widget := Row{
		Width: Cells(20),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
		},
	}
	AssertSnapshot(t, widget, 25, 3)
}

func TestSnapshot_Dimension_PercentWidth100(t *testing.T) {
	// 100% of 20 = 20 cells (should fill parent completely)
	widget := Row{
		Width: Cells(20),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Full", Width: Percent(100), Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 3)
}

func TestSnapshot_Dimension_PercentTwoChildren(t *testing.T) {
	// 30% + 70% = 100% (should fill parent completely)
	widget := Row{
		Width: Cells(30),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "30%", Width: Percent(30), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "70%", Width: Percent(70), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 35, 3)
}

func TestSnapshot_Dimension_PercentOverflow(t *testing.T) {
	// 60% + 60% = 120% (intentional overflow beyond parent)
	widget := Row{
		Width: Cells(20),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "60%", Width: Percent(60), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "60%", Width: Percent(60), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 3)
}

func TestSnapshot_Dimension_PercentZero(t *testing.T) {
	// 0% should result in zero width
	widget := Row{
		Width: Cells(20),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "0%", Width: Percent(0), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Auto", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 3)
}

// --- Percentage Height Tests (in Cells containers) ---

func TestSnapshot_Dimension_PercentHeight(t *testing.T) {
	// 50% of 10 = 5 cells tall
	widget := Column{
		Height: Cells(10),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "50%", Height: Percent(50), Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 12)
}

func TestSnapshot_Dimension_PercentInColumn(t *testing.T) {
	// 25% + 25% + 50% = 100% (should fill parent completely)
	widget := Column{
		Height: Cells(10),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "25%", Height: Percent(25), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "25%", Height: Percent(25), Style: Style{BackgroundColor: layoutGreen}},
			Text{Content: "50%", Height: Percent(50), Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 12)
}

// --- Percentage Mixed with Other Dimension Types ---

func TestSnapshot_Dimension_PercentMixedWithCells(t *testing.T) {
	// Fixed 10 cells + 50% of 30 = 10 + 15 = 25 cells (5 cells gap)
	widget := Row{
		Width: Cells(30),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "Fixed", Width: Cells(10), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 35, 3)
}

func TestSnapshot_Dimension_PercentMixedWithFlex(t *testing.T) {
	// 30% of 30 = 9 cells, Flex(1) fills remaining 21 cells
	widget := Row{
		Width: Cells(30),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "30%", Width: Percent(30), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Flex", Width: Flex(1), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 35, 3)
}

func TestSnapshot_Dimension_PercentMixedWithAuto(t *testing.T) {
	// 50% of 30 = 15 cells, Auto = content width (4 cells for "Auto")
	widget := Row{
		Width: Cells(30),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Auto", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 35, 3)
}

// =============================================================================
// Dimension Constraint Tests
// =============================================================================

func TestSnapshot_Dimension_AutoHeightWithMaxHeight(t *testing.T) {
	content := "one two three four five six seven eight nine ten eleven twelve thirteen"
	widget := Column{
		Style: Style{Width: Cells(12), Height: Cells(6), BackgroundColor: layoutGray},
		Children: []Widget{
			Text{
				Content: content,
				Wrap:    WrapSoft,
				Style: Style{
					BackgroundColor: layoutBlue,
					Height:          Auto,
					MaxHeight:       Cells(3),
				},
			},
		},
	}
	AssertSnapshot(t, widget, 14, 8,
		"Blue wrapped text would exceed 3 lines but is clamped by MaxHeight. Gray container remains 12x6.")
}

func TestSnapshot_Dimension_PercentHeightClampsTallContent(t *testing.T) {
	content := "alpha beta gamma delta epsilon zeta eta theta iota kappa lambda mu nu xi omicron pi rho"
	widget := Column{
		Style: Style{Width: Cells(12), Height: Cells(20), BackgroundColor: layoutGray},
		Children: []Widget{
			Text{
				Content: content,
				Wrap:    WrapSoft,
				Style: Style{
					BackgroundColor: layoutGreen,
					Height:          Percent(25), // 5 rows of 20
				},
			},
		},
	}
	AssertSnapshot(t, widget, 14, 22,
		"Green text is constrained to 25% height (5 rows) even though content would wrap taller. Gray container is 12x20.")
}

func TestSnapshot_Dimension_FlexHeightWithMaxHeight(t *testing.T) {
	content := "red orange yellow green blue indigo violet black white gray brown magenta cyan"
	widget := Column{
		Style: Style{Width: Cells(12), Height: Cells(20), BackgroundColor: layoutGray},
		Children: []Widget{
			Text{
				Content: content,
				Wrap:    WrapSoft,
				Style: Style{
					BackgroundColor: layoutPurple,
					Height:          Flex(1),
					MaxHeight:       Cells(5),
				},
			},
		},
	}
	AssertSnapshot(t, widget, 14, 22,
		"Purple text is flexed but capped at MaxHeight (5 rows) despite available space. Gray container is 12x20.")
}

// --- Percentage Inside Non-Cells Containers ---

func TestSnapshot_Dimension_PercentInsideFlexContainer(t *testing.T) {
	// Outer Row is 40 cells, inner Row is Flex(1) so fills 40 cells
	// 50% of 40 = 20 cells
	widget := Row{
		Width: Cells(40),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Width: Flex(1),
				Style: Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 45, 3)
}

func TestSnapshot_Dimension_PercentInsideFlexContainerMultiple(t *testing.T) {
	// Two Flex(1) containers each get 20 cells (half of 40)
	// 50% inside each = 10 cells each
	widget := Row{
		Width: Cells(40),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Width: Flex(1),
				Style: Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
				},
			},
			Row{
				Width: Flex(1),
				Style: Style{BackgroundColor: layoutPurple},
				Children: []Widget{
					Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutBlue}},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 45, 3)
}

func TestSnapshot_Dimension_PercentInsideAutoContainer(t *testing.T) {
	// Auto container shrink-wraps to content
	// When parent Row is 40 cells and inner Row is Auto, inner gets constraint max=40
	// 50% of 40 = 20 cells
	widget := Row{
		Width: Cells(40),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Width: Auto,
				Style: Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 45, 3)
}

func TestSnapshot_Dimension_PercentInsidePercentContainer(t *testing.T) {
	// Outer: 50% of 40 = 20 cells
	// Inner: 50% of 20 = 10 cells
	widget := Row{
		Width: Cells(40),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Width: Percent(50),
				Style: Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 45, 3)
}

func TestSnapshot_Dimension_PercentInsidePercentContainerDeep(t *testing.T) {
	// 3 levels: 50% of 50% of 50% of 40 = 5 cells
	widget := Row{
		Width: Cells(40),
		Style: Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Row{
				Width: Percent(50), // 20 cells
				Style: Style{BackgroundColor: layoutTeal},
				Children: []Widget{
					Row{
						Width: Percent(50), // 10 cells
						Style: Style{BackgroundColor: layoutPurple},
						Children: []Widget{
							Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}}, // 5 cells
						},
					},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 45, 3)
}

// --- Percentage in Dock Layout ---

func TestSnapshot_Dimension_PercentInDock(t *testing.T) {
	widget := Dock{
		Style: Style{BackgroundColor: layoutGray},
		Top: []Widget{
			Text{Content: "Header", Height: Percent(20), Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Body", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

// --- Percentage in Stack Layout ---

func TestSnapshot_Dimension_PercentInStackWidth(t *testing.T) {
	// Stack is 20x5, child is 50% width = 10 cells
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(5),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
		},
	}
	AssertSnapshot(t, widget, 25, 7)
}

func TestSnapshot_Dimension_PercentInStackHeight(t *testing.T) {
	// Stack is 20x10, child is 50% height = 5 rows
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(10),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Column{
				Width:  Cells(10),
				Height: Percent(50),
				Style:  Style{BackgroundColor: layoutBlue},
				Children: []Widget{
					Text{Content: "50%"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 12)
}

func TestSnapshot_Dimension_PercentInStackBothAxes(t *testing.T) {
	// Stack is 20x10, child is 50% width (10 cells) and 50% height (5 rows)
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(10),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Column{
				Width:  Percent(50),
				Height: Percent(50),
				Style:  Style{BackgroundColor: layoutRed},
				Children: []Widget{
					Text{Content: "50x50"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 12)
}

func TestSnapshot_Dimension_PercentInStackPositioned(t *testing.T) {
	// Stack is 20x10, positioned child at top-left with 50% width
	widget := Stack{
		Width:  Cells(20),
		Height: Cells(10),
		Style:  Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Positioned{
				Top:  IntPtr(0),
				Left: IntPtr(0),
				Child: Text{Content: "50%", Width: Percent(50), Style: Style{BackgroundColor: layoutRed}},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 12)
}
