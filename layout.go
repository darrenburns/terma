package terma

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
	// CrossAxisStretch stretches children to fill the cross axis (default).
	CrossAxisStretch CrossAxisAlign = iota
	// CrossAxisStart aligns children at the start of the cross axis.
	CrossAxisStart
	// CrossAxisCenter centers children along the cross axis.
	CrossAxisCenter
	// CrossAxisEnd aligns children at the end of the cross axis.
	CrossAxisEnd
)

// Row arranges its children horizontally.
type Row struct {
	ID         string         // Optional unique identifier for the widget
	Width      Dimension      // Optional width (zero value = auto)
	Height     Dimension      // Optional height (zero value = auto)
	Style      Style          // Optional styling (background color)
	Spacing    int            // Space between children
	MainAlign  MainAxisAlign  // Main axis (horizontal) alignment
	CrossAlign CrossAxisAlign // Cross axis (vertical) alignment
	Children   []Widget
	Click      func()     // Optional callback invoked when clicked
	Hover      func(bool) // Optional callback invoked when hover state changes
}

// GetDimensions returns the width and height dimension preferences.
// Row defaults to Fr(1) width if not explicitly set.
func (r Row) GetDimensions() (width, height Dimension) {
	w := r.Width
	if w.IsUnset() {
		w = Fr(1)
	}
	return w, r.Height
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
func (r Row) Layout(ctx BuildContext, constraints Constraints) Size {
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
		built := child.Build(ctx)
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
			// For non-stretch cross-axis alignment, let children size naturally
			childMinHeight := constraints.MinHeight
			if r.CrossAlign != CrossAxisStretch {
				childMinHeight = 0
			}
			childConstraints := Constraints{
				MinWidth:  0,
				MaxWidth:  100000, // Unconstrained width - let children take natural size
				MinHeight: childMinHeight,
				MaxHeight: constraints.MaxHeight - children[i].vInset,
			}
			size := layoutable.Layout(ctx, childConstraints)
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

		// For non-stretch cross-axis alignment, let children size naturally
		childMinHeight := constraints.MinHeight
		if r.CrossAlign != CrossAxisStretch {
			childMinHeight = 0
		}
		childConstraints := Constraints{
			MinWidth:  childWidth - children[i].hInset,
			MaxWidth:  childWidth - children[i].hInset,
			MinHeight: childMinHeight,
			MaxHeight: constraints.MaxHeight - children[i].vInset,
		}
		size := children[i].layoutable.Layout(ctx, childConstraints)
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
		height   int
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
		built := child.Build(ctx.buildContext)

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
				// Get padding, margin, and border insets BEFORE layout
				var hInset, vInset int
				if styled, ok := built.(Styled); ok {
					style := styled.GetStyle()
					borderWidth := style.Border.Width()
					hInset = style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					vInset = style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
				}
				childConstraints := Constraints{
					MinWidth:  0,
					MaxWidth:  100000, // Unconstrained width - let children take natural size
					MinHeight: 0,
					MaxHeight: ctx.Height - vInset,
				}
				size := layoutable.Layout(ctx.buildContext, childConstraints)
				children[i].width = size.Width + hInset
				children[i].height = size.Height + vInset
				totalFixedWidth += children[i].width
			}
		}
	}

	// Pass 2: Calculate fr widths and measure fr children heights
	remainingWidth := availableWidth - totalFixedWidth
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	for i, child := range r.Children {
		if !children[i].isFr {
			continue
		}
		frValue := children[i].widthDim.FrValue()
		if totalFr > 0 {
			children[i].width = int(float64(remainingWidth) * frValue / totalFr)
		}
		// Measure height for Fr children
		built := child.Build(ctx.buildContext)
		if layoutable, ok := built.(Layoutable); ok {
			childConstraints := Constraints{
				MinWidth:  children[i].width,
				MaxWidth:  children[i].width,
				MinHeight: 0,
				MaxHeight: ctx.Height,
			}
			size := layoutable.Layout(ctx.buildContext, childConstraints)
			childHeight := size.Height
			if styled, ok := built.(Styled); ok {
				style := styled.GetStyle()
				borderWidth := style.Border.Width()
				childHeight += style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
			}
			children[i].height = childHeight
		}
	}

	// Calculate total content width for main axis alignment
	totalContentWidth := 0
	for _, child := range children {
		totalContentWidth += child.width
	}
	if len(r.Children) > 1 {
		totalContentWidth += r.Spacing * (len(r.Children) - 1)
	}

	// Calculate main axis starting offset
	xOffset := 0
	switch r.MainAlign {
	case MainAxisStart:
		xOffset = 0
	case MainAxisCenter:
		xOffset = (ctx.Width - totalContentWidth) / 2
	case MainAxisEnd:
		xOffset = ctx.Width - totalContentWidth
	}
	if xOffset < 0 {
		xOffset = 0
	}

	// Render children with calculated widths and alignment
	for i, child := range r.Children {
		childWidth := children[i].width
		childHeight := children[i].height
		renderHeight := childHeight

		// Calculate cross-axis (vertical) offset for this child
		yOffset := 0
		switch r.CrossAlign {
		case CrossAxisStretch:
			yOffset = 0
			renderHeight = ctx.Height
		case CrossAxisStart:
			yOffset = 0
		case CrossAxisCenter:
			yOffset = (ctx.Height - childHeight) / 2
		case CrossAxisEnd:
			yOffset = ctx.Height - childHeight
		}
		if yOffset < 0 {
			yOffset = 0
		}

		ctx.RenderChild(i, child, xOffset, yOffset, childWidth, renderHeight)
		xOffset += childWidth
		// Add spacing after each child except the last
		if i < len(r.Children)-1 {
			xOffset += r.Spacing
		}
	}
}

