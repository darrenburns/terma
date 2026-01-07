package layout

import (
	"terma"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxModel_DimensionMethods(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:        terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	// Content: 100x50
	// + Padding (5+5, 5+5): 110x60
	// + Border (1+1, 1+1): 112x62
	// + Margin (10+10, 10+10): 132x82

	t.Run("PaddingBoxWidth", func(t *testing.T) {
		assert.Equal(t, 110, box.PaddingBoxWidth())
	})

	t.Run("PaddingBoxHeight", func(t *testing.T) {
		assert.Equal(t, 60, box.PaddingBoxHeight())
	})

	t.Run("BorderBoxWidth", func(t *testing.T) {
		assert.Equal(t, 112, box.BorderBoxWidth())
	})

	t.Run("BorderBoxHeight", func(t *testing.T) {
		assert.Equal(t, 62, box.BorderBoxHeight())
	})

	t.Run("MarginBoxWidth", func(t *testing.T) {
		assert.Equal(t, 132, box.MarginBoxWidth())
	})

	t.Run("MarginBoxHeight", func(t *testing.T) {
		assert.Equal(t, 82, box.MarginBoxHeight())
	})
}

func TestBoxModel_RectMethods(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:        terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	t.Run("ContentBox", func(t *testing.T) {
		rect := box.ContentBox()
		// X = margin.left + border.left + padding.left = 10 + 1 + 5 = 16
		// Y = margin.top + border.top + padding.top = 10 + 1 + 5 = 16
		assert.Equal(t, terma.Rect{X: 16, Y: 16, Width: 100, Height: 50}, rect)
	})

	t.Run("PaddingBox", func(t *testing.T) {
		rect := box.PaddingBox()
		// X = margin.left + border.left = 10 + 1 = 11
		// Y = margin.top + border.top = 10 + 1 = 11
		assert.Equal(t, terma.Rect{X: 11, Y: 11, Width: 110, Height: 60}, rect)
	})

	t.Run("BorderBox", func(t *testing.T) {
		rect := box.BorderBox()
		// X = margin.left = 10
		// Y = margin.top = 10
		assert.Equal(t, terma.Rect{X: 10, Y: 10, Width: 112, Height: 62}, rect)
	})

	t.Run("MarginBox", func(t *testing.T) {
		rect := box.MarginBox()
		// Always at origin
		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 132, Height: 82}, rect)
	})
}

func TestBoxModel_ContentOrigin(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:        terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	x, y := box.ContentOrigin()
	// X = margin.left + border.left + padding.left = 10 + 1 + 5 = 16
	// Y = margin.top + border.top + padding.top = 10 + 1 + 5 = 16
	assert.Equal(t, 16, x)
	assert.Equal(t, 16, y)
}

func TestBoxModel_AsymmetricInsets(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 2, Right: 4, Bottom: 6, Left: 8},
		Border:        terma.EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Margin:        terma.EdgeInsets{Top: 5, Right: 10, Bottom: 15, Left: 20},
	}

	// Padding: horizontal = 8 + 4 = 12, vertical = 2 + 6 = 8
	// Border: horizontal = 2 + 2 = 4, vertical = 1 + 1 = 2
	// Margin: horizontal = 20 + 10 = 30, vertical = 5 + 15 = 20

	t.Run("PaddingBoxDimensions", func(t *testing.T) {
		assert.Equal(t, 112, box.PaddingBoxWidth())
		assert.Equal(t, 58, box.PaddingBoxHeight())
	})

	t.Run("BorderBoxDimensions", func(t *testing.T) {
		assert.Equal(t, 116, box.BorderBoxWidth())
		assert.Equal(t, 60, box.BorderBoxHeight())
	})

	t.Run("MarginBoxDimensions", func(t *testing.T) {
		assert.Equal(t, 146, box.MarginBoxWidth())
		assert.Equal(t, 80, box.MarginBoxHeight())
	})

	t.Run("ContentBox position", func(t *testing.T) {
		rect := box.ContentBox()
		// X = margin.left + border.left + padding.left = 20 + 2 + 8 = 30
		// Y = margin.top + border.top + padding.top = 5 + 1 + 2 = 8
		assert.Equal(t, 30, rect.X)
		assert.Equal(t, 8, rect.Y)
	})
}

