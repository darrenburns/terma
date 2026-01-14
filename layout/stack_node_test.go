package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper to create a simple box node for stack tests.
func stackBox(w, h int) *BoxNode {
	return &BoxNode{Width: w, Height: h}
}

// Helper to create an int pointer.
func intPtr(n int) *int {
	return &n
}

func TestStackNode_BasicStacking(t *testing.T) {
	t.Run("SingleChild", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(50, 30)},
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Stack size is determined by child
		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 30, result.Box.Height)
		assert.Len(t, result.Children, 1)

		// Child at top-left by default
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
	})

	t.Run("MultipleChildren_SizeFromLargest", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
				{Node: stackBox(50, 40)}, // Largest
				{Node: stackBox(40, 35)},
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Stack size from largest child
		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 40, result.Box.Height)
		assert.Len(t, result.Children, 3)
	})

	t.Run("ZOrder_FirstAtBottom", func(t *testing.T) {
		// Children order in result matches input (first = bottom, last = top)
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(10, 10)}, // Bottom
				{Node: stackBox(20, 20)}, // Middle
				{Node: stackBox(15, 15)}, // Top
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		assert.Len(t, result.Children, 3)
		// First child in result is the bottom layer
		assert.Equal(t, 10, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[1].Layout.Box.Width)
		assert.Equal(t, 15, result.Children[2].Layout.Box.Width)
	})
}

func TestStackNode_Alignment(t *testing.T) {
	t.Run("TopStart_Default", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
	})

	t.Run("Center", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			DefaultHAlign: HAlignCenter,
			DefaultVAlign: VAlignCenter,
			ExpandWidth:   true,
			ExpandHeight:  true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// Child centered: (100-30)/2 = 35, (100-20)/2 = 40
		assert.Equal(t, 35, result.Children[0].X)
		assert.Equal(t, 40, result.Children[0].Y)
	})

	t.Run("BottomEnd", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			DefaultHAlign: HAlignEnd,
			DefaultVAlign: VAlignBottom,
			ExpandWidth:   true,
			ExpandHeight:  true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// Child at bottom-right: (100-30) = 70, (100-20) = 80
		assert.Equal(t, 70, result.Children[0].X)
		assert.Equal(t, 80, result.Children[0].Y)
	})

	t.Run("CenterStart", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			DefaultHAlign: HAlignStart,
			DefaultVAlign: VAlignCenter,
			ExpandWidth:   true,
			ExpandHeight:  true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 40, result.Children[0].Y) // (100-20)/2
	})

	t.Run("TopCenter", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			DefaultHAlign: HAlignCenter,
			DefaultVAlign: VAlignTop,
			ExpandWidth:   true,
			ExpandHeight:  true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 35, result.Children[0].X) // (100-30)/2
		assert.Equal(t, 0, result.Children[0].Y)
	})
}

