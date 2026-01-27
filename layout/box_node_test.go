package layout

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxNode_FixedSize(t *testing.T) {
	t.Run("LooseConstraints", func(t *testing.T) {
		node := &BoxNode{
			Width:  50,
			Height: 30,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 50, result.Box.Width)
		assert.Equal(t, 30, result.Box.Height)
		assert.Nil(t, result.Children)
	})

	t.Run("TightConstraints_Stretch", func(t *testing.T) {
		// Parent says "be exactly 100x80"
		node := &BoxNode{
			Width:  50,
			Height: 30,
		}

		result := node.ComputeLayout(Tight(100, 80))

		// Should stretch to parent's tight constraints
		assert.Equal(t, 100, result.Box.Width)
		assert.Equal(t, 80, result.Box.Height)
	})

	t.Run("TightHeight_Stretch", func(t *testing.T) {
		// Parent says "be exactly 80 tall, width flexible"
		node := &BoxNode{
			Width:  50,
			Height: 30,
		}

		result := node.ComputeLayout(TightHeight(100, 80))

		assert.Equal(t, 50, result.Box.Width, "width stays at preferred")
		assert.Equal(t, 80, result.Box.Height, "height stretches to tight constraint")
	})

	t.Run("ConstrainedByMaxWidth", func(t *testing.T) {
		node := &BoxNode{
			Width:  200,
			Height: 30,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 100, result.Box.Width, "clamped to max")
		assert.Equal(t, 30, result.Box.Height)
	})
}

func TestBoxNode_NodeConstraints(t *testing.T) {
	t.Run("MinWidth", func(t *testing.T) {
		node := &BoxNode{
			Width:    20,
			Height:   30,
			MinWidth: 50,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 50, result.Box.Width, "clamped to node's MinWidth")
	})

	t.Run("MaxWidth", func(t *testing.T) {
		node := &BoxNode{
			Width:    200,
			Height:   30,
			MaxWidth: 80,
		}

		result := node.ComputeLayout(Loose(100, 100))

		// Node's MaxWidth (80) is applied first, then parent's (100)
		assert.Equal(t, 80, result.Box.Width, "clamped to node's MaxWidth")
	})

	t.Run("MinHeight", func(t *testing.T) {
		node := &BoxNode{
			Width:     50,
			Height:    10,
			MinHeight: 40,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 40, result.Box.Height, "clamped to node's MinHeight")
	})

	t.Run("MaxHeight", func(t *testing.T) {
		node := &BoxNode{
			Width:     50,
			Height:    200,
			MaxHeight: 60,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 60, result.Box.Height, "clamped to node's MaxHeight")
	})

	t.Run("NodeMinExceedsParentMax", func(t *testing.T) {
		// Node wants at least 80, but parent only allows up to 50.
		// Parent constraints are inviolable - they define the available space.
		// If node's min exceeds parent's max, node gets clamped to parent's max.
		// This prevents returning sizes that exceed available space (buffer overflow).
		node := &BoxNode{
			Width:    100,
			Height:   30,
			MinWidth: 80,
		}

		result := node.ComputeLayout(Loose(50, 100))

		// Node's MinWidth (80) > parent's MaxWidth (50) creates a conflict.
		// Parent wins: node is clamped to 50 (the available space).
		assert.Equal(t, 50, result.Box.Width, "parent max wins when it conflicts with node min")
	})
}

