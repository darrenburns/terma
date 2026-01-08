package layout

import (
	"terma"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoxModel_DimensionMethods(t *testing.T) {
	// Border-box semantics: Width/Height is the border-box dimension.
	// Content is computed by subtracting padding and border.
	box := BoxModel{
		Width:   112, // Border-box width
		Height:  62,  // Border-box height
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	// Border-box: 112x62
	// - Border (1+1, 1+1): PaddingBox = 110x60
	// - Padding (5+5, 5+5): Content = 100x50
	// + Margin (10+10, 10+10): MarginBox = 132x82

	t.Run("ContentWidth", func(t *testing.T) {
		// 112 - 10 (padding) - 2 (border) = 100
		assert.Equal(t, 100, box.ContentWidth())
	})

	t.Run("ContentHeight", func(t *testing.T) {
		// 62 - 10 (padding) - 2 (border) = 50
		assert.Equal(t, 50, box.ContentHeight())
	})

	t.Run("PaddingBoxWidth", func(t *testing.T) {
		// 112 - 2 (border) = 110
		assert.Equal(t, 110, box.PaddingBoxWidth())
	})

	t.Run("PaddingBoxHeight", func(t *testing.T) {
		// 62 - 2 (border) = 60
		assert.Equal(t, 60, box.PaddingBoxHeight())
	})

	t.Run("BorderBoxWidth", func(t *testing.T) {
		// Direct: 112
		assert.Equal(t, 112, box.BorderBoxWidth())
	})

	t.Run("BorderBoxHeight", func(t *testing.T) {
		// Direct: 62
		assert.Equal(t, 62, box.BorderBoxHeight())
	})

	t.Run("MarginBoxWidth", func(t *testing.T) {
		// 112 + 20 (margin) = 132
		assert.Equal(t, 132, box.MarginBoxWidth())
	})

	t.Run("MarginBoxHeight", func(t *testing.T) {
		// 62 + 20 (margin) = 82
		assert.Equal(t, 82, box.MarginBoxHeight())
	})
}

func TestBoxModel_ContentClampedToZero(t *testing.T) {
	// When insets exceed border-box, content should clamp to 0
	t.Run("InsetExceedsBorderBox", func(t *testing.T) {
		box := BoxModel{
			Width:   20,
			Height:  20,
			Padding: terma.EdgeInsets{Top: 15, Right: 15, Bottom: 15, Left: 15},
		}
		assert.Equal(t, 0, box.ContentWidth(), "content clamped to 0")
		assert.Equal(t, 0, box.ContentHeight(), "content clamped to 0")
	})

	t.Run("PaddingBoxClampedToZero", func(t *testing.T) {
		box := BoxModel{
			Width:  10,
			Height: 10,
			Border: terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
		}
		assert.Equal(t, 0, box.PaddingBoxWidth(), "padding-box clamped to 0")
		assert.Equal(t, 0, box.PaddingBoxHeight(), "padding-box clamped to 0")
	})

	t.Run("ContentOriginWhenClamped", func(t *testing.T) {
		// Edge case: padding exceeds border-box width.
		// ContentWidth is clamped to 0, but where is the content origin?
		//
		// CSS semantics: ContentOrigin is computed from insets regardless of
		// available space. This means the origin can be positioned outside
		// the visual border-box bounds when insets exceed the box size.
		// This is intentional - content size is 0, and the origin is where
		// content "would start" if there were space.
		box := BoxModel{
			Width:   10, // Very small border-box
			Padding: terma.EdgeInsets{Left: 20}, // Exceeds width
			Margin:  terma.EdgeInsets{Left: 5},
		}

		// ContentWidth is 0 (clamped)
		assert.Equal(t, 0, box.ContentWidth())

		// ContentOrigin X = Margin.Left + Border.Left + Padding.Left = 5 + 0 + 20 = 25
		// This is outside the border-box (which spans X: 5 to 15).
		// This is intentional: the origin reflects the inset calculation,
		// not a clipped position within the box.
		x, _ := box.ContentOrigin()
		assert.Equal(t, 25, x)

		// Similarly for ContentBox rect
		rect := box.ContentBox()
		assert.Equal(t, 25, rect.X)
		assert.Equal(t, 0, rect.Width)
	})
}

func TestBoxModel_RectMethods(t *testing.T) {
	box := BoxModel{
		Width:   112, // Border-box width
		Height:  62,  // Border-box height
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	t.Run("ContentBox", func(t *testing.T) {
		rect := box.ContentBox()
		// X = margin.left + border.left + padding.left = 10 + 1 + 5 = 16
		// Y = margin.top + border.top + padding.top = 10 + 1 + 5 = 16
		// Width = 112 - 10 - 2 = 100, Height = 62 - 10 - 2 = 50
		assert.Equal(t, terma.Rect{X: 16, Y: 16, Width: 100, Height: 50}, rect)
	})

	t.Run("PaddingBox", func(t *testing.T) {
		rect := box.PaddingBox()
		// X = margin.left + border.left = 10 + 1 = 11
		// Y = margin.top + border.top = 10 + 1 = 11
		// Width = 112 - 2 = 110, Height = 62 - 2 = 60
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
			Width:           100, // Border-box width
			Height:          50,  // Border-box height
			VirtualWidth:    300, // Triggers horizontal scrolling
			VirtualHeight:   200, // Triggers vertical scrolling
			ScrollbarWidth:  2,   // Vertical scrollbar takes width
			ScrollbarHeight: 1,   // Horizontal scrollbar takes height
		}

		contentBox := scrollableBox.ContentBox()
		usableBox := scrollableBox.UsableContentBox()

		// ContentBox = computed content (what's left after padding/border)
		assert.Equal(t, 100, contentBox.Width, "ContentBox returns computed content width")
		assert.Equal(t, 50, contentBox.Height, "ContentBox returns computed content height")

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
		Width:   112,
		Height:  62,
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	x, y := box.ContentOrigin()
	// X = margin.left + border.left + padding.left = 10 + 1 + 5 = 16
	// Y = margin.top + border.top + padding.top = 10 + 1 + 5 = 16
	assert.Equal(t, 16, x)
	assert.Equal(t, 16, y)
}

func TestBoxModel_AsymmetricInsets(t *testing.T) {
	// Border-box: 116x60
	// Content should be 116 - 12 (padding) - 4 (border) = 100 for width
	// Content should be 60 - 8 (padding) - 2 (border) = 50 for height
	box := BoxModel{
		Width:   116,
		Height:  60,
		Padding: terma.EdgeInsets{Top: 2, Right: 4, Bottom: 6, Left: 8},
		Border:  terma.EdgeInsets{Top: 1, Right: 2, Bottom: 1, Left: 2},
		Margin:  terma.EdgeInsets{Top: 5, Right: 10, Bottom: 15, Left: 20},
	}

	// Padding: horizontal = 8 + 4 = 12, vertical = 2 + 6 = 8
	// Border: horizontal = 2 + 2 = 4, vertical = 1 + 1 = 2
	// Margin: horizontal = 20 + 10 = 30, vertical = 5 + 15 = 20

	t.Run("ContentDimensions", func(t *testing.T) {
		// 116 - 12 - 4 = 100
		assert.Equal(t, 100, box.ContentWidth())
		// 60 - 8 - 2 = 50
		assert.Equal(t, 50, box.ContentHeight())
	})

	t.Run("PaddingBoxDimensions", func(t *testing.T) {
		// 116 - 4 = 112
		assert.Equal(t, 112, box.PaddingBoxWidth())
		// 60 - 2 = 58
		assert.Equal(t, 58, box.PaddingBoxHeight())
	})

	t.Run("BorderBoxDimensions", func(t *testing.T) {
		assert.Equal(t, 116, box.BorderBoxWidth())
		assert.Equal(t, 60, box.BorderBoxHeight())
	})

	t.Run("MarginBoxDimensions", func(t *testing.T) {
		// 116 + 30 = 146
		assert.Equal(t, 146, box.MarginBoxWidth())
		// 60 + 20 = 80
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
		Width:  100,
		Height: 50,
	}

	t.Run("AllBoxesSameSize", func(t *testing.T) {
		assert.Equal(t, 100, box.ContentWidth())
		assert.Equal(t, 50, box.ContentHeight())
		assert.Equal(t, 100, box.PaddingBoxWidth())
		assert.Equal(t, 50, box.PaddingBoxHeight())
		assert.Equal(t, 100, box.BorderBoxWidth())
		assert.Equal(t, 50, box.BorderBoxHeight())
		assert.Equal(t, 100, box.MarginBoxWidth())
		assert.Equal(t, 50, box.MarginBoxHeight())
	})

	t.Run("ContentAtOrigin", func(t *testing.T) {
		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 100, Height: 50}, box.ContentBox())
	})
}

func TestBoxModel_TotalInsets(t *testing.T) {
	box := BoxModel{
		Width:   112,
		Height:  62,
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
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

func TestBoxModel_Scrolling(t *testing.T) {
	t.Run("NonScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
		}

		assert.False(t, box.IsScrollable())
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})

	t.Run("VerticallyScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:         100,
			Height:        50,
			VirtualHeight: 200,
		}

		assert.True(t, box.IsScrollable())
		assert.False(t, box.IsScrollableX())
		assert.True(t, box.IsScrollableY())
	})

	t.Run("HorizontallyScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:        100,
			Height:       50,
			VirtualWidth: 300,
		}

		assert.True(t, box.IsScrollable())
		assert.True(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})

	t.Run("BothScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:         100,
			Height:        50,
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
		Width:         100,
		Height:        50,
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
			Width:  100,
			Height: 50,
		}

		assert.Equal(t, 0, nonScrollable.MaxScrollX())
		assert.Equal(t, 0, nonScrollable.MaxScrollY())
	})
}

