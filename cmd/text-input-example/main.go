package main

import (
	"fmt"
	"log"

	t "terma"
)

// TextInputDemo demonstrates the TextInput widget with multiple inputs,
// showing placeholder text, change tracking, and form submission.
type TextInputDemo struct {
	nameState    *t.TextInputState
	emailState   *t.TextInputState
	messageState *t.TextInputState

	submittedData t.Signal[string]
	charCount     t.Signal[int]
}

func NewTextInputDemo() *TextInputDemo {
	return &TextInputDemo{
		nameState:     t.NewTextInputState(""),
		emailState:    t.NewTextInputState(""),
		messageState:  t.NewTextInputState(""),
		submittedData: t.NewSignal(""),
		charCount:     t.NewSignal(0),
	}
}

func (d *TextInputDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()

	return t.Dock{
		ID: "text-input-demo-root",
		Bottom: []t.Widget{
			t.KeybindBar{},
		},
		Body: t.Column{
			Spacing: 1,
			Style: t.Style{
				Padding: t.EdgeInsetsXY(2, 1),
			},
			Children: []t.Widget{
				// Title
				t.Text{
					Content: " TextInput Demo ",
					Style: t.Style{
						ForegroundColor: theme.TextOnPrimary,
						BackgroundColor: theme.Primary,
					},
				},

				// Instructions
				t.Text{
					Spans: t.ParseMarkup(
						"[b $Primary]Tab[/] to switch fields • [b $Primary]Enter[/] to submit • [b $Primary]Ctrl+C[/] to quit",
						theme,
					),
				},

				t.Text{Content: ""}, // Spacer

				// Name field
				d.buildField(ctx, "Name", "name-input", d.nameState, "Enter your name...", nil),

				// Email field
				d.buildField(ctx, "Email", "email-input", d.emailState, "Enter your email...", nil),

				// Message field with character counter
				d.buildField(ctx, "Message", "message-input", d.messageState, "Type a message...", func(text string) {
					d.charCount.Set(len([]rune(text)))
				}),

				// Character count display
				t.Text{
					Content: fmt.Sprintf("Character count: %d", d.charCount.Get()),
					Style: t.Style{
						ForegroundColor: theme.TextMuted,
					},
				},

				t.Text{Content: ""}, // Spacer

				// Submit button
				&t.Button{
					ID:    "submit-btn",
					Label: " Submit Form ",
					Width: t.Auto,
					Style: t.Style{
						BackgroundColor: theme.Primary,
						ForegroundColor: theme.TextOnPrimary,
					},
					OnPress: d.handleSubmit,
				},

				// Submission result
				t.ShowWhen(d.submittedData.Get() != "", t.Column{
					Spacing: 0,
					Children: []t.Widget{
						t.Text{Content: ""}, // Spacer
						t.Text{
							Content: "Submitted:",
							Style: t.Style{
								ForegroundColor: theme.Success,
							},
						},
						t.Text{
							Content: d.submittedData.Get(),
							Style: t.Style{
								ForegroundColor: theme.Text,
							},
						},
					},
				}),
			},
		},
	}
}

// buildField creates a labeled text input field.
func (d *TextInputDemo) buildField(ctx t.BuildContext, label, id string, state *t.TextInputState, placeholder string, onChange func(string)) t.Widget {
	theme := ctx.Theme()

	return t.Row{
		Spacing:    1,
		CrossAlign: t.CrossAxisCenter,
		Children: []t.Widget{
			t.Text{
				Content: fmt.Sprintf("%8s:", label),
				Style: t.Style{
					ForegroundColor: theme.Text,
				},
			},
			t.TextInput{
				ID:          id,
				State:       state,
				Placeholder: placeholder,
				Width:       t.Cells(40),
				Style: t.Style{
					BackgroundColor: theme.Surface,
					ForegroundColor: theme.Text,
				},
				OnChange: onChange,
				OnSubmit: func(text string) {
					d.handleSubmit()
				},
			},
		},
	}
}

func (d *TextInputDemo) handleSubmit() {
	name := d.nameState.GetText()
	email := d.emailState.GetText()
	message := d.messageState.GetText()

	if name == "" && email == "" && message == "" {
		d.submittedData.Set("(empty form)")
		return
	}

	result := fmt.Sprintf("Name: %s | Email: %s | Message: %s", name, email, message)
	d.submittedData.Set(result)
}

func main() {
	app := NewTextInputDemo()
	if err := t.Run(app); err != nil {
		log.Fatal(err)
	}
}
