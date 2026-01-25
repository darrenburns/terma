package terma

import "testing"

// --- Grapheme Helper Tests ---

func TestSplitGraphemes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", nil},
		{"ascii", "hello", []string{"h", "e", "l", "l", "o"}},
		{"unicode", "héllo", []string{"h", "é", "l", "l", "o"}},
		{"emoji", "hi👋", []string{"h", "i", "👋"}},
		{"spaces", "a b", []string{"a", " ", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitGraphemes(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitGraphemes(%q) = %v, want %v", tt.input, result, tt.expected)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("splitGraphemes(%q)[%d] = %q, want %q", tt.input, i, result[i], tt.expected[i])
				}
			}
		})
	}
}

func TestJoinGraphemes(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected string
	}{
		{"nil", nil, ""},
		{"empty", []string{}, ""},
		{"single", []string{"a"}, "a"},
		{"multiple", []string{"h", "e", "l", "l", "o"}, "hello"},
		{"unicode", []string{"h", "é", "l", "l", "o"}, "héllo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinGraphemes(tt.input)
			if result != tt.expected {
				t.Errorf("joinGraphemes(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGraphemeWidth(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"a", 1},
		{"é", 1},
		{"👋", 2}, // Emoji is typically 2 cells wide
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := graphemeWidth(tt.input)
			if result != tt.expected {
				t.Errorf("graphemeWidth(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"a", true},
		{"Z", true},
		{"0", true},
		{"_", true},
		{" ", false},
		{"-", false},
		{".", false},
		{"é", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isWordChar(tt.input)
			if result != tt.expected {
				t.Errorf("isWordChar(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// --- TextInputState Tests ---

func TestNewTextInputState(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		state := NewTextInputState("")
		if state.GetText() != "" {
			t.Errorf("GetText() = %q, want empty", state.GetText())
		}
		if state.CursorIndex.Peek() != 0 {
			t.Errorf("CursorIndex = %d, want 0", state.CursorIndex.Peek())
		}
	})

	t.Run("with text", func(t *testing.T) {
		state := NewTextInputState("hello")
		if state.GetText() != "hello" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello")
		}
		// Cursor should be at end
		if state.CursorIndex.Peek() != 5 {
			t.Errorf("CursorIndex = %d, want 5", state.CursorIndex.Peek())
		}
	})
}

func TestTextInputState_SetText(t *testing.T) {
	state := NewTextInputState("hello")
	state.SetText("world")

	if state.GetText() != "world" {
		t.Errorf("GetText() = %q, want %q", state.GetText(), "world")
	}
}

func TestTextInputState_Insert(t *testing.T) {
	t.Run("insert at end", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.Insert(" world")
		if state.GetText() != "hello world" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello world")
		}
		if state.CursorIndex.Peek() != 11 {
			t.Errorf("CursorIndex = %d, want 11", state.CursorIndex.Peek())
		}
	})

	t.Run("insert at beginning", func(t *testing.T) {
		state := NewTextInputState("world")
		state.CursorHome()
		state.Insert("hello ")
		if state.GetText() != "hello world" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello world")
		}
	})

	t.Run("insert in middle", func(t *testing.T) {
		state := NewTextInputState("helo")
		state.CursorIndex.Set(2) // After "he"
		state.Insert("l")
		if state.GetText() != "hello" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello")
		}
	})
}

func TestTextInputState_DeleteBackward(t *testing.T) {
	t.Run("delete at end", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.DeleteBackward()
		if state.GetText() != "hell" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hell")
		}
		if state.CursorIndex.Peek() != 4 {
			t.Errorf("CursorIndex = %d, want 4", state.CursorIndex.Peek())
		}
	})

	t.Run("delete at beginning - no op", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.CursorHome()
		state.DeleteBackward()
		if state.GetText() != "hello" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello")
		}
	})

	t.Run("delete in middle", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.CursorIndex.Set(2) // After "he"
		state.DeleteBackward()
		if state.GetText() != "hllo" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hllo")
		}
	})
}

func TestTextInputState_DeleteForward(t *testing.T) {
	t.Run("delete at beginning", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.CursorHome()
		state.DeleteForward()
		if state.GetText() != "ello" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "ello")
		}
	})

	t.Run("delete at end - no op", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.DeleteForward()
		if state.GetText() != "hello" {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello")
		}
	})
}

func TestTextInputState_DeleteToBeginning(t *testing.T) {
	state := NewTextInputState("hello world")
	state.CursorIndex.Set(6) // After "hello "
	state.DeleteToBeginning()

	if state.GetText() != "world" {
		t.Errorf("GetText() = %q, want %q", state.GetText(), "world")
	}
	if state.CursorIndex.Peek() != 0 {
		t.Errorf("CursorIndex = %d, want 0", state.CursorIndex.Peek())
	}
}

func TestTextInputState_DeleteToEnd(t *testing.T) {
	state := NewTextInputState("hello world")
	state.CursorIndex.Set(5) // After "hello"
	state.DeleteToEnd()

	if state.GetText() != "hello" {
		t.Errorf("GetText() = %q, want %q", state.GetText(), "hello")
	}
}

func TestTextInputState_DeleteWordBackward(t *testing.T) {
	t.Run("delete word", func(t *testing.T) {
		state := NewTextInputState("hello world")
		state.DeleteWordBackward()
		if state.GetText() != "hello " {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello ")
		}
	})

	t.Run("delete word with trailing space", func(t *testing.T) {
		state := NewTextInputState("hello world ")
		state.DeleteWordBackward()
		if state.GetText() != "hello " {
			t.Errorf("GetText() = %q, want %q", state.GetText(), "hello ")
		}
	})
}

