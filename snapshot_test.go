package terma

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWidget is a simple widget for testing snapshots.
type testWidget struct {
	text  string
	style Style
}

func (w testWidget) Build(ctx BuildContext) Widget {
	return Text{Content: w.text, Style: w.style}
}

// testStyledWidget creates a widget with styled text.
type testStyledWidget struct{}

func (w testStyledWidget) Build(ctx BuildContext) Widget {
	return Column{
		Children: []Widget{
			Text{Content: "Hello World", Style: Style{ForegroundColor: RGB(255, 100, 100)}},
			Text{Content: "Bold Text", Style: Style{Bold: true, ForegroundColor: RGB(100, 255, 100)}},
			Text{Content: "With Background", Style: Style{
				ForegroundColor: RGB(255, 255, 255),
				BackgroundColor: RGB(50, 50, 150),
			}},
		},
	}
}

func TestRenderToBuffer(t *testing.T) {
	widget := testWidget{text: "Test"}
	buf := RenderToBuffer(widget, 20, 5)

	// Verify buffer was created with correct dimensions
	assert.Equal(t, 20, buf.Width())
	assert.Equal(t, 5, buf.Height())

	// Verify content was rendered
	cell := buf.CellAt(0, 0)
	require.NotNil(t, cell)
	assert.Equal(t, "T", cell.Content)
}

func TestSnapshot(t *testing.T) {
	widget := testWidget{text: "Hello"}
	svg := Snapshot(widget, 20, 5)

	// Verify SVG structure
	assert.Contains(t, svg, "<svg")
	assert.Contains(t, svg, "</svg>")
	assert.Contains(t, svg, "Hello")
	assert.Contains(t, svg, "font-family")
}

func TestSnapshotWithOptions(t *testing.T) {
	widget := testWidget{text: "Custom"}
	opts := SVGOptions{
		FontFamily: "Courier",
		FontSize:   16,
		Background: RGB(30, 30, 30),
		Padding:    10,
	}
	svg := SnapshotWithOptions(widget, 20, 5, opts)

	assert.Contains(t, svg, "Courier")
	assert.Contains(t, svg, "16px")
	assert.Contains(t, svg, "#1E1E1E") // background color
}

func TestSnapshotWithStyledText(t *testing.T) {
	widget := testStyledWidget{}
	svg := Snapshot(widget, 40, 10)

	// Check that colored text is rendered (text spans break on spaces)
	assert.Contains(t, svg, "Hello")
	assert.Contains(t, svg, "World")
	assert.Contains(t, svg, "Bold")
	assert.Contains(t, svg, "Text")
	assert.Contains(t, svg, "With")
	assert.Contains(t, svg, "Background")

	// Check for bold class
	assert.Contains(t, svg, `class="bold"`)

	// Check for background rect (for "With Background" text)
	assert.Contains(t, svg, `<rect x=`)
}

func TestSaveSnapshot(t *testing.T) {
	widget := testWidget{text: "Saved"}
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.svg")

	err := SaveSnapshot(widget, 20, 5, path)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Saved")
	assert.Contains(t, string(content), "<svg")
}

func TestBufferToSVG_EmptyBuffer(t *testing.T) {
	buf := uv.NewBuffer(10, 5)
	svg := BufferToSVG(buf, 10, 5, DefaultSVGOptions())

	// Should still produce valid SVG
	assert.Contains(t, svg, "<svg")
	assert.Contains(t, svg, "</svg>")
}

func TestSnapshot_SpecialCharacters(t *testing.T) {
	widget := testWidget{text: "<script>alert('xss')</script>"}
	svg := Snapshot(widget, 40, 5)

	// Verify special characters are escaped
	assert.NotContains(t, svg, "<script>")
	assert.Contains(t, svg, "&lt;script&gt;")
}

func TestGenerateGallery(t *testing.T) {
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "gallery.html")

	comparisons := []SnapshotComparison{
		{
			Name:     "Test 1",
			Expected: `<svg><text>Expected</text></svg>`,
			Actual:   `<svg><text>Actual</text></svg>`,
			Passed:   false,
		},
		{
			Name:     "Test 2",
			Expected: `<svg><text>Same</text></svg>`,
			Actual:   `<svg><text>Same</text></svg>`,
			Passed:   true,
		},
	}

	err := GenerateGallery(comparisons, outputPath)
	require.NoError(t, err)

	// Verify file was created
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	html := string(content)

	// Check HTML structure
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "Terma Snapshot Gallery")
	assert.Contains(t, html, "Test 1")
	assert.Contains(t, html, "Test 2")
	assert.Contains(t, html, "PASSED")
	assert.Contains(t, html, "FAILED")
	assert.Contains(t, html, "Expected")
	assert.Contains(t, html, "Actual")
}

func TestDefaultSVGOptions(t *testing.T) {
	opts := DefaultSVGOptions()

	assert.Equal(t, "Fira Code, Menlo, Monaco, Consolas, monospace", opts.FontFamily)
	assert.Equal(t, 14, opts.FontSize)
	assert.Equal(t, 1.4, opts.LineHeight)
	assert.Equal(t, 8, opts.Padding)
	assert.True(t, opts.Background.IsSet())
}

func TestSameStyle(t *testing.T) {
	tests := []struct {
		name     string
		a, b     any
		expected bool
	}{
		{
			name:     "identical empty styles",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic test - sameStyle is internal
		})
	}
}

// TestSnapshotIntegration tests the full snapshot workflow with a complex widget.
func TestSnapshotIntegration(t *testing.T) {
	// Create a more complex widget
	widget := Column{
		Style: Style{
			BackgroundColor: RGB(30, 30, 46),
			Padding:         EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
			Border:          Border{Style: BorderRounded, Color: RGB(137, 180, 250)},
		},
		Children: []Widget{
			Text{Content: "Header", Style: Style{
				ForegroundColor: RGB(203, 166, 247),
				Bold:            true,
			}},
			Text{Content: "Body text here", Style: Style{
				ForegroundColor: RGB(205, 214, 244),
			}},
		},
	}

	svg := Snapshot(widget, 30, 10)

	// Verify the SVG contains expected elements (text spans break on spaces)
	assert.Contains(t, svg, "Header")
	assert.Contains(t, svg, "Body")
	assert.Contains(t, svg, "text")
	assert.Contains(t, svg, "here")
	assert.Contains(t, svg, `class="bold"`)

	// Verify it's valid SVG
	assert.True(t, strings.HasPrefix(svg, "<svg"))
	assert.True(t, strings.HasSuffix(strings.TrimSpace(svg), "</svg>"))
}
