package terma

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	uv "github.com/charmbracelet/ultraviolet"
)

var keyNames sync.Map // name -> rune

// namedKeyCode looks up a key name ("enter", "escape", "pgdown", "f5") the
// way Keybind.Key spells it. It asks ultraviolet which key code the name
// matches, so the two can't disagree.
func namedKeyCode(name string) (rune, bool) {
	if code, ok := keyNames.Load(name); ok {
		return code.(rune), true
	}
	// Main keys come before their keypad twins, so "enter" finds KeyEnter.
	for _, span := range [][2]rune{{0, 0x80}, {uv.KeyExtended + 1, uv.KeyExtended + 512}} {
		for code := span[0]; code < span[1]; code++ {
			if (uv.Key{Code: code}).MatchString(name) {
				keyNames.Store(name, code)
				return code, true
			}
		}
	}
	return 0, false
}

var keyModNames = map[string]uv.KeyMod{
	"ctrl":  uv.ModCtrl,
	"alt":   uv.ModAlt,
	"shift": uv.ModShift,
	"meta":  uv.ModMeta,
	"super": uv.ModSuper,
	"hyper": uv.ModHyper,
}

// keyPress turns a key as Keybind.Key spells it ("tab", "ctrl+s",
// "shift+down", "?", "+") into the key press a terminal would report.
func keyPress(s string) (uv.KeyPressEvent, error) {
	var mod uv.KeyMod
	name := s
	// The key itself may be "+", as in "ctrl++".
	for len(name) > 1 {
		i := strings.Index(name, "+")
		if i <= 0 || i == len(name)-1 {
			break
		}
		m, ok := keyModNames[name[:i]]
		if !ok {
			return uv.KeyPressEvent{}, fmt.Errorf("unknown modifier %q in key %q", name[:i], s)
		}
		mod |= m
		name = name[i+1:]
	}

	if code, ok := namedKeyCode(name); ok && utf8.RuneCountInString(name) > 1 {
		key := uv.Key{Code: code, Mod: mod}
		if code == uv.KeySpace && mod&^uv.ModShift == 0 {
			key.Text = " "
		}
		return uv.KeyPressEvent(key), nil
	}
	if utf8.RuneCountInString(name) != 1 {
		return uv.KeyPressEvent{}, fmt.Errorf("unknown key %q", s)
	}
	r, _ := utf8.DecodeRuneInString(name)
	key := uv.Key{Code: r, Mod: mod}
	// A printable key with no ctrl or alt types its character.
	if mod&^uv.ModShift == 0 && unicode.IsPrint(r) {
		key.Text = name
		if mod&uv.ModShift != 0 {
			key.Text = string(unicode.ToUpper(r))
			key.ShiftedCode = unicode.ToUpper(r)
		}
	}
	return uv.KeyPressEvent(key), nil
}

// typedKey is the key press for typing r.
func typedKey(r rune) uv.KeyPressEvent {
	if r == ' ' {
		return uv.KeyPressEvent{Code: uv.KeySpace, Text: " "}
	}
	if r == '\n' {
		return uv.KeyPressEvent{Code: uv.KeyEnter}
	}
	if r == '\t' {
		return uv.KeyPressEvent{Code: uv.KeyTab}
	}
	return uv.KeyPressEvent{Code: r, Text: string(r)}
}
