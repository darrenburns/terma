# Layout

Terma uses a constraint-based layout system with two passes: parents provide constraints to children, children compute their sizes, and parents position children based on alignment and spacing.

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

=== "Code"

    ```go
    // Sidebar (1/3) and main content (2/3) split available space
    Row{
        Children: []Widget{
            Column{Width: Flex(1), Children: []Widget{sidebar}},
            Column{Width: Flex(2), Children: []Widget{main}},
        },
    }
    ```

=== "Diagram"

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

## Row and Column

The primary layout widgets for arranging children linearly.

### Row

Arranges children horizontally (left to right).

=== "Code"

    ```go
    Row{
        Children: []Widget{
            Text{Content: "Left"},
            Text{Content: "Center"},
            Text{Content: "Right"},
        },
    }
    ```

=== "Diagram"

    ```
    Row (main axis →)
    ┌─────────────────────────────────────┐
    │  [A]    [B]    [C]        →         │
    └─────────────────────────────────────┘
    ```

### Column

Arranges children vertically (top to bottom).

=== "Code"

    ```go
    Column{
        Children: []Widget{
            Text{Content: "Top"},
            Text{Content: "Middle"},
            Text{Content: "Bottom"},
        },
    }
    ```

=== "Diagram"

    ```
    Column (main axis ↓)
    ┌─────────┐
    │  [A]    │
    │   ↓     │
    │  [B]    │
    │   ↓     │
    │  [C]    │
    └─────────┘
    ```

### Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `Children` | `[]Widget` | `nil` | Widgets to arrange |
| `Spacing` | `int` | `0` | Gap between children in cells |
| `MainAlign` | `MainAxisAlign` | `MainAxisStart` | Alignment along main axis |
| `CrossAlign` | `CrossAxisAlign` | `CrossAxisStretch` | Alignment along cross axis |
| `Width` | `Dimension` | unset | Width dimension |
| `Height` | `Dimension` | unset | Height dimension |
| `Style` | `Style` | `Style{}` | Padding, margin, border, colors |

## Main and Cross Axis

Linear layouts have two axes:

- **Main axis**: The direction children are arranged
- **Cross axis**: Perpendicular to the main axis

| Widget | Main Axis | Cross Axis |
|--------|-----------|------------|
| Row | Horizontal (left-right) | Vertical (top-bottom) |
| Column | Vertical (top-bottom) | Horizontal (left-right) |

### Main Axis Alignment

Controls how children are distributed along the main axis when there's extra space.

| Value | Description |
|-------|-------------|
| `MainAxisStart` | Pack at the start (default) |
| `MainAxisCenter` | Center children |
| `MainAxisEnd` | Pack at the end |

=== "Code"

    ```go
    // Center children horizontally in a row
    Row{
        MainAlign: MainAxisCenter,
        Children:  []Widget{...},
    }

    // Push children to the bottom of a column
    Column{
        Height:    Flex(1),
        MainAlign: MainAxisEnd,
        Children:  []Widget{...},
    }
    ```

=== "Diagram"

    ```
    MainAxisStart (default)          MainAxisCenter              MainAxisEnd
    ┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
    │ [A] [B] [C]         │     │      [A] [B] [C]    │     │         [A] [B] [C] │
    └─────────────────────┘     └─────────────────────┘     └─────────────────────┘
    ```

### Cross Axis Alignment

Controls how children are positioned along the cross axis.

| Value | Description |
|-------|-------------|
| `CrossAxisStretch` | Stretch to fill cross axis (default) |
| `CrossAxisStart` | Align at start of cross axis |
| `CrossAxisCenter` | Center along cross axis |
| `CrossAxisEnd` | Align at end of cross axis |

Cross axis alignment in a Row (cross axis is vertical):

=== "Code"

    ```go
    // Center items vertically in a row
    Row{
        Height:     Cells(10),
        CrossAlign: CrossAxisCenter,
        Children:   []Widget{...},
    }

    // Align items to the right in a column
    Column{
        Width:      Flex(1),
        CrossAlign: CrossAxisEnd,
        Children:   []Widget{...},
    }
    ```

