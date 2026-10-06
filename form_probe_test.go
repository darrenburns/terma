package terma

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

// This fixture exercises only public field/form state and the rendered event route.
type formProbePair struct {
	left, right           *TextInputState
	leftField, rightField *FieldState
	leftForm, rightForm   *FormState
	leftCalls, rightCalls int
	modal                 bool
	scroll                *ScrollState
}

func newFormProbePair(modal bool) *formProbePair {
	p := &formProbePair{left: NewTextInputState("left"), right: NewTextInputState("right"), modal: modal, scroll: NewScrollState()}
	p.leftField = NewTextInputField("probe-left", p.left, Required("Left required"))
	p.rightField = NewTextInputField("probe-right", p.right, Required("Right required"))
	p.leftForm = NewFormState(p.leftField)
	p.rightForm = NewFormState(p.rightField)
	return p
}
func (p *formProbePair) Build(BuildContext) Widget {
	left := Form{State: p.leftForm, OnSubmit: func() { p.leftCalls++ }, Child: Field{State: p.leftField, Label: "Left", Child: TextInput{State: p.left}}}
	right := Form{State: p.rightForm, OnSubmit: func() { p.rightCalls++ }, Child: Field{State: p.rightField, Label: "Right", Child: TextInput{State: p.right}}}
	body := Widget(FocusTrap{ID: "probe-trap", Active: true, Child: Scrollable{State: p.scroll, Style: Style{Height: Cells(9), Width: Cells(36)}, Child: Column{Spacing: 1, Children: []Widget{left, right}}}})
	if p.modal {
		return Dialog{ID: "probe-dialog", Visible: true, Content: body, Style: Style{Width: Cells(42)}}
	}
	return body
}
func TestFormProbeSiblingSubmissionRouting(t *testing.T) {
	for _, modal := range []bool{false, true} {
		name := "trap-scroll"
		if modal {
			name = "dialog-trap-scroll"
		}
		t.Run(name, func(t *testing.T) {
			seq := newReactivitySequence(t, 60, 20, func() *formProbePair { return newFormProbePair(modal) })
			seq.frame("initial", nil)
			for _, id := range []string{"probe-right", "probe-left"} {
				seq.focus(id)
				seq.frame("focus "+id, nil)
				for _, side := range []*reactivitySurface[*formProbePair]{seq.actual, seq.expected} {
					dispatchKey(side.renderer, side.focus, side.root, makeKeyEvent(uv.KeyEnter, 0))
				}
				seq.frame("submit "+id, nil)
			}
			require.Equal(t, 1, seq.actual.root.leftCalls)
			require.Equal(t, 1, seq.actual.root.rightCalls)
		})
	}
}

