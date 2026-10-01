package terma

import (
	"strings"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/require"
)

type cursorRevealArea struct {
	state  *TextAreaState
	scroll *ScrollState
}

func newCursorRevealArea() *cursorRevealArea {
	state := NewTextAreaState("START\n" + strings.Repeat("wrapped words ", 80) + "\nTARGET")
	state.CursorIndex.Set(0)
	return &cursorRevealArea{state, NewScrollState()}
}

func (a *cursorRevealArea) Build(BuildContext) Widget {
	return Scrollable{State: a.scroll, Width: Flex(1), Height: Flex(1), Child: TextArea{ID: "area", State: a.state, ScrollState: a.scroll, Width: Flex(1)}}
}

func TestTextAreaDirectCursorRevealFirstFrame(t *testing.T) {
	for _, initial := range []bool{false, true} {
		name := "move"
		if initial {
			name = "initial"
		}
		t.Run(name, func(t *testing.T) {
			a := newCursorRevealArea()
			r := NewRenderer(uv.NewBuffer(24, 5), 24, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
			if !initial {
				r.Update(a)
			}
			a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET"))
			r.Update(a)
			require.Contains(t, r.ScreenText(), "TARGET", "first rendered frame must show directly moved cursor")
			a.scroll.ScrollUp(3)
			r.Update(a)
			require.NotContains(t, r.ScreenText(), "TARGET", "wheel must stay away from unchanged cursor")
			r.Render(a)
			require.NotContains(t, r.ScreenText(), "TARGET", "forced rebuild must preserve wheel position")
			a.state.CursorIndex.Set(0)
			r.Update(a)
			require.True(t, strings.HasPrefix(r.ScreenText(), "START"), "direct move back must reveal first line immediately")
			require.Zero(t, a.scroll.GetOffset())
		})
	}
}

func TestTextAreaCursorRevealOnWrappedResize(t *testing.T) {
	a := newCursorRevealArea()
	r := NewRenderer(uv.NewBuffer(24, 5), 24, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Update(a)
	a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET"))
	r.Update(a)
	r.Resize(14, 5)
	r.Update(a)
	require.Contains(t, r.ScreenText(), "TARGET", "narrowing wrapped text must reveal cursor in the resize frame")
}

func TestReactivityTextAreaDirectCursorReveal(t *testing.T) {
	seq := newReactivitySequence(t, 24, 5, newCursorRevealArea)
	seq.frame("initial", nil)
	seq.frame("direct cursor", func(a *cursorRevealArea) { a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET")) })
	require.Contains(t, seq.actual.renderer.ScreenText(), "TARGET")
	seq.frame("wheel", func(a *cursorRevealArea) { a.scroll.ScrollUp(3) })
	require.NotContains(t, seq.actual.renderer.ScreenText(), "TARGET")
	seq.frame("return", func(a *cursorRevealArea) { a.state.CursorIndex.Set(0) })
	seq.frame("cursor again", func(a *cursorRevealArea) { a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET")) })
	require.Contains(t, seq.actual.renderer.ScreenText(), "TARGET")
}

func TestSnapshot_TextAreaInitialCursorReveal(t *testing.T) {
	a := newCursorRevealArea()
	a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET"))
	AssertSnapshot(t, a, 24, 5, "Text area initially scrolled to a directly positioned cursor on TARGET")
}

func TestTextAreaCursorRevealOnHeightShrink(t *testing.T) {
	a := newCursorRevealArea()
	r := NewRenderer(uv.NewBuffer(24, 5), 24, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Update(a)
	a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET"))
	r.Update(a)
	r.Resize(24, 3)
	r.Update(a)
	require.Contains(t, r.ScreenText(), "TARGET", "shortening viewport must retain the cursor in the first resize frame")
	a.scroll.ScrollUp(10)
	r.Update(a)
	require.NotContains(t, r.ScreenText(), "TARGET")
	r.Resize(24, 2)
	r.Update(a)
	require.NotContains(t, r.ScreenText(), "TARGET", "shrinking after wheel must preserve an intentionally hidden cursor")
}

func TestTextAreaCursorRevealSameRowAfterWheel(t *testing.T) {
	a := newCursorRevealArea()
	r := NewRenderer(uv.NewBuffer(24, 5), 24, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Update(a)
	a.state.CursorIndex.Set(1)
	r.Update(a)
	a.scroll.ScrollDown(10)
	r.Update(a)
	require.NotContains(t, r.ScreenText(), "START")
	a.state.CursorIndex.Set(2)
	r.Update(a)
	require.Contains(t, r.ScreenText(), "START", "moving within a hidden cursor row must reveal it immediately")
	require.Zero(t, a.scroll.GetOffset())
}

type nestedCursorRevealArea struct {
	*cursorRevealArea
	label Signal[string]
}

func (a *nestedCursorRevealArea) Build(BuildContext) Widget {
	return Column{Width: Flex(1), Height: Flex(1), Children: []Widget{
		SignalText(a.label, func(s string) string { return s }),
		Row{Width: Flex(1), Height: Flex(1), Children: []Widget{a.cursorRevealArea}},
	}}
}

func TestReactivityNestedTextAreaCursorReveal(t *testing.T) {
	seq := newReactivitySequence(t, 24, 6, func() *nestedCursorRevealArea {
		return &nestedCursorRevealArea{newCursorRevealArea(), NewSignal("before")}
	})
	seq.frame("initial", nil)
	seq.frame("direct cursor", func(a *nestedCursorRevealArea) { a.state.CursorIndex.Set(len(a.state.Content.Peek()) - len("TARGET")) })
	require.Contains(t, seq.actual.renderer.ScreenText(), "TARGET")
	seq.frame("sibling update", func(a *nestedCursorRevealArea) { a.label.Set("after!") })
	require.Contains(t, seq.actual.renderer.ScreenText(), "TARGET", "ancestor cache must not restore the pre-reveal offset")
	seq.frame("wheel", func(a *nestedCursorRevealArea) { a.scroll.ScrollUp(3) })
	require.NotContains(t, seq.actual.renderer.ScreenText(), "TARGET")
	seq.frame("sibling update after wheel", func(a *nestedCursorRevealArea) { a.label.Set("again!") })
	require.NotContains(t, seq.actual.renderer.ScreenText(), "TARGET")
}

func TestTextAreaVisibleCursorMovesStayPaintOnly(t *testing.T) {
	a := newCursorRevealArea()
	r := NewRenderer(uv.NewBuffer(24, 5), 24, 5, NewFocusManager(), NewAnySignal[Focusable](nil), NewAnySignal[Widget](nil))
	r.Update(a)
	for i := 0; i < 20; i++ {
		a.state.CursorIndex.Set(1 + i%2)
		r.Update(a)
		require.Zero(t, r.Stats().LayoutCount, "visible cursor moves must reuse measured geometry")
	}
	a.scroll.ScrollDown(10)
	r.Update(a)
	require.NotContains(t, r.ScreenText(), "START", "wheel must not replay the preceding cursor movement")
	r.Render(a)
	require.NotContains(t, r.ScreenText(), "START", "forced layout must not replay a cursor move preceding wheel")
	a.state.CursorIndex.Set(3)
	r.Update(a)
	require.Contains(t, r.ScreenText(), "START", "same-row movement after wheel must reveal immediately")
}
