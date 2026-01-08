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

	t.Run("EmptyWithPadding_ConstrainedOnce", func(t *testing.T) {
		// Bug: emptyLayout was applying constraints twice:
		// 1. Constrain(0,0) → min (e.g., 100)
		// 2. Add insets → 120
		// 3. Constrain again → clamp to max (e.g., 110)
		// Result: 110, but should be 100 (the minimum)
		//
		// An empty container's natural size is just its insets.
		// That should be clamped to constraints once.
		row := &RowNode{
			Padding: EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10}, // 20x20 insets
		}
		// Min=100, Max=110 - empty box with 20px insets should clamp to min (100)
		result := row.ComputeLayout(Constraints{
			MinWidth:  100,
			MaxWidth:  110,
			MinHeight: 100,
			MaxHeight: 110,
		})

		// Natural size (20) clamped to min (100), not min+insets clamped to max (110)
		assert.Equal(t, 100, result.Box.Width, "should clamp insets to min, not add insets to min")
		assert.Equal(t, 100, result.Box.Height, "should clamp insets to min, not add insets to min")
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
		assert.Equal(t, 25, result.Children[1].X) // 20 + 5
		assert.Equal(t, 60, result.Children[2].X) // 25 + 30 + 5
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
		assert.Equal(t, 40, result.Children[1].X) // 0 + 20 + 20
		assert.Equal(t, 80, result.Children[2].X) // 40 + 20 + 20
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
		assert.Equal(t, 40, result.Children[1].X) // 10 + 20 + 10
		assert.Equal(t, 70, result.Children[2].X) // 40 + 20 + 10
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
			result.Children[0].X, // leading gap
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

	t.Run("SpaceAround_GapsDistributedEvenly", func(t *testing.T) {
		// 3 children of 21 width each = 63, container = 100
		// Extra = 37, SpaceAround distributes as: half + full + full + half = 3 units
		// With cumulative distribution, remainder spreads evenly
		row := &RowNode{
			MainAlign: MainAxisSpaceAround,
			Children: []LayoutNode{
				box(21, 10),
				box(21, 10),
				box(21, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Calculate actual gaps
		gaps := []int{
			result.Children[0].X, // leading gap (half)
			result.Children[1].X - (result.Children[0].X + 21), // gap 0-1 (full)
			result.Children[2].X - (result.Children[1].X + 21), // gap 1-2 (full)
			100 - (result.Children[2].X + 21),                  // trailing gap (half)
		}

		// Between gaps (full units) should be roughly equal
		betweenGaps := []int{gaps[1], gaps[2]}
		minBetween, maxBetween := betweenGaps[0], betweenGaps[0]
		for _, g := range betweenGaps {
			if g < minBetween {
				minBetween = g
			}
			if g > maxBetween {
				maxBetween = g
			}
		}
		assert.LessOrEqual(t, maxBetween-minBetween, 1, "between gaps should differ by at most 1")

		// Edge gaps should be roughly half of between gaps (within rounding tolerance)
		// For 37/3 ≈ 12.33 per unit: full ≈ 12, half ≈ 6
		edgeGaps := []int{gaps[0], gaps[3]}
		for i, edge := range edgeGaps {
			// Allow edge gaps to be slightly more or less than half due to cumulative rounding
			assert.GreaterOrEqual(t, edge, 5, "edge gap %d too small", i)
			assert.LessOrEqual(t, edge, 8, "edge gap %d too large", i)
		}

		// Verify total gaps sum to extra space (most important check)
		totalGaps := 0
		for _, g := range gaps {
			totalGaps += g
		}
		assert.Equal(t, 37, totalGaps, "gaps should sum to extra space")
	})

	// --- 0-child edge cases (ensure no crashes) ---

	t.Run("SpaceBetween_NoChildren", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children:  []LayoutNode{}, // Empty
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Should not crash, returns empty layout
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)
		assert.Empty(t, result.Children)
	})

	t.Run("SpaceAround_NoChildren", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisSpaceAround,
			Children:  []LayoutNode{},
		}
		result := row.ComputeLayout(Tight(100, 20))
		assert.Equal(t, 100, result.Box.Width)
		assert.Empty(t, result.Children)
	})

	t.Run("SpaceEvenly_NoChildren", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisSpaceEvenly,
			Children:  []LayoutNode{},
		}
		result := row.ComputeLayout(Tight(100, 20))
		assert.Equal(t, 100, result.Box.Width)
		assert.Empty(t, result.Children)
	})

	// --- Single-child edge cases ---

	t.Run("SpaceBetween_SingleChild", func(t *testing.T) {
		// Single child with SpaceBetween: no "between" to space, child at 0
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children:  []LayoutNode{box(20, 10)},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Single child starts at 0 (explicit handling in code)
		assert.Equal(t, 0, result.Children[0].X)
	})

	t.Run("SpaceAround_SingleChild", func(t *testing.T) {
		// Single child with SpaceAround: equal space on both sides = centered
		row := &RowNode{
			MainAlign: MainAxisSpaceAround,
			Children:  []LayoutNode{box(20, 10)},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// extraSpace = 80, formula: (80 * 1) / 2 = 40
		assert.Equal(t, 40, result.Children[0].X)
	})

	t.Run("SpaceEvenly_SingleChild", func(t *testing.T) {
		// Single child with SpaceEvenly: equal gaps everywhere = centered
		row := &RowNode{
			MainAlign: MainAxisSpaceEvenly,
			Children:  []LayoutNode{box(20, 10)},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// extraSpace = 80, numGaps = 2, spaceBefore = (80 * 1) / 2 = 40
		assert.Equal(t, 40, result.Children[0].X)
	})

	// --- Spacing field ignored for space-* alignments ---

	t.Run("SpaceBetween_IgnoresSpacingField", func(t *testing.T) {
		// Setting Spacing has no effect when MainAlign is SpaceBetween
		// This documents intentional behavior: space-* alignments calculate their own gaps
		rowWithSpacing := &RowNode{
			Spacing:   100, // Large spacing - should be ignored
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 10),
			},
		}
		rowWithoutSpacing := &RowNode{
			Spacing:   0,
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				box(20, 10),
				box(20, 10),
			},
		}

		r1 := rowWithSpacing.ComputeLayout(Tight(100, 20))
		r2 := rowWithoutSpacing.ComputeLayout(Tight(100, 20))

		// Same positions - Spacing field is ignored for SpaceBetween
		assert.Equal(t, r1.Children[0].X, r2.Children[0].X)
		assert.Equal(t, r1.Children[1].X, r2.Children[1].X)
	})

	t.Run("SpaceEvenly_IgnoresSpacingField", func(t *testing.T) {
		rowWithSpacing := &RowNode{
			Spacing:   100,
			MainAlign: MainAxisSpaceEvenly,
			Children:  []LayoutNode{box(20, 10), box(20, 10)},
		}
		rowWithoutSpacing := &RowNode{
			Spacing:   0,
			MainAlign: MainAxisSpaceEvenly,
			Children:  []LayoutNode{box(20, 10), box(20, 10)},
		}

		r1 := rowWithSpacing.ComputeLayout(Tight(100, 20))
		r2 := rowWithoutSpacing.ComputeLayout(Tight(100, 20))

		assert.Equal(t, r1.Children[0].X, r2.Children[0].X)
		assert.Equal(t, r1.Children[1].X, r2.Children[1].X)
	})

	t.Run("SpaceAround_IgnoresSpacingField", func(t *testing.T) {
		rowWithSpacing := &RowNode{
			Spacing:   100,
			MainAlign: MainAxisSpaceAround,
			Children:  []LayoutNode{box(20, 10), box(20, 10)},
		}
		rowWithoutSpacing := &RowNode{
			Spacing:   0,
			MainAlign: MainAxisSpaceAround,
			Children:  []LayoutNode{box(20, 10), box(20, 10)},
		}

		r1 := rowWithSpacing.ComputeLayout(Tight(100, 20))
		r2 := rowWithoutSpacing.ComputeLayout(Tight(100, 20))

		assert.Equal(t, r1.Children[0].X, r2.Children[0].X)
		assert.Equal(t, r1.Children[1].X, r2.Children[1].X)
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

	t.Run("Stretch_OverridesChildConstraints", func(t *testing.T) {
		// A child with MaxHeight=15 says "I cannot be taller than 15"
		// But a sibling is 25 tall, and CrossAxisStretch forces container to 25
		// Parent constraint (stretch to 25) MUST win over child's MaxHeight
		//
		// This matches Flutter's philosophy: parent constraints are authoritative.
		// When min > max due to conflict, min wins. The child may look ugly
		// (content overflow), but the layout is deterministic.
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				&BoxNode{Width: 20, Height: 10, MaxHeight: 15}, // "I cannot be > 15"
				box(20, 25), // Sibling forces container to 25
			},
		}
		result := row.ComputeLayout(Loose(100, 100))

		// Container shrink-wraps to tallest child (25)
		assert.Equal(t, 25, result.Box.Height)

		// Child 0 is forced to 25, even though it said MaxHeight=15
		// Parent constraints override child preferences
		assert.Equal(t, 25, result.Children[0].Layout.Box.Height,
			"parent stretch constraint should override child's MaxHeight")

		// Child 1 stays at 25 (its natural size)
		assert.Equal(t, 25, result.Children[1].Layout.Box.Height)
	})

	t.Run("Stretch_WithChildMargins", func(t *testing.T) {
		// Container is 50px tall, child has 5px top/bottom margin
		// Child should stretch to fill available space (50 - 10 = 40px)
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				boxWithMargin(20, 10, EdgeInsets{Top: 5, Right: 0, Bottom: 5, Left: 0}),
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// Child border-box should be 40 (container 50 - margins 10)
		assert.Equal(t, 40, result.Children[0].Layout.Box.Height,
			"stretched child should account for its own margins")

		// Child margin-box should equal container
		assert.Equal(t, 50, result.Children[0].Layout.Box.MarginBoxHeight(),
			"child margin-box should fit exactly in container")

		// Child positioned at margin offset
		assert.Equal(t, 5, result.Children[0].Y)

		// Ensure Main Axis (Width) remains untouched (should be natural size 20)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width,
			"Main axis width should preserve natural size during cross-axis stretch")
	})

	t.Run("Stretch_ShrinksLargerChildren", func(t *testing.T) {
		// Child is 80px tall but container is only 30px
		// Stretch forces the child's layout box to match container exactly,
		// even if that means shrinking. Content may overflow visually,
		// but layout integrity is preserved.
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 80), // Wants 80px height
			},
		}
		result := row.ComputeLayout(Tight(100, 30))

		// Child is shrunk to container height
		assert.Equal(t, 30, result.Children[0].Layout.Box.Height,
			"stretch should shrink larger children to fit container")

		// Container is the tight constraint size
		assert.Equal(t, 30, result.Box.Height)
	})

	t.Run("Stretch_ShrinksLargerChildrenWithMargins", func(t *testing.T) {
		// Child wants 80px but container is 40px, child has 5px margins
		// Available for border-box = 40 - 10 = 30
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				boxWithMargin(20, 80, EdgeInsets{Top: 5, Right: 0, Bottom: 5, Left: 0}),
			},
		}
		result := row.ComputeLayout(Tight(100, 40))

		// Child border-box shrunk to 30 (40 - margins)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Height,
			"stretch should shrink to available space after margins")

		// Margin-box equals container
		assert.Equal(t, 40, result.Children[0].Layout.Box.MarginBoxHeight())
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

