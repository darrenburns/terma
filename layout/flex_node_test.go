package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// FlexNode in Isolation Tests
// =============================================================================

func TestFlexNode_OutsideLinearNode(t *testing.T) {
	// When FlexNode is used outside a LinearNode context (e.g., as root),
	// it should delegate to its child and ignore the Flex value.

	t.Run("DelegatesToChild", func(t *testing.T) {
		flex := &FlexNode{
			Flex:  2,
			Child: box(30, 20),
		}
		result := flex.ComputeLayout(Loose(100, 100))

		// Should produce the same result as the child directly
		assert.Equal(t, 30, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)
	})

	t.Run("FlexValue_DefaultsTo1_WhenZero", func(t *testing.T) {
		flex := &FlexNode{
			Flex:  0,
			Child: box(30, 20),
		}

		assert.Equal(t, 1.0, flex.FlexValue(), "FlexValue() should return 1 for Flex: 0")
		assert.Equal(t, 0.0, flex.Flex, "original Flex field should not be mutated")
	})

	t.Run("FlexValue_DefaultsTo1_WhenNegative", func(t *testing.T) {
		flex := &FlexNode{
			Flex:  -5,
			Child: box(30, 20),
		}

		assert.Equal(t, 1.0, flex.FlexValue(), "FlexValue() should return 1 for negative Flex")
		assert.Equal(t, -5.0, flex.Flex, "original Flex field should not be mutated")
	})

	t.Run("FlexValue_ReturnsActualValue_WhenPositive", func(t *testing.T) {
		flex := &FlexNode{
			Flex:  2.5,
			Child: box(30, 20),
		}

		assert.Equal(t, 2.5, flex.FlexValue(), "FlexValue() should return actual value")
	})

	t.Run("ComputeLayout_DoesNotMutate", func(t *testing.T) {
		// Ensure ComputeLayout doesn't mutate the struct
		flex := &FlexNode{
			Flex:  0,
			Child: box(30, 20),
		}
		_ = flex.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 0.0, flex.Flex, "ComputeLayout should not mutate Flex field")
	})
}

func TestFlexNode_IsFlexNode(t *testing.T) {
	t.Run("FlexNode_ReturnsTrue", func(t *testing.T) {
		node := &FlexNode{Flex: 1, Child: box(10, 10)}
		flex, ok := IsFlexNode(node)

		assert.True(t, ok)
		assert.Equal(t, node, flex)
	})

	t.Run("BoxNode_ReturnsFalse", func(t *testing.T) {
		node := box(10, 10)
		flex, ok := IsFlexNode(node)

		assert.False(t, ok)
		assert.Nil(t, flex)
	})
}

// =============================================================================
// FlexNode within LinearNode Tests
// =============================================================================

func TestLinearNode_FlexBasics(t *testing.T) {
	t.Run("SingleFlexChild_GetsAllRemainingSpace", func(t *testing.T) {
		// Container: 100px, Fixed child: 30px, Remaining: 70px
		// The single FlexNode should get ALL 70px.
		row := &RowNode{
			Children: []LayoutNode{
				box(30, 10), // Fixed 30px
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 100, result.Box.Width)
		assert.Len(t, result.Children, 2)

		// Fixed child at position 0
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)

		// Flex child gets remaining 70px
		assert.Equal(t, 30, result.Children[1].X)
		assert.Equal(t, 70, result.Children[1].Layout.Box.Width,
			"single flex child should get all remaining space")
	})

	t.Run("TwoEqualFlexChildren_SplitEvenly", func(t *testing.T) {
		// Container: 100px, both Flex: 1
		// Each should get 50px (100 / 2)
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 50, result.Children[1].X)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Width)
	})

	t.Run("UnequalFlexRatio_1to2", func(t *testing.T) {
		// Container: 90px, Flex 1:2 ratio
		// Child 0: 90 * 1/3 = 30px
		// Child 1: 90 * 2/3 = 60px
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 2, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(90, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width,
			"Flex 1 should get 1/3 of space")

		assert.Equal(t, 30, result.Children[1].X)
		assert.Equal(t, 60, result.Children[1].Layout.Box.Width,
			"Flex 2 should get 2/3 of space")
	})

	t.Run("MixedFixedAndFlex", func(t *testing.T) {
		// Container: 100px
		// Fixed: 20px at start
		// Remaining: 80px split between two Flex: 1 children (40px each)
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10), // Fixed
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Fixed child
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)

		// First flex child: 80 * 1/2 = 40
		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 40, result.Children[1].Layout.Box.Width)

		// Second flex child: 80 * 1/2 = 40
		assert.Equal(t, 60, result.Children[2].X)
		assert.Equal(t, 40, result.Children[2].Layout.Box.Width)
	})

	t.Run("FlexBetweenFixed", func(t *testing.T) {
		// Container: 100px
		// Fixed: 20px at start, 20px at end
		// Flex: 60px in the middle
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),                            // Fixed start
				&FlexNode{Flex: 1, Child: box(0, 10)}, // Flex middle
				box(20, 10),                            // Fixed end
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 60, result.Children[1].Layout.Box.Width,
			"flex child should fill space between fixed children")

		assert.Equal(t, 80, result.Children[2].X)
		assert.Equal(t, 20, result.Children[2].Layout.Box.Width)
	})
}

