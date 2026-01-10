package terma

import "testing"

func TestSpacer_GetDimensions_DefaultsToFlex1(t *testing.T) {
	s := Spacer{}
	w, h := s.GetDimensions()

	if !w.IsFlex() || w.FlexValue() != 1 {
		t.Errorf("expected Width to default to Flex(1), got %v", w)
	}
	if !h.IsFlex() || h.FlexValue() != 1 {
		t.Errorf("expected Height to default to Flex(1), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitWidth(t *testing.T) {
	s := Spacer{Width: Cells(10)}
	w, h := s.GetDimensions()

	if !w.IsCells() || w.CellsValue() != 10 {
		t.Errorf("expected Width to be Cells(10), got %v", w)
	}
	if !h.IsFlex() || h.FlexValue() != 1 {
		t.Errorf("expected Height to default to Flex(1), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitHeight(t *testing.T) {
	s := Spacer{Height: Cells(5)}
	w, h := s.GetDimensions()

	if !w.IsFlex() || w.FlexValue() != 1 {
		t.Errorf("expected Width to default to Flex(1), got %v", w)
	}
	if !h.IsCells() || h.CellsValue() != 5 {
		t.Errorf("expected Height to be Cells(5), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsExplicitBoth(t *testing.T) {
	s := Spacer{Width: Cells(10), Height: Cells(5)}
	w, h := s.GetDimensions()

	if !w.IsCells() || w.CellsValue() != 10 {
		t.Errorf("expected Width to be Cells(10), got %v", w)
	}
	if !h.IsCells() || h.CellsValue() != 5 {
		t.Errorf("expected Height to be Cells(5), got %v", h)
	}
}

func TestSpacer_GetDimensions_RespectsFlexValues(t *testing.T) {
	s := Spacer{Width: Flex(2), Height: Flex(3)}
	w, h := s.GetDimensions()

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
	w, h := s.GetDimensions()

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
	w, h := s.GetDimensions()

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
	w, h := s.GetDimensions()

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
