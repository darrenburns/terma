package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to create a simple fixed-size BoxNode
func box(w, h int) *BoxNode {
	return &BoxNode{Width: w, Height: h}
}

// Helper to create a BoxNode with margin
func boxWithMargin(w, h int, margin EdgeInsets) *BoxNode {
	return &BoxNode{Width: w, Height: h, Margin: margin}
}

func TestLinearNode_EmptyChildren(t *testing.T) {
	t.Run("Row", func(t *testing.T) {
		row := &RowNode{}
		result := row.ComputeLayout(Loose(100, 50))

		assert.Equal(t, 0, result.Box.Width)
		assert.Equal(t, 0, result.Box.Height)
		assert.Nil(t, result.Children)
	})

	t.Run("Column", func(t *testing.T) {
		col := &ColumnNode{}
		result := col.ComputeLayout(Loose(100, 50))

		assert.Equal(t, 0, result.Box.Width)
		assert.Equal(t, 0, result.Box.Height)
		assert.Nil(t, result.Children)
	})

	t.Run("EmptyWithPadding", func(t *testing.T) {
		row := &RowNode{
			Padding: EdgeInsets{Top: 5, Right: 10, Bottom: 5, Left: 10},
		}
		result := row.ComputeLayout(Loose(100, 50))

		// Empty but has padding
		assert.Equal(t, 20, result.Box.Width)  // just padding
		assert.Equal(t, 10, result.Box.Height) // just padding
	})
}

func TestLinearNode_BasicLayout(t *testing.T) {
	t.Run("Row_SingleChild", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{box(30, 20)},
		}
		result := row.ComputeLayout(Loose(100, 50))

		assert.Equal(t, 30, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)
		assert.Len(t, result.Children, 1)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
	})

	t.Run("Row_MultipleChildren", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),
				box(30, 15),
				box(25, 20),
			},
		}
		result := row.ComputeLayout(Loose(200, 50))

		// Total width = 20 + 30 + 25 = 75
		// Max height = 20
		assert.Equal(t, 75, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)

		// Children positioned left to right
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 50, result.Children[2].X)
	})

	t.Run("Column_MultipleChildren", func(t *testing.T) {
		col := &ColumnNode{
			Children: []LayoutNode{
				box(20, 10),
				box(30, 15),
				box(25, 20),
			},
		}
		result := col.ComputeLayout(Loose(50, 200))

		// Max width = 30
		// Total height = 10 + 15 + 20 = 45
		assert.Equal(t, 30, result.Box.Width)
		assert.Equal(t, 45, result.Box.Height)

		// Children positioned top to bottom
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 10, result.Children[1].Y)
		assert.Equal(t, 25, result.Children[2].Y)
	})
}

func TestLinearNode_Spacing(t *testing.T) {
	t.Run("Row_WithSpacing", func(t *testing.T) {
		row := &RowNode{
			Spacing: 5,
			Children: []LayoutNode{
				box(20, 10),
				box(30, 10),
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Loose(200, 50))

		// Total width = 20 + 5 + 30 + 5 + 20 = 80
		assert.Equal(t, 80, result.Box.Width)

		// Positions include spacing
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 25, result.Children[1].X)  // 20 + 5
		assert.Equal(t, 60, result.Children[2].X)  // 25 + 30 + 5
	})

	t.Run("Column_WithSpacing", func(t *testing.T) {
		col := &ColumnNode{
			Spacing: 10,
			Children: []LayoutNode{
				box(20, 15),
				box(20, 25),
			},
		}
		result := col.ComputeLayout(Loose(50, 200))

		// Total height = 15 + 10 + 25 = 50
		assert.Equal(t, 50, result.Box.Height)

		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 25, result.Children[1].Y) // 15 + 10
	})
}

