package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// keybindBarRoot provides root keybinds for a KeybindBar to show.
type keybindBarRoot struct{ child Widget }

func (r keybindBarRoot) Build(ctx BuildContext) Widget { return r.child }
func (r keybindBarRoot) Keybinds() []Keybind {
	return []Keybind{
		{Key: "ctrl+j", Name: "Send", Action: func() {}},
		{Key: "ctrl+t", Name: "Method", Action: func() {}},
		{Key: "ctrl+l", Name: "Focus URL", Action: func() {}},
		{Key: "f1", Name: "Help", Action: func() {}},
	}
}

// keybindBarSelectionScene offers Copy only while the focused input has a
// selection, as an editor's footer does.
type keybindBarSelectionScene struct{ input *TextInputState }

func (s *keybindBarSelectionScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		keybindBarCopyKeys{input: s.input},
		KeybindBar{},
	}}
}

type keybindBarCopyKeys struct{ input *TextInputState }

func (k keybindBarCopyKeys) Build(BuildContext) Widget {
	return TextInput{ID: "input", State: k.input}
}

func (k keybindBarCopyKeys) Keybinds() []Keybind {
	if !k.input.HasSelection() {
		return nil
	}
	return []Keybind{{Key: "ctrl+c", Name: "Copy", Action: func() {}}}
}

func TestReactivityKeybindBarFollowsSelection(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 2, func() *keybindBarSelectionScene {
		return &keybindBarSelectionScene{input: NewTextInputState("hello")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Selecting text shows Copy", func(s *keybindBarSelectionScene) { s.input.SelectAll() })
	sequence.frame("Clearing the selection hides Copy", func(s *keybindBarSelectionScene) { s.input.ClearSelection() })
}

func TestKeybindBar_DefaultsToFlexWidth(t *testing.T) {
	width, height := KeybindBar{}.GetContentDimensions()
	assert.Equal(t, Flex(1), width)
	assert.Equal(t, Cells(1), height)

	width, _ = KeybindBar{Width: Cells(12)}.GetContentDimensions()
	assert.Equal(t, Cells(12), width)
}

func TestSnapshot_KeybindBar_SharesRowWithSibling(t *testing.T) {
	widget := keybindBarRoot{child: Row{Children: []Widget{
		KeybindBar{},
		Text{Content: "Posting 3.0.0"},
	}}}
	AssertSnapshot(t, widget, 40, 1,
		"The bar takes the 27 cells the version text leaves and shows only the hints that fit whole ('ctrl+j Send ctrl+t Method'); 'Posting 3.0.0' is drawn intact at the right.")
}

func TestSnapshot_KeybindBar_ExplicitFlexWidthInRow(t *testing.T) {
	widget := keybindBarRoot{child: Row{Children: []Widget{
		KeybindBar{Width: Flex(1)},
		Text{Content: "Posting 3.0.0", Style: Style{Padding: EdgeInsets{Left: 2}}},
	}}}
	AssertSnapshot(t, widget, 30, 1,
		"With Width: Flex(1) the bar gets the 15 cells left of the padded version text: only 'ctrl+j Send' fits, and nothing runs into 'Posting 3.0.0'.")
}

func TestSnapshot_KeybindBar_AllHintsFit(t *testing.T) {
	widget := keybindBarRoot{child: KeybindBar{}}
	AssertSnapshot(t, widget, 60, 1, "All four hints fit in 60 cells.")
}

func TestSnapshot_KeybindBar_FixedWidth(t *testing.T) {
	widget := keybindBarRoot{child: Row{Children: []Widget{
		KeybindBar{Width: Cells(20), Style: Style{BackgroundColor: Hex("#333333")}},
		Text{Content: "|after"},
	}}}
	AssertSnapshot(t, widget, 40, 1,
		"A 20-cell bar (grey) holds only 'ctrl+j Send'; '|after' starts right at cell 20.")
}
