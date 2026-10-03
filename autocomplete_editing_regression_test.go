package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutocomplete_UnicodeSuggestionRemainsEditable(t *testing.T) {
	for _, kind := range []string{"input", "area"} {
		t.Run(kind, func(t *testing.T) {
			popup := NewAutocompleteState()
			popup.SetSuggestions([]Suggestion{{Label: "developer", Value: "👩‍💻"}})
			popup.Show()
			if kind == "input" {
				state := NewTextInputState("")
				input := TextInput{State: state}
				ac := Autocomplete{State: popup, Child: input}
				ac.onEnterTextInput()
				assert.Equal(t, 1, state.CursorIndex.Peek())
				require.NotPanics(t, func() { input.HandlePaste("!") })
				assert.Equal(t, "👩‍💻!", state.GetText())
			} else {
				state := NewTextAreaState("")
				area := TextArea{State: state}
				ac := Autocomplete{State: popup, Child: area}
				ac.onEnterTextArea()
				assert.Equal(t, 1, state.CursorIndex.Peek())
				require.NotPanics(t, func() { area.HandlePaste("!") })
				assert.Equal(t, "👩‍💻!", state.GetText())
			}
		})
	}
}

func TestAutocomplete_TriggerAfterGrapheme(t *testing.T) {
	state := NewTextInputState("👩‍💻 @jo")
	popup := NewAutocompleteState()
	popup.SetSuggestions([]Suggestion{{Label: "john", Value: "@john"}})
	ac := Autocomplete{State: popup, Child: TextInput{State: state}, TriggerChars: []rune{'@'}}
	ac.updateTriggerAndQuery(ac.getChildTextAndCursor())
	assert.True(t, popup.IsVisible())
	assert.Equal(t, "jo", popup.filterQuery)
	ac.onEnterTextInput()
	assert.Equal(t, "👩‍💻 @john", state.GetText())
	assert.Equal(t, 7, state.CursorIndex.Peek())
}

func TestAutocomplete_TextAreaEnterReplacesSelection(t *testing.T) {
	state := NewTextAreaState("hello world")
	state.SelectionAnchor.Set(6)
	var changed string
	ac := Autocomplete{
		State: NewAutocompleteState(),
		Child: TextArea{State: state, OnChange: func(text string) { changed = text }},
	}
	ac.onEnterTextArea()
	assert.Equal(t, "hello \n", state.GetText())
	assert.Equal(t, 7, state.CursorIndex.Peek())
	assert.Equal(t, "hello \n", changed)
	assert.False(t, state.HasSelection())
}

func TestAutocomplete_InsertStrategiesUseGraphemePositions(t *testing.T) {
	tests := []struct {
		name     string
		strategy InsertStrategy
		text     string
		cursor   int
		trigger  int
		value    string
		wantText string
		wantPos  int
	}{
		{"replace joined emoji", InsertReplace, "old", 3, -1, "👩‍💻", "👩‍💻", 1},
		{"replace combining mark", InsertReplace, "old", 3, -1, "cafe\u0301", "cafe\u0301", 4},
		{"trigger after joined emoji", InsertFromTrigger, "👩‍💻 @de tail", 5, 2, "@dev", "👩‍💻 @dev tail", 6},
		{"trigger after combining mark", InsertFromTrigger, "e\u0301 @a tail", 4, 2, "@b", "e\u0301 @b tail", 4},
		{"insert after joined emoji", InsertAtCursor, "👩‍💻!", 1, -1, "👨‍💻", "👩‍💻👨‍💻!", 2},
		{"insert after combining mark", InsertAtCursor, "e\u0301!", 1, -1, "a\u0301", "e\u0301a\u0301!", 2},
		{"insert merges combining mark", InsertAtCursor, "e!", 1, -1, "\u0301", "e\u0301!", 1},
		{"trigger replacement merges combining mark", InsertFromTrigger, "e@x!", 3, 1, "\u0301", "e\u0301!", 1},
		{"word after joined emoji", InsertReplaceWord, "👩‍💻 hel tail", 5, -1, "👨‍💻", "👩‍💻 👨‍💻 tail", 3},
		{"word after combining mark", InsertReplaceWord, "e\u0301 hel tail", 5, -1, "cafe\u0301", "e\u0301 cafe\u0301 tail", 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, cursor := tt.strategy(tt.text, tt.cursor, Suggestion{Value: tt.value}, tt.trigger)
			assert.Equal(t, tt.wantText, text)
			assert.Equal(t, tt.wantPos, cursor)
		})
	}
}

func TestAutocomplete_QueriesUseGraphemePositions(t *testing.T) {
	for _, text := range []string{"👩‍💻", "e\u0301"} {
		t.Run(text, func(t *testing.T) {
			ac := Autocomplete{State: NewAutocompleteState()}
			ac.updateTriggerAndQuery(text, 1)
			assert.Equal(t, text, ac.State.filterQuery)

			ac.TriggerChars = []rune{'@'}
			ac.updateTriggerAndQuery(text+" @"+text, 4)
			assert.Equal(t, 2, ac.State.triggerPosition)
			assert.Equal(t, text, ac.State.filterQuery)
			assert.True(t, ac.State.IsVisible())
		})
	}
}
