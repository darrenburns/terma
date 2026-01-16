package terma

import (
	"testing"
)

// =============================================================================
// Border Style Tests
// =============================================================================

func TestSnapshot_Style_BorderSquare(t *testing.T) {
	widget := Column{
		Width:  Cells(15),
		Height: Cells(5),
		Style: Style{
			Border: SquareBorder(RGB(200, 200, 200)),
		},
		Children: []Widget{
			Text{Content: "Square"},
		},
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Style_BorderRounded(t *testing.T) {
	widget := Column{
		Width:  Cells(15),
		Height: Cells(5),
		Style: Style{
			Border: RoundedBorder(RGB(200, 200, 200)),
		},
		Children: []Widget{
			Text{Content: "Rounded"},
		},
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Style_BorderDouble(t *testing.T) {
	widget := Column{
		Width:  Cells(15),
		Height: Cells(5),
		Style: Style{
			Border: DoubleBorder(RGB(200, 200, 200)),
		},
		Children: []Widget{
			Text{Content: "Double"},
		},
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Style_BorderHeavy(t *testing.T) {
	widget := Column{
		Width:  Cells(15),
		Height: Cells(5),
		Style: Style{
			Border: HeavyBorder(RGB(200, 200, 200)),
		},
		Children: []Widget{
			Text{Content: "Heavy"},
		},
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Style_BorderAscii(t *testing.T) {
	widget := Column{
		Width:  Cells(15),
		Height: Cells(5),
		Style: Style{
			Border: AsciiBorder(RGB(200, 200, 200)),
		},
		Children: []Widget{
			Text{Content: "ASCII"},
		},
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Style_BorderWithTitle(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(5),
		Style: Style{
			Border: SquareBorder(RGB(200, 200, 200), BorderTitle("Title")),
		},
		Children: []Widget{
			Text{Content: "Content"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Style_BorderWithSubtitle(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(5),
		Style: Style{
			Border: RoundedBorder(RGB(200, 200, 200), BorderSubtitle("Footer")),
		},
		Children: []Widget{
			Text{Content: "Body"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

// =============================================================================
// Padding Tests
// =============================================================================

func TestSnapshot_Style_PaddingAllSides(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(7),
		Style: Style{
			Padding:         EdgeInsetsAll(2),
			BackgroundColor: RGB(50, 50, 100),
		},
		Children: []Widget{
			Text{Content: "Padded"},
		},
	}
	AssertSnapshot(t, widget, 20, 7)
}

func TestSnapshot_Style_PaddingAsymmetric(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(7),
		Style: Style{
			Padding:         EdgeInsetsTRBL(1, 3, 1, 2),
			BackgroundColor: RGB(50, 100, 50),
		},
		Children: []Widget{
			Text{Content: "Asymmetric"},
		},
	}
	AssertSnapshot(t, widget, 20, 7)
}

func TestSnapshot_Style_PaddingXY(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(7),
		Style: Style{
			Padding:         EdgeInsetsXY(3, 1),
			BackgroundColor: RGB(100, 50, 50),
		},
		Children: []Widget{
			Text{Content: "XY Padding"},
		},
	}
	AssertSnapshot(t, widget, 20, 7)
}

// =============================================================================
// Margin Tests
// =============================================================================

func TestSnapshot_Style_MarginAllSides(t *testing.T) {
	widget := Column{
		Style: Style{
			BackgroundColor: RGB(30, 30, 60),
		},
		Children: []Widget{
			Column{
				Width:  Cells(15),
				Height: Cells(3),
				Style: Style{
					Margin:          EdgeInsetsAll(1),
					BackgroundColor: RGB(100, 100, 150),
				},
				Children: []Widget{
					Text{Content: "Margin"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 20, 7)
}

// =============================================================================
// Color Tests
// =============================================================================

func TestSnapshot_Style_BackgroundColor(t *testing.T) {
	widget := Text{
		Content: "With Background",
		Style: Style{
			BackgroundColor: RGB(100, 50, 150),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_BackdropGradient(t *testing.T) {
	widget := Column{
		Width:  Flex(1),
		Height: Flex(1),
		Style: Style{
			BackgroundColor: RGB(20, 20, 20),
		},
		Children: []Widget{
			Row{
				Width:  Flex(1),
				Height: Cells(3),
				Style: Style{
					BackgroundColor: NewGradient(
						RGB(255, 120, 120).WithAlpha(0.5),
						RGB(120, 120, 255).WithAlpha(0.5),
					).WithAngle(90),
				},
				Children: []Widget{
					Text{
						Content: "Gradient",
						Style: Style{
							ForegroundColor: RGB(240, 240, 240),
						},
					},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Style_ForegroundColor(t *testing.T) {
	widget := Text{
		Content: "Colored Text",
		Style: Style{
			ForegroundColor: RGB(255, 128, 0),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_BothColors(t *testing.T) {
	widget := Text{
		Content: "Full Color",
		Style: Style{
			ForegroundColor: RGB(255, 255, 255),
			BackgroundColor: RGB(0, 100, 200),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

// =============================================================================
// Text Style Tests
// =============================================================================

func TestSnapshot_Style_Bold(t *testing.T) {
	widget := Text{
		Content: "Bold Text",
		Style:   Style{Bold: true},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_Italic(t *testing.T) {
	widget := Text{
		Content: "Italic Text",
		Style:   Style{Italic: true},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_Underline(t *testing.T) {
	widget := Text{
		Content: "Underlined Text",
		Style:   Style{Underline: UnderlineSingle},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_Strikethrough(t *testing.T) {
	widget := Text{
		Content: "Struck Text",
		Style:   Style{Strikethrough: true},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_CombinedTextStyles(t *testing.T) {
	widget := Text{
		Content: "Combined",
		Style: Style{
			Bold:   true,
			Italic: true,
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

// =============================================================================
// Border + Padding Combined Tests
// =============================================================================

func TestSnapshot_Style_BorderAndPadding(t *testing.T) {
	widget := Column{
		Width:  Cells(20),
		Height: Cells(7),
		Style: Style{
			Border:  RoundedBorder(RGB(100, 200, 100)),
			Padding: EdgeInsetsAll(1),
		},
		Children: []Widget{
			Text{Content: "Boxed"},
		},
	}
	AssertSnapshot(t, widget, 20, 7)
}

func TestSnapshot_Style_FullStyleStack(t *testing.T) {
	widget := Column{
		Width:  Cells(25),
		Height: Cells(9),
		Style: Style{
			Border:          SquareBorder(RGB(200, 200, 200), BorderTitle("Window")),
			Padding:         EdgeInsetsAll(1),
			BackgroundColor: RGB(30, 30, 50),
		},
		Children: []Widget{
			Text{
				Content: "Hello",
				Style: Style{
					ForegroundColor: RGB(255, 200, 100),
					Bold:            true,
				},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 9)
}

// =============================================================================
// Span Style Tests
// =============================================================================

func TestSnapshot_Style_SpanForeground(t *testing.T) {
	widget := Text{
		Spans: []Span{
			ColorSpan("Red", RGB(255, 0, 0)),
			PlainSpan(" and "),
			ColorSpan("Blue", RGB(0, 0, 255)),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_SpanBold(t *testing.T) {
	widget := Text{
		Spans: []Span{
			BoldSpan("Important"),
			PlainSpan(" text"),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Style_SpanItalic(t *testing.T) {
	widget := Text{
		Spans: []Span{
			ItalicSpan("Emphasis"),
			PlainSpan(" here"),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

// =============================================================================
// Named Color Tests
// =============================================================================

func TestSnapshot_Style_NamedColors(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Text{Content: "Red", Style: Style{ForegroundColor: Red}},
			Text{Content: "Green", Style: Style{ForegroundColor: Green}},
			Text{Content: "Blue", Style: Style{ForegroundColor: Blue}},
			Text{Content: "Yellow", Style: Style{ForegroundColor: Yellow}},
			Text{Content: "Magenta", Style: Style{ForegroundColor: Magenta}},
			Text{Content: "Cyan", Style: Style{ForegroundColor: Cyan}},
		},
	}
	AssertSnapshot(t, widget, 20, 8)
}

// =============================================================================
// Nested Style Tests
// =============================================================================

func TestSnapshot_Style_NestedBorders(t *testing.T) {
	widget := Column{
		Width:  Cells(25),
		Height: Cells(10),
		Style: Style{
			Border: RoundedBorder(RGB(100, 100, 200)),
		},
		Children: []Widget{
			Column{
				Style: Style{
					Border: SquareBorder(RGB(200, 100, 100)),
				},
				Children: []Widget{
					Text{Content: "Inner"},
				},
			},
		},
	}
	AssertSnapshot(t, widget, 25, 10)
}

func TestSnapshot_Style_RowWithStyledChildren(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{Content: "A", Style: Style{ForegroundColor: Red}},
			Text{Content: "B", Style: Style{ForegroundColor: Green}},
			Text{Content: "C", Style: Style{ForegroundColor: Blue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}