func TestLinearNode_FlexWithSpacing(t *testing.T) {
	t.Run("TwoFlexChildren_WithSpacing", func(t *testing.T) {
		// Container: 100px, Spacing: 10px
		// Content space = 100 - 10 = 90px
		// Each flex child: 45px
		row := &RowNode{
			Spacing: 10,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 45, result.Children[0].Layout.Box.Width)

		// Position = 45 + 10 spacing = 55
		assert.Equal(t, 55, result.Children[1].X)
		assert.Equal(t, 45, result.Children[1].Layout.Box.Width)
	})

	t.Run("ThreeFlexChildren_WithSpacing", func(t *testing.T) {
		// Container: 100px, Spacing: 5px (2 gaps = 10px total)
		// Content space = 100 - 10 = 90px
		// Each flex child (equal): 30px
		row := &RowNode{
			Spacing: 5,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 35, result.Children[1].X) // 30 + 5
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)

		assert.Equal(t, 70, result.Children[2].X) // 35 + 30 + 5
		assert.Equal(t, 30, result.Children[2].Layout.Box.Width)
	})

	t.Run("MixedFixedAndFlex_WithSpacing", func(t *testing.T) {
		// Container: 100px, Spacing: 10px (2 gaps = 20px)
		// Fixed: 20px
		// Content space for flex = 100 - 20 - 20 = 60px, split 2 ways = 30px each
		row := &RowNode{
			Spacing: 10,
			Children: []LayoutNode{
				box(20, 10), // Fixed
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 30, result.Children[1].X) // 20 + 10 spacing
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)

		assert.Equal(t, 70, result.Children[2].X) // 30 + 30 + 10 spacing
		assert.Equal(t, 30, result.Children[2].Layout.Box.Width)
	})
}

func TestLinearNode_FlexWithConstraints(t *testing.T) {
	t.Run("FlexChild_RespectsMaxWidth", func(t *testing.T) {
		// Container: 100px
		// Flex child has MaxWidth: 30
		// Would get 100px but capped at 30px
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 30, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 30, result.Children[0].Layout.Box.Width,
			"flex child should respect its MaxWidth")
	})

	t.Run("FlexChild_MinWidth_ClampedToAllocated", func(t *testing.T) {
		// Design decision: Parent constraints are authoritative.
		// If a flex child's MinWidth exceeds its allocated share, the child
		// is clamped to the allocated share. This prevents layout overflow
		// and keeps TUI rendering predictable.
		//
		// Container: 100px, two Flex: 1 children
		// Each allocated: 50px
		// Child has MinWidth: 60, but gets clamped to 50px
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MinWidth: 60, Height: 10}},
				&FlexNode{Flex: 1, Child: &BoxNode{MinWidth: 60, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Both children clamped to their allocated share (50px)
		// MinWidth is a preference, but parent constraints are authoritative
		assert.Equal(t, 50, result.Children[0].Layout.Box.Width,
			"flex child clamped to allocated share, not MinWidth")
		assert.Equal(t, 50, result.Children[1].Layout.Box.Width,
			"flex child clamped to allocated share, not MinWidth")
	})

	t.Run("FlexChild_MinWidth_WorksWithinAllocation", func(t *testing.T) {
		// When MinWidth is less than allocated share, it works normally.
		// Container: 100px
		// Flex child with MinWidth: 30 is allocated 100px → gets at least 30px
		// (will actually get 100px since no other children)
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MinWidth: 30, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Child gets full allocation since MinWidth (30) < allocated (100)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width)
	})

	t.Run("TwoFlexChildren_BothHitMaxWidth_LeftoverExists", func(t *testing.T) {
		// Container: 100px
		// Two Flex: 1 children, each with MaxWidth: 30
		// Each would get 50px but capped at 30px
		// Total used = 60px, leftover = 40px
		// With MainAxisStart, leftover is at the end (default behavior)
		row := &RowNode{
			MainAlign: MainAxisStart,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 30, Height: 10}},
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 30, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Children at start, leftover at end
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 30, result.Children[1].X)
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)
	})
}

