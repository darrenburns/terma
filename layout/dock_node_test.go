package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to create a simple box node for dock tests
func dockBox(w, h int) *BoxNode {
	return &BoxNode{Width: w, Height: h}
}

func TestDockNode_SingleEdge(t *testing.T) {
	t.Run("TopOnly", func(t *testing.T) {
		dock := &DockNode{
			Top: []LayoutNode{dockBox(100, 20)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.Len(t, result.Children, 1)

		// Top child at (0, 0), takes full width
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Height)
	})

	t.Run("BottomOnly", func(t *testing.T) {
		dock := &DockNode{
			Bottom: []LayoutNode{dockBox(100, 20)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 1)

		// Bottom child at (0, 80)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 80, result.Children[0].Y)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Height)
	})

	t.Run("LeftOnly", func(t *testing.T) {
		dock := &DockNode{
			Left: []LayoutNode{dockBox(30, 100)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 1)

		// Left child at (0, 0), takes full height
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Height)
	})

	t.Run("RightOnly", func(t *testing.T) {
		dock := &DockNode{
			Right: []LayoutNode{dockBox(30, 100)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 1)

		// Right child at (70, 0)
		assert.Equal(t, 70, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Height)
	})
}

func TestDockNode_MultipleEdges(t *testing.T) {
	t.Run("TopAndBottom", func(t *testing.T) {
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(100, 20)},
			Bottom: []LayoutNode{dockBox(100, 15)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 2)

		// Top at (0, 0)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Height)

		// Bottom at (0, 85) - 100 - 15 = 85
		assert.Equal(t, 0, result.Children[1].X)
		assert.Equal(t, 85, result.Children[1].Y)
		assert.Equal(t, 15, result.Children[1].Layout.Box.Height)
	})

	t.Run("LeftAndRight", func(t *testing.T) {
		dock := &DockNode{
			Left:  []LayoutNode{dockBox(25, 100)},
			Right: []LayoutNode{dockBox(30, 100)},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 2)

		// Left at (0, 0)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 25, result.Children[0].Layout.Box.Width)

		// Right at (70, 0) - 100 - 30 = 70
		assert.Equal(t, 70, result.Children[1].X)
		assert.Equal(t, 0, result.Children[1].Y)
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)
	})

	t.Run("AllEdges_DefaultOrder", func(t *testing.T) {
		// Default order: Top, Bottom, Left, Right
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(100, 10)},
			Bottom: []LayoutNode{dockBox(100, 15)},
			Left:   []LayoutNode{dockBox(20, 75)},  // height = 100 - 10 - 15 = 75
			Right:  []LayoutNode{dockBox(25, 75)},  // height = 75
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 4)

		// Top: full width at top
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width, "top takes full width")
		assert.Equal(t, 10, result.Children[0].Layout.Box.Height)

		// Bottom: full width at bottom (y = 100 - 15 = 85)
		assert.Equal(t, 0, result.Children[1].X)
		assert.Equal(t, 85, result.Children[1].Y)
		assert.Equal(t, 100, result.Children[1].Layout.Box.Width, "bottom takes full width")

		// Left: after top/bottom consume space (y = 10, height = 75)
		assert.Equal(t, 0, result.Children[2].X)
		assert.Equal(t, 10, result.Children[2].Y)
		assert.Equal(t, 20, result.Children[2].Layout.Box.Width)
		assert.Equal(t, 75, result.Children[2].Layout.Box.Height, "left gets remaining height")

		// Right: after top/bottom/left (x = 100 - 25 = 75)
		assert.Equal(t, 75, result.Children[3].X)
		assert.Equal(t, 10, result.Children[3].Y)
		assert.Equal(t, 25, result.Children[3].Layout.Box.Width)
		assert.Equal(t, 75, result.Children[3].Layout.Box.Height)
	})
}

func TestDockNode_WithBody(t *testing.T) {
	t.Run("BodyFillsRemaining", func(t *testing.T) {
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(100, 20)},
			Bottom: []LayoutNode{dockBox(100, 10)},
			Left:   []LayoutNode{dockBox(15, 70)},
			Right:  []LayoutNode{dockBox(25, 70)},
			Body:   dockBox(0, 0), // Will be stretched to fill
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 5)

		// Body is last child, fills remaining space
		body := result.Children[4]
		assert.Equal(t, 15, body.X, "body x = left width")
		assert.Equal(t, 20, body.Y, "body y = top height")
		// Remaining: width = 100 - 15 - 25 = 60, height = 100 - 20 - 10 = 70
		assert.Equal(t, 60, body.Layout.Box.Width, "body fills remaining width")
		assert.Equal(t, 70, body.Layout.Box.Height, "body fills remaining height")
	})

	t.Run("BodyOnly", func(t *testing.T) {
		dock := &DockNode{
			Body: dockBox(0, 0),
		}

		result := dock.ComputeLayout(Tight(100, 80))

		assert.Len(t, result.Children, 1)

		// Body fills entire container
		body := result.Children[0]
		assert.Equal(t, 0, body.X)
		assert.Equal(t, 0, body.Y)
		assert.Equal(t, 100, body.Layout.Box.Width)
		assert.Equal(t, 80, body.Layout.Box.Height)
	})
}

