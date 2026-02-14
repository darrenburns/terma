package main

import t "terma"

// FocusAwareSplitPane makes SplitPane focus opt-in for tab traversal.
type FocusAwareSplitPane struct {
	t.SplitPane
	AllowFocus     bool
	EnableKeybinds bool
}

func (s FocusAwareSplitPane) IsFocusable() bool {
	return s.AllowFocus && s.SplitPane.IsFocusable()
}

func (s FocusAwareSplitPane) Keybinds() []t.Keybind {
	if !s.EnableKeybinds {
		return nil
	}
	keybinds := s.SplitPane.Keybinds()
	for i := range keybinds {
		switch keybinds[i].Key {
		case "left", "h":
			keybinds[i].Name = "Shrink sidebar"
		case "right", "l":
			keybinds[i].Name = "Grow sidebar"
		}
	}
	return keybinds
}
