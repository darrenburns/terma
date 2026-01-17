package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
	"terma/layout"
)

// WrapMode defines how text should wrap within available width.
type WrapMode int

const (
	// WrapNone disables wrapping - text is truncated if too long.
	WrapNone WrapMode = iota
	// WrapSoft breaks at word boundaries (spaces), only breaking words if necessary (default).
	WrapSoft
	// WrapHard breaks at exact character boundary when line exceeds width.
	WrapHard
)

// Text is a leaf widget that displays text content.
type Text struct {
	ID      string     // Optional unique identifier for the widget
	Content string     // Plain text (used if Spans is empty)
	Spans   []Span     // Rich text segments (takes precedence if non-empty)
	Width   Dimension  // Optional width (zero value = auto)
	Height  Dimension  // Optional height (zero value = auto)
	Wrap    WrapMode   // Wrapping mode (default = WrapNone)
	Style   Style      // Optional styling (colors, inherited by spans)
	Click   func()     // Optional callback invoked when clicked
	Hover   func(bool) // Optional callback invoked when hover state changes
}

// Build returns itself as Text is a leaf widget.
func (t Text) Build(ctx BuildContext) Widget {
	return t
}

// WidgetID returns the text widget's unique identifier.
// Implements the Identifiable interface.
func (t Text) WidgetID() string {
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

// BuildLayoutNode builds a layout node for this Text widget.
// Implements the LayoutNodeBuilder interface.
func (t Text) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	// Get the text content (spans concatenated or plain content)
	content := t.textContent()

	minWidth, maxWidth := dimensionToMinMax(t.Width)
	minHeight, maxHeight := dimensionToMinMax(t.Height)

	return &layout.TextNode{
		Content:   content,
		Wrap:      toLayoutWrapMode(t.Wrap),
		Padding:   toLayoutEdgeInsets(t.Style.Padding),
		Border:    borderToEdgeInsets(t.Style.Border),
		Margin:    toLayoutEdgeInsets(t.Style.Margin),
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
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

// wrapText wraps the given text content to fit within maxWidth based on the wrap mode.
func wrapText(content string, maxWidth int, mode WrapMode) []string {
	if maxWidth <= 0 || mode == WrapNone {
		return strings.Split(content, "\n")
	}

	inputLines := strings.Split(content, "\n")
	var result []string

	for _, line := range inputLines {
		lineWidth := ansi.StringWidth(line)

		// If line fits, add as-is
		if lineWidth <= maxWidth {
			result = append(result, line)
			continue
		}

		switch mode {
		case WrapHard:
			// Hard wrap: break at exact character boundary using Truncate
			remaining := line
			for len(remaining) > 0 {
				if ansi.StringWidth(remaining) <= maxWidth {
					result = append(result, remaining)
					break
				}
				chunk := ansi.Truncate(remaining, maxWidth, "")
				result = append(result, chunk)
				remaining = remaining[len(chunk):]
			}
		case WrapSoft:
			// Soft wrap: break at word boundaries
			wrapped := ansi.Wordwrap(line, maxWidth, "")
			wrappedLines := strings.Split(wrapped, "\n")
			for _, wl := range wrappedLines {
				if ansi.StringWidth(wl) > maxWidth {
					// Word longer than maxWidth, hard-break it
					remaining := wl
					for len(remaining) > 0 {
						if ansi.StringWidth(remaining) <= maxWidth {
							result = append(result, remaining)
							break
						}
						chunk := ansi.Truncate(remaining, maxWidth, "")
						result = append(result, chunk)
						remaining = remaining[len(chunk):]
					}
				} else {
					result = append(result, wl)
				}
			}
		}
	}

	return result
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
	// Start with the full style, then ensure foreground color has a default
	style := t.Style
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = ctx.buildContext.Theme().Text
	}
	drawStyle := style
	if drawStyle.BackgroundColor != nil && drawStyle.BackgroundColor.IsSet() {
		drawStyle.BackgroundColor = nil
	}

	// Get lines with wrapping applied
	lines := wrapText(t.Content, ctx.Width, t.Wrap)

	// Check if we need to draw text and padding separately
	// (when strikethrough/underline is set but FillLine is false)
	hasLineDecoration := style.Strikethrough || style.Underline != UnderlineNone
	separatePadding := hasLineDecoration && !style.FillLine

	for i := 0; i < ctx.Height; i++ {
		var line string
		if i < len(lines) {
			line = lines[i]
		}
		// Truncate line if it exceeds width (fallback for WrapNone or edge cases)
		lineWidth := ansi.StringWidth(line)
		if lineWidth > ctx.Width {
			line = ansi.Truncate(line, ctx.Width, "")
			lineWidth = ctx.Width
		}

		if separatePadding && lineWidth < ctx.Width {
			// Draw text with full style (including strikethrough/underline)
			ctx.DrawStyledText(0, i, line, drawStyle)
			// Draw padding without strikethrough/underline
			paddingStyle := drawStyle
			paddingStyle.Strikethrough = false
			paddingStyle.Underline = UnderlineNone
			padding := strings.Repeat(" ", ctx.Width-lineWidth)
			ctx.DrawStyledText(lineWidth, i, padding, paddingStyle)
		} else {
			// Pad line to fill the full width (for background colors)
			if lineWidth < ctx.Width {
				line = line + strings.Repeat(" ", ctx.Width-lineWidth)
			}
			ctx.DrawStyledText(0, i, line, drawStyle)
		}
	}
}

// renderSpans renders rich text with multiple styled spans.
func (t Text) renderSpans(ctx *RenderContext) {
	// Start with the full style, then ensure foreground color has a default
	baseStyle := t.Style
	if baseStyle.ForegroundColor == nil || !baseStyle.ForegroundColor.IsSet() {
		baseStyle.ForegroundColor = ctx.buildContext.Theme().Text
	}
	drawBaseStyle := baseStyle
	if drawBaseStyle.BackgroundColor != nil && drawBaseStyle.BackgroundColor.IsSet() {
		drawBaseStyle.BackgroundColor = nil
	}

	x, y := 0, 0

	for _, span := range t.Spans {
		parts := strings.Split(span.Text, "\n")
		for partIdx, part := range parts {
			// Handle explicit newline
			if partIdx > 0 {
				x = 0
				y++
				if y >= ctx.Height {
					return
				}
			}

			if len(part) == 0 {
				continue
			}

			// Process this part with wrapping
			remaining := part
			for len(remaining) > 0 {
				if y >= ctx.Height {
					return
				}

				availableWidth := ctx.Width - x
				if availableWidth <= 0 {
					x = 0
					y++
					availableWidth = ctx.Width
					if y >= ctx.Height {
						return
					}
				}

				partWidth := ansi.StringWidth(remaining)

				// If it fits or no wrapping, draw and continue
				if partWidth <= availableWidth || t.Wrap == WrapNone {
					chunk := remaining
					if partWidth > availableWidth {
						chunk = ansi.Truncate(remaining, availableWidth, "")
					}
					if len(chunk) > 0 {
						partSpan := Span{Text: chunk, Style: span.Style}
						ctx.DrawSpan(x, y, partSpan, drawBaseStyle)
						x += ansi.StringWidth(chunk)
					}
					break
				}

				// Need to wrap - find break point
				chunk, rest := t.findWrapPoint(remaining, availableWidth)

				if len(chunk) > 0 {
					partSpan := Span{Text: chunk, Style: span.Style}
					ctx.DrawSpan(x, y, partSpan, drawBaseStyle)
				}

				remaining = rest
				if len(remaining) > 0 {
					x = 0
					y++
				}
			}
		}
	}
}

// findWrapPoint finds where to break text for wrapping, returning the chunk to render
// and the remaining text.
func (t Text) findWrapPoint(text string, availableWidth int) (chunk, remaining string) {
	if t.Wrap == WrapHard {
		chunk = ansi.Truncate(text, availableWidth, "")
		remaining = text[len(chunk):]
		return
	}

	// Soft wrap: find last space within available width
	chunk = ansi.Truncate(text, availableWidth, "")
	lastSpace := strings.LastIndex(chunk, " ")

	if lastSpace > 0 {
		chunk = chunk[:lastSpace]
		remaining = strings.TrimPrefix(text[lastSpace:], " ")
	} else {
		// No space found, must break the word (fallback to hard wrap)
		remaining = text[len(chunk):]
	}
	return
}
