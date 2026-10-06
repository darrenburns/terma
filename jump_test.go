package terma

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJumpHints(t *testing.T) {
	// Enough characters: one each, in alphabet order.
	require.Equal(t, []string{"a", "s", "d"}, jumpHints(3, "asdf", nil))

	// Too few: the first hints are extended, so shorter hints come first and
	// none is a prefix of another.
	require.Equal(t, []string{"s", "d", "f", "aa", "as", "ad"}, jumpHints(6, "asdf", nil))
	hints := jumpHints(40, "asdf", nil)
	require.Len(t, hints, 40)
	requirePrefixFree(t, hints)
	for i := 1; i < len(hints); i++ {
		require.LessOrEqual(t, len(hints[i-1]), len(hints[i]), "shorter hints first")
	}

	// Static keys are never hints, and never start or are started by one.
	require.Equal(t, []string{"s", "d", "f"}, jumpHints(3, "asdf", []string{"a"}))
	hints = jumpHints(5, "gas", []string{"gt", "ga"})
	require.Equal(t, []string{"s", "gg", "gs", "ag", "aa"}, hints)
	requirePrefixFree(t, append(hints, "gt", "ga"))

	// Repeated characters count once; one character can make only one hint.
	require.Equal(t, []string{"a", "s"}, jumpHints(2, "aass", nil))
	require.Equal(t, []string{"a"}, jumpHints(5, "a", nil))
	require.Nil(t, jumpHints(0, "asdf", nil))
}

func requirePrefixFree(t *testing.T, keys []string) {
	t.Helper()
	for i, a := range keys {
		for j, b := range keys {
			if i != j {
				require.False(t, strings.HasPrefix(b, a), "%q is a prefix of %q", a, b)
			}
		}
	}
}

// jumpScene is a small mail app: a sidebar of folders, a compose form and a
// preview line. "1" jumps to the sidebar (a container), "n" to the name
// field, and "p" runs an action.
type jumpScene struct {
	jump      *JumpState
	dynamic   bool
	hints     string
	name      *TextInputState
	previewed Signal[int]
	dialog    Signal[bool]
	unmatched []string
}

func newJumpScene(dynamic bool, hints string) func() *jumpScene {
	return func() *jumpScene {
		return &jumpScene{
			jump:      NewJumpState(),
			dynamic:   dynamic,
			hints:     hints,
			name:      NewTextInputState(""),
			previewed: NewSignal(0),
			dialog:    NewSignal(false),
		}
	}
}

func (s *jumpScene) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	panel := Style{Border: RoundedBorder(theme.Border), Padding: EdgeInsetsXY(1, 0)}
	return Jumper{
		State: s.jump,
		Targets: []JumpTarget{
			{Key: "1", ID: "folders"},
			{Key: "n", ID: "name"},
			{Key: "p", ID: "preview", Action: func() { s.previewed.Update(func(n int) int { return n + 1 }) }},
		},
		Dynamic:   s.dynamic,
		Hints:     s.hints,
		Unmatched: func(event KeyEvent) { s.unmatched = append(s.unmatched, event.Key()) },
		Child: Row{
			Spacing: 1,
			Children: []Widget{
				Column{ID: "folders", Style: panel, Children: []Widget{
					Button{ID: "inbox", Label: "Inbox"},
					Button{ID: "drafts", Label: "Drafts"},
					Button{ID: "sent", Label: "Sent"},
				}},
				Column{Width: Flex(1), Style: panel, Spacing: 1, Children: []Widget{
					TextInput{ID: "name", State: s.name, Width: Cells(20), Placeholder: "Name"},
					Row{Spacing: 1, Children: []Widget{
						Button{ID: "save", Label: "Save"},
						Button{ID: "discard", Label: "Discard"},
					}},
					Text{ID: "preview", Content: "Preview: 0 views"},
				}},
				Dialog{
					ID:      "confirm",
					Visible: s.dialog.Get(),
					Title:   "Discard draft?",
					Content: Text{Content: "This can't be undone."},
					Buttons: []Button{{ID: "keep", Label: "Keep"}, {ID: "drop", Label: "Discard"}},
				},
			},
		},
	}
}

// press sends keys to both sides as the app would, checking the frame drawn
// after each.
func (s *reactivitySequence[T]) press(name string, keys ...string) {
	s.t.Helper()
	for _, key := range keys {
		ev, err := keyPress(key)
		require.NoError(s.t, err)
		for _, side := range []*reactivitySurface[T]{s.actual, s.expected} {
			dispatchKey(side.renderer, side.focus, side.root, KeyEvent{event: ev})
			// The app applies focus requests after its next render; the sides
			// share the request, so apply it to each as it's made.
			if pendingFocusID != "" {
				side.focus.FocusByID(pendingFocusID)
				side.focused.Set(side.focus.Focused())
				pendingFocusID = ""
			}
		}
		s.frame(name+": "+key, nil)
	}
}

func jumpLabelKeys(state *JumpState) map[string]string {
	keys := make(map[string]string)
	for _, label := range state.labels {
		keys[label.focusID] = label.key
	}
	return keys
}