=== "Diagram"

    ```
    CrossAxisStretch        CrossAxisStart        CrossAxisCenter         CrossAxisEnd
    ┌───────────────┐       ┌───────────────┐     ┌───────────────┐      ┌───────────────┐
    │███████████████│       │ [A]  [B]  [C] │     │               │      │               │
    │███████████████│       │               │     │ [A]  [B]  [C] │      │               │
    │███████████████│       │               │     │               │      │ [A]  [B]  [C] │
    └───────────────┘       └───────────────┘     └───────────────┘      └───────────────┘
      (fills height)          (top)                 (middle)               (bottom)
    ```

### Alignment Examples

```go
// Centered content (both axes)
Column{
    Width:      Flex(1),
    Height:     Flex(1),
    MainAlign:  MainAxisCenter,
    CrossAlign: CrossAxisCenter,
    Children: []Widget{
        Text{Content: "Centered"},
    },
}

// Bottom-right alignment
Column{
    Width:      Flex(1),
    Height:     Flex(1),
    MainAlign:  MainAxisEnd,
    CrossAlign: CrossAxisEnd,
    Children: []Widget{
        Text{Content: "Bottom Right"},
    },
}
```

## Dock

Edge-docking layout which allows you to stick widgets to an edge.
Useful for headers, sidebars, and footers.
Widgets dock to edges in order, consuming space, with the body filling the remainder.

=== "Code"

    ```go
    Dock{
        Top:    []Widget{Header{}},
        Bottom: []Widget{StatusBar{}},
        Left:   []Widget{Sidebar{}},
        Body:   MainContent{},
    }
    ```

=== "Diagram"

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

### DockOrder

Control edge processing order with `DockOrder`. This changes which edges get priority.

=== "Code"

    ```go
    // Process Left first - sidebar gets full height
    Dock{
        DockOrder: []Edge{Left, Right, Top, Bottom},
        Left:      []Widget{Sidebar{}},
        Top:       []Widget{Header{}},
        Body:      Content{},
    }
    ```

=== "Diagram"

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

1. Edges are processed in order (default: Top, Bottom, Left, Right)
2. Each edge's widgets consume space from that direction
3. The body fills whatever space remains

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

### Dock Fields

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

### Dock Example

```go
func (a *App) Build(ctx BuildContext) Widget {
    return Dock{
        Top: []Widget{
            Text{Content: "Header", Style: Style{
                BackgroundColor: ctx.Theme().Primary,
                Padding:         EdgeInsetsXY(2, 0),
            }},
        },
        Bottom: []Widget{
            KeybindBar{},
        },
        Left: []Widget{
            Column{
                Width: Cells(20),
                Children: []Widget{
                    Text{Content: "Sidebar"},
                },
            },
        },
        Body: Column{
            Children: []Widget{
                Text{Content: "Main content area"},
            },
        },
    }
}
```

## Spacer

A non-visual widget that occupies space. Defaults to `Flex(1)` in both dimensions.

=== "Code"

    ```go
    // Push items to opposite ends
    Row{
        Children: []Widget{
            Text{Content: "Left"},
            Spacer{},
            Text{Content: "Right"},
        },
    }
    ```

=== "Diagram"

    ```
    Without Spacer:                    With Spacer{}:
    ┌─────────────────────────┐        ┌─────────────────────────┐
    │ [Left] [Right]          │        │ [Left]          [Right] │
    └─────────────────────────┘        └─────────────────────────┘
                                              ↑ Spacer fills gap
    ```

See [Spacer](widgets/spacer.md) for detailed documentation.

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
        Margin: EdgeInsetsXY(4, 1),  // 4 horizontal, 1 vertical
    },
}
```

### EdgeInsets Helpers

| Helper | Description |
|--------|-------------|
| `EdgeInsetsAll(n)` | Same value for all 4 sides |
| `EdgeInsetsXY(h, v)` | Horizontal and vertical values |
| `EdgeInsetsTRBL(t, r, b, l)` | Top, right, bottom, left individually |

### Visual Comparison

```
Margin vs Padding vs Spacing:

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

### Comparison

