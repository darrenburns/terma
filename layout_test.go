package terma

import (
	"testing"

	"terma/layout"
)

// Mock widget for testing layout behavior.
// IMPORTANT: When layoutSize is set, it represents the widget's CONTENT size
// (the size inside padding/margin/border). The parent layout container
// (Row/Column/Dock) will add the style insets when calculating total space.
type mockWidget struct {
	width      Dimension
	height     Dimension
	style      Style
	layoutSize Size // Content size returned by Layout() (excluding padding/margin/border)
}

func (m mockWidget) Build(ctx BuildContext) Widget {
	return m
}

func (m mockWidget) GetDimensions() (Dimension, Dimension) {
	return m.width, m.height
}

func (m mockWidget) GetStyle() Style {
	return m.style
}

func (m mockWidget) Layout(ctx BuildContext, constraints Constraints) Size {
	// If layoutSize is set, use it; otherwise calculate based on dimensions
	if m.layoutSize.Width > 0 || m.layoutSize.Height > 0 {
		return m.layoutSize
	}

	width := 0
	height := 0

	switch {
	case m.width.IsCells():
		width = m.width.CellsValue()
	case m.width.IsFr():
		width = constraints.MaxWidth
	default: // Auto
		width = 10 // Default auto size
	}

	switch {
	case m.height.IsCells():
		height = m.height.CellsValue()
	case m.height.IsFr():
		height = constraints.MaxHeight
	default: // Auto
		height = 5 // Default auto size
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

func (m mockWidget) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	// Create a BoxNode with the mock widget's size
	width := 0
	height := 0

	if m.layoutSize.Width > 0 || m.layoutSize.Height > 0 {
		width = m.layoutSize.Width
		height = m.layoutSize.Height
	} else {
		switch {
		case m.width.IsCells():
			width = m.width.CellsValue()
		default: // Auto
			width = 10 // Default auto size
		}

		switch {
		case m.height.IsCells():
			height = m.height.CellsValue()
		default: // Auto
			height = 5 // Default auto size
		}
	}

	return &layout.BoxNode{
		Width:   width,
		Height:  height,
		Padding: toLayoutEdgeInsets(m.style.Padding),
		Border:  borderToEdgeInsets(m.style.Border),
		Margin:  toLayoutEdgeInsets(m.style.Margin),
	}
}

// Helper to create test build context
func testContext() BuildContext {
	return BuildContext{}
}

// ROW TESTS

func TestRow_EmptyChildren(t *testing.T) {
	row := Row{Children: []Widget{}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Empty row should have zero width and height (Auto dimension with no content)
	if size.Width != 0 {
		t.Errorf("expected Width = 0 for empty row, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("expected Height = 0 for empty row, got %d", size.Height)
	}
}

func TestRow_SingleChildWithAuto(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row with Auto dimensions should wrap its child
	if size.Width != 20 {
		t.Errorf("expected Width = 20 (child width), got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("expected Height = 10 (child height), got %d", size.Height)
	}
}

func TestRow_SingleChildWithCells(t *testing.T) {
	child := mockWidget{width: Cells(30), height: Cells(15)}
	row := Row{Width: Cells(50), Height: Cells(25), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row with Cells dimensions should use specified size
	if size.Width != 50 {
		t.Errorf("expected Width = 50 (row width), got %d", size.Width)
	}
	if size.Height != 25 {
		t.Errorf("expected Height = 25 (row height), got %d", size.Height)
	}
}

func TestRow_SingleChildWithFr(t *testing.T) {
	child := mockWidget{width: Fr(1), height: Fr(1)}
	row := Row{Width: Fr(1), Height: Fr(1), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row with Fr dimensions should fill available space
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (max width), got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50 (max height), got %d", size.Height)
	}
}

func TestRow_MultipleChildrenWithAuto(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	child2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 30, Height: 15}}
	child3 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 25, Height: 12}}
	row := Row{Children: []Widget{child1, child2, child3}}
	constraints := Constraints{MaxWidth: 200, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Total width should be sum of child widths
	expectedWidth := 20 + 30 + 25
	if size.Width != expectedWidth {
		t.Errorf("expected Width = %d (sum of children), got %d", expectedWidth, size.Width)
	}
	// Height should be max of child heights
	if size.Height != 15 {
		t.Errorf("expected Height = 15 (max child height), got %d", size.Height)
	}
}

func TestRow_WithSpacing(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	child2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 30, Height: 10}}
	child3 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 25, Height: 10}}
	row := Row{Spacing: 5, Children: []Widget{child1, child2, child3}}
	constraints := Constraints{MaxWidth: 200, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Total width = sum of children + spacing between them
	// 20 + 30 + 25 + (2 * 5) = 85
	expectedWidth := 20 + 30 + 25 + (2 * 5)
	if size.Width != expectedWidth {
		t.Errorf("expected Width = %d (children + spacing), got %d", expectedWidth, size.Width)
	}
}

