package terma

import "strings"

// KeybindBar displays available keybinds based on the currently focused widget.
// It automatically updates when focus changes, showing keybinds from the focused
// widget and its ancestors in the widget tree.
//
// Keybinds are deduplicated by key, with the focused widget taking precedence
// over ancestors. Keybinds with Hidden=true are not displayed.
//
// Consecutive keybinds with the same Name are grouped together, displaying
// their keys joined with "/" (e.g., "enter/space Press").
type KeybindBar struct {
	Style  Style // Optional styling (background, padding, etc.)
	Width  Dimension // Deprecated: use Style.Width
	Height Dimension // Deprecated: use Style.Height

	// FormatKey transforms key strings for display. If nil, uses minimal
	// normalization (e.g., " " → "space"). Use preset formatters like
	// FormatKeyCaret, FormatKeyEmacs, FormatKeyVim, or FormatKeyVerbose,
	// or provide a custom function.
	FormatKey func(string) string
}

// GetContentDimensions returns the width and height dimension preferences.
// Width defaults to Flex(1) if not explicitly set, as KeybindBar typically fills width.
// Height defaults to Cells(1) if not explicitly set, as KeybindBar is a single-line widget.
func (f KeybindBar) GetContentDimensions() (width, height Dimension) {
	dims := f.Style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = f.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = f.Height
	}
	dims = dims.WithDefaults(Flex(1), Cells(1))
	return dims.Width, dims.Height
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
	dims := f.Style.GetDimensions()
	if dims.Width.IsUnset() {
		dims.Width = f.Width
	}
	if dims.Height.IsUnset() {
		dims.Height = f.Height
	}
	dims = dims.WithDefaults(Flex(1), Cells(1))
	style := f.Style
	style.Width = dims.Width
	style.Height = dims.Height

	if len(keybinds) == 0 {
		return Text{Style: style}
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

	for _, g := range groups {
		if len(spans) > 0 {
			spans = append(spans, PlainSpan(" "))
		}

		// Join keys with /
		keyStr := strings.Join(g.keys, "/")
		spans = append(spans, ColorSpan(keyStr, theme.Accent))
		spans = append(spans, ColorSpan(" "+g.name, theme.TextMuted))
	}

	return Text{
		Spans: spans,
		Style: style,
	}
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
