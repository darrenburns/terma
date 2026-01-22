# Chapter 1: Getting Started with Terma

Terma is a declarative terminal UI framework for Go. Instead of manually drawing characters to the screen, you describe *what* your UI should look like, and Terma handles the rendering.

This tutorial will teach you the fundamentals by building progressively more complex applications.

## Your First Terma App

Let's start with the simplest possible Terma application:

```go
// cmd/tutorial/01-hello-world/main.go
```

Run it with:

```bash
go run ./cmd/tutorial/01-hello-world
```

You'll see "Hello, Terma!" in your terminal. Press `Ctrl+C` to exit.

Let's break down what's happening.

### The App Struct

```go
type App struct{}
```

In Terma, your application is a **widget**. A widget is any Go type that has a `Build` method. The `App` struct is your root widget—it represents the entire application.

### The Build Method

```go
func (a *App) Build(ctx t.BuildContext) t.Widget {
    return t.Text{Content: "Hello, Terma!"}
}
```

The `Build` method returns a **widget tree** that describes your UI. Here we're returning a single `Text` widget. When Terma needs to render your app, it calls `Build` and draws whatever widgets you return.

The `ctx` parameter provides access to things like theming and focus state—we'll use it later.

### Running the App

```go
func main() {
    app := &App{}
    if err := t.Run(app); err != nil {
        log.Fatal(err)
    }
}
```

`t.Run(app)` takes over your terminal, renders your widget tree, and starts the event loop. The event loop handles keyboard input, window resizing, and other events. It returns when the user quits (usually `Ctrl+C`).

## Adding Interactivity with Signals

A static "Hello World" isn't very useful. Let's build something interactive: a counter.

```go
// cmd/tutorial/02-counter/main.go
```

Run it:

```bash
go run ./cmd/tutorial/02-counter
```

Press `Up` to increment, `Down` to decrement, and `q` to quit.

### Signals: Reactive State

The key addition here is the **Signal**:

```go
type App struct {
    count t.Signal[int]
}
```

A Signal wraps a value and tracks when it changes. When you read from a Signal inside `Build`, Terma remembers that your widget depends on that Signal. When the Signal's value changes, Terma automatically rebuilds your widget.

Create a Signal with `NewSignal`:

```go
app := &App{
    count: t.NewSignal(0),  // Initial value is 0
}
```

Read the current value with `Get()`:

```go
a.count.Get()  // Returns the current count
```

Update the value with `Set()`:

```go
a.count.Set(a.count.Get() + 1)  // Increment by 1
```

### Layout with Column

```go
return t.Column{
    Spacing: 1,
    Children: []t.Widget{
        t.Text{Content: fmt.Sprintf("Count: %d", a.count.Get())},
        t.Text{Content: "Press Up to increment..."},
    },
}
```

`Column` arranges its children vertically. `Spacing: 1` adds one blank line between each child. There's also `Row` for horizontal layout—we'll cover that later.

Notice how we use `a.count.Get()` inside `Build`. This creates a dependency: whenever `count` changes, Terma will call `Build` again, and the Text widget will display the new value.

### Handling Keyboard Input

```go
func (a *App) Keybinds() []t.Keybind {
    return []t.Keybind{
        {Key: "up", Name: "Increment", Action: func() {
            a.count.Set(a.count.Get() + 1)
        }},
        {Key: "down", Name: "Decrement", Action: func() {
            a.count.Set(a.count.Get() - 1)
        }},
        {Key: "q", Name: "Quit", Action: t.Quit},
    }
}
```

The `Keybinds()` method returns a list of keyboard shortcuts. Each keybind has:

- `Key`: The key to listen for (e.g., `"up"`, `"down"`, `"enter"`, `"q"`)
- `Name`: A human-readable description (shown in the KeybindBar widget, which we'll cover later)
- `Action`: A function to call when the key is pressed

When you press `Up`, the action runs: it reads the current count, adds 1, and sets the new value. Because `count` changed, Terma rebuilds the UI, and you see the updated number.

## The Declarative Model

This is the core idea of Terma:

1. **State lives in Signals** — Your data is stored in Signals on your app struct
2. **Build describes the UI** — Your `Build` method returns widgets that reflect the current state
3. **Actions update Signals** — User interactions (key presses, clicks) update Signal values
4. **Terma rebuilds automatically** — When a Signal changes, Terma calls `Build` again

You never manually update the screen. You never say "change this text to show 5". Instead, you update the `count` Signal, and the UI updates automatically because it depends on that Signal.

This is called **declarative UI**: you declare what the UI should look like for any given state, rather than imperatively describing how to transition between states.

## Summary

You've learned:

- **Widgets** are the building blocks of Terma UIs. Your app is a widget.
- **Build** returns a tree of widgets that describes your UI.
- **Signals** hold reactive state. Read with `Get()`, write with `Set()`.
- **Column** arranges children vertically with optional spacing.
- **Keybinds** define keyboard shortcuts with actions that update state.
- **Declarative UI** means you describe the desired state, not the transitions.

In the next chapter, we'll build a list with selection—a common pattern in terminal apps.
