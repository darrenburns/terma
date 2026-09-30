// Command markdown-demo demonstrates rendered Markdown, safe link callbacks,
// whole-document copy, dynamic source replacement and incomplete fenced code.
package main

import (
	"flag"
	"fmt"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

const document = "# Terma reading room\n\nA **bold** introduction with *emphasis* and `inline code`. Read the [Terma documentation guide](https://example.com/guide) or [local notes](notes.md).\n\n## Lists and quotations\n\n3. Preserve ordered starting numbers.\n4. A longer item demonstrates hanging indentation when the window gets narrow.\n   - Nested detail with Unicode: café é 👩‍💻 中文.\n   - Another useful detail.\n\n> A quotation wraps with a marker on every line.\n>\n> > Nested quotations keep their context.\n\n## Code\n\n```go\nfunc main() {\n\tprintln(\"Hello, terminal\")\n}\n```\n\n## Safe content\n\n[Unsafe link](javascript:alert) stays inert. <b>HTML is literal</b>.\n\n![A fox](fox.png) is represented by its alt text.\n\n---\n\n[Final link](#finish) is reachable by keyboard even when below the viewport."

type demo struct {
	state  *t.MarkdownState
	scroll *t.ScrollState
	status t.Signal[string]
}

func (a *demo) InitialFocus() string { return "markdown" }
func (a *demo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "1", Name: "Document", Action: func() { a.state.SetSource(document); a.scroll.SetOffset(0); a.status.Set("Full document restored") }},
		{Key: "2", Name: "Open fence", Action: func() {
			a.state.SetSource("# Streaming example\n\nAn unfinished *emphasis and [link](\n\n```go\nlong_identifier_without_spaces := 123456789\n\tprintln(\"waiting\")")
			a.scroll.SetOffset(0)
			a.status.Set("Fence is incomplete; press 3 to append its end")
		}},
		{Key: "3", Name: "Append", Action: func() {
			a.state.Append("\n```\n\n**Finished** and [new link](https://example.com/complete).")
			a.status.Set("Appended closing fence and a new paragraph")
		}},
		{Key: "4", Name: "Empty", Action: func() { a.state.SetSource(""); a.status.Set("Empty document") }},
		{Key: "5", Name: "Controls", Action: func() {
			a.state.SetSource("# Safe display\n\nESC: \x1b[31m Bell: \x07 Entity: &#27;\n\n<script>alert('literal')</script>\n\n[Unsafe](javascript:alert) and [safe](https://example.com/safe).")
			a.scroll.SetOffset(0)
			a.status.Set("Control characters are replaced; HTML remains literal")
		}},
		{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme},
	}
}

func (a *demo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	return t.Column{Height: t.Flex(1), Width: t.Flex(1), Children: []t.Widget{
		demokit.Header{Title: "Markdown", Tagline: "Read, navigate and copy"},
		t.Text{Content: "1 document · 2 open fence · 3 append · 4 empty · 5 controls", Style: t.Style{ForegroundColor: theme.TextMuted}},
		t.Scrollable{State: a.scroll, Height: t.Flex(1), Width: t.Flex(1), Child: t.Markdown{
			ID: "markdown", State: a.state, ScrollState: a.scroll, CodeTheme: "monokai",
			Style:  t.Style{Width: t.Flex(1), Padding: t.EdgeInsetsTRBL(1, 2, 1, 2)},
			OnLink: func(destination string) { a.status.Set("Activated: " + destination) },
			OnCopy: func(text string) {
				a.status.Set(fmt.Sprintf("Copied %d characters: %s", len([]rune(text)), strings.Split(text, "\n")[0]))
			},
		}},
		t.SignalText(a.status, func(value string) string { return value }),
		t.KeybindBar{Style: t.Style{BackgroundColor: theme.Surface}},
	}}
}

func main() {
	probe := flag.Bool("probe", false, "Run nested-layout and independent-viewer verification")
	flag.Parse()
	if *probe {
		demokit.Run(newProbeDemo())
		return
	}
	demokit.Run(&demo{state: t.NewMarkdownState(document), scroll: t.NewScrollState(), status: t.NewSignal("Left/Right chooses links · Enter or mouse activates · Ctrl+A then Y copies")})
}
