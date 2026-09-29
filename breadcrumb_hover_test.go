package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBreadcrumbHover_ClickableSegmentsAndSeparators(t *testing.T) {
	for _, id := range []string{"path", ""} {
		t.Run(id, func(t *testing.T) {
			var selected []int
			crumbs := Breadcrumbs{ID: id, Path: []string{"Root", "Child"}, OnSelect: func(i int) { selected = append(selected, i) }}
			scene := newClickScene(t, hoverScreen{crumbs}, 30, 3)
			theme := getTheme()
			tinted := theme.Hover.BlendOver(theme.Background)
			scene.hover(2, 0)
			assert.Equal(t, tinted, scene.bgAt(2, 0))
			assert.Equal(t, theme.Background, scene.bgAt(9, 0))
			firstID := scene.router.hover.currentID
			require.NotEmpty(t, firstID)
			scene.click(2, 0, 0)
			scene.hover(9, 0)
			assert.Equal(t, theme.Background, scene.bgAt(2, 0))
			assert.Equal(t, tinted, scene.bgAt(9, 0), "current segment retains its click policy and gains hover")
			assert.NotEqual(t, firstID, scene.router.hover.currentID)
			scene.click(9, 0, 0)
			scene.hover(6, 0)
			scene.click(6, 0, 0)
			assert.Equal(t, theme.Background, scene.bgAt(6, 0), "separator is not tinted")
			assert.Equal(t, theme.Background, scene.bgAt(9, 0))
			assert.Equal(t, []int{0, 1}, selected, "separator is not clickable")
			scene.hover(25, 2)
			assert.Equal(t, theme.Background, scene.bgAt(2, 0))
		})
	}
}

func TestBreadcrumbHover_AnonymousPathsHaveDistinctStableIdentities(t *testing.T) {
	crumbs := Breadcrumbs{Path: []string{"Root", "Child"}, OnSelect: func(int) {}}
	scene := newClickScene(t, hoverScreen{Column{Children: []Widget{crumbs, crumbs}}}, 30, 3)
	scene.hover(2, 0)
	firstID := scene.router.hover.currentID
	scene.draw()
	assert.Equal(t, firstID, scene.renderer.WidgetAt(2, 0).ID)
	scene.hover(2, 1)
	assert.NotEqual(t, firstID, scene.router.hover.currentID)
	assert.Equal(t, getTheme().Background, scene.bgAt(2, 0))
	assert.Equal(t, getTheme().Hover.BlendOver(getTheme().Background), scene.bgAt(2, 1))
}

func TestBreadcrumbHover_StaticPathDoesNotHighlight(t *testing.T) {
	scene := newClickScene(t, hoverScreen{Breadcrumbs{Path: []string{"Root", "Child"}}}, 30, 3)
	scene.hover(2, 0)
	assert.Equal(t, getTheme().Background, scene.bgAt(2, 0))
	scene.hover(9, 0)
	assert.Equal(t, getTheme().Background, scene.bgAt(9, 0))
}

func TestBreadcrumbHover_DisabledPathDoesNotHighlightOrSelect(t *testing.T) {
	selected := false
	crumbs := Breadcrumbs{Path: []string{"Root"}, OnSelect: func(int) { selected = true }}
	scene := newClickScene(t, hoverScreen{DisabledWhen(true, crumbs)}, 30, 3)
	scene.hover(2, 0)
	scene.click(2, 0, 0)
	assert.Equal(t, getTheme().Background, scene.bgAt(2, 0))
	assert.False(t, selected)
}

func TestBreadcrumbHover_TintsExplicitBackground(t *testing.T) {
	crumbs := Breadcrumbs{Path: []string{"Root"}, OnSelect: func(int) {}, Style: Style{BackgroundColor: getTheme().Surface}}
	scene := newClickScene(t, hoverScreen{crumbs}, 30, 3)
	scene.hover(2, 0)
	assert.Equal(t, getTheme().Hover.BlendOver(getTheme().Surface), scene.bgAt(2, 0))
	scene.hover(25, 2)
	assert.Equal(t, getTheme().Surface, scene.bgAt(2, 0))
}
