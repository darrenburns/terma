# Jumper

`Jumper` adds a jump mode to your app, inspired by [Posting](https://github.com/darrenburns/posting)'s jump mode and Vimium's link hints. Press `ctrl+o` and a short label appears over each target; type a label and focus moves straight there.

```go
type App struct {
    jump *terma.JumpState
    // ...
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
    return terma.Jumper{
        State: a.jump, // terma.NewJumpState()
        Targets: []terma.JumpTarget{
            {Key: "1", ID: "collections"},
            {Key: "2", ID: "url"},
            {Key: "s", ID: "send", Action: a.send},
        },
        Dynamic: true,
        Child:   a.body(ctx),
    }
}
```

Run `go run ./cmd/jump-example` to try it.

## Static and dynamic targets

**Static jump maps** (`Targets`) bind a key to a widget ID. The same key always goes to the same place, so users learn them, the way Posting's number keys work.

- If the widget is focusable, the jump focuses it.
- If it isn't focusable (a panel, say), the jump focuses the first focusable widget inside it.
- If `Action` is set, it runs instead of moving focus, for example to switch tabs or press a button. The label still appears on the widget with `ID`.
- A key can be longer than one character (`"gt"`). It is typed one character at a time. Don't let one key be a prefix of another.

**Dynamic hints** (`Dynamic: true`) are for what can't be given a key in advance, because it comes from data. Every **item** in view gets its own hint, the way Vimium labels every link on a page:

| Widget | Each hint lands on | Jumping to it |
|--------|--------------------|---------------|
| `List` | a visible row | moves the cursor there (as a click does) and focuses the list |
| `Tree` | a visible node | moves the cursor there and focuses the tree |
| `Table` | a visible row, or a cell in cursor selection mode | moves the cursor there and focuses the table |
| `TabBar` | a tab | activates it |

Rows scrolled out of view get no hint. Any other focusable widget without a static key gets a hint of its own.

To make your own data-driven widgets jump targets (cards, links, search results), implement `Jumpable`. Jumping calls `Jump()`, then focuses the focusable widget that contains the item, if there is one:

```go
func (c ResultCard) Jump() { c.results.Select(c.index) }
```

How hints are generated works like Vimium:

- Hints are made from `Hints` (default `asdfghjklqwertyuiopzxcvbnm`, home row first).
- Hints are as short as the number of targets allows.
- They are assigned in reading order, top to bottom and then left to right.
- No hint starts with another hint or with a static key, so typing one always leads to exactly one target.
- With more targets than characters, some hints get two characters. The labels narrow as each character is typed, and `backspace` undoes the last one.

The two combine. Static keys go to the fixed landmarks of your layout, and dynamic hints go to whatever content happens to be on screen. A list can have both: a static key on the list itself, and a hint on each row.

## Behaviour

- **Visibility:** only widgets currently on screen get a label. Widgets scrolled out of view are skipped.
- **Focus traps:** while focus is inside a modal (or any `FocusTrap`), only targets inside it are labelled.
- **Stacking and focus:** the labels sit above every other overlay, dialogs included. Jump mode doesn't take focus, so the focused widget keeps its focus styling while you choose.
- **Leaving:** `escape`, the toggle key, a click, or any key that matches no label leaves jump mode without jumping.
- **The toggle key:** `Key` (default `ctrl+o`) works wherever focus is, including inside dialogs. It also appears in the `KeybindBar` while focus is inside `Child`. To enter jump mode another way, call `State.Activate()` from your own keybind.
- **Reading jump state:** `State.IsActive()` and `State.Typed()` are reactive, so a header or status bar can show that jump mode is on.

## Fields

| Field | Purpose |
|-------|---------|
| `State` | Required. `*JumpState` from `NewJumpState()` |
| `Targets` | Static jump map: `[]JumpTarget{Key, ID, Action}` |
| `Dynamic` | Hint every List/Tree/Table row, tab and `Jumpable` in view, and any focusable widget without a static key |
| `Hints` | Characters dynamic hints are made from |
| `Key` | Toggle key, default `ctrl+o` |
| `Child` | The content. `Jumper` never affects layout |
| `LabelStyle` | Label colors (defaults to the theme's accent) |
| `Backdrop` | Tint over the screen while jumping. Use a fully transparent color to turn it off |
