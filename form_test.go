package terma

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldRequiredUnicodeWhitespace(t *testing.T) {
	validate := Required("Provide text")
	for _, value := range []string{"", " ", "\t\n", "\u2003\u3000"} {
		assert.Equal(t, "Provide text", validate(value))
	}
	for _, value := range []string{"日本語", "é", "x\ny", " x "} {
		assert.Empty(t, validate(value))
	}
	assert.Equal(t, "Required", Required("")(""))
}

func TestFieldTimingAndProgrammaticChanges(t *testing.T) {
	input := NewTextInputState("")
	field := NewTextInputField("name", input, Required("Required name"))
	assert.False(t, field.Valid())
	assert.Empty(t, field.Errors(), "validity reads do not reveal")
	assert.False(t, field.Touched())
	assert.False(t, field.Dirty())
	input.SetText(" ")
	assert.True(t, field.Dirty())
	assert.Empty(t, field.Errors(), "untouched edits stay quiet")
	field.Touch()
	assert.True(t, field.Touched())
	assert.Equal(t, []string{"Required name"}, field.Errors())
	input.SetText("日本語")
	assert.Empty(t, field.Errors(), "programmatic updates immediately correct errors")
	assert.True(t, field.Valid())
	input.SetText("")
	assert.False(t, field.Dirty(), "exact revert to baseline is pristine")
	assert.Equal(t, []string{"Required name"}, field.Errors())
	field.Reset()
	assert.False(t, field.Touched())
	assert.Empty(t, field.Errors())
	assert.False(t, field.Validate())
	assert.False(t, field.Touched(), "submit validation doesn't claim a blur")
	assert.Equal(t, []string{"Required name"}, field.Errors())
}

func TestFieldAllValidatorsRunInOrder(t *testing.T) {
	value := NewSignal(7)
	order := []int{}
	validators := []Validator[int]{
		func(n int) string { assert.Equal(t, 7, n); order = append(order, 1); return "first" },
		nil,
		func(int) string { order = append(order, 2); return " \t" },
		func(int) string { order = append(order, 3); return "third" },
		func(int) string { order = append(order, 4); return "first" },
	}
	field := NewFieldState("number", FieldBinding[int]{Read: value.Get, Write: value.Set}, validators...)
	validators[0] = nil // constructor owns its validator list
	field.Touch()
	assert.Equal(t, []string{"first", "third", "first"}, field.Errors())
	assert.Equal(t, []int{1, 2, 3, 4}, order)
}

func TestFieldResetAndAccept(t *testing.T) {
	input := NewTextInputState("initial")
	field := NewTextInputField("value", input)
	input.SetText("edited")
	field.Touch()
	field.Reset()
	assert.Equal(t, "initial", input.GetText())
	assert.False(t, field.Dirty())
	assert.False(t, field.Touched())
	input.SetText("loaded")
	field.Accept()
	assert.Equal(t, "loaded", input.GetText())
	assert.False(t, field.Dirty())
	input.SetText("later")
	field.Reset()
	assert.Equal(t, "loaded", input.GetText(), "Reset uses the most recently accepted value")
	area := NewTextAreaState("one\ntwo")
	areaField := NewTextAreaField("notes", area)
	area.SetText("three")
	areaField.Reset()
	assert.Equal(t, "one\ntwo", area.GetText())
}

func TestFieldConfigurationErrors(t *testing.T) {
	binding := FieldBinding[int]{Read: func() int { return 1 }, Write: func(int) {}}
	assert.Panics(t, func() { NewFieldState("", binding) })
	assert.Panics(t, func() { NewFieldState("x", FieldBinding[int]{}) })
	assert.Panics(t, func() { NewTextInputField("x", nil) })
	assert.Panics(t, func() { NewTextAreaField("x", nil) })
	field := NewTextInputField("x", NewTextInputState(""))
	assert.Panics(t, func() { Field{State: field, Child: TextInput{State: NewTextInputState("")}}.Build(BuildContext{}) })
}

