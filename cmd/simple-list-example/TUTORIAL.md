# Building a Simple List with Terma

This tutorial walks through building a minimal interactive list. By the end, you'll understand the core concepts of Terma: widgets, signals, and declarative UI.

## The App is a Widget

In Terma, everything is a widget—including your application itself. A widget is any type that implements the `Build` method:

```go
func (d *SimpleListDemo) Build(ctx t.BuildContext) t.Widget
```

The `Build` method returns a tree of widgets that describes your UI. When you call `t.Run(app)`, Terma renders this tree to the terminal and handles user input.

```go
type SimpleListDemo struct {
    cursorIndex t.Signal[int]
    selectedMsg t.Signal[string]
}
```

Our app struct holds the state that drives the UI. Notice we're not storing UI elements—just data.

## Declarative UI with Signals

Terma uses a declarative model: you never update the UI directly. Instead, you update *state*, and the UI rebuilds automatically.

This is where **Signals** come in. A Signal wraps a value and notifies Terma when it changes:

```go
func NewSimpleListDemo() *SimpleListDemo {
    return &SimpleListDemo{
        cursorIndex: t.NewSignal(0),
        selectedMsg: t.NewSignal("No selection yet"),
    }
}
```

When a Signal's value changes, any widget that read from that Signal during its last build will automatically rebuild. You don't wire up subscriptions manually—Terma tracks dependencies for you.

To read a Signal, call `Get()`. To write, call `Set()`:

```go
d.selectedMsg.Get()                              // read
d.selectedMsg.Set("You selected: Apple")         // write
```

## Layout with Column and Row

The `Column` widget arranges children vertically. Its counterpart, `Row`, arranges children horizontally. Here's our root layout using a Column:

```go
return t.Column{
    ID:      "simple-list-root",
    Spacing: 1,
    Style: t.Style{
        Padding: t.EdgeInsetsXY(2, 1),
    },
    Children: []t.Widget{
        // ...
    },
}
```

`Spacing: 1` adds one blank line between each child. The `Style` field applies visual properties—here we add horizontal padding of 2 cells and vertical padding of 1 cell with `EdgeInsetsXY`.

## Displaying Text

The `Text` widget displays content. At its simplest:

```go
t.Text{
    Content: "Simple String List Example",
    Style: t.Style{
        ForegroundColor: t.Black,
        BackgroundColor: t.Cyan,
        Padding:         t.EdgeInsetsXY(1, 0),
    },
}
```

For richer formatting, use `Spans` instead of `Content`. Each Span can have its own style:

```go
t.Text{
    Spans: []t.Span{
        t.PlainSpan("Use "),
        t.BoldSpan("↑/↓", t.BrightCyan),
        t.PlainSpan(" to navigate"),
    },
}
```

`PlainSpan` creates unstyled text. `BoldSpan` creates bold text with a specified color. You can combine as many spans as needed.

## The List Widget

`List` is a generic widget that displays a slice of items with keyboard navigation. First, we define our data:

```go
items := []string{
    "Apple",
    "Banana",
    "Cherry",
    "Date",
    "Elderberry",
}
```

Then we create the List, passing in this slice:

```go
&t.List[string]{
    ID:          "simple-string-list",
    Items:       items,
    CursorIndex: d.cursorIndex,
    OnSelect: func(item string) {
        d.selectedMsg.Set(fmt.Sprintf("You selected: %s", item))
    },
}
```

A few things to note:

**Generic type parameter**: `List[string]` works with any type. You could use `List[User]` or `List[MenuItem]` just as easily.

**External cursor state**: We pass in `d.cursorIndex`, a Signal that the List reads and writes. When the user presses ↓, the List updates this Signal, which triggers a rebuild. This keeps cursor state outside the List itself—useful when you need to access or modify it elsewhere.

**Selection callback**: `OnSelect` fires when the user presses Enter. Here we update `selectedMsg`, which causes the status text below the list to rebuild with the new message.

**Default rendering**: Without a `RenderItem` function, List renders each item using its string representation. For custom layouts—like showing a title and description for each item—you'd supply a `RenderItem` function that maps each item to a widget tree.

**Scrolling**: For lists longer than the available space, you can wrap the List in a `Scrollable` and provide a `ScrollController`. This example doesn't need scrolling with only five items.

## Running the App

The `main` function creates the app and runs it:

```go
func main() {
    app := NewSimpleListDemo()
    if err := t.Run(app); err != nil {
        log.Fatal(err)
    }
}
```

`t.Run` takes over the terminal, renders your widget tree, and enters an event loop. It returns when the user exits (Ctrl+C).

## Summary

The key ideas:

1. **Widgets** describe UI. Your app is a widget that returns other widgets from `Build`.
2. **Signals** hold state. Update a Signal, and dependent widgets rebuild automatically.
3. **Declarative** means you describe what the UI should look like given the current state—not how to transition between states.
4. **Layout widgets** like `Column` and `Row` arrange children vertically and horizontally. Use `Spacing` and `Style` to control appearance.
5. **List** handles keyboard navigation and selection. Pass a cursor Signal for state, and an `OnSelect` callback for actions.