func TestLinearNode_NodeConstraints(t *testing.T) {
	t.Run("Row_MinHeight_WithStretch", func(t *testing.T) {
		// Key test case: Row has MinHeight=50 on the node itself.
		// Parent gives loose constraints. Children are 20px tall.
		// CrossAlign: Stretch.
		//
		// Expected: Container should be 50px (due to MinHeight).
		// Children should stretch to 50px.
		row := &RowNode{
			MinHeight:  50,
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 20),
			},
		}
		result := row.ComputeLayout(Loose(100, 100))

		// Container respects its own MinHeight
		assert.Equal(t, 50, result.Box.Height,
			"container should respect its own MinHeight")

		// Child stretches to fill container's MinHeight
		assert.Equal(t, 50, result.Children[0].Layout.Box.Height,
			"child should stretch to container's MinHeight")
	})

	t.Run("Column_MinWidth_WithStretch", func(t *testing.T) {
		// Same test but for Column (cross-axis is width)
		col := &ColumnNode{
			MinWidth:   50,
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(20, 20),
			},
		}
		result := col.ComputeLayout(Loose(100, 100))

		// Container respects its own MinWidth
		assert.Equal(t, 50, result.Box.Width,
			"container should respect its own MinWidth")

		// Child stretches to fill container's MinWidth
		assert.Equal(t, 50, result.Children[0].Layout.Box.Width,
			"child should stretch to container's MinWidth")
	})

	t.Run("Row_MinWidth_MainAxis", func(t *testing.T) {
		// MinWidth on Row affects the main axis (width)
		row := &RowNode{
			MinWidth:  80,
			MainAlign: MainAxisCenter,
			Children: []LayoutNode{
				box(20, 10),
			},
		}
		result := row.ComputeLayout(Loose(100, 50))

		// Container is at least MinWidth
		assert.Equal(t, 80, result.Box.Width)

		// Child is centered in the larger container
		assert.Equal(t, 30, result.Children[0].X) // (80-20)/2
	})

	t.Run("Row_MaxHeight_Limits", func(t *testing.T) {
		// MaxHeight limits container even if children are taller.
		// Container constraints flow down to children - they get clamped too.
		row := &RowNode{
			MaxHeight: 30,
			Children: []LayoutNode{
				box(20, 50), // Wants 50, but will be constrained to 30
			},
		}
		result := row.ComputeLayout(Loose(100, 100))

		// Container is clamped to MaxHeight
		assert.Equal(t, 30, result.Box.Height,
			"container should respect its own MaxHeight")

		// Child is also clamped - container constraints flow down to children
		assert.Equal(t, 30, result.Children[0].Layout.Box.Height,
			"child should be clamped to container's MaxHeight")
	})

	t.Run("NodeConstraints_CombineWithParent", func(t *testing.T) {
		// Node says MinHeight=50, parent says MaxHeight=40
		// Parent constraints are inviolable - they define available space.
		// Node gets clamped to parent's max.
		row := &RowNode{
			MinHeight: 50,
			Children: []LayoutNode{
				box(20, 20),
			},
		}
		result := row.ComputeLayout(Constraints{
			MinWidth:  0,
			MaxWidth:  100,
			MinHeight: 0,
			MaxHeight: 40, // Less than node's MinHeight
		})

		// Node's MinHeight (50) > parent's MaxHeight (40) creates conflict.
		// Parent wins: node is clamped to 40 (the available space).
		// This prevents returning sizes that exceed available space.
		assert.Equal(t, 40, result.Box.Height,
			"parent max wins when it conflicts with node min")
	})

	t.Run("Empty_RespectsMinSize", func(t *testing.T) {
		// Empty row with MinHeight should still be that height
		row := &RowNode{
			MinHeight: 30,
			MinWidth:  50,
		}
		result := row.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 30, result.Box.Height)
	})

	t.Run("NavBar_UseCase", func(t *testing.T) {
		// Real-world use case: NavBar component that's always at least 3 cells tall
		navBar := &RowNode{
			MinHeight:  3,
			CrossAlign: CrossAxisStretch,
			Padding:    EdgeInsets{Left: 1, Right: 1},
			Children: []LayoutNode{
				box(10, 1), // Small button
				box(10, 1), // Small button
			},
		}
		result := navBar.ComputeLayout(Loose(80, 24))

		// NavBar respects MinHeight even though content is only 1 tall
		assert.Equal(t, 3, result.Box.Height)

		// Children stretch to fill (3 - 0 padding vertical = 3)
		assert.Equal(t, 3, result.Children[0].Layout.Box.Height)
		assert.Equal(t, 3, result.Children[1].Layout.Box.Height)
	})

	t.Run("Sidebar_UseCase", func(t *testing.T) {
		// Real-world use case: Sidebar that's always at least 20 cells wide
		sidebar := &ColumnNode{
			MinWidth:   20,
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				box(10, 5), // Menu item
				box(15, 5), // Menu item
			},
		}
		result := sidebar.ComputeLayout(Loose(80, 24))

		// Sidebar respects MinWidth even though content is only 15 wide
		assert.Equal(t, 20, result.Box.Width)

		// Children stretch to fill MinWidth
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[1].Layout.Box.Width)
	})

	t.Run("UserConfigError_MinExceedsMax", func(t *testing.T) {
		// User misconfigures node with Min > Max (e.g., MinHeight: 60, MaxHeight: 40)
		// This is a configuration error, but we handle it gracefully.
		// Min wins as it represents a hard functional requirement.
		row := &RowNode{
			MinHeight: 60,
			MaxHeight: 40, // Error: less than MinHeight
			Children: []LayoutNode{
				box(20, 20),
			},
		}
		result := row.ComputeLayout(Loose(100, 100))

		// Min wins: container is 60, not 40
		assert.Equal(t, 60, result.Box.Height,
			"min should win when user misconfigures min > max")
	})

}

