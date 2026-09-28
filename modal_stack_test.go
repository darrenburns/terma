package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type stackedModalsScene struct {
	topOpen Signal[bool]
}

func (s *stackedModalsScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		Button{ID: "base", Label: "Base"},
		Floating{
			Visible: true,
			Config:  FloatConfig{Modal: true, Position: FloatPositionCenter},
			Child:   Button{ID: "lower", Label: "Lower"},
		},
		Floating{
			Visible: s.topOpen.Get(),
			Config:  FloatConfig{Modal: true, Position: FloatPositionTopCenter},
			Child:   Button{ID: "upper", Label: "Upper"},
		},
	}}
}

// Two modals open at once must not take focus from each other: the one on
// top gets it, and the one beneath gets it back when the top one closes.
func TestTopmostModalTakesFocus(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 8, func() *stackedModalsScene {
		return &stackedModalsScene{topOpen: NewSignal(false)}
	})
	sequence.frame("Lower modal open", nil)
	require.Equal(t, "lower", sequence.actual.focus.FocusedID())

	sequence.frame("Upper modal opens on top", func(s *stackedModalsScene) { s.topOpen.Set(true) })
	require.Equal(t, "upper", sequence.actual.focus.FocusedID())

	sequence.frame("Upper modal closes", func(s *stackedModalsScene) { s.topOpen.Set(false) })
	require.Equal(t, "lower", sequence.actual.focus.FocusedID())
}
