package terma

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Replacing a long, scrolled value with a short one used to leave the view
// scrolled past the end of the new text, so the input looked empty.
func TestSnapshot_TextInput_SetTextShorterAfterScrolling(t *testing.T) {
	state := NewTextInputState("")
	state.Insert(strings.Repeat("x", 150))
	widget := TextInput{ID: "url", State: state, Style: Style{Width: Cells(100)}}
	RenderToBuffer(widget, 100, 1)
	assert.Equal(t, 51, state.scrollOffset, "the long value scrolls to show the cursor at its end")

	state.SetText("https://short.test")

	AssertSnapshot(t, widget, 100, 1,
		"After SetText with a short URL the input shows 'https://short.test' from its left edge, with the cursor after it.")
}

// The clamp also covers content replaced without SetText.
func TestTextInput_ScrollOffsetClampedWhenContentShrinks(t *testing.T) {
	state := NewTextInputState(strings.Repeat("x", 150))
	widget := TextInput{ID: "url", State: state, Style: Style{Width: Cells(100)}}
	RenderToBuffer(widget, 100, 1)
	assert.Equal(t, 51, state.scrollOffset)

	state.Content.Set(splitGraphemes(strings.Repeat("y", 120)))
	state.clampCursor()
	RenderToBuffer(widget, 100, 1)

	assert.Equal(t, 21, state.scrollOffset, "scrolled only far enough to show the end and the cursor cell")
}

func TestTextInputState_SetTextResetsScroll(t *testing.T) {
	state := NewTextInputState("")
	state.scrollOffset = 40
	state.SetText("short")
	assert.Equal(t, 0, state.scrollOffset)
}

// An unfocused input shows the start of a value that doesn't fit, even though
// its cursor is at the end.
func TestSnapshot_TextInput_UnfocusedShowsStartOfLongValue(t *testing.T) {
	widget := Column{
		Spacing: 1,
		Children: []Widget{
			TextInput{ID: "focused", State: NewTextInputState("Content-Type"), Style: Style{Width: Cells(8)}},
			TextInput{ID: "unfocused", State: NewTextInputState("Content-Type"), Style: Style{Width: Cells(8)}},
		},
	}
	AssertSnapshot(t, widget, 10, 3,
		"Two 8-wide inputs holding 'Content-Type'. The focused one (top) shows the end 'nt-Type' with the cursor after it; the unfocused one (bottom) shows the start 'Content-'.")
}
