package terma

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMarkdownProbeCodeThemePreservesFollowingParagraph(t *testing.T) {
	source := "```go\nx := 1\n```\n\nafter"
	plain := RenderToBuffer(Markdown{State: NewMarkdownState(source)}, 20, 8)
	highlighted := RenderToBuffer(Markdown{State: NewMarkdownState(source), CodeTheme: "monokai"}, 20, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 20; x++ {
			require.Equal(t, plain.CellAt(x, y).Content, highlighted.CellAt(x, y).Content, "highlighting must preserve content geometry at %d,%d", x, y)
		}
	}
	require.Equal(t, "a", highlighted.CellAt(0, 2).Content, "one paragraph gap follows the code block")
}

func TestMarkdownProbeNestedScrollableRevealsLink(t *testing.T) {
	scroll := NewScrollState()
	var opened []string
	widget := Scrollable{State: scroll, Height: Cells(1), Child: Column{Children: []Widget{
		Text{Content: "before"},
		Markdown{ID: "nested-markdown", State: NewMarkdownState("[target](#target)"), ScrollState: scroll, OnLink: func(s string) { opened = append(opened, s) }},
	}}}
	p := NewPilot(t, widget, 24, 1)
	require.Equal(t, "nested-markdown", p.FocusedID())
	p.Press("right")
	require.Contains(t, p.ScreenText(), "target", "keyboard-selected link must be brought into the containing viewport")
	p.Press("enter")
	require.Equal(t, []string{"#target"}, opened)
}

func TestMarkdownProbeRevealThroughPaddedNestedViewports(t *testing.T) {
	outer, inner := NewScrollState(), NewScrollState()
	var opened []string
	source := strings.Repeat("ordinary paragraph\n\n", 20) + "[last target](#last)"
	md := Markdown{ID: "deep-markdown", State: NewMarkdownState(source), ScrollState: inner,
		Style:  Style{Padding: EdgeInsetsAll(1), Margin: EdgeInsets{Top: 2}},
		OnLink: func(s string) { opened = append(opened, s) }}
	widget := Scrollable{State: outer, Height: Cells(20), Style: Style{Padding: EdgeInsetsAll(1)}, Child: Column{
		Children: []Widget{Text{Content: "Outer heading"}, Scrollable{State: inner, Height: Cells(9),
			Style: Style{Padding: EdgeInsetsAll(1), Border: RoundedBorder(Blue)}, Child: Column{
				Style: Style{Padding: EdgeInsetsAll(1)}, Children: []Widget{Text{Content: "before\nanother"}, DisabledWhen(false, md)},
			}},
		},
	}}
	p := NewPilot(t, widget, 42, 24)
	require.Equal(t, "deep-markdown", p.FocusedID())
	p.Press("right")
	require.Contains(t, p.ScreenText(), "last target")
	require.Zero(t, outer.GetOffset(), "only the containing viewport needs to move")
	require.Positive(t, inner.GetOffset())
	p.Press("enter")
	require.Equal(t, []string{"#last"}, opened)
	p.AssertSnapshot("nested_reveal", "A link that began fully off-screen is revealed through preceding siblings, padding, margin and a nested viewport")
}

type markdownProbeRetainedApp struct {
	first, second *MarkdownState
	style         Signal[int]
}

func (a *markdownProbeRetainedApp) Build(ctx BuildContext) Widget {
	return Column{Style: Style{Padding: EdgeInsetsAll(a.style.Get())}, Children: []Widget{
		Markdown{ID: "probe-first", State: a.first, CodeTheme: "monokai", OnLink: func(string) {}},
		Markdown{ID: "probe-second", State: a.second, Style: Style{BackgroundColor: ctx.Theme().Surface}, OnLink: func(string) {}},
		Button{ID: "after", Label: "After"},
	}}
}

func TestMarkdownProbeRetainedDocumentStyleAndResize(t *testing.T) {
	seq := newReactivitySequence(t, 46, 22, func() *markdownProbeRetainedApp {
		return &markdownProbeRetainedApp{
			first: NewMarkdownState("# One\n\n[first](#first)"), second: NewMarkdownState("# Two\n\n[second](#second)"), style: NewSignal(0),
		}
	})
	seq.frame("initial", nil)
	seq.frame("style padding", func(a *markdownProbeRetainedApp) { a.style.Set(2) })
	seq.focus("probe-second")
	seq.frame("second focused", nil)
	for i, source := range []string{"```go\nfunc main() {\n\tprintln(1)\n}\n```\n\nafter", "> - **e**́ 👩‍💻 中文 [target](#unicode)", "\xff\xfe\x00 &#27; [inert](javascript:alert)", "", "# Restored"} {
		seq.frame(fmt.Sprintf("document %d", i), func(a *markdownProbeRetainedApp) { a.first.SetSource(source) })
	}
	for _, size := range [][2]int{{13, 11}, {2, 4}, {60, 30}, {25, 8}} {
		seq.resize(size[0], size[1])
		seq.frame(fmt.Sprintf("resize %dx%d", size[0], size[1]), nil)
	}
	seq.frame("style reset", func(a *markdownProbeRetainedApp) { a.style.Set(0) })
}