func TestJumpStaticTargets(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 12, newJumpScene(false, ""))
	scene := sequence.actual.root
	sequence.frame("Initial", nil)
	require.Equal(t, "inbox", sequence.actual.focus.FocusedID())

	sequence.press("Jump mode shows static labels", "ctrl+o")
	require.True(t, scene.jump.active.Peek())
	require.Equal(t, map[string]string{"inbox": "1", "name": "n", "": "p"}, jumpLabelKeys(scene.jump))

	sequence.press("Jump to the name field", "n")
	require.False(t, scene.jump.active.Peek())
	require.Equal(t, "name", sequence.actual.focus.FocusedID())
	require.Equal(t, "", scene.name.GetText(), "the key is captured, not typed")

	// A container target focuses the first focusable inside it.
	sequence.press("Jump to the folders panel", "ctrl+o", "1")
	require.Equal(t, "inbox", sequence.actual.focus.FocusedID())

	// An action runs instead of moving focus.
	sequence.press("Jump runs the preview action", "ctrl+o", "p")
	require.Equal(t, 1, scene.previewed.Peek())
	require.Equal(t, "inbox", sequence.actual.focus.FocusedID())
}

func TestJumpLeavingWithoutJumping(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 12, newJumpScene(false, ""))
	scene := sequence.actual.root
	sequence.frame("Initial", nil)

	for _, key := range []string{"escape", "ctrl+o", "x", "down"} {
		sequence.press("Enter jump mode", "ctrl+o")
		require.True(t, scene.jump.active.Peek())
		sequence.press("Leave with "+key, key)
		require.False(t, scene.jump.active.Peek(), key)
		require.Equal(t, "inbox", sequence.actual.focus.FocusedID(), key)
	}
	// Keys that match no label reach Unmatched; escape and the toggle key
	// only leave.
	require.Equal(t, []string{"x", "down"}, scene.unmatched)

	// A click anywhere lands on the overlay and leaves jump mode.
	sequence.press("Enter jump mode", "ctrl+o")
	for _, side := range []*reactivitySurface[*jumpScene]{sequence.actual, sequence.expected} {
		entry := side.renderer.WidgetAt(30, 5)
		require.NotNil(t, entry)
		require.Equal(t, jumpOverlayID, entry.ID)
		entry.EventWidget.(Clickable).OnClick(MouseEvent{})
	}
	sequence.frame("Click leaves jump mode", nil)
	require.False(t, scene.jump.active.Peek())
}

func TestJumpDynamicHintsNarrowAsTyped(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 12, newJumpScene(true, "as"))
	scene := sequence.actual.root
	sequence.frame("Initial", nil)

	sequence.press("Jump mode labels every focusable", "ctrl+o")
	keys := jumpLabelKeys(scene.jump)
	// Static targets keep their keys; the rest get hints in reading order.
	// The folders panel's jump already reaches the inbox. Two characters can
	// only tell four targets apart with two-character hints.
	require.Equal(t, map[string]string{
		"inbox": "1", "name": "n", "": "p",
		"drafts": "aa", "sent": "as", "save": "sa", "discard": "ss",
	}, keys)

	sequence.press("Typing the first character narrows the labels", "a")
	require.True(t, scene.jump.active.Peek())
	require.Equal(t, "a", scene.jump.typed.Peek())

	sequence.press("Backspace widens them again", "backspace")
	require.Equal(t, "", scene.jump.typed.Peek())

	sequence.press("Typing a full hint jumps", "a", "s")
	require.False(t, scene.jump.active.Peek())
	require.Equal(t, "sent", sequence.actual.focus.FocusedID())
}

func TestJumpStaysInsideModal(t *testing.T) {
	sequence := newReactivitySequence(t, 60, 12, newJumpScene(true, ""))
	scene := sequence.actual.root
	sequence.frame("Dialog opens", func(s *jumpScene) { s.dialog.Set(true) })
	require.Equal(t, "keep", sequence.actual.focus.FocusedID())

	// The overlay goes above the dialog, even though the dialog registered
	// its overlay later, and only the dialog's buttons are labelled.
	sequence.press("Jump mode inside the dialog", "ctrl+o")
	require.Equal(t, map[string]string{"keep": "a", "drop": "s"}, jumpLabelKeys(scene.jump))
	require.True(t, sequence.actual.renderer.TopFloat().topmost)

	sequence.press("Jump to the dialog's other button", "s")
	require.Equal(t, "drop", sequence.actual.focus.FocusedID())
}

func TestJumpSnapshots(t *testing.T) {
	scene := func(dynamic bool, hints string, setup func(*jumpScene)) *jumpScene {
		s := newJumpScene(dynamic, hints)()
		s.jump.Activate()
		if setup != nil {
			setup(s)
		}
		return s
	}
	AssertSnapshotNamed(t, "TestJumpSnapshots_static", scene(false, "", nil), 60, 12,
		"Static jump map: labels '1' on the Folders panel's top-left corner, 'n' on the name field and 'p' on the preview line. The screen beneath is lightly dimmed, and Inbox keeps its focus styling.")
	AssertSnapshotNamed(t, "TestJumpSnapshots_dynamic", scene(true, "", nil), 60, 12,
		"Static and dynamic together: '1', 'n' and 'p' as before, plus one-letter hints from the home row on Drafts, Sent, Save and Discard, assigned in reading order.")
	AssertSnapshotNamed(t, "TestJumpSnapshots_narrowed", scene(true, "as", func(s *jumpScene) { s.jump.typed.Set("a") }), 60, 12,
		"Two-character hints after typing 'a': only the labels starting with 'a' remain (on Drafts and Sent), with the typed 'a' faded and the character still to type in bold.")
	AssertSnapshotNamed(t, "TestJumpSnapshots_modal", scene(true, "", func(s *jumpScene) { s.dialog.Set(true) }), 60, 12,
		"With a dialog open, only the dialog's Keep ('a') and Discard ('s') buttons are labelled, and the labels are drawn above the dialog.")
}
