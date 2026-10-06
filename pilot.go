package terma

import (
	"context"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
)

// Pilot drives an app in a test the way a user would: it presses keys, types,
// pastes, clicks and resizes, then lets you read the screen, the focused
// widget or the clipboard, or compare the screen against a snapshot.
//
// A Pilot runs the app on the same retained renderer, focus manager and event
// routing that [Run] uses, with an in-memory screen in place of a terminal.
// After every input it draws frames until the app settles: work queued with
// [Dispatch], signal updates, focus requests and modal focus all apply before
// the next call returns.
//
// Time is frozen. Animations, spinners, shimmers and the cursor blink move
// only when [Pilot.Advance] moves the clock, so frames are deterministic.
// Background goroutines still run for real; use [Pilot.WaitUntil] to wait
// for the state they produce.
//
// A Pilot stands in for the running app, so only one can exist at a time and
// tests that use one must not call t.Parallel.
type Pilot struct {
	t       testing.TB
	session *appSession
	buf     *uv.Buffer
	width   int
	height  int

	now       time.Time
	nextBlink time.Time
	animation *AnimationController
	trigger   chan struct{}
	cancel    context.CancelFunc
	exited    bool

	// heldButton is the mouse button down between MouseDown and MouseUp, so
	// MouseMove reports a drag.
	heldButton uv.MouseButton
	clipboard  map[ClipboardSelection]string
}

// pilotEpoch is where a Pilot's clock starts.
var pilotEpoch = time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

// pilotFrameInterval is how far Advance moves the clock per frame while
// animations run, matching the frame rate of a running app.
const pilotFrameInterval = time.Second / defaultFPS

// maxPilotSettleFrames bounds how many frames one input may take to settle.
const maxPilotSettleFrames = 32

