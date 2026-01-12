package terma

import (
	"fmt"
	"html"
	"image/color"
	"os"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
)

// SVGOptions configures SVG output generation.
type SVGOptions struct {
	FontFamily string  // Default: "Menlo, Monaco, Consolas, monospace"
	FontSize   int     // Default: 14
	LineHeight float64 // Default: 1.4
	CellWidth  float64 // Default: calculated from font size (fontSize * 0.6)
	Background Color   // Default: black
	Padding    int     // Default: 8
}

// DefaultSVGOptions returns sensible defaults for SVG generation.
func DefaultSVGOptions() SVGOptions {
	return SVGOptions{
		FontFamily: "Fira Code, Menlo, Monaco, Consolas, monospace",
		FontSize:   14,
		LineHeight: 1.4,
		CellWidth:  0, // 0 means calculate from font size
		Background: RGB(0, 0, 0),
		Padding:    8,
	}
}

// RenderToBuffer renders a widget to a headless buffer.
// The returned buffer can be inspected with CellAt() or converted to SVG.
func RenderToBuffer(widget Widget, width, height int) *uv.Buffer {
	buf := uv.NewBuffer(width, height)

	// Create focus manager and signals (required for rendering)
	focusManager := NewFocusManager()
	focusManager.SetRootWidget(widget)
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)

	// Create renderer and render the widget
	renderer := NewRenderer(buf, width, height, focusManager, focusedSignal, hoveredSignal)
	renderer.Render(widget)

	return buf
}

// Snapshot renders a widget and returns SVG with default options.
func Snapshot(widget Widget, width, height int) string {
	return SnapshotWithOptions(widget, width, height, DefaultSVGOptions())
}

// SnapshotWithOptions renders a widget and returns SVG with custom options.
func SnapshotWithOptions(widget Widget, width, height int, opts SVGOptions) string {
	buf := RenderToBuffer(widget, width, height)
	return BufferToSVG(buf, width, height, opts)
}

// SaveSnapshot renders a widget and writes the SVG to a file.
func SaveSnapshot(widget Widget, width, height int, path string) error {
	svg := Snapshot(widget, width, height)
	return os.WriteFile(path, []byte(svg), 0644)
}

// SaveSnapshotWithOptions renders a widget and writes the SVG to a file with custom options.
func SaveSnapshotWithOptions(widget Widget, width, height int, path string, opts SVGOptions) error {
	svg := SnapshotWithOptions(widget, width, height, opts)
	return os.WriteFile(path, []byte(svg), 0644)
}

