# FilePicker

`FilePicker` is an embeddable path-selection widget. It reads directory entries and file metadata; it never creates, writes, overwrites, or deletes a selected file. Compose it with `Dialog` for an outer modal. A stable, unique `ID` namespaces its input/list/button/overwrite-dialog IDs.

## Quick start

Create the state once during setup. Add `picker` to your widget tree and handle
`OnSelect` and `OnCancel` in your application.

```go
state := terma.NewFilePickerState(startDirectory)
picker := terma.FilePicker{
	ID:          "open-file",
	State:       state,
	Mode:        terma.FilePickerOpen, // Also FilePickerSave or FilePickerDirectory
	Filters:     []terma.FileFilter{{Label: "Go source", Patterns: []string{"*.go"}}},
	MultiSelect: true, // Open only
	OnSelect:    func(paths []string) { /* caller opens or saves */ },
	OnCancel:    func() {},
	Style:       terma.Style{Width: terma.Flex(1), Height: terma.Cells(20)},
}
```

Use `Navigate` rather than assigning `Directory` directly, so entries and selection move together. State exposes `Directory`, `Error`, `ShowHidden`, `FilterIndex`, `PathInput`, and `FilenameInput`. Use `Navigate(path)` and `Refresh()` to perform filesystem reads and update state outside `Build`. `NewFilePickerStateWithFileSystem` accepts a `FilePickerFileSystem` (`ReadDir`, `Stat`, `Lstat`) for custom filesystem access. By default, the picker uses the operating system filesystem.

`FilePicker.Submit()` and `Cancel()` perform the same actions as the picker buttons; callbacks occur at most once per selection/cancellation attempt. Completing or cancelling latches the state until `State.Reset()` or successful navigation reopens it. Create state and call navigation methods outside `Build`.

### Navigation, paths, and errors

- Empty starting path means the working directory. Relative paths resolve against the current directory, or process working directory during initialization. Absolute lexical paths are returned; `~` and environment variables are not expanded. Spaces and Unicode are preserved exactly. Whitespace is a legal filename and is never trimmed.
- A starting/path-field file navigates to its parent and seeds the filename and row cursor. Missing or unreadable starting paths leave a visible error and empty list. Failed later navigation preserves the previous directory/list/selection and exposes the error. Parent of filesystem root is root. Refresh clears selections to avoid stale index transfer.
- Listing is one directory deep, directories first then case-sensitive lexical filename order. There is no recursive preload or cycle traversal. Symlink targets are inspected; valid directory symlinks can be navigated explicitly, regular file symlinks selected, and broken links/special files are unavailable. Paths remain lexical (symlink aliases are not deduplicated). Final metadata is revalidated on submit; callers must still handle races and permissions when opening/writing.
- Reads occur synchronously in setup/handlers, never Build. Work is bounded to one directory listing plus metadata for symlinks. Very large or slow remote directories can block the event loop; asynchronous loading is not supported.

### Visibility and selection

- Hidden means basename starts with `.`. The toggle defaults off. Filters use `filepath.Match` on basenames, with case sensitivity defined by that function. Empty filter lists/pattern lists match all files. Invalid patterns show an error and block submission. No extension is appended automatically.
- Directories are always visible and navigable regardless of file filter; directory mode hides files. Hidden toggle still applies to directories. Filters and hidden controls clear the existing selection and anchor; direct signal/config changes are defensively rechecked during submission, so hidden/stale items cannot be submitted.
- Open mode accepts existing regular files only. Enter/double-click on a directory navigates; Enter/double-click on a file submits that file. The Open button accepts multi-selected visible files if present, otherwise the visible cursor file. Multi-selection belongs to one directory and emits paths in current display order, once, without duplicates. Selecting a directory in a range never implicitly selects its contents.
- Directory mode's Choose button selects the current directory, including an empty one; Enter on directory rows navigates into them. MultiSelect has no effect.

### Save and overwrite

- Save mode shows a filename field. It accepts one nonempty basename, preserving spaces/Unicode, rejecting `.`, `..`, separators and NUL. Navigate using the path field instead of placing parent components in the filename. The active file filter applies to the candidate; mismatches are errors, not silently rewritten names.
- Missing candidate is returned to the caller immediately. Existing regular file (including a valid symlink to one) requires an explicit overwrite confirmation. An existing directory, broken symlink or special file is rejected.
- Overwrite confirmation is a nested modal with **Keep existing** first and **Overwrite** second. **Keep existing** or Escape cancels confirmation and returns focus to the filename field.
- **Overwrite** rechecks the exact candidate. A changed filename, directory, mode, or target type rejects confirmation. If the target vanished, the picker can accept the path as a new file after checking the parent directory.
- With OS metadata, confirmation compares file identity, size, mode, and modification time. A replacement file or retargeted symlink requires fresh confirmation. Virtual filesystems without OS identity use the metadata comparison. Callers must still handle changes after confirmation or same-file changes that preserve the checked metadata.
- While confirmation is open, **Keep existing**, `Cancel()`, or Escape dismisses only confirmation. Otherwise cancelling the picker calls OnCancel once and does not call OnSelect. Confirming an already completed attempt does nothing. The caller owns hiding or replacing the widget after completion.

### Keyboard, mouse, focus, layout

Tab traverses controls; arrows/Home/End/Page keys navigate the list. In Open multi-select mode, Shift+movement, Shift+click or dragging selects a range; Space toggles the cursor file within that selection. Plain movement or clicking clears the previous selection, matching List behavior. To select two adjacent files, position on the first and press Shift+Down, then activate Open. Enter or double-click activates a row. Buttons provide mouse access to parent, refresh, hidden toggle, filter cycling, choose/open/save and cancel. Ctrl+L focuses path, Alt+Up goes to parent, Ctrl+F cycles filters, Ctrl+Enter submits, Escape cancels. Text inputs retain their normal editing shortcuts.

The list uses a bounded vertical Scrollable; outer layout clips safely on narrow/short terminals. Supply a useful bounded height (recommended at least 14 rows). Long paths scroll within TextInput; errors wrap. At extremely small sizes some controls may be clipped, but keyboard shortcuts and resizing remain available.

## Demo

Run `go run ./cmd/filepicker-demo` to try Open, Save and Directory modes using
temporary example files.
