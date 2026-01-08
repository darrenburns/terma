package layout

import (
	"fmt"

	"terma"
)

// BoxModel describes a rectangular region with CSS border-box semantics.
// Width and Height refer to the border-box (content + padding + border).
// Content dimensions are computed by subtracting padding and border from the border-box.
// Margin is added externally to get the margin-box (total allocated space).
//
// The model consists of four nested boxes:
//   - Content box: computed as border-box minus padding and border
//   - Padding box: border-box minus border
//   - Border box: the stored Width/Height dimensions
//   - Margin box: border-box plus margin (the outermost boundary)
type BoxModel struct {
	// Border-box dimensions (content + padding + border)
	Width  int
	Height int

	// Insets (padding and border shrink inward, margin expands outward)
	Padding terma.EdgeInsets
	Border  terma.EdgeInsets // Typically uniform (1,1,1,1) for borders
	Margin  terma.EdgeInsets

	// Scrolling (optional - zero values mean no scrolling)
	// VirtualWidth/Height represent the total scrollable content size.
	// When set, content can extend beyond the visible viewport.
	VirtualWidth  int // Total virtual content width (0 = same as computed ContentWidth)
	VirtualHeight int // Total virtual content height (0 = same as computed ContentHeight)
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
// Note: Negative margins are allowed (CSS behavior) for overlapping/pull effects.
func (b BoxModel) Validate() {
	// Border-box dimensions
	if b.Width < 0 {
		panic("BoxModel: Width cannot be negative")
	}
	if b.Height < 0 {
		panic("BoxModel: Height cannot be negative")
	}

	// Insets (padding and border must be non-negative, margin can be negative)
	validateEdgeInsets(b.Padding, "Padding")
	validateEdgeInsets(b.Border, "Border")
	// Margin is not validated - negative margins are allowed (CSS behavior)

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
}

// validateEdgeInsets checks that all EdgeInsets values are non-negative.
func validateEdgeInsets(e terma.EdgeInsets, name string) {
	if e.Top < 0 || e.Right < 0 || e.Bottom < 0 || e.Left < 0 {
		panic(fmt.Sprintf("BoxModel: %s cannot have negative values", name))
	}
}

// --- Box dimension methods ---

// ContentWidth computes the content width by subtracting padding and border from the border-box.
// Returns 0 if insets exceed the border-box width.
func (b BoxModel) ContentWidth() int {
	return max(0, b.Width-b.Padding.Horizontal()-b.Border.Horizontal())
}

// ContentHeight computes the content height by subtracting padding and border from the border-box.
// Returns 0 if insets exceed the border-box height.
func (b BoxModel) ContentHeight() int {
	return max(0, b.Height-b.Padding.Vertical()-b.Border.Vertical())
}

// PaddingBoxWidth returns the width of the padding box (border-box minus border).
func (b BoxModel) PaddingBoxWidth() int {
	return max(0, b.Width-b.Border.Horizontal())
}

// PaddingBoxHeight returns the height of the padding box (border-box minus border).
func (b BoxModel) PaddingBoxHeight() int {
	return max(0, b.Height-b.Border.Vertical())
}

// BorderBoxWidth returns the width of the border box.
// This is the stored Width value.
func (b BoxModel) BorderBoxWidth() int {
	return b.Width
}

// BorderBoxHeight returns the height of the border box.
// This is the stored Height value.
func (b BoxModel) BorderBoxHeight() int {
	return b.Height
}

// MarginBoxWidth returns the width of the margin box (border-box plus horizontal margin).
// This is the total width including margin.
func (b BoxModel) MarginBoxWidth() int {
	return b.Width + b.Margin.Horizontal()
}

// MarginBoxHeight returns the height of the margin box (border-box plus vertical margin).
// This is the total height including margin.
func (b BoxModel) MarginBoxHeight() int {
	return b.Height + b.Margin.Vertical()
}

// --- Rect-returning methods ---
// All rects are positioned relative to the margin box origin (0,0).

// ContentBox returns the computed content area as a Rect.
// Content dimensions are computed by subtracting padding and border from the border-box.
// Use for: laying out content within the available space.
// The position is relative to the margin box origin.
func (b BoxModel) ContentBox() terma.Rect {
	return terma.Rect{
		X:      b.Margin.Left + b.Border.Left + b.Padding.Left,
		Y:      b.Margin.Top + b.Border.Top + b.Padding.Top,
		Width:  b.ContentWidth(),
		Height: b.ContentHeight(),
	}
}

// UsableContentBox returns the content area available for child widgets.
// This subtracts space reserved for scrollbars from the computed content area:
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
	contentWidth := b.ContentWidth()
	if b.IsScrollableY() && b.ScrollbarWidth > 0 {
		return max(0, contentWidth-b.ScrollbarWidth)
	}
	return contentWidth
}

// usableContentHeight returns the content height available for child widgets.
// This accounts for horizontal scrollbar height when horizontal scrolling is enabled.
func (b BoxModel) usableContentHeight() int {
	contentHeight := b.ContentHeight()
	if b.IsScrollableX() && b.ScrollbarHeight > 0 {
		return max(0, contentHeight-b.ScrollbarHeight)
	}
	return contentHeight
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

// --- Scrolling methods ---

// EffectiveVirtualWidth returns the virtual content width.
// Returns computed ContentWidth() if VirtualWidth is not set (0).
func (b BoxModel) EffectiveVirtualWidth() int {
	if b.VirtualWidth > 0 {
		return b.VirtualWidth
	}
	return b.ContentWidth()
}

// EffectiveVirtualHeight returns the virtual content height.
// Returns computed ContentHeight() if VirtualHeight is not set (0).
func (b BoxModel) EffectiveVirtualHeight() int {
	if b.VirtualHeight > 0 {
		return b.VirtualHeight
	}
	return b.ContentHeight()
}

// IsScrollableX returns true if horizontal scrolling is possible.
// This is true when virtual width exceeds the computed content width.
func (b BoxModel) IsScrollableX() bool {
	return b.EffectiveVirtualWidth() > b.ContentWidth()
}

// IsScrollableY returns true if vertical scrolling is possible.
// This is true when virtual height exceeds the computed content height.
func (b BoxModel) IsScrollableY() bool {
	return b.EffectiveVirtualHeight() > b.ContentHeight()
}

// IsScrollable returns true if scrolling is possible in either direction.
func (b BoxModel) IsScrollable() bool {
	return b.IsScrollableX() || b.IsScrollableY()
}

// MaxScrollX returns the maximum valid horizontal scroll offset.
// Returns 0 if not scrollable.
func (b BoxModel) MaxScrollX() int {
	maxVal := b.EffectiveVirtualWidth() - b.ContentWidth()
	if maxVal < 0 {
		return 0
	}
	return maxVal
}

// MaxScrollY returns the maximum valid vertical scroll offset.
// Returns 0 if not scrollable.
func (b BoxModel) MaxScrollY() int {
	maxVal := b.EffectiveVirtualHeight() - b.ContentHeight()
	if maxVal < 0 {
		return 0
	}
	return maxVal
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
		Width:  b.ContentWidth(),
		Height: b.ContentHeight(),
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

// WithSize returns a new BoxModel with the specified border-box dimensions.
// Negative values are clamped to 0, since dimensions often come from
// layout calculations that can legitimately underflow (e.g., when the terminal
// is resized too small).
func (b BoxModel) WithSize(width, height int) BoxModel {
	result := b
	result.Width = max(0, width)
	result.Height = max(0, height)
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
