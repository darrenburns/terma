package layout

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to create a fixed-size box for testing
func fixedBox(width, height int) *BoxNode {
	return &BoxNode{Width: width, Height: height}
}

func TestScrollableNode_NoScrollNeeded(t *testing.T) {
	t.Run("ContentFitsInViewport", func(t *testing.T) {
		// Child is 50x30, viewport is 100x100 - no scrolling needed
		scrollable := &ScrollableNode{
			Child: fixedBox(50, 30),
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.Equal(t, 30, result.Box.VirtualHeight) // Matches child
		assert.False(t, result.Box.IsScrollableY())
		assert.Equal(t, 0, result.Box.ScrollbarWidth) // No scrollbar when not needed
	})

	t.Run("ContentExactlyFits", func(t *testing.T) {
		// Child exactly matches viewport
		scrollable := &ScrollableNode{
			Child: fixedBox(100, 100),
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.False(t, result.Box.IsScrollableY())
		assert.Equal(t, 0, result.Box.ScrollbarWidth)
	})
}

func TestScrollableNode_VerticalScroll(t *testing.T) {
	t.Run("ContentTallerThanViewport", func(t *testing.T) {
		// Child is 80x200, viewport is 100x100 - needs vertical scroll
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.Equal(t, 200, result.Box.VirtualHeight)
		assert.True(t, result.Box.IsScrollableY())
		assert.Equal(t, 1, result.Box.ScrollbarWidth)
	})

	t.Run("ScrollOffsetApplied", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollOffsetY:  50,
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 50, result.Box.ScrollOffsetY)
	})

	t.Run("ScrollOffsetClampedToMax", func(t *testing.T) {
		// Virtual height 200, viewport 100, max scroll = 100 + scrollbar
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollOffsetY:  500, // Way too high
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Should be clamped to max valid scroll
		assert.LessOrEqual(t, result.Box.ScrollOffsetY, 200-100+1)
	})

	t.Run("ScrollOffsetClampedToZero", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollOffsetY:  -50, // Negative
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 0, result.Box.ScrollOffsetY)
	})
}

func TestScrollableNode_ScrollbarReservation(t *testing.T) {
	t.Run("ScrollbarSpaceReserved", func(t *testing.T) {
		// With scrollbar width 2, usable content width should be reduced
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollbarWidth: 2,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 2, result.Box.ScrollbarWidth)
		// Usable content width should account for scrollbar
		usableBox := result.Box.UsableContentBox()
		assert.Equal(t, result.Box.ContentWidth()-2, usableBox.Width)
	})

	t.Run("AlwaysReserveScrollbarSpace_WhenNeeded", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:                       fixedBox(80, 200),
			ScrollbarWidth:              1,
			AlwaysReserveScrollbarSpace: true,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 1, result.Box.ScrollbarWidth)
	})

	t.Run("AlwaysReserveScrollbarSpace_WhenNotNeeded", func(t *testing.T) {
		// Content fits, but we still want scrollbar space reserved
		scrollable := &ScrollableNode{
			Child:                       fixedBox(50, 50),
			ScrollbarWidth:              1,
			AlwaysReserveScrollbarSpace: true,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Scrollbar space should be reserved even though not scrollable
		assert.Equal(t, 1, result.Box.ScrollbarWidth)
	})

	t.Run("ZeroScrollbarWidth_NoSpaceReserved", func(t *testing.T) {
		// ScrollbarWidth: 0 means no space reserved (e.g., for overlay scrollbars)
		scrollable := &ScrollableNode{
			Child:          fixedBox(80, 200),
			ScrollbarWidth: 0,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Zero means no scrollbar space - user explicitly opted out
		assert.Equal(t, 0, result.Box.ScrollbarWidth)
		// Content still scrolls, just no space reserved for scrollbar
		assert.True(t, result.Box.IsScrollableY())
	})
}

func TestScrollableNode_WithInsets(t *testing.T) {
	t.Run("PaddingReducesContentArea", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:   fixedBox(50, 150),
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		// Content area should be reduced by padding
		assert.Equal(t, 90, result.Box.ContentWidth())
		assert.Equal(t, 90, result.Box.ContentHeight())
	})

	t.Run("BorderReducesContentArea", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:  fixedBox(50, 150),
			Border: EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 98, result.Box.ContentWidth())
		assert.Equal(t, 98, result.Box.ContentHeight())
	})

	t.Run("CombinedInsets", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:   fixedBox(50, 150),
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Margin:  EdgeInsets{Top: 2, Right: 2, Bottom: 2, Left: 2},
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Border-box is 100x100, margin is stored but doesn't affect Width/Height
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		// Content width = 100 - padding(10) - border(2) = 88
		assert.Equal(t, 88, result.Box.ContentWidth())
	})

	t.Run("AsymmetricInsets_WidthHeightCorrect", func(t *testing.T) {
		// This test catches bugs where horizontal insets are used for height calculations
		// or vice versa. With asymmetric insets, such bugs produce wrong dimensions.
		scrollable := &ScrollableNode{
			Child:   fixedBox(50, 50),
			Padding: EdgeInsets{Top: 20, Right: 5, Bottom: 20, Left: 5}, // vInset=40, hInset=10
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		// Content width = 100 - hInset(10) = 90
		assert.Equal(t, 90, result.Box.ContentWidth())
		// Content height = 100 - vInset(40) = 60
		assert.Equal(t, 60, result.Box.ContentHeight())
	})
}

