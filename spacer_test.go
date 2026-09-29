package terma

import "testing"

func TestSpacer_GetDimensions_DefaultsToFlex1(t *testing.T) {
	s := Spacer{}
	w, h := s.GetContentDimensions()

	if !w.IsFlex() || w.FlexValue() != 1 {
		t.Errorf("expected Width to default to Flex(1), got %v", w)
	}
	if !h.IsFlex() || h.FlexValue() != 1 {
		t.Errorf("expected Height to default to Flex(1), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitWidth(t *testing.T) {
	s := Spacer{Width: Cells(10)}
	w, h := s.GetContentDimensions()

	if !w.IsCells() || w.CellsValue() != 10 {
		t.Errorf("expected Width to be Cells(10), got %v", w)
	}
	// When only Width is set, Height defaults to Auto (not Flex(1))
	// This makes Spacer{Width: Flex(1)} behave as a horizontal-only spacer
	if !h.IsAuto() {
		t.Errorf("expected Height to default to Auto when Width is set, got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitHeight(t *testing.T) {
	s := Spacer{Height: Cells(5)}
	w, h := s.GetContentDimensions()

	// When only Height is set, Width defaults to Auto (not Flex(1))
	// This makes Spacer{Height: Flex(1)} behave as a vertical-only spacer
	if !w.IsAuto() {
		t.Errorf("expected Width to default to Auto when Height is set, got %v", w)
	}
	if !h.IsCells() || h.CellsValue() != 5 {
		t.Errorf("expected Height to be Cells(5), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitBoth(t *testing.T) {
	s := Spacer{Width: Cells(10), Height: Cells(5)}
	w, h := s.GetContentDimensions()

	if !w.IsCells() || w.CellsValue() != 10 {
		t.Errorf("expected Width to be Cells(10), got %v", w)
	}
	if !h.IsCells() || h.CellsValue() != 5 {
		t.Errorf("expected Height to be Cells(5), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsFlexValues(t *testing.T) {
	s := Spacer{Width: Flex(2), Height: Flex(3)}
	w, h := s.GetContentDimensions()

	if !w.IsFlex() || w.FlexValue() != 2 {
		t.Errorf("expected Width to be Flex(2), got %v", w)
	}
	if !h.IsFlex() || h.FlexValue() != 3 {
		t.Errorf("expected Height to be Flex(3), got %v", h)
	}
}

func TestSpacer_Build_ReturnsSelf(t *testing.T) {
	s := Spacer{Width: Cells(10)}
	result := s.Build(BuildContext{})

	if result != s {
		t.Error("expected Build() to return self")
	}
}

func TestSpacer_BuildLayoutNode_ReturnsBoxNode(t *testing.T) {
	s := Spacer{Width: Cells(10), Height: Cells(5)}
	node := s.BuildLayoutNode(BuildContext{})

	if node == nil {
		t.Fatal("expected BuildLayoutNode to return non-nil node")
	}
}

func TestSpacer_GetDimensions_ExplicitAutoPassesThrough(t *testing.T) {
	// Explicitly setting Auto should NOT default to Flex(1).
	// Note: Auto on a Spacer means "fit content" = 0 size (no content).
	s := Spacer{Width: Auto, Height: Auto}
	w, h := s.GetContentDimensions()

	if !w.IsAuto() {
		t.Errorf("expected Width to be Auto when explicitly set, got %v", w)
	}
	if !h.IsAuto() {
		t.Errorf("expected Height to be Auto when explicitly set, got %v", h)
	}
}

func TestSpacer_GetDimensions_FlexZeroPassesThrough(t *testing.T) {
	// Flex(0) means "take no share of remaining space" = 0 size
	s := Spacer{Width: Flex(0), Height: Flex(0)}
	w, h := s.GetContentDimensions()

	if !w.IsFlex() || w.FlexValue() != 0 {
		t.Errorf("expected Width to be Flex(0), got %v", w)
	}
	if !h.IsFlex() || h.FlexValue() != 0 {
		t.Errorf("expected Height to be Flex(0), got %v", h)
	}
}

func TestSpacer_GetDimensions_CellsZeroPassesThrough(t *testing.T) {
	// Cells(0) means explicit 0 fixed size
	s := Spacer{Width: Cells(0), Height: Cells(0)}
	w, h := s.GetContentDimensions()

	if !w.IsCells() || w.CellsValue() != 0 {
		t.Errorf("expected Width to be Cells(0), got %v", w)
	}
	if !h.IsCells() || h.CellsValue() != 0 {
		t.Errorf("expected Height to be Cells(0), got %v", h)
	}
}

func TestSpacer_Render_IsNoOp(t *testing.T) {
	// Render should not panic and should do nothing
	s := Spacer{}
	s.Render(nil) // Should not panic even with nil context
}

// A bare Spacer pushing buttons apart in a Row must not make the Row claim
// the height a Flex(1) sibling needs.
func TestSnapshot_Spacer_InRowFlexesHorizontallyOnly(t *testing.T) {
	widget := Column{
		Style: Style{Width: Cells(30), Height: Cells(6)},
		Children: []Widget{
			TextArea{ID: "body", State: NewTextAreaState("line one\nline two"), Style: Style{Height: Flex(1)}},
			Row{Children: []Widget{
				Text{Content: "[x] wrap"},
				Spacer{},
				Text{Content: "Cancel"},
				Text{Content: " Save"},
			}},
		},
	}
	AssertSnapshot(t, widget, 30, 6,
		"The text area fills the top 5 rows; the button row is 1 row tall at the bottom, '[x] wrap' left and 'Cancel Save' pushed right.")
}

// A bare Spacer in an auto-width Column must not make the Column as wide as
// it can be.
func TestSnapshot_Spacer_InColumnFlexesVerticallyOnly(t *testing.T) {
	widget := Row{
		Style: Style{Width: Cells(20), Height: Cells(4)},
		Children: []Widget{
			Column{
				Style:    Style{BackgroundColor: Hex("#444444")},
				Children: []Widget{Text{Content: "top"}, Spacer{}, Text{Content: "end"}},
			},
			Text{Content: "|side"},
		},
	}
	AssertSnapshot(t, widget, 20, 4,
		"A 3-wide grey column ('top' at the top, 'end' at the bottom) with '|side' right beside it.")
}
