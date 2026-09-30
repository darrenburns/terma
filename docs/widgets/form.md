# Fields and forms

`Field` composes a label, an ordinary input, help text and validation messages.
`Form` coordinates explicitly registered field states and successful submission.
Input state stays in its existing widget; field state owns presentation timing and
an accepted baseline for dirty/reset behaviour.

## Interface

```go
name := terma.NewTextInputState("")
nameField := terma.NewTextInputField("name", name, terma.Required("Enter a name"))
formState := terma.NewFormState(nameField)

form := terma.Form{State: formState, OnSubmit: save}
form.Child = terma.Column{Spacing: 1, Children: []terma.Widget{
    terma.Field{
        Label: "Name", Help: "Shown publicly", State: nameField,
        Child: terma.TextInput{State: name},
    },
    terma.Button{ID: "save", Label: "Save", OnPress: func() { form.Submit() }},
}}
```

`NewTextInputField` and `NewTextAreaField` supply reactive bindings to existing
text state. For a signal or another state module, use the typed constructor:

```go
accepted := terma.NewSignal(false)
acceptedField := terma.NewFieldState("accepted", terma.FieldBinding[bool]{
    Read: accepted.Get, Write: accepted.Set,
}, func(value bool) string {
    if !value { return "Accept the terms" }
    return ""
})
```

`FieldBinding[T comparable]` is a read/write seam. Read must subscribe to the
actual source (use `Get`, not `Peek`), and Write must update it. Values must be
stable reflexive comparable values. Constructor IDs must be nonempty and bindings
must supply both functions; invalid configuration panics before registration.
Text helper constructors reject nil input state; using a different input state as
the adapted child also panics. `InputID()` returns the stable focus target for
custom inputs. Initial text is not normalized.

Field state exposes `Errors()`, `Touched()`, `Dirty()`, `Enabled()`, `Valid()`,
`Touch()`, `Validate()`, `SetEnabled(bool)`, `Reset()` and `Accept()`.
`Errors`, `Touched`, `Dirty`, `Enabled` and `Valid` are reactive reads.
Form state exposes `SetFields`, `Fields`, `Valid`, `Dirty`, `Validate`, `Reset`
and `Accept`. `Form.Submit()` validates, focuses the first invalid field, and
invokes `OnSubmit` only when valid; it returns whether submission succeeded.

## Behaviour specification and edge cases

### Validation and visibility

* A validator is a pure synchronous `func(T) string`. Empty/whitespace-only
  messages mean success. Every non-nil validator runs in declaration order and
  all nonempty messages appear in that order. Duplicate messages are retained.
  Validators must not mutate signals or perform side effects: reactive rendering
  may reevaluate them more than once. No async validation or network validation
  is included.
* Untouched fields initially hide errors. Typing alone does not reveal them.
  Blurring an adapted input calls `Touch`, which sets touched and reveals errors.
  `Validate` reveals errors without marking untouched fields as touched.
  Submitting validates/reveals every enabled registered field, not just the first
  failing one. `Valid` queries validity without revealing errors.
* Once revealed, errors derive from current source values on every reactive read.
  They update on both typing and programmatic setters without requiring a second
  change callback. Correcting a value removes errors immediately. Returning to an
  invalid value shows errors again. A pure cross-field validator can read another
  field's signal with `Get`; its dependencies update errors reactively. Validation
  is not called recursively across field states.
* `Required(message)` rejects empty and Unicode-whitespace-only strings using
  `strings.TrimSpace`. It does not rewrite the source. Unicode text, punctuation
  and embedded newlines otherwise pass; other constraints belong in validators.
  An empty message chooses the default “Required”. Text length policy is left to
  the application (bytes, runes and grapheme clusters differ).

### Dirty, reset and programmatic changes

* Dirty is exact inequality from the accepted baseline. The baseline is the source
  value at construction, or the value at the latest `Accept`. Editing back to it
  makes Dirty false. Whitespace differences count; no implicit trimming occurs.
* `Reset` restores the baseline through the binding, clears touched and hides
  validation messages. `Accept` retains the current value, makes it the baseline,
  clears touched and hides messages. Neither calls input OnChange or OnSubmit.
  Form Reset/Accept applies to all registered fields, including disabled ones.
