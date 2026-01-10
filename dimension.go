package terma

// dimensionUnit represents the type of dimension measurement.
type dimensionUnit int

const (
	unitAuto dimensionUnit = iota
	unitCells
	unitFlex
)

// Dimension represents a size specification for widgets.
// The zero value represents auto-sizing (fit content).
type Dimension struct {
	value float64
	unit  dimensionUnit
}

// Auto represents an auto-sizing dimension that fits content.
// Note: Auto has value=1 to distinguish it from the zero value (unset).
var Auto = Dimension{value: 1, unit: unitAuto}

// Cells returns a fixed dimension measured in terminal cells.
func Cells(n int) Dimension {
	return Dimension{value: float64(n), unit: unitCells}
}

// Flex returns a flexible dimension for proportional space distribution.
// Children with Flex dimensions share remaining space proportionally.
// For example, Flex(1) and Flex(2) siblings get 1/3 and 2/3 of remaining space.
func Flex(n float64) Dimension {
	return Dimension{value: n, unit: unitFlex}
}

// IsAuto returns true if this is an auto-sizing dimension.
func (d Dimension) IsAuto() bool {
	return d.unit == unitAuto
}

// IsCells returns true if this is a fixed cell dimension.
func (d Dimension) IsCells() bool {
	return d.unit == unitCells
}

// IsFlex returns true if this is a flexible dimension.
func (d Dimension) IsFlex() bool {
	return d.unit == unitFlex
}

// CellsValue returns the fixed cell count (only valid if IsCells() is true).
func (d Dimension) CellsValue() int {
	return int(d.value)
}

// FlexValue returns the flex value (only valid if IsFlex() is true).
func (d Dimension) FlexValue() float64 {
	return d.value
}

// IsUnset returns true if this dimension was not explicitly set (zero value).
func (d Dimension) IsUnset() bool {
	return d == Dimension{}
}
