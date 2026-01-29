package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Insert Strategy Tests ---

func TestInsertReplace(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		cursor         int
		suggestion     Suggestion
		triggerPos     int
		expectedText   string
		expectedCursor int
	}{
		{
			name:           "replace entire text",
			text:           "hello",
			cursor:         5,
			suggestion:     Suggestion{Value: "world"},
			triggerPos:     -1,
			expectedText:   "world",
			expectedCursor: 5,
		},
		{
			name:           "replace with label when value empty",
			text:           "test",
			cursor:         4,
			suggestion:     Suggestion{Label: "replacement"},
			triggerPos:     -1,
			expectedText:   "replacement",
			expectedCursor: 11,
		},
		{
			name:           "replace empty text",
			text:           "",
			cursor:         0,
			suggestion:     Suggestion{Value: "new"},
			triggerPos:     -1,
			expectedText:   "new",
			expectedCursor: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newText, newCursor := InsertReplace(tt.text, tt.cursor, tt.suggestion, tt.triggerPos)
			assert.Equal(t, tt.expectedText, newText)
			assert.Equal(t, tt.expectedCursor, newCursor)
		})
	}
}

func TestInsertFromTrigger(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		cursor         int
		suggestion     Suggestion
		triggerPos     int
		expectedText   string
		expectedCursor int
	}{
		{
			name:           "replace from @ trigger",
			text:           "hello @joh",
			cursor:         10,
			suggestion:     Suggestion{Value: "@john"},
			triggerPos:     6,
			expectedText:   "hello @john",
			expectedCursor: 11,
		},
		{
			name:           "replace from # trigger at start",
			text:           "#issu",
			cursor:         5,
			suggestion:     Suggestion{Value: "#issue-123"},
			triggerPos:     0,
			expectedText:   "#issue-123",
			expectedCursor: 10,
		},
		{
			name:           "no trigger position uses start",
			text:           "test",
			cursor:         4,
			suggestion:     Suggestion{Value: "replaced"},
			triggerPos:     -1,
			expectedText:   "replaced",
			expectedCursor: 8,
		},
		{
			name:           "preserve text after cursor",
			text:           "hello @joh world",
			cursor:         10,
			suggestion:     Suggestion{Value: "@john"},
			triggerPos:     6,
			expectedText:   "hello @john world",
			expectedCursor: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newText, newCursor := InsertFromTrigger(tt.text, tt.cursor, tt.suggestion, tt.triggerPos)
			assert.Equal(t, tt.expectedText, newText)
			assert.Equal(t, tt.expectedCursor, newCursor)
		})
	}
}

func TestInsertAtCursor(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		cursor         int
		suggestion     Suggestion
		triggerPos     int
		expectedText   string
		expectedCursor int
	}{
		{
			name:           "insert at middle",
			text:           "hello world",
			cursor:         5,
			suggestion:     Suggestion{Value: " there"},
			triggerPos:     -1,
			expectedText:   "hello there world",
			expectedCursor: 11,
		},
		{
			name:           "insert at start",
			text:           "world",
			cursor:         0,
			suggestion:     Suggestion{Value: "hello "},
			triggerPos:     -1,
			expectedText:   "hello world",
			expectedCursor: 6,
		},
		{
			name:           "insert at end",
			text:           "hello",
			cursor:         5,
			suggestion:     Suggestion{Value: " world"},
			triggerPos:     -1,
			expectedText:   "hello world",
			expectedCursor: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newText, newCursor := InsertAtCursor(tt.text, tt.cursor, tt.suggestion, tt.triggerPos)
			assert.Equal(t, tt.expectedText, newText)
			assert.Equal(t, tt.expectedCursor, newCursor)
		})
	}
}

