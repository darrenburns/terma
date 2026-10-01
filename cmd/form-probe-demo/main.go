// Command form-probe-demo exercises validation-driven focus through nested scroll views.
package main

import (
	"flag"
	"fmt"
	t "github.com/darrenburns/terma"
	"log"
	"strings"
)

type scene struct {
	head, tail, other                *t.TextInputState
	headField, tailField, otherField *t.FieldState
	main, secondary                  *t.FormState
	inner, outer                     *t.ScrollState
	saves, otherSaves                t.Signal[int]
}

func newScene() *scene {
	s := &scene{head: t.NewTextInputState("ready"), tail: t.NewTextInputState(""), other: t.NewTextInputState("independent"), inner: t.NewScrollState(), outer: t.NewScrollState(), saves: t.NewSignal(0), otherSaves: t.NewSignal(0)}
	s.headField = t.NewTextInputField("probe-head", s.head, t.Required("Header required"))
	s.tailField = t.NewTextInputField("probe-tail", s.tail, t.Required("Detail required"), func(value string) string {
		if value != "" && value != strings.Join(s.head.Content.Get(), "") {
			return "Detail must match header"
		}
		return ""
	})
	s.otherField = t.NewTextInputField("probe-other", s.other, t.Required("Other required"))
	s.main = t.NewFormState(s.tailField, s.headField) // Explicit order differs from screen order.
	s.secondary = t.NewFormState(s.otherField)
	return s
}
func (s *scene) mainForm() t.Form {
	return t.Form{State: s.main, OnSubmit: func() { s.saves.Update(func(v int) int { return v + 1 }) }}
}
func (s *scene) top() { s.inner.SetOffset(0); s.outer.SetOffset(0) }
func (s *scene) reset() {
	s.main.Reset()
	s.head.SetText("")
	s.tail.SetText("")
	s.top()
	t.RequestFocus("probe-head")
}
func (s *scene) load() { s.head.SetText("ready"); s.tail.SetText("ready") }
func (s *scene) Keybinds() []t.Keybind {
	return []t.Keybind{{Key: "f2", Name: "Submit main", Action: func() { s.mainForm().Submit() }}, {Key: "f6", Name: "Scroll top", Action: s.top}, {Key: "f7", Name: "Reset invalid", Action: s.reset}, {Key: "f8", Name: "Valid values", Action: s.load}}
}
func (s *scene) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	input := t.Style{Width: t.Flex(1), BackgroundColor: theme.Surface}
	form := s.mainForm()
	form.Child = t.Column{Spacing: 1, Children: []t.Widget{
		t.Row{Spacing: 1, Children: []t.Widget{t.Button{ID: "probe-submit", Label: "Submit F2", OnPress: func() { form.Submit() }}, t.Button{ID: "probe-top", Label: "Top F6", OnPress: s.top}, t.Button{ID: "probe-reset", Label: "Invalid F7", OnPress: s.reset}, t.Button{ID: "probe-load", Label: "Valid F8", OnPress: s.load}}},
		t.Scrollable{State: s.outer, Style: t.Style{Width: t.Cells(58), Height: t.Cells(11)}, Child: t.Column{Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
			t.Text{Content: "Outer scrolling panel\nRegistry validates Detail before Header."},
			t.FocusTrap{ID: "probe-inner-trap", Active: true, Child: t.Scrollable{State: s.inner, Style: t.Style{Width: t.Cells(52), Height: t.Cells(7)}, Child: t.Column{Spacing: 1, Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
				t.Field{State: s.headField, Label: "Header", Child: t.TextInput{State: s.head, Style: input}},
				t.Text{Content: "Additional section 1\nAdditional section 2\nAdditional section 3\nAdditional section 4\nAdditional section 5"},
				t.Field{State: s.tailField, Label: "Detail (first in validation order)", Child: t.TextInput{State: s.tail, Style: input}},
			}}}},
		}}},
	}}
	secondary := t.Form{State: s.secondary, OnSubmit: func() { s.otherSaves.Update(func(v int) int { return v + 1 }) }, Child: t.Field{Label: "Independent form · Enter submits only this form", State: s.otherField, Child: t.TextInput{State: s.other, Style: input}}}
	focus := "none"
	if f, ok := ctx.Focused().(t.Identifiable); ok {
		focus = f.WidgetID()
	}
	return t.Column{Spacing: 1, Style: t.Style{Padding: t.EdgeInsetsAll(1), Width: t.Flex(1)}, Children: []t.Widget{
		t.Text{Content: "FORM PROBE · nested scrolling and independent forms", Style: t.Style{Bold: true, ForegroundColor: theme.Primary}},
		form, secondary,
		t.Text{Content: fmt.Sprintf("Main=%d Other=%d Valid=%t Focus=%s", s.saves.Get(), s.otherSaves.Get(), s.main.Valid(), focus)},
		t.Text{Content: fmt.Sprintf("Header errors=%d Detail errors=%d", len(s.headField.Errors()), len(s.tailField.Errors()))},
		t.Text{Content: "F6 scrolls without moving focus · Ctrl+S retries current form"},
	}}
}
func main() {
	modal := flag.Bool("dialog", false, "embed probe inside a Dialog")
	flag.Parse()
	var root t.Widget = newScene()
	if *modal {
		root = t.Dialog{ID: "form-probe-modal", Visible: true, Title: "Form probe in Dialog", Content: root, Style: t.Style{Width: t.Percent(95)}}
	}
	if err := t.Run(root); err != nil {
		log.Fatal(err)
	}
}
