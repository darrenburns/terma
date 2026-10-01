package terma

import "strings"

// Validator checks a value and returns an error message, or an empty string.
// Validators must be pure and synchronous: reactive reads can run them repeatedly.
// Read cross-field dependencies with Get so their changes invalidate the view.
type Validator[T any] func(T) string

// FieldBinding connects field metadata to an existing value source. Read must
// read that source reactively; Write restores it on Reset. Both are required.
// Use stable, reflexive comparable values (for example strings, bools or IDs).
type FieldBinding[T comparable] struct {
	Read  func() T
	Write func(T)
}

// FieldState owns validation visibility and an accepted baseline, not the input
// value. Construct it once outside Build and explicitly register it with a form.
type FieldState struct {
	inputID       string
	revealRequest Signal[uint64]
	revealHandled uint64
	errors        func() []string
	dirty         func() bool
	reset         func()
	accept        func()
	source        any
	touched       Signal[bool]
	revealed      Signal[bool]
	enabled       Signal[bool]
}

// NewFieldState binds an existing value and ordered validators to an input ID.
// Empty IDs and incomplete bindings panic because they are configuration errors.
func NewFieldState[T comparable](inputID string, binding FieldBinding[T], validators ...Validator[T]) *FieldState {
	if strings.TrimSpace(inputID) == "" {
		panic("terma: field input ID must not be empty")
	}
	if binding.Read == nil || binding.Write == nil {
		panic("terma: field binding requires Read and Write")
	}
	baseline := NewSignal(binding.Read())
	checks := append([]Validator[T](nil), validators...)
	return &FieldState{
		inputID:       inputID,
		revealRequest: NewSignal(uint64(0)),
		errors: func() []string {
			value := binding.Read()
			var errors []string
			for _, validate := range checks {
				if validate != nil {
					if message := validate(value); strings.TrimSpace(message) != "" {
						errors = append(errors, message)
					}
				}
			}
			return errors
		},
		dirty:    func() bool { return binding.Read() != baseline.Get() },
		reset:    func() { binding.Write(baseline.Peek()) },
		accept:   func() { baseline.Set(binding.Read()) },
		touched:  NewSignal(false),
		revealed: NewSignal(false),
		enabled:  NewSignal(true),
	}
}

// NewTextInputField binds a field to an existing single-line input. The binding
// reads Content reactively, unlike TextInputState.GetText's non-reactive read.
func NewTextInputField(inputID string, input *TextInputState, validators ...Validator[string]) *FieldState {
	if input == nil {
		panic("terma: text input field requires input state")
	}
	field := NewFieldState(inputID, FieldBinding[string]{
		Read: func() string { return joinGraphemes(input.Content.Get()) }, Write: input.SetText,
	}, validators...)
	field.source = input
	return field
}

// NewTextAreaField binds a field to an existing multiline input.
func NewTextAreaField(inputID string, input *TextAreaState, validators ...Validator[string]) *FieldState {
	if input == nil {
		panic("terma: text area field requires input state")
	}
	field := NewFieldState(inputID, FieldBinding[string]{
		Read: func() string { return joinGraphemes(input.Content.Get()) }, Write: input.SetText,
	}, validators...)
	field.source = input
	return field
}

// Required rejects empty and Unicode-whitespace-only strings without modifying
// them. A blank message uses "Required". Other constraints belong in validators.
func Required(message string) Validator[string] {
	if strings.TrimSpace(message) == "" {
		message = "Required"
	}
	return func(value string) string {
		if strings.TrimSpace(value) == "" {
			return message
		}
		return ""
	}
}

// InputID returns the stable focus target assigned to adapted inputs.
func (s *FieldState) InputID() string { return s.inputID }

// Errors returns the current visible errors, in validator order. Until Touch or
// Validate reveals errors, or while disabled, it returns nil. It is reactive.
func (s *FieldState) Errors() []string {
	if !s.enabled.Get() || !s.revealed.Get() {
		return nil
	}
	return s.errors()
}

// Touched reports whether the field has blurred since construction/Reset/Accept.
func (s *FieldState) Touched() bool { return s.touched.Get() }

// Dirty compares the current value with the accepted baseline. It is reactive.
func (s *FieldState) Dirty() bool { return s.dirty() }

// Enabled reports whether the field participates in form validation.
func (s *FieldState) Enabled() bool { return s.enabled.Get() }

// SetEnabled controls validation participation and Field's disabled presentation.
// Re-enabling preserves the value, baseline, touched and error visibility state.
func (s *FieldState) SetEnabled(enabled bool) { s.enabled.Set(enabled) }

// Valid checks the current value without revealing errors. Disabled fields are
// valid for form purposes. It tracks reactive source and validator dependencies.
func (s *FieldState) Valid() bool { return !s.enabled.Get() || len(s.errors()) == 0 }

// Touch marks an enabled field as touched and reveals its current errors.
func (s *FieldState) Touch() {
	if s.enabled.Peek() {
		s.touched.Set(true)
		s.revealed.Set(true)
	}
}

// Validate reveals current errors without marking the field touched.
func (s *FieldState) Validate() bool {
	if s.enabled.Peek() {
		s.revealed.Set(true)
	}
	return s.Valid()
}

