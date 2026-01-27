package terma

import "testing"

func TestSnapshot_Dialog_WithTitleAndTwoButtons(t *testing.T) {
	widget := Dialog{
		ID:      "dlg",
		Visible: true,
		Title:   "Confirm",
		Content: Text{Content: "Are you sure you want to proceed?"},
		Buttons: []Button{
			{Label: "Cancel"},
			{Label: "OK"},
		},
	}
	AssertSnapshot(t, widget, 50, 12,
		"Centered modal dialog with rounded border, title 'Confirm' at top-center. "+
			"Body text 'Are you sure you want to proceed?'. Two buttons 'Cancel' and 'OK' right-aligned at bottom. "+
			"Semi-transparent backdrop behind dialog.")
}

func TestSnapshot_Dialog_WithoutTitle(t *testing.T) {
	widget := Dialog{
		ID:      "dlg-no-title",
		Visible: true,
		Content: Text{Content: "Something happened."},
		Buttons: []Button{
			{Label: "OK"},
		},
	}
	AssertSnapshot(t, widget, 50, 10,
		"Centered modal dialog with rounded border but NO title. "+
			"Body text 'Something happened.'. Single 'OK' button right-aligned.")
}

func TestSnapshot_Dialog_SingleButton(t *testing.T) {
	widget := Dialog{
		ID:      "dlg-single",
		Visible: true,
		Title:   "Info",
		Content: Text{Content: "Operation complete."},
		Buttons: []Button{
			{Label: "Done"},
		},
	}
	AssertSnapshot(t, widget, 50, 10,
		"Centered modal dialog with title 'Info'. Body text 'Operation complete.'. "+
			"Single 'Done' button right-aligned.")
}

func TestSnapshot_Dialog_VariantButtons(t *testing.T) {
	widget := Dialog{
		ID:      "dlg-variant",
		Visible: true,
		Title:   "Delete Item",
		Content: Text{Content: "This action cannot be undone."},
		Buttons: []Button{
			{Label: "Cancel"},
			{Label: "Delete", Variant: ButtonError},
		},
	}
	AssertSnapshot(t, widget, 50, 12,
		"Centered modal dialog with title 'Delete Item'. Body text 'This action cannot be undone.'. "+
			"Two buttons: 'Cancel' (default styling) and 'Delete' (error/red variant) right-aligned.")
}
