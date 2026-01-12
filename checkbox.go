package terma

// Checkbox is a focusable widget that displays a toggleable checked/unchecked state.
// It can be toggled with Space or Enter when focused, or by clicking.
type Checkbox struct {
	ID       string     // Unique identifier for the checkbox (required for focus management)
	Label    string     // Optional text displayed next to the checkbox
	Checked  bool       // Current checked state (pass value from signal, not the signal itself)
	OnToggle func(bool) // Callback invoked when toggled, receives the new state
	Width    Dimension  // Optional width (zero value = auto)
	Height   Dimension  // Optional height (zero value = auto)
	Style    Style      // Optional styling (colors) applied when not focused
	Click    func()     // Optional callback invoked when clicked
	Hover    func(bool) // Optional callback invoked when hover state changes
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
// The checkbox responds to Space (shown in keybind bar) and Enter (hidden) to toggle.
// Implements the KeybindProvider interface.
func (c *Checkbox) Keybinds() []Keybind {
	return []Keybind{
		{Key: " ", Name: "Toggle", Action: c.toggle},
		{Key: "enter", Name: "Toggle", Action: c.toggle, Hidden: true},
	}
}

// toggle invokes the OnToggle callback with the inverted checked state.
func (c *Checkbox) toggle() {
	if c.OnToggle != nil {
		c.OnToggle(!c.Checked)
	}
}

// OnKey handles keys not covered by declarative keybindings.
// Since Space and Enter are handled via Keybinds(), this returns false.
// Implements the Focusable interface.
func (c *Checkbox) OnKey(event KeyEvent) bool {
	return false
}

// Build returns a Text widget displaying the checkbox indicator and label.
// When focused, the checkbox is highlighted with theme colors.
// If no explicit style colors are set, theme defaults are applied.
func (c *Checkbox) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	style := c.Style

	// Determine the indicator based on checked state
	indicator := "[ ]"
	if c.Checked {
		indicator = "[✓]"
	}

	// Compose content with optional label
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

	if ctx.IsFocused(c) {
		// Highlight with theme colors when focused
		style.BackgroundColor = theme.Primary
		style.ForegroundColor = theme.TextOnPrimary
	}

	return Text{
		Content: content,
		Width:   c.Width,
		Height:  c.Height,
		Style:   style,
	}
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (c *Checkbox) GetDimensions() (width, height Dimension) {
	return c.Width, c.Height
}

// GetStyle returns the checkbox's style.
// Implements the Styled interface.
func (c *Checkbox) GetStyle() Style {
	return c.Style
}

// OnClick is called when the widget is clicked.
// It toggles the checkbox and invokes the optional Click callback.
// Implements the Clickable interface.
func (c *Checkbox) OnClick() {
	c.toggle()
	if c.Click != nil {
		c.Click()
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (c *Checkbox) OnHover(hovered bool) {
	if c.Hover != nil {
		c.Hover(hovered)
	}
}
