package terma

import (
	"testing"
)

// =============================================================================
// Text Widget Tests
// =============================================================================

func TestSnapshot_Text_PlainContent(t *testing.T) {
	widget := Text{Content: "Hello, World!!"}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Hello, World!!' text at top-left. Width auto-sized to 14 characters.")
}

func TestSnapshot_Text_RichSpans(t *testing.T) {
	widget := Text{
		Spans: []Span{
			{Text: "Bold", Style: SpanStyle{Bold: true}},
			{Text: " and "},
			{Text: "Italic", Style: SpanStyle{Italic: true}},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"Rich text: 'Bold' in bold, ' and ' in normal, 'Italic' in italic. All white on black, single line.")
}

func TestSnapshot_Text_WrapNone(t *testing.T) {
	widget := Text{
		Content: "This is a very long line that should not wrap",
		Wrap:    WrapNone,
	}
	AssertSnapshot(t, widget, 20, 3,
		"Long text extends beyond 20-cell boundary, no wrapping. Only first 20 characters visible: 'This is a very long '.")
}

func TestSnapshot_Text_WrapSoft(t *testing.T) {
	widget := Text{
		Content: "This is a line that should wrap at word boundaries",
		Wrap:    WrapSoft,
		Width:   Cells(15),
	}
	AssertSnapshot(t, widget, 15, 5,
		"Text wraps at word boundaries within 15-cell width. Multiple lines, words not broken mid-word.")
}

func TestSnapshot_Text_WrapHard(t *testing.T) {
	widget := Text{
		Content: "Supercalifragilisticexpialidocious",
		Wrap:    WrapHard,
		Width:   Cells(10),
	}
	AssertSnapshot(t, widget, 10, 5,
		"Long word broken at exactly 10 characters per line. Word split mid-character across multiple lines.")
}

func TestSnapshot_Text_BoldItalicUnderline(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Text{Content: "Bold", Style: Style{Bold: true}},
			Text{Content: "Italic", Style: Style{Italic: true}},
			Text{Content: "Underline", Style: Style{Underline: UnderlineSingle}},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Three text rows showing style variations. 'Bold' in bold on row 1, 'Italic' in italic on row 2, 'Underline' underlined on row 3.")
}

func TestSnapshot_Text_WithBackground(t *testing.T) {
	widget := Text{
		Content: "Highlighted",
		Style: Style{
			BackgroundColor: RGB(100, 100, 200),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"White 'Highlighted' text on purple/blue background. Background extends to text width only.")
}

func TestSnapshot_Text_Multiline(t *testing.T) {
	widget := Text{Content: "Line 1\nLine 2\nLine 3"}
	AssertSnapshot(t, widget, 20, 5,
		"Three lines of text from explicit newlines. 'Line 1' on row 1, 'Line 2' on row 2, 'Line 3' on row 3.")
}

func TestSnapshot_Text_WithForegroundColor(t *testing.T) {
	widget := Text{
		Content: "Colored",
		Style: Style{
			ForegroundColor: RGB(255, 100, 100),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Red/pink 'Colored' text on black background. Text color is RGB(255,100,100).")
}

// =============================================================================
// Text Alignment Tests
// =============================================================================

func TestSnapshot_Text_AlignLeft(t *testing.T) {
	widget := Text{
		Content:   "Left",
		TextAlign: TextAlignLeft,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Text 'Left' aligned to the left edge within 20-cell width. Default alignment behavior.")
}

func TestSnapshot_Text_AlignCenter(t *testing.T) {
	widget := Text{
		Content:   "Center",
		TextAlign: TextAlignCenter,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Text 'Center' horizontally centered within 20-cell width. Equal spacing on both sides.")
}

func TestSnapshot_Text_AlignRight(t *testing.T) {
	widget := Text{
		Content:   "Right",
		TextAlign: TextAlignRight,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Text 'Right' aligned to the right edge within 20-cell width. Text starts at column 15.")
}

func TestSnapshot_Text_AlignCenter_Multiline(t *testing.T) {
	widget := Text{
		Content:   "Line 1\nLonger Line 2\nL3",
		TextAlign: TextAlignCenter,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 5,
		"Three centered lines. Each line independently centered within 20-cell width.")
}

func TestSnapshot_Text_AlignRight_Multiline(t *testing.T) {
	widget := Text{
		Content:   "Short\nMedium text\nA",
		TextAlign: TextAlignRight,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 5,
		"Three right-aligned lines. Each line independently aligned to right edge.")
}

func TestSnapshot_Text_AlignCenter_WithWrap(t *testing.T) {
	widget := Text{
		Content:   "This is a line that wraps",
		TextAlign: TextAlignCenter,
		Wrap:      WrapSoft,
		Width:     Cells(15),
	}
	AssertSnapshot(t, widget, 15, 5,
		"Text wraps at word boundaries, each wrapped line is centered within 15-cell width.")
}

func TestSnapshot_Text_AlignRight_WithWrap(t *testing.T) {
	widget := Text{
		Content:   "This text will wrap",
		TextAlign: TextAlignRight,
		Wrap:      WrapSoft,
		Width:     Cells(12),
	}
	AssertSnapshot(t, widget, 12, 5,
		"Text wraps at word boundaries, each wrapped line is right-aligned within 12-cell width.")
}

func TestSnapshot_Text_AlignCenter_Spans(t *testing.T) {
	widget := Text{
		Spans: []Span{
			{Text: "Bold", Style: SpanStyle{Bold: true}},
			{Text: " text"},
		},
		TextAlign: TextAlignCenter,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Rich text 'Bold text' centered. 'Bold' in bold style, ' text' in normal style.")
}

func TestSnapshot_Text_AlignRight_Spans(t *testing.T) {
	widget := Text{
		Spans: []Span{
			{Text: "Right ", Style: SpanStyle{Italic: true}},
			{Text: "aligned"},
		},
		TextAlign: TextAlignRight,
		Width:     Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Rich text 'Right aligned' aligned to right edge. 'Right ' in italic, 'aligned' in normal.")
}

// =============================================================================
// Button Widget Tests
// =============================================================================

func TestSnapshot_Button_DefaultState(t *testing.T) {
	widget := Button{
		ID:    "btn1",
		Label: "Click Me",
	}
	AssertSnapshot(t, widget, 20, 3,
		"Button with 'Click Me' label. Default styling, width auto-sized to label.")
}

func TestSnapshot_Button_CustomStyle(t *testing.T) {
	widget := Button{
		ID:    "btn2",
		Label: "Styled",
		Style: Style{
			ForegroundColor: RGB(255, 255, 255),
			BackgroundColor: RGB(100, 50, 150),
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Button with 'Styled' label. White text on purple background (RGB 100,50,150).")
}

func TestSnapshot_Button_WithWidth(t *testing.T) {
	widget := Button{
		ID:    "btn3",
		Label: "Wide",
		Width: Cells(15),
	}
	AssertSnapshot(t, widget, 20, 3,
		"Button 'Wide' with fixed 15-cell width. Label centered within the button area.")
}

// =============================================================================
// List Widget Tests
// =============================================================================

func TestSnapshot_List_SingleSelect(t *testing.T) {
	state := NewListState([]string{"Item 1", "Item 2", "Item 3"})
	widget := List[string]{
		ID:    "list1",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 5,
		"List with 3 items vertically stacked. First item 'Item 1' is active (highlighted). Items 2 and 3 below.")
}

func TestSnapshot_List_ActiveItem(t *testing.T) {
	state := NewListState([]string{"First", "Second", "Third"})
	state.SelectIndex(1) // Select "Second"
	widget := List[string]{
		ID:    "list2",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 5,
		"List with 3 items. 'Second' (index 1) is active and highlighted. 'First' above, 'Third' below.")
}

func TestSnapshot_List_Empty(t *testing.T) {
	state := NewListState([]string{})
	widget := List[string]{
		ID:    "list3",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 5,
		"Empty list with no items. Should render as empty space with no visible content.")
}

func TestSnapshot_List_CustomRenderItem(t *testing.T) {
	state := NewListState([]string{"A", "B", "C"})
	widget := List[string]{
		ID:    "list4",
		State: state,
		RenderItem: func(item string, active bool, selected bool) Widget {
			prefix := "[ ] "
			if active {
				prefix = "[*] "
			}
			return Text{Content: prefix + item}
		},
	}
	AssertSnapshot(t, widget, 30, 5,
		"List with custom render showing checkboxes. '[*] A' active on row 1, '[ ] B' and '[ ] C' below.")
}

func TestSnapshot_List_MultiSelect(t *testing.T) {
	state := NewListState([]string{"Option 1", "Option 2", "Option 3"})
	state.Select(0)
	state.Select(2)
	widget := List[string]{
		ID:          "list5",
		State:       state,
		MultiSelect: true,
	}
	AssertSnapshot(t, widget, 30, 5,
		"Multi-select list with items 0 and 2 selected. 'Option 1' and 'Option 3' shown as selected, 'Option 2' unselected.")
}

// =============================================================================
// ProgressBar Widget Tests
// =============================================================================

func TestSnapshot_ProgressBar_ZeroProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.0,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"20-cell progress bar at 0%. Entire bar shows unfilled/empty state.")
}

func TestSnapshot_ProgressBar_HalfProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.5,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"20-cell progress bar at 50%. Left 10 cells filled, right 10 cells unfilled.")
}

func TestSnapshot_ProgressBar_FullProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 1.0,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"20-cell progress bar at 100%. Entire bar shows filled state.")
}

func TestSnapshot_ProgressBar_WithColors(t *testing.T) {
	widget := ProgressBar{
		Progress:      0.75,
		Width:         Cells(20),
		FilledColor:   RGB(0, 200, 100),
		UnfilledColor: RGB(50, 50, 50),
	}
	AssertSnapshot(t, widget, 20, 3,
		"20-cell progress bar at 75%. Green filled portion (15 cells), dark gray unfilled (5 cells).")
}

func TestSnapshot_ProgressBar_QuarterProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.25,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3,
		"20-cell progress bar at 25%. Left 5 cells filled, right 15 cells unfilled.")
}

// =============================================================================
// Spacer Widget Tests
// =============================================================================

func TestSnapshot_Spacer_FlexDefault(t *testing.T) {
	widget := Row{
		Width: Cells(30),
		Children: []Widget{
			Text{Content: "Left"},
			Spacer{},
			Text{Content: "Right"},
		},
	}
	AssertSnapshot(t, widget, 30, 3,
		"30-cell row with 'Left' at column 1 and 'Right' at far right. Spacer fills gap between them.")
}

func TestSnapshot_Spacer_FixedCells(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{Content: "A"},
			Spacer{Width: Cells(5)},
			Text{Content: "B"},
		},
	}
	AssertSnapshot(t, widget, 20, 3,
		"Row with 'A' at column 1, 5-cell fixed gap, then 'B' at column 7. Total width is 7 cells.")
}

func TestSnapshot_Spacer_InColumn(t *testing.T) {
	widget := Column{
		Height: Cells(10),
		Children: []Widget{
			Text{Content: "Top"},
			Spacer{},
			Text{Content: "Bottom"},
		},
	}
	AssertSnapshot(t, widget, 20, 10,
		"10-row column with 'Top' at row 1 and 'Bottom' at row 10. Spacer fills 8 rows between them.")
}

func TestSnapshot_Spacer_MultipleSpacers(t *testing.T) {
	widget := Row{
		Width: Cells(40),
		Children: []Widget{
			Text{Content: "A"},
			Spacer{},
			Text{Content: "B"},
			Spacer{},
			Text{Content: "C"},
		},
	}
	AssertSnapshot(t, widget, 40, 3,
		"40-cell row with 'A', 'B', 'C' evenly distributed. Two spacers split remaining space equally.")
}

// =============================================================================
// Conditional Widget Tests
// =============================================================================

func TestSnapshot_ShowWhen_True(t *testing.T) {
	widget := Column{
		Children: []Widget{
			ShowWhen(true, Text{Content: "Visible"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Column with 'Visible' on row 1 (shown because condition is true), 'Always' on row 2.")
}

func TestSnapshot_ShowWhen_False(t *testing.T) {
	widget := Column{
		Children: []Widget{
			ShowWhen(false, Text{Content: "Hidden"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Column with only 'Always' on row 1. 'Hidden' is removed (condition false), takes no space.")
}

func TestSnapshot_HideWhen_True(t *testing.T) {
	widget := Column{
		Children: []Widget{
			HideWhen(true, Text{Content: "Hidden"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Column with only 'Always' on row 1. 'Hidden' is removed (hide condition true), takes no space.")
}

func TestSnapshot_HideWhen_False(t *testing.T) {
	widget := Column{
		Children: []Widget{
			HideWhen(false, Text{Content: "Visible"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Column with 'Visible' on row 1 (shown because hide condition is false), 'Always' on row 2.")
}

// =============================================================================
// Switcher Widget Tests
// =============================================================================

func TestSnapshot_Switcher_ActiveChild(t *testing.T) {
	widget := Switcher{
		Active: "page1",
		Children: map[string]Widget{
			"page1": Text{Content: "Page One"},
			"page2": Text{Content: "Page Two"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Only 'Page One' visible (active key is 'page1'). 'Page Two' not rendered.")
}

func TestSnapshot_Switcher_DifferentActive(t *testing.T) {
	widget := Switcher{
		Active: "page2",
		Children: map[string]Widget{
			"page1": Text{Content: "First"},
			"page2": Text{Content: "Second"},
			"page3": Text{Content: "Third"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Only 'Second' visible (active key is 'page2'). 'First' and 'Third' not rendered.")
}

func TestSnapshot_Switcher_NoActiveMatch(t *testing.T) {
	widget := Switcher{
		Active: "nonexistent",
		Children: map[string]Widget{
			"page1": Text{Content: "Page One"},
		},
	}
	AssertSnapshot(t, widget, 20, 5,
		"Empty/no content visible. Active key 'nonexistent' doesn't match any child key.")
}
