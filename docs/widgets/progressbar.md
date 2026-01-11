# ProgressBar

A horizontal bar that displays progress from 0% to 100%.

```go
package main

import t "terma"

type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.ProgressBar{
        Progress: 0.7,
        Width:    t.Cells(30),
    }
}

func main() {
    t.Run(&App{})
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `Progress` | `float64` | — | Value from 0.0 to 1.0 |
| `Width` | `Dimension` | `Flex(1)` | Bar width |
| `Height` | `Dimension` | `Cells(1)` | Bar height |
| `Style` | `Style` | — | Padding, margin, border |
| `FilledColor` | `Color` | theme.Primary | Color of filled portion |
| `UnfilledColor` | `Color` | theme.Surface | Color of unfilled portion |

## Basic Usage

```go
// 50% progress with default styling
ProgressBar{Progress: 0.5}

// Fixed width with custom colors
ProgressBar{
    Progress:      0.75,
    Width:         Cells(40),
    FilledColor:   theme.Success,
    UnfilledColor: theme.Background,
}
```

## Sub-Cell Precision

ProgressBar uses Unicode block characters for smooth rendering:

```
▏ ▎ ▍ ▌ ▋ ▊ ▉ █
```

This provides 8 levels of precision per character cell, making animations appear smooth even at narrow widths.

## Animated Progress

Combine with `AnimatedValue` for smooth transitions when progress changes:

```go
type App struct {
    progress *t.AnimatedValue[float64]
}

func NewApp() *App {
    return &App{
        progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
            Initial:  0,
            Duration: 300 * time.Millisecond,
            Easing:   t.EaseOutCubic,
        }),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.ProgressBar{
        Progress: a.progress.Get(),
        Width:    t.Cells(40),
    }
}

func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "+", Name: "Increase", Action: func() {
            current := a.progress.Target()
            if current < 1.0 {
                a.progress.Set(current + 0.1)
            }
        }},
        {Key: "-", Name: "Decrease", Action: func() {
            current := a.progress.Target()
            if current > 0.0 {
                a.progress.Set(current - 0.1)
            }
        }},
    }
}
```

## Auto-Advancing Progress

Use `Animation` for continuous progress animations:

```go
type App struct {
    anim *t.Animation[float64]
}

func NewApp() *App {
    anim := t.NewAnimation(t.AnimationConfig[float64]{
        From:     0,
        To:       1,
        Duration: 3 * time.Second,
        Easing:   t.EaseInOutSine,
        OnComplete: func() {
            // Loop the animation
        },
    })
    anim.Start()

    return &App{anim: anim}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.ProgressBar{
        Progress: a.anim.Value().Get(),
        Width:    t.Cells(40),
    }
}
```

## Multiple Progress Bars

Display several progress bars with different styles:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Spacing: 1,
        Style:   t.Style{Padding: t.EdgeInsetsAll(2)},
        Children: []t.Widget{
            t.Row{Children: []t.Widget{
                t.Text{Content: "Downloads ", Style: t.Style{Width: t.Cells(12)}},
                t.ProgressBar{Progress: 0.45, FilledColor: theme.Primary},
            }},
            t.Row{Children: []t.Widget{
                t.Text{Content: "Uploads   ", Style: t.Style{Width: t.Cells(12)}},
                t.ProgressBar{Progress: 0.82, FilledColor: theme.Success},
            }},
            t.Row{Children: []t.Widget{
                t.Text{Content: "Processing", Style: t.Style{Width: t.Cells(12)}},
                t.ProgressBar{Progress: 0.23, FilledColor: theme.Warning},
            }},
        },
    }
}
```

## With Labels

Combine with Text widgets for labeled progress:

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    progress := a.progress.Get()
    percent := int(progress * 100)

    return t.Column{
        Children: []t.Widget{
            t.Row{
                Children: []t.Widget{
                    t.Text{Content: "Installing..."},
                    t.Spacer{},
                    t.Text{Content: fmt.Sprintf("%d%%", percent)},
                },
            },
            t.ProgressBar{
                Progress: progress,
                Width:    t.Flex(1),
            },
        },
    }
}
```

## Styling

Apply borders and padding through the Style field:

```go
ProgressBar{
    Progress: 0.6,
    Width:    Cells(30),
    Style: Style{
        Border:      BorderRounded,
        BorderColor: theme.TextMuted,
        Padding:     EdgeInsetsHV(1, 0),
    },
    FilledColor:   theme.Accent,
    UnfilledColor: theme.Background,
}
```

## Notes

- Progress values outside 0.0-1.0 are automatically clamped
- Default width is `Flex(1)`, expanding to fill available space
- Colors default to theme values if not specified
- Height defaults to `Cells(1)` for a single-row bar