// BufferToSVG converts a cell buffer to an SVG string.
func BufferToSVG(buf CellBuffer, width, height int, opts SVGOptions) string {
	// Apply defaults for zero values
	if opts.FontFamily == "" {
		opts.FontFamily = "Fira Code, Menlo, Monaco, Consolas, monospace"
	}
	if opts.FontSize == 0 {
		opts.FontSize = 14
	}
	if opts.LineHeight == 0 {
		opts.LineHeight = 1.4
	}
	if opts.CellWidth == 0 {
		opts.CellWidth = float64(opts.FontSize) * 0.6
	}
	if !opts.Background.IsSet() {
		opts.Background = RGB(0, 0, 0)
	}

	cellHeight := float64(opts.FontSize) * opts.LineHeight
	svgWidth := float64(opts.Padding*2) + float64(width)*opts.CellWidth
	svgHeight := float64(opts.Padding*2) + float64(height)*cellHeight

	var sb strings.Builder

	// SVG header
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">`,
		svgWidth, svgHeight, svgWidth, svgHeight))
	sb.WriteString("\n")

	// Style block with Google Fonts import for Fira Code
	sb.WriteString(fmt.Sprintf(`  <style>
    @import url('https://fonts.googleapis.com/css2?family=Fira+Code:wght@400;700&amp;display=swap');
    text { font-family: %s; font-size: %dpx; dominant-baseline: text-before-edge; }
    .bold { font-weight: bold; }
    .italic { font-style: italic; }
    .underline { text-decoration: underline; }
    .strikethrough { text-decoration: line-through; }
  </style>`, opts.FontFamily, opts.FontSize))
	sb.WriteString("\n")

	// Background
	sb.WriteString(fmt.Sprintf(`  <rect width="100%%" height="100%%" fill="%s"/>`, opts.Background.Hex()))
	sb.WriteString("\n")

	// Render each row
	for y := 0; y < height; y++ {
		rowY := float64(opts.Padding) + float64(y)*cellHeight

		// First pass: render background rects
		x := 0
		for x < width {
			cell := buf.CellAt(x, y)
			if cell == nil {
				x++
				continue
			}

			// Check for background color
			if cell.Style.Bg != nil {
				bgColor := FromANSI(cell.Style.Bg)
				if bgColor.IsSet() && bgColor.Hex() != opts.Background.Hex() {
					cellX := float64(opts.Padding) + float64(x)*opts.CellWidth
					cellW := opts.CellWidth
					if cell.Width > 1 {
						cellW = float64(cell.Width) * opts.CellWidth
					}
					sb.WriteString(fmt.Sprintf(`  <rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s"/>`,
						cellX, rowY, cellW, cellHeight, bgColor.Hex()))
					sb.WriteString("\n")
				}
			}

			// Advance by cell width
			if cell.Width > 1 {
				x += cell.Width
			} else {
				x++
			}
		}

		// Second pass: render text
		// Group consecutive characters with the same style for efficiency
		x = 0
		for x < width {
			cell := buf.CellAt(x, y)
			if cell == nil || cell.Content == "" || cell.Content == " " {
				x++
				continue
			}

			// Collect consecutive cells with the same style
			startX := x
			var textContent strings.Builder
			textContent.WriteString(cell.Content)
			baseStyle := cell.Style
			baseFg := FromANSI(cell.Style.Fg)

			// Advance past this cell
			if cell.Width > 1 {
				x += cell.Width
			} else {
				x++
			}

			// Look ahead for same-style cells
			for x < width {
				nextCell := buf.CellAt(x, y)
				if nextCell == nil || nextCell.Content == "" {
					break
				}
				nextFg := FromANSI(nextCell.Style.Fg)
				if !sameStyle(baseStyle, nextCell.Style) || baseFg.Hex() != nextFg.Hex() {
					break
				}
				textContent.WriteString(nextCell.Content)
				if nextCell.Width > 1 {
					x += nextCell.Width
				} else {
					x++
				}
			}

			// Render the text span
			textX := float64(opts.Padding) + float64(startX)*opts.CellWidth
			textY := rowY

			// Build style classes
			var classes []string
			if baseStyle.Attrs&uv.AttrBold != 0 {
				classes = append(classes, "bold")
			}
			if baseStyle.Attrs&uv.AttrItalic != 0 {
				classes = append(classes, "italic")
			}
			if baseStyle.Underline != uv.UnderlineNone {
				classes = append(classes, "underline")
			}
			if baseStyle.Attrs&uv.AttrStrikethrough != 0 {
				classes = append(classes, "strikethrough")
			}

			classAttr := ""
			if len(classes) > 0 {
				classAttr = fmt.Sprintf(` class="%s"`, strings.Join(classes, " "))
			}

			fillAttr := ""
			if baseFg.IsSet() {
				fillAttr = fmt.Sprintf(` fill="%s"`, baseFg.Hex())
			} else {
				fillAttr = ` fill="#FFFFFF"` // default to white text
			}

			sb.WriteString(fmt.Sprintf(`  <text x="%.1f" y="%.1f"%s%s>%s</text>`,
				textX, textY, classAttr, fillAttr, html.EscapeString(textContent.String())))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("</svg>\n")
	return sb.String()
}

// sameStyle checks if two uv.Style values have the same attributes (ignoring colors).
func sameStyle(a, b uv.Style) bool {
	return a.Attrs == b.Attrs && a.Underline == b.Underline
}

// SnapshotStats contains comparison statistics between two snapshots.
type SnapshotStats struct {
	TotalCells      int     // Total number of cells compared
	MatchingCells   int     // Number of cells that match exactly
	MismatchedCells int     // Number of cells that differ
	Similarity      float64 // Percentage of matching cells (0-100)
}

// CompareBuffers compares two buffers and returns comparison statistics.
func CompareBuffers(expected, actual *uv.Buffer, width, height int) SnapshotStats {
	stats := SnapshotStats{}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			stats.TotalCells++

			expectedCell := expected.CellAt(x, y)
			actualCell := actual.CellAt(x, y)

			if cellsEqual(expectedCell, actualCell) {
				stats.MatchingCells++
			} else {
				stats.MismatchedCells++
			}
		}
	}

	if stats.TotalCells > 0 {
		stats.Similarity = float64(stats.MatchingCells) / float64(stats.TotalCells) * 100
	}

	return stats
}

