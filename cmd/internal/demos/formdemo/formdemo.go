// Package formdemo exercises validation, focus, dirty state and input adapters.
package formdemo

import (
	"fmt"
	"strings"
	"unicode/utf8"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

type demo struct {
	scroll                                        *t.ScrollState
	name, confirm                                 *t.TextInputState
	bio                                           *t.TextAreaState
	terms                                         *t.CheckboxState
	nameField, confirmField, bioField, termsField *t.FieldState
	form                                          *t.FormState
	submits, changes, blurs                       t.Signal[int]
}

func newDemo() *demo {
	d := &demo{scroll: t.NewScrollState(), name: t.NewTextInputState(""), confirm: t.NewTextInputState(""), bio: t.NewTextAreaState(""), terms: t.NewCheckboxState(false), submits: t.NewSignal(0), changes: t.NewSignal(0), blurs: t.NewSignal(0)}
	d.nameField = t.NewTextInputField("name", d.name, t.Required("Enter a name"), func(value string) string {
		if value != "" && utf8.RuneCountInString(strings.TrimSpace(value)) < 3 {
			return "Use at least 3 Unicode characters"
		}
		return ""
	})
	d.confirmField = t.NewTextInputField("confirm", d.confirm, t.Required("Confirm your name"), func(value string) string {
		if value != "" && value != strings.Join(d.name.Content.Get(), "") {
			return "Must match Name"
		}
		return ""
	})
	d.bioField = t.NewTextAreaField("bio", d.bio)
	d.termsField = t.NewFieldState("terms", t.FieldBinding[bool]{Read: d.terms.Checked.Get, Write: d.terms.SetChecked}, func(value bool) string {
		if !value {
			return "Accept the terms to continue"
		}
		return ""
	})
	d.form = t.NewFormState(d.nameField, d.confirmField, d.bioField, d.termsField)
	return d
}

func (d *demo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	inputStyle := t.Style{Width: t.Flex(1), BackgroundColor: theme.Surface}
	focus := "none"
	if focused, ok := ctx.Focused().(t.Identifiable); ok {
		focus = focused.WidgetID()
	}
	form := t.Form{State: d.form, OnSubmit: func() { d.submits.Update(func(n int) int { return n + 1 }) }}
	form.Child = t.Column{Spacing: 1, Style: t.Style{Width: t.Flex(1), MaxWidth: t.Cells(64)}, Children: []t.Widget{
		t.Field{Label: "Name", Help: "Public name; Unicode accepted", State: d.nameField, Child: t.TextInput{State: d.name, Style: inputStyle, OnChange: func(string) { d.changes.Update(func(n int) int { return n + 1 }) }, Blur: func() { d.blurs.Update(func(n int) int { return n + 1 }) }}},
		t.Field{Label: "Confirm name", State: d.confirmField, Child: t.TextInput{State: d.confirm, Style: inputStyle}},
		t.Field{Label: "Bio (optional)", Help: "Enter inserts a newline here. Ctrl+S submits.", State: d.bioField, Child: t.TextArea{State: d.bio, Style: t.Style{Width: t.Flex(1), Height: t.Cells(2), BackgroundColor: theme.Surface}}},
		t.Field{State: d.termsField, Child: &t.Checkbox{State: d.terms, Label: "Accept terms"}},
		t.Row{Spacing: 1, Children: []t.Widget{
			t.Button{ID: "save", Label: "Submit", OnPress: func() { form.Submit() }},
			t.Button{ID: "reset", Label: "Reset", OnPress: d.form.Reset},
			t.Button{ID: "accept", Label: "Accept baseline", OnPress: d.form.Accept},
		}},
		t.Row{Spacing: 1, Children: []t.Widget{
			t.Button{ID: "load", Label: "Load Alice", OnPress: func() {
				d.name.SetText("Alice")
				d.confirm.SetText("Alice")
				d.terms.SetChecked(true)
				d.bio.SetText("Loaded from code")
			}},
			t.Button{ID: "toggle", Label: "Toggle confirm", OnPress: func() { d.confirmField.SetEnabled(!d.confirmField.Enabled()) }},
		}},
		t.Text{Content: fmt.Sprintf("Valid=%t Dirty=%t Submissions=%d Focus=%s", d.form.Valid(), d.form.Dirty(), d.submits.Get(), focus)},
		t.Text{Content: fmt.Sprintf("Name: touched=%t errors=%d changes=%d blurs=%d", d.nameField.Touched(), len(d.nameField.Errors()), d.changes.Get(), d.blurs.Get())},
	}}
	body := t.Column{Style: t.Style{Padding: t.EdgeInsets{Left: 2, Top: 1}}, Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Field + Form · validation without duplicate input state", Style: t.Style{Bold: true}},
		t.Text{Content: "Tab validates on blur · Enter submits single line · Ctrl+S submits"}, form,
	}}
	return t.Dock{Style: t.Style{BackgroundColor: theme.Background}, Top: []t.Widget{demokit.Header{Title: "Forms", Tagline: "Validation and input state"}}, Bottom: []t.Widget{demokit.Footer(theme)}, Body: t.Scrollable{State: d.scroll, Width: t.Flex(1), Height: t.Flex(1), Child: body}}
}

// Info describes this demo in the gallery.
var Info = demokit.Info{Key: "form", Title: "Forms", Description: "Validation, dirty state, reset and first-invalid focus"}

// New creates a gallery or standalone demo.
func New() demokit.Demo              { return newDemo() }
func (d *demo) InitialFocus() string { return d.nameField.InputID() }
func (d *demo) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme}}
}