| Method | Applied To | Use Case |
|--------|------------|----------|
| `Spacing` | Row/Column children | Uniform gaps between all children |
| `Padding` | Any widget | Space inside, around content |
| `Margin` | Any widget | Space outside, around widget |
| `Spacer{}` | Specific gap | Variable or precise gap between specific children |

## Common Layout Patterns

### Header, Body, Footer

=== "Code"

    ```go
    // Using Dock (preferred)
    Dock{
        Top:    []Widget{Header{}},
        Bottom: []Widget{Footer{}},
        Body:   Body{},
    }

    // Using Column
    Column{
        Height: Flex(1),
        Children: []Widget{
            Header{},
            Column{Height: Flex(1), Children: []Widget{Body{}}},
            Footer{},
        },
    }
    ```

=== "Diagram"

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

=== "Code"

    ```go
    // Using Row
    Row{
        Children: []Widget{
            Column{Width: Cells(25), Children: []Widget{Sidebar{}}},
            Column{Width: Flex(1), Children: []Widget{MainContent{}}},
        },
    }

    // Using Dock
    Dock{
        Left: []Widget{Column{Width: Cells(25), Children: []Widget{Sidebar{}}}},
        Body: MainContent{},
    }
    ```

=== "Diagram"

    ```
    ┌─────────┬───────────────────────┐
    │░░░░░░░░░│                       │
    │░░░░░░░░░│                       │
    │░Sidebar░│      Main Content     │
    │░░░░░░░░░│                       │
    │░░░░░░░░░│                       │
    └─────────┴───────────────────────┘
       Fixed            Flex(1)
    ```

### Centered Content

=== "Code"

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

=== "Diagram"

    ```
    ┌─────────────────────────┐
    │                         │
    │                         │
    │       [Centered]        │
    │                         │
    │                         │
    └─────────────────────────┘
    ```

### Push to Edges

=== "Code"

    ```go
    Row{
        Children: []Widget{
            Text{Content: "Left"},
            Spacer{},
            Text{Content: "Right"},
        },
    }
    ```

=== "Diagram"

    ```
    ┌─────────────────────────┐
    │ [Left]          [Right] │
    └─────────────────────────┘
         ↑ Spacer{} ↑
    ```

### Split Panels (Proportional)

=== "Code"

    ```go
    Row{
        Children: []Widget{
            Panel{Width: Flex(1)},  // 1/3
            Panel{Width: Flex(2)},  // 2/3
        },
    }
    ```

=== "Diagram"

    ```
    ┌────────────┬────────────────────────┐
    │            │                        │
    │  Flex(1)   │        Flex(2)         │
    │    1/3     │          2/3           │
    │            │                        │
    └────────────┴────────────────────────┘
    ```

### Fixed + Flexible

=== "Code"

    ```go
    Row{
        Children: []Widget{
            Sidebar{Width: Cells(30)},   // Fixed 30 cells
            Content{Width: Flex(1)},     // Takes remaining space
        },
    }
    ```

=== "Diagram"

    ```
    ┌──────────────┬──────────────────────────────┐
    │              │                              │
    │   Cells(30)  │          Flex(1)             │
    │    Fixed     │    Takes remaining space     │
    │              │                              │
    └──────────────┴──────────────────────────────┘
    ```

## Layout Decision Guide

**Choose Row/Column when:**

- Arranging widgets in a single direction
- Need alignment control (MainAlign, CrossAlign)
- Simple linear layouts

**Choose Dock when:**

- Building app shells with header/footer/sidebar
- Edges should consume only needed space
- Body should fill remaining area

**Choose Spacer when:**

- Pushing widgets to opposite ends
- Creating flexible gaps between specific items
- Need proportional spacing with Flex values

**Use Spacing field when:**

- Uniform gaps between all Row/Column children
- Simpler than adding Spacers between every child

**Use Flex dimensions when:**

- Content should fill available space
- Proportional sizing between siblings

**Use Cells dimensions when:**

- Precise pixel-perfect (cell-perfect) control needed
- Fixed-width sidebars or panels

**Use Auto dimensions when:**

- Widget should be exactly as large as its content
- Don't want extra space around content
