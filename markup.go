package terma

import (
	"strings"
)

// ParseMarkup parses a markup string and returns a slice of Spans.
// Supports styles like [bold], [italic], [underline] (or [b], [i], [u]),
// theme colors like [$Primary], background colors like [on $Surface],
// and literal hex colors like [#ff5500].
//
// Examples:
//
//	ParseMarkup("Hello [bold]World[/]", theme)
//	ParseMarkup("Press [b $Accent]Enter[/] to continue", theme)
//	ParseMarkup("[bold $Error on $Background]Warning![/]", theme)
//
// Style nesting is supported: [bold]Hello [italic]World[/][/]
// Use [[ to insert a literal [ character.
// Invalid markup is returned as literal text (graceful fallback).
func ParseMarkup(markup string, theme ThemeData) []Span {
	p := &markupParser{
		input:      markup,
		theme:      theme,
		styleStack: []SpanStyle{{}}, // start with empty base style
	}
	return p.parse()
}

// ParseMarkupToText parses a markup string and returns a Text widget.
// This is a convenience wrapper around ParseMarkup.
//
// Example:
//
//	ParseMarkupToText("Press [b $Warning]Tab[/] to switch focus", theme)
func ParseMarkupToText(markup string, theme ThemeData) Text {
	return Text{Spans: ParseMarkup(markup, theme)}
}

type markupParser struct {
	input      string
	pos        int
	theme      ThemeData
	styleStack []SpanStyle
	spans      []Span
}

func (p *markupParser) parse() []Span {
	var textBuf strings.Builder

	for p.pos < len(p.input) {
		ch := p.input[p.pos]

		if ch == '[' {
			// Check for escape sequence [[
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '[' {
				textBuf.WriteByte('[')
				p.pos += 2
				continue
			}

			// Save current style before parsing tag
			styleBeforeTag := p.currentStyle()
			oldPos := p.pos

			// Try to parse the tag
			if p.parseTag() {
				// Tag was valid - flush any accumulated text with the style before the tag
				if textBuf.Len() > 0 {
					p.spans = append(p.spans, Span{
						Text:  textBuf.String(),
						Style: styleBeforeTag,
					})
					textBuf.Reset()
				}
			} else {
				// Invalid tag - treat [ as literal
				p.pos = oldPos
				textBuf.WriteByte('[')
				p.pos++
			}
		} else if ch == ']' {
			// Check for escape sequence ]]
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == ']' {
				textBuf.WriteByte(']')
				p.pos += 2
				continue
			}
			textBuf.WriteByte(ch)
			p.pos++
		} else {
			textBuf.WriteByte(ch)
			p.pos++
		}
	}

	// Flush remaining text
	if textBuf.Len() > 0 {
		p.spans = append(p.spans, Span{
			Text:  textBuf.String(),
			Style: p.currentStyle(),
		})
	}

	return p.spans
}

func (p *markupParser) currentStyle() SpanStyle {
	if len(p.styleStack) == 0 {
		return SpanStyle{}
	}
	return p.styleStack[len(p.styleStack)-1]
}

func (p *markupParser) parseTag() bool {
	// We're at '[', find the closing ']'
	start := p.pos + 1
	end := strings.IndexByte(p.input[start:], ']')
	if end == -1 {
		return false // No closing bracket
	}
	end += start

	tagContent := strings.TrimSpace(p.input[start:end])
	p.pos = end + 1

	// Check for closing tag [/] or [/tagname]
	if strings.HasPrefix(tagContent, "/") {
		// Pop style from stack
		if len(p.styleStack) > 1 {
			p.styleStack = p.styleStack[:len(p.styleStack)-1]
		}
		return true
	}

	// Parse opening tag - extract styles and colors
	style := p.parseTagContent(tagContent)
	p.styleStack = append(p.styleStack, style)
	return true
}

func (p *markupParser) parseTagContent(content string) SpanStyle {
	// Start with current style (for inheritance)
	style := p.currentStyle()

	// Split by whitespace, but handle "on" specially
	tokens := tokenizeTagContent(content)

	expectingBackground := false
	for _, token := range tokens {
		lower := strings.ToLower(token)

		// Check for "on" keyword
		if lower == "on" {
			expectingBackground = true
			continue
		}

		// Check for style modifiers
		switch lower {
		case "bold", "b":
			style.Bold = true
			continue
		case "italic", "i":
			style.Italic = true
			continue
		case "underline", "u":
			style.Underline = UnderlineSingle
			continue
		}

		// Try to parse as color
		if color, ok := p.resolveColor(token); ok {
			if expectingBackground {
				style.Background = color
				expectingBackground = false
			} else {
				style.Foreground = color
			}
		}
	}

	return style
}

func tokenizeTagContent(content string) []string {
	var tokens []string
	var current strings.Builder

	for i := 0; i < len(content); i++ {
		ch := content[i]
		if ch == ' ' || ch == '\t' {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func (p *markupParser) resolveColor(token string) (Color, bool) {
	// Hex color: #rrggbb or #rgb
	if strings.HasPrefix(token, "#") {
		return Hex(token), true
	}

	// Theme variable: $Primary, $primary, $text_muted, etc.
	if strings.HasPrefix(token, "$") {
		name := token[1:]
		return p.resolveThemeColor(name)
	}

	return Color{}, false
}

func (p *markupParser) resolveThemeColor(name string) (Color, bool) {
	// Normalize: lowercase and remove underscores for comparison
	normalized := strings.ToLower(strings.ReplaceAll(name, "_", ""))

	switch normalized {
	case "primary":
		return p.theme.Primary, true
	case "secondary":
		return p.theme.Secondary, true
	case "accent":
		return p.theme.Accent, true
	case "background":
		return p.theme.Background, true
	case "surface":
		return p.theme.Surface, true
	case "surfacehover":
		return p.theme.SurfaceHover, true
	case "text":
		return p.theme.Text, true
	case "textmuted":
		return p.theme.TextMuted, true
	case "textonprimary":
		return p.theme.TextOnPrimary, true
	case "border":
		return p.theme.Border, true
	case "focusring":
		return p.theme.FocusRing, true
	case "error":
		return p.theme.Error, true
	case "warning":
		return p.theme.Warning, true
	case "success":
		return p.theme.Success, true
	case "info":
		return p.theme.Info, true
	default:
		return Color{}, false
	}
}
