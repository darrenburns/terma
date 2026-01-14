package terma

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// snapshotRegistry collects all snapshot comparisons during a test run.
// Use SnapshotTestMain to generate a combined gallery after tests complete.
var (
	snapshotRegistry   []SnapshotComparison
	snapshotRegistryMu sync.Mutex
)

// registerComparison adds a comparison to the global registry.
func registerComparison(c SnapshotComparison) {
	snapshotRegistryMu.Lock()
	defer snapshotRegistryMu.Unlock()
	snapshotRegistry = append(snapshotRegistry, c)
}

// SnapshotTestMain generates a gallery of all snapshot comparisons.
// Call this from TestMain after m.Run() completes.
//
// Example:
//
//	func TestMain(m *testing.M) {
//	    code := m.Run()
//	    terma.SnapshotTestMain("testdata/snapshot_gallery.html")
//	    os.Exit(code)
//	}
func SnapshotTestMain(galleryPath string) {
	snapshotRegistryMu.Lock()
	defer snapshotRegistryMu.Unlock()

	if len(snapshotRegistry) == 0 {
		return
	}

	// Generate gallery for all comparisons (pass or fail)
	if err := GenerateGallery(snapshotRegistry, galleryPath); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to generate snapshot gallery: %v\n", err)
		return
	}

	absPath, _ := filepath.Abs(galleryPath)

	// Print gallery link with visual emphasis
	fmt.Printf("\n")
	fmt.Printf("╭─────────────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│  Snapshot gallery (%d snapshots)                                 \n", len(snapshotRegistry))
	fmt.Printf("│  %s\n", absPath)
	fmt.Printf("╰─────────────────────────────────────────────────────────────────╯\n")
	fmt.Printf("\n")
}

// AssertSnapshot renders a widget and compares it against a golden file.
//
// The golden file is stored in testdata/<TestName>.svg relative to the test file.
// If the file doesn't exist or UPDATE_SNAPSHOTS=1 is set, the file is created/updated.
//
// All failures are collected and can be viewed in a combined gallery by calling
// SnapshotTestMain from your TestMain function.
//
// Example:
//
//	func TestMyWidget(t *testing.T) {
//	    widget := Text{Content: "Hello"}
//	    AssertSnapshot(t, widget, 20, 5)
//	}
//
// To update golden files after intentional changes:
//
//	UPDATE_SNAPSHOTS=1 go test ./...
func AssertSnapshot(t *testing.T, widget Widget, width, height int) {
	t.Helper()
	AssertSnapshotWithOptions(t, widget, width, height, DefaultSVGOptions())
}

// AssertSnapshotWithOptions renders a widget with custom SVG options and compares against a golden file.
func AssertSnapshotWithOptions(t *testing.T, widget Widget, width, height int, opts SVGOptions) {
	t.Helper()
	assertSnapshotNamed(t, t.Name(), widget, width, height, opts)
}

// AssertSnapshotNamed renders a widget and compares against a golden file with a custom name.
// Use this when you need multiple snapshots in a single test.
//
// Example:
//
//	func TestMyWidget(t *testing.T) {
//	    AssertSnapshotNamed(t, "initial", widget, 20, 5)
//	    widget.Update()
//	    AssertSnapshotNamed(t, "after_update", widget, 20, 5)
//	}
func AssertSnapshotNamed(t *testing.T, name string, widget Widget, width, height int) {
	t.Helper()
	assertSnapshotNamed(t, name, widget, width, height, DefaultSVGOptions())
}

// AssertSnapshotNamedWithOptions renders a widget with custom options and compares against a named golden file.
func AssertSnapshotNamedWithOptions(t *testing.T, name string, widget Widget, width, height int, opts SVGOptions) {
	t.Helper()
	assertSnapshotNamed(t, name, widget, width, height, opts)
}

