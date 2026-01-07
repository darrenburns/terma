package layout

import (
	"fmt"

	"terma"
)

// BoxModel describes a rectangular region with a CSS-like box model.
// The model consists of four nested boxes (from innermost to outermost):
//
//   - Content box: the actual content area
//   - Padding box: content + padding
//   - Border box: content + padding + border
//   - Margin box: content + padding + border + margin (the bounding box)
//
// Unlike CSS where the "box" typically refers to the border-box,
// in this model the "box" refers to the margin-box (outermost boundary).
type BoxModel struct {
	// Content dimensions (innermost area)
	ContentWidth  int
	ContentHeight int

	// Min/max constraints (0 = no constraint)
	MinContentWidth  int
	MaxContentWidth  int
	MinContentHeight int
	MaxContentHeight int

	// Insets (from content outward)
	Padding terma.EdgeInsets
	Border  terma.EdgeInsets // Typically uniform (1,1,1,1) for borders
	Margin  terma.EdgeInsets

	// Scrolling (optional - zero values mean no scrolling)
	// VirtualWidth/Height represent the total scrollable content size.
	// When set, content can extend beyond the visible viewport.
	VirtualWidth  int // Total content width (0 = same as ContentWidth)
	VirtualHeight int // Total content height (0 = same as ContentHeight)
	ScrollOffsetX int // Horizontal scroll offset
	ScrollOffsetY int // Vertical scroll offset

	// ScrollbarWidth is the space reserved for a vertical scrollbar.
	// This reduces the usable content width when vertical scrolling is enabled.
	ScrollbarWidth int

	// ScrollbarHeight is the space reserved for a horizontal scrollbar.
	// This reduces the usable content height when horizontal scrolling is enabled.
	ScrollbarHeight int
}

// --- Validation methods ---

// Validate checks that all BoxModel fields have valid values.
// Panics if any field has an invalid value (negative dimensions, invalid constraints, etc.).
func (b BoxModel) Validate() {
	// Content dimensions
	if b.ContentWidth < 0 {
		panic("BoxModel: ContentWidth cannot be negative")
	}
	if b.ContentHeight < 0 {
		panic("BoxModel: ContentHeight cannot be negative")
	}

	// Insets
	validateEdgeInsets(b.Padding, "Padding")
	validateEdgeInsets(b.Border, "Border")
	validateEdgeInsets(b.Margin, "Margin")

	// Virtual dimensions
	if b.VirtualWidth < 0 {
		panic("BoxModel: VirtualWidth cannot be negative")
	}
	if b.VirtualHeight < 0 {
		panic("BoxModel: VirtualHeight cannot be negative")
	}

	// Scrollbars
	if b.ScrollbarWidth < 0 {
		panic("BoxModel: ScrollbarWidth cannot be negative")
	}
	if b.ScrollbarHeight < 0 {
		panic("BoxModel: ScrollbarHeight cannot be negative")
	}

	// Min/max constraints must be non-negative
	if b.MinContentWidth < 0 {
		panic("BoxModel: MinContentWidth cannot be negative")
	}
	if b.MaxContentWidth < 0 {
		panic("BoxModel: MaxContentWidth cannot be negative")
	}
	if b.MinContentHeight < 0 {
		panic("BoxModel: MinContentHeight cannot be negative")
	}
	if b.MaxContentHeight < 0 {
		panic("BoxModel: MaxContentHeight cannot be negative")
	}

	// Constraint validity (min <= max when both are set)
	if b.MinContentWidth > 0 && b.MaxContentWidth > 0 && b.MinContentWidth > b.MaxContentWidth {
		panic("BoxModel: MinContentWidth cannot exceed MaxContentWidth")
	}
	if b.MinContentHeight > 0 && b.MaxContentHeight > 0 && b.MinContentHeight > b.MaxContentHeight {
		panic("BoxModel: MinContentHeight cannot exceed MaxContentHeight")
	}
}

