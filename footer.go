package terma

import "strings"

// KeybindFooter displays available keybinds based on the currently focused widget.
// It automatically updates when focus changes, showing keybinds from the focused
// widget and its ancestors in the widget tree.
//
// Keybinds are deduplicated by key, with the focused widget taking precedence
// over ancestors. Keybinds with Hidden=true are not displayed.
//
// Consecutive keybinds with the same Name are grouped together, displaying
// their keys joined with "/" (e.g., "enter/space Press").
type KeybindFooter struct {
	Style Style // Optional styling (background, padding, etc.)
}

// keybindGroup represents a group of keys that share the same action name.
type keybindGroup struct {
	keys []string
	name string
}

// Build constructs the footer by collecting active keybinds from context.
func (f KeybindFooter) Build(ctx BuildContext) Widget {
	keybinds := ctx.ActiveKeybinds()
	theme := ctx.Theme()

	if len(keybinds) == 0 {
		return Text{Height: Cells(1), Style: f.Style}
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
		key := normalizeKeyForDisplay(kb.Key)

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
		Spans:  spans,
		Style:  f.Style,
		Height: Cells(1),
	}
}

// normalizeKeyForDisplay converts key patterns to user-friendly display names.
func normalizeKeyForDisplay(key string) string {
	if key == " " {
		return "space"
	}
	return key
}