func TestLinearNode_MainAxisAlignment(t *testing.T) {
	// Setup: 3 children of 20px each = 60px, in a 100px container
	// Extra space = 40px
	makeRow := func(align MainAxisAlignment) *RowNode {
		return &RowNode{
			MainAlign: align,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 10),
				box(20, 10),
			},
		}
	}

	t.Run("Start", func(t *testing.T) {
		result := makeRow(MainAxisStart).ComputeLayout(Tight(100, 50))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 40, result.Children[2].X)
	})

	t.Run("Center", func(t *testing.T) {
		result := makeRow(MainAxisCenter).ComputeLayout(Tight(100, 50))

		// Extra 40px, centered = start at 20
		assert.Equal(t, 20, result.Children[0].X)
		assert.Equal(t, 40, result.Children[1].X)
		assert.Equal(t, 60, result.Children[2].X)
	})

	t.Run("End", func(t *testing.T) {
		result := makeRow(MainAxisEnd).ComputeLayout(Tight(100, 50))

		// Extra 40px at start
		assert.Equal(t, 40, result.Children[0].X)
		assert.Equal(t, 60, result.Children[1].X)
		assert.Equal(t, 80, result.Children[2].X)
	})

	t.Run("SpaceBetween", func(t *testing.T) {
		result := makeRow(MainAxisSpaceBetween).ComputeLayout(Tight(100, 50))

		// 40px extra / 2 gaps = 20px each
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 40, result.Children[1].X)  // 0 + 20 + 20
		assert.Equal(t, 80, result.Children[2].X)  // 40 + 20 + 20
	})

	t.Run("SpaceAround", func(t *testing.T) {
		result := makeRow(MainAxisSpaceAround).ComputeLayout(Tight(100, 50))

		// 40px extra, 3 children → each "unit" = 40/3 ≈ 13.33
		// Edge gaps = half unit, between gaps = full unit
		// Cumulative distribution spreads remainder across positions:
		// Child 0: 40*1/6 = 6
		// Child 1: 40*3/6 = 20, + child0 width = 40
		// Child 2: 40*5/6 = 33, + children 0-1 width = 73
		assert.Equal(t, 6, result.Children[0].X)
		assert.Equal(t, 40, result.Children[1].X)
		assert.Equal(t, 73, result.Children[2].X)
	})

	t.Run("SpaceEvenly", func(t *testing.T) {
		result := makeRow(MainAxisSpaceEvenly).ComputeLayout(Tight(100, 50))

		// 40px extra / 4 gaps = 10px each
		assert.Equal(t, 10, result.Children[0].X)
		assert.Equal(t, 40, result.Children[1].X)  // 10 + 20 + 10
		assert.Equal(t, 70, result.Children[2].X)  // 40 + 20 + 10
	})

	t.Run("SpaceEvenly_NoPixelLoss", func(t *testing.T) {
		// 3 children of 21 each = 63, container = 100
		// Extra = 37, 4 gaps → 37/4 = 9.25 → must distribute remainder
		// Gaps will be: 9, 9, 9, 10 (last gap gets the extra pixel)
		row := &RowNode{
			MainAlign: MainAxisSpaceEvenly,
			Children: []LayoutNode{
				box(21, 10),
				box(21, 10),
				box(21, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// SpaceEvenly has a trailing gap, so last child is NOT flush.
		// Instead verify all extra space is used (gaps sum to 37)
		// Positions: 9, 9+21+9=39, 39+21+9=69
		assert.Equal(t, 9, result.Children[0].X)
		assert.Equal(t, 39, result.Children[1].X)
		assert.Equal(t, 69, result.Children[2].X)

		// Last child ends at 69+21=90, trailing gap is 10, total=100
		lastChildEnd := result.Children[2].X + result.Children[2].Layout.Box.Width
		trailingGap := 100 - lastChildEnd
		assert.Equal(t, 90, lastChildEnd)
		assert.Equal(t, 10, trailingGap, "trailing gap gets the remainder")
	})

	t.Run("SpaceBetween_NoPixelLoss", func(t *testing.T) {
		// Children: 20 + 21 + 20 = 61, container = 100
		// Extra = 39, 2 gaps → 39/2 = 19.5 → must distribute remainder
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				box(20, 10),
				box(21, 10),
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// First child at 0
		assert.Equal(t, 0, result.Children[0].X)

		// Last child must end exactly at container width
		lastChild := result.Children[2]
		lastChildEnd := lastChild.X + lastChild.Layout.Box.Width
		assert.Equal(t, 100, lastChildEnd, "last child should be flush with container end")
	})

	t.Run("SpaceBetween_GapsDistributedEvenly", func(t *testing.T) {
		// 4 children of 20 width each = 80, container = 100
		// Extra = 20, 3 gaps → 20/3 = 6.67
		// Gaps should be 6 or 7 (differ by at most 1)
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 10),
				box(20, 10),
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Calculate actual gaps
		gaps := []int{
			result.Children[1].X - (result.Children[0].X + 20), // gap 0-1
			result.Children[2].X - (result.Children[1].X + 20), // gap 1-2
			result.Children[3].X - (result.Children[2].X + 20), // gap 2-3
		}

		// Find min and max gap
		minGap, maxGap := gaps[0], gaps[0]
		for _, g := range gaps {
			if g < minGap {
				minGap = g
			}
			if g > maxGap {
				maxGap = g
			}
		}

		// Key uniformity check: no gap differs from another by more than 1
		assert.LessOrEqual(t, maxGap-minGap, 1, "gaps should differ by at most 1 cell")

		// All gaps should be floor or ceil of 20/3
		for i, gap := range gaps {
			assert.GreaterOrEqual(t, gap, 6, "gap %d too small", i)
			assert.LessOrEqual(t, gap, 7, "gap %d too large", i)
		}

		// Verify total: gaps + children = container width
		assert.Equal(t, 100, result.Box.Width)
	})

	t.Run("SpaceEvenly_GapsDistributedEvenly", func(t *testing.T) {
		// 3 children of 21 width each = 63, container = 100
		// Extra = 37, 4 gaps (n+1) → 37/4 = 9.25
		// Gaps should be 9 or 10 (differ by at most 1)
		row := &RowNode{
			MainAlign: MainAxisSpaceEvenly,
			Children: []LayoutNode{
				box(21, 10),
				box(21, 10),
				box(21, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Calculate actual gaps
		gaps := []int{
			result.Children[0].X,                               // leading gap
			result.Children[1].X - (result.Children[0].X + 21), // gap 0-1
			result.Children[2].X - (result.Children[1].X + 21), // gap 1-2
			100 - (result.Children[2].X + 21),                  // trailing gap
		}

		// Find min and max gap
		minGap, maxGap := gaps[0], gaps[0]
		for _, g := range gaps {
			if g < minGap {
				minGap = g
			}
			if g > maxGap {
				maxGap = g
			}
		}

		// Key uniformity check: no gap differs from another by more than 1
		assert.LessOrEqual(t, maxGap-minGap, 1, "gaps should differ by at most 1 cell")

		// All gaps should be floor or ceil of 37/4
		for i, gap := range gaps {
			assert.GreaterOrEqual(t, gap, 9, "gap %d too small", i)
			assert.LessOrEqual(t, gap, 10, "gap %d too large", i)
		}

		// Verify gaps sum to total extra space
		totalGaps := 0
		for _, g := range gaps {
			totalGaps += g
		}
		assert.Equal(t, 37, totalGaps, "gaps should sum to extra space")

		// Verify container size
		assert.Equal(t, 100, result.Box.Width)
	})
}

func TestLinearNode_CrossAxisAlignment(t *testing.T) {
	// Setup: children of different heights in a row
	makeRow := func(align CrossAxisAlignment) *RowNode {
		return &RowNode{
			CrossAlign: align,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 20),
				box(20, 15),
			},
		}
	}

	t.Run("Start", func(t *testing.T) {
		result := makeRow(CrossAxisStart).ComputeLayout(Tight(100, 30))

		// All children at Y=0
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 0, result.Children[1].Y)
		assert.Equal(t, 0, result.Children[2].Y)
	})

	t.Run("Center", func(t *testing.T) {
		result := makeRow(CrossAxisCenter).ComputeLayout(Tight(100, 30))

		// Container cross = 30
		// Child 0: height 10, centered = (30-10)/2 = 10
		// Child 1: height 20, centered = (30-20)/2 = 5
		// Child 2: height 15, centered = (30-15)/2 = 7
		assert.Equal(t, 10, result.Children[0].Y)
		assert.Equal(t, 5, result.Children[1].Y)
		assert.Equal(t, 7, result.Children[2].Y)
	})

	t.Run("End", func(t *testing.T) {
		result := makeRow(CrossAxisEnd).ComputeLayout(Tight(100, 30))

		// Child 0: height 10, at bottom = 30 - 10 = 20
		// Child 1: height 20, at bottom = 30 - 20 = 10
		// Child 2: height 15, at bottom = 30 - 15 = 15
		assert.Equal(t, 20, result.Children[0].Y)
		assert.Equal(t, 10, result.Children[1].Y)
		assert.Equal(t, 15, result.Children[2].Y)
	})

	t.Run("Stretch", func(t *testing.T) {
		result := makeRow(CrossAxisStretch).ComputeLayout(Tight(100, 30))

		// All children should be stretched to height 30
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 0, result.Children[1].Y)
		assert.Equal(t, 0, result.Children[2].Y)

		// And their heights should be 30
		assert.Equal(t, 30, result.Children[0].Layout.Box.Height)
		assert.Equal(t, 30, result.Children[1].Layout.Box.Height)
		assert.Equal(t, 30, result.Children[2].Layout.Box.Height)
	})

	t.Run("Stretch_DoesNotForceMainAxisMin", func(t *testing.T) {
		// Row with tight constraints (100x50)
		// Child is 20x10 naturally
		// With CrossAxisStretch, child should stretch to height 50 (cross-axis)
		// BUT should NOT be forced to width 100 (the parent's main-axis min)
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// Child should be stretched on cross-axis (height = 50)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Height)

		// Child should NOT be forced to expand on main-axis
		// It should remain at its natural width of 20
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width,
			"stretch should only affect cross-axis, not force main-axis to parent's min")
	})

	t.Run("Stretch_LooseConstraints_ChildrenMatchTallest", func(t *testing.T) {
		// Design decision: With loose constraints, container shrink-wraps to children.
		// Stretch then makes children match each other (the tallest), not fill parent.
		// This matches Flutter and CSS Flexbox behavior.
		// If user wants children to fill parent, they should use tight constraints.
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 25), // tallest
				box(20, 15),
			},
		}
		result := row.ComputeLayout(Loose(100, 50))

		// Container shrink-wraps to tallest child (25), not parent's max (50)
		assert.Equal(t, 25, result.Box.Height,
			"container should shrink-wrap to tallest child with loose constraints")

		// All children stretched to match tallest (25)
		assert.Equal(t, 25, result.Children[0].Layout.Box.Height)
		assert.Equal(t, 25, result.Children[1].Layout.Box.Height)
		assert.Equal(t, 25, result.Children[2].Layout.Box.Height)
	})

	t.Run("Stretch_TightConstraints_ChildrenFillContainer", func(t *testing.T) {
		// With tight constraints, container fills the space.
		// Stretch then makes children fill the container.
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 25),
				box(20, 15),
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// Container fills tight constraint (50)
		assert.Equal(t, 50, result.Box.Height,
			"container should fill tight constraint")

		// All children stretched to fill container (50)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Height)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Height)
		assert.Equal(t, 50, result.Children[2].Layout.Box.Height)
	})
}