func TestFormExplicitRegistryAndDuplicates(t *testing.T) {
	a := NewTextInputField("a", NewTextInputState(""), Required("A required"))
	b := NewTextInputField("b", NewTextInputState("ok"))
	form := NewFormState(nil, a, a, b, nil)
	assert.Equal(t, []*FieldState{a, b}, form.Fields())
	copy := form.Fields()
	copy[0] = nil
	assert.Same(t, a, form.Fields()[0])
	duplicate := NewTextInputField("a", NewTextInputState(""))
	assert.Panics(t, func() { form.SetFields(b, a, duplicate) })
	assert.Equal(t, []*FieldState{a, b}, form.Fields(), "invalid replacement must be atomic")
	a.Touch()
	form.SetFields(b)
	assert.True(t, form.Valid())
	assert.True(t, a.Touched(), "unregistration doesn't erase independent state")
	form.SetFields(nil)
	assert.Empty(t, form.Fields())
	assert.True(t, form.Valid())
}

func TestFormDisabledParticipation(t *testing.T) {
	input := NewTextInputState("")
	field := NewTextInputField("name", input, Required("Name required"))
	form := NewFormState(field)
	field.Touch()
	input.SetText(" ")
	assert.False(t, form.Valid())
	assert.True(t, form.Dirty())
	field.SetEnabled(false)
	assert.True(t, form.Valid())
	assert.False(t, form.Dirty())
	assert.Empty(t, field.Errors())
	assert.True(t, field.Touched())
	assert.True(t, field.Dirty(), "field metadata survives participation changes")
	field.SetEnabled(true)
	assert.False(t, form.Valid())
	assert.Equal(t, []string{"Name required"}, field.Errors())
	field.SetEnabled(false)
	form.Reset()
	assert.Empty(t, input.GetText(), "disabled fields also reset")
	assert.False(t, field.Touched())
	input.SetText("accepted while disabled")
	form.Accept()
	input.SetText("other")
	form.Reset()
	assert.Equal(t, "accepted while disabled", input.GetText())
}

func TestFormSubmitValidatesAllAndFocusesFirst(t *testing.T) {
	previous := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previous })
	pendingFocusID = ""
	aInput, bInput := NewTextInputState(""), NewTextInputState("")
	a := NewTextInputField("a", aInput, Required("A required"))
	b := NewTextInputField("b", bInput, Required("B required"))
	state := NewFormState(b, a)
	calls := 0
	form := Form{State: state, OnSubmit: func() {
		calls++
		assert.True(t, state.Valid())
		assert.True(t, a.revealed.Peek())
		assert.True(t, b.revealed.Peek())
	}}
	assert.False(t, form.Submit())
	assert.Equal(t, "b", pendingFocusID, "registry order owns focus")
	assert.Equal(t, []string{"A required"}, a.Errors())
	assert.Equal(t, []string{"B required"}, b.Errors())
	assert.Zero(t, calls)
	bInput.SetText("good")
	assert.False(t, form.Submit())
	assert.Equal(t, "a", pendingFocusID)
	aInput.SetText("good")
	assert.True(t, form.Submit())
	assert.Equal(t, 1, calls)
	assert.True(t, state.Dirty(), "success does not silently accept")
	assert.True(t, form.Submit())
	assert.Equal(t, 2, calls, "separate attempts each submit")
}

func TestFormNilEmptyAndReentrantSubmission(t *testing.T) {
	assert.False(t, (Form{}).Submit())
	assert.Empty(t, (Form{}).Keybinds())
	assert.NotPanics(t, func() { RenderToBuffer(Form{}, 20, 5) })
	assert.NotPanics(t, func() { RenderToBuffer(Field{Label: "Plain", Help: "Help"}, 20, 5) })
	state := NewFormState()
	calls := 0
	form := Form{State: state}
	form.OnSubmit = func() { calls++; assert.False(t, form.Submit(), "callback reentry is suppressed") }
	assert.True(t, form.Submit())
	assert.Equal(t, 1, calls)
	assert.False(t, state.submitting)
}

