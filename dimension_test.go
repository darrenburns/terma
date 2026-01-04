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

func TestFr_ReturnsFractionalDimension(t *testing.T) {
	d := Fr(1)

	if !d.IsFr() {
		t.Error("expected IsFr() to be true")
	}
	if d.FrValue() != 1 {
		t.Errorf("expected FrValue() = 1, got %f", d.FrValue())
	}
}

func TestFr_ZeroValue(t *testing.T) {
	d := Fr(0)

	if !d.IsFr() {
		t.Error("expected IsFr() to be true for Fr(0)")
	}
	if d.FrValue() != 0 {
		t.Errorf("expected FrValue() = 0, got %f", d.FrValue())
	}
}

func TestFr_FractionalValue(t *testing.T) {
	d := Fr(2.5)

	if !d.IsFr() {
		t.Error("expected IsFr() to be true")
	}
	if d.FrValue() != 2.5 {
		t.Errorf("expected FrValue() = 2.5, got %f", d.FrValue())
	}
}

func TestFr_LargeValue(t *testing.T) {
	d := Fr(100)

	if !d.IsFr() {
		t.Error("expected IsFr() to be true")
	}
	if d.FrValue() != 100 {
		t.Errorf("expected FrValue() = 100, got %f", d.FrValue())
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

func TestAuto_IsNotFr(t *testing.T) {
	if Auto.IsFr() {
		t.Error("expected Auto.IsFr() to be false")
	}
}

func TestCells_IsAutoFalse(t *testing.T) {
	d := Cells(10)

	if d.IsAuto() {
		t.Error("expected Cells(10).IsAuto() to be false")
	}
}

func TestCells_IsFrFalse(t *testing.T) {
	d := Cells(10)

	if d.IsFr() {
		t.Error("expected Cells(10).IsFr() to be false")
	}
}

func TestFr_IsAutoFalse(t *testing.T) {
	d := Fr(1)

	if d.IsAuto() {
		t.Error("expected Fr(1).IsAuto() to be false")
	}
}

func TestFr_IsCellsFalse(t *testing.T) {
	d := Fr(1)

	if d.IsCells() {
		t.Error("expected Fr(1).IsCells() to be false")
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

func TestFr_IsNotUnset(t *testing.T) {
	d := Fr(1)

	if d.IsUnset() {
		t.Error("expected Fr(1) to not be unset")
	}
}

// Table-driven test for dimension type exclusivity
func TestDimension_TypeExclusivity(t *testing.T) {
	tests := []struct {
		name    string
		dim     Dimension
		isAuto  bool
		isCells bool
		isFr    bool
	}{
		{"Auto", Auto, true, false, false},
		{"Cells(0)", Cells(0), false, true, false},
		{"Cells(10)", Cells(10), false, true, false},
		{"Fr(0)", Fr(0), false, false, true},
		{"Fr(1)", Fr(1), false, false, true},
		{"Fr(2.5)", Fr(2.5), false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.dim.IsAuto() != tt.isAuto {
				t.Errorf("IsAuto() = %v, want %v", tt.dim.IsAuto(), tt.isAuto)
			}
			if tt.dim.IsCells() != tt.isCells {
				t.Errorf("IsCells() = %v, want %v", tt.dim.IsCells(), tt.isCells)
			}
			if tt.dim.IsFr() != tt.isFr {
				t.Errorf("IsFr() = %v, want %v", tt.dim.IsFr(), tt.isFr)
			}
		})
	}
}
