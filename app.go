package terma

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
)

// ErrPanicked is returned by Run when the application panicked.
// The detailed panic message is printed to stderr.
var ErrPanicked = errors.New("terma: application panicked (see stderr for details)")

// appCancel holds the cancel function for the currently running app.
var appCancel func()

// appRenderer holds the current renderer for screen export.
var appRenderer *Renderer

// renderTrigger signals the event loop to re-render when a signal changes.
// Buffered with size 1 to avoid blocking signal setters.
var renderTrigger chan struct{}

const (
	clickChainTimeout = 500 * time.Millisecond
	defaultFPS        = 60
)

var terminalEnableSequences = []string{
	ansi.SetModeMouseNormal,
	ansi.SetModeMouseAnyEvent,
	ansi.SetModeMouseButtonEvent,
	ansi.SetModeMouseExtSgr,
}

var terminalDisableSequences = []string{
	ansi.ResetModeMouseNormal,
	ansi.ResetModeMouseAnyEvent,
	ansi.ResetModeMouseButtonEvent,
	ansi.ResetModeMouseExtSgr,
}

func writeTerminalSequences(writeString func(string) (int, error), sequences []string) {
	for _, seq := range sequences {
		_, _ = writeString(seq)
	}
}

func boolEnv(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	return v != "" && v != "0" && v != "false" && v != "no"
}

func kittyKeyboardDisabledByEnv() bool {
	return boolEnv("TERMA_DISABLE_KITTY_KEYBOARD")
}

func kittyKeyboardEnabledByEnv() bool {
	return boolEnv("TERMA_ENABLE_KITTY_KEYBOARD")
}

func enableTerminalInputModes(writeString func(string) (int, error), enableKittyKeyboard bool) {
	writeTerminalSequences(writeString, terminalEnableSequences)
	if enableKittyKeyboard {
		// Preserve any pre-existing Kitty keyboard state by using stack push.
		_, _ = writeString(ansi.PushKittyKeyboard(ansi.KittyAllFlags))
	}
}

func disableTerminalInputModes(writeString func(string) (int, error), enableKittyKeyboard bool, aggressive bool) {
	writeTerminalSequences(writeString, terminalDisableSequences)
	if enableKittyKeyboard {
		// Restore the previous Kitty keyboard state from the terminal stack.
		_, _ = writeString(ansi.PopKittyKeyboard(1))
	}
	if aggressive {
		// Optional hard reset path for terminals that still misbehave.
		extra := []string{
			ansi.ResetModeMouseX10,
			ansi.ResetModeMouseExtUtf8,
			ansi.ResetModeMouseExtUrxvt,
			ansi.ResetModeMouseExtSgrPixel,
			ansi.ResetModeBracketedPaste,
			ansi.ResetModeCursorKeys,
			ansi.ResetModeKeyboardAction,
			ansi.ResetModeInsertReplace,
			ansi.ResetModeNumericKeypad,
			ansi.KeypadNumericMode,
			ansi.ResetModeFocusEvent,
			"\x1b[>4;0m", // xterm modifyOtherKeys
		}
		writeTerminalSequences(writeString, extra)
		if enableKittyKeyboard {
			_, _ = writeString(ansi.KittyKeyboard(0, 1))
		}
	}
}

func snapshotTTYState(f *os.File) *term.State {
	if f == nil {
		return nil
	}
	fd := f.Fd()
	if !term.IsTerminal(fd) {
		return nil
	}
	state, err := term.GetState(fd)
	if err != nil {
		return nil
	}
	return state
}

func restoreTTYState(f *os.File, state *term.State) {
	if f == nil || state == nil {
		return
	}
	_ = term.Restore(f.Fd(), state)
}

type mouseClickTracker struct {
	lastClickTime time.Time
	lastTargetID  string
	lastButton    uv.MouseButton
	lastX, lastY  int
	clickCount    int

	lastDownTargetID string
	lastDownButton   uv.MouseButton
	lastDownCount    int
}

