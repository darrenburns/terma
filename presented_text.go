package terma

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/darrenburns/terma/layout"
)

// PresentedTextPaint describes the concrete text payload a PresentedText leaf
// should draw during the paint phase.
type PresentedTextPaint struct {
	Content string
	Spans   []Span
	Style   Style
}

// PresentedText is a small leaf helper for the common "stable layout, reactive
// paint" pattern. Layout uses LayoutText/LayoutStyle, while Paint may react to
// transient state such as cursor, focus, or selection.
type PresentedText struct {
	LayoutText  string
	LayoutStyle Style
	Wrap        WrapMode
	TextAlign   TextAlign

	CurrentLayoutText func() string
	CurrentStyle      func() Style
	Paint             func(*RenderContext) PresentedTextPaint
	WidthHint         func() int
	HeightHint        func(width int) int
}

func (t PresentedText) Build(ctx BuildContext) Widget {
	return t
}

func (t PresentedText) GetContentDimensions() (width, height Dimension) {
	dims := t.LayoutStyle.GetDimensions()
	return dims.Width, dims.Height
}

func (t PresentedText) GetStyle() Style {
	if t.CurrentStyle != nil {
		return t.CurrentStyle()
	}
	return t.LayoutStyle
}

func (t PresentedText) layoutText() string {
	if t.CurrentLayoutText != nil {
		return t.CurrentLayoutText()
	}
	return t.LayoutText
}

func (t PresentedText) BuildLayoutNode(ctx BuildContext) layout.LayoutNode {
	return Text{
		Content:   t.layoutText(),
		Wrap:      t.Wrap,
		TextAlign: t.TextAlign,
		Style:     t.LayoutStyle,
	}.BuildLayoutNode(ctx)
}

func (t PresentedText) Render(ctx *RenderContext) {
	if t.Paint == nil {
		Text{
			Content:   t.layoutText(),
			Wrap:      t.Wrap,
			TextAlign: t.TextAlign,
			Style:     t.GetStyle(),
		}.Render(ctx)
		return
	}

	output := t.Paint(ctx)
	text := Text{
		Wrap:      t.Wrap,
		TextAlign: t.TextAlign,
		Style:     output.Style,
	}
	if len(output.Spans) > 0 {
		text.Spans = output.Spans
	} else {
		text.Content = output.Content
	}
	text.Render(ctx)
}

func (t PresentedText) ContentWidthHint() int {
	if t.WidthHint != nil {
		return t.WidthHint()
	}
	return ansi.StringWidth(t.layoutText())
}

func (t PresentedText) ContentHeightHint(width int) int {
	if t.HeightHint != nil {
		return t.HeightHint(width)
	}
	lines := wrapText(t.layoutText(), width, t.Wrap)
	if len(lines) == 0 {
		return 1
	}
	return len(lines)
}

// PresentText creates a PresentedText with a custom paint callback.
func PresentText(layoutText string, layoutStyle Style, currentStyle func() Style, paint func(*RenderContext) PresentedTextPaint) PresentedText {
	return PresentedText{
		LayoutText:   layoutText,
		LayoutStyle:  layoutStyle,
		CurrentStyle: currentStyle,
		Paint:        paint,
	}
}

// ComputedText creates a PresentedText whose visible content is recomputed at
// paint time, while layout remains based on the provided stable layout text.
func ComputedText(layoutText string, compute func() string) PresentedText {
	return PresentText(layoutText, Style{}, nil, func(*RenderContext) PresentedTextPaint {
		content := layoutText
		if compute != nil {
			content = compute()
		}
		return PresentedTextPaint{
			Content: content,
			Style:   Style{},
		}
	})
}

// ComputedSpans creates a PresentedText whose visible spans are recomputed at
// paint time, while layout remains based on the provided stable layout text.
func ComputedSpans(layoutText string, compute func(theme ThemeData) []Span) PresentedText {
	return PresentText(layoutText, Style{}, nil, func(ctx *RenderContext) PresentedTextPaint {
		spans := []Span(nil)
		if compute != nil {
			spans = compute(ctx.buildContext.Theme())
		}
		return PresentedTextPaint{
			Spans: spans,
			Style: Style{},
		}
	})
}

// SignalText renders a comparable signal's formatted value as reactive text.
// The signal is tracked in the paint phase, while layout sizing follows the
// current peeked text so auto-sized uses can still promote to layout correctly.
func SignalText[T comparable](signal Signal[T], format func(T) string) PresentedText {
	valueText := func(value T) string {
		if format != nil {
			return format(value)
		}
		return ""
	}
	currentText := func() string {
		return valueText(signal.Peek())
	}
	return PresentedText{
		CurrentLayoutText: currentText,
		Paint: func(*RenderContext) PresentedTextPaint {
			return PresentedTextPaint{
				Content: valueText(signal.Get()),
				Style:   Style{},
			}
		},
	}
}

