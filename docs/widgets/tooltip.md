# Tooltip

Displays contextual help text when a child widget has focus.
Tooltips appear as floating overlays positioned relative to their child widget.

## Overview

`Tooltip` wraps a child widget and shows a floating text overlay when that child receives focus. This is useful for providing additional context, keyboard shortcuts, or help text without cluttering the interface.

```go
Tooltip{
    Content: "Submit the form",
    Child:   &Button{ID: "submit", Label: "Submit"},
}
```

When the button is focused (via Tab navigation), the tooltip appears above it.

## Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ID` | `string` | auto | Optional identifier for the tooltip |
| `Content` | `string` | `""` | Plain text to display |
| `Spans` | `[]Span` | `nil` | Rich text (takes precedence over Content) |
| `Child` | `Widget` | — | The widget that triggers the tooltip |
| `Position` | `TooltipPosition` | `TooltipTop` | Where tooltip appears relative to child |
| `Offset` | `int` | `0` | Gap in cells between child and tooltip |
| `Style` | `Style` | theme defaults | Custom tooltip styling |

## Positions

Control where the tooltip appears relative to the child widget:

```go
// Above the child (default)
Tooltip{
    Content:  "Appears above",
    Position: TooltipTop,
    Child:    &Button{ID: "btn", Label: "Top"},
}

// Below the child
Tooltip{
    Content:  "Appears below",
    Position: TooltipBottom,
    Child:    &Button{ID: "btn", Label: "Bottom"},
}

// To the left of the child
Tooltip{
    Content:  "Appears left",
    Position: TooltipLeft,
    Child:    &Button{ID: "btn", Label: "Left"},
}

// To the right of the child
Tooltip{
    Content:  "Appears right",
    Position: TooltipRight,
    Child:    &Button{ID: "btn", Label: "Right"},
}
```

## Rich Text

Use `Spans` for styled tooltip content with bold, italic, or colored text:

```go
Tooltip{
    Spans: []Span{
        BoldSpan("Ctrl+S"),
        PlainSpan(" to save, "),
        BoldSpan("Ctrl+Q"),
        PlainSpan(" to quit"),
    },
    Child: &Button{ID: "help", Label: "Shortcuts"},
}
```

Or use markup parsing:

```go
func (a *App) Build(ctx BuildContext) Widget {
    return Tooltip{
        Spans: ParseMarkup("[b]Ctrl+S[/] to save", ctx.Theme()),
        Child: &Button{ID: "save", Label: "Save"},
    }
}
```

## Custom Styling

Override the default tooltip appearance with the `Style` field:

```go
// Warning-styled tooltip
Tooltip{
    Content: "This action cannot be undone",
    Style: Style{
        BackgroundColor: theme.Warning,
        ForegroundColor: RGB(0, 0, 0),
        Padding:         EdgeInsetsXY(2, 0),
    },
    Child: &Button{ID: "delete", Label: "Delete"},
}

// Error-styled tooltip
Tooltip{
    Content: "Invalid input",
    Style: Style{
        BackgroundColor: theme.Error,
        ForegroundColor: RGB(255, 255, 255),
        Border:          Border{Style: BorderRounded, Color: theme.Error},
    },
    Child: &TextInput{ID: "email", State: emailState},
}
```

## Offset

Add space between the child and tooltip with the `Offset` field:

```go
// 2-cell gap between button and tooltip
Tooltip{
    Content: "With spacing",
    Offset:  2,
    Child:   &Button{ID: "btn", Label: "Spaced"},
}
```

## Form Field Help

Tooltips work well with input fields to show validation rules or hints:

```go
func (a *App) Build(ctx BuildContext) Widget {
    return Column{
        Spacing: 1,
        Children: []Widget{
            Tooltip{
                Content:  "3-20 characters, letters and numbers only",
                Position: TooltipRight,
                Child:    &TextInput{ID: "username", State: a.usernameState, Placeholder: "Username"},
            },
            Tooltip{
                Content:  "Must be at least 8 characters",
                Position: TooltipRight,
                Child:    &TextInput{ID: "password", State: a.passwordState, Placeholder: "Password"},
            },
        },
    }
}
```

## Notes

- Tooltips only appear when the child widget has focus (keyboard navigation)
- The child must be a focusable widget (Button, TextInput, etc.) for the tooltip to show
- Default styling uses `theme.Surface` background and `theme.Text` foreground
- Tooltips render as floating overlays and won't affect layout of other widgets
- If no `ID` is provided, one is auto-generated
