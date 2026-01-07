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

	t.Run("ContentBoxVsUsableContentBox", func(t *testing.T) {
		scrollableBox := BoxModel{
			ContentWidth:    100,
			ContentHeight:   50,
			VirtualWidth:    300, // Triggers horizontal scrolling
			VirtualHeight:   200, // Triggers vertical scrolling
			ScrollbarWidth:  2,   // Vertical scrollbar takes width
			ScrollbarHeight: 1,   // Horizontal scrollbar takes height
		}

		contentBox := scrollableBox.ContentBox()
		usableBox := scrollableBox.UsableContentBox()

		// ContentBox = allocated space (what parent gave us)
		assert.Equal(t, 100, contentBox.Width, "ContentBox returns allocated width")
		assert.Equal(t, 50, contentBox.Height, "ContentBox returns allocated height")

		// UsableContentBox = available space for children (after scrollbars)
		assert.Equal(t, 98, usableBox.Width, "UsableContentBox subtracts vertical scrollbar")
		assert.Equal(t, 49, usableBox.Height, "UsableContentBox subtracts horizontal scrollbar")

		// Both have the same position (content origin)
		assert.Equal(t, contentBox.X, usableBox.X, "same X origin")
		assert.Equal(t, contentBox.Y, usableBox.Y, "same Y origin")
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
		// margin (10+10) + border (1+1) + padding (5+5) = 20 + 2 + 10 = 32
		assert.Equal(t, 32, box.TotalHorizontalInset())
	})

	t.Run("TotalVerticalInset", func(t *testing.T) {
		// margin (10+10) + border (1+1) + padding (5+5) = 20 + 2 + 10 = 32
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

func TestBoxModel_UsableContentBox(t *testing.T) {
	// --- Vertical scrollbar tests (affects width) ---

	t.Run("VerticalScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:   100,
			ContentHeight:  50,
			VirtualHeight:  200, // Makes it vertically scrollable
			ScrollbarWidth: 1,
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 99, usable.Width, "vertical scrollbar reduces width")
		assert.Equal(t, 50, usable.Height, "height unchanged")
	})

	t.Run("VerticalScrollbarNotScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:   100,
			ContentHeight:  50,
			ScrollbarWidth: 1, // Has scrollbar width but not scrollable
		}

		// Not scrollable, so scrollbar width shouldn't be subtracted
		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width)
	})

	t.Run("VerticalScrollableNoScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualHeight: 200, // Scrollable but no scrollbar width set
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width)
	})

	// --- Horizontal scrollbar tests (affects height) ---

	t.Run("HorizontalScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:    100,
			ContentHeight:   50,
			VirtualWidth:    200, // Makes it horizontally scrollable
			ScrollbarHeight: 1,
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width, "width unchanged")
		assert.Equal(t, 49, usable.Height, "horizontal scrollbar reduces height")
	})

	t.Run("HorizontalScrollbarNotScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:    100,
			ContentHeight:   50,
			ScrollbarHeight: 1, // Has scrollbar height but not scrollable
		}

		// Not scrollable, so scrollbar height shouldn't be subtracted
		usable := box.UsableContentBox()
		assert.Equal(t, 50, usable.Height)
	})

	t.Run("HorizontalScrollableNoScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			VirtualWidth:  200, // Scrollable but no scrollbar height set
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 50, usable.Height)
	})

	// --- Both scrollbars ---

	t.Run("BothScrollbars", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:    100,
			ContentHeight:   50,
			VirtualWidth:    200, // Horizontally scrollable
			VirtualHeight:   200, // Vertically scrollable
			ScrollbarWidth:  2,
			ScrollbarHeight: 1,
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 98, usable.Width, "vertical scrollbar reduces width")
		assert.Equal(t, 49, usable.Height, "horizontal scrollbar reduces height")
	})

	// --- No scrollbar ---

	t.Run("MatchesContentBoxWhenNoScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth:  100,
			ContentHeight: 50,
			Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}

		content := box.ContentBox()
		usable := box.UsableContentBox()

		// Without scrollbar, both should be identical
		assert.Equal(t, content, usable)
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

	t.Run("WithScrollbarHeight", func(t *testing.T) {
		result := box.WithScrollbarHeight(3)
		assert.Equal(t, 3, result.ScrollbarHeight)
	})

	t.Run("WithScrollbars", func(t *testing.T) {
		result := box.WithScrollbars(2, 3)
		assert.Equal(t, 2, result.ScrollbarWidth)
		assert.Equal(t, 3, result.ScrollbarHeight)
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

func TestBoxModel_ValidationPanics(t *testing.T) {
	// Note: WithContent and WithVirtualSize clamp negative values instead of panicking,
	// since these often come from layout calculations that can legitimately underflow.
	// See TestBoxModel_ClampingBehavior for those tests.

	t.Run("NegativePadding", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithPadding(terma.EdgeInsets{Left: -1})
		})
	})

	t.Run("NegativeBorder", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithBorder(terma.EdgeInsets{Top: -1})
		})
	})

	t.Run("NegativeMargin", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithMargin(terma.EdgeInsets{Right: -1})
		})
	})

	t.Run("NegativeScrollbarWidth", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithScrollbarWidth(-1)
		})
	})

	t.Run("NegativeScrollbarHeight", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithScrollbarHeight(-1)
		})
	})

	t.Run("NegativeScrollbars", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithScrollbars(-1, 1)
		})
		assert.Panics(t, func() {
			BoxModel{}.WithScrollbars(1, -1)
		})
	})

	t.Run("MinWidthExceedsMax", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{MaxContentWidth: 50}.WithMinContent(100, 0)
		})
	})

	t.Run("MinHeightExceedsMax", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{MaxContentHeight: 50}.WithMinContent(0, 100)
		})
	})

	t.Run("NegativeMinContentWidth", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithMinContent(-10, 0)
		})
	})

	t.Run("NegativeMinContentHeight", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithMinContent(0, -10)
		})
	})

	t.Run("NegativeMaxContentWidth", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithMaxContent(-10, 0)
		})
	})

	t.Run("NegativeMaxContentHeight", func(t *testing.T) {
		assert.Panics(t, func() {
			BoxModel{}.WithMaxContent(0, -10)
		})
	})
}

