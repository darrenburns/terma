# Animation

Terma provides a complete animation system for smooth transitions, loading indicators, and dynamic visual effects. Animations integrate with the reactive signal system, automatically triggering widget rebuilds when values change.

## Animation Types

| Type | Purpose |
|------|---------|
| `Animation[T]` | Interpolate between two values over time |
| `AnimatedValue[T]` | Wrap a value that animates on change |
| `FrameAnimation[T]` | Cycle through discrete frames |
| `Spinner` | Pre-built loading indicators |

## Animation[T]

Smoothly interpolates between a start and end value over a duration.

```go
anim := terma.NewAnimation(terma.AnimationConfig[float64]{
    From:     0,
    To:       100,
    Duration: 500 * time.Millisecond,
    Easing:   terma.EaseOutCubic,
})

anim.Start()
```

### Using in Build()

Access the animated value through the `Value()` signal:

```go
func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    progress := a.anim.Value().Get()  // Subscribes to updates

    return terma.Text{
        Content: fmt.Sprintf("Progress: %.0f%%", progress),
    }
}
```

### Configuration

```go
type AnimationConfig[T any] struct {
    From         T                // Start value (required)
    To           T                // End value (required)
    Duration     time.Duration    // Animation length (required)
    Easing       EasingFunc       // Easing function (default: EaseLinear)
    Interpolator Interpolator[T]  // Custom interpolator (optional)
    Delay        time.Duration    // Delay before starting (optional)
    OnComplete   func()           // Called when animation finishes
    OnUpdate     func(T)          // Called on each value change
}
```

### Control Methods

```go
anim.Start()              // Begin animation
anim.Stop()               // Halt immediately
anim.Pause()              // Pause (can resume)
anim.Resume()             // Continue from pause
anim.Reset()              // Restart from beginning

anim.IsRunning() bool     // Currently animating?
anim.IsComplete() bool    // Finished?
anim.Progress() float64   // Progress from 0.0 to 1.0
anim.Get() T              // Current value (no subscription)
anim.Value() AnySignal[T] // Signal for reactive updates
```

### Color Animation Example

```go
type App struct {
    colorAnim *terma.Animation[terma.Color]
}

func NewApp() *App {
    return &App{
        colorAnim: terma.NewAnimation(terma.AnimationConfig[terma.Color]{
            From:     terma.RGB(50, 50, 200),   // Blue
            To:       terma.RGB(200, 50, 50),   // Red
            Duration: 2 * time.Second,
            Easing:   terma.EaseInOutSine,
        }),
    }
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    color := a.colorAnim.Value().Get()

    return terma.Text{
        Content: "████████████████",
        Style:   terma.Style{ForegroundColor: color},
    }
}

func (a *App) Keybinds() []terma.Keybind {
    return []terma.Keybind{
        {Key: "space", Name: "Animate", Action: func() {
            a.colorAnim.Reset()
            a.colorAnim.Start()
        }},
    }
}
```

## AnimatedValue[T]

Wraps a value that automatically animates when changed. Call `Set()` to smoothly transition to a new value. If called mid-animation, it retargets from the current position.

```go
progress := terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
    Initial:  0,
    Duration: 300 * time.Millisecond,
    Easing:   terma.EaseOutQuad,
})
```

### Basic Usage

```go
type App struct {
    progress *terma.AnimatedValue[float64]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    value := a.progress.Get()  // Current animated value

    return terma.Column{
        Children: []terma.Widget{
            terma.Text{Content: fmt.Sprintf("%.0f%%", value)},
            a.buildProgressBar(value),
        },
    }
}

func (a *App) Keybinds() []terma.Keybind {
    return []terma.Keybind{
        {Key: "+", Name: "Increase", Action: func() {
            current := a.progress.Target()
            if current < 100 {
                a.progress.Set(current + 10)  // Animates to new value
            }
        }},
        {Key: "-", Name: "Decrease", Action: func() {
            current := a.progress.Target()
            if current > 0 {
                a.progress.Set(current - 10)
            }
        }},
    }
}
```

### Methods

```go
av.Get() T               // Current value (subscribes in Build())
av.Peek() T              // Current value (no subscription)
av.Target() T            // Target value (what it's animating toward)
av.Set(value T)          // Animate to new value
av.SetImmediate(value T) // Set instantly without animation
av.IsAnimating() bool    // Currently in transition?
av.Signal() Signal[T]    // Underlying signal
```

### Retargeting

When `Set()` is called during an animation, `AnimatedValue` smoothly retargets from the current interpolated position:

```go
// Value is at 0
a.progress.Set(100)  // Starts animating toward 100

// Mid-animation, value is around 50
a.progress.Set(25)   // Smoothly redirects toward 25 from current position
```

### AnyAnimatedValue[T]

For non-comparable types, use `AnyAnimatedValue[T]`:

```go
colorValue := terma.NewAnyAnimatedValue(terma.AnyAnimatedValueConfig[terma.Color]{
    Initial:  terma.RGB(255, 0, 0),
    Duration: 500 * time.Millisecond,
    Easing:   terma.EaseOutCubic,
})

colorValue.Set(terma.RGB(0, 255, 0))  // Animate to green
```

## FrameAnimation[T]

Cycles through discrete frames at a fixed interval. Use for spinners, sprite animations, or any sequence-based animation.

```go
frames := terma.NewFrameAnimation(terma.FrameAnimationConfig[string]{
    Frames:    []string{"◐", "◓", "◑", "◒"},
    FrameTime: 100 * time.Millisecond,
    Loop:      true,
})

frames.Start()
```

### Usage

```go
type App struct {
    spinner *terma.FrameAnimation[string]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    frame := a.spinner.Value().Get()

    return terma.Row{
        Children: []terma.Widget{
            terma.Text{Content: frame},
            terma.Text{Content: " Loading..."},
        },
    }
}
```

### Configuration

```go
type FrameAnimationConfig[T any] struct {
    Frames     []T           // Frames to cycle through (required)
    FrameTime  time.Duration // Duration per frame (required)
    Loop       bool          // Repeat indefinitely
    OnComplete func()        // Called when non-looping animation ends
}
```

### Methods

```go
fa.Start()              // Begin animation
fa.Stop()               // Halt
fa.Reset()              // Return to first frame
fa.Get() T              // Current frame value
fa.Value() AnySignal[T] // Signal for reactive updates
fa.Index() int          // Current frame index
fa.IsRunning() bool     // Currently animating?
```

### Custom Frame Animation

```go
// Animated ellipsis
ellipsis := terma.NewFrameAnimation(terma.FrameAnimationConfig[string]{
    Frames:    []string{"", ".", "..", "..."},
    FrameTime: 300 * time.Millisecond,
    Loop:      true,
})

// Progress indicator
blocks := terma.NewFrameAnimation(terma.FrameAnimationConfig[string]{
    Frames:    []string{"▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
    FrameTime: 80 * time.Millisecond,
    Loop:      true,
})
```

## Spinner

Pre-built loading indicators using `FrameAnimation[string]`.

```go
type App struct {
    spinner *terma.SpinnerState
}

func NewApp() *App {
    spinner := terma.NewSpinnerState(terma.SpinnerDots)
    spinner.Start()

    return &App{spinner: spinner}
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    return terma.Row{
        Children: []terma.Widget{
            terma.Spinner{State: a.spinner},
            terma.Text{Content: " Processing..."},
        },
    }
}
```

### Spinner Styles

| Style | Frames | Description |
|-------|--------|-------------|
| `SpinnerDots` | ⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏ | Classic braille dots |
| `SpinnerLine` | - \ \| / | Rotating line |
| `SpinnerCircle` | ◐ ◓ ◑ ◒ | Quarter-filled circle |
| `SpinnerBounce` | ⠁ ⠂ ⠄ ⠂ | Bouncing dot |
| `SpinnerArrow` | ← ↖ ↑ ↗ → ↘ ↓ ↙ | Rotating arrow |
| `SpinnerBraille` | ⣾ ⣽ ⣻ ⢿ ⡿ ⣟ ⣯ ⣷ | Detailed braille |
| `SpinnerGrow` | ▁ ▂ ▃ ▄ ▅ ▆ ▇ █ | Growing bar |
| `SpinnerPulse` | █ ▓ ▒ ░ ▒ ▓ | Pulsing block |
| `SpinnerClock` | 🕐 🕑 🕒 ... | Clock faces |
| `SpinnerMoon` | 🌑 🌒 🌓 🌔 🌕 🌖 🌗 🌘 | Moon phases |
| `SpinnerDotsBounce` | ⠁ ⠂ ⠄ ⡀ ⢀ ⠠ ⠐ ⠈ | Bouncing braille |

### SpinnerState Methods

```go
state.Start()           // Begin spinning
state.Stop()            // Stop spinning
state.IsRunning() bool  // Currently active?
```

### Multiple Spinners

