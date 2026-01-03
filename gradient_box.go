package terma

// GradientBox is a container widget that renders a vertical gradient background
// behind its child. The gradient flows from top to bottom.
type GradientBox struct {
	ID       string    // Optional unique identifier for the widget
	Gradient Gradient  // The gradient to render as background
	Child    Widget    // The child widget to render on top
	Width    Dimension // Optional width (zero value = auto)
	Height   Dimension // Optional height (zero value = auto)
}

// WidgetID returns the widget's unique identifier.
func (g GradientBox) WidgetID() string {
	return g.ID
}

// Build returns itself as GradientBox manages its own child.
func (g GradientBox) Build(ctx BuildContext) Widget {
	return g
}

// GetDimensions returns the width and height dimension preferences.
func (g GradientBox) GetDimensions() (width, height Dimension) {
	return g.Width, g.Height
}

// Layout computes the size of the gradient box.
func (g GradientBox) Layout(ctx BuildContext, constraints Constraints) Size {
	// Determine our size based on dimensions or child
	var width, height int

	// Get child size if we have one
	var childWidth, childHeight int
	if g.Child != nil {
		built := g.Child.Build(ctx)
		if layoutable, ok := built.(Layoutable); ok {
			size := layoutable.Layout(ctx, constraints)
			childWidth = size.Width
			childHeight = size.Height
		}
	}

	// Width
	switch {
	case g.Width.IsCells():
		width = g.Width.CellsValue()
	case g.Width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = childWidth
	}

	// Height
	switch {
	case g.Height.IsCells():
		height = g.Height.CellsValue()
	case g.Height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = childHeight
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

// Render draws the gradient background and then the child.
func (g GradientBox) Render(ctx *RenderContext) {
	// Render vertical gradient background (top to bottom)
	if ctx.Height > 0 {
		for row := 0; row < ctx.Height; row++ {
			// Calculate gradient position (0 at top, 1 at bottom)
			t := float64(row) / float64(ctx.Height-1)
			if ctx.Height == 1 {
				t = 0.5
			}
			color := g.Gradient.At(t)
			ctx.FillRect(0, row, ctx.Width, 1, color)
		}
	}

	// Render child on top
	if g.Child != nil {
		ctx.RenderChild(0, g.Child, 0, 0, ctx.Width, ctx.Height)
	}
}
