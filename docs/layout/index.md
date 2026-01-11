# Layout

Terma uses a constraint-based layout system with two passes: parents provide constraints to children, children compute their sizes, and parents position children based on alignment and spacing.

## Layout Widgets

- [Row & Column](row-column.md) - Linear layouts (horizontal and vertical)
- [Dock](dock.md) - Edge-docking layout for app shells
- [Scrollable](scrollable.md) - Scrolling container with scrollbar
- [Spacer](spacer.md) - Flexible empty space

## Dimensions

Every widget can specify its width and height using one of three dimension types.

### Auto

Fits the content exactly. The widget measures its content and takes only the space it needs.

```go
// Button takes only the space needed for its label
Button{
    ID:    "submit",
    Label: "Submit",
    Width: Auto,  // Fits the text "Submit"
}
```

### Cells

Fixed size in terminal cells. Use when you need precise control over dimensions.

```go
// Fixed 30-cell wide sidebar
Column{
    Width: Cells(30),
    Children: []Widget{...},
}
```

### Flex

Proportional space distribution. Flex widgets share remaining space after Auto and Cells widgets are sized.

The flex value determines the proportion. `Flex(1)` and `Flex(2)` siblings receive space in a 1:2 ratio.

```go
// Sidebar (1/3) and main content (2/3) split available space
Row{
    Children: []Widget{
        Column{Width: Flex(1), Children: []Widget{sidebar}},
        Column{Width: Flex(2), Children: []Widget{main}},
    },
}
```

```
Available width: 60 cells
┌──────────────────┬────────────────────────────────────────┐
│    Flex(1)       │               Flex(2)                  │
│    20 cells      │               40 cells                 │
└──────────────────┴────────────────────────────────────────┘
       1/3                          2/3
```

### Dimension Comparison

| Dimension | Syntax | Behavior | Use Case |
|-----------|--------|----------|----------|
| Auto | `Auto` | Fit content exactly | Text, buttons, leaf widgets |
| Cells | `Cells(n)` | Fixed n terminal cells | Precise sizing, fixed sidebars |
| Flex | `Flex(n)` | Proportional share of remaining space | Responsive layouts, fill space |

### Default Dimensions

When you don't specify dimensions, widgets use sensible defaults:

- Most widgets default to `Auto` (fit their content)
- `Dock` defaults to `Flex(1)` for both dimensions
- `Spacer` defaults to `Flex(1)` for both dimensions

## Spacing, Padding, and Margin

Three ways to add space in your layouts:

### Spacing

Uniform gap between children in a Row or Column. Specified in cells.

```go
Column{
    Spacing: 2,  // 2-cell gap between each child
    Children: []Widget{
        Text{Content: "Item 1"},
        Text{Content: "Item 2"},
        Text{Content: "Item 3"},
    },
}
```

### Padding

Space inside a widget, between its border and its content.

```go
Column{
    Style: Style{
        Padding: EdgeInsetsAll(2),  // 2-cell padding on all sides
    },
    Children: []Widget{...},
}
```

### Margin

Space outside a widget, around its border.

```go
Text{
    Content: "With margin",
    Style: Style{
        Margin: EdgeInsetsHV(4, 1),  // 4 horizontal, 1 vertical
    },
}
```

### EdgeInsets Helpers

| Helper | Description |
|--------|-------------|
| `EdgeInsetsAll(n)` | Same value for all 4 sides |
| `EdgeInsetsHV(h, v)` | Horizontal and vertical values |
| `EdgeInsetsTRBL(t, r, b, l)` | Top, right, bottom, left individually |

### Visual Comparison

```
┌─ Margin ───────────────────────────────────────────────────────┐
│                                                                │
│    ┌─ Border ───────────────────────────────────────────┐      │
│    │                                                    │      │
│    │    ┌─ Padding ──────────────────────────────┐      │      │
│    │    │                                        │      │      │
│    │    │    ┌─────────┐          ┌─────────┐    │      │      │
│    │    │    │ Child A │ Spacing  │ Child B │    │      │      │
│    │    │    └─────────┘    ↔     └─────────┘    │      │      │
│    │    │                                        │      │      │
│    │    └────────────────────────────────────────┘      │      │
│    │                                                    │      │
│    └────────────────────────────────────────────────────┘      │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

### When to Use Each

| Method | Applied To | Use Case |
|--------|------------|----------|
| `Spacing` | Row/Column children | Uniform gaps between all children |
| `Padding` | Any widget | Space inside, around content |
| `Margin` | Any widget | Space outside, around widget |
| `Spacer{}` | Specific gap | Variable or precise gap between specific children |

## Common Layout Patterns

### Header, Body, Footer

```go
Dock{
    Top:    []Widget{Header{}},
    Bottom: []Widget{Footer{}},
    Body:   Content{},
}
```

```
┌─────────────────────────┐
│░░░░░░░ Header ░░░░░░░░░░│
├─────────────────────────┤
│                         │
│          Body           │
│                         │
├─────────────────────────┤
│░░░░░░░ Footer ░░░░░░░░░░│
└─────────────────────────┘
```

### Sidebar Layout

```go
Row{
    Children: []Widget{
        Column{Width: Cells(25), Children: []Widget{Sidebar{}}},
        Column{Width: Flex(1), Children: []Widget{MainContent{}}},
    },
}
```

```
┌─────────┬───────────────────────┐
│░░░░░░░░░│                       │
│░Sidebar░│      Main Content     │
│░░░░░░░░░│                       │
└─────────┴───────────────────────┘
   Fixed            Flex(1)
```

### Centered Content

```go
Column{
    Width:      Flex(1),
    Height:     Flex(1),
    MainAlign:  MainAxisCenter,
    CrossAlign: CrossAxisCenter,
    Children: []Widget{
        Text{Content: "Centered!"},
    },
}
```

### Push to Edges

```go
Row{
    Children: []Widget{
        Text{Content: "Left"},
        Spacer{},
        Text{Content: "Right"},
    },
}
```

## Layout Decision Guide

| Use Case | Widget/Approach |
|----------|-----------------|
| Horizontal arrangement | `Row` |
| Vertical arrangement | `Column` |
| App shell (header/footer/sidebar) | `Dock` |
| Scrolling content | `Scrollable` |
| Push items apart | `Spacer` |
| Uniform gaps | `Spacing` field |
| Fill available space | `Flex(n)` dimension |
| Fixed size | `Cells(n)` dimension |
| Fit content | `Auto` dimension |