* A programmatic source update changes Dirty and any revealed errors reactively,
  but never marks Touched or calls submit/change hooks. Call Accept explicitly
  after loading a new accepted record. Build never changes source or field state.

### Participation, identity and focus

* Form fields are explicitly registered in validation/focus order. Nil entries are
  ignored. Repeated pointers are de-duplicated; different states with the same ID
  panic instead of ambiguously sharing focus. `SetFields` replaces the registry
  atomically after checking identities; removed states retain their own metadata.
  `Fields` returns a defensive copy. Empty forms are valid and can submit.
* `SetEnabled(false)` excludes a field from form validity/dirty/submit, hides its
  errors and disables the Field child subtree. Re-enabling preserves its value,
  baseline, touched/revealed metadata; existing errors become visible again.
  Generic `DisabledWhen` only controls presentation: use SetEnabled or remove the
  field from SetFields when it should not participate in validation.
* Hidden or unmounted registered fields continue to validate. The caller must
  remove or disable them when they should not participate. An invalid unmounted
  field blocks submission; requesting its absent focus ID is harmless and does
  not silently validate it. No global tree registration or Build mutation occurs.
* A Field with nil state renders its label/help/child unchanged and has no
  validation behaviour. A nil Child is safe. A Form with nil state cannot submit
  and exposes no submit bindings. Field and Form styling use ordinary Style.
* Adapted TextInput, TextArea and Checkbox children receive the field state's ID,
  ensuring the first-invalid focus target matches the actual input. For custom
  children, use that ID explicitly and call field.Touch from their blur handler.

### Input and submission

* Fields adapt direct TextInput/TextArea values or pointers and Checkbox pointers.
  Existing OnChange, Blur, OnSubmit, paste, mouse and extra-key hooks are preserved.
  Blur marks the field touched before invoking the original Blur hook.
* A single-line Field TextInput with no explicit OnSubmit lets Enter bubble to
  its enclosing Form. Its explicit OnSubmit or extra Enter keybind takes
  precedence and prevents an additional form submission. Outside a Form, an
  unconfigured Enter is simply unhandled. Multiline TextArea retains Enter as
  newline; Ctrl+S submits from anywhere in the form unless an input overrides it.
* Form declares Enter and Ctrl+S through Keybinds. Child widgets keep precedence
  for keys they handle (buttons press and checkboxes toggle normally).
* Submit validates all enabled fields before requesting first-invalid focus or
  calling OnSubmit. Invalid attempts never invoke OnSubmit. Valid attempts invoke
  it once after validation, even for an empty form. Reentrant submission from a
  validator/callback is suppressed during the current attempt. Separate user
  attempts can each submit; there is no implicit “only once forever” policy.
* Failed submission reveals the first-invalid input through every enclosing Scrollable after validation messages have reflowed. A later failed attempt reveals it again even when it already has focus; manual scrolling is otherwise preserved. Custom children can still use InputID, but their first row is the reveal target.
* Validation order is registry order, not visual tree order. Validators do not
  modify values. Failed submission preserves values and dirty state. Successful
  submission does not automatically Accept; callers choose when saving succeeds.

## Validation plan

Unit and retained-render tests cover generic/text bindings, Required Unicode
whitespace, all validator results/order, untouched/change/blur/submit timing,
source changes and cross-field dependencies, dirty/revert/Reset/Accept, disabled
and dynamic registration, nil/duplicate/missing IDs, first-invalid focus and
callback ordering/reentrancy. Real key and mouse routing tests cover Enter,
Ctrl+S, multiline newline, hooks and first-invalid focus. Namespaced SVG snapshots
cover pristine/touched/submit errors/corrected/disabled/cross-field/multiline and
narrow layouts. The browser demo shows live status and submission counts with
real keyboard and mouse inputs at default and fixed terminal sizes.

For modal composition checks, run `go run ./cmd/form-demo -dialog`. The same
form is mounted inside an ordinary `Dialog`; input editing and submission stay
within the modal focus scope.
