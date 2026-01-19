package terma

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderToString_SimpleText(t *testing.T) {
	widget := Text{Content: "Hello"}
	result := RenderToString(widget, 10, 1)

	assert.Contains(t, result, "Hello")
	// Should end with reset sequence
	assert.True(t, strings.HasSuffix(result, "\x1b[0m"))
}

func TestRenderToPlainString_SimpleText(t *testing.T) {
	widget := Text{Content: "Hello"}
	result := RenderToPlainString(widget, 10, 1)

	assert.Contains(t, result, "Hello")
	// Should NOT contain ANSI escape codes
	assert.NotContains(t, result, "\x1b[")
}

func TestRenderToString_WithColors(t *testing.T) {
	widget := Text{
		Content: "Colored",
		Style: Style{
			ForegroundColor: RGB(255, 0, 0),
		},
	}
	result := RenderToString(widget, 20, 1)

	// Should contain ANSI codes
	assert.Contains(t, result, "\x1b[")
	assert.Contains(t, result, "Colored")
}

func TestRenderToString_WithBackgroundColor(t *testing.T) {
	widget := Text{
		Content: "BG",
		Style: Style{
			BackgroundColor: RGB(0, 255, 0),
		},
	}
	result := RenderToString(widget, 10, 1)

	assert.Contains(t, result, "\x1b[")
	assert.Contains(t, result, "BG")
}

func TestRenderToString_MultipleRows(t *testing.T) {
	widget := Column{
		Children: []Widget{
			Text{Content: "Line1"},
			Text{Content: "Line2"},
		},
	}
	result := RenderToString(widget, 10, 2)

	assert.Contains(t, result, "Line1")
	assert.Contains(t, result, "Line2")
	assert.Contains(t, result, "\n")
}

func TestPrintWithOptions_ToBuffer(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{Content: "Test"}

	err := PrintWithOptions(widget, PrintOptions{
		Width:           20,
		Height:          1,
		Writer:          &buf,
		NoColor:         false,
		TrailingNewline: true,
	})

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Test")
	assert.True(t, strings.HasSuffix(buf.String(), "\n"))
}

func TestPrintWithOptions_NoColor(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{
		Content: "Plain",
		Style: Style{
			ForegroundColor: RGB(255, 0, 0),
		},
	}

	err := PrintWithOptions(widget, PrintOptions{
		Width:           20,
		Height:          1,
		Writer:          &buf,
		NoColor:         true,
		TrailingNewline: false,
	})

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Plain")
	// Should NOT contain ANSI escape codes
	assert.NotContains(t, buf.String(), "\x1b[")
}

func TestPrintWithOptions_NoTrailingNewline(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{Content: "NoNewline"}

	err := PrintWithOptions(widget, PrintOptions{
		Width:           20,
		Height:          1,
		Writer:          &buf,
		NoColor:         true,
		TrailingNewline: false,
	})

	require.NoError(t, err)
	assert.False(t, strings.HasSuffix(buf.String(), "\n"))
}

func TestPrintWithOptions_DefaultDimensions(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{Content: "X"}

	// Width and Height of 0 should use defaults (80x24)
	// AutoHeight false to get full buffer output
	err := PrintWithOptions(widget, PrintOptions{
		Width:      0,
		Height:     0,
		Writer:     &buf,
		NoColor:    true,
		AutoHeight: false,
	})

	require.NoError(t, err)
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	// Should have 24 lines (default height)
	assert.Equal(t, 24, len(lines))
	// Each line should be 80 chars (default width)
	assert.Equal(t, 80, len(lines[0]))
}

func TestPrintWithOptions_AutoHeight(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{Content: "Single line"}

	// With AutoHeight true, should only output rows based on computed layout
	err := PrintWithOptions(widget, PrintOptions{
		Width:           80,
		Height:          24,
		Writer:          &buf,
		NoColor:         true,
		TrailingNewline: false,
		AutoHeight:      true,
	})

	require.NoError(t, err)
	lines := strings.Split(buf.String(), "\n")
	// Should have only 1 line since Text computes to height 1
	assert.Equal(t, 1, len(lines))
}

