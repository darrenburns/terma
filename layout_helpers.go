package terma

import (
	"strings"

	"terma/layout"
)

// toLayoutEdgeInsets converts terma.EdgeInsets to layout.EdgeInsets.
func toLayoutEdgeInsets(e EdgeInsets) layout.EdgeInsets {
	return layout.EdgeInsets{
		Top:    e.Top,
		Right:  e.Right,
		Bottom: e.Bottom,
		Left:   e.Left,
	}
}

// borderToEdgeInsets converts a Border to layout.EdgeInsets based on border width.
func borderToEdgeInsets(b Border) layout.EdgeInsets {
	w := b.Width()
	return layout.EdgeInsetsAll(w)
}

// toLayoutWrapMode converts terma.WrapMode to layout.WrapMode.
func toLayoutWrapMode(w WrapMode) layout.WrapMode {
	switch w {
	case WrapNone:
		return layout.WrapNone
	case WrapHard:
		return layout.WrapChar
	default: // WrapSoft
		return layout.WrapWord
	}
}

// toLayoutMainAlign converts terma.MainAxisAlign to layout.MainAxisAlignment.
func toLayoutMainAlign(a MainAxisAlign) layout.MainAxisAlignment {
	switch a {
	case MainAxisCenter:
		return layout.MainAxisCenter
	case MainAxisEnd:
		return layout.MainAxisEnd
	default: // MainAxisStart
		return layout.MainAxisStart
	}
}

// toLayoutCrossAlign converts terma.CrossAxisAlign to layout.CrossAxisAlignment.
func toLayoutCrossAlign(a CrossAxisAlign) layout.CrossAxisAlignment {
	switch a {
	case CrossAxisStart:
		return layout.CrossAxisStart
	case CrossAxisCenter:
		return layout.CrossAxisCenter
	case CrossAxisEnd:
		return layout.CrossAxisEnd
	default: // CrossAxisStretch
		return layout.CrossAxisStretch
	}
}

// spansToPlainText extracts plain text content from a slice of Spans.
func spansToPlainText(spans []Span) string {
	var result strings.Builder
	for _, span := range spans {
		result.WriteString(span.Text)
	}
	return result.String()
}

// dimensionToMinMax converts a terma Dimension to min/max constraints.
// For Cells (fixed), both min and max are set to the value.
// For Auto or Fr, returns 0,0 (no constraints from dimension).
func dimensionToMinMax(d Dimension) (min, max int) {
	if d.IsCells() {
		v := d.CellsValue()
		return v, v
	}
	return 0, 0
}
