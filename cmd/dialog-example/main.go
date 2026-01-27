package main

import (
	"fmt"
	"log"

	t "terma"
)

type DialogDemo struct {
	activeDialog t.Signal[string] // which dialog is open ("" = none)
	statusMsg    t.Signal[string]
}

func NewDialogDemo() *DialogDemo {
	return &DialogDemo{
		activeDialog: t.NewSignal(""),
		statusMsg:    t.NewSignal("Select a dialog type to preview"),
	}
}

func (d *DialogDemo) dismiss() {
	d.activeDialog.Set("")
}

func (d *DialogDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	active := d.activeDialog.Get()

	return t.Dock{
		Top: []t.Widget{
			t.Column{
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(2, 1),
					BackgroundColor: theme.Surface,
				},
				Spacing: 1,
				Children: []t.Widget{
					t.Text{
						Spans: t.ParseMarkup("[b $Primary]Dialog Examples[/]", theme),
					},
					t.Text{
						Content: "Each button opens a different style of dialog.",
						Style:   t.Style{ForegroundColor: theme.TextMuted},
					},
				},
			},
		},
		Bottom: []t.Widget{
			t.Row{
				Style: t.Style{
					Padding:         t.EdgeInsetsXY(2, 0),
					BackgroundColor: theme.Surface,
				},
				Children: []t.Widget{
					t.Text{
						Content: d.statusMsg.Get(),
						Style:   t.Style{ForegroundColor: theme.TextMuted, Italic: true},
					},
				},
			},
			t.KeybindBar{},
		},
		Body: t.Column{
			Style: t.Style{
				Padding:         t.EdgeInsetsXY(2, 1),
				BackgroundColor: theme.Background,
			},
			Spacing: 1,
			Children: []t.Widget{
				d.section("Informational", []t.Widget{
					t.Button{
						ID:    "btn-info",
						Label: "Info",
						OnPress: func() {
							d.activeDialog.Set("info")
						},
					},
					t.Button{
						ID:    "btn-success",
						Label: "Success",
						OnPress: func() {
							d.activeDialog.Set("success")
						},
					},
					t.Button{
						ID:    "btn-warning",
						Label: "Warning",
						OnPress: func() {
							d.activeDialog.Set("warning")
						},
					},
				}),
				d.section("Actions", []t.Widget{
					t.Button{
						ID:    "btn-confirm",
						Label: "Confirm Action",
						OnPress: func() {
							d.activeDialog.Set("confirm")
						},
					},
					t.Button{
						ID:    "btn-delete",
						Label: "Delete Item",
						OnPress: func() {
							d.activeDialog.Set("delete")
						},
					},
					t.Button{
						ID:    "btn-unsaved",
						Label: "Unsaved Changes",
						OnPress: func() {
							d.activeDialog.Set("unsaved")
						},
					},
				}),
				d.section("Content", []t.Widget{
					t.Button{
						ID:    "btn-rich",
						Label: "Rich Content",
						OnPress: func() {
							d.activeDialog.Set("rich")
						},
					},
					t.Button{
						ID:    "btn-form",
						Label: "Form Dialog",
						OnPress: func() {
							d.activeDialog.Set("form")
						},
					},
				}),

				// Dialogs (Floating-based, so they register with the float collector
				// and render as overlays regardless of where they appear in the tree)

				// Info dialog
				t.Dialog{
					ID:      "dlg-info",
					Visible: active == "info",
					Title:   "Information",
					Content: t.Text{
						Content: "This operation completed successfully. No further action is required.",
						Wrap:    t.WrapSoft,
					},
					Buttons: []t.Button{
						{Label: "OK", Variant: t.ButtonInfo, OnPress: func() {
							d.statusMsg.Set("Info dialog dismissed")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Success dialog
				t.Dialog{
					ID:      "dlg-success",
					Visible: active == "success",
					Title:   "Success",
					Content: t.Text{
						Content: "Your changes have been saved.",
						Wrap:    t.WrapSoft,
					},
					Buttons: []t.Button{
						{Label: "Great!", Variant: t.ButtonSuccess, OnPress: func() {
							d.statusMsg.Set("Success acknowledged")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Warning dialog
				t.Dialog{
					ID:      "dlg-warning",
					Visible: active == "warning",
					Title:   "Warning",
					Content: t.Text{
						Content: "Your session will expire in 5 minutes. Save your work to avoid losing changes.",
						Wrap:    t.WrapSoft,
					},
					Buttons: []t.Button{
						{Label: "Dismiss", OnPress: func() {
							d.statusMsg.Set("Warning dismissed")
							d.dismiss()
						}},
						{Label: "Save Now", Variant: t.ButtonWarning, OnPress: func() {
							d.statusMsg.Set("Work saved")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Confirm dialog
				t.Dialog{
					ID:      "dlg-confirm",
					Visible: active == "confirm",
					Title:   "Confirm Action",
					Content: t.Text{
						Content: "Are you sure you want to proceed? This will apply the pending changes.",
						Wrap:    t.WrapSoft,
					},
					Buttons: []t.Button{
						{Label: "Cancel", OnPress: func() {
							d.statusMsg.Set("Action cancelled")
							d.dismiss()
						}},
						{Label: "Confirm", Variant: t.ButtonPrimary, OnPress: func() {
							d.statusMsg.Set("Action confirmed!")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Delete dialog
				t.Dialog{
					ID:      "dlg-delete",
					Visible: active == "delete",
					Title:   "Delete Item",
					Content: t.Column{
						Spacing: 1,
						Children: []t.Widget{
							t.Text{
								Content: "Are you sure you want to delete this item?",
								Wrap:    t.WrapSoft,
							},
							t.Text{
								Content: "This action cannot be undone.",
								Wrap:    t.WrapSoft,
								Style:   t.Style{ForegroundColor: theme.TextMuted},
							},
						},
					},
					Buttons: []t.Button{
						{Label: "Cancel", OnPress: func() {
							d.statusMsg.Set("Delete cancelled")
							d.dismiss()
						}},
						{Label: "Delete", Variant: t.ButtonError, OnPress: func() {
							d.statusMsg.Set("Item deleted")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Unsaved changes dialog
				t.Dialog{
					ID:      "dlg-unsaved",
					Visible: active == "unsaved",
					Title:   "Unsaved Changes",
					Content: t.Text{
						Content: "You have unsaved changes. What would you like to do?",
						Wrap:    t.WrapSoft,
					},
					Buttons: []t.Button{
						{Label: "Discard", Variant: t.ButtonError, OnPress: func() {
							d.statusMsg.Set("Changes discarded")
							d.dismiss()
						}},
						{Label: "Cancel", OnPress: func() {
							d.statusMsg.Set("Resumed editing")
							d.dismiss()
						}},
						{Label: "Save", Variant: t.ButtonSuccess, OnPress: func() {
							d.statusMsg.Set("Changes saved")
							d.dismiss()
						}},
					},
					OnDismiss: d.dismiss,
				},

				// Rich content dialog
				d.richContentDialog(active == "rich", theme),

				// Form dialog
				d.formDialog(active == "form", theme),
			},
		},
	}
}

func (d *DialogDemo) section(title string, buttons []t.Widget) t.Widget {
	return t.Column{
		Spacing: 1,
		Children: []t.Widget{
			t.Text{
				Content: title,
				Style:   t.Style{Bold: true},
			},
			t.Row{
				Spacing:  2,
				Children: buttons,
			},
		},
	}
}

func (d *DialogDemo) richContentDialog(visible bool, theme t.ThemeData) t.Widget {
	return t.Dialog{
		ID:      "dlg-rich",
		Visible: visible,
		Title:   "Release Notes",
		Content: t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Text{
					Spans: t.ParseMarkup("[b]Version 2.0[/] is now available!", theme),
				},
				t.Text{Content: "What's new:", Style: t.Style{ForegroundColor: theme.TextMuted}},
				t.Column{
					Children: []t.Widget{
						t.Text{Spans: t.ParseMarkup("  [b $Success]\u2713[/] New dialog widget", theme)},
						t.Text{Spans: t.ParseMarkup("  [b $Success]\u2713[/] Button variants", theme)},
						t.Text{Spans: t.ParseMarkup("  [b $Success]\u2713[/] Improved focus management", theme)},
					},
				},
			},
		},
		Buttons: []t.Button{
			{Label: "Later", OnPress: func() {
				d.statusMsg.Set("Remind me later")
				d.dismiss()
			}},
			{Label: "Update Now", Variant: t.ButtonAccent, OnPress: func() {
				d.statusMsg.Set("Updating...")
				d.dismiss()
			}},
		},
		OnDismiss: d.dismiss,
	}
}

func (d *DialogDemo) formDialog(visible bool, theme t.ThemeData) t.Widget {
	items := []string{"Bug Report", "Feature Request", "Question", "Other"}
	return t.Dialog{
		ID:      "dlg-form",
		Visible: visible,
		Title:   "Submit Feedback",
		Content: t.Column{
			Spacing: 1,
			Children: []t.Widget{
				t.Text{Content: "Select a feedback category:"},
				t.Column{
					Children: func() []t.Widget {
						var ws []t.Widget
						for i, item := range items {
							ws = append(ws, t.Text{
								Content: fmt.Sprintf("  %d. %s", i+1, item),
								Style:   t.Style{ForegroundColor: theme.TextMuted},
							})
						}
						return ws
					}(),
				},
			},
		},
		Buttons: []t.Button{
			{Label: "Cancel", OnPress: func() {
				d.statusMsg.Set("Feedback cancelled")
				d.dismiss()
			}},
			{Label: "Submit", Variant: t.ButtonPrimary, OnPress: func() {
				d.statusMsg.Set("Feedback submitted, thank you!")
				d.dismiss()
			}},
		},
		OnDismiss: d.dismiss,
	}
}

func main() {
	_ = t.InitLogger()
	app := NewDialogDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