func TestBoxModel_ClampScrollOffset(t *testing.T) {
	box := BoxModel{
		Width:         100,
		Height:        50,
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
		Width:         100,
		Height:        50,
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
		Width:         100,
		Height:        50,
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
			Width:         100,
			Height:        50,
			VirtualWidth:  300,
			VirtualHeight: 200,
		}

		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 300, Height: 200}, box.VirtualContentRect())
	})

	t.Run("WithoutVirtualSize", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
		}

		assert.Equal(t, terma.Rect{X: 0, Y: 0, Width: 100, Height: 50}, box.VirtualContentRect())
	})
}

func TestBoxModel_UsableContentBox(t *testing.T) {
	// --- Vertical scrollbar tests (affects width) ---

	t.Run("VerticalScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width:          100,
			Height:         50,
			VirtualHeight:  200, // Makes it vertically scrollable
			ScrollbarWidth: 1,
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 99, usable.Width, "vertical scrollbar reduces width")
		assert.Equal(t, 50, usable.Height, "height unchanged")
	})

	t.Run("VerticalScrollbarNotScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:          100,
			Height:         50,
			ScrollbarWidth: 1, // Has scrollbar width but not scrollable
		}

		// Not scrollable, so scrollbar width shouldn't be subtracted
		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width)
	})

	t.Run("VerticalScrollableNoScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width:         100,
			Height:        50,
			VirtualHeight: 200, // Scrollable but no scrollbar width set
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width)
	})

	// --- Horizontal scrollbar tests (affects height) ---

	t.Run("HorizontalScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width:           100,
			Height:          50,
			VirtualWidth:    200, // Makes it horizontally scrollable
			ScrollbarHeight: 1,
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width, "width unchanged")
		assert.Equal(t, 49, usable.Height, "horizontal scrollbar reduces height")
	})

	t.Run("HorizontalScrollbarNotScrollable", func(t *testing.T) {
		box := BoxModel{
			Width:           100,
			Height:          50,
			ScrollbarHeight: 1, // Has scrollbar height but not scrollable
		}

		// Not scrollable, so scrollbar height shouldn't be subtracted
		usable := box.UsableContentBox()
		assert.Equal(t, 50, usable.Height)
	})

	t.Run("HorizontalScrollableNoScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width:        100,
			Height:       50,
			VirtualWidth: 200, // Scrollable but no scrollbar height set
		}

		usable := box.UsableContentBox()
		assert.Equal(t, 50, usable.Height)
	})

	// --- Both scrollbars ---

	t.Run("BothScrollbars", func(t *testing.T) {
		box := BoxModel{
			Width:           100,
			Height:          50,
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
			Width:   110,
			Height:  60,
			Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}

		content := box.ContentBox()
		usable := box.UsableContentBox()

		// Without scrollbar, both should be identical
		assert.Equal(t, content, usable)
	})

	// --- Padding + Scrollbar interaction ---

	t.Run("PaddingAndScrollbarSubtractFromContent", func(t *testing.T) {
		// In border-box, both padding and scrollbar reduce available content space.
		// The scrollbar subtracts from ContentWidth (after padding), not from BorderBox.
		box := BoxModel{
			Width:          100,
			Height:         100,
			VirtualHeight:  200, // Triggers vertical scrollbar
			ScrollbarWidth: 10,
			Padding:        terma.EdgeInsetsAll(10), // 10 on each side
		}

		// 1. BorderBox = 100
		// 2. PaddingBox = 100 (no border)
		// 3. ContentWidth = 100 - 10(L) - 10(R) = 80
		// 4. UsableWidth = 80 - 10(Scrollbar) = 70

		assert.Equal(t, 100, box.BorderBoxWidth())
		assert.Equal(t, 100, box.PaddingBoxWidth())
		assert.Equal(t, 80, box.ContentWidth())
		assert.Equal(t, 70, box.UsableContentBox().Width)
	})

	t.Run("PaddingBorderAndScrollbarAllSubtract", func(t *testing.T) {
		// Full chain: border-box -> padding-box -> content -> usable content
		box := BoxModel{
			Width:          100,
			Height:         100,
			Border:         terma.EdgeInsetsAll(2),  // 2 on each side = 4 total
			Padding:        terma.EdgeInsetsAll(8),  // 8 on each side = 16 total
			VirtualHeight:  200,                     // Triggers vertical scrollbar
			ScrollbarWidth: 5,
		}

		// 1. BorderBox = 100
		// 2. PaddingBox = 100 - 4 = 96
		// 3. ContentWidth = 100 - 4 - 16 = 80
		// 4. UsableWidth = 80 - 5 = 75

		assert.Equal(t, 100, box.BorderBoxWidth())
		assert.Equal(t, 96, box.PaddingBoxWidth())
		assert.Equal(t, 80, box.ContentWidth())
		assert.Equal(t, 75, box.UsableContentBox().Width)
	})

	t.Run("ScrollbarExceedsPaddingReducedContent", func(t *testing.T) {
		// Edge case: Padding reduces content, then scrollbar exceeds that reduced content.
		// Width=50, Padding=10+10=20, ContentWidth=30, ScrollbarWidth=40
		// UsableWidth should be 0, not -10.
		box := BoxModel{
			Width:          50,
			Height:         50,
			Padding:        terma.EdgeInsets{Left: 10, Right: 10},
			VirtualHeight:  200, // Triggers vertical scrollbar
			ScrollbarWidth: 40,  // Exceeds ContentWidth of 30
		}

		// ContentWidth = 50 - 20 = 30
		assert.Equal(t, 30, box.ContentWidth())

		// UsableWidth = max(0, 30 - 40) = max(0, -10) = 0
		assert.Equal(t, 0, box.UsableContentBox().Width, "should clamp to 0, not go negative")
	})

	t.Run("ScrollbarExceedsBorderPaddingReducedContent", func(t *testing.T) {
		// Full chain with scrollbar exceeding available space
		box := BoxModel{
			Width:           60,
			Height:          60,
			Border:          terma.EdgeInsetsAll(5),  // 10 total
			Padding:         terma.EdgeInsetsAll(10), // 20 total
			VirtualHeight:   200,
			ScrollbarWidth:  50, // Exceeds ContentWidth
			VirtualWidth:    200,
			ScrollbarHeight: 50, // Exceeds ContentHeight
		}

		// ContentWidth = 60 - 10 - 20 = 30
		// ContentHeight = 60 - 10 - 20 = 30
		assert.Equal(t, 30, box.ContentWidth())
		assert.Equal(t, 30, box.ContentHeight())

		// UsableWidth = max(0, 30 - 50) = 0
		// UsableHeight = max(0, 30 - 50) = 0
		usable := box.UsableContentBox()
		assert.Equal(t, 0, usable.Width, "clamped to 0")
		assert.Equal(t, 0, usable.Height, "clamped to 0")
	})
}