func TestDockNode_CustomOrder(t *testing.T) {
	t.Run("LeftRightFirst", func(t *testing.T) {
		// Process Left and Right before Top and Bottom
		// This means left/right get full height, top/bottom get reduced width
		dock := &DockNode{
			Top:       []LayoutNode{dockBox(0, 10)},
			Bottom:    []LayoutNode{dockBox(0, 10)},
			Left:      []LayoutNode{dockBox(20, 0)},
			Right:     []LayoutNode{dockBox(30, 0)},
			DockOrder: []DockEdge{DockLeft, DockRight, DockTop, DockBottom},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 4)

		// Left: at (0, 0), full height
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Height, "left gets full height when processed first")

		// Right: at (70, 0), full height
		assert.Equal(t, 70, result.Children[1].X)
		assert.Equal(t, 100, result.Children[1].Layout.Box.Height)

		// Top: reduced width (100 - 20 - 30 = 50)
		assert.Equal(t, 20, result.Children[2].X, "top starts after left")
		assert.Equal(t, 50, result.Children[2].Layout.Box.Width, "top gets reduced width")

		// Bottom: also reduced width
		assert.Equal(t, 20, result.Children[3].X)
		assert.Equal(t, 50, result.Children[3].Layout.Box.Width)
	})
}

func TestDockNode_MultipleChildrenPerEdge(t *testing.T) {
	t.Run("TwoTops", func(t *testing.T) {
		dock := &DockNode{
			Top: []LayoutNode{
				dockBox(100, 15), // First header
				dockBox(100, 10), // Second header
			},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 2)

		// First top at y=0
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 15, result.Children[0].Layout.Box.Height)

		// Second top at y=15
		assert.Equal(t, 15, result.Children[1].Y)
		assert.Equal(t, 10, result.Children[1].Layout.Box.Height)
	})

	t.Run("TwoLefts", func(t *testing.T) {
		dock := &DockNode{
			Left: []LayoutNode{
				dockBox(20, 100),
				dockBox(15, 100),
			},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 2)

		// First left at x=0
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)

		// Second left at x=20
		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 15, result.Children[1].Layout.Box.Width)
	})
}

func TestDockNode_Empty(t *testing.T) {
	t.Run("NoChildren", func(t *testing.T) {
		dock := &DockNode{}

		result := dock.ComputeLayout(Tight(100, 80))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 80, result.Box.Height)
		assert.Empty(t, result.Children)
	})
}

func TestDockNode_WithPadding(t *testing.T) {
	t.Run("PaddingReducesContentArea", func(t *testing.T) {
		dock := &DockNode{
			Padding: EdgeInsets{Top: 5, Right: 10, Bottom: 5, Left: 10},
			Body:    dockBox(0, 0),
		}

		result := dock.ComputeLayout(Tight(100, 100))

		// Container is 100x100, but content area is 80x90 due to padding
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)

		// Body fills content area: 100 - 20 = 80 width, 100 - 10 = 90 height
		body := result.Children[0]
		assert.Equal(t, 80, body.Layout.Box.Width)
		assert.Equal(t, 90, body.Layout.Box.Height)
	})
}

func TestDockNode_WithBorder(t *testing.T) {
	t.Run("BorderReducesContentArea", func(t *testing.T) {
		dock := &DockNode{
			Border: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Body:   dockBox(0, 0),
		}

		result := dock.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)

		// Body fills content area: 100 - 2 = 98 each dimension
		body := result.Children[0]
		assert.Equal(t, 98, body.Layout.Box.Width)
		assert.Equal(t, 98, body.Layout.Box.Height)
	})
}

func TestDockNode_WithMargin(t *testing.T) {
	t.Run("ChildMarginAffectsPosition", func(t *testing.T) {
		dock := &DockNode{
			Top: []LayoutNode{
				&BoxNode{
					Width:  100,
					Height: 20,
					Margin: EdgeInsets{Top: 5, Left: 10},
				},
			},
		}

		result := dock.ComputeLayout(Tight(100, 100))

		// Child position accounts for margin
		assert.Equal(t, 10, result.Children[0].X, "x offset by left margin")
		assert.Equal(t, 5, result.Children[0].Y, "y offset by top margin")
	})

	t.Run("ChildMarginConsumesSpace", func(t *testing.T) {
		dock := &DockNode{
			Top: []LayoutNode{
				&BoxNode{
					Width:  100,
					Height: 20,
					Margin: EdgeInsets{Top: 5, Bottom: 10},
				},
			},
			Body: dockBox(0, 0),
		}

		result := dock.ComputeLayout(Tight(100, 100))

		// Top consumes 20 + 5 + 10 = 35 vertical space
		// Body starts at y = 35
		body := result.Children[1]
		assert.Equal(t, 35, body.Y)
		assert.Equal(t, 65, body.Layout.Box.Height, "body gets remaining height")
	})
}

