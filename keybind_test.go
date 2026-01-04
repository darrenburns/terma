package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
)

// Helper to create KeyEvent for testing
func makeKeyEvent(code rune, mod uv.KeyMod) KeyEvent {
	return KeyEvent{
		event: uv.KeyPressEvent(uv.Key{
			Code: code,
			Mod:  mod,
		}),
	}
}

// Helper to create KeyEvent for printable character
func makeCharEvent(char rune) KeyEvent {
	return KeyEvent{
		event: uv.KeyPressEvent(uv.Key{
			Code: char,
			Text: string(char),
		}),
	}
}

func TestKeybind_MatchSimpleKey_Enter(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'enter' to match")
	}
}

func TestKeybind_MatchSimpleKey_Tab(t *testing.T) {
	event := makeKeyEvent(uv.KeyTab, 0)

	keybinds := []Keybind{
		{Key: "tab", Name: "Next", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'tab' to match")
	}
}

func TestKeybind_MatchSimpleKey_Escape(t *testing.T) {
	event := makeKeyEvent(uv.KeyEscape, 0)

	keybinds := []Keybind{
		{Key: "escape", Name: "Cancel", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'escape' to match")
	}
}

func TestKeybind_MatchSimpleKey_Space(t *testing.T) {
	event := makeKeyEvent(uv.KeySpace, 0)

	keybinds := []Keybind{
		{Key: "space", Name: "Toggle", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'space' to match")
	}
}

func TestKeybind_MatchLetter(t *testing.T) {
	event := makeCharEvent('a')

	keybinds := []Keybind{
		{Key: "a", Name: "Action A", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'a' to match")
	}
}

func TestKeybind_MatchLetterDifferentKey(t *testing.T) {
	event := makeCharEvent('b')

	keybinds := []Keybind{
		{Key: "a", Name: "Action A", Action: func() {}},
	}

	if matchKeybind(event, keybinds) {
		t.Error("expected 'b' not to match 'a'")
	}
}

func TestKeybind_MatchWithModifier_CtrlC(t *testing.T) {
	event := makeKeyEvent('c', uv.ModCtrl)

	keybinds := []Keybind{
		{Key: "ctrl+c", Name: "Copy", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'ctrl+c' to match")
	}
}

func TestKeybind_MatchWithModifier_CtrlS(t *testing.T) {
	event := makeKeyEvent('s', uv.ModCtrl)

	keybinds := []Keybind{
		{Key: "ctrl+s", Name: "Save", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'ctrl+s' to match")
	}
}

func TestKeybind_MatchWithModifier_AltTab(t *testing.T) {
	event := makeKeyEvent(uv.KeyTab, uv.ModAlt)

	keybinds := []Keybind{
		{Key: "alt+tab", Name: "Switch", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'alt+tab' to match")
	}
}

func TestKeybind_MatchWithModifier_ShiftEnter(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, uv.ModShift)

	keybinds := []Keybind{
		{Key: "shift+enter", Name: "New Line", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'shift+enter' to match")
	}
}

func TestKeybind_NoMatchDifferentKey(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "escape", Name: "Cancel", Action: func() {}},
	}

	if matchKeybind(event, keybinds) {
		t.Error("expected 'enter' not to match 'escape'")
	}
}

func TestKeybind_NoMatchDifferentModifier(t *testing.T) {
	event := makeKeyEvent('s', 0) // 's' without modifier

	keybinds := []Keybind{
		{Key: "ctrl+s", Name: "Save", Action: func() {}},
	}

	if matchKeybind(event, keybinds) {
		t.Error("expected 's' not to match 'ctrl+s'")
	}
}

func TestKeybind_ActionExecuted(t *testing.T) {
	executed := false
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: func() { executed = true }},
	}

	matchKeybind(event, keybinds)

	if !executed {
		t.Error("expected action to be executed")
	}
}

func TestKeybind_ActionNotExecutedOnNoMatch(t *testing.T) {
	executed := false
	event := makeKeyEvent(uv.KeyEscape, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: func() { executed = true }},
	}

	matchKeybind(event, keybinds)

	if executed {
		t.Error("expected action not to be executed when key doesn't match")
	}
}

func TestKeybind_NilActionDoesNotPanic(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: nil},
	}

	// Should not panic
	result := matchKeybind(event, keybinds)

	if !result {
		t.Error("expected match even with nil action")
	}
}

func TestKeybind_MultipleBindings_FirstMatches(t *testing.T) {
	callOrder := []string{}
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "First", Action: func() { callOrder = append(callOrder, "first") }},
		{Key: "enter", Name: "Second", Action: func() { callOrder = append(callOrder, "second") }},
	}

	matchKeybind(event, keybinds)

	if len(callOrder) != 1 || callOrder[0] != "first" {
		t.Errorf("expected only 'first' to be called, got %v", callOrder)
	}
}

