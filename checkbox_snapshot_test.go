package terma

import "testing"

func TestSnapshot_Checkbox_Unchecked_Unfocused(t *testing.T) {
	state := NewCheckboxState(false)
	// Put a Button first so it takes focus, leaving the checkbox unfocused
	widget := Column{
		Children: []Widget{
			&Button{ID: "focus-stealer", Label: ""},
			&Checkbox{
				ID:    "checkbox-unfocused",
				State: state,
				Label: "Accept terms",
			},
		},
	}
	AssertSnapshot(t, widget, 25, 2, "Unchecked checkbox with label in default Text color (unfocused)")
}

func TestSnapshot_Checkbox_Checked_Unfocused(t *testing.T) {
	state := NewCheckboxState(true)
	// Put a Button first so it takes focus, leaving the checkbox unfocused
	widget := Column{
		Children: []Widget{
			&Button{ID: "focus-stealer", Label: ""},
			&Checkbox{
				ID:    "checkbox-checked-unfocused",
				State: state,
				Label: "Accept terms",
			},
		},
	}
	AssertSnapshot(t, widget, 25, 2, "Checked checkbox with label in default Text color (unfocused)")
}

func TestSnapshot_Checkbox_Unchecked_Focused(t *testing.T) {
	state := NewCheckboxState(false)
	// Checkbox is only focusable widget, so it gets focus
	widget := &Checkbox{
		ID:    "checkbox-focused",
		State: state,
		Label: "Accept terms",
	}
	AssertSnapshot(t, widget, 25, 1, "Unchecked checkbox with ActiveCursor background and SelectionText foreground")
}

func TestSnapshot_Checkbox_Checked_Focused(t *testing.T) {
	state := NewCheckboxState(true)
	// Checkbox is only focusable widget, so it gets focus
	widget := &Checkbox{
		ID:    "checkbox-checked-focused",
		State: state,
		Label: "Accept terms",
	}
	AssertSnapshot(t, widget, 25, 1, "Checked checkbox with ActiveCursor background and SelectionText foreground")
}

func TestSnapshot_Checkbox_Unchecked_Disabled(t *testing.T) {
	state := NewCheckboxState(false)
	widget := DisabledWhen(true, &Checkbox{
		ID:    "checkbox-disabled",
		State: state,
		Label: "Accept terms",
	})
	AssertSnapshot(t, widget, 25, 1, "Unchecked disabled checkbox with TextDisabled color")
}

func TestSnapshot_Checkbox_Checked_Disabled(t *testing.T) {
	state := NewCheckboxState(true)
	widget := DisabledWhen(true, &Checkbox{
		ID:    "checkbox-checked-disabled",
		State: state,
		Label: "Accept terms",
	})
	AssertSnapshot(t, widget, 25, 1, "Checked disabled checkbox with TextDisabled color")
}

func TestSnapshot_Checkbox_NoLabel(t *testing.T) {
	state := NewCheckboxState(true)
	// Checkbox is only focusable widget, so it gets focus
	widget := &Checkbox{
		ID:    "checkbox-no-label",
		State: state,
	}
	AssertSnapshot(t, widget, 3, 1, "Checked checkbox indicator only, no label text")
}

func TestSnapshot_Checkbox_Unchecked_NoLabel(t *testing.T) {
	state := NewCheckboxState(false)
	// Checkbox is only focusable widget, so it gets focus
	widget := &Checkbox{
		ID:    "checkbox-unchecked-no-label",
		State: state,
	}
	AssertSnapshot(t, widget, 3, 1, "Unchecked checkbox indicator only, no label text")
}
