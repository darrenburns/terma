package terma

import (
	"strings"

	"terma/layout"
)

// progressBarChars contains Unicode block characters for smooth progress rendering.
// Each character represents a fraction of a full cell width.
// Index 0 = empty (0/8), Index 8 = full (8/8).
var progressBarChars = []string{
	" ", // 0/8 - empty
	"▏", // 1/8 - left one eighth block
	"▎", // 2/8 - left one quarter block
	"▍", // 3/8 - left three eighths block
	"▌", // 4/8 - left half block
	"▋", // 5/8 - left five eighths block
	"▊", // 6/8 - left three quarters block
	"▉", // 7/8 - left seven eighths block
	"█", // 8/8 - full block
}

// ProgressBar displays a horizontal progress indicator.
// Progress is specified as a float64 from 0.0 (empty) to 1.0 (full).
// Uses Unicode block characters for smooth sub-character rendering.
//
// Example:
//
//	ProgressBar{
//	    Progress:    0.65,
//	    Width:       Cells(20),
//	    FilledColor: ctx.Theme().Primary,
//	}
type ProgressBar struct {
	ID string // Optional unique identifier

	// Core fields
	Progress float64 // 0.0 to 1.0

	// Dimensions
	Width  Dimension // Default: Flex(1)
	Height Dimension // Default: Cells(1)

	// Styling
	Style         Style // General styling (padding, margins, border)
	FilledColor   Color // Color of filled portion (default: theme Primary)
	UnfilledColor Color // Color of unfilled portion (default: theme Surface)
	MinMaxDimensions
}

// Build returns itself as ProgressBar is a leaf widget.
func (p ProgressBar) Build(ctx BuildContext) Widget {
	return p
}

// WidgetID returns the progress bar's unique identifier.
// Implements the Identifiable interface.
func (p ProgressBar) WidgetID() string {
	return p.ID
}

// GetContentDimensions returns the width and height dimension preferences.
// Width defaults to Flex(1), Height defaults to Cells(1).
func (p ProgressBar) GetContentDimensions() (width, height Dimension) {
	w, h := p.Width, p.Height
	if w.IsUnset() {
		w = Flex(1)
	}
	if h.IsUnset() {
		h = Cells(1)
	}
	return w, h
}

// GetStyle returns the style of the progress bar.
func (p ProgressBar) GetStyle() Style {
	return p.Style
}

// BuildLayoutNode builds a layout node for this ProgressBar widget.
func (p ProgressBar) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	padding := toLayoutEdgeInsets(p.Style.Padding)
	border := borderToEdgeInsets(p.Style.Border)

	return &layout.BoxNode{
		Padding:   padding,
		Border:    border,
		Margin:    toLayoutEdgeInsets(p.Style.Margin),
	}
}

// Render draws the progress bar to the render context.
func (p ProgressBar) Render(ctx *RenderContext) {
	if ctx.Width <= 0 || ctx.Height <= 0 {
		return
	}

	// Clamp progress to valid range
	progress := p.Progress
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	// Determine colors (use theme defaults if not specified)
	filledColor := p.FilledColor
	if !filledColor.IsSet() {
		filledColor = ctx.buildContext.Theme().Primary
	}
	unfilledColor := p.UnfilledColor
	if !unfilledColor.IsSet() {
		unfilledColor = ctx.buildContext.Theme().Surface
	}

	// Calculate filled width with sub-character precision
	// Total width in "eighths" (each cell has 8 sub-positions)
	totalEighths := ctx.Width * 8
	filledEighths := int(progress * float64(totalEighths))

	// How many full cells are completely filled
	fullCells := filledEighths / 8
	// Remaining eighths for the partial cell (0-7)
	partialEighths := filledEighths % 8

	// Build the filled portion directly
	var sb strings.Builder
	for i := 0; i < fullCells; i++ {
		sb.WriteString(progressBarChars[8]) // █
	}
	if partialEighths > 0 && fullCells < ctx.Width {
		sb.WriteString(progressBarChars[partialEighths])
	}
	filledText := sb.String()

	// Calculate filled width in cells
	filledWidth := fullCells
	if partialEighths > 0 && fullCells < ctx.Width {
		filledWidth++
	}

	// Render each row (progress bars are typically 1 row, but support multi-row)
	for row := 0; row < ctx.Height; row++ {
		// Draw filled part
		if filledWidth > 0 {
			ctx.DrawStyledText(0, row, filledText, Style{
				ForegroundColor: filledColor,
				BackgroundColor: unfilledColor,
			})
		}

		// Draw unfilled part
		if filledWidth < ctx.Width {
			unfilledText := strings.Repeat(" ", ctx.Width-filledWidth)
			ctx.DrawStyledText(filledWidth, row, unfilledText, Style{
				BackgroundColor: unfilledColor,
			})
		}
	}
}
