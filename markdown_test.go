package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownPlainText(t *testing.T) {
	cases := []struct{ name, source, want string }{
		{"empty", "", ""},
		{"inline", "# Hello **bold** and *italic*\n\nUse `a < b` &amp; \\*literal\\*.", "Hello bold and italic\n\nUse a < b & *literal*."},
		{"breaks", "soft\nline  \nhard\\\nbreak", "soft line\nhard\nbreak"},
		{"ordered", "3. first\n4. second", "3. first\n4. second"},
		{"empty item", "-", "• "},
		{"list-only item", "- - child", "• \n  • child"},
		{"reference", "Read [guide][ref].\n\n[ref]: ../guide.md", "Read guide."},
		{"nested", "- one\n  - two\n    - three", "• one\n  • two\n    • three"},
		{"quote", "> quote\n>\n> > nested", "│ quote\n\n│ │ nested"},
		{"unclosed", "*unfinished [link](\n\n```go\nfmt.Println(1)", "*unfinished [link](\n\nfmt.Println(1)"},
		{"code span", "` a  b ` and `` `x` ``", "a  b and `x`"},
		{"image", "![tiny **fox**](https://example.com/fox.png)", "[image: tiny fox]"},
		{"html", "<script>alert('x')</script>\n\ninline <b>bold</b>", "<script>alert('x')</script>\n\ninline <b>bold</b>"},
		{"unicode", "# café e\u0301 👩‍💻 中文", "café e\u0301 👩‍💻 中文"},
		{"controls", "hello\x1b[31mred\x07 &#27; &#x9b;", "hello�[31mred� � ›"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := NewMarkdownState(tc.source)
			assert.Equal(t, tc.source, state.Source())
			assert.Equal(t, tc.want, state.PlainText())
		})
	}
}

func TestMarkdownStateUpdates(t *testing.T) {
	state := NewMarkdownState("```go\nx := 1")
	assert.Equal(t, "x := 1", state.PlainText())
	state.Append("\n```\n\n**done**")
	assert.Equal(t, "x := 1\n\ndone", state.PlainText())
	state.SetSource("")
	assert.Empty(t, state.PlainText())
	state.Append("# Replacement")
	assert.Equal(t, "Replacement", state.PlainText())
}

func markdownKey(code rune, mod uv.KeyMod) KeyEvent {
	return KeyEvent{event: uv.KeyPressEvent{Code: code, Mod: mod}}
}

func markdownDispatchKey(view Focusable, event KeyEvent) bool {
	if provider, ok := view.(KeybindProvider); ok && matchKeybind(event, provider.Keybinds()) {
		return true
	}
	return view.OnKey(event)
}

func TestMarkdownLinkActivation(t *testing.T) {
	var got []string
	source := "[first](https://example.com/a) [relative](../guide.md) [email](mailto:hi@example.com) [anchor](#part) [unsafe](javascript:alert) [protocol relative](//evil.test) [encoded](java&#x73;cript:alert)"
	state := NewMarkdownState(source)
	widget := Markdown{State: state, OnLink: func(s string) { got = append(got, s) }}
	view := widget.Build(newTestBuildContext()).(Focusable)
	assert.False(t, markdownDispatchKey(view, markdownKey(uv.KeyEnter, 0)))
	for i := 0; i < 5; i++ {
		require.True(t, markdownDispatchKey(view, markdownKey(uv.KeyRight, 0)))
		require.True(t, markdownDispatchKey(view, markdownKey(uv.KeyEnter, 0)))
	}
	assert.Equal(t, []string{"https://example.com/a", "../guide.md", "mailto:hi@example.com", "#part", "https://example.com/a"}, got)
	require.True(t, markdownDispatchKey(view, markdownKey(uv.KeyLeft, 0)))
	require.True(t, markdownDispatchKey(view, markdownKey(uv.KeyEnter, 0)))
	assert.Equal(t, "#part", got[len(got)-1])
	assert.False(t, markdownDispatchKey(view, markdownKey(uv.KeyTab, 0)))
	assert.False(t, markdownDispatchKey(view, markdownKey(uv.KeyDown, 0)))
	state.SetSource("[replacement](https://replacement.test)")
	view = widget.Build(newTestBuildContext()).(Focusable)
	assert.False(t, markdownDispatchKey(view, markdownKey(uv.KeyEnter, 0)), "replacement clears active link")
}

