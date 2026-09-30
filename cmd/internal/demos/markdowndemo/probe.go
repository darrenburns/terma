package markdowndemo

import (
	"fmt"
	"strings"

	t "github.com/darrenburns/terma"
	"github.com/darrenburns/terma/cmd/internal/demokit"
)

const probeSource = "# Nested viewer\n\n[First wrapped link label with café é 👩‍💻 中文](#first)\n\n```go\n\tprintln(\"code  spaces\")\n```\n\n[Last target](#last)"

type probeDemo struct {
	state, second *t.MarkdownState
	scroll        *t.ScrollState
	modal         t.Signal[bool]
	status        t.Signal[string]
}

func newProbeDemo() *probeDemo {
	return &probeDemo{state: t.NewMarkdownState(probeSource), second: t.NewMarkdownState("## Independent viewer\n\n[Other link](#other)"), scroll: t.NewScrollState(), modal: t.NewSignal(false), status: t.NewSignal("Press Right: first viewer starts below the viewport")}
}
func (a *probeDemo) InitialFocus() string { return "probe-markdown" }
func (a *probeDemo) Keybinds() []t.Keybind {
	return []t.Keybind{
		{Key: "r", Name: "Reset", Action: func() {
			a.state.SetSource(probeSource)
			a.scroll.SetOffset(0)
			a.status.Set("Reset: select an offscreen link with Right")
			t.RequestFocus("probe-markdown")
		}},
		{Key: "2", Name: "Replace", Action: func() {
			a.state.SetSource("# Replaced\n\n[Replacement](#replacement)\n\nUnicode: é 👩‍💻 中文. Control: \x1b and invalid byte: \xff.")
			a.status.Set("Replaced: prior link/document selection cleared")
		}},
		{Key: "m", Name: "Modal", Action: func() { a.modal.Set(!a.modal.Get()); t.RequestFocus("probe-markdown") }},
		{Key: "ctrl+t", Name: "Theme", Action: demokit.NextTheme},
	}
}
func (a *probeDemo) Build(ctx t.BuildContext) t.Widget {
	theme := ctx.Theme()
	copied := func(value string) {
		a.status.Set(fmt.Sprintf("Copied %d runes: %s", len([]rune(value)), strings.Split(value, "\n")[0]))
	}
	opened := func(value string) { a.status.Set("Activated: " + value) }
	body := t.Column{Spacing: 1, Children: []t.Widget{
		t.Text{Content: "Right reveals nested links · Tab switches viewer · Ctrl+A then Y copies", Style: t.Style{ForegroundColor: theme.TextMuted}},
		t.Scrollable{State: a.scroll, Height: t.Cells(9), Width: t.Flex(1), Style: t.Style{Padding: t.EdgeInsetsAll(1), Border: t.RoundedBorder(theme.Border)}, Child: t.Column{Style: t.Style{Padding: t.EdgeInsetsAll(1)}, Children: []t.Widget{
			t.Text{Content: strings.Repeat("Content before the viewer\n", 16)},
			t.Markdown{ID: "probe-markdown", State: a.state, ScrollState: a.scroll, CodeTheme: "monokai", Style: t.Style{Padding: t.EdgeInsetsAll(1), Margin: t.EdgeInsets{Top: 1}}, OnLink: opened, OnCopy: copied},
		}}},
		t.Markdown{ID: "probe-other", State: a.second, OnLink: opened, OnCopy: copied},
		t.SignalText(a.status, func(s string) string { return s }),
		t.KeybindBar{},
	}}
	if a.modal.Get() {
		return t.Dialog{ID: "markdown-probe-dialog", Visible: true, Title: "Markdown nested probe", Style: t.Style{Width: t.Percent(94)}, Content: body, OnDismiss: func() { a.modal.Set(false) }}
	}
	return t.Column{Width: t.Flex(1), Children: []t.Widget{demokit.Header{Title: "Markdown probe", Tagline: "Nested layout and independent viewers"}, body}}
}