func TestDockNode_Constraints(t *testing.T) {
	t.Run("MinWidth", func(t *testing.T) {
		dock := &DockNode{
			MinWidth: 80,
			Body:     dockBox(0, 0),
		}

		result := dock.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 100, result.Box.Width, "uses max when unconstrained")
	})

	t.Run("MaxWidth", func(t *testing.T) {
		dock := &DockNode{
			MaxWidth: 60,
			Body:     dockBox(0, 0),
		}

		result := dock.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 60, result.Box.Width, "clamped to own MaxWidth")
		assert.Equal(t, 60, result.Children[0].Layout.Box.Width, "body gets reduced width")
	})

	t.Run("MaxHeight", func(t *testing.T) {
		dock := &DockNode{
			MaxHeight: 50,
			Top:       []LayoutNode{dockBox(100, 20)},
			Body:      dockBox(0, 0),
		}

		result := dock.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 50, result.Box.Height)

		// Top still gets 20
		assert.Equal(t, 20, result.Children[0].Layout.Box.Height)

		// Body gets remaining: 50 - 20 = 30
		assert.Equal(t, 30, result.Children[1].Layout.Box.Height)
	})
}

func TestDockNode_RealWorldScenarios(t *testing.T) {
	t.Run("AppLayout_HeaderFooterBody", func(t *testing.T) {
		// Typical app: header at top, footer at bottom, content fills middle
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(0, 3)},  // Header: 3 rows
			Bottom: []LayoutNode{dockBox(0, 1)},  // Footer: 1 row (keybind bar)
			Body:   dockBox(0, 0),                 // Content fills remainder
		}

		result := dock.ComputeLayout(Tight(80, 24)) // 80x24 terminal

		assert.Equal(t, 80, result.Box.Width)
		assert.Equal(t, 24, result.Box.Height)

		// Header
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 80, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 3, result.Children[0].Layout.Box.Height)

		// Footer
		assert.Equal(t, 23, result.Children[1].Y) // 24 - 1 = 23
		assert.Equal(t, 80, result.Children[1].Layout.Box.Width)
		assert.Equal(t, 1, result.Children[1].Layout.Box.Height)

		// Body
		assert.Equal(t, 3, result.Children[2].Y)
		assert.Equal(t, 80, result.Children[2].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[2].Layout.Box.Height) // 24 - 3 - 1 = 20
	})

	t.Run("AppLayout_WithSidebar", func(t *testing.T) {
		// Header, footer, left sidebar, main content
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(0, 2)},  // Header
			Bottom: []LayoutNode{dockBox(0, 1)},  // Footer
			Left:   []LayoutNode{dockBox(20, 0)}, // Sidebar: 20 cols wide
			Body:   dockBox(0, 0),                 // Main content
		}

		result := dock.ComputeLayout(Tight(80, 24))

		// Sidebar: y=2 (after header), height=21 (24-2-1)
		sidebar := result.Children[2]
		assert.Equal(t, 0, sidebar.X)
		assert.Equal(t, 2, sidebar.Y)
		assert.Equal(t, 20, sidebar.Layout.Box.Width)
		assert.Equal(t, 21, sidebar.Layout.Box.Height)

		// Main content: x=20, width=60 (80-20)
		body := result.Children[3]
		assert.Equal(t, 20, body.X)
		assert.Equal(t, 2, body.Y)
		assert.Equal(t, 60, body.Layout.Box.Width)
		assert.Equal(t, 21, body.Layout.Box.Height)
	})

	t.Run("IDE_Layout", func(t *testing.T) {
		// IDE-style: toolbar, status bar, file tree, editor, terminal
		dock := &DockNode{
			Top:    []LayoutNode{dockBox(0, 1)},  // Toolbar
			Bottom: []LayoutNode{dockBox(0, 1)},  // Status bar
			Left:   []LayoutNode{dockBox(25, 0)}, // File tree
			Body:   &DockNode{ // Editor + Terminal area
				Bottom: []LayoutNode{dockBox(0, 8)}, // Terminal: 8 rows
				Body:   dockBox(0, 0),                // Editor
			},
		}

		result := dock.ComputeLayout(Tight(120, 40))

		// Main container
		assert.Equal(t, 120, result.Box.Width)
		assert.Equal(t, 40, result.Box.Height)

		// File tree
		fileTree := result.Children[2]
		assert.Equal(t, 0, fileTree.X)
		assert.Equal(t, 1, fileTree.Y)
		assert.Equal(t, 25, fileTree.Layout.Box.Width)
		assert.Equal(t, 38, fileTree.Layout.Box.Height) // 40 - 1 - 1

		// Inner dock (editor + terminal area)
		innerDock := result.Children[3]
		assert.Equal(t, 25, innerDock.X)
		assert.Equal(t, 1, innerDock.Y)
		assert.Equal(t, 95, innerDock.Layout.Box.Width)  // 120 - 25
		assert.Equal(t, 38, innerDock.Layout.Box.Height) // 40 - 1 - 1

		// Terminal (inside inner dock)
		terminal := innerDock.Layout.Children[0]
		assert.Equal(t, 8, terminal.Layout.Box.Height)

		// Editor (inside inner dock)
		editor := innerDock.Layout.Children[1]
		assert.Equal(t, 30, editor.Layout.Box.Height) // 38 - 8
	})
}
