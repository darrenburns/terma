package terma

import "terma/layout"

// TooltipPosition specifies where the tooltip appears relative to the child.
type TooltipPosition int

const (
	// TooltipTop positions the tooltip above the child (default).
	TooltipTop TooltipPosition = iota
	// TooltipBottom positions the tooltip below the child.
	TooltipBottom
	// TooltipLeft positions the tooltip to the left of the child.
	TooltipLeft
	// TooltipRight positions the tooltip to the right of the child.
	TooltipRight
)


// Tooltip displays contextual help when hovering over or focusing on an element.
// It wraps a child widget and shows a floating tooltip based on the configured trigger.
//
// Example - simple hover tooltip:
//
//	Tooltip{
//	    ID:      "submit-tooltip",
//	    Content: "Submit the form",
//	    Child:   Button{ID: "submit", Label: "Submit"},
//	}
//
// Example - positioned tooltip with focus trigger:
//
//	Tooltip{
//	    ID:       "password-tooltip",
//	    Content:  "3-20 characters",
//	    Position: TooltipRight,
//	    Child:    TextInput{ID: "password", State: passState},
//	}
//
// Example - rich text tooltip:
//
//	Tooltip{
//	    ID:       "save-tooltip",
//	    Spans:    ParseMarkup("[b]Ctrl+S[/] to save", ctx.Theme()),
//	    Position: TooltipBottom,
//	    Child:    SaveIcon{},
//	}
type Tooltip struct {
	// ID optionally identifies this tooltip for anchor positioning.
	// If not provided, an auto-generated ID is used.
	ID string

	// Content is the plain text to display in the tooltip.
	// Used if Spans is empty.
	Content string

	// Spans is the rich text to display in the tooltip.
	// Takes precedence over Content if non-empty.
	Spans []Span

	// Child is the widget that triggers the tooltip.
	Child Widget

	// Position specifies where the tooltip appears relative to the child.
	// Default: TooltipTop
	Position TooltipPosition

	// Offset is the gap in cells between the child and the tooltip.
	// Default: 0 (no gap)
	Offset int

	// Style overrides the default tooltip styling.
	Style Style
}

// WidgetID returns the tooltip's ID for anchor positioning.
// Returns the explicit ID if set, otherwise empty string to use auto-ID.
func (t Tooltip) WidgetID() string {
	return t.ID
}

// anchorID returns the ID used for anchor positioning.
// Uses explicit ID if set, otherwise falls back to auto-generated ID.
func (t Tooltip) anchorID(ctx BuildContext) string {
	if t.ID != "" {
		return t.ID
	}
	return ctx.AutoID()
}

// Build constructs the tooltip widget tree.
func (t Tooltip) Build(ctx BuildContext) Widget {
	// Determine visibility (tooltip shows when child is focused)
	visible := t.isVisible(ctx)

	// Register tooltip overlay if visible
	if visible {
		// Get the anchor ID (explicit or auto-generated)
		anchorID := t.anchorID(ctx)
		Floating{
			Visible: true,
			Config: FloatConfig{
				AnchorID: anchorID,
				Anchor:   t.anchorPoint(),
				Offset:   t.offsetValue(),
			},
			Child: t.buildContent(ctx),
		}.Build(ctx)
	}

	// Return self - Tooltip acts as the anchor widget
	return t
}

// ChildWidgets returns the wrapped child for render tree building.
func (t Tooltip) ChildWidgets() []Widget {
	if t.Child != nil {
		return []Widget{t.Child}
	}
	return nil
}

// BuildLayoutNode wraps the child in a container for layout.
func (t Tooltip) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	if t.Child == nil {
		return &layout.BoxNode{}
	}

	childCtx := ctx.PushChild(0)

	// Build the child to get its layout node
	built := t.Child.Build(childCtx)
	var childNode layout.LayoutNode
	if builder, ok := built.(LayoutNodeBuilder); ok {
		childNode = builder.BuildLayoutNode(childCtx)
	} else {
		childNode = &layout.BoxNode{}
	}

	// Wrap in ColumnNode so ComputedLayout.Children is populated
	return &layout.ColumnNode{
		Children: []layout.LayoutNode{childNode},
	}
}

// isVisible determines if the tooltip should be shown (when child is focused).
func (t Tooltip) isVisible(ctx BuildContext) bool {
	if t.Child == nil {
		return false
	}
	return ctx.IsFocused(t.Child)
}

// anchorPoint maps TooltipPosition to AnchorPoint.
func (t Tooltip) anchorPoint() AnchorPoint {
	switch t.Position {
	case TooltipTop:
		return AnchorTopCenter
	case TooltipBottom:
		return AnchorBottomCenter
	case TooltipLeft:
		return AnchorLeftCenter
	case TooltipRight:
		return AnchorRightCenter
	default:
		return AnchorTopCenter
	}
}

// offsetValue calculates the offset based on position and gap.
func (t Tooltip) offsetValue() Offset {
	gap := t.Offset
	switch t.Position {
	case TooltipTop:
		return Offset{Y: -gap}
	case TooltipBottom:
		return Offset{Y: gap}
	case TooltipLeft:
		return Offset{X: -gap}
	case TooltipRight:
		return Offset{X: gap}
	default:
		return Offset{Y: -gap}
	}
}

// buildContent creates the tooltip content widget with styling.
func (t Tooltip) buildContent(ctx BuildContext) Widget {
	style := t.tooltipStyle(ctx)

	if len(t.Spans) > 0 {
		return Text{Spans: t.Spans, Wrap: WrapSoft, Style: style}
	}
	return Text{Content: t.Content, Wrap: WrapSoft, Style: style}
}

// tooltipStyle returns the style for the tooltip, applying theme defaults.
func (t Tooltip) tooltipStyle(ctx BuildContext) Style {
	style := t.Style
	theme := ctx.Theme()

	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = theme.Text
	}
	if style.Padding.Top == 0 && style.Padding.Right == 0 && style.Padding.Bottom == 0 && style.Padding.Left == 0 {
		style.Padding = EdgeInsetsXY(1, 0) // horizontal padding
	}

	return style
}