// A small independent state-machine oracle uses classified literal values.
// It drives the public API only; no signal metadata or implementation fields are read.
func formProbeModel(t *testing.T, operations []byte) {
	t.Helper()
	previousFocus := pendingFocusID
	t.Cleanup(func() { pendingFocusID = previousFocus })
	values := []string{"", " \t", "Alice", "Bob", "日本語", "x\ny"}
	nonempty := []bool{false, false, true, true, true, true}
	inputs := [2]*TextInputState{NewTextInputState(""), NewTextInputState("")}
	fields := [2]*FieldState{NewTextInputField("model-a", inputs[0], Required("required")), nil}
	fields[1] = NewTextInputField("model-b", inputs[1], Required("required"), func(s string) string {
		if s != "" && s != strings.Join(inputs[0].Content.Get(), "") {
			return "mismatch"
		}
		return ""
	})
	state := NewFormState(fields[0], fields[1])
	current, baseline := [2]int{}, [2]int{}
	enabled := [2]bool{true, true}
	revealed, touched := [2]bool{}, [2]bool{}
	registered := [2]bool{true, true}
	callbacks := 0
	form := Form{State: state}
	form.OnSubmit = func() { callbacks++; require.False(t, form.Submit(), "callback reentry must be suppressed") }
	errors := func(i int) []string {
		var result []string
		if !nonempty[current[i]] {
			result = append(result, "required")
		}
		if i == 1 && current[i] != 0 && current[i] != current[0] {
			result = append(result, "mismatch")
		}
		return result
	}
	valid := func() bool {
		for i := range fields {
			if registered[i] && enabled[i] && len(errors(i)) > 0 {
				return false
			}
		}
		return true
	}
	reset := func(i int) { current[i] = baseline[i]; revealed[i] = false; touched[i] = false }
	accept := func(i int) { baseline[i] = current[i]; revealed[i] = false; touched[i] = false }
	for step, b := range operations {
		i := int(b>>4) % 2
		value := int(b>>3) % len(values)
		switch b % 12 {
		case 0:
			current[i] = value
			inputs[i].SetText(values[value])
		case 1:
			fields[i].Touch()
			if enabled[i] {
				revealed[i] = true
				touched[i] = true
			}
		case 2:
			fields[i].Validate()
			if enabled[i] {
				revealed[i] = true
			}
		case 3:
			enabled[i] = !enabled[i]
			fields[i].SetEnabled(enabled[i])
		case 4:
			fields[i].Reset()
			reset(i)
		case 5:
			fields[i].Accept()
			accept(i)
		case 6:
			registered[i] = !registered[i]
			var next []*FieldState
			for j := range fields {
				if registered[j] {
					next = append(next, nil, fields[j], fields[j])
				}
			}
			state.SetFields(next...)
		case 7:
			state.Reset()
			for j := range fields {
				if registered[j] {
					reset(j)
				}
			}
		case 8:
			state.Accept()
			for j := range fields {
				if registered[j] {
					accept(j)
				}
			}
		case 9:
			expected := valid()
			before := callbacks
			require.Equal(t, expected, form.Submit(), "step %d", step)
			count := before
			if expected {
				count++
			}
			require.Equal(t, count, callbacks)
			for j := range fields {
				if registered[j] && enabled[j] {
					revealed[j] = true
				}
			}
		case 10:
			before := state.Fields()
			duplicate := NewTextInputField("model-a", NewTextInputState(""))
			require.Panics(t, func() { state.SetFields(fields[0], duplicate) })
			require.Equal(t, before, state.Fields(), "registry replacement is atomic")
		case 11:
			require.Equal(t, valid(), state.Validate())
			for j := range fields {
				if registered[j] && enabled[j] {
					revealed[j] = true
				}
			}
		}
		require.Equal(t, valid(), state.Valid(), "step %d", step)
		dirty := false
		for j := range fields {
			require.Equal(t, values[current[j]], inputs[j].GetText(), "step %d field %d", step, j)
			require.Equal(t, enabled[j], fields[j].Enabled())
			require.Equal(t, touched[j], fields[j].Touched())
			require.Equal(t, current[j] != baseline[j], fields[j].Dirty())
			require.Equal(t, !enabled[j] || len(errors(j)) == 0, fields[j].Valid())
			var want []string
			if enabled[j] && revealed[j] {
				want = errors(j)
			}
			require.Equal(t, want, fields[j].Errors(), "step %d field %d", step, j)
			dirty = dirty || (registered[j] && enabled[j] && current[j] != baseline[j])
		}
		require.Equal(t, dirty, state.Dirty(), "step %d", step)
	}
}
func TestFormProbeBoundedStateModel(t *testing.T) {
	for seed := int64(0); seed < 32; seed++ {
		t.Run(fmt.Sprint(seed), func(t *testing.T) {
			random := rand.New(rand.NewSource(seed))
			ops := make([]byte, 128)
			_, _ = random.Read(ops)
			formProbeModel(t, ops)
		})
	}
}
func FuzzFormProbeStateModel(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	f.Add([]byte{16, 24, 40, 52, 63, 91, 201, 255})
	f.Fuzz(func(t *testing.T, ops []byte) {
		if len(ops) > 128 {
			ops = ops[:128]
		}
		formProbeModel(t, ops)
	})
}

func TestFormProbeInvalidOffscreenFieldIsRevealed(t *testing.T) {
	first, last := NewTextInputState("ready"), NewTextInputState("")
	firstField := NewTextInputField("scroll-first", first, Required("First required"))
	lastField := NewTextInputField("scroll-last", last, Required("Last required"))
	scroll := NewScrollState()
	form := Form{State: NewFormState(firstField, lastField), Child: Column{Children: []Widget{
		Field{State: firstField, Child: TextInput{State: first}},
		Field{State: lastField, Child: TextInput{State: last}},
	}}}
	root := Scrollable{State: scroll, Style: Style{Width: Cells(40), Height: Cells(1)}, Child: form}
	p := NewPilot(t, root, 50, 24)
	require.Equal(t, "scroll-first", p.FocusedID())
	p.Press("ctrl+s")
	require.Equal(t, "scroll-last", p.FocusedID())
	input := p.session.renderer.WidgetByID("scroll-last")
	require.NotNil(t, input)
	require.Greater(t, input.Visible.Height, 0, "the first invalid input should be visible after submission focuses it")

	// Manual scrolling stays free until another submit requests the invalid input.
	scroll.SetOffset(0)
	require.Equal(t, "scroll-last", p.FocusedID())
	_, drawn := p.Bounds("scroll-last")
	require.False(t, drawn)
	p.Press("ctrl+s")
	_, drawn = p.Bounds("scroll-last")
	require.True(t, drawn, "retry reveals the already-focused invalid input")
}

func TestFormProbeInvalidRevealAfterErrorGrowth(t *testing.T) {
	first, last := NewTextInputState(""), NewTextInputState("")
	firstField := NewTextInputField("scroll-first", first, Required("First required"))
	lastField := NewTextInputField("scroll-last", last, Required("Last required"))
	scroll := NewScrollState()
	form := Form{State: NewFormState(lastField, firstField), Child: Column{Children: []Widget{
		Field{State: firstField, Child: TextInput{State: first}},
		Field{State: lastField, Child: TextInput{State: last}},
	}}}
	root := Scrollable{State: scroll, Style: Style{Width: Cells(40), Height: Cells(1)}, Child: form}
	p := NewPilot(t, root, 50, 24)
	require.Equal(t, "scroll-first", p.FocusedID())
	p.Press("ctrl+s")
	require.Equal(t, "scroll-last", p.FocusedID())
	input := p.session.renderer.WidgetByID("scroll-last")
	require.NotNil(t, input)
	require.Greater(t, input.Visible.Height, 0, "the first invalid input should be visible after submission focuses it")

}

