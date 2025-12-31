package terma

// dimensionUnit represents the type of dimension measurement.
type dimensionUnit int

const (
	unitAuto dimensionUnit = iota
	unitCells
	unitFr
)

// Dimension represents a size specification for widgets.
// The zero value represents auto-sizing (fit content).
type Dimension struct {
	value float64
	unit  dimensionUnit
}

// Auto represents an auto-sizing dimension that fits content.
var Auto = Dimension{unit: unitAuto}

// Cells returns a fixed dimension measured in terminal cells.
func Cells(n int) Dimension {
	return Dimension{value: float64(n), unit: unitCells}
}

// Fr returns a fractional dimension for proportional space distribution.
// Children with Fr dimensions share remaining space proportionally.
// For example, Fr(1) and Fr(2) siblings get 1/3 and 2/3 of remaining space.
func Fr(n float64) Dimension {
	return Dimension{value: n, unit: unitFr}
}

// IsAuto returns true if this is an auto-sizing dimension.
func (d Dimension) IsAuto() bool {
	return d.unit == unitAuto
}

// IsCells returns true if this is a fixed cell dimension.
func (d Dimension) IsCells() bool {
	return d.unit == unitCells
}

// IsFr returns true if this is a fractional dimension.
func (d Dimension) IsFr() bool {
	return d.unit == unitFr
}

// CellsValue returns the fixed cell count (only valid if IsCells() is true).
func (d Dimension) CellsValue() int {
	return int(d.value)
}

// FrValue returns the fractional value (only valid if IsFr() is true).
func (d Dimension) FrValue() float64 {
	return d.value
}
