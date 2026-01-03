package terma

import "fmt"

// Button is a focusable widget that renders as styled text.
// It can be pressed with Enter or Space when focused.
type Button struct {
	ID      string    // Unique identifier for the button (required for focus management)
	Label   string    // Display text for the button
	OnPress func()    // Callback invoked when button is pressed
	Width   Dimension // Optional width (zero value = auto)
	Height  Dimension // Optional height (zero value = auto)
	Style   Style     // Optional styling (colors) applied when not focused
}

// WidgetID returns the button's unique identifier.
// Implements the Identifiable interface.
func (b *Button) WidgetID() string {
	return b.ID
}

// IsFocusable returns true, indicating this button can receive keyboard focus.
// Implements the Focusable interface.
func (b *Button) IsFocusable() bool {
	return true
}

// Keybinds returns the declarative keybindings for this button.
// The button responds to Enter and Space to trigger the OnPress callback.
// Implements the KeybindProvider interface.
func (b *Button) Keybinds() []Keybind {
	return []Keybind{
		{Key: "enter", Name: "Press", Action: b.press},
		{Key: " ", Name: "Press", Action: b.press},
	}
}

// press invokes the OnPress callback if it's set.
func (b *Button) press() {
	if b.OnPress != nil {
		b.OnPress()
	}
}

// OnKey handles keys not covered by declarative keybindings.
// Since Enter and Space are handled via Keybinds(), this returns false.
// Implements the Focusable interface.
func (b *Button) OnKey(event KeyEvent) bool {
	return false
}

// Build returns a Text widget with appropriate styling based on focus state.
// When focused, the button is highlighted with theme colors.
// If no explicit style colors are set, theme defaults are applied.
func (b *Button) Build(ctx BuildContext) Widget {
	label := fmt.Sprintf("%s", b.Label)
	theme := ctx.Theme()
	style := b.Style

	// Apply theme defaults if no explicit colors set
	if !style.ForegroundColor.IsSet() {
		style.ForegroundColor = theme.Text
	}
	if !style.BackgroundColor.IsSet() {
		style.BackgroundColor = theme.Surface
	}

	if ctx.IsFocused(b) {
		// Highlight with theme colors when focused
		style.BackgroundColor = theme.Primary
		style.ForegroundColor = theme.TextOnPrimary
	}

	return Text{
		Content: label,
		Width:   b.Width,
		Height:  b.Height,
		Style:   style,
	}
}

// GetDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (b *Button) GetDimensions() (width, height Dimension) {
	return b.Width, b.Height
}
