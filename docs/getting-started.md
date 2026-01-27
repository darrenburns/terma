# Getting Started

Terma is a declarative terminal UI framework for Go. Instead of manually drawing characters to the screen, you describe *what* your UI should look like, and Terma handles the rendering.

This guide will teach you the fundamentals by building progressively more complex applications.

## Your First Terma App

Let's start with the simplest possible Terma application. Create this file or find it at `cmd/tutorial/01-hello-world/main.go`:

```go title="cmd/tutorial/01-hello-world/main.go"
--8<-- "cmd/tutorial/01-hello-world/main.go"
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

The name `App` is just a convention. You can call it whatever you like—`Counter`, `MyTUI`, `Dashboard`—as long as it has a `Build` method, Terma can render it.

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

```go title="cmd/tutorial/02-counter/main.go"
--8<-- "cmd/tutorial/02-counter/main.go"
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

### Styling Text with Markup

Notice the instruction text uses `ParseMarkupToText` instead of a plain `Text` widget:

```go
t.ParseMarkupToText("Press [b $Accent]Up[/] to increment...", theme)
```

This returns a `Text` widget with styled spans. The markup syntax is `[styles]text[/]` where:

- `b` makes text bold
- `$Accent`, `$Primary`, `$Error`, etc. apply theme colors
- `[/]` closes the styled section

You get the theme from the build context with `ctx.Theme()`. Using theme colors instead of hardcoded values ensures your app looks consistent and adapts to different color schemes.

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

The `Keybinds()` method returns a list of keyboard shortcuts. Any widget can implement this method, not just your root App—this lets you define context-specific shortcuts on individual components. Each keybind has:

- `Key`: The key to listen for (e.g., `"up"`, `"down"`, `"enter"`, `"q"`)
- `Name`: A human-readable description (shown in the KeybindBar widget, which we'll cover later)
- `Action`: A function to call when the key is pressed

When you press `Up`, the action runs: it reads the current count, adds 1, and sets the new value. Because `count` changed, Terma rebuilds the UI, and you see the updated number.

## Building a Progress Bar

Let's build something more visual using Terma's `ProgressBar` widget. This will teach you about using built-in widgets, conditional styling, and centering content.

<video autoplay loop muted playsinline src="../assets/progressbar-tutorial.mp4"></video>

```go title="cmd/tutorial/03-progress-bar/main.go"
--8<-- "cmd/tutorial/03-progress-bar/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/03-progress-bar
```

Press `Up` to fill the bar, `Down` to empty it, and `q` to quit. When the bar reaches 100%, it turns green.

### The ProgressBar Widget

```go
t.ProgressBar{
    Progress:    float64(progress) / float64(maxProgress),
    Width:       t.Flex(1),
    FilledColor: fillColor,
}
```

`ProgressBar` is a built-in widget that displays a horizontal bar. It takes:

- `Progress`: A float from 0.0 to 1.0 representing completion
- `Width`: How wide the bar should be (here we use `Flex(1)` to fill available space)
- `FilledColor`: The color of the filled portion (defaults to `theme.Primary`)
- `UnfilledColor`: The color of the empty portion (defaults to `theme.Surface`)

### Conditional Styling

```go
fillColor := theme.Primary
if progress == maxProgress {
    fillColor = theme.Success
}
```

This is standard Go—no special syntax needed. When the progress reaches the maximum, we switch from the primary color to the success color (green). Because `Build` is called on every state change, the color updates automatically when the bar fills up.

### Guarding State Changes

```go
{Key: "up", Name: "Increase", Action: func() {
    if a.progress.Get() < maxProgress {
        a.progress.Set(a.progress.Get() + 1)
    }
}},
```

We check bounds before updating the signal. This prevents the progress from going below 0 or above the maximum. You could also clamp values in `Build`, but checking at the source keeps your rendering logic clean.

### Centering Content

```go
return t.Column{
    Width:      t.Flex(1),
    Height:     t.Flex(1),
    MainAlign:  t.MainAxisCenter,
    CrossAlign: t.CrossAxisCenter,
    Style:      t.Style{BackgroundColor: theme.Background},
    Children:   []t.Widget{...},
}
```

To center content on screen, we use an outer `Column` that fills the entire terminal:

- `Width: t.Flex(1)` and `Height: t.Flex(1)` make the column expand to fill all available space
- `MainAlign: t.MainAxisCenter` centers children along the main axis (vertically for a Column)
- `CrossAlign: t.CrossAxisCenter` centers children along the cross axis (horizontally for a Column)
- `Style: t.Style{BackgroundColor: theme.Background}` sets the app's background color

The actual content is wrapped in an inner `Column` so it's treated as a single unit for alignment.

## Adding Animation

Our progress bar works, but the jumps between values feel abrupt. Let's add smooth animation with minimal changes. We'll also set a custom theme.

<video autoplay loop muted playsinline src="../assets/animation-tutorial.mp4"></video>

```go title="cmd/tutorial/04-animation/main.go"
--8<-- "cmd/tutorial/04-animation/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/04-animation
```

Press `Up` to increase by 20, `Down` to decrease. Watch how the bar smoothly animates between values instead of jumping instantly.

### From Signal to AnimatedValue

The key change is replacing `Signal[int]` with `AnimatedValue[float64]`:

```go
// Before: instant updates
progress t.Signal[int]

// After: smooth animations
progress *t.AnimatedValue[float64]
```

Create it with `NewAnimatedValue`:

```go
progress: t.NewAnimatedValue(t.AnimatedValueConfig[float64]{
    Initial:  0,
    Duration: 300 * time.Millisecond,
    Easing:   t.EaseOutCubic,
}),
```

The config specifies:

- `Initial`: Starting value
- `Duration`: How long the animation takes
- `Easing`: The animation curve (`EaseOutCubic` starts fast and slows down)

### Using AnimatedValue

The API is almost identical to Signal. Reading works the same—call `Get()`:

```go
progress := a.progress.Get()  // Returns the current animated value
```

Setting triggers a smooth animation from the current value to the new one:

```go
a.progress.Set(a.progress.Target() + 20)
```

Notice we use `Target()` instead of `Get()` when calculating the next value. `Target()` returns where the animation is heading, while `Get()` returns the current interpolated value. This prevents compounding errors if the user presses keys quickly during an animation.

### Setting a Theme

```go
t.SetTheme("catppuccin")
```

Terma includes several built-in themes. Available themes include `"catppuccin"`, `"dracula"`, `"tokyo-night"`, `"gruvbox"`, `"nord"`, and more—each with light and dark variants.

## The Declarative Model

This is the core idea of Terma:

1. **State lives in Signals** — Your data is stored in Signals on your app struct
2. **Build describes the UI** — Your `Build` method returns widgets that reflect the current state
3. **Actions update Signals** — User interactions (key presses, clicks) update Signal values
4. **Terma rebuilds automatically** — When a Signal changes, Terma calls `Build` again

You never manually update the screen. You never say "change this text to show 5". Instead, you update the `count` Signal, and the UI updates automatically because it depends on that Signal.

This is called **declarative UI**: you declare what the UI should look like for any given state, rather than imperatively describing how to transition between states.

## Building a TODO List

Let's put together what we've learned by building a TODO list application. We'll start simple and progressively add features: scrolling and empty state handling.

### Basic TODO List

<video autoplay loop muted playsinline src="../assets/todo-tutorial-a.mp4"></video>

```go title="cmd/tutorial/05a-todo-list/main.go"
--8<-- "cmd/tutorial/05a-todo-list/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/05a-todo-list
```

Use `Up`/`Down` or `j`/`k` to navigate, `a` to add a task, `d` to delete the selected task, and `q` to quit.

#### ListState: Managing List Data

```go
type App struct {
    listState *t.ListState[string]
    taskCount int
}
```

`ListState` is a specialized state container for lists. It holds both the items and the cursor position. Unlike a plain Signal, ListState provides methods for common list operations.

Create it with `NewListState`:

```go
listState: t.NewListState([]string{"Task 1", "Task 2", "Task 3"})
```

#### The List Widget

```go
t.List[string]{
    ID:    "todo-list",
    State: a.listState,
}
```

The `List` widget renders a navigable list of items. It requires a `State` field pointing to a `ListState`. The ID is optional but recommended for focus management.

List handles keyboard navigation automatically—you don't need to implement `OnKey` for basic up/down movement.

#### Modifying List Data

```go
// Add an item to the end
a.listState.Append(fmt.Sprintf("Task %d", a.taskCount))

// Remove the item at the cursor position
a.listState.RemoveAt(a.listState.CursorIndex.Peek())
```

`Append()` adds items to the list. `RemoveAt()` removes an item by index. We use `CursorIndex.Peek()` to get the current cursor position without subscribing to changes (since we're in an action callback, not in Build).

### Adding Scrolling

<video autoplay loop muted playsinline src="../assets/todo-tutorial-b.mp4"></video>

When your list grows beyond the available space, you need scrolling. Terma provides the `Scrollable` widget for this.

```go title="cmd/tutorial/05b-todo-scrollable/main.go"
--8<-- "cmd/tutorial/05b-todo-scrollable/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/05b-todo-scrollable
```

Add several tasks with `a` to see the scrollbar appear when the list exceeds 10 items.

#### ScrollState and Shared State

```go
type App struct {
    listState   *t.ListState[string]
    scrollState *t.ScrollState
    taskCount   int
}
```

We add a `ScrollState` to track the scroll position. This state is shared between `Scrollable` and `List` so they can coordinate—when you navigate to an item outside the viewport, the list tells the scrollable to scroll it into view.

Create it with `NewScrollState`:

```go
scrollState: t.NewScrollState()
```

#### Wrapping with Scrollable

```go
t.Scrollable{
    State:  a.scrollState,
    Height: t.Cells(10),
    Child: t.List[string]{
        ID:          "todo-list",
        State:       a.listState,
        ScrollState: a.scrollState,  // Share the state
    },
}
```

The pattern is:

1. Wrap your content with `Scrollable`
2. Pass the same `ScrollState` to both `Scrollable` and the child that needs scroll-into-view
3. Set a fixed `Height` on `Scrollable` to create a viewport

The `List` widget uses `ScrollState` to call `ScrollToView()` when the cursor moves, ensuring the selected item is always visible.

### Handling Empty State

<video autoplay loop muted playsinline src="../assets/todo-tutorial-c.mp4"></video>

A good user experience shows helpful messages when there's no data. Let's add empty state handling.

```go title="cmd/tutorial/05c-todo-empty-state/main.go"
--8<-- "cmd/tutorial/05c-todo-empty-state/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/05c-todo-empty-state
```

The app starts empty, showing a helpful message. Press `a` to add tasks. Press `c` to clear all tasks and see the empty state again.

#### Conditional Rendering with ShowWhen/HideWhen

```go
isEmpty := a.listState.ItemCount() == 0

t.ShowWhen(isEmpty, emptyMessage),
t.HideWhen(isEmpty, scrollableList),
```

`ShowWhen(condition, widget)` renders the widget when the condition is true, otherwise renders nothing (takes no space). `HideWhen` is the inverse.

This is different from making a widget invisible—`ShowWhen(false, widget)` completely removes the widget from layout, while `InvisibleWhen(true, widget)` reserves space but doesn't render.

#### Checking Item Count

```go
isEmpty := a.listState.ItemCount() == 0
```

`ItemCount()` returns the number of items in the list. We use this to determine whether to show the empty message or the list.

#### Guarding Operations

```go
{Key: "d", Name: "Delete", Action: func() {
    if a.listState.ItemCount() > 0 {
        a.listState.RemoveAt(a.listState.CursorIndex.Peek())
    }
}},
```

When the list might be empty, guard operations that would fail on an empty list. Here we check `ItemCount() > 0` before attempting to delete.

#### Clearing All Items

```go
{Key: "c", Name: "Clear", Action: func() {
    a.listState.Clear()
}},
```

`Clear()` removes all items from the list and resets the cursor to 0.

### Adding Multi-Select for Bulk Delete

<video autoplay loop muted playsinline src="../assets/todo-tutorial-d.mp4"></video>

With just a few changes, we can enable multi-select to delete multiple items at once.

```go title="cmd/tutorial/05d-todo-multiselect/main.go"
--8<-- "cmd/tutorial/05d-todo-multiselect/main.go"
```

Run it:

```bash
go run ./cmd/tutorial/05d-todo-multiselect
```

Press `Space` to toggle selection on items, then `d` to delete all selected items at once.

#### Enabling Multi-Select

```go
t.List[string]{
    ID:          "todo-list",
    State:       a.listState,
    ScrollState: a.scrollState,
    MultiSelect: true,
}
```

Just add `MultiSelect: true` to the List widget. This enables:

- `Space` to toggle selection on the current item
- `Shift+j/k` (capital `J`/`K`) to extend selection while navigating

#### Bulk Delete with SelectedIndices

```go
{Key: "d", Name: "Delete", Action: func() {
    indices := a.listState.SelectedIndices()
    if len(indices) > 0 {
        // Delete selected items (in reverse order to preserve indices)
        for i := len(indices) - 1; i >= 0; i-- {
            a.listState.RemoveAt(indices[i])
        }
        a.listState.ClearSelection()
    } else if a.listState.ItemCount() > 0 {
        // No selection - delete item at cursor
        a.listState.RemoveAt(a.listState.CursorIndex.Peek())
    }
}},
```

`SelectedIndices()` returns the indices of all selected items in ascending order. We delete in reverse order so that removing an item doesn't shift the indices of items we still need to delete. After deleting, `ClearSelection()` removes all selection state.

If nothing is selected, we fall back to deleting the item at the cursor—giving users a consistent experience whether they use selection or not.

### What We've Learned

In this section, you learned:

- **ListState** manages list data and cursor position
- **List** renders a navigable, focusable list of items
- **Append/RemoveAt/Clear** modify list contents
- **CursorIndex.Peek()** gets the cursor position in callbacks
- **ScrollState** tracks scroll position
- **Scrollable** adds scrolling with a viewport
- **Shared state** between Scrollable and List enables scroll-into-view
- **ShowWhen/HideWhen** conditionally include or exclude widgets
- **ItemCount()** checks if a list is empty
- **MultiSelect** enables selection of multiple items
- **SelectedIndices/ClearSelection** manage multi-select state

## Summary

You've learned:

- **Widgets** are the building blocks of Terma UIs. Your app is a widget.
- **Build** returns a tree of widgets that describes your UI.
- **Signals** hold reactive state. Read with `Get()`, write with `Set()`.
- **AnimatedValue** adds smooth transitions—just swap `Signal` for `AnimatedValue`.
- **Column** arranges children vertically, **Row** arranges them horizontally.
- **Flex dimensions** (`Flex(1)`) make widgets expand to fill available space.
- **Alignment** (`MainAlign`, `CrossAlign`) controls how children are positioned.
- **Style** customizes appearance with colors, padding, and more.
- **Themes** can be set with `SetTheme()` for different color schemes.
- **Markup** styles text inline with `[b $Color]text[/]` syntax.
- **Keybinds** define keyboard shortcuts with actions that update state.
- **Conditional logic** in `Build` lets you change appearance based on state.
- **Declarative UI** means you describe the desired state, not the transitions.

## Next Steps

Now that you understand the basics, explore the other sections of the documentation:

- [Signals](signals.md) — Deep dive into reactive state management
- [Layout](layout/index.md) — Learn about Column, Row, and other layout widgets
- [Styling](styling.md) — Add colors, padding, borders, and more
- [Focus & Keyboard](focus-keyboard.md) — Handle keyboard navigation and focus
