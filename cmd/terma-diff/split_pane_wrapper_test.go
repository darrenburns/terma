package main

import (
	"testing"

	t "terma"

	"github.com/stretchr/testify/require"
)

func TestFocusAwareSplitPane_KeybindNamesUseSidebarTerminology(tt *testing.T) {
	widget := FocusAwareSplitPane{
		SplitPane: t.SplitPane{
			State:       t.NewSplitPaneState(0.30),
			Orientation: t.SplitHorizontal,
		},
		EnableKeybinds: true,
	}

	keybinds := widget.Keybinds()
	require.NotEmpty(tt, keybinds)

	namesByKey := map[string]string{}
	for _, keybind := range keybinds {
		namesByKey[keybind.Key] = keybind.Name
	}

	require.Equal(tt, "Shrink sidebar", namesByKey["left"])
	require.Equal(tt, "Shrink sidebar", namesByKey["h"])
	require.Equal(tt, "Grow sidebar", namesByKey["right"])
	require.Equal(tt, "Grow sidebar", namesByKey["l"])
}