func TestRow_WithFrChildren_EqualDistribution(t *testing.T) {
	child1 := mockWidget{width: Fr(1), height: Cells(10)}
	child2 := mockWidget{width: Fr(1), height: Cells(10)}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row should distribute space equally to Fr(1) children
	// Each child should get 50 width
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (auto wraps to max), got %d", size.Width)
	}
}

func TestRow_WithFrChildren_ProportionalDistribution(t *testing.T) {
	child1 := mockWidget{width: Fr(1), height: Cells(10)}
	child2 := mockWidget{width: Fr(2), height: Cells(10)}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 90, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Total Fr = 3, so child1 gets 30 (1/3), child2 gets 60 (2/3)
	if size.Width != 90 {
		t.Errorf("expected Width = 90, got %d", size.Width)
	}
}

func TestRow_MixedDimensions_AutoAndFr(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	child2 := mockWidget{width: Fr(1), height: Cells(10)}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Auto child takes 20, Fr child gets remaining 80
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
}

func TestRow_MixedDimensions_CellsAndFr(t *testing.T) {
	child1 := mockWidget{width: Cells(30), height: Cells(10)}
	child2 := mockWidget{width: Fr(1), height: Cells(10)}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Cells child takes 30, Fr child gets remaining 70
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
}

func TestRow_WithPadding(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside padding)
		style:      Style{Padding: EdgeInsetsAll(5)},
	}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside padding)
	// Padding adds 10 horizontal (5 left + 5 right), 10 vertical (5 top + 5 bottom)
	// Row adds insets to child size: Total width = 20 + 10 = 30, height = 10 + 10 = 20
	if size.Width != 30 {
		t.Errorf("expected Width = 30 (content + padding), got %d", size.Width)
	}
	if size.Height != 20 {
		t.Errorf("expected Height = 20 (content + padding), got %d", size.Height)
	}
}

func TestRow_WithMargin(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside margin)
		style:      Style{Margin: EdgeInsetsAll(5)},
	}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside margin)
	// Margin adds 10 horizontal (5 left + 5 right), 10 vertical (5 top + 5 bottom)
	// Row adds insets to child size: Total width = 20 + 10 = 30, height = 10 + 10 = 20
	if size.Width != 30 {
		t.Errorf("expected Width = 30 (content + margin), got %d", size.Width)
	}
	if size.Height != 20 {
		t.Errorf("expected Height = 20 (content + margin), got %d", size.Height)
	}
}

func TestRow_WithBorder(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside border)
		style:      Style{Border: SquareBorder(RGB(255, 255, 255))},
	}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside border)
	// Border adds 1 cell on each side: 2 horizontal (left + right), 2 vertical (top + bottom)
	// Row adds border to child size: Total width = 20 + 2 = 22, height = 10 + 2 = 12
	if size.Width != 22 {
		t.Errorf("expected Width = 22 (content + border), got %d", size.Width)
	}
	if size.Height != 12 {
		t.Errorf("expected Height = 12 (content + border), got %d", size.Height)
	}
}

func TestRow_ConstraintClamping_MinWidth(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 10, Height: 10}}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MinWidth: 50, MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row content is 10, but MinWidth is 50
	if size.Width != 50 {
		t.Errorf("expected Width = 50 (clamped to MinWidth), got %d", size.Width)
	}
}

func TestRow_ConstraintClamping_MaxWidth(t *testing.T) {
	child := mockWidget{width: Cells(150), height: Cells(10)}
	row := Row{Width: Cells(150), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Row wants 150, but MaxWidth is 100
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (clamped to MaxWidth), got %d", size.Width)
	}
}

func TestRow_FrWithSpacing(t *testing.T) {
	child1 := mockWidget{width: Fr(1), height: Cells(10)}
	child2 := mockWidget{width: Fr(1), height: Cells(10)}
	row := Row{Spacing: 10, Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Available width = 100 - 10 (spacing) = 90
	// Each Fr(1) child gets 45 width (90 / 2)
	// Total width = 45 + 45 + 10 (spacing) = 100
	// Height = max of children = 10
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("expected Height = 10 (max child height), got %d", size.Height)
	}
}

// COLUMN TESTS