// validateEdgeInsets checks that all EdgeInsets values are non-negative.
func validateEdgeInsets(e terma.EdgeInsets, name string) {
	if e.Top < 0 || e.Right < 0 || e.Bottom < 0 || e.Left < 0 {
		panic(fmt.Sprintf("BoxModel: %s cannot have negative values", name))
	}
}

// --- Box dimension methods ---

// PaddingBoxWidth returns the width of the padding box (content + horizontal padding).
func (b BoxModel) PaddingBoxWidth() int {
	return b.ContentWidth + b.Padding.Horizontal()
}

// PaddingBoxHeight returns the height of the padding box (content + vertical padding).
func (b BoxModel) PaddingBoxHeight() int {
	return b.ContentHeight + b.Padding.Vertical()
}

// BorderBoxWidth returns the width of the border box (padding box + horizontal border).
func (b BoxModel) BorderBoxWidth() int {
	return b.PaddingBoxWidth() + b.Border.Horizontal()
}

// BorderBoxHeight returns the height of the border box (padding box + vertical border).
func (b BoxModel) BorderBoxHeight() int {
	return b.PaddingBoxHeight() + b.Border.Vertical()
}

// MarginBoxWidth returns the width of the margin box (border box + horizontal margin).
// This is the total width including all insets.
func (b BoxModel) MarginBoxWidth() int {
	return b.BorderBoxWidth() + b.Margin.Horizontal()
}

// MarginBoxHeight returns the height of the margin box (border box + vertical margin).
// This is the total height including all insets.
func (b BoxModel) MarginBoxHeight() int {
	return b.BorderBoxHeight() + b.Margin.Vertical()
}

// --- Rect-returning methods ---
// All rects are positioned relative to the margin box origin (0,0).

// ContentBox returns the full allocated content area as a Rect.
// This is the space given by the parent layout, before any scrollbar is subtracted.
// Use for: positioning this widget, drawing backgrounds, border rendering.
// The position is relative to the margin box origin.
func (b BoxModel) ContentBox() terma.Rect {
	return terma.Rect{
		X:      b.Margin.Left + b.Border.Left + b.Padding.Left,
		Y:      b.Margin.Top + b.Border.Top + b.Padding.Top,
		Width:  b.ContentWidth,
		Height: b.ContentHeight,
	}
}

// UsableContentBox returns the content area available for child widgets.
// This subtracts space reserved for scrollbars from the allocated content area:
//   - Vertical scrollbar (ScrollbarWidth) reduces width when IsScrollableY()
//   - Horizontal scrollbar (ScrollbarHeight) reduces height when IsScrollableX()
//
// Use for: laying out children, determining available space for content.
// The position is relative to the margin box origin.
func (b BoxModel) UsableContentBox() terma.Rect {
	return terma.Rect{
		X:      b.Margin.Left + b.Border.Left + b.Padding.Left,
		Y:      b.Margin.Top + b.Border.Top + b.Padding.Top,
		Width:  b.usableContentWidth(),
		Height: b.usableContentHeight(),
	}
}

// usableContentWidth returns the content width available for child widgets.
// This accounts for vertical scrollbar width when vertical scrolling is enabled.
func (b BoxModel) usableContentWidth() int {
	if b.IsScrollableY() && b.ScrollbarWidth > 0 {
		result := b.ContentWidth - b.ScrollbarWidth
		if result < 0 {
			return 0
		}
		return result
	}
	return b.ContentWidth
}

// usableContentHeight returns the content height available for child widgets.
// This accounts for horizontal scrollbar height when horizontal scrolling is enabled.
func (b BoxModel) usableContentHeight() int {
	if b.IsScrollableX() && b.ScrollbarHeight > 0 {
		result := b.ContentHeight - b.ScrollbarHeight
		if result < 0 {
			return 0
		}
		return result
	}
	return b.ContentHeight
}