type mouseDragState struct {
	isDragging    bool
	dragWidgetID  string
	pressedButton uv.MouseButton
}

func (t *mouseClickTracker) nextClick(targetID string, button uv.MouseButton, x, y int, now time.Time) int {
	samePosition := targetID == t.lastTargetID && button == t.lastButton && x == t.lastX && y == t.lastY
	if samePosition && now.Sub(t.lastClickTime) <= clickChainTimeout {
		t.clickCount++
	} else {
		t.clickCount = 1
	}
	t.lastTargetID = targetID
	t.lastButton = button
	t.lastX = x
	t.lastY = y
	t.lastClickTime = now

	t.lastDownTargetID = targetID
	t.lastDownButton = button
	t.lastDownCount = t.clickCount

	return t.clickCount
}

func (t *mouseClickTracker) releaseCount(targetID string, button uv.MouseButton) int {
	if targetID == t.lastDownTargetID && (button == t.lastDownButton || button == uv.MouseNone) {
		return t.lastDownCount
	}
	return 1
}

func buildMouseEvent(m uv.Mouse, entry *WidgetEntry, clickCount int) MouseEvent {
	widgetID := ""
	localX, localY := m.X, m.Y
	if entry != nil {
		widgetID = entry.ID
		localX = m.X - entry.Bounds.X
		localY = m.Y - entry.Bounds.Y
	}
	return MouseEvent{
		X:          m.X,
		Y:          m.Y,
		LocalX:     localX,
		LocalY:     localY,
		Button:     m.Button,
		Mod:        m.Mod,
		ClickCount: clickCount,
		WidgetID:   widgetID,
	}
}

// Quit exits the running application gracefully.
// This performs the same teardown as pressing Ctrl+C.
func Quit() {
	if appCancel != nil {
		appCancel()
	}
}

// ScreenText returns the current screen content as plain text.
// Returns empty string if no app is running.
func ScreenText() string {
	if appRenderer == nil {
		return ""
	}
	return appRenderer.ScreenText()
}

