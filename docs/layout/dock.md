# Dock

Edge-docking layout for building app shells with headers, footers, and sidebars. Widgets dock to edges in order, consuming space, with the body filling the remainder.

```go
Dock{
    Top:    []Widget{Header{}},
    Bottom: []Widget{StatusBar{}},
    Left:   []Widget{Sidebar{}},
    Body:   MainContent{},
}
```

```
┌─────────────────────────────────┐
│░░░░░░░░░░░░ Top ░░░░░░░░░░░░░░░░│
├───────┬─────────────────┬───────┤
│░░░░░░░│                 │░░░░░░░│
│░Left░░│      Body       │░Right░│
│░░░░░░░│                 │░░░░░░░│
├───────┴─────────────────┴───────┤
│░░░░░░░░░░░ Bottom ░░░░░░░░░░░░░░│
└─────────────────────────────────┘
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `Top` | `[]Widget` | `nil` | Widgets docked to top edge |
| `Bottom` | `[]Widget` | `nil` | Widgets docked to bottom edge |
| `Left` | `[]Widget` | `nil` | Widgets docked to left edge |
| `Right` | `[]Widget` | `nil` | Widgets docked to right edge |
| `Body` | `Widget` | `nil` | Fills remaining space |
| `DockOrder` | `[]Edge` | `[Top, Bottom, Left, Right]` | Edge processing order |
| `Width` | `Dimension` | `Flex(1)` | Width dimension |
| `Height` | `Dimension` | `Flex(1)` | Height dimension |
| `Style` | `Style` | `Style{}` | Padding, margin, border |

## How Docking Works

Edges are processed in order (default: Top, Bottom, Left, Right). Each edge consumes space from that direction, and the body fills whatever remains.

```
Step 1: Top consumes      Step 2: Bottom consumes   Step 3: Left consumes     Step 4: Body fills rest
┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
│░░░░░░ Top ░░░░░░░░░░│   │░░░░░░ Top ░░░░░░░░░░│   │░░░░░░ Top ░░░░░░░░░░│   │░░░░░░ Top ░░░░░░░░░░│
├─────────────────────┤   ├─────────────────────┤   ├─────┬───────────────┤   ├─────┬───────────────┤
│                     │   │                     │   │░░░░░│               │   │░░░░░│               │
│                     │   │                     │   │░░░░░│               │   │░░░░░│     Body      │
│                     │   │                     │   │Left │               │   │Left │               │
│                     │   ├─────────────────────┤   ├─────┴───────────────┤   ├─────┴───────────────┤
│                     │   │░░░░ Bottom ░░░░░░░░░│   │░░░░ Bottom ░░░░░░░░░│   │░░░░ Bottom ░░░░░░░░░│
└─────────────────────┘   └─────────────────────┘   └─────────────────────┘   └─────────────────────┘
```

## DockOrder

Control which edges get priority by changing the processing order.

```go
// Left sidebar gets full height
Dock{
    DockOrder: []Edge{Left, Right, Top, Bottom},
    Left:      []Widget{Sidebar{}},
    Top:       []Widget{Header{}},
    Body:      Content{},
}
```

```
DockOrder: [Left, Right, Top, Bottom] - Left gets full height

┌─────┬─────────────────────┐
│░░░░░│░░░░░░ Top ░░░░░░░░░░│
│░░░░░├─────────────────────┤
│░░░░░│                     │
│Left │        Body         │
│░░░░░│                     │
│░░░░░├─────────────────────┤
│░░░░░│░░░░ Bottom ░░░░░░░░░│
└─────┴─────────────────────┘
```

Default order: `[Top, Bottom, Left, Right]`

## Examples

### Basic App Shell

```go
func (a *App) Build(ctx BuildContext) Widget {
    theme := ctx.Theme()

    return Dock{
        Top: []Widget{
            Text{
                Content: "My App",
                Style: Style{
                    BackgroundColor: theme.Primary,
                    ForegroundColor: theme.Primary.AutoText(),
                    Padding:         EdgeInsetsHV(2, 0),
                },
            },
        },
        Bottom: []Widget{
            KeybindBar{},
        },
        Body: Column{
            Style: Style{Padding: EdgeInsetsAll(1)},
            Children: []Widget{
                Text{Content: "Main content area"},
            },
        },
    }
}
```

### With Sidebar

```go
Dock{
    Top: []Widget{Header{}},
    Bottom: []Widget{StatusBar{}},
    Left: []Widget{
        Column{
            Width: Cells(25),
            Style: Style{
                BackgroundColor: theme.Surface,
                Padding:         EdgeInsetsAll(1),
            },
            Children: []Widget{
                Text{Content: "Navigation"},
                List[string]{State: navState},
            },
        },
    },
    Body: MainContent{},
}
```

### Multiple Widgets Per Edge

Each edge accepts a slice of widgets, stacked in the dock direction:

```go
Dock{
    Top: []Widget{
        MenuBar{},      // First row
        Toolbar{},      // Second row
        Breadcrumbs{},  // Third row
    },
    Body: Content{},
}
```

### Full-Height Sidebar

Use `DockOrder` to give the sidebar priority:

```go
Dock{
    DockOrder: []Edge{Left, Top, Bottom, Right},
    Left: []Widget{
        Column{
            Width: Cells(30),
            Children: []Widget{
                Logo{},
                Navigation{},
                Spacer{},
                UserMenu{},
            },
        },
    },
    Top: []Widget{SearchBar{}},
    Body: Content{},
}
```

### Nested Docks

Docks can be nested for complex layouts:

```go
Dock{
    Top: []Widget{GlobalHeader{}},
    Body: Dock{
        Left: []Widget{Sidebar{}},
        Body: Dock{
            Top:  []Widget{LocalToolbar{}},
            Body: Editor{},
        },
    },
}
```

## Dock vs Row/Column

| Use Case | Dock | Row/Column |
|----------|------|------------|
| App shell with header/footer | Preferred | Possible |
| Edges consume only needed space | Yes | Manual sizing |
| Body fills remainder | Automatic | Needs `Flex(1)` |
| Edge processing order control | `DockOrder` | N/A |
| Simple linear arrangement | Overkill | Preferred |

**Use Dock when:**

- Building app shells with fixed headers, footers, or sidebars
- You want edges to automatically size to their content
- The body should fill all remaining space

**Use Row/Column when:**

- Simple linear arrangement of widgets
- You need alignment control (MainAlign, CrossAlign)
- Proportional sizing between siblings
