package terma

import "github.com/darrenburns/terma/layout"

// MainAxisAlign specifies how children are aligned along the main axis.
// For Row, main axis is horizontal. For Column, main axis is vertical.
type MainAxisAlign int

const (
	// MainAxisStart aligns children at the start (default).
	MainAxisStart MainAxisAlign = iota
	// MainAxisCenter centers children along the main axis.
	MainAxisCenter
	// MainAxisEnd aligns children at the end.
	MainAxisEnd
	// Future: MainAxisSpaceBetween, MainAxisSpaceAround, MainAxisSpaceEvenly
)

// CrossAxisAlign specifies how children are aligned along the cross axis.
// For Row, cross axis is vertical. For Column, cross axis is horizontal.
type CrossAxisAlign int

const (
	// CrossAxisStart aligns children at the start of the cross axis.
	CrossAxisStart CrossAxisAlign = iota
	// CrossAxisStretch stretches children to fill the cross axis (default).
	CrossAxisStretch
	// CrossAxisCenter centers children along the cross axis.
	CrossAxisCenter
	// CrossAxisEnd aligns children at the end of the cross axis.
	CrossAxisEnd
)

// Row arranges its children horizontally.
type Row struct {
	ID         string         // Optional unique identifier for the widget
	Width      Dimension      // Deprecated: use Style.Width
	Height     Dimension      // Deprecated: use Style.Height
	Style      Style          // Optional styling (background color)
	Spacing    int            // Space between children
	MainAlign  MainAxisAlign  // Main axis (horizontal) alignment
	CrossAlign CrossAxisAlign // Cross axis (vertical) alignment
	Children   []Widget
	Click      func(MouseEvent) // Optional callback invoked when clicked
	MouseDown  func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp    func(MouseEvent) // Optional callback invoked when mouse is released
	Hover      func(HoverEvent) // Optional callback invoked when hover state changes
}

// GetContentDimensions returns the width and height dimension preferences.
func (r Row) GetContentDimensions() (width, height Dimension) {
	dims := r.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = r.Width
	}
	if height.IsUnset() {
		height = r.Height
	}
	return width, height
}

// GetStyle returns the style of the row.
func (r Row) GetStyle() Style {
	return r.Style
}

