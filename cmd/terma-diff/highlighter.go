package main

import t "terma"

// DiffHighlighter applies token-based highlighting over rendered diff text.
type DiffHighlighter struct {
	Tokens  []HighlightToken
	Palette ThemePalette
}

func (h DiffHighlighter) Highlight(_ string, graphemes []string) []t.TextHighlight {
	if len(h.Tokens) == 0 || len(graphemes) == 0 {
		return nil
	}

	byteToGrapheme := make(map[int]int, len(graphemes)+1)
	bytePos := 0
	byteToGrapheme[0] = 0
	for i, g := range graphemes {
		bytePos += len(g)
		byteToGrapheme[bytePos] = i + 1
	}

	highlights := make([]t.TextHighlight, 0, len(h.Tokens))
	for _, tok := range h.Tokens {
		start, okStart := byteToGrapheme[tok.StartByte]
		end, okEnd := byteToGrapheme[tok.EndByte]
		if !okStart || !okEnd || end <= start {
			continue
		}
		style, ok := h.Palette.StyleForRole(tok.Role)
		if !ok {
			continue
		}
		highlights = append(highlights, t.TextHighlight{
			Start: start,
			End:   end,
			Style: style,
		})
	}

	return highlights
}