func TestBoxModel_ZeroInsets(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
	}

	t.Run("AllBoxesSameSize", func(t *testing.T) {
		assert.Equal(t, 100, box.MarginBoxWidth())
		assert.Equal(t, 50, box.MarginBoxHeight())
	})

	t.Run("ContentAtOrigin", func(t *testing.T) {
		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 100, Height: 50}, box.ContentBox())
	})
}

func TestBoxModel_TotalInsets(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:        terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	t.Run("TotalHorizontalInset", func(t *testing.T) {
		// 10 + 10 (margin) + 2 (border) + 10 (padding) = 32
		assert.Equal(t, 32, box.TotalHorizontalInset())
	})

	t.Run("TotalVerticalInset", func(t *testing.T) {
		// 10 + 10 (margin) + 2 (border) + 10 (padding) = 32
		assert.Equal(t, 32, box.TotalVerticalInset())
	})
}

func TestBoxModel_Constraints(t *testing.T) {
	t.Run("ClampContentWidth", func(t *testing.T) {
		box := BoxModel{
			MinContentWidth: 50,
			MaxContentWidth: 150,
		}

		assert.Equal(t, 50, box.ClampContentWidth(30), "below min")
		assert.Equal(t, 50, box.ClampContentWidth(50), "at min")
		assert.Equal(t, 100, box.ClampContentWidth(100), "in range")
		assert.Equal(t, 150, box.ClampContentWidth(150), "at max")
		assert.Equal(t, 150, box.ClampContentWidth(200), "above max")
	})

	t.Run("ClampContentHeight", func(t *testing.T) {
		box := BoxModel{
			MinContentHeight: 50,
			MaxContentHeight: 150,
		}

		assert.Equal(t, 50, box.ClampContentHeight(30), "below min")
		assert.Equal(t, 50, box.ClampContentHeight(50), "at min")
		assert.Equal(t, 100, box.ClampContentHeight(100), "in range")
		assert.Equal(t, 150, box.ClampContentHeight(150), "at max")
		assert.Equal(t, 150, box.ClampContentHeight(200), "above max")
	})

	t.Run("NoConstraints", func(t *testing.T) {
		box := BoxModel{} // zero constraints mean no constraint

		assert.Equal(t, 0, box.ClampContentWidth(0))
		assert.Equal(t, 1000, box.ClampContentWidth(1000))
	})

	t.Run("OnlyMinConstraint", func(t *testing.T) {
		box := BoxModel{MinContentWidth: 50}

		assert.Equal(t, 50, box.ClampContentWidth(30))
		assert.Equal(t, 1000, box.ClampContentWidth(1000), "no max constraint")
	})

	t.Run("OnlyMaxConstraint", func(t *testing.T) {
		box := BoxModel{MaxContentWidth: 100}

		assert.Equal(t, 0, box.ClampContentWidth(0), "no min constraint")
		assert.Equal(t, 100, box.ClampContentWidth(150))
	})
}

func TestBoxModel_SatisfiesConstraints(t *testing.T) {
	t.Run("WithinConstraints", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:     100,
			ContentHeight:    50,
			MinContentWidth:  50,
			MaxContentWidth:  150,
			MinContentHeight: 25,
			MaxContentHeight: 75,
		}

		assert.True(t, box.SatisfiesConstraints())
	})

	t.Run("BelowMinWidth", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:    30,
			ContentHeight:   50,
			MinContentWidth: 50,
		}

		assert.False(t, box.SatisfiesWidthConstraints())
		assert.False(t, box.SatisfiesConstraints())
	})

	t.Run("AboveMaxHeight", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:     100,
			ContentHeight:    100,
			MaxContentHeight: 75,
		}

		assert.False(t, box.SatisfiesHeightConstraints())
		assert.False(t, box.SatisfiesConstraints())
	})

	t.Run("NoConstraints", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
		}

		assert.True(t, box.SatisfiesConstraints())
	})
}

func TestBoxModel_WithClampedContent(t *testing.T) {
	box := BoxModel{
		ContentWidth:     200,
		ContentHeight:    10,
		MinContentWidth:  50,
		MaxContentWidth:  150,
		MinContentHeight: 25,
		MaxContentHeight: 75,
	}

	clamped := box.WithClampedContent()

	assert.Equal(t, 150, clamped.ContentWidth)
	assert.Equal(t, 25, clamped.ContentHeight)

	// Original should be unchanged
	assert.Equal(t, 200, box.ContentWidth, "original should be unchanged")
}

