package terma

// ButtonVariant represents the semantic color variant for a Button.
type ButtonVariant int

const (
	ButtonDefault ButtonVariant = iota
	ButtonPrimary
	ButtonAccent
	ButtonSuccess
	ButtonError
	ButtonWarning
	ButtonInfo
)

// buttonVariantColors returns the foreground and background colors for a button variant.
func buttonVariantColors(variant ButtonVariant, theme ThemeData) (fg, bg Color) {
	switch variant {
	case ButtonPrimary:
		return theme.TextOnPrimary, theme.Primary
	case ButtonAccent:
		return theme.TextOnAccent, theme.Accent
	case ButtonSuccess:
		return theme.TextOnSuccess, theme.Success
	case ButtonError:
		return theme.TextOnError, theme.Error
	case ButtonWarning:
		return theme.TextOnWarning, theme.Warning
	case ButtonInfo:
		return theme.TextOnInfo, theme.Info
	default:
		return theme.Text, theme.Surface
	}
}

// Button is a focusable widget that renders as styled text.
// It can be pressed with Enter or Space when focused.
type Button struct {
	ID           string        // Optional unique identifier for the button
	DisableFocus bool          // If true, prevent keyboard focus
	Label        string        // Display text for the button
	Variant      ButtonVariant // Semantic color variant (default: ButtonDefault)
	OnPress      func()        // Callback invoked when button is pressed
	Width        Dimension     // Deprecated: use Style.Width
	Height       Dimension     // Deprecated: use Style.Height
	Style        Style         // Optional styling (colors) applied when not focused
	Click        func(MouseEvent) // Optional callback invoked when clicked
	MouseDown func(MouseEvent) // Optional callback invoked when mouse is pressed
	MouseUp   func(MouseEvent) // Optional callback invoked when mouse is released
	Hover   func(bool) // Optional callback invoked when hover state changes
}

// WidgetID returns the button's unique identifier.
// Implements the Identifiable interface.
func (b Button) WidgetID() string {
	return b.ID
}

// IsFocusable returns true, indicating this button can receive keyboard focus.
// Implements the Focusable interface.
func (b Button) IsFocusable() bool {
	return !b.DisableFocus
}

// Keybinds returns the declarative keybindings for this button.
// The button responds to Enter and Space to trigger the OnPress callback.
// Implements the KeybindProvider interface.
func (b Button) Keybinds() []Keybind {
	return []Keybind{
		{Key: "enter", Name: "Press", Action: b.press},
		{Key: " ", Name: "Press", Action: b.press},
	}
}

// press invokes the OnPress callback if it's set.
func (b Button) press() {
	if b.OnPress != nil {
		b.OnPress()
	}
}

// OnKey handles keys not covered by declarative keybindings.
// Since Enter and Space are handled via Keybinds(), this returns false.
// Implements the Focusable interface.
func (b Button) OnKey(event KeyEvent) bool {
	return false
}

// Build returns a Text widget with appropriate styling based on focus state.
// Buttons are rendered with bracket affordance: [label]
// When focused, the button is highlighted with variant colors (or theme.Primary for default).
// When disabled, the button shows disabled styling and brackets are faded.
// If no explicit style colors are set, variant-derived defaults are applied.
func (b Button) Build(ctx BuildContext) Widget {
	theme := ctx.Theme()
	style := b.Style
	if style.Width.IsUnset() {
		style.Width = b.Width
	}
	if style.Height.IsUnset() {
		style.Height = b.Height
	}

	// Resolve variant colors as defaults
	variantFg, variantBg := buttonVariantColors(b.Variant, theme)

	// Apply variant defaults if no explicit colors set
	if style.ForegroundColor == nil || !style.ForegroundColor.IsSet() {
		style.ForegroundColor = variantFg
	}
	if style.BackgroundColor == nil || !style.BackgroundColor.IsSet() {
		style.BackgroundColor = variantBg
	}

	// Get actual Color values for blending (ColorProvider could be Color or Gradient)
	// For blending, we use ColorAt(1,1,0,0) to get a representative color
	fg := style.ForegroundColor.ColorAt(1, 1, 0, 0)
	bg := style.BackgroundColor.ColorAt(1, 1, 0, 0)
	var bracketColor Color

	// Handle disabled state
	if ctx.IsDisabled() {
		style.ForegroundColor = theme.TextDisabled
		// Brackets blend 70% toward background (very faded)
		bracketColor = theme.TextDisabled.Blend(bg, 0.7)
		return Text{
			Spans: []Span{
				ColorSpan("[", bracketColor),
				PlainSpan(b.Label),
				ColorSpan("]", bracketColor),
			},
			Style:  style,
		}
	}

	if ctx.IsFocused(b) {
		// Highlight with variant colors when focused
		style.BackgroundColor = variantBg
		style.ForegroundColor = variantFg
		// Focused: brackets blend 55% toward background (visible but subtle)
		bracketColor = variantFg.Blend(variantBg, 0.55)
	} else {
		// Unfocused: brackets blend 85% toward background (very faded)
		bracketColor = fg.Blend(bg, 0.85)
	}

	return Text{
		Spans: []Span{
			ColorSpan("[", bracketColor),
			PlainSpan(b.Label),
			ColorSpan("]", bracketColor),
		},
		Style:  style,
	}
}

// GetContentDimensions returns the width and height dimension preferences.
// Implements the Dimensioned interface.
func (b Button) GetContentDimensions() (width, height Dimension) {
	dims := b.Style.GetDimensions()
	width, height = dims.Width, dims.Height
	if width.IsUnset() {
		width = b.Width
	}
	if height.IsUnset() {
		height = b.Height
	}
	return width, height
}

// OnClick is called when the widget is clicked.
// Implements the Clickable interface.
func (b Button) OnClick(event MouseEvent) {
	if b.Click != nil {
		b.Click(event)
	}
}

// OnMouseDown is called when the mouse is pressed on the widget.
// Implements the MouseDownHandler interface.
func (b Button) OnMouseDown(event MouseEvent) {
	if b.MouseDown != nil {
		b.MouseDown(event)
	}
}

// OnMouseUp is called when the mouse is released on the widget.
// Implements the MouseUpHandler interface.
func (b Button) OnMouseUp(event MouseEvent) {
	if b.MouseUp != nil {
		b.MouseUp(event)
	}
}

// OnHover is called when the hover state changes.
// Implements the Hoverable interface.
func (b Button) OnHover(hovered bool) {
	if b.Hover != nil {
		b.Hover(hovered)
	}
}
