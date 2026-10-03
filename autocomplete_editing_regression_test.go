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

func TestSnapshot_AutocompleteEditing(t *testing.T) {
	t.Run("unicode", func(t *testing.T) {
		state := NewTextInputState("👩‍💻 @jo")
		popup := NewAutocompleteState()
		popup.SetSuggestions([]Suggestion{{Label: "john", Value: "@john"}})
		input := TextInput{State: state, Width: Cells(24)}
		ac := Autocomplete{State: popup, Child: input, TriggerChars: []rune{'@'}}
		ac.updateTriggerAndQuery(ac.getChildTextAndCursor())
		ac.onEnterTextInput()
		input.HandlePaste("!")
		AssertSnapshot(t, input, 26, 3, "Autocomplete preserves a joined emoji and replaces the complete trigger query before further typing.")
	})

	t.Run("selection", func(t *testing.T) {
		state := NewTextAreaState("hello world")
		state.SelectionAnchor.Set(6)
		area := TextArea{State: state, Width: Cells(24), Height: Cells(3)}
		ac := Autocomplete{State: NewAutocompleteState(), Child: area}
		ac.onEnterTextArea()
		area.HandlePaste("next line")
		AssertSnapshot(t, area, 26, 4, "Enter replaces the selected word when autocomplete has no suggestions, and typing continues on the next line.")
	})
}

func TestAutocomplete_SeparatelyInsertedCombiningMark(t *testing.T) {
	for _, kind := range []string{"input", "area"} {
		for _, scenario := range []string{"insert before suffix", "query before suffix"} {
			t.Run(kind+"/"+scenario, func(t *testing.T) {
				var moveLeft func()
				var getText func() string
				var paste func(string) bool
				var query string
				ac := Autocomplete{
					State:         NewAutocompleteState(),
					Insert:        InsertAtCursor,
					OnQueryChange: func(value string) { query = value },
				}
				if scenario == "query before suffix" {
					ac.TriggerChars = []rune{'@'}
				}
				if kind == "input" {
					state := NewTextInputState("")
					ac.Child = TextInput{State: state}
					child := ac.wrapTextInput(ac.Child.(TextInput), BuildContext{}, true).(TextInput)
					paste, moveLeft, getText = child.HandlePaste, state.CursorLeft, state.GetText
				} else {
					state := NewTextAreaState("")
					ac.Child = TextArea{State: state}
					child := ac.wrapTextArea(ac.Child.(TextArea), BuildContext{}, true).(TextArea)
					paste, moveLeft, getText = child.HandlePaste, state.CursorLeft, state.GetText
				}

				paste("e")
				paste("\u0301")
				if scenario == "insert before suffix" {
					paste("x")
					moveLeft()
					ac.selectSuggestion(Suggestion{Value: "!"})
					assert.Equal(t, "e\u0301!x", getText())
					paste("?")
					assert.Equal(t, "e\u0301!?x", getText())
				} else {
					paste(" @ab")
					moveLeft()
					ac.updateTriggerAndQuery(ac.getChildTextAndCursor())
					assert.Equal(t, "a", ac.State.filterQuery)
					assert.Equal(t, 2, ac.State.triggerPosition)
					paste("c")
					assert.Equal(t, "e\u0301 @acb", getText())
					assert.Equal(t, "ac", ac.State.filterQuery)
					assert.Equal(t, "ac", query)
				}
			})
		}
	}
}