```go
type App struct {
    dots   *terma.SpinnerState
    circle *terma.SpinnerState
    grow   *terma.SpinnerState
}

func NewApp() *App {
    return &App{
        dots:   terma.NewSpinnerState(terma.SpinnerDots),
        circle: terma.NewSpinnerState(terma.SpinnerCircle),
        grow:   terma.NewSpinnerState(terma.SpinnerGrow),
    }
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    return terma.Column{
        Spacing: 1,
        Children: []terma.Widget{
            terma.Row{Children: []terma.Widget{
                terma.Spinner{State: a.dots},
                terma.Text{Content: " Dots"},
            }},
            terma.Row{Children: []terma.Widget{
                terma.Spinner{State: a.circle},
                terma.Text{Content: " Circle"},
            }},
            terma.Row{Children: []terma.Widget{
                terma.Spinner{State: a.grow},
                terma.Text{Content: " Grow"},
            }},
        },
    }
}
```

## Easing Functions

Easing functions control the rate of change over time. All functions take a progress value `t` (0.0 to 1.0) and return the eased value.

### Available Easing Functions

**Linear:**
- `EaseLinear` - Constant velocity, no acceleration

**Quadratic (t²):**
- `EaseInQuad` - Slow start, accelerates
- `EaseOutQuad` - Fast start, decelerates
- `EaseInOutQuad` - Slow start and end

**Cubic (t³):**
- `EaseInCubic` - Slower start than quadratic
- `EaseOutCubic` - Slower deceleration than quadratic
- `EaseInOutCubic` - Smooth start and end

**Quartic (t⁴) and Quintic (t⁵):**
- `EaseInQuart`, `EaseOutQuart`, `EaseInOutQuart`
- `EaseInQuint`, `EaseOutQuint`, `EaseInOutQuint`

**Sinusoidal:**
- `EaseInSine` - Gentle acceleration
- `EaseOutSine` - Gentle deceleration
- `EaseInOutSine` - Natural feel, good for UI

**Exponential:**
- `EaseInExpo` - Very slow start, rapid acceleration
- `EaseOutExpo` - Rapid deceleration
- `EaseInOutExpo` - Dramatic start and end

**Circular:**
- `EaseInCirc`, `EaseOutCirc`, `EaseInOutCirc`

**Elastic (spring-like oscillation):**
- `EaseInElastic` - Winds up before moving
- `EaseOutElastic` - Overshoots and springs back
- `EaseInOutElastic` - Both ends

**Back (overshoot):**
- `EaseInBack` - Pulls back before moving forward
- `EaseOutBack` - Overshoots target, then returns
- `EaseInOutBack` - Both ends

**Bounce:**
- `EaseInBounce` - Bounces at start
- `EaseOutBounce` - Bounces at end
- `EaseInOutBounce` - Bounces at both ends

### Choosing an Easing Function

| Use Case | Recommended Easing |
|----------|-------------------|
| UI transitions | `EaseOutQuad`, `EaseOutCubic` |
| Attention-grabbing | `EaseOutElastic`, `EaseOutBack` |
| Natural movement | `EaseInOutSine`, `EaseInOutQuad` |
| Mechanical feel | `EaseLinear` |
| Playful UI | `EaseOutBounce` |
| Progress bars | `EaseOutCubic`, `EaseOutQuart` |

### Easing Comparison Example

```go
type App struct {
    linear  *terma.AnimatedValue[float64]
    quad    *terma.AnimatedValue[float64]
    cubic   *terma.AnimatedValue[float64]
    elastic *terma.AnimatedValue[float64]
}

func NewApp() *App {
    duration := 1 * time.Second
    return &App{
        linear: terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
            Duration: duration,
            Easing:   terma.EaseLinear,
        }),
        quad: terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
            Duration: duration,
            Easing:   terma.EaseOutQuad,
        }),
        cubic: terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
            Duration: duration,
            Easing:   terma.EaseOutCubic,
        }),
        elastic: terma.NewAnimatedValue(terma.AnimatedValueConfig[float64]{
            Duration: duration,
            Easing:   terma.EaseOutElastic,
        }),
    }
}

func (a *App) Keybinds() []terma.Keybind {
    return []terma.Keybind{
        {Key: "space", Name: "Animate", Action: func() {
            a.linear.Set(100)
            a.quad.Set(100)
            a.cubic.Set(100)
            a.elastic.Set(100)
        }},
        {Key: "r", Name: "Reset", Action: func() {
            a.linear.SetImmediate(0)
            a.quad.SetImmediate(0)
            a.cubic.SetImmediate(0)
            a.elastic.SetImmediate(0)
        }},
    }
}
```