func TestBoxModel_ValidationValid(t *testing.T) {
	t.Run("ZeroValues", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BoxModel{}.WithContent(0, 0)
		})
	})

	t.Run("ValidConstraints", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BoxModel{}.WithMinContent(50, 50).WithMaxContent(100, 100)
		})
	})

	t.Run("EqualMinMax", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BoxModel{}.WithMinContent(50, 50).WithMaxContent(50, 50)
		})
	})
}

func TestBoxModel_ClampingBehavior(t *testing.T) {
	// WithContent and WithVirtualSize clamp negative values to 0 instead of panicking,
	// because these values often come from layout calculations that can legitimately
	// underflow (e.g., when the terminal is resized too small).

	t.Run("WithContentClampsNegativeWidth", func(t *testing.T) {
		box := BoxModel{}.WithContent(-50, 100)
		assert.Equal(t, 0, box.ContentWidth, "negative width clamped to 0")
		assert.Equal(t, 100, box.ContentHeight, "positive height unchanged")
	})

	t.Run("WithContentClampsNegativeHeight", func(t *testing.T) {
		box := BoxModel{}.WithContent(100, -50)
		assert.Equal(t, 100, box.ContentWidth, "positive width unchanged")
		assert.Equal(t, 0, box.ContentHeight, "negative height clamped to 0")
	})

	t.Run("WithContentClampsBothNegative", func(t *testing.T) {
		box := BoxModel{}.WithContent(-100, -50)
		assert.Equal(t, 0, box.ContentWidth)
		assert.Equal(t, 0, box.ContentHeight)
	})

	t.Run("WithContentPreservesZero", func(t *testing.T) {
		box := BoxModel{}.WithContent(0, 0)
		assert.Equal(t, 0, box.ContentWidth)
		assert.Equal(t, 0, box.ContentHeight)
	})

	t.Run("WithContentPreservesPositive", func(t *testing.T) {
		box := BoxModel{}.WithContent(100, 50)
		assert.Equal(t, 100, box.ContentWidth)
		assert.Equal(t, 50, box.ContentHeight)
	})

	t.Run("WithVirtualSizeClampsNegativeWidth", func(t *testing.T) {
		box := BoxModel{}.WithVirtualSize(-50, 100)
		assert.Equal(t, 0, box.VirtualWidth, "negative width clamped to 0")
		assert.Equal(t, 100, box.VirtualHeight, "positive height unchanged")
	})

	t.Run("WithVirtualSizeClampsNegativeHeight", func(t *testing.T) {
		box := BoxModel{}.WithVirtualSize(100, -50)
		assert.Equal(t, 100, box.VirtualWidth, "positive width unchanged")
		assert.Equal(t, 0, box.VirtualHeight, "negative height clamped to 0")
	})

	t.Run("WithVirtualSizeClampsBothNegative", func(t *testing.T) {
		box := BoxModel{}.WithVirtualSize(-100, -50)
		assert.Equal(t, 0, box.VirtualWidth)
		assert.Equal(t, 0, box.VirtualHeight)
	})

	t.Run("WithVirtualSizePreservesZero", func(t *testing.T) {
		box := BoxModel{}.WithVirtualSize(0, 0)
		assert.Equal(t, 0, box.VirtualWidth)
		assert.Equal(t, 0, box.VirtualHeight)
	})

	t.Run("WithVirtualSizePreservesPositive", func(t *testing.T) {
		box := BoxModel{}.WithVirtualSize(300, 200)
		assert.Equal(t, 300, box.VirtualWidth)
		assert.Equal(t, 200, box.VirtualHeight)
	})

	// Realistic scenario: layout calculation underflow
	t.Run("LayoutUnderflowScenario", func(t *testing.T) {
		containerWidth := 100
		padding := 60
		border := 60
		// This calculation legitimately goes negative when container is too small
		availableWidth := containerWidth - padding - border // = -20

		box := BoxModel{}.WithContent(availableWidth, 50)
		assert.Equal(t, 0, box.ContentWidth, "underflow clamped to 0")
		assert.Equal(t, 50, box.ContentHeight)
	})
}

