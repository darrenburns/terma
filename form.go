package terma

// FormState is an explicit ordered registry of fields. It does not discover the
// widget tree: remove or disable hidden/unmounted fields that must not validate.
// State mutations and submission run on the UI event loop, like input callbacks.
type FormState struct {
	fields     AnySignal[[]*FieldState]
	submitting bool
}

// NewFormState registers fields in validation and first-invalid focus order.
// Nil entries are ignored and repeated pointers de-duplicated. Distinct states
// sharing one input ID panic because focus would otherwise be ambiguous.
func NewFormState(fields ...*FieldState) *FormState {
	s := &FormState{fields: NewAnySignal([]*FieldState(nil))}
	s.SetFields(fields...)
	return s
}

// SetFields atomically replaces the registry after checking identities. It
// leaves removed fields' values and metadata unchanged. Never call during Build.
func (s *FormState) SetFields(fields ...*FieldState) {
	seen := make(map[string]*FieldState)
	registered := make([]*FieldState, 0, len(fields))
	for _, field := range fields {
		if field == nil {
			continue
		}
		if previous, exists := seen[field.inputID]; exists {
			if previous != field {
				panic("terma: duplicate form field input ID: " + field.inputID)
			}
			continue
		}
		seen[field.inputID] = field
		registered = append(registered, field)
	}
	s.fields.Set(registered)
}

// Fields returns a reactive defensive copy of the ordered registry.
func (s *FormState) Fields() []*FieldState { return append([]*FieldState(nil), s.fields.Get()...) }

// Valid checks all enabled fields without revealing their errors.
func (s *FormState) Valid() bool {
	valid := true
	for _, field := range s.fields.Get() {
		if !field.Valid() {
			valid = false
		}
	}
	return valid
}

// Dirty reports whether any enabled field differs from its accepted baseline.
func (s *FormState) Dirty() bool {
	dirty := false
	for _, field := range s.fields.Get() {
		if field.Enabled() && field.Dirty() {
			dirty = true
		}
	}
	return dirty
}

func (s *FormState) firstInvalid() *FieldState {
	var first *FieldState
	for _, field := range s.fields.Peek() {
		if !field.Validate() && first == nil {
			first = field
		}
	}
	return first
}

// Validate reveals errors on every enabled registered field without moving focus.
func (s *FormState) Validate() bool { return s.firstInvalid() == nil }

// Reset restores all registered fields, including disabled ones, to baseline.
func (s *FormState) Reset() {
	for _, field := range s.fields.Peek() {
		field.Reset()
	}
}

// Accept accepts current values and clears validation metadata on all fields.
func (s *FormState) Accept() {
	for _, field := range s.fields.Peek() {
		field.Accept()
	}
}

// Form provides submission coordination around an ordinary widget subtree.
// Enter from an unconfigured Field TextInput and Ctrl+S submit through declarative
// keybindings. Child keybindings and explicit input OnSubmit handlers win.
type Form struct {
	State    *FormState
	Child    Widget
	OnSubmit func()
	Style    Style
}

func (f Form) GetContentDimensions() (Dimension, Dimension) {
	dims := f.Style.GetDimensions()
	return dims.Width, dims.Height
}

func (f Form) Build(BuildContext) Widget {
	if f.Child == nil {
		return EmptyWidget{}
	}
	return Column{Style: f.Style, Children: []Widget{f.Child}}
}

// Keybinds exposes submission while preserving child-specific Enter behaviour.
func (f Form) Keybinds() []Keybind {
	if f.State == nil {
		return nil
	}
	return []Keybind{
		{Key: "ctrl+s", Name: "Submit", Action: func() { f.Submit() }},
		{Key: "enter", Name: "Submit", Hidden: true, Action: func() { f.Submit() }},
	}
}

// Submit validates every enabled field, requests first-invalid focus on failure,
// and invokes OnSubmit once on success. Reentrant calls during this attempt are
// suppressed. A successful submission does not automatically accept/reset values.
func (f Form) Submit() bool {
	if f.State == nil || f.State.submitting {
		return false
	}
	f.State.submitting = true
	defer func() { f.State.submitting = false }()
	if field := f.State.firstInvalid(); field != nil {
		field.revealRequest.Update(func(n uint64) uint64 { return n + 1 })
		RequestFocus(field.inputID)
		return false
	}
	if f.OnSubmit != nil {
		f.OnSubmit()
	}
	return true
}
