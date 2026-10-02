# SelectBox

`SelectBox[T comparable]` chooses a committed value from a fixed set of labelled
options. Use stable scalar IDs for records containing slices or maps.

## Quick start

```go
state := terma.NewSelectState[string]() // allocate once, outside Build
widget := terma.SelectBox[string]{
	ID:    "environment",
	State: state,
	Options: []terma.SelectOption[string]{
		{Label: "Development", Value: "dev"},
		{Label: "Production", Value: "prod", Disabled: true},
	},
	Placeholder: "Choose environment",
	Searchable:  true,
	MaxVisible:  8,
	Style:       terma.Style{Width: terma.Cells(30)},
	OnChange:    func(value string) { /* a different value was committed */ },
}
value, set := state.Value() // reactive; zero and unset are different
state.SetValue("dev")       // programmatic updates do not call OnChange
state.Clear()               // return to unset; closes any open popup
```

`Style` styles the control; `PopupStyle` styles the popup. `Disabled` prevents
interaction and focus, like wrapping the widget in `DisabledWhen`. A nil State
renders an inert placeholder. IDs are optional but stable explicit IDs are
recommended. `MaxVisible <= 0` defaults to eight option rows.

## Interaction and selection

* Enter, Space, Up or Down opens the popup without changing the value. It initially
  highlights the first enabled option matching the committed value, otherwise the
  first enabled option. Up/Down move through enabled matches and stop at the ends;
  Home/End jump to the first/last enabled match. Page Up/Down move a page.
* Enter commits the highlight and closes. Clicking an enabled row commits directly.
  The callback runs once, after state is updated and the popup has closed, only
  if the committed value changed (including unset to the zero value).
* Escape, Tab, Shift+Tab, blur, and outside click cancel. Tab retains normal focus
  traversal. Focus remains on the control while searching and choosing. Clicking
  the control toggles its popup. An outside click dismisses the popup without
  activating the content underneath, including when used inside a dialog.
* Searchable controls accept printable text, including spaces, directly. Typing opens the popup, except that Space on a closed control opens
  it without adding a search character. Search is Unicode case-insensitive substring matching on labels.
  Backspace removes one Unicode code point and Ctrl+U clears the search. Search is
  transient and resets every time the popup closes. Search does not normalize Unicode or use fuzzy matching. Nonsearchable controls leave printable keys unhandled.
* Disabled options remain visible, are dimmed and marked, and cannot be highlighted
  or committed. Empty lists say "No options"; searches with no results say "No
  matches"; all-disabled results are visible with no highlight. Enter does nothing
  in these states. Escape and Tab still work.
* Values use Go equality and should be stable, reflexive comparable keys (not NaN,
  and not interface values containing maps/slices). Duplicate labels are supported.
  Duplicate values refer to the same selection; the first matching option supplies
  the closed label. Recommitting a duplicate does not call OnChange.
* Removing a selected option does not silently change application state: the
  control displays "Unavailable selection" until the option returns or the caller
  changes/clears it. Disabled committed options retain their label. If the active
  option disappears or becomes disabled while open, the first enabled match is
  used. Reordering options preserves the highlight by value.
* Popup rows are bounded by MaxVisible and the viewport; keyboard navigation and
  the mouse wheel reveal the active row. A range footer reports the visible slice.
  Popup width is at least the control's width and fits the longest label when
  possible; long lines are clipped without wrapping. Floating placement follows
  Terma's standard screen clamping, which can overlap an anchor near the bottom.
  Very small terminals prioritize option rows over search/status decorations.
* Change state in event handlers or setup code, outside `Build`. Changes to
  `Options` take effect on the next rebuild, as with other widget properties.

## Demo

Run `go run ./cmd/select-demo` to try searching, disabled options and selection
callbacks.
