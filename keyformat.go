package terma

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// keyParts represents the parsed components of a key string.
type keyParts struct {
	ctrl  bool
	alt   bool
	shift bool
	key   string // the base key as written (e.g., "x", "X", "tab", "enter")
}

// parseKey parses a key string like "ctrl+shift+x" into its components.
// Modifier names match in any case; the base key keeps its case, so "D" and
// "d" stay distinct.
func parseKey(s string) keyParts {
	var parts keyParts
	name := s
	// The key itself may be "+", as in "ctrl++".
	for len(name) > 1 {
		i := strings.Index(name, "+")
		if i <= 0 || i == len(name)-1 {
			break
		}
		switch strings.ToLower(strings.TrimSpace(name[:i])) {
		case "ctrl", "control":
			parts.ctrl = true
		case "alt", "meta", "opt", "option":
			parts.alt = true
		case "shift":
			parts.shift = true
		}
		name = name[i+1:]
	}
	if name != " " {
		name = strings.TrimSpace(name)
	}
	parts.key = name
	return parts
}

// formatChord formats each key press of a chord ("ctrl+k ctrl+s") with
// format and joins them with sep. A single key is formatted as is.
func formatChord(key string, sep string, format func(string) string) string {
	steps := keySteps(key)
	if len(steps) == 1 {
		return format(steps[0])
	}
	formatted := make([]string, len(steps))
	for i, step := range steps {
		formatted[i] = format(step)
	}
	return strings.Join(formatted, sep)
}

// titleCase capitalises the first letter of a multi-character key name
// ("pgup" → "Pgup"). A single character is returned as is: its case is part
// of the key.
func titleCase(key string) string {
	if utf8.RuneCountInString(key) <= 1 {
		return key
	}
	return strings.ToUpper(key[:1]) + key[1:]
}

// FormatKeyCaret formats keys in classic Unix caret notation.
// ctrl+x → ^x, other modifiers fall back to verbose style. The letter keeps
// its case: ^D means ctrl with an uppercase D (ctrl+D or ctrl+shift+d).
// Chords join their key presses with spaces.
//
// Examples:
//   - "ctrl+c" → "^c"
//   - "ctrl+shift+x" → "^X"
//   - "shift+tab" → "Shift+Tab"
//   - "enter" → "Enter"
//   - " " → "Space"
//   - "ctrl+x ctrl+s" → "^x ^s"
func FormatKeyCaret(key string) string {
	return formatChord(key, " ", formatKeyCaret)
}

func formatKeyCaret(key string) string {
	p := parseKey(key)
	if p.ctrl && !p.alt && utf8.RuneCountInString(p.key) == 1 && p.key != " " {
		if !p.shift {
			return "^" + p.key
		}
		if r := []rune(p.key)[0]; unicode.IsLetter(r) {
			return "^" + string(unicode.ToUpper(r))
		}
	}
	return formatKeyVerbose(key)
}

// FormatKeyEmacs formats keys in Emacs style.
// ctrl → C-, alt/meta → M-, shift → S-. Chords join their key presses with
// spaces.
//
// Examples:
//   - "ctrl+x" → "C-x"
//   - "alt+x" → "M-x"
//   - "ctrl+alt+x" → "C-M-x"
//   - "shift+tab" → "S-TAB"
//   - "enter" → "RET"
//   - " " → "SPC"
//   - "ctrl+x ctrl+s" → "C-x C-s"
func FormatKeyEmacs(key string) string {
	return formatChord(key, " ", formatKeyEmacs)
}

func formatKeyEmacs(key string) string {
	p := parseKey(key)

	baseKey := p.key
	switch strings.ToLower(baseKey) {
	case " ", "space":
		baseKey = "SPC"
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
// Named and modified keys are wrapped in angle brackets with modifier
// prefixes. Chords are written as one sequence, as Vim does.
//
// Examples:
//   - "ctrl+x" → "<C-x>"
//   - "alt+x" → "<M-x>"
//   - "shift+tab" → "<S-Tab>"
//   - "enter" → "<CR>"
//   - " " → "<Space>"
//   - "g g" → "gg"
//   - "ctrl+w v" → "<C-w>v"
func FormatKeyVim(key string) string {
	return formatChord(key, "", formatKeyVim)
}

func formatKeyVim(key string) string {
	p := parseKey(key)

	baseKey := p.key
	needsBrackets := p.ctrl || p.alt || p.shift
	switch strings.ToLower(baseKey) {
	case " ", "space":
		baseKey = "Space"
		needsBrackets = true
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
		if utf8.RuneCountInString(baseKey) > 1 {
			baseKey = titleCase(baseKey)
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

// FormatKeyVerbose formats keys in readable title case. Named keys are
// capitalised; a character key keeps its case, so "d" and "D" stay distinct.
// Chords join their key presses with spaces.
//
// Examples:
//   - "ctrl+x" → "Ctrl+x"
//   - "ctrl+X" → "Ctrl+X"
//   - "shift+tab" → "Shift+Tab"
//   - "enter" → "Enter"
//   - " " → "Space"
//   - "ctrl+k ctrl+s" → "Ctrl+k Ctrl+s"
func FormatKeyVerbose(key string) string {
	return formatChord(key, " ", formatKeyVerbose)
}

func formatKeyVerbose(key string) string {
	p := parseKey(key)

	baseKey := p.key
	switch strings.ToLower(baseKey) {
	case " ", "space":
		baseKey = "Space"
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
		baseKey = titleCase(baseKey)
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