func TestBoxNode_MeasureFunc(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		node := &BoxNode{
			Width:  100, // Ignored when MeasureFunc is set
			Height: 100,
			MeasureFunc: func(constraints Constraints) (int, int) {
				// Simulate text that takes 60x20
				return 60, 20
			},
		}

		result := node.ComputeLayout(Loose(200, 200))

		assert.Equal(t, 60, result.Box.Width)
		assert.Equal(t, 20, result.Box.Height)
	})

	t.Run("ReceivesEffectiveConstraints", func(t *testing.T) {
		var received Constraints
		node := &BoxNode{
			MeasureFunc: func(constraints Constraints) (int, int) {
				received = constraints
				return 50, 30
			},
		}

		node.ComputeLayout(Loose(150, 80))

		assert.Equal(t, 0, received.MinWidth)
		assert.Equal(t, 150, received.MaxWidth)
		assert.Equal(t, 0, received.MinHeight)
		assert.Equal(t, 80, received.MaxHeight)
	})

	t.Run("ReceivesEffectiveConstraints_WithNodeMax", func(t *testing.T) {
		// Node's MaxWidth should tighten the effective constraint
		var received Constraints
		node := &BoxNode{
			MaxWidth: 100, // Tighter than parent's 200
			MeasureFunc: func(constraints Constraints) (int, int) {
				received = constraints
				return constraints.MaxWidth, 30
			},
		}

		node.ComputeLayout(Loose(200, 80))

		assert.Equal(t, 100, received.MaxWidth, "effective MaxWidth should be node's tighter constraint")
	})

	t.Run("ConstrainedByParent", func(t *testing.T) {
		node := &BoxNode{
			MeasureFunc: func(constraints Constraints) (int, int) {
				return 200, 200 // Wants more than available
			},
		}

		result := node.ComputeLayout(Loose(100, 80))

		assert.Equal(t, 100, result.Box.Width, "clamped to parent max")
		assert.Equal(t, 80, result.Box.Height, "clamped to parent max")
	})

	t.Run("WithNodeConstraints", func(t *testing.T) {
		node := &BoxNode{
			MinWidth: 40,
			MaxWidth: 70,
			MeasureFunc: func(constraints Constraints) (int, int) {
				return 30, 50 // Below MinWidth
			},
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 40, result.Box.Width, "clamped to node's MinWidth")
	})

	t.Run("TextReflowScenario", func(t *testing.T) {
		// This test verifies the reflow fix:
		// A text measurer needs to know the actual width limit to calculate correct height.
		// Without the fix, MeasureFunc would receive parent's MaxWidth (500),
		// calculate height for 500-wide text, then get clamped to 200 - wrong height!
		node := &BoxNode{
			MaxWidth: 200, // Node says "I can only be 200 wide"
			MeasureFunc: func(constraints Constraints) (int, int) {
				// Simulate text that wraps: narrower width = taller height
				if constraints.MaxWidth <= 200 {
					return 200, 100 // Wrapped: 200 wide, 100 tall (more lines)
				}
				return 400, 50 // Unwrapped: 400 wide, 50 tall (fewer lines)
			},
		}

		result := node.ComputeLayout(Loose(500, 500)) // Parent allows up to 500

		assert.Equal(t, 200, result.Box.Width, "width should be node's max")
		assert.Equal(t, 100, result.Box.Height, "height should be calculated for 200-wide text, not 500")
	})

	t.Run("SpacerExpandsWhenTight", func(t *testing.T) {
		// A spacer should expand to fill tight constraints
		node := &BoxNode{
			MeasureFunc: func(constraints Constraints) (int, int) {
				// Spacer: return min if tight, 0 if loose
				w := 0
				if constraints.IsTightWidth() {
					w = constraints.MinWidth
				}
				h := 0
				if constraints.IsTightHeight() {
					h = constraints.MinHeight
				}
				return w, h
			},
		}

		// Loose constraints - spacer returns 0
		looseResult := node.ComputeLayout(Loose(100, 100))
		assert.Equal(t, 0, looseResult.Box.Width)
		assert.Equal(t, 0, looseResult.Box.Height)

		// Tight constraints - spacer expands
		tightResult := node.ComputeLayout(Tight(50, 30))
		assert.Equal(t, 50, tightResult.Box.Width)
		assert.Equal(t, 30, tightResult.Box.Height)
	})

	t.Run("ContentBoxSemantics_WithPadding", func(t *testing.T) {
		// MeasureFunc receives content-box constraints and returns content-box dimensions.
		// ComputeLayout adds padding/border back automatically.
		var received Constraints
		node := &BoxNode{
			Padding: EdgeInsets{Top: 5, Right: 10, Bottom: 5, Left: 10}, // 20 horizontal, 10 vertical
			MeasureFunc: func(constraints Constraints) (int, int) {
				received = constraints
				// Return content size (e.g., measured text)
				return 60, 20
			},
		}

		result := node.ComputeLayout(Loose(200, 100))

		// MeasureFunc should receive content-box constraints:
		// MaxWidth: 200 - 20 (padding) = 180
		// MaxHeight: 100 - 10 (padding) = 90
		assert.Equal(t, 180, received.MaxWidth, "MeasureFunc should receive content-box MaxWidth")
		assert.Equal(t, 90, received.MaxHeight, "MeasureFunc should receive content-box MaxHeight")

		// Result should be content + padding = border-box:
		// Width: 60 + 20 = 80
		// Height: 20 + 10 = 30
		assert.Equal(t, 80, result.Box.Width, "result should be content + padding (border-box)")
		assert.Equal(t, 30, result.Box.Height, "result should be content + padding (border-box)")

		// Content dimensions should match what MeasureFunc returned
		assert.Equal(t, 60, result.Box.ContentWidth())
		assert.Equal(t, 20, result.Box.ContentHeight())
	})

	t.Run("ContentBoxSemantics_WithPaddingAndBorder", func(t *testing.T) {
		var received Constraints
		node := &BoxNode{
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}, // 10 horizontal, 10 vertical
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1}, // 2 horizontal, 2 vertical
			MeasureFunc: func(constraints Constraints) (int, int) {
				received = constraints
				return 50, 30
			},
		}

		result := node.ComputeLayout(Loose(100, 80))

		// Content-box constraints: 100 - 12 = 88 width, 80 - 12 = 68 height
		assert.Equal(t, 88, received.MaxWidth)
		assert.Equal(t, 68, received.MaxHeight)

		// Border-box result: 50 + 12 = 62 width, 30 + 12 = 42 height
		assert.Equal(t, 62, result.Box.Width)
		assert.Equal(t, 42, result.Box.Height)
	})

	t.Run("ContentBoxSemantics_TextMeasurement", func(t *testing.T) {
		// Simulates real text measurement: user just returns text dimensions,
		// doesn't need to think about padding/border at all.
		node := &BoxNode{
			Padding: EdgeInsets{Top: 8, Right: 16, Bottom: 8, Left: 16}, // Typical text padding
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			MeasureFunc: func(constraints Constraints) (int, int) {
				// Simple text measurement - just return the text size
				// User doesn't need to add padding/border!
				textWidth := 100
				textHeight := 20
				return textWidth, textHeight
			},
		}

		result := node.ComputeLayout(Loose(500, 500))

		// Border-box = content + padding + border
		// Width: 100 + 32 + 2 = 134
		// Height: 20 + 16 + 2 = 38
		assert.Equal(t, 134, result.Box.Width)
		assert.Equal(t, 38, result.Box.Height)

		// Content dimensions match what we measured
		assert.Equal(t, 100, result.Box.ContentWidth())
		assert.Equal(t, 20, result.Box.ContentHeight())
	})

	t.Run("InsetsExceedAvailableSpace", func(t *testing.T) {
		// When padding/border exceed parent's max, the box is clamped
		// and content area collapses to zero.
		var receivedMaxWidth int
		node := &BoxNode{
			Padding: EdgeInsets{Top: 0, Right: 30, Bottom: 0, Left: 30}, // 60 horizontal
			MeasureFunc: func(constraints Constraints) (int, int) {
				receivedMaxWidth = constraints.MaxWidth
				return 0, 20 // Content can't fit, return minimal
			},
		}

		result := node.ComputeLayout(Loose(50, 100)) // Only 50 available, but 60 padding

		// MeasureFunc receives max(0, 50-60) = 0 content space
		assert.Equal(t, 0, receivedMaxWidth)

		// Final size is clamped to parent's max (50), not padding (60)
		assert.Equal(t, 50, result.Box.Width, "clamped to parent constraint")

		// Content area collapses to zero (50 - 60 padding = negative, clamped to 0)
		assert.Equal(t, 0, result.Box.ContentWidth(), "content collapses when insets exceed box")
	})
}

