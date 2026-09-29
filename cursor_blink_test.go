package terma

import (
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enableCursorBlinkForTest(t *testing.T) {
	t.Helper()
	SetCursorBlink(true)
	t.Cleanup(func() { SetCursorBlink(false) })
}

func TestCursorBlink_OffByDefault(t *testing.T) {
	assert.False(t, CursorBlink())
	toggleCursorBlink()
	assert.True(t, cursorVisible(), "the cursor never hides while blinking is off")
}

func TestCursorBlink_TogglesAndInputShowsCursor(t *testing.T) {
	enableCursorBlinkForTest(t)

	assert.True(t, cursorVisible())
	toggleCursorBlink()
	assert.False(t, cursorVisible())
	showCursorForInput()
	assert.True(t, cursorVisible())
	toggleCursorBlink()

	SetCursorBlink(false)
	assert.True(t, cursorVisible(), "turning blinking off shows the cursor")
}

func blinkTestInput() Widget {
	return TextInput{ID: "input", State: NewTextInputState("hello"), Style: Style{Width: Cells(10)}}
}

func TestSnapshot_TextInput_CursorBlinkShown(t *testing.T) {
	enableCursorBlinkForTest(t)
	AssertSnapshot(t, blinkTestInput(), 10, 1, "Focused input, blink phase shown: reverse-video cursor after 'hello'.")
}

func TestSnapshot_TextInput_CursorBlinkHidden(t *testing.T) {
	enableCursorBlinkForTest(t)
	toggleCursorBlink()
	AssertSnapshot(t, blinkTestInput(), 10, 1, "Focused input, blink phase hidden: 'hello' with no cursor.")
}

func TestSnapshot_TextArea_CursorBlinkHidden(t *testing.T) {
	enableCursorBlinkForTest(t)
	toggleCursorBlink()
	widget := TextArea{ID: "area", State: NewTextAreaState("line one\nline two"), Style: Style{Width: Cells(12), Height: Cells(2)}}
	AssertSnapshot(t, widget, 12, 2, "Focused text area in the hidden blink phase: two lines, no cursor.")
}

// A blink toggle repaints the focused input without a full render.
func TestRenderer_CursorBlinkRepaintsFocusedInput(t *testing.T) {
	enableCursorBlinkForTest(t)

	screen := newTrackingScreen(10, 1)
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)
	renderer := NewRenderer(screen, 10, 1, focusManager, focusedSignal, NewAnySignal[Widget](nil))
	widget := TextInput{ID: "input", State: NewTextInputState("hi"), Style: Style{Width: Cells(10)}}

	focusManager.focusedID = "input"
	focusedSignal.Set(widget)
	renderer.Update(widget)
	cursorCell := func() *uv.Cell { return screen.CellAt(2, 0) }
	require.NotZero(t, cursorCell().Style.Attrs&uv.AttrReverse, "cursor shown")

	screen.resetTouched()
	toggleCursorBlink()
	renderer.Update(widget)

	assert.Equal(t, rendererFramePartial, renderer.lastFrameMode)
	assert.NotEmpty(t, screen.touched)
	assert.Zero(t, cursorCell().Style.Attrs&uv.AttrReverse, "cursor hidden")
}
