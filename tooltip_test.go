package terma

import (
	"os"
	"path/filepath"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/darrenburns/terma/layout"
)

// renderToBufferWithFocus renders a widget with the given widget ID focused.
// This is used for testing tooltip visibility (tooltips show when child is focused).
// If focusID is empty, no widget will be focused (skips SetFocusables to avoid auto-focus).
func renderToBufferWithFocus(widget Widget, width, height int, focusID string) *uv.Buffer {
	buf := uv.NewBuffer(width, height)

	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)

	renderer := NewRenderer(buf, width, height, focusManager, focusedSignal, hoveredSignal)

	// First render to collect focusables
	focusables := renderer.Render(widget)

	// Only set focusables and re-render if we want focus
	// (SetFocusables auto-focuses first widget, so skip it for no-focus case)
	if focusID != "" {
		focusManager.SetFocusables(focusables)
		focusManager.FocusByID(focusID)

		// Re-render with focus set
		buf = uv.NewBuffer(width, height)
		renderer = NewRenderer(buf, width, height, focusManager, focusedSignal, hoveredSignal)
		renderer.Render(widget)
	}

	return buf
}

// snapshotWithFocus renders a widget with focus state and returns SVG.
func snapshotWithFocus(widget Widget, width, height int, focusID string) string {
	buf := renderToBufferWithFocus(widget, width, height, focusID)
	return BufferToSVG(buf, width, height, DefaultSVGOptions())
}

// assertSnapshotFromSVG compares an already-rendered SVG against a golden file.
// This is used when we need custom render setup (like setting focus state).
func assertSnapshotFromSVG(t *testing.T, actualSVG string, description string) {
	t.Helper()
	assertSnapshotFromSVGNamed(t, t.Name(), actualSVG, description)
}

// assertSnapshotFromSVGNamed compares an SVG against a named golden file.
func assertSnapshotFromSVGNamed(t *testing.T, name string, actualSVG string, description string) {
	t.Helper()

	sanitizedName := sanitizeFilename(name)
	goldenPath := filepath.Join("testdata", sanitizedName+".svg")

	// Check if we should update snapshots
	if os.Getenv("UPDATE_SNAPSHOTS") == "1" {
		// Ensure testdata directory exists
		if err := os.MkdirAll("testdata", 0755); err != nil {
			t.Fatalf("failed to create testdata directory: %v", err)
		}
		// Write the new golden file
		err := os.WriteFile(goldenPath, []byte(actualSVG), 0644)
		if err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read the expected golden file
	expectedSVG, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		registerComparison(SnapshotComparison{
			Name:        name,
			Description: description,
			Expected:    "<!-- File not found -->",
			Actual:      actualSVG,
			Passed:      false,
		})
		t.Fatalf("golden file not found: %s\n    Run with UPDATE_SNAPSHOTS=1 to create it", goldenPath)
	}
	if err != nil {
		t.Fatalf("failed to read golden file: %v", err)
	}

	passed := string(expectedSVG) == actualSVG

	// Register comparison for gallery
	registerComparison(SnapshotComparison{
		Name:        name,
		Description: description,
		Expected:    string(expectedSVG),
		Actual:      actualSVG,
		Passed:      passed,
	})

	if !passed {
		t.Errorf("snapshot mismatch for %s\n    Run with UPDATE_SNAPSHOTS=1 to update", name)
	}
}

func TestTooltip_ChildRendersWithoutFocus(t *testing.T) {
	// When child is not focused, only the child should render
	widget := Tooltip{
		Content: "Help text",
		Child:   Button{ID: "btn", Label: "Click me"},
	}
	// Use custom render to avoid auto-focus (AssertSnapshot auto-focuses first widget)
	svg := snapshotWithFocus(widget, 20, 5, "") // empty string = no focus
	assertSnapshotFromSVG(t, svg, "Button '[Click me]' at top-left. No tooltip visible because button is not focused.")
}

func TestTooltip_Position_Top_Visible(t *testing.T) {
	// Tooltip positioned above the child when focused
	// Position button lower so tooltip has room above
	widget := Column{
		Children: []Widget{
			Spacer{Height: Cells(3)},
			Tooltip{
				Content:  "Help text",
				Position: TooltipTop,
				Child:    Button{ID: "top-btn", Label: "Target"},
			},
		},
	}
	// Render with focus to show tooltip
	svg := snapshotWithFocus(widget, 25, 6, "top-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Target]' at row 3. Tooltip ' Help text ' on surface background positioned directly ABOVE button (no gap). Tooltip horizontally centered over button.")
}

func TestTooltip_Position_Bottom_Visible(t *testing.T) {
	widget := Tooltip{
		Content:  "Help text",
		Position: TooltipBottom,
		Child:    Button{ID: "bottom-btn", Label: "Target"},
	}
	svg := snapshotWithFocus(widget, 25, 4, "bottom-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Target]' at row 0. Tooltip ' Help text ' on surface background positioned directly BELOW button (no gap). Tooltip horizontally centered under button.")
}