func TestFieldAdaptersPreserveHooks(t *testing.T) {
	input := NewTextInputState("")
	field := NewTextInputField("bound", input)
	var events []string
	widget := TextInput{ID: "overridden", State: input, OnChange: func(string) { events = append(events, "change") }, Blur: func() { assert.True(t, field.Touched()); events = append(events, "blur") }, OnSubmit: func(string) { events = append(events, "submit") }, OnPaste: func(string) bool { events = append(events, "paste"); return false }}
	adapted := Field{State: field}.adaptChild(&widget).(fieldTextInput)
	assert.Equal(t, "bound", adapted.WidgetID())
	assert.Equal(t, "overridden", widget.ID, "adaptation does not mutate caller's widget")
	adapted.OnKey(makeCharEvent('x'))
	adapted.OnBlur()
	require.True(t, matchKeybind(makeKeyEvent(uv.KeyEnter, 0), adapted.Keybinds()))
	adapted.HandlePaste("y")
	assert.Equal(t, []string{"change", "blur", "submit", "paste", "change"}, events)
	field.Reset()
	assert.Len(t, events, 5, "programmatic reset does not invoke input hooks")
}

func TestFieldTextAreaAndCheckboxAdapters(t *testing.T) {
	area := NewTextAreaState("")
	field := NewTextAreaField("area", area)
	blurs := 0
	adapted := Field{State: field}.adaptChild(&TextArea{State: area, Blur: func() { blurs++ }}).(TextArea)
	assert.Equal(t, "area", adapted.ID)
	adapted.OnBlur()
	assert.True(t, field.Touched())
	assert.Equal(t, 1, blurs)
	check := NewCheckboxState(false)
	checkField := NewFieldState("check", FieldBinding[bool]{Read: check.Checked.Get, Write: check.SetChecked}, func(b bool) string {
		if !b {
			return "Check required"
		}
		return ""
	})
	checkInput := Field{State: checkField}.adaptChild(&Checkbox{State: check}).(fieldCheckbox)
	checkInput.OnBlur()
	assert.Equal(t, []string{"Check required"}, checkField.Errors())
	matchKeybind(makeKeyEvent(uv.KeyEnter, 0), checkInput.Keybinds())
	assert.True(t, check.IsChecked())
	assert.Empty(t, checkField.Errors())
}

type formFixture struct {
	name, confirm                      *TextInputState
	area                               *TextAreaState
	nameField, confirmField, areaField *FieldState
	state                              *FormState
	submits                            int
	onNameSubmit                       func(string)
	extra                              []Keybind
}

func newFormFixture() *formFixture {
	f := &formFixture{name: NewTextInputState(""), confirm: NewTextInputState(""), area: NewTextAreaState("")}
	f.nameField = NewTextInputField("name", f.name, Required("Enter a name"))
	f.confirmField = NewTextInputField("confirm", f.confirm, Required("Confirm name"), func(value string) string {
		if value != "" && value != strings.Join(f.name.Content.Get(), "") {
			return "Must match Name"
		}
		return ""
	})
	f.areaField = NewTextAreaField("area", f.area)
	f.state = NewFormState(f.nameField, f.confirmField, f.areaField)
	return f
}

