# Styling

Terma provides a flexible styling system for controlling colors, padding, margins, and borders. Themes allow you to define consistent color palettes across your application.

## Tall borders

`TallBorder` ports Textual's `tall` border: thin strokes at the very top and
bottom of the border cells leave more visual space around the content. The
fractional-block sides blend into the parent background. Like other borders,
it occupies one cell on each side and supports titles, subtitles, and gradients.

```go
terma.Style{
    Border: terma.TallBorder(terma.RGB(180, 150, 230),
        terma.BorderTitle("Settings"),
        terma.BorderSubtitleRight("Saved"),
    ),
}
```

You can also use `terma.Border{Style: terma.BorderTall, Color: color}`.
Run `go run ./cmd/border-example` to compare the border styles.

## Mouse pointers

Set `Style.Pointer` to change the mouse pointer over a widget:

```go
terma.Button{
    Label: "Open",
    Style: terma.Style{Pointer: terma.PointerPointer},
}
```

Shapes include `PointerText`, `PointerGrab`, `PointerGrabbing`, `PointerWait`,
`PointerNotAllowed`, `PointerEWResize`, and `PointerNSResize`. Terma provides
constants for all 30 shapes in the [Kitty pointer shape protocol](https://sw.kovidgoyal.net/kitty/pointer-shapes/).
This controls the mouse pointer, not the text cursor.

The zero value (`PointerInherit`) uses the nearest ancestor's pointer.
`PointerDefault` explicitly restores the terminal's default shape, even inside
a parent with a custom pointer. Disabled widgets and modal backdrops use the
default. Overlays do not inherit from widgets behind them. During a drag, the
pressed widget keeps control of the pointer until release. Pointer styles also
update when a rebuild or layout change occurs beneath a stationary mouse.

`SplitPane` automatically uses the appropriate resize pointer over its divider
and while dragging. Set its `Style.Pointer` to override that behavior.

Pointer shapes require a terminal that supports OSC 22, such as Kitty or
Ghostty; other terminals leave the pointer unchanged. `terma-browser` also
supports Terma's pointer shapes. The pointer resets when the app exits or
suspends to run an external command.