func TestBoxModel_ClampWithNegativeInput(t *testing.T) {
	box := BoxModel{MinContentWidth: 50, MinContentHeight: 50}

	t.Run("NegativeWidthPanics", func(t *testing.T) {
		assert.Panics(t, func() {
			box.ClampContentWidth(-10)
		})
	})

	t.Run("NegativeHeightPanics", func(t *testing.T) {
		assert.Panics(t, func() {
			box.ClampContentHeight(-10)
		})
	})

	t.Run("ZeroIsValid", func(t *testing.T) {
		assert.NotPanics(t, func() {
			box.ClampContentWidth(0)
			box.ClampContentHeight(0)
		})
	})
}

func TestBoxModel_ContentAtConstraintBoundaries(t *testing.T) {
	t.Run("WidthAtMin", func(t *testing.T) {
		box := BoxModel{ContentWidth: 50, MinContentWidth: 50}
		assert.True(t, box.SatisfiesWidthConstraints())
	})

	t.Run("WidthAtMax", func(t *testing.T) {
		box := BoxModel{ContentWidth: 100, MaxContentWidth: 100}
		assert.True(t, box.SatisfiesWidthConstraints())
	})

	t.Run("HeightAtMin", func(t *testing.T) {
		box := BoxModel{ContentHeight: 50, MinContentHeight: 50}
		assert.True(t, box.SatisfiesHeightConstraints())
	})

	t.Run("HeightAtMax", func(t *testing.T) {
		box := BoxModel{ContentHeight: 100, MaxContentHeight: 100}
		assert.True(t, box.SatisfiesHeightConstraints())
	})

	t.Run("BothAtBoundaries", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 50, ContentHeight: 100,
			MinContentWidth: 50, MaxContentHeight: 100,
		}
		assert.True(t, box.SatisfiesConstraints())
	})
}

func TestBoxModel_VirtualSizeEdgeCases(t *testing.T) {
	t.Run("VirtualEqualsContent", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 100, VirtualHeight: 50,
		}
		assert.False(t, box.IsScrollableX(), "equal virtual width should not be scrollable")
		assert.False(t, box.IsScrollableY(), "equal virtual height should not be scrollable")
		assert.Equal(t, 0, box.MaxScrollX())
		assert.Equal(t, 0, box.MaxScrollY())
	})

	t.Run("VirtualSmallerThanContent", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 50, VirtualHeight: 25,
		}
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
		assert.Equal(t, 0, box.MaxScrollX(), "max scroll should be 0, not negative")
		assert.Equal(t, 0, box.MaxScrollY(), "max scroll should be 0, not negative")
	})
}

func TestBoxModel_ScrollbarWidthEdgeCases(t *testing.T) {
	t.Run("ScrollbarEqualsContentWidth", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualHeight: 200, ScrollbarWidth: 100,
		}
		assert.Equal(t, 0, box.UsableContentBox().Width)
	})

	t.Run("ScrollbarExceedsContentWidth", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualHeight: 200, ScrollbarWidth: 150,
		}
		// Usable width is clamped to 0 when scrollbar exceeds content width
		assert.Equal(t, 0, box.UsableContentBox().Width)
	})

	t.Run("HorizontalScrollOnlyWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, ScrollbarWidth: 10,
		}
		// Scrollbar only applies to vertical scrolling
		assert.Equal(t, 100, box.UsableContentBox().Width)
	})

	t.Run("BothScrollableWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 10,
		}
		assert.Equal(t, 90, box.UsableContentBox().Width)
	})
}