// PaddingBox returns the padding box as a Rect.
// The position is relative to the margin box origin.
func (b BoxModel) PaddingBox() terma.Rect {
	return terma.Rect{
		X:      b.Margin.Left + b.Border.Left,
		Y:      b.Margin.Top + b.Border.Top,
		Width:  b.PaddingBoxWidth(),
		Height: b.PaddingBoxHeight(),
	}
}

// BorderBox returns the border box as a Rect.
// The position is relative to the margin box origin.
func (b BoxModel) BorderBox() terma.Rect {
	return terma.Rect{
		X:      b.Margin.Left,
		Y:      b.Margin.Top,
		Width:  b.BorderBoxWidth(),
		Height: b.BorderBoxHeight(),
	}
}

// MarginBox returns the margin box as a Rect.
// This is always positioned at (0,0) since it's the outermost boundary.
func (b BoxModel) MarginBox() terma.Rect {
	return terma.Rect{
		X:      0,
		Y:      0,
		Width:  b.MarginBoxWidth(),
		Height: b.MarginBoxHeight(),
	}
}

// ContentOrigin returns the offset from the margin box origin to the content origin.
// This is useful for positioning content within a rendered box.
func (b BoxModel) ContentOrigin() (x, y int) {
	return b.Margin.Left + b.Border.Left + b.Padding.Left,
		b.Margin.Top + b.Border.Top + b.Padding.Top
}

// BorderOrigin returns the offset from the margin box origin to the border origin.
// This is useful for drawing borders.
func (b BoxModel) BorderOrigin() (x, y int) {
	return b.Margin.Left, b.Margin.Top
}

// --- Inset calculations ---

// TotalHorizontalInset returns the total horizontal space taken by all insets.
func (b BoxModel) TotalHorizontalInset() int {
	return b.Padding.Horizontal() + b.Border.Horizontal() + b.Margin.Horizontal()
}

// TotalVerticalInset returns the total vertical space taken by all insets.
func (b BoxModel) TotalVerticalInset() int {
	return b.Padding.Vertical() + b.Border.Vertical() + b.Margin.Vertical()
}

// --- Constraint methods ---

// ClampContentWidth clamps the given width to the min/max content width constraints.
// A constraint of 0 means no constraint on that bound.
// Panics if width is negative.
func (b BoxModel) ClampContentWidth(width int) int {
	if width < 0 {
		panic("BoxModel: ClampContentWidth cannot accept negative width")
	}
	if b.MinContentWidth > 0 && width < b.MinContentWidth {
		return b.MinContentWidth
	}
	if b.MaxContentWidth > 0 && width > b.MaxContentWidth {
		return b.MaxContentWidth
	}
	return width
}

// ClampContentHeight clamps the given height to the min/max content height constraints.
// A constraint of 0 means no constraint on that bound.
// Panics if height is negative.
func (b BoxModel) ClampContentHeight(height int) int {
	if height < 0 {
		panic("BoxModel: ClampContentHeight cannot accept negative height")
	}
	if b.MinContentHeight > 0 && height < b.MinContentHeight {
		return b.MinContentHeight
	}
	if b.MaxContentHeight > 0 && height > b.MaxContentHeight {
		return b.MaxContentHeight
	}
	return height
}

// WithClampedContent returns a new BoxModel with content dimensions clamped to constraints.
func (b BoxModel) WithClampedContent() BoxModel {
	result := b
	result.ContentWidth = b.ClampContentWidth(b.ContentWidth)
	result.ContentHeight = b.ClampContentHeight(b.ContentHeight)
	return result
}

// SatisfiesWidthConstraints returns true if the content width is within constraints.
func (b BoxModel) SatisfiesWidthConstraints() bool {
	if b.MinContentWidth > 0 && b.ContentWidth < b.MinContentWidth {
		return false
	}
	if b.MaxContentWidth > 0 && b.ContentWidth > b.MaxContentWidth {
		return false
	}
	return true
}

