package main

import (
	"fmt"

	"terma"
)

type App struct {
	// Form checkboxes
	termsAccepted    *terma.CheckboxState
	newsletter       *terma.CheckboxState
	marketingEmails  *terma.CheckboxState

	// Feature toggles
	darkMode         *terma.CheckboxState
	notifications    *terma.CheckboxState
	autoSave         *terma.CheckboxState

	// Status tracking
	lastChanged      terma.Signal[string]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()
	lastChanged := a.lastChanged.Get()

	// Calculate if form is complete (all required checked)
	formComplete := a.termsAccepted.Checked.Get()

	return terma.Column{
		Spacing: 1,
		Style: terma.Style{
			Padding:         terma.EdgeInsetsAll(2),
			BackgroundColor: theme.Background,
		},
		Children: []terma.Widget{
			// Header
			terma.Text{
				Content: "Checkbox Widget Demo",
				Style: terma.Style{
					ForegroundColor: theme.Primary,
					Bold:            true,
				},
			},

			terma.Text{
				Content: "Use arrow keys to navigate, Space/Enter to toggle",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Status
			terma.Text{
				Content: fmt.Sprintf("Last changed: %s", lastChanged),
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Registration Form
			terma.Text{
				Content: "Registration Form",
				Style: terma.Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},

			a.buildCheckbox(ctx, "terms", a.termsAccepted, "I accept the Terms of Service (required)"),
			a.buildCheckbox(ctx, "newsletter", a.newsletter, "Subscribe to newsletter"),
			a.buildCheckbox(ctx, "marketing", a.marketingEmails, "Receive marketing emails"),

			terma.Spacer{Height: terma.Cells(1)},

			// Submit button (disabled until terms accepted)
			terma.Row{
				Spacing: 1,
				Children: []terma.Widget{
					terma.DisabledWhen(!formComplete, terma.Button{
						ID:    "submit",
						Label: "Submit Registration",
						OnPress: func() {
							a.lastChanged.Set("Form submitted!")
						},
					}),
					terma.ShowWhen(!formComplete, terma.Text{
						Content: "(Accept terms to enable)",
						Style:   terma.Style{ForegroundColor: theme.TextMuted},
					}),
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Settings
			terma.Text{
				Content: "Settings",
				Style: terma.Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},

			a.buildCheckbox(ctx, "darkmode", a.darkMode, "Enable dark mode"),
			a.buildCheckbox(ctx, "notifications", a.notifications, "Enable notifications"),
			a.buildCheckbox(ctx, "autosave", a.autoSave, "Auto-save documents"),

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Disabled checkboxes
			terma.Text{
				Content: "Disabled Checkboxes (for comparison)",
				Style: terma.Style{
					ForegroundColor: theme.Accent,
					Bold:            true,
				},
			},

			terma.DisabledWhen(true, &terma.Checkbox{
				ID:    "disabled-unchecked",
				State: terma.NewCheckboxState(false),
				Label: "Disabled unchecked",
			}),

			terma.DisabledWhen(true, &terma.Checkbox{
				ID:    "disabled-checked",
				State: terma.NewCheckboxState(true),
				Label: "Disabled checked",
			}),

			terma.Spacer{Height: terma.Cells(1)},

			// Current state summary
			terma.Text{
				Content: "Current State:",
				Style: terma.Style{
					ForegroundColor: theme.Secondary,
					Bold:            true,
				},
			},

			a.buildStatusLine(ctx, "Terms accepted", a.termsAccepted.Checked.Get()),
			a.buildStatusLine(ctx, "Newsletter", a.newsletter.Checked.Get()),
			a.buildStatusLine(ctx, "Marketing", a.marketingEmails.Checked.Get()),
			a.buildStatusLine(ctx, "Dark mode", a.darkMode.Checked.Get()),
			a.buildStatusLine(ctx, "Notifications", a.notifications.Checked.Get()),
			a.buildStatusLine(ctx, "Auto-save", a.autoSave.Checked.Get()),
		},
	}
}

func (a *App) buildCheckbox(ctx terma.BuildContext, id string, state *terma.CheckboxState, label string) terma.Widget {
	return &terma.Checkbox{
		ID:    id,
		State: state,
		Label: label,
		OnChange: func(checked bool) {
			status := "unchecked"
			if checked {
				status = "checked"
			}
			a.lastChanged.Set(fmt.Sprintf("%s: %s", label, status))
		},
	}
}

func (a *App) buildStatusLine(ctx terma.BuildContext, label string, checked bool) terma.Widget {
	theme := ctx.Theme()
	indicator := "[ ]"
	color := theme.TextMuted
	if checked {
		indicator = "[x]"
		color = theme.Success
	}
	return terma.Text{
		Content: fmt.Sprintf("  %s %s", indicator, label),
		Style:   terma.Style{ForegroundColor: color},
	}
}

func main() {
	app := &App{
		termsAccepted:   terma.NewCheckboxState(false),
		newsletter:      terma.NewCheckboxState(true),
		marketingEmails: terma.NewCheckboxState(false),
		darkMode:        terma.NewCheckboxState(true),
		notifications:   terma.NewCheckboxState(true),
		autoSave:        terma.NewCheckboxState(false),
		lastChanged:     terma.NewSignal("(none)"),
	}
	terma.Run(app)
}
