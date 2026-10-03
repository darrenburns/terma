package terma

import (
	"testing"

	"github.com/darrenburns/terma/layout"
	"github.com/stretchr/testify/require"
)

func TestVisibleWhenPreservesChildOutput(t *testing.T) {
	child := Text{Content: "visible child"}
	expected := RenderToBuffer(child, 20, 4)
	actual := RenderToBuffer(VisibleWhen(true, child), 20, 4)
	require.Zero(t, CompareBuffers(expected, actual, 20, 4).MismatchedCells, "VisibleWhen(true) must render its child")
}

func TestVisibleWhenHiddenPreservesChildSize(t *testing.T) {
	child := Text{Content: "hidden child"}
	_, width, height := RenderToBufferWithSize(VisibleWhen(false, child), 20, 4)
	require.Equal(t, 12, width, "hidden child must reserve its width")
	require.Equal(t, 1, height, "hidden child must reserve its height")
}

type visibilityScene struct {
	visible Signal[bool]
	content Signal[string]
}

func (s *visibilityScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		VisibleWhen(s.visible.Get(), Column{Children: []Widget{
			reactivityBuilder{build: func(BuildContext) Widget { return Text{Content: s.content.Get()} }},
			Button{ID: "inside", Label: "Inside"},
		}}),
		Button{ID: "outside", Label: "Outside"},
	}}
}

func TestReactivityVisibilityPreservesLayoutAndInteraction(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 8, func() *visibilityScene {
		return &visibilityScene{visible: NewSignal(true), content: NewSignal("one")}
	})
	sequence.frame("visible", nil)
	require.Equal(t, "inside", sequence.actual.focusables[0].ID)
	outside := sequence.actual.renderer.WidgetByID("outside")
	require.NotNil(t, outside)
	originalY := outside.Bounds.Y
	sequence.frame("hidden", func(s *visibilityScene) { s.visible.Set(false) })
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Inside")
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "one")
	require.Len(t, sequence.actual.focusables, 1)
	require.Equal(t, "outside", sequence.actual.focusables[0].ID)
	require.Nil(t, sequence.actual.renderer.WidgetByID("inside"))
	require.Equal(t, originalY, sequence.actual.renderer.WidgetByID("outside").Bounds.Y)
	sequence.frame("hidden content grows", func(s *visibilityScene) { s.content.Set("one\ntwo\nthree") })
	require.Equal(t, originalY+2, sequence.actual.renderer.WidgetByID("outside").Bounds.Y)
	sequence.frame("visible again", func(s *visibilityScene) { s.visible.Set(true) })
	require.Len(t, sequence.actual.focusables, 2)
	require.NotNil(t, sequence.actual.renderer.WidgetByID("inside"))
	require.Contains(t, sequence.actual.renderer.ScreenText(), "one")
}

func TestVisibleWhenSnapshots(t *testing.T) {
	for _, visible := range []bool{false, true} {
		name := "TestVisibleWhenSnapshots_hidden"
		if visible {
			name = "TestVisibleWhenSnapshots_visible"
		}
		widget := Column{Children: []Widget{
			VisibleWhen(visible, Column{Style: Style{Padding: EdgeInsets{Left: 1, Right: 1}, Border: RoundedBorder(Cyan)}, Children: []Widget{
				Text{Content: "Reserved panel"},
				Button{ID: "inside", Label: "Inside"},
			}}),
			Text{Content: "Footer"},
		}}
		AssertSnapshotNamed(t, name, widget, 24, 8)
		AssertSnapshotNamed(t, name+"_text", Column{Children: []Widget{
			VisibleWhen(visible, Text{Content: "Text child"}),
			Text{Content: "Footer"},
		}}, 24, 4)
	}
}

type visibilityFloatScene struct{ visible Signal[bool] }

func (s *visibilityFloatScene) Build(BuildContext) Widget {
	return VisibleWhen(s.visible.Get(), Column{Children: []Widget{
		Text{Content: "inline"},
		Floating{Visible: true, Child: Button{ID: "floating", Label: "FLOAT"}},
	}})
}

