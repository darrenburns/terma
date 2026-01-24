package terma

// CheckboxState holds the state for a Checkbox widget.
// It is the source of truth for the checked state and must be provided to Checkbox.
type CheckboxState struct {
	Checked Signal[bool] // Reactive checked state
}

// NewCheckboxState creates a new CheckboxState with the given initial checked value.
func NewCheckboxState(checked bool) *CheckboxState {
	return &CheckboxState{
		Checked: NewSignal(checked),
	}
}

// Toggle flips the checked state.
func (s *CheckboxState) Toggle() {
	s.Checked.Update(func(c bool) bool { return !c })
}

// SetChecked sets the checked state to the given value.
func (s *CheckboxState) SetChecked(checked bool) {
	s.Checked.Set(checked)
}

// IsChecked returns the current checked state without subscribing to changes.
func (s *CheckboxState) IsChecked() bool {
	return s.Checked.Peek()
}

// Checkbox is a focusable widget that displays a checkable box with an optional label.
// It can be toggled with Enter or Space when focused.
type Checkbox struct {
	ID        string          // Unique identifier for the checkbox (required for focus management)
	State     *CheckboxState  // Required - holds checked state
	Label     string          // Optional text displayed after the indicator
	Width     Dimension       // Deprecated: use Style.Width
	Height    Dimension       // Deprecated: use Style.Height
	Style     Style           // Optional styling applied when not focused
	OnChange  func(bool)      // Optional callback invoked after state changes
	Click     func(MouseEvent) // Optional callback invoked when clicked
	MouseDown func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent) // Optional callback invoked when mouse is released
	Hover     func(bool)      // Optional callback invoked when hover state changes
}

// WidgetID returns the checkbox's unique identifier.
// Implements the Identifiable interface.
func (c *Checkbox) WidgetID() string {
	return c.ID
}

// IsFocusable returns true, indicating this checkbox can receive keyboard focus.
// Implements the Focusable interface.
func (c *Checkbox) IsFocusable() bool {
	return true
}

// Keybinds returns the declarative keybindings for this checkbox.
// The checkbox responds to Enter and Space to toggle the checked state.
// Implements the KeybindProvider interface.
func (c *Checkbox) Keybinds() []Keybind {
	return []Keybind{
		{Key: "enter", Name: "Toggle", Action: c.toggle},
		{Key: " ", Name: "Toggle", Action: c.toggle},
	}
}

// toggle flips the checked state and invokes the OnChange callback.
func (c *Checkbox) toggle() {
	if c.State != nil {
		c.State.Toggle()
		if c.OnChange != nil {
			c.OnChange(c.State.IsChecked())
		}
	}
}

// OnKey handles keys not covered by declarative keybindings.
// Since Enter and Space are handled via Keybinds(), this returns false.
// Implements the Focusable interface.
func (c *Checkbox) OnKey(event KeyEvent) bool {
	return false
}

// Build returns a Text widget with the checkbox indicator and label.
// The checkbox is rendered with appropriate styling based on focus and disabled state.
func (c *Checkbox) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	style := c.Style
	if style.Width.IsUnset() {
		style.Width = c.Width
	}
	if style.Height.IsUnset() {
		style.Height = c.Height
	}

	// Subscribe to state changes and get current checked value
	checked := false
	if c.State != nil {
		checked = c.State.Checked.Get()
	}

	// Determine indicator character
	indicator := "☐"
	if checked {
		indicator = "☑"
	}

	// Build content string
	content := indicator
	if c.Label != "" {
		content = indicator + " " + c.Label
	}

	// Apply theme defaults if no explicit colors set
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = theme.Text
	}
	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}

	// Handle disabled state
	if ctx.IsDisabled() {
		style.ForegroundColor = theme.TextDisabled
		return Text{
			Content: content,
			Style:   style,
		}
	}

	// Handle focused state
	if ctx.IsFocused(c) {
		style.BackgroundColor = theme.ActiveCursor
		style.ForegroundColor = theme.SelectionText
	}

	return Text{
		Content: content,
		Style:   style,
	}
}

// GetContentDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (c *Checkbox) GetContentDimensions() (width, height Dimension) {
	dims := c.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = c.Width
	}
	if height.IsUnset() {
		height = c.Height
	}
	return width, height
}

// OnClick is called when the widget is clicked.
// It toggles the checkbox state and invokes the Click callback.
// Implements the Clickable interface.
func (c *Checkbox) OnClick(event MouseEvent) {
	c.toggle()
	if c.Click != nil {
		c.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (c *Checkbox) OnMouseDown(event MouseEvent) {
	if c.MouseDown != nil {
		c.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (c *Checkbox) OnMouseUp(event MouseEvent) {
	if c.MouseUp != nil {
		c.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (c *Checkbox) OnHover(hovered bool) {
	if c.Hover != nil {
		c.Hover(hovered)
	}
}
