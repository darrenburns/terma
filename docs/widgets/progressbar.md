# ProgressBar

A horizontal bar that displays progress from 0% to 100%.
Expands horizontally to fill parent container.
Unicode block characters provide 8 levels of sub-cell precision for smooth rendering even at narrow widths.

=== "Demo"

    <video autoplay loop muted playsinline src="../../assets/progressbar-demo.mp4"></video>

=== "Code"

    ```go
    --8<-- "cmd/progressbar-example/main.go"
    ```

## Overview

`ProgressBar` displays a value between 0.0 and 1.0 as a horizontal bar. Pass the current progress to the `Progress` field and optionally customize colors with `FilledColor` and `UnfilledColor`.

```go
--8<-- "docs/minimal-examples/progressbar-basic/main.go"
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

## Animated Progress

Combine with `AnimatedValue` for smooth transitions when progress changes. `AnimatedValue` wraps a value and animates between old and new values when you call `Set()`. Use `Get()` in your `Build` method to read the current animated value, and `Target()` to read the target value (useful for incrementing).

Key configuration options:

- `Initial`: Starting value
- `Duration`: How long the animation takes
- `Easing`: The easing function (e.g., `EaseOutCubic`, `EaseInOutSine`)

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

While `AnimatedValue` animates between values you set manually, `Animation` drives progress automatically over time—useful for loading indicators, countdowns, or any progress that advances without user input.

`Animation` animates from the value `From` to the value `To` over a specified `Duration`, with an optional `Easing` function. `OnComplete` is called when the animation finishes. Call `Start()` to begin the animation. The `Value()` method returns a `Signal[T]`—call `Value().Get()` in your `Build` method to read the current value and automatically subscribe to updates.

```go
--8<-- "docs/minimal-examples/progressbar-animation/main.go"
```

### Looping

To loop an animation, call `Reset()` and `Start()` in the `OnComplete` callback. Since the callback references the animation variable, declare it first:

```go
type App struct {
    anim *t.Animation[float64]
}

func NewApp() *App {
    app := &App{}
    app.anim = t.NewAnimation(t.AnimationConfig[float64]{
        From:     0,
        To:       1,
        Duration: 3 * time.Second,
        OnComplete: func() {
            app.anim.Reset()
            app.anim.Start()
        },
    })
    app.anim.Start()
    return app
}
```

### Controlling Animations

`Animation` provides methods for controlling playback:

| Method | Description |
|--------|-------------|
| `Start()` | Begin or restart the animation |
| `Stop()` | Halt the animation without completing |
| `Pause()` | Temporarily suspend the animation |
| `Resume()` | Continue a paused animation |
| `Reset()` | Reset to the beginning (call `Start()` to play again) |
| `IsRunning()` | Returns true if currently animating |
| `IsComplete()` | Returns true if animation has finished |

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

Combine with `Text` widgets for labeled progress:

```go
type App struct {
    progress t.Signal[float64]
}

func NewApp() *App {
    return &App{progress: t.NewSignal(0.65)}
}

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
            t.ProgressBar{Progress: progress},
        },
    }
}
```

## Styling

Apply borders and padding through the Style field:

```go
--8<-- "docs/minimal-examples/progressbar-styling/main.go"
```

## Notes

- Progress values outside 0.0-1.0 are automatically clamped
- Default width is `Flex(1)`, expanding to fill available space
- Colors default to theme values if not specified
- Height defaults to `Cells(1)` for a single-row bar