func TestBoxNode_Insets(t *testing.T) {
	t.Run("PassedToBoxModel", func(t *testing.T) {
		node := &BoxNode{
			Width:   100,
			Height:  50,
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
			Margin:  EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
		}

		result := node.ComputeLayout(Loose(200, 200))

		assert.Equal(t, node.Padding, result.Box.Padding)
		assert.Equal(t, node.Border, result.Box.Border)
		assert.Equal(t, node.Margin, result.Box.Margin)
	})

	t.Run("ContentDimensionsComputed", func(t *testing.T) {
		node := &BoxNode{
			Width:   100, // Border-box
			Height:  50,
			Padding: EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
			Border:  EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}

		result := node.ComputeLayout(Loose(200, 200))

		// Content = BorderBox - Padding - Border
		// Width: 100 - 10 - 2 = 88
		// Height: 50 - 10 - 2 = 38
		assert.Equal(t, 88, result.Box.ContentWidth())
		assert.Equal(t, 38, result.Box.ContentHeight())
	})
}

func TestBoxNode_ZeroSize(t *testing.T) {
	t.Run("ZeroWidthHeight", func(t *testing.T) {
		node := &BoxNode{
			Width:  0,
			Height: 0,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 0, result.Box.Width)
		assert.Equal(t, 0, result.Box.Height)
	})

	t.Run("ZeroWithMinConstraint", func(t *testing.T) {
		node := &BoxNode{
			Width:     0,
			Height:    0,
			MinWidth:  20,
			MinHeight: 10,
		}

		result := node.ComputeLayout(Loose(100, 100))

		assert.Equal(t, 20, result.Box.Width)
		assert.Equal(t, 10, result.Box.Height)
	})
}