func TestFormProbeNestedRevealAndUnmount(t *testing.T) {
	value := NewTextInputState("")
	field := NewTextInputField("nested-invalid", value, Required("Detail required"))
	state := NewFormState(field)
	inner, outer := NewScrollState(), NewScrollState()
	form := Form{State: state}
	content := Column{Style: Style{Padding: EdgeInsetsAll(1)}, Children: []Widget{Text{Content: "one\ntwo\nthree\nfour"}, Field{Label: "Detail", State: field, Child: TextInput{State: value}}}}
	form.Child = Column{Children: []Widget{Button{ID: "nested-submit", Label: "Submit", OnPress: func() { form.Submit() }}, Scrollable{State: inner, Style: Style{Width: Cells(32), Height: Cells(3)}, Child: content}}}
	dialogContent := NewAnySignal[Widget](FocusTrap{ID: "nested-trap", Active: true, Child: Scrollable{State: outer, Style: Style{Width: Cells(36), Height: Cells(4)}, Child: Column{Children: []Widget{Text{Content: "above\nabove"}, form}}}})
	root := pilotBuilder(func(BuildContext) Widget {
		return Dialog{ID: "probe-nested-dialog", Visible: true, Style: Style{Width: Cells(42)}, Content: dialogContent.Get()}
	})
	p := NewPilot(t, root, 50, 24)
	p.AssertSnapshot("before", "Padded nested scroll areas before an invalid submission, with the auto-focused Submit button drawn focused")
	p.Press("enter")
	require.Equal(t, "nested-invalid", p.FocusedID())
	input := p.session.renderer.WidgetByID("nested-invalid")
	require.NotNil(t, input)
	require.Greater(t, input.Visible.Height, 0)
	p.AssertSnapshot("invalid", "First-invalid input is visible through both scroll viewports after submission")
	// A removed invalid input still blocks submit, without scrolling its old tree.
	dialogContent.Set(Button{ID: "replacement", Label: "Replacement"})
	p.settle()
	oldInner, oldOuter := inner.GetOffset(), outer.GetOffset()
	require.False(t, form.Submit())
	p.settle()
	require.Equal(t, oldInner, inner.GetOffset())
	require.Equal(t, oldOuter, outer.GetOffset())
}

type formProbeScrolled struct {
	fixture *formFixture
	scroll  *ScrollState
}

func (p *formProbeScrolled) Build(BuildContext) Widget {
	return Scrollable{State: p.scroll, Style: Style{Width: Cells(44), Height: Cells(5)}, Child: p.fixture}
}
func TestFormProbeRetainedRevealAndManualScroll(t *testing.T) {
	seq := newReactivitySequence(t, 44, 8, func() *formProbeScrolled {
		f := newFormFixture()
		f.state.SetFields(f.confirmField, f.nameField, f.areaField)
		return &formProbeScrolled{fixture: f, scroll: NewScrollState()}
	})
	seq.frame("initial", nil)
	press := func() {
		for _, side := range []*reactivitySurface[*formProbeScrolled]{seq.actual, seq.expected} {
			dispatchKey(side.renderer, side.focus, side.root, makeKeyEvent('s', uv.ModCtrl))
			if pendingFocusID != "" {
				side.focus.FocusByID(pendingFocusID)
				side.focused.Set(side.focus.Focused())
				pendingFocusID = ""
			}
		}
	}
	press()
	seq.frame("error reflow then reveal", nil)
	require.Equal(t, "confirm", seq.actual.focus.FocusedID())
	require.NotNil(t, seq.actual.renderer.WidgetByID("confirm"))
	assertBufferSnapshot(t, "FormProbe_error_reflow_reveal", seq.actual.buffer, 44, 8, DefaultSVGOptions(), "Registry-first Confirm remains visibly focused after Name error growth")
	seq.frame("manual scroll stays free", func(p *formProbeScrolled) { p.scroll.SetOffset(0) })
	require.Nil(t, seq.actual.renderer.WidgetByID("confirm"))
	assertBufferSnapshot(t, "FormProbe_manual_scroll", seq.actual.buffer, 44, 8, DefaultSVGOptions(), "Manual scrolling can hide the focused field until another explicit submit")
	press()
	seq.frame("retry while same field focused", nil)
	require.NotNil(t, seq.actual.renderer.WidgetByID("confirm"))
	assertBufferSnapshot(t, "FormProbe_retry_reveal", seq.actual.buffer, 44, 8, DefaultSVGOptions(), "Another failed submit reveals the same focused field again")
}
