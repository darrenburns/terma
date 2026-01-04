package terma

// Edge represents a docking edge.
type Edge int

const (
	Top Edge = iota
	Right
	Bottom
	Left
)

// Dock arranges children by docking them to edges, with a body filling the remaining space.
// Docked widgets consume space from the available area (they don't overlay).
// Use DockOrder to control which edges claim space first.
type Dock struct {
	ID        string    // Optional unique identifier
	DockOrder []Edge    // Priority order for edges. Default: Top, Bottom, Left, Right
	Top       []Widget  // Widgets docked to top (stack top-to-bottom)
	Right     []Widget  // Widgets docked to right (stack right-to-left)
	Bottom    []Widget  // Widgets docked to bottom (stack bottom-to-top)
	Left      []Widget  // Widgets docked to left (stack left-to-right)
	Body      Widget    // Main content, fills remaining space
	Width     Dimension
	Height    Dimension
	Style     Style
	Click     func()
	Hover     func(bool)
}

// defaultDockOrder is used when DockOrder is not specified.
var defaultDockOrder = []Edge{Top, Bottom, Left, Right}

// GetDimensions returns the width and height dimension preferences.
func (d Dock) GetDimensions() (width, height Dimension) {
	return d.Width, d.Height
}

// GetStyle returns the style of the dock.
func (d Dock) GetStyle() Style {
	return d.Style
}

// WidgetID returns the dock's unique identifier.
func (d Dock) WidgetID() string {
	return d.ID
}

// OnClick is called when the widget is clicked.
func (d Dock) OnClick() {
	if d.Click != nil {
		d.Click()
	}
}

// OnHover is called when the hover state changes.
func (d Dock) OnHover(hovered bool) {
	if d.Hover != nil {
		d.Hover(hovered)
	}
}

// Build returns itself as Dock manages its own children.
func (d Dock) Build(ctx BuildContext) Widget {
	return d
}

// dockOrder returns the effective dock order.
func (d Dock) dockOrder() []Edge {
	if len(d.DockOrder) > 0 {
		return d.DockOrder
	}
	return defaultDockOrder
}

// edgeWidgets returns the widgets for a given edge.
func (d Dock) edgeWidgets(edge Edge) []Widget {
	switch edge {
	case Top:
		return d.Top
	case Right:
		return d.Right
	case Bottom:
		return d.Bottom
	case Left:
		return d.Left
	default:
		return nil
	}
}

// Layout computes the size of the dock and positions children.
func (d Dock) Layout(ctx BuildContext, constraints Constraints) Size {
	// Track remaining space after docking
	remainingX := 0
	remainingY := 0
	remainingWidth := constraints.MaxWidth
	remainingHeight := constraints.MaxHeight

	// Process edges in dock order
	for _, edge := range d.dockOrder() {
		widgets := d.edgeWidgets(edge)
		if len(widgets) == 0 {
			continue
		}

		switch edge {
		case Top:
			for _, widget := range widgets {
				built := widget.Build(ctx)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  remainingWidth,
						MinHeight: 0,
						MaxHeight: remainingHeight,
					}
					size := layoutable.Layout(ctx, childConstraints)
					height := size.Height
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						height += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
					}
					remainingY += height
					remainingHeight -= height
					if remainingHeight < 0 {
						remainingHeight = 0
					}
				}
			}
		case Bottom:
			for _, widget := range widgets {
				built := widget.Build(ctx)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  remainingWidth,
						MinHeight: 0,
						MaxHeight: remainingHeight,
					}
					size := layoutable.Layout(ctx, childConstraints)
					height := size.Height
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						height += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
					}
					remainingHeight -= height
					if remainingHeight < 0 {
						remainingHeight = 0
					}
				}
			}
		case Left:
			for _, widget := range widgets {
				built := widget.Build(ctx)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  remainingWidth,
						MinHeight: 0,
						MaxHeight: remainingHeight,
					}
					size := layoutable.Layout(ctx, childConstraints)
					width := size.Width
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						width += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					}
					remainingX += width
					remainingWidth -= width
					if remainingWidth < 0 {
						remainingWidth = 0
					}
				}
			}
		case Right:
			for _, widget := range widgets {
				built := widget.Build(ctx)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  remainingWidth,
						MinHeight: 0,
						MaxHeight: remainingHeight,
					}
					size := layoutable.Layout(ctx, childConstraints)
					width := size.Width
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						width += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					}
					remainingWidth -= width
					if remainingWidth < 0 {
						remainingWidth = 0
					}
				}
			}
		}
	}

	// Layout body in remaining space
	if d.Body != nil {
		built := d.Body.Build(ctx)
		if layoutable, ok := built.(Layoutable); ok {
			childConstraints := Constraints{
				MinWidth:  remainingWidth,
				MaxWidth:  remainingWidth,
				MinHeight: remainingHeight,
				MaxHeight: remainingHeight,
			}
			layoutable.Layout(ctx, childConstraints)
		}
	}

	// Determine final dimensions
	var width int
	switch {
	case d.Width.IsCells():
		width = d.Width.CellsValue()
	case d.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = constraints.MaxWidth
	}

	var height int
	switch {
	case d.Height.IsCells():
		height = d.Height.CellsValue()
	case d.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = constraints.MaxHeight
	}

	// Clamp to constraints
	if width < constraints.MinWidth {
		width = constraints.MinWidth
	}
	if width > constraints.MaxWidth {
		width = constraints.MaxWidth
	}
	if height < constraints.MinHeight {
		height = constraints.MinHeight
	}
	if height > constraints.MaxHeight {
		height = constraints.MaxHeight
	}

	return Size{Width: width, Height: height}
}

