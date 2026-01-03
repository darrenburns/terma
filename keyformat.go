package terma

import (
	"strings"
)

// keyParts represents the parsed components of a key string.
type keyParts struct {
	ctrl  bool
	alt   bool
	shift bool
	key   string // the base key (e.g., "x", "tab", "enter")
}

// parseKey parses a key string like "ctrl+shift+x" into its components.
func parseKey(s string) keyParts {
	var parts keyParts
	s = strings.ToLower(s)

	// Split by + and process each part
	segments := strings.Split(s, "+")
	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		switch seg {
		case "ctrl", "control":
			parts.ctrl = true
		case "alt", "meta", "opt", "option":
			parts.alt = true
		case "shift":
			parts.shift = true
		default:
			parts.key = seg
		}
	}

	return parts
}

// FormatKeyCaret formats keys in classic Unix caret notation.
// ctrl+x → ^X, other modifiers fall back to verbose style.
//
// Examples:
//   - "ctrl+c" → "^C"
//   - "ctrl+x" → "^X"
//   - "shift+tab" → "Shift+Tab"
//   - "enter" → "Enter"
//   - " " → "Space"
func FormatKeyCaret(key string) string {
	if key == " " {
		return "Space"
	}

	p := parseKey(key)

	// Pure ctrl+key → ^X
	if p.ctrl && !p.alt && !p.shift && len(p.key) == 1 {
		return "^" + strings.ToUpper(p.key)
	}

	// Fall back to verbose for anything else
	return FormatKeyVerbose(key)
}

// FormatKeyEmacs formats keys in Emacs style.
// ctrl → C-, alt/meta → M-, shift → S-
//
// Examples:
//   - "ctrl+x" → "C-x"
//   - "alt+x" → "M-x"
//   - "ctrl+alt+x" → "C-M-x"
//   - "shift+tab" → "S-TAB"
//   - "enter" → "RET"
//   - " " → "SPC"
func FormatKeyEmacs(key string) string {
	if key == " " {
		return "SPC"
	}

	p := parseKey(key)

	// Handle special keys
	baseKey := p.key
	switch strings.ToLower(baseKey) {
	case "enter", "return":
		baseKey = "RET"
	case "escape", "esc":
		baseKey = "ESC"
	case "backspace":
		baseKey = "DEL"
	case "delete":
		baseKey = "delete"
	case "tab":
		baseKey = "TAB"
	default:
		// Keep as-is
	}

	var prefix string
	if p.ctrl {
		prefix += "C-"
	}
	if p.alt {
		prefix += "M-"
	}
	if p.shift {
		prefix += "S-"
	}

	return prefix + baseKey
}

// FormatKeyVim formats keys in Vim style.
// Keys are wrapped in angle brackets with modifier prefixes.
//
// Examples:
//   - "ctrl+x" → "<C-x>"
//   - "alt+x" → "<M-x>"
//   - "shift+tab" → "<S-Tab>"
//   - "enter" → "<CR>"
//   - " " → "<Space>"
func FormatKeyVim(key string) string {
	if key == " " {
		return "<Space>"
	}

	p := parseKey(key)

	// Handle special keys
	baseKey := p.key
	needsBrackets := p.ctrl || p.alt || p.shift
	switch strings.ToLower(baseKey) {
	case "enter", "return":
		baseKey = "CR"
		needsBrackets = true
	case "escape", "esc":
		baseKey = "Esc"
		needsBrackets = true
	case "backspace":
		baseKey = "BS"
		needsBrackets = true
	case "delete":
		baseKey = "Del"
		needsBrackets = true
	case "tab":
		baseKey = "Tab"
		needsBrackets = true
	default:
		// Capitalize first letter for special keys
		if len(baseKey) > 1 {
			baseKey = strings.ToUpper(baseKey[:1]) + baseKey[1:]
			needsBrackets = true
		}
	}

	var prefix string
	if p.ctrl {
		prefix += "C-"
	}
	if p.alt {
		prefix += "M-"
	}
	if p.shift {
		prefix += "S-"
	}

	if needsBrackets {
		return "<" + prefix + baseKey + ">"
	}
	return baseKey
}

// FormatKeyVerbose formats keys in readable title case.
//
// Examples:
//   - "ctrl+x" → "Ctrl+X"
//   - "shift+tab" → "Shift+Tab"
//   - "enter" → "Enter"
//   - " " → "Space"
func FormatKeyVerbose(key string) string {
	if key == " " {
		return "Space"
	}

	p := parseKey(key)

	// Handle special key names
	baseKey := p.key
	switch strings.ToLower(baseKey) {
	case "enter", "return":
		baseKey = "Enter"
	case "escape", "esc":
		baseKey = "Esc"
	case "backspace":
		baseKey = "Backspace"
	case "delete":
		baseKey = "Delete"
	case "tab":
		baseKey = "Tab"
	default:
		// Single char → uppercase, multi-char → title case
		if len(baseKey) == 1 {
			baseKey = strings.ToUpper(baseKey)
		} else if len(baseKey) > 1 {
			baseKey = strings.ToUpper(baseKey[:1]) + baseKey[1:]
		}
	}

	var parts []string
	if p.ctrl {
		parts = append(parts, "Ctrl")
	}
	if p.alt {
		parts = append(parts, "Alt")
	}
	if p.shift {
		parts = append(parts, "Shift")
	}
	parts = append(parts, baseKey)

	return strings.Join(parts, "+")
}