func TestInsertReplaceWord(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		cursor         int
		suggestion     Suggestion
		triggerPos     int
		expectedText   string
		expectedCursor int
	}{
		{
			name:           "replace partial word",
			text:           "hello wor",
			cursor:         9,
			suggestion:     Suggestion{Value: "world"},
			triggerPos:     -1,
			expectedText:   "hello world",
			expectedCursor: 11,
		},
		{
			name:           "replace word in middle",
			text:           "the quick fox",
			cursor:         9, // cursor at end of "quick"
			suggestion:     Suggestion{Value: "slow"},
			triggerPos:     -1,
			expectedText:   "the slow fox",
			expectedCursor: 8,
		},
		{
			name:           "replace first word",
			text:           "hello world",
			cursor:         5,
			suggestion:     Suggestion{Value: "hi"},
			triggerPos:     -1,
			expectedText:   "hi world",
			expectedCursor: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newText, newCursor := InsertReplaceWord(tt.text, tt.cursor, tt.suggestion, tt.triggerPos)
			assert.Equal(t, tt.expectedText, newText)
			assert.Equal(t, tt.expectedCursor, newCursor)
		})
	}
}

func TestInsertStrategy_MultiByteChars(t *testing.T) {
	t.Run("InsertFromTrigger with unicode", func(t *testing.T) {
		text := "Hello @日本"
		cursor := 10 // after 日本 (rune count: H e l l o   @ 日 本 = 9 + cursor at end)
		suggestion := Suggestion{Value: "@日本語"}
		triggerPos := 6 // @

		newText, newCursor := InsertFromTrigger(text, cursor, suggestion, triggerPos)
		assert.Equal(t, "Hello @日本語", newText)
		assert.Equal(t, 10, newCursor) // 6 (before @) + 4 (runes in @日本語)
	})

	t.Run("InsertReplaceWord with emoji", func(t *testing.T) {
		text := "test 👋hel"
		cursor := 9
		suggestion := Suggestion{Value: "👋hello"}

		newText, _ := InsertReplaceWord(text, cursor, suggestion, -1)
		assert.Equal(t, "test 👋hello", newText)
	})
}

// --- Trigger Detection Tests ---

