// Command ansi-to-svg renders ANSI-styled terminal text (for example the output
// of `tmux capture-pane -p -e`) to an SVG using Terma's snapshot renderer.
//
//	tmux capture-pane -p -e -t demo | go run ./cmd/ansi-to-svg -w 120 -h 40 > frame.svg
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	t "github.com/darrenburns/terma"
)

type screen struct{ *uv.Buffer }

func (screen) WidthMethod() uv.WidthMethod { return ansi.GraphemeWidth }

func main() {
	width := flag.Int("w", 80, "screen width in cells")
	height := flag.Int("h", 24, "screen height in cells")
	flag.Parse()

	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		lines[i] = expandTabs(line)
	}
	// capture-pane ends lines with plain newlines; the styled-string parser
	// needs carriage returns too to start each line at column zero.
	text := strings.Join(lines, "\r\n")

	buf := uv.NewBuffer(*width, *height)
	uv.NewStyledString(text).Draw(screen{buf}, buf.Bounds())
	svg := t.BufferToSVG(buf, *width, *height, t.DefaultSVGOptions())
	// Each cell is its own rect at fractional coordinates; anti-aliasing their
	// shared edges draws faint seams between cells when rasterised to PNG.
	svg = strings.Replace(svg, "<style>", "<style>\n    rect { shape-rendering: crispEdges; }", 1)
	fmt.Print(svg)
}

// expandTabs replaces each tab with spaces up to the next 8-column tab stop.
// tmux captures a tab when the program used one to move the cursor; the
// styled-string parser would otherwise draw it as a single cell and shift the
// rest of the line left. The spaces take the style in effect at the tab, which
// is the style the skipped cells were last painted with.
func expandTabs(line string) string {
	if !strings.Contains(line, "\t") {
		return line
	}
	var b strings.Builder
	for _, r := range line {
		if r != '\t' {
			b.WriteRune(r)
			continue
		}
		col := ansi.StringWidth(b.String())
		b.WriteString(strings.Repeat(" ", 8-col%8))
	}
	return b.String()
}
