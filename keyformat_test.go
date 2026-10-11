package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatKey_KeepsCharacterCase(t *testing.T) {
	cases := []struct {
		key, verbose, caret, emacs, vim string
	}{
		{"d", "d", "d", "d", "d"},
		{"D", "D", "D", "D", "D"},
		{"G", "G", "G", "G", "G"},
		{"ctrl+d", "Ctrl+d", "^d", "C-d", "<C-d>"},
		{"ctrl+D", "Ctrl+D", "^D", "C-D", "<C-D>"},
		{"ctrl+shift+d", "Ctrl+Shift+d", "^D", "C-S-d", "<C-S-d>"},
		{"shift+tab", "Shift+Tab", "Shift+Tab", "S-TAB", "<S-Tab>"},
		{"alt+x", "Alt+x", "Alt+x", "M-x", "<M-x>"},
		{"ctrl+alt+x", "Ctrl+Alt+x", "Ctrl+Alt+x", "C-M-x", "<C-M-x>"},
		{"enter", "Enter", "Enter", "RET", "<CR>"},
		{" ", "Space", "Space", "SPC", "<Space>"},
		{"ctrl+space", "Ctrl+Space", "Ctrl+Space", "C-SPC", "<C-Space>"},
		{"f1", "F1", "F1", "f1", "<F1>"},
		{"+", "+", "+", "+", "+"},
		{"ctrl++", "Ctrl++", "^+", "C-+", "<C-+>"},
	}
	for _, c := range cases {
		assert.Equal(t, c.verbose, FormatKeyVerbose(c.key), "verbose %q", c.key)
		assert.Equal(t, c.caret, FormatKeyCaret(c.key), "caret %q", c.key)
		assert.Equal(t, c.emacs, FormatKeyEmacs(c.key), "emacs %q", c.key)
		assert.Equal(t, c.vim, FormatKeyVim(c.key), "vim %q", c.key)
	}
}

func TestFormatKey_Chords(t *testing.T) {
	cases := []struct {
		key, def, verbose, caret, emacs, vim string
	}{
		{"ctrl+k ctrl+s", "ctrl+k ctrl+s", "Ctrl+k Ctrl+s", "^k ^s", "C-k C-s", "<C-k><C-s>"},
		{"g g", "g g", "g g", "g g", "g g", "gg"},
		{"space f", "space f", "Space f", "Space f", "SPC f", "<Space>f"},
		{"ctrl+w V", "ctrl+w V", "Ctrl+w V", "^w V", "C-w V", "<C-w>V"},
	}
	for _, c := range cases {
		assert.Equal(t, c.def, KeybindBar{}.formatKey(c.key), "default %q", c.key)
		assert.Equal(t, c.verbose, FormatKeyVerbose(c.key), "verbose %q", c.key)
		assert.Equal(t, c.caret, FormatKeyCaret(c.key), "caret %q", c.key)
		assert.Equal(t, c.emacs, FormatKeyEmacs(c.key), "emacs %q", c.key)
		assert.Equal(t, c.vim, FormatKeyVim(c.key), "vim %q", c.key)
	}
}