func TestBoxModel_EffectiveVirtualDimensions(t *testing.T) {
	t.Run("WithVirtual", func(t *testing.T) {
		box := BoxModel{
			Width:         100,
			Height:        50,
			VirtualWidth:  300,
			VirtualHeight: 200,
		}

		assert.Equal(t, 300, box.EffectiveVirtualWidth())
		assert.Equal(t, 200, box.EffectiveVirtualHeight())
	})

	t.Run("WithoutVirtual", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
		}

		assert.Equal(t, 100, box.EffectiveVirtualWidth())
		assert.Equal(t, 50, box.EffectiveVirtualHeight())
	})
}

func TestBoxModel_BuilderMethods(t *testing.T) {
	box := BoxModel{Width: 100, Height: 50}

	t.Run("WithSize", func(t *testing.T) {
		result := box.WithSize(200, 100)
		assert.Equal(t, 200, result.Width)
		assert.Equal(t, 100, result.Height)
		// Original unchanged
		assert.Equal(t, 100, box.Width, "original should be unchanged")
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
			WithSize(200, 100).
			WithPadding(terma.EdgeInsetsAll(5)).
			WithBorder(terma.EdgeInsetsAll(1)).
			WithMargin(terma.EdgeInsetsAll(10))

		// Border-box is 200, MarginBox = 200 + 20 = 220
		assert.Equal(t, 220, result.MarginBoxWidth())
	})
}