// Column arranges its children vertically.
type Column struct {
	ID         string         // Optional unique identifier for the widget
	Width      Dimension      // Optional width (zero value = auto)
	Height     Dimension      // Optional height (zero value = auto)
	Style      Style          // Optional styling (background color)
	Spacing    int            // Space between children
	MainAlign  MainAxisAlign  // Main axis (vertical) alignment
	CrossAlign CrossAxisAlign // Cross axis (horizontal) alignment
	Children   []Widget
	Click      func()     // Optional callback invoked when clicked
	Hover      func(bool) // Optional callback invoked when hover state changes
}

// GetDimensions returns the width and height dimension preferences.
// Column defaults to Fr(1) height if not explicitly set.
func (c Column) GetDimensions() (width, height Dimension) {
	h := c.Height
	if h.IsUnset() {
		h = Fr(1)
	}
	return c.Width, h
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
func (c Column) Layout(ctx BuildContext, constraints Constraints) Size {
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
		built := child.Build(ctx)
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
			// For non-stretch cross-axis alignment, let children size naturally
			childMinWidth := constraints.MinWidth
			if c.CrossAlign != CrossAxisStretch {
				childMinWidth = 0
			}
			childConstraints := Constraints{
				MinWidth:  childMinWidth,
				MaxWidth:  constraints.MaxWidth - children[i].hInset,
				MinHeight: 0,
				MaxHeight: 100000, // Unconstrained height - let children take natural size
			}
			size := layoutable.Layout(ctx, childConstraints)
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

		// For non-stretch cross-axis alignment, let children size naturally
		childMinWidth := constraints.MinWidth
		if c.CrossAlign != CrossAxisStretch {
			childMinWidth = 0
		}
		childConstraints := Constraints{
			MinWidth:  childMinWidth,
			MaxWidth:  constraints.MaxWidth - children[i].hInset,
			MinHeight: childHeight - children[i].vInset,
			MaxHeight: childHeight - children[i].vInset,
		}
		size := children[i].layoutable.Layout(ctx, childConstraints)
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
		width     int
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
		built := child.Build(ctx.buildContext)

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
				// Get padding, margin, and border insets BEFORE layout
				var hInset, vInset int
				if styled, ok := built.(Styled); ok {
					style := styled.GetStyle()
					borderWidth := style.Border.Width()
					hInset = style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
					vInset = style.Padding.Vertical() + style.Margin.Vertical() + borderWidth*2
				}
				childConstraints := Constraints{
					MinWidth:  0,
					MaxWidth:  ctx.Width - hInset,
					MinHeight: 0,
					MaxHeight: 100000, // Unconstrained height - let children take natural size
				}
				size := layoutable.Layout(ctx.buildContext, childConstraints)
				children[i].height = size.Height + vInset
				children[i].width = size.Width + hInset
				totalFixedHeight += children[i].height
			}
		}
	}

	// Pass 2: Calculate fr heights and measure fr children widths
	remainingHeight := availableHeight - totalFixedHeight
	if remainingHeight < 0 {
		remainingHeight = 0
	}

	for i, child := range c.Children {
		if !children[i].isFr {
			continue
		}
		frValue := children[i].heightDim.FrValue()
		if totalFr > 0 {
			children[i].height = int(float64(remainingHeight) * frValue / totalFr)
		}
		// Measure width for Fr children
		built := child.Build(ctx.buildContext)
		if layoutable, ok := built.(Layoutable); ok {
			childConstraints := Constraints{
				MinWidth:  0,
				MaxWidth:  ctx.Width,
				MinHeight: children[i].height,
				MaxHeight: children[i].height,
			}
			size := layoutable.Layout(ctx.buildContext, childConstraints)
			childWidth := size.Width
			if styled, ok := built.(Styled); ok {
				style := styled.GetStyle()
				borderWidth := style.Border.Width()
				childWidth += style.Padding.Horizontal() + style.Margin.Horizontal() + borderWidth*2
			}
			children[i].width = childWidth
		}
	}

	// Calculate total content height for main axis alignment
	totalContentHeight := 0
	for _, child := range children {
		totalContentHeight += child.height
	}
	if len(c.Children) > 1 {
		totalContentHeight += c.Spacing * (len(c.Children) - 1)
	}

	// Calculate main axis starting offset
	yOffset := 0
	switch c.MainAlign {
	case MainAxisStart:
		yOffset = 0
	case MainAxisCenter:
		yOffset = (ctx.Height - totalContentHeight) / 2
	case MainAxisEnd:
		yOffset = ctx.Height - totalContentHeight
	}
	if yOffset < 0 {
		yOffset = 0
	}

	// Render children with calculated heights and alignment
	for i, child := range c.Children {
		childHeight := children[i].height
		childWidth := children[i].width
		renderWidth := childWidth

		// Calculate cross-axis (horizontal) offset for this child
		xOffset := 0
		switch c.CrossAlign {
		case CrossAxisStretch:
			xOffset = 0
			renderWidth = ctx.Width
		case CrossAxisStart:
			xOffset = 0
		case CrossAxisCenter:
			xOffset = (ctx.Width - childWidth) / 2
		case CrossAxisEnd:
			xOffset = ctx.Width - childWidth
		}
		if xOffset < 0 {
			xOffset = 0
		}

		ctx.RenderChild(i, child, xOffset, yOffset, renderWidth, childHeight)
		yOffset += childHeight
		// Add spacing after each child except the last
		if i < len(c.Children)-1 {
			yOffset += c.Spacing
		}
	}
}