func TestBoxModel_Scrolling(t *testing.T) {
	t.Run("NonScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
		}

		assert.False(t, box.IsScrollable())
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})

	t.Run("VerticallyScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualHeight: 200,
		}

		assert.True(t, box.IsScrollable())
		assert.False(t, box.IsScrollableX())
		assert.True(t, box.IsScrollableY())
	})

	t.Run("HorizontallyScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualWidth:  300,
		}

		assert.True(t, box.IsScrollable())
		assert.True(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})

	t.Run("BothScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualWidth:  300,
			VirtualHeight: 200,
		}

		assert.True(t, box.IsScrollable())
		assert.True(t, box.IsScrollableX())
		assert.True(t, box.IsScrollableY())
	})
}

func TestBoxModel_MaxScroll(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		VirtualWidth:  300,
		VirtualHeight: 200,
	}

	t.Run("MaxScrollX", func(t *testing.T) {
		// 300 - 100 = 200
		assert.Equal(t, 200, box.MaxScrollX())
	})

	t.Run("MaxScrollY", func(t *testing.T) {
		// 200 - 50 = 150
		assert.Equal(t, 150, box.MaxScrollY())
	})

	t.Run("NonScrollableMaxScroll", func(t *testing.T) {
		nonScrollable := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
		}

		assert.Equal(t, 0, nonScrollable.MaxScrollX())
		assert.Equal(t, 0, nonScrollable.MaxScrollY())
	})
}

func TestBoxModel_ClampScrollOffset(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		VirtualWidth:  300,
		VirtualHeight: 200,
	}
	// MaxScrollX = 200, MaxScrollY = 150

	t.Run("ClampScrollOffsetX", func(t *testing.T) {
		assert.Equal(t, 0, box.ClampScrollOffsetX(-10), "negative")
		assert.Equal(t, 0, box.ClampScrollOffsetX(0), "at min")
		assert.Equal(t, 100, box.ClampScrollOffsetX(100), "in range")
		assert.Equal(t, 200, box.ClampScrollOffsetX(200), "at max")
		assert.Equal(t, 200, box.ClampScrollOffsetX(300), "above max")
	})

	t.Run("ClampScrollOffsetY", func(t *testing.T) {
		assert.Equal(t, 0, box.ClampScrollOffsetY(-10), "negative")
		assert.Equal(t, 0, box.ClampScrollOffsetY(0), "at min")
		assert.Equal(t, 75, box.ClampScrollOffsetY(75), "in range")
		assert.Equal(t, 150, box.ClampScrollOffsetY(150), "at max")
		assert.Equal(t, 150, box.ClampScrollOffsetY(200), "above max")
	})
}

func TestBoxModel_WithClampedScrollOffset(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		VirtualWidth:  300,
		VirtualHeight: 200,
		ScrollOffsetX: 500, // exceeds max
		ScrollOffsetY: -10, // negative
	}

	clamped := box.WithClampedScrollOffset()

	assert.Equal(t, 200, clamped.ScrollOffsetX)
	assert.Equal(t, 0, clamped.ScrollOffsetY)

	// Original should be unchanged
	assert.Equal(t, 500, box.ScrollOffsetX, "original should be unchanged")
}

func TestBoxModel_VisibleContentRect(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		VirtualWidth:  300,
		VirtualHeight: 200,
		ScrollOffsetX: 50,
		ScrollOffsetY: 25,
	}

	assert.Equal(t, terma.Rect{X: 50, Y: 25, Width: 100, Height: 50}, box.VisibleContentRect())
}

func TestBoxModel_VirtualContentRect(t *testing.T) {
	t.Run("WithVirtualSize", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualWidth:  300,
			VirtualHeight: 200,
		}

		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 300, Height: 200}, box.VirtualContentRect())
	})

	t.Run("WithoutVirtualSize", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
		}

		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 100, Height: 50}, box.VirtualContentRect())
	})
}

