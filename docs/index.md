# Terma

Terma is a Go library for terminal user interfaces. Compose widgets and update their state in event handlers.

```go
--8<-- "docs/minimal-examples/widget-counter/main.go"
```

Reading a signal in `Build()` subscribes that widget to rebuilds when the value changes.
For text that changes frequently, use the [reactive text helpers](widgets/presentedtext.md) to avoid rebuilding the surrounding widgets.

## Documentation

- [Getting Started](getting-started.md): Installation and building your first app
- [Hello World](getting-started.md#your-first-terma-app): Your first widget and Run()
- [Widgets](widgets/index.md): Overview of available widgets
- [Signals](signals.md): Reactive state management
- [Async Tasks](async.md): Run blocking work without freezing the UI
- [Layout](layout/index.md): Layout system and dimensions
- [Styling](styling.md): Colors, padding, margins, and theming
- [Focus & Keyboard](focus-keyboard.md): Focus management and keybindings
- [Conditional Rendering](conditional.md): ShowWhen, Switcher, and visibility control
- [Animation](animation.md): Smooth transitions, spinners, and easing
- [Floating](floating.md): Overlays, modals, dropdowns, and tooltips
- [Examples](examples.md): Example applications
