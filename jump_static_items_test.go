package terma

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// jumpStaticTabsScene gives some tabs static keys, with a button beside the
// tab bar so there is something else for dynamic hints to label.
type jumpStaticTabsScene struct {
	jump *JumpState
	tabs *TabState
	keys map[string]string // tab key -> static jump key
}

func newJumpStaticTabsScene(keys map[string]string) func() *jumpStaticTabsScene {
	return func() *jumpStaticTabsScene {
		return &jumpStaticTabsScene{
			jump: NewJumpState(),
			tabs: NewTabState([]Tab{{Key: "one", Label: "One"}, {Key: "two", Label: "Two"}, {Key: "three", Label: "Three"}}),
			keys: keys,
		}
	}
}

func (s *jumpStaticTabsScene) Build(ctx BuildContext) Widget {
	bar := TabBar{ID: "tabs", State: s.tabs}
	var targets []JumpTarget
	for _, tab := range []string{"one", "two", "three"} {
		if key, ok := s.keys[tab]; ok {
			targets = append(targets, JumpTarget{Key: key, ID: bar.TabID(tab)})
		}
	}
	return Jumper{
		State:   s.jump,
		Targets: targets,
		Dynamic: true,
		Child:   Column{Children: []Widget{bar, Button{ID: "save", Label: "Save"}}},
	}
}

// A tab bar whose tabs carry static keys gets no hint of its own: it would
// sit on its first tab, covering that tab's label.
func TestJumpStaticKeysOnTabsCoverTheirTabBar(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 4, newJumpStaticTabsScene(map[string]string{"one": "1", "two": "2", "three": "3"}))
	scene := sequence.actual.root
	sequence.frame("Initial", nil)
	sequence.press("Jump mode", "ctrl+o")

	var keys []string
	for _, label := range scene.jump.labels {
		keys = append(keys, label.key+"→"+label.focusID)
	}
	require.ElementsMatch(t, []string{"1→tabs", "2→tabs", "3→tabs", "a→save"}, keys)

	sequence.press("Jump to the second tab", "2")
	require.Equal(t, "two", scene.tabs.ActiveKeyPeek())
	require.Equal(t, "tabs", sequence.actual.focus.FocusedID())
}

// With only some tabs keyed, the rest get dynamic hints, and the tab bar
// still gets none of its own.
func TestJumpSomeStaticTabsLeaveTheRestDynamic(t *testing.T) {
	sequence := newReactivitySequence(t, 40, 4, newJumpStaticTabsScene(map[string]string{"one": "1"}))
	scene := sequence.actual.root
	sequence.frame("Initial", nil)
	sequence.press("Jump mode", "ctrl+o")

	tabLabels := 0
	for _, label := range scene.jump.labels {
		if label.focusID == "tabs" {
			require.NotNil(t, label.action, "the tab bar itself is not a target")
			tabLabels++
		}
	}
	require.Equal(t, 3, tabLabels, "one static and two dynamic tab hints")
}

func TestJumpStaticTabsSnapshot(t *testing.T) {
	scene := newJumpStaticTabsScene(map[string]string{"one": "1", "two": "2", "three": "3"})()
	scene.jump.Activate()
	AssertSnapshot(t, scene, 40, 4,
		"Static keys on tabs: '1', '2' and '3' sit on the One, Two and Three tabs, and 'a' on the Save button. The tab bar gets no hint of its own, so nothing covers the first tab.")
}
