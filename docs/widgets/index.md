# Widgets

Widgets are the building blocks of Terma applications. Everything in Terma is a widget, from simple text to complex interactive lists.

## Widget Categories

### Layout Widgets

Widgets that arrange and position other widgets:

- [Row & Column](../layout/row-column.md) - Arrange children linearly
- [Dock](../layout/dock.md) - Edge-docking layout
- [Scrollable](../layout/scrollable.md) - Scrolling container
- [Floating](../floating.md) - Overlays, modals, and dropdowns
- [Spacer](../layout/spacer.md) - Empty space for layout control

### Content Widgets

Widgets that display content:

- Text - Display plain or rich text
- [TextInput](textinput.md) - Single-line text entry
- Button - Focusable button with press handler
- List - Generic navigable list
- Table - Navigable multi-column table
- [Tree](tree.md) - Hierarchical expandable list
- [ProgressBar](progressbar.md) - Horizontal progress indicator
- [Tabs](tabs.md) - TabBar and TabView for tab navigation

### Conditional & Switching Widgets

- [Switcher](switcher.md) - Show one widget at a time from a keyed collection
- [ShowWhen / HideWhen](../conditional.md#showwhen--hidewhen) - Toggle widget presence
- [VisibleWhen / InvisibleWhen](../conditional.md#visiblewhen--invisiblewhen) - Toggle visibility while preserving space

### Utility Widgets

- KeybindBar - Display active keybindings
- [Spacer](spacer.md) - Empty space for layout control
- [Spinner](../animation.md#spinner) - Animated loading indicators
- [Tooltip](tooltip.md) - Contextual help text on focus

## Creating Custom Widgets

Every widget implements the `Widget` interface:

```go
type Widget interface {
    Build(ctx BuildContext) Widget
}
```

Leaf widgets (those that render directly) return themselves from `Build()`. Composite widgets return a tree of other widgets.

```go
// Leaf widget example
type MyLeafWidget struct{}

func (w MyLeafWidget) Build(ctx BuildContext) Widget {
    return w  // Returns itself
}

// Composite widget example
type MyCompositeWidget struct {
    Title string
}

func (w MyCompositeWidget) Build(ctx BuildContext) Widget {
    return Column{
        Children: []Widget{
            Text{Content: w.Title},
            // ... more widgets
        },
    }
}
```
