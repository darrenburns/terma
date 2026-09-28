package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// directPanel returns a focusable widget straight from Build, with no
// container in between.
type directPanel struct {
	child Widget
}

func (p directPanel) Build(BuildContext) Widget { return p.child }

func (p directPanel) Keybinds() []Keybind {
	return []Keybind{{Key: "ctrl+x", Name: "Panel action", Action: func() {}}}
}

func TestComponentReturningFocusableWidgetIsFocusable(t *testing.T) {
	cases := map[string]Widget{
		"list":       List[string]{ID: "direct", State: NewListState([]string{"a", "b"})},
		"text input": TextInput{ID: "direct", State: NewTextInputState("")},
		"text area":  TextArea{ID: "direct", State: NewTextAreaState("")},
		"split pane": SplitPane{ID: "direct", State: NewSplitPaneState(0.5), First: Text{Content: "a"}, Second: Text{Content: "b"}},
	}
	for name, child := range cases {
		t.Run(name, func(t *testing.T) {
			root := Column{Children: []Widget{
				Button{ID: "before", Label: "Before"},
				directPanel{child: child},
			}}
			focus := NewFocusManager()
			focus.SetRootWidget(root)
			renderer := NewRenderer(reactivityScreen{uv.NewBuffer(20, 5)}, 20, 5, focus, NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
			t.Cleanup(func() {
				if renderer.rootNode != nil {
					renderer.rootNode.dispose()
				}
			})

			focusables := renderer.Render(root)
			focus.SetFocusables(focusables)

			var ids []string
			for _, entry := range focusables {
				ids = append(ids, entry.ID)
			}
			require.Equal(t, []string{"before", "direct"}, ids)

			focus.FocusByID("direct")
			var names []string
			for _, kb := range focus.ActiveKeybinds() {
				names = append(names, kb.Name)
			}
			require.Contains(t, names, "Panel action", "the component's keybinds apply while its child is focused")
		})
	}
}
