# Testing

Terma has two ways to test an app. `AssertSnapshot` renders one widget once and compares it with a golden SVG. A `Pilot` runs the whole app, lets a test press keys, type, paste, click and resize, and then checks what is on screen.

## Snapshot a widget

```go
func TestGreeting(t *testing.T) {
    terma.AssertSnapshot(t, Greeting{Name: "Ada"}, 30, 3, "The greeting reads 'Hello, Ada'")
}
```

The golden file is `testdata/TestGreeting.svg`. Create or update it with:

```bash
UPDATE_SNAPSHOTS=1 go test ./...
```

Each run also writes `testdata/snapshot_gallery.html`, which shows every snapshot next to its description. A failing snapshot shows the expected and actual output with the differing cells marked.

## Drive an app with a Pilot

A snapshot of a static widget can't show what happens after input. Use a `Pilot` for that:

```go
func TestSignup(t *testing.T) {
    app := NewSignupApp()
    p := terma.NewPilot(t, app, 80, 24)

    p.Type("Ada")
    p.Press("tab", "enter")

    if got := p.TextOf("status"); got != "Welcome, Ada" {
        t.Fatalf("status = %q", got)
    }
    if got := p.FocusedID(); got != "ok" {
        t.Fatalf("focus = %q, want the dialog's OK button", got)
    }
    p.AssertSnapshot("welcome", "The welcome dialog is open with OK focused")

    p.Click("ok")
}
```

`NewPilot(t, root, width, height)` starts the app on an in-memory screen and draws its first frame. The app stops when the test ends.

The Pilot runs the app on the same retained renderer, focus manager and event routing as `Run`. Only the terminal is replaced. Keybinds, Tab focus order, focus traps, modal focus, Escape on overlays, mouse hit testing and paste handling all behave as they do for a user. A click on a button that a modal covers goes to the modal's backdrop, not the button.

After each input the Pilot draws frames until nothing is left to do. Work queued with `Dispatch`, signal changes and `RequestFocus` calls have all applied by the time the call returns.

### Input

| Method | What it does |
|--------|--------------|
| `Press(keys...)` | Presses keys in turn. Spell keys as in `Keybind.Key`: `"tab"`, `"enter"`, `"escape"`, `"ctrl+s"`, `"shift+down"`, `"q"`, `"?"` |
| `Type(text)` | Types text one character at a time |
| `Paste(text)` | Pastes text in one piece, as bracketed paste delivers it |
| `Click(id)` | Clicks the middle of the visible part of the widget with that ID |
| `ClickAt(x, y)` | Clicks the left button at a screen cell |
| `MouseDown(x, y, button, mod)` / `MouseUp(...)` | Presses or releases a button with modifiers, for shift-click, right-click or the start and end of a drag |
| `MouseMove(x, y)` | Moves the pointer. Between `MouseDown` and `MouseUp` this drags |
| `Hover(id)` | Moves the pointer over a widget |
| `Scroll(x, y, lines)` | Turns the mouse wheel. Positive lines scroll down |
| `Resize(width, height)` | Resizes the screen |
| `Send(events...)` | Sends raw ultraviolet events for anything else |

`Click`, `TextOf` and `Hover` fail the test if no widget with that ID is on screen.

### Reading the result

| Method | Returns |
|--------|---------|
| `ScreenText()` | The screen as plain text, one line per row |
| `TextOf(id)` | The text inside one widget, with trailing spaces trimmed |
| `Bounds(id)` | The widget's screen rectangle |
| `FocusedID()` | The ID of the focused widget |
| `Keybinds()` | The keybinds active for the focused widget, as `KeybindBar` sees them |
| `Clipboard()` | What the app last copied with `SetClipboard` |
| `Exited()` | Whether the app quit through `Quit` or ctrl+c |
| `Buffer()` | The screen's cells, for checks on colors or styles |
| `AssertSnapshot(name, description)` | Compares the screen with `testdata/<TestName>_<name>.svg` |

The Pilot plays the terminal's part for the clipboard. `SetClipboard` writes to it, and `ReadClipboard` reads from it, so a copy and paste round trip works in a test.

### Time

The Pilot's clock starts at a fixed time and only moves when you call `Advance`. Spinners, shimmers, animations and the cursor blink stay still until then, so every frame is the same on every run.

```go
spinner.Start()
p.Advance(100 * time.Millisecond) // the spinner shows its second frame
```

While animations run, `Advance` steps the clock one frame (1/60 s) at a time and draws each frame, as a running app does.

Clicks on the same cell form a double click while the clock stands still. Call `p.Advance(time.Second)` between two clicks to keep them separate.

### Background work

Goroutines started by the app, such as a `Task` or code that sets a signal from another goroutine, run in real time. `WaitUntil` draws frames as their updates arrive and returns when a condition holds:

```go
p.Press("r") // starts a refresh Task
p.WaitUntil(func() bool { return p.TextOf("status") == "Loaded" }, 0)
```

The second argument is a real-time timeout. Zero means one second. The test fails if the condition still doesn't hold by then.

### Limits

- The Pilot takes the place of the running app, so a test can have one Pilot at a time. Tests that use a Pilot can't call `t.Parallel`.
- There is no terminal to suspend. `RunExternal` runs its command directly, and ctrl+z reaches the app as an ordinary key.
- If the app changes state on every frame (for example, a `Dispatch` call from `OnLayout` that never settles), the Pilot fails the test after 32 frames instead of looping.

## Check a real terminal

The Pilot covers the app's logic. To see the real output in a real terminal, drive the built program through tmux with `scripts/tui-capture.sh`, which captures the screen after each key as ANSI, SVG and PNG:

```bash
go build -o /tmp/demo ./cmd/form-demo
scripts/tui-capture.sh -s 80x24 -o snapshot-output/tui/form /tmp/demo Tab Ada Enter
```