func TestKeybind_MultipleBindings_MatchesCorrectOne(t *testing.T) {
	executed := ""
	event := makeKeyEvent(uv.KeyTab, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Enter", Action: func() { executed = "enter" }},
		{Key: "tab", Name: "Tab", Action: func() { executed = "tab" }},
		{Key: "escape", Name: "Escape", Action: func() { executed = "escape" }},
	}

	matchKeybind(event, keybinds)

	if executed != "tab" {
		t.Errorf("expected 'tab' to be executed, got '%s'", executed)
	}
}

func TestKeybind_EmptyKeybinds(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{}

	if matchKeybind(event, keybinds) {
		t.Error("expected no match with empty keybinds")
	}
}

func TestKeybind_HiddenStillMatches(t *testing.T) {
	executed := false
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: func() { executed = true }, Hidden: true},
	}

	result := matchKeybind(event, keybinds)

	if !result {
		t.Error("expected hidden keybind to still match")
	}
	if !executed {
		t.Error("expected hidden keybind action to be executed")
	}
}

func TestKeybind_FunctionKeys(t *testing.T) {
	tests := []struct {
		key     rune
		pattern string
	}{
		{uv.KeyF1, "f1"},
		{uv.KeyF2, "f2"},
		{uv.KeyF3, "f3"},
		{uv.KeyF10, "f10"},
		{uv.KeyF12, "f12"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			event := makeKeyEvent(tt.key, 0)
			keybinds := []Keybind{
				{Key: tt.pattern, Name: "Fn", Action: func() {}},
			}

			if !matchKeybind(event, keybinds) {
				t.Errorf("expected %s to match", tt.pattern)
			}
		})
	}
}

func TestKeybind_ArrowKeys(t *testing.T) {
	tests := []struct {
		key     rune
		pattern string
	}{
		{uv.KeyUp, "up"},
		{uv.KeyDown, "down"},
		{uv.KeyLeft, "left"},
		{uv.KeyRight, "right"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			event := makeKeyEvent(tt.key, 0)
			keybinds := []Keybind{
				{Key: tt.pattern, Name: "Arrow", Action: func() {}},
			}

			if !matchKeybind(event, keybinds) {
				t.Errorf("expected %s to match", tt.pattern)
			}
		})
	}
}

func TestKeybind_NavigationKeys(t *testing.T) {
	tests := []struct {
		key     rune
		pattern string
	}{
		{uv.KeyHome, "home"},
		{uv.KeyEnd, "end"},
		{uv.KeyPgUp, "pgup"},
		{uv.KeyPgDown, "pgdown"},
		{uv.KeyDelete, "delete"},
		{uv.KeyBackspace, "backspace"},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			event := makeKeyEvent(tt.key, 0)
			keybinds := []Keybind{
				{Key: tt.pattern, Name: "Nav", Action: func() {}},
			}

			if !matchKeybind(event, keybinds) {
				t.Errorf("expected %s to match", tt.pattern)
			}
		})
	}
}

func TestKeybind_CtrlWithArrow(t *testing.T) {
	event := makeKeyEvent(uv.KeyRight, uv.ModCtrl)

	keybinds := []Keybind{
		{Key: "ctrl+right", Name: "Word Right", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'ctrl+right' to match")
	}
}

func TestKeybind_ShiftWithArrow(t *testing.T) {
	event := makeKeyEvent(uv.KeyDown, uv.ModShift)

	keybinds := []Keybind{
		{Key: "shift+down", Name: "Select Down", Action: func() {}},
	}

	if !matchKeybind(event, keybinds) {
		t.Error("expected 'shift+down' to match")
	}
}

// Test Keybind struct fields
func TestKeybind_StructFields(t *testing.T) {
	action := func() {}
	kb := Keybind{
		Key:    "ctrl+s",
		Name:   "Save",
		Action: action,
		Hidden: true,
	}

	if kb.Key != "ctrl+s" {
		t.Errorf("expected Key 'ctrl+s', got '%s'", kb.Key)
	}
	if kb.Name != "Save" {
		t.Errorf("expected Name 'Save', got '%s'", kb.Name)
	}
	if kb.Action == nil {
		t.Error("expected Action to be set")
	}
	if !kb.Hidden {
		t.Error("expected Hidden to be true")
	}
}

func TestKeybind_ReturnsTrueOnMatch(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "enter", Name: "Submit", Action: func() {}},
	}

	result := matchKeybind(event, keybinds)

	if !result {
		t.Error("expected matchKeybind to return true on match")
	}
}

func TestKeybind_ReturnsFalseOnNoMatch(t *testing.T) {
	event := makeKeyEvent(uv.KeyEnter, 0)

	keybinds := []Keybind{
		{Key: "escape", Name: "Cancel", Action: func() {}},
	}

	result := matchKeybind(event, keybinds)

	if result {
		t.Error("expected matchKeybind to return false on no match")
	}
}