func TestBoxModel_BorderOrigin(t *testing.T) {
	box := BoxModel{
		Width:   112,
		Height:  62,
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 15, Bottom: 10, Left: 20},
	}

	x, y := box.BorderOrigin()
	// Border starts after margin
	assert.Equal(t, 20, x)
	assert.Equal(t, 10, y)
}

func TestBoxModel_ValidationPanics(t *testing.T) {
	// Note: WithSize and WithVirtualSize clamp negative values instead of panicking,
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

	// Note: Negative margins are allowed (CSS behavior) - see TestBoxModel_NegativeMargin

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
}

func TestBoxModel_ValidationValid(t *testing.T) {
	t.Run("ZeroValues", func(t *testing.T) {
		assert.NotPanics(t, func() {
			BoxModel{}.WithSize(0, 0)
		})
	})
}

func TestBoxModel_ClampingBehavior(t *testing.T) {
	// WithSize and WithVirtualSize clamp negative values to 0 instead of panicking,
	// because these values often come from layout calculations that can legitimately
	// underflow (e.g., when the terminal is resized too small).

	t.Run("WithSizeClampsNegativeWidth", func(t *testing.T) {
		box := BoxModel{}.WithSize(-50, 100)
		assert.Equal(t, 0, box.Width, "negative width clamped to 0")
		assert.Equal(t, 100, box.Height, "positive height unchanged")
	})

	t.Run("WithSizeClampsNegativeHeight", func(t *testing.T) {
		box := BoxModel{}.WithSize(100, -50)
		assert.Equal(t, 100, box.Width, "positive width unchanged")
		assert.Equal(t, 0, box.Height, "negative height clamped to 0")
	})

	t.Run("WithSizeClampsBothNegative", func(t *testing.T) {
		box := BoxModel{}.WithSize(-100, -50)
		assert.Equal(t, 0, box.Width)
		assert.Equal(t, 0, box.Height)
	})

	t.Run("WithSizePreservesZero", func(t *testing.T) {
		box := BoxModel{}.WithSize(0, 0)
		assert.Equal(t, 0, box.Width)
		assert.Equal(t, 0, box.Height)
	})

	t.Run("WithSizePreservesPositive", func(t *testing.T) {
		box := BoxModel{}.WithSize(100, 50)
		assert.Equal(t, 100, box.Width)
		assert.Equal(t, 50, box.Height)
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

		box := BoxModel{}.WithSize(availableWidth, 50)
		assert.Equal(t, 0, box.Width, "underflow clamped to 0")
		assert.Equal(t, 50, box.Height)
	})
}

