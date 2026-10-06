// Package hyperlinkdemo demonstrates OSC 8 terminal hyperlinks in markup,
// spans, Markdown and a scrolling list, and opening a link with OpenURL.
package hyperlinkdemo

import (
	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

// Info describes the demo for the gallery.
var Info = demokit.Info{
	Key:         "hyperlinks",
	Title:       "Hyperlinks",
	Description: "OSC 8 links in markup, spans, Markdown and lists",
}

type resource struct {
	name, url string
}

var resources = []resource{
	{"Go", "https://go.dev"},
	{"Terma on GitHub", "https://github.com/darrenburns/terma"},
	{"OSC 8 specification", "https://gist.github.com/egmontkob/eb114294efbcd5adb1944c9f3cb5feda"},
	{"Ultraviolet", "https://github.com/charmbracelet/ultraviolet"},
	{"CommonMark", "https://commonmark.org"},
	{"日本語のページ", "https://ja.wikipedia.org/wiki/日本語"},
	{"Unicode", "https://unicode.org"},
	{"xterm control sequences", "https://invisible-island.net/xterm/ctlseqs/ctlseqs.html"},
}

const markdown = "Markdown links such as [CommonMark](https://commonmark.org) are hyperlinks too. Relative ones like [notes](notes.md) stay plain."

type demo struct {
	list      *t.ListState[resource]
	scroll    *t.ScrollState
	markdown  *t.MarkdownState
	status    t.Signal[string]
	highlight t.Signal[bool]
}

// New creates the demo.
func New() demokit.Demo {
	return &demo{
		list:      t.NewListState(resources),
		scroll:    t.NewScrollState(),
		markdown:  t.NewMarkdownState(markdown),
		status:    t.NewSignal("Cmd- or Ctrl-click a link in a terminal that supports OSC 8"),
		highlight: t.NewSignal(false),
	}
}

// InitialFocus focuses the list so the arrow keys move between links.
func (d *demo) InitialFocus() string { return "hyperlink-list" }

func (d *demo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "o", Name: "Open", Action: d.open},
		{Key: "h", Name: "Highlight", Action: func() { d.highlight.Set(!d.highlight.Get()) }},
		{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme},
	}
}

func (d *demo) open() {
	item, ok := d.list.SelectedItem()
	if !ok {
		return
	}
	d.status.Set("Opening " + item.url)
	go func() {
		if err := t.OpenURL(item.url); err != nil {
			d.status.Set(err.Error())
			return
		}
		d.status.Set("Opened " + item.url)
	}()
}

func (d *demo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	// Restyling a link repaints only its cells; it must stay a link afterwards.
	first := "[link=https://go.dev]a link[/]"
	if d.highlight.Get() {
		first = "[link=https://go.dev $Background on $Warning]a link[/]"
	}
	list := t.List[resource]{
		ID:          "hyperlink-list",
		State:       d.list,
		ScrollState: d.scroll,
		Style:       t.Style{Width: t.Flex(1)},
		RenderItem: func(item resource, active, _ bool) t.Widget {
			prefix, style := "  ", t.Style{}
			if active {
				prefix, style = "▶ ", t.Style{BackgroundColor: theme.SurfaceHover}
			}
			return t.Text{Width: t.Flex(1), Style: style, Spans: []t.Span{
				t.PlainSpan(prefix),
				t.LinkSpan(item.name, item.url, theme.Link),
				t.ColorSpan("  "+item.url, theme.TextMuted),
			}}
		},
	}
	return t.Dock{
		ID:    "hyperlink-demo-root",
		Style: t.Style{BackgroundColor: theme.Background},
		Top:   []t.Widget{demokit.Header{Title: "Hyperlinks", Tagline: "OSC 8 links your terminal can open"}},
		Bottom: []t.Widget{
			demokit.Footer(theme),
			t.Row{Width: t.Flex(1), Style: t.Style{Padding: t.EdgeInsetsXY(1, 0)}, Children: []t.Widget{
				t.SignalText(d.status, func(s string) string { return s }),
			}},
		},
		Body: t.Column{Width: t.Flex(1), Height: t.Flex(1), Style: t.Style{Padding: t.EdgeInsetsXY(1, 0)}, Children: []t.Widget{
			t.Column{Style: demokit.PanelStyle(theme, "Markup", false), Width: t.Flex(1), Children: []t.Widget{
				t.Text{Wrap: t.WrapSoft, Width: t.Flex(1), Spans: t.ParseMarkup(
					"Write "+first+" in markup, or give it [link=https://go.dev/doc $Accent]its own color[/]. "+
						"Touching links ([link=https://example.com/a]one[/][link=https://example.com/b]two[/]) stay separate, "+
						"[link=https://ja.wikipedia.org/wiki/日本語]日本語[/] keeps its link across wide cells, and "+
						"[link=https://example.com/long]a long link label wraps onto the next line and is clickable on both lines[/].",
					theme)},
			}},
			t.Column{Style: demokit.PanelStyle(theme, "Markdown", false), Width: t.Flex(1), Children: []t.Widget{
				t.Markdown{State: d.markdown, DisableFocus: true},
			}},
			t.Column{
				Style:  demokit.PanelStyle(theme, "List · o opens with OpenURL", ctx.IsFocused(list)),
				Width:  t.Flex(1),
				Height: t.Flex(1),
				Children: []t.Widget{
					t.Scrollable{State: d.scroll, Style: t.Style{Width: t.Flex(1), Height: t.Flex(1)}, Child: list},
				},
			},
		}},
	}
}
