package terma

import (
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// KeybindBar displays available keybinds based on the currently focused widget.
// It automatically updates when focus changes, showing keybinds from the focused
// widget and its ancestors in the widget tree.
//
// Keybinds are deduplicated by key, with the focused widget taking precedence
// over ancestors. Keybinds with Hidden=true are not displayed.
//
// Consecutive keybinds with the same Name are grouped together, displaying
// their keys joined with "/" (e.g., "enter/space Press").
//
// Hints that don't fit in the bar's width are left out, from the end, so the
// bar never draws past its box.
type KeybindBar struct {
	Style  Style     // Optional styling (background, padding, etc.)
	Width  Dimension // Width dimension (default: Flex(1) to fill available width)
	Height Dimension // Height dimension (default: Cells(1) for single-line bar)

	// FormatKey transforms key strings for display. If nil, uses minimal
	// normalization (e.g., " " → "space"). Use preset formatters like
	// FormatKeyCaret, FormatKeyEmacs, FormatKeyVim, or FormatKeyVerbose,
	// or provide a custom function.
	FormatKey func(string) string
}

// GetContentDimensions returns the width and height dimension preferences.
// Width defaults to Flex(1), filling the space its siblings leave. Height
// defaults to Cells(1), as KeybindBar is a single-line widget.
func (f KeybindBar) GetContentDimensions() (width, height Dimension) {
	w, h := f.Width, f.Height
	if w.IsUnset() {
		w = Flex(1)
	}
	if h.IsUnset() {
		h = Cells(1)
	}
	return w, h
}

// GetStyle returns the style.
func (f KeybindBar) GetStyle() Style {
	return f.Style
}

// keybindGroup represents a group of keys that share the same action name.
type keybindGroup struct {
	keys []string
	name string
}

// Build constructs the keybind bar by collecting active keybinds from context.
func (f KeybindBar) Build(ctx BuildContext) Widget {
	keybinds := ctx.ActiveKeybinds()
	theme := ctx.Theme()
	width, height := f.GetContentDimensions()

	if len(keybinds) == 0 {
		return Text{Width: width, Height: height, Style: f.Style}
	}

	// Filter out hidden keybinds and deduplicate by key
	var visible []Keybind
	seenKeys := make(map[string]bool)

	for _, kb := range keybinds {
		if kb.Hidden || seenKeys[kb.Key] {
			continue
		}
		seenKeys[kb.Key] = true
		visible = append(visible, kb)
	}

	// Group consecutive keybinds with the same Name
	var groups []keybindGroup

	for _, kb := range visible {
		key := f.formatKey(kb.Key)

		// Check if we can add to the last group (same Name)
		if len(groups) > 0 && groups[len(groups)-1].name == kb.Name {
			groups[len(groups)-1].keys = append(groups[len(groups)-1].keys, key)
		} else {
			groups = append(groups, keybindGroup{keys: []string{key}, name: kb.Name})
		}
	}

	// Build spans from groups
	var spans []Span
	hints := make([][]Span, 0, len(groups))

	for _, g := range groups {
		if len(spans) > 0 {
			spans = append(spans, PlainSpan(" "))
		}

		// Join keys with /
		keyStr := strings.Join(g.keys, "/")
		hint := []Span{ColorSpan(keyStr, theme.Accent), ColorSpan(" "+g.name, theme.TextMuted)}
		spans = append(spans, hint...)
		hints = append(hints, hint)
	}

	return keybindBarLine{
		Text: Text{
			Spans:  spans,
			Style:  f.Style,
			Width:  width,
			Height: height,
		},
		hints: hints,
	}
}

// keybindBarLine is the text of a KeybindBar. It lays out like the Text it
// embeds, but paints only the hints that fit whole: a hint cut off mid-word
// would read as a different one.
type keybindBarLine struct {
	Text
	hints [][]Span
}

func (l keybindBarLine) Build(ctx BuildContext) Widget {
	return l
}

func (l keybindBarLine) Render(ctx *RenderContext) {
	var spans []Span
	used := 0
	for _, hint := range l.hints {
		width := 0
		for _, span := range hint {
			width += ansi.StringWidth(span.Text)
		}
		if len(spans) > 0 {
			width++ // The space before the hint.
		}
		if used+width > ctx.Width {
			break
		}
		if len(spans) > 0 {
			spans = append(spans, PlainSpan(" "))
		}
		spans = append(spans, hint...)
		used += width
	}
	l.Text.Spans = spans
	l.Text.Render(ctx)
}

// formatKey applies the custom FormatKey function if set, otherwise uses
// minimal normalization.
func (f KeybindBar) formatKey(key string) string {
	if f.FormatKey != nil {
		return f.FormatKey(key)
	}
	// Default: minimal normalization
	if key == " " {
		return "space"
	}
	return key
}
