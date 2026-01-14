package main

import (
	"fmt"
	"os"
	"path/filepath"

	"terma"
)

// DemoApp is a sample widget for demonstrating snapshots.
type DemoApp struct{}

func (d DemoApp) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Column{
		Style: terma.Style{
			BackgroundColor: terma.RGB(30, 30, 46),
			Padding:         terma.EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
			Border:          terma.Border{Style: terma.BorderRounded, Color: terma.RGB(137, 180, 250)},
		},
		Children: []terma.Widget{
			terma.Text{Content: "Terma Snapshot Demo", Style: terma.Style{
				ForegroundColor: terma.RGB(203, 166, 247),
				Bold:            true,
			}},
			terma.Spacer{Height: terma.Cells(1)},
			terma.Text{Content: "This is a demo of the snapshot feature.", Style: terma.Style{
				ForegroundColor: terma.RGB(205, 214, 244),
			}},
			terma.Text{Content: "It can capture terminal UIs as SVG.", Style: terma.Style{
				ForegroundColor: terma.RGB(166, 173, 200),
				Italic:          true,
			}},
		},
	}
}

// AltApp is a slightly different widget for comparison.
type AltApp struct{}

func (a AltApp) Build(ctx terma.BuildContext) terma.Widget {
	return terma.Column{
		Style: terma.Style{
			BackgroundColor: terma.RGB(30, 30, 46),
			Padding:         terma.EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
			Border:          terma.Border{Style: terma.BorderRounded, Color: terma.RGB(137, 180, 250)},
		},
		Children: []terma.Widget{
			terma.Text{Content: "Terma Snapshot Demo", Style: terma.Style{
				ForegroundColor: terma.RGB(203, 166, 247),
				Bold:            true,
			}},
			terma.Spacer{Height: terma.Cells(1)},
			terma.Text{Content: "This is a MODIFIED demo.", Style: terma.Style{
				ForegroundColor: terma.RGB(255, 100, 100), // Different color
			}},
			terma.Text{Content: "Notice the differences!", Style: terma.Style{
				ForegroundColor: terma.RGB(166, 173, 200),
				Italic:          true,
			}},
		},
	}
}

func main() {
	// Create output directory
	outDir := "snapshot-output"
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		return
	}

	width, height := 50, 12
	opts := terma.DefaultSVGOptions()

	// Render widgets to buffers
	expectedBuf := terma.RenderToBuffer(DemoApp{}, width, height)
	actualBuf := terma.RenderToBuffer(AltApp{}, width, height)

	// Compare buffers to get stats
	diffStats := terma.CompareBuffers(expectedBuf, actualBuf, width, height)
	sameStats := terma.CompareBuffers(expectedBuf, expectedBuf, width, height)

	fmt.Printf("Comparison stats: %d cells, %d mismatched (%.1f%% similar)\n",
		diffStats.TotalCells, diffStats.MismatchedCells, diffStats.Similarity)

	// Convert buffers to SVG
	expectedSVG := terma.BufferToSVG(expectedBuf, width, height, opts)
	actualSVG := terma.BufferToSVG(actualBuf, width, height, opts)

	// Generate diff SVGs for highlighting differences
	diffSVG := terma.GenerateDiffSVG(expectedBuf, actualBuf, width, height, opts)
	sameDiffSVG := terma.GenerateDiffSVG(expectedBuf, expectedBuf, width, height, opts)

	// Save SVGs
	expectedPath := filepath.Join(outDir, "expected.svg")
	actualPath := filepath.Join(outDir, "actual.svg")

	if err := os.WriteFile(expectedPath, []byte(expectedSVG), 0644); err != nil {
		fmt.Printf("Error saving expected SVG: %v\n", err)
		return
	}
	fmt.Printf("Saved: %s\n", expectedPath)

	if err := os.WriteFile(actualPath, []byte(actualSVG), 0644); err != nil {
		fmt.Printf("Error saving actual SVG: %v\n", err)
		return
	}
	fmt.Printf("Saved: %s\n", actualPath)

	// Generate comparison gallery
	comparisons := []terma.SnapshotComparison{
		{
			Name:     "Basic Demo Widget",
			Expected: expectedSVG,
			Actual:   actualSVG,
			DiffSVG:  diffSVG,
			Passed:   diffStats.MismatchedCells == 0,
			Stats:    diffStats,
		},
		{
			Name:     "Same Widget (should pass)",
			Expected: expectedSVG,
			Actual:   expectedSVG,
			DiffSVG:  sameDiffSVG,
			Passed:   sameStats.MismatchedCells == 0,
			Stats:    sameStats,
		},
	}

	galleryPath := filepath.Join(outDir, "gallery.html")
	if err := terma.GenerateGallery(comparisons, galleryPath); err != nil {
		fmt.Printf("Error generating gallery: %v\n", err)
		return
	}
	fmt.Printf("Saved: %s\n", galleryPath)

	fmt.Println("\nOpen gallery.html in a browser to view the snapshot comparison!")
}