func TestColumn_EmptyChildren(t *testing.T) {
	col := Column{Children: []Widget{}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// Empty column should have zero width and height
	if size.Width != 0 {
		t.Errorf("expected Width = 0 for empty column, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("expected Height = 0 for empty column, got %d", size.Height)
	}
}

func TestColumn_SingleChildWithAuto(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// Column with Auto dimensions should wrap its child
	if size.Width != 20 {
		t.Errorf("expected Width = 20 (child width), got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("expected Height = 10 (child height), got %d", size.Height)
	}
}

func TestColumn_SingleChildWithCells(t *testing.T) {
	child := mockWidget{width: Cells(30), height: Cells(15)}
	col := Column{Width: Cells(50), Height: Cells(25), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// Column with Cells dimensions should use specified size
	if size.Width != 50 {
		t.Errorf("expected Width = 50 (column width), got %d", size.Width)
	}
	if size.Height != 25 {
		t.Errorf("expected Height = 25 (column height), got %d", size.Height)
	}
}

func TestColumn_SingleChildWithFr(t *testing.T) {
	child := mockWidget{width: Fr(1), height: Fr(1)}
	col := Column{Width: Fr(1), Height: Fr(1), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// Column with Fr dimensions should fill available space
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (max width), got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50 (max height), got %d", size.Height)
	}
}

func TestColumn_MultipleChildrenWithAuto(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	child2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 30, Height: 15}}
	child3 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 25, Height: 12}}
	col := Column{Children: []Widget{child1, child2, child3}}
	constraints := Constraints{MaxWidth: 200, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Width should be max of child widths
	if size.Width != 30 {
		t.Errorf("expected Width = 30 (max child width), got %d", size.Width)
	}
	// Total height should be sum of child heights
	expectedHeight := 10 + 15 + 12
	if size.Height != expectedHeight {
		t.Errorf("expected Height = %d (sum of children), got %d", expectedHeight, size.Height)
	}
}

func TestColumn_WithSpacing(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	child2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 15}}
	child3 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 12}}
	col := Column{Spacing: 5, Children: []Widget{child1, child2, child3}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 200}

	size := col.Layout(testContext(), constraints)

	// Total height = sum of children + spacing between them
	// 10 + 15 + 12 + (2 * 5) = 47
	expectedHeight := 10 + 15 + 12 + (2 * 5)
	if size.Height != expectedHeight {
		t.Errorf("expected Height = %d (children + spacing), got %d", expectedHeight, size.Height)
	}
}

func TestColumn_WithFrChildren_EqualDistribution(t *testing.T) {
	child1 := mockWidget{width: Cells(10), height: Fr(1)}
	child2 := mockWidget{width: Cells(10), height: Fr(1)}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Column should distribute space equally to Fr(1) children
	// Each child should get 50 height
	if size.Height != 100 {
		t.Errorf("expected Height = 100 (auto wraps to max), got %d", size.Height)
	}
}

func TestColumn_WithFrChildren_ProportionalDistribution(t *testing.T) {
	child1 := mockWidget{width: Cells(10), height: Fr(1)}
	child2 := mockWidget{width: Cells(10), height: Fr(2)}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 90}

	size := col.Layout(testContext(), constraints)

	// Total Fr = 3, so child1 gets 30 (1/3), child2 gets 60 (2/3)
	if size.Height != 90 {
		t.Errorf("expected Height = 90, got %d", size.Height)
	}
}

func TestColumn_MixedDimensions_AutoAndFr(t *testing.T) {
	child1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 20}}
	child2 := mockWidget{width: Cells(20), height: Fr(1)}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Auto child takes 20 height, Fr child gets remaining 80
	if size.Height != 100 {
		t.Errorf("expected Height = 100, got %d", size.Height)
	}
}

func TestColumn_MixedDimensions_CellsAndFr(t *testing.T) {
	child1 := mockWidget{width: Cells(20), height: Cells(30)}
	child2 := mockWidget{width: Cells(20), height: Fr(1)}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Cells child takes 30 height, Fr child gets remaining 70
	if size.Height != 100 {
		t.Errorf("expected Height = 100, got %d", size.Height)
	}
}

func TestColumn_WithPadding(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside padding)
		style:      Style{Padding: EdgeInsetsAll(5)},
	}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside padding)
	// Padding adds 10 horizontal (5 left + 5 right), 10 vertical (5 top + 5 bottom)
	// Column adds insets to child size: Total width = 20 + 10 = 30, height = 10 + 10 = 20
	if size.Width != 30 {
		t.Errorf("expected Width = 30 (content + padding), got %d", size.Width)
	}
	if size.Height != 20 {
		t.Errorf("expected Height = 20 (content + padding), got %d", size.Height)
	}
}

