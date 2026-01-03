package terma

// EdgeInsets represents spacing around the four edges of a widget.
type EdgeInsets struct {
	Top, Right, Bottom, Left int
}

// EdgeInsetsAll creates EdgeInsets with the same value for all sides.
func EdgeInsetsAll(value int) EdgeInsets {
	return EdgeInsets{Top: value, Right: value, Bottom: value, Left: value}
}

// EdgeInsetsXY creates EdgeInsets with separate horizontal and vertical values.
func EdgeInsetsXY(horizontal, vertical int) EdgeInsets {
	return EdgeInsets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// EdgeInsetsTRBL creates EdgeInsets with individual values for each side.
func EdgeInsetsTRBL(top, right, bottom, left int) EdgeInsets {
	return EdgeInsets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Horizontal returns the total horizontal inset (Left + Right).
func (e EdgeInsets) Horizontal() int {
	return e.Left + e.Right
}

// Vertical returns the total vertical inset (Top + Bottom).
func (e EdgeInsets) Vertical() int {
	return e.Top + e.Bottom
}

// BorderStyle defines the visual style of a border.
type BorderStyle int

// Border style constants.
const (
	BorderNone BorderStyle = iota
	BorderSquare
	BorderRounded
)

// DecorationPosition defines where a decoration appears on the border.
type DecorationPosition int

// Decoration position constants.
const (
	DecorationTopLeft DecorationPosition = iota
	DecorationTopCenter
	DecorationTopRight
	DecorationBottomLeft
	DecorationBottomCenter
	DecorationBottomRight
)

// BorderDecoration defines a text label on a border edge.
type BorderDecoration struct {
	Text     string
	Position DecorationPosition
	Color    Color // If unset (zero value), inherits border color
}

// BorderTitle creates a title decoration at the top-left of the border.
func BorderTitle(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationTopLeft}
}

// BorderTitleCenter creates a title decoration at the top-center of the border.
func BorderTitleCenter(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationTopCenter}
}

// BorderTitleRight creates a title decoration at the top-right of the border.
func BorderTitleRight(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationTopRight}
}

// BorderSubtitle creates a subtitle decoration at the bottom-left of the border.
func BorderSubtitle(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationBottomLeft}
}

// BorderSubtitleCenter creates a subtitle decoration at the bottom-center of the border.
func BorderSubtitleCenter(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationBottomCenter}
}

// BorderSubtitleRight creates a subtitle decoration at the bottom-right of the border.
func BorderSubtitleRight(text string) BorderDecoration {
	return BorderDecoration{Text: text, Position: DecorationBottomRight}
}

// Border defines the border appearance for a widget.
type Border struct {
	Style       BorderStyle
	Color       Color
	Decorations []BorderDecoration
}

// SquareBorder creates a square border with the given color and optional decorations.
func SquareBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderSquare, Color: color, Decorations: decorations}
}

// RoundedBorder creates a rounded border with the given color and optional decorations.
func RoundedBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderRounded, Color: color, Decorations: decorations}
}

// IsZero returns true if no border is set.
func (b Border) IsZero() bool {
	return b.Style == BorderNone
}

// Width returns the border width (1 if border is set, 0 otherwise).
// Borders consume 1 cell on each side.
func (b Border) Width() int {
	if b.Style == BorderNone {
		return 0
	}
	return 1
}

// Style defines the visual appearance of a widget.
type Style struct {
	ForegroundColor Color
	BackgroundColor Color
	Padding         EdgeInsets
	Margin          EdgeInsets
	Border          Border
}

// IsZero returns true if the style has no values set.
func (s Style) IsZero() bool {
	return !s.ForegroundColor.IsSet() &&
		!s.BackgroundColor.IsSet() &&
		s.Padding == (EdgeInsets{}) &&
		s.Margin == (EdgeInsets{}) &&
		s.Border.IsZero()
}

// SpanStyle defines text attributes for a span (colors + formatting).
type SpanStyle struct {
	Foreground Color
	Background Color
	Bold       bool
	Italic     bool
	Underline  bool
}

// Span represents a segment of text with its own styling.
type Span struct {
	Text  string
	Style SpanStyle
}

// PlainSpan creates a span with no styling.
func PlainSpan(text string) Span {
	return Span{Text: text}
}

// ColorSpan creates a span with a foreground color.
func ColorSpan(text string, fg Color) Span {
	return Span{Text: text, Style: SpanStyle{Foreground: fg}}
}

// BoldSpan creates a bold span with optional foreground color.
func BoldSpan(text string, fg ...Color) Span {
	s := SpanStyle{Bold: true}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// ItalicSpan creates an italic span with optional foreground color.
func ItalicSpan(text string, fg ...Color) Span {
	s := SpanStyle{Italic: true}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// UnderlineSpan creates an underlined span with optional foreground color.
func UnderlineSpan(text string, fg ...Color) Span {
	s := SpanStyle{Underline: true}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// StyledSpan creates a span with full style control.
func StyledSpan(text string, style SpanStyle) Span {
	return Span{Text: text, Style: style}
}