func TestMarkdownProbeIndependentSelectionCopyAndLinks(t *testing.T) {
	var copies, opened []string
	first, second := NewMarkdownState("# First\n\n[same](#one) [same](#two)"), NewMarkdownState("# 第二\n\n[other](#three)")
	widget := Column{Children: []Widget{
		Markdown{ID: "first", State: first, OnCopy: func(s string) { copies = append(copies, s) }, OnLink: func(s string) { opened = append(opened, s) }},
		Markdown{ID: "second", State: second, OnCopy: func(s string) { copies = append(copies, s) }, OnLink: func(s string) { opened = append(opened, s) }},
	}}
	p := NewPilot(t, widget, 32, 12)
	p.Press("ctrl+a", "y")
	require.Equal(t, []string{"First\n\nsame same"}, copies)
	p.Press("right", "right", "enter")
	require.Equal(t, []string{"#two"}, opened)
	p.Press("tab")
	require.Equal(t, "second", p.FocusedID())
	p.Press("y")
	require.Equal(t, []string{"First\n\nsame same"}, copies, "selection must remain instance-local")
	p.Press("ctrl+a", "y")
	require.Equal(t, []string{"First\n\nsame same", "第二\n\nother"}, copies)
	second.SetSource("new")
	p.settle()
	p.Press("y")
	require.Equal(t, []string{"First\n\nsame same", "第二\n\nother"}, copies, "source replacement clears only its document selection")
	p.AssertSnapshot("multiple_instances", "Two independent viewers retain separate sources, focus and document-selection state")
}

func TestMarkdownProbeBoundedInteractionModel(t *testing.T) {
	sources := []string{"# A\n\n[first](#a) [second](#b)", "# B\n\n[third](#c)"}
	texts := []string{"A\n\nfirst second", "B\n\nthird"}
	destinations := [][]string{{"#a", "#b"}, {"#c"}}
	var copies, opened []string
	state := NewMarkdownState(sources[0])
	p := NewPilot(t, Markdown{ID: "model", State: state, OnCopy: func(s string) { copies = append(copies, s) }, OnLink: func(s string) { opened = append(opened, s) }}, 24, 10)
	source, plain, index, active, selected := sources[0], texts[0], 0, -1, false
	rng := rand.New(rand.NewSource(20260930))
	var wantCopies, wantOpened []string
	for step := 0; step < 160; step++ {
		switch rng.Intn(8) {
		case 0:
			p.Press("ctrl+a")
			selected = true
		case 1:
			p.Press("y")
			if selected {
				wantCopies = append(wantCopies, plain)
			}
		case 2:
			p.Press("escape")
			selected = false
		case 3:
			p.Press("right")
			active = (active + 1) % len(destinations[index])
			selected = false
		case 4:
			p.Press("left")
			if active < 0 {
				active = 0
			}
			active = (active - 1 + len(destinations[index])) % len(destinations[index])
			selected = false
		case 5:
			p.Press("enter")
			if active >= 0 {
				wantOpened = append(wantOpened, destinations[index][active])
			}
		case 6:
			index = 1 - index
			source, plain = sources[index], texts[index]
			state.SetSource(source)
			active, selected = -1, false
		case 7:
			source += "\n\nend"
			plain += "\n\nend"
			state.Append("\n\nend")
			active, selected = -1, false
		}
		p.settle()
		require.Equal(t, source, state.Source(), "step %d source", step)
		require.Equal(t, plain, state.PlainText(), "step %d text", step)
		require.Equal(t, wantCopies, copies, "step %d copy callbacks", step)
		require.Equal(t, wantOpened, opened, "step %d link callbacks", step)
	}
}

func TestMarkdownProbeAdversarialText(t *testing.T) {
	cases := []struct{ source, want string }{
		{"\xff\xfe\x00", "���"},
		{"```unknown\n\t  spaced  \n```", "\t  spaced  "},
		{"> > - **e**́ 👩‍💻 中文", "│ │ • é 👩‍💻 中文"},
		{"``a\nb``", "a b"},
		{"[a](https://example.com/" + strings.Repeat("x", 2048) + ")", "a"},
		{strings.Repeat("> ", 64) + "deep", strings.Repeat("│ ", 64) + "deep"},
	}
	for i, tc := range cases {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			state := NewMarkdownState(tc.source)
			require.Equal(t, tc.want, state.PlainText())
			for _, width := range []int{0, 1, 2, 9, 80} {
				require.NotPanics(t, func() { RenderToBuffer(Markdown{State: state, CodeTheme: "monokai"}, width, 12) })
			}
		})
	}
}

func TestMarkdownProbeUnicodeLinkKeepsBorderCellsAligned(t *testing.T) {
	scroll := NewScrollState()
	widget := Scrollable{State: scroll, Width: Cells(28), Height: Cells(3), Style: Style{Border: RoundedBorder(Blue)}, Child: Markdown{
		ID: "unicode-border", State: NewMarkdownState("[café é 👩‍💻 中文](#unicode)"), ScrollState: scroll, OnLink: func(string) {},
	}}
	p := NewPilot(t, widget, 40, 8)
	p.Press("right")
	buf := p.Buffer()
	borderX := -1
	for x := 0; x < 40; x++ {
		if buf.CellAt(x, 0).Content == "╮" {
			borderX = x
			break
		}
	}
	require.Positive(t, borderX)
	for y := 1; y < 4; y++ {
		require.Equal(t, "│", buf.CellAt(borderX, y).Content, "right border row %d", y)
		for x := borderX + 1; x < 40; x++ {
			require.Empty(t, strings.TrimSpace(buf.CellAt(x, y).Content), "outside border %d,%d", x, y)
		}
	}
	control := widget
	control.Child = Text{Content: "café é 👩‍💻 中文"}
	controlBuffer := RenderToBuffer(control, 40, 8)
	for y := 0; y < 8; y++ {
		for x := 0; x < 40; x++ {
			require.Equal(t, controlBuffer.CellAt(x, y).Content, buf.CellAt(x, y).Content, "plain Text control at %d,%d", x, y)
		}
	}
	p.AssertSnapshot("unicode_border", "The cell buffer keeps a single aligned right border and blank exterior beside a selected ZWJ/CJK link")
}