func (f *formFixture) Build(ctx BuildContext) Widget {
	form := Form{State: f.state, OnSubmit: func() { f.submits++ }}
	form.Child = Column{Spacing: 1, Children: []Widget{
		Field{Label: "Name", Help: "Public name", State: f.nameField, Child: TextInput{State: f.name, OnSubmit: f.onNameSubmit, ExtraKeybinds: f.extra, Style: Style{Width: Flex(1), BackgroundColor: ctx.Theme().Surface}}},
		Field{Label: "Confirm", State: f.confirmField, Child: TextInput{State: f.confirm, Style: Style{Width: Flex(1), BackgroundColor: ctx.Theme().Surface}}},
		Field{Label: "Notes", State: f.areaField, Child: TextArea{State: f.area, Style: Style{Width: Flex(1), Height: Cells(2), BackgroundColor: ctx.Theme().Surface}}},
		Button{ID: "save", Label: "Submit", OnPress: func() { form.Submit() }},
	}}
	return form
}

// formInputHarness drives the actual renderer, focus manager and key dispatcher.
func formInputHarness(t *testing.T, root Widget) (*mouseRouter, *Renderer, func(KeyEvent)) {
	t.Helper()
	previous := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previous })
	pendingFocusID = ""
	router, renderer := renderForMouse(root, 50, 24)
	return router, renderer, func(event KeyEvent) {
		dispatchKey(renderer, router.focusManager, root, event)
		requested := pendingFocusID
		pendingFocusID = ""
		router.focusManager.SetFocusables(renderer.Render(root))
		if requested != "" {
			router.focusManager.FocusByID(requested)
		}
		renderer.Render(root)
	}
}

func TestFormKeyboardRoutingAndFocus(t *testing.T) {
	f := newFormFixture()
	router, _, key := formInputHarness(t, f)
	assert.Equal(t, "name", router.focusManager.FocusedID())
	key(makeKeyEvent(uv.KeyEnter, 0))
	assert.Zero(t, f.submits)
	assert.Equal(t, "name", router.focusManager.FocusedID())
	assert.NotEmpty(t, f.nameField.Errors())
	assert.NotEmpty(t, f.confirmField.Errors())
	for _, r := range "日本語" {
		key(makeCharEvent(r))
	}
	key(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, "confirm", router.focusManager.FocusedID())
	assert.True(t, f.nameField.Touched())
	for _, r := range "日本語" {
		key(makeCharEvent(r))
	}
	key(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, 1, f.submits)
	key(makeKeyEvent(uv.KeyTab, 0))
	assert.Equal(t, "area", router.focusManager.FocusedID())
	key(makeCharEvent('x'))
	key(makeKeyEvent(uv.KeyEnter, 0))
	key(makeCharEvent('y'))
	assert.Equal(t, "x\ny", f.area.GetText())
	assert.Equal(t, 1, f.submits, "multiline Enter never submits")
	key(makeKeyEvent('s', uv.ModCtrl))
	assert.Equal(t, 2, f.submits)
}

func TestFormExplicitEnterHooksTakePrecedence(t *testing.T) {
	for _, kind := range []string{"OnSubmit", "ExtraKeybind"} {
		t.Run(kind, func(t *testing.T) {
			f := newFormFixture()
			f.name.SetText("ok")
			f.confirm.SetText("ok")
			calls := 0
			if kind == "OnSubmit" {
				f.onNameSubmit = func(string) { calls++ }
			} else {
				f.extra = []Keybind{{Key: "enter", Action: func() { calls++ }}}
			}
			_, _, key := formInputHarness(t, f)
			key(makeKeyEvent(uv.KeyEnter, 0))
			assert.Equal(t, 1, calls)
			assert.Zero(t, f.submits)
		})
	}
}

