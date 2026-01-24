package terma

// dimensionUnit represents the type of dimension measurement.
type dimensionUnit int

const (
	unitAuto dimensionUnit = iota
	unitCells
	unitFlex
	unitPercent
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

// Percent returns a dimension as a percentage of the parent's available space.
// For example, Percent(50) means 50% of the parent's width or height.
// Unlike Flex, percentages are calculated from the total available space,
// not from the remaining space after fixed children.
func Percent(n float64) Dimension {
	return Dimension{value: n, unit: unitPercent}
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

// IsPercent returns true if this is a percentage dimension.
func (d Dimension) IsPercent() bool {
	return d.unit == unitPercent
}

// CellsValue returns the fixed cell count (only valid if IsCells() is true).
func (d Dimension) CellsValue() int {
	return int(d.value)
}

// FlexValue returns the flex value (only valid if IsFlex() is true).
func (d Dimension) FlexValue() float64 {
	return d.value
}

// PercentValue returns the percentage value (only valid if IsPercent() is true).
func (d Dimension) PercentValue() float64 {
	return d.value
}

// IsUnset returns true if this dimension was not explicitly set (zero value).
func (d Dimension) IsUnset() bool {
	return d == Dimension{}
}

// DimensionSet groups size preferences and constraints for a widget.
// Width/Height describe the preferred content-box size.
// Min/Max fields constrain the content-box size range.
type DimensionSet struct {
	Width, Height           Dimension
	MinWidth, MinHeight     Dimension
	MaxWidth, MaxHeight     Dimension
}

// WithDefaults applies default width/height if unset.
func (d DimensionSet) WithDefaults(width, height Dimension) DimensionSet {
	if d.Width.IsUnset() {
		d.Width = width
	}
	if d.Height.IsUnset() {
		d.Height = height
	}
	return d
}