// Run starts the application with the given root widget and blocks until it exits.
// The root widget can implement KeyHandler to receive key events that bubble up
// from focused descendants.
func Run(root Widget) (runErr error) {
	t := uv.DefaultTerminal()
	origStdinState := snapshotTTYState(os.Stdin)
	origStdoutState := snapshotTTYState(os.Stdout)
	restoreOriginalTTY := func() {
		restoreTTYState(os.Stdin, origStdinState)
		restoreTTYState(os.Stdout, origStdoutState)
	}

	if err := t.Start(); err != nil {
		return err
	}
	// Keep Kitty keyboard protocol enabled by default, but allow opt-out for
	// environments with buggy pager/terminal combinations.
	enableKittyKeyboard := !kittyKeyboardDisabledByEnv()
	if kittyKeyboardEnabledByEnv() {
		enableKittyKeyboard = true
	}

	t.EnterAltScreen()

	// Enable input reporting modes used by Terma (mouse + Kitty keyboard).
	enableTerminalInputModes(t.WriteString, enableKittyKeyboard)

	// shutdownTerminal restores the terminal to its normal state.
	// Safe to call multiple times (Shutdown is idempotent).
	shutdownTerminal := func() {
		// First, disable modes while the terminal session is still active.
		// Some emulators/shell multiplexer stacks can scope keyboard protocol
		// state to screen buffers, so doing this before shutdown is more
		// reliable than only restoring after shutdown.
		preRestoreDone := false
		disableTerminalInputModes(t.WriteString, enableKittyKeyboard, false)
		if err := t.Flush(); err == nil {
			preRestoreDone = true
		}

		shutdownErr := t.Shutdown(context.Background())
		aggressiveRestore := shutdownErr != nil || boolEnv("TERMA_AGGRESSIVE_TERMINAL_RESTORE")
		// Write restore sequences directly to stdout after Shutdown.
		//
		// When a panic occurs before the first Display() call, Shutdown's
		// restoreState has no recorded lastState, so it skips exiting alt
		// screen and showing the cursor. However, Shutdown DOES flush the
		// internal buffer which contains the enter-alt-screen and mouse-enable
		// sequences that were buffered during Start(). This leaves the
		// terminal stuck in alt screen with mouse tracking enabled.
		//
		// Writing these sequences directly to stdout (the output device used
		// by DefaultTerminal) ensures the terminal is fully restored. These
		// are idempotent — harmless if Shutdown already handled them.
		_, _ = os.Stdout.WriteString(ansi.ResetModeAltScreenSaveCursor)
		_, _ = os.Stdout.WriteString(ansi.SetModeTextCursorEnable)
		// If pre-shutdown restore succeeded, avoid a second Kitty pop on stdout.
		postRestoreKitty := enableKittyKeyboard && !preRestoreDone
		disableTerminalInputModes(os.Stdout.WriteString, postRestoreKitty, aggressiveRestore)
		if boolEnv("TERMA_FORCE_TERMINAL_RIS") {
			_, _ = os.Stdout.WriteString(ansi.ResetInitialState)
		}
		restoreOriginalTTY()
	}

	ctx, cancel := context.WithCancel(context.Background())
	appCancel = cancel

	// Create animation controller for this app
	animController := NewAnimationController(defaultFPS)
	currentController = animController

	// Create render trigger channel for signal-driven re-renders
	renderTrigger = make(chan struct{}, 1)

	// Track event loop goroutine so we can wait for it during shutdown.
	eventLoopDone := make(chan struct{})
	eventLoopStarted := false

	// Single cleanup defer handles both normal exit and panic recovery.
	// On panic: recover captures it, then we cancel context, wait for the
	// event loop goroutine to exit, and restore the terminal cleanly.
	// Without recovery, panics write to the alternate screen buffer which
	// is discarded on exit, making the error invisible.
	defer func() {
		if r := recover(); r != nil {
			recordPanic(Panic{
				Message:    fmt.Sprint(r),
				StackTrace: string(debug.Stack()),
			})
			runErr = ErrPanicked
		}

		cancel()
		if eventLoopStarted {
			<-eventLoopDone
		}

		appCancel = nil
		appRenderer = nil
		renderTrigger = nil
		currentController = nil
		animController.Stop()

		shutdownTerminal()
		renderPanics()
	}()

	// Get initial terminal size
	size := t.Size()
	width, height := size.Width, size.Height
	debugOverlayEnabled := os.Getenv("TERMA_DEBUG_OVERLAY") != ""
	if debugOverlayEnabled {
		EnableDebugRenderCause()
	}

	// Create focus manager and focused signal
	focusManager := NewFocusManager()
	focusManager.SetRootWidget(root)
	focusedSignal := NewAnySignal[Focusable](nil)
	lastFocusedID := ""

	// Create hovered widget signal (tracks the currently hovered widget)
	hoveredSignal := NewAnySignal[Widget](nil)

	// Create renderer with focus manager and signal
	renderer := NewRenderer(t, width, height, focusManager, focusedSignal, hoveredSignal)
	appRenderer = renderer

	updateFocusedSignal := func() bool {
		focusedID := focusManager.FocusedID()
		if focusedID == lastFocusedID {
			return false
		}
		lastFocusedID = focusedID
		focusedSignal.Set(focusManager.Focused())
		return true
	}

	var (
		coalescedRenderRequests int
		overrunFrames           int
		lastFrameDuration       time.Duration
		lastOverlayWidth        int
		lastCauseOverlayWidth   int
	)

	drawDebugOverlay := func() {
		if !debugOverlayEnabled || width <= 0 || height <= 0 {
			return
		}

		frameMs := float64(lastFrameDuration.Microseconds()) / 1000.0
		text := fmt.Sprintf("frame %.2fms coalesced %d overrun %d", frameMs, coalescedRenderRequests, overrunFrames)
		textWidth := ansi.StringWidth(text)
		if textWidth < lastOverlayWidth {
			text += strings.Repeat(" ", lastOverlayWidth-textWidth)
		} else {
			lastOverlayWidth = textWidth
		}

		cause := LastRenderCause()
		if cause == "" {
			cause = "(none)"
		}
		causeText := fmt.Sprintf("cause %s", cause)
		causeWidth := ansi.StringWidth(causeText)
		if causeWidth < lastCauseOverlayWidth {
			causeText += strings.Repeat(" ", lastCauseOverlayWidth-causeWidth)
		} else {
			lastCauseOverlayWidth = causeWidth
		}

		ctx := NewRenderContext(t, width, height, nil, nil, BuildContext{}, nil)
		ctx.DrawStyledText(0, 0, text, Style{
			ForegroundColor: BrightWhite,
			BackgroundColor: Black,
		})
		if height > 1 {
			ctx.DrawStyledText(0, 1, causeText, Style{
				ForegroundColor: BrightWhite,
				BackgroundColor: Black,
			})
		}
	}

	renderInterval := time.Second / time.Duration(defaultFPS)
	lastModalCount := 0

	// Render and update focusables
	display := func() {
		startTime := time.Now()
		screen.Clear(t)
		// Update the focused signal BEFORE render so widgets can read it
		updateFocusedSignal()

		focusables := renderer.Render(root)
		focusManager.SetFocusables(focusables)

		// If focus changed after render (auto-focus or focus removal), re-render
		if updateFocusedSignal() {
			renderer.Render(root)
		}

		// Manage modal focus transitions (open/close) and keep focus inside topmost modal.
		modalCount := renderer.ModalCount()
		openedModals := 0
		closedModals := 0
		if modalCount != lastModalCount {
			if modalCount > lastModalCount {
				openedModals = modalCount - lastModalCount
			} else {
				closedModals = lastModalCount - modalCount
			}
		}
		for i := 0; i < openedModals; i++ {
			focusManager.SaveFocus()
		}

		for i := 0; i < closedModals; i++ {
			focusManager.RestoreFocus()
		}

		// Apply pending focus request from ctx.RequestFocus() after modal
		// restore logic so explicit focus requests win.
		if pendingFocusID != "" {
			focusManager.FocusByID(pendingFocusID)
			pendingFocusID = ""
			// Update the signal and re-render so the focused widget shows focus style
			if updateFocusedSignal() {
				renderer.Render(root)
			}
		}

		lastModalCount = modalCount
		// Update the signal and re-render so the focused widget shows focus style
		if updateFocusedSignal() {
			renderer.Render(root)
		}
		// Position terminal cursor for IME support (emoji picker, input methods)
		// Must be before Display() since MoveTo only takes effect on next Display call
		if focusedID := focusManager.FocusedID(); focusedID != "" {
			if entry := renderer.WidgetByID(focusedID); entry != nil {
				if textInput, ok := entry.Widget.(TextInput); ok {
					cursorX := textInput.CursorScreenPosition(entry.Bounds.X)
					cursorY := entry.Bounds.Y
					t.MoveTo(cursorX, cursorY)
				}
			}
		}

		drawDebugOverlay()
		_ = t.Display()

		elapsed := time.Since(startTime)
		lastFrameDuration = elapsed
		if elapsed > renderInterval {
			overrunFrames++
		}

		Log("Render complete in %.3fms, %d widgets registered", float64(elapsed.Microseconds())/1000.0, len(renderer.widgetRegistry.entries))
	}

	clickTracker := &mouseClickTracker{}
	dragState := &mouseDragState{}
	currentHoveredID := ""

	resolveMouseTarget := func(x, y int, allowDismiss bool) (*WidgetEntry, bool) {
		// Check if click is on a float
		if renderer.FloatAt(x, y) != nil {
			return renderer.WidgetAt(x, y), false
		}

		// Click is outside all floats - check for dismissal or modal blocking
		if renderer.HasFloats() {
			if allowDismiss {
				topFloat := renderer.TopFloat()
				if topFloat != nil && topFloat.Config.shouldDismissOnClickOutside() && topFloat.Config.OnDismiss != nil {
					topFloat.Config.OnDismiss()
					return nil, true
				}
			}

			// For modal floats, block the click from reaching underlying widgets
			if renderer.HasModalFloat() {
				return nil, true
			}
		}

		return renderer.WidgetAt(x, y), false
	}

	// focusAt finds the innermost focusable widget at (x, y) and focuses it.
	// This is separate from WidgetAt because the clicked widget (for OnClick)
	// may be different from the focusable widget (e.g., clicking Text inside a List).
	focusAt := func(x, y int) {
		entry := renderer.FocusableAt(x, y)
		if entry != nil {
			focusManager.FocusByID(entry.ID)
		}
	}

	// Get root's key handling interfaces (if any) for the no-focusables case
	rootHandler, _ := root.(KeyHandler)
	rootKeybindProvider, _ := root.(KeybindProvider)

	var (
		lastRender    time.Time
		renderPending bool
		renderTimer   *time.Timer
		renderTimerCh <-chan time.Time
	)

	stopRenderTimer := func() {
		if renderTimer == nil {
			renderTimerCh = nil
			return
		}
		if !renderTimer.Stop() {
			select {
			case <-renderTimer.C:
			default:
			}
		}
		renderTimerCh = nil
	}

	renderNow := func() {
		stopRenderTimer()
		renderPending = false
		display()
		lastRender = time.Now()
	}

	requestRender := func() {
		now := time.Now()
		if lastRender.IsZero() || now.Sub(lastRender) >= renderInterval {
			renderNow()
			return
		}
		if renderPending {
			coalescedRenderRequests++
			return
		}
		renderPending = true
		wait := renderInterval - now.Sub(lastRender)
		if renderTimer == nil {
			renderTimer = time.NewTimer(wait)
		} else {
			if !renderTimer.Stop() {
				select {
				case <-renderTimer.C:
				default:
				}
			}
			renderTimer.Reset(wait)
		}
		renderTimerCh = renderTimer.C
	}

	// Initial render
	renderNow()

	// Event loop
	eventLoopStarted = true
	go func() {
		defer close(eventLoopDone)
		defer func() {
			if r := recover(); r != nil {
				recordPanic(Panic{
					Message:    fmt.Sprint(r),
					StackTrace: string(debug.Stack()),
				})
				cancel()
			}
		}()
		termEvents := t.Events()
		for {
			select {
			case <-ctx.Done():
				return
			case <-renderTrigger:
				requestRender()
			case <-animController.Tick():
				animController.Update()
				requestRender()
			case <-renderTimerCh:
				if renderPending {
					renderNow()
				}
			case ev, ok := <-termEvents:
				if !ok {
					return
				}
				switch ev := ev.(type) {
				case uv.WindowSizeEvent:
					_ = t.Resize(ev.Width, ev.Height)
					renderer.Resize(ev.Width, ev.Height)
					width = ev.Width
					height = ev.Height
					t.Erase()
					requestRender()
				case uv.KeyPressEvent:
					// Check for app-level quit keys
					if ev.MatchString("ctrl+c") {
						cancel()
						return
					}

					// Screen export keybind
					if ev.MatchString("ctrl+shift+s") {
						exportScreenToFile()
						continue
					}

					// Suspend on Ctrl+Z
					if ev.MatchString("ctrl+z") {
						// Disable input reporting modes before suspending so
						// the shell gets plain keyboard input while suspended.
						disableTerminalInputModes(t.WriteString, enableKittyKeyboard, false)

						// Exit alternate screen to show shell
						t.ExitAltScreen()

						// Pause input reading and suspend process
						_ = t.Pause()
						_ = uv.Suspend() // Blocks until resumed via `fg`

						// Resume input reading
						_ = t.Resume()

						// Re-enter alternate screen
						t.EnterAltScreen()

						// Re-enable mouse tracking
						enableTerminalInputModes(t.WriteString, enableKittyKeyboard)

						// Redraw the screen
						requestRender()
						continue
					}

					// Check for Escape to dismiss floats
					if ev.MatchString("escape") {
						if topFloat := renderer.TopFloat(); topFloat != nil {
							if topFloat.Config.shouldDismissOnEsc() && topFloat.Config.OnDismiss != nil {
								topFloat.Config.OnDismiss()
								requestRender()
								continue
							}
						}
					}

					// Route key event through focus manager (bubbles through widget tree)
					keyEvent := KeyEvent{event: ev}
					handled := focusManager.HandleKey(keyEvent)

					// If not handled, try root's keybindings and handler directly
					// (handles case when there are no focusable widgets)
					if !handled {
						if rootKeybindProvider != nil {
							handled = matchKeybind(keyEvent, rootKeybindProvider.Keybinds())
						}
						if !handled && rootHandler != nil {
							rootHandler.OnKey(keyEvent)
						}
					}

					// Re-render after key press (for signal updates and focus changes)
					requestRender()

				case uv.MouseClickEvent:
					Log("MouseClickEvent at X=%d Y=%d Button=%v", ev.X, ev.Y, ev.Button)

					entry, handled := resolveMouseTarget(ev.X, ev.Y, true)
					if handled {
						Log("  Mouse click handled by float logic")
						requestRender()
						continue
					}

					if entry != nil {
						Log("  Found widget: ID=%q Type=%T", entry.ID, entry.EventWidget)
						focusAt(ev.X, ev.Y)
						clickCount := clickTracker.nextClick(entry.ID, ev.Button, ev.X, ev.Y, time.Now())
						mouseEvent := buildMouseEvent(uv.Mouse(ev), entry, clickCount)

						// Set drag state for mouse move tracking
						dragState.isDragging = true
						dragState.dragWidgetID = entry.ID
						dragState.pressedButton = ev.Button

						if downHandler, ok := entry.EventWidget.(MouseDownHandler); ok {
							Log("  Widget has OnMouseDown")
							downHandler.OnMouseDown(mouseEvent)
						}

						if clickable, ok := entry.EventWidget.(Clickable); ok {
							Log("  Widget is Clickable, calling OnClick")
							clickable.OnClick(mouseEvent)
						} else {
							Log("  Widget is NOT Clickable")
						}
					} else {
						Log("  No widget found at position")
						LogWidgetRegistry(renderer.widgetRegistry)
					}

					// Re-render after click
					requestRender()

				case uv.MouseReleaseEvent:
					Log("MouseReleaseEvent at X=%d Y=%d Button=%v", ev.X, ev.Y, ev.Button)

					// Clear drag state
					dragState.isDragging = false
					dragState.dragWidgetID = ""
					dragState.pressedButton = uv.MouseNone

					entry, handled := resolveMouseTarget(ev.X, ev.Y, false)
					if handled {
						Log("  Mouse release blocked by float logic")
						requestRender()
						continue
					}

					if entry != nil {
						Log("  Found widget: ID=%q Type=%T", entry.ID, entry.EventWidget)
						clickCount := clickTracker.releaseCount(entry.ID, ev.Button)
						mouseEvent := buildMouseEvent(uv.Mouse(ev), entry, clickCount)

						if upHandler, ok := entry.EventWidget.(MouseUpHandler); ok {
							Log("  Widget has OnMouseUp")
							upHandler.OnMouseUp(mouseEvent)
						}
					} else {
						Log("  No widget found at position")
					}

					// Re-render after mouse up
					requestRender()

				case uv.MouseMotionEvent:
					// Log("MouseMotionEvent at X=%d Y=%d", ev.X, ev.Y)

					// Handle drag - dispatch to the widget that received the mouse down
					if dragState.isDragging && dragState.dragWidgetID != "" {
						if dragEntry := renderer.WidgetByID(dragState.dragWidgetID); dragEntry != nil {
							if moveHandler, ok := dragEntry.EventWidget.(MouseMoveHandler); ok {
								// Build mouse event with local coordinates relative to the drag widget
								localX := ev.X - dragEntry.Bounds.X
								localY := ev.Y - dragEntry.Bounds.Y
								mouseEvent := MouseEvent{
									X:          ev.X,
									Y:          ev.Y,
									LocalX:     localX,
									LocalY:     localY,
									Button:     dragState.pressedButton,
									Mod:        ev.Mod,
									ClickCount: 1,
									WidgetID:   dragEntry.ID,
								}
								moveHandler.OnMouseMove(mouseEvent)
								display()
							}
						}
					}

					// Find the widget under the cursor
					entry := renderer.WidgetAt(ev.X, ev.Y)
					var newHovered Widget
					newHoveredID := ""
					if entry != nil {
						newHovered = entry.EventWidget
						newHoveredID = entry.ID
					}

					// Only update if hover changed (compare by ID to avoid incomparable type issues)
					if newHoveredID != currentHoveredID {
						Log("  Hover changed: %q -> %q", currentHoveredID, newHoveredID)

						// Notify old widget it's no longer hovered
						oldHovered := hoveredSignal.Get()
						if oldHovered != nil {
							if hoverable, ok := oldHovered.(Hoverable); ok {
								Log("  Calling OnHover(false) on %q", currentHoveredID)
								hoverable.OnHover(false)
							}
						}

						// Update the hovered signal
						hoveredSignal.Set(newHovered)
						currentHoveredID = newHoveredID

						// Notify new widget it's now hovered
						if entry != nil {
							if hoverable, ok := entry.EventWidget.(Hoverable); ok {
								Log("  Calling OnHover(true) on %q", newHoveredID)
								hoverable.OnHover(true)
							}
						}

						// Re-render after hover change
						requestRender()
					}

				case uv.MouseWheelEvent:
					dispatchMouseWheel(renderer, ev.X, ev.Y, ev.Button)
					requestRender()

				default:
					// Log other event types for debugging
					Log("Unhandled event: %T %v", ev, ev)
				}
			}
		}
	}()

	<-ctx.Done()
	return runErr
}

