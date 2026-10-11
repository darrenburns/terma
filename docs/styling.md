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
