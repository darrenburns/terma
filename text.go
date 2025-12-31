package terma

import "strings"

// Text is a leaf widget that displays text content.
type Text struct {
	ID      string     // Optional unique identifier for the widget
	Content string     // Plain text (used if Spans is empty)
	Spans   []Span     // Rich text segments (takes precedence if non-empty)
	Width   Dimension  // Optional width (zero value = auto)
	Height  Dimension  // Optional height (zero value = auto)
	Style   Style      // Optional styling (colors, inherited by spans)
	Click   func()     // Optional callback invoked when clicked
	Hover   func(bool) // Optional callback invoked when hover state changes
}

// Build returns itself as Text is a leaf widget.
func (t Text) Build(ctx BuildContext) Widget {
	return t
}

// Key returns the text widget's unique identifier.
// Implements the Keyed interface.
func (t Text) Key() string {
	return t.ID
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (t Text) OnClick() {
	if t.Click != nil {
		t.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (t Text) OnHover(hovered bool) {
	if t.Hover != nil {
		t.Hover(hovered)
	}
}

// GetDimensions returns the width and height dimension preferences.
func (t Text) GetDimensions() (width, height Dimension) {
	return t.Width, t.Height
}

// GetStyle returns the style of the text widget.
func (t Text) GetStyle() Style {
	return t.Style
}

// textContent returns the effective text content.
// If Spans is non-empty, concatenates all span text; otherwise returns Content.
func (t Text) textContent() string {
	if len(t.Spans) > 0 {
		var sb strings.Builder
		for _, span := range t.Spans {
			sb.WriteString(span.Text)
		}
		return sb.String()
	}
	return t.Content
}

// Layout computes the size of the text widget.
func (t Text) Layout(constraints Constraints) Size {
	// Calculate natural size from content (spans or plain text)
	content := t.textContent()
	lines := strings.Split(content, "\n")
	naturalHeight := len(lines)
	naturalWidth := 0
	for _, line := range lines {
		if len(line) > naturalWidth {
			naturalWidth = len(line)
		}
	}

	// Determine width based on dimension type
	var width int
	switch {
	case t.Width.IsCells():
		width = t.Width.CellsValue()
	case t.Width.IsFr():
		// Fr dimensions use the constraint max (parent allocates space)
		width = constraints.MaxWidth
	default: // Auto
		width = naturalWidth
	}

	// Determine height based on dimension type
	var height int
	switch {
	case t.Height.IsCells():
		height = t.Height.CellsValue()
	case t.Height.IsFr():
		// Fr dimensions use the constraint max (parent allocates space)
		height = constraints.MaxHeight
	default: // Auto
		height = naturalHeight
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

// Render draws the text to the render context.
func (t Text) Render(ctx *RenderContext) {
	if len(t.Spans) > 0 {
		t.renderSpans(ctx)
	} else {
		t.renderPlain(ctx)
	}
}

// renderPlain renders plain text content.
func (t Text) renderPlain(ctx *RenderContext) {
	lines := strings.Split(t.Content, "\n")
	for i := 0; i < ctx.Height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		// Truncate line if it exceeds width
		if len(line) > ctx.Width {
			line = line[:ctx.Width]
		}
		// Pad line to fill the full width (for background colors)
		if len(line) < ctx.Width {
			line = line + strings.Repeat(" ", ctx.Width-len(line))
		}
		ctx.DrawStyledText(0, i, line, t.Style)
	}
}

// renderSpans renders rich text with multiple styled spans.
func (t Text) renderSpans(ctx *RenderContext) {
	x, y := 0, 0

	for _, span := range t.Spans {
		// Handle newlines within span text
		parts := strings.Split(span.Text, "\n")
		for partIdx, part := range parts {
			// Move to next line if this isn't the first part (after a newline)
			if partIdx > 0 {
				// Pad remainder of current line with spaces for background
				if x < ctx.Width && t.Style.BackgroundColor != DefaultColor {
					padding := strings.Repeat(" ", ctx.Width-x)
					ctx.DrawStyledText(x, y, padding, t.Style)
				}
				x = 0
				y++
				if y >= ctx.Height {
					return
				}
			}

			// Skip if we've run out of horizontal space
			if x >= ctx.Width {
				continue
			}

			// Draw this part of the span
			if len(part) > 0 {
				// Create a span for just this part
				partSpan := Span{Text: part, Style: span.Style}
				ctx.DrawSpan(x, y, partSpan, t.Style)
				x += len(part)
			}
		}
	}

	// Pad remainder of last line with spaces for background
	if x < ctx.Width && t.Style.BackgroundColor != DefaultColor {
		padding := strings.Repeat(" ", ctx.Width-x)
		ctx.DrawStyledText(x, y, padding, t.Style)
	}

	// Fill remaining lines with spaces for background
	for row := y + 1; row < ctx.Height; row++ {
		if t.Style.BackgroundColor != DefaultColor {
			padding := strings.Repeat(" ", ctx.Width)
			ctx.DrawStyledText(0, row, padding, t.Style)
		}
	}
}
