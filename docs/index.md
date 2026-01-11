# Terma

Build beautiful terminal UIs with Go. Declarative. Reactive. Simple.

```go
type App struct {
    count *terma.Signal[int]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    return terma.Column{
        Children: []terma.Widget{
            terma.Text{Content: fmt.Sprintf("Count: %d", a.count.Get())},
            terma.Button{
                ID:    "increment",
                Label: "Increment",
                OnPress: func() {
                    a.count.Set(a.count.Get() + 1)
                },
            },
        },
    }
}

func main() {
    terma.Run(&App{count: terma.NewSignal(0)})
}
```

No manual redraws. No event wiring. Just declare your UI and let Terma handle the rest.

## Why Terma?

- **Declarative** — Describe what your UI looks like, not how to update it
- **Reactive** — State changes automatically trigger rebuilds
- **Composable** — Build complex UIs from simple, reusable widgets
- **Familiar** — If you know React or Flutter, you'll feel right at home

## Documentation

- [Getting Started](getting-started.md) — Installation and building your first app
- [Widgets](widgets/index.md) — Overview of available widgets
- [Signals](signals.md) — Reactive state management
- [Layout](layout.md) — Layout system and dimensions
- [Styling](styling.md) — Colors, padding, margins, and theming
- [Focus & Keyboard](focus-keyboard.md) — Focus management and keybindings
- [Conditional Rendering](conditional.md) — ShowWhen, Switcher, and visibility control
- [Animation](animation.md) — Smooth transitions, spinners, and easing
- [Examples](examples.md) — Example applications
