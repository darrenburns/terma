package terma

import "testing"

// DOCK TESTS

func TestDock_EmptyDock(t *testing.T) {
	dock := Dock{}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Empty dock should fill available space (Auto defaults to max for Dock)
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_BodyOnly(t *testing.T) {
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 50, Height: 30}}
	dock := Dock{Body: body}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Dock fills available space even with just body
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_TopEdge_SingleWidget(t *testing.T) {
	top := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 30}}
	dock := Dock{
		Top:  []Widget{top},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Dock should consume full space
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_BottomEdge_SingleWidget(t *testing.T) {
	bottom := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 30}}
	dock := Dock{
		Bottom: []Widget{bottom},
		Body:   body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_LeftEdge_SingleWidget(t *testing.T) {
	left := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 50}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 50}}
	dock := Dock{
		Left: []Widget{left},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_RightEdge_SingleWidget(t *testing.T) {
	right := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 50}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 50}}
	dock := Dock{
		Right: []Widget{right},
		Body:  body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_AllEdges(t *testing.T) {
	top := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}
	bottom := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}
	left := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 15, Height: 30}}
	right := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 15, Height: 30}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 30}}

	dock := Dock{
		Top:    []Widget{top},
		Bottom: []Widget{bottom},
		Left:   []Widget{left},
		Right:  []Widget{right},
		Body:   body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Dock should fill available space
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_MultipleWidgetsPerEdge_Top(t *testing.T) {
	top1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 8}}
	top2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 7}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 30}}

	dock := Dock{
		Top:  []Widget{top1, top2},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Multiple top widgets should stack and consume space
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_MultipleWidgetsPerEdge_Left(t *testing.T) {
	left1 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 12, Height: 50}}
	left2 := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 13, Height: 50}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 50}}

	dock := Dock{
		Left: []Widget{left1, left2},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Multiple left widgets should stack horizontally
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_DefaultDockOrder(t *testing.T) {
	dock := Dock{}
	order := dock.dockOrder()

	// Default order should be Top, Bottom, Left, Right
	if len(order) != 4 {
		t.Fatalf("expected 4 edges in default order, got %d", len(order))
	}
	if order[0] != Top {
		t.Errorf("expected first edge to be Top, got %v", order[0])
	}
	if order[1] != Bottom {
		t.Errorf("expected second edge to be Bottom, got %v", order[1])
	}
	if order[2] != Left {
		t.Errorf("expected third edge to be Left, got %v", order[2])
	}
	if order[3] != Right {
		t.Errorf("expected fourth edge to be Right, got %v", order[3])
	}
}

func TestDock_CustomDockOrder(t *testing.T) {
	customOrder := []Edge{Left, Right, Top, Bottom}
	dock := Dock{DockOrder: customOrder}
	order := dock.dockOrder()

	// Should return custom order
	if len(order) != 4 {
		t.Fatalf("expected 4 edges in custom order, got %d", len(order))
	}
	if order[0] != Left {
		t.Errorf("expected first edge to be Left, got %v", order[0])
	}
	if order[1] != Right {
		t.Errorf("expected second edge to be Right, got %v", order[1])
	}
	if order[2] != Top {
		t.Errorf("expected third edge to be Top, got %v", order[2])
	}
	if order[3] != Bottom {
		t.Errorf("expected fourth edge to be Bottom, got %v", order[3])
	}
}

func TestDock_EdgeWidgets_Top(t *testing.T) {
	widget1 := mockWidget{width: Auto, height: Auto}
	widget2 := mockWidget{width: Auto, height: Auto}
	dock := Dock{Top: []Widget{widget1, widget2}}

	widgets := dock.edgeWidgets(Top)

	if len(widgets) != 2 {
		t.Fatalf("expected 2 top widgets, got %d", len(widgets))
	}
}

func TestDock_EdgeWidgets_Bottom(t *testing.T) {
	widget1 := mockWidget{width: Auto, height: Auto}
	dock := Dock{Bottom: []Widget{widget1}}

	widgets := dock.edgeWidgets(Bottom)

	if len(widgets) != 1 {
		t.Fatalf("expected 1 bottom widget, got %d", len(widgets))
	}
}

func TestDock_EdgeWidgets_Left(t *testing.T) {
	widget1 := mockWidget{width: Auto, height: Auto}
	widget2 := mockWidget{width: Auto, height: Auto}
	widget3 := mockWidget{width: Auto, height: Auto}
	dock := Dock{Left: []Widget{widget1, widget2, widget3}}

	widgets := dock.edgeWidgets(Left)

	if len(widgets) != 3 {
		t.Fatalf("expected 3 left widgets, got %d", len(widgets))
	}
}

func TestDock_EdgeWidgets_Right(t *testing.T) {
	dock := Dock{Right: []Widget{}}

	widgets := dock.edgeWidgets(Right)

	if len(widgets) != 0 {
		t.Fatalf("expected 0 right widgets, got %d", len(widgets))
	}
}

func TestDock_WithPadding(t *testing.T) {
	top := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 100, Height: 10},
		style:      Style{Padding: EdgeInsetsAll(2)},
	}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 30}}

	dock := Dock{
		Top:  []Widget{top},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Top widget with padding should consume 10 (content) + 4 (padding vertical) = 14 height
	// Body should get remaining 36 height
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_WithMargin(t *testing.T) {
	left := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 20, Height: 50},
		style:      Style{Margin: EdgeInsetsAll(3)},
	}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 50}}

	dock := Dock{
		Left: []Widget{left},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Left widget with margin should consume 20 (content) + 6 (margin horizontal) = 26 width
	// Body should get remaining 74 width
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_WithBorder(t *testing.T) {
	top := mockWidget{
		width:      Auto,
		height:     Auto,
		layoutSize: Size{Width: 100, Height: 10},
		style:      Style{Border: RoundedBorder(RGB(255, 255, 255))},
	}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 30}}

	dock := Dock{
		Top:  []Widget{top},
		Body: body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Top widget with border should consume 10 (content) + 2 (border vertical) = 12 height
	// Body should get remaining 38 height
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_ConstraintClamping_WithCellsDimensions(t *testing.T) {
	dock := Dock{
		Width:  Cells(150),
		Height: Cells(150),
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Dock wants 150x150, but should be clamped to constraints
	if size.Width != 100 {
		t.Errorf("expected Width = 100 (clamped), got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50 (clamped), got %d", size.Height)
	}
}

func TestDock_WithFrDimensions(t *testing.T) {
	dock := Dock{
		Width:  Fr(1),
		Height: Fr(1),
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Fr(1) should fill available space
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}

func TestDock_GetDimensions(t *testing.T) {
	dock := Dock{Width: Cells(80), Height: Fr(1)}

	width, height := dock.GetDimensions()

	if !width.IsCells() || width.CellsValue() != 80 {
		t.Errorf("expected width = Cells(80), got %v", width)
	}
	if !height.IsFr() || height.FrValue() != 1 {
		t.Errorf("expected height = Fr(1), got %v", height)
	}
}

func TestDock_WidgetID(t *testing.T) {
	dock := Dock{ID: "test-dock"}

	if dock.WidgetID() != "test-dock" {
		t.Errorf("expected WidgetID = 'test-dock', got '%s'", dock.WidgetID())
	}
}

func TestDock_PartialDockOrder(t *testing.T) {
	// Test with partial dock order (only 2 edges specified)
	customOrder := []Edge{Left, Top}
	dock := Dock{DockOrder: customOrder}
	order := dock.dockOrder()

	// Should return exactly what was specified
	if len(order) != 2 {
		t.Fatalf("expected 2 edges in custom order, got %d", len(order))
	}
	if order[0] != Left {
		t.Errorf("expected first edge to be Left, got %v", order[0])
	}
	if order[1] != Top {
		t.Errorf("expected second edge to be Top, got %v", order[1])
	}
}

func TestDock_DockOrderAffectsLayout(t *testing.T) {
	// This tests that dock order actually affects space allocation
	// When Top is processed first, it gets full width
	// When Left is processed first, Top gets reduced width

	top := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}
	left := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 20, Height: 50}}
	body := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 70, Height: 30}}

	// Order 1: Top, Left (default-like)
	dock1 := Dock{
		DockOrder: []Edge{Top, Left},
		Top:       []Widget{top},
		Left:      []Widget{left},
		Body:      body,
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}
	size1 := dock1.Layout(testContext(), constraints)

	// Order 2: Left, Top
	dock2 := Dock{
		DockOrder: []Edge{Left, Top},
		Top:       []Widget{top},
		Left:      []Widget{left},
		Body:      body,
	}
	size2 := dock2.Layout(testContext(), constraints)

	// Both should produce same final dock size (fill constraints)
	if size1.Width != 100 || size1.Height != 50 {
		t.Errorf("dock1: expected 100x50, got %dx%d", size1.Width, size1.Height)
	}
	if size2.Width != 100 || size2.Height != 50 {
		t.Errorf("dock2: expected 100x50, got %dx%d", size2.Width, size2.Height)
	}
}

func TestDock_NoBody(t *testing.T) {
	top := mockWidget{width: Auto, height: Auto, layoutSize: Size{Width: 100, Height: 10}}

	dock := Dock{
		Top: []Widget{top},
		// No body
	}
	constraints := Constraints{MaxWidth: 100, MaxHeight: 50}

	size := dock.Layout(testContext(), constraints)

	// Should still work without a body
	if size.Width != 100 {
		t.Errorf("expected Width = 100, got %d", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected Height = 50, got %d", size.Height)
	}
}