func TestBoxModel_VirtualSizeEdgeCases(t *testing.T) {
	// VirtualSize semantics:
	// - VirtualSize = 0: "not set" - EffectiveVirtual returns ContentSize (no scrolling)
	// - VirtualSize > 0: explicit virtual size - used as-is
	// - VirtualSize < ContentSize: valid "scale-down" case (no scrolling, rare in TUI)
	// - VirtualSize > ContentSize: triggers scrolling

	t.Run("VirtualZeroDefaultsToContent", func(t *testing.T) {
		// VirtualSize = 0 is a sentinel meaning "use content size"
		box := BoxModel{
			Width:  100,
			Height: 50,
			// VirtualWidth and VirtualHeight are 0 (default)
		}

		// EffectiveVirtual returns ContentSize when Virtual is 0
		assert.Equal(t, 100, box.EffectiveVirtualWidth(), "defaults to ContentWidth")
		assert.Equal(t, 50, box.EffectiveVirtualHeight(), "defaults to ContentHeight")

		// Not scrollable because EffectiveVirtual == Content
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})

	t.Run("VirtualZeroWithPadding", func(t *testing.T) {
		// Verify VirtualSize=0 defaults to computed ContentSize (after padding)
		box := BoxModel{
			Width:   100,
			Height:  50,
			Padding: terma.EdgeInsetsAll(10), // ContentWidth = 80, ContentHeight = 30
		}

		// EffectiveVirtual returns the computed ContentSize
		assert.Equal(t, 80, box.ContentWidth())
		assert.Equal(t, 30, box.ContentHeight())
		assert.Equal(t, 80, box.EffectiveVirtualWidth(), "defaults to computed ContentWidth")
		assert.Equal(t, 30, box.EffectiveVirtualHeight(), "defaults to computed ContentHeight")
	})

	t.Run("VirtualEqualsContent", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 100, VirtualHeight: 50,
		}
		assert.False(t, box.IsScrollableX(), "equal virtual width should not be scrollable")
		assert.False(t, box.IsScrollableY(), "equal virtual height should not be scrollable")
		assert.Equal(t, 0, box.MaxScrollX())
		assert.Equal(t, 0, box.MaxScrollY())
	})

	t.Run("VirtualSmallerThanContent", func(t *testing.T) {
		// "Scale-down" case: Virtual is smaller than Content.
		// Rare in TUI, but valid. EffectiveVirtual returns the explicit value.
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 50, VirtualHeight: 25,
		}

		// EffectiveVirtual returns the explicit (smaller) value, not ContentSize
		assert.Equal(t, 50, box.EffectiveVirtualWidth(), "uses explicit value, not content")
		assert.Equal(t, 25, box.EffectiveVirtualHeight(), "uses explicit value, not content")

		// Not scrollable because Virtual < Content
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
		assert.Equal(t, 0, box.MaxScrollX(), "max scroll should be 0, not negative")
		assert.Equal(t, 0, box.MaxScrollY(), "max scroll should be 0, not negative")
	})

	t.Run("VirtualLargerThanContent", func(t *testing.T) {
		// Normal scrolling case: Virtual > Content
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, VirtualHeight: 150,
		}

		assert.Equal(t, 200, box.EffectiveVirtualWidth())
		assert.Equal(t, 150, box.EffectiveVirtualHeight())
		assert.True(t, box.IsScrollableX())
		assert.True(t, box.IsScrollableY())
		assert.Equal(t, 100, box.MaxScrollX()) // 200 - 100
		assert.Equal(t, 100, box.MaxScrollY()) // 150 - 50
	})

	t.Run("VirtualOneIsNotZero", func(t *testing.T) {
		// Edge case: VirtualSize = 1 is an explicit value, not the default
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 1, VirtualHeight: 1,
		}

		// Returns 1, not ContentSize
		assert.Equal(t, 1, box.EffectiveVirtualWidth())
		assert.Equal(t, 1, box.EffectiveVirtualHeight())

		// Not scrollable (Virtual < Content)
		assert.False(t, box.IsScrollableX())
		assert.False(t, box.IsScrollableY())
	})
}