func TestMarkdownUnsafeDestinations(t *testing.T) {
	for _, destination := range []string{"javascript:alert(1)", "data:text/plain,hello", "file:///tmp/a", "//evil.test", "https://example.com/\x1b", "https://a.test/ space", "\\\\server\\path", "http:", "mailto:"} {
		t.Run(destination, func(t *testing.T) { assert.False(t, markdownSafeLink(destination)) })
	}
}

func TestMarkdownCopyAndSelection(t *testing.T) {
	state := NewMarkdownState("# Title\n\n**body**")
	var copied []string
	widget := Markdown{State: state, OnCopy: func(s string) { copied = append(copied, s) }}
	view := widget.Build(newTestBuildContext()).(Focusable)
	assert.False(t, markdownDispatchKey(view, markdownKey('y', 0)))
	assert.True(t, markdownDispatchKey(view, markdownKey('a', uv.ModCtrl)))
	assert.True(t, markdownDispatchKey(view, markdownKey('y', 0)))
	assert.Equal(t, []string{"Title\n\nbody"}, copied)
	assert.True(t, markdownDispatchKey(view, markdownKey(uv.KeyEscape, 0)))
	assert.False(t, markdownDispatchKey(view, markdownKey('y', 0)))
	markdownDispatchKey(view, markdownKey('a', uv.ModCtrl))
	state.SetSource("changed")
	assert.False(t, markdownDispatchKey(view, markdownKey('y', 0)))
	withRunningApp(t)
	widget.OnCopy = nil
	view = widget.Build(newTestBuildContext()).(Focusable)
	markdownDispatchKey(view, markdownKey('a', uv.ModCtrl))
	markdownDispatchKey(view, markdownKey('y', 0))
	assert.Equal(t, []string{ansi.SetClipboard(SystemClipboard, "changed")}, takeTerminalWrites())
}

func TestMarkdownMouseLinksAfterWrapping(t *testing.T) {
	var got []string
	widget := Markdown{ID: "md", State: NewMarkdownState("before [longlabelwrapped](https://example.com) after"), Style: Style{Padding: EdgeInsetsAll(1)}, OnLink: func(s string) { got = append(got, s) }}
	view := widget.Build(newTestBuildContext()).(*markdownView)
	// Render through the public widget seam, then dispatch against the actual hit target.
	seq := newReactivitySequence(t, 12, 8, func() Markdown { return widget })
	seq.frame("initial", nil)
	target := seq.actual.renderer.WidgetAt(2, 2)
	require.NotNil(t, target)
	// Get the current rendered focusable so its wrapped geometry is available.
	rendered, ok := seq.actual.focus.Focused().(Clickable)
	require.True(t, ok)
	rendered.OnClick(MouseEvent{LocalX: 2, LocalY: 2, Button: uv.MouseLeft})
	assert.Equal(t, []string{"https://example.com"}, got)
	rendered.OnClick(MouseEvent{LocalX: 2, LocalY: 3, Button: uv.MouseLeft})
	assert.Len(t, got, 2, "continuation of the wrapped link activates the same destination")
	rendered.OnClick(MouseEvent{LocalX: 0, LocalY: 0, Button: uv.MouseLeft})
	rendered.OnClick(MouseEvent{LocalX: 10, LocalY: 3, Button: uv.MouseLeft})
	rendered.OnClick(MouseEvent{LocalX: 2, LocalY: 2, Button: uv.MouseRight})
	assert.Len(t, got, 2)
	view.owner.OnLink = nil
	assert.False(t, markdownDispatchKey(view, markdownKey(uv.KeyRight, 0)))
}

