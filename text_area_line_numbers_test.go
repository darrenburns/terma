package terma

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTextAreaLineNumbersRendering(t *testing.T) {
	for _, tc := range []struct {
		name, text, placeholder string
		wrap                    WrapMode
		width                   int
		want                    string
	}{
		{"soft wrap and blank lines", "alpha beta\n\nZ\n", "", WrapSoft, 9, "1 alpha\n  beta\n2\n3 Z\n4"},
		{"hard wrap and Unicode", "界界界界\nend", "", WrapHard, 8, "1 界界\n  界界\n2 end"},
		{"placeholder", "", "Type here", WrapSoft, 9, "1 Type\n  here"},
		{"empty", "", "", WrapNone, 9, "1"},
		{"two digit numbers", strings.Repeat("x\n", 9), "", WrapNone, 9, " 1 x\n 2 x\n 3 x\n 4 x\n 5 x\n 6 x\n 7 x\n 8 x\n 9 x\n10"},
		{"too narrow for gutter", "abc", "", WrapNone, 3, "ab"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := NewTextAreaState(tc.text)
			state.CursorIndex.Set(0)
			state.WrapMode.Set(tc.wrap)
			height := strings.Count(tc.want, "\n") + 1
			area := TextArea{ID: "area", State: state, ShowLineNumbers: true, Placeholder: tc.placeholder,
				Style: Style{Width: Flex(1), Height: Flex(1)}}
			p := NewPilot(t, area, tc.width, height)
			rows := strings.Split(p.ScreenText(), "\n")
			for i := range rows {
				rows[i] = strings.TrimRight(rows[i], " ")
			}
			require.Equal(t, tc.want, strings.Join(rows, "\n"))
		})
	}
}

func TestTextAreaLineNumbersHorizontalScrollAndSelection(t *testing.T) {
	state := NewTextAreaState("0123456789ABCDEF\nshort")
	state.CursorIndex.Set(0)
	state.WrapMode.Set(WrapNone)
	area := TextArea{ID: "area", State: state, ShowLineNumbers: true,
		LineNumberStyle: SpanStyle{Foreground: Yellow, Background: Blue},
		LineHighlights:  []LineHighlight{{StartLine: 0, EndLine: 1, Style: Style{BackgroundColor: Red}}},
		Style:           Style{Width: Flex(1), Height: Flex(1)}}
	p := NewPilot(t, area, 10, 3)
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 0123456 "))
	p.Press("end")
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 9ABCDEF "))
	x, y := state.CursorScreenPosition(20, 30)
	require.Equal(t, 29, x, "cursor coordinates include the fixed gutter")
	require.Equal(t, 30, y)
	p.ClickAt(4, 0) // B is document column 11, despite being screen column 4.
	require.Equal(t, 11, state.CursorIndex.Peek())
	drag(p, 3, 0, 6, 0)
	require.Equal(t, "ABC", state.GetSelectedText())
	require.Equal(t, Blue, bgAt(p, 0, 0), "line highlights and selection must not paint over the gutter")
	require.Equal(t, Blue, bgAt(p, 1, 0), "the gutter's gap gets its background too")
	require.Equal(t, Yellow, FromANSI(p.Buffer().CellAt(0, 0).Style.Fg))
	p.Press("home")
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 0123456 "))
}

func TestTextAreaLineNumbersWrappedNavigationAndMouse(t *testing.T) {
	state := NewTextAreaState("abcdefghij\nnext")
	state.CursorIndex.Set(0)
	state.WrapMode.Set(WrapHard)
	area := TextArea{ID: "area", State: state, ShowLineNumbers: true,
		Style: Style{Width: Flex(1), Height: Flex(1), Border: RoundedBorder(Blue), Padding: EdgeInsetsXY(1, 1)}}
	p := NewPilot(t, area, 13, 8) // 9 content cells, 2 gutter cells, 6 text cells.
	p.ClickAt(5, 2)
	require.Equal(t, 1, state.CursorIndex.Peek(), "mouse coordinates exclude border, padding and gutter")
	p.Press("down")
	require.Equal(t, 7, state.CursorIndex.Peek(), "vertical navigation uses the reduced wrap width")
	p.Advance(time.Second) // Start a drag rather than double-clicking the earlier cell.
	drag(p, 5, 2, 6, 3)
	require.Equal(t, "bcdefgh", state.GetSelectedText())
}

func TestTextAreaLineNumbersGrowShrinkAndReveal(t *testing.T) {
	state := NewTextAreaState(strings.Repeat("x\n", 8) + "last")
	state.CursorIndex.Set(0)
	scroll := NewScrollState()
	area := TextArea{ID: "area", State: state, ScrollState: scroll, ShowLineNumbers: true, Style: Style{Width: Flex(1)}}
	p := NewPilot(t, Scrollable{State: scroll, Width: Flex(1), Height: Flex(1), Child: area}, 12, 4)
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 x"))
	state.CursorIndex.Set(len(state.Content.Peek()))
	require.Contains(t, p.ScreenText(), "9 last")
	p.Press("enter")
	require.Contains(t, p.ScreenText(), " 9 last")
	require.Contains(t, p.ScreenText(), "10 ")
	p.Press("backspace")
	state.CursorIndex.Set(0)
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 x"), "removing line 10 shrinks the gutter")
	state.SetText("abcdefghij\nTARGET")
	state.CursorIndex.Set(len(state.Content.Peek()))
	require.Contains(t, p.ScreenText(), "2 TARGET")
	p.Resize(8, 3)
	require.Contains(t, p.ScreenText(), "2 TARG")
	require.Contains(t, p.ScreenText(), "  ET")
	state.SetText("")
	require.True(t, strings.HasPrefix(p.ScreenText(), "1 "), "clearing a scrolled document restores the empty first line")
}
