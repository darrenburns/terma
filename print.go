package terma

import (
	"image/color"
	"io"
	"os"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/term"
)

// PrintOptions configures widget printing behavior.
type PrintOptions struct {
	Width           int       // 0 = auto-detect or default 80
	Height          int       // 0 = auto-detect or default 24
	Writer          io.Writer // nil = os.Stdout
	NoColor         bool      // Force plain text output (no ANSI)
	TrailingNewline bool      // Add newline after output (default: true)
	AutoHeight      bool      // Use widget's computed height instead of buffer height (default: true)
}

// DefaultPrintOptions returns sensible defaults for printing.
func DefaultPrintOptions() PrintOptions {
	return PrintOptions{
		Width:           0,    // Auto-detect
		Height:          0,    // Auto-detect
		Writer:          nil,  // os.Stdout
		NoColor:         false,
		TrailingNewline: true,
		AutoHeight:      true, // Use computed layout height
	}
}

// Print renders a widget to stdout with ANSI styling.
// Auto-detects terminal dimensions or falls back to 80x24.
// Output height is determined by the widget's computed layout.
func Print(widget Widget) error {
	return PrintTo(os.Stdout, widget)
}

// PrintTo renders a widget to the specified writer with ANSI styling.
func PrintTo(w io.Writer, widget Widget) error {
	opts := DefaultPrintOptions()
	opts.Writer = w

	// Auto-detect TTY and dimensions
	if f, ok := w.(*os.File); ok {
		fd := f.Fd()
		if term.IsTerminal(fd) {
			width, height, err := term.GetSize(fd)
			if err == nil {
				opts.Width = width
				opts.Height = height
			}
		} else {
			opts.NoColor = true // Graceful degradation for pipes
		}
	} else {
		// Non-file writer: assume no color support
		opts.NoColor = true
	}

	return PrintWithOptions(widget, opts)
}

// PrintWithSize renders a widget to stdout at specific dimensions.
func PrintWithSize(widget Widget, width, height int) error {
	opts := DefaultPrintOptions()
	opts.Width = width
	opts.Height = height

	// Still detect TTY for color support
	if term.IsTerminal(os.Stdout.Fd()) {
		opts.NoColor = false
	} else {
		opts.NoColor = true
	}

	return PrintWithOptions(widget, opts)
}

// PrintWithOptions renders a widget using custom options.
func PrintWithOptions(widget Widget, opts PrintOptions) error {
	// Apply defaults
	if opts.Width <= 0 {
		opts.Width = 80
	}
	if opts.Height <= 0 {
		opts.Height = 24
	}
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}

	// Render to buffer and get computed layout size (border-box dimensions)
	buf, layoutWidth, layoutHeight := RenderToBufferWithSize(widget, opts.Width, opts.Height)

	// Determine output dimensions
	outputWidth := opts.Width
	outputHeight := opts.Height
	if opts.AutoHeight {
		outputWidth = layoutWidth
		outputHeight = layoutHeight
	}

	// Convert to output
	var output string
	if opts.NoColor {
		output = bufferToPlainText(buf, outputWidth, outputHeight)
	} else {
		output = BufferToANSI(buf, outputWidth, outputHeight)
	}

	if opts.TrailingNewline {
		output += "\n"
	}

	_, err := opts.Writer.Write([]byte(output))
	return err
}

// RenderToString renders a widget to an ANSI-styled string.
// Uses the widget's computed layout dimensions (border-box size).
func RenderToString(widget Widget, width, height int) string {
	buf, layoutWidth, layoutHeight := RenderToBufferWithSize(widget, width, height)
	return BufferToANSI(buf, layoutWidth, layoutHeight)
}

// RenderToPlainString renders a widget to a plain text string (no ANSI codes).
// Uses the widget's computed layout dimensions (border-box size).
func RenderToPlainString(widget Widget, width, height int) string {
	buf, layoutWidth, layoutHeight := RenderToBufferWithSize(widget, width, height)
	return bufferToPlainText(buf, layoutWidth, layoutHeight)
}

// BufferToANSI converts a rendered buffer to an ANSI-styled string.
func BufferToANSI(buf CellBuffer, width, height int) string {
	var sb strings.Builder
	sb.Grow(width * height * 2) // Pre-allocate estimate

	for y := 0; y < height; y++ {
		x := 0
		for x < width {
			cell := buf.CellAt(x, y)

			if cell == nil || cell.Content == "" {
				sb.WriteByte(' ')
				x++
				continue
			}

			// Collect consecutive cells with same style for efficiency
			var text strings.Builder
			baseStyle := cell.Style

			for x < width {
				c := buf.CellAt(x, y)
				if c == nil {
					break
				}
				if !uvStylesEqual(&c.Style, &baseStyle) {
					break
				}

				content := c.Content
				if content == "" {
					content = " "
				}
				text.WriteString(content)

				if c.Width > 1 {
					x += c.Width
				} else {
					x++
				}
			}

			// Apply style and write
			styled := baseStyle.Styled(text.String())
			sb.WriteString(styled)
		}

		if y < height-1 {
			sb.WriteByte('\n')
		}
	}

	// Ensure reset at end
	sb.WriteString("\x1b[0m")

	return sb.String()
}

// bufferToPlainText converts a rendered buffer to plain text (no ANSI codes).
func bufferToPlainText(buf CellBuffer, width, height int) string {
	var sb strings.Builder
	sb.Grow(width * height)

	for y := 0; y < height; y++ {
		x := 0
		for x < width {
			cell := buf.CellAt(x, y)

			if cell == nil || cell.Content == "" {
				sb.WriteByte(' ')
				x++
				continue
			}

			sb.WriteString(cell.Content)

			// Skip continuation cells for wide characters
			if cell.Width > 1 {
				x += cell.Width
			} else {
				x++
			}
		}

		if y < height-1 {
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}

// uvStylesEqual compares two uv.Style values for equality.
func uvStylesEqual(a, b *uv.Style) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Compare Fg, Bg colors and Attrs
	return uvColorsEqual(a.Fg, b.Fg) &&
		uvColorsEqual(a.Bg, b.Bg) &&
		a.Attrs == b.Attrs &&
		a.Underline == b.Underline
}

// uvColorsEqual compares two color.Color values for equality.
func uvColorsEqual(a, b color.Color) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	r1, g1, b1, a1 := a.RGBA()
	r2, g2, b2, a2 := b.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}
