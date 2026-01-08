package layout

// Rect represents a rectangle with position and size.
type Rect struct {
	X, Y          int
	Width, Height int
}

// Contains returns true if the point (x, y) is within this rectangle.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width &&
		y >= r.Y && y < r.Y+r.Height
}

// EdgeInsets represents spacing around the four edges of a box.
type EdgeInsets struct {
	Top, Right, Bottom, Left int
}

// EdgeInsetsAll creates EdgeInsets with the same value for all sides.
func EdgeInsetsAll(value int) EdgeInsets {
	return EdgeInsets{Top: value, Right: value, Bottom: value, Left: value}
}

// EdgeInsetsXY creates EdgeInsets with separate horizontal and vertical values.
func EdgeInsetsXY(horizontal, vertical int) EdgeInsets {
	return EdgeInsets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// EdgeInsetsTRBL creates EdgeInsets with individual values for each side.
func EdgeInsetsTRBL(top, right, bottom, left int) EdgeInsets {
	return EdgeInsets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Horizontal returns the total horizontal inset (Left + Right).
func (e EdgeInsets) Horizontal() int {
	return e.Left + e.Right
}

// Vertical returns the total vertical inset (Top + Bottom).
func (e EdgeInsets) Vertical() int {
	return e.Top + e.Bottom
}