// cellsEqual compares two cells for equality.
func cellsEqual(a, b *uv.Cell) bool {
	// Both nil = equal
	if a == nil && b == nil {
		return true
	}
	// One nil = not equal
	if a == nil || b == nil {
		return false
	}
	// Compare content
	if a.Content != b.Content {
		return false
	}
	// Compare foreground color
	if !colorsEqual(a.Style.Fg, b.Style.Fg) {
		return false
	}
	// Compare background color
	if !colorsEqual(a.Style.Bg, b.Style.Bg) {
		return false
	}
	// Compare attributes
	if a.Style.Attrs != b.Style.Attrs {
		return false
	}
	return true
}

// colorsEqual compares two ANSI colors for equality.
func colorsEqual(a, b color.Color) bool {
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

// SnapshotComparison represents a comparison between expected and actual snapshots.
type SnapshotComparison struct {
	Name     string        // Test name / description
	Expected string        // SVG content or path
	Actual   string        // SVG content or path
	Passed   bool          // Whether they match
	Stats    SnapshotStats // Comparison statistics (optional)
}

// GenerateGallery creates an HTML page comparing actual vs expected snapshots.
func GenerateGallery(comparisons []SnapshotComparison, outputPath string) error {
	// Calculate pass/fail counts
	passedCount := 0
	failedCount := 0
	for _, comp := range comparisons {
		if comp.Passed {
			passedCount++
		} else {
			failedCount++
		}
	}

	var sb strings.Builder

	sb.WriteString(`<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Terma Snapshot Gallery</title>
  <style>
    * { box-sizing: border-box; }
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      background: #1a1a2e;
      color: #eee;
      margin: 0;
      padding: 20px;
    }
    h1 { margin: 0 0 20px; color: #fff; }
    .toolbar {
      display: flex;
      gap: 20px;
      margin-bottom: 20px;
      flex-wrap: wrap;
      align-items: center;
    }
    .toolbar-group {
      display: flex;
      gap: 8px;
      align-items: center;
    }
    .toolbar-label {
      font-size: 12px;
      color: #888;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }
    .toolbar button {
      background: #2d2d44;
      border: none;
      color: #eee;
      padding: 8px 16px;
      border-radius: 4px;
      cursor: pointer;
      font-size: 14px;
    }
    .toolbar button:hover { background: #3d3d54; }
    .toolbar button.active { background: #5a5a8a; }
    .comparison {
      margin-bottom: 30px;
      padding: 20px;
      background: #2d2d44;
      border-radius: 8px;
    }
    .comparison.failed { border: 2px solid #ff4444; }
    .comparison.passed { border: 2px solid #44ff44; }
    .comparison-header {
      display: flex;
      align-items: center;
      gap: 10px;
      margin-bottom: 15px;
    }
    .comparison-name { font-size: 18px; font-weight: 600; }
    .status-badge {
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 12px;
      font-weight: 600;
    }
    .status-badge.passed { background: #44ff44; color: #000; }
    .status-badge.failed { background: #ff4444; color: #fff; }
    .stats {
      display: flex;
      gap: 15px;
      margin-left: auto;
      font-size: 15px;
      color: #aaa;
    }
    .stat {
      display: flex;
      align-items: center;
      gap: 5px;
    }
    .stat-value {
      font-weight: 600;
      color: #fff;
    }
    .stat-value.good { color: #44ff44; }
    .stat-value.bad { color: #ff4444; }

    /* Side-by-side view */
    .view-sidebyside .snapshots { display: flex; gap: 20px; }
    .view-sidebyside .snapshot-container { flex: 1; }
    .view-sidebyside .diff-view { display: none; }
    .snapshot-label { font-size: 14px; color: #aaa; margin-bottom: 8px; }
    .snapshot {
      background: #1a1a2e;
      border-radius: 4px;
      overflow: hidden;
    }
    .snapshot svg { display: block; max-width: 100%; height: auto; }

    /* Overlay view */
    .view-overlay .snapshots { display: none; }
    .view-overlay .diff-view { display: block; }
    .diff-view { display: none; }
    .diff-layers { position: relative; }
    .diff-layers .expected-layer,
    .diff-layers .actual-layer {
      border-radius: 4px;
      overflow: hidden;
      background: #1a1a2e;
    }
    .diff-layers .expected-layer { position: relative; }
    .diff-layers .actual-layer {
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      opacity: 0.5;
      pointer-events: none;
    }
    .diff-layers svg { display: block; max-width: 100%; height: auto; }
    .diff-controls {
      margin-top: 10px;
      display: flex;
      align-items: center;
      gap: 10px;
    }
    .diff-controls label { font-size: 14px; color: #aaa; }
    .diff-controls input[type="range"] { flex: 1; max-width: 300px; }

    /* Slider view */
    .view-slider .snapshots { display: none; }
    .view-slider .diff-view { display: block; }
    .view-slider .diff-layers .actual-layer {
      opacity: 1;
      clip-path: inset(0 50% 0 0);
      transition: none;
    }
    .view-slider .diff-controls { display: block; }
    .slider-label { font-size: 12px; color: #666; }

    /* Difference blend mode */
    .view-difference .snapshots { display: none; }
    .view-difference .diff-view { display: block; }
    .view-difference .diff-layers .actual-layer {
      opacity: 1;
      mix-blend-mode: difference;
    }
    .view-difference .diff-controls { display: none; }

    .hidden { display: none !important; }
    .help-text {
      font-size: 12px;
      color: #666;
    }
    .header-bar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 20px;
    }
    .summary {
      display: flex;
      gap: 15px;
      font-size: 15px;
    }
    .summary-item {
      display: flex;
      align-items: center;
      gap: 6px;
    }
    .summary-count {
      font-weight: 700;
      font-size: 18px;
    }
    .summary-count.passed { color: #44ff44; }
    .summary-count.failed { color: #ff4444; }
  </style>
</head>
<body>
`)

	sb.WriteString(fmt.Sprintf(`  <div class="header-bar">
    <h1 style="margin: 0;">Terma Snapshot Gallery</h1>
    <div class="summary">
      <div class="summary-item"><span class="summary-count passed">%d</span> passed</div>
      <div class="summary-item"><span class="summary-count failed">%d</span> failed</div>
    </div>
  </div>
`, passedCount, failedCount))

	sb.WriteString(`  <div class="toolbar">
    <div class="toolbar-group">
      <span class="toolbar-label">Filter:</span>
      <button class="filter-btn active" data-filter="all">All</button>
      <button class="filter-btn" data-filter="failed">Failed</button>
      <button class="filter-btn" data-filter="passed">Passed</button>
    </div>
    <div class="toolbar-group">
      <span class="toolbar-label">View:</span>
      <button class="view-btn active" data-view="sidebyside">Side by Side</button>
      <button class="view-btn" data-view="overlay">Overlay</button>
      <button class="view-btn" data-view="slider">Slider</button>
      <button class="view-btn" data-view="difference">Difference</button>
    </div>
    <span class="help-text">Difference mode: black = identical, colored = different</span>
  </div>
`)

	for i, comp := range comparisons {
		status := "passed"
		if !comp.Passed {
			status = "failed"
		}

		// Build stats HTML if stats are available
		statsHTML := ""
		if comp.Stats.TotalCells > 0 {
			similarityClass := "good"
			if comp.Stats.Similarity < 100 {
				similarityClass = "bad"
			}
			mismatchClass := ""
			if comp.Stats.MismatchedCells > 0 {
				mismatchClass = "bad"
			}
			statsHTML = fmt.Sprintf(`
      <div class="stats">
        <div class="stat">Similarity: <span class="stat-value %s">%.1f%%</span></div>
        <div class="stat">Cells: <span class="stat-value">%d</span></div>
        <div class="stat">Mismatched: <span class="stat-value %s">%d</span></div>
      </div>`, similarityClass, comp.Stats.Similarity, comp.Stats.TotalCells, mismatchClass, comp.Stats.MismatchedCells)
		}

		sb.WriteString(fmt.Sprintf(`  <div class="comparison %s view-sidebyside" data-status="%s" data-index="%d">
    <div class="comparison-header">
      <span class="comparison-name">%s</span>
      <span class="status-badge %s">%s</span>%s
    </div>
    <div class="snapshots">
      <div class="snapshot-container">
        <div class="snapshot-label">Expected</div>
        <div class="snapshot expected">
%s
        </div>
      </div>
      <div class="snapshot-container">
        <div class="snapshot-label">Actual</div>
        <div class="snapshot actual">
%s
        </div>
      </div>
    </div>
    <div class="diff-view">
      <div class="snapshot-label"><span class="diff-mode-label">Overlay</span>: Expected + Actual</div>
      <div class="diff-layers">
        <div class="expected-layer">
%s
        </div>
        <div class="actual-layer">
%s
        </div>
      </div>
      <div class="diff-controls">
        <label class="slider-label-text">Actual opacity:</label>
        <input type="range" min="0" max="100" value="50" class="opacity-slider">
        <span class="opacity-value">50%%</span>
      </div>
    </div>
  </div>
`, status, status, i, html.EscapeString(comp.Name), status, strings.ToUpper(status), statsHTML,
			indentSVG(comp.Expected, "          "),
			indentSVG(comp.Actual, "          "),
			indentSVG(comp.Expected, "        "),
			indentSVG(comp.Actual, "        ")))
	}

	sb.WriteString(`  <script>
    // Filter buttons
    document.querySelectorAll('.filter-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        const filter = btn.dataset.filter;
        document.querySelectorAll('.comparison').forEach(el => {
          if (filter === 'all') {
            el.classList.remove('hidden');
          } else {
            el.classList.toggle('hidden', el.dataset.status !== filter);
          }
        });
      });
    });

    // View mode buttons
    document.querySelectorAll('.view-btn').forEach(btn => {
      btn.addEventListener('click', () => {
        document.querySelectorAll('.view-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        const view = btn.dataset.view;
        document.querySelectorAll('.comparison').forEach(el => {
          el.classList.remove('view-sidebyside', 'view-overlay', 'view-slider', 'view-difference');
          el.classList.add('view-' + view);

          // Update mode label
          const label = el.querySelector('.diff-mode-label');
          if (label) {
            const labels = { overlay: 'Overlay', slider: 'Slider', difference: 'Difference' };
            label.textContent = labels[view] || 'Overlay';
          }

          // Update slider label
          const sliderLabel = el.querySelector('.slider-label-text');
          if (sliderLabel) {
            sliderLabel.textContent = view === 'slider' ? 'Slider position:' : 'Actual opacity:';
          }

          // Reset actual layer styles and slider
          const actualLayer = el.querySelector('.actual-layer');
          const slider = el.querySelector('.opacity-slider');
          if (actualLayer && slider) {
            // Clear inline styles so CSS takes over
            actualLayer.style.clipPath = '';
            actualLayer.style.opacity = '';
            slider.value = 50;
            const valueDisplay = el.querySelector('.opacity-value');
            if (valueDisplay) valueDisplay.textContent = '50%';
          }
        });
        // Update help text
        const helpText = document.querySelector('.help-text');
        if (view === 'difference') {
          helpText.textContent = 'Difference mode: black = identical, colored = different';
        } else if (view === 'slider') {
          helpText.textContent = 'Slider mode: drag to reveal expected vs actual';
        } else if (view === 'overlay') {
          helpText.textContent = 'Overlay mode: adjust opacity to compare';
        } else {
          helpText.textContent = '';
        }
      });
    });

    // Opacity/slider controls
    document.querySelectorAll('.opacity-slider').forEach(slider => {
      slider.addEventListener('input', (e) => {
        const comparison = e.target.closest('.comparison');
        const value = e.target.value;
        updateSlider(comparison, value);
      });
    });

    function updateSlider(comparison, value) {
      const actualLayer = comparison.querySelector('.actual-layer');
      const valueDisplay = comparison.querySelector('.opacity-value');
      if (comparison.classList.contains('view-slider')) {
        // Slider mode: clip-path
        actualLayer.style.clipPath = 'inset(0 ' + (100 - value) + '% 0 0)';
        actualLayer.style.opacity = '1';
        valueDisplay.textContent = value + '%';
      } else {
        // Overlay mode: opacity
        actualLayer.style.clipPath = '';
        actualLayer.style.opacity = value / 100;
        valueDisplay.textContent = value + '%';
      }
    }
  </script>
</body>
</html>
`)

	return os.WriteFile(outputPath, []byte(sb.String()), 0644)
}

// indentSVG adds indentation to each line of an SVG string.
func indentSVG(svg string, indent string) string {
	lines := strings.Split(strings.TrimSpace(svg), "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}