func TestBoxModel_ScrollbarWidthEdgeCases(t *testing.T) {
	t.Run("ScrollbarEqualsContentWidth", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualHeight: 200, ScrollbarWidth: 100,
		}
		assert.Equal(t, 0, box.UsableContentBox().Width)
	})

	t.Run("ScrollbarExceedsContentWidth", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualHeight: 200, ScrollbarWidth: 150,
		}
		// Usable width is clamped to 0 when scrollbar exceeds content width
		assert.Equal(t, 0, box.UsableContentBox().Width)
	})

	t.Run("HorizontalScrollOnlyWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, ScrollbarWidth: 10,
		}
		// Scrollbar only applies to vertical scrolling
		assert.Equal(t, 100, box.UsableContentBox().Width)
	})

	t.Run("BothScrollableWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 10,
		}
		assert.Equal(t, 90, box.UsableContentBox().Width)
	})
}

func TestBoxModel_ScrollbarHeightEdgeCases(t *testing.T) {
	t.Run("ScrollbarEqualsContentHeight", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, ScrollbarHeight: 50,
		}
		assert.Equal(t, 0, box.UsableContentBox().Height)
	})

	t.Run("ScrollbarExceedsContentHeight", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, ScrollbarHeight: 75,
		}
		// Usable height is clamped to 0 when scrollbar exceeds content height
		assert.Equal(t, 0, box.UsableContentBox().Height)
	})

	t.Run("VerticalScrollOnlyWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualHeight: 200, ScrollbarHeight: 10,
		}
		// Horizontal scrollbar only applies to horizontal scrolling
		assert.Equal(t, 50, box.UsableContentBox().Height)
	})

	t.Run("BothScrollableWithScrollbar", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarHeight: 10,
		}
		assert.Equal(t, 40, box.UsableContentBox().Height)
	})
}

