# Inline Mode

`Run` takes over the whole terminal in the alternate screen. `RunInline` runs the same widget tree in the normal screen instead, below the cursor, the way a CLI prompt, a progress view or a chat composer does. The shell history above stays visible, and anything the app prints stays in the scrollback after it exits.

```go
func main() {
	app := &App{input: terma.NewTextInputState("")}
	if err := terma.RunInline(app, terma.InlineOptions{}); err != nil {
		log.Fatal(err)
	}
}
```

Everything else works as it does in a fullscreen app: signals, focus, keybinds, floats, animation and `Dispatch`.

## The live region

The app draws into a *live region* that starts on the row where the cursor was. The region's height follows the root widget's height, so a root with `Auto` height grows and shrinks with its content. When the region reaches the bottom of the screen, the terminal scrolls the rows above it up.

`InlineOptions.MaxHeight` caps the region. Zero means the terminal height. The root is laid out with that height as its limit:

- An `Auto` height follows the content. Use this for most inline apps.
- A `Flex` height fills `MaxHeight`. A `Dock` whose body is `Flex(1)` therefore takes the whole cap.
- Content taller than the cap is cut off at the bottom. Put it in a `Scrollable` with a fixed height if it can grow without limit.

Floats such as menus, autocomplete popups and tooltips can extend below the root widget. The region grows to fit them while they are open, up to `MaxHeight`, and shrinks again when they close. A float is positioned against the full `MaxHeight` area, so a centered `Dialog` makes the region grow to about half the terminal height. Prefer anchored floats in inline apps.

## Printing above the region

`PrintAbove` writes finished output permanently above the live region. The output becomes part of the terminal scrollback, and the region moves down to make room for it.

```go
func (a *App) submit(text string) {
	terma.PrintAbove(terma.ParseMarkupToText("[b $Primary]You[/] "+text, a.theme))
	a.input.SetText("")
}
```

`PrintAbove` takes a widget and renders it at the terminal width, so styles, borders and wrapped `Text` all work. `PrintAboveText` takes a string, which may contain ANSI escape sequences. It does not wrap. A line wider than the terminal is cut off at the edge, so use `PrintAbove(terma.Text{Content: s, Wrap: terma.WrapSoft})` for long text.

Both functions are safe to call from any goroutine. The output is written with the next frame. With no app running, they print to standard output directly. A fullscreen app ignores them.

## Leaving the app

`InlineOptions.OnExit` decides what happens to the live region when the app exits:

| Value | Result |
|-------|--------|
| `InlineExitKeep` (default) | The final frame stays in the scrollback. The cursor moves to the line below it. |
| `InlineExitClear` | The region is erased. The cursor returns to the row where the app started. |

Output written with `PrintAbove` stays in both cases. For a prompt whose answer you print yourself, use `InlineExitClear` and print the answer with `PrintAbove` before calling `Quit`.

## Mouse input

Mouse reporting is off by default in inline mode. While it is on, the terminal sends wheel events to the app, and the wheel no longer scrolls the terminal's own scrollback. Set `InlineOptions.Mouse` to turn it on. Mouse positions are translated so that widgets receive coordinates relative to the region, as in a fullscreen app. To do this, Terma asks the terminal for the cursor position at startup and after a resize. Mouse events that arrive before the terminal answers are ignored.

## Suspending and external programs

`ctrl+z` and `RunExternal` work inline. Before handing over the terminal, Terma erases the live region, so the shell or the external program writes from the region's first row. When the app takes the terminal back, it draws the region again below that output.

## Resizing

When the terminal is resized, Terma erases the region and draws it again at the new size. Some terminals rewrap old lines when the width changes. If a terminal rewraps the region's old rows before Terma erases them, a few stale rows can remain above the new region.

## Limitations

- Sixel images are positioned in absolute screen coordinates and don't display correctly in inline mode.
- The terminal reports the cursor position only when asked, so if another program writes to the terminal while the app runs, the region and mouse positions can be off until the next resize.

See `cmd/inline-example` for a complete chat composer that uses a text input, an autocomplete popup, a streamed reply and `PrintAbove`.
