package terma

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
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
var renderTriggerMu sync.RWMutex

func currentRenderTrigger() chan struct{} {
	renderTriggerMu.RLock()
	defer renderTriggerMu.RUnlock()
	return renderTrigger
}

func swapRenderTrigger(next chan struct{}) chan struct{} {
	renderTriggerMu.Lock()
	defer renderTriggerMu.Unlock()
	previous := renderTrigger
	renderTrigger = next
	return previous
}

const defaultFPS = 60

var mouseEnableSequences = []string{
	ansi.SetModeMouseNormal,
	ansi.SetModeMouseButtonEvent,
	// Some terminals treat 1002/1003 as mutually exclusive tracking modes.
	// Enable AnyEvent (1003) after ButtonEvent (1002) so plain hover motion
	// is reported even when no mouse button is pressed.
	ansi.SetModeMouseAnyEvent,
	ansi.SetModeMouseExtSgr,
}

var terminalEnableSequences = []string{
	// Pastes arrive as one PasteEvent instead of a burst of key presses.
	ansi.SetModeBracketedPaste,
}

var terminalDisableSequences = []string{
	ansi.ResetModeMouseAnyEvent,
	ansi.ResetModeMouseButtonEvent,
	ansi.ResetModeMouseNormal,
	ansi.ResetModeMouseExtSgr,
	ansi.ResetModeBracketedPaste,
	// Switched on later if the terminal supports it (see pixelPointer).
	ansi.ResetModeMouseExtSgrPixel,
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

// resolveKittyKeyboardMode decides whether Kitty keyboard protocol should be
// enabled or force-disabled for this app session.
//
// Default is force-disabled to avoid duplicate/synthetic keypress behavior on
// some terminal stacks. TERMA_ENABLE_KITTY_KEYBOARD opts in and takes
// precedence when both env vars are set.
func resolveKittyKeyboardMode() (enableKittyKeyboard bool, forceDisableKittyKeyboard bool) {
	kittyDisabledByEnv := kittyKeyboardDisabledByEnv()
	kittyEnabledByEnv := kittyKeyboardEnabledByEnv()

	enableKittyKeyboard = kittyEnabledByEnv
	if kittyDisabledByEnv && !kittyEnabledByEnv {
		enableKittyKeyboard = false
	}

	forceDisableKittyKeyboard = !enableKittyKeyboard
	return
}

func enableTerminalInputModes(writeString func(string) (int, error), mouse bool, enableKittyKeyboard bool, forceDisableKittyKeyboard bool) {
	if mouse {
		writeTerminalSequences(writeString, mouseEnableSequences)
	}
	writeTerminalSequences(writeString, terminalEnableSequences)
	if enableKittyKeyboard {
		// Preserve any pre-existing Kitty keyboard state by using stack push.
		_, _ = writeString(ansi.PushKittyKeyboard(ansi.KittyAllFlags))
	} else if forceDisableKittyKeyboard {
		// Explicit opt-out should override any pre-existing Kitty keyboard state
		// for this app session. Push 0 so cleanup can safely restore prior state.
		_, _ = writeString(ansi.PushKittyKeyboard(0))
	}
}

func disableTerminalInputModes(writeString func(string) (int, error), enableKittyKeyboard bool, forceDisableKittyKeyboard bool, aggressive bool) {
	writeTerminalSequences(writeString, terminalDisableSequences)
	if enableKittyKeyboard || forceDisableKittyKeyboard {
		// Restore the previous Kitty keyboard state from the terminal stack.
		_, _ = writeString(ansi.PopKittyKeyboard(1))
	}
	if aggressive {
		// Optional hard reset path for terminals that still misbehave.
		extra := []string{
			ansi.ResetModeMouseX10,
			ansi.ResetModeMouseExtUtf8,
			ansi.ResetModeMouseExtUrxvt,
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

// startTerminal starts the terminal with hard-tab cursor movement disabled on
// the given ttys. See disableHardTabs.
func startTerminal(t *uv.Terminal, ttys ...*os.File) error {
	restores := make([]func(), 0, len(ttys))
	for _, f := range ttys {
		restores = append(restores, disableHardTabs(f))
	}
	err := t.Start()
	for i := len(restores) - 1; i >= 0; i-- {
		restores[i]()
	}
	return err
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

// CurrentRenderStats returns debug information for the most recent frame.
// Returns zero values if no app is currently running.
func CurrentRenderStats() RenderStats {
	if appRenderer == nil {
		return RenderStats{}
	}
	return appRenderer.Stats()
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

	if err := startTerminal(t, os.Stdin, os.Stdout); err != nil {
		restoreOriginalTTY()
		return err
	}
	// Keep Kitty keyboard protocol disabled by default, but allow explicit opt-in.
	enableKittyKeyboard, forceDisableKittyKeyboard := resolveKittyKeyboardMode()

	t.EnterAltScreen()

	// Enable input reporting modes used by Terma (mouse + Kitty keyboard).
	enableTerminalInputModes(t.WriteString, true, enableKittyKeyboard, forceDisableKittyKeyboard)
	// Ask whether the mouse can be reported in pixels; see pixelPointer.
	pointer := newPixelPointer()
	windowSize := uv.NewSizeNotifier(os.Stdout)
	pointer.readWindow = func() (windowGeometry, bool) {
		cells, pixels, err := windowSize.GetWindowSize()
		return windowGeometry{cells.Width, cells.Height, pixels.Width, pixels.Height}, err == nil
	}
	_, _ = t.WriteString(pointer.query())
	// Reassembles mouse reports the decoder splits; see sgrMouseRepair.
	var sgrRepair sgrMouseRepair
	images := newImageSession(t, pointer)

	// awaitCellSizeReply waits a little for the reply to a cell size query
	// that is still out, so it isn't left for the shell (or the program run
	// next) to read as typing. Other input read meanwhile is dropped: it was
	// sent as the app was being left.
	awaitCellSizeReply := func() {
		if !pointer.replyDue() {
			return
		}
		timeout := time.NewTimer(cellSizeReplyTimeout)
		defer timeout.Stop()
		for pointer.replyDue() {
			select {
			case ev, ok := <-t.Events():
				if !ok {
					return
				}
				if reply, isReply := ev.(uv.CellSizeEvent); isReply {
					pointer.recordCellSize(reply)
				}
			case <-timeout.C:
				return
			}
		}
	}

	// shutdownTerminal restores the terminal to its normal state.
	// Safe to call multiple times (Shutdown is idempotent).
	shutdownTerminal := func() {
		awaitCellSizeReply()
		images.close(t)
		// First, disable modes while the terminal session is still active.
		// Some emulators/shell multiplexer stacks can scope keyboard protocol
		// state to screen buffers, so doing this before shutdown is more
		// reliable than only restoring after shutdown.
		preRestoreDone := false
		disableTerminalInputModes(t.WriteString, enableKittyKeyboard, forceDisableKittyKeyboard, false)
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
		postRestoreKitty := (enableKittyKeyboard || forceDisableKittyKeyboard) && !preRestoreDone
		disableTerminalInputModes(os.Stdout.WriteString, postRestoreKitty, false, aggressiveRestore)
		if boolEnv("TERMA_FORCE_TERMINAL_RIS") {
			_, _ = os.Stdout.WriteString(ansi.ResetInitialState)
		}
		restoreOriginalTTY()
	}

	ctx, cancel := context.WithCancel(context.Background())
	appCancel = cancel
	setAppRuntimeState(ctx, newDispatchQueue())

	// Create animation controller for this app
	animController := NewAnimationController(defaultFPS)
	currentController = animController

	// Create render trigger channel for signal-driven re-renders
	swapRenderTrigger(make(chan struct{}, 1))

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
		setSuspender(nil)
		takeTerminalWrites()
		resetClipboardReads()
		swapRenderTrigger(nil)
		currentController = nil
		clearAppRuntimeState()
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
	pointer.setWindow(windowGeometry{cols: width, rows: height})
	images.geometry(renderer)
	appRenderer = renderer

	updateFocusedSignal := func() bool {
		keybindsChanged := focusManager.syncKeybinds()
		focusedID := focusManager.FocusedID()
		if focusedID == lastFocusedID {
			return keybindsChanged
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
		lastStatsOverlayWidth   int
		lastDamageOverlayWidth  int
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

		stats := renderer.Stats()
		mode := stats.FrameMode
		if mode == "" {
			mode = "(none)"
		}
		modeLabel := mode
		switch mode {
		case string(rendererFramePartial):
			modeLabel = "partial repaint"
		case string(rendererFrameFull):
			modeLabel = "full render"
		case string(rendererFrameReflow):
			modeLabel = "relayout + partial repaint"
		}
		statsText := fmt.Sprintf(
			"last frame: %s | rebuilt %d | relaid out %d | repainted %d",
			modeLabel,
			stats.BuildCount,
			stats.LayoutCount,
			stats.PaintCount,
		)
		statsWidth := ansi.StringWidth(statsText)
		if statsWidth < lastStatsOverlayWidth {
			statsText += strings.Repeat(" ", lastStatsOverlayWidth-statsWidth)
		} else {
			lastStatsOverlayWidth = statsWidth
		}

		damageText := "repaint area: none"
		if len(stats.DamagedRects) > 0 {
			damageUnion := stats.DamagedRects[0]
			for _, rect := range stats.DamagedRects[1:] {
				damageUnion = damageUnion.Union(rect)
			}
			damageText = fmt.Sprintf(
				"repaint area: %dx%d at %d,%d",
				damageUnion.Width,
				damageUnion.Height,
				damageUnion.X,
				damageUnion.Y,
			)
		}
		damageWidth := ansi.StringWidth(damageText)
		if damageWidth < lastDamageOverlayWidth {
			damageText += strings.Repeat(" ", lastDamageOverlayWidth-damageWidth)
		} else {
			lastDamageOverlayWidth = damageWidth
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
		if height > 2 {
			ctx.DrawStyledText(0, 2, statsText, Style{
				ForegroundColor: BrightWhite,
				BackgroundColor: Black,
			})
		}
		if height > 3 {
			ctx.DrawStyledText(0, 3, damageText, Style{
				ForegroundColor: BrightWhite,
				BackgroundColor: Black,
			})
		}
	}

	renderInterval := time.Second / time.Duration(defaultFPS)
	lastModalCount := 0
	mouse := newMouseRouter(renderer, focusManager, hoveredSignal)

	// Render and update focusables
	display := func() {
		startTime := time.Now()
		drainPendingDispatches()
		// Update the focused signal BEFORE render so widgets can read it
		updateFocusedSignal()

		focusables := renderer.Update(root)
		focusManager.SetFocusables(focusables)

		// If focus changed after render (auto-focus or focus removal), re-render
		if updateFocusedSignal() {
			renderer.Update(root)
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
				renderer.Update(root)
			}
		}

		lastModalCount = modalCount
		// Update the signal and re-render so the focused widget shows focus style
		if updateFocusedSignal() {
			renderer.Update(root)
		}

		// Reconcile hover after render so enter/leave transitions still fire when
		// layout changes under a stationary pointer.
		if mouse.reconcileHover() {
			renderer.Update(root)
		}
		positionCursor(t, renderer.WidgetByID(focusManager.FocusedID()))

		// Sequences queued with WriteTerminal (clipboard writes and reads)
		// go out with this frame.
		for _, seq := range takeTerminalWrites() {
			_, _ = t.WriteString(seq)
		}
		debugRows := 0
		if debugOverlayEnabled {
			debugRows = min(height, 4)
		}
		if err := images.present(t, renderer, drawDebugOverlay, debugRows); err != nil {
			Log("Image presentation: %v", err)
		}

		elapsed := time.Since(startTime)
		lastFrameDuration = elapsed
		if elapsed > renderInterval {
			overrunFrames++
		}

		Log("Render complete in %.3fms, %d widgets registered", float64(elapsed.Microseconds())/1000.0, len(renderer.widgetRegistry.entries))
	}

	scheduler := newFrameScheduler(renderInterval, display)
	renderNow := scheduler.renderNow
	requestRender := func() {
		if scheduler.request() {
			coalescedRenderRequests++
		}
	}

	// suspend hands the terminal back to the shell, runs fn, then takes the
	// terminal back and redraws everything. Used for ctrl+z and RunExternal.
	suspend := func(fn func() error) error {
		awaitCellSizeReply()
		// Disable input reporting modes so the shell (or the program run)
		// gets plain keyboard input.
		disableTerminalInputModes(t.WriteString, enableKittyKeyboard, forceDisableKittyKeyboard, false)
		t.ExitAltScreen()
		// Pause stops reading input and restores the tty.
		_ = t.Pause()

		err := fn()

		_ = t.Resume()
		t.EnterAltScreen()
		enableTerminalInputModes(t.WriteString, true, enableKittyKeyboard, forceDisableKittyKeyboard)
		_, _ = t.WriteString(pointer.resume())
		// The screen was used by something else meanwhile; repaint it all.
		// Schedule the frame rather than drawing it here: fn may have been
		// run from a Dispatch callback, inside a frame.
		images.reset(t)
		renderer.fullRenderRequired = true
		t.Erase()
		scheduleRender()
		return err
	}
	setSuspender(suspend)

	// The text cursor blinks on this ticker when SetCursorBlink is on. Input
	// restarts it so the cursor stays shown while someone types.
	blinkTicker := time.NewTicker(cursorBlinkInterval)
	defer blinkTicker.Stop()
	restartCursorBlink := func() {
		showCursorForInput()
		blinkTicker.Reset(cursorBlinkInterval)
	}

	// Initial render
	renderNow()

	// Event loop
	eventLoopStarted = true
	go func() {
		defer close(eventLoopDone)
		defer scheduler.stopTimer()
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
			case <-currentRenderTrigger():
				requestRender()
			case <-animController.Tick():
				animController.Update()
				requestRender()
			case <-scheduler.timerCh:
				if scheduler.pending {
					renderNow()
				}
			case <-blinkTicker.C:
				toggleCursorBlink()
			case ev, ok := <-termEvents:
				if !ok {
					return
				}
				// Motion can arrive faster than it is handled (the terminal
				// reports every cell crossed, or every pixel with pixel
				// reporting); only the latest position matters. The pixel
				// pointer ignores mouse events, so it can't change mid-burst.
				if motion, isMotion := ev.(uv.MouseMotionEvent); isMotion {
					sgrRepair.reset()
					latest, next := coalesceMouseMotion(motion, termEvents)
					if mouse.motion(pointer.locateMotion(latest)) {
						requestRender()
					}
					if next == nil {
						continue
					}
					ev = next
				}
				// A report with a negative coordinate (the pointer outside
				// the window, in pixel mode) arrives in pieces, mostly as key
				// presses; hold them and handle the mouse event they make.
				if ev = sgrRepair.feed(ev); ev == nil {
					continue
				}
				if seq := pointer.handle(ev); seq != "" {
					_, _ = t.WriteString(seq)
					_ = t.Flush()
				}
				images.handle(ev, renderer)
				if renderer.imageUsed {
					switch ev.(type) {
					case uv.CellSizeEvent, uv.WindowPixelSizeEvent, uv.ModeReportEvent, uv.PrimaryDeviceAttributesEvent, uv.TerminalVersionEvent, uv.KittyGraphicsEvent:
						requestRender()
					}
				}
				// Pixel positions become cells, keeping the pointer's place
				// within its cell for widgets that track it precisely.
				ev, subX, subY := pointer.locateEvent(ev)
				switch ev := ev.(type) {
				case uv.WindowSizeEvent:
					_ = t.Resize(ev.Width, ev.Height)
					renderer.Resize(ev.Width, ev.Height)
					images.erased()
					width = ev.Width
					height = ev.Height
					t.Erase()
					requestRender()
				case uv.WindowPixelSizeEvent, uv.CellSizeEvent, uv.ModeReportEvent:
					// Shared mouse/image geometry and capability replies (above).
				case uv.KeyPressEvent:
					if scheduler.pending {
						renderNow()
					}
					keyEvent := KeyEvent{event: ev}
					captured := focusManager.capturesKey(keyEvent)
					// Check for app-level quit keys
					if !captured && ev.MatchString("ctrl+c") {
						cancel()
						return
					}

					// Screen export keybind
					if !captured && ev.MatchString("ctrl+shift+s") {
						exportScreenToFile()
						continue
					}

					// Suspend on Ctrl+Z
					if !captured && ev.MatchString("ctrl+z") {
						_ = suspend(func() error {
							return uv.Suspend() // Blocks until resumed via `fg`
						})
						continue
					}

					restartCursorBlink()
					dispatchKey(renderer, focusManager, root, keyEvent)

					// Re-render after key press (for signal updates and focus changes)
					requestRender()

				case uv.PasteEvent:
					// Like a key, a paste goes to what's on screen.
					if scheduler.pending {
						renderNow()
					}
					restartCursorBlink()
					if !dispatchPaste(focusManager, root, ev.Content) {
						Log("Paste not handled (%d bytes)", len(ev.Content))
					}
					requestRender()

				case uv.PasteStartEvent, uv.PasteEndEvent:
					// The terminal reader assembles the paste into a PasteEvent.

				case uv.ClipboardEvent:
					deliverClipboard(ev.Content)
					requestRender()

				case uv.MouseClickEvent:
					restartCursorBlink()
					mouse.press(ev, subX, subY, time.Now())
					requestRender()

				case uv.MouseReleaseEvent:
					mouse.release(ev, subX, subY)
					requestRender()

				case uv.MouseMotionEvent:
					if mouse.motion(ev, subX, subY) {
						requestRender()
					}

				case uv.MouseWheelEvent:
					if mouse.wheelAt(ev, subX, subY) {
						requestRender()
					}

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

// dispatchKey routes a key press. An overlay capturing keys (jump mode) takes
// it first; then Escape dismisses the top overlay if it allows that; then it
// goes to the focused widget, bubbling up through its ancestors; then to
// global keybinds (Jumper's toggle key), which work even with focus in an
// overlay; and finally to the root widget, which handles keys when nothing is
// focused.
func dispatchKey(renderer *Renderer, focusManager *FocusManager, root Widget, event KeyEvent) {
	if event.MatchString("escape") && renderer.cancelDrag() {
		return
	}
	if renderer.captureKey(event) {
		return
	}

	if event.MatchString("escape") {
		if topFloat := renderer.TopFloat(); topFloat != nil {
			if topFloat.Config.shouldDismissOnEsc() && topFloat.Config.OnDismiss != nil {
				topFloat.Config.OnDismiss()
				return
			}
		}
	}

	if focusManager.HandleKey(event) {
		return
	}
	if matchKeybind(event, renderer.focusCollector.globalKeybinds) {
		return
	}
	if provider, ok := root.(KeybindProvider); ok && matchKeybind(event, provider.Keybinds()) {
		return
	}
	if handler, ok := root.(KeyHandler); ok {
		handler.OnKey(event)
	}
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