func TestBoxModel_BothScrollbarsEdgeCases(t *testing.T) {
	t.Run("BothScrollbarsReduceBothDimensions", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 10, ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 90, usable.Width, "vertical scrollbar reduces width")
		assert.Equal(t, 45, usable.Height, "horizontal scrollbar reduces height")
	})

	t.Run("BothScrollbarsExceedDimensions", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
			VirtualWidth: 200, VirtualHeight: 200,
			ScrollbarWidth: 150, ScrollbarHeight: 75,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 0, usable.Width, "clamped to 0")
		assert.Equal(t, 0, usable.Height, "clamped to 0")
	})

	t.Run("OnlyVerticalScrollable", func(t *testing.T) {
		box := BoxModel{
			Width: 100, Height: 50,
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
			Width: 100, Height: 50,
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
			Width: 100, Height: 50,
			// No virtual dimensions, so not scrollable
			ScrollbarWidth:  10,
			ScrollbarHeight: 5,
		}
		usable := box.UsableContentBox()
		assert.Equal(t, 100, usable.Width, "no scrollbar applied")
		assert.Equal(t, 50, usable.Height, "no scrollbar applied")
	})
}

func TestBoxModel_ZeroBorderBoxWithInsets(t *testing.T) {
	box := BoxModel{
		Width:   0,
		Height:  0,
		Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:  terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		Margin:  terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
	}

	t.Run("ContentClampedToZero", func(t *testing.T) {
		assert.Equal(t, 0, box.ContentWidth())
		assert.Equal(t, 0, box.ContentHeight())
	})

	t.Run("PaddingBoxClampedToZero", func(t *testing.T) {
		assert.Equal(t, 0, box.PaddingBoxWidth())
		assert.Equal(t, 0, box.PaddingBoxHeight())
	})

	t.Run("BorderBoxIsZero", func(t *testing.T) {
		assert.Equal(t, 0, box.BorderBoxWidth())
		assert.Equal(t, 0, box.BorderBoxHeight())
	})

	t.Run("MarginBoxIncludesMargin", func(t *testing.T) {
		assert.Equal(t, 20, box.MarginBoxWidth())  // 0 + 10 + 10
		assert.Equal(t, 20, box.MarginBoxHeight()) // 0 + 10 + 10
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
			Width:   110,
			Height:  60,
			Padding: terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		}
		// Content = 110 - 10 = 100, Height = 60 - 10 = 50
		assert.Equal(t, 100, box.ContentWidth())
		assert.Equal(t, 50, box.ContentHeight())
		assert.Equal(t, 110, box.MarginBoxWidth())
		assert.Equal(t, 60, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 5, rect.X)
		assert.Equal(t, 5, rect.Y)
	})

	t.Run("OnlyBorder", func(t *testing.T) {
		box := BoxModel{
			Width:  102,
			Height: 52,
			Border: terma.EdgeInsets{Top: 1, Right: 1, Bottom: 1, Left: 1},
		}
		// Content = 102 - 2 = 100, Height = 52 - 2 = 50
		assert.Equal(t, 100, box.ContentWidth())
		assert.Equal(t, 50, box.ContentHeight())
		assert.Equal(t, 102, box.MarginBoxWidth())
		assert.Equal(t, 52, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 1, rect.X)
		assert.Equal(t, 1, rect.Y)
	})

	t.Run("OnlyMargin", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
			Margin: terma.EdgeInsets{Top: 10, Right: 10, Bottom: 10, Left: 10},
		}
		// Content = 100 (no padding/border), MarginBox = 100 + 20 = 120
		assert.Equal(t, 100, box.ContentWidth())
		assert.Equal(t, 50, box.ContentHeight())
		assert.Equal(t, 120, box.MarginBoxWidth())
		assert.Equal(t, 70, box.MarginBoxHeight())
		rect := box.ContentBox()
		assert.Equal(t, 10, rect.X)
		assert.Equal(t, 10, rect.Y)
	})
}