func TestReactivityVisibilitySuppressesFloatingChildren(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 8, func() *visibilityFloatScene {
		return &visibilityFloatScene{visible: NewSignal(true)}
	})
	sequence.frame("visible float", nil)
	require.Contains(t, sequence.actual.renderer.ScreenText(), "FLOAT")
	sequence.frame("hidden float", func(s *visibilityFloatScene) { s.visible.Set(false) })
	require.Empty(t, sequence.actual.renderer.retainedFloats)
	require.Empty(t, sequence.actual.focusables)
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "FLOAT")
	sequence.frame("visible float again", func(s *visibilityFloatScene) { s.visible.Set(true) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "FLOAT")
	require.Len(t, sequence.actual.focusables, 1)
}

func TestVisibleWhenLegacyRenderTree(t *testing.T) {
	for _, visible := range []bool{false, true} {
		floats := NewFloatCollector()
		focus := NewFocusCollector()
		renderer, _ := geometryRenderer(t, 24, 8)
		ctx := NewBuildContext(NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil), floats)
		widget := VisibleWhen(visible, Column{Children: []Widget{
			Button{ID: "inside", Label: "Inside"},
			Floating{Visible: true, Child: Text{Content: "FLOAT"}},
		}})
		tree := BuildRenderTree(widget, ctx, layout.Loose(24, 8), focus)
		renderCtx := NewRenderContext(renderer.terminal, 24, 8, focus, renderer.focusManager, ctx, renderer.widgetRegistry)
		renderer.renderTree(renderCtx, tree, 0, 0)
		if visible {
			require.Contains(t, renderer.ScreenText(), "Inside")
			require.Len(t, focus.Focusables(), 1)
			require.NotEmpty(t, floats.Entries())
		} else {
			require.NotContains(t, renderer.ScreenText(), "Inside")
			require.Empty(t, focus.Focusables())
			require.Empty(t, floats.Entries())
			require.Nil(t, renderer.WidgetByID("inside"))
		}
	}
}

func TestVisibleWhenPreservesContainerDimensions(t *testing.T) {
	for _, width := range []Dimension{Auto, Cells(5), Flex(1), Percent(50)} {
		for _, makeParent := range []func(Widget) Widget{
			func(child Widget) Widget { return Row{Children: []Widget{child, Text{Content: "after"}}} },
			func(child Widget) Widget { return Column{Children: []Widget{child, Text{Content: "after"}}} },
		} {
			child := Text{Content: "content", Style: Style{Width: width, Padding: EdgeInsets{Left: 1, Right: 1}, Border: RoundedBorder(Cyan), Margin: EdgeInsets{Left: 1, Top: 1}}}
			expected := RenderToBuffer(makeParent(child), 24, 8)
			actual := RenderToBuffer(makeParent(VisibleWhen(true, child)), 24, 8)
			require.Zero(t, CompareBuffers(expected, actual, 24, 8).MismatchedCells, "wrapper must preserve child dimensions %v", width)
		}
	}
}

func TestReactivityInitiallyHiddenVisibility(t *testing.T) {
	sequence := newReactivitySequence(t, 24, 8, func() *visibilityScene {
		return &visibilityScene{visible: NewSignal(false), content: NewSignal("one")}
	})
	sequence.frame("initially hidden", nil)
	originalY := sequence.actual.renderer.WidgetByID("outside").Bounds.Y
	sequence.frame("hidden content grows", func(s *visibilityScene) { s.content.Set("one\ntwo\nthree") })
	require.Equal(t, originalY+2, sequence.actual.renderer.WidgetByID("outside").Bounds.Y)
	require.NotContains(t, sequence.actual.renderer.ScreenText(), "Inside")
	sequence.frame("revealed", func(s *visibilityScene) { s.visible.Set(true) })
	require.Contains(t, sequence.actual.renderer.ScreenText(), "three")
}

func TestVisibleWhenLegacyDoesNotDuplicateFloatingChildren(t *testing.T) {
	for _, child := range []Widget{
		Floating{Visible: true, Child: Text{Content: "FLOAT"}},
		Column{Children: []Widget{Floating{Visible: true, Child: Text{Content: "FLOAT"}}}},
	} {
		countFloats := func(widget Widget) int {
			floats := NewFloatCollector()
			ctx := NewBuildContext(NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil), floats)
			BuildRenderTree(widget, ctx, layout.Loose(24, 8), nil)
			return len(floats.Entries())
		}
		require.Equal(t, countFloats(child), countFloats(VisibleWhen(true, child)))
	}
}
