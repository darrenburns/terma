package main

import (
	"fmt"

	"terma"
)

type App struct {
	formValid   terma.Signal[bool]
	clickCount  terma.Signal[int]
	lastClicked terma.Signal[string]
}

func (a *App) Build(ctx terma.BuildContext) terma.Widget {
	theme := ctx.Theme()
	valid := a.formValid.Get()
	count := a.clickCount.Get()
	last := a.lastClicked.Get()

	return terma.Column{
		Spacing: 1,
		Style: terma.Style{
			Padding:         terma.EdgeInsetsAll(2),
			BackgroundColor: theme.Background,
		},
		Children: []terma.Widget{
			// Header
			terma.Text{
				Content: "Button Disabled State Demo",
				Style: terma.Style{
					ForegroundColor: theme.Primary,
					Bold:            true,
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Status display
			terma.Text{
				Content: fmt.Sprintf("Form valid: %v | Clicks: %d | Last: %s", valid, count, last),
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Toggle button (always enabled)
			terma.Row{
				Spacing: 1,
				Children: []terma.Widget{
					terma.Text{Content: "Toggle form validity:"},
					&terma.Button{
						ID:    "toggle",
						Label: "Toggle Valid",
						OnPress: func() {
							a.formValid.Set(!a.formValid.Get())
						},
					},
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Conditionally disabled buttons
			terma.Text{
				Content: "Conditionally Disabled (disabled when form invalid):",
				Style:   terma.Style{ForegroundColor: theme.Accent},
			},

			terma.Row{
				Spacing: 2,
				Children: []terma.Widget{
					terma.DisabledWhen(!valid, &terma.Button{
						ID:    "submit",
						Label: "Submit",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Submit")
						},
					}),
					terma.DisabledWhen(!valid, &terma.Button{
						ID:    "save",
						Label: "Save Draft",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Save Draft")
						},
					}),
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Using EnabledWhen (inverse)
			terma.Text{
				Content: "Using EnabledWhen (enabled when form valid):",
				Style:   terma.Style{ForegroundColor: theme.Accent},
			},

			terma.Row{
				Spacing: 2,
				Children: []terma.Widget{
					terma.EnabledWhen(valid, &terma.Button{
						ID:    "publish",
						Label: "Publish",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Publish")
						},
					}),
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Always enabled buttons for comparison
			terma.Text{
				Content: "Always Enabled (for comparison):",
				Style:   terma.Style{ForegroundColor: theme.Accent},
			},

			terma.Row{
				Spacing: 2,
				Children: []terma.Widget{
					&terma.Button{
						ID:    "cancel",
						Label: "Cancel",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Cancel")
						},
					},
					&terma.Button{
						ID:    "reset",
						Label: "Reset",
						OnPress: func() {
							a.clickCount.Set(0)
							a.lastClicked.Set("Reset")
						},
					},
				},
			},

			terma.Spacer{Height: terma.Cells(1)},

			// Section: Nested disabled (whole group disabled)
			terma.Text{
				Content: "Nested DisabledWhen (entire group disabled when invalid):",
				Style:   terma.Style{ForegroundColor: theme.Accent},
			},

			terma.DisabledWhen(!valid, terma.Row{
				Spacing: 2,
				Children: []terma.Widget{
					&terma.Button{
						ID:    "action1",
						Label: "Action 1",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Action 1")
						},
					},
					&terma.Button{
						ID:    "action2",
						Label: "Action 2",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Action 2")
						},
					},
					&terma.Button{
						ID:    "action3",
						Label: "Action 3",
						OnPress: func() {
							a.clickCount.Update(func(c int) int { return c + 1 })
							a.lastClicked.Set("Action 3")
						},
					},
				},
			}),

			terma.Spacer{Height: terma.Flex(1)},

			// Footer
			terma.Text{
				Content: "Tab/Shift+Tab to navigate, Enter/Space to press, q to quit",
				Style:   terma.Style{ForegroundColor: theme.TextMuted},
			},
		},
	}
}

func (a *App) Keybinds() []terma.Keybind {
	return []terma.Keybind{
		{Key: "q", Name: "Quit", Action: terma.Quit},
	}
}

func main() {
	terma.Run(&App{
		formValid:   terma.NewSignal(true),
		clickCount:  terma.NewSignal(0),
		lastClicked: terma.NewSignal("(none)"),
	})
}