func TestTextInputState_CursorMovement(t *testing.T) {
	state := NewTextInputState("hello")

	t.Run("cursor left", func(t *testing.T) {
		state.CursorEnd()
		state.CursorLeft()
		if state.CursorIndex.Peek() != 4 {
			t.Errorf("CursorIndex = %d, want 4", state.CursorIndex.Peek())
		}
	})

	t.Run("cursor right", func(t *testing.T) {
		state.CursorRight()
		if state.CursorIndex.Peek() != 5 {
			t.Errorf("CursorIndex = %d, want 5", state.CursorIndex.Peek())
		}
	})

	t.Run("cursor home", func(t *testing.T) {
		state.CursorHome()
		if state.CursorIndex.Peek() != 0 {
			t.Errorf("CursorIndex = %d, want 0", state.CursorIndex.Peek())
		}
	})

	t.Run("cursor end", func(t *testing.T) {
		state.CursorEnd()
		if state.CursorIndex.Peek() != 5 {
			t.Errorf("CursorIndex = %d, want 5", state.CursorIndex.Peek())
		}
	})

	t.Run("cursor left at beginning - no op", func(t *testing.T) {
		state.CursorHome()
		state.CursorLeft()
		if state.CursorIndex.Peek() != 0 {
			t.Errorf("CursorIndex = %d, want 0", state.CursorIndex.Peek())
		}
	})

	t.Run("cursor right at end - no op", func(t *testing.T) {
		state.CursorEnd()
		state.CursorRight()
		if state.CursorIndex.Peek() != 5 {
			t.Errorf("CursorIndex = %d, want 5", state.CursorIndex.Peek())
		}
	})
}

func TestTextInputState_CursorWordMovement(t *testing.T) {
	state := NewTextInputState("hello world foo")

	t.Run("word left from end", func(t *testing.T) {
		state.CursorEnd()
		state.CursorWordLeft()
		if state.CursorIndex.Peek() != 12 { // Before "foo"
			t.Errorf("CursorIndex = %d, want 12", state.CursorIndex.Peek())
		}
	})

	t.Run("word left again", func(t *testing.T) {
		state.CursorWordLeft()
		if state.CursorIndex.Peek() != 6 { // Before "world"
			t.Errorf("CursorIndex = %d, want 6", state.CursorIndex.Peek())
		}
	})

	t.Run("word right from beginning", func(t *testing.T) {
		state.CursorHome()
		state.CursorWordRight()
		if state.CursorIndex.Peek() != 6 { // After "hello "
			t.Errorf("CursorIndex = %d, want 6", state.CursorIndex.Peek())
		}
	})
}

func TestTextInputState_CursorDisplayX(t *testing.T) {
	t.Run("ascii", func(t *testing.T) {
		state := NewTextInputState("hello")
		state.CursorIndex.Set(3)
		if state.cursorDisplayX() != 3 {
			t.Errorf("cursorDisplayX() = %d, want 3", state.cursorDisplayX())
		}
	})

	t.Run("with wide char", func(t *testing.T) {
		state := NewTextInputState("hi👋")
		state.CursorEnd()
		// "hi" = 2 cells, "👋" = 2 cells, total = 4
		if state.cursorDisplayX() != 4 {
			t.Errorf("cursorDisplayX() = %d, want 4", state.cursorDisplayX())
		}
	})
}

func TestTextInputState_ContentWidth(t *testing.T) {
	t.Run("ascii", func(t *testing.T) {
		state := NewTextInputState("hello")
		if state.contentWidth() != 5 {
			t.Errorf("contentWidth() = %d, want 5", state.contentWidth())
		}
	})

	t.Run("with emoji", func(t *testing.T) {
		state := NewTextInputState("hi👋")
		// "hi" = 2 cells, "👋" = 2 cells
		if state.contentWidth() != 4 {
			t.Errorf("contentWidth() = %d, want 4", state.contentWidth())
		}
	})
}

func TestTextInput_SpaceKeyRepresentation(t *testing.T) {
	// Test that we can detect a space character correctly
	key := " "
	runes := []rune(key)
	if len(runes) != 1 {
		t.Errorf("space key has %d runes, expected 1", len(runes))
	}
	if runes[0] != ' ' {
		t.Errorf("space rune is %q, expected ' '", runes[0])
	}
}

// --- Snapshot Tests ---

func TestSnapshot_TextInput_PlaceholderFocused(t *testing.T) {
	state := NewTextInputState("")
	widget := TextInput{
		ID:          "textinput-placeholder-focused",
		State:       state,
		Placeholder: "Type here...",
		Width:       Cells(20),
	}

	AssertSnapshotNamed(t, "focused", widget, 20, 1,
		"Empty TextInput with placeholder, focused. First placeholder character should be visible under cursor (reversed).")
}

func TestSnapshot_TextInput_PlaceholderUnfocused(t *testing.T) {
	state := NewTextInputState("")
	// Put a Button first so it takes focus, leaving the TextInput unfocused
	widget := Column{
		Children: []Widget{
			&Button{ID: "focus-stealer", Label: ""},
			TextInput{
				ID:          "textinput-placeholder-unfocused",
				State:       state,
				Placeholder: "Type here...",
				Width:       Cells(20),
			},
		},
	}

	AssertSnapshotNamed(t, "unfocused", widget, 20, 2,
		"Empty TextInput with placeholder, unfocused. Full placeholder text visible without cursor.")
}
