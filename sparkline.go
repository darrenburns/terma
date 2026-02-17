package terma

import (
	"math"
	"strings"

	"github.com/darrenburns/terma/layout"
)

var sparklineBars = []string{
	"▁",
	"▂",
	"▃",
	"▄",
	"▅",
	"▆",
	"▇",
	"█",
}

// Sparkline renders a compact inline chart using Unicode bar characters.
//
// Width defaults to the number of values, Height defaults to 1 cell.
// Set ColorByValue to vary bar colors by magnitude; ValueColorScale customizes the gradient.
type Sparkline struct {
	ID string // Optional unique identifier

	Values []float64 // Data points to render

	Width  Dimension // Deprecated: use Style.Width
	Height Dimension // Deprecated: use Style.Height

	Style Style // General styling (padding, margin, border)

	// Bars allows customizing the character set from low to high.
	// If empty or too short, the default sparkline bars are used.
	Bars []string

	// ColorByValue enables per-bar coloring based on normalized value.
	// If ValueColorScale is unset, a theme-based gradient is used.
	ColorByValue    bool
	ValueColorScale Gradient

	// Optional scaling overrides. When set, these bounds are used instead of auto min/max.
	MinValue *float64
	MaxValue *float64
}

// Build returns itself as Sparkline is a leaf widget.
func (s Sparkline) Build(ctx BuildContext) Widget {
	return s
}

// WidgetID returns the sparkline's unique identifier.
// Implements the Identifiable interface.
func (s Sparkline) WidgetID() string {
	return s.ID
}

// GetContentDimensions returns the width and height dimension preferences.
func (s Sparkline) GetContentDimensions() (width, height Dimension) {
	w, h := s.Style.GetDimensions().Width, s.Style.GetDimensions().Height
	if w.IsUnset() {
		w = s.Width
	}
	if h.IsUnset() {
		h = s.Height
	}
	if w.IsUnset() {
		w = Cells(len(s.Values))
	}
	if h.IsUnset() {
		h = Cells(1)
	}
	return w, h
}

// GetStyle returns the style of the sparkline.
func (s Sparkline) GetStyle() Style {
	return s.Style
}

// BuildLayoutNode builds a layout node for this Sparkline widget.
func (s Sparkline) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	padding := toLayoutEdgeInsets(s.Style.Padding)
	border := borderToEdgeInsets(s.Style.Border)
	dims := s.Style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = s.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = s.Height
	}
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	node := layout.LayoutNode(&layout.BoxNode{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
		Padding:   padding,
		Border:    border,
		Margin:    toLayoutEdgeInsets(s.Style.Margin),
		MeasureFunc: func(constraints layout.Constraints) (int, int) {
			size := s.Layout(ctx, Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight,
			})
			return size.Width, size.Height
		},
	})

	if hasPercentMinMax(dims) {
		node = &percentConstraintWrapper{
			child:     node,
			minWidth:  dims.MinWidth,
			maxWidth:  dims.MaxWidth,
			minHeight: dims.MinHeight,
			maxHeight: dims.MaxHeight,
			padding:   padding,
			border:    border,
		}
	}

	return node
}

// Layout computes the size of the sparkline.
func (s Sparkline) Layout(ctx BuildContext, constraints Constraints) Size {
	dims := s.Style.GetDimensions()
	widthDim := dims.Width
	heightDim := dims.Height
	if widthDim.IsUnset() {
		widthDim = s.Width
	}
	if heightDim.IsUnset() {
		heightDim = s.Height
	}
	var width int
	switch {
	case widthDim.IsCells():
		width = widthDim.CellsValue()
	case widthDim.IsFlex(), widthDim.IsPercent():
		width = constraints.MaxWidth
	default:
		width = len(s.Values)
	}

	var height int
	switch {
	case heightDim.IsCells():
		height = heightDim.CellsValue()
	case heightDim.IsFlex(), heightDim.IsPercent():
		height = constraints.MaxHeight
	default:
		height = 1
	}

	width = clampInt(width, constraints.MinWidth, constraints.MaxWidth)
	height = clampInt(height, constraints.MinHeight, constraints.MaxHeight)

	return Size{Width: width, Height: height}
}