func TestScrollableNode_Constraints(t *testing.T) {
	t.Run("MinWidthRespected", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:    fixedBox(30, 50),
			MinWidth: 80,
		}
		result := scrollable.ComputeLayout(Loose(100, 100))

		assert.GreaterOrEqual(t, result.Box.Width, 80)
	})

	t.Run("MinHeightRespected", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:     fixedBox(30, 50),
			MinHeight: 80,
		}
		result := scrollable.ComputeLayout(Loose(100, 100))

		assert.GreaterOrEqual(t, result.Box.Height, 80)
	})

	t.Run("MaxWidthRespected", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:    fixedBox(150, 50),
			MaxWidth: 80,
		}
		result := scrollable.ComputeLayout(Loose(200, 200))

		assert.LessOrEqual(t, result.Box.Width, 80)
	})

	t.Run("MaxHeightRespected", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:     fixedBox(50, 150),
			MaxHeight: 80,
		}
		result := scrollable.ComputeLayout(Loose(200, 200))

		assert.LessOrEqual(t, result.Box.Height, 80)
	})
}

func TestScrollableNode_ChildPositioning(t *testing.T) {
	t.Run("ChildAtOrigin", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child: fixedBox(80, 200),
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Len(t, result.Children, 1)
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
	})

	t.Run("ChildLayoutPreserved", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child: fixedBox(80, 200),
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Child's computed layout should reflect its natural size
		childBox := result.Children[0].Layout.Box
		assert.Equal(t, 80, childBox.Width)
		assert.Equal(t, 200, childBox.Height)
	})
}

func TestScrollableNode_ShrinkWrap(t *testing.T) {
	t.Run("ShrinkWrapsToContentWithLooseConstraints", func(t *testing.T) {
		// ScrollableNode shrink-wraps to content size, acting as a boundary.
		// Unlike layout containers that expand to fill space, scrollable
		// contains content tightly.
		scrollable := &ScrollableNode{
			Child: fixedBox(50, 50),
		}
		result := scrollable.ComputeLayout(Loose(100, 100))

		// Should shrink-wrap to content, not expand to fill
		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 50, result.Box.Height)
		assert.False(t, result.Box.IsScrollableX())
		assert.False(t, result.Box.IsScrollableY())
	})

	t.Run("ShrinkWrapsWithUnboundedConstraints", func(t *testing.T) {
		// With fully unbounded constraints, shrink-wraps to content size.
		// Scrollables act as a boundary on infinity.
		scrollable := &ScrollableNode{
			Child: fixedBox(50, 50),
		}
		result := scrollable.ComputeLayout(Unbounded())

		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 50, result.Box.Height)
	})

	t.Run("ShrinkWrapsWidthWithUnboundedWidth", func(t *testing.T) {
		// Unbounded width, bounded height - shrinks width, caps height
		scrollable := &ScrollableNode{
			Child:          fixedBox(50, 200),
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Constraints{
			MinWidth: 0, MaxWidth: math.MaxInt32,
			MinHeight: 0, MaxHeight: 100,
		})

		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.True(t, result.Box.IsScrollableY()) // 200 > 100
	})

	t.Run("CapsAtMaxWhenContentOverflows", func(t *testing.T) {
		// Content larger than available space - caps at max, enables scrolling
		scrollable := &ScrollableNode{
			Child:          fixedBox(200, 300),
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.True(t, result.Box.IsScrollableX())
		assert.True(t, result.Box.IsScrollableY())
	})

	t.Run("RespectsMinConstraints", func(t *testing.T) {
		// Content smaller than min constraints - expands to min
		scrollable := &ScrollableNode{
			Child: fixedBox(20, 20),
		}
		result := scrollable.ComputeLayout(Constraints{
			MinWidth: 50, MaxWidth: 100,
			MinHeight: 50, MaxHeight: 100,
		})

		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 50, result.Box.Height)
	})
}