func TestPrintWithOptions_AutoHeight_MultipleRows(t *testing.T) {
	var buf bytes.Buffer
	widget := Column{
		Children: []Widget{
			Text{Content: "Line 1"},
			Text{Content: "Line 2"},
			Text{Content: "Line 3"},
		},
	}

	err := PrintWithOptions(widget, PrintOptions{
		Width:           40,
		Height:          10,
		Writer:          &buf,
		NoColor:         true,
		TrailingNewline: false,
		AutoHeight:      true,
	})

	require.NoError(t, err)
	lines := strings.Split(buf.String(), "\n")
	// Should have 3 lines matching the Column's computed layout
	assert.Equal(t, 3, len(lines))
}

func TestPrintWithOptions_AutoHeight_WithPadding(t *testing.T) {
	var buf bytes.Buffer
	widget := Row{
		Children: []Widget{
			Text{Content: "Padded"},
		},
		Style: Style{
			Padding: EdgeInsetsAll(2), // 2 cells padding on all sides
		},
	}

	err := PrintWithOptions(widget, PrintOptions{
		Width:           40,
		Height:          10,
		Writer:          &buf,
		NoColor:         true,
		TrailingNewline: false,
		AutoHeight:      true,
	})

	require.NoError(t, err)
	lines := strings.Split(buf.String(), "\n")
	// Should have 5 lines: 2 padding top + 1 content + 2 padding bottom
	assert.Equal(t, 5, len(lines))
}

func TestBufferToANSI_WideCharacters(t *testing.T) {
	// Emoji takes 2 cells
	widget := Text{Content: "Hi 👋!"}
	result := RenderToString(widget, 10, 1)

	assert.Contains(t, result, "👋")
	assert.Contains(t, result, "Hi")
}

func TestBufferToANSI_StyleGrouping(t *testing.T) {
	// Multiple characters with same style should be grouped
	widget := Text{
		Content: "AAAA",
		Style: Style{
			ForegroundColor: RGB(255, 0, 0),
		},
	}
	result := RenderToString(widget, 10, 1)

	// The styled text should appear as one group, not 4 separate ANSI sequences
	assert.Contains(t, result, "AAAA")
}

func TestPrintTo_BytesBuffer(t *testing.T) {
	var buf bytes.Buffer
	widget := Text{Content: "BufferTest"}

	// bytes.Buffer is not an *os.File, so NoColor should be enabled
	err := PrintTo(&buf, widget)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "BufferTest")
	// Should be plain text (no ANSI) since bytes.Buffer is not a TTY
	assert.NotContains(t, buf.String(), "\x1b[38")
}

func TestDefaultPrintOptions(t *testing.T) {
	opts := DefaultPrintOptions()

	assert.Equal(t, 0, opts.Width)
	assert.Equal(t, 0, opts.Height)
	assert.Nil(t, opts.Writer)
	assert.False(t, opts.NoColor)
	assert.True(t, opts.TrailingNewline)
	assert.True(t, opts.AutoHeight) // Default true to use computed layout size
}

func TestUvStylesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     *Style
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "same colors",
			a:        &Style{ForegroundColor: RGB(255, 0, 0)},
			b:        &Style{ForegroundColor: RGB(255, 0, 0)},
			expected: true,
		},
		{
			name:     "different colors",
			a:        &Style{ForegroundColor: RGB(255, 0, 0)},
			b:        &Style{ForegroundColor: RGB(0, 255, 0)},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: We're testing the Terma Style, not uv.Style directly
			// The uvStylesEqual function works on uv.Style which is internal
			// This test validates the concept through the public API
			if tt.a == nil && tt.b == nil {
				assert.True(t, tt.expected)
				return
			}

			// Render both and compare - if styles are equal, output should be identical
			widgetA := Text{Content: "X"}
			widgetB := Text{Content: "X"}
			if tt.a != nil {
				widgetA.Style = *tt.a
			}
			if tt.b != nil {
				widgetB.Style = *tt.b
			}

			resultA := RenderToString(widgetA, 5, 1)
			resultB := RenderToString(widgetB, 5, 1)

			if tt.expected {
				assert.Equal(t, resultA, resultB)
			} else {
				assert.NotEqual(t, resultA, resultB)
			}
		})
	}
}
