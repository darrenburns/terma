package terma

import (
	"github.com/stretchr/testify/require"
	"testing"
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
		name := "hidden"
		if visible {
			name = "visible"
		}
		widget := Column{Children: []Widget{
			VisibleWhen(visible, Column{Style: Style{Padding: EdgeInsets{Left: 1, Right: 1}, Border: RoundedBorder(Cyan)}, Children: []Widget{
				Text{Content: "Reserved panel"},
				Button{ID: "inside", Label: "Inside"},
			}}),
			Text{Content: "Footer"},
		}}
		AssertSnapshotNamed(t, name, widget, 24, 8)
	}
}