func TestLinearNode_ColumnCrossAxisAlignment(t *testing.T) {
	// Verify cross-axis works correctly for Column (horizontal alignment)
	makeCol := func(align CrossAxisAlignment) *ColumnNode {
		return &ColumnNode{
			CrossAlign: align,
			Children: []LayoutNode{
				box(10, 20),
				box(20, 20),
				box(15, 20),
			},
		}
	}

	t.Run("Start", func(t *testing.T) {
		result := makeCol(CrossAxisStart).ComputeLayout(Tight(30, 100))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[1].X)
		assert.Equal(t, 0, result.Children[2].X)
	})

	t.Run("Center", func(t *testing.T) {
		result := makeCol(CrossAxisCenter).ComputeLayout(Tight(30, 100))

		assert.Equal(t, 10, result.Children[0].X) // (30-10)/2
		assert.Equal(t, 5, result.Children[1].X)  // (30-20)/2
		assert.Equal(t, 7, result.Children[2].X)  // (30-15)/2
	})

	t.Run("Stretch", func(t *testing.T) {
		result := makeCol(CrossAxisStretch).ComputeLayout(Tight(30, 100))

		// All children stretched to width 30
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)
		assert.Equal(t, 30, result.Children[2].Layout.Box.Width)
	})
}

func TestLinearNode_ContainerInsets(t *testing.T) {
	t.Run("Padding", func(t *testing.T) {
		row := &RowNode{
			Padding: EdgeInsets{Top: 5, Right: 10, Bottom: 5, Left: 10},
			Children: []LayoutNode{
				box(30, 20),
			},
		}
		result := row.ComputeLayout(Loose(200, 100))

		// Border-box size = content + padding
		assert.Equal(t, 50, result.Box.Width)  // 30 + 20
		assert.Equal(t, 30, result.Box.Height) // 20 + 10

		// Child position is relative to content area (after padding)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
	})

	t.Run("Border", func(t *testing.T) {
		row := &RowNode{
			Border: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Children: []LayoutNode{
				box(30, 20),
			},
		}
		result := row.ComputeLayout(Loose(200, 100))

		assert.Equal(t, 32, result.Box.Width)  // 30 + 2
		assert.Equal(t, 22, result.Box.Height) // 20 + 2
	})

	t.Run("PaddingAndBorder", func(t *testing.T) {
		row := &RowNode{
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Children: []LayoutNode{
				box(30, 20),
			},
		}
		result := row.ComputeLayout(Loose(200, 100))

		// Border-box = content + padding + border
		assert.Equal(t, 42, result.Box.Width)  // 30 + 10 + 2
		assert.Equal(t, 32, result.Box.Height) // 20 + 10 + 2
	})

	t.Run("Margin", func(t *testing.T) {
		row := &RowNode{
			Margin: EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
			Children: []LayoutNode{
				box(30, 20),
			},
		}
		result := row.ComputeLayout(Loose(200, 100))

		// Border-box doesn't include margin
		assert.Equal(t, 30, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)

		// But margin-box does
		assert.Equal(t, 50, result.Box.MarginBoxWidth())
		assert.Equal(t, 40, result.Box.MarginBoxHeight())
	})
}

