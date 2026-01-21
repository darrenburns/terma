package terma

// MinMaxDimensions holds optional min/max size preferences for a widget.
// Embed this in widget structs to satisfy MinMaxDimensioned without boilerplate.
type MinMaxDimensions struct {
	MinWidth  Dimension
	MaxWidth  Dimension
	MinHeight Dimension
	MaxHeight Dimension
}

// GetMinMaxDimensions returns the min/max dimension preferences.
func (m MinMaxDimensions) GetMinMaxDimensions() (minWidth, maxWidth, minHeight, maxHeight Dimension) {
	return m.MinWidth, m.MaxWidth, m.MinHeight, m.MaxHeight
}
