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
	return &cursorRevealArea{NewTextAreaState(strings.Repeat("wrapped words ", 80) + "\nTARGET"), NewScrollState()}
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
			require.True(t, strings.HasPrefix(r.ScreenText(), "wrapped words"), "direct move back must reveal first line immediately")
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