// SatisfiesHeightConstraints returns true if the content height is within constraints.
func (b BoxModel) SatisfiesHeightConstraints() bool {
	if b.MinContentHeight > 0 && b.ContentHeight < b.MinContentHeight {
		return false
	}
	if b.MaxContentHeight > 0 && b.ContentHeight > b.MaxContentHeight {
		return false
	}
	return true
}

// SatisfiesConstraints returns true if content dimensions are within all constraints.
func (b BoxModel) SatisfiesConstraints() bool {
	return b.SatisfiesWidthConstraints() && b.SatisfiesHeightConstraints()
}

// --- Scrolling methods ---

// EffectiveVirtualWidth returns the virtual content width.
// Returns ContentWidth if VirtualWidth is not set (0).
func (b BoxModel) EffectiveVirtualWidth() int {
	if b.VirtualWidth > 0 {
		return b.VirtualWidth
	}
	return b.ContentWidth
}

// EffectiveVirtualHeight returns the virtual content height.
// Returns ContentHeight if VirtualHeight is not set (0).
func (b BoxModel) EffectiveVirtualHeight() int {
	if b.VirtualHeight > 0 {
		return b.VirtualHeight
	}
	return b.ContentHeight
}

// IsScrollableX returns true if horizontal scrolling is possible.
// This is true when virtual width exceeds the visible content width.
func (b BoxModel) IsScrollableX() bool {
	return b.EffectiveVirtualWidth() > b.ContentWidth
}

// IsScrollableY returns true if vertical scrolling is possible.
// This is true when virtual height exceeds the visible content height.
func (b BoxModel) IsScrollableY() bool {
	return b.EffectiveVirtualHeight() > b.ContentHeight
}

// IsScrollable returns true if scrolling is possible in either direction.
func (b BoxModel) IsScrollable() bool {
	return b.IsScrollableX() || b.IsScrollableY()
}

// MaxScrollX returns the maximum valid horizontal scroll offset.
// Returns 0 if not scrollable.
func (b BoxModel) MaxScrollX() int {
	max := b.EffectiveVirtualWidth() - b.ContentWidth
	if max < 0 {
		return 0
	}
	return max
}

// MaxScrollY returns the maximum valid vertical scroll offset.
// Returns 0 if not scrollable.
func (b BoxModel) MaxScrollY() int {
	max := b.EffectiveVirtualHeight() - b.ContentHeight
	if max < 0 {
		return 0
	}
	return max
}

// ClampScrollOffsetX clamps the given offset to valid horizontal scroll bounds.
func (b BoxModel) ClampScrollOffsetX(offset int) int {
	if offset < 0 {
		return 0
	}
	max := b.MaxScrollX()
	if offset > max {
		return max
	}
	return offset
}

// ClampScrollOffsetY clamps the given offset to valid vertical scroll bounds.
func (b BoxModel) ClampScrollOffsetY(offset int) int {
	if offset < 0 {
		return 0
	}
	max := b.MaxScrollY()
	if offset > max {
		return max
	}
	return offset
}

// WithClampedScrollOffset returns a new BoxModel with scroll offsets clamped to valid bounds.
func (b BoxModel) WithClampedScrollOffset() BoxModel {
	result := b
	result.ScrollOffsetX = b.ClampScrollOffsetX(b.ScrollOffsetX)
	result.ScrollOffsetY = b.ClampScrollOffsetY(b.ScrollOffsetY)
	return result
}

// VisibleContentRect returns the visible portion of the virtual content.
// The rect is in virtual content coordinates (not screen coordinates).
// For non-scrollable boxes, this returns a rect starting at (0,0).
func (b BoxModel) VisibleContentRect() terma.Rect {
	return terma.Rect{
		X:      b.ScrollOffsetX,
		Y:      b.ScrollOffsetY,
		Width:  b.ContentWidth,
		Height: b.ContentHeight,
	}
}

