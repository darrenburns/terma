package terma

import "testing"

func TestCells_ReturnsFixedDimension(t *testing.T) {
	d := Cells(10)

	if !d.IsCells() {
		t.Error("expected IsCells() to be true")
	}
	if d.CellsValue() != 10 {
		t.Errorf("expected CellsValue() = 10, got %d", d.CellsValue())
	}
}

func TestCells_ZeroValue(t *testing.T) {
	d := Cells(0)

	if !d.IsCells() {
		t.Error("expected IsCells() to be true for Cells(0)")
	}
	if d.CellsValue() != 0 {
		t.Errorf("expected CellsValue() = 0, got %d", d.CellsValue())
	}
}

func TestCells_LargeValue(t *testing.T) {
	d := Cells(10000)

	if !d.IsCells() {
		t.Error("expected IsCells() to be true")
	}
	if d.CellsValue() != 10000 {
		t.Errorf("expected CellsValue() = 10000, got %d", d.CellsValue())
	}
}

func TestFlex_ReturnsFlexibleDimension(t *testing.T) {
	d := Flex(1)

	if !d.IsFlex() {
		t.Error("expected IsFlex() to be true")
	}
	if d.FlexValue() != 1 {
		t.Errorf("expected FlexValue() = 1, got %f", d.FlexValue())
	}
}

func TestFlex_ZeroValue(t *testing.T) {
	d := Flex(0)

	if !d.IsFlex() {
		t.Error("expected IsFlex() to be true for Flex(0)")
	}
	if d.FlexValue() != 0 {
		t.Errorf("expected FlexValue() = 0, got %f", d.FlexValue())
	}
}

func TestFlex_FractionalValue(t *testing.T) {
	d := Flex(2.5)

	if !d.IsFlex() {
		t.Error("expected IsFlex() to be true")
	}
	if d.FlexValue() != 2.5 {
		t.Errorf("expected FlexValue() = 2.5, got %f", d.FlexValue())
	}
}

func TestFlex_LargeValue(t *testing.T) {
	d := Flex(100)

	if !d.IsFlex() {
		t.Error("expected IsFlex() to be true")
	}
	if d.FlexValue() != 100 {
		t.Errorf("expected FlexValue() = 100, got %f", d.FlexValue())
	}
}

func TestAuto_IsAutoTrue(t *testing.T) {
	if !Auto.IsAuto() {
		t.Error("expected Auto.IsAuto() to be true")
	}
}

func TestAuto_IsNotCells(t *testing.T) {
	if Auto.IsCells() {
		t.Error("expected Auto.IsCells() to be false")
	}
}

func TestAuto_IsNotFlex(t *testing.T) {
	if Auto.IsFlex() {
		t.Error("expected Auto.IsFlex() to be false")
	}
}

func TestCells_IsAutoFalse(t *testing.T) {
	d := Cells(10)

	if d.IsAuto() {
		t.Error("expected Cells(10).IsAuto() to be false")
	}
}

func TestCells_IsFlexFalse(t *testing.T) {
	d := Cells(10)

	if d.IsFlex() {
		t.Error("expected Cells(10).IsFlex() to be false")
	}
}

func TestFlex_IsAutoFalse(t *testing.T) {
	d := Flex(1)

	if d.IsAuto() {
		t.Error("expected Flex(1).IsAuto() to be false")
	}
}

func TestFlex_IsCellsFalse(t *testing.T) {
	d := Flex(1)

	if d.IsCells() {
		t.Error("expected Flex(1).IsCells() to be false")
	}
}

func TestDimension_ZeroValue_IsUnset(t *testing.T) {
	var d Dimension

	if !d.IsUnset() {
		t.Error("expected zero value Dimension to be unset")
	}
}

func TestDimension_ZeroValue_IsAuto(t *testing.T) {
	// The zero value has unit=unitAuto (0), so it should be considered Auto
	var d Dimension

	if !d.IsAuto() {
		t.Error("expected zero value Dimension to be Auto")
	}
}

func TestAuto_IsNotUnset(t *testing.T) {
	// Auto has value=1, so it's not the zero value
	if Auto.IsUnset() {
		t.Error("expected Auto to not be unset")
	}
}

func TestCells_IsNotUnset(t *testing.T) {
	d := Cells(5)

	if d.IsUnset() {
		t.Error("expected Cells(5) to not be unset")
	}
}

func TestFlex_IsNotUnset(t *testing.T) {
	d := Flex(1)

	if d.IsUnset() {
		t.Error("expected Flex(1) to not be unset")
	}
}

// Table-driven test for dimension type exclusivity
func TestDimension_TypeExclusivity(t *testing.T) {
	tests := []struct {
		name    string
		dim     Dimension
		isAuto  bool
		isCells bool
		isFlex  bool
	}{
		{"Auto", Auto, true, false, false},
		{"Cells(0)", Cells(0), false, true, false},
		{"Cells(10)", Cells(10), false, true, false},
		{"Flex(0)", Flex(0), false, false, true},
		{"Flex(1)", Flex(1), false, false, true},
		{"Flex(2.5)", Flex(2.5), false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dim.IsAuto() != tt.isAuto {
				t.Errorf("IsAuto() = %v, want %v", tt.dim.IsAuto(), tt.isAuto)
			}
			if tt.dim.IsCells() != tt.isCells {
				t.Errorf("IsCells() = %v, want %v", tt.dim.IsCells(), tt.isCells)
			}
			if tt.dim.IsFlex() != tt.isFlex {
				t.Errorf("IsFlex() = %v, want %v", tt.dim.IsFlex(), tt.isFlex)
			}
		})
	}
}