// NewPilot starts root on a width×height screen and draws its first frame.
// The app stops when the test ends.
func NewPilot(t testing.TB, root Widget, width, height int) *Pilot {
	t.Helper()
	if currentAppContext() != nil {
		t.Fatalf("terma: NewPilot needs the app runtime to itself; another app or Pilot is running")
	}

	p := &Pilot{
		t:         t,
		buf:       uv.NewBuffer(width, height),
		width:     width,
		height:    height,
		now:       pilotEpoch,
		trigger:   make(chan struct{}, 1),
		clipboard: map[ClipboardSelection]string{},
	}
	p.nextBlink = p.now.Add(cursorBlinkInterval)
	p.session = newAppSession(root, p.buf, width, height, p.Now)
	p.session.onInput = func() {
		showCursorForInput()
		p.nextBlink = p.now.Add(cursorBlinkInterval)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.animation = NewAnimationController(defaultFPS)
	p.animation.now = p.Now

	previousTrigger := swapRenderTrigger(p.trigger)
	previousController := currentController
	previousFocusID := pendingFocusID
	pendingFocusID = ""
	takeTerminalWrites()
	resetClipboardReads()
	setAppRuntimeState(ctx, newDispatchQueue())
	appCancel = cancel
	appRenderer = p.session.renderer
	currentController = p.animation

	t.Cleanup(func() {
		cancel()
		p.animation.Stop()
		appCancel = nil
		appRenderer = nil
		currentController = previousController
		clearAppRuntimeState()
		swapRenderTrigger(previousTrigger)
		takeTerminalWrites()
		resetClipboardReads()
		pendingFocusID = previousFocusID
	})

	p.settle()
	return p
}

// Press presses each key in turn. Keys are spelled as in [Keybind]: "tab",
// "enter", "ctrl+s", "shift+down", "q", "?".
func (p *Pilot) Press(keys ...string) {
	p.t.Helper()
	for _, key := range keys {
		ev, err := keyPress(key)
		if err != nil {
			p.t.Fatalf("terma: Pilot.Press: %v", err)
		}
		p.Send(ev)
	}
}

// Type types text one character at a time, as key presses. Use [Pilot.Paste]
// to deliver text in one piece.
func (p *Pilot) Type(text string) {
	p.t.Helper()
	for _, r := range text {
		p.Send(typedKey(r))
	}
}

// Paste pastes text as the terminal does with bracketed paste.
func (p *Pilot) Paste(text string) {
	p.t.Helper()
	p.Send(uv.PasteEvent{Content: text})
}

// Click clicks the left mouse button in the middle of the visible part of the
// widget with the given ID. The click is hit tested like a real one, so a
// widget covered by an overlay doesn't receive it.
func (p *Pilot) Click(id string) {
	p.t.Helper()
	x, y := p.center(id)
	p.ClickAt(x, y)
}

// ClickAt presses and releases the left mouse button at screen cell (x, y).
// Clicks on the same cell form a double (or triple) click unless the clock
// moves on by more than half a second between them.
func (p *Pilot) ClickAt(x, y int) {
	p.t.Helper()
	p.MouseDown(x, y, uv.MouseLeft, 0)
	p.MouseUp(x, y, uv.MouseLeft, 0)
}

// MouseDown presses a mouse button at (x, y) with the given modifiers held.
func (p *Pilot) MouseDown(x, y int, button uv.MouseButton, mod uv.KeyMod) {
	p.t.Helper()
	p.heldButton = button
	p.Send(uv.MouseClickEvent{X: x, Y: y, Button: button, Mod: mod})
}

// MouseUp releases a mouse button at (x, y).
func (p *Pilot) MouseUp(x, y int, button uv.MouseButton, mod uv.KeyMod) {
	p.t.Helper()
	p.heldButton = uv.MouseNone
	p.Send(uv.MouseReleaseEvent{X: x, Y: y, Button: button, Mod: mod})
}

// MouseMove moves the pointer to (x, y). Between MouseDown and MouseUp this
// drags.
func (p *Pilot) MouseMove(x, y int) {
	p.t.Helper()
	p.Send(uv.MouseMotionEvent{X: x, Y: y, Button: p.heldButton})
}

// Hover moves the pointer over the middle of the widget with the given ID.
func (p *Pilot) Hover(id string) {
	p.t.Helper()
	x, y := p.center(id)
	p.MouseMove(x, y)
}

// Scroll turns the mouse wheel at (x, y) by lines notches: positive scrolls
// down, negative up.
func (p *Pilot) Scroll(x, y, lines int) {
	p.t.Helper()
	button := uv.MouseWheelDown
	if lines < 0 {
		button, lines = uv.MouseWheelUp, -lines
	}
	for range lines {
		p.Send(uv.MouseWheelEvent{X: x, Y: y, Button: button})
	}
}

// Resize resizes the screen, as when the terminal window changes size.
func (p *Pilot) Resize(width, height int) {
	p.t.Helper()
	p.Send(uv.WindowSizeEvent{Width: width, Height: height})
}

// Send delivers raw terminal events (ultraviolet key, mouse, paste and window
// size events) through the app's event routing, settling after each. The
// helpers above cover common input; Send is for the rest, such as a click
// with modifiers or a key with unusual fields.
func (p *Pilot) Send(events ...uv.Event) {
	p.t.Helper()
	for _, ev := range events {
		if p.exited {
			p.t.Fatalf("terma: Pilot: the app has quit; can't send %T", ev)
		}
		if size, ok := ev.(uv.WindowSizeEvent); ok {
			p.width, p.height = size.Width, size.Height
			p.session.renderer.Resize(size.Width, size.Height)
			// Run erases the terminal on resize; the full frame that follows
			// paints only what widgets draw.
			p.buf.Clear()
		} else if p.session.handle(ev, 0.5, 0.5) == inputQuit {
			p.cancel()
		}
		p.settle()
	}
}

// Advance moves the clock on by d. Running animations step a frame at a time
// at the app's frame rate, and the cursor blinks if blinking is on.
func (p *Pilot) Advance(d time.Duration) {
	p.t.Helper()
	end := p.now.Add(d)
	for p.now.Before(end) {
		step := end.Sub(p.now)
		if p.animation.HasActiveAnimations() {
			step = min(step, pilotFrameInterval)
		}
		if blink := p.nextBlink.Sub(p.now); CursorBlink() && blink > 0 {
			step = min(step, blink)
		}
		p.now = p.now.Add(step)
		p.animation.Update()
		for !p.nextBlink.After(p.now) {
			toggleCursorBlink()
			p.nextBlink = p.nextBlink.Add(cursorBlinkInterval)
		}
		p.settle()
	}
}

// WaitUntil draws frames as work arrives from other goroutines (signal
// updates, [Dispatch]) until cond reports true. It fails the test if cond
// still doesn't hold after timeout of real time (one second if zero).
func (p *Pilot) WaitUntil(cond func() bool, timeout time.Duration) {
	p.t.Helper()
	if timeout <= 0 {
		timeout = time.Second
	}
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	for {
		p.settle()
		if cond() {
			return
		}
		select {
		case <-p.trigger:
			// settle redraws for it on the next pass; put it back so it
			// counts as pending work.
			scheduleRender()
		case <-deadline.C:
			p.t.Fatalf("terma: Pilot.WaitUntil: condition not met within %v", timeout)
		}
	}
}

// Now is the Pilot's clock. It starts at a fixed time and moves only with
// [Pilot.Advance].
func (p *Pilot) Now() time.Time {
	return p.now
}

// ScreenText is the screen as plain text, one line per row.
func (p *Pilot) ScreenText() string {
	p.t.Helper()
	p.settle()
	return p.session.renderer.ScreenText()
}

// TextOf is the plain text in the visible part of the widget with the given
// ID, one line per row with trailing spaces removed.
func (p *Pilot) TextOf(id string) string {
	p.t.Helper()
	p.settle()
	area := p.visible(id)
	rows := make([]string, 0, area.Height)
	for y := area.Y; y < area.Y+area.Height; y++ {
		var row strings.Builder
		for x := area.X; x < area.X+area.Width; {
			cell := p.buf.CellAt(x, y)
			if cell == nil || cell.Content == "" {
				row.WriteByte(' ')
				x++
				continue
			}
			row.WriteString(cell.Content)
			x += max(1, cell.Width)
		}
		rows = append(rows, strings.TrimRight(row.String(), " "))
	}
	return strings.Join(rows, "\n")
}

// Bounds is the screen area of the widget with the given ID, and whether a
// widget with that ID was drawn.
func (p *Pilot) Bounds(id string) (Rect, bool) {
	p.t.Helper()
	p.settle()
	entry := p.session.renderer.WidgetByID(id)
	if entry == nil {
		return Rect{}, false
	}
	return entry.Bounds, true
}

// FocusedID is the ID of the focused widget, or "" if nothing is focused.
func (p *Pilot) FocusedID() string {
	p.t.Helper()
	p.settle()
	return p.session.focus.FocusedID()
}

// Keybinds are the keybinds active for the focused widget and its ancestors,
// as a [KeybindBar] would show them (hidden ones included).
func (p *Pilot) Keybinds() []Keybind {
	p.t.Helper()
	p.settle()
	return p.session.focus.ActiveKeybinds()
}

// Clipboard is what the app last put on the system clipboard with
// [SetClipboard]. [ReadClipboard] reads it back.
func (p *Pilot) Clipboard() string {
	p.t.Helper()
	p.settle()
	return p.clipboard[SystemClipboard]
}

// Exited reports whether the app has quit, through [Quit] or ctrl+c.
func (p *Pilot) Exited() bool {
	p.t.Helper()
	p.settle()
	return p.exited
}

// Buffer is the screen's cells, for checks the other methods don't cover.
func (p *Pilot) Buffer() *uv.Buffer {
	p.t.Helper()
	p.settle()
	return p.buf
}

// AssertSnapshot compares the screen with the golden file
// testdata/<TestName>_<name>.svg, as [AssertSnapshot] does for a widget.
// Run the tests with UPDATE_SNAPSHOTS=1 to write it.
func (p *Pilot) AssertSnapshot(name string, description ...string) {
	p.t.Helper()
	p.settle()
	t, ok := p.t.(*testing.T)
	if !ok {
		p.t.Fatalf("terma: Pilot.AssertSnapshot needs a *testing.T")
	}
	assertBufferSnapshot(t, t.Name()+"_"+name, p.buf, p.width, p.height, DefaultSVGOptions(), strings.Join(description, " "))
}

// settle draws frames until no work is left: nothing dispatched, no signal
// changed and no clipboard reply due.
func (p *Pilot) settle() {
	p.t.Helper()
	if p.exited {
		return
	}
	for range maxPilotSettleFrames {
		if ctx := currentAppContext(); ctx == nil || ctx.Err() != nil {
			p.exited = true
			return
		}
		select {
		case <-p.trigger:
		default:
		}
		p.session.frame()
		replies := p.terminalWrites()
		for _, reply := range replies {
			p.session.handle(uv.ClipboardEvent{Content: reply}, 0, 0)
		}
		select {
		case <-p.trigger:
			scheduleRender()
			continue
		default:
		}
		if len(replies) == 0 {
			return
		}
	}
	p.t.Fatalf("terma: Pilot: the app was still changing after %d frames; something updates state on every frame", maxPilotSettleFrames)
}

// terminalWrites plays the terminal's part for sequences the app wrote this
// frame: it keeps OSC 52 clipboard writes and answers clipboard reads,
// returning the replies.
func (p *Pilot) terminalWrites() []string {
	var replies []string
	for _, seq := range takeTerminalWrites() {
		body, ok := strings.CutPrefix(seq, "\x1b]52;")
		if !ok || len(body) < 2 {
			continue
		}
		selection := ClipboardSelection(body[0])
		data := strings.TrimSuffix(strings.TrimSuffix(body[2:], "\x07"), "\x1b\\")
		if data == "?" {
			replies = append(replies, p.clipboard[selection])
			continue
		}
		decoded, err := base64.StdEncoding.DecodeString(data)
		if err != nil {
			continue
		}
		p.clipboard[selection] = string(decoded)
	}
	return replies
}

func (p *Pilot) visible(id string) Rect {
	p.t.Helper()
	entry := p.session.renderer.WidgetByID(id)
	if entry == nil {
		p.t.Fatalf("terma: Pilot: no widget with ID %q on screen", id)
	}
	if entry.Visible.Width <= 0 || entry.Visible.Height <= 0 {
		p.t.Fatalf("terma: Pilot: widget %q is scrolled or clipped out of view", id)
	}
	return entry.Visible
}

func (p *Pilot) center(id string) (int, int) {
	p.t.Helper()
	p.settle()
	area := p.visible(id)
	return area.X + area.Width/2, area.Y + area.Height/2
}
