# Snapshot Testing

Terma includes built-in snapshot testing for capturing terminal UI as SVG images. This enables visual regression testing and documentation generation.

## Quick Start

```go
package myapp_test

import (
    "testing"
    "terma"
)

func TestMyWidget(t *testing.T) {
    widget := MyWidget{Title: "Hello"}

    // Generate SVG snapshot
    svg := terma.Snapshot(widget, 80, 24)

    // Save to file
    terma.SaveSnapshot(widget, 80, 24, "testdata/my_widget.svg")
}
```

## API Reference

### Generating Snapshots

```go
// Render widget to SVG with default options
svg := terma.Snapshot(widget, width, height)

// Render with custom options
opts := terma.SVGOptions{
    FontFamily: "Fira Code",
    FontSize:   16,
    Background: terma.RGB(30, 30, 46),
}
svg := terma.SnapshotWithOptions(widget, width, height, opts)

// Save directly to file
terma.SaveSnapshot(widget, 80, 24, "output.svg")
terma.SaveSnapshotWithOptions(widget, 80, 24, "output.svg", opts)
```

### SVGOptions

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `FontFamily` | `string` | `"Fira Code, Menlo, Monaco, Consolas, monospace"` | CSS font-family |
| `FontSize` | `int` | `14` | Font size in pixels |
| `LineHeight` | `float64` | `1.4` | Line height multiplier |
| `CellWidth` | `float64` | `fontSize * 0.6` | Width per character cell |
| `Background` | `Color` | `RGB(0,0,0)` | SVG background color |
| `Padding` | `int` | `8` | Padding around content |

### Low-Level API

```go
// Render to buffer for inspection
buf := terma.RenderToBuffer(widget, width, height)

// Access individual cells
cell := buf.CellAt(x, y)
fmt.Println(cell.Content, cell.Style.Fg)

// Convert buffer to SVG
svg := terma.BufferToSVG(buf, width, height, terma.DefaultSVGOptions())
```

### Comparing Snapshots

```go
// Render both widgets to buffers
expectedBuf := terma.RenderToBuffer(expectedWidget, width, height)
actualBuf := terma.RenderToBuffer(actualWidget, width, height)

// Compare buffers to get statistics
stats := terma.CompareBuffers(expectedBuf, actualBuf, width, height)

fmt.Printf("Similarity: %.1f%%\n", stats.Similarity)
fmt.Printf("Total cells: %d\n", stats.TotalCells)
fmt.Printf("Mismatched: %d\n", stats.MismatchedCells)

// Include stats in comparison for gallery
comparison := terma.SnapshotComparison{
    Name:     "My Test",
    Expected: terma.BufferToSVG(expectedBuf, width, height, opts),
    Actual:   terma.BufferToSVG(actualBuf, width, height, opts),
    Passed:   stats.MismatchedCells == 0,
    Stats:    stats,  // Stats will be displayed in the gallery
}
```

## Comparison Gallery

Generate an HTML page showing expected vs actual snapshots side-by-side:

```go
comparisons := []terma.SnapshotComparison{
    {
        Name:     "Login Form",
        Expected: expectedSVG,  // SVG string
        Actual:   actualSVG,    // SVG string
        Passed:   expectedSVG == actualSVG,
    },
    {
        Name:     "Dashboard",
        Expected: dashboardExpected,
        Actual:   dashboardActual,
        Passed:   true,
    },
}

terma.GenerateGallery(comparisons, "snapshot-gallery.html")
```

The gallery includes:
- **Filter buttons** - Show All / Failed Only / Passed Only
- **View modes** for comparing snapshots:
  - **Side by Side** - Traditional two-column view
  - **Overlay** - Stack images with adjustable opacity slider
  - **Slider** - Drag to reveal expected vs actual (like a before/after comparison)
  - **Difference** - Uses CSS blend mode to highlight differences (black = identical, colored = different)

## Example Test Pattern

```go
package myapp_test

import (
    "os"
    "path/filepath"
    "testing"
    "terma"
)

func TestWidgetSnapshots(t *testing.T) {
    tests := []struct {
        name   string
        widget terma.Widget
        width  int
        height int
    }{
        {"empty_list", MyList{Items: nil}, 40, 10},
        {"populated_list", MyList{Items: []string{"a", "b", "c"}}, 40, 10},
        {"with_border", MyBox{Border: true}, 30, 8},
    }

    var comparisons []terma.SnapshotComparison

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            actual := terma.Snapshot(tt.widget, tt.width, tt.height)

            goldenPath := filepath.Join("testdata", tt.name+".svg")

            if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
                // Update golden files
                os.WriteFile(goldenPath, []byte(actual), 0644)
                return
            }

            expected, err := os.ReadFile(goldenPath)
            if err != nil {
                t.Fatalf("missing golden file %s (run with UPDATE_SNAPSHOTS=1)", goldenPath)
            }

            passed := string(expected) == actual
            comparisons = append(comparisons, terma.SnapshotComparison{
                Name:     tt.name,
                Expected: string(expected),
                Actual:   actual,
                Passed:   passed,
            })

            if !passed {
                t.Errorf("snapshot mismatch for %s", tt.name)
            }
        })
    }

    // Generate comparison gallery on failure
    terma.GenerateGallery(comparisons, "snapshot-failures.html")
}
```

Run tests normally:
```bash
go test ./...
```

Update golden files when UI changes intentionally:
```bash
UPDATE_SNAPSHOTS=1 go test ./...
```

## Demo

Run the included demo to see snapshot generation in action:

```bash
go run ./cmd/snapshot-demo/main.go
```

This creates:
- `snapshot-output/expected.svg` - Sample widget snapshot
- `snapshot-output/actual.svg` - Modified variant
- `snapshot-output/gallery.html` - Comparison gallery

Open `gallery.html` in a browser to view the side-by-side comparison.