func TestLinearNode_FlexWithAlignment(t *testing.T) {
	// FlexNode + MainAxisAlignment compose: flex children absorb space first,
	// then alignment distributes leftover. This matches CSS flexbox.

	t.Run("FlexAbsorbsAllSpace_NoLeftoverForAlignment", func(t *testing.T) {
		// Container: 100px
		// Single flex child (no MaxWidth), absorbs all 100px
		// MainAxisCenter has no effect since no leftover
		row := &RowNode{
			MainAlign: MainAxisCenter,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Child fills entire container, starts at 0
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width)
	})

	t.Run("FlexHitsMaxWidth_CenterDistributesLeftover", func(t *testing.T) {
		// Container: 100px
		// Single flex child with MaxWidth: 40
		// Leftover = 60px, centered = 30px offset
		row := &RowNode{
			MainAlign: MainAxisCenter,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 40, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Child is centered: (100 - 40) / 2 = 30
		assert.Equal(t, 30, result.Children[0].X)
		assert.Equal(t, 40, result.Children[0].Layout.Box.Width)
	})

	t.Run("FlexHitsMaxWidth_EndDistributesLeftover", func(t *testing.T) {
		// Container: 100px
		// Single flex child with MaxWidth: 40
		// Leftover = 60px, all at start (child at end)
		row := &RowNode{
			MainAlign: MainAxisEnd,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 40, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Child at end: 100 - 40 = 60
		assert.Equal(t, 60, result.Children[0].X)
		assert.Equal(t, 40, result.Children[0].Layout.Box.Width)
	})

	t.Run("TwoFlex_BothHitMax_SpaceBetween", func(t *testing.T) {
		// Container: 100px
		// Two Flex: 1, each MaxWidth: 30
		// Total used = 60px, leftover = 40px
		// SpaceBetween: first at 0, last at 100-30 = 70
		row := &RowNode{
			MainAlign: MainAxisSpaceBetween,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 30, Height: 10}},
				&FlexNode{Flex: 1, Child: &BoxNode{MaxWidth: 30, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 30, result.Children[0].Layout.Box.Width)

		assert.Equal(t, 70, result.Children[1].X) // 100 - 30 = 70
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)
	})
}

func TestLinearNode_FlexTransparency(t *testing.T) {
	// FlexNode is transparent: it doesn't appear in output ComputedLayout.Children.
	// Output indices must match input indices for widget-to-layout mapping.

	t.Run("OutputIndicesMatchInput", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),                           // Index 0
				&FlexNode{Flex: 1, Child: box(0, 10)}, // Index 1
				box(20, 10),                           // Index 2
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Output should have 3 children (same as input)
		assert.Len(t, result.Children, 3,
			"output children count must match input")

		// Each child's layout should be the actual content, not a FlexNode wrapper
		// Index 1 should be the unwrapped BoxNode, not the FlexNode
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width) // Fixed
		assert.Equal(t, 60, result.Children[1].Layout.Box.Width) // Flex fills remaining
		assert.Equal(t, 20, result.Children[2].Layout.Box.Width) // Fixed
	})

	t.Run("AllFlexChildren_IndicesMatch", func(t *testing.T) {
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 2, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Container: 100px, Flex 1:2:1 = 25:50:25
		assert.Len(t, result.Children, 3)
		assert.Equal(t, 25, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Width)
		assert.Equal(t, 25, result.Children[2].Layout.Box.Width)
	})
}

