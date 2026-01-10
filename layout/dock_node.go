package layout

// DockEdge specifies which edge a child is docked to.
type DockEdge int

const (
	DockTop DockEdge = iota
	DockBottom
	DockLeft
	DockRight
)

// DockNode lays out children by docking them to edges (WPF DockPanel style).
// Each docked child consumes space from the remaining area in order.
// The body fills whatever space remains after all edges are processed.
//
// Processing order is determined by DockOrder. If not specified, the default
// order is: Top, Bottom, Left, Right. This matches WPF's LastChildFill behavior
// when Body is set.
//
// Example: With Top=[header], Bottom=[footer], Body=content:
//   - header takes full width at top, consumes its height
//   - footer takes full width at bottom, consumes its height
//   - content fills the remaining space in the middle
type DockNode struct {
	// Docked children for each edge.
	// Multiple children on the same edge are processed in order,
	// each consuming space from the remaining area.
	Top    []LayoutNode
	Bottom []LayoutNode
	Left   []LayoutNode
	Right  []LayoutNode

	// Body fills the remaining space after all edges are processed.
	// Optional - if nil, the dock only contains edge children.
	Body LayoutNode

	// DockOrder specifies the order in which edges are processed.
	// If nil or empty, defaults to [DockTop, DockBottom, DockLeft, DockRight].
	// Each edge in this list is processed once, laying out all children for that edge.
	DockOrder []DockEdge

	// Container's own insets (optional).
	Padding EdgeInsets
	Border  EdgeInsets
	Margin  EdgeInsets

	// Container's own size constraints (optional, 0 means unconstrained).
	MinWidth  int
	MaxWidth  int
	MinHeight int
	MaxHeight int
}

// ComputeLayout computes the layout for this dock container.
func (d *DockNode) ComputeLayout(constraints Constraints) ComputedLayout {
	// Combine parent constraints with node's own constraints
	effective := d.effectiveConstraints(constraints)

	// Convert to content-box constraints (space available for children)
	contentConstraints := d.toContentConstraints(effective)

	// Track remaining space as we dock children
	remaining := remainingRect{
		x:      0,
		y:      0,
		width:  contentConstraints.MaxWidth,
		height: contentConstraints.MaxHeight,
	}

	// Collect all positioned children
	var positioned []PositionedChild

	// Get dock order (default if not specified)
	order := d.dockOrder()

	// Process each edge in order
	for _, edge := range order {
		children := d.childrenForEdge(edge)
		for _, child := range children {
			pos, layout := d.dockChild(child, edge, &remaining)
			positioned = append(positioned, PositionedChild{
				X:      pos.x,
				Y:      pos.y,
				Layout: layout,
			})
		}
	}

	// Layout body in remaining space
	if d.Body != nil {
		bodyConstraints := Tight(remaining.width, remaining.height)
		bodyLayout := d.Body.ComputeLayout(bodyConstraints)

		// Adjust for body's margin
		x := remaining.x + bodyLayout.Box.Margin.Left
		y := remaining.y + bodyLayout.Box.Margin.Top

		positioned = append(positioned, PositionedChild{
			X:      x,
			Y:      y,
			Layout: bodyLayout,
		})
	}

	// Build the final BoxModel
	return d.buildResult(effective, contentConstraints, positioned)
}

// effectiveConstraints combines parent constraints with node's own min/max constraints.
func (d *DockNode) effectiveConstraints(parent Constraints) Constraints {
	return parent.WithNodeConstraints(d.MinWidth, d.MaxWidth, d.MinHeight, d.MaxHeight)
}

// toContentConstraints converts border-box constraints to content-box constraints.
func (d *DockNode) toContentConstraints(constraints Constraints) Constraints {
	hInset := d.Padding.Horizontal() + d.Border.Horizontal()
	vInset := d.Padding.Vertical() + d.Border.Vertical()

	return Constraints{
		MinWidth:  max(0, constraints.MinWidth-hInset),
		MaxWidth:  max(0, constraints.MaxWidth-hInset),
		MinHeight: max(0, constraints.MinHeight-vInset),
		MaxHeight: max(0, constraints.MaxHeight-vInset),
	}
}

