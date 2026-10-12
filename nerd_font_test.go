package terma_test

import (
	"strings"
	"testing"

	"github.com/darrenburns/terma"
	"github.com/stretchr/testify/assert"
)

func TestNerdFontsOffByDefault(t *testing.T) {
	assert.False(t, terma.NerdFontsEnabled())
	assert.Equal(t, "[dir]", terma.NerdFontSymbol("\uf07b", "[dir]"))
}

func TestNerdFontSymbol(t *testing.T) {
	t.Cleanup(func() { terma.SetNerdFontsEnabled(false) })
	cases := []struct {
		name, symbol, fallback string
	}{
		{"ASCII fallback", "\uf07b", "[dir]"},
		{"Unicode fallback", "\U000f0256", "→"},
		{"padding and multiple characters", "\uf07b \uf15b ", " dir/file "},
		{"empty fallback", "\uf07b", ""},
		{"empty symbol", "", "[dir]"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			terma.SetNerdFontsEnabled(false)
			assert.Equal(t, c.fallback, terma.NerdFontSymbol(c.symbol, c.fallback))
			terma.SetNerdFontsEnabled(true)
			assert.Equal(t, c.symbol, terma.NerdFontSymbol(c.symbol, c.fallback))
			terma.SetNerdFontsEnabled(false)
			assert.Equal(t, c.fallback, terma.NerdFontSymbol(c.symbol, c.fallback))
		})
	}
}

type nerdFontLabel struct {
	symbol, fallback string
}

func (l nerdFontLabel) Build(terma.BuildContext) terma.Widget {
	return terma.Text{Content: terma.NerdFontSymbol(l.symbol, l.fallback)}
}

func TestNerdFontsSettingUpdatesLabels(t *testing.T) {
	t.Cleanup(func() { terma.SetNerdFontsEnabled(false) })
	p := terma.NewPilot(t, terma.Row{Spacing: 1, Children: []terma.Widget{
		nerdFontLabel{symbol: "\uf07b", fallback: "[dir]"},
		nerdFontLabel{symbol: "\uf15b", fallback: "[file]"},
		terma.Text{Content: "| next"},
	}}, 30, 1)
	assert.Equal(t, "[dir] [file] | next", strings.TrimSpace(p.ScreenText()))

	terma.SetNerdFontsEnabled(true)
	assert.True(t, terma.NerdFontsEnabled())
	assert.Equal(t, "\uf07b \uf15b | next", strings.TrimSpace(p.ScreenText()))

	terma.SetNerdFontsEnabled(false)
	assert.False(t, terma.NerdFontsEnabled())
	assert.Equal(t, "[dir] [file] | next", strings.TrimSpace(p.ScreenText()))
}