func TestLinearNode_ChildMargins(t *testing.T) {
	t.Run("ChildWithMargin", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				boxWithMargin(20, 10, EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}),
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Loose(200, 100))

		// First child margin-box = 30x20
		// Total width = 30 + 20 = 50
		assert.Equal(t, 50, result.Box.Width)

		// First child positioned at its margin.Left
		assert.Equal(t, 5, result.Children[0].X)
		// Second child starts after first child's margin-box
		assert.Equal(t, 30, result.Children[1].X)
	})
}

func TestLinearNode_NestedLayouts(t *testing.T) {
	t.Run("RowInColumn", func(t *testing.T) {
		col := &ColumnNode{
			Children: []LayoutNode{
				box(50, 20), // Header
				&RowNode{
					Children: []LayoutNode{
						box(30, 40),
						box(30, 40),
					},
				},
				box(50, 20), // Footer
			},
		}
		result := col.ComputeLayout(Loose(200, 200))

		// Column width = max child width = 60 (the row with 30+30)
		// Column height = 20 + 40 + 20 = 80 (row height is max of children = 40)
		assert.Equal(t, 60, result.Box.Width)
		assert.Equal(t, 80, result.Box.Height)

		// Row is the second child
		rowLayout := result.Children[1].Layout
		assert.Equal(t, 60, rowLayout.Box.Width)
		assert.Equal(t, 40, rowLayout.Box.Height)

		// Row's children
		assert.Equal(t, 0, rowLayout.Children[0].X)
		assert.Equal(t, 30, rowLayout.Children[1].X)
	})
}

func TestLinearNode_Constraints(t *testing.T) {
	t.Run("TightConstraints_FillsSpace", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisCenter,
			Children: []LayoutNode{
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// With tight constraints, container fills the space
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 50, result.Box.Height)

		// Child is centered
		assert.Equal(t, 40, result.Children[0].X) // (100-20)/2
	})

	t.Run("LooseConstraints_ShrinkWraps", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Loose(100, 50))

		// With loose constraints, container shrink-wraps content
		assert.Equal(t, 20, result.Box.Width)
		assert.Equal(t, 10, result.Box.Height)
	})

	t.Run("MinConstraints", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),
			},
		}
		// Min 50x30, max 100x50
		result := row.ComputeLayout(Constraints{
			MinWidth:  50,
			MaxWidth:  100,
			MinHeight: 30,
			MaxHeight: 50,
		})

		// Content is 20x10, but min constraints enforce 50x30
		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 30, result.Box.Height)
	})
}
