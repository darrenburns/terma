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
	// capture-pane ends lines with plain newlines; the styled-string parser
	// needs carriage returns too to start each line at column zero.
	text := strings.ReplaceAll(string(input), "\n", "\r\n")

	buf := uv.NewBuffer(*width, *height)
	uv.NewStyledString(text).Draw(screen{buf}, buf.Bounds())
	svg := t.BufferToSVG(buf, *width, *height, t.DefaultSVGOptions())
	// Each cell is its own rect at fractional coordinates; anti-aliasing their
	// shared edges draws faint seams between cells when rasterised to PNG.
	svg = strings.Replace(svg, "<style>", "<style>\n    rect { shape-rendering: crispEdges; }", 1)
	fmt.Print(svg)
}