// dockOrder returns the edge processing order.
func (d *DockNode) dockOrder() []DockEdge {
	if len(d.DockOrder) > 0 {
		return d.DockOrder
	}
	// Default order: Top, Bottom, Left, Right
	return []DockEdge{DockTop, DockBottom, DockLeft, DockRight}
}

// childrenForEdge returns the children docked to the given edge.
func (d *DockNode) childrenForEdge(edge DockEdge) []LayoutNode {
	switch edge {
	case DockTop:
		return d.Top
	case DockBottom:
		return d.Bottom
	case DockLeft:
		return d.Left
	case DockRight:
		return d.Right
	default:
		return nil
	}
}

// remainingRect tracks the remaining space as edges consume it.
type remainingRect struct {
	x, y          int
	width, height int
}

// position holds x, y coordinates.
type position struct {
	x, y int
}

// dockChild lays out a child at the given edge and updates remaining space.
func (d *DockNode) dockChild(child LayoutNode, edge DockEdge, remaining *remainingRect) (position, ComputedLayout) {
	var childConstraints Constraints
	var pos position

	switch edge {
	case DockTop:
		// Top: full width, flexible height up to remaining
		childConstraints = Constraints{
			MinWidth:  remaining.width,
			MaxWidth:  remaining.width,
			MinHeight: 0,
			MaxHeight: remaining.height,
		}
	case DockBottom:
		// Bottom: full width, flexible height up to remaining
		childConstraints = Constraints{
			MinWidth:  remaining.width,
			MaxWidth:  remaining.width,
			MinHeight: 0,
			MaxHeight: remaining.height,
		}
	case DockLeft:
		// Left: flexible width, full height
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  remaining.width,
			MinHeight: remaining.height,
			MaxHeight: remaining.height,
		}
	case DockRight:
		// Right: flexible width, full height
		childConstraints = Constraints{
			MinWidth:  0,
			MaxWidth:  remaining.width,
			MinHeight: remaining.height,
			MaxHeight: remaining.height,
		}
	}

	// Layout the child
	layout := child.ComputeLayout(childConstraints)
	marginBoxWidth := layout.Box.MarginBoxWidth()
	marginBoxHeight := layout.Box.MarginBoxHeight()

	// Position and consume space based on edge
	switch edge {
	case DockTop:
		pos = position{x: remaining.x + layout.Box.Margin.Left, y: remaining.y + layout.Box.Margin.Top}
		remaining.y += marginBoxHeight
		remaining.height = max(0, remaining.height-marginBoxHeight)

	case DockBottom:
		bottomY := remaining.y + remaining.height - marginBoxHeight
		pos = position{x: remaining.x + layout.Box.Margin.Left, y: bottomY + layout.Box.Margin.Top}
		remaining.height = max(0, remaining.height-marginBoxHeight)

	case DockLeft:
		pos = position{x: remaining.x + layout.Box.Margin.Left, y: remaining.y + layout.Box.Margin.Top}
		remaining.x += marginBoxWidth
		remaining.width = max(0, remaining.width-marginBoxWidth)

	case DockRight:
		rightX := remaining.x + remaining.width - marginBoxWidth
		pos = position{x: rightX + layout.Box.Margin.Left, y: remaining.y + layout.Box.Margin.Top}
		remaining.width = max(0, remaining.width-marginBoxWidth)
	}

	return pos, layout
}

// buildResult constructs the final ComputedLayout.
func (d *DockNode) buildResult(
	constraints Constraints,
	contentConstraints Constraints,
	children []PositionedChild,
) ComputedLayout {
	// Dock fills available space (uses max constraints for content area)
	contentWidth := contentConstraints.MaxWidth
	contentHeight := contentConstraints.MaxHeight

	// Convert content size to border-box
	hInset := d.Padding.Horizontal() + d.Border.Horizontal()
	vInset := d.Padding.Vertical() + d.Border.Vertical()
	borderBoxWidth := contentWidth + hInset
	borderBoxHeight := contentHeight + vInset

	// Clamp to original constraints
	borderBoxWidth, borderBoxHeight = constraints.Constrain(borderBoxWidth, borderBoxHeight)

	return ComputedLayout{
		Box: BoxModel{
			Width:   borderBoxWidth,
			Height:  borderBoxHeight,
			Padding: d.Padding,
			Border:  d.Border,
			Margin:  d.Margin,
		},
		Children: children,
	}
}
