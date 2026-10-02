# Input and action widget verification

The source audit covers Button, Checkbox, TextInput, TextArea, Autocomplete, CommandPalette, Menu, and Dialog at commit 5728b68d01669f8e99cc9621b6d79463d8e16f93.
The accompanying JSON records the implementation evidence for 133 prose and image-alt claims.
Each image is registered against the same Go function included in its page.
The parent task owns image generation, visual inspection, browser verification, and the independent sentence audit.

## Corrections to previous documentation

- TextInput binds Ctrl+A to SelectAll, rather than moving to the beginning.
- TextInputState has no Clear method, so the page describes SetText instead.
- TextInput renders its placeholder whenever content is empty, including while focused.
- Dialog comments say a backdrop click dismisses it, but Dialog.Build passes Modal true and no DismissOnClickOutside override.
  FloatConfig.shouldDismissOnClickOutside defaults to false for modal floats.
  The new page documents the implemented behavior.
- CommandPalette FilterText replaces the searchable label.
  A failed FilterText match returns before checking Label, so hidden keywords must include any label words that should remain searchable.

## Checks run

- PASS. `go test ./docs/widget-examples/inputs` compiled the examples after parent integration.
- PASS. `go test . -run 'Test(Button|Checkbox|NewCheckboxState|TextInput|NewTextInputState|TextArea|Autocomplete|Insert|Menu|CommandPalette|Snapshot_(Button|Checkbox|TextInput|TextArea|Autocomplete|Menu|CommandPalette|Dialog))' -count=1` exercised existing behavior and snapshots.
- PASS. `gofmt` formatted the example source.

## Example details

State is allocated by each constructor, outside Build methods.
Menu and Dialog handlers change visibility signals that their Build methods only read.
The Menu constructor requests focus for its initial open menu.
The Autocomplete example sets DismissOnBlur false to keep its suggestions visible before focus arrives.
Button and CommandPalette actions update visible status text.
