# FilePicker

## Specification

`FilePicker` is an embeddable path-selection widget. It reads directory entries and file metadata; it never creates, writes, overwrites, or deletes a selected file. Compose it with `Dialog` for an outer modal. A stable, unique `ID` namespaces its input/list/button/overwrite-dialog IDs.

```go
state := NewFilePickerState(startDirectory)
FilePicker{
    ID: "open-file",
    State: state,
    Mode: FilePickerOpen, // FilePickerSave or FilePickerDirectory
    Filters: []FileFilter{{Label: "Go source", Patterns: []string{"*.go"}}},
    MultiSelect: true, // Open only
    OnSelect: func(paths []string) { /* caller opens or saves */ },
    OnCancel: func() {},
    Style: Style{Width: Flex(1), Height: Flex(1)},
}
```

Use `Navigate` rather than assigning `Directory` directly, so entries and selection move together. State exposes `Directory`, `Error`, `ShowHidden`, `FilterIndex`, `PathInput`, and `FilenameInput`. Use `Navigate(path)` and `Refresh()` to perform filesystem reads and update state outside `Build`. `NewFilePickerStateWithFileSystem` accepts a `FilePickerFileSystem` (`ReadDir`, `Stat`, `Lstat`) for deterministic virtual-filesystem and permission-error tests. The OS adapter is the default.

`FilePicker.Submit()` and `Cancel()` cross the same seam as UI buttons; callbacks occur at most once per selection/cancellation attempt. Completing or cancelling latches the state until `State.Reset()` or successful navigation reopens it. Build is pure with respect to signals and filesystem access.

### Navigation, paths, and errors

- Empty starting path means the working directory. Relative paths resolve against the current directory, or process working directory during initialization. Absolute lexical paths are returned; `~` and environment variables are not expanded. Spaces and Unicode are preserved exactly. Whitespace is a legal filename and is never trimmed.
- A starting/path-field file navigates to its parent and seeds the filename and row cursor. Missing or unreadable starting paths leave a visible error and empty list. Failed later navigation preserves the previous directory/list/selection and exposes the error. Parent of filesystem root is root. Refresh clears selections to avoid stale index transfer.
- Listing is one directory deep, directories first then case-sensitive lexical filename order. There is no recursive preload or cycle traversal. Symlink targets are inspected; valid directory symlinks can be navigated explicitly, regular file symlinks selected, and broken links/special files are unavailable. Paths remain lexical (symlink aliases are not deduplicated). Final metadata is revalidated on submit; callers must still handle races and permissions when opening/writing.
- Reads occur synchronously in setup/handlers, never Build. Work is bounded to one directory listing plus metadata for symlinks. Very large or slow remote directories can block the event loop; asynchronous loading is outside this first implementation.

### Visibility and selection

- Hidden means basename starts with `.`. The toggle defaults off. Filters use `filepath.Match` on basenames, with case sensitivity defined by that function. Empty filter lists/pattern lists match all files. Invalid patterns show an error and block submission. No extension is appended automatically.
- Directories are always visible and navigable regardless of file filter; directory mode hides files. Hidden toggle still applies to directories. Filters and hidden controls clear the existing selection and anchor; direct signal/config changes are defensively rechecked during submission, so hidden/stale items cannot be submitted.
- Open mode accepts existing regular files only. Enter/double-click on a directory navigates; Enter/double-click on a file submits that file. The Open button accepts multi-selected visible files if present, otherwise the visible cursor file. Multi-selection belongs to one directory and emits paths in current display order, once, without duplicates. Selecting a directory in a range never implicitly selects its contents.
- Directory mode's Choose button selects the current directory, including an empty one; Enter on directory rows navigates into them. MultiSelect has no effect.

### Save and overwrite

- Save mode shows a filename field. It accepts one nonempty basename, preserving spaces/Unicode, rejecting `.`, `..`, separators and NUL. Navigate using the path field instead of placing parent components in the filename. The active file filter applies to the candidate; mismatches are errors, not silently rewritten names.
- Missing candidate is returned to the caller immediately. Existing regular file (including a valid symlink to one) requires an explicit overwrite confirmation. An existing directory, broken symlink or special file is rejected.
- Overwrite confirmation is a nested modal with Cancel first. Escape/dismiss cancels only confirmation and returns focus to filename. Confirm rechecks the exact candidate and rejects changed filename/directory/mode/target type. For OS metadata it also compares file identity, so a renamed replacement or retargeted symlink cannot reuse approval even when size, mode and modification time match. Virtual filesystems without OS identity use the metadata comparison. Same-file changes that preserve checked metadata and changes after confirmation remain the caller’s responsibility. An existing target that vanishes before confirmation may still be accepted as a new file. The widget itself never overwrites anything.
- While confirmation is open, Cancel or Escape dismisses only confirmation. Otherwise cancelling the picker calls OnCancel once and does not call OnSelect. Confirming an already completed attempt does nothing. The caller owns hiding or replacing the widget after completion.

### Keyboard, mouse, focus, layout

Tab traverses controls; arrows/Home/End/Page keys navigate the list. In Open multi-select mode, Shift+movement, Shift+click or dragging selects a range; Space toggles the cursor file within that selection. Plain movement or clicking clears the previous selection, matching List behavior. To select two adjacent files, position on the first and press Shift+Down, then activate Open. Enter or double-click activates a row. Buttons provide mouse access to parent, refresh, hidden toggle, filter cycling, choose/open/save and cancel. Ctrl+L focuses path, Alt+Up goes to parent, Ctrl+F cycles filters, Ctrl+Enter submits, Escape cancels. Text inputs retain their normal editing shortcuts.

The list uses a bounded vertical Scrollable; outer layout clips safely on narrow/short terminals. Supply a useful bounded height (recommended at least 14 rows). Long paths scroll within TextInput; errors wrap. At extremely small sizes some controls may be clipped, but keyboard shortcuts and resizing remain available. Nested modal focus is delegated to existing Dialog machinery.

## Expected edge behavior and test plan

| Condition | Expected | Planned proof |
|---|---|---|
| Nonexistent/unreadable start | Error with no selected result | State/unit + error SVG |
| Failed later navigation | Existing contents/selection retained | State regression |
| Relative/space/Unicode paths | Correct absolute paths and exact names | OS fixture tests |
| Broken/directory symlinks | Broken rejected; directory explicit navigation only | OS + fake FS tests |
| Empty directory | Empty message; Directory mode still chooses it | SVG + callback test |
| Hidden/filter change | No invisible file can be submitted | Unit + browser |
| Invalid filter patterns | Clear error, no callback | Unit + SVG |
| Multi-select across directories | Clear at navigation; visible regular files only, sorted order | Unit + browser |
| Save new/existing/directory | Immediate/new, explicit confirm/existing, error/directory | Unit + browser + SVG |
| Filename/filter/target changes during confirm | Revalidate; no stale overwrite decision | Unit tests |
| Repeated confirm/cancel | At most one final callback until reset | Cardinality tests |
| Build/repaint | No filesystem reads or signal writes | Build-count test + retained-render sequence |
| Tiny layout / modal nesting | Safe clipping and focus containment | SVGs + browser |

The demo accepts `-probe` to start Save mode inside a parent Dialog. In that
disposable-fixture mode only, F7 simulates another writer replacing `alpha.go`
with identical size, permissions and modification time. Confirming the stale
prompt reports that the target changed; resubmit to review the replacement.