// dispatchMouseWheel routes wheel events to scrollable widgets under the cursor.
// Scrollables are tried from innermost to outermost until one handles the event.
func dispatchMouseWheel(renderer *Renderer, x int, y int, button uv.MouseButton) bool {
	if renderer == nil {
		return false
	}
	for _, scrollable := range renderer.ScrollablesAt(x, y) {
		var handled bool
		switch button {
		case uv.MouseWheelUp:
			handled = scrollable.ScrollUp(1)
		case uv.MouseWheelDown:
			handled = scrollable.ScrollDown(1)
		case uv.MouseWheelLeft:
			handled = scrollable.ScrollLeft(1)
		case uv.MouseWheelRight:
			handled = scrollable.ScrollRight(1)
		}
		if handled {
			return true
		}
	}
	return false
}

// exportScreenToFile saves the current screen content to a timestamped file.
func exportScreenToFile() {
	if appRenderer == nil {
		return
	}

	text := appRenderer.ScreenText()
	timestamp := time.Now().Format("20060102-150405")
	filename := fmt.Sprintf("terma-screenshot-%s.txt", timestamp)

	if err := os.WriteFile(filename, []byte(text), 0644); err != nil {
		Log("Screen export failed: %v", err)
		return
	}

	Log("Screen exported to %s", filename)
}