func TestMarkdownNarrowWidthsAndGraphemes(t *testing.T) {
	widget := Markdown{State: NewMarkdownState("- **e**\u0301 👩‍💻 中文 abcdefghijklmnop\n\n```\n\tcode  with  spaces\n```")}
	for _, width := range []int{0, 1, 2, 3, 5, 12, 40} {
		t.Run(string(rune('A'+width)), func(t *testing.T) { require.NotPanics(t, func() { RenderToBuffer(widget, width, 50) }) })
	}
	buffer := RenderToBuffer(Markdown{State: NewMarkdownState("**e**\u0301 👩‍💻 中文")}, 5, 5)
	assert.Equal(t, "e\u0301", buffer.CellAt(0, 0).Content)
	assert.Equal(t, "👩‍💻", buffer.CellAt(2, 0).Content)
}

func TestMarkdownDisabledAndNilState(t *testing.T) {
	assert.NotPanics(t, func() { RenderToBuffer(Markdown{}, 10, 3) })
	called := false
	widget := Markdown{State: NewMarkdownState("[link](https://example.com)"), OnLink: func(string) { called = true }, OnCopy: func(string) { called = true }}
	view := widget.Build(newTestBuildContext().WithDisabled()).(*markdownView)
	assert.False(t, view.IsFocusable())
	for _, event := range []KeyEvent{markdownKey(uv.KeyRight, 0), markdownKey(uv.KeyEnter, 0), markdownKey('a', uv.ModCtrl), markdownKey('y', 0)} {
		assert.False(t, markdownDispatchKey(view, event))
	}
	view.OnClick(MouseEvent{Button: uv.MouseLeft})
	assert.False(t, called)
}

func TestMarkdownReactivity(t *testing.T) {
	seq := newReactivitySequence(t, 28, 18, func() Markdown {
		return Markdown{ID: "document", State: NewMarkdownState("# Initial\n\n[one](https://one.test)"), OnLink: func(string) {}}
	})
	seq.frame("initial", nil)
	work := seq.frame("select first link", func(m Markdown) {
		markdownDispatchKey(m.Build(newTestBuildContext()).(Focusable), markdownKey(uv.KeyRight, 0))
	})
	assert.Zero(t, work.BuildCount, "link highlight must not rebuild document")
	work = seq.frame("select whole document", func(m Markdown) {
		markdownDispatchKey(m.Build(newTestBuildContext()).(Focusable), markdownKey('a', uv.ModCtrl))
	})
	assert.Zero(t, work.BuildCount, "selection must not rebuild document")
	seq.frame("clear selection", func(m Markdown) {
		markdownDispatchKey(m.Build(newTestBuildContext()).(Focusable), markdownKey(uv.KeyEscape, 0))
	})
	seq.frame("grow and wrap", func(m Markdown) {
		m.State.SetSource("# Changed\n\n- one\n  - nested long text wraps over lines\n\n```go\nx := 1")
	})
	seq.frame("complete fence", func(m Markdown) { m.State.Append("\n```\n\n[replacement](https://two.test)") })
	seq.frame("shrink", func(m Markdown) { m.State.SetSource("short") })
	seq.frame("empty", func(m Markdown) { m.State.SetSource("") })
	seq.frame("restore", func(m Markdown) { m.State.SetSource("[new](https://new.test)") })
}

func TestMarkdownCodeThemeDoesNotChangeText(t *testing.T) {
	source := "```go\nfunc main() {\n\tprintln(\"hi\")\n}\n```"
	a := RenderToBuffer(Markdown{State: NewMarkdownState(source)}, 40, 10)
	b := RenderToBuffer(Markdown{State: NewMarkdownState(source), CodeTheme: "monokai"}, 40, 10)
	c := RenderToBuffer(Markdown{State: NewMarkdownState(source), CodeTheme: "missing"}, 40, 10)
	for y := 0; y < 10; y++ {
		for x := 0; x < 40; x++ {
			assert.Equal(t, a.CellAt(x, y).Content, b.CellAt(x, y).Content)
			assert.Equal(t, a.CellAt(x, y).Content, c.CellAt(x, y).Content)
		}
	}
}