func TestColumn_WithMargin(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside margin)
		style:      Style{Margin: EdgeInsetsAll(5)},
	}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside margin)
	// Margin adds 10 horizontal (5 left + 5 right), 10 vertical (5 top + 5 bottom)
	// Column adds insets to child size: Total width = 20 + 10 = 30, height = 10 + 10 = 20
	if size.Width != 30 {
		t.Errorf("expected Width = 30 (content + margin), got %d", size.Width)
	}
	if size.Height != 20 {
		t.Errorf("expected Height = 20 (content + margin), got %d", size.Height)
	}
}

func TestColumn_WithBorder(t *testing.T) {
	child := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 10}, // Content size (inside border)
		style:      Style{Border: SquareBorder(RGB(255, 255, 255))},
	}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := col.Layout(testContext(), constraints)

	// layoutSize {20, 10} is the content size (inside border)
	// Border adds 1 cell on each side: 2 horizontal (left + right), 2 vertical (top + bottom)
	// Column adds border to child size: Total width = 20 + 2 = 22, height = 10 + 2 = 12
	if size.Width != 22 {
		t.Errorf("expected Width = 22 (content + border), got %d", size.Width)
	}
	if size.Height != 12 {
		t.Errorf("expected Height = 12 (content + border), got %d", size.Height)
	}
}

func TestColumn_ConstraintClamping_MinHeight(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 10, Height: 10}}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MinHeight: 50, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Column content is 10, but MinHeight is 50
	if size.Height != 50 {
		t.Errorf("expected Height = 50 (clamped to MinHeight), got %d", size.Height)
	}
}

func TestColumn_ConstraintClamping_MaxHeight(t *testing.T) {
	child := mockWidget{width: Cells(10), height: Cells(150)}
	col := Column{Height: Cells(150), Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Column wants 150, but MaxHeight is 100
	if size.Height != 100 {
		t.Errorf("expected Height = 100 (clamped to MaxHeight), got %d", size.Height)
	}
}

func TestColumn_FrWithSpacing(t *testing.T) {
	child1 := mockWidget{width: Cells(10), height: Fr(1)}
	child2 := mockWidget{width: Cells(10), height: Fr(1)}
	col := Column{Spacing: 10, Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Available height = 100 - 10 (spacing) = 90
	// Each Fr(1) child gets 45 height (90 / 2)
	// Total height = 45 + 45 + 10 (spacing) = 100
	// Width = max of children = 10
	if size.Height != 100 {
		t.Errorf("expected Height = 100, got %d", size.Height)
	}
	if size.Width != 10 {
		t.Errorf("expected Width = 10 (max child width), got %d", size.Width)
	}
}

// NESTED LAYOUT TESTS

func TestRow_NestedColumn(t *testing.T) {
	innerChild1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	innerChild2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 15}}
	innerCol := Column{Children: []Widget{innerChild1, innerChild2}}

	row := Row{Children: []Widget{innerCol}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := row.Layout(testContext(), constraints)

	// Inner column should stack children vertically: height = 10 + 15 = 25, width = 20
	if size.Width != 20 {
		t.Errorf("expected Width = 20, got %d", size.Width)
	}
	if size.Height != 25 {
		t.Errorf("expected Height = 25, got %d", size.Height)
	}
}

func TestColumn_NestedRow(t *testing.T) {
	innerChild1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	innerChild2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 30, Height: 10}}
	innerRow := Row{Children: []Widget{innerChild1, innerChild2}}

	col := Column{Children: []Widget{innerRow}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Inner row should stack children horizontally: width = 20 + 30 = 50, height = 10
	if size.Width != 50 {
		t.Errorf("expected Width = 50, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("expected Height = 10, got %d", size.Height)
	}
}

func TestRow_ComplexNesting(t *testing.T) {
	// Create a complex nested structure: Row [ Column [ Row [ child ] ] ]
	innerChild := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	innerRow := Row{Children: []Widget{innerChild}}
	midCol := Column{Children: []Widget{innerRow}}
	outerRow := Row{Children: []Widget{midCol}}

	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}
	size := outerRow.Layout(testContext(), constraints)

	// Should propagate dimensions through nesting
	if size.Width != 20 {
		t.Errorf("expected Width = 20, got %d", size.Width)
	}
	if size.Height != 10 {
		t.Errorf("expected Height = 10, got %d", size.Height)
	}
}

// ALIGNMENT TESTS (these test the Layout pass, not Render)