// Render draws the dock's children.
func (d Dock) Render(ctx *RenderContext) {
	// Track each widget's render position and size
	type widgetLayout struct {
		widget Widget
		x, y   int
		width  int
		height int
	}

	var layouts []widgetLayout

	// Track the remaining content area as edges consume space
	contentX := 0
	contentY := 0
	contentWidth := ctx.Width
	contentHeight := ctx.Height

	// Process edges in dock order
	for _, edge := range d.dockOrder() {
		widgets := d.edgeWidgets(edge)
		if len(widgets) == 0 {
			continue
		}

		switch edge {
		case Top:
			// Top widgets stack downward from current contentY
			for _, widget := range widgets {
				built := widget.Build(ctx.buildContext)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  contentWidth,
						MinHeight: 0,
						MaxHeight: contentHeight,
					}
					size := layoutable.Layout(ctx.buildContext, childConstraints)
					totalHeight := size.Height
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						totalHeight += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
					}
					layouts = append(layouts, widgetLayout{
						widget: widget,
						x:      contentX,
						y:      contentY,
						width:  contentWidth,
						height: totalHeight,
					})
					contentY += totalHeight
					contentHeight -= totalHeight
					if contentHeight < 0 {
						contentHeight = 0
					}
				}
			}

		case Bottom:
			// Bottom widgets stack upward from the bottom of content area
			// Measure all bottom widgets first to find total height
			var bottomWidgets []widgetLayout
			totalBottomHeight := 0
			for _, widget := range widgets {
				built := widget.Build(ctx.buildContext)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  contentWidth,
						MinHeight: 0,
						MaxHeight: contentHeight - totalBottomHeight,
					}
					size := layoutable.Layout(ctx.buildContext, childConstraints)
					totalHeight := size.Height
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						totalHeight += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
					}
					bottomWidgets = append(bottomWidgets, widgetLayout{
						widget: widget,
						x:      contentX,
						width:  contentWidth,
						height: totalHeight,
					})
					totalBottomHeight += totalHeight
				}
			}
			// Position bottom widgets from the bottom of content area
			bottomY := contentY + contentHeight - totalBottomHeight
			for i := range bottomWidgets {
				bottomWidgets[i].y = bottomY
				bottomY += bottomWidgets[i].height
			}
			layouts = append(layouts, bottomWidgets...)
			contentHeight -= totalBottomHeight
			if contentHeight < 0 {
				contentHeight = 0
			}

		case Left:
			// Left widgets stack rightward from current contentX
			for _, widget := range widgets {
				built := widget.Build(ctx.buildContext)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  contentWidth,
						MinHeight: 0,
						MaxHeight: contentHeight,
					}
					size := layoutable.Layout(ctx.buildContext, childConstraints)
					totalWidth := size.Width
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						totalWidth += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					}
					layouts = append(layouts, widgetLayout{
						widget: widget,
						x:      contentX,
						y:      contentY,
						width:  totalWidth,
						height: contentHeight,
					})
					contentX += totalWidth
					contentWidth -= totalWidth
					if contentWidth < 0 {
						contentWidth = 0
					}
				}
			}

		case Right:
			// Right widgets stack leftward from the right of content area
			// Measure all right widgets first to find total width
			var rightWidgets []widgetLayout
			totalRightWidth := 0
			for _, widget := range widgets {
				built := widget.Build(ctx.buildContext)
				if layoutable, ok := built.(Layoutable); ok {
					childConstraints := Constraints{
						MinWidth:  0,
						MaxWidth:  contentWidth - totalRightWidth,
						MinHeight: 0,
						MaxHeight: contentHeight,
					}
					size := layoutable.Layout(ctx.buildContext, childConstraints)
					totalWidth := size.Width
					if styled, ok := built.(Styled); ok {
						style := styled.GetStyle()
						borderWidth := style.Border.Width()
						totalWidth += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					}
					rightWidgets = append(rightWidgets, widgetLayout{
						widget: widget,
						y:      contentY,
						width:  totalWidth,
						height: contentHeight,
					})
					totalRightWidth += totalWidth
				}
			}
			// Position right widgets from the right of content area
			rightX := contentX + contentWidth - totalRightWidth
			for i := range rightWidgets {
				rightWidgets[i].x = rightX
				rightX += rightWidgets[i].width
			}
			layouts = append(layouts, rightWidgets...)
			contentWidth -= totalRightWidth
			if contentWidth < 0 {
				contentWidth = 0
			}
		}
	}

	// Render all edge widgets
	for i, layout := range layouts {
		ctx.RenderChild(i, layout.widget, layout.x, layout.y, layout.width, layout.height)
	}

	// Render body in remaining content area
	if d.Body != nil {
		ctx.RenderChild(len(layouts), d.Body, contentX, contentY, contentWidth, contentHeight)
	}
}
