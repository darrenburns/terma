package terma

import (
	"sync/atomic"
	"time"
)

// cursorBlinkInterval is how long the text cursor stays shown, then hidden,
// while it blinks.
const cursorBlinkInterval = 530 * time.Millisecond

var (
	cursorBlinkEnabled atomic.Bool
	// cursorBlinkShown is the blink phase. Focused text widgets read it while
	// they paint, so each toggle repaints only them.
	cursorBlinkShown = NewSignal(true)
)

// SetCursorBlink turns blinking of the text cursor in TextInput and TextArea
// on or off for the whole app. It is off by default. While the cursor
// blinks, typing, pasting and clicking show it and restart the blink.
func SetCursorBlink(enabled bool) {
	cursorBlinkEnabled.Store(enabled)
	if !enabled {
		cursorBlinkShown.Set(true)
	}
}

// CursorBlink reports whether the text cursor blinks. See [SetCursorBlink].
func CursorBlink() bool {
	return cursorBlinkEnabled.Load()
}

// cursorVisible reports whether a focused text widget should draw its cursor
// now. It subscribes the caller to the blink phase when blinking is on.
func cursorVisible() bool {
	if !cursorBlinkEnabled.Load() {
		return true
	}
	return cursorBlinkShown.Get()
}

// showCursorForInput shows the cursor after user input, so it doesn't vanish
// while someone types.
func showCursorForInput() {
	if cursorBlinkEnabled.Load() {
		cursorBlinkShown.Set(true)
	}
}

// toggleCursorBlink flips the blink phase.
func toggleCursorBlink() {
	if cursorBlinkEnabled.Load() {
		cursorBlinkShown.Set(!cursorBlinkShown.Peek())
	}
}
