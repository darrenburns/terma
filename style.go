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
	BorderDouble
	BorderHeavy
	BorderDashed
	BorderAscii
)

// BorderCharSet contains the characters used to render a border.
// Some border styles use different characters for each edge.
type BorderCharSet struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Top         string // Horizontal character for top edge
	Bottom      string // Horizontal character for bottom edge
	Left        string // Vertical character for left edge
	Right       string // Vertical character for right edge
}

// GetBorderCharSet returns the character set for a given border style.
func GetBorderCharSet(style BorderStyle) BorderCharSet {
	switch style {
	case BorderSquare:
		return BorderCharSet{
			TopLeft: "┌", TopRight: "┐", BottomLeft: "└", BottomRight: "┘",
			Top: "─", Bottom: "─", Left: "│", Right: "│",
		}
	case BorderRounded:
		return BorderCharSet{
			TopLeft: "╭", TopRight: "╮", BottomLeft: "╰", BottomRight: "╯",
			Top: "─", Bottom: "─", Left: "│", Right: "│",
		}
	case BorderDouble:
		return BorderCharSet{
			TopLeft: "╔", TopRight: "╗", BottomLeft: "╚", BottomRight: "╝",
			Top: "═", Bottom: "═", Left: "║", Right: "║",
		}
	case BorderHeavy:
		return BorderCharSet{
			TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
			Top: "━", Bottom: "━", Left: "┃", Right: "┃",
		}
	case BorderDashed:
		return BorderCharSet{
			TopLeft: "┏", TopRight: "┓", BottomLeft: "┗", BottomRight: "┛",
			Top: "╍", Bottom: "╍", Left: "╏", Right: "╏",
		}
	case BorderAscii:
		return BorderCharSet{
			TopLeft: "+", TopRight: "+", BottomLeft: "+", BottomRight: "+",
			Top: "-", Bottom: "-", Left: "|", Right: "|",
		}
	default:
		return BorderCharSet{}
	}
}

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
	Text     string             // Plain text (used if Markup is empty)
	Markup   string             // Markup string, parsed at render time for styled text
	Position DecorationPosition
	Color    ColorProvider // Fallback color if markup has no color (or for plain text)
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

// BorderTitleMarkup creates a title decoration with markup at the top-left of the border.
// The markup is parsed at render time to access the theme.
func BorderTitleMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationTopLeft}
}

// BorderTitleCenterMarkup creates a title decoration with markup at the top-center of the border.
func BorderTitleCenterMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationTopCenter}
}

// BorderTitleRightMarkup creates a title decoration with markup at the top-right of the border.
func BorderTitleRightMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationTopRight}
}

// BorderSubtitleMarkup creates a subtitle decoration with markup at the bottom-left of the border.
func BorderSubtitleMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationBottomLeft}
}

// BorderSubtitleCenterMarkup creates a subtitle decoration with markup at the bottom-center of the border.
func BorderSubtitleCenterMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationBottomCenter}
}

// BorderSubtitleRightMarkup creates a subtitle decoration with markup at the bottom-right of the border.
func BorderSubtitleRightMarkup(markup string) BorderDecoration {
	return BorderDecoration{Markup: markup, Position: DecorationBottomRight}
}

// Border defines the border appearance for a widget.
type Border struct {
	Style       BorderStyle
	Color       ColorProvider // Can be Color or Gradient
	Decorations []BorderDecoration
}

// SquareBorder creates a square border with the given color and optional decorations.
//
//	┌───┐
//	│   │
//	└───┘
func SquareBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderSquare, Color: color, Decorations: decorations}
}

// RoundedBorder creates a rounded border with the given color and optional decorations.
//
//	╭───╮
//	│   │
//	╰───╯
func RoundedBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderRounded, Color: color, Decorations: decorations}
}

// DoubleBorder creates a double-line border with the given color and optional decorations.
//
//	╔═══╗
//	║   ║
//	╚═══╝
func DoubleBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderDouble, Color: color, Decorations: decorations}
}

// HeavyBorder creates a heavy/thick border with the given color and optional decorations.
//
//	┏━━━┓
//	┃   ┃
//	┗━━━┛
func HeavyBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderHeavy, Color: color, Decorations: decorations}
}

// DashedBorder creates a dashed border with the given color and optional decorations.
func DashedBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderDashed, Color: color, Decorations: decorations}
}

