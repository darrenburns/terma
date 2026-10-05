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
			p := NewPilot(t, hoverScreen{crumbs}, 30, 3)
			theme := getTheme()
			tinted := theme.Hover.BlendOver(theme.Background)
			p.MouseMove(2, 0)
			assert.Equal(t, tinted, bgAt(p, 2, 0))
			assert.Equal(t, theme.Background, bgAt(p, 9, 0))
			firstID := p.session.mouse.hover.currentID
			require.NotEmpty(t, firstID)
			p.ClickAt(2, 0)
			p.MouseMove(9, 0)
			assert.Equal(t, theme.Background, bgAt(p, 2, 0))
			assert.Equal(t, tinted, bgAt(p, 9, 0), "current segment retains its click policy and gains hover")
			assert.NotEqual(t, firstID, p.session.mouse.hover.currentID)
			p.ClickAt(9, 0)
			p.MouseMove(6, 0)
			p.ClickAt(6, 0)
			assert.Equal(t, theme.Background, bgAt(p, 6, 0), "separator is not tinted")
			assert.Equal(t, theme.Background, bgAt(p, 9, 0))
			assert.Equal(t, []int{0, 1}, selected, "separator is not clickable")
			p.MouseMove(25, 2)
			assert.Equal(t, theme.Background, bgAt(p, 2, 0))
		})
	}
}

func TestBreadcrumbHover_AnonymousPathsHaveDistinctStableIdentities(t *testing.T) {
	crumbs := Breadcrumbs{Path: []string{"Root", "Child"}, OnSelect: func(int) {}}
	p := NewPilot(t, hoverScreen{Column{Children: []Widget{crumbs, crumbs}}}, 30, 3)
	p.MouseMove(2, 0)
	firstID := p.session.mouse.hover.currentID
	p.settle()
	assert.Equal(t, firstID, p.session.renderer.WidgetAt(2, 0).ID)
	p.MouseMove(2, 1)
	assert.NotEqual(t, firstID, p.session.mouse.hover.currentID)
	assert.Equal(t, getTheme().Background, bgAt(p, 2, 0))
	assert.Equal(t, getTheme().Hover.BlendOver(getTheme().Background), bgAt(p, 2, 1))
}

func TestBreadcrumbHover_StaticPathDoesNotHighlight(t *testing.T) {
	p := NewPilot(t, hoverScreen{Breadcrumbs{Path: []string{"Root", "Child"}}}, 30, 3)
	p.MouseMove(2, 0)
	assert.Equal(t, getTheme().Background, bgAt(p, 2, 0))
	p.MouseMove(9, 0)
	assert.Equal(t, getTheme().Background, bgAt(p, 9, 0))
}

func TestBreadcrumbHover_DisabledPathDoesNotHighlightOrSelect(t *testing.T) {
	selected := false
	crumbs := Breadcrumbs{Path: []string{"Root"}, OnSelect: func(int) { selected = true }}
	p := NewPilot(t, hoverScreen{DisabledWhen(true, crumbs)}, 30, 3)
	p.MouseMove(2, 0)
	p.ClickAt(2, 0)
	assert.Equal(t, getTheme().Background, bgAt(p, 2, 0))
	assert.False(t, selected)
}

func TestBreadcrumbHover_TintsExplicitBackground(t *testing.T) {
	crumbs := Breadcrumbs{Path: []string{"Root"}, OnSelect: func(int) {}, Style: Style{BackgroundColor: getTheme().Surface}}
	p := NewPilot(t, hoverScreen{crumbs}, 30, 3)
	p.MouseMove(2, 0)
	assert.Equal(t, getTheme().Hover.BlendOver(getTheme().Surface), bgAt(p, 2, 0))
	p.MouseMove(25, 2)
	assert.Equal(t, getTheme().Surface, bgAt(p, 2, 0))
}
