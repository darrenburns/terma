package terma

import (
	"fmt"
	"strings"
	"testing"
	"time"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func enableCursorBlinkForTest(t testing.TB) {
	t.Helper()
	SetCursorBlink(true)
	t.Cleanup(func() { SetCursorBlink(false) })
}

func TestCursorBlink_OffByDefault(t *testing.T) {
	assert.False(t, CursorBlink())
}

func TestSnapshot_TextInput_CursorBlinkShown(t *testing.T) {
	enableCursorBlinkForTest(t)
	widget := TextInput{ID: "input", State: NewTextInputState("hello"), Style: Style{Width: Cells(10)}}
	AssertSnapshot(t, widget, 10, 1, "Focused input at full pulse brightness: reverse-video cursor after 'hello'.")
}

// Test through the public widgets and clock: only the cursor's cell may change,
// even when its widget is auto-sized, contains a wide glyph, or has no content.
func TestCursorPulse_CellOnly(t *testing.T) {
	originalTheme := CurrentThemeName()
	RegisterTheme(t.Name(), ExtendTheme(ThemeNameRosePine, WithCursor(RGB(160, 100, 240))))
	SetTheme(t.Name())
	t.Cleanup(func() {
		SetTheme(originalTheme)
		delete(themeRegistry, t.Name())
	})
	for _, kind := range []string{"input", "area"} {
		for _, sizing := range []string{"fixed", "auto"} {
			for _, text := range []string{"", "ab界d"} {
				t.Run(kind+"/"+sizing+"/"+text, func(t *testing.T) {
					enableCursorBlinkForTest(t)
					style := Style{ForegroundColor: RGB(220, 200, 160), BackgroundColor: RGB(20, 40, 60)}
					if sizing == "fixed" {
						style.Width = Cells(12)
						style.Height = Cells(2)
					}
					highlights := 0
					highlighter := HighlighterFunc(func(string, []string) []TextHighlight {
						highlights++
						return nil
					})
					var widget Widget
					if kind == "input" {
						widget = TextInput{ID: "editor", State: NewTextInputState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
					} else {
						widget = TextArea{ID: "editor", State: NewTextAreaState(text), Placeholder: "Hint", Style: style, Highlighter: highlighter}
					}
					p := NewPilot(t, Column{Children: []Widget{widget, Text{Content: "unchanged sibling"}}}, 24, 5)
					positions := []int{0}
					if text != "" {
						positions = []int{5, 2} // End of text, then the two-cell glyph.
					}
					for _, x := range positions {
						if x == 2 {
							p.Press("left", "left")
						}
						if kind == "input" && sizing == "auto" && text != "" {
							// Auto input is five cells wide; it scrolls one cell
							// to make room for the cursor after the five-cell text.
							x--
						}
						before := make([]uv.Cell, 24*5)
						for y := range 5 {
							for col := range 24 {
								before[y*24+col] = *p.Buffer().CellAt(col, y)
							}
						}
						assert.Equal(t, RGB(160, 100, 240), FromANSI(before[x].Style.Fg), "reverse-video cursor uses the theme color, not the text color")
						fullFrames := p.session.renderer.Stats().FullRenderCount
						highlightsBefore := highlights
						p.Advance(1200 * time.Millisecond) // 400 ms hold + half a cycle.
						stats := p.session.renderer.Stats()
						assert.Equal(t, "cursor", stats.FrameMode)
						assert.Zero(t, stats.BuildCount)
						assert.Zero(t, stats.LayoutCount)
						assert.Zero(t, stats.PaintCount)
						assert.Equal(t, highlightsBefore, highlights, "pulses must not rerun highlighting")
						assert.Equal(t, fullFrames, stats.FullRenderCount)
						cell := p.Buffer().CellAt(x, 0)
						require.Equal(t, before[x].Content, cell.Content)
						assert.Equal(t, before[x].Width, cell.Width)
						assert.Equal(t, RGB(20, 40, 60), FromANSI(cell.Style.Fg))
						assert.Equal(t, RGB(69, 61, 123), FromANSI(cell.Style.Bg))
						for y := range 5 {
							for col := range 24 {
								if y == 0 && col >= x && col < x+cell.Width {
									continue
								}
								assert.Equal(t, before[y*24+col], *p.Buffer().CellAt(col, y), "non-cursor cell (%d,%d)", col, y)
							}
						}
					}
				})
			}
		}
	}
}

func TestCursorPulse_ScrollingAndOverlays(t *testing.T) {
	enableCursorBlinkForTest(t)
	state := NewTextAreaState(strings.Repeat("row of text\n", 1000))
	state.CursorIndex.Set(0)
	scroll := NewScrollState()
	covered := NewSignal(false)
	style := Style{ForegroundColor: RGB(220, 200, 160), BackgroundColor: RGB(20, 40, 60)}
	highlights := 0
	root := pilotBuilder(func(ctx BuildContext) Widget {
		coverStyle := style
		coverStyle.ForegroundColor = ctx.Theme().Cursor
		coverStyle.Reverse = true
		return Column{Children: []Widget{
			Scrollable{State: scroll, Width: Cells(20), Height: Cells(4), Child: TextArea{
				ID: "editor", State: state, ScrollState: scroll, Width: Flex(1), Style: style,
				Highlighter: HighlighterFunc(func(string, []string) []TextHighlight { highlights++; return nil }),
			}},
			Floating{Visible: covered.Get(), Config: FloatConfig{Position: FloatPositionTopLeft, PointerPassthrough: true},
				Child: Text{Content: "r", Style: coverStyle}},
		}}
	})
	p := NewPilot(t, root, 24, 6)
	buffer := p.Buffer()
	bright := *buffer.CellAt(0, 0)
	initialHighlights := highlights
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(0, 0).Style)
	assert.Equal(t, initialHighlights, highlights, "a large scrolled document is not highlighted on pulse ticks")
	assert.Zero(t, p.session.renderer.Stats().PaintCount)

	scroll.Offset.Set(10)
	p.ScreenText()
	offscreen := *buffer.CellAt(0, 0)
	initialHighlights = highlights
	p.Advance(2 * time.Second)
	assert.Equal(t, offscreen, *buffer.CellAt(0, 0), "no stale cursor after scrolling it out of view")
	assert.Equal(t, initialHighlights, highlights)
	scroll.Offset.Set(0)
	p.ScreenText()
	assert.Equal(t, bright, *buffer.CellAt(0, 0))
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(0, 0).Style, "scrolling back resumes the pulse")

	covered.Set(true)
	p.ScreenText()
	require.Equal(t, bright, *buffer.CellAt(0, 0), "the covering cell deliberately looks identical to the full-brightness cursor")
	p.Advance(2 * time.Second)
	assert.Equal(t, bright, *buffer.CellAt(0, 0), "a passive overlay must not inherit the underlying cursor pulse")
	covered.Set(false)
	p.ScreenText()
	p.Press("right")
	oldPosition, newPosition := *buffer.CellAt(0, 0), *buffer.CellAt(1, 0)
	p.Advance(1200 * time.Millisecond)
	assert.Equal(t, oldPosition, *buffer.CellAt(0, 0), "movement discards the old cached position")
	assert.NotEqual(t, newPosition.Style, buffer.CellAt(1, 0).Style)
	p.Resize(18, 5)
	newPosition = *buffer.CellAt(1, 0)
	p.Advance(400 * time.Millisecond)
	assert.NotEqual(t, newPosition.Style, buffer.CellAt(1, 0).Style, "resize refreshes the cached cell")
}