// AsciiBorder creates an ASCII-only border with the given color and optional decorations.
// Use this for maximum terminal compatibility.
//
//	+---+
//	|   |
//	+---+
func AsciiBorder(color Color, decorations ...BorderDecoration) Border {
	return Border{Style: BorderAscii, Color: color, Decorations: decorations}
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

// UnderlineStyle defines the visual style of underlined text.
type UnderlineStyle int

// Underline style constants.
const (
	UnderlineNone UnderlineStyle = iota
	UnderlineSingle
	UnderlineDouble
	UnderlineCurly
	UnderlineDotted
	UnderlineDashed
)

// Style defines the visual appearance of a widget.
type Style struct {
	ForegroundColor ColorProvider // Can be Color or Gradient
	BackgroundColor ColorProvider // Can be Color or Gradient

	// Text attributes
	Bold           bool
	Faint          bool
	Italic         bool
	Underline      UnderlineStyle
	UnderlineColor Color
	Blink          bool
	Reverse        bool
	Conceal        bool
	Strikethrough  bool
	FillLine       bool // Extend underline/strikethrough to fill the line width

	// Layout
	Padding EdgeInsets
	Margin  EdgeInsets
	Border  Border

	// Dimensions (content-box)
	Width     Dimension
	Height    Dimension
	MinWidth  Dimension
	MinHeight Dimension
	MaxWidth  Dimension
	MaxHeight Dimension
}

// IsZero returns true if the style has no values set.
func (s Style) IsZero() bool {
	fgSet := s.ForegroundColor != nil && s.ForegroundColor.IsSet()
	bgSet := s.BackgroundColor != nil && s.BackgroundColor.IsSet()
	return !fgSet &&
		!bgSet &&
		!s.Bold &&
		!s.Faint &&
		!s.Italic &&
		s.Underline == UnderlineNone &&
		!s.UnderlineColor.IsSet() &&
		!s.Blink &&
		!s.Reverse &&
		!s.Conceal &&
		!s.Strikethrough &&
		!s.FillLine &&
		s.Padding == (EdgeInsets{}) &&
		s.Margin == (EdgeInsets{}) &&
		s.Border.IsZero() &&
		s.Width.IsUnset() &&
		s.Height.IsUnset() &&
		s.MinWidth.IsUnset() &&
		s.MinHeight.IsUnset() &&
		s.MaxWidth.IsUnset() &&
		s.MaxHeight.IsUnset()
}

// GetDimensions returns the style's dimension fields.
func (s Style) GetDimensions() DimensionSet {
	return DimensionSet{
		Width:     s.Width,
		Height:    s.Height,
		MinWidth:  s.MinWidth,
		MinHeight: s.MinHeight,
		MaxWidth:  s.MaxWidth,
		MaxHeight: s.MaxHeight,
	}
}

// SpanStyle defines text attributes for a span (colors + formatting).
type SpanStyle struct {
	Foreground     Color
	Background     Color
	Bold           bool
	Faint          bool
	Italic         bool
	Underline      UnderlineStyle
	UnderlineColor Color
	Blink          bool
	Reverse        bool
	Conceal        bool
	Strikethrough  bool
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
	s := SpanStyle{Underline: UnderlineSingle}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// FaintSpan creates a faint/dim span with optional foreground color.
func FaintSpan(text string, fg ...Color) Span {
	s := SpanStyle{Faint: true}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// StrikethroughSpan creates a strikethrough span with optional foreground color.
func StrikethroughSpan(text string, fg ...Color) Span {
	s := SpanStyle{Strikethrough: true}
	if len(fg) > 0 {
		s.Foreground = fg[0]
	}
	return Span{Text: text, Style: s}
}

// StyledSpan creates a span with full style control.
func StyledSpan(text string, style SpanStyle) Span {
	return Span{Text: text, Style: style}
}

// CursorStyle configures the visual appearance of cursor and selection in list-like widgets.
// Embed this anonymously in widgets to get CursorPrefix/SelectedPrefix fields.
//
// By default, no prefixes are shown - the cursor and selection are indicated via background
// color highlighting only. Users who prefer or need visual indicators (e.g., for accessibility)
// can set CursorPrefix to "▶ " or similar.
//
// Example:
//
//	List[string]{
//	    CursorPrefix:   "▶ ",  // Show arrow on cursor row
//	    SelectedPrefix: "* ",  // Show asterisk on selected rows
//	    State: myState,
//	}
type CursorStyle struct {
	CursorPrefix   string // Prefix shown on active/cursor item (default: "")
	SelectedPrefix string // Prefix shown on selected items in multi-select (default: "")
}