// WidgetID returns the row's unique identifier.
// Implements the Identifiable interface.
func (r Row) WidgetID() string {
	return r.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (r Row) OnClick(event MouseEvent) {
	if r.Click != nil {
		r.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (r Row) OnMouseDown(event MouseEvent) {
	if r.MouseDown != nil {
		r.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (r Row) OnMouseUp(event MouseEvent) {
	if r.MouseUp != nil {
		r.MouseUp(event)
	}
}

// OnHover is called on hover enter/leave transitions.
// Implements the Hoverable interface.
func (r Row) OnHover(event HoverEvent) {
	if r.Hover != nil {
		r.Hover(event)
	}
}

// Build returns itself as Row manages its own children.
func (r Row) Build(ctx BuildContext) Widget {
	return r
}

// BuildLayoutNode builds a layout node for this Row widget.
// Implements the LayoutNodeBuilder interface.
func (r Row) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	children := make([]layout.LayoutNode, len(r.Children))
	for i, child := range r.Children {
		childCtx := ctx.PushChild(i)
		built := child.Build(childCtx)

		// Build the child's layout node
		var childNode layout.LayoutNode
		if builder, ok := built.(LayoutNodeBuilder); ok {
			childNode = builder.BuildLayoutNode(childCtx)
		} else {
			// Fallback: create a BoxNode for widgets without LayoutNodeBuilder
			childNode = buildFallbackLayoutNode(built, childCtx)
		}

		// Wrap in FlexNode or PercentNode if child has Flex/Percent width (Row's main axis is horizontal)
		mainAxisDim := getChildMainAxisDimension(built, true)
		childNode = wrapInPercentIfNeeded(childNode, mainAxisDim, layout.Horizontal)
		children[i] = wrapInFlexIfNeeded(childNode, mainAxisDim)
	}

	padding := toLayoutEdgeInsets(r.Style.Padding)
	border := borderToEdgeInsets(r.Style.Border)
	dims := GetWidgetDimensionSet(r)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	// Explicit Auto means "fit content, don't stretch" - set preserve flags
	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	node := layout.LayoutNode(&layout.RowNode{
		Spacing:        r.Spacing,
		MainAlign:      toLayoutMainAlign(r.MainAlign),
		CrossAlign:     toLayoutCrossAlign(r.CrossAlign),
		Children:       children,
		Padding:        padding,
		Border:         border,
		Margin:         toLayoutEdgeInsets(r.Style.Margin),
		MinWidth:       minWidth,
		MaxWidth:       maxWidth,
		MinHeight:      minHeight,
		MaxHeight:      maxHeight,
		ExpandWidth:    dims.Width.IsFlex(),
		ExpandHeight:   dims.Height.IsFlex(),
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
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

// Render is a no-op for Row when using the RenderTree-based rendering path.
// Child positioning is handled by renderTree() which uses the computed layout.
// Row has no content of its own to render - it's purely a layout container.
func (r Row) Render(ctx *RenderContext) {
	// No-op: children are positioned by renderTree() using ComputedLayout.Children
}

// Column arranges its children vertically.
type Column struct {
	ID         string         // Optional unique identifier for the widget
	Width      Dimension      // Deprecated: use Style.Width
	Height     Dimension      // Deprecated: use Style.Height
	Style      Style          // Optional styling (background color)
	Spacing    int            // Space between children
	MainAlign  MainAxisAlign  // Main axis (vertical) alignment
	CrossAlign CrossAxisAlign // Cross axis (horizontal) alignment
	Children   []Widget
	Click      func(MouseEvent) // Optional callback invoked when clicked
	MouseDown  func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp    func(MouseEvent) // Optional callback invoked when mouse is released
	Hover      func(HoverEvent) // Optional callback invoked when hover state changes
}

// GetContentDimensions returns the width and height dimension preferences.
func (c Column) GetContentDimensions() (width, height Dimension) {
	dims := c.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = c.Width
	}
	if height.IsUnset() {
		height = c.Height
	}
	return width, height
}

// GetStyle returns the style of the column.
func (c Column) GetStyle() Style {
	return c.Style
}

// WidgetID returns the column's unique identifier.
// Implements the Identifiable interface.
func (c Column) WidgetID() string {
	return c.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (c Column) OnClick(event MouseEvent) {
	if c.Click != nil {
		c.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (c Column) OnMouseDown(event MouseEvent) {
	if c.MouseDown != nil {
		c.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (c Column) OnMouseUp(event MouseEvent) {
	if c.MouseUp != nil {
		c.MouseUp(event)
	}
}

// OnHover is called on hover enter/leave transitions.
// Implements the Hoverable interface.
func (c Column) OnHover(event HoverEvent) {
	if c.Hover != nil {
		c.Hover(event)
	}
}

// Build returns itself as Column manages its own children.
func (c Column) Build(ctx BuildContext) Widget {
	return c
}

// BuildLayoutNode creates a ColumnNode for the layout system.
func (c Column) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	children := make([]layout.LayoutNode, len(c.Children))
	for i, child := range c.Children {
		childCtx := ctx.PushChild(i)
		built := child.Build(childCtx)

		// Build the child's layout node
		var childNode layout.LayoutNode
		if builder, ok := built.(LayoutNodeBuilder); ok {
			childNode = builder.BuildLayoutNode(childCtx)
		} else {
			// Fallback: create a BoxNode for widgets without LayoutNodeBuilder
			childNode = buildFallbackLayoutNode(built, childCtx)
		}

		// Wrap in FlexNode or PercentNode if child has Flex/Percent height (Column's main axis is vertical)
		mainAxisDim := getChildMainAxisDimension(built, false)
		childNode = wrapInPercentIfNeeded(childNode, mainAxisDim, layout.Vertical)
		children[i] = wrapInFlexIfNeeded(childNode, mainAxisDim)
	}

	padding := toLayoutEdgeInsets(c.Style.Padding)
	border := borderToEdgeInsets(c.Style.Border)
	dims := GetWidgetDimensionSet(c)
	minWidth, maxWidth, minHeight, maxHeight := dimensionSetToMinMax(dims, padding, border)

	// Explicit Auto means "fit content, don't stretch" - set preserve flags
	preserveWidth := dims.Width.IsAuto() && !dims.Width.IsUnset()
	preserveHeight := dims.Height.IsAuto() && !dims.Height.IsUnset()

	node := layout.LayoutNode(&layout.ColumnNode{
		Spacing:        c.Spacing,
		MainAlign:      toLayoutMainAlign(c.MainAlign),
		CrossAlign:     toLayoutCrossAlign(c.CrossAlign),
		Children:       children,
		Padding:        padding,
		Border:         border,
		Margin:         toLayoutEdgeInsets(c.Style.Margin),
		MinWidth:       minWidth,
		MaxWidth:       maxWidth,
		MinHeight:      minHeight,
		MaxHeight:      maxHeight,
		ExpandWidth:    dims.Width.IsFlex(),
		ExpandHeight:   dims.Height.IsFlex(),
		PreserveWidth:  preserveWidth,
		PreserveHeight: preserveHeight,
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

// Render is a no-op for Column when using the RenderTree-based rendering path.
func (c Column) Render(ctx *RenderContext) {
	// No-op: children are positioned by renderTree() using ComputedLayout.Children
}