func TestLinearNode_Overflow(t *testing.T) {
	t.Run("Center_ChildrenExceedContainer", func(t *testing.T) {
		// Children total 80px in 50px container
		row := &RowNode{
			MainAlign: MainAxisCenter,
			Children: []LayoutNode{
				box(40, 10),
				box(40, 10),
			},
		}
		result := row.ComputeLayout(Tight(50, 20))

		// Content is centered: overflows equally on both sides
		// extraSpace = 50 - 80 = -30, centered = -15
		assert.Equal(t, -15, result.Children[0].X)
		assert.Equal(t, 25, result.Children[1].X) // -15 + 40
	})

	t.Run("End_ChildrenExceedContainer", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisEnd,
			Children: []LayoutNode{
				box(40, 10),
				box(40, 10),
			},
		}
		result := row.ComputeLayout(Tight(50, 20))

		// Last child flush with end, overflow on left
		assert.Equal(t, -30, result.Children[0].X)
		assert.Equal(t, 10, result.Children[1].X)
		// Last child ends at container edge
		assert.Equal(t, 50, result.Children[1].X+40)
	})

	t.Run("SpaceBetween_ChildrenExceedContainer", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				box(40, 10),
				box(40, 10),
			},
		}
		result := row.ComputeLayout(Tight(50, 20))

		// SpaceBetween invariant maintained: first at 0, last ends at container edge
		// Children overlap due to negative gap
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 10, result.Children[1].X) // gap = -30, so 40 - 30 = 10
		assert.Equal(t, 50, result.Children[1].X+40)
	})

	t.Run("Start_ChildrenExceedContainer", func(t *testing.T) {
		row := &RowNode{
			MainAlign: MainAxisStart,
			Children: []LayoutNode{
				box(40, 10),
				box(40, 10),
			},
		}
		result := row.ComputeLayout(Tight(50, 20))

		// Children positioned normally, overflow to the right
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 40, result.Children[1].X) // second child starts at 40, ends at 80
	})
}

func TestLinearNode_InvalidInputConstraints(t *testing.T) {
	t.Run("MinExceedsMax_Normalized", func(t *testing.T) {
		// Parent passes invalid constraints (MinWidth > MaxWidth)
		row := &RowNode{
			Children: []LayoutNode{box(50, 20)},
		}
		result := row.ComputeLayout(Constraints{
			MinWidth:  100,
			MaxWidth:  90, // Invalid: less than MinWidth
			MinHeight: 0,
			MaxHeight: 50,
		})

		// Min wins: container is 100 (the minimum)
		// System normalizes constraints so Max becomes 100
		assert.Equal(t, 100, result.Box.Width)
	})

	t.Run("MinHeightExceedsMaxHeight_Normalized", func(t *testing.T) {
		col := &ColumnNode{
			Children: []LayoutNode{box(20, 30)},
		}
		result := col.ComputeLayout(Constraints{
			MinWidth:  0,
			MaxWidth:  100,
			MinHeight: 80,
			MaxHeight: 60, // Invalid: less than MinHeight
		})

		// Min wins: container is 80
		assert.Equal(t, 80, result.Box.Height)
	})
}