func TestBoxModel_EffectiveContentWidth(t *testing.T) {
	t.Run("WithScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:   100,
			ContentHeight:  50,
			VirtualHeight:  200, // Makes it scrollable
			ScrollbarWidth: 1,
		}

		assert.Equal(t, 99, box.EffectiveContentWidth())
	})

	t.Run("NotScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:   100,
			ContentHeight:  50,
			ScrollbarWidth: 1, // Has scrollbar width but not scrollable
		}

		// Not scrollable, so scrollbar width shouldn't be subtracted
		assert.Equal(t, 100, box.EffectiveContentWidth())
	})

	t.Run("NoScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualHeight: 200, // Scrollable but no scrollbar
		}

		assert.Equal(t, 100, box.EffectiveContentWidth())
	})
}

func TestBoxModel_EffectiveVirtualDimensions(t *testing.T) {
	t.Run("WithVirtual", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualWidth:  300,
			VirtualHeight: 200,
		}

		assert.Equal(t, 300, box.EffectiveVirtualWidth())
		assert.Equal(t, 200, box.EffectiveVirtualHeight())
	})

	t.Run("WithoutVirtual", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
		}

		assert.Equal(t, 100, box.EffectiveVirtualWidth())
		assert.Equal(t, 50, box.EffectiveVirtualHeight())
	})
}

func TestBoxModel_BuilderMethods(t *testing.T) {
	box := BoxModel{ContentWidth: 100, ContentHeight: 50}

	t.Run("WithContent", func(t *testing.T) {
		result := box.WithContent(200, 100)
		assert.Equal(t, 200, result.ContentWidth)
		assert.Equal(t, 100, result.ContentHeight)
		// Original unchanged
		assert.Equal(t, 100, box.ContentWidth, "original should be unchanged")
	})

	t.Run("WithPadding", func(t *testing.T) {
		padding := terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5}
		result := box.WithPadding(padding)
		assert.Equal(t, padding, result.Padding)
	})

	t.Run("WithBorder", func(t *testing.T) {
		border := terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1}
		result := box.WithBorder(border)
		assert.Equal(t, border, result.Border)
	})

	t.Run("WithMargin", func(t *testing.T) {
		margin := terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10}
		result := box.WithMargin(margin)
		assert.Equal(t, margin, result.Margin)
	})

	t.Run("WithMinContent", func(t *testing.T) {
		result := box.WithMinContent(50, 25)
		assert.Equal(t, 50, result.MinContentWidth)
		assert.Equal(t, 25, result.MinContentHeight)
	})

	t.Run("WithMaxContent", func(t *testing.T) {
		result := box.WithMaxContent(150, 75)
		assert.Equal(t, 150, result.MaxContentWidth)
		assert.Equal(t, 75, result.MaxContentHeight)
	})

	t.Run("WithVirtualSize", func(t *testing.T) {
		result := box.WithVirtualSize(300, 200)
		assert.Equal(t, 300, result.VirtualWidth)
		assert.Equal(t, 200, result.VirtualHeight)
	})

	t.Run("WithScrollOffset", func(t *testing.T) {
		result := box.WithScrollOffset(50, 25)
		assert.Equal(t, 50, result.ScrollOffsetX)
		assert.Equal(t, 25, result.ScrollOffsetY)
	})

	t.Run("WithScrollbarWidth", func(t *testing.T) {
		result := box.WithScrollbarWidth(2)
		assert.Equal(t, 2, result.ScrollbarWidth)
	})

	t.Run("Chaining", func(t *testing.T) {
		result := box.
			WithContent(200, 100).
			WithPadding(terma.EdgeInsetsAll(5)).
			WithBorder(terma.EdgeInsetsAll(1)).
			WithMargin(terma.EdgeInsetsAll(10))

		// MarginBoxWidth = 200 + 10 + 2 + 20 = 232
		assert.Equal(t, 232, result.MarginBoxWidth())
	})
}

func TestBoxModel_BorderOrigin(t *testing.T) {
	box := BoxModel{
		ContentWidth:  100,
		ContentHeight: 50,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:        terma.EdgeInsets{Top: 10, Right: 15, Bottom: 10, Left: 20},
	}

	x, y := box.BorderOrigin()
	// Border starts after margin
	assert.Equal(t, 20, x)
	assert.Equal(t, 10, y)
}
