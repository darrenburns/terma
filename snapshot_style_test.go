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
	AssertSnapshot(t, widget, 15, 5,
		"15x5 column with gray square border (┌─┐│└─┘ characters). 'Square' text inside, inset by 1 cell.")
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
	AssertSnapshot(t, widget, 15, 5,
		"15x5 column with gray rounded border (╭─╮│╰─╯ characters). 'Rounded' text inside, corners are curved.")
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
	AssertSnapshot(t, widget, 15, 5,
		"15x5 column with gray double-line border (╔═╗║╚═╝ characters). 'Double' text inside.")
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
	AssertSnapshot(t, widget, 15, 5,
		"15x5 column with gray heavy/thick border (┏━┓┃┗━┛ characters). 'Heavy' text inside.")
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
	AssertSnapshot(t, widget, 15, 5,
		"15x5 column with gray ASCII border (+-+|+-+ characters). 'ASCII' text inside.")
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
	AssertSnapshot(t, widget, 20, 5,
		"20x5 column with square border. 'Title' text embedded in top border line. 'Content' inside.")
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
	AssertSnapshot(t, widget, 20, 5,
		"20x5 column with rounded border. 'Footer' text embedded in bottom border line. 'Body' inside.")
}

func TestSnapshot_Style_BorderWithMarkupTitle(t *testing.T) {
	widget := Column{
		Width:  Cells(25),
		Height: Cells(5),
		Style: Style{
			Border: SquareBorder(RGB(200, 200, 200), BorderTitleMarkup("[b]Bold Title[/]")),
		},
		Children: []Widget{
			Text{Content: "Content"},
		},
	}
	AssertSnapshot(t, widget, 25, 5,
		"25x5 column with square border. 'Bold Title' in bold text embedded in top border. 'Content' inside.")
}