func (s *FieldState) clearPresentation() {
	s.touched.Set(false)
	s.revealed.Set(false)
}

// Reset restores the accepted baseline and clears touched/error visibility. It
// does not call input callbacks. Use in handlers or setup, never during Build.
func (s *FieldState) Reset() {
	s.reset()
	s.clearPresentation()
}

// Accept retains the current value as the new baseline and clears touched/error
// visibility. Call after loading or successfully saving an accepted value.
func (s *FieldState) Accept() {
	s.accept()
	s.clearPresentation()
}

// Field composes a label, input, help and visible errors. Direct TextInput and
// TextArea values/pointers and Checkbox pointers receive the state's input ID
// and blur tracking. Custom inputs should use InputID and call Touch on blur.
// A nil State is a presentation-only field.
type Field struct {
	Label string
	Help  string
	State *FieldState
	Child Widget
	Style Style
}

func (f Field) GetContentDimensions() (Dimension, Dimension) {
	dims := f.Style.GetDimensions()
	return dims.Width, dims.Height
}

func (f Field) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	children := make([]Widget, 0, 4)
	if f.Label != "" {
		children = append(children, Text{Content: f.Label, Wrap: WrapSoft, Style: Style{Bold: true, ForegroundColor: theme.Text}})
	}
	child := f.Child
	if f.State != nil {
		child = f.adaptChild(child)
		if !f.State.Enabled() && child != nil {
			child = DisabledWhen(true, child)
		}
	}
	inputIndex := -1
	if child != nil {
		inputIndex = len(children)
		children = append(children, child)
	}
	if f.Help != "" {
		children = append(children, Text{Content: f.Help, Wrap: WrapSoft, Style: Style{ForegroundColor: theme.TextMuted}})
	}
	if f.State != nil && !ctx.IsDisabled() {
		for _, message := range f.State.Errors() {
			children = append(children, Text{Content: "! " + message, Wrap: WrapSoft, Style: Style{ForegroundColor: theme.Error}})
		}
	}
	return fieldContainer{Column: Column{Style: f.Style, Children: children}, state: f.State, inputIndex: inputIndex}
}

func (f Field) checkSource(source any) {
	if f.State.source != nil && f.State.source != source {
		panic("terma: field child does not match its bound input state")
	}
}

func (f Field) adaptChild(child Widget) Widget {
	switch input := child.(type) {
	case *TextInput:
		if input == nil {
			return nil
		}
		return f.adaptChild(*input)
	case TextInput:
		f.checkSource(input.State)
		input.ID = f.State.inputID
		return fieldTextInput{TextInput: input, field: f.State}
	case *TextArea:
		if input == nil {
			return nil
		}
		return f.adaptChild(*input)
	case TextArea:
		f.checkSource(input.State)
		input.ID = f.State.inputID
		original := input.Blur
		input.Blur = func() {
			f.State.Touch()
			if original != nil {
				original()
			}
		}
		return input
	case *Checkbox:
		if input == nil {
			return nil
		}
		copy := *input
		copy.ID = f.State.inputID
		return fieldCheckbox{Checkbox: &copy, field: f.State}
	default:
		return child
	}
}

// fieldTextInput preserves the input's editing and hooks while allowing an
// unconfigured Enter to reach the enclosing form. A custom handler takes priority.
type fieldTextInput struct {
	TextInput
	field *FieldState
}

func (f fieldTextInput) Build(BuildContext) Widget { return f }

func (f fieldTextInput) Keybinds() []Keybind {
	if f.OnSubmit != nil {
		return f.TextInput.Keybinds()
	}
	input := f.TextInput
	input.ExtraKeybinds = nil
	keys := append([]Keybind(nil), f.ExtraKeybinds...)
	for _, key := range input.Keybinds() {
		if key.Key != "enter" {
			keys = append(keys, key)
		}
	}
	return keys
}

func (f fieldTextInput) OnBlur() {
	f.field.Touch()
	f.TextInput.OnBlur()
}

type fieldCheckbox struct {
	*Checkbox
	field *FieldState
}

func (f fieldCheckbox) OnBlur() { f.field.Touch() }

// fieldContainer schedules submit-time reveal after validation has reflowed
// labels/errors. Dispatch updates scroll state between frames, never mid-layout.
type fieldContainer struct {
	Column
	state      *FieldState
	inputIndex int
}

func (c fieldContainer) Build(BuildContext) Widget { return c }
func (c fieldContainer) ChildWidgets() []Widget    { return c.Children }
func (c fieldContainer) OnLayout(ctx BuildContext, metrics LayoutMetrics) {
	if c.state == nil {
		return
	}
	request := c.state.revealRequest.Get()
	if request == c.state.revealHandled || ctx.scrollToView == nil {
		return
	}
	if bounds, ok := metrics.ChildBounds(c.inputIndex); ok && bounds.Height > 0 {
		c.state.revealHandled = request
		Dispatch(func() {
			if c.state.revealRequest.Peek() == request {
				ctx.scrollToView(nil, bounds.Y, 1)
			}
		})
	}
}