func TestLinearNode_FlexColumn(t *testing.T) {
	// Verify flex works correctly for Column (vertical axis)

	t.Run("ColumnFlexBasic", func(t *testing.T) {
		// Container: 100px tall
		// Fixed: 20px, Flex: 1 gets remaining 80px
		col := &ColumnNode{
			Children: []LayoutNode{
				box(50, 20), // Fixed height
				&FlexNode{Flex: 1, Child: box(50, 0)},
			},
		}
		result := col.ComputeLayout(Tight(50, 100))

		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Height)

		assert.Equal(t, 20, result.Children[1].Y)
		assert.Equal(t, 80, result.Children[1].Layout.Box.Height,
			"flex child should fill remaining vertical space")
	})

	t.Run("ColumnTwoFlexEqual", func(t *testing.T) {
		// Container: 100px tall, two equal flex children
		col := &ColumnNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(50, 0)},
				&FlexNode{Flex: 1, Child: box(50, 0)},
			},
		}
		result := col.ComputeLayout(Tight(50, 100))

		assert.Equal(t, 0, result.Children[0].Y)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Height)

		assert.Equal(t, 50, result.Children[1].Y)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Height)
	})
}

func TestLinearNode_FlexEdgeCases(t *testing.T) {
	t.Run("NoFlexChildren_ExistingBehavior", func(t *testing.T) {
		// Verify existing behavior unchanged when no FlexNodes present
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),
				box(30, 10),
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Children should not expand to fill space
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 20, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 20, result.Children[1].X)
		assert.Equal(t, 30, result.Children[1].Layout.Box.Width)
	})

	t.Run("FlexZero_TreatedAsDefault", func(t *testing.T) {
		// Flex: 0 should be normalized to Flex: 1
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 0, Child: box(0, 10)},
				&FlexNode{Flex: 0, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Both treated as Flex: 1, so equal split
		assert.Equal(t, 50, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Width)
	})

	t.Run("NegativeRemaining_FlexChildrenGetZero", func(t *testing.T) {
		// Fixed children exceed container, negative remaining space
		// Flex children should get 0 space (or minimum)
		row := &RowNode{
			Children: []LayoutNode{
				box(60, 10),                           // Fixed
				box(60, 10),                           // Fixed (total 120 > 100)
				&FlexNode{Flex: 1, Child: box(0, 10)}, // Flex
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Flex child gets 0 or its natural minimum
		// Key: no crash, and fixed children are positioned correctly
		assert.Equal(t, 0, result.Children[0].X)
		assert.Equal(t, 60, result.Children[1].X)
		// Flex child position and size are implementation-defined
		// but it should exist in output
		assert.Len(t, result.Children, 3)
	})

	t.Run("VerySmallFlex_RoundingBehavior", func(t *testing.T) {
		// Container: 100px
		// Flex 1:100 ratio in a 100px container
		// Child 0: 100 * 1/101 ≈ 0.99 → rounds to 0 or 1
		// Child 1: 100 * 100/101 ≈ 99.01 → rounds to 99 or 100
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 100, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Total should still be ~100 (accounting for rounding)
		total := result.Children[0].Layout.Box.Width + result.Children[1].Layout.Box.Width
		assert.InDelta(t, 100, total, 2, "total should be approximately 100")

		// The larger flex should get significantly more space
		assert.Greater(t, result.Children[1].Layout.Box.Width, result.Children[0].Layout.Box.Width)
	})

	t.Run("LooseConstraints_FlexChildrenExpand", func(t *testing.T) {
		// Design decision: With loose constraints and flex children,
		// container expands to maxWidth and flex children fill space.
		// This matches TUI user expectations where Fr(1) means "fill remaining".
		// If there are NO flex children, container shrink-wraps (existing behavior).
		// If there ARE flex children, container expands to give them space.
		row := &RowNode{
			Children: []LayoutNode{
				box(20, 10),
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Loose(100, 20))

		// Container expands to maxWidth because there's a flex child
		assert.Equal(t, 100, result.Box.Width,
			"container should expand to maxWidth when flex children present")

		// Flex child fills remaining space
		assert.Equal(t, 80, result.Children[1].Layout.Box.Width,
			"flex child should fill remaining space")
	})
}

func TestLinearNode_FlexCrossAxis(t *testing.T) {
	// Flex only affects main axis. Cross-axis behavior should be unchanged.

	t.Run("FlexChildren_CrossAxisStretch", func(t *testing.T) {
		// Container: 100x50
		// Flex children should stretch vertically (cross-axis)
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)}, // Natural height 10
				&FlexNode{Flex: 1, Child: box(0, 15)}, // Natural height 15
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// Both should stretch to container height (50)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Height,
			"flex children should stretch on cross-axis")
		assert.Equal(t, 50, result.Children[1].Layout.Box.Height)
	})

	t.Run("FlexChild_StretchPreservesFlexWidth", func(t *testing.T) {
		// Child wants to be 50x50 natural.
		// Flex allocation gives it 100 width.
		// CrossAxisStretch should force it to 200 height.
		// BUG: Without the fix, width reverts to 50 during stretch re-layout.
		row := &RowNode{
			CrossAlign: CrossAxisStretch,
			Children: []LayoutNode{
				&FlexNode{
					Flex:  1,
					Child: box(50, 50),
				},
			},
		}

		// Parent provides 100 width, 200 height
		result := row.ComputeLayout(Tight(100, 200))

		assert.Equal(t, 200, result.Children[0].Layout.Box.Height,
			"should stretch vertically")
		assert.Equal(t, 100, result.Children[0].Layout.Box.Width,
			"should maintain allocated flex width, not revert to natural width")
	})

	t.Run("FlexChildren_CrossAxisCenter", func(t *testing.T) {
		// Container: 100x50
		// Flex children should be centered vertically
		row := &RowNode{
			CrossAlign: CrossAxisCenter,
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 20)},
			},
		}
		result := row.ComputeLayout(Tight(100, 50))

		// Child 0: (50-10)/2 = 20
		// Child 1: (50-20)/2 = 15
		assert.Equal(t, 20, result.Children[0].Y)
		assert.Equal(t, 15, result.Children[1].Y)
	})
}

