package terma

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestElideStart(t *testing.T) {
	assert.Equal(t, "short", elideStart("short", 10))
	assert.Equal(t, "…/models/base.en", elideStart("/Users/darren/models/base.en", 16))
	assert.Equal(t, "…", elideStart("abc", 1))
	assert.Equal(t, "", elideStart("abc", 0))
	assert.Equal(t, "…界", elideStart("世界", 3), "wide graphemes are kept whole")
}

func TestTruncateSpans(t *testing.T) {
	spans := []Span{{Text: "ab"}, {Text: "cde"}, {Text: "f"}}
	assert.Equal(t, []Span{{Text: "ab"}, {Text: "c"}}, truncateSpans(spans, 3))
}

func paletteForRows(items []CommandPaletteItem) CommandPalette {
	state := NewCommandPaletteState("Models", items)
	state.Visible.Set(true)
	return CommandPalette{
		ID:       "palette-rows",
		State:    state,
		Style:    Style{Width: Cells(40)},
		Position: FloatPositionTopLeft,
		Offset:   Offset{X: 1, Y: 1},
	}
}

func TestSnapshot_CommandPalette_LongHintElidedFromStart(t *testing.T) {
	palette := paletteForRows([]CommandPaletteItem{
		{Label: "sample_base.en", Hint: "/Users/darren/code/whisper/models/sample_base.en"},
		{Label: "tiny", Hint: "/Users/darren/code/whisper/models/tiny"},
		{Label: "Open Recent", Hint: "Ctrl+R"},
	})
	AssertSnapshot(t, palette, 44, 9,
		"Labels show in full. The long path hints keep their ends ('…models/sample_base.en', '…darren/code/whisper/models/tiny'), right-aligned, 2 cells after the label; 'Ctrl+R' fits whole.")
}

func TestSnapshot_CommandPalette_LabelWiderThanRow(t *testing.T) {
	palette := paletteForRows([]CommandPaletteItem{
		{Label: "A label far too long to fit in the palette row at all", Hint: "Ctrl+L"},
		{Label: "Nested with a long label that runs to the edge", Children: func() []CommandPaletteItem { return nil }},
	})
	AssertSnapshot(t, palette, 44, 8,
		"The first label is cut with an ellipsis and its hint is dropped; the second is cut before the ▸ nested indicator, which stays at the right edge.")
}

func TestSnapshot_CommandPalette_CurrentItemMarked(t *testing.T) {
	palette := paletteForRows([]CommandPaletteItem{
		{Label: "Dark"},
		{Label: "Dracula", Hint: "ctrl+d", Current: true},
		{Label: "Light"},
	})
	RenderToBuffer(palette, 44, 8)
	palette.moveCursor(1) // Off the current item, onto "Light".
	AssertSnapshot(t, palette, 44, 8,
		"The cursor is on 'Light'; 'Dracula', the current item, still shows an accent-colored check at the right edge, after its hint.")
}