func TestRow_GetDimensions(t *testing.T) {
	row := Row{Width: Cells(50), Height: Cells(25)}

	width, height := row.GetDimensions()

	if !width.IsCells() || width.CellsValue() != 50 {
		t.Errorf("expected width = Cells(50), got %v", width)
	}
	if !height.IsCells() || height.CellsValue() != 25 {
		t.Errorf("expected height = Cells(25), got %v", height)
	}
}

func TestColumn_GetDimensions(t *testing.T) {
	col := Column{Width: Fr(1), Height: Cells(100)}

	width, height := col.GetDimensions()

	if !width.IsFr() || width.FrValue() != 1 {
		t.Errorf("expected width = Fr(1), got %v", width)
	}
	if !height.IsCells() || height.CellsValue() != 100 {
		t.Errorf("expected height = Cells(100), got %v", height)
	}
}

// EDGE CASE TESTS

func TestRow_ZeroSizeConstraints(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	row := Row{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 0, MaxHeight: 0}

	size := row.Layout(testContext(), constraints)

	// With zero constraints, children get clamped to 0
	// Row wraps content, which is also 0
	if size.Width != 0 {
		t.Errorf("expected Width = 0 with zero constraints, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("expected Height = 0 with zero constraints, got %d", size.Height)
	}
}

func TestColumn_ZeroSizeConstraints(t *testing.T) {
	child := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 10}}
	col := Column{Children: []Widget{child}}
	constraints := Constraints{MaxWidth: 0, MaxHeight: 0}

	size := col.Layout(testContext(), constraints)

	// With zero constraints, children get clamped to 0
	// Column wraps content, which is also 0
	if size.Width != 0 {
		t.Errorf("expected Width = 0 with zero constraints, got %d", size.Width)
	}
	if size.Height != 0 {
		t.Errorf("expected Height = 0 with zero constraints, got %d", size.Height)
	}
}

func TestRow_ChildrenExceedMaxWidth(t *testing.T) {
	// Test what happens when fixed-size children exceed available width
	child1 := mockWidget{width: Cells(60), height: Cells(10)}
	child2 := mockWidget{width: Cells(60), height: Cells(10)}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// Each child is clamped to fit available space
	// After child1 (60), only 40 remains for child2, so child2 gets clamped to 40
	// Total = 60 + 40 = 100
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (children clamped to fit), got %d", size.Width)
	}
}

func TestColumn_ChildrenExceedMaxHeight(t *testing.T) {
	// Test what happens when fixed-size children exceed available height
	child1 := mockWidget{width: Cells(10), height: Cells(60)}
	child2 := mockWidget{width: Cells(10), height: Cells(60)}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// Each child is clamped to fit available space
	// After child1 (60), only 40 remains for child2, so child2 gets clamped to 40
	// Total = 60 + 40 = 100
	if size.Height != 100 {
		t.Errorf("expected Height = 100 (children clamped to fit), got %d", size.Height)
	}
}

func TestRow_FrChildrenWithPaddingDistribution(t *testing.T) {
	// Test that Fr distribution correctly accounts for padding
	child1 := mockWidget{
		width:      Fr(1),
		height:     Cells(10),
		style:      Style{Padding: EdgeInsetsXY(5, 0)}, // 10 horizontal padding
	}
	child2 := mockWidget{
		width:  Fr(1),
		height: Cells(10),
		// No padding
	}
	row := Row{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := row.Layout(testContext(), constraints)

	// child1: Fr(1) with 10 horizontal padding
	// child2: Fr(1) with 0 padding
	// Available width for Fr distribution = 100
	// child1 content gets: ~50 - 10 = ~40, total space: 50
	// child2 content gets: ~50, total space: 50
	// Row total: 100
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
}

func TestColumn_FrChildrenWithMarginDistribution(t *testing.T) {
	// Test that Fr distribution correctly accounts for margin
	child1 := mockWidget{
		width:  Cells(10),
		height: Fr(1),
		style:  Style{Margin: EdgeInsetsXY(0, 5)}, // 10 vertical margin
	}
	child2 := mockWidget{
		width:  Cells(10),
		height: Fr(1),
		// No margin
	}
	col := Column{Children: []Widget{child1, child2}}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 100}

	size := col.Layout(testContext(), constraints)

	// child1: Fr(1) with 10 vertical margin
	// child2: Fr(1) with 0 margin
	// Available height for Fr distribution = 100
	// child1 content gets: ~50 - 10 = ~40, total space: 50
	// child2 content gets: ~50, total space: 50
	// Column total: 100
	if size.Height != 100 {
		t.Errorf("expected Height = 100, got %d", size.Height)
	}
}