// Helper to create a box with minimum width (can't shrink below this)
func minWidthBox(minW, height int) *BoxNode {
	return &BoxNode{MinWidth: minW, Height: height}
}

func TestScrollableNode_ScrollbarInteraction(t *testing.T) {
	t.Run("VerticalScrollbarCausesHorizontalOverflow", func(t *testing.T) {
		// Scenario: Content has minimum width equal to viewport, but is tall (needs vertical scroll).
		// When vertical scrollbar space is reserved, viewport width reduces.
		// Content can't shrink below MinWidth, so it now overflows horizontally.
		//
		// Initial: MinWidth=100, Height=200, viewport=100x100
		// After vertical scrollbar (width 1): usable viewport is 99w
		// Content MinWidth=100 > 99, so needs horizontal scroll too
		scrollable := &ScrollableNode{
			Child:           minWidthBox(100, 200), // MinWidth prevents shrinking
			ScrollbarWidth:  1,
			ScrollbarHeight: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Should need vertical scroll (200 > 100)
		assert.True(t, result.Box.IsScrollableY())
		assert.Equal(t, 1, result.Box.ScrollbarWidth)

		// After reserving vertical scrollbar space, content (100w) > usable viewport (99w)
		// So horizontal scroll should now be needed too
		assert.True(t, result.Box.IsScrollableX(), "horizontal scroll should be triggered after vertical scrollbar reduces width")
		assert.Equal(t, 1, result.Box.ScrollbarHeight)
	})

	t.Run("VerticalScrollbarDoesNotCauseHorizontalOverflow", func(t *testing.T) {
		// Content is narrower than viewport minus scrollbar, so no horizontal scroll
		scrollable := &ScrollableNode{
			Child:           fixedBox(90, 200), // Narrow enough to fit with scrollbar
			ScrollbarWidth:  1,
			ScrollbarHeight: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.True(t, result.Box.IsScrollableY())
		assert.Equal(t, 1, result.Box.ScrollbarWidth)

		// Content (90w) < viewport after scrollbar (99w), no horizontal scroll
		assert.False(t, result.Box.IsScrollableX())
		assert.Equal(t, 0, result.Box.ScrollbarHeight)
	})
}

func TestScrollableNode_EdgeCases(t *testing.T) {
	t.Run("ZeroSizeChild", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child: fixedBox(0, 0),
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.Equal(t, 0, result.Box.VirtualHeight)
		assert.False(t, result.Box.IsScrollableY())
	})

	t.Run("VeryLargeChild", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:          fixedBox(1000, 10000),
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)
		assert.Equal(t, 10000, result.Box.VirtualHeight)
		assert.True(t, result.Box.IsScrollableY())
	})

	t.Run("ScrollOffsetWithZeroVirtualContent", func(t *testing.T) {
		scrollable := &ScrollableNode{
			Child:         fixedBox(50, 50),
			ScrollOffsetY: 100, // Offset set but no scrolling needed
		}
		result := scrollable.ComputeLayout(Tight(100, 100))

		// Offset should be clamped to 0 since no scrolling is possible
		assert.Equal(t, 0, result.Box.ScrollOffsetY)
	})
}

func TestScrollableNode_RealWorldScenarios(t *testing.T) {
	t.Run("TypicalScrollableList", func(t *testing.T) {
		// Simulate a list with 50 items, each 1 cell tall
		listContent := fixedBox(78, 50) // 78 wide to leave room for scrollbar

		scrollable := &ScrollableNode{
			Child:          listContent,
			ScrollbarWidth: 1,
			ScrollOffsetY:  10,
		}
		result := scrollable.ComputeLayout(Tight(80, 24)) // Terminal viewport

		assert.Equal(t, 80, result.Box.Width)
		assert.Equal(t, 24, result.Box.Height)
		assert.Equal(t, 50, result.Box.VirtualHeight)
		assert.True(t, result.Box.IsScrollableY())
		assert.Equal(t, 10, result.Box.ScrollOffsetY)
	})

	t.Run("ScrollableWithBorder", func(t *testing.T) {
		// Scrollable container with border (common pattern)
		content := fixedBox(76, 100)

		scrollable := &ScrollableNode{
			Child:          content,
			Border:         EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			ScrollbarWidth: 1,
		}
		result := scrollable.ComputeLayout(Tight(80, 24))

		// Border box is 80x24
		assert.Equal(t, 80, result.Box.Width)
		assert.Equal(t, 24, result.Box.Height)
		// Content area is 78x22 (minus border)
		assert.Equal(t, 78, result.Box.ContentWidth())
		assert.Equal(t, 22, result.Box.ContentHeight())
		assert.True(t, result.Box.IsScrollableY())
	})
}
