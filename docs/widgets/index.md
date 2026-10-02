# Widgets

A Terma widget implements `Build(BuildContext) Widget`.

## Run an example

The examples included below use `t` as the import alias for `github.com/darrenburns/terma`.
Each example function returns a widget that can be passed to `t.Run`.
Create state once, outside `Build()`, so rebuilds preserve edits, selection, and scroll position.

This complete program displays a text widget.

```go
--8<-- "docs/minimal-examples/widget-start/main.go"
```

From the repository root, run a documented example with `-widget`.

```sh
go run ./docs/widget-examples -widget button
go run ./docs/widget-examples -list
```

`-list` prints all available example names.
The [getting started guide](../getting-started.md) walks through a complete application.

## Text and status

- [Text](text.md)
- [PresentedText](presentedtext.md)
- [Image](image.md)
- [Markdown](markdown.md)
- [ProgressBar](progressbar.md)
- [Sparkline](sparkline.md)
- [Spinner](spinner.md)

## Input and actions

- [Button](button.md)
- [Checkbox](checkbox.md)
- [TextInput](textinput.md)
- [TextArea](textarea.md)
- [Autocomplete](autocomplete.md)
- [SelectBox](select.md)
- [Form and Field](form.md)
- [FilePicker](filepicker.md)
- [CommandPalette](commandpalette.md)
- [Menu](menu.md)
- [Dialog](dialog.md)

## Collections and navigation

- [List](list.md)
- [Table](table.md)
- [Tree](tree.md)
- [DirectoryTree](directorytree.md)
- [TabBar and TabView](tabs.md)
- [Breadcrumbs](breadcrumbs.md)
- [KeybindBar](keybindbar.md)
- [Jumper](jumper.md)

## Composition

- [Switcher](switcher.md)
- [FocusTrap](focustrap.md)
- [Tooltip](tooltip.md)
- [EmptyWidget](empty.md)

## Layout

[Row and Column](../layout/row-column.md), [Dock](../layout/dock.md), [Scrollable](../layout/scrollable.md), [SplitPane](../layout/splitpane.md), and [Stack](../layout/stack.md) arrange child widgets.
[Spacer](../layout/spacer.md) reserves empty layout space.
[Floating](../floating.md) places content above the main layout.

## Custom widgets

[Custom widgets](custom-widgets.md) compose existing widgets or implement their own rendering.
