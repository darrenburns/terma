# Spacer

A non-visual widget that occupies space in layouts. Use it to push widgets apart, create flexible gaps, or add fixed-size empty regions.

## Basic Usage

A bare `Spacer{}` expands to fill available space (defaults to `Flex(1)` in both dimensions):

```go
// Push items to opposite ends of a row
Row{
    Children: []Widget{
        Text{Content: "Left"},
        Spacer{},
        Text{Content: "Right"},
    },
}
```

## Fixed-Size Gaps

Use `Cells()` for fixed-size spacing:

```go
// 5-cell gap between items
Row{
    Children: []Widget{
        Text{Content: "A"},
        Spacer{Width: Cells(5)},
        Text{Content: "B"},
    },
}
```

## Proportional Spacing

Use different `Flex()` values for proportional distribution:

```go
// B is closer to C (1:2 ratio of gaps)
Row{
    Children: []Widget{
        Text{Content: "A"},
        Spacer{Width: Flex(1)},  // 1/3 of remaining space
        Text{Content: "B"},
        Spacer{Width: Flex(2)},  // 2/3 of remaining space
        Text{Content: "C"},
    },
}
```

## Vertical Spacing

Spacers work in both directions. In a Column, the Height dimension matters:

```go
Column{
    Height: Cells(10),
    Children: []Widget{
        Text{Content: "Top"},
        Spacer{},  // Pushes Bottom to the end
        Text{Content: "Bottom"},
    },
}
```

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Width` | `Dimension` | `Flex(1)` | Horizontal size |
| `Height` | `Dimension` | `Flex(1)` | Vertical size |

## Dimension Options

- **`Flex(n)`** - Take a proportional share of remaining space (default)
- **`Cells(n)`** - Fixed size in terminal cells
- **`Auto`** - Fit content (results in 0 size for Spacer since it has no content)

!!! warning "Auto results in zero size"
    Since Spacer has no content, setting `Auto` results in 0 size. `Flex(1)` (the default) causes the spacer to expand.

## Example Application

```go
package main

import (
    "log"
    t "terma"
)

type App struct{}

func (a *App) Build(ctx t.BuildContext) t.Widget {
    theme := ctx.Theme()

    return t.Column{
        Style: t.Style{Padding: t.EdgeInsetsAll(1)},
        Children: []t.Widget{
            // Header pushed to top, footer to bottom
            t.Text{Content: "Header", Style: t.Style{
                ForegroundColor: theme.Primary.AutoText(),
                BackgroundColor: theme.Primary,
            }},
            t.Spacer{},
            t.Text{Content: "Footer", Style: t.Style{
                ForegroundColor: theme.Surface.AutoText(),
                BackgroundColor: theme.Surface,
            }},
        },
    }
}

func main() {
    if err := t.Run(&App{}); err != nil {
        log.Fatal(err)
    }
}
```

## Spacer vs Other Spacing Methods

| Method | Use Case |
|--------|----------|
| `Spacer{}` | Flexible space between specific children |
| `Spacing` field on Row/Column | Uniform gaps between all children |
| `Margin` on widgets | Space around a specific widget |
| `Padding` on containers | Space inside a container, around content |

## When to Use Spacer

**Use Spacer when you need:**

- To push widgets to opposite ends of a container
- Variable-sized gaps that respond to available space
- Precise control over spacing between specific children

**Use `Spacing` field instead when:**

- You want uniform gaps between all children in a Row or Column

```go
// Uniform 2-cell gaps between all children
Column{
    Spacing: 2,
    Children: []Widget{
        Text{Content: "A"},
        Text{Content: "B"},
        Text{Content: "C"},
    },
}
```
