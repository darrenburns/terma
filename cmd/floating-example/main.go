package main

import (
	"fmt"
	"log"

	t "terma"
)

type App struct {
	showConfirm t.Signal[bool]
	showForm    t.Signal[bool]
	showInfo    t.Signal[bool]
	statusMsg   t.Signal[string]
	nameInput   *t.TextInputState
	emailInput  *t.TextInputState
}

func (a *App) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "q", Name: "Quit", Action: t.Quit},
	}
}

func (a *App) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		Bottom: []t.Widget{t.KeybindBar{}},
		Body: t.Column{
			Width:  t.Flex(1),
			Height: t.Flex(1),
			Style: t.Style{
				Padding:         t.EdgeInsetsXY(3, 1),
				BackgroundColor: theme.Background,
			},
			Spacing: 1,
			Children: []t.Widget{
				// Header
				t.Text{
					Spans: []t.Span{
						{Text: "Modal Dialogs", Style: t.SpanStyle{Bold: true}},
					},
				},
				t.Text{
					Content: "Open a modal to see focus trapping in action. Tab/Shift+Tab stays inside the dialog.",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Spacer{Height: t.Cells(1)},

				// Trigger buttons
				t.Row{
					Spacing: 2,
					Children: []t.Widget{
						&t.Button{
							ID:    "confirm-trigger",
							Label: "Confirmation",
							OnPress: func() {
								a.showConfirm.Set(true)
							},
						},
						&t.Button{
							ID:    "form-trigger",
							Label: "Form",
							OnPress: func() {
								a.nameInput.SetText("")
								a.emailInput.SetText("")
								a.showForm.Set(true)
							},
						},
						&t.Button{
							ID:    "info-trigger",
							Label: "Info",
							OnPress: func() {
								a.showInfo.Set(true)
							},
						},
					},
				},

				// Status
				t.ShowWhen(a.statusMsg.Get() != "", t.Text{
					Spans: []t.Span{
						{Text: a.statusMsg.Get(), Style: t.SpanStyle{Foreground: theme.Success}},
					},
				}),

				// Modals
				a.confirmDialog(ctx),
				a.formDialog(ctx),
				a.infoDialog(ctx),
			},
		},
	}
}

// confirmDialog is a simple yes/no confirmation modal.
func (a *App) confirmDialog(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Floating{
		Visible: a.showConfirm.Get(),
		Config: t.FloatConfig{
			Position:  t.FloatPositionCenter,
			Modal:     true,
			OnDismiss: func() { a.showConfirm.Set(false) },
		},
		Child: t.Column{
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Border: t.Border{
					Style: t.BorderRounded,
					Color: theme.Border,
					Decorations: []t.BorderDecoration{
						t.BorderTitleMarkup("[b $Primary] Confirm [/]"),
					},
				},
				Padding: t.EdgeInsets{Left: 3, Right: 3, Top: 1, Bottom: 1},
			},
			Spacing: 1,
			Children: []t.Widget{
				t.Text{Content: "Delete all items?"},
				t.Text{
					Content: "This action cannot be undone.",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						&t.Button{
							ID:    "confirm-yes",
							Label: "Delete",
							Style: t.Style{BackgroundColor: theme.Error, ForegroundColor: theme.TextOnPrimary},
							OnPress: func() {
								a.statusMsg.Set("Deleted.")
								a.showConfirm.Set(false)
							},
						},
						&t.Button{
							ID:    "confirm-no",
							Label: "Cancel",
							OnPress: func() {
								a.statusMsg.Set("Cancelled.")
								a.showConfirm.Set(false)
							},
						},
					},
				},
			},
		},
	}
}

// formDialog is a modal with text inputs demonstrating focus trapping.
func (a *App) formDialog(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Floating{
		Visible: a.showForm.Get(),
		Config: t.FloatConfig{
			Position:  t.FloatPositionCenter,
			Modal:     true,
			OnDismiss: func() { a.showForm.Set(false) },
		},
		Child: t.Column{
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Border: t.Border{
					Style: t.BorderRounded,
					Color: theme.Border,
					Decorations: []t.BorderDecoration{
						t.BorderTitleMarkup("[b $Primary] New User [/]"),
					},
				},
				Padding: t.EdgeInsets{Left: 3, Right: 3, Top: 1, Bottom: 1},
			},
			Spacing: 1,
			Children: []t.Widget{
				// Name field
				t.Text{Content: "Name"},
				t.TextInput{
					ID:          "form-name",
					State:       a.nameInput,
					Placeholder: "Jane Doe",
					Style:       t.Style{Width: t.Cells(30)},
				},

				// Email field
				t.Text{Content: "Email"},
				t.TextInput{
					ID:          "form-email",
					State:       a.emailInput,
					Placeholder: "jane@example.com",
					Style:       t.Style{Width: t.Cells(30)},
				},

				// Actions
				t.Row{
					Spacing: 1,
					Children: []t.Widget{
						&t.Button{
							ID:    "form-submit",
							Label: "Create",
							OnPress: func() {
								name := a.nameInput.GetText()
								email := a.emailInput.GetText()
								if name == "" {
									name = "(empty)"
								}
								a.statusMsg.Set(fmt.Sprintf("Created user %s <%s>", name, email))
								a.showForm.Set(false)
							},
						},
						&t.Button{
							ID:    "form-cancel",
							Label: "Cancel",
							OnPress: func() {
								a.showForm.Set(false)
							},
						},
					},
				},
			},
		},
	}
}

// infoDialog is a simple informational modal with a single dismiss button.
func (a *App) infoDialog(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Floating{
		Visible: a.showInfo.Get(),
		Config: t.FloatConfig{
			Position:  t.FloatPositionCenter,
			Modal:     true,
			OnDismiss: func() { a.showInfo.Set(false) },
		},
		Child: t.Column{
			Style: t.Style{
				BackgroundColor: theme.Surface,
				Border: t.Border{
					Style: t.BorderRounded,
					Color: theme.Border,
					Decorations: []t.BorderDecoration{
						t.BorderTitleMarkup("[b $Info] About [/]"),
					},
				},
				Padding: t.EdgeInsets{Left: 3, Right: 3, Top: 1, Bottom: 1},
			},
			Spacing: 1,
			Children: []t.Widget{
				t.Text{
					Spans: []t.Span{
						{Text: "Terma", Style: t.SpanStyle{Bold: true}},
						{Text: " is a declarative terminal UI framework for Go."},
					},
				},
				t.Text{
					Content: "Modal floats automatically trap focus — Tab and",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Text{
					Content: "Shift+Tab only cycle within the dialog. Press",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				t.Text{
					Content: "Escape or click the button below to close.",
					Style:   t.Style{ForegroundColor: theme.TextMuted},
				},
				&t.Button{
					ID:    "info-ok",
					Label: "OK",
					OnPress: func() {
						a.showInfo.Set(false)
					},
				},
			},
		},
	}
}

func main() {
	app := &App{
		showConfirm: t.NewSignal(false),
		showForm:    t.NewSignal(false),
		showInfo:    t.NewSignal(false),
		statusMsg:   t.NewSignal(""),
		nameInput:   t.NewTextInputState(""),
		emailInput:  t.NewTextInputState(""),
	}
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