func TestLinearNode_FlexMinWidthConflict(t *testing.T) {
	// Design decision: Parent constraints are authoritative.
	// Flex allocation is the "parent constraint" for flex children.
	// MinWidth is a preference that gets clamped to available space.

	t.Run("TwoFlexChildren_OneHasLargeMinWidth_BothClampedToShare", func(t *testing.T) {
		// Container: 100px
		// Flex 1:1 would give 50px each
		// Child 1 has MinWidth: 70, but gets clamped to 50px (its share)
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: &BoxNode{MinWidth: 70, Height: 10}},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Both children get their allocated share (50px each)
		// MinWidth: 70 is clamped to available 50px
		assert.Len(t, result.Children, 2)
		assert.Equal(t, 50, result.Children[0].Layout.Box.Width)
		assert.Equal(t, 50, result.Children[1].Layout.Box.Width,
			"MinWidth clamped to allocated share")
	})
}

// =============================================================================
// Rounding and Pixel Distribution Tests
// =============================================================================

func TestLinearNode_FlexRounding(t *testing.T) {
	t.Run("ThreeEqualFlex_In100px", func(t *testing.T) {
		// 100 / 3 = 33.33, must distribute remainder
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(100, 20))

		// Total must be exactly 100
		total := result.Children[0].Layout.Box.Width +
			result.Children[1].Layout.Box.Width +
			result.Children[2].Layout.Box.Width
		assert.Equal(t, 100, total, "total width must exactly match container")

		// Each child should be 33 or 34 (differ by at most 1)
		widths := []int{
			result.Children[0].Layout.Box.Width,
			result.Children[1].Layout.Box.Width,
			result.Children[2].Layout.Box.Width,
		}
		minW, maxW := widths[0], widths[0]
		for _, w := range widths {
			if w < minW {
				minW = w
			}
			if w > maxW {
				maxW = w
			}
		}
		assert.LessOrEqual(t, maxW-minW, 1, "widths should differ by at most 1")
	})

	t.Run("FlexWithRemainder_NoPixelLoss", func(t *testing.T) {
		// Container: 103px, two equal flex = 51.5 each
		row := &RowNode{
			Children: []LayoutNode{
				&FlexNode{Flex: 1, Child: box(0, 10)},
				&FlexNode{Flex: 1, Child: box(0, 10)},
			},
		}
		result := row.ComputeLayout(Tight(103, 20))

		// Total must be exactly 103
		total := result.Children[0].Layout.Box.Width + result.Children[1].Layout.Box.Width
		assert.Equal(t, 103, total, "no pixel loss allowed")

		// Widths should be 51 and 52 (or both could be handled differently)
		assert.InDelta(t, 51.5, float64(result.Children[0].Layout.Box.Width), 0.5)
		assert.InDelta(t, 51.5, float64(result.Children[1].Layout.Box.Width), 0.5)
	})
}