// VirtualContentRect returns the full virtual content area.
// For non-scrollable boxes, this equals the content dimensions.
func (b BoxModel) VirtualContentRect() terma.Rect {
	return terma.Rect{
		X:      0,
		Y:      0,
		Width:  b.EffectiveVirtualWidth(),
		Height: b.EffectiveVirtualHeight(),
	}
}

// --- Builder-style methods for convenience ---

// WithContent returns a new BoxModel with the specified content dimensions.
// Negative values are clamped to 0, since content dimensions often come from
// layout calculations that can legitimately underflow (e.g., when the terminal
// is resized too small).
func (b BoxModel) WithContent(width, height int) BoxModel {
	result := b
	result.ContentWidth = max(0, width)
	result.ContentHeight = max(0, height)
	result.Validate()
	return result
}

// WithPadding returns a new BoxModel with the specified padding.
// Panics if any padding value is negative.
func (b BoxModel) WithPadding(padding terma.EdgeInsets) BoxModel {
	result := b
	result.Padding = padding
	result.Validate()
	return result
}

// WithBorder returns a new BoxModel with the specified border.
// Panics if any border value is negative.
func (b BoxModel) WithBorder(border terma.EdgeInsets) BoxModel {
	result := b
	result.Border = border
	result.Validate()
	return result
}

// WithMargin returns a new BoxModel with the specified margin.
// Panics if any margin value is negative.
func (b BoxModel) WithMargin(margin terma.EdgeInsets) BoxModel {
	result := b
	result.Margin = margin
	result.Validate()
	return result
}

// WithMinContent returns a new BoxModel with min content constraints.
// Panics if min exceeds max (when both are set).
func (b BoxModel) WithMinContent(minWidth, minHeight int) BoxModel {
	result := b
	result.MinContentWidth = minWidth
	result.MinContentHeight = minHeight
	result.Validate()
	return result
}

// WithMaxContent returns a new BoxModel with max content constraints.
// Panics if max is less than min (when both are set).
func (b BoxModel) WithMaxContent(maxWidth, maxHeight int) BoxModel {
	result := b
	result.MaxContentWidth = maxWidth
	result.MaxContentHeight = maxHeight
	result.Validate()
	return result
}

// WithVirtualSize returns a new BoxModel with virtual content dimensions for scrolling.
// Negative values are clamped to 0, since virtual dimensions may come from
// layout calculations that can legitimately underflow.
func (b BoxModel) WithVirtualSize(virtualWidth, virtualHeight int) BoxModel {
	result := b
	result.VirtualWidth = max(0, virtualWidth)
	result.VirtualHeight = max(0, virtualHeight)
	result.Validate()
	return result
}

// WithScrollOffset returns a new BoxModel with the specified scroll offset.
func (b BoxModel) WithScrollOffset(offsetX, offsetY int) BoxModel {
	result := b
	result.ScrollOffsetX = offsetX
	result.ScrollOffsetY = offsetY
	return result
}

// WithScrollbarWidth returns a new BoxModel with the specified vertical scrollbar width.
// Panics if width is negative.
func (b BoxModel) WithScrollbarWidth(width int) BoxModel {
	result := b
	result.ScrollbarWidth = width
	result.Validate()
	return result
}

// WithScrollbarHeight returns a new BoxModel with the specified horizontal scrollbar height.
// Panics if height is negative.
func (b BoxModel) WithScrollbarHeight(height int) BoxModel {
	result := b
	result.ScrollbarHeight = height
	result.Validate()
	return result
}

// WithScrollbars returns a new BoxModel with the specified scrollbar dimensions.
// Panics if width or height is negative.
func (b BoxModel) WithScrollbars(width, height int) BoxModel {
	result := b
	result.ScrollbarWidth = width
	result.ScrollbarHeight = height
	result.Validate()
	return result
}