func TestBoxModel_ScrollbarHeightEdgeCases(t *testing.T) {
	t.Run("ScrollbarEqualsContentHeight", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, ScrollbarHeight: 50,
		}
		assert.Equal(t, 0, box.UsableContentBox().Height)
	})

	t.Run("ScrollbarExceedsContentHeight", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, ScrollbarHeight: 75,
		}
		// Usable height is clamped to 0 when scrollbar exceeds content height
		assert.Equal(t, 0, box.UsableContentBox().Height)
	})

	t.Run("VerticalScrollOnlyWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualHeight: 200, ScrollbarHeight: 10,
		}
		// Horizontal scrollbar only applies to horizontal scrolling
		assert.Equal(t, 50, box.UsableContentBox().Height)
	})

	t.Run("BothScrollableWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarHeight: 10,
		}
		assert.Equal(t, 40, box.UsableContentBox().Height)
	})
}

func TestBoxModel_BothScrollbarsEdgeCases(t *testing.T) {
	t.Run("BothScrollbarsReduceBothDimensions", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 10, ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 90, usable.Width, "vertical scrollbar reduces width")
		assert.Equal(t, 45, usable.Height, "horizontal scrollbar reduces height")
	})

	t.Run("BothScrollbarsExceedDimensions", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 150, ScrollbarHeight: 75,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 0, usable.Width, "clamped to 0")
		assert.Equal(t, 0, usable.Height, "clamped to 0")
	})

	t.Run("OnlyVerticalScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualHeight:   200, // Only vertical scrolling
			ScrollbarWidth:  10,
			ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 90, usable.Width, "vertical scrollbar applies")
		assert.Equal(t, 50, usable.Height, "horizontal scrollbar doesn't apply")
	})

	t.Run("OnlyHorizontalScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			VirtualWidth:    200, // Only horizontal scrolling
			ScrollbarWidth:  10,
			ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width, "vertical scrollbar doesn't apply")
		assert.Equal(t, 45, usable.Height, "horizontal scrollbar applies")
	})

	t.Run("NeitherScrollable", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			// No virtual dimensions, so not scrollable
			ScrollbarWidth:  10,
			ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width, "no scrollbar applied")
		assert.Equal(t, 50, usable.Height, "no scrollbar applied")
	})
}

func TestBoxModel_ZeroContentWithInsets(t *testing.T) {
	box := BoxModel{
		ContentWidth: 0, ContentHeight: 0,
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	t.Run("BoxDimensions", func(t *testing.T) {
		assert.Equal(t, 10, box.PaddingBoxWidth())  // 0 + 5 + 5
		assert.Equal(t, 10, box.PaddingBoxHeight()) // 0 + 5 + 5
		assert.Equal(t, 12, box.BorderBoxWidth())   // 10 + 1 + 1
		assert.Equal(t, 12, box.BorderBoxHeight())  // 10 + 1 + 1
		assert.Equal(t, 32, box.MarginBoxWidth())   // 12 + 10 + 10
		assert.Equal(t, 32, box.MarginBoxHeight())  // 12 + 10 + 10
	})

	t.Run("ContentBoxPosition", func(t *testing.T) {
		rect := box.ContentBox()
		assert.Equal(t, 16, rect.X)  // 10 + 1 + 5
		assert.Equal(t, 16, rect.Y)  // 10 + 1 + 5
		assert.Equal(t, 0, rect.Width)
		assert.Equal(t, 0, rect.Height)
	})
}

func TestBoxModel_PartialInsets(t *testing.T) {
	t.Run("OnlyPadding", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}
		assert.Equal(t, 110, box.MarginBoxWidth())
		assert.Equal(t, 60, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 5, rect.X)
		assert.Equal(t, 5, rect.Y)
	})

	t.Run("OnlyBorder", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			Border: terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		assert.Equal(t, 102, box.MarginBoxWidth())
		assert.Equal(t, 52, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 1, rect.X)
		assert.Equal(t, 1, rect.Y)
	})

	t.Run("OnlyMargin", func(t *testing.T) {
		box := BoxModel{
			ContentWidth: 100, ContentHeight: 50,
			Margin: terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
		}
		assert.Equal(t, 120, box.MarginBoxWidth())
		assert.Equal(t, 70, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 10, rect.X)
		assert.Equal(t, 10, rect.Y)
	})
}

func TestBoxModel_HeightOnlyConstraints(t *testing.T) {
	t.Run("OnlyMinHeightConstraint", func(t *testing.T) {
		box := BoxModel{MinContentHeight: 50}
		assert.Equal(t, 50, box.ClampContentHeight(30))
		assert.Equal(t, 1000, box.ClampContentHeight(1000), "no max constraint")
	})

	t.Run("OnlyMaxHeightConstraint", func(t *testing.T) {
		box := BoxModel{MaxContentHeight: 100}
		assert.Equal(t, 0, box.ClampContentHeight(0), "no min constraint")
		assert.Equal(t, 100, box.ClampContentHeight(150))
	})
}
