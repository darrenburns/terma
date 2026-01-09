package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRect_Intersect(t *testing.T) {
	t.Run("overlapping rects", func(t *testing.T) {
		r1 := Rect{X: 0, Y: 0, Width: 100, Height: 100}
		r2 := Rect{X: 50, Y: 50, Width: 100, Height: 100}
		result := r1.Intersect(r2)
		assert.Equal(t, Rect{X: 50, Y: 50, Width: 50, Height: 50}, result)
	})

	t.Run("contained rect", func(t *testing.T) {
		outer := Rect{X: 0, Y: 0, Width: 100, Height: 100}
		inner := Rect{X: 10, Y: 10, Width: 20, Height: 20}
		result := outer.Intersect(inner)
		assert.Equal(t, inner, result)
	})

	t.Run("non-overlapping rects", func(t *testing.T) {
		r1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
		r2 := Rect{X: 20, Y: 20, Width: 10, Height: 10}
		result := r1.Intersect(r2)
		assert.True(t, result.IsEmpty())
	})

	t.Run("adjacent rects (touching but not overlapping)", func(t *testing.T) {
		r1 := Rect{X: 0, Y: 0, Width: 10, Height: 10}
		r2 := Rect{X: 10, Y: 0, Width: 10, Height: 10}
		result := r1.Intersect(r2)
		assert.True(t, result.IsEmpty())
	})
}

func TestRect_IsEmpty(t *testing.T) {
	assert.True(t, Rect{}.IsEmpty())
	assert.True(t, Rect{X: 10, Y: 10, Width: 0, Height: 10}.IsEmpty())
	assert.True(t, Rect{X: 10, Y: 10, Width: 10, Height: 0}.IsEmpty())
	assert.True(t, Rect{X: 10, Y: 10, Width: -1, Height: 10}.IsEmpty())
	assert.False(t, Rect{X: 10, Y: 10, Width: 1, Height: 1}.IsEmpty())
}

func TestSubContext_ClipRectPropagation(t *testing.T) {
	t.Run("SubContext clips to child bounds", func(t *testing.T) {
		// Root context with clip = screen bounds
		root := &RenderContext{
			X: 0, Y: 0,
			Width: 100, Height: 100,
			clip: Rect{X: 0, Y: 0, Width: 100, Height: 100},
		}

		// Create a sub-context at offset (10, 10) with size (50, 50)
		sub := root.SubContext(10, 10, 50, 50)

		// Sub-context should have:
		// - Position at (10, 10)
		// - Clip rect = intersection of parent clip and child bounds
		assert.Equal(t, 10, sub.X)
		assert.Equal(t, 10, sub.Y)
		assert.Equal(t, 50, sub.Width)
		assert.Equal(t, 50, sub.Height)
		assert.Equal(t, Rect{X: 10, Y: 10, Width: 50, Height: 50}, sub.clip)
	})

	t.Run("nested SubContexts accumulate clipping", func(t *testing.T) {
		// Root context
		root := &RenderContext{
			X: 0, Y: 0,
			Width: 100, Height: 100,
			clip: Rect{X: 0, Y: 0, Width: 100, Height: 100},
		}

		// First level: offset (10, 10), size (80, 80)
		level1 := root.SubContext(10, 10, 80, 80)
		assert.Equal(t, Rect{X: 10, Y: 10, Width: 80, Height: 80}, level1.clip)

		// Second level: offset (5, 5) relative to level1, size (30, 30)
		// Absolute position: (15, 15)
		// Clip should be intersection of level1.clip and child bounds
		level2 := level1.SubContext(5, 5, 30, 30)
		assert.Equal(t, 15, level2.X)
		assert.Equal(t, 15, level2.Y)
		assert.Equal(t, Rect{X: 15, Y: 15, Width: 30, Height: 30}, level2.clip)
	})

	t.Run("SubContext clips content that extends beyond parent", func(t *testing.T) {
		// Parent context with limited clip area
		parent := &RenderContext{
			X: 10, Y: 10,
			Width: 50, Height: 50,
			clip: Rect{X: 10, Y: 10, Width: 50, Height: 50},
		}

		// Child tries to extend beyond parent's right edge
		// Child at (40, 0) with size (30, 30) would extend to X=80
		// But parent clip ends at X=60
		child := parent.SubContext(40, 0, 30, 30)

		// Child bounds: (50, 10) to (80, 40)
		// Parent clip: (10, 10) to (60, 60)
		// Intersection: (50, 10) to (60, 40) = width 10, height 30
		assert.Equal(t, 50, child.X)
		assert.Equal(t, 10, child.Y)
		assert.Equal(t, Rect{X: 50, Y: 10, Width: 10, Height: 30}, child.clip)
	})

	t.Run("SubContext with negative offset clips to parent bounds", func(t *testing.T) {
		// This test captures the bug: when content is positioned with
		// MainAxisEnd alignment, child content may have negative-looking
		// positions relative to the parent's content origin, but the
		// clip rect must still be constrained to the parent's content area.
		parent := &RenderContext{
			X: 10, Y: 10,
			Width: 50, Height: 50,
			clip: Rect{X: 10, Y: 10, Width: 50, Height: 50},
		}

		// Child at offset (-5, 0) - starts before parent's origin
		// This simulates content that would clip on the left
		child := parent.SubContext(-5, 0, 30, 30)

		// Child position would be (5, 10)
		// Child bounds: (5, 10) to (35, 40)
		// Parent clip: (10, 10) to (60, 60)
		// Intersection: (10, 10) to (35, 40) = X=10, Y=10, width=25, height=30
		assert.Equal(t, 5, child.X)
		assert.Equal(t, 10, child.Y)
		assert.Equal(t, Rect{X: 10, Y: 10, Width: 25, Height: 30}, child.clip)
	})
}