func TestAutocomplete_FindTriggerPosition(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		cursorPos    int
		triggerChars []rune
		wordBoundary bool
		expected     int
	}{
		{
			name:         "find @ trigger",
			text:         "hello @john",
			cursorPos:    11,
			triggerChars: []rune{'@'},
			wordBoundary: true,
			expected:     6,
		},
		{
			name:         "find # trigger",
			text:         "fix #123",
			cursorPos:    8,
			triggerChars: []rune{'#'},
			wordBoundary: true,
			expected:     4,
		},
		{
			name:         "no trigger found",
			text:         "hello world",
			cursorPos:    11,
			triggerChars: []rune{'@'},
			wordBoundary: true,
			expected:     -1,
		},
		{
			name:         "@ in email ignored with word boundary",
			text:         "user@example.com",
			cursorPos:    16,
			triggerChars: []rune{'@'},
			wordBoundary: true,
			expected:     -1, // @ is not at word boundary
		},
		{
			name:         "multiple triggers first match",
			text:         "hello @user #tag",
			cursorPos:    16,
			triggerChars: []rune{'@', '#'},
			wordBoundary: true,
			expected:     12, // #tag is closer to cursor
		},
		{
			name:         "trigger at start of text",
			text:         "@mention",
			cursorPos:    8,
			triggerChars: []rune{'@'},
			wordBoundary: true,
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := Autocomplete{
				TriggerChars:          tt.triggerChars,
				TriggerAtWordBoundary: tt.wordBoundary,
			}
			result := ac.findTriggerPosition(tt.text, tt.cursorPos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutocomplete_ExtractQuery(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		cursorPos    int
		triggerPos   int
		triggerChars []rune
		expected     string
	}{
		{
			name:         "extract query after @",
			text:         "hello @john",
			cursorPos:    11,
			triggerPos:   6,
			triggerChars: []rune{'@'},
			expected:     "john",
		},
		{
			name:         "partial query",
			text:         "hello @jo",
			cursorPos:    9,
			triggerPos:   6,
			triggerChars: []rune{'@'},
			expected:     "jo",
		},
		{
			name:         "empty query after trigger",
			text:         "hello @",
			cursorPos:    7,
			triggerPos:   6,
			triggerChars: []rune{'@'},
			expected:     "",
		},
		{
			name:         "no trigger - always on mode",
			text:         "hello",
			cursorPos:    5,
			triggerPos:   -1,
			triggerChars: nil, // always-on mode
			expected:     "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := Autocomplete{
				TriggerChars: tt.triggerChars,
			}
			result := ac.extractQuery(tt.text, tt.cursorPos, tt.triggerPos)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// --- State Tests ---

func TestAutocompleteState_Visibility(t *testing.T) {
	state := NewAutocompleteState()

	assert.False(t, state.IsVisible())

	state.Show()
	assert.True(t, state.IsVisible())

	state.Hide()
	assert.False(t, state.IsVisible())
}

func TestAutocompleteState_Suggestions(t *testing.T) {
	state := NewAutocompleteState()

	suggestions := []Suggestion{
		{Label: "apple", Value: "apple"},
		{Label: "banana", Value: "banana"},
	}

	state.SetSuggestions(suggestions)

	// Verify suggestions are set
	got := state.Suggestions.Peek()
	require.Len(t, got, 2)
	assert.Equal(t, "apple", got[0].Label)
	assert.Equal(t, "banana", got[1].Label)
}

// --- MinChars Tests ---

func TestAutocomplete_MinChars(t *testing.T) {
	t.Run("shows popup when rune count meets MinChars", func(t *testing.T) {
		state := NewAutocompleteState()
		state.SetSuggestions([]Suggestion{{Label: "test"}})

		inputState := NewTextInputState("日本") // 2 runes, 6 bytes

		ac := Autocomplete{
			State:        state,
			Child:        TextInput{ID: "input", State: inputState},
			MinChars:     2,
			TriggerChars: nil, // always-on mode
		}

		// Simulate text change
		ac.updateTriggerAndQuery("日本", 2)

		assert.True(t, state.IsVisible(), "should show popup when 2+ runes")
	})

	t.Run("hides popup when rune count below MinChars", func(t *testing.T) {
		state := NewAutocompleteState()
		state.SetSuggestions([]Suggestion{{Label: "test"}})

		inputState := NewTextInputState("日") // 1 rune

		ac := Autocomplete{
			State:        state,
			Child:        TextInput{ID: "input", State: inputState},
			MinChars:     2,
			TriggerChars: nil,
		}

		ac.updateTriggerAndQuery("日", 1)

		assert.False(t, state.IsVisible(), "should hide popup when less than 2 runes")
	})
}

// --- Filtering Tests ---

func TestAutocomplete_FilteredCount(t *testing.T) {
	t.Run("contains match", func(t *testing.T) {
		state := NewAutocompleteState()
		suggestions := []Suggestion{
			{Label: "apple"},
			{Label: "application"},
			{Label: "banana"},
		}
		state.listState.SetItems(suggestions)
		state.filterState.Mode.Set(FilterContains)

		state.filterState.Query.Set("app")
		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 2, state.listState.FilteredCount())

		state.filterState.Query.Set("xyz")
		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 0, state.listState.FilteredCount())
	})

	t.Run("fuzzy match", func(t *testing.T) {
		state := NewAutocompleteState()
		suggestions := []Suggestion{
			{Label: "file_open"},
			{Label: "file_close"},
			{Label: "folder_open"},
		}
		state.listState.SetItems(suggestions)
		state.filterState.Mode.Set(FilterFuzzy)

		// "fop" should match "file_open" and "folder_open" (f-o-p in sequence)
		state.filterState.Query.Set("fop")
		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 2, state.listState.FilteredCount())

		state.filterState.Query.Set("xyz")
		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 0, state.listState.FilteredCount())
	})

	t.Run("empty query matches all", func(t *testing.T) {
		state := NewAutocompleteState()
		suggestions := []Suggestion{
			{Label: "a"},
			{Label: "b"},
			{Label: "c"},
		}
		state.listState.SetItems(suggestions)

		state.filterState.Query.Set("")
		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 3, state.listState.FilteredCount())
	})

	t.Run("empty suggestions returns zero", func(t *testing.T) {
		state := NewAutocompleteState()
		state.listState.SetItems([]Suggestion{})

		state.listState.ApplyFilter(state.filterState, suggestionMatchItem)
		assert.Equal(t, 0, state.listState.FilteredCount())
	})
}

// --- Snapshot Tests ---

func TestSnapshot_Autocomplete_Hidden(t *testing.T) {
	inputState := NewTextInputState("hello")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "hello"},
		{Label: "help"},
	})
	// Don't show the popup
	acState.Hide()

	widget := Autocomplete{
		ID:    "ac-hidden",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(20)},
	}

	AssertSnapshot(t, widget, 30, 5, "Autocomplete with popup hidden, just shows input")
}

