# Text

A leaf widget that displays plain or rich text content, with optional wrapping and alignment.

Set `Style.BackgroundColor` directly to a gradient; no container wrapper is needed:

```go
terma.Text{
    Content: "Gradient background",
    Style: terma.Style{
        Width: terma.Cells(30),
        BackgroundColor: terma.NewGradient(
            terma.Hex("#962814"),
            terma.Hex("#142896"),
        ).WithAngle(90),
    },
}
```

Background gradients span the Text's border box, including padding and borders.
Wrapped lines and aligned text keep the gradient at their cell positions.
`SpanStyle.Background` overrides the gradient for that span; translucent span
backgrounds blend over it. Translucent Text backgrounds blend over the underlying
background once, and foreground colors blend over the resulting background.
A wide terminal glyph uses its leading cell's color for the whole glyph.
