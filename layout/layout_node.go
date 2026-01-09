package layout

// Constraints represents the min/max size constraints passed from parent to child.
// This allows expressing tight constraints (min == max), loose constraints (min = 0),
// and range constraints (min < max).
type Constraints struct {
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// --- Constraint constructors ---

// Tight creates constraints where the node must be exactly the given size.
// Both min and max are set to the same value.
func Tight(width, height int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose creates constraints where the node can be any size from 0 up to the given max.
func Loose(maxWidth, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: 0,
		MaxHeight: maxHeight,
	}
}

// TightWidth creates constraints with a fixed width but flexible height.
func TightWidth(width, maxHeight int) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: maxHeight,
	}
}

// TightHeight creates constraints with a fixed height but flexible width.
func TightHeight(maxWidth, height int) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  maxWidth,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Unbounded creates constraints with no limits.
// Useful for measuring intrinsic size.
func Unbounded() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  maxInt,
		MinHeight: 0,
		MaxHeight: maxInt,
	}
}

// maxInt is the maximum int value, used for unbounded constraints.
const maxInt = int(^uint(0) >> 1)

// --- Constraint methods ---

// IsTight returns true if both dimensions are tightly constrained (min == max).
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsTightWidth returns true if width is tightly constrained (min == max).
func (c Constraints) IsTightWidth() bool {
	return c.MinWidth == c.MaxWidth
}

// IsTightHeight returns true if height is tightly constrained (min == max).
func (c Constraints) IsTightHeight() bool {
	return c.MinHeight == c.MaxHeight
}

// Constrain clamps the given width and height to satisfy these constraints.
func (c Constraints) Constrain(width, height int) (int, int) {
	w := max(c.MinWidth, min(c.MaxWidth, width))
	h := max(c.MinHeight, min(c.MaxHeight, height))
	return w, h
}

// WithNodeConstraints applies a node's own min/max constraints to parent constraints.
// Node constraints use 0 to mean "unconstrained".
// Parent constraints are inviolable - node constraints are clamped to parent bounds:
//   - Node's min is clamped to parent's max (can't exceed available space)
//   - Node's max is raised to parent's min (can't refuse required minimum, e.g., stretch)
//
// If the resulting constraints are invalid (min > max), min wins. This handles
// user configuration errors like MinHeight: 60, MaxHeight: 40.
func (c Constraints) WithNodeConstraints(minW, maxW, minH, maxH int) Constraints {
	effective := c

	// Apply node's max constraints first - these represent explicit size limits
	// like Width: Cells(4) which should be respected even against parent's min.
	if maxW > 0 {
		effective.MaxWidth = min(effective.MaxWidth, maxW)
		// If parent's min exceeds node's explicit max, lower it.
		// This ensures explicit width constraints like Width: Cells(4) are respected
		// even when CrossAxisStretch wants to force a larger size.
		if c.MinWidth > maxW {
			effective.MinWidth = maxW
		}
	}
	if maxH > 0 {
		effective.MaxHeight = min(effective.MaxHeight, maxH)
		if c.MinHeight > maxH {
			effective.MinHeight = maxH
		}
	}

	// Apply node's min constraints, but clamp to parent's max.
	if minW > 0 {
		effective.MinWidth = max(effective.MinWidth, min(minW, c.MaxWidth))
	}
	if minH > 0 {
		effective.MinHeight = max(effective.MinHeight, min(minH, c.MaxHeight))
	}

	// Sanity check: ensure min <= max. If user misconfigured node constraints
	// (e.g., MinHeight > MaxHeight), min wins as it represents a hard requirement.
	if effective.MinWidth > effective.MaxWidth {
		effective.MaxWidth = effective.MinWidth
	}
	if effective.MinHeight > effective.MaxHeight {
		effective.MaxHeight = effective.MinHeight
	}

	return effective
}

// LayoutNode represents a node in the layout tree.
// It can be a leaf (BoxNode) or a container (RowNode, ColumnNode, etc.).
type LayoutNode interface {
	// ComputeLayout computes this node's size and positions all children.
	// Constraints specify the min/max bounds the node must fit within.
	// Returns a ComputedLayout containing the resulting BoxModel and positioned children.
	ComputeLayout(constraints Constraints) ComputedLayout
}

// ComputedLayout is the result of layout computation.
// It contains the computed BoxModel for a node and its positioned children (if any).
type ComputedLayout struct {
	// Box is the computed BoxModel for this node.
	Box BoxModel

	// Children contains the positioned child layouts.
	// nil for leaf nodes.
	Children []PositionedChild
}

// PositionedChild is a child with its computed position.
//
// Coordinate system: X and Y specify the child's border-box position
// relative to the parent's content-area origin (after padding and border).
//
// The renderer must add the parent's padding and border offsets when
// translating to screen coordinates. For example:
//
//	screenX := parentScreenX + parent.Box.Padding.Left + parent.Box.Border.Left + child.X
//	screenY := parentScreenY + parent.Box.Padding.Top + parent.Box.Border.Top + child.Y
//
// Child margins are already accounted for in X,Y: if a child has Margin.Left=5,
// its X is offset by 5 from where its margin-box starts. This means X,Y point
// directly to where the child's visible border-box begins.
type PositionedChild struct {
	// X, Y is the child's border-box position relative to parent's content-area.
	X, Y int

	// Layout is the child's computed layout (recursive).
	Layout ComputedLayout
}