func TestSnapshot_Style_BorderWithMarkupColors(t *testing.T) {
	widget := Column{
		Width:  Cells(30),
		Height: Cells(5),
		Style: Style{
			Border: SquareBorder(RGB(200, 200, 200), BorderTitleMarkup("[b $Accent]ESC[/] close")),
		},
		Children: []Widget{
			Text{Content: "Dialog content"},
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"30x5 column with square border. Title 'ESC close' where 'ESC' is bold and accent-colored. 'Dialog content' inside.")
}

func TestSnapshot_Style_BorderMixedDecorations(t *testing.T) {
	widget := Column{
		Width:  Cells(30),
		Height: Cells(5),
		Style: Style{
			Border: SquareBorder(RGB(200, 200, 200),
				BorderTitleMarkup("[i]Styled[/]"),
				BorderTitleRight("Plain"),
			),
		},
		Children: []Widget{
			Text{Content: "Mixed decorations"},
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"30x5 column with square border. 'Styled' in italic at top-left, 'Plain' at top-right. 'Mixed decorations' inside.")
}

func TestSnapshot_Style_BorderGradientWithMarkupTitle(t *testing.T) {
	// Test that markup title text without explicit color samples from the gradient border
	widget := Column{
		Width:  Cells(30),
		Height: Cells(5),
		Style: Style{
			Border: Border{
				Style: BorderRounded,
				Color: NewGradient(RGB(255, 0, 0), RGB(0, 0, 255)).WithAngle(90), // Red to blue horizontal
				Decorations: []BorderDecoration{
					BorderTitleMarkup("[b]Gradient Title[/]"), // Bold, but no color - should sample from gradient
				},
			},
		},
		Children: []Widget{
			Text{Content: "Content"},
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"30x5 column with rounded gradient border (red to blue). 'Gradient Title' in bold, color sampled from gradient at title position.")
}

func TestSnapshot_Style_BorderGradientWithMarkupTitleExplicitColor(t *testing.T) {
	// Test that markup title with explicit color overrides the gradient
	widget := Column{
		Width:  Cells(30),
		Height: Cells(5),
		Style: Style{
			Border: Border{
				Style: BorderRounded,
				Color: NewGradient(RGB(255, 0, 0), RGB(0, 0, 255)).WithAngle(90), // Red to blue horizontal
				Decorations: []BorderDecoration{
					BorderTitleMarkup("[b #00ff00]Green Title[/]"), // Explicit green color should override gradient
				},
			},
		},
		Children: []Widget{
			Text{Content: "Content"},
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"30x5 column with rounded gradient border (red to blue). 'Green Title' in bold green, overriding the gradient color.")
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
	AssertSnapshot(t, widget, 20, 7,
		"20x7 dark blue column with 2-cell padding on all sides. 'Padded' text inset by 2 cells from each edge.")
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
	AssertSnapshot(t, widget, 20, 7,
		"20x7 dark green column with asymmetric padding: top=1, right=3, bottom=1, left=2. 'Asymmetric' text offset accordingly.")
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
	AssertSnapshot(t, widget, 20, 7,
		"20x7 dark red column with horizontal padding=3, vertical padding=1. 'XY Padding' text inset 3 from sides, 1 from top/bottom.")
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
	AssertSnapshot(t, widget, 20, 7,
		"Dark blue outer column. Light purple 15x3 inner column with 1-cell margin on all sides. Gap between inner and outer visible.")
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
	AssertSnapshot(t, widget, 20, 3,
		"White text 'With Background' on purple background (RGB 100,50,150). Background extends to text width.")
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
	AssertSnapshot(t, widget, 20, 3,
		"Orange text 'Colored Text' (RGB 255,128,0) on black background.")
}

func TestSnapshot_Style_BothColors(t *testing.T) {
	widget := Text{
		Content: "Full Color",
		Style: Style{
			ForegroundColor: RGB(255, 255, 255),
			BackgroundColor: RGB(0, 100, 200),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White text 'Full Color' on blue background (RGB 0,100,200). Both foreground and background colors applied.")
}

// =============================================================================
// Text Style Tests
// =============================================================================

func TestSnapshot_Style_Bold(t *testing.T) {
	widget := Text{
		Content: "Bold Text",
		Style:   Style{Bold: true},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Bold Text' in bold weight at top-left on black background.")
}

func TestSnapshot_Style_Italic(t *testing.T) {
	widget := Text{
		Content: "Italic Text",
		Style:   Style{Italic: true},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Italic Text' in italic style at top-left on black background.")
}

func TestSnapshot_Style_Underline(t *testing.T) {
	widget := Text{
		Content: "Underlined Text",
		Style:   Style{Underline: UnderlineSingle},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Underlined Text' with single underline at top-left on black background.")
}

func TestSnapshot_Style_Strikethrough(t *testing.T) {
	widget := Text{
		Content: "Struck Text",
		Style:   Style{Strikethrough: true},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Struck Text' with strikethrough line at top-left on black background.")
}

func TestSnapshot_Style_CombinedTextStyles(t *testing.T) {
	widget := Text{
		Content: "Combined",
		Style: Style{
			Bold:   true,
			Italic: true,
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Combined' text in both bold and italic at top-left on black background.")
}

func TestSnapshot_Style_Reverse(t *testing.T) {
	widget := Text{
		Content: "Reversed Text",
		Style:   Style{Reverse: true},
	}
	AssertSnapshot(t, widget, 20, 3,
		"'Reversed Text' with reversed colors - theme text color becomes background, black text. Background should be continuous across the space.")
}

func TestSnapshot_Style_ReverseWithColors(t *testing.T) {
	widget := Text{
		Content: "Reversed",
		Style: Style{
			ForegroundColor: RGB(255, 100, 100),
			BackgroundColor: RGB(50, 50, 150),
			Reverse:         true,
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"'Reversed' text with colors swapped - light red background (#FF6464), dark blue text (#323296).")
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
	AssertSnapshot(t, widget, 20, 7,
		"20x7 column with green rounded border. 'Boxed' text inset by border (1 cell) plus padding (1 cell) = 2 cells from each edge.")
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
	AssertSnapshot(t, widget, 25, 9,
		"25x9 column with gray square border, 'Window' title in top border, dark blue background. Orange bold 'Hello' text inset 2 cells (border+padding).")
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
	AssertSnapshot(t, widget, 20, 3,
		"Single line with mixed colors: 'Red' in red, ' and ' in white, 'Blue' in blue. All on black background.")
}

func TestSnapshot_Style_SpanBold(t *testing.T) {
	widget := Text{
		Spans: []Span{
			BoldSpan("Important"),
			PlainSpan(" text"),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Single line with 'Important' in bold followed by ' text' in normal weight. White on black.")
}

func TestSnapshot_Style_SpanItalic(t *testing.T) {
	widget := Text{
		Spans: []Span{
			ItalicSpan("Emphasis"),
			PlainSpan(" here"),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Single line with 'Emphasis' in italic followed by ' here' in normal style. White on black.")
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
	AssertSnapshot(t, widget, 20, 8,
		"Six text rows showing named colors. 'Red' in red on row 1, 'Green' in green on row 2, 'Blue' in blue on row 3, 'Yellow' in yellow on row 4, 'Magenta' in magenta on row 5, 'Cyan' in cyan on row 6.")
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
	AssertSnapshot(t, widget, 25, 10,
		"Outer 25x10 column with blue rounded border. Inner column with red square border nested inside. 'Inner' text inside the inner border.")
}

func TestSnapshot_Style_RowWithStyledChildren(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{Content: "A", Style: Style{ForegroundColor: Red}},
			Text{Content: "B", Style: Style{ForegroundColor: Green}},
			Text{Content: "C", Style: Style{ForegroundColor: Blue}},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Row with three colored letters: red 'A', green 'B', blue 'C' arranged horizontally from left to right.")
}
