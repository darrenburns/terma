package terma

import (
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/charmbracelet/x/ansi"
)

// Syntax highlighting is optional and uses the Chroma dependency already used
// by Terma demos. Unknown languages and styles retain readable plain code.
func markdownCodeRuns(block markdownBlock, themeName string) []markdownRun {
	if !block.code || themeName == "" || block.language == "" {
		return block.runs
	}
	theme, ok := styles.Registry[themeName]
	if !ok {
		return block.runs
	}
	lexer := lexers.Get(block.language)
	if lexer == nil {
		return block.runs
	}
	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, block.runs[0].text)
	if err != nil {
		return block.runs
	}
	var result []markdownRun
	for token := iterator(); token != chroma.EOF; token = iterator() {
		entry := theme.Get(token.Type)
		style := SpanStyle{Bold: entry.Bold == chroma.Yes, Italic: entry.Italic == chroma.Yes}
		if entry.Colour.IsSet() {
			style.Foreground = RGB(entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue())
		}
		result = append(result, markdownRun{text: token.Value, style: style, link: -1, code: true})
	}
	return result
}

func (v *markdownView) blockGlyphs(block markdownBlock) []markdownGlyph {
	runs := block.runs
	var content strings.Builder
	for _, run := range runs {
		content.WriteString(run.text)
	}
	remaining := content.String()
	var glyphs []markdownGlyph
	runIndex, runOffset := 0, 0
	for remaining != "" {
		for runIndex < len(runs) && runOffset >= len(runs[runIndex].text) {
			runOffset -= len(runs[runIndex].text)
			runIndex++
		}
		if runIndex == len(runs) {
			break
		}
		run := runs[runIndex]
		cluster, width := ansi.FirstGraphemeCluster(remaining, ansi.GraphemeWidth)
		remaining = remaining[len(cluster):]
		runOffset += len(cluster)
		style := run.style
		if run.code && !block.code {
			style.Foreground = v.theme.Accent
			style.Background = v.theme.Surface
		}
		if run.link >= 0 && v.owner.OnLink != nil {
			style.Foreground = v.theme.Link
			style.Underline = UnderlineSingle
		}
		glyphs = append(glyphs, markdownGlyph{text: cluster, width: width, style: style, link: run.link})
	}
	return glyphs
}

func (v *markdownView) wrap(width int) []markdownLine {
	if width <= 0 {
		return nil
	}
	var lines []markdownLine
	for blockIndex, block := range v.blocks {
		if blockIndex > 0 && block.gap {
			lines = append(lines, markdownLine{})
		}
		var line markdownLine
		newLine := func(prefix string) {
			line = markdownLine{code: block.code}
			// Prefer readable content to decoration when indentation fills the viewport.
			if ansi.StringWidth(prefix) >= width {
				prefix = ""
			}
			for prefix != "" {
				cluster, w := ansi.FirstGraphemeCluster(prefix, ansi.GraphemeWidth)
				prefix = prefix[len(cluster):]
				line.glyphs = append(line.glyphs, markdownGlyph{text: cluster, width: w, style: SpanStyle{Foreground: v.theme.TextMuted}, link: -1})
				line.width += w
			}
		}
		newLine(block.firstPrefix)
		prefixWidth := line.width
		finish := func() { lines = append(lines, line); newLine(block.nextPrefix); prefixWidth = line.width }
		appendGlyph := func(g markdownGlyph) {
			if g.text == "\t" { // Fixed four-column tab stops, relative to content.
				g.text = strings.Repeat(" ", 4-(line.width-prefixWidth)%4)
				g.width = len(g.text)
				for _, r := range g.text {
					if line.width >= width {
						finish()
					}
					line.glyphs = append(line.glyphs, markdownGlyph{text: string(r), width: 1, style: g.style, link: g.link})
					line.width++
				}
				return
			}
			if g.width == 0 {
				return
			}
			if line.width+g.width > width && line.width > prefixWidth {
				finish()
			}
			if line.width+g.width > width { // e.g. a 2-cell grapheme in a 1-cell viewport.
				return
			}
			line.glyphs = append(line.glyphs, g)
			line.width += g.width
		}
		if block.rule {
			ruleWidth := width
			// Intrinsic measurement may use an effectively unbounded sentinel.
			// Give a rule a small natural width instead of allocating that size.
			if ruleWidth > 100_000 {
				ruleWidth = max(line.width+8, 8)
			}
			for line.width < ruleWidth {
				line.glyphs = append(line.glyphs, markdownGlyph{text: "─", width: 1, style: SpanStyle{Foreground: v.theme.TextMuted}, link: -1})
				line.width++
			}
			lines = append(lines, line)
			continue
		}
		glyphs := v.blockGlyphs(block)
		if block.code {
			for _, g := range glyphs {
				if g.text == "\n" {
					finish()
				} else {
					appendGlyph(g)
				}
			}
		} else {
			var spaces []markdownGlyph
			for i := 0; i < len(glyphs); {
				if glyphs[i].text == "\n" {
					spaces = nil
					finish()
					i++
					continue
				}
				if glyphs[i].text == " " || glyphs[i].text == "\t" {
					spaces = append(spaces, glyphs[i])
					i++
					continue
				}
				end, wordWidth := i, 0
				for end < len(glyphs) && glyphs[end].text != " " && glyphs[end].text != "\t" && glyphs[end].text != "\n" {
					wordWidth += glyphs[end].width
					end++
				}
				spaceWidth := 0
				for _, g := range spaces {
					if g.text == "\t" {
						spaceWidth += 4 - (line.width-prefixWidth+spaceWidth)%4
					} else {
						spaceWidth += g.width
					}
				}
				if line.width > prefixWidth && line.width+spaceWidth+wordWidth > width {
					finish()
					spaces = nil
				}
				for _, g := range spaces {
					appendGlyph(g)
				}
				spaces = nil
				for ; i < end; i++ {
					appendGlyph(glyphs[i])
				}
			}
		}
		lines = append(lines, line)
	}
	return lines
}