func TestTooltip_Position_Left_Visible(t *testing.T) {
	// Position child to the right so tooltip has room on left
	widget := Row{
		Children: []Widget{
			Spacer{Width: Cells(10)},
			Tooltip{
				Content:  "Help",
				Position: TooltipLeft,
				Child:    Button{ID: "left-btn", Label: "Target"},
			},
		},
	}
	svg := snapshotWithFocus(widget, 26, 1, "left-btn")
	assertSnapshotFromSVG(t, svg, "Tooltip ' Help ' on left, then button '[Target]' on right (no gap between them).")
}

func TestTooltip_Position_Right_Visible(t *testing.T) {
	widget := Tooltip{
		Content:  "Help",
		Position: TooltipRight,
		Child:    Button{ID: "right-btn", Label: "Target"},
	}
	svg := snapshotWithFocus(widget, 20, 1, "right-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Target]' on left, then tooltip ' Help ' on right (no gap between them).")
}

func TestTooltip_RichText_Visible(t *testing.T) {
	// Tooltip with rich text spans - use Bottom position so tooltip doesn't overlap button
	widget := Tooltip{
		Spans: []Span{
			BoldSpan("Ctrl+S"),
			PlainSpan(" to save"),
		},
		Position: TooltipBottom,
		Child:    Button{ID: "rich-btn", Label: "Save"},
	}
	svg := snapshotWithFocus(widget, 20, 3, "rich-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Save]' at top. Tooltip below with ' Ctrl+S to save ' where 'Ctrl+S' is BOLD. Surface background, 1-cell horizontal padding.")
}

func TestTooltip_CustomStyle_Visible(t *testing.T) {
	// Use Bottom position so tooltip appears below button
	widget := Tooltip{
		Content:  "Styled",
		Position: TooltipBottom,
		Style: Style{
			BackgroundColor: RGB(50, 50, 100),
			ForegroundColor: RGB(255, 255, 255),
			Border:          Border{Style: BorderDouble, Color: RGB(100, 100, 200)},
			Padding:         EdgeInsetsAll(1),
		},
		Child: Button{ID: "styled-btn", Label: "Target"},
	}
	svg := snapshotWithFocus(widget, 16, 6, "styled-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Target]' at top. Tooltip below with DOUBLE-LINE border, dark blue background (#323264), white text 'Styled', 1 cell padding on all sides.")
}

func TestTooltip_CustomOffset_Visible(t *testing.T) {
	// Use Bottom position with 2-cell gap
	widget := Tooltip{
		Content:  "Help",
		Position: TooltipBottom,
		Offset:   2, // 2 cell gap
		Child:    Button{ID: "offset-btn", Label: "Target"},
	}
	svg := snapshotWithFocus(widget, 16, 5, "offset-btn")
	assertSnapshotFromSVG(t, svg, "Button '[Target]' at top. Tooltip ' Help ' below with 2 empty rows between button and tooltip.")
}

func TestTooltip_InColumn_Layout(t *testing.T) {
	// Tooltip within a column layout (no focus on tooltip child)
	widget := Column{
		Children: []Widget{
			Text{Content: "Header"},
			Tooltip{
				Content: "This is a tooltip",
				Child:   Button{ID: "btn", Label: "Click me"},
			},
			Text{Content: "Footer"},
		},
	}
	// Use custom render to avoid auto-focus
	svg := snapshotWithFocus(widget, 30, 3, "")
	assertSnapshotFromSVG(t, svg, "Vertical stack: 'Header' at top, '[Click me]' button in middle, 'Footer' at bottom. NO tooltip visible.")
}

func TestTooltip_InRow_Layout(t *testing.T) {
	// Tooltip within a row layout (no focus)
	widget := Row{
		Spacing: 2,
		Children: []Widget{
			Text{Content: "Left"},
			Tooltip{
				Content: "Tooltip help",
				Child:   Button{ID: "center-btn", Label: "Center"},
			},
			Text{Content: "Right"},
		},
	}
	// Use custom render to avoid auto-focus
	svg := snapshotWithFocus(widget, 30, 1, "")
	assertSnapshotFromSVG(t, svg, "Horizontal row: 'Left', then '[Center]' button, then 'Right'. NO tooltip visible.")
}

// TestTooltip_VisibleOnFocus tests that the tooltip appears when child is focused.
func TestTooltip_VisibleOnFocus(t *testing.T) {
	button := Button{ID: "focus-btn", Label: "Focus me"}

	widget := Tooltip{
		Content: "Focus tooltip!",
		Child:   button,
	}

	// Render with focus on the button
	buf := renderToBufferWithFocus(widget, 30, 8, "focus-btn")

	// Check that tooltip content is rendered somewhere in the buffer
	found := false
	for y := 0; y < 8; y++ {
		line := ""
		for x := 0; x < 30; x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				line += cell.Content
			}
		}
		if containsString(line, "Focus tooltip!") {
			found = true
			break
		}
	}
	assert.True(t, found, "Tooltip content should be visible when child is focused")
}