func TestStackNode_Positioned(t *testing.T) {
	t.Run("TopLeft", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(20, 15),
					IsPositioned: true,
					Top:          intPtr(5),
					Left:         intPtr(10),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 10, result.Children[0].X)
		assert.Equal(t, 5, result.Children[0].Y)
	})

	t.Run("TopRight", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(20, 15),
					IsPositioned: true,
					Top:          intPtr(5),
					Right:        intPtr(10),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// X = 100 - 20 - 10 = 70
		assert.Equal(t, 70, result.Children[0].X)
		assert.Equal(t, 5, result.Children[0].Y)
	})

	t.Run("BottomLeft", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(20, 15),
					IsPositioned: true,
					Bottom:       intPtr(10),
					Left:         intPtr(5),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// Y = 100 - 15 - 10 = 75
		assert.Equal(t, 5, result.Children[0].X)
		assert.Equal(t, 75, result.Children[0].Y)
	})

	t.Run("BottomRight", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(20, 15),
					IsPositioned: true,
					Bottom:       intPtr(10),
					Right:        intPtr(5),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// X = 100 - 20 - 5 = 75, Y = 100 - 15 - 10 = 75
		assert.Equal(t, 75, result.Children[0].X)
		assert.Equal(t, 75, result.Children[0].Y)
	})

	t.Run("Fill_TopBottomLeftRight", func(t *testing.T) {
		// Positioned.fill equivalent: all edges at 0
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(10, 10), // Small initial size
					IsPositioned: true,
					Top:          intPtr(0),
					Right:        intPtr(0),
					Bottom:       intPtr(0),
					Left:         intPtr(0),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 80))

		// Child stretched to fill entire stack
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 80, result.Children[0].Layout.Box.Height)
	})

	t.Run("TopAndBottom_StretchHeight", func(t *testing.T) {
		// Top=10, Bottom=20 -> height = 100 - 10 - 20 = 70
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(30, 10), // Width stays, height stretched
					IsPositioned: true,
					Top:          intPtr(10),
					Bottom:       intPtr(20),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 0, result.Children[0].X) // Defaults to 0 when Left not set
		assert.Equal(t, 10, result.Children[0].Y)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)  // Original width
		assert.Equal(t, 70, result.Children[0].Layout.Box.Height) // Stretched: 100-10-20
	})

	t.Run("LeftAndRight_StretchWidth", func(t *testing.T) {
		// Left=15, Right=25 -> width = 100 - 15 - 25 = 60
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(10, 40), // Height stays, width stretched
					IsPositioned: true,
					Left:         intPtr(15),
					Right:        intPtr(25),
				},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		assert.Equal(t, 15, result.Children[0].X)
		assert.Equal(t, 0, result.Children[0].Y) // Defaults to 0 when Top not set
		assert.Equal(t, 60, result.Children[0].Layout.Box.Width)  // Stretched: 100-15-25
		assert.Equal(t, 40, result.Children[0].Layout.Box.Height) // Original height
	})
}

func TestStackNode_PositionedDoesNotAffectSize(t *testing.T) {
	t.Run("PositionedChildIgnoredForSizing", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(50, 40)}, // Non-positioned: determines size
				{
					Node:         stackBox(200, 200), // Large positioned child
					IsPositioned: true,
					Top:          intPtr(0),
					Left:         intPtr(0),
				},
			},
		}

		result := stack.ComputeLayout(Loose(300, 300))

		// Stack size from non-positioned child only
		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 40, result.Box.Height)
	})

	t.Run("OnlyPositionedChildren_ZeroSize", func(t *testing.T) {
		// All children are positioned -> stack has zero intrinsic size
		stack := &StackNode{
			Children: []StackChild{
				{
					Node:         stackBox(50, 40),
					IsPositioned: true,
					Top:          intPtr(5),
					Left:         intPtr(5),
				},
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Stack has no intrinsic size (no non-positioned children)
		assert.Equal(t, 0, result.Box.Width)
		assert.Equal(t, 0, result.Box.Height)
	})
}

func TestStackNode_MixedChildren(t *testing.T) {
	t.Run("CenteredContentWithCornerBadges", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(60, 40)}, // Main content (centered)
				{
					Node:         stackBox(10, 10), // Top-left badge
					IsPositioned: true,
					Top:          intPtr(0),
					Left:         intPtr(0),
				},
				{
					Node:         stackBox(10, 10), // Bottom-right badge
					IsPositioned: true,
					Bottom:       intPtr(0),
					Right:        intPtr(0),
				},
			},
			DefaultHAlign: HAlignCenter,
			DefaultVAlign: VAlignCenter,
			ExpandWidth:   true,
			ExpandHeight:  true,
		}

		result := stack.ComputeLayout(Tight(100, 80))

		assert.Len(t, result.Children, 3)

		// Main content centered
		main := result.Children[0]
		assert.Equal(t, 20, main.X) // (100-60)/2
		assert.Equal(t, 20, main.Y) // (80-40)/2

		// Top-left badge
		topLeft := result.Children[1]
		assert.Equal(t, 0, topLeft.X)
		assert.Equal(t, 0, topLeft.Y)

		// Bottom-right badge
		bottomRight := result.Children[2]
		assert.Equal(t, 90, bottomRight.X) // 100 - 10 - 0
		assert.Equal(t, 70, bottomRight.Y) // 80 - 10 - 0
	})
}

