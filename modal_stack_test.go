package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
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

type modalFocusRequestScene struct {
	open    Signal[bool]
	editing Signal[bool]
}

func (s *modalFocusRequestScene) Build(BuildContext) Widget {
	children := []Widget{Button{ID: "first", Label: "First"}, Button{ID: "second", Label: "Second"}}
	if s.editing.Get() {
		children = append(children, Button{ID: "editor", Label: "Editor"})
	}
	return Column{Children: []Widget{
		Button{ID: "base", Label: "Base"},
		Floating{
			Visible: s.open.Get(),
			Config:  FloatConfig{Modal: true, Position: FloatPositionCenter},
			Child:   Column{Children: children},
		},
	}}
}

// A focus request for a widget inside a modal is kept rather than replaced
// by the modal's first focusable: both when the request opens the modal and
// when it moves focus off a widget that is being removed from the modal.
func TestModalKeepsPendingFocusRequestInside(t *testing.T) {
	oldPending := pendingFocusID
	defer func() { pendingFocusID = oldPending }()

	scene := &modalFocusRequestScene{open: NewSignal(false), editing: NewSignal(true)}
	buffer := uv.NewBuffer(30, 8)
	focus := NewFocusManager()
	focus.SetRootWidget(scene)
	focused := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(reactivityScreen{buffer}, 30, 8, focus, focused, NewAnySignal[Widget](nil))
	// Render as the app loop does: requests made before a frame are applied
	// after it.
	frame := func() {
		focus.SetFocusables(renderer.Update(scene))
		if pendingFocusID != "" {
			focus.FocusByID(pendingFocusID)
			pendingFocusID = ""
		}
		focused.Set(focus.Focused())
	}
	pendingFocusID = ""
	frame()
	require.Equal(t, "base", focus.FocusedID())

	scene.open.Set(true)
	RequestFocus("second")
	frame()
	require.Equal(t, "second", focus.FocusedID())

	RequestFocus("editor")
	frame()
	require.Equal(t, "editor", focus.FocusedID())

	scene.editing.Set(false)
	RequestFocus("second")
	frame()
	require.Equal(t, "second", focus.FocusedID())

	// Without a request, a modal still pulls focus inside when it opens.
	scene.open.Set(false)
	RequestFocus("base")
	frame()
	scene.open.Set(true)
	frame()
	require.Equal(t, "first", focus.FocusedID())
}