// SignalMarkup renders a comparable signal's formatted value as reactive markup.
// The markup string is reparsed during paint, while layout sizing follows the
// current peeked plain-text content.
func SignalMarkup[T comparable](signal Signal[T], format func(T) string) PresentedText {
	valueMarkup := func(value T) string {
		if format != nil {
			return format(value)
		}
		return ""
	}
	currentText := func() string {
		return spansText(ParseMarkup(valueMarkup(signal.Peek()), getTheme()))
	}
	return PresentedText{
		CurrentLayoutText: currentText,
		Paint: func(ctx *RenderContext) PresentedTextPaint {
			return PresentedTextPaint{
				Spans: ParseMarkup(valueMarkup(signal.Get()), ctx.buildContext.Theme()),
				Style: Style{},
			}
		},
	}
}

// AnySignalText renders an AnySignal's formatted value as reactive text.
func AnySignalText[T any](signal AnySignal[T], format func(T) string) PresentedText {
	valueText := func(value T) string {
		if format != nil {
			return format(value)
		}
		return ""
	}
	currentText := func() string {
		return valueText(signal.Peek())
	}
	return PresentedText{
		CurrentLayoutText: currentText,
		Paint: func(*RenderContext) PresentedTextPaint {
			return PresentedTextPaint{
				Content: valueText(signal.Get()),
				Style:   Style{},
			}
		},
	}
}

// AnySignalMarkup renders an AnySignal's formatted value as reactive markup.
func AnySignalMarkup[T any](signal AnySignal[T], format func(T) string) PresentedText {
	valueMarkup := func(value T) string {
		if format != nil {
			return format(value)
		}
		return ""
	}
	currentText := func() string {
		return spansText(ParseMarkup(valueMarkup(signal.Peek()), getTheme()))
	}
	return PresentedText{
		CurrentLayoutText: currentText,
		Paint: func(ctx *RenderContext) PresentedTextPaint {
			return PresentedTextPaint{
				Spans: ParseMarkup(valueMarkup(signal.Get()), ctx.buildContext.Theme()),
				Style: Style{},
			}
		},
	}
}

func spansText(spans []Span) string {
	if len(spans) == 0 {
		return ""
	}
	result := make([]byte, 0)
	for _, span := range spans {
		result = append(result, span.Text...)
	}
	return string(result)
}

// PresentStyledText creates a PresentedText whose paint output is the layout
// text with a paint-phase style.
func PresentStyledText(layoutText string, layoutStyle Style, currentStyle func() Style, paintStyle func(*RenderContext) Style) PresentedText {
	return PresentText(layoutText, layoutStyle, currentStyle, func(ctx *RenderContext) PresentedTextPaint {
		style := layoutStyle
		if paintStyle != nil {
			style = paintStyle(ctx)
		}
		return PresentedTextPaint{
			Content: layoutText,
			Style:   style,
		}
	})
}

// PresentHighlightedText creates a PresentedText that highlights matched text
// ranges during paint while keeping layout stable.
func PresentHighlightedText(layoutText string, layoutStyle Style, currentStyle func() Style, paintStyle func(*RenderContext) Style, match func(*RenderContext) MatchResult) PresentedText {
	return PresentText(layoutText, layoutStyle, currentStyle, func(ctx *RenderContext) PresentedTextPaint {
		style := layoutStyle
		if paintStyle != nil {
			style = paintStyle(ctx)
		}
		if match != nil {
			result := match(ctx)
			if result.Matched && len(result.Ranges) > 0 {
				return PresentedTextPaint{
					Spans: HighlightSpans(layoutText, result.Ranges, MatchHighlightStyle(ctx.buildContext.Theme())),
					Style: style,
				}
			}
		}
		return PresentedTextPaint{
			Content: layoutText,
			Style:   style,
		}
	})
}

// PresentPrefixedText creates a PresentedText that prepends a paint-phase
// prefix while keeping layout stable. Highlight ranges apply only to the main
// content, not the prefix.
func PresentPrefixedText(layoutText, content string, layoutStyle Style, currentStyle func() Style, paintStyle func(*RenderContext) Style, prefix func(*RenderContext) string, match func(*RenderContext) MatchResult) PresentedText {
	return PresentText(layoutText, layoutStyle, currentStyle, func(ctx *RenderContext) PresentedTextPaint {
		style := layoutStyle
		if paintStyle != nil {
			style = paintStyle(ctx)
		}
		currentPrefix := ""
		if prefix != nil {
			currentPrefix = prefix(ctx)
		}
		if match != nil {
			result := match(ctx)
			if result.Matched && len(result.Ranges) > 0 {
				spans := make([]Span, 0, 1+len(result.Ranges)*2)
				if currentPrefix != "" {
					spans = append(spans, PlainSpan(currentPrefix))
				}
				spans = append(spans, HighlightSpans(content, result.Ranges, MatchHighlightStyle(ctx.buildContext.Theme()))...)
				return PresentedTextPaint{
					Spans: spans,
					Style: style,
				}
			}
		}
		return PresentedTextPaint{
			Content: currentPrefix + content,
			Style:   style,
		}
	})
}