func TestStackNode_WithPadding(t *testing.T) {
	t.Run("PaddingReducesContentArea", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			Padding:      EdgeInsets{Top: 5, Right: 10, Bottom: 5, Left: 10},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Tight(100, 100))

		// Border-box size is 100x100
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 100, result.Box.Height)

		// Non-positioned child is aligned in content area but positioned
		// relative to border-box origin. With padding Left=10, Top=5,
		// the content area starts at (10, 5).
		assert.Equal(t, 10, result.Children[0].X)
		assert.Equal(t, 5, result.Children[0].Y)
	})
}

func TestStackNode_WithMargin(t *testing.T) {
	t.Run("ChildMarginAffectsPosition", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: &BoxNode{
					Width:  30,
					Height: 20,
					Margin: EdgeInsets{Top: 5, Left: 10},
				}},
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Child position accounts for margin
		assert.Equal(t, 10, result.Children[0].X)
		assert.Equal(t, 5, result.Children[0].Y)
	})
}

func TestStackNode_Constraints(t *testing.T) {
	t.Run("ExplicitWidthHeight", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			MinWidth:  60,
			MaxWidth:  60,
			MinHeight: 50,
			MaxHeight: 50,
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Explicit size constraints override child-based sizing
		assert.Equal(t, 60, result.Box.Width)
		assert.Equal(t, 50, result.Box.Height)
	})

	t.Run("ExpandFlags", func(t *testing.T) {
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(30, 20)},
			},
			ExpandWidth:  true,
			ExpandHeight: true,
		}

		result := stack.ComputeLayout(Loose(100, 80))

		// Expand to fill available space
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 80, result.Box.Height)
	})
}

func TestStackNode_Empty(t *testing.T) {
	t.Run("NoChildren", func(t *testing.T) {
		stack := &StackNode{}

		result := stack.ComputeLayout(Tight(100, 80))

		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 80, result.Box.Height)
		assert.Empty(t, result.Children)
	})
}

func TestStackNode_RealWorldScenarios(t *testing.T) {
	t.Run("LoadingOverlay", func(t *testing.T) {
		// Content with a loading spinner overlay
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(80, 60)}, // Main content
				{
					Node:         stackBox(20, 20), // Loading spinner (centered)
					IsPositioned: true,
					Top:          intPtr(0),
					Right:        intPtr(0),
					Bottom:       intPtr(0),
					Left:         intPtr(0),
				},
			},
			DefaultHAlign: HAlignCenter,
			DefaultVAlign: VAlignCenter,
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Stack sized by main content
		assert.Equal(t, 80, result.Box.Width)
		assert.Equal(t, 60, result.Box.Height)

		// Spinner fills stack (positioned fill)
		spinner := result.Children[1]
		assert.Equal(t, 80, spinner.Layout.Box.Width)
		assert.Equal(t, 60, spinner.Layout.Box.Height)
	})

	t.Run("BadgeOnIcon", func(t *testing.T) {
		// Icon with notification badge in top-right corner
		stack := &StackNode{
			Children: []StackChild{
				{Node: stackBox(24, 24)}, // Icon
				{
					Node:         stackBox(8, 8), // Badge
					IsPositioned: true,
					Top:          intPtr(-2), // Slightly outside
					Right:        intPtr(-2),
				},
			},
		}

		result := stack.ComputeLayout(Loose(100, 100))

		// Stack sized by icon
		assert.Equal(t, 24, result.Box.Width)
		assert.Equal(t, 24, result.Box.Height)

		// Badge positioned at top-right with negative offset
		badge := result.Children[1]
		assert.Equal(t, 18, badge.X) // 24 - 8 - (-2) = 18
		assert.Equal(t, -2, badge.Y)
	})
}