func TestFormMouseSubmitFocusAndBlur(t *testing.T) {
	f := newFormFixture()
	router, renderer, _ := formInputHarness(t, f)
	button := renderer.WidgetByID("save")
	require.NotNil(t, button)
	x, y := button.Bounds.X, button.Bounds.Y
	router.press(uv.MouseClickEvent{X: x, Y: y, Button: uv.MouseLeft}, .5, .5, time.Now())
	router.release(uv.MouseReleaseEvent{X: x, Y: y, Button: uv.MouseLeft}, .5, .5)
	assert.True(t, f.nameField.Touched(), "mouse focus change validates blur")
	assert.Equal(t, "name", pendingFocusID, "failed mouse submit requests first invalid")
	assert.Zero(t, f.submits)
	f.name.SetText("ok")
	f.confirm.SetText("ok")
	renderer.Render(f)
	button = renderer.WidgetByID("save")
	x, y = button.Bounds.X, button.Bounds.Y
	router.press(uv.MouseClickEvent{X: x, Y: y, Button: uv.MouseLeft}, .5, .5, time.Now())
	router.release(uv.MouseReleaseEvent{X: x, Y: y, Button: uv.MouseLeft}, .5, .5)
	assert.Equal(t, 1, f.submits)
}

func TestFormRegisteredUnmountedFieldStillBlocks(t *testing.T) {
	previous := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previous })
	pendingFocusID = ""
	field := NewTextInputField("missing", NewTextInputState(""), Required("Missing required"))
	form := Form{State: NewFormState(field), Child: Button{ID: "visible", Label: "Visible"}, OnSubmit: func() { t.Fatal("invalid unmounted field submitted") }}
	router, _ := renderForMouse(form, 30, 6)
	assert.False(t, form.Submit())
	router.focusManager.FocusByID(pendingFocusID)
	assert.Equal(t, "visible", router.focusManager.FocusedID(), "missing focus request is harmless")
	assert.Equal(t, []string{"Missing required"}, field.Errors())
}

func TestFieldBuildPurityAndNilChildren(t *testing.T) {
	f := newFormFixture()
	f.nameField.Touch()
	beforeTouched, beforeRevealed, beforeText := f.nameField.touched.Peek(), f.nameField.revealed.Peek(), f.name.GetText()
	RenderToBuffer(f, 40, 20)
	assert.Equal(t, beforeTouched, f.nameField.touched.Peek())
	assert.Equal(t, beforeRevealed, f.nameField.revealed.Peek())
	assert.Equal(t, beforeText, f.name.GetText())
	var input *TextInput
	var area *TextArea
	var checkbox *Checkbox
	for _, child := range []Widget{nil, input, area, checkbox} {
		assert.NotPanics(t, func() { RenderToBuffer(Field{State: f.nameField, Child: child}, 20, 4) })
	}
}

func TestFormReactiveCrossFieldAndProgrammaticUpdates(t *testing.T) {
	seq := newReactivitySequence(t, 44, 22, newFormFixture)
	seq.frame("pristine", nil)
	seq.frame("untouched edit", func(f *formFixture) { f.name.SetText("Alice") })
	seq.frame("reveal all", func(f *formFixture) { f.state.Validate() })
	seq.frame("matching", func(f *formFixture) { f.confirm.SetText("Alice") })
	seq.frame("cross-field mismatch", func(f *formFixture) { f.name.SetText("Bob") })
	assert.Equal(t, []string{"Must match Name"}, seq.actual.root.confirmField.Errors())
	seq.frame("disable invalid", func(f *formFixture) { f.confirmField.SetEnabled(false) })
	seq.frame("reenable errors", func(f *formFixture) { f.confirmField.SetEnabled(true) })
	seq.frame("accept", func(f *formFixture) { f.state.Accept() })
	seq.frame("programmatic invalid", func(f *formFixture) { f.name.SetText(""); f.nameField.Touch() })
	seq.frame("reset", func(f *formFixture) { f.state.Reset() })
}