func TestBoxModel_NegativeMargin(t *testing.T) {
	// CSS allows negative margins for overlapping and pull effects.
	// Negative margins shrink the margin box and shift origins.

	t.Run("NegativeMarginShrinksBox", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
			Margin: terma.EdgeInsets{Top: -10, Right: -10, Bottom: -10, Left: -10},
		}

		// MarginBox = Width + Margin.Left + Margin.Right
		// 100 + (-10) + (-10) = 80
		assert.Equal(t, 80, box.MarginBoxWidth())
		assert.Equal(t, 30, box.MarginBoxHeight()) // 50 + (-10) + (-10) = 30

		// MarginBox origin is at (0,0), but BorderBox shifts inward
		// BorderOrigin X = Margin.Left = -10 (pulls left)
		x, y := box.BorderOrigin()
		assert.Equal(t, -10, x)
		assert.Equal(t, -10, y)

		// MarginBox rect
		rect := box.MarginBox()
		assert.Equal(t, 0, rect.X)
		assert.Equal(t, 0, rect.Y)
		assert.Equal(t, 80, rect.Width)
		assert.Equal(t, 30, rect.Height)

		// BorderBox rect (shifted by negative margin)
		borderRect := box.BorderBox()
		assert.Equal(t, -10, borderRect.X)
		assert.Equal(t, -10, borderRect.Y)
		assert.Equal(t, 100, borderRect.Width)
		assert.Equal(t, 50, borderRect.Height)
	})

	t.Run("MixedPositiveNegativeMargins", func(t *testing.T) {
		box := BoxModel{
			Width:  100,
			Height: 50,
			Margin: terma.EdgeInsets{Top: 10, Right: -5, Bottom: 10, Left: 20},
		}

		// MarginBoxWidth = 100 + 20 + (-5) = 115
		// MarginBoxHeight = 50 + 10 + 10 = 70
		assert.Equal(t, 115, box.MarginBoxWidth())
		assert.Equal(t, 70, box.MarginBoxHeight())

		// BorderOrigin reflects actual margin values
		x, y := box.BorderOrigin()
		assert.Equal(t, 20, x)  // Margin.Left
		assert.Equal(t, 10, y)  // Margin.Top
	})

	t.Run("NegativeMarginWithBuilder", func(t *testing.T) {
		// Verify builder accepts negative margins without panicking
		assert.NotPanics(t, func() {
			BoxModel{}.WithMargin(terma.EdgeInsets{Left: -10, Top: -5})
		})

		box := BoxModel{}.
			WithSize(100, 50).
			WithMargin(terma.EdgeInsets{Left: -10, Right: -10})

		assert.Equal(t, 80, box.MarginBoxWidth())
	})
}

func TestBoxModel_ScrollingWithInsets(t *testing.T) {
	// When we have insets, content is computed from border-box
	box := BoxModel{
		Width:         120, // Border-box
		Height:        70,
		Padding:       terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		Border:        terma.EdgeInsets{Top: 5, Right: 5, Bottom: 5, Left: 5},
		VirtualWidth:  200,
		VirtualHeight: 150,
	}
	// Content = 120 - 10 - 10 = 100 for width
	// Content = 70 - 10 - 10 = 50 for height

	t.Run("ContentDimensions", func(t *testing.T) {
		assert.Equal(t, 100, box.ContentWidth())
		assert.Equal(t, 50, box.ContentHeight())
	})

	t.Run("ScrollableWithInsets", func(t *testing.T) {
		// Virtual 200 > Content 100
		assert.True(t, box.IsScrollableX())
		// Virtual 150 > Content 50
		assert.True(t, box.IsScrollableY())
	})

	t.Run("MaxScrollWithInsets", func(t *testing.T) {
		// MaxScrollX = 200 - 100 = 100
		assert.Equal(t, 100, box.MaxScrollX())
		// MaxScrollY = 150 - 50 = 100
		assert.Equal(t, 100, box.MaxScrollY())
	})
}