func FuzzMarkdownRender(f *testing.F) {
	for _, source := range []string{"", "# hello", "- a\n  - b", "```go\na", "[x](javascript:alert)", "👩‍💻e\u0301\x1b[2J"} {
		f.Add(source, uint8(20))
	}
	f.Fuzz(func(t *testing.T, source string, width uint8) {
		if len(source) > 8192 {
			t.Skip()
		}
		state := NewMarkdownState(source)
		require.NotContains(t, state.PlainText(), "\x1b")
		RenderToBuffer(Markdown{State: state}, int(width%80)+1, 20)
	})
}

func TestMarkdownMouseRoutingAndSourceReplacement(t *testing.T) {
	var got []string
	state := NewMarkdownState("prefix [link](https://one.test)")
	widget := Markdown{ID: "md", State: state, Style: Style{Padding: EdgeInsetsAll(1)}, OnLink: func(s string) { got = append(got, s) }}
	scene := newClickScene(t, widget, 20, 8)
	scene.click(9, 1, 0)
	require.Equal(t, []string{"https://one.test"}, got)
	state.SetSource("[new](https://two.test)")
	scene.draw()
	scene.click(2, 1, 0)
	require.Equal(t, []string{"https://one.test", "https://two.test"}, got)
	scene.click(9, 1, 0)
	require.Len(t, got, 2, "the old link target disappeared")
}

func TestMarkdownKeyboardRevealsLinks(t *testing.T) {
	state := NewMarkdownState("[first](https://one.test)\n\n" + strings.Repeat("paragraph\n\n", 20) + "[last](https://two.test)")
	scroll := NewScrollState()
	widget := Scrollable{State: scroll, Height: Cells(6), Child: Markdown{ID: "md", State: state, ScrollState: scroll, OnLink: func(string) {}}}
	scene := newClickScene(t, widget, 30, 6)
	require.True(t, scene.focus.HandleKey(markdownKey(uv.KeyRight, 0)))
	scene.draw()
	assert.Zero(t, scroll.GetOffset())
	require.True(t, scene.focus.HandleKey(markdownKey(uv.KeyRight, 0)))
	scene.draw()
	assert.Positive(t, scroll.GetOffset(), "keyboard selection reveals an off-screen link")
	require.True(t, scene.focus.HandleKey(markdownKey(uv.KeyLeft, 0)))
	scene.draw()
	assert.Zero(t, scroll.GetOffset(), "previous link reveals the top again")
}

func TestMarkdownThematicRuleStaysOneLine(t *testing.T) {
	for _, width := range []int{1, 3, 20} {
		buffer := RenderToBuffer(Markdown{State: NewMarkdownState("---\n\nafter")}, width, 10)
		assert.Equal(t, "─", buffer.CellAt(0, 0).Content)
		assert.Equal(t, "a", buffer.CellAt(0, 2).Content, "rule must not wrap into multiple rows")
	}
}

type markdownKeybindScene struct{ state *MarkdownState }

func (s *markdownKeybindScene) Build(BuildContext) Widget {
	return Column{Children: []Widget{Markdown{ID: "md", State: s.state, OnLink: func(string) {}}, KeybindBar{}}}
}
func TestMarkdownKeybindBarFollowsSource(t *testing.T) {
	seq := newReactivitySequence(t, 70, 10, func() *markdownKeybindScene {
		return &markdownKeybindScene{NewMarkdownState("[link](https://example.com)")}
	})
	seq.frame("links present", nil)
	seq.frame("links removed", func(s *markdownKeybindScene) { s.state.SetSource("plain") })
	seq.frame("links restored", func(s *markdownKeybindScene) { s.state.SetSource("[new](https://new.test)") })
	seq.frame("empty", func(s *markdownKeybindScene) { s.state.SetSource("") })
}
