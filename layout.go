package terma

// Row arranges its children horizontally.
type Row struct {
	ID       string    // Optional unique identifier for the widget
	Width    Dimension // Optional width (zero value = auto)
	Height   Dimension // Optional height (zero value = auto)
	Style    Style     // Optional styling (background color)
	Spacing  int       // Space between children
	Children []Widget
	Click    func()     // Optional callback invoked when clicked
	Hover    func(bool) // Optional callback invoked when hover state changes
}

// GetDimensions returns the width and height dimension preferences.
func (r Row) GetDimensions() (width, height Dimension) {
	return r.Width, r.Height
}

// GetStyle returns the style of the row.
func (r Row) GetStyle() Style {
	return r.Style
}

// Key returns the row's unique identifier.
// Implements the Keyed interface.
func (r Row) Key() string {
	return r.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (r Row) OnClick() {
	if r.Click != nil {
		r.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (r Row) OnHover(hovered bool) {
	if r.Hover != nil {
		r.Hover(hovered)
	}
}

// Build returns itself as Row manages its own children.
func (r Row) Build(ctx BuildContext) Widget {
	return r
}

// Layout computes the size of the row and positions children.
func (r Row) Layout(constraints Constraints) Size {
	// Two-pass layout algorithm for fractional dimensions:
	// Pass 1: Measure fixed/auto children, collect fr children
	// Pass 2: Distribute remaining space to fr children

	type childInfo struct {
		built      Widget
		layoutable Layoutable
		widthDim   Dimension
		size       Size
		isFr       bool
		hInset     int // horizontal padding + margin
		vInset     int // vertical padding + margin
	}

	children := make([]childInfo, len(r.Children))
	totalFixedWidth := 0
	totalFr := 0.0
	maxHeight := 0

	// Account for spacing
	spacingTotal := 0
	if len(r.Children) > 1 {
		spacingTotal = r.Spacing * (len(r.Children) - 1)
	}
	availableWidth := constraints.MaxWidth - spacingTotal

	// Pass 1: Measure fixed/auto children and collect fr info
	for i, child := range r.Children {
		built := child.Build(BuildContext{})
		layoutable, ok := built.(Layoutable)
		if !ok {
			continue
		}

		children[i].built = built
		children[i].layoutable = layoutable

		// Get padding/margin/border insets
		if styled, ok := built.(Styled); ok {
			style := styled.GetStyle()
			borderWidth := style.Border.Width()
			children[i].hInset = style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
			children[i].vInset = style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
		}

		// Check if child has dimension preferences
		var widthDim Dimension
		if dimensioned, ok := built.(Dimensioned); ok {
			widthDim, _ = dimensioned.GetDimensions()
		}
		children[i].widthDim = widthDim

		if widthDim.IsFr() {
			children[i].isFr = true
			totalFr += widthDim.FrValue()
		} else {
			// Fixed or auto - measure now
			childConstraints := Constraints{
				MinWidth:  0,
				MaxWidth:  availableWidth - totalFixedWidth - children[i].hInset,
				MinHeight: constraints.MinHeight,
				MaxHeight: constraints.MaxHeight - children[i].vInset,
			}
			size := layoutable.Layout(childConstraints)
			children[i].size = size
			totalFixedWidth += size.Width + children[i].hInset
			totalHeight := size.Height + children[i].vInset
			if totalHeight > maxHeight {
				maxHeight = totalHeight
			}
		}
	}

	// Pass 2: Distribute remaining space to fr children
	remainingWidth := availableWidth - totalFixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	for i := range children {
		if !children[i].isFr || children[i].layoutable == nil {
			continue
		}

		// Calculate this child's share of remaining space
		frValue := children[i].widthDim.FrValue()
		childWidth := 0
		if totalFr > 0 {
			childWidth = int(float64(remainingWidth) * frValue / totalFr)
		}

		childConstraints := Constraints{
			MinWidth:  childWidth - children[i].hInset,
			MaxWidth:  childWidth - children[i].hInset,
			MinHeight: constraints.MinHeight,
			MaxHeight: constraints.MaxHeight - children[i].vInset,
		}
		size := children[i].layoutable.Layout(childConstraints)
		children[i].size = size
		totalHeight := size.Height + children[i].vInset
		if totalHeight > maxHeight {
			maxHeight = totalHeight
		}
	}

	// Calculate total width including spacing and insets
	totalWidth := 0
	for _, child := range children {
		totalWidth += child.size.Width + child.hInset
	}
	// Add spacing between children
	if len(r.Children) > 1 {
		totalWidth += r.Spacing * (len(r.Children) - 1)
	}

	// Determine final dimensions
	var width int
	switch {
	case r.Width.IsCells():
		width = r.Width.CellsValue()
	case r.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = totalWidth
	}

	var height int
	switch {
	case r.Height.IsCells():
		height = r.Height.CellsValue()
	case r.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = maxHeight
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

// Render draws the row's children.
func (r Row) Render(ctx *RenderContext) {
	// Calculate child widths using the same algorithm as Layout
	type childInfo struct {
		widthDim Dimension
		width    int
		isFr     bool
	}

	children := make([]childInfo, len(r.Children))
	totalFixedWidth := 0
	totalFr := 0.0

	// Account for spacing in available width
	spacingTotal := 0
	if len(r.Children) > 1 {
		spacingTotal = r.Spacing * (len(r.Children) - 1)
	}
	availableWidth := ctx.Width - spacingTotal

	// Pass 1: Measure fixed/auto children
	for i, child := range r.Children {
		built := child.Build(BuildContext{})

		var widthDim Dimension
		if dimensioned, ok := built.(Dimensioned); ok {
			widthDim, _ = dimensioned.GetDimensions()
		}
		children[i].widthDim = widthDim

		if widthDim.IsFr() {
			children[i].isFr = true
			totalFr += widthDim.FrValue()
		} else {
			// Fixed or auto - measure now
			if layoutable, ok := built.(Layoutable); ok {
				childConstraints := Constraints{
					MinWidth:  0,
					MaxWidth:  availableWidth - totalFixedWidth,
					MinHeight: 0,
					MaxHeight: ctx.Height,
				}
				size := layoutable.Layout(childConstraints)
				childWidth := size.Width
				// Add padding, margin, and border to get total space needed
				if styled, ok := built.(Styled); ok {
					style := styled.GetStyle()
					borderWidth := style.Border.Width()
					childWidth += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
				}
				children[i].width = childWidth
				totalFixedWidth += childWidth
			}
		}
	}

	// Pass 2: Calculate fr widths
	remainingWidth := availableWidth - totalFixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	for i := range children {
		if !children[i].isFr {
			continue
		}
		frValue := children[i].widthDim.FrValue()
		if totalFr > 0 {
			children[i].width = int(float64(remainingWidth) * frValue / totalFr)
		}
	}

	// Render children with calculated widths
	xOffset := 0
	for i, child := range r.Children {
		childWidth := children[i].width
		ctx.RenderChild(i, child, xOffset, 0, childWidth, ctx.Height)
		xOffset += childWidth
		// Add spacing after each child except the last
		if i < len(r.Children)-1 {
			xOffset += r.Spacing
		}
	}
}

// Column arranges its children vertically.
type Column struct {
	ID       string    // Optional unique identifier for the widget
	Width    Dimension // Optional width (zero value = auto)
	Height   Dimension // Optional height (zero value = auto)
	Style    Style     // Optional styling (background color)
	Spacing  int       // Space between children
	Children []Widget
	Click    func()     // Optional callback invoked when clicked
	Hover    func(bool) // Optional callback invoked when hover state changes
}

// GetDimensions returns the width and height dimension preferences.
func (c Column) GetDimensions() (width, height Dimension) {
	return c.Width, c.Height
}

// GetStyle returns the style of the column.
func (c Column) GetStyle() Style {
	return c.Style
}

// Key returns the column's unique identifier.
// Implements the Keyed interface.
func (c Column) Key() string {
	return c.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (c Column) OnClick() {
	if c.Click != nil {
		c.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (c Column) OnHover(hovered bool) {
	if c.Hover != nil {
		c.Hover(hovered)
	}
}

// Build returns itself as Column manages its own children.
func (c Column) Build(ctx BuildContext) Widget {
	return c
}

// Layout computes the size of the column and positions children.
func (c Column) Layout(constraints Constraints) Size {
	// Two-pass layout algorithm for fractional dimensions:
	// Pass 1: Measure fixed/auto children, collect fr children
	// Pass 2: Distribute remaining space to fr children

	type childInfo struct {
		built      Widget
		layoutable Layoutable
		heightDim  Dimension
		size       Size
		isFr       bool
		hInset     int // horizontal padding + margin
		vInset     int // vertical padding + margin
	}

	children := make([]childInfo, len(c.Children))
	totalFixedHeight := 0
	totalFr := 0.0
	maxWidth := 0

	// Account for spacing
	spacingTotal := 0
	if len(c.Children) > 1 {
		spacingTotal = c.Spacing * (len(c.Children) - 1)
	}
	availableHeight := constraints.MaxHeight - spacingTotal

	// Pass 1: Measure fixed/auto children and collect fr info
	for i, child := range c.Children {
		built := child.Build(BuildContext{})
		layoutable, ok := built.(Layoutable)
		if !ok {
			continue
		}

		children[i].built = built
		children[i].layoutable = layoutable

		// Get padding/margin/border insets
		if styled, ok := built.(Styled); ok {
			style := styled.GetStyle()
			borderWidth := style.Border.Width()
			children[i].hInset = style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
			children[i].vInset = style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
		}

		// Check if child has dimension preferences
		var heightDim Dimension
		if dimensioned, ok := built.(Dimensioned); ok {
			_, heightDim = dimensioned.GetDimensions()
		}
		children[i].heightDim = heightDim

		if heightDim.IsFr() {
			children[i].isFr = true
			totalFr += heightDim.FrValue()
		} else {
			// Fixed or auto - measure now
			childConstraints := Constraints{
				MinWidth:  constraints.MinWidth,
				MaxWidth:  constraints.MaxWidth - children[i].hInset,
				MinHeight: 0,
				MaxHeight: availableHeight - totalFixedHeight - children[i].vInset,
			}
			size := layoutable.Layout(childConstraints)
			children[i].size = size
			totalFixedHeight += size.Height + children[i].vInset
			totalWidth := size.Width + children[i].hInset
			if totalWidth > maxWidth {
				maxWidth = totalWidth
			}
		}
	}

	// Pass 2: Distribute remaining space to fr children
	remainingHeight := availableHeight - totalFixedHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	for i := range children {
		if !children[i].isFr || children[i].layoutable == nil {
			continue
		}

		// Calculate this child's share of remaining space
		frValue := children[i].heightDim.FrValue()
		childHeight := 0
		if totalFr > 0 {
			childHeight = int(float64(remainingHeight) * frValue / totalFr)
		}

		childConstraints := Constraints{
			MinWidth:  constraints.MinWidth,
			MaxWidth:  constraints.MaxWidth - children[i].hInset,
			MinHeight: childHeight - children[i].vInset,
			MaxHeight: childHeight - children[i].vInset,
		}
		size := children[i].layoutable.Layout(childConstraints)
		children[i].size = size
		totalWidth := size.Width + children[i].hInset
		if totalWidth > maxWidth {
			maxWidth = totalWidth
		}
	}

	// Calculate total height including spacing and insets
	totalHeight := 0
	for _, child := range children {
		totalHeight += child.size.Height + child.vInset
	}
	// Add spacing between children
	if len(c.Children) > 1 {
		totalHeight += c.Spacing * (len(c.Children) - 1)
	}

	// Determine final dimensions
	var width int
	switch {
	case c.Width.IsCells():
		width = c.Width.CellsValue()
	case c.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = maxWidth
	}

	var height int
	switch {
	case c.Height.IsCells():
		height = c.Height.CellsValue()
	case c.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = totalHeight
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

// Render draws the column's children.
func (c Column) Render(ctx *RenderContext) {
	// Calculate child heights using the same algorithm as Layout
	type childInfo struct {
		heightDim Dimension
		height    int
		isFr      bool
	}

	children := make([]childInfo, len(c.Children))
	totalFixedHeight := 0
	totalFr := 0.0

	// Account for spacing in available height
	spacingTotal := 0
	if len(c.Children) > 1 {
		spacingTotal = c.Spacing * (len(c.Children) - 1)
	}
	availableHeight := ctx.Height - spacingTotal

	// Pass 1: Measure fixed/auto children
	for i, child := range c.Children {
		built := child.Build(BuildContext{})

		var heightDim Dimension
		if dimensioned, ok := built.(Dimensioned); ok {
			_, heightDim = dimensioned.GetDimensions()
		}
		children[i].heightDim = heightDim

		if heightDim.IsFr() {
			children[i].isFr = true
			totalFr += heightDim.FrValue()
		} else {
			// Fixed or auto - measure now
			if layoutable, ok := built.(Layoutable); ok {
				childConstraints := Constraints{
					MinWidth:  0,
					MaxWidth:  ctx.Width,
					MinHeight: 0,
					MaxHeight: availableHeight - totalFixedHeight,
				}
				size := layoutable.Layout(childConstraints)
				childHeight := size.Height
				// Add padding, margin, and border to get total space needed
				if styled, ok := built.(Styled); ok {
					style := styled.GetStyle()
					borderWidth := style.Border.Width()
					childHeight += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
				}
				children[i].height = childHeight
				totalFixedHeight += childHeight
			}
		}
	}

	// Pass 2: Calculate fr heights
	remainingHeight := availableHeight - totalFixedHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	for i := range children {
		if !children[i].isFr {
			continue
		}
		frValue := children[i].heightDim.FrValue()
		if totalFr > 0 {
			children[i].height = int(float64(remainingHeight) * frValue / totalFr)
		}
	}

	// Render children with calculated heights
	yOffset := 0
	for i, child := range c.Children {
		childHeight := children[i].height
		ctx.RenderChild(i, child, 0, yOffset, ctx.Width, childHeight)
		yOffset += childHeight
		// Add spacing after each child except the last
		if i < len(c.Children)-1 {
			yOffset += c.Spacing
		}
	}
}
