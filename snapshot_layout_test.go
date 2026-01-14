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
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Column_MainAlignStart(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisStart,
		Children: []Widget{
			Text{Content: "Top", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10)
}

func TestSnapshot_Column_MainAlignCenter(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisCenter,
		Children: []Widget{
			Text{Content: "Centered", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10)
}

func TestSnapshot_Column_MainAlignEnd(t *testing.T) {
	widget := Column{
		Height:    Cells(10),
		MainAlign: MainAxisEnd,
		Children: []Widget{
			Text{Content: "Bottom", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 10)
}

func TestSnapshot_Column_CrossAlignStretch(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisStretch,
		Children: []Widget{
			Text{Content: "Stretched", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Column_CrossAlignStart(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisStart,
		Children: []Widget{
			Text{Content: "Left", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Column_CrossAlignCenter(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisCenter,
		Children: []Widget{
			Text{Content: "CenterH", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Column_CrossAlignEnd(t *testing.T) {
	widget := Column{
		Width:      Cells(20),
		CrossAlign: CrossAxisEnd,
		Children: []Widget{
			Text{Content: "Right", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 10)
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
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 10)
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
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Row_MainAlignStart(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisStart,
		Children: []Widget{
			Text{Content: "Left", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Row_MainAlignCenter(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisCenter,
		Children: []Widget{
			Text{Content: "Mid", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Row_MainAlignEnd(t *testing.T) {
	widget := Row{
		Width:     Cells(20),
		MainAlign: MainAxisEnd,
		Children: []Widget{
			Text{Content: "Right", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Row_CrossAlignStretch(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisStretch,
		Children: []Widget{
			Text{Content: "Tall", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Row_CrossAlignStart(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisStart,
		Children: []Widget{
			Text{Content: "Top", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Row_CrossAlignCenter(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisCenter,
		Children: []Widget{
			Text{Content: "CenterV", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Row_CrossAlignEnd(t *testing.T) {
	widget := Row{
		Height:     Cells(5),
		CrossAlign: CrossAxisEnd,
		Children: []Widget{
			Text{Content: "Bottom", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 3)
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
	AssertSnapshot(t, widget, 30, 3)
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
	AssertSnapshot(t, widget, 30, 3)
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
	AssertSnapshot(t, widget, 30, 10)
}

func TestSnapshot_Dock_BottomOnly(t *testing.T) {
	widget := Dock{
		Bottom: []Widget{
			Text{Content: "Footer", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Body", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

func TestSnapshot_Dock_LeftOnly(t *testing.T) {
	widget := Dock{
		Left: []Widget{
			Text{Content: "Side", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Main", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

func TestSnapshot_Dock_RightOnly(t *testing.T) {
	widget := Dock{
		Right: []Widget{
			Text{Content: "Aside", Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Main", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

func TestSnapshot_Dock_AllEdges(t *testing.T) {
	widget := Dock{
		Top:    []Widget{Text{Content: "T0p", Style: Style{BackgroundColor: layoutRed}}},
		Bottom: []Widget{Text{Content: "Bottom", Style: Style{BackgroundColor: layoutOrange}}},
		Left:   []Widget{Text{Content: "Left", Style: Style{BackgroundColor: layoutGreen}}},
		Right:  []Widget{Text{Content: "Right", Style: Style{BackgroundColor: layoutPurple}}},
		Body:   Text{Content: "Center", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 40, 10)
}

func TestSnapshot_Dock_BodyFillsRemainder(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Text{Content: "Header", Height: Cells(2), Style: Style{BackgroundColor: layoutRed}},
		},
		Body: Text{Content: "Content fills the rest", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

func TestSnapshot_Dock_MultipleTop(t *testing.T) {
	widget := Dock{
		Top: []Widget{
			Text{Content: "Toolbar1", Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Toolbar2", Style: Style{BackgroundColor: layoutOrange}},
		},
		Body: Text{Content: "Content", Style: Style{BackgroundColor: layoutBlue}},
	}
	AssertSnapshot(t, widget, 30, 10)
}

// =============================================================================
// Dimension Tests
// =============================================================================

func TestSnapshot_Dimension_AutoWidth(t *testing.T) {
	widget := Text{Content: "Auto sized", Style: Style{BackgroundColor: layoutBlue}}
	AssertSnapshot(t, widget, 20, 3)
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
	AssertSnapshot(t, widget, 20, 10)
}

func TestSnapshot_Dimension_FlexProportional(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "1", Width: Flex(1), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "2", Width: Flex(2), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 3)
}

func TestSnapshot_Dimension_FlexVsCells(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "Fixed", Width: Cells(10), Style: Style{BackgroundColor: layoutRed}},
			Text{Content: "Flex", Width: Flex(1), Style: Style{BackgroundColor: layoutGreen}},
		},
	}
	AssertSnapshot(t, widget, 30, 3)
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
	AssertSnapshot(t, widget, 20, 10)
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
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 30, 10)
}