func assertSnapshotNamed(t *testing.T, name string, widget Widget, width, height int, opts SVGOptions) {
	t.Helper()

	sanitizedName := sanitizeFilename(name)

	// Render to buffer first (we need this for DiffSVG generation)
	actualBuf := RenderToBuffer(widget, width, height)
	actualSVG := BufferToSVG(actualBuf, width, height, opts)

	goldenPath := filepath.Join("testdata", sanitizedName+".svg")
	bufferPath := filepath.Join("testdata", sanitizedName+".buf.json")

	// Update mode: write the golden file and buffer data
	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		// Ensure testdata directory exists
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata directory: %v", err)
		}
		if err := os.WriteFile(goldenPath, []byte(actualSVG), 0644); err != nil {
			t.Fatalf("failed to write golden file %s: %v", goldenPath, err)
		}

		// Save buffer data for DiffSVG generation
		bufData := SerializeBuffer(actualBuf, width, height)
		bufJSON, err := json.Marshal(bufData)
		if err != nil {
			t.Fatalf("failed to serialize buffer: %v", err)
		}
		if err := os.WriteFile(bufferPath, bufJSON, 0644); err != nil {
			t.Fatalf("failed to write buffer file %s: %v", bufferPath, err)
		}

		t.Logf("updated golden file: %s", goldenPath)
		return
	}

	// Compare mode: read golden file and compare
	expectedBytes, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		registerComparison(SnapshotComparison{
			Name:     name,
			Expected: fmt.Sprintf("<!-- File not found: %s -->", goldenPath),
			Actual:   actualSVG,
			Passed:   false,
		})
		t.Fatalf("golden file not found: %s\nRun with UPDATE_SNAPSHOTS=1 to create it", goldenPath)
	}
	if err != nil {
		t.Fatalf("failed to read golden file %s: %v", goldenPath, err)
	}

	expectedSVG := string(expectedBytes)
	passed := expectedSVG == actualSVG

	// Generate DiffSVG and Stats if there's a mismatch
	var diffSVG string
	var stats SnapshotStats
	if !passed {
		// Try to load expected buffer for diff generation
		bufJSON, err := os.ReadFile(bufferPath)
		if err == nil {
			var bufData SerializedBuffer
			if json.Unmarshal(bufJSON, &bufData) == nil {
				expectedBuf := bufData.ToBuffer()
				diffSVG = GenerateDiffSVG(expectedBuf, actualBuf, width, height, opts)
				stats = CompareBuffers(expectedBuf, actualBuf, width, height)
			}
		}
	}

	// Register comparison for combined gallery
	registerComparison(SnapshotComparison{
		Name:     name,
		Expected: expectedSVG,
		Actual:   actualSVG,
		DiffSVG:  diffSVG,
		Passed:   passed,
		Stats:    stats,
	})

	if !passed {
		t.Errorf("snapshot mismatch for %s\n"+
			"  Golden file: %s\n"+
			"\nRun with UPDATE_SNAPSHOTS=1 to update the golden file if this change is intentional",
			name, goldenPath)
	}
}

// sanitizeFilename converts a test name to a valid filename.
// Replaces path separators and other problematic characters.
func sanitizeFilename(name string) string {
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch c {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			result = append(result, '_')
		default:
			result = append(result, c)
		}
	}
	return string(result)
}

// SnapshotSuite provides a convenient way to run multiple snapshot assertions with custom options.
// Results are automatically collected into the global registry for the combined gallery.
//
// Example:
//
//	func TestWidgets(t *testing.T) {
//	    suite := terma.NewSnapshotSuite(t)
//	    suite.Assert("button", buttonWidget, 20, 5)
//	    suite.Assert("list", listWidget, 30, 10)
//	}
type SnapshotSuite struct {
	t    *testing.T
	opts SVGOptions
}

// NewSnapshotSuite creates a new snapshot suite.
func NewSnapshotSuite(t *testing.T) *SnapshotSuite {
	return &SnapshotSuite{
		t:    t,
		opts: DefaultSVGOptions(),
	}
}

// WithOptions sets custom SVG options for the suite.
func (s *SnapshotSuite) WithOptions(opts SVGOptions) *SnapshotSuite {
	s.opts = opts
	return s
}

// Assert renders a widget and compares against a golden file.
// Results are collected into the global registry for the combined gallery.
func (s *SnapshotSuite) Assert(name string, widget Widget, width, height int) {
	s.t.Helper()
	assertSnapshotNamed(s.t, name, widget, width, height, s.opts)
}
