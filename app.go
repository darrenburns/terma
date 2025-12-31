package terma

import (
	"context"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/ultraviolet/screen"
	"github.com/charmbracelet/x/ansi"
)

// appCancel holds the cancel function for the currently running app.
var appCancel func()

// Quit exits the running application gracefully.
// This performs the same teardown as pressing Ctrl+C.
func Quit() {
	if appCancel != nil {
		appCancel()
	}
}

// Run starts the application with the given root widget and blocks until it exits.
// The root widget can implement KeyHandler to receive key events that bubble up
// from focused descendants.
func Run(root Widget) error {
	t := uv.DefaultTerminal()

	if err := t.Start(); err != nil {
		return err
	}

	t.EnterAltScreen()

	// Enable mouse tracking (all mouse events including hover + SGR extended encoding)
	t.WriteString(ansi.SetModeMouseAnyEvent)
	t.WriteString(ansi.SetModeMouseExtSgr)

	ctx, cancel := context.WithCancel(context.Background())
	appCancel = cancel
	defer func() {
		appCancel = nil
		cancel()
	}()

	// Get initial terminal size
	size := t.Size()
	width, height := size.Width, size.Height

	// Create focus manager and focused signal
	focusManager := NewFocusManager()
	focusedSignal := NewAnySignal[Focusable](nil)

	// Create hovered widget signal (tracks the currently hovered widget)
	hoveredSignal := NewAnySignal[Widget](nil)

	// Create renderer with focus manager and signal
	renderer := NewRenderer(t, width, height, focusManager, focusedSignal, hoveredSignal)

	// Render and update focusables
	display := func() {
		startTime := time.Now()
		screen.Clear(t)
		// Update the focused signal BEFORE render so widgets can read it
		focusedSignal.Set(focusManager.Focused())

		focusables := renderer.Render(root)
		focusManager.SetFocusables(focusables)
		t.Display()

		elapsed := time.Since(startTime)
		Log("Render complete in %.3fms, %d widgets registered", float64(elapsed.Microseconds())/1000.0, len(renderer.widgetRegistry.entries))
	}

	// Get root's key handling interfaces (if any) for the no-focusables case
	rootHandler, _ := root.(KeyHandler)
	rootKeybindProvider, _ := root.(KeybindProvider)

	// Initial render
	display()

	// Event loop
	go func() {
		for ev := range t.Events() {
			switch ev := ev.(type) {
			case uv.WindowSizeEvent:
				t.Resize(ev.Width, ev.Height)
				renderer.Resize(ev.Width, ev.Height)
				t.Erase()
				display()
			case uv.KeyPressEvent:
				// Check for app-level quit keys
				if ev.MatchString("ctrl+c") {
					cancel()
					return
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
				display()

			case uv.MouseClickEvent:
				Log("MouseClickEvent at X=%d Y=%d Button=%v", ev.X, ev.Y, ev.Button)

				// Find the widget under the cursor
				entry := renderer.WidgetAt(ev.X, ev.Y)
				if entry != nil {
					Log("  Found widget: Key=%q Type=%T", entry.Key, entry.Widget)
					// If the widget implements Clickable, call OnClick
					if clickable, ok := entry.Widget.(Clickable); ok {
						Log("  Widget is Clickable, calling OnClick")
						clickable.OnClick()
					} else {
						Log("  Widget is NOT Clickable")
					}
				} else {
					Log("  No widget found at position")
					LogWidgetRegistry(renderer.widgetRegistry)
				}

				// Re-render after click
				display()

			case uv.MouseMotionEvent:
				// Log("MouseMotionEvent at X=%d Y=%d", ev.X, ev.Y)

				// Find the widget under the cursor
				entry := renderer.WidgetAt(ev.X, ev.Y)
				var newHovered Widget
				newHoveredKey := ""
				if entry != nil {
					newHovered = entry.Widget
					newHoveredKey = entry.Key
					// Log("  Widget at position: Key=%q Type=%T", entry.Key, entry.Widget)
				} else {
					// Log("  No widget at position")
				}

				// Get old hovered widget and its key
				oldHovered := hoveredSignal.Get()
				oldHoveredKey := ""
				if oldHovered != nil {
					if keyed, ok := oldHovered.(Keyed); ok {
						oldHoveredKey = keyed.Key()
					}
				}

				// Only update if hover changed (compare by key to avoid incomparable type issues)
				if newHoveredKey != oldHoveredKey {
					Log("  Hover changed: %q -> %q", oldHoveredKey, newHoveredKey)

					// Notify old widget it's no longer hovered
					if oldHovered != nil {
						if hoverable, ok := oldHovered.(Hoverable); ok {
							Log("  Calling OnHover(false) on %q", oldHoveredKey)
							hoverable.OnHover(false)
						}
					}

					// Update the hovered signal
					hoveredSignal.Set(newHovered)

					// Notify new widget it's now hovered
					if entry != nil {
						if hoverable, ok := entry.Widget.(Hoverable); ok {
							Log("  Calling OnHover(true) on %q", newHoveredKey)
							hoverable.OnHover(true)
						}
					}

					// Re-render after hover change
					display()
				}

			case uv.MouseWheelEvent:
				// Find the scrollable widget under the cursor (may not be focused)
				scrollable := renderer.ScrollableAt(ev.X, ev.Y)
				if scrollable != nil {
					Log("MouseWheelEvent %v at X=%d Y=%d on Scrollable[%s]", ev.Button, ev.X, ev.Y, scrollable.ID)
					switch ev.Button {
					case uv.MouseWheelUp:
						scrollable.ScrollUp(1)
					case uv.MouseWheelDown:
						scrollable.ScrollDown(1)
					}
					// Re-render after scroll
					display()
				} else {
					Log("MouseWheelEvent %v at X=%d Y=%d (no scrollable)", ev.Button, ev.X, ev.Y)
				}

			default:
				// Log other event types for debugging
				Log("Unhandled event: %T %v", ev, ev)
			}
		}
	}()

	<-ctx.Done()

	// Disable mouse tracking before shutdown
	t.WriteString(ansi.ResetModeMouseAnyEvent)
	t.WriteString(ansi.ResetModeMouseExtSgr)

	return t.Shutdown(context.Background())
}
