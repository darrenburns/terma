package terma

import (
	"testing"
)

// =============================================================================
// Text Widget Tests
// =============================================================================

func TestSnapshot_Text_PlainContent(t *testing.T) {
	widget := Text{Content: "Hello, World!!"}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Text_RichSpans(t *testing.T) {
	widget := Text{
		Spans: []Span{
			{Text: "Bold", Style: SpanStyle{Bold: true}},
			{Text: " and "},
			{Text: "Italic", Style: SpanStyle{Italic: true}},
		},
	}
	AssertSnapshot(t, widget, 30, 3)
}

func TestSnapshot_Text_WrapNone(t *testing.T) {
	widget := Text{
		Content: "This is a very long line that should not wrap",
		Wrap:    WrapNone,
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Text_WrapSoft(t *testing.T) {
	widget := Text{
		Content: "This is a line that should wrap at word boundaries",
		Wrap:    WrapSoft,
		Width:   Cells(15),
	}
	AssertSnapshot(t, widget, 15, 5)
}

func TestSnapshot_Text_WrapHard(t *testing.T) {
	widget := Text{
		Content: "Supercalifragilisticexpialidocious",
		Wrap:    WrapHard,
		Width:   Cells(10),
	}
	AssertSnapshot(t, widget, 10, 5)
}

func TestSnapshot_Text_BoldItalicUnderline(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Text{Content: "Bold", Style: Style{Bold: true}},
			Text{Content: "Italic", Style: Style{Italic: true}},
			Text{Content: "Underline", Style: Style{Underline: UnderlineSingle}},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Text_WithBackground(t *testing.T) {
	widget := Text{
		Content: "Highlighted",
		Style: Style{
			BackgroundColor: RGB(100, 100, 200),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Text_Multiline(t *testing.T) {
	widget := Text{Content: "Line 1\nLine 2\nLine 3"}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Text_WithForegroundColor(t *testing.T) {
	widget := Text{
		Content: "Colored",
		Style: Style{
			ForegroundColor: RGB(255, 100, 100),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

// =============================================================================
// Button Widget Tests
// =============================================================================

func TestSnapshot_Button_DefaultState(t *testing.T) {
	widget := &Button{
		ID:    "btn1",
		Label: "Click Me",
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Button_CustomStyle(t *testing.T) {
	widget := &Button{
		ID:    "btn2",
		Label: "Styled",
		Style: Style{
			ForegroundColor: RGB(255, 255, 255),
			BackgroundColor: RGB(100, 50, 150),
		},
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_Button_WithWidth(t *testing.T) {
	widget := &Button{
		ID:    "btn3",
		Label: "Wide",
		Width: Cells(15),
	}
	AssertSnapshot(t, widget, 20, 3)
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
	AssertSnapshot(t, widget, 30, 5)
}

func TestSnapshot_List_ActiveItem(t *testing.T) {
	state := NewListState([]string{"First", "Second", "Third"})
	state.SelectIndex(1) // Select "Second"
	widget := List[string]{
		ID:    "list2",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 5)
}

func TestSnapshot_List_Empty(t *testing.T) {
	state := NewListState([]string{})
	widget := List[string]{
		ID:    "list3",
		State: state,
	}
	AssertSnapshot(t, widget, 30, 5)
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
	AssertSnapshot(t, widget, 30, 5)
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
	AssertSnapshot(t, widget, 30, 5)
}

// =============================================================================
// ProgressBar Widget Tests
// =============================================================================

func TestSnapshot_ProgressBar_ZeroProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.0,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_ProgressBar_HalfProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.5,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_ProgressBar_FullProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 1.0,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_ProgressBar_WithColors(t *testing.T) {
	widget := ProgressBar{
		Progress:      0.75,
		Width:         Cells(20),
		FilledColor:   RGB(0, 200, 100),
		UnfilledColor: RGB(50, 50, 50),
	}
	AssertSnapshot(t, widget, 20, 3)
}

func TestSnapshot_ProgressBar_QuarterProgress(t *testing.T) {
	widget := ProgressBar{
		Progress: 0.25,
		Width:    Cells(20),
	}
	AssertSnapshot(t, widget, 20, 3)
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
	AssertSnapshot(t, widget, 30, 3)
}

func TestSnapshot_Spacer_FixedCells(t *testing.T) {
	widget := Row{
		Children: []Widget{
			Text{Content: "A"},
			Spacer{Width: Cells(5)},
			Text{Content: "B"},
		},
	}
	AssertSnapshot(t, widget, 20, 3)
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
	AssertSnapshot(t, widget, 20, 10)
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
	AssertSnapshot(t, widget, 40, 3)
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
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_ShowWhen_False(t *testing.T) {
	widget := Column{
		Children: []Widget{
			ShowWhen(false, Text{Content: "Hidden"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_HideWhen_True(t *testing.T) {
	widget := Column{
		Children: []Widget{
			HideWhen(true, Text{Content: "Hidden"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_HideWhen_False(t *testing.T) {
	widget := Column{
		Children: []Widget{
			HideWhen(false, Text{Content: "Visible"}),
			Text{Content: "Always"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 5)
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
	AssertSnapshot(t, widget, 20, 5)
}

func TestSnapshot_Switcher_NoActiveMatch(t *testing.T) {
	widget := Switcher{
		Active: "nonexistent",
		Children: map[string]Widget{
			"page1": Text{Content: "Page One"},
		},
	}
	AssertSnapshot(t, widget, 20, 5)
}