// TestTooltip_NotVisibleWhenNotFocused tests that tooltip doesn't show without focus.
func TestTooltip_NotVisibleWhenNotFocused(t *testing.T) {
	widget := Tooltip{
		Content: "Should not appear",
		Child:   Button{ID: "btn", Label: "Target"},
	}

	// Render WITHOUT focus
	buf := renderToBufferWithFocus(widget, 30, 8, "")

	// Check that tooltip content is NOT rendered
	found := false
	for y := 0; y < 8; y++ {
		line := ""
		for x := 0; x < 30; x++ {
			cell := buf.CellAt(x, y)
			if cell != nil {
				line += cell.Content
			}
		}
		if containsString(line, "Should not appear") {
			found = true
			break
		}
	}
	assert.False(t, found, "Tooltip should NOT appear when child is not focused")
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test that Tooltip implements the necessary interfaces
func TestTooltip_Interfaces(t *testing.T) {
	tooltip := Tooltip{
		ID:    "test-tooltip",
		Child: Text{Content: "Child"},
	}

	// Should implement Identifiable
	var _ Identifiable = tooltip
	assert.Equal(t, "test-tooltip", tooltip.WidgetID())

	// Should implement Widget
	var _ Widget = tooltip

	// Should implement ChildProvider
	var _ ChildProvider = tooltip
	children := tooltip.ChildWidgets()
	assert.Len(t, children, 1)

	// Should implement LayoutNodeBuilder
	var _ LayoutNodeBuilder = tooltip
}

// Test layout node building
func TestTooltip_BuildLayoutNode(t *testing.T) {
	tooltip := Tooltip{
		ID:    "test-tooltip",
		Child: Text{Content: "Test", Width: Cells(10), Height: Cells(2)},
	}

	fm := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	fc := NewFloatCollector()
	ctx := NewBuildContext(fm, focusedSignal, hoveredSignal, fc)

	node := tooltip.BuildLayoutNode(ctx)
	computed := node.ComputeLayout(layout.Tight(10, 2))

	// Should have the same size as the child
	assert.Equal(t, 10, computed.Box.Width)
	assert.Equal(t, 2, computed.Box.Height)
}

// Test with nil child
func TestTooltip_NilChild(t *testing.T) {
	tooltip := Tooltip{
		ID:    "test-tooltip",
		Child: nil,
	}

	children := tooltip.ChildWidgets()
	assert.Nil(t, children)

	fm := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	hoveredSignal := NewAnySignal[Widget](nil)
	fc := NewFloatCollector()
	ctx := NewBuildContext(fm, focusedSignal, hoveredSignal, fc)

	// Should return empty BoxNode for nil child
	node := tooltip.BuildLayoutNode(ctx)
	assert.NotNil(t, node)

	// isVisible should return false for nil child
	assert.False(t, tooltip.isVisible(ctx))
}

// Test offset calculation for each position
func TestTooltip_OffsetValue(t *testing.T) {
	tests := []struct {
		name     string
		position TooltipPosition
		offset   int
		expected Offset
	}{
		{"Top default", TooltipTop, 0, Offset{Y: 0}},
		{"Top custom", TooltipTop, 3, Offset{Y: -3}},
		{"Bottom default", TooltipBottom, 0, Offset{Y: 0}},
		{"Bottom custom", TooltipBottom, 2, Offset{Y: 2}},
		{"Left default", TooltipLeft, 0, Offset{X: 0}},
		{"Left custom", TooltipLeft, 4, Offset{X: -4}},
		{"Right default", TooltipRight, 0, Offset{X: 0}},
		{"Right custom", TooltipRight, 5, Offset{X: 5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tooltip := Tooltip{
				ID:       "test",
				Content:  "test",
				Position: tt.position,
				Offset:   tt.offset,
				Child:    Text{Content: "target"},
			}
			assert.Equal(t, tt.expected, tooltip.offsetValue())
		})
	}
}

// Test anchor point mapping
func TestTooltip_AnchorPoint(t *testing.T) {
	tests := []struct {
		position TooltipPosition
		expected AnchorPoint
	}{
		{TooltipTop, AnchorTopCenter},
		{TooltipBottom, AnchorBottomCenter},
		{TooltipLeft, AnchorLeftCenter},
		{TooltipRight, AnchorRightCenter},
	}

	for _, tt := range tests {
		tooltip := Tooltip{
			ID:       "test",
			Content:  "test",
			Position: tt.position,
			Child:    Text{Content: "target"},
		}
		assert.Equal(t, tt.expected, tooltip.anchorPoint())
	}
}