// Render draws the sparkline to the render context.
func (s Sparkline) Render(ctx *RenderContext) {
	if ctx.Width <= 0 || ctx.Height <= 0 {
		return
	}

	values := sparklineResample(s.Values, ctx.Width)
	if len(values) == 0 {
		return
	}

	bars := s.Bars
	if len(bars) < 2 {
		bars = sparklineBars
	}

	minVal, maxVal := sparklineMinMax(values)
	if s.MinValue != nil {
		minVal = *s.MinValue
	}
	if s.MaxValue != nil {
		maxVal = *s.MaxValue
	}
	if maxVal < minVal {
		maxVal = minVal
	}

	theme := ctx.buildContext.Theme()
	baseStyle := s.Style
	if baseStyle.ForegroundColor == nil || !baseStyle.ForegroundColor.IsSet() {
		baseStyle.ForegroundColor = theme.Primary
	}

	useValueColor := s.ColorByValue || s.ValueColorScale.IsSet()
	colorScale := s.ValueColorScale
	if useValueColor && !colorScale.IsSet() {
		colorScale = NewGradient(theme.TextMuted, theme.Primary)
	}

	if !useValueColor {
		var sb strings.Builder
		sb.Grow(ctx.Width)
		for _, v := range values {
			norm := sparklineNormalize(v, minVal, maxVal)
			sb.WriteString(bars[sparklineBarIndex(norm, len(bars))])
		}
		line := sb.String()
		for row := 0; row < ctx.Height; row++ {
			ctx.DrawStyledText(0, row, line, baseStyle)
		}
		return
	}

	for row := 0; row < ctx.Height; row++ {
		for i, v := range values {
			norm := sparklineNormalize(v, minVal, maxVal)
			barStyle := baseStyle
			barStyle.ForegroundColor = colorScale.At(norm)
			ctx.DrawStyledText(i, row, bars[sparklineBarIndex(norm, len(bars))], barStyle)
		}
	}
}

func sparklineNormalize(value, minVal, maxVal float64) float64 {
	if maxVal == minVal {
		return 0
	}
	norm := (value - minVal) / (maxVal - minVal)
	if norm < 0 {
		return 0
	}
	if norm > 1 {
		return 1
	}
	return norm
}

func sparklineBarIndex(norm float64, barCount int) int {
	if barCount <= 1 {
		return 0
	}
	idx := int(math.Round(norm * float64(barCount-1)))
	if idx < 0 {
		return 0
	}
	if idx >= barCount {
		return barCount - 1
	}
	return idx
}

func sparklineMinMax(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}
	minVal, maxVal := values[0], values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	return minVal, maxVal
}

func sparklineResample(values []float64, width int) []float64 {
	if width <= 0 || len(values) == 0 {
		return nil
	}
	if len(values) == width {
		return append([]float64(nil), values...)
	}
	if width == 1 {
		return []float64{values[len(values)-1]}
	}
	if len(values) > width {
		return sparklineDownsample(values, width)
	}
	return sparklineUpsample(values, width)
}

func sparklineDownsample(values []float64, width int) []float64 {
	result := make([]float64, width)
	step := float64(len(values)) / float64(width)
	for i := 0; i < width; i++ {
		start := int(math.Floor(float64(i) * step))
		end := int(math.Floor(float64(i+1) * step))
		if end <= start {
			end = start + 1
		}
		if end > len(values) {
			end = len(values)
		}
		sum := 0.0
		for j := start; j < end; j++ {
			sum += values[j]
		}
		result[i] = sum / float64(end-start)
	}
	return result
}

func sparklineUpsample(values []float64, width int) []float64 {
	result := make([]float64, width)
	if len(values) == 1 {
		for i := range result {
			result[i] = values[0]
		}
		return result
	}
	step := float64(len(values)-1) / float64(width-1)
	for i := 0; i < width; i++ {
		pos := float64(i) * step
		lower := int(math.Floor(pos))
		upper := int(math.Ceil(pos))
		if upper >= len(values) {
			upper = len(values) - 1
		}
		if lower == upper {
			result[i] = values[lower]
			continue
		}
		t := pos - float64(lower)
		result[i] = values[lower] + (values[upper]-values[lower])*t
	}
	return result
}