func TestCursorPulse_ImageCanvas(t *testing.T) {
	enableCursorBlinkForTest(t)
	p := NewPilot(t, Row{Children: []Widget{
		TextInput{ID: "input", State: NewTextInputState("hi"), Width: Cells(10)},
		Image{Source: testImage(t, 2, 2), Style: Style{Width: Cells(2), Height: Cells(2)}},
	}}, 12, 2)
	buffer := p.Buffer()
	bright, imageCell := *buffer.CellAt(2, 0), *buffer.CellAt(10, 0)
	p.Advance(1200 * time.Millisecond)
	assert.NotEqual(t, bright.Style, buffer.CellAt(2, 0).Style, "the cursor reaches the presentation buffer without copying the image canvas")
	assert.Equal(t, imageCell, *buffer.CellAt(10, 0))
	assert.Zero(t, p.session.renderer.Stats().PaintCount)
}

func TestCursorPulse_ThemeChanges(t *testing.T) {
	for _, kind := range []string{"input", "area"} {
		t.Run(kind, func(t *testing.T) {
			originalTheme := CurrentThemeName()
			t.Cleanup(func() { SetTheme(originalTheme) })
			SetTheme(ThemeNameAurora)
			style := Style{BackgroundColor: RGB(20, 40, 60)}
			var widget Widget = TextInput{ID: "editor", State: NewTextInputState("hi"), Width: Cells(10), Style: style}
			if kind == "area" {
				widget = TextArea{ID: "editor", State: NewTextAreaState("hi"), Width: Cells(10), Style: style}
			}
			p := NewPilot(t, widget, 10, 2)
			assert.Equal(t, Hex("#45FFB3"), FromANSI(p.Buffer().CellAt(2, 0).Style.Fg), "steady cursor uses Aurora's caret color")
			enableCursorBlinkForTest(t)
			p.Advance(1200 * time.Millisecond)
			SetTheme(ThemeNameMidnightEmber)
			p.ScreenText()  // Repaint after the theme changes, refreshing the palette.
			p.Press("left") // Restart at full brightness over a glyph.
			assert.Equal(t, Hex("#FF9A5C"), FromANSI(p.Buffer().CellAt(1, 0).Style.Fg))
			p.Advance(1200 * time.Millisecond)
			assert.Equal(t, RGB(102, 80, 71), FromANSI(p.Buffer().CellAt(1, 0).Style.Bg), "35% Midnight Ember cursor over the explicit widget background")
			assert.Zero(t, p.session.renderer.Stats().PaintCount)
		})
	}
}

func BenchmarkCursorPulse(b *testing.B) {
	for _, lines := range []int{10, 10000} {
		b.Run(fmt.Sprintf("%d_lines", lines), func(b *testing.B) {
			enableCursorBlinkForTest(b)
			state := NewTextAreaState(strings.Repeat("a line of text\n", lines))
			state.CursorIndex.Set(0)
			scroll := NewScrollState()
			p := NewPilot(b, Scrollable{State: scroll, Width: Cells(80), Height: Cells(24), Child: TextArea{
				ID: "editor", State: state, ScrollState: scroll, Width: Flex(1),
			}}, 80, 24)
			p.Advance(time.Second) // Warm the cursor and damage bookkeeping.
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				p.now = p.now.Add(cursorPulseInterval)
				p.session.advanceCursorPulse()
			}
		})
	}
}