## Custom Interpolators

For custom types, register an interpolator:

```go
// Define a custom type
type Point struct {
    X, Y float64
}

// Register interpolator at init
func init() {
    terma.RegisterInterpolator(func(from, to Point, t float64) Point {
        return Point{
            X: from.X + (to.X-from.X)*t,
            Y: from.Y + (to.Y-from.Y)*t,
        }
    })
}

// Use in animation
anim := terma.NewAnimation(terma.AnimationConfig[Point]{
    From:     Point{0, 0},
    To:       Point{100, 100},
    Duration: 500 * time.Millisecond,
})
```

### Built-in Interpolators

These types work automatically:
- `float64`, `float32`
- `int`, `int64`
- `terma.Color` (smooth color blending)

## Complete Example

A demo application showing multiple animation types:

```go
package main

import (
    "fmt"
    "strings"
    "time"

    t "terma"
)

type App struct {
    // Animated progress bar
    progress *t.AnimatedValue[float64]

    // Color animation
    colorAnim *t.Animation[t.Color]

    // Loading spinner
    spinner *t.SpinnerState

    // Status
    running t.Signal[bool]
}

func NewApp() *App {
    spinner := t.NewSpinnerState(t.SpinnerDots)
    spinner.Start()

    return &App{
        progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
            Initial:  0,
            Duration: 300 * time.Millisecond,
            Easing:   t.EaseOutCubic,
        }),
        colorAnim: t.NewAnimation(t.AnimationConfig[t.Color]{
            From:     t.RGB(100, 100, 255),
            To:       t.RGB(255, 100, 100),
            Duration: 2 * time.Second,
            Easing:   t.EaseInOutSine,
        }),
        spinner: spinner,
        running: t.NewSignal(true),
    }
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()
    progress := a.progress.Get()
    color := a.colorAnim.Value().Get()

    return t.Column{
        Style: t.Style{Padding: t.EdgeInsetsAll(2)},
        Children: []t.Widget{
            // Title
            t.Text{
                Content: "Animation Demo",
                Style:   t.Style{ForegroundColor: theme.Primary},
            },
            t.Spacer{Height: t.Cells(1)},

            // Progress bar
            t.Text{Content: fmt.Sprintf("Progress: %.0f%%", progress)},
            t.Text{Content: a.renderProgressBar(progress, 30)},
            t.Spacer{Height: t.Cells(1)},

            // Color animation
            t.Text{Content: "Color Animation:"},
            t.Text{
                Content: "████████████████████",
                Style:   t.Style{ForegroundColor: color},
            },
            t.Spacer{Height: t.Cells(1)},

            // Spinner
            t.Row{
                Children: []t.Widget{
                    t.ShowWhen(a.running.Get(), t.Spinner{State: a.spinner}),
                    t.ShowWhen(a.running.Get(), t.Text{Content: " Processing..."}),
                    t.HideWhen(a.running.Get(), t.Text{Content: "Stopped"}),
                },
            },
        },
    }
}

func (a *App) renderProgressBar(value float64, width int) string {
    filled := int(value / 100 * float64(width))
    if filled > width {
        filled = width
    }
    return "[" + strings.Repeat("█", filled) + strings.Repeat("░", width-filled) + "]"
}

func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "+", Name: "Increase", Action: func() {
            current := a.progress.Target()
            if current < 100 {
                a.progress.Set(current + 10)
            }
        }},
        {Key: "-", Name: "Decrease", Action: func() {
            current := a.progress.Target()
            if current > 0 {
                a.progress.Set(current - 10)
            }
        }},
        {Key: "c", Name: "Color", Action: func() {
            a.colorAnim.Reset()
            a.colorAnim.Start()
        }},
        {Key: "space", Name: "Toggle", Action: func() {
            if a.running.Get() {
                a.spinner.Stop()
                a.running.Set(false)
            } else {
                a.spinner.Start()
                a.running.Set(true)
            }
        }},
        {Key: "q", Name: "Quit", Action: func() {
            // Handle quit
        }},
    }
}

func main() {
    t.Run(NewApp())
}
```

## Performance

The animation system is optimized for efficiency:

- **Lazy ticker**: Animation ticker only runs when animations are active
- **60 FPS default**: Configurable frame rate for smooth updates
- **Automatic cleanup**: Completed animations are automatically unregistered
- **Signal-based updates**: Only rebuilds widgets that depend on animated values

Animations are managed by a global controller created by `Run()`. You don't need to manually manage the animation loop.