func TestBoxNode_IsLeafNode(t *testing.T) {
	node := &BoxNode{Width: 50, Height: 30}
	result := node.ComputeLayout(Loose(100, 100))

	assert.Nil(t, result.Children, "BoxNode is a leaf node with no children")
}

func TestBoxNode_ExpandInUnboundedContext_Panics(t *testing.T) {
	t.Run("ExpandHeight with unbounded height", func(t *testing.T) {
		node := &BoxNode{ExpandHeight: true}
		assert.Panics(t, func() {
			node.ComputeLayout(Constraints{
				MinWidth: 0, MaxWidth: 100,
				MinHeight: 0, MaxHeight: maxInt,
			})
		})
	})

	t.Run("ExpandHeight with unbounded height after inset subtraction", func(t *testing.T) {
		node := &BoxNode{ExpandHeight: true}
		assert.Panics(t, func() {
			node.ComputeLayout(Constraints{
				MinWidth: 0, MaxWidth: 100,
				MinHeight: 0, MaxHeight: maxInt - 20,
			})
		})
	})

	t.Run("ExpandWidth with unbounded width", func(t *testing.T) {
		node := &BoxNode{ExpandWidth: true}
		assert.Panics(t, func() {
			node.ComputeLayout(Constraints{
				MinWidth: 0, MaxWidth: maxInt,
				MinHeight: 0, MaxHeight: 100,
			})
		})
	})

	t.Run("ExpandHeight with bounded height does not panic", func(t *testing.T) {
		node := &BoxNode{ExpandHeight: true}
		assert.NotPanics(t, func() {
			node.ComputeLayout(Loose(100, 50))
		})
	})

	t.Run("ExpandWidth with bounded width does not panic", func(t *testing.T) {
		node := &BoxNode{ExpandWidth: true}
		assert.NotPanics(t, func() {
			node.ComputeLayout(Loose(100, 50))
		})
	})
}