func TestSnapshot_Autocomplete_WithSuggestions(t *testing.T) {
	inputState := NewTextInputState("he")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "hello", Description: "greeting"},
		{Label: "help", Description: "assistance"},
		{Label: "helicopter"},
	})
	acState.Show()

	// Set up filter state
	acState.listState.SetItems(acState.Suggestions.Peek())
	acState.filterState.Query.Set("he")

	ac := Autocomplete{
		ID:    "ac-visible",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(25)},
	}

	AssertSnapshot(t, ac, 35, 12, "Autocomplete showing dropdown with 3 matching suggestions")
}

func TestSnapshot_Autocomplete_ActiveItem(t *testing.T) {
	inputState := NewTextInputState("app")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "apple"},
		{Label: "application"},
		{Label: "appetite"},
	})
	acState.Show()

	// Set up filter state
	acState.listState.SetItems(acState.Suggestions.Peek())
	acState.filterState.Query.Set("app")

	ac := Autocomplete{
		ID:    "ac-active",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(20)},
	}

	// Select second item
	acState.listState.SelectNext()

	AssertSnapshotNamed(t, "second_selected", ac, 30, 10,
		"Autocomplete with second item (application) highlighted")
}

func TestSnapshot_Autocomplete_EmptyResults(t *testing.T) {
	inputState := NewTextInputState("xyz")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "hello"},
		{Label: "world"},
	})

	// Set up filter state with non-matching query
	acState.listState.SetItems(acState.Suggestions.Peek())
	acState.filterState.Query.Set("xyz")

	ac := Autocomplete{
		ID:    "ac-empty",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(20)},
	}

	// Popup should not show when no matches
	AssertSnapshot(t, ac, 30, 8, "Autocomplete with no matching results - popup hidden")
}

func TestSnapshot_Autocomplete_WithTrigger(t *testing.T) {
	inputState := NewTextInputState("Hello @jo")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "john", Value: "@john"},
		{Label: "joe", Value: "@joe"},
		{Label: "jordan", Value: "@jordan"},
	})
	acState.Show()

	// Set up filter state
	acState.listState.SetItems(acState.Suggestions.Peek())
	acState.filterState.Query.Set("jo")

	ac := Autocomplete{
		ID:           "ac-trigger",
		State:        acState,
		TriggerChars: []rune{'@'},
		Child:        TextInput{ID: "input", State: inputState, Width: Cells(25)},
	}

	AssertSnapshot(t, ac, 35, 10, "Autocomplete with @ trigger showing matching usernames")
}

func TestSnapshot_Autocomplete_CustomRender(t *testing.T) {
	inputState := NewTextInputState("cmd")
	acState := NewAutocompleteState()
	acState.SetSuggestions([]Suggestion{
		{Label: "Open File", Icon: "O", Description: "Ctrl+O"},
		{Label: "Save", Icon: "S", Description: "Ctrl+S"},
		{Label: "Quit", Icon: "Q", Description: "Ctrl+Q"},
	})
	acState.Show()

	// Set up filter state (empty query shows all)
	acState.listState.SetItems(acState.Suggestions.Peek())
	acState.filterState.Query.Set("")

	ac := Autocomplete{
		ID:    "ac-custom",
		State: acState,
		Child: TextInput{ID: "input", State: inputState, Width: Cells(30)},
		RenderSuggestion: func(s Suggestion, active bool, match MatchResult, ctx BuildContext) Widget {
			theme := ctx.Theme()
			style := Style{Padding: EdgeInsets{Left: 1, Right: 1}}
			if active {
				style.BackgroundColor = theme.Primary
				style.ForegroundColor = theme.TextOnPrimary
			}
			return Row{
				Style: style,
				Width: Flex(1),
				Children: []Widget{
					Text{Content: "[" + s.Icon + "] ", Style: Style{ForegroundColor: theme.Accent}},
					Text{Content: s.Label},
					Spacer{Width: Flex(1)},
					Text{Content: s.Description, Style: Style{ForegroundColor: theme.TextMuted}},
				},
			}
		},
	}

	AssertSnapshot(t, ac, 45, 12, "Autocomplete with custom icon and shortcut rendering")
}