func TestSnapshot_Form_States(t *testing.T) {
	for _, scenario := range []string{"pristine", "blur-errors", "submit-errors", "corrected", "cross-field", "disabled", "accepted", "multiline", "multiple-errors"} {
		t.Run(scenario, func(t *testing.T) {
			f := newFormFixture()
			switch scenario {
			case "blur-errors":
				f.nameField.Touch()
			case "submit-errors":
				f.state.Validate()
			case "corrected":
				f.state.Validate()
				f.name.SetText("日本語")
				f.confirm.SetText("日本語")
			case "cross-field":
				f.name.SetText("Alice")
				f.confirm.SetText("Bob")
				f.state.Validate()
			case "disabled":
				f.state.Validate()
				f.confirmField.SetEnabled(false)
			case "accepted":
				f.name.SetText("Accepted")
				f.confirm.SetText("Accepted")
				f.state.Accept()
			case "multiline":
				f.area.SetText("First line\n日本語 second line")
			case "multiple-errors":
				f.nameField = NewTextInputField("name", f.name, Required("Name is required"), func(string) string { return "Second independent constraint" })
				f.state.SetFields(f.nameField, f.confirmField)
				f.state.Validate()
			}
			AssertSnapshot(t, f, 44, 22, "Form: "+scenario)
		})
	}
}

func TestSnapshot_Form_Narrow(t *testing.T) {
	for _, width := range []int{18, 8} {
		t.Run(fmt.Sprint(width), func(t *testing.T) {
			f := newFormFixture()
			f.name.SetText("Alice")
			f.confirm.SetText("Bob")
			f.state.Validate()
			AssertSnapshot(t, f, width, 26, "Narrow form wraps labels, help and error messages")
		})
	}
}

func TestFieldDisabledAndCheckboxFocusRouting(t *testing.T) {
	checked := NewCheckboxState(false)
	state := NewFieldState("consent", FieldBinding[bool]{Read: checked.Checked.Get, Write: checked.SetChecked}, func(value bool) string {
		if !value {
			return "Consent required"
		}
		return ""
	})
	root := Column{Children: []Widget{
		Field{State: state, Child: &Checkbox{State: checked, Label: "Accept"}},
		Button{ID: "next", Label: "Next"},
	}}
	router, _, key := formInputHarness(t, root)
	assert.Equal(t, "consent", router.focusManager.FocusedID())
	key(makeKeyEvent(uv.KeyTab, 0))
	assert.True(t, state.Touched(), "real focus traversal reaches the checkbox adapter's blur handler")
	assert.Equal(t, []string{"Consent required"}, state.Errors())
	state.SetEnabled(false)
	router, _, _ = formInputHarness(t, root)
	assert.Equal(t, "next", router.focusManager.FocusedID(), "disabled field subtree leaves the focus order")
}

func TestFormDialogSubmissionAndFocus(t *testing.T) {
	fixture := newFormFixture()
	root := Dialog{ID: "form-test-dialog", Visible: true, Title: "Edit profile", Content: fixture, Style: Style{Width: Cells(46)}}
	router, _, key := formInputHarness(t, root)
	assert.Equal(t, "name", router.focusManager.FocusedID())
	key(makeKeyEvent('s', uv.ModCtrl))
	assert.Zero(t, fixture.submits)
	assert.Equal(t, "name", router.focusManager.FocusedID(), "invalid modal submit focuses the first field")
	assert.Equal(t, []string{"Enter a name"}, fixture.nameField.Errors())
	fixture.name.SetText("Alice")
	key(makeKeyEvent('s', uv.ModCtrl))
	assert.Equal(t, "confirm", router.focusManager.FocusedID())
	for _, r := range "Alice" {
		key(makeCharEvent(r))
	}
	key(makeKeyEvent(uv.KeyEnter, 0))
	assert.Equal(t, 1, fixture.submits, "Enter submits the nested form exactly once")
	key(makeKeyEvent(uv.KeyTab, 0))
	assert.Equal(t, "area", router.focusManager.FocusedID())
	key(makeCharEvent('x'))
	key(makeKeyEvent(uv.KeyEnter, 0))
	key(makeCharEvent('y'))
	assert.Equal(t, "x\ny", fixture.area.GetText())
	assert.Equal(t, 1, fixture.submits, "multiline Enter remains local inside a modal")
	key(makeKeyEvent('s', uv.ModCtrl))
	assert.Equal(t, 2, fixture.submits)
}
