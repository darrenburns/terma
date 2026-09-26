package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type reactivitySizedListScene struct {
	state  *ListState[string]
	height func(bool) Dimension
}

func (s *reactivitySizedListScene) Build(BuildContext) Widget {
	return List[string]{
		ID: "sized-list", State: s.state, Width: Cells(20), Height: Cells(10),
		RenderItem: func(item string, active, selected bool) Widget {
			return Text{Content: item, Style: Style{Height: s.height(active)}}
		},
	}
}

// Full and incremental rendering can agree while both lose the custom row's
// declared size. Check the actual layout as well as comparing render paths.
func TestReactivityCustomListRowDimensions(t *testing.T) {
	for _, tc := range []struct {
		name   string
		height func(bool) Dimension
		first  []listItemLayout
		moved  []listItemLayout
	}{
		{"flex", func(bool) Dimension { return Flex(1) }, []listItemLayout{{0, 5}, {5, 5}}, []listItemLayout{{0, 5}, {5, 5}}},
		{"percent", func(bool) Dimension { return Percent(50) }, []listItemLayout{{0, 5}, {5, 5}}, []listItemLayout{{0, 5}, {5, 5}}},
		{"cursor-dependent", func(active bool) Dimension {
			if active {
				return Percent(70)
			}
			return Percent(30)
		}, []listItemLayout{{0, 7}, {7, 3}}, []listItemLayout{{0, 3}, {3, 7}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sequence := newReactivitySequence(t, 20, 10, func() *reactivitySizedListScene {
				return &reactivitySizedListScene{state: NewListState([]string{"one", "two"}), height: tc.height}
			})
			sequence.frame("Initial", nil)
			require.Equal(t, tc.first, sequence.actual.root.state.itemLayouts)
			work := sequence.frame("Move cursor", func(s *reactivitySizedListScene) { s.state.SelectNext() })
			require.Equal(t, tc.moved, sequence.actual.root.state.itemLayouts)
			require.LessOrEqual(t, work.BuildCount, 4, "only the two affected rows and their children rebuild")
		})
	}
}

type reactivityBackdropScene struct {
	color  Signal[Color]
	nested bool
}

func (s *reactivityBackdropScene) Build(BuildContext) Widget {
	owner := reactivityBuilder{ID: "backdrop-owner", build: func(BuildContext) Widget {
		return Floating{
			Visible: true, Config: FloatConfig{Modal: true, BackdropColor: s.color.Get()},
			Child: Text{Content: "modal"},
		}
	}}
	var overlay Widget = owner
	if s.nested {
		overlay = Floating{Visible: true, Child: owner}
	}
	return Column{Children: []Widget{
		Text{Content: "background", Style: Style{BackgroundColor: RGB(255, 255, 255), Width: Cells(20), Height: Cells(8)}},
		overlay,
	}}
}

func TestReactivityModalBackdropColor(t *testing.T) {
	for _, nested := range []bool{false, true} {
		name := "direct"
		if nested {
			name = "nested"
		}
		t.Run(name, func(t *testing.T) {
			sequence := newReactivitySequence(t, 20, 10, func() *reactivityBackdropScene {
				return &reactivityBackdropScene{color: NewSignal(RGBA(0, 0, 0, 0.5)), nested: nested}
			})
			sequence.frame("Initial black backdrop", nil)
			sequence.frame("Change to red backdrop", func(s *reactivityBackdropScene) { s.color.Set(RGBA(255, 0, 0, 0.5)) })
			sequence.frame("Remove backdrop opacity", func(s *reactivityBackdropScene) { s.color.Set(RGBA(255, 0, 0, 0)) })
			sequence.frame("Use theme backdrop", func(s *reactivityBackdropScene) { s.color.Set(Color{}) })
		})
	}
}
