package terma

// HorizontalAlignment specifies horizontal positioning within available space.
type HorizontalAlignment int

const (
	// HAlignStart aligns content at the start (left) of the available space.
	HAlignStart HorizontalAlignment = iota
	// HAlignCenter centers content horizontally.
	HAlignCenter
	// HAlignEnd aligns content at the end (right) of the available space.
	HAlignEnd
)

// VerticalAlignment specifies vertical positioning within available space.
type VerticalAlignment int

const (
	// VAlignTop aligns content at the top of the available space.
	VAlignTop VerticalAlignment = iota
	// VAlignCenter centers content vertically.
	VAlignCenter
	// VAlignBottom aligns content at the bottom of the available space.
	VAlignBottom
)

// Alignment combines horizontal and vertical alignment for 2D positioning.
type Alignment struct {
	Horizontal HorizontalAlignment
	Vertical   VerticalAlignment
}

// Predefined alignments for common positioning patterns.
var (
	AlignTopStart    = Alignment{HAlignStart, VAlignTop}
	AlignTopCenter   = Alignment{HAlignCenter, VAlignTop}
	AlignTopEnd      = Alignment{HAlignEnd, VAlignTop}
	AlignCenterStart = Alignment{HAlignStart, VAlignCenter}
	AlignCenter      = Alignment{HAlignCenter, VAlignCenter}
	AlignCenterEnd   = Alignment{HAlignEnd, VAlignCenter}
	AlignBottomStart = Alignment{HAlignStart, VAlignBottom}
	AlignBottomCenter= Alignment{HAlignCenter, VAlignBottom}
	AlignBottomEnd   = Alignment{HAlignEnd, VAlignBottom}
)
