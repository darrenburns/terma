package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splitPaneHoverPoint(t *testing.T, scene *clickScene, pane SplitPane) (int, int) {
	t.Helper()
	entry := scene.renderer.WidgetByID(pane.ID)
	require.NotNil(t, entry)
	cache := pane.State.layoutCache
	x, y := entry.Bounds.X+cache.contentOffsetX, entry.Bounds.Y+cache.contentOffsetY
	if pane.Orientation == SplitVertical {
		y += cache.dividerPos
	} else {
		x += cache.dividerPos
	}
	return x, y
}

func newSplitPaneHoverScene(t *testing.T, pane SplitPane) *clickScene {
	t.Helper()
	return newClickScene(t, hoverScreen{Column{Children: []Widget{
		Button{ID: "other-focus", Label: "Other"}, pane,
	}}}, 30, 12)
}

func TestSplitPaneHover_OnlyDivider(t *testing.T) {
	for _, orientation := range []SplitPaneOrientation{SplitHorizontal, SplitVertical} {
		t.Run(map[SplitPaneOrientation]string{SplitHorizontal: "horizontal", SplitVertical: "vertical"}[orientation], func(t *testing.T) {
			pane := SplitPane{
				ID: "split", State: NewSplitPaneState(0.5), Orientation: orientation,
				First: EmptyWidget{}, Second: EmptyWidget{}, DividerSize: 2,
				Style: Style{Width: Cells(20), Height: Cells(8), Padding: EdgeInsetsAll(1)},
			}
			scene := newSplitPaneHoverScene(t, pane)
			x, y := splitPaneHoverPoint(t, scene, pane)
			base := scene.bgAt(x, y)
			baseForeground := FromANSI(scene.buf.CellAt(x, y).Style.Fg)
			scene.hover(x, y)
			assert.True(t, pane.State.hovered.Peek())
			assert.Equal(t, base, scene.bgAt(x, y), "hover leaves the background unchanged")
			assert.Equal(t, getTheme().Hover.BlendOver(baseForeground), FromANSI(scene.buf.CellAt(x, y).Style.Fg))
			assert.False(t, pane.State.dragging.Peek(), "hover does not start dragging")
			assert.Equal(t, "other-focus", scene.focus.FocusedID())
			assert.Equal(t, 0.5, pane.State.GetPosition())

			entry := scene.renderer.WidgetByID(pane.ID)
			paddingX, paddingY := x, entry.Bounds.Y
			if orientation == SplitVertical {
				paddingX, paddingY = entry.Bounds.X, y
			}
			scene.hover(paddingX, paddingY)
			assert.False(t, pane.State.hovered.Peek(), "padding along the divider's axis is not the divider")
			assert.Equal(t, base, scene.bgAt(x, y))
			assert.Equal(t, baseForeground, FromANSI(scene.buf.CellAt(x, y).Style.Fg))
			scene.hover(entry.Bounds.X+pane.State.layoutCache.contentOffsetX, entry.Bounds.Y+pane.State.layoutCache.contentOffsetY)
			assert.False(t, pane.State.hovered.Peek(), "pane content has no divider hover")

			scene.hover(x, y)
			pane.State.SetPosition(0.8)
			scene.draw()
			scene.router.reconcileHover()
			scene.draw()
			assert.False(t, pane.State.hovered.Peek(), "moving the divider away clears stationary-pointer hover")

			x, y = splitPaneHoverPoint(t, scene, pane)
			toX, toY := x+2, y
			if orientation == SplitVertical {
				toX, toY = x, y-1
			}
			before := pane.State.GetPosition()
			scene.drag(x, y, toX, toY)
			assert.NotEqual(t, before, pane.State.GetPosition(), "existing mouse capture still resizes")
			assert.False(t, pane.State.dragging.Peek(), "release ends dragging")
		})
	}
}

func TestSplitPaneHover_PreservesGradientAndFocusColors(t *testing.T) {
	background := NewGradient(RGB(20, 40, 60), RGB(70, 90, 110)).WithAngle(90)
	focusBackground := NewGradient(RGB(60, 30, 80), RGB(120, 70, 140)).WithAngle(90)
	foreground := NewGradient(RGB(100, 120, 140), RGB(150, 170, 190)).WithAngle(90)
	focusForeground := NewGradient(RGB(180, 200, 220), RGB(210, 220, 230)).WithAngle(90)
	pane := SplitPane{
		ID: "split", State: NewSplitPaneState(0.5), First: EmptyWidget{}, Second: EmptyWidget{},
		Style:             Style{Width: Cells(20), Height: Cells(8)},
		DividerBackground: background, DividerForeground: foreground,
		DividerFocusBackground: focusBackground, DividerFocusForeground: focusForeground,
	}
	scene := newSplitPaneHoverScene(t, pane)
	x, y := splitPaneHoverPoint(t, scene, pane)
	scene.hover(x, y)
	cache := pane.State.layoutCache
	for localY := 0; localY < cache.contentHeight; localY++ {
		assert.Equal(t, background.ColorAt(cache.contentWidth, cache.contentHeight, x, localY), scene.bgAt(x, y+localY))
		assert.Equal(t, getTheme().Hover.BlendOver(foreground.ColorAt(cache.contentWidth, cache.contentHeight, x, localY)), FromANSI(scene.buf.CellAt(x, y+localY).Style.Fg))
	}
	scene.focus.FocusByID(pane.ID)
	scene.draw()
	assert.Equal(t, focusBackground.ColorAt(cache.contentWidth, cache.contentHeight, x, 0), scene.bgAt(x, y))
	assert.Equal(t, getTheme().Hover.BlendOver(focusForeground.ColorAt(cache.contentWidth, cache.contentHeight, x, 0)), FromANSI(scene.buf.CellAt(x, y).Style.Fg))
	scene.hover(29, 11)
	assert.Equal(t, focusBackground.ColorAt(cache.contentWidth, cache.contentHeight, x, 0), scene.bgAt(x, y), "hover leave retains focus styling")
	assert.Equal(t, focusForeground.ColorAt(cache.contentWidth, cache.contentHeight, x, 0), FromANSI(scene.buf.CellAt(x, y).Style.Fg))
	scene.focus.FocusByID("other-focus")
	pane.State.setDragging(true)
	scene.draw()
	assert.Equal(t, focusForeground.ColorAt(cache.contentWidth, cache.contentHeight, x, 0), FromANSI(scene.buf.CellAt(x, y).Style.Fg), "dragging still uses focus colors")
}

func TestSplitPaneHover_DisabledDoesNotTint(t *testing.T) {
	pane := SplitPane{ID: "split", State: NewSplitPaneState(0.5), First: EmptyWidget{}, Second: EmptyWidget{}, Style: Style{Width: Cells(20), Height: Cells(8)}}
	scene := newClickScene(t, hoverScreen{DisabledWhen(true, pane)}, 30, 12)
	x, y := splitPaneHoverPoint(t, scene, pane)
	base := scene.bgAt(x, y)
	baseForeground := FromANSI(scene.buf.CellAt(x, y).Style.Fg)
	scene.hover(x, y)
	assert.False(t, pane.State.hovered.Peek())
	assert.Equal(t, base, scene.bgAt(x, y))
	assert.Equal(t, baseForeground, FromANSI(scene.buf.CellAt(x, y).Style.Fg))
}
