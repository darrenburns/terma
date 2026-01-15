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
		Top:    []Widget{Text{Content: "Broken", Style: Style{BackgroundColor: layoutRed}}},
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
	AssertSnapshot(t, widget, 25, 7)
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
	AssertSnapshot(t, widget, 30, 8)
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
	AssertSnapshot(t, widget, 30, 5)
}

// Alignment tests for non-positioned children

func TestSnapshot_Stack_AlignTopStart(t *testing.T) {
	widget := Stack{
		Width:     Cells(20),
		Height:   Cells(6),
		Alignment: AlignTopStart,
		Style:     Style{BackgroundColor: layoutGray},
		Children: []Widget{
			Text{Content: "TopLeft", Style: Style{BackgroundColor: layoutBlue}},
		},
	}
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 25, 7)
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
	AssertSnapshot(t, widget, 25, 7)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 20, 7)
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
	AssertSnapshot(t, widget, 15, 5)
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
	AssertSnapshot(t, widget, 25, 7)
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
	AssertSnapshot(t, widget, 30, 10)
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
	AssertSnapshot(t, widget, 25, 7)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 27, 9)
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
	AssertSnapshot(t, widget, 25, 8)
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
	AssertSnapshot(t, widget, 30, 6)
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
	AssertSnapshot(t, widget, 30, 10)
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
	AssertSnapshot(t, widget, 25, 8)
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
