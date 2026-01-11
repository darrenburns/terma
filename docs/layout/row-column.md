# Row & Column

The primary layout widgets for arranging children linearly.

## Row

Arranges children horizontally (left to right).

```go
Row{
    Children: []Widget{
        Text{Content: "Left"},
        Text{Content: "Center"},
        Text{Content: "Right"},
    },
}
```

```
Row (main axis →)
┌─────────────────────────────────────┐
│  [A]    [B]    [C]        →         │
└─────────────────────────────────────┘
```

## Column

Arranges children vertically (top to bottom).

```go
Column{
    Children: []Widget{
        Text{Content: "Top"},
        Text{Content: "Middle"},
        Text{Content: "Bottom"},
    },
}
```

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

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | `""` | Optional identifier |
| `Children` | `[]Widget` | `nil` | Widgets to arrange |
| `Spacing` | `int` | `0` | Gap between children in cells |
| `MainAlign` | `MainAxisAlign` | `MainAxisStart` | Alignment along main axis |
| `CrossAlign` | `CrossAxisAlign` | `CrossAxisStretch` | Alignment along cross axis |
| `Width` | `Dimension` | — | Width dimension |
| `Height` | `Dimension` | — | Height dimension |
| `Style` | `Style` | `Style{}` | Padding, margin, border, colors |

## Main and Cross Axis

Linear layouts have two axes:

- **Main axis**: The direction children are arranged
- **Cross axis**: Perpendicular to the main axis

| Widget | Main Axis | Cross Axis |
|--------|-----------|------------|
| Row | Horizontal (left-right) | Vertical (top-bottom) |
| Column | Vertical (top-bottom) | Horizontal (left-right) |

## Main Axis Alignment

Controls how children are distributed along the main axis when there's extra space.

| Value | Description |
|-------|-------------|
| `MainAxisStart` | Pack at the start (default) |
| `MainAxisCenter` | Center children |
| `MainAxisEnd` | Pack at the end |

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

```
MainAxisStart (default)          MainAxisCenter              MainAxisEnd
┌─────────────────────┐     ┌─────────────────────┐     ┌─────────────────────┐
│ [A] [B] [C]         │     │      [A] [B] [C]    │     │         [A] [B] [C] │
└─────────────────────┘     └─────────────────────┘     └─────────────────────┘
```

## Cross Axis Alignment

Controls how children are positioned along the cross axis.

| Value | Description |
|-------|-------------|
| `CrossAxisStretch` | Stretch to fill cross axis (default) |
| `CrossAxisStart` | Align at start of cross axis |
| `CrossAxisCenter` | Center along cross axis |
| `CrossAxisEnd` | Align at end of cross axis |

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

Cross axis alignment in a Row (cross axis is vertical):

```
CrossAxisStretch        CrossAxisStart        CrossAxisCenter         CrossAxisEnd
┌───────────────┐       ┌───────────────┐     ┌───────────────┐      ┌───────────────┐
│███████████████│       │ [A]  [B]  [C] │     │               │      │               │
│███████████████│       │               │     │ [A]  [B]  [C] │      │               │
│███████████████│       │               │     │               │      │ [A]  [B]  [C] │
└───────────────┘       └───────────────┘     └───────────────┘      └───────────────┘
  (fills height)          (top)                 (middle)               (bottom)
```

## Examples

### Centered Content

```go
Column{
    Width:      Flex(1),
    Height:     Flex(1),
    MainAlign:  MainAxisCenter,
    CrossAlign: CrossAxisCenter,
    Children: []Widget{
        Text{Content: "Centered"},
    },
}
```

### Bottom-Right Alignment

```go
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

```
┌─────────────────────────┐
│ [Left]          [Right] │
└─────────────────────────┘
```

### Proportional Split

```go
Row{
    Children: []Widget{
        Panel{Width: Flex(1)},  // 1/3
        Panel{Width: Flex(2)},  // 2/3
    },
}
```

```
┌────────────┬────────────────────────┐
│            │                        │
│  Flex(1)   │        Flex(2)         │
│    1/3     │          2/3           │
└────────────┴────────────────────────┘
```

### Fixed + Flexible

```go
Row{
    Children: []Widget{
        Sidebar{Width: Cells(30)},   // Fixed 30 cells
        Content{Width: Flex(1)},     // Takes remaining space
    },
}
```

### Spacing Between Children

```go
Column{
    Spacing: 1,
    Children: []Widget{
        Text{Content: "Line 1"},
        Text{Content: "Line 2"},
        Text{Content: "Line 3"},
    },
}
```

### Nested Layouts

```go
Column{
    Children: []Widget{
        // Header row
        Row{
            Children: []Widget{
                Text{Content: "Title"},
                Spacer{},
                Button{ID: "menu", Label: "Menu"},
            },
        },
        // Content
        Text{Content: "Body content here"},
    },
}
```
