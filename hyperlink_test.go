package terma

import (
	"regexp"
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

var sgrPattern = regexp.MustCompile("\x1b\\[[0-9;:]*m")

func TestParseMarkup_Link(t *testing.T) {
	theme := testTheme
	theme.Link = Hex("#00aaff")

	spans := ParseMarkup("See [link=https://example.com/a?b=C]the docs[/] now", theme)
	require.Len(t, spans, 3)
	require.Equal(t, "", spans[0].Style.Link)
	require.Equal(t, Span{Text: "the docs", Style: SpanStyle{
		Link: "https://example.com/a?b=C", Underline: UnderlineSingle, Foreground: theme.Link,
	}}, spans[1])
	require.Equal(t, SpanStyle{}, spans[2].Style, "the link ends at the closing tag")

	for _, markup := range []string{"[link=https://x.test $Error]x[/]", "[$Error link=https://x.test]x[/]"} {
		spans = ParseMarkup(markup, theme)
		require.Equal(t, theme.Error, spans[0].Style.Foreground, "%s: an explicit color wins in either order", markup)
		require.Equal(t, "https://x.test", spans[0].Style.Link)
	}

	spans = ParseMarkup("[link=https://x.test]a [b]bold[/] b[/]", theme)
	require.Len(t, spans, 3)
	for _, span := range spans {
		require.Equal(t, "https://x.test", span.Style.Link, "nested tags inherit the link: %q", span.Text)
	}
	require.True(t, spans[1].Style.Bold)

	spans = ParseMarkup("[$Link]styled[/]", theme)
	require.Equal(t, theme.Link, spans[0].Style.Foreground)
}

func TestTerminalLink(t *testing.T) {
	require.Equal(t, uv.Link{}, terminalLink(""))
	require.Equal(t, uv.Link{URL: "https://example.com/a?b=1#c"}, terminalLink("https://example.com/a?b=1#c"))
	require.Equal(t, uv.Link{URL: "https://example.com/caf%C3%A9%20menu"}, terminalLink("https://example.com/café menu"))
	for _, unsafe := range []string{"https://x.test/\x1b]8;;\x07", "https://x.test/\x07", "https://x.test/\n", "https://x.test/\x7f"} {
		require.Equal(t, uv.Link{}, terminalLink(unsafe), "%q", unsafe)
	}
}

func TestBufferToANSI_LinkBoundaries(t *testing.T) {
	widget := Text{Spans: []Span{
		PlainSpan("a "),
		LinkSpan("日本", "https://one.test"),
		LinkSpan("x", "https://two.test"),
		PlainSpan(" b"),
	}}
	out := sgrPattern.ReplaceAllString(RenderToString(widget, 12, 1), "")
	require.Equal(t,
		"a \x1b]8;;https://one.test\x07日本\x1b]8;;\x07\x1b]8;;https://two.test\x07x\x1b]8;;\x07 b",
		out,
		"each link is opened before its first cell and closed before the next link or unlinked text")
}

func TestDrawSpan_LinkCoversEveryCellOfTheSpan(t *testing.T) {
	buf := RenderToBuffer(Text{Spans: []Span{PlainSpan("a"), LinkSpan("日b", "https://x.test"), PlainSpan("c")}}, 6, 1)
	links := make([]string, 0, 5)
	for x := 0; x < 5; x++ {
		if cell := buf.CellAt(x, 0); cell != nil && cell.Width > 0 {
			links = append(links, cell.Content+"="+cell.Link.URL)
		}
	}
	require.Equal(t, []string{"a=", "日=https://x.test", "b=https://x.test", "c="}, links)
}

func TestSnapshot_TextLinks(t *testing.T) {
	theme := ThemeData{Link: Hex("#7aa2f7"), Text: Hex("#c0caf5"), Error: Hex("#f7768e")}
	markup := strings.Join([]string{
		"Read [link=https://one.test]the guide[/] first.",
		"[link=https://one.test]one[/][link=https://two.test]two[/] touch.",
		"Wide [link=https://wide.test]日本語[/] link.",
		"[link=https://red.test $Error]Custom color[/]",
	}, "\n")
	AssertSnapshot(t, Text{Spans: ParseMarkup(markup, theme)}, 30, 4,
		"Each link is underlined in the theme Link color and wrapped in its own <a href>. Adjacent links stay separate, wide glyphs keep their link, and an explicit color overrides the Link color.")
}

type reactivityLinkScene struct {
	label Signal[string]
	url   Signal[string]
}

func (s *reactivityLinkScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{
		reactivityBuilder{ID: "link", build: func(BuildContext) Widget {
			return Row{Children: []Widget{
				Text{Spans: []Span{LinkSpan(s.label.Get(), s.url.Get()), PlainSpan(" tail")}},
			}}
		}},
		Text{Content: "static line"},
	}}
}

func TestReactivityLinksRepaintWithoutBleeding(t *testing.T) {
	sequence := newReactivitySequence(t, 30, 3, func() *reactivityLinkScene {
		return &reactivityLinkScene{label: NewSignal("long link label"), url: NewSignal("https://one.test")}
	})
	sequence.frame("Initial", nil)
	sequence.frame("Same text, new URL", func(s *reactivityLinkScene) { s.url.Set("https://two.test") })
	sequence.frame("Shorter label", func(s *reactivityLinkScene) { s.label.Set("short") })
	sequence.frame("Wide label", func(s *reactivityLinkScene) { s.label.Set("日本語") })
	sequence.frame("Link removed", func(s *reactivityLinkScene) { s.url.Set("") })
	for x := 0; x < 30; x++ {
		if cell := sequence.actual.buffer.CellAt(x, 0); cell != nil {
			require.Empty(t, cell.Link.URL, "no cell keeps a removed link (column %d)", x)
		}
	}
}

func TestBorderTitleLink(t *testing.T) {
	theme := ThemeData{Link: Hex("#7aa2f7")}
	widget := Column{Width: Cells(20), Height: Cells(3), Style: Style{
		Border: RoundedBorder(theme.Link, BorderTitleMarkup("[link=https://x.test]Docs[/]")),
	}}
	buf := RenderToBuffer(widget, 20, 3)
	var linked strings.Builder
	for x := 0; x < 20; x++ {
		if cell := buf.CellAt(x, 0); cell != nil && cell.Link.URL == "https://x.test" {
			linked.WriteString(cell.Content)
		}
	}
	require.Equal(t, "Docs", linked.String())
}
